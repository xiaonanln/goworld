package uuid

import "testing"

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
