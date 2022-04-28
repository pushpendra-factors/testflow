package postgres

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForWebsiteSessions(projectID uint64, reqID string) (map[string]interface{}, int) {
	config := model.KPIConfigForWebsiteSessions
	kpiPropertiesFromContentGroup := pg.getWebsiteRelatedContentGroupPropertiesForKPI(projectID)
	config["properties"] = append(model.KPIPropertiesForWebsiteSessions, kpiPropertiesFromContentGroup...)
	config["metrics"] = model.GetMetricsForDisplayCategory(model.WebsiteSessionDisplayCategory)
	return config, http.StatusOK
}

func (pg *Postgres) getWebsiteRelatedContentGroupPropertiesForKPI(projectID uint64) []map[string]string {
	contentGroups, _ := pg.GetAllContentGroups(projectID)

	resultantKPIProperties := make([]map[string]string, 0)
	// contentGroupNames := make([]string, 0)
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
func (pg *Postgres) ValidateKPISessions(projectID uint64, kpiQuery model.KPIQuery) bool {
	return model.ValidateKPIQueryMetricsForWebsiteSession(kpiQuery.Metrics) ||
		pg.validateKPIQueryFiltersForWebsiteSession(projectID, kpiQuery.Filters) ||
		pg.validateKPIQueryGroupByForWebsiteSession(projectID, kpiQuery.GroupBy)
}

func (pg *Postgres) validateKPIQueryFiltersForWebsiteSession(projectID uint64, kpiQueryFilters []model.KPIFilter) bool {
	mapOfPropertyName := make(map[string]struct{})
	for _, propertyData := range model.KPIPropertiesForWebsiteSessions {
		mapOfPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, filter := range kpiQueryFilters {
		if _, exists := mapOfPropertyName[filter.PropertyName]; !exists {
			if !pg.IsDuplicateNameCheck(projectID, filter.PropertyName) {
				return false
			}
		}
	}
	return true
}

func (pg *Postgres) validateKPIQueryGroupByForWebsiteSession(projectID uint64, kpiQueryGroupBys []model.KPIGroupBy) bool {
	mapOfPredefinedKPIPropertyName := make(map[string]struct{})
	for _, propertyData := range model.KPIPropertiesForWebsiteSessions {
		mapOfPredefinedKPIPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, groupBy := range kpiQueryGroupBys {
		if _, exists := mapOfPredefinedKPIPropertyName[groupBy.PropertyName]; !exists {
			if !pg.IsDuplicateNameCheck(projectID, groupBy.PropertyName) {
				return false
			}
		}
	}
	return true
}
