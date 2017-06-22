package netutil

import (
	"fmt"
	"net"

	"encoding/binary"

	"sync"

	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	MAX_PACKET_SIZE    = 1 * 1024 * 1024
	SIZE_FIELD_SIZE    = 4
	PREPAYLOAD_SIZE    = SIZE_FIELD_SIZE
	MAX_PAYLOAD_LENGTH = MAX_PACKET_SIZE - PREPAYLOAD_SIZE
)

var (
	NETWORK_ENDIAN = binary.LittleEndian
	messagePool    = sync.Pool{
		New: func() interface{} {
			return &Packet{
				refcount: 0,
			}
		},
	}
)

type PacketConnection struct {
	binconn BinaryConnection
}

func NewPacketConnection(conn net.Conn) PacketConnection {
	return PacketConnection{
		binconn: NewBinaryConnection(conn),
	}
}

func allocPacket() *Packet {
	pkt := messagePool.Get().(*Packet)
	//gwlog.Debug("ALLOC %p", pkt)
	if pkt.refcount != 0 {
		gwlog.Panicf("packet must be released when allocated from pool, but refcount=%d", pkt.refcount)
	}
	pkt.refcount = 1
	return pkt
}

func NewPacket() *Packet {
	return allocPacket()
}

func (pc PacketConnection) NewPacket() *Packet {
	return allocPacket()
}

func (pc PacketConnection) SendPacket(packet *Packet) error {
	if consts.DEBUG_PACKETS {
		gwlog.Debug("%s SEND PACKET: %v", pc, packet.bytes[:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
	}
	err := pc.binconn.SendAll(packet.bytes[:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
	return err
}

func (pc PacketConnection) RecvPacket() (*Packet, error) {
	packet := allocPacket()

	payloadLenBuf := packet.bytes[:SIZE_FIELD_SIZE]
	err := pc.binconn.RecvAll(payloadLenBuf)
	if err != nil {
		packet.Release()
		return nil, err
	}

	var payloadLen uint32 = NETWORK_ENDIAN.Uint32(payloadLenBuf)

	if payloadLen > MAX_PAYLOAD_LENGTH {
		// p size is too large
		packet.Release()
		return nil, fmt.Errorf("message packet too large: %v", payloadLen)
	}

	err = pc.binconn.RecvAll(packet.bytes[PREPAYLOAD_SIZE : PREPAYLOAD_SIZE+payloadLen]) // receive the packet type and payload
	if err != nil {
		packet.Release()
		return nil, err
	}

	//gwlog.Debug("<<< RecvMsg: payloadLen=%v, packet=%v", payloadLen, packet.bytes[:PREPAYLOAD_SIZE+payloadLen])
	return packet, nil
}

func (pc PacketConnection) Close() {
	pc.binconn.Close()
}

func (pc PacketConnection) RemoteAddr() net.Addr {
	return pc.binconn.RemoteAddr()
}

func (pc PacketConnection) LocalAddr() net.Addr {
	return pc.binconn.LocalAddr()
}

func (pc PacketConnection) String() string {
	return fmt.Sprintf("[%s >>> %s]", pc.LocalAddr(), pc.RemoteAddr())
}

//type PacketConnectionWithQueue struct {
//	PacketConnection
//	queue sync_queue.SyncQueue
//}
//
//func NewPacketConnectionWithQueue(conn net.Conn) PacketConnectionWithQueue {
//	pc := PacketConnectionWithQueue{
//		PacketConnection: NewPacketConnection(conn),
//		queue:            sync_queue.NewSyncQueue(),
//	}
//
//	go pc.sendRoutine()
//	return pc
//}
//
//func (pc PacketConnectionWithQueue) PushPacket(packet *Packet) {
//	pc.queue.Push(packet)
//}
//
//func (pc PacketConnectionWithQueue) SendInstantPacket(packet *Packet) error {
//	return pc.PacketConnection.SendPacket(packet)
//}
//
//func (pc PacketConnectionWithQueue) SendPacket(packet *Packet) {
//	gwlog.Panicf("DO NOT USE SendPacket")
//}
//
//func (pc PacketConnectionWithQueue) sendRoutine() {
//	for {
//		packet := pc.queue.Pop().(*Packet)
//		pc.PacketConnection.SendPacket(packet)
//	}
//}
