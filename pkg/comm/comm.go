package comm

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"club.asynclab/asrp/pkg/packet"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 32*1024*1024)
		return &buf
	},
}

func SendPacket(conn net.Conn, p packet.IPacket) (int, error) {
	data, err := packet.ToNetPacket(p).Serialize()
	if err != nil {
		return 0, err
	}
	length := uint32(len(data))
	buffer := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buffer[:4], length)
	copy(buffer[4:], data)
	return conn.Write(buffer)
}

func ReceivePacket(conn net.Conn) (packet.IPacket, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	buf := make([]byte, length)
	n, err := io.ReadFull(conn, buf)
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

func ReadForBytes(conn net.Conn) ([]byte, error) {
	bufPtr := bufPool.Get().(*[]byte)
	buf := *bufPtr
	defer bufPool.Put(bufPtr)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	res := make([]byte, n)
	copy(res, buf[:n])
	return res, nil
}
