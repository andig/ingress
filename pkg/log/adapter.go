package log

import (
	"reflect"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/sirupsen/logrus"
	"gopkg.in/birkirb/loggers.v1"
)

func convertMap(fields ...interface{}) []interface{} {
	if len(fields) != 1 {
		return fields
	}
	v := reflect.ValueOf(fields[0])
	if v.Kind() != reflect.Map {
		return fields
	}

	// convert map to array
	l := v.Len()
	list := make([]interface{}, 0, 2*l+2) // reserve capacity for module name
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		list = append(list, key, val)
	}
	return list
}

// WithContext returns advanced logger with fields context
func WithContext(logger loggers.Contextual, fields ...interface{}) loggers.Advanced {
	fields = convertMap(fields...)
	return logger.WithFields(fields...)
}

// WithModule returns advanced logger with module and fields context
func WithModule(logger loggers.Contextual, module string, fields ...interface{}) loggers.Advanced {
	fields = append(convertMap(fields...), "_module", module)
	return logger.WithFields(fields...)
}

// NewLogger creates default contextual logger
func NewLogger(level string) loggers.Contextual {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Fatal("invalid log level " + level)
	}

	logrusLogger := logrus.New()
	logrusLogger.Level = lvl
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})

	return &mapper.Logger{
		Logger: logrusLogger,
	}
}
