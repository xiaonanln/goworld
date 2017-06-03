package goworld_proto

import (
	"bytes"
	"encoding/json"
)

type JSONMsgPacker struct {
}

func (mp JSONMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(buf)
	jsonEncoder := json.NewEncoder(buffer)
	err := jsonEncoder.Encode(msg)
	if err != nil {
		return buf, err
	}
	buf = buffer.Bytes()
	return buf[:len(buf)-1], nil // encoder always put '\n' at the end, we trim it
}

func (mp JSONMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	err := json.Unmarshal(data, msg)
	return err
}
