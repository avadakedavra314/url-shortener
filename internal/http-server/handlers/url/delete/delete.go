package delete

import (
	"log/slog"
	"net/http"

	"github.com/avadakedavra314/url-shortener/internal/lib/api/response"
	"github.com/avadakedavra314/url-shortener/internal/lib/logger/sl"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type UrlDeletter interface {
	DeleteUrl(urlAlias string) error
}

func New(log *slog.Logger, urlDeletter UrlDeletter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}
		err := urlDeletter.DeleteUrl(alias)
		if err != nil {
			log.Error("falied to delete url", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		render.JSON(w, r, response.OK())
	}

}
