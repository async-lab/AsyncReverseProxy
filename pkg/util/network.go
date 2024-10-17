package util

import (
	"errors"
	"io"
	"net"
)

func IsNetClose(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)
}
