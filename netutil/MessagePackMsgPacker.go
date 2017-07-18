package netutil

import (
	"bytes"

	"github.com/xiaonanln/goworld/gwlog"
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
	if pv, ok := msg.(*interface{}); ok {
		*pv = mp.convertToStringKeys(*pv)
	} else if pv, ok := msg.(*[]interface{}); ok {
		*pv = mp.convertToStringKeys(*pv).([]interface{})
	} else if pv, ok := msg.(*map[string]interface{}); ok {
		*pv = mp.convertToStringKeys(*pv).(map[string]interface{})
	}

	return err
}

func (mp MessagePackMsgPacker) convertToStringKeys(v interface{}) interface{} {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.Error("MessagePackMsgPacker.convertToStringKeys failed while converting: %v", v)
			panic(err)
		}
	}()

	if rv, ok := v.(map[interface{}]interface{}); ok {
		rrv := make(map[string]interface{}, len(rv))
		for k, _v := range rv {
			ks, ok := k.(string)
			if !ok {
				gwlog.Panicf("%v is not string, but %T", k, k)
			}

			rrv[ks] = mp.convertToStringKeys(_v)
		}
		return rrv
	}

	if rv, ok := v.(map[string]interface{}); ok {
		for k, _v := range rv {
			rv[k] = mp.convertToStringKeys(_v)
		}
		return rv
	}

	if rv, ok := v.([]interface{}); ok {
		for i, _v := range rv {
			rv[i] = mp.convertToStringKeys(_v)
		}
		return rv
	}

	return v
}
