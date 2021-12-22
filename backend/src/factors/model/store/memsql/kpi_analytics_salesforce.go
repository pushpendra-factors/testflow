package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForSalesforceUsers(projectID uint64, reqID string) (map[string]interface{}, int) {
	return store.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceUsersDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForSalesforceAccounts(projectID uint64, reqID string) (map[string]interface{}, int) {
	return store.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceAccountsDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForSalesforceOpportunities(projectID uint64, reqID string) (map[string]interface{}, int) {
	return store.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceOpportunitiesDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForSalesforce(projectID uint64, reqID string, displayCategory string) (map[string]interface{}, int) {
	salesforceProjectSettings, errCode := store.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting salesforce project settings.")
		return nil, http.StatusOK
	}
	if len(salesforceProjectSettings) == 0 {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn("Salesforce integration is not available.")
		return nil, http.StatusOK
	}

	return store.getConfigForSpecificSalesforceCategory(projectID, reqID, displayCategory), http.StatusOK
}

func (store *MemSQL) getConfigForSpecificSalesforceCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		"properties":       store.getPropertiesForSalesforce(projectID, reqID),
	}
}

func (store *MemSQL) getPropertiesForSalesforce(projectID uint64, reqID string) []map[string]string {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	properties, propertiesToDisplayNames, err := store.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$salesforce")
}
