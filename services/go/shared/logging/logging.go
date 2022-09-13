package logging

import (
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehhttp"
	"github.com/lefinal/meh/mehlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

// NewLogger creates a new zap.Logger. Don't forget to call Sync() on the
// returned logged before exiting!
func NewLogger(serviceName string, level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	config.OutputPaths = []string{"stdout"}
	config.Level = zap.NewAtomicLevelAt(level)
	config.DisableCaller = true
	config.DisableStacktrace = true
	logger, err := config.Build()
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "new zap production logger", meh.Details{"config": config})
	}
	return logger.Named(serviceName), nil
}

var (
	debugLogger      *zap.Logger
	debugLoggerMutex sync.RWMutex
)

var defaultLevelTranslator map[meh.Code]zapcore.Level
var defaultLevelTranslatorMutex sync.RWMutex

func init() {
	defaultLevelTranslator = make(map[meh.Code]zapcore.Level)
	AddToDefaultLevelTranslator(meh.ErrNotFound, zap.DebugLevel)
	AddToDefaultLevelTranslator(meh.ErrUnauthorized, zap.DebugLevel)
	AddToDefaultLevelTranslator(meh.ErrForbidden, zap.DebugLevel)
	AddToDefaultLevelTranslator(meh.ErrBadInput, zap.DebugLevel)
	AddToDefaultLevelTranslator(mehhttp.ErrCommunication, zap.DebugLevel)
	mehlog.SetDefaultLevelTranslator(func(code meh.Code) zapcore.Level {
		defaultLevelTranslatorMutex.RLock()
		defer defaultLevelTranslatorMutex.RUnlock()
		if level, ok := defaultLevelTranslator[code]; ok {
			return level
		}
		return zap.ErrorLevel
	})
}

// AddToDefaultLevelTranslator adds the given case to the translation map.
func AddToDefaultLevelTranslator(code meh.Code, level zapcore.Level) {
	defaultLevelTranslatorMutex.Lock()
	defaultLevelTranslator[code] = level
	defaultLevelTranslatorMutex.Unlock()
}

// DebugLogger returns the logger set via SetDebugLogger. If none is set, a
// zap.NewProduction will be created.
func DebugLogger() *zap.Logger {
	debugLoggerMutex.RLock()
	defer debugLoggerMutex.RUnlock()
	if debugLogger == nil {
		tempLogger, _ := NewLogger("debug", zap.InfoLevel)
		return tempLogger
	}
	return debugLogger
}

// SetDebugLogger sets the logger that can be retrieved with DebugLogger.
func SetDebugLogger(logger *zap.Logger) {
	debugLoggerMutex.Lock()
	defer debugLoggerMutex.Unlock()
	debugLogger = logger
}
