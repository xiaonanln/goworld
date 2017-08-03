package netutil

import (
	"net"

	"testing"

	"sync"

	"fmt"

	"time"

	"github.com/xiaonanln/goworld/engine/gwlog"
)

var (
	listener net.Listener
)

func init() {
	var err error
	listener, err = net.Listen("tcp", "0.0.0.0:13001")
	paniciferror(err)
	gwlog.Info("Listening ...")
}

func paniciferror(err error) {
	if err != nil {
		gwlog.Panicf("error: %s", err.Error())
	}
}

func TestCompressedConnection_Read(t *testing.T) {
	var wait sync.WaitGroup
	wait.Add(2)
	MSGCOUNT := 1000
	go func() {
		defer wait.Done()
		conn, err := net.Dial("tcp", "localhost:13001")
		println("dial", conn, err)
		paniciferror(err)
		cc := NewCompressedConnection(NetConnection{conn})
		println("NewCompressedConnection", cc)
		for i := 0; i < MSGCOUNT; i++ {
			cc.Write([]byte(fmt.Sprintf("test message %d", i)))
			cc.Flush()
		}
	}()
	go func() {
		defer wait.Done()

		conn, err := listener.Accept()
		println("accept", conn, err)
		paniciferror(err)
		cc := NewCompressedConnection(NetConnection{conn})
		println("NewCompressedConnection", cc)

		buf := make([]byte, 1024)
		for i := 0; i < MSGCOUNT; i++ {
			cc.SetReadDeadline(time.Now().Add(time.Millisecond))
			n, err := cc.Read(buf)
			if err != nil && IsTemporaryNetError(err) {
				continue
			}

			paniciferror(err)
			msg := string(buf[:n])
			if msg != fmt.Sprintf("test message %d", i) {
				t.Error("wrong message")
			}
		}
	}()

	wait.Wait()
}
