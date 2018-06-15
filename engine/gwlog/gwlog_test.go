package gwlog

import "testing"

func TestGWLog(t *testing.T) {
	SetSource("gwlog_test")
	SetOutput([]string{"stderr", "gwlog_test.log"})
	SetLevel(DebugLevel)

	if lv := ParseLevel("debug"); lv != DebugLevel {
		t.Fail()
	}
	if lv := ParseLevel("info"); lv != InfoLevel {
		t.Fail()
	}
	if lv := ParseLevel("warn"); lv != WarnLevel {
		t.Fail()
	}
	if lv := ParseLevel("error"); lv != ErrorLevel {
		t.Fail()
	}
	if lv := ParseLevel("panic"); lv != PanicLevel {
		t.Fail()
	}
	if lv := ParseLevel("fatal"); lv != FatalLevel {
		t.Fail()
	}

	Debugf("this is a debug %d", 1)
	SetLevel(InfoLevel)
	Debugf("SHOULD NOT SEE THIS!")
	Infof("this is an info %d", 2)
	Warnf("this is a warning %d", 3)
	TraceError("this is a trace error %d", 4)
	func() {
		defer func() {
			_ = recover()
		}()
		Panicf("this is a panic %d", 4)
	}()

	func() {
		defer func() {
			_ = recover()
		}()
		//Fatalf("this is a fatal %d", 5)
	}()
}
