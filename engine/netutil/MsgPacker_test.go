package netutil

import (
	"testing"

	"github.com/xiaonanln/goworld/engine/uuid"
)

type testMsg struct {
	ID        string
	F1        float64
	F2        int
	ListField []interface{}
	MapField  map[string]interface{}
}

func BenchmarkMessagePackMsgPacker(b *testing.B) {
	benchmarkMsgPacker(b, &MessagePackMsgPacker{})
}

func benchmarkMsgPacker(b *testing.B, packer MsgPacker) {
	b.Logf("Testing MsgPacker %T ...", packer)
	msg := testMsg{
		ID:        "abc",
		F1:        0.123124234,
		ListField: []interface{}{1, 2, 3, "abc", "def"},
		MapField:  map[string]interface{}{},
	}
	for i := 0; i < 100; i++ {
		msg.MapField[uuid.GenUUID()] = uuid.GenUUID()
	}

	var totalSize int64
	for i := 0; i < b.N; i++ {

		buf := make([]byte, 0, 100)
		buf, _ = packer.PackMsg(msg, buf)
		totalSize += int64(len(buf))

		var restoreMsg map[string]interface{}
		_ = packer.UnpackMsg(buf, &restoreMsg)
		//if msg.ID != restoreMsg.ID {
		//	b.Fail()
		//}
	}
	b.Logf("average size: %d", totalSize/int64(b.N))
}

func TestMessagePackMsgPacker_UnpackMsg(t *testing.T) {
	msg := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": map[string]interface{}{
			"d": 1,
		},
	}
	buf := make([]byte, 0)
	buf, err := MessagePackMsgPacker{}.PackMsg(msg, buf)
	if err != nil {
		t.Error(err)
	}
	var outmsg map[string]interface{}
	MessagePackMsgPacker{}.UnpackMsg(buf, &outmsg)
	t.Logf("outmsg %T %v", outmsg, outmsg)
	if _, ok := outmsg["c"].(map[interface{}]interface{}); ok {
		t.Errorf("should not unpack with type map[interface{}]interface{}")
	}
}

func BenchmarkMessagePackMsgPacker_PackMsg_Array_AllInOne(b *testing.B) {
	packer := MessagePackMsgPacker{}
	items := []testMsg{}
	for i := 0; i < 3; i++ {
		items = append(items, testMsg{
			ID:        "abc",
			F1:        0.123124234,
			ListField: []interface{}{1, 2, 3, "abc", "def"},
			MapField:  map[string]interface{}{},
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		packer.PackMsg(items, []byte{})
	}
}

func BenchmarkMessagePackMsgPacker_PackMsg_Array_OneByOne(b *testing.B) {
	packer := MessagePackMsgPacker{}
	items := []testMsg{}
	for i := 0; i < 3; i++ {
		items = append(items, testMsg{
			ID:        "abc",
			F1:        0.123124234,
			ListField: []interface{}{1, 2, 3, "abc", "def"},
			MapField:  map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, item := range items {
			packer.PackMsg(item, []byte{})
		}
	}
}
