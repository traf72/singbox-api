package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name        string
		envKey      string
		envIsSet    bool
		envVal      string
		defaultVal  string
		expectedVal string
	}{
		{
			name:        "Environment variable is set",
			envKey:      "TEST_ENV",
			envIsSet:    true,
			envVal:      "set_value",
			defaultVal:  "default_value",
			expectedVal: "set_value",
		},
		{
			name:        "Environment variable is unset",
			envKey:      "TEST_ENV",
			envIsSet:    false,
			envVal:      "",
			defaultVal:  "default_value",
			expectedVal: "default_value",
		},
		{
			name:        "Environment variable is explicitly empty",
			envKey:      "TEST_ENV",
			envIsSet:    true,
			envVal:      "",
			defaultVal:  "default_value",
			expectedVal: "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Unsetenv(tt.envKey)
			assert.NoError(t, err)

			if tt.envIsSet {
				err := os.Setenv(tt.envKey, tt.envVal)
				assert.NoError(t, err)
			}

			result := GetEnv(tt.envKey, tt.defaultVal)
			assert.Equal(t, tt.expectedVal, result)
		})
	}
}
