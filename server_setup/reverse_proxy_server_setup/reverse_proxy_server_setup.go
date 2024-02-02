// File: reverse_proxy_server_setup.go
// Package: reverse_proxy_server_setup

// Program Description:
// This file handles the setup of the reverse proxy server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 30, 2024

package reverse_proxy_server_setup

import (
	"encoding/json"
	"fmt"
	"jinx/internal/reverse_proxy"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"os"
	"path/filepath"
)

func ReverseProxyServerSetup(config types.ReverseProxyConfig) {

	//Create a directory for logs
	logRoot := filepath.Join(constant.BASE, string(constant.REVERSE_PROXY), constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Fatalf("unable to create log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
	}

	port := config.Port
	_, validationErr := helper.ValidatePort(port)
	if validationErr != nil {
		log.Fatalf(validationErr.Error())
	}

	ipAddress := config.IP
	if ipAddress == "" {
		ipAddress = constant.DEFAULT_IP
	}

	certFile := config.CertFile
	if certFile != "" {
		if _, certFileErr := os.Stat(certFile); certFileErr != nil {
			log.Fatalf("%s: %v", certFile, certFileErr)
		}
	}

	keyFile := config.KeyFile
	if keyFile != "" {
		if _, keyFileErr := os.Stat(keyFile); keyFileErr != nil {
			log.Fatalf("%s: %v", keyFile, keyFileErr)
		}
	}

	routeTablePath := config.RoutingTable
	if routeTablePath == "" {
		log.Fatalln("a route file must be provided")
	}

	if routeTableValidationErr := ValidateRouteTablePath(routeTablePath); routeTableValidationErr != nil {
		log.Fatalf("route table validation error: %v", routeTableValidationErr)
	}

	routeTable, err := LoadRouteTable(routeTablePath)
	if err != nil {
		log.Fatalf("error occurred while reading route table: %v", err)
	}

	jinxReversProxyConfig := types.JinxReverseProxyServerConfig{
		IP:         ipAddress,
		Port:       port,
		LogRoot:    logRoot,
		RouteTable: routeTable,
		CertFile:   certFile,
		KeyFile:    keyFile,
	}

	jinx := reverse_proxy.NewJinxReverseProxyServer(jinxReversProxyConfig, filepath.Join(constant.BASE, string(constant.REVERSE_PROXY)))
	jinx.Start()
	fmt.Println("Reverse Proxy Started...")
}

// ValidateRouteTablePath verifies the existence and format of the route table file specified by the path.
// This function is designed to ensure that the provided path points to a valid, accessible JSON file
// which is expected to contain routing information for a reverse proxy setup. The validation process
// involves checking the file's existence and verifying its extension to be '.json'.
//
// Parameters:
// - path: A string representing the filesystem path to the route table file. This file is expected
//         to be a JSON file containing routing configurations.
//
// Returns:
// - An error if the file at the given path does not exist, is not accessible, or does not have a '.json'
//   extension, indicating that it might not be a valid JSON file. If the file passes all checks, nil is
//   returned, indicating that the path is valid.
//
// The function performs the following checks:
// 1. File Existence: Utilizes os.Stat to determine if the file exists at the specified path. If the file
//    does not exist or is not accessible due to permission issues, an error is returned.
// 2. File Extension: Checks the file's extension to ensure it is '.json'. This is a basic validation step
//    to help ensure that the file format is as expected for JSON content. If the extension is not '.json',
//    an error indicating "invalid extension" is returned.
//
// This validation function is crucial for pre-validation steps in applications that rely on external
// configuration files, particularly when such configurations are critical for the application's routing
// logic or other functionalities.

func ValidateRouteTablePath(path string) error {

	if _, statErr := os.Stat(path); statErr != nil {
		return statErr
	}

	if pathExt := filepath.Ext(path); pathExt != ".json" {
		return os.ErrInvalid
	}

	return nil
}

// LoadRouteTable reads a JSON-formatted route table file from the specified path and decodes it into
// a RouteTable type. The route table is expected to be a JSON object with string keys and values,
// representing the mapping of request paths to addresses of upstream servers. This function is crucial
// for initializing routing configurations in applications that require dynamic request forwarding,
// such as reverse proxies.
//
// Parameters:
// - path: A string specifying the filesystem path to the JSON file containing the route table.
//
// Returns:
// - A populated RouteTable instance if the file is successfully read and decoded. The RouteTable
//   type is defined as a map[string]string, where keys are request paths and values are the
//   corresponding upstream server addresses.
// - An error if the file cannot be opened, read, or if the JSON decoding fails. This includes
//   scenarios where the file does not exist, is not accessible, or contains invalid JSON.
//
// Usage Example:
// Assuming a valid JSON file at "./config/routes.json", the function can be called as follows:
//   routeTable, err := LoadRouteTable("./config/routes.json")
//   if err != nil {
//       log.Fatalf("Failed to load route table: %v", err)
//   }
//   // Use routeTable for request routing...

func LoadRouteTable(path string) (types.RouteTable, error) {
	routeTable := make(types.RouteTable)

	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	if decodeErr := decoder.Decode(&routeTable); decodeErr != nil {
		return nil, decodeErr
	}

	return routeTable, nil
}
