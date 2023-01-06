package task

import (
	log "github.com/sirupsen/logrus"
)

var taskLog = log.New()

func init() {
	taskLog.SetFormatter(&log.JSONFormatter{})
}
