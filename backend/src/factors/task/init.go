package task

import (
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const notify_SOURCE = "task"

var taskLog = log.New()

func init() {
	taskLog.Formatter = new(prefixed.TextFormatter)
}
