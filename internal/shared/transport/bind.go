package transport

import (
	"errors"
	"net/http"

	"github.com/libtnb/chix/v2"
	"github.com/libtnb/validator"
)

// Bind binds and validates the request against the given validator.
func Bind[T any](r *http.Request, v *validator.Validator) (*T, error) {
	req := new(T)

	binder := chix.NewBind(r)
	defer binder.Release()
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if r.ContentLength > 0 {
			if err := binder.Body(req); err != nil {
				return nil, err
			}
		}
	}
	if err := binder.Query(req); err != nil {
		return nil, err
	}
	if err := binder.URI(req); err != nil {
		return nil, err
	}

	if hook, ok := any(req).(WithPrepare); ok {
		if err := hook.Prepare(r); err != nil {
			return nil, err
		}
	}

	vd, err := v.Struct(req)
	if err != nil {
		return nil, err
	}
	if hook, ok := any(req).(WithRules); ok {
		for field, expr := range hook.Rules(r) {
			if err := vd.AddRules(field, expr); err != nil {
				return nil, err
			}
		}
	}
	if hook, ok := any(req).(WithFilters); ok {
		for field, filters := range hook.Filters(r) {
			if err := vd.AddFilters(field, filters); err != nil {
				return nil, err
			}
		}
	}
	if hook, ok := any(req).(WithMessages); ok {
		if messages := hook.Messages(r); messages != nil {
			if err := vd.AddMessages(messages); err != nil {
				return nil, err
			}
		}
	}

	if err = vd.ValidateAs(r.Context(), req); err != nil {
		if fields, ok := validator.AsErrors(err); ok {
			return nil, errors.New(fields.One())
		}
		return nil, err
	}

	return req, nil
}
