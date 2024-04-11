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
	"errors"
	"jinx/internal/reverse_proxy"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/error_handler"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"net"
	"os"
	"path/filepath"
)

// ReverseProxyServerSetup initializes and configures a reverse proxy server based on the provided
// configuration and server root directory. This setup process includes creating a directory for logs,
// validating the specified port, IP address, and paths for SSL certificate and key files, and loading
// the routing table from a specified file. It ensures that all necessary preconditions for running the
// reverse proxy server are met, including permission checks for creating directories and file existence
// checks for critical configuration files.
//
// Parameters:
//   - config: A types.ReverseProxyConfig object containing configuration options for the reverse proxy server,
//     including the port, IP address, paths to SSL certificate and key files, and the path to the routing table file.
//   - serverRootDir: The root directory path for the server where logs and other server-related files will be stored.
//
// The method performs the following key actions:
//   1. Creates a log directory within the server root directory to store log files.
//   2. Validates the specified port to ensure it is within the acceptable range and format.
//   3. Validates the IP address, defaulting to the loopback address if the specified IP is invalid.
//   4. Checks for the existence of the SSL certificate and key files if specified, ensuring secure connections can be established.
//   5. Validates the presence and format of the routing table file, which is crucial for determining the forwarding rules for incoming requests.
//   6. Loads the routing table from the specified file, parsing it to set up forwarding rules for the reverse proxy server.
//
// If any step in the setup process fails (e.g., due to invalid configuration options, permission issues, or missing files),
// the function returns nil and an error detailing the reason for the failure. This ensures that the server does not start
// in an improperly configured state, minimizing potential runtime errors and security risks.
//
// Returns:
//   - A types.JinxServer instance configured as a reverse proxy server, ready to start serving requests based on the routing table.
//   - A pointer to an error_handler.JinxError if an error occurs during the setup process. If setup is successful, nil is returned.
//
// Usage:
//   - This method is intended to be called during the initial setup phase of the reverse proxy server, typically at application
//     startup. It provides a streamlined process for preparing the server environment, loading configuration settings, and ensuring
//     that the server is ready to handle requests according to the defined forwarding rules.

func ReverseProxyServerSetup(config types.ReverseProxyConfig, serverRootDir string) (types.JinxServer, *error_handler.JinxError) {

	//Create a directory for logs
	logRoot := filepath.Join(serverRootDir, string(constant.REVERSE_PROXY), constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Printf("unable to create log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
		return nil, error_handler.NewJinxError(constant.ERR_CREATE_DIR, mkLogDirErr)
	}

	port := config.Port
	_, validationErr := helper.ValidatePort(port)
	if validationErr != nil {
		log.Printf(validationErr.Error())
		return nil, error_handler.NewJinxError(constant.INVALID_PORT, validationErr)
	}

	ipAddress := net.IP(config.IP)
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

	routeTablePath := config.RoutingTable
	if routeTablePath == "" {
		log.Println("a route file must be provided")
		return nil, error_handler.NewJinxError(constant.ERR_INVALID_ROUTE_TABLE, errors.New("no route table"))
	}

	if routeTableValidationErr := ValidateRouteTablePath(routeTablePath); routeTableValidationErr != nil {
		log.Printf("route table validation error: %v", routeTableValidationErr)
		return nil, error_handler.NewJinxError(constant.ERR_INVALID_ROUTE_TABLE, validationErr)
	}

	routeTable, err := LoadRouteTable(routeTablePath)
	if err != nil {
		log.Printf("error occurred while reading route table: %v", err)
		return nil, error_handler.NewJinxError(constant.ERR_INVALID_ROUTE_TABLE, err)
	}

	jinxReversProxyConfig := types.JinxReverseProxyServerConfig{
		IP:         string(ipAddress),
		Port:       port,
		LogRoot:    logRoot,
		RouteTable: routeTable,
		CertFile:   certFile,
		KeyFile:    keyFile,
	}

	jinx := reverse_proxy.NewJinxReverseProxyServer(jinxReversProxyConfig, filepath.Join(serverRootDir, string(constant.REVERSE_PROXY)))
	return jinx, nil
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
