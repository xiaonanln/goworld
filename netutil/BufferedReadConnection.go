package netutil

import "github.com/xiaonanln/goworld/gwlog"

type BufferedReadConnection struct {
	Connection
	buffer   [8192 * 2]byte
	readpos  int
	writepos int
}

func NewBufferedReadConnection(conn Connection) *BufferedReadConnection {
	return &BufferedReadConnection{
		Connection: conn,
	}
}

func (brc *BufferedReadConnection) Read(p []byte) (int, error) {
	//gwlog.Info("BufferedReadConnection reading %d", len(p))
	totalN := 0
	if !brc.isBufferEmpty() {
		n := brc.readTo(p)
		totalN += n
		p = p[n:]
	}

	if len(p) == 0 {
		return totalN, nil
	}

	// need to read more data
	if !brc.isBufferEmpty() {
		gwlog.Panicf("buffer should be empty")
	}

	if len(p) >= len(brc.buffer) {
		// the upper-layer has a recv buffer which is even larger than mine, so just read to upper-layer buffer
		n, err := brc.Connection.Read(p)
		totalN += n
		return totalN, err
	}

	//start reading from sub connection
	var subN int
	var suberr error
	subN, suberr = brc.Connection.Read(brc.buffer[:])
	//if subN > 0 {
	//	gwlog.Info("READ TO BUFFER: %d", subN)
	//}
	brc.writepos = subN

	if subN != 0 {
		// read more from subn
		n := brc.readTo(p)
		totalN += n
	}

	if totalN > 0 {
		return totalN, nil
	} else {
		return 0, suberr
	}
}

func (brc *BufferedReadConnection) isBufferEmpty() bool {
	return brc.writepos == 0
}

func (brc *BufferedReadConnection) readTo(p []byte) int {
	n := copy(p, brc.buffer[brc.readpos:brc.writepos])
	brc.readpos += n
	if brc.readpos == brc.writepos {
		// all is read
		brc.reset()
	}

	return n
}

func (brc *BufferedReadConnection) reset() {
	brc.readpos = 0
	brc.writepos = 0
}
