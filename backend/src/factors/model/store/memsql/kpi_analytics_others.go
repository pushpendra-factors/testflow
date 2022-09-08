package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForOthers(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	config := model.KpiOtherConfig
	config["properties"] = make([]map[string]string, 0)

	rMetrics := store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.OthersDisplayCategory, includeDerivedKPIs)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
