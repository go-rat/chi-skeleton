package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/libtnb/chi-skeleton/internal/shared/registry"
)

func TestRunHealthChecks(t *testing.T) {
	checks := registry.HealthChecks{
		{Name: "one", Check: func(context.Context) error { return nil }},
		{Name: "two", Check: func(context.Context) error { return nil }},
	}

	name, err := runHealthChecks(t.Context(), checks)
	require.NoError(t, err)
	require.Empty(t, name)
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
	require.ErrorIs(t, err, want)
	require.Equal(t, "database", name)
	require.Eventually(t, func() bool {
		select {
		case <-finished:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond)
}

func TestRunHealthChecksHonorsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	checks := registry.HealthChecks{{Name: "slow", Check: func(ctx context.Context) error {
		<-ctx.Done()
		return context.Cause(ctx)
	}}}

	name, err := runHealthChecks(ctx, checks)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, "readiness", name)
}
