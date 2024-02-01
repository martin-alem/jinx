package test

import (
	"jinx/pkg/util/types"
	"jinx/server_setup/reverse_proxy_server_setup"
	"os"
	"reflect"
	"testing"
)

// Assuming the types package and RouteTable type are defined as follows:
// package types
// type RouteTable map[string]string

func TestLoadRouteTable(t *testing.T) {
	// Create a temporary JSON file with route table content for testing.
	tempFile, err := os.CreateTemp("", "routeTable*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}() // Clean up after the test.

	// Sample route table content.
	routes := `{"path1":"address1", "path2":"address2"}`
	if _, err := tempFile.WriteString(routes); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tempFile.Close() // Close the file to ensure the content is flushed.

	tests := []struct {
		name    string
		path    string
		want    types.RouteTable
		wantErr bool
	}{
		{
			name: "ValidRouteTable",
			path: tempFile.Name(),
			want: types.RouteTable{"path1": "address1", "path2": "address2"},
		},
		{
			name:    "NonexistentFile",
			path:    "nonexistent.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reverse_proxy_server_setup.LoadRouteTable(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRouteTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadRouteTable() got = %v, want %v", got, tt.want)
			}
		})
	}
}
