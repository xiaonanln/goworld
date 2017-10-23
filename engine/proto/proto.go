package proto

import (
	"unsafe"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// MsgType is the type of message types
type MsgType uint16

const (
	// MT_INVALID is the invalid message type
	MT_INVALID = iota
	// MT_SET_GAME_ID is a message type for game
	MT_SET_GAME_ID
	// MT_SET_GATE_ID is a message type for gate
	MT_SET_GATE_ID
	// MT_NOTIFY_CREATE_ENTITY is a message type for creating entities
	MT_NOTIFY_CREATE_ENTITY
	// MT_NOTIFY_DESTROY_ENTITY is a message type for destroying entities
	MT_NOTIFY_DESTROY_ENTITY
	// MT_DECLARE_SERVICE is a message type for declaring services
	MT_DECLARE_SERVICE
	// MT_UNDECLARE_SERVICE is a message type for undeclaring services
	MT_UNDECLARE_SERVICE
	// MT_CALL_ENTITY_METHOD is a message type for calling entity methods
	MT_CALL_ENTITY_METHOD
	// MT_CREATE_ENTITY_ANYWHERE is a message type for creating entities
	MT_CREATE_ENTITY_ANYWHERE
	// MT_LOAD_ENTITY_ANYWHERE is a message type loading entities
	MT_LOAD_ENTITY_ANYWHERE
	// MT_NOTIFY_CLIENT_CONNECTED is a message type for clients
	MT_NOTIFY_CLIENT_CONNECTED
	// MT_NOTIFY_CLIENT_DISCONNECTED is a message type for clients
	MT_NOTIFY_CLIENT_DISCONNECTED
	// MT_CALL_ENTITY_METHOD_FROM_CLIENT is a message type for clients
	MT_CALL_ENTITY_METHOD_FROM_CLIENT
	// MT_SYNC_POSITION_YAW_FROM_CLIENT is a message type for clients
	MT_SYNC_POSITION_YAW_FROM_CLIENT
	// MT_NOTIFY_ALL_GAMES_CONNECTED is a message type to notify all games connected
	MT_NOTIFY_ALL_GAMES_CONNECTED
	// MT_NOTIFY_GATE_DISCONNECTED is a message type to notify gate disconnected
	MT_NOTIFY_GATE_DISCONNECTED
	// MT_START_FREEZE_GAME is a message type for hot swapping
	MT_START_FREEZE_GAME
	// MT_START_FREEZE_GAME_ACK is a message type for hot swapping
	MT_START_FREEZE_GAME_ACK

	// Message types for migrating

	// MT_MIGRATE_REQUEST is a message type for entity migrations
	MT_MIGRATE_REQUEST
	// MT_REAL_MIGRATE is a message type for entity migrations
	MT_REAL_MIGRATE
)

const (
	// MT_GATE_SERVICE_MSG_TYPE_START is the first message types that should be handled by GateService
	MT_GATE_SERVICE_MSG_TYPE_START = 1000 + iota
	// MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START is the first message type that should be redirected to client proxy
	MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_START
	// MT_CREATE_ENTITY_ON_CLIENT message type
	MT_CREATE_ENTITY_ON_CLIENT
	// MT_DESTROY_ENTITY_ON_CLIENT message type
	MT_DESTROY_ENTITY_ON_CLIENT
	// MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT message type
	MT_NOTIFY_MAP_ATTR_CHANGE_ON_CLIENT
	// MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT message type
	MT_NOTIFY_MAP_ATTR_DEL_ON_CLIENT
	// MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT message type
	MT_NOTIFY_LIST_ATTR_CHANGE_ON_CLIENT
	// MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT message type
	MT_NOTIFY_LIST_ATTR_POP_ON_CLIENT
	// MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT message type
	MT_NOTIFY_LIST_ATTR_APPEND_ON_CLIENT
	// MT_CALL_ENTITY_METHOD_ON_CLIENT message type
	MT_CALL_ENTITY_METHOD_ON_CLIENT
	// MT_UPDATE_POSITION_ON_CLIENT message type
	MT_UPDATE_POSITION_ON_CLIENT
	// MT_UPDATE_YAW_ON_CLIENT message type
	MT_UPDATE_YAW_ON_CLIENT
	// MT_SET_CLIENTPROXY_FILTER_PROP message type
	MT_SET_CLIENTPROXY_FILTER_PROP
	// MT_CLEAR_CLIENTPROXY_FILTER_PROPS message type
	MT_CLEAR_CLIENTPROXY_FILTER_PROPS
	// MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP message type
	MT_REDIRECT_TO_GATEPROXY_MSG_TYPE_STOP = 1499
)

const (
	// MT_CALL_FILTERED_CLIENTS message type: messages to be processed by GateService from Dispatcher, but not redirected to clients
	MT_CALL_FILTERED_CLIENTS = 1501 + iota
	// MT_SYNC_POSITION_YAW_ON_CLIENTS message type
	MT_SYNC_POSITION_YAW_ON_CLIENTS
	// MT_GATE_SERVICE_MSG_TYPE_STOP message type
	MT_GATE_SERVICE_MSG_TYPE_STOP = 1999
)

// Messages types that is sent directly between Gate & Client
const (
	// MT_SET_CLIENT_CLIENTID message is sent to client to set its clientid
	MT_SET_CLIENT_CLIENTID = 2001 + iota
	MT_UDP_SYNC_CONN_NOTIFY_CLIENTID
	MT_UDP_SYNC_CONN_NOTIFY_CLIENTID_ACK
	// MT_HEARTBEAT_FROM_CLIENT is sent by client to notify the gate server that the client is alive
	MT_HEARTBEAT_FROM_CLIENT
)

const (
	// SYNC_INFO_SIZE_PER_ENTITY is the size of sync info per entity
	SYNC_INFO_SIZE_PER_ENTITY = 16
	UDP_SYNC_PACKET_SIZE      = common.ENTITYID_LENGTH + SYNC_INFO_SIZE_PER_ENTITY
)

// EntitySyncInfo defines fields of entity sync info
type EntitySyncInfo struct {
	X, Y, Z float32
	Yaw     float32
}

func init() {
	if unsafe.Sizeof(EntitySyncInfo{}) != SYNC_INFO_SIZE_PER_ENTITY {
		gwlog.Fatalf("Wrong type definition for EntitySyncInfo: size is %d, but should be %d", unsafe.Sizeof(EntitySyncInfo{}), SYNC_INFO_SIZE_PER_ENTITY)
	}
}
