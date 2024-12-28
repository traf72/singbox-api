package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
)

func TestRuleType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		kind     RuleType
		expected bool
	}{
		{"Suffix", Suffix, true},
		{"Keyword", Keyword, true},
		{"Domain", Domain, true},
		{"Regex", Regex, true},
		{"Unknown", RuleType(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.kind.isValid())
		})
	}
}

func TestRule_Validate(t *testing.T) {
	tests := []struct {
		name     string
		rule     Rule
		expected *apperr.Err
	}{
		{"Rule_Suffix_Proxy", Rule{kind: Suffix, mode: config.RouteProxy, domain: ".com"}, nil},
		{"Rule_Keyword_Block", Rule{kind: Keyword, mode: config.RouteBlock, domain: "google"}, nil},
		{"Rule_Domain_Direct", Rule{kind: Domain, mode: config.RouteDirect, domain: "google.com"}, nil},
		{"Rule_Regexp_Proxy", Rule{kind: Regex, mode: config.RouteProxy, domain: "you.*be"}, nil},
		{"Rule_Empty", Rule{kind: Suffix, mode: config.RouteProxy, domain: ""}, errEmptyDomain},
		{"Rule_WithSpace", Rule{kind: Keyword, mode: config.RouteProxy, domain: "google com"}, errDomainHasSpaces("google com")},
		{"Rule_WithLineBreak", Rule{kind: Keyword, mode: config.RouteProxy, domain: "google\ncom"}, errDomainHasSpaces("google\ncom")},
		{"Rule_WithTab", Rule{kind: Keyword, mode: config.RouteProxy, domain: "google\tcom"}, errDomainHasSpaces("google\tcom")},
		{"Kind_Invalid", Rule{kind: RuleType(-1), mode: config.RouteProxy, domain: "google.com"}, errInvalidRuleType},
		{"RouteMode_Invalid", Rule{kind: Suffix, mode: "Unknown", domain: "google.com"}, apperr.NewValidationErr("DNSRule_InvalidRouteMode", "invalid route mode 'Unknown'")},
		{"Domain_Invalid", Rule{kind: Domain, mode: config.RouteProxy, domain: ".com"}, errInvalidDomain(".com")},
		{"Regex_Invalid", Rule{kind: Regex, mode: config.RouteProxy, domain: "[a-z"}, errInvalidRegexp("[a-z")},
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
		kind          RuleType
		mode          config.RouteMode
		domain        string
		expected      *Rule
		expectedError *apperr.Err
	}{
		{"Suffix_Direct", Suffix, config.RouteDirect, " .Com ", &Rule{Suffix, config.RouteDirect, ".com"}, nil},
		{"Suffix_Block", Suffix, config.RouteBlock, " .Com ", &Rule{Suffix, config.RouteBlock, ".com"}, nil},
		{"Suffix_Proxy", Suffix, config.RouteProxy, " .Com ", &Rule{Suffix, config.RouteProxy, ".com"}, nil},
		{"Domain_Proxy", Domain, config.RouteProxy, "Google.Com\n", &Rule{Domain, config.RouteProxy, "google.com"}, nil},
		{"Keyword_Block", Keyword, config.RouteBlock, "google", &Rule{Keyword, config.RouteBlock, "google"}, nil},
		{"EmptyDomain", Suffix, config.RouteProxy, "", nil, errEmptyDomain},
		{"WhiteSpaceOnlyDomain", Keyword, config.RouteProxy, " \n\r\t", nil, errEmptyDomain},
		{"DomainWithSpace", Suffix, config.RouteProxy, "google com", nil, errDomainHasSpaces("google com")},
		{"Kind_Invalid", RuleType(-1), config.RouteProxy, "google.com", nil, errInvalidRuleType},
		{"RouteMode_Invalid", Suffix, "Unknown", "google.com", nil, apperr.NewValidationErr("DNSRule_InvalidRouteMode", "invalid route mode 'Unknown'")},
		{"Domain_Invalid", Domain, config.RouteProxy, "@com", nil, errInvalidDomain("@com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewRule(tt.kind, tt.mode, tt.domain)
			assert.Equal(t, tt.expected, rule)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
