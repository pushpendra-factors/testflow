package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForFormSubmissions(projectID int64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	config := model.KPIConfigForFormSubmissions
	config["metrics"] = model.GetMetricsForDisplayCategory(model.FormSubmissionsDisplayCategory)
	return config, http.StatusOK
}
