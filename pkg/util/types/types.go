package types

import "net/http"

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
}

type JinxResourceResponse struct {
	Res      *http.Response
	Filename string
}

type ServerMode string

type LoadBalancerAlgo string

type RouteTable map[string]string
