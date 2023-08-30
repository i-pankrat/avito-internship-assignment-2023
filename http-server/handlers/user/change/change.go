package change

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/response"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/storage"
)

type Response struct {
	response.Response
}

type Request struct {
	UserId           int64    `json:"user_id" validate:"required"`
	SegmentsToAdd    []string `json:"segments_to_add,omitempty"`
	SegmentsToDelete []string `json:"segments_to_delete,omitempty"`
}

type UserSegmentChanger interface {
	ChangeUserSegments(id int64, segmentsToAdd, segmentsToDelete []string) error
}

func New(usrSgmChanger UserSegmentChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var request Request
		render.DecodeJSON(r.Body, &request)

		if request.SegmentsToAdd == nil && request.SegmentsToDelete == nil {
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		if err := validator.New().Struct(request); err != nil {
			render.JSON(w, r, response.Error("can not validate request"))
			return
		}

		err := usrSgmChanger.ChangeUserSegments(request.UserId, request.SegmentsToAdd, request.SegmentsToDelete)

		if err != nil {

			if errors.Is(err, storage.ErrSegmentDoesNotExist) || errors.Is(err, storage.ErrUserHasAlreadyAddedToSegment) {
				render.JSON(w, r, response.Error(err.Error()))
				return
			}

			render.JSON(w, r, response.Error("internal error"))
			return
		}

		render.JSON(w, r, Response{response.OK()})
	}
}
