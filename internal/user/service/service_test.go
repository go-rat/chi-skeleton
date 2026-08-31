package service_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-rio/rio"
	"github.com/libtnb/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksbiz "github.com/libtnb/chi-skeleton/internal/mocks/user/biz"
	"github.com/libtnb/chi-skeleton/internal/user/biz"
	"github.com/libtnb/chi-skeleton/internal/user/service"
)

// newTestRouter wires the service against a mocked repo and a real validator.
func newTestRouter(t *testing.T) (*chi.Mux, *mocksbiz.UserRepo) {
	t.Helper()

	repo := mocksbiz.NewUserRepo(t)
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
	repo.EXPECT().List(mock.Anything, 1, 10).
		Return([]*biz.User{{ID: 1, Name: "alice"}}, int64(1), nil)

	w := do(router, http.MethodGet, "/users", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alice")
}

func TestUserGet(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.EXPECT().Get(mock.Anything, uint(1)).
		Return(&biz.User{ID: 1, Name: "alice"}, nil)

	w := do(router, http.MethodGet, "/users/1", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserGet_NotFoundMapsTo404(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.EXPECT().Get(mock.Anything, uint(9)).
		Return(nil, rio.ErrNotFound)

	w := do(router, http.MethodGet, "/users/9", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserCreate(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.EXPECT().ExistsName(mock.Anything, "alice").Return(false, nil)
	repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *biz.User) bool {
		return u.Name == "alice"
	})).Return(nil)

	w := do(router, http.MethodPost, "/users", `{"name":"alice"}`)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserCreate_NameTakenMapsToConflict(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.EXPECT().ExistsName(mock.Anything, "alice").Return(true, nil)

	w := do(router, http.MethodPost, "/users", `{"name":"alice"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserCreate_RejectsShortName(t *testing.T) {
	router, _ := newTestRouter(t) // no repo expectations: validation must fail first

	w := do(router, http.MethodPost, "/users", `{"name":"ab"}`)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestUserUpdate_NotFoundMapsTo404(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *biz.User) bool {
		return u.ID == 9 && u.Name == "alice"
	})).Return(nil, rio.ErrNotFound)

	w := do(router, http.MethodPut, "/users/9", `{"name":"alice"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserDelete(t *testing.T) {
	router, repo := newTestRouter(t)
	repo.EXPECT().Delete(mock.Anything, uint(1)).Return(nil)

	w := do(router, http.MethodDelete, "/users/1", "")

	assert.Equal(t, http.StatusOK, w.Code)
}
