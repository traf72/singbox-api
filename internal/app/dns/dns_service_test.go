package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/config/dns"
)

func TestParseType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    dns.RuleType
		expectedErr apperr.Err
	}{
		{"Full", "full", dns.Domain, nil},
		{"Full_TrimSpaces_LowerCase", "\tFull\r\n", dns.Domain, nil},
		{"Keyword", "keyword", dns.Keyword, nil},
		{"Keyword_TrimSpaces_LowerCase", "  KEYWORD\n", dns.Keyword, nil},
		{"Domain", "domain", dns.Suffix, nil},
		{"Domain_TrimSpaces", "\tdomain  \n", dns.Suffix, nil},
		{"Regex", "regexp", dns.Regex, nil},
		{"Regex_TrimSpaces", "\tregexp\r\n", dns.Regex, nil},
		{"EmptyInput", "", -1, errEmptyType},
		{"SpaceOnlyInput", " \n\r\t", -1, errEmptyType},
		{"UnknownType", "unknown", -1, errUnknownType("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseType(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestToDNSRule(t *testing.T) {
	tests := []struct {
		name          string
		rule          *Rule
		expected      *dns.Rule
		expectedError apperr.Err
	}{
		{
			name: "Domain_Proxy_TrimSpace_LowerCase",
			rule: &Rule{Domain: "GOOGLE.com", RouteMode: "\t PROXY\n"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Domain, config.RouteProxy, "google.com")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "Domain_Full_Direct",
			rule: &Rule{Domain: "full:google.com", RouteMode: "direct"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Domain, config.RouteDirect, "google.com")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "Suffix_Block",
			rule: &Rule{Domain: "DOMAIN:mail.google.com", RouteMode: "block"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Suffix, config.RouteBlock, "mail.google.com")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "Keyword_Direct",
			rule: &Rule{Domain: "keyword:Google", RouteMode: "direct"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Keyword, config.RouteDirect, "google")
				return t
			}(),
			expectedError: nil,
		},
		{
			name: "Regex_Proxy",
			rule: &Rule{Domain: "regexp:you.*be", RouteMode: "proxy"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Regex, config.RouteProxy, "you.*be")
				return t
			}(),
			expectedError: nil,
		},
		{
			name:          "Rule_Empty",
			rule:          &Rule{Domain: "", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyRule,
		},
		{
			name:          "Domain_SpaceOnly",
			rule:          &Rule{Domain: "\r\n\r ", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyRule,
		},
		{
			name:          "Domain_TooManyParts",
			rule:          &Rule{Domain: "domain:google.com:extra", RouteMode: "direct"},
			expected:      nil,
			expectedError: errTooManyParts("domain:google.com:extra"),
		},
		{
			name:          "RuleType_Empty",
			rule:          &Rule{Domain: ":google.com", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyType,
		},
		{
			name:          "RuleType_SpaceOnly",
			rule:          &Rule{Domain: "\t\r\n :google.com", RouteMode: "direct"},
			expected:      nil,
			expectedError: errEmptyType,
		},
		{
			name:          "RuleType_Unknown",
			rule:          &Rule{Domain: "bad:google.com", RouteMode: "direct"},
			expected:      nil,
			expectedError: errUnknownType("bad"),
		},
		{
			name:          "RouteMode_Empty",
			rule:          &Rule{Domain: "domain:google.com", RouteMode: ""},
			expected:      nil,
			expectedError: apperr.NewValidationErr("DNSRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:          "RouteMode_SpaceOnly",
			rule:          &Rule{Domain: "domain:google.com", RouteMode: "\r\n\t "},
			expected:      nil,
			expectedError: apperr.NewValidationErr("DNSRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:          "RouteMode_Unknown",
			rule:          &Rule{Domain: "domain:google.com", RouteMode: "bad"},
			expected:      nil,
			expectedError: apperr.NewValidationErr("DNSRule_InvalidRouteMode", "route mode 'bad' is unknown"),
		},
		{
			name:          "Domain_Empty",
			rule:          &Rule{Domain: "domain:", RouteMode: "direct"},
			expected:      nil,
			expectedError: apperr.NewValidationErr("DNSRule_EmptyDomain", "domain is empty"),
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
