package gwlog

import (
	"runtime/debug"

	"strings"

	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
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
	Error  func(args ...interface{})
	Fatal  func(args ...interface{})
	Panic  func(args ...interface{})
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
		if source != "" {
			logger = logger.With(zap.String("source", source))
		}
		setSugar(logger.Sugar())
	} else {
		panic(err)
	}
}

func setSugar(sugar_ *zap.SugaredLogger) {
	sugar = sugar_
	Debugf = sugar.Debugf
	Infof = sugar.Infof
	Warnf = sugar.Warnf
	Errorf = sugar.Errorf
	Error = sugar.Error
	Panicf = sugar.Panicf
	Panic = sugar.Panic
	Fatalf = sugar.Fatalf
	Fatal = sugar.Fatal
}
