package postgres

import (
	"factors/model/model"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForFormSubmissions(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForFormSubmissions
	config["metrics"] = model.GetMetricsForDisplayCategory(model.FormSubmissionsDisplayCategory)
	return config, http.StatusOK
}
