package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
)

func TestDNSRuleType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		kind     DNSRuleType
		expected bool
	}{
		{"Suffix", DNSRuleSuffix, true},
		{"Keyword", DNSRuleKeyword, true},
		{"Domain", DNSRuleDomain, true},
		{"Regex", DNSRuleRegex, true},
		{"Unknown", DNSRuleType(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.kind.isValid())
		})
	}
}

func TestDNSRule_Validate(t *testing.T) {
	tests := []struct {
		name     string
		rule     DNSRule
		expected *apperr.Err
	}{
		{"Rule_Suffix_Proxy", DNSRule{kind: DNSRuleSuffix, mode: RouteProxy, domain: ".com"}, nil},
		{"Rule_Keyword_Block", DNSRule{kind: DNSRuleKeyword, mode: RouteBlock, domain: "google"}, nil},
		{"Rule_Domain_Direct", DNSRule{kind: DNSRuleDomain, mode: RouteDirect, domain: "google.com"}, nil},
		{"Rule_Regexp_Proxy", DNSRule{kind: DNSRuleRegex, mode: RouteProxy, domain: "you.*be"}, nil},
		{"Rule_Empty", DNSRule{kind: DNSRuleSuffix, mode: RouteProxy, domain: ""}, errEmptyDomain},
		{"Rule_WithSpace", DNSRule{kind: DNSRuleKeyword, mode: RouteProxy, domain: "google com"}, errDomainHasSpaces("google com")},
		{"Rule_WithLineBreak", DNSRule{kind: DNSRuleKeyword, mode: RouteProxy, domain: "google\ncom"}, errDomainHasSpaces("google\ncom")},
		{"Rule_WithTab", DNSRule{kind: DNSRuleKeyword, mode: RouteProxy, domain: "google\tcom"}, errDomainHasSpaces("google\tcom")},
		{"Kind_Invalid", DNSRule{kind: DNSRuleType(-1), mode: RouteProxy, domain: "google.com"}, errInvalidRuleType},
		{"RouteMode_Invalid", DNSRule{kind: DNSRuleSuffix, mode: "Unknown", domain: "google.com"}, errUnknownRouteMode("Unknown")},
		{"Domain_Invalid", DNSRule{kind: DNSRuleDomain, mode: RouteProxy, domain: ".com"}, errInvalidDomain(".com")},
		{"Regex_Invalid", DNSRule{kind: DNSRuleRegex, mode: RouteProxy, domain: "[a-z"}, errInvalidRegexp("[a-z")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.rule.validate())
		})
	}
}

func TestNewDNSRule(t *testing.T) {
	tests := []struct {
		name          string
		kind          DNSRuleType
		mode          RouteMode
		domain        string
		expected      *DNSRule
		expectedError *apperr.Err
	}{
		{"Suffix_Direct", DNSRuleSuffix, RouteDirect, " .Com ", &DNSRule{DNSRuleSuffix, RouteDirect, ".com"}, nil},
		{"Suffix_Block", DNSRuleSuffix, RouteBlock, " .Com ", &DNSRule{DNSRuleSuffix, RouteBlock, ".com"}, nil},
		{"Suffix_Proxy", DNSRuleSuffix, RouteProxy, " .Com ", &DNSRule{DNSRuleSuffix, RouteProxy, ".com"}, nil},
		{"Domain_Proxy", DNSRuleDomain, RouteProxy, "Google.Com\n", &DNSRule{DNSRuleDomain, RouteProxy, "google.com"}, nil},
		{"Keyword_Block", DNSRuleKeyword, RouteBlock, "google", &DNSRule{DNSRuleKeyword, RouteBlock, "google"}, nil},
		{"EmptyDomain", DNSRuleSuffix, RouteProxy, "", nil, errEmptyDomain},
		{"WhiteSpaceOnlyDomain", DNSRuleKeyword, RouteProxy, " \n\r\t", nil, errEmptyDomain},
		{"DomainWithSpace", DNSRuleSuffix, RouteProxy, "google com", nil, errDomainHasSpaces("google com")},
		{"Kind_Invalid", DNSRuleType(-1), RouteProxy, "google.com", nil, errInvalidRuleType},
		{"RouteMode_Invalid", DNSRuleSuffix, "Unknown", "google.com", nil, errUnknownRouteMode("Unknown")},
		{"Domain_Invalid", DNSRuleDomain, RouteProxy, "@com", nil, errInvalidDomain("@com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewDNSRule(tt.kind, tt.mode, tt.domain)
			assert.Equal(t, tt.expected, rule)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
