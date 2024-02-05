package test_test

import (
	"io"
	"jinx/internal/jinx_http"
	"jinx/pkg/util/types"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()
	tempDir := os.TempDir()
	serverRootDir := filepath.Join(tempDir, "jinx")
	defaultWebRoot := filepath.Join(serverRootDir, "www")
	hostWebRoot := filepath.Join(serverRootDir, "example.com")

	// Create directories to simulate different host directories and default content
	_ = os.MkdirAll(defaultWebRoot, 0755)
	_ = os.MkdirAll(hostWebRoot, 0755)

	// Create dummy files for testing
	_ = os.WriteFile(filepath.Join(defaultWebRoot, "index.html"), []byte("Default Index"), 0644)
	_ = os.WriteFile(filepath.Join(hostWebRoot, "index.html"), []byte("Host Index"), 0644)
	_ = os.WriteFile(filepath.Join(defaultWebRoot, "404.html"), []byte("Default 404"), 0644)
	_ = os.WriteFile(filepath.Join(hostWebRoot, "404.html"), []byte("Host 404"), 0644)

	config := types.JinxHttpServerConfig{
		IP:          "127.0.0.1",
		Port:        8080,
		LogRoot:     serverRootDir,
		KeyFile:     "",
		CertFile:    "",
		WebsiteRoot: serverRootDir,
	}

	jx := jinx_http.NewJinxHttpServer(config, serverRootDir)

	tests := []struct {
		name    string
		request string
		want    string
	}{
		{
			name:    "Root URL for Default Host",
			request: "http://localhost/",
			want:    "Default Index",
		},
		{
			name:    "Root URL for Specific Host",
			request: "http://example.com/",
			want:    "Host Index",
		},
		{
			name:    "Non-existent File",
			request: "http://localhost/nonexistent.html",
			want:    "Default 404",
		},

		{
			name:    "Non-existent File for specific host",
			request: "http://example.com/nonexistent.html",
			want:    "Host 404",
		},
		{
			name:    "Directory Path",
			request: "http://localhost/somedir/",
			want:    "Default 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.request, nil)
			w := httptest.NewRecorder()
			jx.ServeHTTP(w, req)

			resp := w.Result()
			defer func() {
				_ = resp.Body.Close()
			}()

			// Check content
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if got := string(body); got != tt.want {
				t.Errorf("serveFile() body = %v, want %v", got, tt.want)
			}
		})
	}
}
