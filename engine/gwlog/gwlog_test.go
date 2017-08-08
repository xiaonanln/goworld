package gwlog

import "testing"

func TestFatalf(t *testing.T) {
	defer func() {
		_ = recover()
	}()
	Fatalf("abc: %d %s", 1, "2")
}
