package server

import (
	"net"

	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

func handleConnection(conn *net.Conn) {

}

func RunServer(listenAddress string) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Error("Error starting server: ", err)
		return
	}
	defer listener.Close()

	logger.Info("Server started, listening on ", listenAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection: ", err)
			continue
		}
		go handleConnection(&conn)
	}
}
