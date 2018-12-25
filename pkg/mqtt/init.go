package mqtt

import (
	"github.com/andig/ingress/pkg/log"
	"gopkg.in/birkirb/loggers.v1"
)

var logger loggers.Contextual

func init() {
	log.Register(setLogger)
}

func setLogger(l loggers.Contextual) {
	logger = l
}

// Log returns a contextual logger
func Log(fields ...interface{}) loggers.Advanced {
	// return log.WithModule(logger, "mqtt", fields...)
	return log.WithContext(logger, fields...)
}
