package netutil

import (
	"compress/flate"
	"io"

	"github.com/xiaonanln/goworld/gwlog"
)

type CompressedConnection struct {
	zipReadCloser io.ReadCloser
	zipWriter     *flate.Writer
	Connection
}

func NewCompressedConnection(conn Connection) *CompressedConnection {
	cc := &CompressedConnection{
		Connection: conn,
	}
	var err error

	cc.zipWriter, err = flate.NewWriter(conn, flate.BestSpeed)
	if err != nil {
		gwlog.Panicf("flate new writer failed: %s", err.Error())
	}
	//cc.zipWriter.Flush()
	cc.zipReadCloser = flate.NewReader(conn)
	//if err != nil {
	//	gwlog.Panicf("flate new reader failed: %s", err.Error())
	//}

	return cc
}

func (cc *CompressedConnection) Write(p []byte) (int, error) {
	n, err := cc.zipWriter.Write(p)
	if n > 0 {
		gwlog.Info("compress write %d %v", n, err)
	}

	return n, err
}

func (cc *CompressedConnection) Read(p []byte) (int, error) {
	n, err := cc.zipReadCloser.Read(p)
	if n > 0 {
		gwlog.Info("compress read %d %v", n, err)
	}
	return n, err
}

func (cc *CompressedConnection) Flush() error {
	gwlog.Info("compress flush")
	err := cc.zipWriter.Flush()
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
	return cc.zipReadCloser.Close()
}
