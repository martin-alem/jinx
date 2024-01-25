package main

import (
	"jinx/internal/http"
	"jinx/pkg/util"
)

func main() {

	config := util.JinxHttpServerConfig{
		Port:    80,
		LogRoot: "C:\\Users\\alema\\OneDrive\\Documents\\logs",
		Root:    "C:\\Users\\alema\\OneDrive\\Documents\\Website",
	}

	jinxHttpServer := http.NewJinxHttpServer(config)
	jinxHttpServer.Start()

}
