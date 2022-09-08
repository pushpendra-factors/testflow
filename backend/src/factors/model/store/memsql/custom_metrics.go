package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateCustomMetric(customMetric model.CustomMetric) (*model.CustomMetric, string, int) {
	logFields := log.Fields{
		"custom_metric": customMetric,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	errMsg, isValidCustomMetric := model.ValidateCustomMetric(customMetric)
	if !isValidCustomMetric {
		logCtx.WithField("customMetric", customMetric).Warn(errMsg)
		return &model.CustomMetric{}, errMsg, http.StatusBadRequest
	}
	customMetric.ID = uuid.New().String()
	err := db.Create(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).WithField("customMetric", customMetric).Warn("Failed while creating custom metric.")
		return &model.CustomMetric{}, err.Error(), http.StatusInternalServerError
	}
	return &customMetric, "", http.StatusCreated
}

func (store *MemSQL) GetCustomMetricsByProjectId(projectID int64) ([]model.CustomMetric, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	if projectID == 0 {
		return make([]model.CustomMetric, 0), "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetrics []model.CustomMetric
	err := db.Where("project_id = ?", projectID).Find(&customMetrics).Error
	if err != nil {
		logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving custom metrics.")
		return make([]model.CustomMetric, 0), err.Error(), http.StatusInternalServerError
	}
	return customMetrics, "", http.StatusFound
}

func (store *MemSQL) GetCustomMetricAndDerivedMetricByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string {
	return append(store.GetCustomKPIMetricsByProjectIdAndDisplayCategory(projectID, displayCategory), store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, displayCategory, includeDerivedKPIs)...)
}

func (store *MemSQL) GetCustomKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string) []map[string]string {
	logCtx := log.WithField("project_id", projectID)
	customMetrics, err, statusCode := store.GetCustomMetricByProjectIdQueryTypeAndObjectType(projectID, model.ProfileQueryType, displayCategory)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
	}
	rCustomMetrics := model.GetKPIConfig(customMetrics)
	for i := range rCustomMetrics {
		rCustomMetrics[i]["kpi_query_type"] = model.KpiCustomQueryType
	}
	return rCustomMetrics
}

func (store *MemSQL) GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string {
	logCtx := log.WithField("project_id", projectID)
	if includeDerivedKPIs == true {
		return make([]map[string]string, 0)
	} else {
		customMetrics, err, statusCode := store.GetCustomMetricByProjectIdQueryTypeAndObjectType(projectID, model.DerivedQueryType, displayCategory)
		if statusCode != http.StatusFound {
			logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
		}
		rCustomMetrics := model.GetKPIConfig(customMetrics)
		for i := range rCustomMetrics {
			rCustomMetrics[i]["kpi_query_type"] = model.KpiDerivedQueryType
		}
		return rCustomMetrics
	}
}

func (store *MemSQL) GetCustomMetricByProjectIdQueryTypeAndObjectType(projectID int64, queryType int, objectType string) ([]model.CustomMetric, string, int) {
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

func (store *MemSQL) GetDerivedKPIMetricsByProjectId(projectID int64) []model.CustomMetric {
	logCtx := log.WithField("project_id", projectID)
	customMetrics, err, statusCode := store.GetCustomMetricByProjectIdAndQueryType(projectID, model.DerivedQueryType)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).Warn("Failed to get the custom Metric by object type")
	}
	return customMetrics
}

func (store *MemSQL) GetCustomMetricByProjectIdAndQueryType(projectID int64, queryType int) ([]model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	customMetrics := make([]model.CustomMetric, 0, 0)
	if projectID == 0 {
		return customMetrics, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	err := db.Where("project_id = ? AND type_of_query = ? ", projectID, queryType).Find(&customMetrics).Error
	if err != nil {
		logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving custom metrics.")
		return make([]model.CustomMetric, 0), err.Error(), http.StatusInternalServerError
	}
	return customMetrics, "", http.StatusFound
}

// Note: Relying on fact that there is a unique name exists for kpi.
func (store *MemSQL) GetKpiRelatedCustomMetricsByName(projectID int64, name string) (model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.CustomMetric{}, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetric model.CustomMetric
	err := db.Where("project_id = ? AND type_of_query IN (?, ?) AND name = ?", projectID, model.ProfileQueryType, model.DerivedQueryType, name).Find(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).Warn("Failed while retrieving custom metrics.")
		return customMetric, err.Error(), http.StatusInternalServerError
	}
	return customMetric, "", http.StatusFound
}

func (store *MemSQL) GetDerivedMetricsByName(projectID int64, name string) (model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.CustomMetric{}, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetric model.CustomMetric
	err := db.Where("project_id = ? AND type_of_query = ? AND name = ?", projectID, model.DerivedQueryType, name).Find(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).Warn("Failed while retrieving custom metrics.")
		return customMetric, err.Error(), http.StatusInternalServerError
	}
	return customMetric, "", http.StatusFound
}

// TODO lets see if unique index can be used for fetching.
func (store *MemSQL) GetCustomMetricsByID(projectID int64, id string) (model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.CustomMetric{}, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetric model.CustomMetric
	err := db.Where("project_id = ? AND id = ?", projectID, id).Find(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).Warn("Failed while retrieving custom metrics.")
		return customMetric, err.Error(), http.StatusInternalServerError
	}
	return customMetric, "", http.StatusFound
}

func (store *MemSQL) DeleteCustomMetricByID(projectID int64, id string) int {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	var customMetric model.CustomMetric
	err := db.Where("project_id = ? AND id = ?", projectID, id).Delete(&customMetric).Error
	if err != nil {
		logCtx.WithError(err).Warn("Failed while deleting custom metrics.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (store *MemSQL) GetDerivedKPIsHavingNameInInternalQueries(projectID int64, customMetricName string) []string {
	customMetrics := store.GetDerivedKPIMetricsByProjectId(projectID)
	rCustomMetrics := make([]string, 0)

	for _, customMetric := range customMetrics {
		customKpi := model.DecodeCustomMetricsTransformation(customMetric)
		if customKpi.ContainsNameInInternalTransformation(customMetricName) {
			rCustomMetrics = append(rCustomMetrics, customMetric.Name)
		}
	}

	return rCustomMetrics
}
