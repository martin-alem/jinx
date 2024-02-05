package test_test

import (
	"io"
	"jinx/internal/jinx_http"
	"jinx/pkg/util/types"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestServe404(t *testing.T) {
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

	// Setup temporary file for testing
	tempFile, err := os.CreateTemp("", "test404Page.*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tempFile.Name())

	_, err = tempFile.WriteString("Custom 404 Page Content")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tempFile.Close()

	tests := []struct {
		name       string
		filePath   string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "File Exists",
			filePath:   tempFile.Name(),
			wantStatus: http.StatusNotFound,
			wantBody:   "Custom 404 Page Content",
		},
		{
			name:       "File Does Not Exist",
			filePath:   "nonexistent.html",
			wantStatus: http.StatusNotFound,
			wantBody:   "404 Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httptest.NewRequest("GET", "http://example.com/foo", nil)
			w := httptest.NewRecorder()

			jx.Serve404(w, tt.filePath)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("serve404() status = %v, wantStatus %v", resp.StatusCode, tt.wantStatus)
			}

			if gotBody := string(body); !strings.Contains(gotBody, tt.wantBody) {
				t.Errorf("serve404() body = %v, wantBody %v", gotBody, tt.wantBody)
			}
		})
	}
}
