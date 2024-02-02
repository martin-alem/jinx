package load_balancing_server_setup

import (
	"encoding/json"
	"fmt"
	"jinx/internal/load_balancer"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"os"
	"path/filepath"
)

func LoadBalancingServerSetup(config types.LoadBalancerConfig) {

	//Create a directory for logs
	logRoot := filepath.Join(constant.BASE, string(constant.LOAD_BALANCER), constant.LOG_ROOT)
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

	algorithm := config.Algo
	if algorithm == "" {
		algorithm = constant.ROUND_ROBIN
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

	serverPoolConfigPath := config.ServerPoolConfigPath
	if serverPoolConfigPath == "" {
		log.Fatalf("a server pool config file must be provided")
	}

	if pathValidationErr := ValidateServerPoolConfigPath(serverPoolConfigPath); pathValidationErr != nil {
		log.Fatalf("server pool config validation error: %v", pathValidationErr)
	}

	serverPool, err := LoadServerPoolConfig(serverPoolConfigPath)
	if err != nil {
		log.Fatalf("error occurred while reading server pool config: %v", err)
	}

	jinxLoadBalancerConfig := types.JinxLoadBalancingServerConfig{
		IP:         ipAddress,
		Port:       port,
		LogRoot:    logRoot,
		CertFile:   certFile,
		KeyFile:    keyFile,
		ServerPool: serverPool,
		Algorithm:  algorithm,
	}

	jinx := load_balancer.NewJinxLoadBalancingServer(jinxLoadBalancerConfig, filepath.Join(constant.BASE, string(constant.LOAD_BALANCER)))
	jinx.Start()
	fmt.Println("Load Balancer Started...")
}

func ValidateServerPoolConfigPath(path string) error {

	if _, statErr := os.Stat(path); statErr != nil {
		return statErr
	}

	if pathExt := filepath.Ext(path); pathExt != ".json" {
		return os.ErrInvalid
	}

	return nil
}

func LoadServerPoolConfig(path string) ([]types.UpStreamServer, error) {
	serverPoolConfig := make(types.ServerPoolConfig)
	serverPool := make([]types.UpStreamServer, 0)

	file, err := os.Open(path)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	if decodeErr := decoder.Decode(&serverPoolConfig); decodeErr != nil {
		return nil, decodeErr
	}

	for _, val := range serverPoolConfig {
		serverPool = append(serverPool, val)
	}

	return serverPool, nil
}
