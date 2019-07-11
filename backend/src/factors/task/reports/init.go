package reports

import (
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var baseLog = log.New()

func init() {
	baseLog.Formatter = new(prefixed.TextFormatter)
	baseLog.SetReportCaller(true)
}
