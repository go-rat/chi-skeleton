package transport_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-rio/rio"
	"github.com/libtnb/assert/must"

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
	must.Equal(t, status, http.StatusNotFound)
	must.Contains(t, body, "not found")
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
		must.Equal(t, status, want, must.Msgf("kind %s", kind))
		must.Contains(t, body, "mod.code")
		must.Contains(t, body, "public detail")
		must.NotContains(t, body, "internal detail")
	}
}

func TestErrorFromUnknownErrorHidesDetails(t *testing.T) {
	status, body := respond(t, errors.New("password=hunter2 exploded"))
	must.Equal(t, status, http.StatusInternalServerError)
	must.NotContains(t, body, "hunter2")
}
