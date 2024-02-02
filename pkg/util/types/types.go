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
	IP         string
	Port       int
	LogRoot    string
	CertFile   string
	KeyFile    string
	ServerPool []UpStreamServer
	Algorithm  LoadBalancerAlgo
}

type JinxResourceResponse struct {
	Res      *http.Response
	Filename string
}

type HttpServerConfig struct {
	Port           int
	IP             string
	CertFile       string
	KeyFile        string
	WebsiteRootDir string
}

type ReverseProxyConfig struct {
	Port         int
	IP           string
	CertFile     string
	KeyFile      string
	RoutingTable string
}

type ForwardProxyConfig struct {
	Port      int
	IP        string
	CertFile  string
	KeyFile   string
	BlackList string
}

type LoadBalancerConfig struct {
	Port                 int
	IP                   string
	CertFile             string
	KeyFile              string
	ServerPoolConfigPath string
	Algo                 LoadBalancerAlgo
}

type JinxServerConfiguration struct {
	Mode               ServerMode
	HttpServerConfig   HttpServerConfig
	ReverseProxyConfig ReverseProxyConfig
	ForwardProxyConfig ForwardProxyConfig
	LoadBalancerConfig LoadBalancerConfig
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
