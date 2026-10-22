package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traf72/singbox-api/internal/apperr"
)

func testVLESSOutboundOptions() *VLESSOutboundOptions {
	return &VLESSOutboundOptions{
		Tag:         "frankfurt-vless",
		Server:      "203.0.113.10",
		ServerPort:  443,
		UUID:        "123e4567-e89b-12d3-a456-426614174000",
		Flow:        "xtls-rprx-vision",
		ServerName:  "example.com",
		Fingerprint: "firefox",
		PublicKey:   "public-key",
		ShortID:     "0123456789abcdef",
	}
}

func TestAddVLESSOutbound(t *testing.T) {
	selector := &outbound{
		Type:      "selector",
		Tag:       "proxy",
		Outbounds: []string{"existing-vless"},
		Default:   "existing-vless",
	}
	conf := &Conf{Outbounds: []*outbound{
		selector,
		{Type: "direct", Tag: "direct"},
	}}

	err := addVLESSOutbound(testVLESSOutboundOptions(), conf)

	require.Nil(t, err)
	require.Len(t, conf.Outbounds, 3)
	assert.Equal(t, []string{"existing-vless", "frankfurt-vless"}, selector.Outbounds)
	assert.Equal(t, &outbound{
		Type:       "vless",
		Tag:        "frankfurt-vless",
		Server:     "203.0.113.10",
		ServerPort: 443,
		UUID:       "123e4567-e89b-12d3-a456-426614174000",
		Flow:       "xtls-rprx-vision",
		TLS: &tls{
			Enabled:    true,
			ServerName: "example.com",
			UTLS: &utls{
				Enabled:     true,
				Fingerprint: "firefox",
			},
			Reality: &reality{
				Enabled:   true,
				PublicKey: "public-key",
				ShortID:   "0123456789abcdef",
			},
		},
	}, conf.Outbounds[2])
}

func TestAddVLESSOutboundRejectsDuplicateTag(t *testing.T) {
	conf := &Conf{Outbounds: []*outbound{{Type: "direct", Tag: "frankfurt-vless"}}}

	err := addVLESSOutbound(testVLESSOutboundOptions(), conf)

	assert.Equal(t, apperr.Conflict, err.Kind())
	assert.Equal(t, "outbound tag 'frankfurt-vless' already exists", err.Msg())
	assert.Len(t, conf.Outbounds, 1)
}

func TestAddVLESSOutboundRequiresProxySelector(t *testing.T) {
	conf := &Conf{Outbounds: []*outbound{{Type: "direct", Tag: "direct"}}}

	err := addVLESSOutbound(testVLESSOutboundOptions(), conf)

	assert.Equal(t, apperr.NotFound, err.Kind())
	assert.Equal(t, "selector outbound tag 'proxy' was not found", err.Msg())
	assert.Len(t, conf.Outbounds, 1)
}

func TestRemoveOutbound(t *testing.T) {
	selector := &outbound{
		Type:      "selector",
		Tag:       "proxy",
		Outbounds: []string{"frankfurt-vless", "direct"},
		Default:   "frankfurt-vless",
	}
	conf := &Conf{Outbounds: []*outbound{
		selector,
		{Type: "direct", Tag: "direct"},
		{Type: "vless", Tag: "frankfurt-vless"},
	}}

	err := removeOutbound("frankfurt-vless", conf)

	require.Nil(t, err)
	assert.Equal(t, []*outbound{selector, {Type: "direct", Tag: "direct"}}, conf.Outbounds)
	assert.Equal(t, []string{"direct"}, selector.Outbounds)
	assert.Equal(t, "direct", selector.Default)
}

func TestRemoveOutboundReturnsNotFound(t *testing.T) {
	conf := &Conf{Outbounds: []*outbound{{Type: "direct", Tag: "direct"}}}

	err := removeOutbound("missing", conf)

	assert.Equal(t, apperr.NotFound, err.Kind())
	assert.Equal(t, "outbound tag 'missing' was not found", err.Msg())
}
