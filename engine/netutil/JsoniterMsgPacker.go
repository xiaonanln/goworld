package netutil

// JsoniterMsgPacker packs and unpacks messages in JSON format
type JsoniterMsgPacker struct{}

// PackMsg packs message to bytes of JSON format
func (mp JsoniterMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	return nil, nil
	//buffer := bytes.NewBuffer(buf)
	//jsonEncoder := jsoniter.NewEncoder(buffer)
	//err := jsonEncoder.Encode(msg)
	//if err != nil {
	//	return buf, err
	//}
	//return buffer.Bytes(), nil
}

// UnpackMsg unpacks bytes of JSON format to message
func (mp JsoniterMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	//err := jsoniter.Unmarshal(data, msg)
	//return err
	return nil
}
