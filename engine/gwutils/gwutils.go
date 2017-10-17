package gwutils

import "github.com/xiaonanln/goworld/engine/gwlog"

// RunPanicless calls a function panic-freely
func RunPanicless(f func()) (paniced bool) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.TraceError("%s panic: %s", f, err)
			paniced = true
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
