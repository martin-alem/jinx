// File: forward_proxy_server_setup.go
// Package: forward_proxy_server_setup

// Program Description:
// This file handles the setup of the forward proxy server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 31, 2024

package forward_proxy_server_setup

import (
	"bufio"
	"fmt"
	"jinx/internal/forward_proxy"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"os"
	"path/filepath"
)

func ForwardProxyServerSetup(config types.ForwardProxyConfig) {

	//Create a directory for logs
	logRoot := filepath.Join(constant.BASE, string(constant.FORWARD_PROXY), constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Fatalf("unable to create a log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
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

	var blackList []string
	var blackListErr error

	blackListPath := config.BlackList
	if blackListPath == "" {
		if blackListValidationPathErr := ValidateBlackListPath(blackListPath); blackListValidationPathErr != nil {
			log.Fatalf("error while parsing blacklist file: %v", blackListValidationPathErr)
		}

		blackList, blackListErr = LoadBlackList(blackListPath)
		if blackListErr != nil {
			log.Fatalf("error while loading the black list: %v", blackListErr)
		}
	}

	jinxForwardProxyConfig := types.JinxForwardProxyServerConfig{
		IP:        ipAddress,
		Port:      port,
		LogRoot:   logRoot,
		BlackList: blackList,
		CertFile:  certFile,
		KeyFile:   keyFile,
	}

	jinx := forward_proxy.NewJinxForwardProxyServer(jinxForwardProxyConfig, filepath.Join(constant.BASE, string(constant.FORWARD_PROXY)))
	jinx.Start()
	fmt.Println("Forward Proxy Started...")
}

func ValidateBlackListPath(path string) error {
	if _, statErr := os.Stat(path); statErr != nil {
		return statErr
	}

	if pathExt := filepath.Ext(path); pathExt != ".txt" {
		return os.ErrInvalid
	}

	return nil
}

func LoadBlackList(path string) ([]string, error) {
	blackList := make([]string, 0)

	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		blackList = append(blackList, line)
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return blackList, nil
}
