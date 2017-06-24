package gwutils

import (
	"io"

	"github.com/xiaonanln/goworld/gwlog"
)

func RunPanicless(f func()) {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
}

type MultiWriter struct {
	subwriters []io.Writer
}

func NewMultiWriter(writers ...io.Writer) io.Writer {
	mw := &MultiWriter{
		subwriters: writers,
	}
	return mw
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	for _, subwriter := range mw.subwriters {
		subwriter.Write(p)
	}
	return
}
