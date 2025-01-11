package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestFromJSON(t *testing.T) {
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
				err := FromJSON(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			case *string:
				err := FromJSON(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			case *[]testStruct:
				err := FromJSON(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			case *[]int:
				err := FromJSON(strings.NewReader(tt.input), target)
				assertResult(t, err, tt.expectedError, tt.expected, *target)
			default:
				t.Fatalf("unsupported target type: %T", tt.target)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		options       *JSONOptions
		expected      string
		expectedError string
	}{
		{
			name:  "Valid Struct with Indent",
			input: testStruct{Name: "John", Age: 30, Email: "john@example.com"},
			options: &JSONOptions{
				Indent: "  ",
			},
			expected: `{
  "name": "John",
  "age": 30,
  "email": "john@example.com"
}
`,
		},
		{
			name:    "Valid Struct with default options",
			input:   testStruct{Name: "<div>John</div>", Age: 30, Email: "john@example.com"},
			options: nil,
			expected: `{"name":"\u003cdiv\u003eJohn\u003c/div\u003e","age":30,"email":"john@example.com"}
`,
		},
		{
			name:  "Valid Array of Structs with Indent",
			input: []testStruct{{Name: "Alice", Age: 25, Email: "alice@example.com"}, {Name: "Bob", Age: 40, Email: "bob@example.com"}},
			options: &JSONOptions{
				Indent: "\t",
			},
			expected: `[
	{
		"name": "Alice",
		"age": 25,
		"email": "alice@example.com"
	},
	{
		"name": "Bob",
		"age": 40,
		"email": "bob@example.com"
	}
]
`,
		},
		{
			name:  "Valid String with EscapeHTML Disabled",
			input: "<div>Hello</div>",
			options: &JSONOptions{
				EscapeHTML: false,
				Indent:     "",
			},
			expected: "\"<div>Hello</div>\"\n",
		},
		{
			name:  "Valid String with EscapeHTML Enabled",
			input: "<div>Hello</div>",
			options: &JSONOptions{
				EscapeHTML: true,
				Indent:     "",
			},
			expected: `"\u003cdiv\u003eHello\u003c/div\u003e"
`,
		},
		{
			name:     "Valid Array of Integers",
			input:    []int{1, 2, 3, 4},
			options:  &JSONOptions{},
			expected: "[1,2,3,4]\n",
		},
		{
			name:     "Nil Input with Default Options",
			input:    nil,
			options:  &JSONOptions{},
			expected: "null\n",
		},
		{
			name:  "Invalid Input (Channel)",
			input: make(chan int),
			options: &JSONOptions{
				Indent: "",
			},
			expected:      "",
			expectedError: "failed to serialize to json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := ToJSON(&buf, tt.input, tt.options)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
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
