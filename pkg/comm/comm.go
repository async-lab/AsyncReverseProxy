package comm

import (
	"encoding/binary"
	"io"

	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/packet"
)

// var logger = logging.GetLogger()

var commDataSize uint64 = 64 * 1024
var commHeaderSize uint64 = 8 + 1024
var commBufSize uint64 = commHeaderSize + commDataSize

var commBufPool = concurrent.NewPool(func() []byte {
	return make([]byte, int(commBufSize))
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
	buf := commBufPool.Get()
	defer commBufPool.Put(buf)

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
	buf := commBufPool.Get()
	defer commBufPool.Put(buf)

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
	buf := commBufPool.Get()
	defer commBufPool.Put(buf)

	n, err := src.Read(buf[:commDataSize])
	if err != nil {
		return nil, err
	}

	return append([]byte(nil), buf[:n]...), nil
}
