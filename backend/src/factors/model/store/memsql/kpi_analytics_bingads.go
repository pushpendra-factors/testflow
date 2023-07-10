package memsql

import (
	"factors/model/model"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForBingAds(projectID int64, reqID string, includeDerivedKPIs bool) (map[string]interface{}, int) {
	projectIDInString := []string{strconv.FormatInt(projectID, 10)}
	isBingAdsIntegrationDone := store.IsBingIntegrationAvailable(projectID)
	if !isBingAdsIntegrationDone {
		log.WithField("projectId", projectIDInString).Warn("Bingads integration not available.")
		return nil, http.StatusOK
	}
	config := model.KpiBingAdsConfig
	BingadsObjectsAndProperties := store.buildObjectAndPropertiesForBingAds(projectID, model.ObjectsForBingads)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(BingadsObjectsAndProperties, "BingAds")

	rMetrics := model.GetKPIMetricsForBingAds()
	rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, model.BingAdsDisplayCategory, includeDerivedKPIs)...)

	config["metrics"] = rMetrics
	return config, http.StatusOK
}
