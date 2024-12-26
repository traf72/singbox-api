package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/err"
)

func TestParseType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected config.DnsRuleType
	}{
		{"Keyword", "keyword", config.Keyword},
		{"Keyword_TrimSpaces", "  keyword\n", config.Keyword},
		{"Domain", "domain", config.Suffix},
		{"Domain_TrimSpaces", "\tdomain  \n", config.Suffix},
		{"EmptyInput", "", -1},
		{"SpaceOnlyInput", " \n\r\t", -1},
		{"UnknownType", "unknown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseKind(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      *config.DnsRule
		expectedError *err.AppErr
	}{
		{
			name:          "Domain",
			input:         "GOOGLE.com",
			expected:      func() *config.DnsRule { t, _ := config.NewDnsRule(config.Domain, "google.com"); return t }(),
			expectedError: nil,
		},
		{
			name:          "Suffix",
			input:         "DOMAIN:mail.google.com",
			expected:      func() *config.DnsRule { t, _ := config.NewDnsRule(config.Suffix, "mail.google.com"); return t }(),
			expectedError: nil,
		},
		{
			name:          "Keyword",
			input:         "\t KEYworD:Google \n",
			expected:      func() *config.DnsRule { t, _ := config.NewDnsRule(config.Keyword, "google"); return t }(),
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
			name:          "EmptyKind",
			input:         ":google.com",
			expected:      nil,
			expectedError: err.NewValidationErr("InvalidRuleType", "rule type is invalid"),
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
