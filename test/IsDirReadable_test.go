package test

import (
	"jinx/pkg/util/helper"
	"os"
	"path/filepath"
	"testing"
)

func TestIsDirReadability(t *testing.T) {

	tempDir, tempDirErr := os.MkdirTemp("", "test")
	if tempDirErr != nil {
		t.Fatal(tempDirErr)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()
	readableDirPath := filepath.Join(tempDir, "sample_readable")
	mkReadableErr := os.Mkdir(readableDirPath, 0644)
	if mkReadableErr != nil {
		t.Fatal(mkReadableErr)
	}

	notReadableDirPath := filepath.Join(tempDir, "sample_not_readable")
	mkNotReadableErr := os.Mkdir(notReadableDirPath, 0300)
	if mkNotReadableErr != nil {
		t.Fatal(mkNotReadableErr)
	}

	readableDir, _ := helper.IsDirReadable(readableDirPath)
	if readableDir != true {
		t.Errorf("expected %v got %v", true, readableDir)
	}

	notReadableDir, _ := helper.IsDirReadable(notReadableDirPath)
	if notReadableDir != false {
		t.Errorf("expected %v got %v", false, notReadableDir)
	}
}
