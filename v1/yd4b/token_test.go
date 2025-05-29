package yd4b_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/aethiopicuschan/yd4b-go/v1/yd4b"
	"github.com/stretchr/testify/assert"
)

func TestTokenRequest_ToRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         *yd4b.TokenRequest
		endpoint    string
		wantMethod  string
		wantURL     string
		wantBody    string
		expectError bool
	}{
		{
			name:       "valid request",
			req:        &yd4b.TokenRequest{GrantType: "client_credentials", ClientID: "ID123", SecretKey: "SK456"},
			endpoint:   "https://api.example.com/token",
			wantMethod: "POST",
			wantURL:    "https://api.example.com/token",
			wantBody:   `{"grant_type":"client_credentials","client_id":"ID123","secret_key":"SK456"}`,
		},
		{
			name:        "invalid URL",
			req:         &yd4b.TokenRequest{GrantType: "a", ClientID: "b", SecretKey: "c"},
			endpoint:    `://bad-url`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := tt.req.ToRequest(tt.endpoint)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantMethod, req.Method)
			assert.Equal(t, tt.wantURL, req.URL.String())
			body, _ := io.ReadAll(req.Body)
			assert.JSONEq(t, tt.wantBody, string(body))
		})
	}
}

func TestClient_GetToken(t *testing.T) {
	tests := []struct {
		name          string
		doFunc        func(req *http.Request) (*http.Response, error)
		wantErrSubstr string
		wantRes       yd4b.TokenResponse
	}{
		{
			name: "success and trim prefix",
			doFunc: func(req *http.Request) (*http.Response, error) {
				json := `{"scope":"J1","token_type":"Bearer","expires_in":3600,"token":"Token: abc123"}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(json))}, nil
			},
			wantRes: yd4b.TokenResponse{Scope: "J1", TokenType: "Bearer", ExpiresIn: 3600, Token: "abc123"},
		},
		{
			name: "non-200 status",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString(`error`))}, nil
			},
			wantErrSubstr: "unexpected status code",
		},
		{
			name: "decoding error",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{bad json}`))}, nil
			},
			wantErrSubstr: "json decoding error",
		},
		{
			name: "client do error",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network fail")
			},
			wantErrSubstr: "client do error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := yd4b.NewClient("https://api.example.com", "ID", "SECRET", "1.2.3.4")
			client.SetDoFunc(tt.doFunc)

			res, err := client.GetToken()

			if tt.wantErrSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSubstr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRes, res)
		})
	}
}
