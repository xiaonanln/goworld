package crontab

import "testing"

func init() {
	Initialize()
}

func TestRegister(t *testing.T) {
	Register(-1, -1, -1, -1, -1, func() {
		t.Logf("crontab every minute")
	})
	check()
}

func TestUnregister(t *testing.T) {
	var h Handle
	Register(-1, -1, -1, -1, -1, func() {
		t.Logf("crontab every minute 1")
	})

	h = Register(-1, -1, -1, -1, -1, func() {
		t.Logf("crontab every minute 2")
		h.Unregister()
	})
	check()
}
