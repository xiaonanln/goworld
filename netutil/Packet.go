package netutil

import (
	"log"

	"encoding/binary"
)

var (
	PACKET_ENDIAN = binary.LittleEndian
)

type Packet struct {
	payloadLen uint32
	bytes      [MAX_PACKET_SIZE]byte
}

func (p *Packet) Payload() []byte {
	return p.bytes[PREPAYLOAD_SIZE : PREPAYLOAD_SIZE+p.payloadLen]
}

func (p *Packet) Release() {
	messagePool.Put(p)
}

func (p *Packet) AppendByte(b byte) {
	p.bytes[PREPAYLOAD_SIZE+p.payloadLen] = b
	p.payloadLen += 1
}

func (p *Packet) prepareSend() {
	NETWORK_ENDIAN.PutUint32(p.bytes[:SIZE_FIELD_SIZE], p.payloadLen)
}

func (p *Packet) AppendUint16(v uint16) {
	payloadEnd := PREPAYLOAD_SIZE + p.payloadLen
	PACKET_ENDIAN.PutUint16(p.bytes[payloadEnd:payloadEnd+2], v)
	p.payloadLen += 2
}

func (p *Packet) AppendUint32(v uint32) {
	payloadEnd := PREPAYLOAD_SIZE + p.payloadLen
	PACKET_ENDIAN.PutUint32(p.bytes[payloadEnd:payloadEnd+4], v)
	p.payloadLen += 4
}

func (p *Packet) AppendUint64(v uint64) {
	payloadEnd := PREPAYLOAD_SIZE + p.payloadLen
	PACKET_ENDIAN.PutUint64(p.bytes[payloadEnd:payloadEnd+8], v)
	p.payloadLen += 8
}

func (p *Packet) AppendBytes(v []byte) {
	payloadEnd := PREPAYLOAD_SIZE + p.payloadLen
	bytesLen := uint32(len(v))
	copy(p.bytes[payloadEnd:payloadEnd+bytesLen], v)
	p.payloadLen += bytesLen
}

func (p *Packet) SetPayloadLen(plen uint32) {
	if plen > MAX_PAYLOAD_LENGTH {
		log.Panicf("payload length too long: %d", plen)
	}

	p.payloadLen = plen
}
