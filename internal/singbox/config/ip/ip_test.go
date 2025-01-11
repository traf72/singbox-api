package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

func TestRule_Validate(t *testing.T) {
	tests := []struct {
		name     string
		rule     Rule
		expected apperr.Err
	}{
		{
			name:     "IP_ValidIPv4_Proxy",
			rule:     Rule{mode: config.RouteProxy, ip: "192.168.0.1"},
			expected: nil,
		},
		{
			name:     "IP_ValidCIDR_Direct",
			rule:     Rule{mode: config.RouteDirect, ip: "10.0.0.0/24"},
			expected: nil,
		},
		{
			name:     "IP_Empty",
			rule:     Rule{mode: config.RouteProxy, ip: ""},
			expected: errEmptyIP,
		},
		{
			name:     "IP_Invalid",
			rule:     Rule{mode: config.RouteProxy, ip: "999.999.999.999"},
			expected: errInvalidIP("999.999.999.999"),
		},
		{
			name:     "IP_InvalidRouteMode",
			rule:     Rule{mode: "Unknown", ip: "192.168.0.1"},
			expected: apperr.NewValidationErr("IPRule_InvalidRouteMode", "invalid route mode 'Unknown'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.rule.validate())
		})
	}
}

func TestNewRule(t *testing.T) {
	tests := []struct {
		name          string
		mode          config.RouteMode
		ip            string
		expected      *Rule
		expectedError apperr.Err
	}{
		{
			name:     "IP_ValidIPv4_Proxy",
			mode:     config.RouteProxy,
			ip:       "192.168.0.1",
			expected: &Rule{mode: config.RouteProxy, ip: "192.168.0.1"},
		},
		{
			name:     "IP_ValidCIDR_Block",
			mode:     config.RouteBlock,
			ip:       "10.0.0.0/24",
			expected: &Rule{mode: config.RouteBlock, ip: "10.0.0.0/24"},
		},
		{
			name:          "IP_Empty",
			mode:          config.RouteProxy,
			ip:            "",
			expected:      nil,
			expectedError: errEmptyIP,
		},
		{
			name:          "IP_Invalid",
			mode:          config.RouteDirect,
			ip:            "999.999.999.999",
			expected:      nil,
			expectedError: errInvalidIP("999.999.999.999"),
		},
		{
			name:          "IP_InvalidRouteMode",
			mode:          "Unknown",
			ip:            "192.168.0.1",
			expected:      nil,
			expectedError: apperr.NewValidationErr("IPRule_InvalidRouteMode", "invalid route mode 'Unknown'"),
		},
		{
			name:     "TrimSpaces",
			mode:     config.RouteProxy,
			ip:       "   192.168.0.100/32  ",
			expected: &Rule{mode: config.RouteProxy, ip: "192.168.0.100/32"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewRule(tt.mode, tt.ip)
			assert.Equal(t, tt.expected, rule)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
