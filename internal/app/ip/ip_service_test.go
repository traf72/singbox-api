package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/config/ip"
)

func TestToConfigRule(t *testing.T) {
	tests := []struct {
		name          string
		rule          *Rule
		expected      *ip.Rule
		expectedError *apperr.Err
	}{
		{
			name: "IP_Proxy_TrimSpace_LowerCase",
			rule: &Rule{IP: "\t 142.250.0.0\r\n", RouteMode: "\t PROXY\n"},
			expected: func() *ip.Rule {
				t, _ := ip.NewRule(config.RouteProxy, "142.250.0.0")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "IP_Direct",
			rule: &Rule{IP: "142.250.0.0", RouteMode: "direct"},
			expected: func() *ip.Rule {
				t, _ := ip.NewRule(config.RouteDirect, "142.250.0.0")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "IP_Block",
			rule: &Rule{IP: "142.250.0.0", RouteMode: "block"},
			expected: func() *ip.Rule {
				t, _ := ip.NewRule(config.RouteBlock, "142.250.0.0")
				return t
			}(),
			expectedError: nil,
		},
		{
			name:          "IP_Empty",
			rule:          &Rule{IP: "", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyIP,
		},
		{
			name:          "IP_SpaceOnly",
			rule:          &Rule{IP: "\r\n\r ", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyIP,
		},
		{
			name:          "RouteMode_Empty",
			rule:          &Rule{IP: "domain:google.com", RouteMode: ""},
			expected:      nil,
			expectedError: apperr.NewValidationErr("IPRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:          "RouteMode_SpaceOnly",
			rule:          &Rule{IP: "domain:google.com", RouteMode: "\r\n\t "},
			expected:      nil,
			expectedError: apperr.NewValidationErr("IPRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:          "RouteMode_Unknown",
			rule:          &Rule{IP: "domain:google.com", RouteMode: "bad"},
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
