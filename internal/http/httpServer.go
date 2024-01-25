// Package http implements the JinxHttpServer, part of the Jinx web server project.
// This package provides the core server functionality including handling of HTTP requests,
// logging, and file serving based on the request.
package http

import (
	"fmt"
	"jinx/pkg/util"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// JinxHttpServer struct represents the HTTP server. It includes the server configuration,
// error logger, and server activity logger.
type JinxHttpServer struct {
	config       util.JinxHttpServerConfig // Server configuration settings.
	errorLogger  *slog.Logger              // Logger for error messages.
	serverLogger *slog.Logger              // Logger for general server activity.
}

// NewJinxHttpServer initializes and returns a new instance of JinxHttpServer.
// It sets up logging for both server activity and errors.
func NewJinxHttpServer(config util.JinxHttpServerConfig) *JinxHttpServer {
	// Open or create the error log file.
	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr) // Fatal log and exit if unable to open/create the error log file.
	}

	// Open or create the server activity log file.
	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr) // Fatal log and exit if unable to open/create the server log file.
	}

	// Return a new JinxHttpServer instance with the specified configuration and loggers.
	return &JinxHttpServer{
		config:       config,
		errorLogger:  slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger: slog.New(slog.NewJSONHandler(serverLogFile, nil)),
	}
}

// Start runs the HTTP server. It configures the server and starts listening on the specified port.
// It logs the server start event and any errors during operation.
func (jx *JinxHttpServer) Start() {
	// Format the address using the configured port.
	addr := fmt.Sprintf(":%d", jx.config.Port)
	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx on %s", addr))

	// Configure and start the HTTP server.
	s := &http.Server{
		Addr:           addr,
		Handler:        NewJinxHttpServer(jx.config),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
		log.Fatal(err) // Fatal log and exit if the server fails to start.
	}
}

// ServeHTTP handles incoming HTTP requests. It serves static files, handles file
// not found scenarios, and logs request details.
func (jx *JinxHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now() // Track start time for response time logging.
	// Log the received request details.
	jx.serverLogger.Info(fmt.Sprintf("Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL, r.RemoteAddr))

	// Get the current working directory to serve local files.
	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		jx.errorLogger.Error(cwdErr.Error())
		log.Fatal(cwdErr) // Fatal log and exit if unable to get current working directory.
	}

	// Extract the host from the request and split if a non-default port is included.
	host := strings.Split(r.Host, ":")[0]
	root := jx.config.Root

	path := r.URL.Path

	// Split the path to directory and file components.
	dir, file := filepath.Split(path)

	// Check if the host is local and set the root directory accordingly.
	if util.IsLocalhostOrIP(host) {
		host = "www"
		root = cwd
	}

	// Set default file to 'index.html' if the request is for the root directory.
	if file == "" || file == "/" {
		file = "index.html"
	}

	// Check if the requested file exists, serve '404.html' if not found.
	ok, file := util.FileExist(filepath.Join(root, host, dir), file)
	if ok {
		jx.serverLogger.Info(fmt.Sprintf("Serving file: %s", file))
	} else {
		jx.serverLogger.Info(fmt.Sprintf("File not found, serving 404 page: %s", file))
	}

	// Set response headers and serve the file.
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("Server", "Jinx")
	http.ServeFile(w, r, file)

	// Log the response details including the duration.
	responseTime := time.Since(startTime)
	jx.serverLogger.Info(fmt.Sprintf("Served response: StatusCode=%d, Duration=%s", http.StatusOK, responseTime))
}
