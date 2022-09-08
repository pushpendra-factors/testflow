package memsql

import (
	"factors/model/model"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForLinkedin(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectIDInString := []string{strconv.FormatInt(projectID, 10)}
	settings, errCode := store.GetLinkedinEnabledProjectSettingsForProjects(projectIDInString)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(settings) == 0 {
		log.WithField("projectId", projectIDInString).Warn("Linkedin integration not available.")
		return nil, http.StatusOK
	}
	config := model.KpiLinkedinConfig
	linkedinObjectsAndProperties := store.buildObjectAndPropertiesForLinkedin(projectID, model.ObjectsForLinkedin)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(linkedinObjectsAndProperties)

	rMetrics := model.GetKPIMetricsForLinkedin()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.LinkedinDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
