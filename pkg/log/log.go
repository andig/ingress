package log

import (
	"reflect"
	"sort"

	mapper "github.com/birkirb/loggers-mapper-logrus"
	"github.com/sirupsen/logrus"
	"gopkg.in/birkirb/loggers.v1"
)

const (
	ID  = "id"   // source
	SRC = "src"  // source
	EV  = "evt"  // event
	VAL = "val"  // value
	TGT = "trgt" // target
)

var logger loggers.Contextual
var level logrus.Level

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

// Log returns a contextual logger
func Log(fields ...interface{}) loggers.Advanced {
	if logger == nil {
		level = logrus.TraceLevel
		logger = NewLogger()
	}

	fields = convertMap(fields...)
	return logger.WithFields(fields...)
}

func contextSort(keys []string) {
	order := map[string]int{
		SRC: 10,
		EV:  20,
		TGT: 40,
		ID:  30,
		VAL: 50,
	}

	// extract fixed keys
	fixed := make([]string, 0)
	for i, key := range keys {
		if _, ok := order[key]; ok {
			fixed = append(fixed, key)
			keys[i] = ""
		}
	}

	// sort fixed keys
	sort.Slice(fixed, func(i, j int) bool {
		return order[fixed[i]] < order[fixed[j]]
	})

	// sort remaining keys
	sort.Strings(keys)

	// add remaining keys
	for _, key := range keys {
		if key != "" {
			fixed = append(fixed, key)
		}
	}

	for i, key := range fixed {
		keys[i] = key
	}
}

// NewLogger creates default contextual logger
func NewLogger() loggers.Contextual {
	logrusLogger := logrus.New()
	logrusLogger.Level = getLevel()
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
		SortingFunc:     contextSort,
	})

	return &mapper.Logger{
		Logger: logrusLogger,
	}
}

func setLevel(lvl string) {
	var err error
	level, err = logrus.ParseLevel(lvl)
	if err != nil {
		logrus.Fatal("invalid log level " + lvl)
	}
}

func getLevel() logrus.Level {
	return level
}

func Configure(lvl string) {
	setLevel(lvl)
	logger = NewLogger()
}
