package netutil

import (
	"fmt"
	"io"
	"net"

	"unsafe"

	"github.com/pkg/errors"
)

// IsConnectionError check if the error is a connection error (close)
func IsConnectionError(_err interface{}) bool {
	err, ok := _err.(error)
	if !ok {
		return false
	}

	err = errors.Cause(err)
	if err == io.EOF || err == io.ErrClosedPipe {
		return true
	}

	neterr, ok := err.(net.Error)
	if !ok {
		return false
	}
	if neterr.Timeout() {
		return false
	}

	return true
}

// ConnectTCP connects to host:port in TCP
func ConnectTCP(host string, port int) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	return conn, err
}

func PutFloat32(b []byte, f float32) {
	NETWORK_ENDIAN.PutUint32(b, *(*uint32)(unsafe.Pointer(&f)))
}
