package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/core"
)

func TestParseDnsType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    core.DnsRuleType
		expectedErr *apperr.Err
	}{
		{"Keyword", "keyword", core.Keyword, nil},
		{"Keyword_TrimSpaces", "  keyword\n", core.Keyword, nil},
		{"Domain", "domain", core.Suffix, nil},
		{"Domain_TrimSpaces", "\tdomain  \n", core.Suffix, nil},
		{"EmptyInput", "", -1, errEmptyDnsRuleType},
		{"SpaceOnlyInput", " \n\r\t", -1, errEmptyDnsRuleType},
		{"UnknownType", "unknown", -1, errUnknownDnsRuleType("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDnsType(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      *core.DnsRule
		expectedError *apperr.Err
	}{
		{
			name:          "Domain",
			input:         "GOOGLE.com",
			expected:      func() *core.DnsRule { t, _ := core.NewDnsRule(core.Domain, "google.com"); return t }(),
			expectedError: nil,
		},
		{
			name:          "Suffix",
			input:         "DOMAIN:mail.google.com",
			expected:      func() *core.DnsRule { t, _ := core.NewDnsRule(core.Suffix, "mail.google.com"); return t }(),
			expectedError: nil,
		},
		{
			name:          "Keyword",
			input:         "\t KEYworD:Google \n",
			expected:      func() *core.DnsRule { t, _ := core.NewDnsRule(core.Keyword, "google"); return t }(),
			expectedError: nil,
		},
		{
			name:          "EmptyInput",
			input:         "",
			expected:      nil,
			expectedError: errEmptyRule,
		},
		{
			name:          "SpaceOnlyInput",
			input:         "\r\n\r ",
			expected:      nil,
			expectedError: errEmptyRule,
		},
		{
			name:          "TooManyParts",
			input:         "domain:google.com:extra",
			expected:      nil,
			expectedError: errTooManyParts("domain:google.com:extra"),
		},
		{
			name:          "EmptyDnsRuleType",
			input:         ":google.com",
			expected:      nil,
			expectedError: errEmptyDnsRuleType,
		},
		{
			name:          "UnknownDnsRuleType",
			input:         "bad:google.com",
			expected:      nil,
			expectedError: errUnknownDnsRuleType("bad"),
		},
		{
			name:          "EmptyDomain",
			input:         "domain:",
			expected:      nil,
			expectedError: apperr.NewValidationErr("EmptyDomain", "domain is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parse(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
