package gwlog

import (
	"runtime/debug"

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
)

func ParseLevel(lvl string) (sublog.Level, error) {
	return sublog.ParseLevel(lvl)
}

func SetLevel(lv sublog.Level) {
	sublog.SetLevel(lv)
}

func TraceError(format string, args ...interface{}) {
	debug.PrintStack()
	Error(format, args...)
}
