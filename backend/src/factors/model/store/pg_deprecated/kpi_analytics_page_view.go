package postgres

import (
	"factors/model/model"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForPageViews(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForPageViews
	config["metrics"] = model.GetMetricsForDisplayCategory(model.PageViewsDisplayCategory)
	return config, http.StatusOK
}
