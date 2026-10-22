package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetRouteModeHandler(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"dns": {"final": "dns-direct"},
		"route": {"final": "direct"}
	}`), 0o600))
	t.Setenv("CONFIG_PATH", configPath)

	router := http.NewServeMux()
	router.Handle("PUT /route/mode", SetRouteModeHandler())

	invalidResponse := serveJSONRequest(router, http.MethodPut, "/route/mode?norestart=true", `{"mode":"block"}`)
	assert.Equal(t, http.StatusBadRequest, invalidResponse.Code)
	assert.Contains(t, invalidResponse.Body.String(), "route mode must be 'proxy' or 'direct'")

	proxyResponse := serveJSONRequest(router, http.MethodPut, "/route/mode?norestart=true", `{"mode":"proxy"}`)
	require.Equal(t, http.StatusNoContent, proxyResponse.Code, proxyResponse.Body.String())
	assertFinals(t, configPath, "proxy", "dns-remote")

	directResponse := serveJSONRequest(router, http.MethodPut, "/route/mode?norestart=true", `{"mode":"direct"}`)
	require.Equal(t, http.StatusNoContent, directResponse.Code, directResponse.Body.String())
	assertFinals(t, configPath, "direct", "dns-direct")
}

func assertFinals(t *testing.T, configPath, expectedRoute, expectedDNS string) {
	t.Helper()

	contents, err := os.ReadFile(configPath)
	require.NoError(t, err)
	var saved struct {
		DNS struct {
			Final string `json:"final"`
		} `json:"dns"`
		Route struct {
			Final string `json:"final"`
		} `json:"route"`
	}
	require.NoError(t, json.Unmarshal(contents, &saved))
	assert.Equal(t, expectedRoute, saved.Route.Final)
	assert.Equal(t, expectedDNS, saved.DNS.Final)
}
