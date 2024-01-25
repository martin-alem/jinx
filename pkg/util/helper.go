package util

import (
	"io/fs"
	"net"
	"path/filepath"
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
//	exists, path := FileExist("/my/directory", "myfile.txt")
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
