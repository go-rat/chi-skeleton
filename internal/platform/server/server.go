package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/libtnb/sessions"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/libtnb/chi-skeleton/internal/platform/conf"
	"github.com/libtnb/chi-skeleton/internal/shared/registry"
	"github.com/libtnb/chi-skeleton/internal/shared/transport"
)

func NewRouter(
	config *conf.Config,
	log *slog.Logger,
	session *sessions.Manager,
	validate *validator.Validator,
	routes registry.Routes,
	version Version,
) (*chi.Mux, error) {
	r := chi.NewRouter()
	r.Use(globalMiddlewares(config, log, session)...)

	HTTP(routes, r)

	if config.HTTP.Docs {
		spec, err := SpecJSON(config.App.Name, version, validate, routes)
		if err != nil {
			return nil, err
		}
		docs := openapi.DocsHTML(config.App.Name, "/openapi.json")
		r.Get("/openapi.json", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(spec)
		})
		r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(docs)
		})
	}

	// framework-level errors leave as JSON too
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		transport.Error(w, http.StatusNotFound, "%s", http.StatusText(http.StatusNotFound))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		transport.Error(w, http.StatusMethodNotAllowed, "%s", http.StatusText(http.StatusMethodNotAllowed))
	})

	return r, nil
}

func NewHTTP(config *conf.Config, router *chi.Mux) *http.Server {
	return &http.Server{
		Addr:           config.HTTP.Address,
		Handler:        http.AllowQuerySemicolons(router),
		MaxHeaderBytes: config.HTTP.HeaderLimit,
		ReadTimeout:    config.HTTP.ReadTimeout,
		WriteTimeout:   config.HTTP.WriteTimeout,
		IdleTimeout:    config.HTTP.IdleTimeout,
	}
}
