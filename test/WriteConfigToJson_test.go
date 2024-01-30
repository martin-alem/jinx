package test

import (
	"jinx/pkg/util/helper"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteConfigToJsonFile uses table-driven tests to verify the WriteConfigToJsonFile function.
func TestWriteConfigToJsonFile(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		config   map[string]interface{}
		wantErr  bool
		validate func(t *testing.T, file string) // Optional: A function to validate file contents if needed
	}{
		{
			name: "SimpleConfig",
			config: map[string]interface{}{
				"port":    8080,
				"enabled": true,
			},
			wantErr: false,
		},
		{
			name: "ComplexConfig",
			config: map[string]interface{}{
				"database": map[string]interface{}{
					"host": "localhost",
					"port": 5432,
				},
				"features": []string{"feature1", "feature2"},
			},
			wantErr: false,
		},
		{
			name:    "EmptyConfig",
			config:  map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file
			tempFile := filepath.Join(t.TempDir(), "sample.config.json")

			// Test the function
			err := helper.WriteConfigToJsonFile(tc.config, tempFile)
			if (err != nil) != tc.wantErr {
				t.Errorf("WriteConfigToJsonFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Optional: Validate file contents if validate function is provided
			if tc.validate != nil {
				tc.validate(t, tempFile)
			} else {
				// Default validation checks if the file is not empty
				info, err := os.Stat(tempFile)
				if err != nil || info.Size() == 0 {
					t.Errorf("File is empty or cannot be accessed")
				}
			}
		})
	}
}
