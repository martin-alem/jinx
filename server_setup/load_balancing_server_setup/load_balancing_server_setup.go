package load_balancing_server_setup

import (
	"encoding/json"
	"jinx/internal/load_balancer"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func LoadBalancingServerSetup(options map[string]string) {

	//Create a directory for logs
	logRoot := filepath.Join(constant.BASE, constant.LOAD_BALANCER, constant.LOG_ROOT)
	if mkLogDirErr := os.MkdirAll(logRoot, 0755); !os.IsExist(mkLogDirErr) && mkLogDirErr != nil {
		log.Fatalf("unable to create log directory. make sure you have the right permissions in %s: %v", logRoot, mkLogDirErr)
	}

	port, portOk := options[constant.PORT]
	if !portOk {
		port = constant.HTTP_PORT
	}

	ipAddress, ipOk := options[constant.IP]
	if !ipOk {
		ipAddress = constant.DEFAULT_IP
	}

	certFile, certFileOk := options[constant.CERT_FILE]
	if certFileOk && certFile != "" {
		if _, certFileErr := os.Stat(certFile); certFileErr != nil {
			log.Fatalf("%s: %v", certFile, certFileErr)
		}
	}

	keyFile, keyFileOk := options[constant.KEY_FILE]
	if keyFileOk && keyFile != "" {
		if _, keyFileErr := os.Stat(keyFile); keyFileErr != nil {
			log.Fatalf("%s: %v", keyFile, keyFileErr)
		}
	}

	if certFileOk && certFile != "" && keyFileOk && keyFile != "" {
		if !portOk {
			port = constant.HTTPS_PORT
		}

	}

	serverPoolConfigPath, pathOk := options[constant.SERVER_POOL_CONFIG]
	if !pathOk {
		log.Fatalf("a server pool config file must be provided")
	}

	if validationErr := ValidateServerPoolConfigPath(serverPoolConfigPath); validationErr != nil {
		log.Fatalf("server pool config validation error: %v", validationErr)
	}

	serverPool, err := LoadServerPoolConfig(serverPoolConfigPath)
	if err != nil {
		log.Fatalf("error occurred while reading server pool config: %v", err)
	}

	portInt, err := strconv.ParseInt(port, 10, 0)
	if err != nil {
		log.Fatalf("%s is not a valid port:", port)
	}

	jinxLoadBalancerConfig := types.JinxLoadBalancingServerConfig{
		IP:             ipAddress,
		Port:           int(portInt),
		LogRoot:        logRoot,
		CertFile:       certFile,
		KeyFile:        keyFile,
		ServerPool:     serverPool,
		Algorithm:      constant.ROUND_ROBIN,
		SSLTermination: false,
	}

	jinx := load_balancer.NewJinxLoadBalancingServer(jinxLoadBalancerConfig, filepath.Join(constant.BASE, constant.LOAD_BALANCER))
	jinx.Start()

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
