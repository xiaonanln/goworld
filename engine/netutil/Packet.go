package netutil

import (
	"encoding/binary"
	"github.com/xiaonanln/pktconn"

	"unsafe"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Packet is a packet for sending data
type Packet pktconn.Packet

// NewPacket allocates a new packet
func NewPacket() *Packet {
	return (*Packet)(pktconn.NewPacket())
}

// Payload returns the total payload of packet
func (p *Packet) Payload() []byte {
	return (*pktconn.Packet)(p).Payload()
}

// UnreadPayload returns the unread payload
func (p *Packet) UnreadPayload() []byte {
	return (*pktconn.Packet)(p).UnreadPayload()
}

// HasUnreadPayload returns if all payload is read
func (p *Packet) HasUnreadPayload() bool {
	return (*pktconn.Packet)(p).HasUnreadPayload()
}

// Release releases the packet to packet pool
func (p *Packet) Release() {
	(*pktconn.Packet)(p).Release()
}

// ClearPayload clears packet payload
func (p *Packet) ClearPayload() {
	(*pktconn.Packet)(p).ClearPayload()
}

// AppendByte appends one byte to the end of payload
func (p *Packet) AppendByte(b byte) {
	(*pktconn.Packet)(p).WriteOneByte(b)
}

// ReadOneByte reads one byte from the beginning
func (p *Packet) ReadOneByte() (v byte) {
	return (*pktconn.Packet)(p).ReadOneByte()
}

// AppendBool appends one byte 1/0 to the end of payload
func (p *Packet) AppendBool(b bool) {
	(*pktconn.Packet)(p).WriteBool(b)
}

// ReadBool reads one byte 1/0 from the beginning of unread payload
func (p *Packet) ReadBool() (v bool) {
	return (*pktconn.Packet)(p).ReadBool()
}

// AppendUint16 appends one uint16 to the end of payload
func (p *Packet) AppendUint16(v uint16) {
	(*pktconn.Packet)(p).WriteUint16(v)
}

// AppendUint32 appends one uint32 to the end of payload
func (p *Packet) AppendUint32(v uint32) {
	(*pktconn.Packet)(p).WriteUint32(v)
}

// AppendUint64 appends one uint64 to the end of payload
func (p *Packet) AppendUint64(v uint64) {
	(*pktconn.Packet)(p).WriteUint64(v)
}

// UnpackFloat32 unpacks float32 in specified byte order
func UnpackFloat32(order binary.ByteOrder, b []byte) (f float32) {
	fi := order.Uint32(b)
	f = *(*float32)(unsafe.Pointer(&fi))
	return
}

// AppendFloat32 appends one float32 to the end of payload
func (p *Packet) AppendFloat32(f float32) {
	(*pktconn.Packet)(p).WriteFloat32(f)
}

// ReadFloat32 reads one float32 from the beginning of unread payload
func (p *Packet) ReadFloat32() float32 {
	return (*pktconn.Packet)(p).ReadFloat32()
}

// AppendFloat64 appends one float64 to the end of payload
func (p *Packet) AppendFloat64(f float64) {
	(*pktconn.Packet)(p).WriteFloat64(f)
}

// ReadFloat64 reads one float64 from the beginning of unread payload
func (p *Packet) ReadFloat64() float64 {
	return (*pktconn.Packet)(p).ReadFloat64()
}

// AppendBytes appends slice of bytes to the end of payload
func (p *Packet) AppendBytes(v []byte) {
	(*pktconn.Packet)(p).WriteBytes(v)
}

// AppendVarStr appends a varsize string to the end of payload
func (p *Packet) AppendVarStr(s string) {
	(*pktconn.Packet)(p).WriteVarStrI(s)
}

// AppendVarBytes appends varsize bytes to the end of payload
func (p *Packet) AppendVarBytes(v []byte) {
	(*pktconn.Packet)(p).WriteVarBytesI(v)
}

// ReadUint16 reads one uint16 from the beginning of unread payload
func (p *Packet) ReadUint16() (v uint16) {
	return (*pktconn.Packet)(p).ReadUint16()
}

// ReadUint32 reads one uint32 from the beginning of unread payload
func (p *Packet) ReadUint32() (v uint32) {
	return (*pktconn.Packet)(p).ReadUint32()
}

// ReadUint64 reads one uint64 from the beginning of unread payload
func (p *Packet) ReadUint64() (v uint64) {
	return (*pktconn.Packet)(p).ReadUint64()
}

// ReadBytes reads bytes from the beginning of unread payload
func (p *Packet) ReadBytes(size uint32) []byte {
	return (*pktconn.Packet)(p).ReadBytes(int(size))
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

func (p *Packet) AppendMapStringString(m map[string]string) {
	p.AppendUint32(uint32(len(m)))
	for k, v := range m {
		p.AppendVarStr(k)
		p.AppendVarStr(v)
	}
}

func (p *Packet) ReadMapStringString() map[string]string {
	size := p.ReadUint32()
	m := make(map[string]string, size)
	for i := uint32(0); i < size; i++ {
		k := p.ReadVarStr()
		v := p.ReadVarStr()
		m[k] = v
	}
	return m
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
	return (*pktconn.Packet)(p).GetPayloadLen()
}

// SetPayloadLen sets the payload l
func (p *Packet) SetPayloadLen(plen uint32) {
	(*pktconn.Packet)(p).SetPayloadLen(plen)
}

func (p *Packet) Retain() {
	(*pktconn.Packet)(p).Retain()
}
