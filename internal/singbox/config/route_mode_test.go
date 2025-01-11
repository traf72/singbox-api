package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteModeValidate(t *testing.T) {
	tests := []struct {
		name     string
		mode     RouteMode
		expected error
	}{
		{"Proxy", RouteProxy, nil},
		{"Direct", RouteDirect, nil},
		{"Block", RouteBlock, nil},
		{"Unknown", "Unknown", errors.New("invalid route mode 'Unknown'")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.Validate())
		})
	}
}

func TestRouteModeFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    RouteMode
		expectedErr error
	}{
		{"Proxy", "proxy", RouteProxy, nil},
		{"Proxy_TrimSpaces_LowerCase", "  PROXY\n", RouteProxy, nil},
		{"Block", "block", RouteBlock, nil},
		{"Direct", "direct", RouteDirect, nil},
		{"EmptyInput", "", "", errors.New("route mode is empty")},
		{"SpaceOnlyInput", " \n\r\t", "", errors.New("route mode is empty")},
		{"UnknownMode", "unknown", "", errors.New("route mode 'unknown' is unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RouteModeFromString(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
