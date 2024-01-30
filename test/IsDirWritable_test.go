package test

import (
	"jinx/pkg/util/helper"
	"os"
	"path/filepath"
	"testing"
)

func TestIsDirWritable(t *testing.T) {

	tempDir, tempDirErr := os.MkdirTemp("", "test")
	if tempDirErr != nil {
		t.Fatal(tempDirErr)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()
	writableDirPath := filepath.Join(tempDir, "sample_writable")
	mkWritableErr := os.Mkdir(writableDirPath, 0755)
	if mkWritableErr != nil {
		t.Fatal(mkWritableErr)
	}

	notWritableDirPath := filepath.Join(tempDir, "sample_not_writable")
	mkNotWritableErr := os.Mkdir(notWritableDirPath, 0555)
	if mkNotWritableErr != nil {
		t.Fatal(mkNotWritableErr)
	}

	writableDir, _ := helper.IsDirWritable(writableDirPath)
	if !writableDir {
		t.Errorf("expected %v got %v", true, writableDir)
	}

	notWritableDir, _ := helper.IsDirWritable(notWritableDirPath)
	if notWritableDir {
		t.Errorf("expected %v got %v", false, notWritableDir)
	}
}
