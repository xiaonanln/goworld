package post

import "testing"

func TestPost(t *testing.T) {
	var a int
	Post(func() {
		a = 1
	})
	Tick()
	if a != 1 {
		t.Errorf("t should be 1")
	}
}
