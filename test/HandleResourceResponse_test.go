package test

import (
	"bytes"
	"io"
	"jinx/pkg/util/types"
	"jinx/server_setup/http_server_setup"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// mockHttpResponse creates a mock HTTP response with the specified content for testing purposes.
func mockHttpResponse(bodyContent string) *http.Response {
	return &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(bodyContent)),
	}
}

// TestHandleResourceResponse tests the HandleResourceResponse function for correctness.
func TestHandleResourceResponse(t *testing.T) {
	// Create a temporary directory to simulate the website root.
	tempDir := t.TempDir()

	// Create a subdirectory within the temporary directory to simulate the images' directory.
	imagesDir := filepath.Join(tempDir, "images")
	if err := os.Mkdir(imagesDir, 0755); err != nil {
		t.Fatalf("Failed to create images directory: %v", err)
	}

	// Define a mock resource response.
	resource := &types.JinxResourceResponse{
		Res:      mockHttpResponse("test content"),
		Filename: "test_file.txt",
	}

	// Call the function with the mock response and temporary directories.
	http_server_setup.HandleResourceResponse(tempDir, imagesDir, resource)

	// Verify the file was created and contains the expected content.
	expectedFilePath := filepath.Join(tempDir, resource.Filename)
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read expected file: %s, error: %v", expectedFilePath, err)
	}

	if string(content) != "test content" {
		t.Errorf("File content does not match expected content. Got: %s, want: test content", string(content))
	}
}
