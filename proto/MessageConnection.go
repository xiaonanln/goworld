package goworld_proto

import (
	"net"

	"encoding/binary"

	"fmt"

	"sync"

	"encoding/json"

	"github.com/xiaonanln/goworld/netutil"
	"github.com/xiaonanln/goworld/uuid"
	"github.com/xiaonanln/vacuum/vlog"
)

const (
	MAX_MESSAGE_SIZE = 1 * 1024 * 1024
	SIZE_FIELD_SIZE  = 4
	TYPE_FIELD_SIZE  = 2
	PREPAYLOAD_SIZE  = SIZE_FIELD_SIZE + TYPE_FIELD_SIZE

	STRING_ID_SIZE        = uuid.UUID_LENGTH
	RELAY_PREPAYLOAD_SIZE = SIZE_FIELD_SIZE + STRING_ID_SIZE + TYPE_FIELD_SIZE

	RELAY_MASK = 0x80000000
)

var (
	NETWORK_ENDIAN = binary.LittleEndian
	messagePool    = sync.Pool{
		New: newMessageInPool,
	}
)

func newMessageInPool() interface{} {
	return &Message{}
}

type MessageConnection struct {
	netutil.BinaryConnection
}

func NewMessageConnection(conn net.Conn) MessageConnection {
	return MessageConnection{BinaryConnection: netutil.NewBinaryConnection(conn)}
}

type Message [MAX_MESSAGE_SIZE]byte

func allocMessage() *Message {
	msg := messagePool.Get().(*Message)
	//vlog.Debug("ALLOC %p", msg)
	return msg
}

func (m *Message) Release() {
	//vlog.Debug("RELEASE %p", m)
	messagePool.Put(m)
}

func toJsonString(msg interface{}) string {
	s, _ := json.Marshal(msg)
	return string(s)
}

// Send msg to/from dispatcher
// Message format: [size*4B][type*2B][payload*NB]
func (mc *MessageConnection) SendMsg(mt MsgType_t, msg interface{}) error {
	return mc.SendMsgEx(mt, msg, MSG_PACKER)
}

func (mc *MessageConnection) SendMsgEx(mt MsgType_t, msg interface{}, msgPacker MsgPacker) error {
	msgbuf := allocMessage()
	defer msgbuf.Release()

	NETWORK_ENDIAN.PutUint16((msgbuf)[SIZE_FIELD_SIZE:SIZE_FIELD_SIZE+TYPE_FIELD_SIZE], uint16(mt))
	payloadBuf := (msgbuf)[PREPAYLOAD_SIZE:PREPAYLOAD_SIZE]
	payloadCap := cap(payloadBuf)
	payloadBuf, err := msgPacker.PackMsg(msg, payloadBuf)
	if err != nil {
		return err
	}

	payloadLen := len(payloadBuf)
	if payloadLen > payloadCap {
		// exceed payload
		return fmt.Errorf("MessageConnection: message paylaod too large(%d): %v", payloadLen, msg)
	}

	var pktSize uint32 = uint32(payloadLen + PREPAYLOAD_SIZE)
	NETWORK_ENDIAN.PutUint32((msgbuf)[:SIZE_FIELD_SIZE], pktSize)
	err = mc.SendAll((msgbuf)[:pktSize])
	vlog.Debug(">>> SendMsg: size=%v, %s%v, error=%v", pktSize, MsgTypeToString(mt), toJsonString(msg), err)
	return err
}

// Send msg to another String through dispatcher
// Message format: [size*4B][stringID][type*2B][payload*NB]
func (mc *MessageConnection) SendRelayMsg(targetID string, mt MsgType_t, msg interface{}) error {
	msgbuf := allocMessage()
	defer msgbuf.Release()
	copy(msgbuf[SIZE_FIELD_SIZE:SIZE_FIELD_SIZE+STRING_ID_SIZE], []byte(targetID))

	NETWORK_ENDIAN.PutUint16((msgbuf)[SIZE_FIELD_SIZE+STRING_ID_SIZE:SIZE_FIELD_SIZE+STRING_ID_SIZE+TYPE_FIELD_SIZE], uint16(mt))
	payloadBuf := (msgbuf)[RELAY_PREPAYLOAD_SIZE:RELAY_PREPAYLOAD_SIZE]
	payloadCap := cap(payloadBuf)
	payloadBuf, err := MSG_PACKER.PackMsg(msg, payloadBuf)
	if err != nil {
		return err
	}

	payloadLen := len(payloadBuf)
	if payloadLen > payloadCap {
		// exceed payload
		return fmt.Errorf("MessageConnection: message paylaod too large(%d): %v", payloadLen, msg)
	}

	var pktSize uint32 = uint32(payloadLen + RELAY_PREPAYLOAD_SIZE)
	NETWORK_ENDIAN.PutUint32((msgbuf)[:SIZE_FIELD_SIZE], pktSize|RELAY_MASK) // set highest bit of size to 1 to indicate a relay msg
	err = mc.SendAll((msgbuf)[:pktSize])
	vlog.Debug(">>> SendRelayMsg: size=%v, targetID=%s, type=%v: %v, error=%v", pktSize, targetID, mt, msg, err)
	return err
}

type MessageHandler interface {
	HandleMsg(msg *Message, pktSize uint32, msgType MsgType_t) error
	HandleRelayMsg(msg *Message, pktSize uint32, targetID string) error
}

func (mc *MessageConnection) RecvMsg(handler MessageHandler) error {
	msg := allocMessage()

	pktSizeBuf := msg[:SIZE_FIELD_SIZE]
	err := mc.RecvAll(pktSizeBuf)
	if err != nil {
		return err
	}

	var pktSize uint32 = NETWORK_ENDIAN.Uint32(pktSizeBuf)
	isRelayMsg := false

	if pktSize&RELAY_MASK != 0 {
		// this is a relay msg
		isRelayMsg = true
		pktSize -= RELAY_MASK
	}

	if pktSize > MAX_MESSAGE_SIZE {
		// pkt size is too large
		msg.Release()
		return fmt.Errorf("message packet too large: %v", pktSize)
	}

	err = mc.RecvAll((msg)[SIZE_FIELD_SIZE:pktSize]) // receive the msg type and payload
	if err != nil {
		msg.Release()
		return err
	}

	vlog.Debug("<<< RecvMsg: pktsize=%v, isRelayMsg=%v, packet=%v", pktSize, isRelayMsg, msg[:pktSize])
	//vlog.WithFields(vlog.Fields{"pktSize": pktSize, "isRelayMsg": isRelayMsg}).Debugf("RecvMsg")
	if isRelayMsg {
		// if it is a relay msg, we just relay what we receive without interpret the payload
		targetID := string(msg[SIZE_FIELD_SIZE : SIZE_FIELD_SIZE+STRING_ID_SIZE])
		err = handler.HandleRelayMsg(msg, pktSize, targetID)
	} else {
		var msgtype MsgType_t
		msgtype = MsgType_t(NETWORK_ENDIAN.Uint16((msg)[SIZE_FIELD_SIZE : SIZE_FIELD_SIZE+TYPE_FIELD_SIZE]))
		err = handler.HandleMsg(msg, pktSize, msgtype)
	}

	return err
}
