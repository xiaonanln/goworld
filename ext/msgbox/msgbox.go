package msgbox

import (
	"fmt"

	"strconv"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
	"github.com/xiaonanln/goworld/engine/kvdb"
	"github.com/xiaonanln/goworld/engine/kvdb/types"
	"github.com/xiaonanln/goworld/engine/netutil"
)

// msgbox is a reliable message box for entities
// Entity can use Msgbox.Send to send message to other entity and use Msgbox.Recv to receive messages
// Message can always be received by target entity even when target entity is not online, in which case target entity
// will receive this message when it is loaded and calls Msgbox.Receive.

const (
	ServiceName             = "MsgboxService"
	_LastMsgboxMsgIdAttrKey = "_lastMsgboxMsgId"
	_MaxMsgId               = 9999999999
)

var (
	msgpacker = netutil.MessagePackMsgPacker{}
)

type Msg map[string]interface{}

type MsgboxService struct {
	entity.Entity
}

// OnInit initialize MsgboxService fields
func (mbs *MsgboxService) OnInit() {
}

// OnCreated is called when MsgboxService is created
func (mbs *MsgboxService) OnCreated() {
	mbs.DeclareService(ServiceName)
	mbs.Attrs.SetDefaultInt("maxMsgId", 0)
}

// RegisterService registeres MsgboxService to goworld
func RegisterService() {
	goworld.RegisterEntity(ServiceName, &MsgboxService{}, true, false).DefineAttrs(map[string][]string{
		"maxMsgId": {"Persistent"},
	})
}

// Send requests MsgboxService to send a message to target entity
func (mbs *MsgboxService) Send(targetID common.EntityID, msg Msg) {
	gwlog.Debugf("%s: Send %s => %T %v", mbs, targetID, msg, msg)
	msgid := mbs.getNextMsgId()
	msgkey := mbs.getMsgKey(targetID, msgid)
	msgBytes, err := msgpacker.PackMsg(msg, nil)
	if err != nil {
		gwlog.Panic(err)
	}

	kvdb.Put(msgkey, string(msgBytes), func(err error) {
		if err != nil {
			gwlog.Panic(err)
		}
		gwlog.Debugf("Msg is sent ok")
	})
}

func (mbs *MsgboxService) Recv(targetID common.EntityID, beginMsgId int64) {
	beginKey := mbs.getMsgKey(targetID, beginMsgId)
	endKey := mbs.getMsgKey(targetID, _MaxMsgId)
	kvdb.GetRange(beginKey, endKey, func(items []kvdbtypes.KVItem, err error) {
		if err != nil {
			gwlog.Panic(err)
		}

		if len(items) == 0 {
			gwlog.Debugf("found no msg")
			return
		}

		var msgs []Msg
		var endMsgId int64 = beginMsgId - 1
		for _, item := range items {
			_, msgid := mbs.parseMsgKey(item.Key)
			endMsgId = msgid

			msgBytes := []byte(item.Val)
			var msg Msg
			err := msgpacker.UnpackMsg(msgBytes, &msg)
			if err != nil {
				gwlog.Panic(err)
			}
			msgs = append(msgs, msg)

		}

		mbs.Call(targetID, "MsgboxOnRecvMsg", beginMsgId, endMsgId, msgs)
	})
}

func (mbs *MsgboxService) getMsgKey(targetID common.EntityID, msgid int64) string {
	return fmt.Sprintf("__Msg_%s_%010d", targetID, msgid)
}

func (mbs *MsgboxService) parseMsgKey(msgkey string) (common.EntityID, int64) {
	if msgkey[:6] != "__Msg_" {
		gwlog.Panicf("not a valid msg key: %s", msgkey)
	}

	targetID := common.EntityID(msgkey[6 : 6+common.ENTITYID_LENGTH])
	msgid, err := strconv.Atoi(msgkey[6+common.ENTITYID_LENGTH+1:])
	if err != nil {
		gwlog.Panic(err)
	}
	return targetID, int64(msgid)
}

func (mbs *MsgboxService) getNextMsgId() int64 {
	id := mbs.GetInt("maxMsgId")
	id++
	mbs.Attrs.SetInt("maxMsgId", id)
	return id
}

// Msgbox is used to send messages among entities: e.x. Msgbox{&a.Entity}.Send(targetID, msg, callback)
type Msgbox struct {
	entity.Component
	msghandler func(msg Msg)
}

func (mb *Msgbox) OnInit() {
	gwlog.Debugf("%s: initializing msgbox ...", mb.Entity)
	mb.Attrs.SetDefaultInt(_LastMsgboxMsgIdAttrKey, 0)
}

func (mb *Msgbox) Send(targetID common.EntityID, msg Msg) {
	mb.CallService(ServiceName, "Send", targetID, msg)
}

func (mb *Msgbox) Recv() {
	mb.CallService(ServiceName, "Recv", mb.ID, mb.getLastMsgId()+1)
}

func (mb *Msgbox) MsgboxOnRecvMsg(beginMsgId int64, endMsgId int64, msgs []Msg) {
	gwlog.Debugf("%s: MsgBox.OnRecvMsg: %d -> %d: msgs %v", mb.Entity, beginMsgId, endMsgId, msgs)
	mb.Attrs.SetInt(_LastMsgboxMsgIdAttrKey, endMsgId)
	for _, msg := range msgs {
		gwutils.RunPanicless(func() {
			mb.msghandler(msg)
		})
	}
}

func (mb *Msgbox) SetMsgHandler(handler func(msg Msg)) {
	mb.msghandler = handler
}

func (mb *Msgbox) getLastMsgId() int64 {
	return mb.Attrs.GetInt(_LastMsgboxMsgIdAttrKey)
}

func (mb *Msgbox) setLastMsgId(id int64) {
	mb.Attrs.SetInt(_LastMsgboxMsgIdAttrKey, id)
}
