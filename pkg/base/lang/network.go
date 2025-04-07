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
