package crontab

import (
	"testing"

	"time"

	"github.com/xiaonanln/goTimer"
)

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
	timerLoop()
}

func TestUnregister(t *testing.T) {
	var h Handle
	Register(-1, -1, -1, -1, -1, func() {
		t.Logf("crontab every minute 1")
	})

	h = Register(-1, -1, -1, -1, -1, func() {
		t.Logf("crontab every minute 2")
		Unregister(h)
	})

	timerLoop()
}

func timerLoop() {
	for quit == 0 {
		timer.Tick()
		time.Sleep(time.Millisecond)
	}
}
