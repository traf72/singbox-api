package application

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/domain"
	"github.com/traf72/singbox-api/internal/err"
)

func TestParseType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected domain.TemplateKind
	}{
		{"Keyword", "keyword", domain.Keyword},
		{"Keyword_TrimSpaces", "  keyword\n", domain.Keyword},
		{"Domain", "domain", domain.Suffix},
		{"Domain_TrimSpaces", "\tdomain  \n", domain.Suffix},
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
		expected      *domain.Template
		expectedError *err.AppErr
	}{
		{
			name:          "Domain",
			input:         "GOOGLE.com",
			expected:      func() *domain.Template { t, _ := domain.NewTemplate(domain.Domain, "google.com"); return t }(),
			expectedError: nil,
		},
		{
			name:          "Suffix",
			input:         "DOMAIN:mail.google.com",
			expected:      func() *domain.Template { t, _ := domain.NewTemplate(domain.Suffix, "mail.google.com"); return t }(),
			expectedError: nil,
		},
		{
			name:          "Keyword",
			input:         "\t KEYworD:Google \n",
			expected:      func() *domain.Template { t, _ := domain.NewTemplate(domain.Keyword, "google"); return t }(),
			expectedError: nil,
		},
		{
			name:          "EmptyInput",
			input:         "",
			expected:      nil,
			expectedError: errEmptyTemplate,
		},
		{
			name:          "SpaceOnlyInput",
			input:         "\r\n\r ",
			expected:      nil,
			expectedError: errEmptyTemplate,
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
			expectedError: err.NewValidationErr("InvalidTemplateKind", fmt.Sprintf("kind '%d' is invalid", -1)),
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
