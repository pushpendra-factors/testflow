package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForHubspotContacts(projectID int64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotContactsDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForHubspotCompanies(projectID int64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotCompaniesDisplayCategory)
}

func (store *MemSQL) GetKPIConfigsForHubspotDeals(projectID int64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForHubspot(projectID, reqID, model.HubspotDealsDisplayCategory)
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
func (store *MemSQL) GetKPIConfigsForHubspot(projectID int64, reqID string, displayCategory string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	hubspotProjectSettings, errCode := store.GetAllHubspotProjectSettingsForProjectID(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting hubspot project settings.")
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn("Hubspot integration is not available.")
		return nil, http.StatusOK
	}

	return store.getConfigForSpecificHubspotCategory(projectID, reqID, displayCategory), http.StatusOK
}

// Removed constants for hubspot and salesforce kpi metrics in PR - pull/3984.
// Only considering hubspot_contacts and salesforce_users for now.
func (store *MemSQL) getConfigForSpecificHubspotCategory(projectID int64, reqID string, displayCategory string) map[string]interface{} {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	customMetrics, err, statusCode := store.GetCustomMetricByProjectIdAndObjectType(projectID, model.ProfileQueryType, displayCategory)
	if statusCode != http.StatusFound {
		logCtx.WithField("err", err).WithField("displayCategory", displayCategory).Warn("Failed to get the custom Metric by object type")
	}
	rCustomMetrics := make([]map[string]string, 0)

	for _, customMetric := range customMetrics {
		currentMetric := make(map[string]string)
		currentMetric["name"] = customMetric.Name
		currentMetric["display_name"] = customMetric.Name
		currentMetric["type"] = ""
		rCustomMetrics = append(rCustomMetrics, currentMetric)
	}

	return map[string]interface{}{
		"category":         model.ProfileCategory,
		"display_category": displayCategory,
		"metrics":          rCustomMetrics,
		"properties":       store.GetPropertiesForHubspot(projectID, reqID),
	}
}

func (store *MemSQL) GetPropertiesForHubspot(projectID int64, reqID string) []map[string]string {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	properties, propertiesToDisplayNames, err := store.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$hubspot")
}
