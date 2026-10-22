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

func TestSubscriptionHandlers(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	original := []byte(`{
		"outbounds": [
			{
				"type": "selector",
				"tag": "proxy",
				"outbounds": ["fixed-vless", "old-subscription"],
				"default": "old-subscription"
			},
			{
				"type": "vless",
				"tag": "fixed-vless",
				"server": "192.0.2.10"
			},
			{
				"type": "direct",
				"tag": "direct"
			},
			{
				"type": "vless",
				"tag": "old-subscription",
				"server": "192.0.2.20"
			}
		]
	}`)
	require.NoError(t, os.WriteFile(configPath, original, 0o600))
	t.Setenv("CONFIG_PATH", configPath)

	router := http.NewServeMux()
	router.Handle("GET /subscription/settings", GetSubscriptionSettingsHandler())
	router.Handle("PUT /subscription/settings", SetSubscriptionSettingsHandler())
	router.Handle("PUT /subscription", UpdateSubscriptionHandler())

	settingsResponse := serveJSONRequest(router, http.MethodPut, "/subscription/settings", `{
		"fixed_outbounds": ["fixed-vless"],
		"fallback_outbound": "fixed-vless"
	}`)
	require.Equal(t, http.StatusNoContent, settingsResponse.Code, settingsResponse.Body.String())

	getSettingsResponse := serveJSONRequest(router, http.MethodGet, "/subscription/settings", "")
	require.Equal(t, http.StatusOK, getSettingsResponse.Code, getSettingsResponse.Body.String())
	assert.JSONEq(t, `{
		"version": 1,
		"fixed_outbounds": ["fixed-vless"],
		"fallback_outbound": "fixed-vless"
	}`, getSettingsResponse.Body.String())

	updateResponse := serveJSONRequest(router, http.MethodPut, "/subscription?norestart=true", `{
		"version": 1,
		"outbounds": [
			{"type":"direct","tag":"direct"},
			{"type":"vless","tag":"📆 Осталось: 31 день"},
			{
				"type": "vless",
				"tag": "new-subscription",
				"server": "vpn.example.com",
				"server_port": 443,
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"packet_encoding": "xudp",
				"flow": "xtls-rprx-vision",
				"tls": {
					"enabled": true,
					"server_name": "example.com",
					"utls": {"enabled": true, "fingerprint": "firefox"},
					"reality": {"enabled": true, "public_key": "public-key", "short_id": "0123456789abcdef"}
				}
			}
		]
	}`)
	require.Equal(t, http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	assert.JSONEq(t, `{
		"added": 1,
		"removed": 1,
		"preserved": 3,
		"skipped": [
			{"tag":"direct","reason":"missing server"},
			{"tag":"📆 Осталось: 31 день","reason":"missing server"}
		],
		"default": "fixed-vless"
	}`, updateResponse.Body.String())

	contents, err := os.ReadFile(configPath)
	require.NoError(t, err)
	var saved struct {
		Outbounds []struct {
			Type      string   `json:"type"`
			Tag       string   `json:"tag"`
			Server    string   `json:"server"`
			Outbounds []string `json:"outbounds"`
			Default   string   `json:"default"`
		} `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(contents, &saved))
	require.Len(t, saved.Outbounds, 4)
	assert.Equal(t, []string{"proxy", "fixed-vless", "direct", "new-subscription"}, []string{
		saved.Outbounds[0].Tag,
		saved.Outbounds[1].Tag,
		saved.Outbounds[2].Tag,
		saved.Outbounds[3].Tag,
	})
	assert.Equal(t, []string{"fixed-vless", "new-subscription"}, saved.Outbounds[0].Outbounds)
	assert.Equal(t, "fixed-vless", saved.Outbounds[0].Default)
	assert.Equal(t, "vpn.example.com", saved.Outbounds[3].Server)

	backup, err := os.ReadFile(configPath + ".subscription.bak")
	require.NoError(t, err)
	assert.JSONEq(t, string(original), string(backup))

	beforeInvalidUpdate := append([]byte(nil), contents...)
	invalidResponse := serveJSONRequest(router, http.MethodPut, "/subscription?norestart=true", `{
		"version": 1,
		"outbounds": [{"type":"vless","tag":"No server"}]
	}`)
	assert.Equal(t, http.StatusBadRequest, invalidResponse.Code)
	contents, err = os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, beforeInvalidUpdate, contents)
}
