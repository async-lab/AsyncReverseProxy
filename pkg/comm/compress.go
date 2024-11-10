package comm

import (
	"bytes"
	"compress/flate"
)

func Compress(data []byte) []byte {
	var buf bytes.Buffer
	zw, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		panic(err)
	}

	_, err = zw.Write(data)
	if err != nil {
		panic(err)
	}
	err = zw.Close()
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func Decompress(data []byte) []byte {
	zr := flate.NewReader(bytes.NewReader(data))
	defer zr.Close()

	var out bytes.Buffer
	_, err := out.ReadFrom(zr)
	if err != nil {
		panic(err)
	}

	return out.Bytes()
}
