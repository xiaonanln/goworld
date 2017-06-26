package gwlog

import (
	"runtime/debug"

	"io"

	"os"

	sublog "github.com/Sirupsen/logrus"
)

var (
	DEBUG = sublog.DebugLevel
	INFO  = sublog.InfoLevel
	WARN  = sublog.WarnLevel
	ERROR = sublog.ErrorLevel
	PANIC = sublog.PanicLevel
	FATAL = sublog.FatalLevel

	Debug  = sublog.Debugf
	Info   = sublog.Infof
	Warn   = sublog.Warnf
	Error  = sublog.Errorf
	Panic  = sublog.Panic
	Panicf = sublog.Panicf
	Fatal  = sublog.Fatalf

	outputWriter io.Writer
)

func init() {
	outputWriter = os.Stderr
	sublog.SetOutput(outputWriter)

	sublog.SetLevel(sublog.DebugLevel)
}

func ParseLevel(lvl string) (sublog.Level, error) {
	return sublog.ParseLevel(lvl)
}

func SetLevel(lv sublog.Level) {
	sublog.SetLevel(lv)
}

func TraceError(format string, args ...interface{}) {
	outputWriter.Write(debug.Stack())
	Error(format, args...)
}

func SetOutput(out io.Writer) {
	outputWriter = out
	sublog.SetOutput(outputWriter)
}

func GetOutput() io.Writer {
	return outputWriter
}
