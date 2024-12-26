package apperr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppErr(t *testing.T) {
	tests := []struct {
		name         string
		appErr       *Err
		expectedMsg  string
		expectedCode string
		expectedKind ErrKind
		expectedErr  string
	}{
		{
			name:         "Validation Error",
			appErr:       NewValidationErr("VAL001", "Invalid input"),
			expectedMsg:  "Invalid input",
			expectedCode: "VAL001",
			expectedKind: Validation,
			expectedErr:  "VAL001: Invalid input",
		},
		{
			name:         "Not Found Error",
			appErr:       NewNotFoundErr("NOT001", "Resource not found"),
			expectedMsg:  "Resource not found",
			expectedCode: "NOT001",
			expectedKind: NotFound,
			expectedErr:  "NOT001: Resource not found",
		},
		{
			name:         "Fatal Error",
			appErr:       NewFatalErr("FAT001", "Internal server failure"),
			expectedMsg:  "Internal server failure",
			expectedCode: "FAT001",
			expectedKind: Fatal,
			expectedErr:  "FAT001: Internal server failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMsg, tt.appErr.Msg(), "unexpected message")
			assert.Equal(t, tt.expectedCode, tt.appErr.Code(), "unexpected code")
			assert.Equal(t, tt.expectedKind, tt.appErr.Kind(), "unexpected kind")
			assert.Equal(t, tt.expectedErr, tt.appErr.Error(), "unexpected error string")
		})
	}
}
