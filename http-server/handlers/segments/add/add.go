package add

import (
	"net/http"

	"errors"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
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

func New(sgmAdder SegmentAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var request Request
		err := render.DecodeJSON(r.Body, &request)

		if err != nil {
			render.JSON(w, r, response.Error("can not decode request"))
			return
		}

		if err = validator.New().Struct(request); err != nil {
			render.JSON(w, r, response.Error("can not validate request"))
			return
		}

		_, err = sgmAdder.AddSegment(request.Slug)

		if err != nil {

			if errors.Is(err, storage.ErrSegmentExists) {
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			render.JSON(w, r, response.Error("internal error"))
			return
		}

		render.JSON(w, r, response.OK())
	}
}
