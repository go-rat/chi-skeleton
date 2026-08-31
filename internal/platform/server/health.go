package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/libtnb/chi-skeleton/internal/shared/registry"
	"github.com/libtnb/chi-skeleton/internal/shared/transport"
)

const readinessTimeout = 5 * time.Second

type healthResult struct {
	name string
	err  error
}

func runHealthChecks(ctx context.Context, checks registry.HealthChecks) (string, error) {
	if err := context.Cause(ctx); err != nil {
		return "readiness", err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan healthResult, len(checks))
	for _, check := range checks {
		go func() {
			if check.Check == nil {
				results <- healthResult{name: check.Name, err: errors.New("health check is nil")}
				return
			}
			results <- healthResult{name: check.Name, err: check.Check(ctx)}
		}()
	}

	for range checks {
		select {
		case result := <-results:
			if result.err != nil {
				cancel()
				return result.name, result.err
			}
		case <-ctx.Done():
			return "readiness", context.Cause(ctx)
		}
	}

	return "", nil
}

// HealthRoutes serves the probes; they stay out of the OpenAPI docs.
func HealthRoutes(checks registry.HealthChecks) transport.Endpoints {
	return transport.Endpoints{
		{Method: http.MethodGet, Path: "/", Handler: func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Hello, World 👋!"))
		}},
		{Method: http.MethodGet, Path: "/healthz", Handler: func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}},
		{Method: http.MethodGet, Path: "/readyz", Handler: func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
			defer cancel()
			if name, err := runHealthChecks(ctx, checks); err != nil {
				transport.Error(w, http.StatusServiceUnavailable, "%s unavailable", name)
				return
			}
			_, _ = w.Write([]byte("ok"))
		}},
	}
}
