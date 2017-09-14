package netutil

import (
	"bufio"

	"github.com/xiaonanln/goworld/engine/consts"
)

// BufferedWriteConnection provides buffered write to connections
type BufferedWriteConnection struct {
	Connection
	bufWriter *bufio.Writer
}

// NewBufferedWriteConnection creates a new connection with buffered write based on underlying connection
func NewBufferedWriteConnection(conn Connection) *BufferedWriteConnection {
	brc := &BufferedWriteConnection{
		Connection: conn,
	}
	brc.bufWriter = bufio.NewWriterSize(conn, consts.BUFFERED_WRITE_BUFFSIZE)
	return brc
}

func (brc *BufferedWriteConnection) Write(p []byte) (int, error) {
	return brc.bufWriter.Write(p)
}

func (brc *BufferedWriteConnection) Flush() error {
	err := brc.bufWriter.Flush()
	if err != nil {
		return err
	}
	return brc.Connection.Flush()
}
