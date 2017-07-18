package gwutils

import "github.com/xiaonanln/goworld/gwlog"

func RunPanicless(f func()) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
}
