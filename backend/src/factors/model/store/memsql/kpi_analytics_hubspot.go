package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
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
		"properties":       store.getPropertiesForHubspot(projectID, reqID),
	}
}

func (store *MemSQL) getPropertiesForHubspot(projectID uint64, reqID string) []map[string]string {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	properties, propertiesToDisplayNames, err := store.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$hubspot")
}
