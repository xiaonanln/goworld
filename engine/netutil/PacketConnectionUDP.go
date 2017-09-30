package netutil

import (
	"net"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// PacketConnectionUDP is a packet connection upon a UDP connection
type PacketConnectionUDP struct {
	*net.UDPConn
}

var (
	errPacketTooLarge = errors.New("packet too large")
)

func NewPacketConnectionUDP(udpConn *net.UDPConn) PacketConnectionUDP {
	return PacketConnectionUDP{udpConn}
}

// SendPacket sends a packet through udp packet connection
func (pc PacketConnectionUDP) SendPacket(packet *Packet) error {
	payload := packet.Payload()
	n, err := pc.Write(payload) // only need to send payload
	if err == nil && n != len(payload) {
		err = errPacketTooLarge
	}
	return err
}

// SendPacketTo sends a packet through udp packet connection to target address
func (pc PacketConnectionUDP) SendPacketTo(packet *Packet, addr *net.UDPAddr) error {
	payload := packet.Payload()
	n, err := pc.WriteToUDP(payload, addr)
	if err == nil && n != len(payload) {
		err = errPacketTooLarge
	}
	return err
}

//
func (pc PacketConnectionUDP) RecvPacket() (*Packet, error) {
	packet := NewPacket()
	if packet.PayloadCap() < consts.UDP_MAX_PACKET_PAYLOAD_SIZE {
		gwlog.Fatalf("packet payload capacity(%d) should be larger than UDP_MAX_PACKET_PAYLOAD_SIZE(%d)", packet.PayloadCap(), consts.UDP_MAX_PACKET_PAYLOAD_SIZE)
	}

	totalPayload := packet.TotalPayload()
	n, err := pc.Read(totalPayload)

	if err == nil && n > len(totalPayload) {
		err = errPacketTooLarge
	}

	if err != nil {
		packet.Release()
		return nil, err
	}

	packet.SetPayloadLen(uint32(n))
	return packet, nil
}

func (pc PacketConnectionUDP) RecvPacketFrom() (*Packet, *net.UDPAddr, error) {
	packet := NewPacket()
	if packet.PayloadCap() < consts.UDP_MAX_PACKET_PAYLOAD_SIZE {
		gwlog.Fatalf("packet payload capacity(%d) should be larger than UDP_MAX_PACKET_PAYLOAD_SIZE(%d)", packet.PayloadCap(), consts.UDP_MAX_PACKET_PAYLOAD_SIZE)
	}

	totalPayload := packet.TotalPayload()
	n, srcAddr, err := pc.ReadFromUDP(totalPayload)

	if err == nil && n > len(totalPayload) {
		err = errPacketTooLarge
	}

	if err != nil {
		packet.Release()
		return nil, srcAddr, err
	}

	packet.SetPayloadLen(uint32(n))
	return packet, srcAddr, nil

}
