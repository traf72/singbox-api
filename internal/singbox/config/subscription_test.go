package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traf72/singbox-api/internal/apperr"
)

func TestApplySubscriptionUpdate(t *testing.T) {
	selector := &outbound{
		Type:      "selector",
		Tag:       "proxy",
		Outbounds: []string{"fixed-vless", "old-subscription"},
		Default:   "old-subscription",
	}
	fixed := &outbound{Type: "vless", Tag: "fixed-vless", Server: "192.0.2.10"}
	direct := &outbound{Type: "direct", Tag: "direct"}
	conf := &Conf{Outbounds: []*outbound{
		selector,
		fixed,
		direct,
		{Type: "vless", Tag: "old-subscription", Server: "192.0.2.20"},
	}}
	settings := &SubscriptionSettings{
		Version:          1,
		FixedOutbounds:   []string{"fixed-vless"},
		FallbackOutbound: "fixed-vless",
	}
	request := &SubscriptionUpdate{
		Version: 1,
		Outbounds: []json.RawMessage{
			json.RawMessage(`{"type":"direct","tag":"direct"}`),
			json.RawMessage(`{"type":"vless","tag":"📆 Осталось: 31 день"}`),
			validSubscriptionOutboundJSON("new-subscription", "vpn.example.com"),
			validSubscriptionOutboundJSON("fixed-vless", "192.0.2.30"),
			json.RawMessage(`{"type":"shadowsocks","tag":"unsupported","server":"192.0.2.40","server_port":443}`),
		},
	}

	result, err := applySubscriptionUpdate(request, settings, conf)

	require.Nil(t, err)
	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Removed)
	assert.Equal(t, 3, result.Preserved)
	assert.Equal(t, "fixed-vless", result.Default)
	assert.Equal(t, []SkippedSubscriptionOutbound{
		{Tag: "direct", Reason: "missing server"},
		{Tag: "📆 Осталось: 31 день", Reason: "missing server"},
		{Tag: "fixed-vless", Reason: "tag is fixed"},
		{Tag: "unsupported", Reason: "unsupported outbound type 'shadowsocks'"},
	}, result.Skipped)
	assert.Equal(t, []string{"fixed-vless", "new-subscription"}, selector.Outbounds)
	assert.Equal(t, "fixed-vless", selector.Default)
	require.Len(t, conf.Outbounds, 4)
	assert.Same(t, selector, conf.Outbounds[0])
	assert.Same(t, fixed, conf.Outbounds[1])
	assert.Same(t, direct, conf.Outbounds[2])
	assert.Equal(t, "new-subscription", conf.Outbounds[3].Tag)
	assert.Equal(t, "vpn.example.com", conf.Outbounds[3].Server)
}

func TestApplySubscriptionUpdateKeepsAvailableDefault(t *testing.T) {
	selector := &outbound{
		Type:      "selector",
		Tag:       "proxy",
		Outbounds: []string{"fixed-vless", "same-tag"},
		Default:   "same-tag",
	}
	conf := &Conf{Outbounds: []*outbound{
		selector,
		{Type: "vless", Tag: "fixed-vless", Server: "192.0.2.10"},
		{Type: "vless", Tag: "same-tag", Server: "192.0.2.20"},
	}}
	settings := &SubscriptionSettings{Version: 1, FixedOutbounds: []string{"fixed-vless"}, FallbackOutbound: "fixed-vless"}
	request := &SubscriptionUpdate{Version: 1, Outbounds: []json.RawMessage{validSubscriptionOutboundJSON("same-tag", "192.0.2.30")}}

	result, err := applySubscriptionUpdate(request, settings, conf)

	require.Nil(t, err)
	assert.Equal(t, "same-tag", result.Default)
	assert.Equal(t, "same-tag", selector.Default)
}

func TestApplySubscriptionUpdateRejectsEmptyUsableSetWithoutMutation(t *testing.T) {
	selector := &outbound{Type: "selector", Tag: "proxy", Outbounds: []string{"fixed-vless"}, Default: "fixed-vless"}
	old := &outbound{Type: "vless", Tag: "old-subscription", Server: "192.0.2.20"}
	conf := &Conf{Outbounds: []*outbound{selector, {Type: "vless", Tag: "fixed-vless"}, old}}
	settings := &SubscriptionSettings{Version: 1, FixedOutbounds: []string{"fixed-vless"}, FallbackOutbound: "fixed-vless"}
	request := &SubscriptionUpdate{Version: 1, Outbounds: []json.RawMessage{json.RawMessage(`{"type":"direct","tag":"direct"}`)}}

	result, err := applySubscriptionUpdate(request, settings, conf)

	assert.Nil(t, result)
	assert.Equal(t, apperr.Validation, err.Kind())
	assert.Equal(t, "subscription contains no usable outbounds", err.Msg())
	assert.Equal(t, []*outbound{selector, {Type: "vless", Tag: "fixed-vless"}, old}, conf.Outbounds)
}

func TestValidateSubscriptionSettings(t *testing.T) {
	selector := &outbound{Type: "selector", Tag: "proxy", Outbounds: []string{"fixed-vless"}}
	conf := &Conf{Outbounds: []*outbound{selector, {Type: "vless", Tag: "fixed-vless"}}}

	t.Run("Valid", func(t *testing.T) {
		settings := &SubscriptionSettings{Version: 1, FixedOutbounds: []string{" fixed-vless "}, FallbackOutbound: " fixed-vless "}

		err := validateSubscriptionSettings(settings, conf)

		require.Nil(t, err)
		assert.Equal(t, []string{"fixed-vless"}, settings.FixedOutbounds)
		assert.Equal(t, "fixed-vless", settings.FallbackOutbound)
	})

	t.Run("Fallback is not fixed", func(t *testing.T) {
		settings := &SubscriptionSettings{Version: 1, FixedOutbounds: []string{"fixed-vless"}, FallbackOutbound: "other"}

		err := validateSubscriptionSettings(settings, conf)

		assert.Equal(t, apperr.Validation, err.Kind())
		assert.Equal(t, "fallback_outbound must be included in fixed_outbounds", err.Msg())
	})
}

func validSubscriptionOutboundJSON(tag, server string) json.RawMessage {
	value := map[string]any{
		"type":            "vless",
		"tag":             tag,
		"server":          server,
		"server_port":     443,
		"uuid":            "123e4567-e89b-12d3-a456-426614174000",
		"flow":            "xtls-rprx-vision",
		"packet_encoding": "xudp",
		"tls": map[string]any{
			"enabled":     true,
			"server_name": "example.com",
			"utls": map[string]any{
				"enabled":     true,
				"fingerprint": "firefox",
			},
			"reality": map[string]any{
				"enabled":    true,
				"public_key": "public-key",
				"short_id":   "0123456789abcdef",
			},
		},
	}
	result, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return result
}
