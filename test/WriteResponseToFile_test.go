package test

import (
	"bytes"
	"io"
	"jinx/server_setup/http_server_setup"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteResponseToFile(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	file1Path := filepath.Join(tempDir, "file.txt")
	file1, err := os.Create(file1Path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = file1.Close()
	}()

	mockResponse := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString("mock response")),
	}
	tests := []struct {
		file string
		err  bool
	}{
		{file1.Name(), false},
		{"/invalid/path/file.txt", true},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := http_server_setup.WriteResponseToFile(test.file, mockResponse)
			if (err == nil) == test.err {
				t.Errorf("expected %v but got %v", test.err, err)
			}
		})
	}
}
