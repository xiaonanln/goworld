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
	PACKET_SIZE_TINY   = 1024
	PACKET_SIZE_NORMAL = 1024 * 64
	PACKET_SIZE_HUGE   = 1024 * 1024 * 4
)

const (
	MAX_PACKET_SIZE    = 4 * 1024
	SIZE_FIELD_SIZE    = 4
	PREPAYLOAD_SIZE    = SIZE_FIELD_SIZE
	MAX_PAYLOAD_LENGTH = MAX_PACKET_SIZE - PREPAYLOAD_SIZE
)

var (
	NETWORK_ENDIAN = binary.LittleEndian
)

type PacketConnection struct {
	binconn      BinaryConnection
	useSendQueue bool
	sendQueue    sync_queue.SyncQueue
}

func NewPacketConnection(conn net.Conn, useSendQueue bool) PacketConnection {
	pc := PacketConnection{
		binconn:      NewBinaryConnection(conn),
		useSendQueue: useSendQueue,
	}
	if useSendQueue {
		pc.sendQueue = sync_queue.NewSyncQueue()
		go pc.sendRoutine()
	}
	return pc
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
	if packet.refcount <= 0 {
		gwlog.Panicf("sending packet with refcount=%d", packet.refcount)
	}
	if pc.useSendQueue {
		packet.AddRefCount(1) // will be released when pop from queue
		pc.sendQueue.Push(packet)
		sendQueueLen := pc.sendQueue.Len()
		if sendQueueLen >= 1000 && sendQueueLen%1000 == 0 {
			gwlog.Warn("%s: send queue length = %d", pc, pc.sendQueue.Len())
		}
		return nil
	} else {
		err := pc.binconn.SendAll(packet.bytes[:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
		return err
	}
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
func (pc PacketConnection) sendRoutine() {
	for {
		packet := pc.sendQueue.Pop().(*Packet)
		pc.binconn.SendAll(packet.bytes[:PREPAYLOAD_SIZE+packet.GetPayloadLen()])
		packet.Release()
	}
}
