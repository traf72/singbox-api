package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traf72/singbox-api/internal/apperr"
)

func TestSetRouteModeValidation(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		message string
	}{
		{name: "Empty", mode: "", message: "route mode is empty"},
		{name: "Unknown", mode: "unknown", message: "route mode 'unknown' is unknown"},
		{name: "Block", mode: "block", message: "route mode must be 'proxy' or 'direct'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetRouteMode(&RouteModeRequest{Mode: tt.mode}, false)

			assert.Equal(t, apperr.Validation, err.Kind())
			assert.Equal(t, tt.message, err.Msg())
		})
	}
}
