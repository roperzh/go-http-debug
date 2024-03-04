package httpdebug

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormatXML(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:  "Valid XML",
			input: []byte("<test><sub>value</sub></test>"),
			expected: []byte(`<test>
  <sub>value</sub>
</test>`),
		},
		{
			name:     "Invalid XML",
			input:    []byte("foo"),
			expected: []byte("foo"),
		},
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: []byte(""),
		},
		{
			name:     "Malformed XML",
			input:    []byte("<test><sub>value</sub>"),
			expected: []byte("<test><sub>value</sub>"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatXML(tt.input)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("formatXML(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "Valid JSON",
			input:    []byte(`{"key":"value"}`),
			expected: []byte("{\n  \"key\": \"value\"\n}\n"),
		},
		{
			name:     "Invalid JSON",
			input:    []byte(`{"key":"value"`),
			expected: []byte(`{"key":"value"`),
		},
		{
			name:     "Empty input",
			input:    []byte(""),
			expected: []byte(""),
		},
		{
			name:     "Nested JSON",
			input:    []byte(`{"key":{"innerKey":"innerValue"}}`),
			expected: []byte("{\n  \"key\": {\n    \"innerKey\": \"innerValue\"\n  }\n}\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatJSON(tt.input)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("formatJSON(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDuplicateRequestBody(t *testing.T) {
	tests := []struct {
		name        string
		bodyContent string
	}{
		{
			name:        "Non-empty body",
			bodyContent: "test content",
		},
		{
			name:        "Empty body",
			bodyContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "http://example.com", strings.NewReader(tt.bodyContent))
			duplicatedBody, err := duplicateRequestBody(req)

			if err != nil {
				t.Fatalf("Did not expect an error but got one: %v", err)
			}

			if string(duplicatedBody) != tt.bodyContent {
				t.Errorf("Expected body to be %q, got %q", tt.bodyContent, string(duplicatedBody))
			}

			// Read the body again to ensure it was correctly reset
			newBody, _ := io.ReadAll(req.Body)
			if string(newBody) != tt.bodyContent {
				t.Errorf("The request body was not correctly reset. Expected %q, got %q", tt.bodyContent, string(newBody))
			}
		})
	}
}

// errReader is a custom io.Reader that always returns an error.
type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("intentional error during read")
}

func TestDuplicateRequestBodyError(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", &errReader{})

	_, err := duplicateRequestBody(req)
	if err == nil {
		t.Errorf("Expected an error but didn't get one")
	}

	// Also check that the request body is not nil after the operation
	if req.Body == nil {
		t.Errorf("Expected the request body not to be nil even after an error")
	}
}

// Define a simple next handler for testing purposes
func testHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("test response"))
}

func TestGenerateTransaction(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		contentType        string
		expectError        bool
		expectedReqBody    string
		expectedRespBody   string
		expectedStatusCode int
	}{
		{
			name:               "Valid request and response",
			requestBody:        "<test>value</test>",
			contentType:        "text/xml; charset=utf-8",
			expectError:        false,
			expectedReqBody:    "<test>value</test>",
			expectedRespBody:   "test response",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Without explicit content type",
			requestBody:        "<foo><bar>value</bar></foo>",
			contentType:        "",
			expectError:        false,
			expectedReqBody:    "<foo><bar>value</bar></foo>",
			expectedRespBody:   "test response",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "JSON body",
			requestBody: `{"foo": "bar"}`,
			contentType: "application/json",
			expectError: false,
			expectedReqBody: `{
  "foo": "bar"
}
`,
			expectedRespBody:   "test response",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "unsupported content type",
			requestBody:        "foo",
			contentType:        "image/foo",
			expectError:        false,
			expectedReqBody:    "content-type image/foo preview not supported",
			expectedRespBody:   "test response",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			trx, err := generateTransaction(testHandler, w, req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Did not expect an error but got: %v", err)
				}

				if trx.Method != "POST" || trx.Path != "/test" {
					t.Errorf("Transaction data incorrect, got: %v", trx)
				}

				if trx.Request.Body != tt.expectedReqBody {
					t.Errorf("Expected request body to be %q, got %q", tt.expectedReqBody, trx.Request.Body)
				}

				if trx.Response.Body != tt.expectedRespBody {
					t.Errorf("Expected response body to be %q, got %q", tt.expectedRespBody, trx.Response.Body)
				}

				if trx.Status != tt.expectedStatusCode {
					t.Errorf("Expected status code to be %d, got %d", tt.expectedStatusCode, trx.Status)
				}
			}
		})
	}
}
