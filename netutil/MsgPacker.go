package netutil

var (
	MSG_PACKER MsgPacker = JSONMsgPacker{}
)

type MsgPacker interface {
	PackMsg(msg interface{}, buf []byte) ([]byte, error)
	UnpackMsg(data []byte, msg interface{}) error
}
