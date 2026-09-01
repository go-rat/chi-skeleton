package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	"github.com/libtnb/chi-skeleton/internal/shared/registry"
	"github.com/libtnb/chi-skeleton/internal/shared/transport"
)

type documentRequest struct {
	ID uint `uri:"id" validate:"required && number"`
}

type documentResponse struct {
	CreatedAt time.Time `json:"created_at"`
}

func requireMap(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := parent[key]
	must.True(t, ok, must.Msgf("missing key %q", key))
	result, ok := value.(map[string]any)
	must.True(t, ok, must.Msgf("%q is %T, not an object", key, value))
	return result
}

func TestSpecJSONUsesTypedSchemasAndNoBodyResponse(t *testing.T) {
	routes := registry.Routes{{
		{Method: http.MethodGet, Path: "/things/{id}", Summary: "Get thing", Tags: []string{"thing"},
			Document: transport.Describe[documentRequest, documentResponse](http.StatusOK)},
		{Method: http.MethodDelete, Path: "/things/{id}", Summary: "Delete thing", Tags: []string{"thing"},
			Document: transport.DescribeNoBody[documentRequest](http.StatusNoContent)},
	}}

	spec, err := SpecJSON("test", "v1", validator.MustNew(), routes)
	must.NoError(t, err)
	must.Contains(t, string(spec), `"format": "date-time"`)

	var document map[string]any
	must.NoError(t, json.Unmarshal(spec, &document))
	version, _ := requireMap(t, document, "info")["version"].(string)
	must.Equal(t, version, "v1")
	path := requireMap(t, requireMap(t, document, "paths"), "/things/{id}")
	getResponse := requireMap(t, requireMap(t, requireMap(t, path, "get"), "responses"), "200")
	must.MapContains(t, getResponse, "content")
	deleteResponse := requireMap(t, requireMap(t, requireMap(t, path, "delete"), "responses"), "204")
	must.NotMapContains(t, deleteResponse, "content")
}

func TestSpecJSONPropagatesGeneratorErrors(t *testing.T) {
	_, err := SpecJSON("", "v1", validator.MustNew(), nil)
	must.Error(t, err)

	routes := registry.Routes{{{
		Method:   http.MethodGet,
		Path:     "/things",
		Document: transport.Describe[documentRequest, documentResponse](0),
	}}}
	_, err = SpecJSON("test", "v1", validator.MustNew(), routes)
	must.Error(t, err)
}
