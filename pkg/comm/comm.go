package comm

import (
	"encoding/binary"
	"io"

	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/packet"
)

// var logger = logging.GetLogger()

var commDataSize uint64 = 64 * 1024
var commBufSize uint64 = 8 + 1024 + commDataSize

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

	length := uint64(len(data))
	if length > commBufSize {
		return 0, io.ErrShortBuffer
	}

	bufPtr := commBufPool.Get()
	defer commBufPool.Put(bufPtr)
	buf := *bufPtr

	buf = buf[:8+length]
	binary.BigEndian.PutUint64(buf[:8], length)
	copy(buf[8:], data)
	return dst.Write(buf)
}

func ReceivePacket(src io.Reader) (packet.IPacket, error) {
	lenBuf := make([]byte, 8)
	_, err := io.ReadFull(src, lenBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint64(lenBuf)
	if length > commBufSize {
		return nil, io.ErrShortBuffer
	}

	bufPtr := commBufPool.Get()
	defer commBufPool.Put(bufPtr)
	buf := *bufPtr

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
