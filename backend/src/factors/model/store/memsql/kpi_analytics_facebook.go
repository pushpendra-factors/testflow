package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForFacebook(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, settings, errCode := store.GetFacebookEnabledIDsAndProjectSettingsForProject([]int64{projectID})
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		return nil, http.StatusOK
	}
	config := model.KpiFacebookConfig
	facebookObjectsAndProperties := store.buildObjectAndPropertiesForFacebook(projectID, model.ObjectsForFacebook)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(facebookObjectsAndProperties)

	rMetrics := model.GetKPIMetricsForFacebook()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.FacebookDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
