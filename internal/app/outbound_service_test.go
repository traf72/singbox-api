package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

const validVLESSURL = "vless://123e4567-e89b-12d3-a456-426614174000@203.0.113.10:443?type=tcp&security=reality&pbk=public-key&fp=firefox&sni=example.com&sid=0123456789abcdef&spx=%2F&flow=xtls-rprx-vision"

func TestOutboundRequestToVLESSOptions(t *testing.T) {
	request := &OutboundRequest{
		Tag: "  frankfurt-vless  ",
		URL: "  " + validVLESSURL + "  ",
	}

	actual, err := request.toVLESSOptions()

	require.Nil(t, err)
	assert.Equal(t, &config.VLESSOutboundOptions{
		Tag:         "frankfurt-vless",
		Server:      "203.0.113.10",
		ServerPort:  443,
		UUID:        "123e4567-e89b-12d3-a456-426614174000",
		Flow:        "xtls-rprx-vision",
		ServerName:  "example.com",
		Fingerprint: "firefox",
		PublicKey:   "public-key",
		ShortID:     "0123456789abcdef",
	}, actual)
}

func TestOutboundRequestToVLESSOptionsValidation(t *testing.T) {
	tests := []struct {
		name       string
		request    *OutboundRequest
		errMessage string
	}{
		{
			name:       "Empty tag",
			request:    &OutboundRequest{URL: validVLESSURL},
			errMessage: "outbound tag is empty",
		},
		{
			name:       "Empty URL",
			request:    &OutboundRequest{Tag: "test"},
			errMessage: "outbound URL is empty",
		},
		{
			name:       "Wrong scheme",
			request:    &OutboundRequest{Tag: "test", URL: "https://example.com"},
			errMessage: "scheme must be 'vless'",
		},
		{
			name:       "Invalid UUID",
			request:    &OutboundRequest{Tag: "test", URL: "vless://not-a-uuid@203.0.113.10:443?type=tcp&security=reality"},
			errMessage: "UUID is missing or invalid",
		},
		{
			name:       "Invalid server",
			request:    &OutboundRequest{Tag: "test", URL: "vless://123e4567-e89b-12d3-a456-426614174000@example.com:443?type=tcp&security=reality"},
			errMessage: "server must be a valid IP address",
		},
		{
			name:       "Missing port",
			request:    &OutboundRequest{Tag: "test", URL: "vless://123e4567-e89b-12d3-a456-426614174000@203.0.113.10?type=tcp&security=reality"},
			errMessage: "port must be between 1 and 65535",
		},
		{
			name:       "Unsupported transport",
			request:    &OutboundRequest{Tag: "test", URL: "vless://123e4567-e89b-12d3-a456-426614174000@203.0.113.10:443?type=grpc&security=reality"},
			errMessage: "type must be 'tcp'",
		},
		{
			name:       "Unsupported security",
			request:    &OutboundRequest{Tag: "test", URL: "vless://123e4567-e89b-12d3-a456-426614174000@203.0.113.10:443?type=tcp&security=tls"},
			errMessage: "security must be 'reality'",
		},
		{
			name:       "Missing public key",
			request:    &OutboundRequest{Tag: "test", URL: "vless://123e4567-e89b-12d3-a456-426614174000@203.0.113.10:443?type=tcp&security=reality&fp=firefox&sni=example.com&sid=01&flow=xtls-rprx-vision"},
			errMessage: "public key query parameter 'pbk' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := tt.request.toVLESSOptions()

			assert.Nil(t, actual)
			require.NotNil(t, err)
			assert.Contains(t, err.Msg(), tt.errMessage)
		})
	}
}
