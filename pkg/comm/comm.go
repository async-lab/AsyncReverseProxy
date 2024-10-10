package comm

import (
	"context"
	"io"
	"net"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

func SendPacket(conn net.Conn, p packet.IPacket) (int, error) {
	b, err := packet.NewNetPacket(p).Serialize()
	if err != nil {
		return 0, err
	}
	return conn.Write(b)
}

func ReceivePacket(conn net.Conn) (packet.IPacket, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	netPacket, err := packet.Deserialize(buf[:n])
	if err != nil {
		return nil, err
	}
	p := packet.FromNetPacket(netPacket)
	return p, nil
}

func CopyData(ctx context.Context, dst io.Writer, src io.Reader) error {
	for {
		_, err := io.Copy(dst, src)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}

func Proxy(ctx context.Context, conn1 net.Conn, conn2 net.Conn) error {
	errChan := make(chan error)
	localCtx, cancel := context.WithCancel(context.Background())
	go func() { errChan <- CopyData(localCtx, conn1, conn2) }()
	go func() { errChan <- CopyData(localCtx, conn2, conn1) }()
	for {
		select {
		case <-ctx.Done():
			cancel()
			return nil
		case err := <-errChan:
			cancel()
			return err
		}
	}
}
