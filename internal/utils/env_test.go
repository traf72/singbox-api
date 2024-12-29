package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		envKey        string
		envIsSet      bool
		envValue      string
		defaultValue  string
		expectedValue string
	}{
		{
			name:          "Environment variable is set",
			envKey:        "TEST_ENV",
			envIsSet:      true,
			envValue:      "set_value",
			defaultValue:  "default_value",
			expectedValue: "set_value",
		},
		{
			name:          "Environment variable is unset",
			envKey:        "TEST_ENV",
			envIsSet:      false,
			envValue:      "",
			defaultValue:  "default_value",
			expectedValue: "default_value",
		},
		{
			name:          "Environment variable is explicitly empty",
			envKey:        "TEST_ENV",
			envIsSet:      true,
			envValue:      "",
			defaultValue:  "default_value",
			expectedValue: "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Unsetenv(tt.envKey)
			assert.NoError(t, err)

			if tt.envIsSet {
				err := os.Setenv(tt.envKey, tt.envValue)
				assert.NoError(t, err)
			}

			result := GetEnv(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}
