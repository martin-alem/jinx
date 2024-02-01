package test

import (
	"errors"
	"jinx/server_setup/reverse_proxy_server_setup"
	"os"
	"path/filepath"
	"testing"
)

// TestValidateRouteTablePath uses table-driven tests to verify the ValidateRouteTablePath function.
func TestValidateRouteTablePath(t *testing.T) {
	// Setup - Create a temporary .json file and a temporary file with an incorrect extension.
	tempDir := t.TempDir()
	validFilePath := filepath.Join(tempDir, "valid.json")
	_, err := os.Create(validFilePath)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	invalidFilePath := filepath.Join(tempDir, "invalid.txt")
	_, err = os.Create(invalidFilePath)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Define test cases
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"ValidFile", validFilePath, false},
		{"InvalidExtension", invalidFilePath, true},
		{"NonExistentFile", filepath.Join(tempDir, "nonexistent.json"), true},
		{"EmptyPath", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the function with the test case parameters
			err := reverse_proxy_server_setup.ValidateRouteTablePath(tc.path)

			// Verify the error based on the expected outcome
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateRouteTablePath(%s) got err: %v, wantErr: %v", tc.path, err, tc.wantErr)
			}

			// Additional check for the error content if an error is expected
			if tc.wantErr && err != nil {
				if tc.path == invalidFilePath && !errors.Is(err, os.ErrInvalid) {
					t.Errorf("ValidateRouteTablePath(%s) expected ErrInvalidExtension, got: %v", tc.path, err)
				}
			}
		})
	}
}
