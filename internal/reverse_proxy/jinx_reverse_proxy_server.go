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
	"context"
	"errors"
	"fmt"
	"jinx/pkg/util/helper"
	"jinx/pkg/util/types"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type JinxReverseProxyServer struct {
	config        types.JinxReverseProxyServerConfig
	errorLogger   *slog.Logger
	serverLogger  *slog.Logger
	serverRootDir string
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
func NewJinxReverseProxyServer(config types.JinxReverseProxyServerConfig, serverRoot string) *JinxReverseProxyServer {

	errorLogFile, errorLogErr := os.OpenFile(filepath.Join(config.LogRoot, "error.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if errorLogErr != nil {
		log.Fatal(errorLogErr)
	}

	serverLogFile, logFileErr := os.OpenFile(filepath.Join(config.LogRoot, "server.log"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logFileErr != nil {
		log.Fatal(logFileErr)
	}

	return &JinxReverseProxyServer{
		config:        config,
		errorLogger:   slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:  slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverRootDir: serverRoot,
	}
}

func (jx *JinxReverseProxyServer) HandleProxyRequest(w http.ResponseWriter, r *http.Request, upstreamURL string) {
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

func (jx *JinxReverseProxyServer) Start() {
	addr := fmt.Sprintf("%s:%d", jx.config.IP, jx.config.Port)

	s := &http.Server{
		Addr:           addr,
		Handler:        jx,
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

	if jx.config.CertFile != "" && jx.config.KeyFile != "" {
		jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Reverse Proxy Sever on %s using HTTPS Protocol", addr))
		err := s.ListenAndServeTLS(jx.config.CertFile, jx.config.KeyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
			log.Fatal(err)
		}
		return
	}

	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Reverse Proxy Sever on %s using HTTP Protocol", addr))
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
		log.Fatal(err)
	}
}

func (jx *JinxReverseProxyServer) DetermineUpstreamURL(r *http.Request) (string, error) {
	path := filepath.Clean(r.URL.Path)

	upStreamUrl, ok := jx.config.RouteTable[path]
	if !ok {
		msg := fmt.Sprintf("%s does not exist in route table:", path)
		return "", errors.New(msg)
	}

	return upStreamUrl, nil

}

func (jx *JinxReverseProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jx.logRequestDetails(r)

	// Example: Determine the upstream URL based on the request
	upstreamURL, err := jx.DetermineUpstreamURL(r)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// Dispatch the request
	jx.HandleProxyRequest(w, r, upstreamURL)

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
func (jx *JinxReverseProxyServer) logRequestDetails(r *http.Request) {
	jx.serverLogger.Info(fmt.Sprintf("Received request: Method=%s, URL=%s, RemoteAddr=%s", r.Method, r.URL.String(), r.RemoteAddr))
}
