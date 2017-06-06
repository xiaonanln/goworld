package proto

import (
	"bytes"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type MessagePackMsgPacker struct{}

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

func (mp MessagePackMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	err := msgpack.Unmarshal(data, msg)
	return err
}
