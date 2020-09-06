package netutil

import (
	"encoding/binary"
	"io"
	"net"

	"unsafe"

	"github.com/pkg/errors"
)

var (
	// NETWORK_ENDIAN is the network Endian of connections
	NETWORK_ENDIAN = binary.LittleEndian
)

// IsConnectionError check if the error is a connection error (close)
func IsConnectionError(_err interface{}) bool {
	err, ok := _err.(error)
	if !ok {
		return false
	}

	err = errors.Cause(err)
	if err == io.EOF {
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
func ConnectTCP(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	return conn, err
}

func PutFloat32(b []byte, f float32) {
	NETWORK_ENDIAN.PutUint32(b, *(*uint32)(unsafe.Pointer(&f)))
}
