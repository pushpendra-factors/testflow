package postgres

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForHubspot(projectID uint64, reqID string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := pg.GetAllHubspotProjectSettingsForProjectID(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		return nil, http.StatusOK
	}
	finalResult := make(map[string]interface{}, 0)
	for _, displayCategory := range model.DisplayCategoriesForHubspot {
		finalResult = U.MergeJSONMaps(finalResult, pg.getConfigForSpecificHubspotCategory(projectID, reqID, displayCategory))
	}
	return finalResult, http.StatusOK
}

func (pg *Postgres) getConfigForSpecificHubspotCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		"properties":       make([]string, 0),
	}
}

// getRequiredUserProperties gives response in following format - each data_type with multiple property_names.
// propertiesToDisplayNames - is having defined list of propertyNames to displaNames.
// Using both the above to transform into name, display_name, cateogry.
// func (pg *Postgres) getPropertiesForHubspot(projectID uint64, reqID string) []map[string]string {
// 	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
// 	properties, propertiesToDisplayNames, err := pg.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
// 	if err != nil {
// 		logCtx.WithError(err).Error("Failed to get hubspot properties. Internal error")
// 		return make([]map[string]string, 0)
// 	}

// 	// transforming to kpi structure.
// 	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames)
// }
