package util

import (
	"factors/services/error_collector"

	"github.com/sirupsen/logrus"
)

type Hook struct {
	C *error_collector.Collector
}

var (
	levels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}
)

func (h *Hook) Levels() []logrus.Level {
	return levels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	h.C.Notice(entry)
	return nil
}
