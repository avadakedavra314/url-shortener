package save

import (
	"log/slog"
	"net/http"

	"github.com/avadakedavra314/url-shortener/internal/lib/api/response"
	"github.com/avadakedavra314/url-shortener/internal/lib/logger/sl"
	"github.com/avadakedavra314/url-shortener/internal/lib/random"
	"github.com/avadakedavra314/url-shortener/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UrlSaver
type UrlSaver interface {
	SaveUrl(urlToSave string, alias string) (int64, error)
}

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

func New(log *slog.Logger, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// validate struct Request by tags, like validate:"reqired"
		if err := validator.New().Struct(req); err != nil {
			log.Error("request is not valid", sl.Err(err))
			render.JSON(w, r, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveUrl(req.Url, alias)

		switch err {
		case nil:
			break
		case storage.ErrUrlExist:
			log.Info("url already exist", slog.String("url", req.Url))
			render.JSON(w, r, response.Error("url already exist"))
			return
		default:
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("url added", slog.Int64("id", id))
		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
