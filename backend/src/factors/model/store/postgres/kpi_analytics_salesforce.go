package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) GetKPIConfigsForSalesforceUsers(projectID uint64, reqID string) (map[string]interface{}, int) {
	return pg.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceUsersDisplayCategory)
}

func (pg *Postgres) GetKPIConfigsForSalesforceAccounts(projectID uint64, reqID string) (map[string]interface{}, int) {
	return pg.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceAccountsDisplayCategory)
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
func (pg *Postgres) GetKPIConfigsForSalesforceOpportunities(projectID uint64, reqID string) (map[string]interface{}, int) {
	return pg.GetKPIConfigsForSalesforce(projectID, reqID, model.SalesforceOpportunitiesDisplayCategory)
}

func (pg *Postgres) GetKPIConfigsForSalesforce(projectID uint64, reqID string, displayCategory string) (map[string]interface{}, int) {
	salesforceProjectSettings, errCode := pg.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting salesforce project settings.")
		return nil, http.StatusOK
	}
	if len(salesforceProjectSettings) == 0 {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn("Salesforce integration is not available.")
		return nil, http.StatusOK
	}

	return pg.getConfigForSpecificSalesforceCategory(projectID, reqID, displayCategory), http.StatusOK
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
// Only considering hubspot_contacts and salesforce_users for now.
func (pg *Postgres) getConfigForSpecificSalesforceCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	customMetrics, err, statusCode := pg.GetCustomMetricByProjectIdAndObjectType(projectID, model.ProfileQueryType, displayCategory)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
	}
	customMetricNames := make([]string, 0)
	for _, customMetric := range customMetrics {
		customMetricNames = append(customMetricNames, customMetric.Name)
	}

	return map[string]interface{}{
		"category":         model.ProfileCategory,
		"display_category": displayCategory,
		"metrics":          customMetricNames,
		"properties":       pg.GetPropertiesForSalesforce(projectID, reqID),
	}
}

func (pg *Postgres) GetPropertiesForSalesforce(projectID uint64, reqID string) []map[string]string {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	properties, propertiesToDisplayNames, err := pg.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$salesforce")
}
