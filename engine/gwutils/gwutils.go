package gwutils

import "github.com/xiaonanln/goworld/engine/gwlog"

// CatchPanic calls a function and returns the error if function paniced
func CatchPanic(f func()) (err interface{}) {
	defer func() {
		err = recover()
		if err != nil {
			gwlog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
	return
}

// RunPanicless calls a function panic-freely
func RunPanicless(f func()) (panicless bool) {
	defer func() {
		err := recover()
		panicless = err == nil
		if err != nil {
			gwlog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
	return
}

// RepeatUntilPanicless runs the function repeatly until there is no panic
func RepeatUntilPanicless(f func()) {
	for !RunPanicless(f) {
	}
}
