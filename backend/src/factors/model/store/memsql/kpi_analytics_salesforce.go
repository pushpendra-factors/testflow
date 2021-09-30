package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForSalesforce(projectID uint64, reqID string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := store.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		return nil, http.StatusOK
	}
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": model.SalesforceDisplayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(model.SalesforceDisplayCategory),
		"properties":       store.getPropertiesForSalesforce(projectID, reqID),
	}, http.StatusOK
}

func (store *MemSQL) getPropertiesForSalesforce(projectID uint64, reqID string) []map[string]string {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	properties, propertiesToDisplayNames, err := store.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames)
}
