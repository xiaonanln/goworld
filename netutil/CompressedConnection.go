package netutil

import (
	"compress/zlib"
	"io"

	"github.com/xiaonanln/goworld/gwlog"
)

type CompressedConnection struct {
	io.ReadCloser
	*zlib.Writer
	Connection
}

func NewCompressedConnection(conn Connection) *CompressedConnection {
	cc := &CompressedConnection{
		Writer:     zlib.NewWriter(conn),
		Connection: conn,
	}
	var err error
	cc.ReadCloser, err = zlib.NewReader(conn)
	if err != nil {
		gwlog.Panicf("zlib new reader failed: %s", err.Error())
	}

	return cc
}

func (cc *CompressedConnection) Write(p []byte) (int, error) {
	return cc.Writer.Write(p)
}

func (cc *CompressedConnection) Read(p []byte) (int, error) {
	return cc.ReadCloser.Read(p)
}

func (cc *CompressedConnection) Flush() error {
	err := cc.Writer.Flush()
	if err != nil {
		return err
	}
	return cc.Connection.Flush()
}

func (cc *CompressedConnection) Close() error {
	err := cc.Connection.Close()
	if err != nil {
		return err
	}
	return cc.ReadCloser.Close()
}
