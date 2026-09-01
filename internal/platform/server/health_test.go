package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libtnb/assert/must"

	"github.com/libtnb/chi-skeleton/internal/shared/registry"
)

func TestRunHealthChecks(t *testing.T) {
	checks := registry.HealthChecks{
		{Name: "one", Check: func(context.Context) error { return nil }},
		{Name: "two", Check: func(context.Context) error { return nil }},
	}

	name, err := runHealthChecks(t.Context(), checks)
	must.NoError(t, err)
	must.Empty(t, name)
}

func TestRunHealthChecksReturnsNamedFailureAndCancelsSiblings(t *testing.T) {
	want := errors.New("down")
	finished := make(chan struct{})
	checks := registry.HealthChecks{
		{Name: "database", Check: func(context.Context) error { return want }},
		{Name: "slow", Check: func(ctx context.Context) error {
			<-ctx.Done()
			close(finished)
			return context.Cause(ctx)
		}},
	}

	name, err := runHealthChecks(t.Context(), checks)
	must.ErrorIs(t, err, want)
	must.Equal(t, name, "database")
	must.Eventually(t, func() bool {
		select {
		case <-finished:
			return true
		default:
			return false
		}
	}, must.Tick(time.Millisecond))
}

func TestRunHealthChecksHonorsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	checks := registry.HealthChecks{{Name: "slow", Check: func(ctx context.Context) error {
		<-ctx.Done()
		return context.Cause(ctx)
	}}}

	name, err := runHealthChecks(ctx, checks)
	must.ErrorIs(t, err, context.DeadlineExceeded)
	must.Equal(t, name, "readiness")
}
