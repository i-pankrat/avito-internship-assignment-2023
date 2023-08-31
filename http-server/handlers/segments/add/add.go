package add

import (
	"net/http"

	"errors"

	"log/slog"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/logger/sl"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/response"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage"
)

type Response struct {
	response.Response
}

type SegmentAdder interface {
	AddSegment(slug string) (int64, error)
}

type Request struct {
	Slug string `json:"slug" validate:"required"`
}

func New(log *slog.Logger, sgmAdder SegmentAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.segments.add.New"
		log = log.With(slog.String("op", op))

		var request Request
		err := render.DecodeJSON(r.Body, &request)

		if err != nil {
			log.Error("invalid json")
			render.JSON(w, r, response.Error("can not decode request"))
			return
		}

		if err = validator.New().Struct(request); err != nil {
			log.Error("can not validate request data")
			render.JSON(w, r, response.Error("can not validate request"))
			return
		}

		_, err = sgmAdder.AddSegment(request.Slug)

		if err != nil {

			if errors.Is(err, storage.ErrSegmentExists) {
				log.Error("segment has not added becaise it's already exist")
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			log.Error("unexpected error", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("add segment", slog.String("slug", request.Slug))
		render.JSON(w, r, response.OK())
	}
}
