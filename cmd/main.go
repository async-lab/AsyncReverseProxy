package main

import (
	client "club.asynclab/asrp/cmd/client"
	server "club.asynclab/asrp/cmd/server"
	logging "club.asynclab/asrp/pkg/logging"
)

func main() {
	logging.GetLogger().Info("Hello, World!")
	go server.RunServer(":7000")
	go client.RunClient()
}
