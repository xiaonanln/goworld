package netutil

import (
	"bufio"

	"github.com/xiaonanln/goworld/engine/consts"
)

// BufferedReadConnection provides buffered read to connections
type BufferedReadConnection struct {
	Connection
	bufReader *bufio.Reader
}

// NewBufferedReadConnection creates a new connection with buffered read based on underlying connection
func NewBufferedReadConnection(conn Connection) *BufferedReadConnection {
	brc := &BufferedReadConnection{
		Connection: conn,
	}
	brc.bufReader = bufio.NewReaderSize(conn, consts.BUFFERED_READ_BUFFSIZE)
	return brc
}

// Read
func (brc *BufferedReadConnection) Read(p []byte) (int, error) {
	return brc.bufReader.Read(p)
}
