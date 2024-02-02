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
	"encoding/json"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"jinx/server_setup/forward_proxy_server_setup"
	"jinx/server_setup/http_server_setup"
	"jinx/server_setup/load_balancing_server_setup"
	"jinx/server_setup/reverse_proxy_server_setup"
	"log"
	"os"
)

var configuration types.JinxServerConfiguration

func init() {
	configFile, openErr := os.Open(constant.CONFIG_FILE_PATH)
	if openErr != nil {
		log.Fatalf("unable to locate configuration file. please make sure %s exist in %s ", constant.CONFIG_FILE, constant.CONFIG_FILE_PATH)
	}

	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&configuration); err != nil {
		log.Fatalf("error occurred while reading configuration file: %v", err)
	}

}

func HandleStart() {

	switch configuration.Mode {
	case constant.HTTP_SERVER:
		http_server_setup.HTTPServerSetup(configuration.HttpServerConfig)
		break
	case constant.REVERSE_PROXY:
		reverse_proxy_server_setup.ReverseProxyServerSetup(configuration.ReverseProxyConfig)
		break
	case constant.FORWARD_PROXY:
		forward_proxy_server_setup.ForwardProxyServerSetup(configuration.ForwardProxyConfig)
		break
	case constant.LOAD_BALANCER:
		load_balancing_server_setup.LoadBalancingServerSetup(configuration.LoadBalancerConfig)
		break
	}
}

func HandleStop() {

}

func HandleRestart() {

}

func HandleDestroy() {

}
