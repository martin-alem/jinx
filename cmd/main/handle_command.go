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
	"path/filepath"
)

var configuration types.JinxServerConfiguration
var server types.JinxServer

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
		httpServerWorkingDir := filepath.Join(constant.BASE, string(constant.HTTP_SERVER))
		jinx, serverErr := http_server_setup.HTTPServerSetup(configuration.HttpServerConfig, httpServerWorkingDir)
		if serverErr != nil {
			log.Fatal(serverErr)
		}
		server = jinx.Start()
		break
	case constant.REVERSE_PROXY:
		reverseProxyWorkingDir := filepath.Join(constant.BASE, string(constant.REVERSE_PROXY))
		jinx, serverErr := reverse_proxy_server_setup.ReverseProxyServerSetup(configuration.ReverseProxyConfig, reverseProxyWorkingDir)
		if serverErr != nil {
			log.Fatal(serverErr)
		}
		server = jinx.Start()
		break
	case constant.FORWARD_PROXY:
		forwardProxyWorkingDir := filepath.Join(constant.BASE, string(constant.FORWARD_PROXY))
		jinx, serverErr := forward_proxy_server_setup.ForwardProxyServerSetup(configuration.ForwardProxyConfig, forwardProxyWorkingDir)
		if serverErr != nil {
			log.Fatal(serverErr)
		}
		server = jinx.Start()
		break
	case constant.LOAD_BALANCER:
		loadBalancerWorkingDir := filepath.Join(constant.BASE, string(constant.LOAD_BALANCER))
		jinx, serverErr := load_balancing_server_setup.LoadBalancingServerSetup(configuration.LoadBalancerConfig, loadBalancerWorkingDir)
		if serverErr != nil {
			log.Fatal(serverErr)
		}
		server = jinx.Start()
		break
	}
}

func HandleStop() {
	server.Stop()
}

func HandleRestart() {
	server.Restart()
}

func HandleDestroy() {
	server.Destroy()
}
