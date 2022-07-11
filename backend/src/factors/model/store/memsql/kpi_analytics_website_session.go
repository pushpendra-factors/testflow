package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForWebsiteSessions(projectID int64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	config := model.KPIConfigForWebsiteSessions
	kpiPropertiesFromContentGroup := store.getWebsiteRelatedContentGroupPropertiesForKPI(projectID)
	config["properties"] = append(model.KPIPropertiesForWebsiteSessions, kpiPropertiesFromContentGroup...)
	config["metrics"] = model.GetMetricsForDisplayCategory(model.WebsiteSessionDisplayCategory)
	return config, http.StatusOK
}

func (store *MemSQL) getWebsiteRelatedContentGroupPropertiesForKPI(projectID int64) []map[string]string {
	contentGroups, _ := store.GetAllContentGroups(projectID)

	resultantKPIProperties := make([]map[string]string, 0)
	for _, contentGroup := range contentGroups {
		currentKPIProperty := make(map[string]string)
		currentKPIProperty["name"] = contentGroup.ContentGroupName
		currentKPIProperty["display_name"] = contentGroup.ContentGroupName
		currentKPIProperty["data_type"] = U.PropertyTypeCategorical
		currentKPIProperty["entity"] = model.EventEntity
		resultantKPIProperties = append(resultantKPIProperties, currentKPIProperty)
	}
	return resultantKPIProperties
}

// Validation methods for website session starts here.
func (store *MemSQL) ValidateKPISessions(projectID int64, kpiQuery model.KPIQuery) bool {
	return model.ValidateKPIQueryMetricsForWebsiteSession(kpiQuery.Metrics) ||
		store.validateKPIQueryFiltersForWebsiteSession(projectID, kpiQuery.Filters) ||
		store.validateKPIQueryGroupByForWebsiteSession(projectID, kpiQuery.GroupBy)
}

func (store *MemSQL) validateKPIQueryFiltersForWebsiteSession(projectID int64, kpiQueryFilters []model.KPIFilter) bool {
	mapOfPredefinedKPIPropertyName := make(map[string]struct{})
	for _, propertyData := range model.KPIPropertiesForWebsiteSessions {
		mapOfPredefinedKPIPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, filter := range kpiQueryFilters {
		if _, exists := mapOfPredefinedKPIPropertyName[filter.PropertyName]; !exists {
			if !store.IsDuplicateNameCheck(projectID, filter.PropertyName) {
				return false
			}
		}
	}
	return true
}

func (store *MemSQL) validateKPIQueryGroupByForWebsiteSession(projectID int64, kpiQueryGroupBys []model.KPIGroupBy) bool {
	mapOfPropertyName := make(map[string]struct{})
	for _, propertyData := range model.KPIPropertiesForWebsiteSessions {
		mapOfPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, groupBy := range kpiQueryGroupBys {
		if _, exists := mapOfPropertyName[groupBy.PropertyName]; !exists {
			if !store.IsDuplicateNameCheck(projectID, groupBy.PropertyName) {
				return false
			}
		}
	}
	return true
}
