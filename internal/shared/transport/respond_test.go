package transport_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-rio/rio"
	"github.com/stretchr/testify/require"

	"github.com/libtnb/chi-skeleton/internal/shared/apperr"
	"github.com/libtnb/chi-skeleton/internal/shared/transport"
)

func respond(t *testing.T, err error) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	transport.ErrorFrom(w, req, err)
	return w.Code, w.Body.String()
}

func TestErrorFromNotFound(t *testing.T) {
	status, body := respond(t, rio.ErrNotFound)
	require.Equal(t, http.StatusNotFound, status)
	require.Contains(t, body, "not found")
}

func TestErrorFromKinds(t *testing.T) {
	for kind, want := range map[apperr.Kind]int{
		apperr.KindInvalid:       http.StatusBadRequest,
		apperr.KindUnauthorized:  http.StatusUnauthorized,
		apperr.KindForbidden:     http.StatusForbidden,
		apperr.KindNotFound:      http.StatusNotFound,
		apperr.KindConflict:      http.StatusConflict,
		apperr.KindUnprocessable: http.StatusUnprocessableEntity,
	} {
		err := apperr.New(kind, "mod.code", "public detail").Errorf("internal detail")
		status, body := respond(t, err)
		require.Equal(t, want, status, "kind %s", kind)
		require.Contains(t, body, "mod.code")
		require.Contains(t, body, "public detail")
		require.NotContains(t, body, "internal detail")
	}
}

func TestErrorFromUnknownErrorHidesDetails(t *testing.T) {
	status, body := respond(t, errors.New("password=hunter2 exploded"))
	require.Equal(t, http.StatusInternalServerError, status)
	require.NotContains(t, body, "hunter2")
}
