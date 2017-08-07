package gwlog

import (
	"runtime/debug"

	"io"

	"os"

	"strings"

	sublog "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var (
	// DebugLevel level
	DebugLevel Level = Level(sublog.DebugLevel)
	// InfoLevel level
	InfoLevel Level = Level(sublog.InfoLevel)
	// WarnLevel level
	WarnLevel Level = Level(sublog.WarnLevel)
	// ErrorLevel level
	ErrorLevel Level = Level(sublog.ErrorLevel)
	// PanicLevel level
	PanicLevel Level = Level(sublog.PanicLevel)
	// FatalLevel level
	FatalLevel Level = Level(sublog.FatalLevel)

	// Debug logs formatted debug message
	Debug = sublog.Debugf
	// Info logs formatted info message
	Info = sublog.Infof
	// Warn logs formatted warn message
	Warn = sublog.Warnf
	// Error logs formatted error message
	Error = sublog.Errorf

	outputWriter io.Writer
)

// Level is type of log levels
type Level uint8

func init() {
	outputWriter = os.Stderr
	sublog.SetOutput(outputWriter)

	sublog.SetLevel(sublog.DebugLevel)
}

// ParseLevel parses log level string to Level
func ParseLevel(lvl string) (sublog.Level, error) {
	return sublog.ParseLevel(lvl)
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	sublog.SetLevel(sublog.Level(lv))
}

// TraceError prints the stack and error
func TraceError(format string, args ...interface{}) {
	outputWriter.Write(debug.Stack())
	Error(format, args...)
}

// Fatalf prints formatted fatal message
func Fatalf(format string, args ...interface{}) {
	debug.PrintStack()
	sublog.Fatalf(format, args...)
}

// Panic panics
func Panic(v interface{}) {
	panic(v)
}

// Panicf prints formatted panic message
func Panicf(format string, args ...interface{}) {
	panic(errors.Errorf(format, args...))
}

// SetOutput sets the output writer
func SetOutput(out io.Writer) {
	outputWriter = out
	sublog.SetOutput(outputWriter)
}

// GetOutput returns the output writer
func GetOutput() io.Writer {
	return outputWriter
}

// StringToLevel converts string to Levels
func StringToLevel(s string) Level {
	if strings.ToLower(s) == "debug" {
		return DebugLevel
	} else if strings.ToLower(s) == "info" {
		return InfoLevel
	} else if strings.ToLower(s) == "warn" || strings.ToLower(s) == "warning" {
		return WarnLevel
	} else if strings.ToLower(s) == "error" {
		return ErrorLevel
	} else if strings.ToLower(s) == "panic" {
		return PanicLevel
	} else if strings.ToLower(s) == "fatal" {
		return FatalLevel
	}
	Error("StringToLevel: unknown level: %s", s)
	return DebugLevel
}
