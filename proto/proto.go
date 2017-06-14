package proto

var (
	msgTypeToString = map[int]string{}
)

type MsgType_t uint16

const (
	MT_INVALID = iota
	// Server Messages
	MT_SET_SERVER_ID           = iota
	MT_NOTIFY_CREATE_ENTITY    = iota
	MT_DECLARE_SERVICE         = iota
	MT_CALL_ENTITY_METHOD      = iota
	MT_CREATE_ENTITY_ANYWHERE  = iota
	MT_LOAD_ENTITY_ANYWHERE    = iota
	MT_NOTIFY_CLIENT_CONNECTED = iota
)
const ( // Message types that should be handled by GateService
	MT_GATE_SERVICE_MSG_TYPE_START = 10000 + iota
	MT_CREATE_ENTITY_ON_CLIENT     = 10000 + iota
	MT_DESTROY_ENTITY_ON_CLIENT    = 10000 + iota
)

func MsgTypeToString(msgType MsgType_t) string {
	return msgTypeToString[int(msgType)]
}
