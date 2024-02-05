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
	"errors"
	"jinx/internal/forward_proxy"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/error_handler"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"net"
	"os"
	"path/filepath"
)

func ForwardProxyServerSetup(config types.ForwardProxyConfig, serverRootDir string) (types.JinxServer, *error_handler.JinxError) {

	//Create a directory for logs
	logRoot := filepath.Join(serverRootDir, string(constant.FORWARD_PROXY), constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Printf("unable to create a log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
		return nil, error_handler.NewJinxError(constant.ERR_CREATE_DIR, mkLogDirErr)
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

	var blackList []string
	var blackListErr error

	blackListPath := config.BlackList
	if blackListPath != "" {
		if blackListValidationPathErr := ValidateBlackListPath(blackListPath); blackListValidationPathErr != nil {
			log.Printf("error while parsing blacklist file: %v", blackListValidationPathErr)
			return nil, error_handler.NewJinxError(constant.ERR_INVALID_BLACK_LIST, errors.New("no black list table"))
		}

		blackList, blackListErr = LoadBlackList(blackListPath)
		if blackListErr != nil {
			log.Printf("error while loading the black list: %v", blackListErr)
			return nil, error_handler.NewJinxError(constant.ERR_INVALID_BLACK_LIST, blackListErr)
		}
	}

	jinxForwardProxyConfig := types.JinxForwardProxyServerConfig{
		IP:        string(ipAddress),
		Port:      port,
		LogRoot:   logRoot,
		BlackList: blackList,
		CertFile:  certFile,
		KeyFile:   keyFile,
	}

	jinx := forward_proxy.NewJinxForwardProxyServer(jinxForwardProxyConfig, filepath.Join(serverRootDir, string(constant.FORWARD_PROXY)))
	return jinx, nil
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
