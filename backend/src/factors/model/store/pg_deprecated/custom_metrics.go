package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Can handle errors specifically if required.
// Need to handle deletion of customMetric. What would be the case?
func (pg *Postgres) CreateCustomMetric(customMetric model.CustomMetric) (*model.CustomMetric, string, int) {
	logCtx := log.WithField("project_id", customMetric.ProjectID)
	db := C.GetServices().Db
	errMsg, isValidCustomMetric := model.ValidateCustomMetric(customMetric)
	if !isValidCustomMetric {
		logCtx.WithField("customMetric", customMetric).Warn(errMsg)
		return &model.CustomMetric{}, errMsg, http.StatusBadRequest
	}

	customMetric.TypeOfQuery = 1
	err := db.Create(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).WithField("customMetric", customMetric).Warn("Failed while creating custom metric.")
		return &model.CustomMetric{}, err.Error(), http.StatusInternalServerError
	}
	return &customMetric, "", http.StatusCreated
}

// Considering profileQueryType as of now.
func (pg *Postgres) GetCustomMetricsByProjectId(projectID uint64) ([]model.CustomMetric, string, int) {
	logCtx := log.WithField("project_id", projectID)
	db := C.GetServices().Db
	if projectID == 0 {
		return make([]model.CustomMetric, 0), "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetrics []model.CustomMetric
	err := db.Where("project_id = ? AND type_of_query = ?", projectID, model.ProfileQueryType).Find(&customMetrics).Error
	if err != nil {
		logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving custom metrics.")
		return make([]model.CustomMetric, 0), err.Error(), http.StatusInternalServerError
	}
	return customMetrics, "", http.StatusFound
}

func (pg *Postgres) GetCustomMetricByProjectIdAndObjectType(projectID uint64, queryType int, objectType string) ([]model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	customMetrics := make([]model.CustomMetric, 0, 0)
	if projectID == 0 {
		return customMetrics, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	err := db.Where("project_id = ? AND type_of_query = ? AND object_type = ? ", projectID, queryType, objectType).Find(&customMetrics).Error
	if err != nil {
		logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving custom metrics.")
		return make([]model.CustomMetric, 0), err.Error(), http.StatusInternalServerError
	}
	return customMetrics, "", http.StatusFound
}

// TODO lets see if unique index can be used for fetching.
func (pg *Postgres) GetCustomMetricsByName(projectID uint64, name string) (model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.CustomMetric{}, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetric model.CustomMetric
	err := db.Where("project_id = ? AND type_of_query = 1 AND name = ?", projectID, name).Find(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).Warn("Failed while retrieving custom metrics.")
		return customMetric, err.Error(), http.StatusInternalServerError
	}
	return customMetric, "", http.StatusFound
}
