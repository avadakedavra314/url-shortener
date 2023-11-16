package redirect

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/avadakedavra314/url-shortener/internal/lib/api/response"
	"github.com/avadakedavra314/url-shortener/internal/lib/logger/sl"
	"github.com/avadakedavra314/url-shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type UrlGetter interface {
	GetUrl(urlAlias string) (string, error)
}

func New(log *slog.Logger, urlGetter UrlGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.New"
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

		url, err := urlGetter.GetUrl(alias)

		switch err {
		case nil:
			break
		case storage.ErrUrlNotFound:
			log.Info("url not found", slog.String("alias", alias))
			render.JSON(w, r, response.Error(fmt.Sprintf("url not found for this alias: %s", alias)))
			return
		default:
			log.Error("falied to get url", sl.Err(err))
			render.JSON(w, r, "internal error")
			return
		}

		log.Info("retrieved url", slog.String("url", url))

		http.Redirect(w, r, url, http.StatusFound)
	}
}
