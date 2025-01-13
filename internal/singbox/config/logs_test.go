package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevel_IsValid(t *testing.T) {
	validLogLevels := []LogLevel{Trace, Debug, Info, Warn, Error, Fatal, Panic, " WaRn "}
	for _, level := range validLogLevels {
		t.Run(string(level), func(t *testing.T) {
			assert.True(t, level.isValid())
		})
	}

	invalidLogLevels := []LogLevel{"verbose", "critical", "alert", "infoo", ""}
	for _, level := range invalidLogLevels {
		t.Run(string(level), func(t *testing.T) {
			assert.False(t, level.isValid())
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{" WaRn ", "warn"},
		{"\tTRACE ", "trace"},
		{" Unknown\n", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}
