package netutil

import (
	"fmt"
	"net"

	"encoding/binary"

	"github.com/xiaonanln/goSyncQueue"
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
	conn         Connection
	useSendQueue bool
	sendQueue    sync_queue.SyncQueue
}

func NewPacketConnection(conn Connection, useSendQueue bool) PacketConnection {
	pc := PacketConnection{
		conn:         conn,
		useSendQueue: useSendQueue,
	}
	if useSendQueue {
		pc.sendQueue = sync_queue.NewSyncQueue()
		go pc.sendRoutine()
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
		gwlog.Debug("%s SEND PACKET: %v", pc, packet.bytes[:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
	}
	if packet.refcount <= 0 {
		gwlog.Panicf("sending p with refcount=%d", packet.refcount)
	}
	if pc.useSendQueue {
		packet.AddRefCount(1) // will be released when pop from queue
		pc.sendQueue.Push(packet)
		sendQueueLen := pc.sendQueue.Len()
		if sendQueueLen >= 10000 && sendQueueLen%10000 == 0 {
			gwlog.Warn("%s: send queue length = %d", pc, sendQueueLen)
		}
		return nil
	} else {
		return WriteAll(pc.conn, packet.data())
	}
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
		// p size is too large
		// todo: reset the connection when p size is invalid
		return nil, fmt.Errorf("message p too large: %v", payloadLen)
	}

	packet := NewPacketWithPayloadLen(payloadLen)
	err = ReadAll(pc.conn, packet.bytes[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE+payloadLen]) // receive the p type and payload
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
//func (pc PacketConnectionWithQueue) PushPacket(p *Packet) {
//	pc.queue.Push(p)
//}
//
//func (pc PacketConnectionWithQueue) SendInstantPacket(p *Packet) error {
//	return pc.PacketConnection.SendPacket(p)
//}
//
//func (pc PacketConnectionWithQueue) SendPacket(p *Packet) {
//	gwlog.Panicf("DO NOT USE SendPacket")
//}
//
func (pc PacketConnection) sendRoutine() {
	for {
		packet := pc.sendQueue.Pop().(*Packet)
		WriteAll(pc.conn, packet.data())
		packet.Release()
	}
}
