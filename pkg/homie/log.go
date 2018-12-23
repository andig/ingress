package homie

import "github.com/sirupsen/logrus"

var Log *logrus.Entry

func InitLog() {
	if Log == nil {
		Log = logrus.WithFields(logrus.Fields{
			"module": "homie",
		})
	}
}
