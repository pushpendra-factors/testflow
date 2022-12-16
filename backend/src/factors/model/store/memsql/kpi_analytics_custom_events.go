package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForCustomEvents(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	config := model.KpiCustomEventsConfig
	config["properties"] = make([]map[string]string, 0)

	rMetrics := store.GetCustomEventAndDerivedMetricByProjectIdAndDisplayCategory(projectID, model.EventsBasedDisplayCategory, includeDerivedKPIs)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
