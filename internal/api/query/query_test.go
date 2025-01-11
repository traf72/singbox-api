package query

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBool(t *testing.T) {
	tests := []struct {
		name        string
		q           url.Values
		key         string
		fallback    bool
		expected    bool
		expectedErr error
	}{
		{
			name:        "Key not present => fallback=false",
			q:           url.Values{},
			key:         "flag",
			fallback:    false,
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "Key not present => fallback=true",
			q:           url.Values{},
			key:         "flag",
			fallback:    true,
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "Key present, empty => true",
			q:           url.Values{"flag": {}},
			key:         "flag",
			fallback:    false,
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "Key present, value='true' => true",
			q:           url.Values{"flag": {"true"}},
			key:         "flag",
			fallback:    false,
			expected:    true,
			expectedErr: nil,
		},
		{
			name:        "Key present, value='false' => false",
			q:           url.Values{"flag": {"false"}},
			key:         "flag",
			fallback:    true,
			expected:    false,
			expectedErr: nil,
		},
		{
			name:        "Invalid boolean => fallback + error",
			q:           url.Values{"flag": {"foobar"}},
			key:         "flag",
			fallback:    true,
			expected:    true,
			expectedErr: errors.New("invalid value 'foobar' for query param 'flag', expected a boolean (true, false)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := GetBool(tt.q, tt.key, tt.fallback)
			assert.Equal(t, tt.expected, val)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
