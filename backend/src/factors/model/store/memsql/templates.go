package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	"github.com/jinzhu/gorm/dialects/postgres"
)

var templateMetrics = []string{
	model.Clicks,
	model.Impressions,
	model.ClickThroughRate,
	model.CostPerClick,
	model.SearchImpressionShare,
	"cost",
	"leads",
	"cost_per_lead",
	"click_to_lead_rate",
}
var templateMetricsMap = map[string]bool{
	model.Clicks:                true,
	model.Impressions:           true,
	model.ClickThroughRate:      true,
	model.CostPerClick:          true,
	model.SearchImpressionShare: true,
	"cost":                      true,
	"leads":                     true,
	"cost_per_lead":             true,
	"click_to_lead_rate":        true,
}

func (store *MemSQL) RunTemplateQuery(projectID uint64, query model.TemplateQuery, reqID string) (model.TemplateResponse, int) {
	if query.Type == 1 {
		if query.Metric == "leads" {
			return model.MockResponseLeads, http.StatusOK
		} else {
			return model.MockResponse, http.StatusOK
		}
	}
	return model.TemplateResponse{}, http.StatusOK
}

//get the list of metrics and thresholds for that project in the form of { metrics: [], thresholds:[]}
func (store *MemSQL) GetTemplateConfig(projectID uint64, templateType int) (model.TemplateConfig, int) {
	if projectID == 0 || templateType < 1 || templateType > 1 {
		return model.TemplateConfig{}, http.StatusBadRequest
	}
	var templateConfig model.TemplateConfig
	templateConfig.Metrics = templateMetrics
	templateThresholds, err := store.getTemplateThresholds(projectID, templateType)
	if err != nil {
		return model.TemplateConfig{}, http.StatusInternalServerError
	}
	templateConfig.Thresholds = templateThresholds

	return templateConfig, http.StatusOK
}
func (store *MemSQL) getTemplateThresholds(projectID uint64, templateType int) ([]model.TemplateThreshold, error) {
	var templateThresholds []model.TemplateThreshold
	db := C.GetServices().Db
	err := db.Table("templates").Select("thresholds").Where("project_id = ? AND type = ?", projectID, templateType).Find(&templateThresholds).Error
	if err != nil {
		return []model.TemplateThreshold{}, err
	}
	return templateThresholds, nil
}

//validates if the thresholds metric is part of allowed metrics and is not repeated. e.g: [{metric: clicks}, {metric: clicks}] not allowed
func validateTemplateThresholds(thresholds []model.TemplateThreshold) bool {
	metricsCountMap := make(map[string]int)
	for _, threshold := range thresholds {
		_, isExistsMetric := templateMetricsMap[threshold.Metric]
		if !isExistsMetric {
			return false
		}
		metricsCountMap[threshold.Metric]++
	}
	for _, count := range metricsCountMap {
		if count > 1 {
			return false
		}
	}
	return true
}
func (store *MemSQL) UpdateTemplateConfig(projectID uint64, templateType int, thresholds []model.TemplateThreshold) ([]model.TemplateThreshold, string) {
	isValidConfig := validateTemplateThresholds(thresholds)
	if !isValidConfig {
		return []model.TemplateThreshold{}, "Invalid config input"
	}
	var template model.Template
	template.ProjectID = projectID
	template.Type = templateType
	jsonThresholds, err := json.Marshal(thresholds)
	if err != nil {
		return []model.TemplateThreshold{}, "Failed to encode thresholds"
	}
	template.Thresholds = &postgres.Jsonb{jsonThresholds}
	db := C.GetServices().Db
	if db.Table("templates").Where("project_id = ? AND type = ?", projectID, templateType).Update(&template).RowsAffected == 0 {
		err = db.Table("templates").Create(&template).Error
		if err != nil {
			return []model.TemplateThreshold{}, "Failed to update thresholds in db"
		}
	}
	var updatedThresholds []model.TemplateThreshold
	err = U.DecodePostgresJsonbToStructType(template.Thresholds, &updatedThresholds)
	if err != nil {
		return []model.TemplateThreshold{}, "Failed to decode thresholds"
	}
	return updatedThresholds, ""
}
