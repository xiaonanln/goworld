package netutil

import (
	"github.com/xiaonanln/goworld/engine/compress/gwsnappy"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type CompressedConnection struct {
	Connection
	compressWriter *gwsnappy.Writer
	compressReader *gwsnappy.Reader
}

// NewCompressedConnection creates a new connection reads and writes compressed data upon underlying connection
func NewCompressedConnection(conn Connection) *CompressedConnection {
	cc := &CompressedConnection{
		Connection: conn,
	}
	cc.compressWriter = gwsnappy.NewWriter(conn)
	cc.compressReader = gwsnappy.NewReader(conn)
	return cc
}

func (cc *CompressedConnection) Read(p []byte) (int, error) {
	//cc.compressReader.Reset(cc.Connection)
	n, err := cc.compressReader.Read(p)
	//if err != nil {
	//	gwlog.Infof("Cleared error %v", err)
	//	cc.compressReader.ClearError()
	//	if n > 0 {
	//		gwlog.Fatalf("CompressedConnection: error %v occured and %d bytes are lost", err, n)
	//	}
	//}

	gwlog.Debugf("CompressedConnection: Read %d %v", n, err)
	return n, err
}

func (cc *CompressedConnection) Write(p []byte) (int, error) {
	n, err := cc.compressWriter.Write(p)
	gwlog.Debugf("CompressedConnection: Write %d/%d bytes, err=%v", len(p), n, err)
	return n, err
}

func (cc *CompressedConnection) Flush() error {
	err := cc.compressWriter.Flush()
	if err == nil {
		err = cc.Connection.Flush()
	}
	//cc.compressWriter.Reset(cc.Connection)
	return err
}
