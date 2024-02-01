// File: http_server_setup.go
// Package: http_server_setup

// Program Description:
// This file handles the setup of the http server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 30, 2024

package http_server_setup

import (
	"io"
	"jinx/internal/jinx_http"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

func HTTPServerSetup(options map[string]string) {

	webRootDir, webRootDirOk := options[constant.WEBSITE_ROOT_DIR]
	if !webRootDirOk || webRootDir == "" {
		webRootDir = constant.DEFAULT_WEBSITE_ROOT_DIR
	} else {
		if readable, readableErr := helper.IsDirReadable(webRootDir); !readable {
			log.Fatalf("unable to read website directory or does not exit: %s: %v", webRootDir, readableErr)
		}
	}

	port, portOk := options[constant.PORT]
	if !portOk {
		port = constant.HTTP_PORT
	}

	ipAddress, ipOk := options[constant.IP]
	if !ipOk {
		ipAddress = constant.DEFAULT_IP
	}

	certFile, certFileOk := options[constant.CERT_FILE]
	if certFileOk && certFile != "" {
		if _, certFileErr := os.Stat(certFile); certFileErr != nil {
			log.Fatalf("%s: %v", certFile, certFileErr)
		}
	}

	keyFile, keyFileOk := options[constant.KEY_FILE]
	if keyFileOk && keyFile != "" {
		if _, keyFileErr := os.Stat(keyFile); keyFileErr != nil {
			log.Fatalf("%s: %v", keyFile, keyFileErr)
		}
	}

	if certFileOk && certFile != "" && keyFileOk && keyFile != "" {
		if !portOk {
			port = constant.HTTPS_PORT
		}
	}

	//Create a directory for logs
	logRoot := filepath.Join(constant.BASE, constant.HTTP_SERVER, constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Fatalf("unable to create log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
	}

	//Create a directory to store default website files
	defaultWebsiteRoot := filepath.Join(constant.BASE, constant.HTTP_SERVER, constant.DEFAULT_WEBSITE_ROOT)
	if mkdirErr := os.MkdirAll(defaultWebsiteRoot, 0755); !os.IsExist(mkdirErr) && mkdirErr != nil {
		log.Fatalf("unable to create default website root. make sure you have the right permissions in %s: %v", defaultWebsiteRoot, mkdirErr)
	}

	imagesDir := filepath.Join(defaultWebsiteRoot, constant.IMAGE_DIR)
	if mkdirErr := os.Mkdir(imagesDir, 0755); !os.IsExist(mkdirErr) && mkdirErr != nil {
		log.Fatalf("unable to create default website image dir. make sure you have the right permissions in %s: %v", imagesDir, mkdirErr)
	}

	resources := map[string]string{
		constant.JINX_INDEX_URL: constant.JINX_INDEX_FILE,
		constant.JINX_404_URL:   constant.JINX_404_FILE,
		constant.JINX_CSS_URL:   constant.JINX_CSS_FILE,
		constant.JINX_ICO_URL:   constant.JINX_ICO_FLE,
		constant.JINX_SVG_URL:   constant.JINX_SVG_FILE,
	}

	FetchServerDefaultWebsite(resources, defaultWebsiteRoot, imagesDir)

	portInt, err := strconv.ParseInt(port, 10, 0)
	if err != nil {
		log.Fatalf("%s is not a valid port:", port)
	}

	jinxHttpConfig := types.JinxHttpServerConfig{
		IP:          ipAddress,
		Port:        int(portInt),
		LogRoot:     logRoot,
		WebsiteRoot: webRootDir,
		CertFile:    certFile,
		KeyFile:     keyFile,
	}

	jinx := jinx_http.NewJinxHttpServer(jinxHttpConfig, filepath.Join(constant.BASE, constant.HTTP_SERVER))
	jinx.Start()
}

// FetchServerDefaultWebsite concurrently downloads a set of predefined resources from the web
// and saves them to the local filesystem, organizing them into the specified directories for the
// server's default website content and images. This function is designed to prepare the server's
// static content by ensuring that essential files like HTML, CSS, icons, and other resources are
// available in the server's web root directory before it starts serving requests.
//
// Parameters:
//   - resources: A map where each key-value pair consists of a URL (key) from which to fetch a resource
//     and a filename (value) under which to save the resource locally. This map defines the
//     resources necessary for the server's default website.
//   - defaultWebsiteRoot: The path to the directory where most resources should be saved. This directory
//     serves as the root for the server's default website content.
//   - imagesDir: The path to the directory within defaultWebsiteRoot dedicated to storing image files.
//     Certain resources, specifically designated as image files (e.g., icons, SVGs), are saved here.
//
// The function operates as follows:
//  1. Initiates concurrent download operations for each resource specified in the resources map using goroutines.
//     Each operation attempts to fetch the resource from its URL and passes the result (or error) through a
//     channel.
//  2. Waits for all download operations to complete, using a sync.WaitGroup to track completion.
//  3. Processes each received resource from the channel, saving it to the appropriate location on the filesystem
//     based on its type. Non-image resources are saved directly under defaultWebsiteRoot, while image resources
//     are saved in imagesDir.
//  4. Closes the response body of each HTTP request to free up network resources.
//
// Errors encountered during resource fetching are logged, but do not halt the execution of other download
// operations. This function ensures that the server has access to all necessary static files for its default
// website, facilitating a quick setup and start-up process.
//
// Note: This function is critical for server initialization, especially in scenarios where the server is
// expected to provide a default set of web pages or images immediately upon start-up.
func FetchServerDefaultWebsite(resources map[string]string, defaultWebsiteRoot string, imagesDir string) {
	var wg sync.WaitGroup
	wg.Add(len(resources)) // Add count to WaitGroup before starting goroutines

	resourceChan := make(chan types.JinxResourceResponse, len(resources))

	for url, file := range resources {
		go func(resourceURL string, fileName string) {
			defer wg.Done() // Ensure wg.Done() is called when goroutine finishes
			res, err := http.Get(resourceURL)
			if err != nil {
				log.Printf("unable to fetch resource from URL %s: %v", resourceURL, err)
				return
			}
			resourceChan <- types.JinxResourceResponse{Res: res, Filename: fileName}
		}(url, file)
	}

	// Close resourceChan after all goroutines have finished
	go func() {
		wg.Wait()
		close(resourceChan)
	}()

	// Process received resources
	for data := range resourceChan {
		HandleResourceResponse(defaultWebsiteRoot, imagesDir, &data)
		_ = data.Res.Body.Close()
	}
}

// HandleResourceResponse processes and saves a single resource fetched from the web to the local filesystem.
// It is typically used to store static assets like CSS, JavaScript, images, or HTML files required for
// the server's web content. The function determines the correct file path based on the resource type
// and writes the content to the file system.
//
// Parameters:
// - websiteRoot: The root directory for the website's static content. It serves as the base path for saving
//                most resources except for specific types like icons or SVG files, which are saved in a
//                designated subdirectory.
// - imageDir:    The directory within websiteRoot dedicated to storing image files. Icons and SVG files
//                fetched as resources are saved here.
// - resource:    A pointer to a JinxResourceResponse, which includes the HTTP response containing the
//                resource's content and the filename under which the resource should be saved.
//
// The function performs the following actions:
// 1. Reads the content from the HTTP response body.
// 2. Determines the appropriate file path based on the filename. Icons and SVG files are saved in
//    the image directory, while other resources are saved directly under the website root.
// 3. Creates or overwrites the file at the determined path with the resource's content.
//
// If the function encounters errors at any step, such as reading the response body or writing to the
// file system, it logs the error and terminates the application. This ensures that the server setup
// process does not proceed with missing or incomplete web content.

func HandleResourceResponse(websiteRoot string, imageDir string, resource *types.JinxResourceResponse) {

	fileContent, err := io.ReadAll(resource.Res.Body)
	if err != nil {
		log.Fatalf("unable to read response for: %v", resource.Filename)
	}

	filePath := filepath.Join(websiteRoot, resource.Filename)

	if resource.Filename == constant.JINX_ICO_FLE || resource.Filename == constant.JINX_SVG_FILE {
		filePath = filepath.Join(imageDir, resource.Filename)
	}

	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	defer func() {
		_ = fileHandle.Close()
	}()
	if err != nil && !os.IsExist(err) {
		log.Fatalf("unable to open file %s: %v", filePath, err)
	}

	if _, writeErr := fileHandle.Write(fileContent); writeErr != nil {
		log.Fatalf("error writing to %s: %v", filePath, writeErr)
	}
}
