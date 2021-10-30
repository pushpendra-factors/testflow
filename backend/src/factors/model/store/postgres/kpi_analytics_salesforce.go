package postgres

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
)

func (pg *Postgres) GetKPIConfigsForSalesforce(projectID uint64, reqID string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := pg.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		return nil, http.StatusOK
	}
	finalResult := make(map[string]interface{}, 0)
	for _, displayCategory := range model.DisplayCategoriesForSalesforce {
		finalResult = U.MergeJSONMaps(finalResult, pg.getConfigForSpecificSalesforceCategory(projectID, reqID, displayCategory))
	}
	return finalResult, http.StatusOK
}

func (pg *Postgres) getConfigForSpecificSalesforceCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		"properties":       make([]string, 0),
	}
}
