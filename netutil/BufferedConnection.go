package netutil

import (
	"bytes"
	"sync"
	"time"

	"fmt"
)

type BufferedConnection struct {
	Connection
	sync.Mutex
	writeBuffer *bytes.Buffer
	delay       time.Duration
	readBuffer  []byte
	unreadBytes []byte
	closed      bool
}

func NewBufferedConnection(conn Connection, delay time.Duration) *BufferedConnection {
	bc := &BufferedConnection{
		Connection:  conn,
		writeBuffer: bytes.NewBuffer([]byte{}),
		delay:       delay,
		readBuffer:  make([]byte, 4096),
		unreadBytes: nil,
	}
	go bc.sendRoutine()
	return bc
}

func (bc *BufferedConnection) String() string {
	return fmt.Sprintf("BufferedConnection<%s>", bc.Connection.RemoteAddr())
}

func (bc *BufferedConnection) sendRoutine() {
sendRoutineLoop:
	for !bc.closed {
		time.Sleep(bc.delay)

		bc.Lock()
		writableLen := bc.writeBuffer.Len()
		if writableLen == 0 {
			bc.Unlock()
			continue
		}

		writeBuffer := bc.writeBuffer
		bc.writeBuffer = bytes.NewBuffer([]byte{}) // replace the write buffer with a new empty one
		bc.Unlock()

		for { // send data in write buffer until it's empty
			_, err := writeBuffer.WriteTo(bc.Connection)
			if err != nil && !IsTemporaryNetError(err) {
				// got bad error, stop the send routine
				break sendRoutineLoop
			}

			if writeBuffer.Len() == 0 {
				break
			}
		}
	}
}

func (bc *BufferedConnection) Close() error {
	bc.closed = true
	return bc.Connection.Close()
}

func (bc *BufferedConnection) Write(p []byte) (n int, err error) {
	bc.Lock()
	n, err = bc.writeBuffer.Write(p)
	bc.Unlock()
	return
}

func (bc *BufferedConnection) Read(p []byte) (n int, err error) {
	//n, err = bc.Connection.Read(p)
	//return
	//
	if len(bc.unreadBytes) == 0 {
		n, err = bc.Connection.Read(bc.readBuffer)
		if n == 0 { // reads none, return immediately
			return
		}
		bc.unreadBytes = bc.readBuffer[:n]
	}

	n = copy(p, bc.unreadBytes)
	bc.unreadBytes = bc.unreadBytes[n:]
	return
}

//func (bc *BufferedConnection) Send(b []byte) error {
//	bc.writeBuffer.Write(b)
//	bc.Connection.Send(bc.writeBuffer.Bytes())
//}
//
//func (bc *BufferedConnection) Recv(b []byte) error {
//
//}
