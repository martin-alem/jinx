// File: jinx_load_balancing_server.go
// Package: load_balancer

// Program Description:
// This file implements a level 4 load balancer

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 31, 2024

package load_balancer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"jinx/internal/load_balancer/algo"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JinxLoadBalancingServer struct {
	config         types.JinxLoadBalancingServerConfig
	errorLogger    *slog.Logger
	serverLogger   *slog.Logger
	serverInstance *http.Server
	serverRootDir  string
	mode           string
	currentServer  int
	mutex          *sync.Mutex
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
		config:         config,
		errorLogger:    slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:   slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverRootDir:  serverRoot,
		serverInstance: nil,
		mode:           loadBalancerMode,
		currentServer:  -1,
		mutex:          &sync.Mutex{},
	}
}

func (jx *JinxLoadBalancingServer) Start() types.JinxServer {
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

	go func() {
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
	}()

	return jx
}

// Stop gracefully shuts down the JinxHttpServer instance, ensuring all ongoing requests are
// completed before closure. This method initiates a graceful shutdown by creating a context
// with a 15-second timeout, signaling the server to cease accepting new requests and wait
// for existing requests to conclude within this timeframe. If the server successfully shuts
// down within the allotted time, it logs a confirmation message. If an error occurs during
// shutdown (e.g., the timeout is exceeded), it logs the error. This method is essential for
// clean server termination, minimizing the risk of interrupting active client connections
// and ensuring resources are properly released.
//
// The method does nothing if the server instance (`serverInstance`) is nil, which implies
// that the server has not been started or has already been stopped. This check prevents
// potential nil pointer dereferences and ensures the method's idempotency, allowing it to
// be safely called multiple times.
//
// Usage:
// - This method should be called when the server needs to be stopped, such as in response
//   to an interrupt signal or a shutdown command. It is designed to be used as part of
//   the server's lifecycle management, facilitating controlled and safe server termination.

func (jx *JinxLoadBalancingServer) Stop() {
	if jx.serverInstance == nil {
		return
	}
	// Create a context with a timeout to tell the server how long to wait for existing requests to finish
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := jx.serverInstance.Shutdown(ctx); err != nil {
		jx.errorLogger.Error(fmt.Sprintf("Server shutdown error: %s", err))
	}

	jx.serverLogger.Info(fmt.Sprintf("Successfully shutdown server manually"))
}

// Restart attempts to gracefully restart the JinxHttpServer instance. It first checks if the server
// is running (`serverInstance` is not nil); if not, it returns nil, indicating there's no server to restart.
// If the server is running, it performs a graceful shutdown by calling the Stop method, which waits
// for ongoing requests to finish before stopping the server. After stopping, it immediately initiates
// the server's restart process in a new goroutine, allowing the method to return without waiting for
// the server to restart. This non-blocking approach facilitates rapid restarts without stalling the
// calling thread or process.
//
// The server is restarted with TLS if both `CertFile` and `KeyFile` are specified in the server's
// configuration (`config`). If these are not provided, it restarts without TLS. If an error occurs
// during the restart process, such as issues with binding to the specified port or problems with
// the TLS configuration, it logs the error and terminates the application with `log.Fatal`.
// This method ensures the server can be dynamically restarted with updated configurations or
// in response to certain runtime conditions without manual intervention.
//
// Usage:
// - This method is useful in scenarios where changes to the server's configuration or runtime
//   environment necessitate a restart, such as after updating TLS certificates or changing server
//   settings. It provides a programmatic way to restart the server, encapsulating the shutdown
//   and restart logic within the JinxHttpServer's lifecycle management.
//
// Returns:
// - A reference to the restarted JinxHttpServer instance (`jx`), allowing for chaining or further
//   actions. Returns nil if the server was not running at the time of the call, indicating there
//   was no server instance to restart.

func (jx *JinxLoadBalancingServer) Restart() types.JinxServer {
	if jx.serverInstance == nil {
		return nil
	}

	jx.Stop()
	go func() {
		if jx.config.CertFile != "" && jx.config.KeyFile != "" {
			err := jx.serverInstance.ListenAndServeTLS(jx.config.CertFile, jx.config.KeyFile)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
				log.Fatal(err)
			}
			return
		}

		// Start the server
		err := jx.serverInstance.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
			log.Fatal(err)
		}
	}()

	return jx
}

// Destroy performs a complete teardown of the JinxHttpServer instance, effectively stopping the server
// and removing its working directory and all contained data. This method first checks if the server instance
// (`serverInstance`) is currently running; if it is not, the method returns immediately, as there is no server
// to stop or resources to clean up. If the server is running, it calls the Stop method to gracefully shut down
// the server, ensuring that all ongoing requests are allowed to complete before the server stops accepting new
// requests. Following the server shutdown, Destroy removes the server's working directory (`serverWorkingDir`),
// which includes all files and subdirectories related to the server's operation. This operation is irreversible
// and should be used with caution, as it results in the loss of any data stored in the server's working directory.
//
// Usage:
//   - The Destroy method is intended for scenarios where a complete cleanup of the server and its resources is
//     required, such as during decommissioning, or in testing and development environments where a fresh start
//     is needed. It provides a way to programmatically remove all traces of the server's operation from the host
//     system.
//
// Note:
//   - Care should be taken when calling this method, as it will delete the server's working directory and all its
//     contents, which may include application data, logs, and configuration files. Ensure that any important data
//     is backed up before calling Destroy.
func (jx *JinxLoadBalancingServer) Destroy() {
	if jx.serverInstance == nil {
		return
	}

	jx.Stop()
	_ = os.RemoveAll(jx.serverRootDir)

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
