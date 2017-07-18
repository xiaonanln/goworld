package netutil

import (
	"bytes"
	"compress/flate"
	"encoding/binary"

	"unsafe"

	"sync/atomic"

	"sync"

	"io"

	"fmt"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
)

const (
	MIN_PAYLOAD_CAP = 128
	CAP_GROW_SHIFT  = uint(2)

	PAYLOAD_LEN_MASK    = 0x7FFFFFFF
	COMPRESSED_BIT_MASK = 0x80000000
)

var (
	PACKET_ENDIAN                = binary.LittleEndian
	PREDEFINE_PAYLOAD_CAPACITIES []uint32

	debugInfo struct {
		NewCount     int64
		AllocCount   int64
		ReleaseCount int64
	}

	packetBufferPools = map[uint32]*sync.Pool{}
	packetPool        = sync.Pool{
		New: func() interface{} {
			p := &Packet{}
			p.bytes = p.initialBytes[:]

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

	if MIN_PAYLOAD_CAP >= consts.PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD {
		gwlog.Fatal("MIN_PAYLOAD_CAP should be smaller than PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD")
	}

	payloadCap := uint32(MIN_PAYLOAD_CAP) << CAP_GROW_SHIFT
	for payloadCap < MAX_PAYLOAD_LENGTH {
		PREDEFINE_PAYLOAD_CAPACITIES = append(PREDEFINE_PAYLOAD_CAPACITIES, payloadCap)
		payloadCap <<= CAP_GROW_SHIFT
	}
	PREDEFINE_PAYLOAD_CAPACITIES = append(PREDEFINE_PAYLOAD_CAPACITIES, MAX_PAYLOAD_LENGTH)

	for _, payloadCap := range PREDEFINE_PAYLOAD_CAPACITIES {
		payloadCap := payloadCap
		packetBufferPools[payloadCap] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, PREPAYLOAD_SIZE+payloadCap)
			},
		}
	}
}

func getPayloadCapOfPayloadLen(payloadLen uint32) uint32 {
	for _, payloadCap := range PREDEFINE_PAYLOAD_CAPACITIES {
		if payloadCap >= payloadLen {
			return payloadCap
		}
	}
	return MAX_PAYLOAD_LENGTH
}

type Packet struct {
	readCursor uint32

	refcount     int64
	bytes        []byte
	initialBytes [PREPAYLOAD_SIZE + MIN_PAYLOAD_CAP]byte
}

func allocPacket() *Packet {
	pkt := packetPool.Get().(*Packet)
	pkt.refcount = 1

	if consts.DEBUG_PACKET_ALLOC {
		atomic.AddInt64(&debugInfo.AllocCount, 1)
	}

	if pkt.GetPayloadLen() != 0 || pkt.isCompressed() {
		gwlog.Panicf("allocPacket: payload should be 0 not not compressed, but is %d, compressed %v", pkt.GetPayloadLen(), pkt.isCompressed())
	}

	return pkt
}

func NewPacket() *Packet {
	return allocPacket()
}

//func NewPacketWithPayloadLen(payloadLen uint32) *Packet {
//	allocPayloadCap := uint32(MAX_PAYLOAD_LENGTH)
//	for _, payloadCap := range PREDEFINE_PAYLOAD_CAPACITIES {
//		if payloadCap >= payloadLen {
//			allocPayloadCap = payloadCap
//			break
//		}
//	}
//
//	return allocPacket(allocPayloadCap)
//}

func (p *Packet) assureCapacity(need uint32) {
	requireCap := p.GetPayloadLen() + need
	oldCap := p.PayloadCap()

	if requireCap <= oldCap { // most case
		return
	}

	// try to find the proper capacity for the need bytes
	resizeToCap := getPayloadCapOfPayloadLen(requireCap)

	buffer := packetBufferPools[resizeToCap].Get().([]byte)
	if len(buffer) != int(resizeToCap+SIZE_FIELD_SIZE) {
		gwlog.Panicf("buffer size should be %d, but is %d", resizeToCap, len(buffer))
	}
	copy(buffer, p.data())
	oldPayloadCap := p.PayloadCap()
	oldBytes := p.bytes
	p.bytes = buffer

	if oldPayloadCap > MIN_PAYLOAD_CAP {
		// release old bytes
		packetBufferPools[oldPayloadCap].Put(oldBytes)
	}
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
		payloadCap := p.PayloadCap()
		if payloadCap > MIN_PAYLOAD_CAP {
			buffer := p.bytes
			p.bytes = p.initialBytes[:]
			packetBufferPools[payloadCap].Put(buffer) // reclaim the buffer
		}

		p.readCursor = 0
		p.setPayloadLenCompressed(0, false)
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
	p.assureCapacity(1)
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
	p.assureCapacity(2)
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	PACKET_ENDIAN.PutUint16(p.bytes[payloadEnd:payloadEnd+2], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 2
}

func (p *Packet) AppendUint32(v uint32) {
	p.assureCapacity(4)
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	PACKET_ENDIAN.PutUint32(p.bytes[payloadEnd:payloadEnd+4], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 4
}

func (p *Packet) PopUint32() (v uint32) {
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
	v = PACKET_ENDIAN.Uint32(p.bytes[payloadEnd-4 : payloadEnd])
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) -= 4
	return
}

func (p *Packet) AppendUint64(v uint64) {
	p.assureCapacity(8)
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
	bytesLen := uint32(len(v))
	p.assureCapacity(bytesLen)
	payloadEnd := PREPAYLOAD_SIZE + p.GetPayloadLen()
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
	if pos > uint32(len(p.bytes)) || pos+size > uint32(len(p.bytes)) {
		gwlog.Panicf("Packet %p bytes is %d, but reading %d+%d", p, len(p.bytes), pos, size)
	}

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
	dataBytes, err := MSG_PACKER.PackMsg(msg, nil)
	if err != nil {
		gwlog.Panic(err)
	}

	p.AppendVarBytes(dataBytes)
}

func (p *Packet) ReadData(msg interface{}) {
	b := p.ReadVarBytes()
	//gwlog.Info("ReadData: %s", string(b))
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
	return *(*uint32)(unsafe.Pointer(&p.bytes[0])) & PAYLOAD_LEN_MASK
}

func (p *Packet) SetPayloadLen(plen uint32) {
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = (*pplen & COMPRESSED_BIT_MASK) | plen
}

func (p *Packet) setPayloadLenCompressed(plen uint32, compressed bool) {
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	if compressed {
		*pplen = COMPRESSED_BIT_MASK | plen
	} else {
		*pplen = plen
	}
}

func (p *Packet) compress(cw *flate.Writer) {
	if p.isCompressed() {
		return
	}

	if p.GetPayloadLen() < consts.PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD {
		return // payload is too short, compress is ignored
	}

	payloadCap := p.PayloadCap()
	compressedBuffer := packetBufferPools[payloadCap].Get().([]byte)
	w := bytes.NewBuffer(compressedBuffer[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE])
	cw.Reset(w)

	oldPayload := p.Payload()
	oldPayloadLen := len(oldPayload)
	if err := WriteAll(cw, oldPayload); err != nil {
		gwlog.Panicf("compress error: %v", err)
	}

	if err := cw.Flush(); err != nil {
		gwlog.Panicf("compress error: %v", err)
	}

	compressedPayload := w.Bytes()
	compressedPayloadLen := len(compressedPayload)

	//gwlog.Info("COMPRESS %v => %v", oldPayload, compressedPayload)
	//gwlog.Info("Old payload len %d, compressed payload len %d", oldPayloadLen, compressedPayloadLen)
	fmt.Printf("(%.1fKB=%.1f%%)", float64(oldPayloadLen)/1024.0, float64(compressedPayloadLen)*100.0/float64(oldPayloadLen))

	if compressedPayloadLen >= oldPayloadLen-4 { // leave 4 bytes for AppendUint32 in the last
		return // compress not useful enough, throw away
	}

	if &compressedPayload[0] != &compressedBuffer[PREPAYLOAD_SIZE] {
		gwlog.Panicf("should equal")
	}

	// reclaim the old payload buffer and use new compressed buffer
	packetBufferPools[payloadCap].Put(p.bytes)
	p.bytes = compressedBuffer
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = COMPRESSED_BIT_MASK | uint32(compressedPayloadLen)

	p.AppendUint32(uint32(oldPayloadLen)) // append the size of old payload to the end of packet
	return
}

func (p *Packet) decompress(cr io.ReadCloser) {
	if !p.isCompressed() {
		return
	}

	// pop the uncompressed payload len from payload
	uncompressedPayloadLen := p.PopUint32()

	oldPayloadCap := p.PayloadCap()
	oldPayload := p.Payload()
	uncompressedBuffer := packetBufferPools[getPayloadCapOfPayloadLen(uncompressedPayloadLen)].Get().([]byte)
	cr.(flate.Resetter).Reset(bytes.NewReader(oldPayload), nil)

	//newPayloadLen, err := cr.Read(uncompressedBuffer[PREPAYLOAD_SIZE:])
	err := ReadAll(cr, uncompressedBuffer[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE+uncompressedPayloadLen])
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "decompress failed"))
	}
	//if err := cr.Close(); err != nil {
	//	gwlog.Panic(errors.Wrap(err, "close uncompressor failed"))
	//}

	//gwlog.Info("Compressed payload: %d, after decompress: %d", len(oldPayload), newPayloadLen)
	//gwlog.Info("UNCOMPRESS: %v => %v", oldPayload, compressedPayload)
	if oldPayloadCap != MIN_PAYLOAD_CAP {
		packetBufferPools[oldPayloadCap].Put(p.bytes)
	}

	p.bytes = uncompressedBuffer
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = uncompressedPayloadLen
}

func (p *Packet) isCompressed() bool {
	return *(*uint32)(unsafe.Pointer(&p.bytes[0]))&COMPRESSED_BIT_MASK != 0
}
