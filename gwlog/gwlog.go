package gwlog

import (
	"runtime/debug"

	"io"

	"os"

	"strings"

	sublog "github.com/Sirupsen/logrus"
)

var (
	DEBUG Level = Level(sublog.DebugLevel)
	INFO  Level = Level(sublog.InfoLevel)
	WARN  Level = Level(sublog.WarnLevel)
	ERROR Level = Level(sublog.ErrorLevel)
	PANIC Level = Level(sublog.PanicLevel)
	FATAL Level = Level(sublog.FatalLevel)

	Debug  = sublog.Debugf
	Info   = sublog.Infof
	Warn   = sublog.Warnf
	Error  = sublog.Errorf
	Panic  = sublog.Panic
	Panicf = sublog.Panicf

	outputWriter io.Writer
)

type Level uint8

func init() {
	outputWriter = os.Stderr
	sublog.SetOutput(outputWriter)

	sublog.SetLevel(sublog.DebugLevel)
}

func ParseLevel(lvl string) (sublog.Level, error) {
	return sublog.ParseLevel(lvl)
}

func SetLevel(lv Level) {
	sublog.SetLevel(sublog.Level(lv))
}

func TraceError(format string, args ...interface{}) {
	outputWriter.Write(debug.Stack())
	Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	debug.PrintStack()
	sublog.Fatalf(format, args...)
}

func SetOutput(out io.Writer) {
	outputWriter = out
	sublog.SetOutput(outputWriter)
}

func GetOutput() io.Writer {
	return outputWriter
}

func StringToLevel(s string) Level {
	if strings.ToLower(s) == "debug" {
		return DEBUG
	} else if strings.ToLower(s) == "info" {
		return INFO
	} else if strings.ToLower(s) == "warn" || strings.ToLower(s) == "warning" {
		return WARN
	} else if strings.ToLower(s) == "error" {
		return ERROR
	} else if strings.ToLower(s) == "panic" {
		return PANIC
	} else if strings.ToLower(s) == "fatal" {
		return FATAL
	}
	Error("StringToLevel: unknown level: %s", s)
	return DEBUG
}
