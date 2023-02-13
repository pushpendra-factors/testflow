package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForSalesforceUsers(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceUsersDisplayCategory, includeDerivedKPIs)
}

func (store *MemSQL) GetKPIConfigsForSalesforceAccounts(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceAccountsDisplayCategory, includeDerivedKPIs)
}

func (store *MemSQL) GetKPIConfigsForSalesforceOpportunities(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceOpportunitiesDisplayCategory, includeDerivedKPIs)
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
func (store *MemSQL) GetKPIConfigsForSalesforce(projectID int64, reqID string, displayCategory string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	salesforceProjectSettings, errCode := store.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting salesforce project settings.")
		return nil, http.StatusOK
	}
	if len(salesforceProjectSettings) == 0 {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn("Salesforce integration is not available.")
		return nil, http.StatusOK
	}

	return store.getConfigForSpecificSalesforceCategory(projectID, reqID, displayCategory, includeDerivedKPIs), http.StatusOK
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
// Only considering hubspot_contacts and salesforce_users for now.
func (store *MemSQL) getConfigForSpecificSalesforceCategory(projectID int64, reqID string, displayCategory string, includeDerivedKPIs bool) map[string]interface{} {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rMetrics := store.GetCustomMetricAndDerivedMetricByProjectIdAndDisplayCategory(projectID, displayCategory, includeDerivedKPIs)

	return map[string]interface{}{
		"category":         model.ProfileCategory,
		"display_category": displayCategory,
		"metrics":          rMetrics,
		"properties":       store.getPropertiesForSalesforceByDisplayCategory(projectID, reqID, displayCategory),
	}
}

func (store *MemSQL) getPropertiesForSalesforceByDisplayCategory(projectID int64, reqID, displayCategory string) []map[string]string {
	finalProperties := make([]map[string]string, 0)
	standardUserProperties := store.GetKPIConfigFromStandardUserProperties(projectID)
	switch displayCategory {
	case model.SalesforceOpportunitiesDisplayCategory:
		finalProperties = store.GetPropertiesForSalesforceOpportunities(projectID, reqID)
	case model.SalesforceAccountsDisplayCategory:
		finalProperties = store.GetPropertiesForSalesforceAccounts(projectID, reqID)
	case model.SalesforceUsersDisplayCategory:
		finalProperties = append(standardUserProperties, store.GetPropertiesForSalesforceUsers(projectID, reqID)...)
	default:
		log.WithFields(log.Fields{"project_id": projectID, "req_id": reqID, "display_category": displayCategory}).
			Error("Invalid category on GetPropertiesForSalesforceByDisplayCategory.")
		return finalProperties
	}

	return finalProperties
}

func (store *MemSQL) GetPropertiesForSalesforceUsers(projectID int64, reqID string) []map[string]string {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	properties, propertiesToDisplayNames, err := store.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$salesforce")
}

func (store *MemSQL) GetPropertiesForSalesforceOpportunities(projectID int64, reqID string) []map[string]string {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	groupProperties, status := store.GetPropertiesByGroup(projectID, model.GetGroupNameByMetricSectionDisplayCategory(model.SalesforceOpportunitiesDisplayCategory), 2500,
		C.GetLookbackWindowForEventUserCache())
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get salesforce opportunities properties. Internal error")
		return make([]map[string]string, 0)
	}

	displayNamesOp := make(map[string]string)
	_, displayNames := store.GetDisplayNamesForAllUserProperties(projectID)
	for _, properties := range groupProperties {
		for _, property := range properties {
			if _, exist := displayNames[property]; exist {
				displayNamesOp[property] = displayNames[property]
			}
		}
	}

	_, displayNames = store.GetDisplayNamesForObjectEntities(projectID)
	for _, properties := range groupProperties {
		for _, property := range properties {
			if _, exist := displayNames[property]; exist {
				displayNamesOp[property] = displayNames[property]
			}
		}
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(groupProperties, displayNamesOp, "$salesforce")
}

func (store *MemSQL) GetPropertiesForSalesforceAccounts(projectID int64, reqID string) []map[string]string {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	groupProperties, status := store.GetPropertiesByGroup(projectID, model.GetGroupNameByMetricSectionDisplayCategory(model.SalesforceAccountsDisplayCategory), 2500,
		C.GetLookbackWindowForEventUserCache())
	if status != http.StatusFound {
		logCtx.WithField("err_code", status).Error("Failed to get salesforce account properties. Internal error")
		return make([]map[string]string, 0)
	}

	displayNamesOp := make(map[string]string)
	_, displayNames := store.GetDisplayNamesForAllUserProperties(projectID)
	for _, properties := range groupProperties {
		for _, property := range properties {
			if _, exist := displayNames[property]; exist {
				displayNamesOp[property] = displayNames[property]
			}
		}
	}

	_, displayNames = store.GetDisplayNamesForObjectEntities(projectID)
	for _, properties := range groupProperties {
		for _, property := range properties {
			if _, exist := displayNames[property]; exist {
				displayNamesOp[property] = displayNames[property]
			}
		}
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(groupProperties, displayNamesOp, "$salesforce")
}
