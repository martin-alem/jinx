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
	"jinx/internal/forward_proxy"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func ForwardProxyServerSetup(options map[string]string) {

	//Create a directory for logs
	logRoot := filepath.Join(constant.BASE, constant.FORWARD_PROXY, constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Fatalf("unable to create a log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
	}

	port, portOk := options[constant.PORT]
	if !portOk {
		port = constant.HTTP_PORT
	}

	ipAddress, ipOk := options[constant.IP]
	if !ipOk {
		ipAddress = constant.DEFAULT_IP
	}

	var blackList []string
	var blackListErr error

	blackListPath, pathOk := options[constant.BLACK_LIST]
	if pathOk {
		if validationErr := ValidateBlackListPath(blackListPath); validationErr != nil {
			log.Fatalf("error while parsing blacklist file: %v", validationErr)
		}

		blackList, blackListErr = LoadBlackList(blackListPath)
		if blackListErr != nil {
			log.Fatalf("error while loading the black list: %v", blackListErr)
		}
	}

	portInt, err := strconv.ParseInt(port, 10, 0)
	if err != nil {
		log.Fatalf("%s is not a valid port:", port)
	}

	jinxForwardProxyConfig := types.JinxForwardProxyServerConfig{
		IP:        ipAddress,
		Port:      int(portInt),
		LogRoot:   logRoot,
		BlackList: blackList,
	}

	jinx := forward_proxy.NewJinxForwardProxyServer(jinxForwardProxyConfig, filepath.Join(constant.BASE, constant.FORWARD_PROXY))
	jinx.Start()
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
