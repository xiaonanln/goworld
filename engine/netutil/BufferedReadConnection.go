package netutil

import "bufio"

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
	brc.bufReader = bufio.NewReaderSize(conn, 8192*2)
	return brc
}

func (brc *BufferedReadConnection) Read(p []byte) (int, error) {
	return brc.bufReader.Read(p)
}
