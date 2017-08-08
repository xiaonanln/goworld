package gwutils

import "github.com/xiaonanln/goworld/engine/gwlog"

// RunPanicless calls a function panic-freely
func RunPanicless(f func()) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
}
