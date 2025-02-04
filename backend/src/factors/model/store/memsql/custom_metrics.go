package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) CreateCustomMetric(customMetric model.CustomMetric) (*model.CustomMetric, string, int) {
	logFields := log.Fields{
		"custom_metric": customMetric,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	logCtx.Warn("Create Custom Metric")

	// objectType used as sectionDisplayCategory in db
	if customMetric.SectionDisplayCategory != "" {
		customMetric.ObjectType = customMetric.SectionDisplayCategory
	}
	db := C.GetServices().Db
	errMsg, isValidCustomMetric := model.ValidateCustomMetric(customMetric)
	if !isValidCustomMetric {
		logCtx.WithField("customMetric", customMetric).Warn(errMsg)
		return &model.CustomMetric{}, errMsg, http.StatusBadRequest
	}
	customMetric.ID = uuid.New().String()
	err := db.Create(&customMetric).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			log.WithError(err).WithField("customMetric", customMetric).Error("Failed to create custom metric. Duplicate.")
			return &model.CustomMetric{}, err.Error(), http.StatusConflict
		}
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
	err := db.Order("name ASC").Where("project_id = ?", projectID).Find(&customMetrics).Error
	if err != nil {
		logCtx.WithError(err).WithField("projectID", projectID).Warn("Failed while retrieving custom metrics.")
		return make([]model.CustomMetric, 0), err.Error(), http.StatusInternalServerError
	}
	// setting sectionDisplayCategory as objectType, as object type in db is used for sectionDisplayCategory
	for i := range customMetrics {
		customMetrics[i].SectionDisplayCategory = customMetrics[i].ObjectType
	}
	return customMetrics, "", http.StatusFound
}

func (store *MemSQL) GetCustomMetricAndDerivedMetricByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string {
	return append(store.getCustomKPIMetricsByProjectIdAndDisplayCategory(projectID, displayCategory), store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, displayCategory, includeDerivedKPIs)...)
}

func (store *MemSQL) GetCustomEventAndDerivedMetricByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string {
	return append(store.getCustomEventKPIMetricsByProjectIdAndDisplayCategory(projectID, displayCategory), store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, displayCategory, includeDerivedKPIs)...)
}

func (store *MemSQL) getCustomKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string) []map[string]string {
	logCtx := log.WithField("project_id", projectID)
	customMetrics, err, statusCode := store.getCustomMetricByProjectIdQueryTypeAndObjectType(projectID, model.ProfileQueryType, displayCategory)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
	}
	return store.getKPIMetricsFromCustomMetric(customMetrics, model.KpiCustomQueryType)
}

func (store *MemSQL) GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string, includeDerivedKPIs bool) []map[string]string {
	logCtx := log.WithField("project_id", projectID)
	if !includeDerivedKPIs {
		return make([]map[string]string, 0)
	} else {
		customMetrics, err, statusCode := store.getCustomMetricByProjectIdQueryTypeAndObjectType(projectID, model.DerivedQueryType, displayCategory)
		if statusCode != http.StatusFound {
			logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
		}
		return store.getKPIMetricsFromCustomMetric(customMetrics, model.KpiDerivedQueryType)
	}
}

func (store *MemSQL) getCustomEventKPIMetricsByProjectIdAndDisplayCategory(projectID int64, displayCategory string) []map[string]string {
	logCtx := log.WithField("project_id", projectID)
	customMetrics, err, statusCode := store.getCustomMetricByProjectIdQueryTypeAndObjectType(projectID, model.EventBasedQueryType, displayCategory)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
	}
	return store.getKPIMetricsFromCustomMetric(customMetrics, model.KpiCustomQueryType)
}

func (store *MemSQL) getCustomMetricByProjectIdQueryTypeAndObjectType(projectID int64, queryType int, objectType string) ([]model.CustomMetric, string, int) {
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
	// setting sectionDisplayCategory as objectType, as object type in db is used for sectionDisplayCategory
	for i := range customMetrics {
		customMetrics[i].SectionDisplayCategory = customMetrics[i].ObjectType
	}
	return customMetrics, "", http.StatusFound
}

func (store *MemSQL) getKPIMetricsFromCustomMetric(customMetrics []model.CustomMetric, kpiQueryType string) []map[string]string {
	rCustomMetrics := model.GetKPIConfig(customMetrics)
	for i := range rCustomMetrics {
		rCustomMetrics[i]["kpi_query_type"] = kpiQueryType
	}
	return rCustomMetrics
}

func (store *MemSQL) getDerivedKPIMetricsByProjectId(projectID int64) []model.CustomMetric {
	logCtx := log.WithField("project_id", projectID)
	customMetrics, err, statusCode := store.GetCustomMetricByProjectIdAndQueryType(projectID, model.DerivedQueryType)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).Warn("Failed to get the custom Metric by object type")
	}
	return customMetrics
}

func (store *MemSQL) GetDisplayCategoriesByProjectIdAndNameFromDerivedCustomKPI(projectID int64, name string) ([]string, string) {
	logCtx := log.WithField("project_id", projectID)

	displayCategories := []string{}
	derivedMetric, errMsg, status := store.GetDerivedCustomMetricByProjectIdName(projectID, name)
	if status != http.StatusFound {
		logCtx.WithField("err", errMsg).Warn("Failed to get the derived Metric by name")
		return nil, errMsg
	}
	var derivedMetricTransformation model.KPIQueryGroup
	err := U.DecodePostgresJsonbToStructType(derivedMetric.Transformations, &derivedMetricTransformation)
	if err != nil {
		logCtx.WithField("err", err).Warn("Error during decode of derived metrics transformations.")
		return nil, "Error during decode of derived metrics transformations."
	}
	for _, kpiQuery := range derivedMetricTransformation.Queries {
		displayCategories = append(displayCategories, kpiQuery.DisplayCategory)
	}
	return displayCategories, ""
}

func (store *MemSQL) GetProfileCustomMetricByProjectIdName(projectID int64, name string) (model.CustomMetric, string, int) {
	return store.getCustomMetricByProjectIdNameAndQueryType(projectID, name, model.ProfileQueryType)
}

func (store *MemSQL) GetDerivedCustomMetricByProjectIdName(projectID int64, name string) (model.CustomMetric, string, int) {
	return store.getCustomMetricByProjectIdNameAndQueryType(projectID, name, model.DerivedQueryType)
}

func (store *MemSQL) GetEventBasedCustomMetricByProjectIdName(projectID int64, name string) (model.CustomMetric, string, int) {
	return store.getCustomMetricByProjectIdNameAndQueryType(projectID, name, model.EventBasedQueryType)
}

func (store *MemSQL) getCustomMetricByProjectIdNameAndQueryType(projectID int64, name string, queryType int) (model.CustomMetric, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"name":       name,
		"query_type": queryType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	if projectID == 0 {
		return model.CustomMetric{}, "Invalid project ID for custom metric", http.StatusBadRequest
	}

	var customMetric model.CustomMetric
	err := db.Where("project_id = ? AND type_of_query = ? AND name = ?", projectID, queryType, name).Find(&customMetric).Error
	if err == gorm.ErrRecordNotFound {
		return customMetric, err.Error(), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Warn("Failed while retrieving custom metrics.")
		return customMetric, err.Error(), http.StatusInternalServerError
	}
	// setting sectionDisplayCategory as objectType, as object type in db is used for sectionDisplayCategory
	customMetric.SectionDisplayCategory = customMetric.ObjectType

	return customMetric, "", http.StatusFound
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
	// setting sectionDisplayCategory as objectType, as object type in db is used for sectionDisplayCategory
	for i := range customMetrics {
		customMetrics[i].SectionDisplayCategory = customMetrics[i].ObjectType
	}
	return customMetrics, "", http.StatusFound
}

// TODO: need to check here if need to be updated for custom events
// Note: Relying on fact that there is a unique name exists for kpi.
func (store *MemSQL) GetKpiRelatedCustomMetricsByName(projectID int64, name string) (model.CustomMetric, string, int) {
	logCtx := log.WithField("projectID", projectID)
	db := C.GetServices().Db
	if projectID == 0 {
		return model.CustomMetric{}, "Invalid project ID for custom metric", http.StatusBadRequest
	}
	var customMetric model.CustomMetric
	arrayOfCRMObjects := make([]interface{}, 0)
	arrayOfCRMObjects = append(arrayOfCRMObjects, model.HubspotContactsDisplayCategory, model.HubspotCompaniesDisplayCategory,
		model.HubspotDealsDisplayCategory,
		model.SalesforceUsersDisplayCategory, model.SalesforceAccountsDisplayCategory, model.SalesforceOpportunitiesDisplayCategory)
	stmnt := "project_id = ? AND type_of_query IN (?, ?) AND name = ? AND object_type IN (?, ?, ?, ?, ?, ?) "
	args := []interface{}{projectID, model.ProfileQueryType, model.DerivedQueryType, name}
	args = append(args, arrayOfCRMObjects...)
	err := db.Where(stmnt, args...).Find(&customMetric).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Warn("Failed while retrieving custom metrics.")
			return customMetric, err.Error(), http.StatusInternalServerError
		}
		return customMetric, err.Error(), http.StatusNotFound
	}
	// setting sectionDisplayCategory as objectType, as object type in db is used for sectionDisplayCategory
	customMetric.SectionDisplayCategory = customMetric.ObjectType

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
	// setting sectionDisplayCategory as objectType, as object type in db is used for sectionDisplayCategory
	customMetric.SectionDisplayCategory = customMetric.ObjectType

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
	customMetrics := store.getDerivedKPIMetricsByProjectId(projectID)
	rCustomMetrics := make([]string, 0)

	for _, customMetric := range customMetrics {
		customKpi := model.DecodeCustomMetricsTransformation(customMetric)
		if customKpi.ContainsNameInInternalTransformation(customMetricName) {
			rCustomMetrics = append(rCustomMetrics, customMetric.Name)
		}
	}

	return rCustomMetrics
}
