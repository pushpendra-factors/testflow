package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForHubspot(projectID uint64, reqID string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := store.GetAllHubspotProjectSettingsForProjectID(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		return nil, http.StatusOK
	}
	finalResult := make(map[string]interface{}, 0)
	for _, displayCategory := range model.DisplayCategoriesForHubspot {
		finalResult = U.MergeJSONMaps(finalResult, store.getConfigForSpecificHubspotCategory(projectID, reqID, displayCategory))
	}
	return finalResult, http.StatusOK
}

func (store *MemSQL) getConfigForSpecificHubspotCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		"properties":       make([]string, 0),
	}
}
