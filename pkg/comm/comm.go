package comm

import (
	"encoding/binary"
	"io"

	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/packet"
)

// var logger = logging.GetLogger()

var commDataSize uint32 = 128 * 1024
var commBufSize uint32 = 4 + commDataSize

var commBufPool = concurrent.NewPool(func() *[]byte {
	buf := make([]byte, commBufSize)
	return &buf
})

func SendPacket(dst io.Writer, p packet.IPacket) (int, error) {
	netPkt, err := packet.ToNetPacket(p)
	if err != nil {
		return 0, err
	}
	data, err := netPkt.Serialize()
	if err != nil {
		return 0, err
	}

	bufPtr := commBufPool.Get()
	defer commBufPool.Put(bufPtr)
	buf := *bufPtr

	length := uint32(len(data))
	if length > commBufSize {
		return 0, io.ErrShortBuffer
	}

	buf = buf[:4+length]
	binary.BigEndian.PutUint32(buf[:4], length)
	copy(buf[4:], data)
	return dst.Write(buf)
}

func ReceivePacket(src io.Reader) (packet.IPacket, error) {
	bufPtr := commBufPool.Get()
	defer commBufPool.Put(bufPtr)
	buf := *bufPtr

	buf = buf[:4]
	_, err := io.ReadFull(src, buf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(buf)
	if length > commBufSize {
		return nil, io.ErrShortBuffer
	}

	buf = buf[:length]
	n, err := io.ReadFull(src, buf)
	if err != nil {
		return nil, err
	}
	netPacket, err := packet.Deserialize(buf[:n])
	if err != nil {
		return nil, err
	}
	p, err := packet.FromNetPacket(netPacket)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func ReadForBytes(src io.Reader) ([]byte, error) {
	bufPtr := commBufPool.Get()
	defer commBufPool.Put(bufPtr)
	buf := *bufPtr

	n, err := src.Read(buf[:commDataSize])
	if err != nil {
		return nil, err
	}
	res := make([]byte, n)
	copy(res, buf[:n])
	return res, nil
}
