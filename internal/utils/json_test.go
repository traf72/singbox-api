package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestParseJson(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		target        any
		expected      any
		expectedError string
	}{
		{
			name:     "Valid Struct",
			input:    `{"name": "John", "age": 30, "email": "john@example.com"}`,
			expected: testStruct{Name: "John", Age: 30, Email: "john@example.com"},
			target:   &testStruct{},
		},
		{
			name:  "Valid Array of Structs",
			input: `[{"name": "Alice", "age": 25, "email": "alice@example.com"}, {"name": "Bob", "age": 40, "email": "bob@example.com"}]`,
			expected: []testStruct{
				{Name: "Alice", Age: 25, Email: "alice@example.com"},
				{Name: "Bob", Age: 40, Email: "bob@example.com"},
			},
			target: &[]testStruct{},
		},
		{
			name:     "Valid Plain String",
			input:    `"Test str"`,
			expected: "Test str",
			target:   new(string),
		},
		{
			name:     "Valid Array of Integers",
			input:    `[1, 2, 3, 4]`,
			expected: []int{1, 2, 3, 4},
			target:   &[]int{},
		},
		{
			name:          "Invalid JSON",
			input:         `{"name": "John", "age": , "email": "john@example.com"}`,
			target:        &testStruct{},
			expectedError: "failed to parse json",
		},
		{
			name:          "Empty Input",
			input:         "",
			target:        &testStruct{},
			expectedError: "failed to parse json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch target := tt.target.(type) {
			case *testStruct:
				err := ParseJson(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			case *string:
				err := ParseJson(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			case *[]testStruct:
				err := ParseJson(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			case *[]int:
				err := ParseJson(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			default:
				t.Fatalf("unsupported target type: %T", tt.target)
			}
		})
	}
}

func assertResult(t *testing.T, err error, expectedError string, expected, target any) {
	t.Helper()

	if expectedError != "" {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedError)
	} else {
		assert.NoError(t, err)
		assert.Equal(t, expected, target)
	}
}
