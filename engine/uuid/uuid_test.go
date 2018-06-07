package uuid

import (
	"strconv"
	"testing"
)

func TestGenUUID(t *testing.T) {
	for i := 0; i < 100; i++ {
		uuid := GenUUID()
		t.Logf("GenUUID: %s", uuid)
		if len(uuid) != UUID_LENGTH {
			t.FailNow()
		}
	}
}

func BenchmarkGenUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenUUID()
	}
}

func TestGenFixedUUID(t *testing.T) {
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		u1 := GenFixedUUID([]byte(s))
		u2 := GenFixedUUID([]byte(s))
		if u1 != u2 {
			t.Fatalf("GenFixedUUID is not fixed")
		}
		t.Logf("GenFixedUUID: %v => %v", i, u1)
	}
}
