package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// IsLocalhostOrIP checks if the provided host name is "localhost" or an IP address in the loopback range.
//
// This function is used to determine if a given host name represents a local server instance. It first checks
// if the host name is exactly "localhost". If not, it attempts to parse the host as an IP address. If the parsing
// is successful and the IP address is within the loopback range (typically 127.0.0.1/8 for IPv4), it returns true.
// Otherwise, it returns false.
//
// Parameters:
//   - host: A string representing the host name to be checked.
//
// Returns:
//   - bool: True if the host is "localhost" or a loopback IP address, false otherwise.
//
// Example:
//
//	fmt.Println(IsLocalhostOrIP("localhost")) // Output: true
//	fmt.Println(IsLocalhostOrIP("127.0.0.1")) // Output: true
//	fmt.Println(IsLocalhostOrIP("192.168.1.1")) // Output: false
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

// InList checks if a specific element is present in a slice based on a provided comparison predicate.
// It is a generic function that works with any data type.
//
// This function iterates through each element of the provided slice and uses the provided predicate
// function to compare it with the specified element. The predicate defines the criteria for considering
// two elements as equal. If the predicate returns true for any element in the slice, the function
// returns true, indicating that the element exists in the slice as per the comparison logic defined
// by the predicate. If no match is found by the end of the slice, the function returns false.
//
// This approach allows the function to be used with complex data types and custom comparison logic.
//
// Type Parameters:
//   - T: The type of the elements in the slice, can be any type.
//
// Parameters:
//   - list: A slice of type T elements to be searched.
//   - element: An element of type T to search for in the list.
//   - predicate: A function that takes two arguments of type T and returns a bool.
//     It defines the criteria for comparing elements in the list with the target element.
//
// Returns:
//   - bool: True if the element is found in the list based on the predicate, false otherwise.
//
// Example:
//
//	isEqual := func(a, b int) bool { return a == b }
//	fmt.Println(InList([]int{1, 2, 3}, 2, isEqual)) // Output: true
//
//	isSameLength := func(a, b string) bool { return len(a) == len(b) }
//	fmt.Println(InList([]string{"apple", "banana", "mango"}, "grape", isSameLength)) // Output: true
func InList[T any](list []T, element T, predicate func(a T, b T) bool) bool {
	for _, el := range list {
		if predicate(el, element) == true {
			return true
		}
	}
	return false
}

// FileExist searches for a specific file within a directory and returns a boolean
// indicating whether the file exists, along with the absolute path of the found file.
//
// The function uses the `filepath.WalkDir` function to traverse the directory tree
// starting from the specified directory. It looks for a file that matches the provided
// file name. When the file is found, the traversal is stopped, and the function returns
// true along with the absolute path of the file. If the file is not found in the given
// directory tree, the function returns false and the input file name as the absolute path.
//
// Note that this function ignores any errors encountered during the directory traversal.
//
// Parameters:
//   - dir: The directory path where the search should start.
//   - file: The name of the file to search for.
//
// Returns:
//   - bool: True if the file is found, false otherwise.
//   - string: The absolute path of the found file, or the input file name if not found.
//
// Example:
//
//	exists, path := FileExist("/my/directory", "my_file.txt")
//	if exists {
//	    fmt.Println("File found at:", path)
//	} else {
//	    fmt.Println("File not found")
//	}
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

// IsDirReadable takes in a directory path and check if it has read permissions.
//
// The function uses the `os.ReadDir` function to attempt reading the directory
// If the directory has read permission, os.ReadDir will be successful
// The function returns a boolean and error
//
// Parameters:
//   - dirPath: The directory path to check it read permission
//
// Returns:
//   - bool: True if the directory is readable
//   - error: The error encountered while trying to read the directory
//
// Example:
//
//		readable, err := IsDirReadable("/my/directory")
//	 if err != nil {
//	    log.Println(err)
//	 }
//	 readable will be true if directory is readable.
func IsDirReadable(dirPath string) (readable bool, err error) {
	// Check read permission by reading the contents of the directory
	if _, err := os.ReadDir(dirPath); err != nil {
		return false, err
	}
	return true, nil
}

// IsDirWritable takes in a directory path and check if it has write permissions.
//
// The function tries to create a temp file in the directory using `os.Create`
// If the directory has write permission, os.Create will be successful
// The function returns a boolean and error
// The function also cleans up by closing the file and removing the temp file
//
// Parameters:
//   - dirPath: The directory path to check it write permission
//
// Returns:
//   - bool: True if the directory is writable
//   - error: The error encountered while trying to write to directory
//
// Example:
//
//		writable, err := IsDirWritable("/my/directory")
//	 if err != nil {
//	    log.Println(err)
//	 }
//	 writable will be true if directory is writable.
func IsDirWritable(dirPath string) (writable bool, err error) {
	// Check write permission by trying to create a temporary file in the directory
	tempFilePath := filepath.Join(dirPath, ".tmp_permission_check")
	file, err := os.Create(tempFilePath)
	if err != nil {
		return false, err
	}
	_ = file.Close()

	// Clean up: remove the temporary file
	if err := os.Remove(tempFilePath); err != nil {
		return false, err
	}

	return true, nil
}

// WriteConfigToJsonFile serializes a configuration map to a JSON-formatted file using the encoding/json
// package to handle serialization of complex data types. This ensures that the output is correctly formatted
// as valid JSON, including proper handling of special characters, nested structures, and arrays. It overwrites
// an existing file or creates a new one at the specified path to save the JSON content.
//
// Parameters:
//   - config: A map[string]interface{} representing the configuration settings to be serialized. The keys
//     are string identifiers for configuration parameters, while the values can be any data type
//     supported by JSON, including nested maps and slices.
//   - file: The file path where the JSON-formatted configuration will be saved. If the file exists, it will
//     be overwritten; if not, a new file will be created.
//
// Returns:
//   - An error if any step of the file writing process fails, including file creation, JSON serialization,
//     or writing to the file. Returns nil if the operation completes successfully.
//
// This function is particularly useful for saving complex configurations that include hierarchical settings
// or multiple data types. It abstracts away the manual construction of JSON strings, relying instead on the
// robust serialization capabilities of the encoding/json package.
func WriteConfigToJsonFile(config map[string]any, file string) error {

	// Marshal the config map to a JSON-formatted byte slice.
	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err // Return serialization errors.
	}

	// Write the JSON data to the specified file.
	err = os.WriteFile(file, jsonData, 0644)
	if err != nil {
		return err // Return file writing errors.
	}

	return nil // Indicate success.
}

func SingleJoiningSlash(base, path string) string {
	baseSlash := strings.HasSuffix(base, "/")
	pathSlash := strings.HasPrefix(path, "/")
	switch {
	case baseSlash && pathSlash:
		return base + path[1:]
	case !baseSlash && !pathSlash:
		return base + "/" + path
	default:
		return base + path
	}
}

func ValidatePort(port int) (bool, error) {
	// Check if port is in the valid range (1-65535)
	if port < 1 || port > 65535 {
		return false, fmt.Errorf("port must be between 1 and 65535")
	}

	return true, nil
}

// transfer is a utility function designed to relay data between two streams: `src` (source) and `dst` (destination).
// It reads data from `src` and writes it to `dst`, facilitating the bidirectional flow of data in scenarios such as
// proxying HTTP requests, handling WebSocket connections, or any other context where data needs to be passed
// between two endpoints. This function is typically used to enable communication between a client and a server
// or between two servers, especially after establishing a connection or upgrading a protocol (e.g., upgrading an
// HTTP connection to a WebSocket).

// Parameters:
//   - dst: The destination writer where data from the source will be written. It must implement the io.WriteCloser
//          interface to allow writing data and closing the stream once the transfer is complete.
//   - src: The source reader from which data will be read. It must implement the io.ReadCloser interface to
//          allow reading data and closing the stream upon completion.

// The function works by continuously reading data from the `src` stream and writing it to the `dst` stream until
// there's no more data to read or an error occurs. It uses the `io.Copy` function, which handles the read-write
// loop efficiently and copies data directly between streams without unnecessary buffering.

// After the data transfer is complete or if an error interrupts the transfer, `transfer` ensures proper cleanup
// by deferring the closure of both the `src` and `dst` streams. This deferred closure is crucial for releasing
// resources and avoiding leaks, especially in network programming and concurrent applications where multiple
// data transfers may occur simultaneously.

// Usage:
//   - This function is invoked in scenarios requiring the relay of data between two endpoints, such as forwarding
//     HTTP requests in a reverse proxy server or managing WebSocket connections. It is typically run in its
//     goroutine to allow simultaneous data flow in both directions (e.g., using two `transfer` calls in separate
//     goroutines for bidirectional communication).

// Note:
//   - While `transfer` abstracts away the complexities of copying data between streams, it's important to handle
//     errors and the lifecycle of connections outside this function. The caller should monitor for termination
//     conditions, such as network errors or signals indicating the end of communication, to gracefully close
//     the connections and terminate the data transfer.

func Transfer(dst io.WriteCloser, src io.ReadCloser) {
	defer func() {
		_ = dst.Close()
		_ = src.Close()
	}()
	_, _ = io.Copy(dst, src)
}
