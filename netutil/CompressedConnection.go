package netutil

import (
	"compress/zlib"
	"io"

	"github.com/xiaonanln/goworld/gwlog"
)

type CompressedConnection struct {
	io.Reader
	io.Writer
	Connection
}

func NewCompressedConnection(conn Connection) *CompressedConnection {
	cc := &CompressedConnection{
		Writer:     zlib.NewWriter(conn),
		Connection: conn,
	}
	var err error
	cc.Reader, err = zlib.NewReader(conn)
	if err != nil {
		gwlog.Panicf("zlib new reader failed: %s", err.Error())
	}

	return cc
}
