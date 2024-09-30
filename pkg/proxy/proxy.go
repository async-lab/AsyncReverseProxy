package proxy

import (
	"io"
	"log"
	"net"
)

func StartProxy(localAddr string, remoteAddr string) error {

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn, remoteAddr)
	}
}

func handleConnection(srcConn net.Conn, remoteAddr string) {
	defer srcConn.Close()

	// 连接到目标服务器
	destConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("Error connecting to remote server: %v", err)
		return
	}
	defer destConn.Close()

	// 使用 goroutine 进行双向转发
	go io.Copy(destConn, srcConn)
	io.Copy(srcConn, destConn)
}
