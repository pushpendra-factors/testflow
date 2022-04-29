package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) GetPropertiesForMarketo(projectID uint64, reqID string) []map[string]string {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	properties, propertiesToDisplayNames, err := pg.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get marketo properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$marketo")
}

func (pg *Postgres) GetKPIConfigsForMarketoLeads(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return pg.GetKPIConfigsForMarketo(projectID, reqID, model.MarketoLeadsDisplayCategory)
}

func (pg *Postgres) GetKPIConfigsForMarketo(projectID uint64, reqID string, displayCategory string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	connectorId, err := pg.GetFiveTranMapping(projectID, model.MarketoIntegration)
	if err != nil || connectorId == "" {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting marketo project settings.")
		return nil, http.StatusOK
	}

	return pg.getConfigForSpecificMarketoCategory(projectID, reqID, displayCategory), http.StatusOK
}

func (pg *Postgres) getConfigForSpecificMarketoCategory(projectID uint64, reqID string, displayCategory string) map[string]interface{} {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
		"properties":       pg.GetPropertiesForMarketo(projectID, reqID),
	}
}
