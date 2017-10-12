package netutil

import (
	"net"
	"testing"
	"time"

	"math/rand"

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

//func TestCompressedConnection(t *testing.T) {
//
//	_conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", PORT))
//	if err != nil {
//		t.Errorf("connect error: %s", err)
//	}
//
//	conn := NewCompressedConnection(NetConnection{_conn})
//	s := ""
//	for i := 0; i < 1000; i++ {
//		s = s + "a"
//	}
//
//	SR := 1
//	for tm := 0; tm < 10; tm++ {
//		for i := 0; i < SR; i++ {
//			n, err := conn.Write([]byte(s))
//			if err != nil {
//				panic(err)
//			}
//			if n != len(s) {
//				panic("write wrong size")
//			}
//		}
//
//		println("written", SR)
//		conn.Flush()
//
//		recvbuf := make([]byte, len(s))
//		for len(recvbuf) > 0 {
//			var n int
//			var err error
//			if len(recvbuf) > 1000 {
//				n, err = conn.Read(recvbuf[:1000])
//			} else {
//				n, err = conn.Read(recvbuf)
//			}
//
//			recvbuf = recvbuf[n:]
//			fmt.Fprintf(os.Stderr, "Read %d/%d: %v\n", len(s)*SR-len(recvbuf), len(s)*SR, err)
//		}
//	}
//}

func TestPacketConnection(t *testing.T) {

	_conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", PORT))
	if err != nil {
		t.Errorf("connect error: %s", err)
	}

	conn := NewPacketConnection(NetConnection{_conn}, nil)

	for i := 0; i < 10; i++ {
		var PAYLOAD_LEN uint32 = uint32(rand.Intn(4096 + 1))
		gwlog.Infof("Testing with payload %v", PAYLOAD_LEN)

		packet := conn.NewPacket()
		for j := uint32(0); j < PAYLOAD_LEN; j++ {
			packet.AppendByte(byte(rand.Intn(256)))
		}
		if packet.GetPayloadLen() != PAYLOAD_LEN {
			t.Errorf("payload should be %d, but is %d", PAYLOAD_LEN, packet.GetPayloadLen())
		}
		conn.SendPacket(packet)
		conn.Flush("Test")
		recvPacket, err := conn.RecvPacket()
		if err != nil {
			t.Error(err)
		}
		if packet.GetPayloadLen() != recvPacket.GetPayloadLen() {
			t.Errorf("send packet len %d, but recv len %d", packet.GetPayloadLen(), recvPacket.GetPayloadLen())
		}
		for i := uint32(0); i < packet.GetPayloadLen(); i++ {
			if packet.Payload()[i] != recvPacket.Payload()[i] {
				t.Errorf("send packet and recv packet mismatch on byte index %d", i)
			}
		}
	}

}
