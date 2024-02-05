// File: jinx_forward_proxy_server.go
// Package: forward_proxy

// Program Description:
// This file implements a reverse proxy with https support.
// Server options are passed as command line options
// These options are used to set up the server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 31, 2024

package forward_proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type JinxForwardProxyServer struct {
	config         types.JinxForwardProxyServerConfig
	errorLogger    *slog.Logger
	serverLogger   *slog.Logger
	serverRootDir  string
	serverInstance *http.Server
}

func NewJinxForwardProxyServer(config types.JinxForwardProxyServerConfig, serverRoot string) *JinxForwardProxyServer {

	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr)
	}

	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr)
	}

	return &JinxForwardProxyServer{
		config:         config,
		errorLogger:    slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:   slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverRootDir:  serverRoot,
		serverInstance: nil,
	}
}

func (jx *JinxForwardProxyServer) Start() types.JinxServer {
	addr := fmt.Sprintf("%s:%d", jx.config.IP, jx.config.Port)

	s := &http.Server{
		Addr:           addr,
		Handler:        jx,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	jx.serverInstance = s

	// Set up a channel to listen for interrupt or termination signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Listen for shutdown signals in a separate goroutine
	go func() {
		sig := <-signalChan
		jx.serverLogger.Info(fmt.Sprintf("Received signal %v: shutting down server...", sig))

		// Create a context with a timeout to tell the server how long to wait for existing requests to finish
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server
		if err := s.Shutdown(ctx); err != nil {
			jx.errorLogger.Error(fmt.Sprintf("Server shutdown error: %s", err))
		}

		jx.serverLogger.Info(fmt.Sprintf("Successfully shutdown server"))
	}()

	if jx.config.CertFile != "" && jx.config.KeyFile != "" {
		jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Forward Proxy Sever on %s using HTTPS Protocol", addr))
		err := s.ListenAndServeTLS(jx.config.CertFile, jx.config.KeyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
			log.Fatal(err)
		}
		return jx
	}

	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Forward Proxy Sever on %s using HTTP Protocol", addr))
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
		log.Fatal(err)
	}

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

func (jx *JinxForwardProxyServer) Stop() {
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

func (jx *JinxForwardProxyServer) Restart() types.JinxServer {
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
func (jx *JinxForwardProxyServer) Destroy() {
	if jx.serverInstance == nil {
		return
	}

	jx.Stop()
	_ = os.RemoveAll(jx.serverRootDir)

}

func (jx *JinxForwardProxyServer) HandleHTTPProxyRequest(w http.ResponseWriter, r *http.Request) {
	jx.serverLogger.Info(fmt.Sprintf("Handling %s request...", r.URL.RequestURI()))
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			jx.errorLogger.Error(err.Error(), err, r)
		},
	}
	proxy.ServeHTTP(w, r)
	jx.serverLogger.Info(fmt.Sprintf("Handling %s request completed...", r.URL.RequestURI()))
}

func (jx *JinxForwardProxyServer) ValidateUpstreamURL(r *http.Request) error {

	reqHost := strings.Split(r.Host, ":")[0]

	if inList := helper.InList[string](jx.config.BlackList, reqHost, func(a string, b string) bool {
		return a == b
	}); inList {
		msg := fmt.Sprintf("%s has been blacklisted", reqHost)
		return errors.New(msg)
	}

	return nil

}

func (jx *JinxForwardProxyServer) handleHTTPSProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "HTTP Server does not support hijacking", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Connect to the destination server
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		_ = clientConn.Close()
		return
	}

	// Send a 200 OK response to client
	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Stream data between the client and the destination server
	go helper.Transfer(clientConn, destConn)
	go helper.Transfer(destConn, clientConn)
}

func (jx *JinxForwardProxyServer) handleWebSocketProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "HTTP Server does not support hijacking", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(clientConn net.Conn) {
		_ = clientConn.Close()
	}(clientConn)

	// Connect to the destination server
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		return
	}
	defer func(destConn net.Conn) {
		_ = destConn.Close()
	}(destConn)

	// Forward the client's WebSocket upgrade request to the destination server
	err = r.Write(destConn)
	if err != nil {
		http.Error(w, "Failed to send WebSocket upgrade request to the destination server", http.StatusInternalServerError)
		return
	}

	// Read the response from the destination server
	response, err := http.ReadResponse(bufio.NewReader(destConn), r)
	if err != nil {
		http.Error(w, "Failed to read WebSocket upgrade response from the destination server", http.StatusInternalServerError)
		return
	}

	// Forward the destination server's response back to the client
	err = response.Write(clientConn)
	if err != nil {
		http.Error(w, "Failed to send WebSocket upgrade request to the client", http.StatusInternalServerError)
		return
	}

	// At this point, the WebSocket handshake is complete, and we can start relaying messages
	go helper.Transfer(destConn, clientConn)
	go helper.Transfer(clientConn, destConn)
}

func (jx *JinxForwardProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jx.logRequestDetails(r)

	// Validate the upstream URL for HTTP requests
	err := jx.ValidateUpstreamURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden) // Use 403 for forbidden access
		return
	}

	// Special handling for HTTPS CONNECT requests
	if r.Method == http.MethodConnect {
		jx.handleHTTPSProxyRequest(w, r)
		return
	}

	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	connectionHeader := strings.ToLower(r.Header.Get("Connection"))

	if upgradeHeader == "websocket" && strings.Contains(connectionHeader, "upgrade") {
		jx.handleWebSocketProxyRequest(w, r)
		return
	}

	// Handle HTTP requests as before
	jx.HandleHTTPProxyRequest(w, r)
}

// logRequestDetails logs the essential details of an incoming HTTP request.
// This method is instrumental in providing visibility into the traffic the server is handling,
// logging vital information such as the HTTP method used, the request URL, and the remote address
// of the client making the request. This information can be invaluable for debugging, monitoring,
// and analyzing the patterns of requests received by the server.
//
// Parameters:
//   - r: The *http.Request object representing the client's request. It contains all the details
//     of the request made to the server, including the method, URL, and remote address.
//
// This function extracts the method, URL, and remote address from the provided request object
// and formats them into a log message. It then uses the server's serverLogger to log this message
// at the Info level. This centralized logging of request details aids in maintaining a clear and
// concise record of server activity, facilitating easier troubleshooting and operational oversight.
func (jx *JinxForwardProxyServer) logRequestDetails(r *http.Request) {
	jx.serverLogger.Info(fmt.Sprintf("Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL.String(), r.RemoteAddr))
}
