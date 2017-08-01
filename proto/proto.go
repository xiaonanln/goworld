package proto

import (
	"unsafe"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
)

var (
	msgTypeToString = map[int]string{}
)

type MsgType_t uint16

const (
	MT_INVALID = iota
	// Server Messages
	MT_SET_GAME_ID
	MT_SET_GATE_ID
	MT_NOTIFY_CREATE_ENTITY
	MT_NOTIFY_DESTROY_ENTITY
	MT_DECLARE_SERVICE
	MT_UNDECLARE_SERVICE
	MT_CALL_ENTITY_METHOD
	MT_CREATE_ENTITY_ANYWHERE
	MT_LOAD_ENTITY_ANYWHERE
	MT_NOTIFY_CLIENT_CONNECTED
	MT_NOTIFY_CLIENT_DISCONNECTED
	MT_CALL_ENTITY_METHOD_FROM_CLIENT
	MT_SYNC_POSITION_YAW_FROM_CLIENT
	MT_NOTIFY_ALL_GAMES_CONNECTED
	MT_NOTIFY_GATE_DISCONNECTED

	MT_START_FREEZE_GAME
	MT_START_FREEZE_GAME_ACK

	// Message types for migrating
	MT_MIGRATE_REQUEST
	MT_REAL_MIGRATE
)

const ( // Message types that should be handled by GateService
	MT_GATE_SERVICE_MSG_TYPE_START          = 1000 + iota
	MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START // messages that should be redirected to client proxy

	MT_CREATE_ENTITY_ON_CLIENT
	MT_DESTROY_ENTITY_ON_CLIENT

	MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT
	MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT
	MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT
	MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT
	MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT

	MT_CALL_ENTITY_METHOD_ON_CLIENT
	MT_UPDATE_POSITION_ON_CLIENT
	MT_UPDATE_YAW_ON_CLIENT

	MT_SET_CLIENTPROXY_FILTER_PROP
	MT_CLEAR_CLIENTPROXY_FILTER_PROPS

	MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP

	MT_CALL_FILTERED_CLIENTS
	MT_SYNC_POSITION_YAW_ON_CLIENTS

	MT_GATE_SERVICE_MSG_TYPE_STOP
)

const (
	SYNC_INFO_SIZE_PER_ENTITY = 16
)

type EntitySyncInfo struct {
	X, Y, Z float32
	Yaw     float32
}

type EntitySyncInfoToClient struct {
	ClientID common.ClientID
	EntityID common.EntityID
	EntitySyncInfo
}

func init() {
	if unsafe.Sizeof(EntitySyncInfo{}) != SYNC_INFO_SIZE_PER_ENTITY {
		gwlog.Fatal("Wrong type defintion for EntitySyncInfo: size is %d, but should be %d", unsafe.Sizeof(EntitySyncInfo{}), SYNC_INFO_SIZE_PER_ENTITY)
	}
	if unsafe.Sizeof(EntitySyncInfoToClient{}) != SYNC_INFO_SIZE_PER_ENTITY+common.CLIENTID_LENGTH+common.ENTITYID_LENGTH {
		gwlog.Fatal("Wrong type defintion for EntitySyncInfoToClient: size is %d, but should be %d", unsafe.Sizeof(EntitySyncInfoToClient{}), SYNC_INFO_SIZE_PER_ENTITY+common.CLIENTID_LENGTH)
	}
}

//const ( // Message types that can be received from client
//
//)

func MsgTypeToString(msgType MsgType_t) string {
	return msgTypeToString[int(msgType)]
}
