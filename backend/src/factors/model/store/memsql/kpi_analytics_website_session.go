package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForWebsiteSessions(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForWebsiteSessions
	config["metrics"] = model.GetMetricsForDisplayCategory(model.WebsiteSessionDisplayCategory)
	return config, http.StatusOK
}
