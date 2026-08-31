package server

import (
	"log/slog"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-cz/httplog"
	"github.com/libtnb/sessions"
	sessionmiddleware "github.com/libtnb/sessions/middleware"

	"github.com/libtnb/chi-skeleton/internal/platform/conf"
)

func globalMiddlewares(config *conf.Config, log *slog.Logger, session *sessions.Manager) []func(http.Handler) http.Handler {
	handlers := []func(http.Handler) http.Handler{
		chimiddleware.Recoverer,
		chimiddleware.RequestSize(int64(config.HTTP.BodyLimit) << 10),
		chimiddleware.StripSlashes,
		chimiddleware.RequestID,
	}

	// CORS only when origins are explicitly allowed; empty = same-origin
	if len(config.HTTP.CorsOrigins) > 0 {
		handlers = append(handlers, cors.Handler(cors.Options{
			AllowedOrigins: config.HTTP.CorsOrigins,
		}))
	}

	return append(handlers,
		httplog.RequestLogger(log, &httplog.Options{
			Level:             slog.LevelInfo,
			LogRequestHeaders: []string{"User-Agent"},
			// probes are noise
			Skip: func(req *http.Request, respStatus int) bool {
				return isProbe(req.URL.Path)
			},
		}),
		requestIDAttr,
		chimiddleware.Compress(5),
		// a session per probe hit would be persisted and GCed for nothing
		skipProbes(sessionmiddleware.StartSession(session)),
	)
}

func isProbe(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func skipProbes(mw func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		wrapped := mw(next)
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if isProbe(req.URL.Path) {
				next.ServeHTTP(w, req)
				return
			}
			wrapped.ServeHTTP(w, req)
		})
	}
}

// requestIDAttr echoes X-Request-Id and attaches the id to the access log.
func requestIDAttr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if id := chimiddleware.GetReqID(req.Context()); id != "" {
			w.Header().Set("X-Request-Id", id)
			httplog.SetAttrs(req.Context(), slog.String("request_id", id))
		}
		next.ServeHTTP(w, req)
	})
}
