package gwlog

import (
	"runtime/debug"

	"strings"

	"encoding/json"

	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// DebugLevel level
	DebugLevel = Level(zap.DebugLevel)
	// InfoLevel level
	InfoLevel = Level(zap.InfoLevel)
	// WarnLevel level
	WarnLevel = Level(zap.WarnLevel)
	// ErrorLevel level
	ErrorLevel = Level(zap.ErrorLevel)
	// PanicLevel level
	PanicLevel = Level(zap.PanicLevel)
	// FatalLevel level
	FatalLevel = Level(zap.FatalLevel)
)

type logFormatFunc func(format string, args ...interface{})

// Level is type of log levels
type Level = zapcore.Level

var (
	cfg          zap.Config
	logger       *zap.Logger
	sugar        *zap.SugaredLogger
	source       string
	currentLevel Level
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
	currentLevel = DebugLevel

	if err = json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	rebuildLoggerFromCfg()
}

// SetSource sets the component name (dispatcher/gate/game) of gwlog module
func SetSource(source_ string) {
	source = source_
	rebuildLoggerFromCfg()
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	currentLevel = lv
	cfg.Level.SetLevel(lv)
}

// GetLevel get the current log level
func GetLevel() Level {
	return currentLevel
}

// TraceError prints the stack and error
func TraceError(format string, args ...interface{}) {
	Error(string(debug.Stack()))
	Errorf(format, args...)
}

// SetOutput sets the output writer
func SetOutput(outputs []string) {
	cfg.OutputPaths = outputs
	rebuildLoggerFromCfg()
}

// ParseLevel converts string to Levels
func ParseLevel(s string) Level {
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
	Errorf("ParseLevel: unknown level: %s", s)
	return DebugLevel
}

func rebuildLoggerFromCfg() {
	if newLogger, err := cfg.Build(); err == nil {
		if logger != nil {
			logger.Sync()
		}
		logger = newLogger
		//logger = logger.With(zap.Time("ts", time.Now()))
		if source != "" {
			logger = logger.With(zap.String("source", source))
		}
		setSugar(logger.Sugar())
	} else {
		panic(err)
	}
}

func Debugf(format string, args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	debug.PrintStack()
	sugar.With(zap.Time("ts", time.Now())).Fatalf(format, args...)
}

func Error(args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Error(args...)
}

func Panic(args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Panic(args...)
}

func Fatal(args ...interface{}) {
	sugar.With(zap.Time("ts", time.Now())).Fatal(args...)
}

func setSugar(sugar_ *zap.SugaredLogger) {
	sugar = sugar_
}
