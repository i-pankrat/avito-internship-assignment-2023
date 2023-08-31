package delete

import (
	"errors"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/logger/sl"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/response"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage"
)

type Response struct {
	response.Response
}

type SegmentDeleter interface {
	RemoveSegment(slug string) error
}

func New(log *slog.Logger, sgmDeleter SegmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.delete.New"
		log = log.With(slog.String("op", op))

		slug := chi.URLParam(r, "slug")

		if len(slug) == 0 {
			log.Error("slug has zero legnth")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid slug"))
			return
		}

		err := sgmDeleter.RemoveSegment(slug)

		if err != nil {

			if errors.Is(err, storage.ErrSegmentDoesNotExist) {
				log.Error("segment does not exist")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			log.Error("unexpected error", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("delete segment", slog.String("slug", slug))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, response.OK())
	}
}
