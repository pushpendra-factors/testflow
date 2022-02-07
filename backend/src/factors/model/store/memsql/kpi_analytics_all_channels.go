package memsql

import (
	"factors/model/model"
	"net/http"
	"time"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForAllChannels(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id": reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return model.GetKPIConfigsForAllChannels(), http.StatusOK
}
