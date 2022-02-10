package memsql

import (
	"factors/model/model"
	"net/http"
	"time"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForPageViews(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id": reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	config := model.KPIConfigForPageViews
	config["metrics"] = model.GetMetricsForDisplayCategory(model.PageViewsDisplayCategory)
	return config, http.StatusOK
}
