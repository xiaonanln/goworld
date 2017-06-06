package proto

var (
	msgTypeToString = map[int]string{}
)

type msgtype_t uint16

const (
	MT_INVALID     = iota
	MT_SET_GAME_ID = iota
)

func MsgTypeToString(msgType msgtype_t) string {
	return msgTypeToString[int(msgType)]
}
