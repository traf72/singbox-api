package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/err"
)

func TestTemplateKind_String(t *testing.T) {
	tests := []struct {
		name     string
		kind     DnsRuleType
		expected string
	}{
		{"Suffix", Suffix, "Suffix"},
		{"Keyword", Keyword, "Keyword"},
		{"Domain", Domain, "Domain"},
		{"Unknown", DnsRuleType(-1), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.kind.String())
		})
	}
}

func TestTemplateKind_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		kind     DnsRuleType
		expected bool
	}{
		{"Suffix", Suffix, true},
		{"Keyword", Keyword, true},
		{"Domain", Domain, true},
		{"Unknown", DnsRuleType(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.kind.isValid())
		})
	}
}

func TestTemplate_Validate(t *testing.T) {
	tests := []struct {
		name     string
		template DnsRule
		expected *err.AppErr
	}{
		{"Template_Suffix", DnsRule{kind: Suffix, domain: ".com"}, nil},
		{"Template_Keyword", DnsRule{kind: Keyword, domain: "google"}, nil},
		{"Template_Domain", DnsRule{kind: Domain, domain: "google.com"}, nil},
		{"Template_Empty", DnsRule{kind: Suffix, domain: ""}, errEmptyDomain},
		{"Template_WithSpace", DnsRule{kind: Keyword, domain: "google com"}, errDomainHasSpaces("google com")},
		{"Template_WithLineBreak", DnsRule{kind: Keyword, domain: "google\ncom"}, errDomainHasSpaces("google\ncom")},
		{"Template_WithTab", DnsRule{kind: Keyword, domain: "google\tcom"}, errDomainHasSpaces("google\tcom")},
		{"Kind_Invalid", DnsRule{kind: DnsRuleType(-1), domain: "google.com"}, errInvalidRuleType},
		{"Domain_Invalid", DnsRule{kind: Domain, domain: ".com"}, errInvalidDomain(".com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.template.validate())
		})
	}
}

func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name          string
		kind          DnsRuleType
		text          string
		expected      *DnsRule
		expectedError *err.AppErr
	}{
		{"Template_Suffix", Suffix, " .Com ", &DnsRule{Suffix, ".com"}, nil},
		{"Template_Domain", Domain, "Google.Com\n", &DnsRule{Domain, "google.com"}, nil},
		{"Template_Keyword", Keyword, "google", &DnsRule{Keyword, "google"}, nil},
		{"Template_Empty", Suffix, "", nil, errEmptyDomain},
		{"Template_WhiteSpaceOnly", Keyword, " \n\r\t", nil, errEmptyDomain},
		{"Template_WithSpace", Suffix, "google com", nil, errDomainHasSpaces("google com")},
		{"Kind_Invalid", DnsRuleType(-1), "google.com", nil, errInvalidRuleType},
		{"Domain_Invalid", Domain, "@com", nil, errInvalidDomain("@com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := NewDnsRule(tt.kind, tt.text)
			assert.Equal(t, tt.expected, template)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
