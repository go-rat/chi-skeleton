// Package service adapts HTTP to the usecase.
package service

import (
	"net/http"

	"github.com/libtnb/validator"

	"github.com/libtnb/chi-skeleton/internal/order/biz"
	"github.com/libtnb/chi-skeleton/internal/shared/transport"
)

// OrderService handles the order HTTP endpoints.
type OrderService struct {
	order    *biz.OrderUsecase
	validate *validator.Validator
}

func NewOrderService(order *biz.OrderUsecase, validate *validator.Validator) *OrderService {
	return &OrderService{
		order:    order,
		validate: validate,
	}
}

func (r *OrderService) List(w http.ResponseWriter, req *http.Request) {
	paginate, err := transport.Bind[transport.Paginate](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	orders, total, err := r.order.List(req.Context(), paginate.Page, paginate.Limit)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, transport.Page[*biz.Order]{
		Total: total,
		Items: orders,
	})
}

func (r *OrderService) Get(w http.ResponseWriter, req *http.Request) {
	orderID, err := transport.Bind[OrderID](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	order, err := r.order.Get(req.Context(), orderID.ID)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, order)
}

func (r *OrderService) Create(w http.ResponseWriter, req *http.Request) {
	create, err := transport.Bind[OrderCreate](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	order, err := r.order.Place(req.Context(), create.UserID, create.Amount)
	if err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success(w, order)
}

func (r *OrderService) Delete(w http.ResponseWriter, req *http.Request) {
	orderID, err := transport.Bind[OrderID](req, r.validate)
	if err != nil {
		transport.Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	if err = r.order.Delete(req.Context(), orderID.ID); err != nil {
		transport.ErrorFrom(w, req, err)
		return
	}

	transport.Success[any](w, nil)
}
