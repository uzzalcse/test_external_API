package services

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

// ErrorReader is a custom reader that always returns an error
type ErrorReader struct{}

func (er ErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

type MockHTTPClient struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

var (
	originalTransport = http.DefaultClient.Transport
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		httpError     error
		useErrorReader bool
		wantErr       bool
		wantEmpty     bool
	}{
		{
			name: "successful request",
			responseBody: `[{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"email": "john@example.com",
				"phone": "1234567890"
			}]`,
			responseStatus: http.StatusOK,
			httpError:     nil,
			useErrorReader: false,
			wantErr:       false,
			wantEmpty:     false,
		},
		{
			name:           "non-200 status code",
			responseBody:   `{}`,
			responseStatus: http.StatusNotFound,
			httpError:     nil,
			useErrorReader: false,
			wantErr:       true,
			wantEmpty:     true,
		},
		{
			name:           "http client error",
			responseBody:   ``,
			responseStatus: 0,
			httpError:     io.ErrUnexpectedEOF,
			useErrorReader: false,
			wantErr:       true,
			wantEmpty:     true,
		},
		{
			name: "invalid json response",
			responseBody: `[{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"email": "john@example.com",
				"phone": "1234567890"`, // incomplete JSON
			responseStatus: http.StatusOK,
			httpError:     nil,
			useErrorReader: false,
			wantErr:       true,
			wantEmpty:     true,
		},
		{
			name:           "response body read error",
			responseBody:   "",
			responseStatus: http.StatusOK,
			httpError:     nil,
			useErrorReader: true,
			wantErr:       true,
			wantEmpty:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var responseBody io.ReadCloser
			if tt.useErrorReader {
				responseBody = io.NopCloser(ErrorReader{})
			} else {
				responseBody = io.NopCloser(bytes.NewBufferString(tt.responseBody))
			}

			// Create a mock response
			response := &http.Response{
				StatusCode: tt.responseStatus,
				Body:       responseBody,
			}

			// Create a mock transport
			mockTransport := &MockHTTPClient{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					if tt.httpError != nil {
						return nil, tt.httpError
					}
					return response, nil
				},
			}

			// Replace the default client's transport with our mock
			http.DefaultClient.Transport = mockTransport

			// Create a hook for testing log output
			var logOutput strings.Builder
			log.SetOutput(&logOutput)

			// Call the function
			users := GetUser()

			// Restore the original transport
			http.DefaultClient.Transport = originalTransport

			// Check for expected errors in log output
			if tt.wantErr && !strings.Contains(logOutput.String(), "Failed") {
				t.Errorf("GetUser() expected error logs, got none")
			}

			// Check if users slice is empty when expected
			if tt.wantEmpty && len(users) > 0 {
				t.Errorf("GetUser() expected empty slice, got %v users", len(users))
			}

			// For successful case, check if we got users
			if !tt.wantEmpty && len(users) == 0 {
				t.Errorf("GetUser() expected non-empty slice, got empty slice")
			}

			// For read error case, verify specific error message
			if tt.useErrorReader {
				expectedError := "Failed to read response body"
				if !strings.Contains(logOutput.String(), expectedError) {
					t.Errorf("Expected error message containing '%s', got '%s'", expectedError, logOutput.String())
				}
			}
		})
	}
}