package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBootstrap_EchoRegisterAllAndPlugins(t *testing.T) {
	app, err := NewAppForTest()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	app.Echo().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestBootstrap_EchoSwaggerRoute(t *testing.T) {
	app, err := NewAppForTest()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	app.Echo().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestBootstrap_EchoSwaggerDocRoute(t *testing.T) {
	app, err := NewAppForTest()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	app.Echo().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"swagger": "2.0"`)
	for _, route := range []string{
		`"/auth/login"`,
		`"/users"`,
		`"/roles"`,
		`"/permissions"`,
		`"/menus"`,
		`"/dictionaries"`,
		`"/domains/me"`,
		`"/media"`,
		`"/api-keys"`,
		`"/schemas"`,
		`"/mcp/health"`,
		`"/logs/operations"`,
		`"/captcha"`,
		`"/init/status"`,
		`"/profile"`,
	} {
		require.Contains(t, w.Body.String(), route)
	}
	require.Contains(t, w.Body.String(), `"name": "Auth"`)
	require.Contains(t, w.Body.String(), `"name": "Users"`)
	require.Contains(t, w.Body.String(), `"name": "Media"`)
}
