package netutil

import (
	"encoding/binary"
	"log"

	"unsafe"

	"sync/atomic"

	"sync"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	PACKET_ENDIAN = binary.LittleEndian
)

var (
	debugInfo struct {
		NewCount     int64
		AllocCount   int64
		ReleaseCount int64
	}

	messagePool = sync.Pool{
		New: func() interface{} {
			p := &Packet{
				refcount: 0,
			}
			if consts.DEBUG_PACKET_ALLOC {
				atomic.AddInt64(&debugInfo.NewCount, 1)
				gwlog.Info("DEBUG PACKETS: ALLOC=%d, RELEASE=%d, NEW=%d",
					atomic.LoadInt64(&debugInfo.AllocCount),
					atomic.LoadInt64(&debugInfo.ReleaseCount),
					atomic.LoadInt64(&debugInfo.NewCount))
			}
			return p
		},
	}
)

type Packet struct {
	readCursor uint32

	refcount int64
	bytes    [MAX_PACKET_SIZE]byte
}

func allocPacket() *Packet {
	pkt := messagePool.Get().(*Packet)
	//gwlog.Debug("ALLOC %p", pkt)
	if pkt.refcount != 0 {
		gwlog.Panicf("packet must be released when allocated from pool, but refcount=%d", pkt.refcount)
	}
	pkt.refcount = 1
	if consts.DEBUG_PACKET_ALLOC {
		atomic.AddInt64(&debugInfo.AllocCount, 1)
	}
	return pkt
}

func (packet *Packet) AddRefCount(add int64) {
	atomic.AddInt64(&packet.refcount, add)
}

func (p *Packet) Payload() []byte {
	return p.bytes[PREPAYLOAD_SIZE : PREPAYLOAD_SIZE+p.GetPayloadLen()]
}

func (p *Packet) FreePayload() []byte {
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	return p.bytes[payloadEnd:]
}

func (p *Packet) Release() {
	refcount := atomic.AddInt64(&p.refcount, -1)
	if refcount == 0 {
		p.SetPayloadLen(0)
		p.readCursor = 0

		messagePool.Put(p)

		if consts.DEBUG_PACKET_ALLOC {
			atomic.AddInt64(&debugInfo.ReleaseCount, 1)
		}
	} else if refcount < 0 {
		gwlog.Panicf("releasing packet with refcount=%d", p.refcount)
	}
}

func (p *Packet) ClearPayload() {
	p.readCursor = 0
	p.SetPayloadLen(0)
}

func (p *Packet) AppendByte(b byte) {
	p.bytes[PREPAYLOAD_SIZE+p.GetPayloadLen()] = b
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 1
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

func (p *Packet) AppendUint16(v uint16) {
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	PACKET_ENDIAN.PutUint16(p.bytes[payloadEnd:payloadEnd+2], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 2
}

func (p *Packet) AppendUint32(v uint32) {
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	PACKET_ENDIAN.PutUint32(p.bytes[payloadEnd:payloadEnd+4], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 4
}

func (p *Packet) AppendUint64(v uint64) {
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	PACKET_ENDIAN.PutUint64(p.bytes[payloadEnd:payloadEnd+8], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 8
}

func (p *Packet) AppendBytes(v []byte) {
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	bytesLen := uint32(len(v))
	copy(p.bytes[payloadEnd:payloadEnd+bytesLen], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += bytesLen
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

func (p *Packet) AppendData(msg interface{}) {
	freePayload := p.FreePayload()

	argsData, err := MSG_PACKER.PackMsg(msg, freePayload[4:4])
	if err != nil {
		gwlog.Panic(err)
	}
	argsDataLen := uint32(len(argsData))
	PACKET_ENDIAN.PutUint32(freePayload[:4], argsDataLen)
	p.SetPayloadLen(p.GetPayloadLen() + 4 + argsDataLen)
}

func (p *Packet) ReadData(msg interface{}) {
	b := p.ReadVarBytes()
	err := MSG_PACKER.UnpackMsg(b, msg)
	if err != nil {
		gwlog.Panic(err)
	}
}

func (p *Packet) AppendStringList(list []string) {
	p.AppendUint16(uint16(len(list)))
	for _, s := range list {
		p.AppendVarStr(s)
	}
}

func (p *Packet) ReadStringList() []string {
	listlen := int(p.ReadUint16())
	list := make([]string, listlen)
	for i := 0; i < listlen; i++ {
		list[i] = p.ReadVarStr()
	}
	return list
}

func (p *Packet) GetPayloadLen() uint32 {
	return *(*uint32)(unsafe.Pointer(&p.bytes[0]))
}

func (p *Packet) SetPayloadLen(plen uint32) {
	if plen > MAX_PAYLOAD_LENGTH {
		log.Panicf("payload length too long: %d", plen)
	}

	*(*uint32)(unsafe.Pointer(&p.bytes[0])) = plen
}
