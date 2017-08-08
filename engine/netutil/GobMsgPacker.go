package netutil

import (
	"bytes"
	"encoding/gob"
)

// GobMsgPacker packs and unpacks message in golang's Gob format
type GobMsgPacker struct{}

// PackMsg packs a message to bytes of gob format
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

// UnpackMsg unpacks bytes of gob format to message
func (mp GobMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(msg)
	return err
}
