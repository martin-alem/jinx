package main

import (
	"flag"
	"fmt"
	"jinx/pkg/util"
	"log"
)

var serverMode util.ServerMode
var websiteRootDirectory string
var sslCertificateFile string
var sslPrivateKeyFile string
var routingRulesFilePath string
var hostnameBlacklist string
var backendServersConfigPath string
var loadBalancingAlgorithm util.LoadBalancerAlgo

var jinxFlags *flag.FlagSet

func init() {
	jinxFlags = flag.NewFlagSet("Jinx Flags", flag.ExitOnError)

	jinxFlags.StringVar((*string)(&serverMode), "mode", string(util.HTTP_SERVER), "Defines the server's operational mode (e.g., http, reverse_proxy, forward_proxy, load_balancer, ftp).")
	jinxFlags.StringVar(&websiteRootDirectory, "website-root-dir", "", "Sets the root directory path for hosting website files.")
	jinxFlags.StringVar(&sslCertificateFile, "cert-file", "", "Specifies the file path to the SSL certificate used for HTTPS connections.")
	jinxFlags.StringVar(&sslPrivateKeyFile, "key-file", "", "Specifies the file path to the SSL certificate's private key.")
	jinxFlags.StringVar(&routingRulesFilePath, "route-table", "", "Path to a JSON file defining routing rules for directing incoming requests to specific backend servers.")
	jinxFlags.StringVar(&hostnameBlacklist, "black-list", "", "Comma-separated list of hostnames to block access to (e.g., google.com, facebook.com).")
	jinxFlags.StringVar(&backendServersConfigPath, "server-pool-config", "", "Path to a JSON file listing backend servers and their configurations for load balancing.")
	jinxFlags.StringVar((*string)(&loadBalancingAlgorithm), "alg", string(util.ROUND_ROBIN), "Specifies the load balancing algorithm to use (e.g., ROUND_ROBIN, LEAST_CONNECTIONS).")
}

func HandleStart(args []string) {

	if err := jinxFlags.Parse(args); err != nil {
		log.Fatal(err)
	}

	switch serverMode {
	case util.HTTP_SERVER:
		fmt.Println("HTTP SERVER")
		break
	case util.REVERSE_PROXY:
		fmt.Println("Proxy")
		break
	case util.FORWARD_PROXY:
		fmt.Println("forward")
		break
	case util.LOAD_BALANCER:
		fmt.Println("load balancer")
		break
	case util.FTP_SERVER:
		fmt.Println("file server")
		break
	default:
		log.Fatalf("%s is not a valid server mod option. valid option includes: http, reverse_proxy, forward_proxy, load_balancer, ftp", serverMode)
	}
}

func HandleStop() {

}

func HandleRestart() {

}

func HandleDestroy() {

}
