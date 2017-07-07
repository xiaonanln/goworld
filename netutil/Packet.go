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

const (
	INITIAL_PACKET_CAPACITY = 128
)

var (
	PACKET_ENDIAN                = binary.LittleEndian
	MIN_PAYLOAD_CAP              = 128
	PREDEFINE_PAYLOAD_CAPACITIES []uint32

	debugInfo struct {
		NewCount     int64
		AllocCount   int64
		ReleaseCount int64
	}

	packetBufferPools = map[uint32]*sync.Pool{}
	packetPool        = sync.Pool{
		New: func() interface{} {
			p := &Packet{
				refcount: 0,
				bytes:    make([]byte, PREPAYLOAD_SIZE+MIN_PAYLOAD_CAP), // 4 for the uint32 payload len
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

func init() {
	payloadCap := uint32(MIN_PAYLOAD_CAP)
	for payloadCap < MAX_PAYLOAD_LENGTH {
		PREDEFINE_PAYLOAD_CAPACITIES = append(PREDEFINE_PAYLOAD_CAPACITIES, payloadCap)
		payloadCap <<= 2
	}
	PREDEFINE_PAYLOAD_CAPACITIES = append(PREDEFINE_PAYLOAD_CAPACITIES, MAX_PAYLOAD_LENGTH)
	gwlog.Info("Predefined payload caps: %v", PREDEFINE_PAYLOAD_CAPACITIES)

	for _, payloadCap := range PREDEFINE_PAYLOAD_CAPACITIES {
		payloadCap := payloadCap
		packetBufferPools[payloadCap] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, PREPAYLOAD_SIZE+payloadCap)
			},
		}
	}
}

type Packet struct {
	readCursor uint32

	refcount     int64
	bytes        []byte
	initialBytes [MIN_PAYLOAD_CAP]byte
}

func allocPacket(payloadCap uint32) *Packet {
	pkt := packetPool.Get().(*Packet)
	pkt.assureCapacity(payloadCap)
	pkt.refcount = 1

	if consts.DEBUG_PACKET_ALLOC {
		atomic.AddInt64(&debugInfo.AllocCount, 1)
	}
	return pkt
}

func NewPacketWithPayloadLen(payloadLen uint32) *Packet {
	allocPayloadCap := uint32(MAX_PAYLOAD_LENGTH)
	for _, payloadCap := range PREDEFINE_PAYLOAD_CAPACITIES {
		if payloadCap >= payloadLen {
			allocPayloadCap = payloadCap
			break
		}
	}
	return allocPacket(allocPayloadCap)
}

func (p *Packet) assureCapacity(need uint32) {
	requireCap := PREPAYLOAD_SIZE + p.GetPayloadLen() + need
	curcap := uint32(len(p.bytes))

	if requireCap <= curcap {
		return
	}

	resizeToCap := curcap << 1
	for resizeToCap < requireCap {
		resizeToCap <<= 1
	}

	newbytes := make([]byte, resizeToCap)
	copy(newbytes, p.data())
	p.bytes = newbytes
}

func (p *Packet) AddRefCount(add int64) {
	atomic.AddInt64(&p.refcount, add)
}

func (p *Packet) Payload() []byte {
	return p.bytes[PREPAYLOAD_SIZE : PREPAYLOAD_SIZE+p.GetPayloadLen()]
}

func (p *Packet) data() []byte {
	return p.bytes[0 : PREPAYLOAD_SIZE+p.GetPayloadLen()]
}

//func (p *Packet) FreePayload() []byte {
//	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
//	return p.bytes[payloadEnd:]
//}

func (p *Packet) PayloadCap() uint32 {
	return uint32(len(p.bytes) - PREPAYLOAD_SIZE)
}

func (p *Packet) Release() {

	refcount := atomic.AddInt64(&p.refcount, -1)

	if refcount == 0 {
		p.SetPayloadLen(0)
		p.readCursor = 0

		payloadCap := p.PayloadCap()
		packetPool := packetBufferPools[payloadCap]
		if packetPool == nil {
			gwlog.Panicf("payload cap is not valid: %v", payloadCap)
		}
		packetPool.Put(p)

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

func (p *Packet) AppendFloat32(f float32) {
	p.AppendUint32(*(*uint32)(unsafe.Pointer(&f)))
}

func (p *Packet) ReadFloat32() float32 {
	v := p.ReadUint32()
	return *(*float32)(unsafe.Pointer(&v))
}

func (p *Packet) AppendFloat64(f float64) {
	p.AppendUint64(*(*uint64)(unsafe.Pointer(&f)))
}

func (p *Packet) ReadFloat64() float64 {
	v := p.ReadUint64()
	return *(*float64)(unsafe.Pointer(&v))
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
	oldPayloadLen := p.GetPayloadLen() // save payload len before pack msg
	p.AppendUint32(0)                  // uint32 for data length

	newData, err := MSG_PACKER.PackMsg(msg, p.data())
	if err != nil {
		gwlog.Panic(err)
	}

	if len(newData) > len(p.bytes) { // data overflow!
		gwlog.Panicf("AppendData failed: overflow")
	}
	newPayloadLen := uint32(len(newData) - PREPAYLOAD_SIZE)
	dataLen := newPayloadLen - oldPayloadLen - 4
	PACKET_ENDIAN.PutUint32(p.bytes[PREPAYLOAD_SIZE+oldPayloadLen:PREPAYLOAD_SIZE+oldPayloadLen+4], dataLen)
	p.SetPayloadLen(newPayloadLen)
}

func (p *Packet) ReadData(msg interface{}) {
	b := p.ReadVarBytes()
	err := MSG_PACKER.UnpackMsg(b, msg)
	if err != nil {
		gwlog.Panic(err)
	}
}

// Append arguments to packet one by one
func (p *Packet) AppendArgs(args []interface{}) {
	argCount := uint16(len(args))
	p.AppendUint16(argCount)

	for _, arg := range args {
		p.AppendData(arg)
	}
}

func (p *Packet) ReadArgs() [][]byte {
	argCount := p.ReadUint16()
	args := make([][]byte, argCount)
	var i uint16
	for i = 0; i < argCount; i++ {
		args[i] = p.ReadVarBytes() // just read bytes, but not parse it
	}
	return args
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
