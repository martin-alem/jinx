// File: jinx_load_balancing_server.go
// Package: load_balancer

// Program Description:
// This file implements a level 4 load balancer

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 31, 2024

package load_balancer

import (
	"crypto/tls"
	"fmt"
	"io"
	"jinx/internal/load_balancer/algo"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type JinxLoadBalancingServer struct {
	config        types.JinxLoadBalancingServerConfig
	errorLogger   *slog.Logger
	serverLogger  *slog.Logger
	serverRootDir string
	mode          string
	currentServer int
	mutex         *sync.Mutex
}

func NewJinxLoadBalancingServer(config types.JinxLoadBalancingServerConfig, serverRoot string) *JinxLoadBalancingServer {
	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr)
	}

	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr)
	}

	loadBalancerMode := "http"
	if config.CertFile != "" && config.KeyFile != "" {
		loadBalancerMode = "https"
	}

	return &JinxLoadBalancingServer{
		config:        config,
		errorLogger:   slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:  slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverRootDir: serverRoot,
		mode:          loadBalancerMode,
		currentServer: -1,
		mutex:         &sync.Mutex{},
	}
}

func (jx *JinxLoadBalancingServer) Start() {
	addr := fmt.Sprintf("%s:%d", jx.config.IP, jx.config.Port)
	var listener net.Listener

	if jx.mode == "https" {
		certificate, certErr := tls.LoadX509KeyPair(jx.config.CertFile, jx.config.KeyFile)
		if certErr != nil {
			msg := fmt.Sprintf("error loading certificate: %v", certErr)
			jx.errorLogger.Error(msg)
		}
		config := &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		l, listenerErr := tls.Listen("tcp", addr, config)
		if listenerErr != nil {
			msg := fmt.Sprintf("error starting https load balancer: %v", listenerErr)
			jx.errorLogger.Error(msg)
		}
		listener = l
	} else {
		l, listenerErr := net.Listen("tcp", addr)
		if listenerErr != nil {
			msg := fmt.Sprintf("error starting http load balancer: %v", listenerErr)
			jx.errorLogger.Error(msg)
		}
		listener = l
	}

	for {
		if listener != nil {
			conn, err := listener.Accept()
			if err != nil {
				msg := fmt.Sprintf("error accepting connection: %v", err)
				jx.errorLogger.Error(msg)
			}
			go jx.ProxyTCP(conn)
		}
	}
}

func (jx *JinxLoadBalancingServer) ProxyTCP(conn net.Conn) {
	upstreamServer := jx.PickAlgorithm()(jx.config.ServerPool, jx.currentServer, jx.mutex)
	addr := fmt.Sprintf("%s:%d", upstreamServer.IP, upstreamServer.Port)

	remoteConn, err := net.Dial("tcp", addr)
	if err != nil {
		jx.errorLogger.Error(fmt.Sprintf("error connecting to remote: %v", err))
		_ = conn.Close() // Only close conn here as remoteConn is not yet established.
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)

	// Client to Remote
	go func() {
		defer wg.Done()
		_, copyErr := io.Copy(remoteConn, conn)
		if copyErr != nil {
			jx.errorLogger.Error(fmt.Sprintf("copying from client to remote failed: %v", copyErr))
		}
	}()

	// Remote to Client
	go func() {
		defer wg.Done()
		_, copyErr := io.Copy(conn, remoteConn)
		if copyErr != nil {
			jx.errorLogger.Error(fmt.Sprintf("copying from remote to client failed: %v", copyErr))
		}
	}()

	wg.Wait()
	_ = remoteConn.Close() // Close remote connection after data transfer is complete.
	_ = conn.Close()
}

func (jx *JinxLoadBalancingServer) PickAlgorithm() types.LoadBalancingAlgorithm {
	switch jx.config.Algorithm {
	case constant.ROUND_ROBIN:
		return algo.RoundRobin
	case constant.LEAST_CONNECTIONS:
		return algo.LeastConnection
	case constant.LEAST_RESPONSE_TIME:
		return algo.LeastResponse
	case constant.HASHING:
		return algo.Hash
	case constant.WEIGHTED_ROUND_ROBIN:
		return algo.WeightedRoundRobin
	case constant.WEIGHTED_LEAST_CONNECTIONS:
		return algo.WeightedLeastConnection
	case constant.WEIGHTED_LEAST_RESPONSE_TIME:
		return algo.WeightedLeastResponse
	case constant.RESOURCE_BASED:
		return algo.ResourceBased
	case constant.GEOGRAPHICAL:
		return algo.Geographical
	default:
		return algo.RoundRobin
	}

}
