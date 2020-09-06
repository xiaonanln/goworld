package netutil

import (
	"bytes"

	"github.com/vmihailenco/msgpack"
)

// MessagePackMsgPacker packs and unpacks message in MessagePack format
type MessagePackMsgPacker struct{}

// PackMsg packs message to bytes in MessagePack format
func (mp MessagePackMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(buf)

	encoder := msgpack.NewEncoder(buffer)
	err := encoder.Encode(msg)
	if err != nil {
		return buf, err
	}
	buf = buffer.Bytes()
	return buf, nil
}

// UnpackMsg unpacksbytes in MessagePack format to message
func (mp MessagePackMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	err := msgpack.Unmarshal(data, msg)
	return err
}
