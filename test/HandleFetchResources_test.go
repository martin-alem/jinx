package test_test

import (
	"jinx/server_setup/http_server_setup"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleFetchResources(t *testing.T) {
	tempDir := t.TempDir()
	fileHandle, err := os.CreateTemp(tempDir, "test.txt")
	defer func() {
		_ = fileHandle.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		resources map[string]string
		expect    bool
	}{
		{resources: map[string]string{"https://google.com": fileHandle.Name(), "https://facebook.com": fileHandle.Name()}, expect: true},
		{resources: map[string]string{"/invalid/url": filepath.Join(tempDir, "test.txt"), "https://facebook.com": filepath.Join(tempDir, "book.txt")}, expect: false},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			result := http_server_setup.HandleFetchResources(test.resources)
			if (len(result) == 0) != test.expect {
				t.Errorf("expected %v got %v", test.expect, (len(result)) > 0)
			}
		})
	}
}
