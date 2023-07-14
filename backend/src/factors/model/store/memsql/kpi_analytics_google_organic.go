package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForGoogleOrganic(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	settings, errCode := store.GetIntGoogleOrganicProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	config := model.KpiGoogleOrganicConfig
	organicObjectsAndProperties := store.buildObjectAndPropertiesForGoogleOrganic(model.ObjectsForGoogleOrganic)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(organicObjectsAndProperties, "Google Organic")

	rMetrics := model.GetKPIMetricsForGoogleOrganic()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.GoogleOrganicDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
