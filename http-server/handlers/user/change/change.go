package change

import (
	"errors"
	"net/http"

	"log/slog"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	seg "github.com/i-pankrat/avito-internship-assignment-2023/internal/data/segment"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/logger/sl"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/response"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage"
)

type Response struct {
	response.Response
}

type Request struct {
	UserId           int64         `json:"user_id" validate:"required"`
	SegmentsToAdd    []seg.Segment `json:"segments_to_add,omitempty"`
	SegmentsToDelete []seg.Slug    `json:"segments_to_delete,omitempty"`
}

type UserSegmentChanger interface {
	ChangeUserSegments(id int64, segmentsToAdd []seg.Segment, segmentsToDelete []seg.Slug) error
}

func New(log *slog.Logger, usrSgmChanger UserSegmentChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.change.New"
		log = log.With(slog.String("op", op))

		var request Request
		err := render.DecodeJSON(r.Body, &request)

		if err != nil {
			log.Error("invalid json")
			render.JSON(w, r, response.Error("can not decode request"))
			return
		}

		if request.SegmentsToAdd == nil && request.SegmentsToDelete == nil {
			log.Error("nothing to do")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		if err := validator.New().Struct(request); err != nil {
			log.Error("can not validate request data")
			render.JSON(w, r, response.Error("can not validate request"))
			return
		}

		err = usrSgmChanger.ChangeUserSegments(request.UserId, request.SegmentsToAdd, request.SegmentsToDelete)

		if err != nil {

			if errors.Is(err, storage.ErrSegmentDoesNotExist) || errors.Is(err, storage.ErrUserHasAlreadyAddedToSegment) {
				log.Error("segment does not exist")
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			log.Error("unexpected error", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("change user segments", slog.Int("id", int(request.UserId)))
		render.JSON(w, r, Response{response.OK()})
	}
}
