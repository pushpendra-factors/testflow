package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) GetKPIConfigsForHubspotContacts(projectID uint64, reqID string) (map[string]interface{}, int) {
	return pg.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotContactsDisplayCategory)
}

func (pg *Postgres) GetKPIConfigsForHubspotCompanies(projectID uint64, reqID string) (map[string]interface{}, int) {
	return pg.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotCompaniesDisplayCategory)
}

func (pg *Postgres) GetKPIConfigsForHubspotDeals(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return pg.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotDealsDisplayCategory)
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
func (pg *Postgres) GetKPIConfigsForHubspot(projectID uint64, reqID string, displayCategory string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := pg.GetAllHubspotProjectSettingsForProjectID(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting hubspot project settings.")
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn("Hubspot integration is not available.")
		return nil, http.StatusOK
	}

	return pg.getConfigForSpecificHubspotCategory(projectID, reqID, displayCategory), http.StatusOK
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
// Only considering hubspot_contacts and salesforce_users for now.
func (pg *Postgres) getConfigForSpecificHubspotCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
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
		"properties":       pg.GetPropertiesForHubspot(projectID, reqID),
	}
}

// getRequiredUserProperties gives response in following format - each data_type with multiple property_names.
// propertiesToDisplayNames - is having defined list of propertyNames to displaNames.
// Using both the above to transform into name, display_name, cateogry.
func (pg *Postgres) GetPropertiesForHubspot(projectID uint64, reqID string) []map[string]string {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	properties, propertiesToDisplayNames, err := pg.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$hubspot")
}
