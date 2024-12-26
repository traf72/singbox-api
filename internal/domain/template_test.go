package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/err"
)

func TestTemplateKind_String(t *testing.T) {
	tests := []struct {
		name     string
		kind     TemplateKind
		expected string
	}{
		{"Suffix", Suffix, "Suffix"},
		{"Keyword", Keyword, "Keyword"},
		{"Domain", Domain, "Domain"},
		{"Unknown", TemplateKind(-1), "Unknown"},
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
		kind     TemplateKind
		expected bool
	}{
		{"Suffix", Suffix, true},
		{"Keyword", Keyword, true},
		{"Domain", Domain, true},
		{"Unknown", TemplateKind(-1), false},
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
		template Template
		expected *err.AppErr
	}{
		{"Template_Suffix", Template{kind: Suffix, text: ".com"}, nil},
		{"Template_Keyword", Template{kind: Keyword, text: "google"}, nil},
		{"Template_Domain", Template{kind: Domain, text: "google.com"}, nil},
		{"Template_Empty", Template{kind: Suffix, text: ""}, errEmptyTemplate},
		{"Template_WithSpace", Template{kind: Keyword, text: "google com"}, errTemplateHasSpaces("google com")},
		{"Template_WithLineBreak", Template{kind: Keyword, text: "google\ncom"}, errTemplateHasSpaces("google\ncom")},
		{"Template_WithTab", Template{kind: Keyword, text: "google\tcom"}, errTemplateHasSpaces("google\tcom")},
		{"Kind_Invalid", Template{kind: TemplateKind(-1), text: "google.com"}, errInvalidKind(TemplateKind(-1))},
		{"Domain_Invalid", Template{kind: Domain, text: ".com"}, errInvalidDomain(".com")},
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
		kind          TemplateKind
		text          string
		expected      *Template
		expectedError *err.AppErr
	}{
		{"Template_Suffix", Suffix, " .Com ", &Template{Suffix, ".com"}, nil},
		{"Template_Domain", Domain, "Google.Com\n", &Template{Domain, "google.com"}, nil},
		{"Template_Keyword", Keyword, "google", &Template{Keyword, "google"}, nil},
		{"Template_Empty", Suffix, "", nil, errEmptyTemplate},
		{"Template_WhiteSpaceOnly", Keyword, " \n\r\t", nil, errEmptyTemplate},
		{"Template_WithSpace", Suffix, "google com", nil, errTemplateHasSpaces("google com")},
		{"Kind_Invalid", TemplateKind(-1), "google.com", nil, errInvalidKind(TemplateKind(-1))},
		{"Domain_Invalid", Domain, "@com", nil, errInvalidDomain("@com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := NewTemplate(tt.kind, tt.text)
			assert.Equal(t, tt.expected, template)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
