package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetPropertiesForMarketo(projectID int64, reqID string) []map[string]string {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	properties, propertiesToDisplayNames, err := store.GetRequiredUserPropertiesByProject(projectID, 2500, C.GetLookbackWindowForEventUserCache())
	if err != nil {
		logCtx.WithError(err).Error("Failed to get marketo properties. Internal error")
		return make([]map[string]string, 0)
	}

	// transforming to kpi structure.
	return model.TransformCRMPropertiesToKPIConfigProperties(properties, propertiesToDisplayNames, "$marketo")
}

func (store *MemSQL) GetKPIConfigsForMarketoLeads(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetKPIConfigsForMarketo(projectID, reqID, model.MarketoLeadsDisplayCategory, includeDerivedKPIs)
}

func (store *MemSQL) GetKPIConfigsForMarketo(projectID int64, reqID string, displayCategory string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	connectorId, err := store.GetFiveTranMapping(projectID, model.MarketoIntegration)
	if err != nil || connectorId == "" {
		log.WithField("projectId", projectID).WithField("reqID", reqID).Warn(" Failed in getting marketo project settings.")
		return nil, http.StatusOK
	}

	return store.getConfigForSpecificMarketoCategory(projectID, reqID, displayCategory, includeDerivedKPIs), http.StatusOK
}

func (store *MemSQL) getConfigForSpecificMarketoCategory(projectID int64, reqID string, displayCategory string, includeDerivedKPIs bool) map[string]interface{} {
	logFields := log.Fields{
		"project_id":       projectID,
		"req_id":           reqID,
		"display_category": displayCategory,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rMetrics := store.GetCustomMetricAndDerivedMetricByProjectIdAndDisplayCategory(projectID, displayCategory, includeDerivedKPIs)
	marketoProperties := store.GetPropertiesForMarketo(projectID, reqID)
	standardUserProperties := store.GetKPIConfigFromStandardUserProperties(projectID)

	return map[string]interface{}{
		"category":         model.ProfileCategory,
		"display_category": displayCategory,
		"metrics":          rMetrics,
		"properties":       append(standardUserProperties, marketoProperties...),
	}
}
