package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForAdwords(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	adwordsSettings, errCode := store.GetIntAdwordsProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(adwordsSettings) == 0 {
		return nil, http.StatusOK
	}
	config := model.KpiAdwordsConfig
	adwordsObjectsAndProperties := store.buildObjectAndPropertiesForAdwords(projectID, model.ObjectsForAdwords)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(adwordsObjectsAndProperties, "Adwords")

	rMetrics := model.GetKPIMetricsForAdwords()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.GoogleAdsDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
