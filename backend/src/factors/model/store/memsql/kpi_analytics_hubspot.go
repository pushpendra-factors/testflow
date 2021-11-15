package memsql

import (
	"factors/model/model"
	"net/http"
)

func (store *MemSQL) GetKPIConfigsForHubspotContacts(projectID uint64, reqID string) (map[string]interface{}, int) {
	return store.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotContactsDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForHubspotCompanies(projectID uint64, reqID string) (map[string]interface{}, int) {
	return store.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotCompaniesDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForHubspot(projectID uint64, reqID string, displayCategory string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := store.GetAllHubspotProjectSettingsForProjectID(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		return nil, http.StatusOK
	}

	return store.getConfigForSpecificHubspotCategory(projectID, reqID, displayCategory), http.StatusOK
}

func (store *MemSQL) getConfigForSpecificHubspotCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		"properties":       make([]string, 0),
	}
}
