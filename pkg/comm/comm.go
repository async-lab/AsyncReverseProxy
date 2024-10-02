package comm

import (
	"io"
	"net"

	"club.asynclab/asrp/pkg/api"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packets"
)

var logger = logging.GetLogger()

func SendPacketData(conn net.Conn, p api.IPacket) (int, error) {
	b, err := p.GetData().Serialize()
	if err != nil {
		return 0, err
	}
	return conn.Write(b)
}

func ReceivePacketData(conn net.Conn) (api.IPacket, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	d, err := api.Deserialize(buf[:n])
	if err != nil {
		return nil, err
	}
	p := packets.GetPacketFromData(d)
	return p, nil
}

func CopyData(dst io.Writer, src io.Reader) error {
	for {
		_, err := io.Copy(dst, src)
		if err != nil {
			return err
		}
	}
}

func Proxy(conn1 net.Conn, conn2 net.Conn) {
	errChan := make(chan error)
	go func() { errChan <- CopyData(conn1, conn2) }()
	go func() { errChan <- CopyData(conn2, conn1) }()
	logger.Error("Error copying data: ", <-errChan)
}
