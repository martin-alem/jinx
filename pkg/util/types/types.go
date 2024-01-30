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

type JinxResourceResponse struct {
	Res      *http.Response
	Filename string
}

type ServerMode string

type LoadBalancerAlgo string
