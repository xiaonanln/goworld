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
			gwlog.TraceError("%v panic: %s", f, err)
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

// PanicFree makes the function panic-free
func PanicFree(f func(), errors ...error) func() {
	return func() {
		defer func() {
			_err := recover()
			if len(errors) == 0 {
				// ignore all errors
				gwlog.Errorf("%v panic: %s", f, _err)
			} else if err, ok := _err.(error); ok {
				// check if err in errors
				for _, ignoreErr := range errors {
					if ignoreErr == err {
						gwlog.Errorf("%v panic: %s", f, err)
						return
					}
				}
				panic(err)
			} else {
				panic(_err)
			}
		}()
		f()
	}
}

// NextLargerKey finds the next key that is larger than the specified key,
// but smaller than any other keys that is larger than the specified key
func NextLargerKey(key string) string {
	return key + "\x00" // the next string that is larger than key, but smaller than any other keys > key
}
