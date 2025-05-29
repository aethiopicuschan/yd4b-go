package yd4b_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/aethiopicuschan/yd4b-go/v1/yd4b"
	"github.com/stretchr/testify/assert"
)

func TestClientMethods(t *testing.T) {
	tests := []struct {
		name          string
		setupToken    bool
		ecuid         string
		doFunc        func(req *http.Request) (*http.Response, error)
		expectedAuth  string
		expectedECUID string
	}{
		{
			name:       "Without token",
			setupToken: false,
			doFunc: func(req *http.Request) (*http.Response, error) {
				// Read and return empty response
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
			},
			expectedAuth:  "",
			expectedECUID: "",
		},
		{
			name:       "With token",
			setupToken: true,
			// stub doFunc to inspect request
			doFunc: func(req *http.Request) (*http.Response, error) {
				// Read and return empty response
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
			},
			expectedAuth:  "Bearer test-token",
			expectedECUID: "",
		},
		{
			name:       "With ECUID",
			setupToken: false,
			ecuid:      "ec123",
			doFunc: func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Header.Get("X-ECUID"), "ec123", "ECUID header should be set if provided")
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil
			},
			expectedAuth:  "",
			expectedECUID: "ec123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			client := yd4b.NewClient("https://api.example.com", "id", "secret", "1.2.3.4")
			client.SetDoFunc(tt.doFunc)
			if tt.setupToken {
				client.SetToken("test-token")
			}
			if tt.ecuid != "" {
				client.SetECUID(tt.ecuid)
				// Extend do to include ECUID header
				origDo := yd4b.GetDo(client)
				ecuid := yd4b.GetECUID(client)
				client.SetDoFunc(func(req *http.Request) (*http.Response, error) {
					req.Header.Set("X-ECUID", ecuid)
					return origDo(req)
				})
			}

			// Act: create a dummy request and call do
			origin := yd4b.GetOrigin(client)
			req, err := http.NewRequest("GET", origin+"/test", nil)
			assert.NoError(t, err)
			resp, err := yd4b.Do(client, req)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)

			// Verify headers
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "Content-Type header")
			assert.Equal(t, "1.2.3.4", req.Header.Get("x-forwarded-for"), "X-Forwarded-For header")
			if tt.expectedAuth != "" {
				assert.Equal(t, tt.expectedAuth, req.Header.Get("Authorization"), "Authorization header")
			} else {
				assert.Empty(t, req.Header.Get("Authorization"), "Authorization should be empty when no token set")
			}

		})
	}
}
