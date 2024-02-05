package test_test

import (
	"jinx/internal/jinx_http"
	"jinx/pkg/util/types"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFilePath(t *testing.T) {
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
		request *http.Request
		want    string
		wantErr bool
	}{
		{
			name:    "Root URL for Default Host",
			request: httptest.NewRequest("GET", "http://localhost/", nil),
			want:    filepath.Join(defaultWebRoot, "index.html"),
			wantErr: false,
		},
		{
			name:    "Root URL for Specific Host",
			request: httptest.NewRequest("GET", "http://example.com/", nil),
			want:    filepath.Join(hostWebRoot, "index.html"),
			wantErr: false,
		},
		{
			name:    "Non-existent File",
			request: httptest.NewRequest("GET", "http://localhost/nonexistent.html", nil),
			want:    filepath.Join(defaultWebRoot, "404.html"),
			wantErr: true,
		},

		{
			name:    "Non-existent File for specific host",
			request: httptest.NewRequest("GET", "http://example.com/nonexistent.html", nil),
			want:    filepath.Join(hostWebRoot, "404.html"),
			wantErr: true,
		},
		{
			name:    "Directory Path",
			request: httptest.NewRequest("GET", "http://localhost/somedir/", nil),
			want:    filepath.Join(defaultWebRoot, "404.html"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jx.ResolveFilePath(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.HasSuffix(got, tt.want) {
				t.Errorf("resolveFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
