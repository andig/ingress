package log

import "gopkg.in/birkirb/loggers.v1"

type ContextualLoggerCallback func(logger loggers.Contextual)

var callbacks []ContextualLoggerCallback

func init() {
	callbacks = make([]ContextualLoggerCallback, 0)
}

func Register(cb ContextualLoggerCallback) {
	callbacks = append(callbacks, cb)
}

func InitLoggers(level string) {
	for _, cb := range callbacks {
		logger := NewLogger(level)
		cb(logger)
	}
}
