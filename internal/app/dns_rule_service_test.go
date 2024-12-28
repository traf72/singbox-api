package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/core"
)

func TestParseDNSType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    core.DNSRuleType
		expectedErr *apperr.Err
	}{
		{"Keyword", "keyword", core.DNSRuleKeyword, nil},
		{"Keyword_TrimSpaces_LowerCase", "  KEYWORD\n", core.DNSRuleKeyword, nil},
		{"Domain", "domain", core.DNSRuleSuffix, nil},
		{"Domain_TrimSpaces", "\tdomain  \n", core.DNSRuleSuffix, nil},
		{"EmptyInput", "", -1, errEmptyType},
		{"SpaceOnlyInput", " \n\r\t", -1, errEmptyType},
		{"UnknownType", "unknown", -1, errUnknownType("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDNSType(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestParseRouteMode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    core.RouteMode
		expectedErr *apperr.Err
	}{
		{"Proxy", "proxy", core.RouteProxy, nil},
		{"Proxy_TrimSpaces_LowerCase", "  PROXY\n", core.RouteProxy, nil},
		{"Block", "domain", core.RouteBlock, nil},
		{"Direct", "domain", core.RouteDirect, nil},
		{"EmptyInput", "", "", errEmptyRouteMode},
		{"SpaceOnlyInput", " \n\r\t", "", errEmptyRouteMode},
		{"UnknownMode", "unknown", "", errUnknownRouteMode("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRouteMode(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestToDNSRule(t *testing.T) {
	tests := []struct {
		name          string
		rule          *DNSRuleDTO
		expected      *core.DNSRule
		expectedError *apperr.Err
	}{
		{
			name: "Domain_Proxy_TrimSpace_LowerCase",
			rule: &DNSRuleDTO{Domain: "GOOGLE.com", RouteMode: "\t PROXY\n"},
			expected: func() *core.DNSRule {
				t, _ := core.NewDNSRule(core.DNSRuleDomain, core.RouteProxy, "google.com")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "Suffix_Block",
			rule: &DNSRuleDTO{Domain: "DOMAIN:mail.google.com", RouteMode: "block"},
			expected: func() *core.DNSRule {
				t, _ := core.NewDNSRule(core.DNSRuleSuffix, core.RouteBlock, "mail.google.com")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "Keyword_Direct",
			rule: &DNSRuleDTO{Domain: "keyword:Google", RouteMode: "direct"},
			expected: func() *core.DNSRule {
				t, _ := core.NewDNSRule(core.DNSRuleKeyword, core.RouteDirect, "google")
				return t
			}(),
			expectedError: nil,
		},
		{
			name:          "Rule_Empty",
			rule:          &DNSRuleDTO{Domain: "", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyRule,
		},
		{
			name:          "Domain_SpaceOnly",
			rule:          &DNSRuleDTO{Domain: "\r\n\r ", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyRule,
		},
		{
			name:          "Domain_TooManyParts",
			rule:          &DNSRuleDTO{Domain: "domain:google.com:extra", RouteMode: "direct"},
			expected:      nil,
			expectedError: errTooManyParts("domain:google.com:extra"),
		},
		{
			name:          "RuleType_Empty",
			rule:          &DNSRuleDTO{Domain: ":google.com", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyType,
		},
		{
			name:          "RuleType_SpaceOnly",
			rule:          &DNSRuleDTO{Domain: "\t\r\n :google.com", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyType,
		},
		{
			name:          "RuleType_Unknown",
			rule:          &DNSRuleDTO{Domain: "bad:google.com", RouteMode: "direct"},
			expected:      nil,
			expectedError: errUnknownType("bad"),
		},
		{
			name:          "RouteMode_Empty",
			rule:          &DNSRuleDTO{Domain: "domain:google.com", RouteMode: ""},
			expected:      nil,
			expectedError: errEmptyRouteMode,
		},
		{
			name:          "RouteMode_SpaceOnly",
			rule:          &DNSRuleDTO{Domain: "domain:google.com", RouteMode: "\r\n\t "},
			expected:      nil,
			expectedError: errEmptyRouteMode,
		},
		{
			name:          "RouteMode_Unknown",
			rule:          &DNSRuleDTO{Domain: "domain:google.com", RouteMode: "bad"},
			expected:      nil,
			expectedError: errUnknownRouteMode("bad"),
		},
		{
			name:          "Domain_Empty",
			rule:          &DNSRuleDTO{Domain: "domain:", RouteMode: "direct"},
			expected:      nil,
			expectedError: apperr.NewValidationErr("DNSRule_EmptyDomain", "domain is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := tt.rule.toDNSRule()
			assert.Equal(t, tt.expected, r)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
