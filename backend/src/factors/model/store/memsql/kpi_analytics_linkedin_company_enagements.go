package memsql

import (
	"factors/model/model"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForLinkedinCompanyEngagements(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
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
	config := model.KpiLinkedinCompanyEngagementsConfig
	log.Warn("kark5-linkedin")
	linkedinObjectsAndProperties := store.buildObjectAndPropertiesForLinkedinCompanyEngagements(projectID, model.ObjectsForLinkedinCompany)
	log.Warn("kark5-linkedin-2")
	config["properties"] = model.TransformLinkedinChannelsPropertiesConfigToKpiPropertiesConfig(linkedinObjectsAndProperties)
	log.Warn("kark5-linkedin-3")

	rMetrics := model.GetKPIMetricsForLinkedin()
	log.Warn("kark5-linkedin-4")
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.LinkedinCompanyEngagementsDisplayCategory, includeDerivedKPIs)...)
	log.Warn("kark5-linkedin-5")

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
