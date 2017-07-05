package netutil

import (
	"fmt"
	"net"

	"encoding/binary"

	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

const ( // Three different level of packet size
	PACKET_SIZE_SMALL = 1024
	PACKET_SIZE_BIG   = 1024 * 64
	PACKET_SIZE_HUGE  = 1024 * 1024 * 4
)

const (
	MAX_PACKET_SIZE    = 4 * 1024 * 1024
	SIZE_FIELD_SIZE    = 4
	PREPAYLOAD_SIZE    = SIZE_FIELD_SIZE
	MAX_PAYLOAD_LENGTH = MAX_PACKET_SIZE - PREPAYLOAD_SIZE
)

var (
	NETWORK_ENDIAN = binary.LittleEndian
)

type PacketConnection struct {
	conn Connection
}

func NewPacketConnection(conn Connection) PacketConnection {
	pc := PacketConnection{
		conn: conn,
	}
	return pc
}

func NewPacketWithPayloadLen(payloadLen uint32) *Packet {
	return allocPacket(payloadLen)
}

func NewPacket() *Packet {
	return allocPacket(INITIAL_PACKET_CAPACITY)
}

func (pc PacketConnection) NewPacket() *Packet {
	return allocPacket(INITIAL_PACKET_CAPACITY)
}

func (pc PacketConnection) SendPacket(packet *Packet) error {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s SEND PACKET: msgtype=%v, payload=%v", pc, PACKET_ENDIAN.Uint16(packet.bytes[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE+2]),
			packet.bytes[PREPAYLOAD_SIZE+2:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
	}
	if packet.refcount <= 0 {
		gwlog.Panicf("sending packet with refcount=%d", packet.refcount)
	}
	return WriteAll(pc.conn, packet.data())
}

func (pc PacketConnection) RecvPacket() (*Packet, error) {
	var _payloadLenBuf [SIZE_FIELD_SIZE]byte
	payloadLenBuf := _payloadLenBuf[:]

	err := ReadAll(pc.conn, payloadLenBuf)
	if err != nil {
		return nil, err
	}

	var payloadLen uint32 = NETWORK_ENDIAN.Uint32(payloadLenBuf)

	if payloadLen > MAX_PAYLOAD_LENGTH {
		// packet size is too large
		// todo: reset the connection when packet size is invalid
		return nil, fmt.Errorf("message packet too large: %v", payloadLen)
	}

	//if payloadLen > 1024 {
	//	fmt.Printf("(%d)", payloadLen)
	//}

	packet := NewPacketWithPayloadLen(payloadLen)
	err = ReadAll(pc.conn, packet.bytes[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE+payloadLen]) // receive the packet type and payload
	if err != nil {
		packet.Release()
		return nil, err
	}

	//gwlog.Debug("<<< RecvMsg: payloadLen=%v, p=%v", payloadLen, p.bytes[:PREPAYLOAD_SIZE+payloadLen])
	packet.SetPayloadLen(payloadLen)
	return packet, nil
}

func (pc PacketConnection) Close() {
	pc.conn.Close()
}

func (pc PacketConnection) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}

func (pc PacketConnection) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

func (pc PacketConnection) String() string {
	return fmt.Sprintf("[%s >>> %s]", pc.LocalAddr(), pc.RemoteAddr())
}
