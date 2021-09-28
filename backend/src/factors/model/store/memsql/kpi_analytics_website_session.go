package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForWebsiteSessions(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForWebsiteSessions
	config["metrics"] = model.GetMetricsForDisplayCategory(model.WebsiteSessionDisplayCategory)
	return model.AddObjectTypeToProperties(config, U.EVENT_NAME_SESSION), http.StatusOK
}
