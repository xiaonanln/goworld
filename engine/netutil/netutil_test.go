package netutil

import (
	"net"
	"testing"

	"math/rand"

	"fmt"

	"github.com/xiaonanln/goworld/engine/gwlog"
)

type testEchoTcpServer struct {
}

func (ts *testEchoTcpServer) ServeTCPConnection(conn net.Conn) {
	buf := make([]byte, 1024*1024, 1024*1024)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			WriteAll(conn, buf[:n])
		}

		if err != nil {
			if IsTemporaryNetError(err) {
				continue
			} else {
				gwlog.Error("read error: %s", err.Error())
				break
			}
		}
	}
}

func TestPacketConnection(t *testing.T) {
	PORT := 4003
	go func() {
		ServeTCP(fmt.Sprintf("localhost:%d", PORT), &testEchoTcpServer{})
	}()

	for compressed := 0; compressed <= 1; compressed += 1 {

		_conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", PORT))
		if err != nil {
			t.Errorf("connect error: %s", err)
		}

		conn := NewPacketConnection(NetConnection{_conn}, compressed != 0)

		for i := 0; i < 10; i++ {
			var PAYLOAD_LEN uint32 = uint32(rand.Intn(4096 + 1))
			gwlog.Info("Testing with payload %v", PAYLOAD_LEN)

			packet := conn.NewPacket()
			for j := uint32(0); j < PAYLOAD_LEN; j++ {
				packet.AppendByte(byte(rand.Intn(256)))
			}
			if packet.GetPayloadLen() != PAYLOAD_LEN {
				t.Errorf("payload should be %d, but is %d", PAYLOAD_LEN, packet.GetPayloadLen())
			}
			conn.SendPacket(packet)
			conn.Flush()
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

}
