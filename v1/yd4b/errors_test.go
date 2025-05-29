package yd4b_test

import (
	"testing"

	"github.com/aethiopicuschan/yd4b-go/v1/yd4b"
	"github.com/stretchr/testify/assert"
)

func TestErrorBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		statusCode   int
		message      string
		expectedJSON string
	}{
		{
			name:         "Not Found Error",
			statusCode:   404,
			message:      "resource not found",
			expectedJSON: `{"status_code":404,"message":"resource not found"}`,
		},
		{
			name:         "Internal Server Error",
			statusCode:   500,
			message:      "internal server error",
			expectedJSON: `{"status_code":500,"message":"internal server error"}`,
		},
		{
			name:         "Custom Error",
			statusCode:   123,
			message:      "something went wrong",
			expectedJSON: `{"status_code":123,"message":"something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create the error
			e := yd4b.NewError(tt.statusCode, tt.message)
			assert.NotNil(t, e, "NewError should return a non-nil *Error")

			// Check the StatusCode field
			assert.Equal(t, tt.statusCode, e.StatusCode, "StatusCode should be set correctly")

			// Check the Message field
			assert.Equal(t, tt.message, e.Message, "Message should be set correctly")

			// Check Error() method output
			assert.Equal(t, tt.message, e.Error(), "Error() should return the message")

			// Check JSON output
			jsonBytes, err := e.ToJSON()
			assert.NoError(t, err, "ToJSON should not return an error")
			assert.JSONEq(t, tt.expectedJSON, string(jsonBytes), "JSON output should match expected")
		})
	}
}
