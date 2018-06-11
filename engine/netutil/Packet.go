package netutil

import (
	"encoding/binary"

	"unsafe"

	"sync/atomic"

	"sync"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/netutil/compress"
)

const (
	_MIN_PAYLOAD_CAP = 128
	_CAP_GROW_SHIFT  = uint(2)

	_PAYLOAD_LEN_MASK            = 0x7FFFFFFF
	_PAYLOAD_COMPRESSED_BIT_MASK = 0x80000000
)

var (
	packetEndian               = binary.LittleEndian
	predefinePayloadCapacities []uint32

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
				gwlog.Infof("DEBUG PACKETS: ALLOC=%d, RELEASE=%d, NEW=%d",
					atomic.LoadInt64(&debugInfo.AllocCount),
					atomic.LoadInt64(&debugInfo.ReleaseCount),
					atomic.LoadInt64(&debugInfo.NewCount))
			}
			return p
		},
	}
)

func init() {

	if _MIN_PAYLOAD_CAP >= consts.PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD {
		gwlog.Fatalf("_MIN_PAYLOAD_CAP should be smaller than PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD")
	}

	payloadCap := uint32(_MIN_PAYLOAD_CAP) << _CAP_GROW_SHIFT
	for payloadCap < _MAX_PAYLOAD_LENGTH {
		predefinePayloadCapacities = append(predefinePayloadCapacities, payloadCap)
		payloadCap <<= _CAP_GROW_SHIFT
	}
	predefinePayloadCapacities = append(predefinePayloadCapacities, _MAX_PAYLOAD_LENGTH)

	for _, payloadCap := range predefinePayloadCapacities {
		payloadCap := payloadCap
		packetBufferPools[payloadCap] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, _PREPAYLOAD_SIZE+payloadCap)
			},
		}
	}
}

func getPayloadCapOfPayloadLen(payloadLen uint32) uint32 {
	for _, payloadCap := range predefinePayloadCapacities {
		if payloadCap >= payloadLen {
			return payloadCap
		}
	}
	return _MAX_PAYLOAD_LENGTH
}

// Packet is a packet for sending data
type Packet struct {
	readCursor uint32

	notCompress  bool
	refcount     int64
	bytes        []byte
	initialBytes [_PREPAYLOAD_SIZE + _MIN_PAYLOAD_CAP]byte
}

func allocPacket() *Packet {
	pkt := packetPool.Get().(*Packet)
	pkt.refcount = 1

	if pkt.notCompress {
		gwlog.Panicf("notCompress should be false")
	}

	if consts.DEBUG_PACKET_ALLOC {
		atomic.AddInt64(&debugInfo.AllocCount, 1)
	}

	if pkt.GetPayloadLen() != 0 || pkt.isCompressed() {
		gwlog.Panicf("allocPacket: payload should be 0 not not compressed, but is %d, compressed %v", pkt.GetPayloadLen(), pkt.isCompressed())
	}

	return pkt
}

// NewPacket allocates a new packet
func NewPacket() *Packet {
	return allocPacket()
}

// SetNotCompress force the packet not to be compressed
func (p *Packet) SetNotCompress() {
	p.notCompress = true
}

func (p *Packet) AssureCapacity(need uint32) {
	requireCap := p.GetPayloadLen() + need
	oldCap := p.PayloadCap()

	if requireCap <= oldCap { // most case
		return
	}

	// try to find the proper capacity for the need bytes
	resizeToCap := getPayloadCapOfPayloadLen(requireCap)

	buffer := packetBufferPools[resizeToCap].Get().([]byte)
	if len(buffer) != int(resizeToCap+_SIZE_FIELD_SIZE) {
		gwlog.Panicf("buffer size should be %d, but is %d", resizeToCap, len(buffer))
	}
	copy(buffer, p.data())
	oldPayloadCap := p.PayloadCap()
	oldBytes := p.bytes
	p.bytes = buffer

	if oldPayloadCap > _MIN_PAYLOAD_CAP {
		// release old bytes
		packetBufferPools[oldPayloadCap].Put(oldBytes)
	}
}

// AddRefCount adds reference count of packet
func (p *Packet) AddRefCount(add int64) {
	atomic.AddInt64(&p.refcount, add)
}

// Payload returns the total payload of packet
func (p *Packet) Payload() []byte {
	return p.bytes[_PREPAYLOAD_SIZE : _PREPAYLOAD_SIZE+p.GetPayloadLen()]
}

// UnwrittenPayload returns the unwritten payload, which is the left payload capacity
func (p *Packet) UnwrittenPayload() []byte {
	payloadLen := p.GetPayloadLen()
	return p.bytes[_PREPAYLOAD_SIZE+payloadLen:]
}

func (p *Packet) TotalPayload() []byte {
	return p.bytes[_PREPAYLOAD_SIZE:]
}

// UnreadPayload returns the unread payload
func (p *Packet) UnreadPayload() []byte {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	payloadEnd := _PREPAYLOAD_SIZE + p.GetPayloadLen()
	return p.bytes[pos:payloadEnd]
}

// HasUnreadPayload returns if all payload is read
func (p *Packet) HasUnreadPayload() bool {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	plen := p.GetPayloadLen()
	return pos < plen
}

func (p *Packet) data() []byte {
	return p.bytes[0 : _PREPAYLOAD_SIZE+p.GetPayloadLen()]
}

// PayloadCap returns the current payload capacity
func (p *Packet) PayloadCap() uint32 {
	return uint32(len(p.bytes) - _PREPAYLOAD_SIZE)
}

// Release releases the packet to packet pool
func (p *Packet) Release() {
	refcount := atomic.AddInt64(&p.refcount, -1)

	if refcount == 0 {
		payloadCap := p.PayloadCap()
		if payloadCap > _MIN_PAYLOAD_CAP {
			buffer := p.bytes
			p.bytes = p.initialBytes[:]
			packetBufferPools[payloadCap].Put(buffer) // reclaim the buffer
		}

		p.readCursor = 0
		p.setPayloadLenCompressed(0, false)
		p.notCompress = false
		packetPool.Put(p)

		if consts.DEBUG_PACKET_ALLOC {
			atomic.AddInt64(&debugInfo.ReleaseCount, 1)
		}
	} else if refcount < 0 {
		gwlog.Panicf("releasing packet with refcount=%d", p.refcount)
	}
}

// ClearPayload clears packet payload
func (p *Packet) ClearPayload() {
	p.readCursor = 0
	p.SetPayloadLen(0)
}

// AppendByte appends one byte to the end of payload
func (p *Packet) AppendByte(b byte) {
	p.AssureCapacity(1)
	p.bytes[_PREPAYLOAD_SIZE+p.GetPayloadLen()] = b
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 1
}

// ReadOneByte reads one byte from the beginning
func (p *Packet) ReadOneByte() (v byte) {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	v = p.bytes[pos]
	p.readCursor += 1
	return
}

// AppendBool appends one byte 1/0 to the end of payload
func (p *Packet) AppendBool(b bool) {
	if b {
		p.AppendByte(1)
	} else {
		p.AppendByte(0)
	}
}

// ReadBool reads one byte 1/0 from the beginning of unread payload
func (p *Packet) ReadBool() (v bool) {
	return p.ReadOneByte() != 0
}

// AppendUint16 appends one uint16 to the end of payload
func (p *Packet) AppendUint16(v uint16) {
	p.AssureCapacity(2)
	payloadEnd := _PREPAYLOAD_SIZE + p.GetPayloadLen()
	packetEndian.PutUint16(p.bytes[payloadEnd:payloadEnd+2], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 2
}

// AppendUint32 appends one uint32 to the end of payload
func (p *Packet) AppendUint32(v uint32) {
	p.AssureCapacity(4)
	payloadEnd := _PREPAYLOAD_SIZE + p.GetPayloadLen()
	packetEndian.PutUint32(p.bytes[payloadEnd:payloadEnd+4], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 4
}

// PopUint32 pops one uint32 from the end of payload
func (p *Packet) PopUint32() (v uint32) {
	payloadEnd := _PREPAYLOAD_SIZE + p.GetPayloadLen()
	v = packetEndian.Uint32(p.bytes[payloadEnd-4 : payloadEnd])
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) -= 4
	return
}

// AppendUint64 appends one uint64 to the end of payload
func (p *Packet) AppendUint64(v uint64) {
	p.AssureCapacity(8)
	payloadEnd := _PREPAYLOAD_SIZE + p.GetPayloadLen()
	packetEndian.PutUint64(p.bytes[payloadEnd:payloadEnd+8], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += 8
}

// PackFloat32 packs float32 in specified byte order
func PackFloat32(order binary.ByteOrder, b []byte, f float32) {
	fi := *(*uint32)(unsafe.Pointer(&f)) // convert bits from float32 to uint32
	order.PutUint32(b, fi)
}

// UnpackFloat32 unpacks float32 in specified byte order
func UnpackFloat32(order binary.ByteOrder, b []byte) (f float32) {
	fi := order.Uint32(b)
	f = *(*float32)(unsafe.Pointer(&fi))
	return
}

// AppendFloat32 appends one float32 to the end of payload
func (p *Packet) AppendFloat32(f float32) {
	p.AppendUint32(*(*uint32)(unsafe.Pointer(&f)))
}

// ReadFloat32 reads one float32 from the beginning of unread payload
func (p *Packet) ReadFloat32() float32 {
	v := p.ReadUint32()
	return *(*float32)(unsafe.Pointer(&v))
}

// AppendFloat64 appends one float64 to the end of payload
func (p *Packet) AppendFloat64(f float64) {
	p.AppendUint64(*(*uint64)(unsafe.Pointer(&f)))
}

// ReadFloat64 reads one float64 from the beginning of unread payload
func (p *Packet) ReadFloat64() float64 {
	v := p.ReadUint64()
	return *(*float64)(unsafe.Pointer(&v))
}

// AppendBytes appends slice of bytes to the end of payload
func (p *Packet) AppendBytes(v []byte) {
	bytesLen := uint32(len(v))
	p.AssureCapacity(bytesLen)
	payloadEnd := _PREPAYLOAD_SIZE + p.GetPayloadLen()
	copy(p.bytes[payloadEnd:payloadEnd+bytesLen], v)
	*(*uint32)(unsafe.Pointer(&p.bytes[0])) += bytesLen
}

// AppendVarStr appends a varsize string to the end of payload
func (p *Packet) AppendVarStr(s string) {
	p.AppendVarBytes([]byte(s))
}

// AppendVarBytes appends varsize bytes to the end of payload
func (p *Packet) AppendVarBytes(v []byte) {
	p.AppendUint32(uint32(len(v)))
	p.AppendBytes(v)
}

// ReadUint16 reads one uint16 from the beginning of unread payload
func (p *Packet) ReadUint16() (v uint16) {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	v = packetEndian.Uint16(p.bytes[pos : pos+2])
	p.readCursor += 2
	return
}

// ReadUint32 reads one uint32 from the beginning of unread payload
func (p *Packet) ReadUint32() (v uint32) {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	v = packetEndian.Uint32(p.bytes[pos : pos+4])
	p.readCursor += 4
	return
}

// ReadUint64 reads one uint64 from the beginning of unread payload
func (p *Packet) ReadUint64() (v uint64) {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	v = packetEndian.Uint64(p.bytes[pos : pos+8])
	p.readCursor += 8
	return
}

// ReadBytes reads bytes from the beginning of unread payload
func (p *Packet) ReadBytes(size uint32) []byte {
	pos := p.readCursor + _PREPAYLOAD_SIZE
	if pos > uint32(len(p.bytes)) || pos+size > uint32(len(p.bytes)) {
		gwlog.Panicf("Packet %p bytes is %d, but reading %d+%d", p, len(p.bytes), pos, size)
	}

	bytes := p.bytes[pos : pos+size] // bytes are not copied
	p.readCursor += size
	return bytes
}

// AppendEntityID appends one Entity ID to the end of payload
func (p *Packet) AppendEntityID(id common.EntityID) {
	if len(id) != common.ENTITYID_LENGTH {
		gwlog.Panicf("AppendEntityID: invalid entity id: %s", id)
	}
	p.AppendBytes([]byte(id))
}

// ReadEntityID reads one EntityID from the beginning of unread  payload
func (p *Packet) ReadEntityID() common.EntityID {
	return common.EntityID(p.ReadBytes(common.ENTITYID_LENGTH))
}

// AppendClientID appends one Client ID to the end of payload
func (p *Packet) AppendClientID(id common.ClientID) {
	if len(id) != common.CLIENTID_LENGTH {
		gwlog.Panicf("AppendEntityID: invalid client id: %s", id)
	}
	p.AppendBytes([]byte(id))
}

// ReadClientID reads one ClientID from the beginning of unread  payload
func (p *Packet) ReadClientID() common.ClientID {
	return common.ClientID(p.ReadBytes(common.CLIENTID_LENGTH))
}

// ReadVarStr reads a varsize string from the beginning of unread  payload
func (p *Packet) ReadVarStr() string {
	b := p.ReadVarBytes()
	return string(b)
}

// ReadVarBytes reads a varsize slice of bytes from the beginning of unread  payload
func (p *Packet) ReadVarBytes() []byte {
	blen := p.ReadUint32()
	return p.ReadBytes(blen)
}

// AppendData appends one data of any type to the end of payload
func (p *Packet) AppendData(msg interface{}) {
	dataBytes, err := MSG_PACKER.PackMsg(msg, nil)
	if err != nil {
		gwlog.Panic(err)
	}

	p.AppendVarBytes(dataBytes)
}

// ReadData reads one data of any type from the beginning of unread payload
func (p *Packet) ReadData(msg interface{}) {
	b := p.ReadVarBytes()
	//gwlog.Infof("ReadData: %s", string(b))
	err := MSG_PACKER.UnpackMsg(b, msg)
	if err != nil {
		gwlog.Panic(err)
	}
}

// AppendArgs appends arguments to the end of payload one by one
func (p *Packet) AppendArgs(args []interface{}) {
	argCount := uint16(len(args))
	p.AppendUint16(argCount)

	for _, arg := range args {
		p.AppendData(arg)
	}
}

// ReadArgs reads a number of arguments from the beginning of unread payload
func (p *Packet) ReadArgs() [][]byte {
	argCount := p.ReadUint16()
	args := make([][]byte, argCount)
	var i uint16
	for i = 0; i < argCount; i++ {
		args[i] = p.ReadVarBytes() // just read bytes, but not parse it
	}
	return args
}

// AppendStringList appends a list of strings to the end of payload
func (p *Packet) AppendStringList(list []string) {
	p.AppendUint16(uint16(len(list)))
	for _, s := range list {
		p.AppendVarStr(s)
	}
}

// ReadStringList reads a list of strings from the beginning of unread payload
func (p *Packet) ReadStringList() []string {
	listlen := int(p.ReadUint16())
	list := make([]string, listlen)
	for i := 0; i < listlen; i++ {
		list[i] = p.ReadVarStr()
	}
	return list
}

func (p *Packet) AppendEntityIDSet(eids common.EntityIDSet) {
	p.AppendUint32(uint32(len(eids)))
	for eid := range eids {
		p.AppendEntityID(eid)
	}
}

func (p *Packet) ReadEntityIDSet() common.EntityIDSet {
	size := p.ReadUint32()
	eids := make(common.EntityIDSet, size)
	for i := uint32(0); i < size; i++ {
		eids.Add(p.ReadEntityID())
	}
	return eids
}

// GetPayloadLen returns the payload length
func (p *Packet) GetPayloadLen() uint32 {
	return *(*uint32)(unsafe.Pointer(&p.bytes[0])) & _PAYLOAD_LEN_MASK
}

// SetPayloadLen sets the payload l
func (p *Packet) SetPayloadLen(plen uint32) {
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = (*pplen & _PAYLOAD_COMPRESSED_BIT_MASK) | plen
}

func (p *Packet) setPayloadLenCompressed(plen uint32, compressed bool) {
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	if compressed {
		*pplen = _PAYLOAD_COMPRESSED_BIT_MASK | plen
	} else {
		*pplen = plen
	}
}

func (p *Packet) requireCompress() bool {
	return !p.notCompress && !p.isCompressed() && p.GetPayloadLen() >= consts.PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD
}

func (p *Packet) compress(compressor compress.Compressor) {
	if !p.requireCompress() {
		return
	}

	payloadCap := p.PayloadCap()
	compressedBuffer := packetBufferPools[payloadCap].Get().([]byte)
	//w := bytes.NewBuffer(compressedBuffer[_PREPAYLOAD_SIZE:_PREPAYLOAD_SIZE])
	//cw.Reset(w)

	oldPayload := p.Payload()
	oldPayloadLen := len(oldPayload)
	compressedPayload, err := compressor.Compress(oldPayload, compressedBuffer[_PREPAYLOAD_SIZE:_PREPAYLOAD_SIZE])
	if err != nil {
		gwlog.Panic(errors.Wrap(err, "compress failed"))
	}

	compressedPayloadLen := len(compressedPayload)

	//fmt.Printf("(%.1fKB=%.1f%%)", float64(oldPayloadLen)/1024.0, float64(compressedPayloadLen)*100.0/float64(oldPayloadLen))

	if compressedPayloadLen >= oldPayloadLen-4 { // leave 4 bytes for AppendUint32 in the last
		return // compress not useful enough, throw away
	}

	if &compressedPayload[0] != &compressedBuffer[_PREPAYLOAD_SIZE] {
		gwlog.Panicf("should equal")
	}

	// reclaim the old payload buffer and use new compressed buffer
	packetBufferPools[payloadCap].Put(p.bytes)
	p.bytes = compressedBuffer
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = _PAYLOAD_COMPRESSED_BIT_MASK | uint32(compressedPayloadLen)

	p.AppendUint32(uint32(oldPayloadLen)) // append the size of old payload to the end of packet
	return
}

func (p *Packet) decompress(compressor compress.Compressor) {
	if !p.isCompressed() {
		return
	}

	// pop the uncompressed payload len from payload
	uncompressedPayloadLen := p.PopUint32()

	oldPayloadCap := p.PayloadCap()
	oldPayload := p.Payload()
	uncompressedBuffer := packetBufferPools[getPayloadCapOfPayloadLen(uncompressedPayloadLen)].Get().([]byte)
	if err := compressor.Decompress(oldPayload, uncompressedBuffer[_PREPAYLOAD_SIZE:_PREPAYLOAD_SIZE+uncompressedPayloadLen]); err != nil {
		gwlog.Panic(errors.Wrap(err, "decompress failed"))
	}

	//gwlog.Infof("Compressed payload: %d, after decompress: %d", len(oldPayload), newPayloadLen)
	//gwlog.Infof("UNCOMPRESS: %v => %v", oldPayload, compressedPayload)
	if oldPayloadCap != _MIN_PAYLOAD_CAP {
		packetBufferPools[oldPayloadCap].Put(p.bytes)
	}

	p.bytes = uncompressedBuffer
	pplen := (*uint32)(unsafe.Pointer(&p.bytes[0]))
	*pplen = uncompressedPayloadLen
}

func (p *Packet) isCompressed() bool {
	return *(*uint32)(unsafe.Pointer(&p.bytes[0]))&_PAYLOAD_COMPRESSED_BIT_MASK != 0
}
