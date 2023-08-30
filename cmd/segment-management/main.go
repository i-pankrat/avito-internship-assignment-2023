package main

import (
	"fmt"
	"log"
	"net/http"

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
	connectionStr := createDBConnectionString(&cfg.Postgres)
	storage, err := postgresql.New(connectionStr)

	if err != nil {
		log.Fatal("Can not connect to database")
	}

	router := chi.NewRouter()

	router.Route("/segments", func(r chi.Router) {
		r.Post("/", add.New(storage))
		r.Delete("/{slug}", delete.New(storage))
	})

	router.Route("/user", func(r chi.Router) {
		r.Get("/{user_id}", get.New(storage))
		r.Post("/", change.New(storage))
	})

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Can not start server")
	}

}

func createDBConnectionString(pc *config.Postgres) string {
	// urlExample := "postgres://username:password@localhost:5432/dbname"
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pc.Username, pc.Password, pc.Host, pc.Port, pc.DBName)
}
