package gwlog

import "testing"

func TestGWLog(t *testing.T) {
	SetComponent("gwlog_test")

	if lv, err := ParseLevel("debug"); err != nil || lv != DebugLevel {
		t.Fail()
	}
	if lv, err := ParseLevel("info"); err != nil || lv != InfoLevel {
		t.Fail()
	}
	if lv, err := ParseLevel("warn"); err != nil || lv != WarnLevel {
		t.Fail()
	}
	if lv, err := ParseLevel("error"); err != nil || lv != ErrorLevel {
		t.Fail()
	}
	if lv, err := ParseLevel("panic"); err != nil || lv != PanicLevel {
		t.Fail()
	}
	if lv, err := ParseLevel("fatal"); err != nil || lv != FatalLevel {
		t.Fail()
	}

	Debugf("this is a debug %d", 1)
	Infof("this is an info %d", 2)
	Warnf("this is a warning %d", 3)

	func() {
		defer func() {
			_ = recover()
		}()
		Panicf("this is a panic %d", 4)
	}()

	func() {
		defer recover()
		// Fatalf("this is a fatal %d", 5)
	}()
}
