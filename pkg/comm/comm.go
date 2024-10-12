package comm

import (
	"context"
	"io"
	"net"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
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
	var ret error
	pattern.SelectContextAndChannel(
		ctx,
		make(chan error),
		func() { ret = nil },
		func(err error) bool { ret = err; return false },
		func(ch chan error) {
			for {
				_, err := io.Copy(dst, src)
				if err != nil {
					if ctx.Err() != nil {
						ch <- nil
						return
					}
					ch <- err
					return
				}
			}
		},
	)
	return ret
}

func Proxy(ctx context.Context, conn1 net.Conn, conn2 net.Conn) error {
	var ret error

	localCtx, cancel := context.WithCancel(context.Background())

	pattern.SelectContextAndChannel(
		ctx,
		make(chan error),
		func() { cancel(); ret = nil },
		func(err error) bool { cancel(); ret = err; return false },
		func(ch chan error) { ch <- CopyData(localCtx, conn1, conn2) },
		func(ch chan error) { ch <- CopyData(localCtx, conn2, conn1) },
	)

	return ret
}
