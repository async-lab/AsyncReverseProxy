package comm

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"club.asynclab/asrp/pkg/packet"
)

var bufSize = 4 + 128*1024
var reserved = 1024 // 读取时为头部和序列化预留的长度

var bufPool = &sync.Pool{
	New: func() interface{} {
		buf := make([]byte, bufSize)
		return &buf
	},
}

func SendPacket(conn net.Conn, p packet.IPacket) (int, error) {
	data, err := packet.ToNetPacket(p).Serialize()
	if err != nil {
		return 0, err
	}
	bufPtr := bufPool.Get().(*[]byte)
	defer bufPool.Put(bufPtr)
	buf := *bufPtr

	length := uint32(len(data))
	if int(length) > bufSize {
		return 0, io.ErrShortBuffer
	}

	buf = buf[:4+length]
	binary.BigEndian.PutUint32(buf[:4], length)
	copy(buf[4:], data)
	return conn.Write(buf)
}

func ReceivePacket(conn net.Conn) (packet.IPacket, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		return nil, err
	}
	bufPtr := bufPool.Get().(*[]byte)
	defer bufPool.Put(bufPtr)
	buf := *bufPtr

	length := binary.BigEndian.Uint32(lengthBuf)
	if int(length) > bufSize {
		return nil, io.ErrShortBuffer
	}

	buf = buf[:length]
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
	defer bufPool.Put(bufPtr)
	buf := *bufPtr

	n, err := conn.Read(buf[:reserved])
	if err != nil {
		return nil, err
	}
	res := make([]byte, n)
	copy(res, buf[:n])
	return res, nil
}
