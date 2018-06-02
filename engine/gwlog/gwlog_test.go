package gwlog

import "testing"

func TestGWLog(t *testing.T) {
	SetSource("gwlog_test")

	if lv := StringToLevel("debug"); lv != DebugLevel {
		t.Fail()
	}
	if lv := StringToLevel("info"); lv != InfoLevel {
		t.Fail()
	}
	if lv := StringToLevel("warn"); lv != WarnLevel {
		t.Fail()
	}
	if lv := StringToLevel("error"); lv != ErrorLevel {
		t.Fail()
	}
	if lv := StringToLevel("panic"); lv != PanicLevel {
		t.Fail()
	}
	if lv := StringToLevel("fatal"); lv != FatalLevel {
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
