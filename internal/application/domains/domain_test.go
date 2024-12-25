package domains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseType(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      kind
		expectedError error
	}{
		{"Keyword", "keyword", Keyword, nil},
		{"Keyword_TrimSpaces", "  keyword\n", Keyword, nil},
		{"Domain", "domain", Suffix, nil},
		{"Domain_TrimSpaces", "\tdomain  \n", Suffix, nil},
		{"EmptyInput", "", -1, errEmptyTemplateType},
		{"SpaceOnlyInput", "", -1, errEmptyTemplateType},
		{"UnknownType", "unknown", -1, invalidTemplateType("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseType(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestParseDomain(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		expectedError error
	}{
		{"TwoLevelDomain", "google.com", "google.com", nil},
		{"ThreeLevelDomain", "\tGmail.Google.COM \n", "gmail.google.com", nil},
		{"FourLevelDomain", "m.gmail.google.com", "m.gmail.google.com", nil},
		{"SingleLevelDomain", "com", "", invalidDomain("com")},
		{"DomainStartingWithDot", ".com", "", invalidDomain(".com")},
		{"EmptyInput", "", "", errEmptyTemplate},
		{"SpaceOnlyInput", "\n\t", "", errEmptyTemplate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDomain(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      template
		expectedError error
	}{
		{"Domain", "GOOGLE.com", template{kind: Domain, text: "google.com"}, nil},
		{"Suffix", "DOMAIN:mail.google.com", template{kind: Suffix, text: "mail.google.com"}, nil},
		{"Keyword", "\t KEYworD:Google \n", template{kind: Keyword, text: "google"}, nil},
		{"EmptyInput", "", template{}, errEmptyTemplate},
		{"SpaceOnlyInput", "\n\r ", template{}, errEmptyTemplate},
		{"EmptyType", ":google.com", template{}, errEmptyTemplateType},
		{"EmptySuffix", "domain:", template{}, errEmptyTemplate},
		{"EmptyKeyword", "keyword:\t ", template{}, errEmptyTemplate},
		{"TooManyParts", "domain:google.com:extra", template{}, tooManyParts("domain:google.com:extra")},
		{"InvalidType", "InvalidType:google.com", template{}, invalidTemplateType("InvalidType")},
		{"InvalidDomain", ".com", template{}, invalidDomain(".com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parse(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
