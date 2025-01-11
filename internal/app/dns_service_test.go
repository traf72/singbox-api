package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
	"github.com/traf72/singbox-api/internal/singbox/config/dns"
)

func TestParseDNSRuleType(t *testing.T) {
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
		{"EmptyInput", "", -1, errDNSEmptyType},
		{"SpaceOnlyInput", " \n\r\t", -1, errDNSEmptyType},
		{"UnknownType", "unknown", -1, errDNSUnknownType("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDNSRuleType(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_DNSRuleToConfigRule(t *testing.T) {
	tests := []struct {
		name        string
		rule        *DNSRule
		expected    *dns.Rule
		expectedErr apperr.Err
	}{
		{
			name: "Domain_Proxy_TrimSpace_LowerCase",
			rule: &DNSRule{Domain: "GOOGLE.com", RouteMode: "\t PROXY\n"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Domain, config.RouteProxy, "google.com")
				return t
			}(),
			expectedErr: nil,
		},
		{
			name: "Domain_Full_Direct",
			rule: &DNSRule{Domain: "full:google.com", RouteMode: "direct"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Domain, config.RouteDirect, "google.com")
				return t
			}(),
			expectedErr: nil,
		},
		{
			name: "Suffix_Block",
			rule: &DNSRule{Domain: "DOMAIN:mail.google.com", RouteMode: "block"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Suffix, config.RouteBlock, "mail.google.com")
				return t
			}(),
			expectedErr: nil,
		},
		{
			name: "Keyword_Direct",
			rule: &DNSRule{Domain: "keyword:Google", RouteMode: "direct"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Keyword, config.RouteDirect, "google")
				return t
			}(),
			expectedErr: nil,
		},
		{
			name: "Regex_Proxy",
			rule: &DNSRule{Domain: "regexp:you.*be", RouteMode: "proxy"},
			expected: func() *dns.Rule {
				t, _ := dns.NewRule(dns.Regex, config.RouteProxy, "you.*be")
				return t
			}(),
			expectedErr: nil,
		},
		{
			name:        "Rule_Empty",
			rule:        &DNSRule{Domain: "", RouteMode: "direct"},
			expected:    nil,
			expectedErr: errDNSEmptyRule,
		},
		{
			name:        "Domain_SpaceOnly",
			rule:        &DNSRule{Domain: "\r\n\r ", RouteMode: "direct"},
			expected:    nil,
			expectedErr: errDNSEmptyRule,
		},
		{
			name:        "Domain_TooManyParts",
			rule:        &DNSRule{Domain: "domain:google.com:extra", RouteMode: "direct"},
			expected:    nil,
			expectedErr: errDNSTooManyParts("domain:google.com:extra"),
		},
		{
			name:        "RuleType_Empty",
			rule:        &DNSRule{Domain: ":google.com", RouteMode: "direct"},
			expected:    nil,
			expectedErr: errDNSEmptyType,
		},
		{
			name:        "RuleType_SpaceOnly",
			rule:        &DNSRule{Domain: "\t\r\n :google.com", RouteMode: "direct"},
			expected:    nil,
			expectedErr: errDNSEmptyType,
		},
		{
			name:        "RuleType_Unknown",
			rule:        &DNSRule{Domain: "bad:google.com", RouteMode: "direct"},
			expected:    nil,
			expectedErr: errDNSUnknownType("bad"),
		},
		{
			name:        "RouteMode_Empty",
			rule:        &DNSRule{Domain: "domain:google.com", RouteMode: ""},
			expected:    nil,
			expectedErr: apperr.NewValidationErr("DNSRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:        "RouteMode_SpaceOnly",
			rule:        &DNSRule{Domain: "domain:google.com", RouteMode: "\r\n\t "},
			expected:    nil,
			expectedErr: apperr.NewValidationErr("DNSRule_InvalidRouteMode", "route mode is empty"),
		},
		{
			name:        "RouteMode_Unknown",
			rule:        &DNSRule{Domain: "domain:google.com", RouteMode: "bad"},
			expected:    nil,
			expectedErr: apperr.NewValidationErr("DNSRule_InvalidRouteMode", "route mode 'bad' is unknown"),
		},
		{
			name:        "Domain_Empty",
			rule:        &DNSRule{Domain: "domain:", RouteMode: "direct"},
			expected:    nil,
			expectedErr: apperr.NewValidationErr("DNSRule_EmptyDomain", "domain is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := tt.rule.toConfigRule()
			assert.Equal(t, tt.expected, r)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
