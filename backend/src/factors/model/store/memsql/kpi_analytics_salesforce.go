package memsql

import (
	"factors/model/model"
	"net/http"
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
		return nil, http.StatusOK
	}
	if len(salesforceProjectSettings) == 0 {
		return nil, http.StatusOK
	}

	return store.getConfigForSpecificSalesforceCategory(projectID, reqID, displayCategory), http.StatusOK
}

func (store *MemSQL) getConfigForSpecificSalesforceCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		"properties":       make([]string, 0),
	}
}
