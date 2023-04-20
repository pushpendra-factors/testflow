package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForFormSubmissions(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	config := model.KPIConfigForFormSubmissions
	rMetrics := model.GetStaticallyDefinedMetricsForDisplayCategory(model.FormSubmissionsDisplayCategory)
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.FormSubmissionsDisplayCategory, includeDerivedKPIs)...)

	standardUserProperties := store.GetKPIConfigFromStandardUserProperties(projectID)
	rProperties := model.MergeKPIPropertiesByConsiderElementsInFirst(model.KPIPropertiesForFormSubmissions, standardUserProperties)

	config["metrics"] = rMetrics
	config["properties"] = rProperties
	return config, http.StatusOK
}
