package log

import "gopkg.in/birkirb/loggers.v1"

const (
	SRC = "src" // source
	EV  = "ev"  // event
	VAL = "val" // value
	TGT = "tgt" // target
)

var logger loggers.Contextual

// Log returns a contextual logger
func Log(fields ...interface{}) loggers.Advanced {
	return WithContext(logger, fields...)
}

func Configure(level string) {
	logger = NewLogger(level)
	InitLoggers(level)
}
