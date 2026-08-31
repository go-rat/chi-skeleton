package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/libtnb/validator"
	"github.com/stretchr/testify/require"

	"github.com/libtnb/chi-skeleton/internal/shared/transport"
)

type createReq struct {
	Name string `json:"name" validate:"required && min:3 && max:10"`
}

func bindOn[T any](t *testing.T, method, target, body string) (*T, int) {
	t.Helper()

	var bound *T
	router := chi.NewRouter()
	handler := func(w http.ResponseWriter, r *http.Request) {
		req, err := transport.Bind[T](r, validator.MustNew())
		if err != nil {
			transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
			return
		}
		bound = req
		transport.Success[any](w, nil)
	}
	router.HandleFunc("/bind", handler)
	router.HandleFunc("/bind/{id}", handler)

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return bound, w.Code
}

func TestBindBodyAndValidate(t *testing.T) {
	got, status := bindOn[createReq](t, http.MethodPost, "/bind", `{"name":"alice"}`)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "alice", got.Name)
}

func TestBindRejectsInvalid(t *testing.T) {
	_, status := bindOn[createReq](t, http.MethodPost, "/bind", `{"name":"ab"}`)
	require.Equal(t, http.StatusUnprocessableEntity, status)
}

func TestBindRunsPrepareHook(t *testing.T) {
	got, status := bindOn[transport.Paginate](t, http.MethodGet, "/bind", "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, got.Page)
	require.Equal(t, 10, got.Limit)

	got, status = bindOn[transport.Paginate](t, http.MethodGet, "/bind?page=3&limit=50", "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, got.Page)
	require.Equal(t, 50, got.Limit)
}

func TestBindQueryOverLimitFailsValidation(t *testing.T) {
	_, status := bindOn[transport.Paginate](t, http.MethodGet, "/bind?limit=5000", "")
	require.Equal(t, http.StatusUnprocessableEntity, status)
}

type uriReq struct {
	ID uint `uri:"id" validate:"required && number"`
}

func TestBindURI(t *testing.T) {
	got, status := bindOn[uriReq](t, http.MethodGet, "/bind/42", "")
	require.Equal(t, http.StatusOK, status)
	require.EqualValues(t, 42, got.ID)
}

func TestBindPropagatesCanceledContext(t *testing.T) {
	var bindErr error
	router := chi.NewRouter()
	router.Post("/bind", func(w http.ResponseWriter, r *http.Request) {
		_, bindErr = transport.Bind[createReq](r, validator.MustNew())
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{"name":"alice"}`)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), req)
	require.ErrorIs(t, bindErr, context.Canceled)
}
