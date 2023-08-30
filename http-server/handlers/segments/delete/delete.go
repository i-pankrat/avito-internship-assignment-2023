package delete

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/response"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage"
)

type Response struct {
	response.Response
}

type SegmentDeleter interface {
	RemoveSegment(slug string) error
}

func New(sgmDeleter SegmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		slug := chi.URLParam(r, "slug")

		if len(slug) == 0 {
			render.JSON(w, r, response.Error("invalid slug"))
			return
		}

		err := sgmDeleter.RemoveSegment(slug)

		if err != nil {

			if errors.Is(err, storage.ErrSegmentDoesNotExist) {
				render.JSON(w, r, response.Error(err.Error()))
				return
			}
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		render.JSON(w, r, response.OK())
	}
}
