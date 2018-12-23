package log

import (
	"github.com/sirupsen/logrus"
)

func GetInstance(module string) *logrus.Entry {
	log := logrus.WithFields(logrus.Fields{
		"module": module,
	})
	return log
}
