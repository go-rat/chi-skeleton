package service_test

import (
	"reflect"
	"testing"

	"github.com/libtnb/validator"
	"github.com/stretchr/testify/assert"

	"github.com/libtnb/chi-skeleton/internal/shared/transport"
	"github.com/libtnb/chi-skeleton/internal/user/service"
)

// TestCheckRules catches invalid validate tags at test time.
func TestCheckRules(t *testing.T) {
	v := validator.MustNew()

	for _, req := range []any{
		transport.Paginate{},
		service.UserID{},
		service.UserAdd{},
		service.UserUpdate{},
	} {
		assert.NoError(t, v.CheckType(reflect.TypeOf(req)), "%T has an invalid validate tag", req)
	}
}
