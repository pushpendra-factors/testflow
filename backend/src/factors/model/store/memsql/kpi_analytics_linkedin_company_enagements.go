package memsql

import (
	C "factors/config"
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
	if !C.IsLinkedinMemberCompanyConfigEnabled(projectID) {
		return nil, http.StatusOK
	}
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
	linkedinObjectsAndProperties := store.buildObjectAndPropertiesForLinkedinCompanyEngagements(projectID, model.ObjectsForLinkedinCompany)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(linkedinObjectsAndProperties)

	rMetrics := model.GetKPIMetricsForLinkedin()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.LinkedinCompanyEngagementsDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
