package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/avadakedavra314/url-shortener/internal/config"
	"github.com/avadakedavra314/url-shortener/internal/http-server/handlers/redirect"
	"github.com/avadakedavra314/url-shortener/internal/http-server/handlers/url/delete"
	"github.com/avadakedavra314/url-shortener/internal/http-server/handlers/url/save"
	"github.com/avadakedavra314/url-shortener/internal/http-server/middleware/mwlogger"
	"github.com/avadakedavra314/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/avadakedavra314/url-shortener/internal/lib/logger/sl"
	"github.com/avadakedavra314/url-shortener/internal/storage/sqlite"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)

	log := setupLogger(cfg.Env)
	log.Info(
		"starting url-shortener",
		slog.String("env", cfg.Env))
	log.Debug("debug logs are enabled")
	log.Error("error logs are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	_ = storage

	router := chi.NewRouter()
	router.Use(middleware.RequestID) // add request id to request context, usable to tracing
	router.Use(mwlogger.New(log))    // log incoming requests
	router.Use(middleware.Recoverer) // recover from panic, log panic and return HTTP 500(internal server error)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HttpServer.User: cfg.HttpServer.Password,
		}))

		r.Post("/url", save.New(log, storage))
		r.Delete("/url/{alias}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.HttpServer.Address))
	srv := http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}
	if err = srv.ListenAndServe(); err != nil {
		log.Error("falied to start server", sl.Err(err))
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		// log = slog.New(
		// 	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		// )
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
