package netutil

import (
	"encoding/binary"
	"log"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	PACKET_ENDIAN = binary.LittleEndian
)

type Packet struct {
	released   bool
	payloadLen uint32
	readCursor uint32
	bytes      [MAX_PACKET_SIZE]byte
}

func (p *Packet) Payload() []byte {
	return p.bytes[PREPAYLOAD_SIZE : PREPAYLOAD_SIZE+p.payloadLen]
}

func (p *Packet) FreePayload() []byte {
	payloadEnd := PREPAYLOAD_SIZE + p.payloadLen
	return p.bytes[payloadEnd:]
}

func (p *Packet) Release() {
	if p.released {
		gwlog.Panicf("packet must not be released multiple times!")
	}

	p.payloadLen = 0
	p.readCursor = 0
	p.released = true
	messagePool.Put(p)
}

func (p *Packet) ClearPayload() {
	p.readCursor = 0
	p.payloadLen = 0
}

func (p *Packet) AppendByte(b byte) {
	p.bytes[PREPAYLOAD_SIZE+p.payloadLen] = b
	p.payloadLen += 1
}

func (p *Packet) ReadByte() (v byte) {
	pos := p.readCursor + PREPAYLOAD_SIZE
	v = p.bytes[pos]
	p.readCursor += 1
	return
}

func (p *Packet) AppendBool(b bool) {
	if b {
		p.AppendByte(1)
	} else {
		p.AppendByte(0)
	}
}

func (p *Packet) ReadBool() (v bool) {
	return p.ReadByte() != 0
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

func (p *Packet) AppendVarStr(s string) {
	p.AppendVarBytes([]byte(s))
}

func (p *Packet) AppendVarBytes(v []byte) {
	p.AppendUint32(uint32(len(v)))
	p.AppendBytes(v)
}

func (p *Packet) ReadUint16() (v uint16) {
	pos := p.readCursor + PREPAYLOAD_SIZE
	v = PACKET_ENDIAN.Uint16(p.bytes[pos : pos+2])
	p.readCursor += 2
	return
}

func (p *Packet) ReadUint32() (v uint32) {
	pos := p.readCursor + PREPAYLOAD_SIZE
	v = PACKET_ENDIAN.Uint32(p.bytes[pos : pos+4])
	p.readCursor += 4
	return
}

func (p *Packet) ReadUint64() (v uint64) {
	pos := p.readCursor + PREPAYLOAD_SIZE
	v = PACKET_ENDIAN.Uint64(p.bytes[pos : pos+8])
	p.readCursor += 8
	return
}

func (p *Packet) ReadBytes(size uint32) []byte {
	pos := p.readCursor + PREPAYLOAD_SIZE
	bytes := p.bytes[pos : pos+size] // bytes are not copied
	p.readCursor += size
	return bytes
}

func (p *Packet) AppendEntityID(id common.EntityID) {
	p.AppendBytes([]byte(id))
}

func (p *Packet) ReadEntityID() common.EntityID {
	return common.EntityID(p.ReadBytes(common.ENTITYID_LENGTH))
}
func (p *Packet) AppendClientID(id common.ClientID) {
	p.AppendBytes([]byte(id))
}

func (p *Packet) ReadClientID() common.ClientID {
	return common.ClientID(p.ReadBytes(common.CLIENTID_LENGTH))
}

func (p *Packet) ReadVarStr() string {
	b := p.ReadVarBytes()
	return string(b)
}

func (p *Packet) ReadVarBytes() []byte {
	blen := p.ReadUint32()
	return p.ReadBytes(blen)
}

func (p *Packet) AppendMessage(msg interface{}) {
	freePayload := p.FreePayload()

	argsData, err := MSG_PACKER.PackMsg(msg, freePayload[4:4])
	if err != nil {
		gwlog.Panic(err)
	}
	argsDataLen := uint32(len(argsData))
	PACKET_ENDIAN.PutUint32(freePayload[:4], argsDataLen)
	p.SetPayloadLen(p.payloadLen + 4 + argsDataLen)
}

func (p *Packet) ReadMessage(msg interface{}) {
	b := p.ReadVarBytes()
	err := MSG_PACKER.UnpackMsg(b, msg)
	if err != nil {
		gwlog.Panic(err)
	}
}

func (p *Packet) GetPayloadLen() uint32 {
	return p.payloadLen
}

func (p *Packet) SetPayloadLen(plen uint32) {
	if plen > MAX_PAYLOAD_LENGTH {
		log.Panicf("payload length too long: %d", plen)
	}

	p.payloadLen = plen
}
