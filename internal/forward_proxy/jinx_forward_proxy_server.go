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
	"context"
	"errors"
	"fmt"
	"io"
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
	config        types.JinxForwardProxyServerConfig
	errorLogger   *slog.Logger
	serverLogger  *slog.Logger
	serverRootDir string
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
		config:        config,
		errorLogger:   slog.New(slog.NewJSONHandler(errorLogFile, nil)),
		serverLogger:  slog.New(slog.NewJSONHandler(serverLogFile, nil)),
		serverRootDir: serverRoot,
	}
}

func (jx *JinxForwardProxyServer) Start() {
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

	jx.serverLogger.Info(fmt.Sprintf("Starting Jinx Forward Proxy Sever on %s using HTTP Protocol", addr))
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		jx.errorLogger.Error(fmt.Sprintf("Failed to start server: %s", err.Error()))
		log.Fatal(err)
	}
}

func (jx *JinxForwardProxyServer) HandleProxyRequest(w http.ResponseWriter, r *http.Request) {
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

func (jx *JinxForwardProxyServer) handleHTTPSConnect(w http.ResponseWriter, r *http.Request) {
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
	go transfer(clientConn, destConn)
	go transfer(destConn, clientConn)
}

func transfer(dst io.WriteCloser, src io.ReadCloser) {
	defer func() {
		_ = dst.Close()
		_ = src.Close()
	}()
	_, _ = io.Copy(dst, src)
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
		jx.handleHTTPSConnect(w, r)
		return
	}

	// Handle HTTP requests as before
	jx.HandleProxyRequest(w, r)
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
