package util

import (
	"io"
	"os"
	"syscall"
)

func IsConnectionClose(err error) bool {
	return err == io.EOF || err == syscall.ECONNRESET || err == os.ErrClosed
}
