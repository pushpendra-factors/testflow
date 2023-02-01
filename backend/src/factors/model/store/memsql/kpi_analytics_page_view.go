package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForPageViews(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	config := model.KPIConfigForPageViews
	rMetrics := model.GetStaticallyDefinedMetricsForDisplayCategory(model.PageViewsDisplayCategory)
	standardUserProperties := store.GetKPIConfigFromStandardUserProperties(projectID)
	rProperties := append(model.KPIPropertiesForPageViews, standardUserProperties...)

	config["metrics"] = rMetrics
	config["properties"] = rProperties
	return config, http.StatusOK
}
