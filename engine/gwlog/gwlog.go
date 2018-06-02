package gwlog

import (
	"runtime/debug"

	"io"

	"strings"

	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	outputWriter io.Writer

	// DebugLevel level
	DebugLevel Level = Level(zap.DebugLevel)
	// InfoLevel level
	InfoLevel Level = Level(zap.InfoLevel)
	// WarnLevel level
	WarnLevel Level = Level(zap.WarnLevel)
	// ErrorLevel level
	ErrorLevel Level = Level(zap.ErrorLevel)
	// PanicLevel level
	PanicLevel Level = Level(zap.PanicLevel)
	// FatalLevel level
	FatalLevel Level = Level(zap.FatalLevel)

	// Debugf logs formatted debug message
	Debugf logFormatFunc
	// Infof logs formatted info message
	Infof logFormatFunc
	// Warnf logs formatted warn message
	Warnf logFormatFunc
	// Errorf logs formatted error message
	Errorf logFormatFunc
	Panicf logFormatFunc
	Fatalf logFormatFunc
	Fatal  func(args ...interface{})
	Panic  func(args ...interface{})
)

type logFormatFunc func(format string, args ...interface{})

// Level is type of log levels
type Level zapcore.Level

var (
	cfg    zap.Config
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

func init() {
	var err error
	cfgJson := []byte(`{
		"level": "debug",
		"outputPaths": ["stderr"],
		"errorOutputPaths": ["stderr"],
		"encoding": "console",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`)

	if err = json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}

	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	setSugar(logger.Sugar())
}

// SetSource sets the component name (dispatcher/gate/game) of gwlog module
func SetSource(comp string) {
	logger = logger.With(zap.String("source", comp))
	setSugar(logger.Sugar())
}

func setSugar(sugar_ *zap.SugaredLogger) {
	sugar = sugar_
	Debugf = sugar.Debugf
	Infof = sugar.Infof
	Warnf = sugar.Warnf
	Errorf = sugar.Errorf
	Panicf = sugar.Panicf
	Panic = sugar.Panic
	Fatalf = sugar.Fatalf
	Fatal = sugar.Fatal
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	//zap.SetLevel(zapcore.Level(lv))
}

// TraceError prints the stack and error
func TraceError(format string, args ...interface{}) {
	outputWriter.Write(debug.Stack())
	Errorf(format, args...)
}

// SetOutput sets the output writer
func SetOutput(out io.Writer) {
	//outputWriter = out
	//zap.SetOutput(outputWriter)
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
	Errorf("StringToLevel: unknown level: %s", s)
	return DebugLevel
}
