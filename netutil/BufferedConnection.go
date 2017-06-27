package netutil

import (
	"bytes"
	"sync"
	"time"

	"fmt"

	"github.com/xiaonanln/goworld/gwlog"
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
	for !bc.closed {
		time.Sleep(bc.delay)

		bc.Lock() // TODO: handle network error
		writableLen := bc.writeBuffer.Len()
		if writableLen == 0 {

			bc.Unlock()
			continue
		}

		n, err := bc.writeBuffer.WriteTo(bc.Connection)
		if int(n) < writableLen || err != nil {
			gwlog.Debug("%s: Write Buffer Write To: %d %v, writableLen=%v", bc, n, err, writableLen)
		}
		bc.Unlock()

		if err != nil && !IsTemporaryNetError(err) {
			// got bad error, stop the send routine
			gwlog.Error("%s send routine quit due to error: %s", bc, err)
			break
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
