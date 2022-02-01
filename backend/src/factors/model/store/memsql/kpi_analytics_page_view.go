package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForPageViews(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForPageViews
	config["metrics"] = model.GetMetricsForDisplayCategory(model.PageViewsDisplayCategory)
	return config, http.StatusOK
}
