package msgbox

import (
	"fmt"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
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

type MsgboxService struct {
	entity.Entity
}

// OnInit initialize MsgboxService fields
func (mbs *MsgboxService) OnInit() {
}

// OnCreated is called when MsgboxService is created
func (mbs *MsgboxService) OnCreated() {
	mbs.DeclareService(ServiceName)
	mbs.Attrs.SetDefault("maxMsgId", 0)
}

// RegisterService registeres MsgboxService to goworld
func RegisterService() {
	goworld.RegisterEntity(ServiceName, &MsgboxService{}, false, false).DefineAttrs(map[string][]string{
		"maxMsgId": {"Persistent"},
	})
}

// Send requests MsgboxService to send a message to target entity
func (mbs *MsgboxService) Send(targetID common.EntityID, msg interface{}) {
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
	})
}

func (mbs *MsgboxService) Recv(targetID common.EntityID, beginMsgId int64) {
	beginKey := mbs.getMsgKey(targetID, beginMsgId)
	endKey := mbs.getMsgKey(targetID, _MaxMsgId)
	kvdb.GetRange(beginKey, endKey, func(items []kvdbtypes.KVItem, err error) {
		if err != nil {
			gwlog.Panic(err)
		}

	})
}

func (mbs *MsgboxService) getMsgKey(targetID common.EntityID, msgid int64) string {
	return fmt.Sprintf("__Msg_%s_%010d", targetID, msgid)
}

func (mbs *MsgboxService) getNextMsgId() int64 {
	id := mbs.GetInt("maxMsgId")
	id++
	mbs.Attrs.Set("maxMsgId", id)
	return id
}

// Msgbox is used to send messages among entities: e.x. Msgbox{&a.Entity}.Send(targetID, msg, callback)
type Msgbox struct {
	entity *entity.Entity
}

func (mb Msgbox) Send(targetID common.EntityID, msg interface{}) {
	mb.entity.CallService(ServiceName, "Send", targetID, msg)
}

func (mb Msgbox) Recv() {
	mb.entity.CallService(ServiceName, "Recv", mb.entity.ID, mb.getLastMsgId()+1)
}

func (mb Msgbox) getLastMsgId() int64 {
	return mb.entity.Attrs.GetInt(_LastMsgboxMsgIdAttrKey)
}

func (mb Msgbox) setLastMsgId(id int64) {
	mb.entity.Attrs.Set(_LastMsgboxMsgIdAttrKey, id)
}
