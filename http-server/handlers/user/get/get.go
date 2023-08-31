package get

import (
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	seg "github.com/i-pankrat/avito-internship-assignment-2023/internal/data/segment"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/logger/sl"
	"github.com/i-pankrat/avito-internship-assignment-2023/internal/lib/api/response"
)

type Response struct {
	response.Response
	Segments []seg.Slug `json:"slugs"`
}

type UserSegmentGetter interface {
	GetUserSegments(id int64) ([]seg.Slug, error)
}

func New(log *slog.Logger, usrSgmGetter UserSegmentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.get.New"
		log = log.With(slog.String("op", op))

		idStr := chi.URLParam(r, "user_id")
		id, err := strconv.ParseInt(idStr, 10, 64)

		if err != nil {
			log.Error("invalid user id")
			render.JSON(w, r, response.Error("invalid user id"))
			return
		}

		sgms, err := usrSgmGetter.GetUserSegments(id)

		if err != nil {
			log.Error("unexpected error", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("return user segments", slog.Int("id", int(id)))
		render.JSON(w, r, Response{response.OK(), sgms})
	}
}
