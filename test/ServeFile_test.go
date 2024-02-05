package test_test

import (
	"io"
	"jinx/internal/jinx_http"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestServeFile tests the serveFile method for serving static files with correct headers and content.
func TestServeFile(t *testing.T) {
	t.Parallel()
	config := types.JinxHttpServerConfig{
		IP:          "127.0.0.1",
		Port:        8080,
		LogRoot:     t.TempDir(),
		KeyFile:     "",
		CertFile:    "",
		WebsiteRoot: "",
	}

	jx := jinx_http.NewJinxHttpServer(config, t.TempDir())

	tempDir, err := os.MkdirTemp("", "static")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up after the test

	tempFilePath := filepath.Join(tempDir, "testfile.txt")
	expectedContent := "Hello, serveFile!"
	err = os.WriteFile(tempFilePath, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Create a mock HTTP request and response recorder
	req := httptest.NewRequest("GET", "http://example.com/static/testfile.txt", nil)
	w := httptest.NewRecorder()

	// Serve the file
	jx.ServeFile(w, req, tempFilePath)

	resp := w.Result()
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check headers
	if got := resp.Header.Get("Cache-Control"); got != "max-age=3600" {
		t.Errorf("serveFile() Cache-Control = %v, want %v", got, "max-age=3600")
	}

	if got := resp.Header.Get("Server"); got != constant.SOFTWARE_NAME {
		t.Errorf("serveFile() Server = %v, want %v", got, constant.SOFTWARE_NAME)
	}

	// Check content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if got := string(body); got != expectedContent {
		t.Errorf("serveFile() body = %v, want %v", got, expectedContent)
	}

	// Optionally, check for correct MIME type based on file extension
	if got, want := resp.Header.Get("Content-Type"), "text/plain; charset=utf-8"; got != want {
		t.Errorf("serveFile() Content-Type = %v, want %v", got, want)
	}
}
