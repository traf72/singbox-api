package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traf72/singbox-api/internal/apperr"
)

func TestSetFinalRouteMode(t *testing.T) {
	tests := []struct {
		name  string
		mode  RouteMode
		route string
		dns   string
	}{
		{name: "Proxy all", mode: RouteProxy, route: "proxy", dns: "dns-remote"},
		{name: "Split tunnel", mode: RouteDirect, route: "direct", dns: "dns-direct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := new(Conf)

			err := setFinalRouteMode(tt.mode, conf)

			require.Nil(t, err)
			assert.Equal(t, tt.route, conf.Route.Final)
			assert.Equal(t, tt.dns, conf.DNS.Final)
		})
	}
}

func TestSetFinalRouteModeRejectsUnsupportedMode(t *testing.T) {
	conf := new(Conf)

	err := setFinalRouteMode(RouteBlock, conf)

	assert.Equal(t, apperr.Validation, err.Kind())
	assert.Equal(t, "route mode 'block' is unsupported; expected 'proxy' or 'direct'", err.Msg())
	assert.Empty(t, conf.Route.Final)
	assert.Empty(t, conf.DNS.Final)
}
