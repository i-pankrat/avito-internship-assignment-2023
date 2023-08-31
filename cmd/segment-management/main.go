package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/i-pankrat/avito-internship-assignment-2023/http-server/handlers/segments/add"
	"github.com/i-pankrat/avito-internship-assignment-2023/http-server/handlers/segments/delete"
	"github.com/i-pankrat/avito-internship-assignment-2023/http-server/handlers/user/change"
	"github.com/i-pankrat/avito-internship-assignment-2023/http-server/handlers/user/get"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/config"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage/postgresql"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoadConfig()
	log := setupLogger(cfg.Env)
	connectionStr := createDBConnectionString(&cfg.Postgres)
	log.Info("connection string", slog.String("ConStr", connectionStr))
	storage, err := postgresql.New(connectionStr)

	if err != nil {
		log.Error("can not connect to database")
		os.Exit(1)
	}

	log.Info("connected to db")

	router := chi.NewRouter()

	router.Route("/segments", func(r chi.Router) {
		r.Post("/", add.New(log, storage))
		r.Delete("/{slug}", delete.New(log, storage))
	})

	router.Route("/user", func(r chi.Router) {
		r.Get("/{user_id}", get.New(log, storage))
		r.Post("/", change.New(log, storage))
	})

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      router,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("starting server")
	go storage.StartTTLChecker(log, cfg.TTLCheckerSeconds)

	if err := srv.ListenAndServe(); err != nil {
		log.Error("can not start server")
		os.Exit(1)
	}
}

func createDBConnectionString(pc *config.Postgres) string {
	// urlExample := "postgres://username:password@localhost:5432/dbname"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pc.Username, pc.Password, pc.Host, pc.Port, pc.DBName)
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envDev:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return logger
}
