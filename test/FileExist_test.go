package test

import (
	"jinx/pkg/util"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExistInDir(t *testing.T) {

	tempDir, tempDirErr := os.MkdirTemp("", "repo")
	if tempDirErr != nil {
		t.Fatalf("Failed to create temporary directory: %v", tempDirErr)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up after the test.

	if mkDirErr := os.Mkdir(filepath.Join(tempDir, "test"), 0755); mkDirErr != nil {
		t.Fatalf("Failed to create test directory: %v", mkDirErr)
	}

	file1, file1Err := os.OpenFile(filepath.Join(tempDir, "test", "file1.txt"), os.O_CREATE, 0644)
	if file1Err != nil {
		t.Fatalf("Failed to create file in test: %v", file1Err)
	}
	defer func() {
		_ = file1.Close()
	}()

	if mkDirErr := os.Mkdir(filepath.Join(tempDir, "another"), 0755); mkDirErr != nil {
		t.Fatalf("Failed to create another directory: %v", mkDirErr)
	}

	file2, file2Err := os.OpenFile(filepath.Join(tempDir, "another", "file2.txt"), os.O_CREATE, 0644)
	if file2Err != nil {
		t.Fatalf("Failed to create file in another: %v", file1Err)
	}
	defer func() {
		_ = file2.Close()
	}()

	ok1, _ := util.FileExist(filepath.Join(tempDir, "test"), "file1.txt")
	if !ok1 {
		t.Errorf("Expected 'file1.txt' to exist, but it does not.")
	}

	ok2, _ := util.FileExist(filepath.Join(tempDir, "another"), "file2.txt")
	if !ok2 {
		t.Errorf("Expected 'file2.txt' to exist, but it does not.")
	}

	ok3, _ := util.FileExist(filepath.Join(tempDir, "another"), "file3.txt")
	if ok3 {
		t.Errorf("Expected 'file2.txt' not to exist, but it does.")
	}
}
