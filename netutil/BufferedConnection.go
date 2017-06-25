package netutil

import (
	"bytes"
	"sync"
	"time"
)

type BufferedConnection struct {
	Connection
	sync.Mutex
	writeBuffer *bytes.Buffer
	delay       time.Duration
	readBuffer  []byte
	unreadBytes []byte
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

func (bc *BufferedConnection) sendRoutine() {
	ticker := time.Tick(bc.delay)
	for {
		<-ticker
		bc.Lock()
		bc.writeBuffer.WriteTo(bc.Connection)
		bc.Unlock()
	}
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
