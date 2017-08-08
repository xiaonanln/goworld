package crontab

import "testing"

var (
	quit int64
)

func init() {
	Initialize()
}

func TestRegister(t *testing.T) {
	count := 0
	Register(-1, -1, -1, -1, -1, func() {
		t.Logf("crontab every minute")
		count += 1
		if count == 2 {
			quit = 1
		}
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

//
//func timerLoop() {
//	for quit == 0 {
//		timer.Tick()
//		time.Sleep(time.Millisecond)
//	}
//}
