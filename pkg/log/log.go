package log

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/sirupsen/logrus"
)

const (
	ID  = "id"   // source
	SRC = "src"  // source
	EV  = "evt"  // event
	VAL = "val"  // value
	TGT = "trgt" // target
)

var logger *logrus.Logger = logrus.StandardLogger()

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

// Context returns a contextual logger
func Context(fields ...interface{}) *logrus.Entry {
	fields = convertMap(fields...)

	// convert
	f := make(map[string]interface{}, len(fields)/2)
	var key, value interface{}
	for i := 0; i+1 < len(fields); i = i + 2 {
		key = fields[i]
		value = fields[i+1]
		if s, ok := key.(string); ok {
			f[s] = value
		} else if s, ok := key.(fmt.Stringer); ok {
			f[s.String()] = value
		}
	}

	return logger.WithFields(f)
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

func Configure(lvl string) {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		logrus.Fatal("invalid log level " + lvl)
	}

	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
		SortingFunc:     contextSort,
	})
}
