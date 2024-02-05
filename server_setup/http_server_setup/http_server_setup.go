// File: http_server_setup.go
// Package: http_server_setup

// Program Description:
// This file handles the setup of the http/https server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 30, 2024

package http_server_setup

import (
	"io"
	"jinx/internal/jinx_http"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/error_handler"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// HTTPServerSetup initializes and configures an HTTP server based on the provided configuration and server root directory.
// It performs several setup tasks including validating the server configuration, ensuring necessary directories (e.g., for logs,
// website files, and images) are created, and fetching initial resources. The function checks for the validity of the website
// root directory, port, IP address, SSL certificate, and key paths, creating necessary directories as needed. It also attempts
// to fetch predefined resources such as the index page, error page, CSS, and image files. If any setup step fails, the function
// returns an error, preventing the server from starting. This comprehensive setup process is designed to ensure that the server
// is correctly and securely configured before it begins serving requests. If all setup tasks complete successfully, the function
// returns a configured JinxServer instance ready to start serving requests, along with nil error. Otherwise, it returns a
// non-nil error detailing what went wrong during setup.
//
// Parameters:
// - config: A types.HttpServerConfig object containing configuration options for the server, such as port, IP address, and paths to SSL certificate and key files.
// - serverRootDir: The root directory path for the server where logs and default website files will be stored.
//
// Returns:
// - A types.JinxServer instance, which is a custom server type that encapsulates the configured HTTP server.
// - A pointer to an error_handler.JinxError if an error occurs during the setup process. If setup is successful, nil is returned.
func HTTPServerSetup(config types.HttpServerConfig, serverRootDir string) (types.JinxServer, *error_handler.JinxError) {

	webRootDir := config.WebsiteRootDir
	if webRootDir == "" {
		webRootDir = string(constant.DEFAULT_WEBSITE_ROOT_DIR)
	} else {
		if readable, readableErr := helper.IsDirReadable(webRootDir); !readable {
			log.Printf("unable to read website directory or does not exit: %s: %v", webRootDir, readableErr)
			return nil, error_handler.NewJinxError(constant.INVALID_WEBSITE_DIR, readableErr)
		}
	}

	port := config.Port
	_, validationErr := helper.ValidatePort(port)
	if validationErr != nil {
		log.Printf(validationErr.Error())
		return nil, error_handler.NewJinxError(constant.INVALID_PORT, validationErr)
	}

	ipAddress := net.ParseIP(config.IP)
	if ipAddress == nil {
		log.Printf("%s is an invalid ip address: using loopback address 127.0.0.1", config.IP)
		ipAddress = net.IP(constant.DEFAULT_IP)
	}

	certFile := config.CertFile
	if certFile != "" {
		if _, certFileErr := os.Stat(certFile); certFileErr != nil {
			log.Printf("%s: %v", certFile, certFileErr)
			return nil, error_handler.NewJinxError(constant.INVALID_CERT_PATH, certFileErr)
		}
	}

	keyFile := config.KeyFile
	if keyFile != "" {
		if _, keyFileErr := os.Stat(keyFile); keyFileErr != nil {
			log.Printf("%s: %v", keyFile, keyFileErr)
			return nil, error_handler.NewJinxError(constant.INVALID_KEY_PATH, keyFileErr)
		}
	}

	//Create a directory for logs
	logRoot := filepath.Join(serverRootDir, string(constant.HTTP_SERVER), constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Printf("unable to create log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
		return nil, error_handler.NewJinxError(constant.ERR_CREATE_DIR, mkLogDirErr)
	}

	//Create a directory to store default website files
	defaultWebsiteRoot := filepath.Join(serverRootDir, string(constant.HTTP_SERVER), constant.DEFAULT_WEBSITE_ROOT)
	if mkdirErr := os.MkdirAll(defaultWebsiteRoot, 0755); !os.IsExist(mkdirErr) && mkdirErr != nil {
		log.Printf("unable to create default website root. make sure you have the right permissions in %s: %v", defaultWebsiteRoot, mkdirErr)
		return nil, error_handler.NewJinxError(constant.ERR_CREATE_DIR, mkdirErr)
	}

	imagesDir := filepath.Join(defaultWebsiteRoot, constant.IMAGE_DIR)
	if mkdirErr := os.Mkdir(imagesDir, 0755); !os.IsExist(mkdirErr) && mkdirErr != nil {
		log.Printf("unable to create default website image dir. make sure you have the right permissions in %s: %v", imagesDir, mkdirErr)
		return nil, error_handler.NewJinxError(constant.ERR_CREATE_DIR, mkdirErr)
	}

	resources := map[string]string{
		constant.JINX_INDEX_URL: filepath.Join(defaultWebsiteRoot, constant.INDEX_FILE),
		constant.JINX_404_URL:   filepath.Join(defaultWebsiteRoot, constant.JINX_404_FILE),
		constant.JINX_CSS_URL:   filepath.Join(defaultWebsiteRoot, constant.JINX_CSS_FILE),
		constant.JINX_ICO_URL:   filepath.Join(imagesDir, constant.JINX_ICO_FILE),
		constant.JINX_SVG_URL:   filepath.Join(imagesDir, constant.JINX_SVG_FILE),
	}

	anyErrors := HandleFetchResources(resources)

	if len(anyErrors) >= 1 {
		//Initiate clean up process and terminate
		_ = os.RemoveAll(serverRootDir)
		log.Printf("errors occured while trying to fetch some resources. Terminating server start up.")
		return nil, nil
	}

	jinxHttpConfig := types.JinxHttpServerConfig{
		IP:          string(ipAddress),
		Port:        port,
		LogRoot:     logRoot,
		WebsiteRoot: webRootDir,
		CertFile:    certFile,
		KeyFile:     keyFile,
	}

	jinx := jinx_http.NewJinxHttpServer(jinxHttpConfig, serverRootDir)
	return jinx, nil
}

// HandleFetchResources concurrently fetches multiple resources specified by the `resources` map, where each key-value
// pair represents a URL and its corresponding file path to store the fetched content. This function orchestrates the
// process of sending HTTP requests to each URL, receiving responses, and writing the response bodies to their
// respective files. It leverages goroutines and channels to perform these operations in parallel, significantly
// improving efficiency when dealing with multiple resources. Error handling is a critical aspect of this function; it
// captures errors from each operation (e.g., network errors, file system errors) and aggregates them into a slice.
// This error aggregation allows the caller to inspect and handle errors after all operations have completed. The use
// of a `sync.WaitGroup` ensures that the function waits for all fetch and write operations to finish before returning
// the collected errors.
//
// Parameters:
//   - resources: A map where each key is a URL of a resource to fetch, and each value is the file path where the
//     resource's content should be saved.
//
// Returns:
//   - A slice of errors encountered during the fetching and writing operations. If all operations succeed, this slice
//     will be empty. Each error in the slice is indicative of a failure in fetching from a URL or writing to a file,
//     corresponding to one of the entries in the `resources` map.
func HandleFetchResources(resources map[string]string) []error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(resources))

	wg.Add(len(resources))

	for url, filePath := range resources {
		go func(url, filePath string) {
			defer wg.Done()

			fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			defer func() {
				if fileHandle != nil {
					_ = fileHandle.Close()
				}
			}()
			if err != nil {
				log.Printf("error opening file: %s:%v", filePath, err)
				errChan <- err
				return
			}

			resp, respErr := FetchResource(url)
			if respErr != nil {
				log.Printf("error fetching from %s:%v", url, respErr)
				errChan <- respErr
				return
			}
			writeErr := WriteResponseToFile(fileHandle.Name(), resp)
			if writeErr != nil {
				log.Printf("error writing response to file %v", writeErr)
				errChan <- writeErr
				return
			}
		}(url, filePath)
	}

	// Close errChan after all goroutines are done
	go func() {
		wg.Wait()
		close(errChan)
	}()

	anyErrors := make([]error, 0)
	for err := range errChan {
		anyErrors = append(anyErrors, err)
	}

	return anyErrors
}

// FetchResource sends an HTTP GET request to the specified URL and returns the response received. It is designed to
// facilitate the retrieval of resources from the web, encapsulating the network request logic and error handling into
// a simple, reusable function. If the function encounters an error while attempting to fetch the resource, such as
// network issues or an invalid URL, it logs the error and returns a JinxError that encapsulates the error details,
// providing a unified error handling mechanism across the application. This function is particularly useful for
// applications that need to fetch and process external resources, offering a straightforward way to initiate HTTP
// requests and handle potential errors in a consistent manner.
//
// Parameters:
// - resource: The URL of the resource to be fetched.
//
// Returns:
//   - A pointer to a http.Response if the request is successful, allowing the caller to access the response body,
//     headers, and other metadata.
//   - A pointer to an error_handler.JinxError if the function encounters an error while fetching the resource, containing
//     details about the failure. If no error occurs, this will be nil.
func FetchResource(resource string) (*http.Response, *error_handler.JinxError) {
	res, err := http.Get(resource)
	if err != nil {
		log.Printf("unable to fetch resource from URL %s: %v", resource, err)
		return nil, error_handler.NewJinxError(constant.FETCH_RESOURCE_ERR, err)
	}

	return res, nil
}

// WriteResponseToFile takes a file path and an HTTP response, writes the response body to the specified file,
// and handles any errors that occur during the process. It attempts to read the entire response body, open (or create if
// it doesn't exist) the file at the given path, and write the response body to it. If any step fails, a JinxError is
// returned detailing the nature of the error, such as issues reading the response body, opening the file, or writing to
// the file. The function ensures the file is properly closed before exiting. This function is useful for persisting
// HTTP response data to the filesystem, enabling offline access or caching of resources.
//
// Parameters:
// - file: The path to the file where the HTTP response body should be written.
// - resource: The HTTP response whose body is to be written to the file.
//
// Returns:
//   - A pointer to an error_handler.JinxError if an error occurs during the process; otherwise, nil if the operation
//     is successful.
func WriteResponseToFile(file string, resource *http.Response) *error_handler.JinxError {

	fileContent, err := io.ReadAll(resource.Body)
	if err != nil {
		log.Printf("unable to read response for: %v", err)
		return error_handler.NewJinxError(constant.READ_RESPONSE_ERR, err)
	}

	filePath := filepath.Join(file)

	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	defer func() {
		_ = fileHandle.Close()
	}()
	if err != nil && !os.IsExist(err) {
		log.Printf("unable to open file %s: %v", filePath, err)
		return error_handler.NewJinxError(constant.OPEN_FILE_ERR, err)
	}

	if _, writeErr := fileHandle.Write(fileContent); writeErr != nil {
		log.Printf("error writing to %s: %v", filePath, writeErr)
		return error_handler.NewJinxError(constant.WRITE_FILE_ERR, writeErr)
	}

	return nil
}
