package load_balancing_server_setup

import (
	"encoding/json"
	"errors"
	"jinx/internal/load_balancer"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/error_handler"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"net"
	"os"
	"path/filepath"
)

func LoadBalancingServerSetup(config types.LoadBalancerConfig, serverRootDir string) (types.JinxServer, *error_handler.JinxError) {

	//Create a directory for logs
	logRoot := filepath.Join(serverRootDir, string(constant.LOAD_BALANCER), constant.LOG_ROOT)
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

	ipAddress := net.ParseIP(config.IP)
	if ipAddress == nil {
		log.Printf("%s is an invalid ip address: using loopback address 127.0.0.1", config.IP)
		ipAddress = net.IP(constant.DEFAULT_IP)
	}

	algorithm := config.Algo
	if algorithm == "" {
		algorithm = constant.ROUND_ROBIN
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

	serverPoolConfigPath := config.ServerPoolConfigPath
	if serverPoolConfigPath == "" {
		log.Println("a server pool config file must be provided")
		return nil, error_handler.NewJinxError(constant.ERR_INVALID_SERVER_POOL_CONFIG, errors.New("no server pool config"))
	}

	if pathValidationErr := ValidateServerPoolConfigPath(serverPoolConfigPath); pathValidationErr != nil {
		log.Printf("server pool config validation error: %v", pathValidationErr)
		return nil, error_handler.NewJinxError(constant.ERR_INVALID_SERVER_POOL_CONFIG, pathValidationErr)
	}

	serverPool, err := LoadServerPoolConfig(serverPoolConfigPath)
	if err != nil {
		log.Printf("error occurred while reading server pool config: %v", err)
		return nil, error_handler.NewJinxError(constant.ERR_INVALID_SERVER_POOL_CONFIG, err)
	}

	jinxLoadBalancerConfig := types.JinxLoadBalancingServerConfig{
		IP:         string(ipAddress),
		Port:       port,
		LogRoot:    logRoot,
		CertFile:   certFile,
		KeyFile:    keyFile,
		ServerPool: serverPool,
		Algorithm:  algorithm,
	}

	jinx := load_balancer.NewJinxLoadBalancingServer(jinxLoadBalancerConfig, filepath.Join(constant.BASE, string(constant.LOAD_BALANCER)))
	return jinx, nil
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
