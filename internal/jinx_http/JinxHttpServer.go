// Package jinx_http implements the JinxHttpServer, part of the Jinx web server project.
// This package provides the core server functionality including handling of HTTP requests,
// logging, and file serving based on the request.
package jinx_http

import (
	"context"
	"errors"
	"fmt"
	"jinx/pkg/util"
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

// JinxHttpServer struct represents the HTTP server. It includes the server configuration,
// error logger, and server activity logger.
type JinxHttpServer struct {
	config        util.JinxHttpServerConfig // Server configuration settings.
	errorLogger   *slog.Logger              // Logger for error messages.
	serverLogger  *slog.Logger              // Logger for general server activity.
	serverRootDir string                    // Server root dir for www dir
}

// NewJinxHttpServer initializes and returns a new instance of JinxHttpServer.
// It sets up logging for both server activity and errors.
func NewJinxHttpServer(config util.JinxHttpServerConfig, serverRoot string) *JinxHttpServer {

	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr)
	}

	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr)
	}

	//Make sure www exists and it's readable
	readable, err := util.IsDirReadable(serverRoot)
	if !readable || err != nil {
		log.Fatalf("%s does not exist or is not readable", serverRoot)
	}

	return &JinxHttpServer{
		config:        config,
		errorLogger:   slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:  slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverRootDir: serverRoot,
	}
}

// Start runs the HTTP server. It configures the server and starts listening on the specified port.
// It logs the server start event and any errors during operation.
func (jx *JinxHttpServer) Start() {
	addr := fmt.Sprintf("%s:%d", jx.config.IP, jx.config.Port)
	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx on %s", addr))

	s := &http.Server{
		Addr:           addr,
		Handler:        NewJinxHttpServer(jx.config, jx.serverRootDir),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

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

	// Start the server
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
		log.Fatal(err)
	}
}

// ServeHTTP handles incoming HTTP requests. It serves static files, handles file
// not found scenarios, and logs request details.
func (jx *JinxHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now() // Track start time for response time logging.

	jx.serverLogger.Info(fmt.Sprintf("Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL, r.RemoteAddr))

	// Extract the host from the request and split if a non-default port is included.
	host := strings.Split(r.Host, ":")[0]
	root := jx.config.WebsiteRoot

	urlPath := path.Clean(r.URL.Path)

	// Check if the host is local and set the root directory accordingly.
	if util.IsLocalhostOrIP(host) {
		host = util.WebRoot
		root = jx.serverRootDir
	} else {
		//Make sure user root/host path exist and is readable and writable
		readable, readErr := util.IsDirReadable(filepath.Join(root, host))
		if readErr != nil {
			jx.serverLogger.Info(fmt.Sprintf("Website directory not found or is not readable %s", readErr.Error()))
		}
		if !readable {
			host = util.WebRoot
			root = jx.serverRootDir
		}
	}

	file := ""
	// Set default file to 'index.html' if the request is for the root directory.
	if urlPath == "" || urlPath == "/" {
		file = filepath.Join(root, host, util.IndexFile)
		jx.serverLogger.Info(fmt.Sprintf("Serving index file: %s", file))
	} else {
		// Check if the requested file exists, serve '404.html' if not found.
		file = filepath.Join(root, host, urlPath)
		info, stateErr := os.Stat(file)
		if stateErr == nil && !info.IsDir() {
			jx.serverLogger.Info(fmt.Sprintf("Serving file: %s", file))
		} else {
			jx.serverLogger.Info(fmt.Sprintf("File not found, serving 404 page: %s", file))
			file = filepath.Join(root, host, util.NotFound)
			Serve404File(file, w)
			return
		}

	}

	// Set response headers and serve the file.
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("Server", util.Software)
	http.ServeFile(w, r, file)

	// Log the response details including the duration.
	responseTime := time.Since(startTime)
	jx.serverLogger.Info(fmt.Sprintf("Served response: StatusCode=%d, Duration=%s", http.StatusOK, responseTime))
}

func Serve404File(file string, w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("Server", util.Software)
	w.WriteHeader(http.StatusNotFound)

	content, readErr := os.ReadFile(file) // Use os.ReadFile to read the file content
	if readErr != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	if _, err := w.Write(content); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
}
