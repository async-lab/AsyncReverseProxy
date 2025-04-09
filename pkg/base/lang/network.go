package lang

import (
	"errors"
	"io"
	"net"
)

func IsNetClose(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)
}

func IsNetLost(err error) bool {
	return IsNetClose(err) || errors.As(err, new(net.Error))
}

func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return true
	}
	return false
}
