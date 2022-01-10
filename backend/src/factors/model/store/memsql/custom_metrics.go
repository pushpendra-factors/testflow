package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateCustomMetric(customMetric model.CustomMetric) (*model.CustomMetric, string, int) {
	logCtx := log.WithField("project_id", customMetric.ProjectID)
	db := C.GetServices().Db
	errMsg, isValidCustomMetric := model.ValidateCustomMetric(customMetric)
	if !isValidCustomMetric {
		logCtx.WithField("customMetric", customMetric).Warn(errMsg)
		return &model.CustomMetric{}, errMsg, http.StatusBadRequest
	}
	customMetric.TypeOfQuery = 1
	customMetric.ID = uuid.New().String()
	err := db.Create(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).WithField("customMetric", customMetric).Warn("Failed while creating custom metric.")
		return &model.CustomMetric{}, err.Error(), http.StatusInternalServerError
	}
	return &customMetric, "", http.StatusCreated
}

func (store *MemSQL) GetCustomMetricsByProjectId(projectID uint64) ([]model.CustomMetric, string, int) {
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
