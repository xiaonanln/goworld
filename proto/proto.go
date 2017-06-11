package proto

var (
	msgTypeToString = map[int]string{}
)

type MsgType_t uint16

const (
	MT_INVALID                = iota
	MT_SET_GAME_ID            = iota
	MT_NOTIFY_CREATE_ENTITY   = iota
	MT_DECLARE_SERVICE        = iota
	MT_CALL_ENTITY_METHOD     = iota
	MT_CREATE_ENTITY_ANYWHERE = iota
	MT_LOAD_ENTITY_ANYWHERE   = iota
)

func MsgTypeToString(msgType MsgType_t) string {
	return msgTypeToString[int(msgType)]
}
