// File: jinx_reverse_proxy_server.go
// Package: reverse_proxy

// Program Description:
// This file implements a reverse proxy with https support.
// Server options are passed as command line options
// These options are used to set up the server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 16, 2024

package reverse_proxy

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
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type JinxReverseProxyServer struct {
	config           types.JinxReverseProxyServerConfig
	errorLogger      *slog.Logger
	serverLogger     *slog.Logger
	serverWorkingDir string
	serverInstance   *http.Server
}

// NewJinxReverseProxyServer initializes a new instance of JinxReverseProxyServer with the provided configuration
// and server root directory. It sets up error and server activity logging by creating and configuring
// log files based on the provided configuration paths. This function ensures that the server is ready
// to log important events and errors before it starts handling requests.
//
// Parameters:
//   - config: A types.JinxHttpServerConfig struct containing configuration settings for the server,
//     including IP, port, log file paths, and SSL certificate paths if HTTPS is to be supported.
//   - serverRoot: A string specifying the path to the server's root directory
//
// Returns:
//   - A pointer to an instance of JinxReverseProxyServer, fully initialized with loggers and configuration.
//     This server instance is ready to start handling requests after this function completes.
//
// The function performs the following operations:
//  1. Attempts to open or create log files for both error logging and server activity logging. If it
//     fails to open or create these files, the program will terminate with an error message.
//
// This initialization function is critical for setting up a JinxReverseProxyServer with the appropriate
// logging mechanisms and verifying that its configuration is viable for handling request
func NewJinxReverseProxyServer(config types.JinxReverseProxyServerConfig, serverWorkingDir string) *JinxReverseProxyServer {

	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr)
	}

	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr)
	}

	return &JinxReverseProxyServer{
		config:           config,
		errorLogger:      slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:     slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverWorkingDir: serverWorkingDir,
		serverInstance:   nil,
	}
}

// Start initiates the JinxReverseProxyServer, making it ready to handle incoming HTTP or HTTPS requests
// based on its configuration. This method configures and starts an internal http.Server with settings
// specified in the JinxReverseProxyServer's configuration, such as IP address, port, and SSL certificates
// for HTTPS. It also sets up a graceful shutdown mechanism to handle interrupt or termination signals,
// ensuring that the server can shut down cleanly without abruptly disconnecting clients.
//
// The server is started with HTTPS if both a certificate file and a key file are provided in the
// configuration; otherwise, it falls back to HTTP. This method includes setting timeouts for reading
// and writing to prevent slow or malicious clients from affecting the server's performance.
//
// Parameters:
//   - None.
//
// Returns:
//   - A reference to the JinxReverseProxyServer instance, allowing for method chaining or capturing the
//     server instance for further operations.
//
// Workflow:
//   1. Constructs the server address from the configured IP and port.
//   2. Creates a new http.Server instance with appropriate timeouts and the JinxReverseProxyServer as the handler.
//   3. Listens for OS interrupt or termination signals in a separate goroutine to gracefully shut down the server
//      when such signals are received.
//   4. Starts the server using HTTPS if SSL certificates are provided; otherwise, starts an HTTP server.
//   5. Logs the server start-up and any errors encountered during operation. If the server is shut down
//      due to a received signal, it attempts a graceful shutdown, waiting for ongoing requests to complete.
//
// Usage:
//   - This method should be called after the JinxReverseProxyServer has been properly configured and is
//     ready to start serving requests. It's typically the last step in the server setup process, transitioning
//     the server from a configured state to an active state.
//
// Note:
//   - This method blocks if the server starts successfully, only returning if an error occurs that
//     prevents the server from operating (excluding http.ErrServerClosed, which is expected during
//     a graceful shutdown). Ensure that any necessary preparations are completed before calling Start.

func (jx *JinxReverseProxyServer) Start() types.JinxServer {
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
		jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Reverse Proxy Sever on %s using HTTPS Protocol", addr))
		err := s.ListenAndServeTLS(jx.config.CertFile, jx.config.KeyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
			log.Fatal(err)
		}
		return jx
	}

	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Reverse Proxy Sever on %s using HTTP Protocol", addr))
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

func (jx *JinxReverseProxyServer) Stop() {
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

func (jx *JinxReverseProxyServer) Restart() types.JinxServer {
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
func (jx *JinxReverseProxyServer) Destroy() {
	if jx.serverInstance == nil {
		return
	}

	jx.Stop()
	_ = os.RemoveAll(jx.serverWorkingDir)

}

// HandleHTTPProxyRequest forwards an incoming HTTP request to an upstream server specified by upstreamURL,
// acting as a reverse proxy. This method dynamically modifies the request to reflect the target
// upstream service's scheme, host, and path, then forwards the request using Go's built-in ReverseProxy.
// It also provides custom error handling, logging any errors that occur during the proxy operation.
//
// The function logs the start of request handling, modifies the request to point to the upstream service,
// and then uses a httputil.ReverseProxy instance to serve the request. If the upstream service encounters
// an error (e.g., connection failure, timeout), the ErrorHandler logs the error before responding to the client.
// After successfully serving the request, it logs the completion of request handling.
//
// Parameters:
//   - w: The http.ResponseWriter that is used to write the HTTP response back to the client. It may be used
//     by the ReverseProxy for writing directly to the client in case of errors or forwarding the response from the upstream service.
//   - r: The *http.Request representing the client's request. This request is modified to direct it to the upstream service.
//   - upstreamURL: A string representing the URL of the upstream service to which the request should be forwarded.
//
// Workflow:
//  1. Logs the initiation of request handling to the specified upstream URL.
//  2. Creates a new httputil.ReverseProxy instance with a Director function that modifies the request to point to the upstream service.
//  3. Sets a custom ErrorHandler on the proxy to log any errors that occur during the request forwarding.
//  4. Calls ServeHTTP on the proxy instance to forward the request and handle the response.
//  5. Logs the completion of request handling.
//
// Usage:
//   - This method is intended to be called from within the ServeHTTP method of JinxReverseProxyServer or similar
//     request handling contexts where requests need to be dynamically forwarded to configured upstream services.
//     It abstracts the complexities of modifying requests and handling errors, making it easier to implement
//     reverse proxy functionality.
//
// Note:
//   - The upstreamURL parameter must be a valid URL, including the scheme (http/https) and host. The path component
//     of upstreamURL is used as the base path for the forwarded request. This method does not validate the availability
//     or responsiveness of the upstream service; it is the caller's responsibility to ensure that the upstreamURL points
//     to a valid and available service.
func (jx *JinxReverseProxyServer) HandleHTTPProxyRequest(w http.ResponseWriter, r *http.Request, upstreamURL string) {
	jx.serverLogger.Info(fmt.Sprintf("Handling %s request...", upstreamURL))
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			target, _ := url.Parse(upstreamURL)
			r.URL.Scheme = target.Scheme
			r.URL.Host = target.Host
			r.Host = target.Host
			r.URL.Path = helper.SingleJoiningSlash(target.Path, r.URL.Path)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			jx.errorLogger.Error(err.Error(), err, r)
		},
	}
	proxy.ServeHTTP(w, r)
	jx.serverLogger.Info(fmt.Sprintf("Handling %s request completed...", upstreamURL))
}

// handleHTTPSProxyRequest manages the forwarding of HTTPS requests through the JinxReverseProxyServer.
// It uses HTTP connection hijacking to intercept the client's request and establish a direct TCP connection
// to the requested destination server. This method allows the proxy server to serve as a transparent intermediary
// for HTTPS traffic, facilitating the secure and encrypted communication between the client and the destination
// server without terminating the SSL/TLS connection.
//
// The process involves the following steps:
//  1. Hijacking the client's HTTP connection to gain direct control over the socket.
//  2. Establishing a TCP connection to the destination server specified in the request's Host header.
//  3. Sending a "200 Connection Established" response back to the client, indicating that the proxy connection
//     has been successfully established.
//  4. Initiating the bidirectional streaming of data between the client and the destination server to facilitate
//     the transparent proxying of HTTPS requests and responses.
//
// Parameters:
//   - w: The http.ResponseWriter that allows access to the underlying network connection via hijacking.
//   - r: The *http.Request representing the client's HTTPS request. The Host header of this request is used to
//     determine the destination server's address.
//
// Workflow:
//   - If the http.ResponseWriter does not support hijacking, an internal server error is returned to the client.
//   - Attempts to hijack the client's connection. On failure, an internal server error is returned to the client.
//   - Establishes a TCP connection to the destination server. If this fails, the client's hijacked connection is
//     closed and the function returns.
//   - Sends a "200 Connection Established" response to the client over the hijacked connection.
//   - Starts two goroutines to stream data bidirectionally between the client and the destination server.
//
// Usage:
//   - This method is specifically designed for handling HTTPS requests in a reverse proxy setup where it's necessary
//     to forward encrypted traffic without decrypting it. It should be called when the proxy server detects an HTTPS
//     CONNECT method, indicating the client's intention to establish a secure connection through the proxy.
//
// Note:
//   - The method relies on the ability to hijack HTTP connections, a feature that must be supported by the server's
//     underlying HTTP library. It is critical for enabling the proxy functionality for HTTPS traffic.
//   - The "transfer" function referenced in the code is responsible for copying data between the client and destination
//     connections. It should handle errors and close both connections when the data transfer is complete or an error occurs.
func (jx *JinxReverseProxyServer) handleHTTPSProxyRequest(w http.ResponseWriter, r *http.Request) {
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

// handleWebSocketConnect establishes a WebSocket connection through the JinxReverseProxyServer by
// hijacking the client's HTTP connection and forwarding the WebSocket upgrade request to the
// destination server. It acts as a transparent intermediary, facilitating WebSocket communication
// between the client and the destination server without modifying or inspecting the transferred data.
//
// This method performs the following steps to establish the WebSocket connection:
//  1. Hijacks the client's HTTP connection to gain direct control over the underlying TCP connection.
//  2. Establishes a new TCP connection to the destination server specified in the request's Host header.
//  3. Forwards the client's WebSocket upgrade request to the destination server and reads the server's
//     upgrade response.
//  4. Forwards the destination server's WebSocket upgrade response back to the client, completing the
//     WebSocket handshake.
//  5. Initiates bidirectional streaming of WebSocket messages between the client and the destination server.
//
// Parameters:
//   - w: The http.ResponseWriter, which allows for hijacking the connection to directly manipulate the TCP socket.
//   - r: The *http.Request representing the client's request, including the WebSocket upgrade headers.
//
// Workflow:
//   - Checks for hijacking support and hijacks the client's connection. If hijacking is not supported or fails,
//     an internal server error is returned to the client.
//   - Connects to the destination server using the address specified in the request's Host header.
//     If the connection fails, the method returns without further action.
//   - Forwards the WebSocket upgrade request to the destination server and reads its response.
//     If forwarding fails or the response cannot be read, an internal server error is returned to the client.
//   - Forwards the destination server's response back to the client, completing the WebSocket handshake.
//   - Starts two goroutines to relay WebSocket messages between the client and the destination server,
//     allowing for full-duplex communication.
//
// Usage:
//   - This method is designed to handle WebSocket connections in a reverse proxy setup, enabling real-time
//     web applications to communicate through the proxy without any modifications to the WebSocket protocol.
//     It should be called when the proxy server detects a WebSocket upgrade request.
//
// Note:
//   - The "transfer" function referenced in the code is responsible for relaying WebSocket messages between
//     the client and destination connections. It should efficiently handle message streaming and close both
//     connections when the WebSocket session ends or an error occurs.
//   - Proper error handling and resource cleanup are crucial in this method to prevent resource leaks and
//     ensure the stability and reliability of the proxy server during WebSocket communication.
func (jx *JinxReverseProxyServer) handleWebSocketConnect(w http.ResponseWriter, r *http.Request) {
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

// DetermineUpstreamURL analyzes the incoming HTTP request to identify the appropriate upstream URL
// based on the request's path. It uses the server's routing table, which maps request paths to upstream
// service URLs, to find the destination URL where the request should be forwarded. This method is a key
// component of the reverse proxy's routing logic, enabling it to dynamically route requests to different
// backend services based on the URL path.
//
// Parameters:
//   - r: The *http.Request object representing the client's request. The URL path of this request is
//     used to look up the corresponding upstream URL in the server's routing table.
//
// Returns:
//   - A string representing the upstream URL to which the request should be forwarded. This URL is
//     retrieved from the server's routing table based on the request's path.
//   - An error if the request's path does not match any entry in the routing table, indicating that
//     there is no configured upstream URL for the requested path. The error message includes the
//     requested path to aid in debugging and configuration adjustments.
//
// Workflow:
//  1. Cleans the request's URL path to ensure a standard, predictable format for lookup.
//  2. Looks up the cleaned path in the server's routing table to find the corresponding upstream URL.
//  3. If the path exists in the routing table, returns the mapped upstream URL.
//  4. If the path does not exist in the routing table, returns an error indicating that the requested
//     path is not configured for forwarding, suggesting a potential misconfiguration or an unsupported request.
//
// Usage:
//   - This method should be called as part of the request handling process in the JinxReverseProxyServer
//     to determine the destination for each incoming request. It allows the reverse proxy to support
//     multiple backend services by routing requests to the appropriate service based on the request path.
//
// Note:
//   - The routing table (`jx.config.RouteTable`) must be properly configured before starting the server
//     to ensure that all expected paths are mapped to their respective upstream URLs. The absence of a
//     path in the routing table will result in an error, preventing the request from being forwarded
func (jx *JinxReverseProxyServer) DetermineUpstreamURL(r *http.Request) (string, error) {
	path := filepath.Clean(r.URL.Path)

	upStreamUrl, ok := jx.config.RouteTable[path]
	if !ok {
		msg := fmt.Sprintf("%s does not exist in route table:", path)
		return "", errors.New(msg)
	}

	return upStreamUrl, nil
}

// ServeHTTP is the core request handler for the JinxReverseProxyServer, implementing the http.Handler
// interface. This method is called for every incoming HTTP request to the server. It orchestrates the
// request processing workflow, including logging the request, determining the appropriate upstream URL
// for the request, and forwarding the request to its destination. Special handling is provided for
// HTTPS CONNECT requests and WebSocket upgrades, enabling the proxy to support a wide range of protocols
// and use cases.
//
// Parameters:
//   - w: The http.ResponseWriter that is used to write the HTTP response to the client. It allows for
//     sending responses, including status codes and headers, back to the requesting client.
//   - r: The *http.Request representing the client's request. It contains all the details of the request,
//     including the method, URL, headers, and body.
//
// Workflow:
//  1. Logs the incoming request, including its method, URL, and the client's remote address, for debugging
//     and monitoring purposes.
//  2. Determines the upstream URL by matching the request's path against the server's routing table. If no
//     match is found, responds with a 404 error.
//  3. For HTTPS CONNECT requests, invokes the handleHTTPSProxyRequest method to establish a tunnel between
//     the client and the destination server.
//  4. For WebSocket connection requests, identified by the "Upgrade: websocket" header, invokes the
//     handleWebSocketConnect method to facilitate the WebSocket handshake and data transfer.
//  5. For all other HTTP requests, forwards the request to the determined upstream URL using the
//     HandleHTTPProxyRequest method.
//
// Usage:
//   - This method is automatically called by the Go HTTP server infrastructure for each incoming request
//     to the JinxReverseProxyServer. It should not be called directly; instead, configure the server to
//     use an instance of JinxReverseProxyServer as its handler.
//
// Note:
//   - The routing table used by DetermineUpstreamURL must be properly configured to ensure correct
//     forwarding of requests. Misconfiguration may lead to requests being incorrectly routed or
//     rejected.
//   - The server must be capable of handling HTTPS CONNECT methods and WebSocket upgrades if these
//     features are to be used. This requires additional configuration, such as specifying SSL/TLS
//     certificates for HTTPS and ensuring the proxy can interpret and forward WebSocket communication.
func (jx *JinxReverseProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jx.serverLogger.Info(fmt.Sprintf("Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL.String(), r.RemoteAddr))

	// Example: Determine the upstream URL based on the request
	upstreamURL, err := jx.DetermineUpstreamURL(r)
	if err != nil {
		http.Error(w, err.Error(), 404)
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
		jx.handleWebSocketConnect(w, r)
		return
	}

	// Handle HTTP request
	jx.HandleHTTPProxyRequest(w, r, upstreamURL)

}
