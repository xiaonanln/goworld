package netutil

import "github.com/golang/snappy"

type CompressedConnection struct {
	Connection
	compressWriter *snappy.Writer
	compressReader *snappy.Reader
}

// NewCompressedConnection creates a new connection reads and writes compressed data upon underlying connection
func NewCompressedConnection(conn Connection) *CompressedConnection {
	cc := &CompressedConnection{
		Connection: conn,
	}
	cc.compressWriter = snappy.NewWriter(conn)
	cc.compressReader = snappy.NewReader(conn)
	return cc
}

func (cc *CompressedConnection) Read(p []byte) (int, error) {
	cc.compressReader.Reset(cc.Connection)
	return cc.compressReader.Read(p)
}

func (cc *CompressedConnection) Write(p []byte) (int, error) {
	return cc.compressWriter.Write(p)
}

func (cc *CompressedConnection) Flush() error {
	err := cc.compressWriter.Flush()
	cc.compressWriter.Reset(cc.Connection)
	return err
}
