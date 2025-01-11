package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
	"github.com/traf72/singbox-api/internal/singbox/config/ip"
)

func Test_IPRuleToConfigRule(t *testing.T) {
	tests := []struct {
		name          string
		rule          *IPRule
		expected      *ip.Rule
		expectedError apperr.Err
	}{
		{
			name: "IP_Proxy_TrimSpace_LowerCase",
			rule: &IPRule{IP: "\t 142.250.0.0\r\n", RouteMode: "\t PROXY\n"},
			expected: func() *ip.Rule {
				t, _ := ip.NewRule(config.RouteProxy, "142.250.0.0")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "IP_Direct",
			rule: &IPRule{IP: "142.250.0.0", RouteMode: "direct"},
			expected: func() *ip.Rule {
				t, _ := ip.NewRule(config.RouteDirect, "142.250.0.0")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "IP_Block",
			rule: &IPRule{IP: "142.250.0.0", RouteMode: "block"},
			expected: func() *ip.Rule {
				t, _ := ip.NewRule(config.RouteBlock, "142.250.0.0")
				return t
			}(),
			expectedError: nil,
		},
		{
			name:          "IP_Empty",
			rule:          &IPRule{IP: "", RouteMode: "direct"},
			expected:      nil,
			expectedError: errIPEmptyRule,
		},
		{
			name:          "IP_SpaceOnly",
			rule:          &IPRule{IP: "\r\n\r ", RouteMode: "direct"},
			expected:      nil,
			expectedError: errIPEmptyRule,
		},
		{
			name:          "RouteMode_Empty",
			rule:          &IPRule{IP: "domain:google.com", RouteMode: ""},
			expected:      nil,
			expectedError: apperr.NewValidationErr("IPRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:          "RouteMode_SpaceOnly",
			rule:          &IPRule{IP: "domain:google.com", RouteMode: "\r\n\t "},
			expected:      nil,
			expectedError: apperr.NewValidationErr("IPRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:          "RouteMode_Unknown",
			rule:          &IPRule{IP: "domain:google.com", RouteMode: "bad"},
			expected:      nil,
			expectedError: apperr.NewValidationErr("IPRule_InvalidRouteMode", "route mode 'bad' is unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := tt.rule.toConfigRule()
			assert.Equal(t, tt.expected, r)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
