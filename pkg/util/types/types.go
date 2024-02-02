package types

import (
	"net/http"
	"sync"
)

type JinxHttpServerConfig struct {
	IP          string
	Port        int
	LogRoot     string
	WebsiteRoot string
	CertFile    string
	KeyFile     string
}

type JinxReverseProxyServerConfig struct {
	IP         string
	Port       int
	LogRoot    string
	RouteTable RouteTable
	CertFile   string
	KeyFile    string
}

type JinxForwardProxyServerConfig struct {
	IP        string
	Port      int
	LogRoot   string
	BlackList []string
	CertFile  string
	KeyFile   string
}

type JinxLoadBalancingServerConfig struct {
	IP             string
	Port           int
	LogRoot        string
	CertFile       string
	KeyFile        string
	ServerPool     []UpStreamServer
	SSLTermination bool
	Algorithm      LoadBalancerAlgo
}

type JinxResourceResponse struct {
	Res      *http.Response
	Filename string
}

type LoadBalancingAlgorithm func([]UpStreamServer, int, *sync.Mutex) UpStreamServer

type UpStreamServer struct {
	IP       string
	Port     int
	Weight   int
	Location string
}

type ServerPoolConfig map[string]UpStreamServer

type ServerMode string

type LoadBalancerAlgo string

type RouteTable map[string]string
