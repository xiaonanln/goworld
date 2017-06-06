package proto

import (
	"bytes"
	"encoding/gob"
)

type GobMsgPacker struct{}

func (mp GobMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	//return msgpack.Marshal(msg)
	buffer := bytes.NewBuffer(buf)

	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(msg)
	if err != nil {
		return buf, err
	}
	buf = buffer.Bytes()
	return buf, nil
}

func (mp GobMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(msg)
	return err
}
