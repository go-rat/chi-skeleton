package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/libtnb/assert/must"
)

// TestGraph builds both generated graphs and exercises their managed cleanup.
func TestGraph(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APP_CONFIG", "../../config/config.example.yml")
	t.Setenv("APP_DATABASE__PATH", filepath.Join(tmp, "test.db"))
	t.Setenv("APP_LOG__OUTPUT", "file")
	t.Setenv("APP_LOG__PATH", filepath.Join(tmp, "test.log"))

	application, cleanupApp, err := InitializeApp("test")
	must.NoError(t, err)
	must.NotNil(t, application)
	must.NoError(t, application.migrator.Up(t.Context()))

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	res := httptest.NewRecorder()
	application.server.Handler.ServeHTTP(res, req)
	must.Equal(t, res.Code, http.StatusOK)
	must.Contains(t, res.Body.String(), `"version": "test"`)
	must.Contains(t, res.Body.String(), `"/users/{id}"`)

	must.NoError(t, cleanupApp())
	must.NoError(t, cleanupApp())

	management, cleanupCLI, err := InitializeCLI()
	must.NoError(t, err)
	must.NotNil(t, management)
	must.NoError(t, cleanupCLI())
	must.NoError(t, cleanupCLI())
}
