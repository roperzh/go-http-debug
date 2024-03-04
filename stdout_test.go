package httpdebug

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRawStdout(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	middleware := RawStdout(nextHandler)
	middleware.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != "test response" {
		t.Errorf("Expected response body to be 'test response', got '%s'", string(body))
	}
}

func TestRawStdoutWithLogCapture(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		// reset the log output to stdout once the test is done
		log.SetOutput(os.Stdout)
	}()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	middleware := RawStdout(nextHandler)
	middleware.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != "test response" {
		t.Errorf("Expected response body to be 'test response', got '%s'", string(body))
	}

	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "test response") || !strings.Contains(logOutput, "Status: 200") {
		t.Errorf("Log output did not contain expected response details. Log output: %s", logOutput)
	}
}

func TestJSONStdoutWithLogCapture(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		// reset log output to stdout after the test
		log.SetOutput(os.Stdout)
	}()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("z", "y")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	var reqBody bytes.Buffer
	reqBody.WriteString(`{"foo": "bar"}`)
	req := httptest.NewRequest("GET", "http://example.com/foo", &reqBody)
	req.Header.Add("foo", "bar")
	w := httptest.NewRecorder()

	middleware := JSONStdout(nextHandler)
	middleware.ServeHTTP(w, req)

	logOutput := logBuffer.String()

	expectedOutput := `{"status":200,"path":"http://example.com/foo","method":"GET","request":{"raw_headers":"Foo: bar\r\n","body":"{\"foo\": \"bar\"}"},"response":{"raw_headers":"Z: y\r\n","body":"test response"}}`
	if !strings.Contains(logOutput, expectedOutput) {
		t.Errorf("Log output did not contain the expected transaction details. Log output: %s, Expected: %s", logOutput, string(expectedOutput))
	}
}
