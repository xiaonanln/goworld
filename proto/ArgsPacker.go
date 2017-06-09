package proto

var (
	ARGS_PACKER MsgPacker = MessagePackMsgPacker{}
)

type MsgPacker interface {
	PackMsg(msg interface{}, buf []byte) ([]byte, error)
	UnpackMsg(data []byte, msg interface{}) error
}
