// Package service adapts HTTP and CLI to the usecase.
package service

import (
	"net/http"

	"github.com/libtnb/validator"

	"github.com/libtnb/chi-skeleton/internal/shared/transport"
	"github.com/libtnb/chi-skeleton/internal/user/biz"
)

type UserService struct {
	user     *biz.UserUsecase
	validate *validator.Validator
}

func NewUserService(user *biz.UserUsecase, validate *validator.Validator) *UserService {
	return &UserService{
		user:     user,
		validate: validate,
	}
}

func (r *UserService) List(w http.ResponseWriter, req *http.Request) {
	paginate, err := transport.Bind[transport.Paginate](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	users, total, err := r.user.List(req.Context(), paginate.Page, paginate.Limit)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, transport.Page[*biz.User]{
		Total: total,
		Items: users,
	})
}

func (r *UserService) Get(w http.ResponseWriter, req *http.Request) {
	userID, err := transport.Bind[UserID](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	user, err := r.user.Get(req.Context(), userID.ID)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, user)
}

func (r *UserService) Create(w http.ResponseWriter, req *http.Request) {
	userAdd, err := transport.Bind[UserAdd](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	user, err := r.user.Create(req.Context(), userAdd.Name)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, user)
}

func (r *UserService) Update(w http.ResponseWriter, req *http.Request) {
	userUpdate, err := transport.Bind[UserUpdate](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	user, err := r.user.Update(req.Context(), userUpdate.ID, userUpdate.Name)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, user)
}

func (r *UserService) Delete(w http.ResponseWriter, req *http.Request) {
	userID, err := transport.Bind[UserID](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if err = r.user.Delete(req.Context(), userID.ID); err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success[any](w, nil)
}
