package netutil

import (
	"bufio"

	"github.com/xiaonanln/goworld/engine/consts"
)

// BufferedConnection provides buffered write to connections
type BufferedConnection struct {
	Connection
	bufReader *bufio.Reader
	bufWriter *bufio.Writer
}

// NewBufferedWriteConnection creates a new connection with buffered write based on underlying connection
func NewBufferedConnection(conn Connection) *BufferedConnection {
	brc := &BufferedConnection{
		Connection: conn,
	}
	brc.bufReader = bufio.NewReaderSize(conn, consts.BUFFERED_READ_BUFFSIZE)
	brc.bufWriter = bufio.NewWriterSize(conn, consts.BUFFERED_WRITE_BUFFSIZE)
	return brc
}

// Read
func (brc *BufferedConnection) Read(p []byte) (int, error) {
	return brc.bufReader.Read(p)
}

func (brc *BufferedConnection) Write(p []byte) (int, error) {
	return brc.bufWriter.Write(p)
}

func (brc *BufferedConnection) Flush() error {
	err := brc.bufWriter.Flush()
	if err != nil {
		return err
	}
	return brc.Connection.Flush()
}
