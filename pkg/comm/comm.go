package comm

import (
	"encoding/binary"
	"io"
	"net"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

func SendPacket(conn net.Conn, p packet.IPacket) (int, error) {
	bytes, err := packet.ToNetPacket(p).Serialize()
	if err != nil {
		return 0, err
	}
	length := uint32(len(bytes))
	buffer := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buffer[:4], length)
	copy(buffer[4:], bytes)
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
	buf := make([]byte, 32*1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
