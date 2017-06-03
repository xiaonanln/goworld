package goworld_proto

import (
	"testing"

	"github.com/xiaonanln/goworld/uuid"
)

type testMsg struct {
	ID        string
	F1        float64
	F2        int
	ListField []interface{}
	MapField  map[string]interface{}
}

func BenchmarkJSONMsgPacker(b *testing.B) {
	benchmarkMsgPacker(b, &JSONMsgPacker{})
}

func BenchmarkMessagePackMsgPacker(b *testing.B) {
	benchmarkMsgPacker(b, &MessagePackMsgPacker{})
}

func BenchmarkGobMsgPacker(b *testing.B) {
	benchmarkMsgPacker(b, &GobMsgPacker{})
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

	for i := 0; i < b.N; i++ {

		buf := make([]byte, 0, 100)
		buf, _ = packer.PackMsg(msg, buf)

		var restoreMsg testMsg
		_ = packer.UnpackMsg(buf, &restoreMsg)
		if msg.ID != restoreMsg.ID {
			b.Fail()
		}
	}
}
