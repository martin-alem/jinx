package util

import (
	"io/fs"
	"net"
	"path/filepath"
)

func IsLocalhostOrIP(host string) bool {
	// Check if the host is "localhost"
	if host == "localhost" {
		return true
	}

	// Parse the host as an IP address
	ip := net.ParseIP(host)
	if ip != nil {
		// Check if the IP address is in the loop back range
		if ip.IsLoopback() {
			return true
		}
	}

	return false
}

func Contains[T comparable](list []T, element T) bool {
	for _, el := range list {
		if el == element {
			return true
		}
	}
	return false
}

func FileExist(dir string, file string) (bool, string) {
	found := false
	abs := file
	walkFunc := func(path string, fileInfo fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() && fileInfo.Name() == file {
			found = true
			abs = path
			return filepath.SkipDir // Stop walking once the file is found
		}

		return nil
	}

	_ = filepath.WalkDir(dir, walkFunc) // Ignore the error

	return found, abs
}
