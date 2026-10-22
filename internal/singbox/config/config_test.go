package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigV114RoundTrip(t *testing.T) {
	input := []byte(`{
		"dns": {
			"independent_cache": true
		},
		"inbounds": [{
			"type": "tun",
			"tag": "tun-in",
			"endpoint_independent_nat": true,
			"dns_mode": "hijack"
		}]
	}`)

	var conf Conf
	require.NoError(t, json.Unmarshal(input, &conf))
	require.Len(t, conf.Inbounds, 1)
	assert.Equal(t, "hijack", conf.Inbounds[0].DNSMode)

	output, err := json.Marshal(&conf)
	require.NoError(t, err)
	assert.NotContains(t, string(output), "independent_cache")
	assert.NotContains(t, string(output), "endpoint_independent_nat")
	assert.Contains(t, string(output), `"dns_mode":"hijack"`)
}

func TestInboundDNSModeIsOptional(t *testing.T) {
	input := []byte(`{"inbounds":[{"type":"tun","tag":"tun-in"}]}`)

	var conf Conf
	require.NoError(t, json.Unmarshal(input, &conf))

	output, err := json.Marshal(&conf)
	require.NoError(t, err)
	assert.NotContains(t, string(output), "dns_mode")
}

func TestServicesRoundTrip(t *testing.T) {
	services := `[
		{
			"type": "api",
			"tag": "management-api",
			"listen": "127.0.0.1",
			"listen_port": 9091,
			"secret": "test-secret",
			"dashboard": {
				"enabled": true
			}
		}
	]`
	input := []byte(`{"services":` + services + `}`)

	var conf Conf
	require.NoError(t, json.Unmarshal(input, &conf))
	require.Len(t, conf.Services, 1)
	assert.Equal(t, "api", conf.Services[0].Type)
	assert.Equal(t, "management-api", conf.Services[0].Tag)
	assert.Equal(t, "127.0.0.1", conf.Services[0].Listen)
	assert.Equal(t, 9091, conf.Services[0].ListenPort)
	assert.Equal(t, "test-secret", conf.Services[0].Secret)
	require.NotNil(t, conf.Services[0].Dashboard)
	assert.True(t, conf.Services[0].Dashboard.Enabled)

	output, err := json.Marshal(&conf)
	require.NoError(t, err)

	var root struct {
		Services json.RawMessage `json:"services"`
	}
	require.NoError(t, json.Unmarshal(output, &root))
	assert.JSONEq(t, services, string(root.Services))
}

func TestSelectorOutboundRoundTrip(t *testing.T) {
	selector := `{
		"type": "selector",
		"tag": "proxy",
		"outbounds": [
			"own-frankfurt-vless-grpc",
			"own-frankfurt-ss"
		],
		"default": "own-frankfurt-vless-grpc",
		"interrupt_exist_connections": true
	}`
	input := []byte(`{"outbounds":[` + selector + `]}`)

	var conf Conf
	require.NoError(t, json.Unmarshal(input, &conf))
	require.Len(t, conf.Outbounds, 1)
	assert.Equal(t, []string{"own-frankfurt-vless-grpc", "own-frankfurt-ss"}, conf.Outbounds[0].Outbounds)
	assert.Equal(t, "own-frankfurt-vless-grpc", conf.Outbounds[0].Default)
	assert.True(t, conf.Outbounds[0].InterruptExistConnections)

	output, err := json.Marshal(&conf)
	require.NoError(t, err)

	var root struct {
		Outbounds []json.RawMessage `json:"outbounds"`
	}
	require.NoError(t, json.Unmarshal(output, &root))
	require.Len(t, root.Outbounds, 1)
	assert.JSONEq(t, selector, string(root.Outbounds[0]))
}
