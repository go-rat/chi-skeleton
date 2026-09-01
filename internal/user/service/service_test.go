package service_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	mocksbiz "github.com/libtnb/chi-skeleton/internal/mocks/user/biz"
	"github.com/libtnb/chi-skeleton/internal/user/biz"
	"github.com/libtnb/chi-skeleton/internal/user/service"
)

// newTestRouter wires the service against a mocked repo and a real validator.
func newTestRouter(t *testing.T) (*chi.Mux, *mocksbiz.UserRepo) {
	t.Helper()

	repo := &mocksbiz.UserRepo{}
	user := service.NewUserService(biz.NewUserUsecase(repo), validator.MustNew())

	router := chi.NewRouter()
	router.Get("/users", user.List)
	router.Post("/users", user.Create)
	router.Get("/users/{id}", user.Get)
	router.Put("/users/{id}", user.Update)
	router.Delete("/users/{id}", user.Delete)

	return router, repo
}

func do(router *chi.Mux, method, target, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestUserList(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.ListFunc = func(context.Context, int, int) ([]*biz.User, int64, error) {
		return []*biz.User{{ID: 1, Name: "alice"}}, 1, nil
	}

	w := do(router, http.MethodGet, "/users", "")

	check.Equal(t, w.Code, http.StatusOK)
	check.Contains(t, w.Body.String(), "alice")
	listed := repo.ListCalls()
	must.Len(t, listed, 1)
	// the paginate defaults reach the usecase
	check.Equal(t, listed[0].Page, 1)
	check.Equal(t, listed[0].Limit, 10)
}

func TestUserGet(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.GetFunc = func(context.Context, uint) (*biz.User, error) {
		return &biz.User{ID: 1, Name: "alice"}, nil
	}

	w := do(router, http.MethodGet, "/users/1", "")

	check.Equal(t, w.Code, http.StatusOK)
	got := repo.GetCalls()
	must.Len(t, got, 1)
	check.Equal(t, got[0].ID, uint(1))
}

func TestUserGet_NotFoundMapsTo404(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.GetFunc = func(context.Context, uint) (*biz.User, error) {
		return nil, rio.ErrNotFound
	}

	w := do(router, http.MethodGet, "/users/9", "")

	check.Equal(t, w.Code, http.StatusNotFound)
}

func TestUserCreate(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.ExistsNameFunc = func(context.Context, string) (bool, error) { return false, nil }
	repo.CreateFunc = func(context.Context, *biz.User) error { return nil }

	w := do(router, http.MethodPost, "/users", `{"name":"alice"}`)

	check.Equal(t, w.Code, http.StatusOK)
	created := repo.CreateCalls()
	must.Len(t, created, 1)
	check.Equal(t, created[0].User.Name, "alice")
}

func TestUserCreate_NameTakenMapsToConflict(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.ExistsNameFunc = func(context.Context, string) (bool, error) { return true, nil }

	w := do(router, http.MethodPost, "/users", `{"name":"alice"}`)

	check.Equal(t, w.Code, http.StatusConflict)
}

func TestUserCreate_RejectsShortName(t *testing.T) {
	router, _ := newTestRouter(t) // no repo funcs: validation must fail first

	w := do(router, http.MethodPost, "/users", `{"name":"ab"}`)

	check.Equal(t, w.Code, http.StatusUnprocessableEntity)
}

func TestUserUpdate_NotFoundMapsTo404(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.UpdateFunc = func(context.Context, *biz.User) (*biz.User, error) {
		return nil, rio.ErrNotFound
	}

	w := do(router, http.MethodPut, "/users/9", `{"name":"alice"}`)

	check.Equal(t, w.Code, http.StatusNotFound)
	updated := repo.UpdateCalls()
	must.Len(t, updated, 1)
	check.Equal(t, updated[0].User.ID, uint(9))
	check.Equal(t, updated[0].User.Name, "alice")
}

func TestUserDelete(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.DeleteFunc = func(context.Context, uint) error { return nil }

	w := do(router, http.MethodDelete, "/users/1", "")

	check.Equal(t, w.Code, http.StatusOK)
	deleted := repo.DeleteCalls()
	must.Len(t, deleted, 1)
	check.Equal(t, deleted[0].ID, uint(1))
}
