package transport

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-rio/rio"
	"github.com/libtnb/chix/v2"
	"github.com/samber/oops"

	"github.com/libtnb/chi-skeleton/internal/shared/apperr"
)

// Envelope is the one response shape, typed so routes can document bodies.
type Envelope[T any] struct {
	Msg  string `json:"msg"`
	Code string `json:"code,omitempty"`
	Data T      `json:"data,omitempty"`
}

// Page is the typed payload of list responses.
type Page[T any] struct {
	Total int64 `json:"total"`
	Items []T   `json:"items"`
}

func Success[T any](w http.ResponseWriter, data T) {
	render := chix.NewRender(w)
	defer render.Release()
	render.JSON(&Envelope[T]{
		Msg:  "success",
		Data: data,
	})
}

func Error(w http.ResponseWriter, code int, format string, args ...any) {
	render := chix.NewRender(w)
	defer render.Release()
	render.Header(chix.HeaderContentType, chix.MIMEApplicationJSONCharsetUTF8) // must before Status()
	render.Status(code)
	render.JSON(&Envelope[any]{
		Msg: fmt.Sprintf(format, args...),
	})
}

// ErrorSystem writes a generic 500 without leaking details.
func ErrorSystem(w http.ResponseWriter) {
	render := chix.NewRender(w)
	defer render.Release()
	render.Header(chix.HeaderContentType, chix.MIMEApplicationJSONCharsetUTF8) // must before Status()
	render.Status(http.StatusInternalServerError)
	render.JSON(&Envelope[any]{
		Msg: http.StatusText(http.StatusInternalServerError),
	})
}

// ErrorFrom maps known errors to their status; anything else logs and 500s.
func ErrorFrom(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, rio.ErrNotFound) {
		Error(w, http.StatusNotFound, "not found")
		return
	}

	if status := statusFromKind(apperr.KindOf(err)); status != 0 {
		render := chix.NewRender(w)
		defer render.Release()
		render.Header(chix.HeaderContentType, chix.MIMEApplicationJSONCharsetUTF8)
		render.Status(status)
		render.JSON(&Envelope[any]{
			Msg:  oops.GetPublic(err, http.StatusText(status)),
			Code: apperr.CodeOf(err),
		})
		return
	}

	slog.ErrorContext(r.Context(), "request failed",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Any("error", err),
	)
	ErrorSystem(w)
}

// statusFromKind maps kinds to statuses; 0 = no kind, do not expose.
func statusFromKind(kind apperr.Kind) int {
	switch kind {
	case apperr.KindInvalid:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindUnprocessable:
		return http.StatusUnprocessableEntity
	default:
		return 0
	}
}
