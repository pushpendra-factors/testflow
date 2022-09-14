package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForAllChannels(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	config := model.GetKPIConfigsForAllChannels()
	objectsAndProperties := store.buildObjectAndPropertiesForAllChannel(projectID, ObjectsForAllChannels)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(objectsAndProperties)

	rMetrics := model.GetKPIMetricsForAllChannels()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.AllChannelsDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
