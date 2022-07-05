package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) RunTemplateQuery(projectID int64, query model.TemplateQuery, reqID string) (model.TemplateResponse, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if query.Type == model.TemplateAliasToType[model.SEMChecklist] {
		templateResponse, errCode := store.ExecuteAdwordsSEMChecklistQuery(projectID, query, reqID)
		if errCode != http.StatusOK {
			return model.TemplateResponse{}, errCode
		}
		return templateResponse, http.StatusOK
	}
	return model.TemplateResponse{}, http.StatusOK
}

//get the list of metrics and thresholds for that project in the form of { metrics: [], thresholds:[]}
func (store *MemSQL) GetTemplateConfig(projectID int64, templateType int) (model.TemplateConfig, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"template_type": templateType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 || templateType < 1 || templateType > 1 {
		return model.TemplateConfig{}, http.StatusBadRequest
	}
	var templateConfig model.TemplateConfig
	templateConfig.Metrics = model.TemplateMetricsForAdwordsWithDisplayName
	templateThresholds, err := store.getTemplateThresholds(projectID, templateType)
	if err != nil {
		return model.TemplateConfig{}, http.StatusInternalServerError
	}
	templateConfig.Thresholds = templateThresholds

	return templateConfig, http.StatusOK
}
func (store *MemSQL) getTemplateThresholds(projectID int64, templateType int) ([]model.TemplateThreshold, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"template_type": templateType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var templateThresholds []model.TemplateThreshold
	db := C.GetServices().Db
	err := db.Table("templates").Select("thresholds").Where("project_id = ? AND type = ?", projectID, templateType).Find(&templateThresholds).Error
	if err != nil {
		return []model.TemplateThreshold{}, err
	}
	return templateThresholds, nil
}

func (store *MemSQL) UpdateTemplateConfig(projectID int64, templateType int, thresholds []model.TemplateThreshold) ([]model.TemplateThreshold, string) {
	logFields := log.Fields{
		"project_id":    projectID,
		"template_type": templateType,
		"thresholds":    thresholds,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	isValidConfig := model.ValidateTemplateThresholds(thresholds)
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
