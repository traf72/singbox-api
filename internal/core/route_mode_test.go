package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		mode     RouteMode
		expected bool
	}{
		{"Proxy", RouteProxy, true},
		{"Direct", RouteDirect, true},
		{"Block", RouteBlock, true},
		{"Unknown", "Unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.isValid())
		})
	}
}
