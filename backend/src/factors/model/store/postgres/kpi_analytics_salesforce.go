package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) GetKPIConfigsForSalesforce(projectID uint64, reqID string) (map[string]interface{}, int) {
	hubspotProjectSettings, errCode := pg.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(hubspotProjectSettings) == 0 {
		return nil, http.StatusOK
	}
	finalResult := make(map[string]interface{}, 0)
	for _, displayCategory := range model.DisplayCategoriesForSalesforce {
		finalResult = U.MergeJSONMaps(finalResult, pg.getConfigForSpecificSalesforceCategory(projectID, reqID, displayCategory))
	}
	return finalResult, http.StatusOK
}

// Move to model?
// Should we still have properties - because user_properties are dynamic here?
func (pg *Postgres) getConfigForSpecificSalesforceCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	return map[string]interface{}{
		"category":         model.EventCategory,
		"display_category": displayCategory,
		"metrics":          model.GetMetricsForDisplayCategory(displayCategory),
		// "properties":       pg.getPropertiesForHubspot(projectID, reqID),
	}
}

func (pg *Postgres) getPropertiesForSalesforce(projectID uint64, reqID string) []map[string]string {
	logCtx := log.WithField("req_id", reqID).WithField("project_id", projectID)
	properties, propertiesToDisplayNames, err := pg.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames)
}
