package netutil

import (
	"net"
	"time"

	"fmt"

	"github.com/xiaonanln/goworld/engine/gwioutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type testEchoTcpServer struct {
}

func (ts *testEchoTcpServer) ServeTCPConnection(conn net.Conn) {
	buf := make([]byte, 1024*1024, 1024*1024)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			gwioutil.WriteAll(conn, buf[:n])
		}

		if err != nil {
			if gwioutil.IsTimeoutError(err) {
				continue
			} else {
				gwlog.Errorf("read error: %s", err.Error())
				break
			}
		}
	}
}

const PORT = 14572

func init() {
	go func() {
		ServeTCP(fmt.Sprintf("localhost:%d", PORT), &testEchoTcpServer{})
	}()
	time.Sleep(time.Millisecond * 200)
}
