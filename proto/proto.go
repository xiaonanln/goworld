package goworld_proto

var (
	msgTypeToString = map[int]string{}
)

type MsgType_t uint16

func MsgTypeToString(msgType MsgType_t) string {
	return msgTypeToString[int(msgType)]
}
