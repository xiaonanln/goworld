package proto

var (
	msgTypeToString = map[int]string{}
)

type MsgType_t uint16

const (
	MT_INVALID = iota
	// Server Messages
	MT_SET_SERVER_ID                  = iota
	MT_NOTIFY_CREATE_ENTITY           = iota
	MT_NOTIFY_DESTROY_ENTITY          = iota
	MT_DECLARE_SERVICE                = iota
	MT_UNDECLARE_SERVICE              = iota
	MT_CALL_ENTITY_METHOD             = iota
	MT_CREATE_ENTITY_ANYWHERE         = iota
	MT_LOAD_ENTITY_ANYWHERE           = iota
	MT_NOTIFY_CLIENT_CONNECTED        = iota
	MT_NOTIFY_CLIENT_DISCONNECTED     = iota
	MT_CALL_ENTITY_METHOD_FROM_CLIENT = iota
	MT_NOTIFY_ALL_SERVERS_CONNECTED   = iota

	// Message types for migrating
	MT_MIGRATE_REQUEST = iota
	MT_REAL_MIGRATE    = iota
)

const ( // Message types that should be handled by GateService
	MT_GATE_SERVICE_MSG_TYPE_START  = 1000 + iota
	MT_CREATE_ENTITY_ON_CLIENT      = 1000 + iota
	MT_DESTROY_ENTITY_ON_CLIENT     = 1000 + iota
	MT_NOTIFY_ATTR_CHANGE_ON_CLIENT = 1000 + iota
	MT_NOTIFY_ATTR_DEL_ON_CLIENT    = 1000 + iota
	MT_CALL_ENTITY_METHOD_ON_CLIENT = 1000 + iota
)

//const ( // Message types that can be received from client
//
//)

func MsgTypeToString(msgType MsgType_t) string {
	return msgTypeToString[int(msgType)]
}
