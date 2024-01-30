// File: handle_command.go
// Package: main

// Program Description:
// This file handles the command line arguments and calls the
// specific server handler

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 30, 2024

package main

import (
	"flag"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"jinx/server_setup/forward_proxy_server_setup"
	"jinx/server_setup/ftp_server_setup"
	"jinx/server_setup/http_server_setup"
	"jinx/server_setup/load_balancing_server_setup"
	"jinx/server_setup/reverse_proxy_server_setup"
	"log"
)

var serverMode types.ServerMode
var ip string
var port string
var websiteRootDirectory string
var sslCertificateFile string
var sslPrivateKeyFile string
var routingRulesFilePath string
var hostnameBlacklist string
var serverPoolConfigPath string
var loadBalancingAlgorithm types.LoadBalancerAlgo

var jinxFlags *flag.FlagSet

func init() {
	jinxFlags = flag.NewFlagSet("Jinx Flags", flag.ExitOnError)

	jinxFlags.StringVar((*string)(&serverMode), constant.MODE, "", "Defines the server's operational mode (e.g., http, reverse_proxy, forward_proxy, load_balancer, ftp).")
	jinxFlags.StringVar(&ip, constant.IP, constant.DEFAULT_IP, "host ip address")
	jinxFlags.StringVar(&port, constant.PORT, constant.HTTP_PORT, "application port")
	jinxFlags.StringVar(&websiteRootDirectory, constant.WEBSITE_ROOT_DIR, "", "Sets the root directory path for hosting website files.")
	jinxFlags.StringVar(&sslCertificateFile, constant.CERT_FILE, "", "Specifies the file path to the SSL certificate used for HTTPS connections.")
	jinxFlags.StringVar(&sslPrivateKeyFile, constant.KEY_FILE, "", "Specifies the file path to the SSL certificate's private key.")
	jinxFlags.StringVar(&routingRulesFilePath, constant.ROUTE_TABLE, "", "Path to a JSON file defining routing rules for directing incoming requests to specific backend servers.")
	jinxFlags.StringVar(&hostnameBlacklist, constant.BLACK_LIST, "", "Comma-separated list of hostnames to block access to (e.g., google.com, facebook.com).")
	jinxFlags.StringVar(&serverPoolConfigPath, constant.SERVER_POOL_CONFIG, "", "Path to a JSON file listing backend servers and their configurations for load balancing.")
	jinxFlags.StringVar((*string)(&loadBalancingAlgorithm), constant.ALGO, string(constant.ROUND_ROBIN), "Specifies the load balancing algorithm to use (e.g., ROUND_ROBIN, LEAST_CONNECTIONS).")
}

// HandleStart is the entry point for configuring and starting different server modes based on the provided
// command-line arguments. This function parses the command-line arguments to configure server options such
// as IP address, port, SSL certificate, and key files. Depending on the specified server mode, it delegates
// to the appropriate setup function for initializing and starting the HTTP server, reverse proxy, forward proxy,
// load balancer, or FTP server.
//
// Parameters:
//   - args: A slice of strings representing the command-line arguments passed to the application. These arguments
//     are expected to include server configuration options like IP, port, and the server mode, among others.
//
// The function performs the following steps:
//  1. Parses the command-line arguments using a predefined flags parser to extract configuration options.
//  2. Constructs a map of options from the parsed arguments, including network settings and file paths for SSL
//     certificates, if applicable.
//  3. Determines the server mode based on the parsed arguments and invokes the corresponding setup function to
//     configure and start the server. Supported server modes are:
//     - HTTP server: Sets up and starts an HTTP server using the specified options.
//     - Reverse proxy: Initializes a reverse proxy server with routing rules defined in a configuration file.
//     - Forward proxy: Configures and starts a forward proxy server with an optional hostname blacklist.
//     - Load balancer: Initiates a load balancing server with a server pool configuration.
//     - FTP server: Sets up and starts an FTP server.
//  4. In case of an unrecognized server mode, logs a fatal error and terminates the application.
//
// Usage:
// The function is designed to be called with command-line arguments typically passed to the main function
// It abstracts away the complexity of server initialization and configuration, providing a
// simplified interface for starting servers in different modes based on runtime configuration.
func HandleStart(args []string) {

	if err := jinxFlags.Parse(args); err != nil {
		log.Fatal(err)
	}

	options := make(map[string]string)
	options[constant.IP] = ip
	options[constant.PORT] = port
	options[constant.CERT_FILE] = sslCertificateFile
	options[constant.KEY_FILE] = sslPrivateKeyFile

	switch serverMode {
	case constant.HTTP_SERVER:
		options[constant.WEBSITE_ROOT_DIR] = websiteRootDirectory
		http_server_setup.HTTPServerSetup(options)
		break
	case constant.REVERSE_PROXY:
		options[constant.ROUTE_TABLE] = routingRulesFilePath
		reverse_proxy_server_setup.ReverseProxyServerSetup(options)
		break
	case constant.FORWARD_PROXY:
		options[constant.BLACK_LIST] = hostnameBlacklist
		forward_proxy_server_setup.ForwardProxyServerSetup(options)
		break
	case constant.LOAD_BALANCER:
		options[constant.SERVER_POOL_CONFIG] = serverPoolConfigPath
		load_balancing_server_setup.LoadBalancingServerSetup(options)
		break
	case constant.FTP_SERVER:
		ftp_server_setup.FTPServerSetup(options)
		break
	default:
		log.Fatalf("%s is not a valid server mod option. valid option includes: http, reverse_proxy, forward_proxy, load_balancer, ftp", serverMode)
	}
}

func HandleStop() {

}

func HandleRestart() {

}

func HandleDestroy() {

}
