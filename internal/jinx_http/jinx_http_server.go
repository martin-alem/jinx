// File: jinx_http_server.go
// Package: jinx_http

// Program Description:
// This file implements a http server with https support.
// Server options are passed as command line options
// These options are used to set up the server

// Author: Martin Alemajoh
// Jinx- v1.0.0
// Created on: January 16, 2024

package jinx_http

import (
	"context"
	"errors"
	"fmt"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type JinxHttpServer struct {
	config           types.JinxHttpServerConfig // Server configuration settings.
	errorLogger      *slog.Logger               // Logger for error messages.
	serverLogger     *slog.Logger               // Logger for general server activity.
	serverWorkingDir string                     // Server root dir where website files are stored
	serverInstance   *http.Server
}

// NewJinxHttpServer initializes a new instance of JinxHttpServer with the provided configuration
// and server root directory. It sets up error and server activity logging by creating and configuring
// log files based on the provided configuration paths. This function ensures that the server is ready
// to log important events and errors before it starts serving requests. It also verifies that the
// specified server root directory is readable, which is essential for serving files correctly.
//
// Parameters:
//   - config: A types.JinxHttpServerConfig struct containing configuration settings for the server,
//             including IP, port, log file paths, and SSL certificate paths if HTTPS is to be supported.
//   - serverRoot: A string specifying the path to the server's root directory from which static files
//                 will be served. This directory should contain the web content (e.g., HTML, CSS, JS files)
//                 that the server will deliver in response to HTTP requests.
//
// Returns:
//   - A pointer to an instance of JinxHttpServer, fully initialized with loggers and configuration.
//     This server instance is ready to start serving requests after this function completes.
//
// The function performs the following operations:
// 1. Attempts to open or create log files for both error logging and server activity logging. If it
//    fails to open or create these files, the program will terminate with an error message.
// 2. Verifies that the server root directory is readable. If the directory does not exist, is not readable,
//    or if any other error occurs during this check, the program will terminate, indicating that the
//    server cannot serve files from an inaccessible directory.
//
// This initialization function is critical for setting up a JinxHttpServer with the appropriate
// logging mechanisms and verifying that its configuration is viable for serving web content.
// It ensures that the server is properly configured and ready to handle HTTP requests efficiently
// and reliably.

func NewJinxHttpServer(config types.JinxHttpServerConfig, serverWorkingDir string) *JinxHttpServer {

	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr)
	}

	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr)
	}

	//Make sure www exists and it's readable
	readable, err := helper.IsDirReadable(serverWorkingDir)
	if !readable || err != nil {
		log.Fatalf("%s does not exist or is not readable", serverWorkingDir)
	}

	return &JinxHttpServer{
		config:           config,
		errorLogger:      slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:     slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverWorkingDir: serverWorkingDir,
		serverInstance:   nil,
	}
}

// Start initializes and runs the Jinx HTTP server, configuring it to listen on the IP and port
// specified in its configuration. This method sets up the server with specified timeouts and
// maximum header sizes to ensure efficient operation. It supports both HTTP and HTTPS (if certificate
// and key files are provided) and implements graceful shutdown to handle ongoing requests properly
// before stopping the server.
//
// The method performs the following operations:
//  1. Logs the server's start-up on the configured IP address and port.
//  2. Configures a http.Server instance with the server's address, read/write timeouts, maximum header
//     size, and sets the current JinxHttpServer instance as the handler for incoming requests.
//  3. Sets up a signal listener to gracefully handle interrupt and termination signals (SIGINT, SIGTERM),
//     allowing the server to finish processing current requests before shutting down.
//  4. Starts listening for incoming HTTP or HTTPS connections, depending on the configuration. For HTTPS,
//     it requires paths to the SSL certificate and key files.
//  5. On receiving a shutdown signal, attempts to gracefully shut down the server, logging any errors
//     encountered during the shutdown process.
//
// If the server fails to start or encounters an error during runtime that isn't related to a normal
// shutdown (ErrServerClosed), the error is logged, and the program is terminated using log.Fatal.
//
// This method encapsulates the entire lifecycle of the server from start-up to graceful shutdown,
// making it easy to manage the server's operation within the context of an application.
func (jx *JinxHttpServer) Start() types.JinxServer {
	addr := fmt.Sprintf("%s:%d", jx.config.IP, jx.config.Port)
	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx on %s", addr))

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
		err := s.ListenAndServeTLS(jx.config.CertFile, jx.config.KeyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
			log.Fatal(err)
		}
		return nil
	}

	// Start the server
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

func (jx *JinxHttpServer) Stop() {
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

func (jx *JinxHttpServer) Restart() types.JinxServer {
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
// - The Destroy method is intended for scenarios where a complete cleanup of the server and its resources is
//   required, such as during decommissioning, or in testing and development environments where a fresh start
//   is needed. It provides a way to programmatically remove all traces of the server's operation from the host
//   system.
//
// Note:
// - Care should be taken when calling this method, as it will delete the server's working directory and all its
//   contents, which may include application data, logs, and configuration files. Ensure that any important data
//   is backed up before calling Destroy.

func (jx *JinxHttpServer) Destroy() {
	if jx.serverInstance == nil {
		return
	}

	jx.Stop()
	_ = os.RemoveAll(jx.serverWorkingDir)

}

// ServeHTTP is the core method implementing the http.Handler interface for JinxHttpServer, making
// it capable of serving HTTP requests. This method orchestrates the server's response to incoming
// requests by logging request details, resolving the appropriate file path based on the request,
// serving the requested file or a custom 404 page if the file cannot be found, and logging the
// response details including the time taken to serve the request. This structured approach ensures
// a consistent and efficient handling of web requests, enhancing the server's reliability and
// maintainability.
//
// Parameters:
//   - w: The http.ResponseWriter that is used to write the HTTP response to the client. It is
//     utilized for setting response headers and writing the response body.
//   - r: The *http.Request representing the client's request, containing all necessary information
//     such as the requested URL, HTTP method, and headers.
//
// Workflow:
//  1. Log the incoming request details for monitoring and debugging purposes.
//  2. Resolve the file path for the requested resource. This involves determining the correct
//     file to serve based on the request URL and the server's configuration. If the file does not
//     exist, or an error occurs in resolving the file path, a custom 404 page is served instead.
//  3. Serve the resolved file to the client, setting appropriate response headers for caching and
//     server identification.
//  4. Log the response details, specifically the duration it took to serve the request, to aid in
//     performance monitoring and optimization efforts.
//
// The ServeHTTP method ensures that all incoming HTTP requests are handled in a uniform manner,
// leveraging the server's configuration and custom logic for file resolution, error handling, and
// logging. This makes JinxHttpServer a flexible and robust solution for serving web content.
func (jx *JinxHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Log the incoming request
	jx.serverLogger.Info(fmt.Sprintf("Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL.String(), r.RemoteAddr))

	// Determine the file to serve
	filePath, err := jx.ResolveFilePath(r)
	if err != nil {
		jx.serverLogger.Info(err.Error())
		jx.Serve404(w, filePath) // Serve the 404 page if an error occurs
		return
	}

	// Serve the file
	jx.ServeFile(w, r, filePath)

	// Log the response details
	responseTime := time.Since(startTime)
	jx.serverLogger.Info(fmt.Sprintf("Served response: Duration=%s", responseTime))
}

// ResolveFilePath determines the absolute file path to serve in response to an HTTP request.
// It dynamically resolves the file path based on the request's host header and the requested URL path,
// taking into account the server's configuration for the website root directory and handling default
// content and not found scenarios.
//
// The function supports serving content from different host directories and defaults to serving from
// a common server root directory if the requested host's directory is not found or is not readable.
// It also defaults to serving an 'index.html' file for root directory requests and returns a path to
// a '404.html' file (or equivalent) if the requested file does not exist or is a directory.
//
// Parameters:
//   - r: The *http.Request object representing the client's request. It contains the Host and URL
//     from which the function extracts information to resolve the file path.
//
// Returns:
//   - A string representing the absolute path to the file that should be served in response to the request.
//     This path is constructed based on the server's configuration, the request's host, and the URL path.
//   - An error if the requested file does not exist or is a directory (indicating a '404 Not Found' scenario),
//     with the error message including the path of the requested file.
//
// The function first extracts the host from the request's Host header and the URL path from the request's URL.
// It then determines the appropriate root directory to use (either a specific directory for the host or the
// server's default root directory) and constructs the absolute file path to serve. If the requested URL path
// points to the root directory or does not specify a file, the function defaults to serving 'index.html' from
// the determined root directory. If the file does not exist or is a directory, it sets up to serve a '404 Not Found'
// page instead, returning its path and an error to indicate the file was not found.
func (jx *JinxHttpServer) ResolveFilePath(r *http.Request) (string, error) {
	host := strings.Split(r.Host, ":")[0]
	root := jx.config.WebsiteRoot
	urlPath := path.Clean(r.URL.Path)

	// Determine the root directory based on the host
	if helper.IsLocalhostOrIP(host) {
		root = jx.serverWorkingDir
		host = constant.DEFAULT_WEBSITE_ROOT
	} else if readable, _ := helper.IsDirReadable(filepath.Join(root, host)); !readable {
		root = jx.serverWorkingDir
		host = constant.DEFAULT_WEBSITE_ROOT
	}

	// Determine the specific file to serve
	file := filepath.Join(root, host, urlPath)
	if urlPath == "" || urlPath == "/" {
		file = filepath.Join(file, constant.INDEX_FILE)
	} else if info, err := os.Stat(file); err != nil || info.IsDir() {
		return filepath.Join(root, host, constant.NOT_FOUND), fmt.Errorf("file not found: %s", file)
	}

	return file, nil
}

// ServeFile sends a static file located at the specified filePath to the client. It sets appropriate
// HTTP headers before sending the file to optimize for caching and to identify the server software.
// This function is primarily used to serve static content like HTML, CSS, JavaScript files, images,
// and more, making it a key component of the server's capability to deliver web resources efficiently.
//
// Parameters:
//   - w: The http.ResponseWriter object used to write the HTTP response headers and content to the client.
//   - r: The *http.Request object representing the client's request. This parameter is required by
//     http.ServeFile to manage specifics of the request, such as range headers.
//   - filePath: A string representing the absolute path to the file that should be served to the client.
//     The function reads and streams this file as the HTTP response body.
//
// This method first sets the "Cache-Control" header to instruct clients and intermediaries to cache the
// response for 3600 seconds (1 hour), reducing the need for subsequent requests for the same resource
// to hit the server. It also sets the "Server" header to the value of constant.SOFTWARE_NAME, which
// identifies the server software to clients without exposing detailed version information for security.
// Finally, it uses the http.ServeFile function to handle the file serving, including support for
// partial content delivery and automatic MIME type detection.
func (jx *JinxHttpServer) ServeFile(w http.ResponseWriter, r *http.Request, filePath string) {
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("Server", constant.SOFTWARE_NAME)
	http.ServeFile(w, r, filePath)
}

// Serve404 sends a 404 Not Found response to the client with the content of a specified file.
// This function is designed to handle scenarios where a requested resource cannot be found on the server.
// It attempts to read the content of the specified file (typically a custom 404 error page) and sends it
// as the response body to provide a more user-friendly error message. If the file cannot be read,
// it sends a default "404 Not Found" error message. Additionally, if there's an error while writing the file
// content to the response, it sends a "500 Internal Server Error" message.
//
// Parameters:
//   - w: The http.ResponseWriter object to write the HTTP response.
//   - filePath: The path to the file that contains the custom 404 error page content. This file is read
//     and its content is sent as the response body for the 404 error.
//
// Note: This function sets the HTTP status code to 404 Not Found when serving the custom error page.
// If an error occurs while reading the custom error file, the status code is still set to 404.
// However, if an error occurs while writing the content to the response, the status code is set to 500 Internal Server Error.
func (jx *JinxHttpServer) Serve404(w http.ResponseWriter, filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	if _, writeErr := w.Write(content); writeErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
