package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const outboundRequestBody = `{
	"tag": "frankfurt-vless",
	"url": "vless://123e4567-e89b-12d3-a456-426614174000@203.0.113.10:443?type=tcp&security=reality&pbk=public-key&fp=firefox&sni=example.com&sid=0123456789abcdef&spx=%2F&flow=xtls-rprx-vision"
}`

func TestOutboundHandlers(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"outbounds": [
			{
				"type": "selector",
				"tag": "proxy",
				"outbounds": ["existing-vless"],
				"default": "existing-vless"
			},
			{
				"type": "vless",
				"tag": "existing-vless"
			}
		]
	}`), 0o600))
	t.Setenv("CONFIG_PATH", configPath)

	router := http.NewServeMux()
	router.Handle("POST /outbounds", AddOutboundHandler())
	router.Handle("DELETE /outbounds/{tag}", RemoveOutboundHandler())

	invalidResponse := serveOutboundRequest(router, http.MethodPost, "/outbounds?norestart=true", `{"tag":"broken","url":"https://example.com"}`)
	assert.Equal(t, http.StatusBadRequest, invalidResponse.Code)
	assert.Contains(t, invalidResponse.Body.String(), "scheme must be 'vless'")

	createResponse := serveOutboundRequest(router, http.MethodPost, "/outbounds?norestart=true", outboundRequestBody)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())

	contents, err := os.ReadFile(configPath)
	require.NoError(t, err)
	var saved struct {
		Outbounds []json.RawMessage `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(contents, &saved))
	require.Len(t, saved.Outbounds, 3)
	var selector struct {
		Outbounds []string `json:"outbounds"`
		Default   string   `json:"default"`
	}
	require.NoError(t, json.Unmarshal(saved.Outbounds[0], &selector))
	assert.Equal(t, []string{"existing-vless", "frankfurt-vless"}, selector.Outbounds)
	assert.Equal(t, "existing-vless", selector.Default)
	assert.JSONEq(t, `{
		"type": "vless",
		"tag": "frankfurt-vless",
		"server": "203.0.113.10",
		"server_port": 443,
		"uuid": "123e4567-e89b-12d3-a456-426614174000",
		"flow": "xtls-rprx-vision",
		"tls": {
			"enabled": true,
			"server_name": "example.com",
			"utls": {
				"enabled": true,
				"fingerprint": "firefox"
			},
			"reality": {
				"enabled": true,
				"public_key": "public-key",
				"short_id": "0123456789abcdef"
			}
		}
	}`, string(saved.Outbounds[2]))

	duplicateResponse := serveOutboundRequest(router, http.MethodPost, "/outbounds?norestart=true", outboundRequestBody)
	assert.Equal(t, http.StatusConflict, duplicateResponse.Code)
	assert.Contains(t, duplicateResponse.Body.String(), "outbound tag 'frankfurt-vless' already exists")

	deleteResponse := serveOutboundRequest(router, http.MethodDelete, "/outbounds/frankfurt-vless?norestart=true", "")
	require.Equal(t, http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())

	contents, err = os.ReadFile(configPath)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(contents, &saved))
	require.Len(t, saved.Outbounds, 2)
	require.NoError(t, json.Unmarshal(saved.Outbounds[0], &selector))
	assert.Equal(t, []string{"existing-vless"}, selector.Outbounds)

	missingResponse := serveOutboundRequest(router, http.MethodDelete, "/outbounds/frankfurt-vless?norestart=true", "")
	assert.Equal(t, http.StatusNotFound, missingResponse.Code)
	assert.Contains(t, missingResponse.Body.String(), "outbound tag 'frankfurt-vless' was not found")
}

func serveOutboundRequest(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
