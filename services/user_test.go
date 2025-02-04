// package services

// import (
// 	"bytes"
// 	"errors"
// 	"io"
// 	"log"
// 	"net/http"
// 	"strings"
// 	"testing"
// )

// // ErrorReader is a custom reader that always returns an error
// type ErrorReader struct{}

// func (er ErrorReader) Read(p []byte) (n int, err error) {
// 	return 0, errors.New("forced read error")
// }

// type MockHTTPClient struct {
// 	RoundTripFunc func(req *http.Request) (*http.Response, error)
// }

// func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
// 	return m.RoundTripFunc(req)
// }

// var (
// 	originalTransport = http.DefaultClient.Transport
// )

// func TestGetUser(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		responseBody   string
// 		responseStatus int
// 		httpError     error
// 		useErrorReader bool
// 		wantErr       bool
// 		wantEmpty     bool
// 	}{
// 		{
// 			name: "successful request",
// 			responseBody: `[{
// 				"id": 1,
// 				"first_name": "John",
// 				"last_name": "Doe",
// 				"email": "john@example.com",
// 				"phone": "1234567890"
// 			}]`,
// 			responseStatus: http.StatusOK,
// 			httpError:     nil,
// 			useErrorReader: false,
// 			wantErr:       false,
// 			wantEmpty:     false,
// 		},
// 		{
// 			name:           "non-200 status code",
// 			responseBody:   `{}`,
// 			responseStatus: http.StatusNotFound,
// 			httpError:     nil,
// 			useErrorReader: false,
// 			wantErr:       true,
// 			wantEmpty:     true,
// 		},
// 		{
// 			name:           "http client error",
// 			responseBody:   ``,
// 			responseStatus: 0,
// 			httpError:     io.ErrUnexpectedEOF,
// 			useErrorReader: false,
// 			wantErr:       true,
// 			wantEmpty:     true,
// 		},
// 		{
// 			name: "invalid json response",
// 			responseBody: `[{
// 				"id": 1,
// 				"first_name": "John",
// 				"last_name": "Doe",
// 				"email": "john@example.com",
// 				"phone": "1234567890"`, // incomplete JSON
// 			responseStatus: http.StatusOK,
// 			httpError:     nil,
// 			useErrorReader: false,
// 			wantErr:       true,
// 			wantEmpty:     true,
// 		},
// 		{
// 			name:           "response body read error",
// 			responseBody:   "",
// 			responseStatus: http.StatusOK,
// 			httpError:     nil,
// 			useErrorReader: true,
// 			wantErr:       true,
// 			wantEmpty:     true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var responseBody io.ReadCloser
// 			if tt.useErrorReader {
// 				responseBody = io.NopCloser(ErrorReader{})
// 			} else {
// 				responseBody = io.NopCloser(bytes.NewBufferString(tt.responseBody))
// 			}

// 			// Create a mock response
// 			response := &http.Response{
// 				StatusCode: tt.responseStatus,
// 				Body:       responseBody,
// 			}

// 			// Create a mock transport
// 			mockTransport := &MockHTTPClient{
// 				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
// 					if tt.httpError != nil {
// 						return nil, tt.httpError
// 					}
// 					return response, nil
// 				},
// 			}

// 			// Replace the default client's transport with our mock
// 			http.DefaultClient.Transport = mockTransport

// 			// Create a hook for testing log output
// 			var logOutput strings.Builder
// 			log.SetOutput(&logOutput)

// 			// Call the function
// 			users := GetUser()

// 			// Restore the original transport
// 			http.DefaultClient.Transport = originalTransport

// 			// Check for expected errors in log output
// 			if tt.wantErr && !strings.Contains(logOutput.String(), "Failed") {
// 				t.Errorf("GetUser() expected error logs, got none")
// 			}

// 			// Check if users slice is empty when expected
// 			if tt.wantEmpty && len(users) > 0 {
// 				t.Errorf("GetUser() expected empty slice, got %v users", len(users))
// 			}

// 			// For successful case, check if we got users
// 			if !tt.wantEmpty && len(users) == 0 {
// 				t.Errorf("GetUser() expected non-empty slice, got empty slice")
// 			}

// 			// For read error case, verify specific error message
// 			if tt.useErrorReader {
// 				expectedError := "Failed to read response body"
// 				if !strings.Contains(logOutput.String(), expectedError) {
// 					t.Errorf("Expected error message containing '%s', got '%s'", expectedError, logOutput.String())
// 				}
// 			}
// 		})
// 	}
// }


package services

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockHTTPClient is a mock implementation of http.RoundTripper
type MockHTTPClient struct {
	mock.Mock
}

// ErrorReader is a custom reader that returns an error on Read
type ErrorReader struct{}

func (er ErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockHTTPClient)
		expectedUsers  []User
		expectedEmpty  bool
	}{
		{
			name: "successful request",
			setupMock: func(m *MockHTTPClient) {
				responseBody := `[{
					"id": 1,
					"first_name": "John",
					"last_name": "Doe",
					"email": "john@example.com",
					"phone": "1234567890"
				}]`
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
				}
				m.On("RoundTrip", mock.Anything).Return(response, nil)
			},
			expectedUsers: []User{
				{
					ID:        1,
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
					Phone:     "1234567890",
				},
			},
			expectedEmpty: false,
		},
		{
			name: "non-200 status code",
			setupMock: func(m *MockHTTPClient) {
				response := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				}
				m.On("RoundTrip", mock.Anything).Return(response, nil)
			},
			expectedUsers: nil,
			expectedEmpty: true,
		},
		{
			name: "http client error",
			setupMock: func(m *MockHTTPClient) {
				m.On("RoundTrip", mock.Anything).Return(nil, io.ErrUnexpectedEOF)
			},
			expectedUsers: nil,
			expectedEmpty: true,
		},
		{
			name: "invalid json response",
			setupMock: func(m *MockHTTPClient) {
				responseBody := `[{
					"id": 1,
					"first_name": "John",` // incomplete JSON
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
				}
				m.On("RoundTrip", mock.Anything).Return(response, nil)
			},
			expectedUsers: nil,
			expectedEmpty: true,
		},
		{
			name: "response body read error",
			setupMock: func(m *MockHTTPClient) {
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(ErrorReader{}), // Use ErrorReader to force read error
				}
				m.On("RoundTrip", mock.Anything).Return(response, nil)
			},
			expectedUsers: nil,
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client and restore original transport after test
			mockClient := new(MockHTTPClient)
			originalTransport := http.DefaultClient.Transport
			defer func() {
				http.DefaultClient.Transport = originalTransport
			}()

			// Setup mock expectations
			tt.setupMock(mockClient)
			http.DefaultClient.Transport = mockClient

			// Call the function under test
			users := GetUser()

			// Assertions using testify
			if tt.expectedEmpty {
				assert.Empty(t, users, "Expected empty slice of users")
			} else {
				require.NotEmpty(t, users, "Expected non-empty slice of users")
				require.Equal(t, len(tt.expectedUsers), len(users), "Expected same number of users")
				
				// Compare each user's fields
				for i, expectedUser := range tt.expectedUsers {
					assert.Equal(t, expectedUser.ID, users[i].ID, "User ID mismatch")
					assert.Equal(t, expectedUser.FirstName, users[i].FirstName, "FirstName mismatch")
					assert.Equal(t, expectedUser.LastName, users[i].LastName, "LastName mismatch")
					assert.Equal(t, expectedUser.Email, users[i].Email, "Email mismatch")
					assert.Equal(t, expectedUser.Phone, users[i].Phone, "Phone mismatch")
				}
			}

			// Verify that all expected mock calls were made
			mockClient.AssertExpectations(t)
		})
	}
}