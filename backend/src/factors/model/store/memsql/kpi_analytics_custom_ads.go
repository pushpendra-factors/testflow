package memsql

import (
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForCustomAds(projectID int64, reqID string, includeDerivedKPIs bool) ([]map[string]interface{}, int) {
	isCustomAdsAvailable := store.IsCustomAdsAvailable(projectID)
	if !isCustomAdsAvailable {
		log.WithField("projectId", projectID).Warn("custom ads integration not available.")
		return nil, http.StatusOK
	}
	configs := store.GetKPIConfigsForCustomAdsFromDB(projectID, includeDerivedKPIs)
	for _, config := range configs {
		CustomadsObjectsAndProperties := store.buildObjectAndPropertiesForCustomAds(projectID, config["display_category"].(string), model.ObjectsForCustomAds)
		properties := model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(CustomadsObjectsAndProperties)
		config["properties"] = properties
	}
	return configs, http.StatusOK
}

func (store *MemSQL) GetKPIConfigsForCustomAdsFromDB(projectID int64, includeDerivedKPIs bool) []map[string]interface{} {
	configs := make([]map[string]interface{}, 0)
	adsImportList, _ := store.GetCustomAdsSourcesByProject(projectID)
	for _, source := range adsImportList {
		config := map[string]interface{}{
			"category":         model.CustomChannelCategory,
			"display_category": source,
		}
		allChannelMetrics := model.GetStaticallyDefinedMetricsForDisplayCategory(model.AllChannelsDisplayCategory)

		rMetrics := append(allChannelMetrics, model.GetStaticallyDefinedMetricsForDisplayCategory(model.CustomAdsDisplayCategory)...)
		rMetrics = append(rMetrics, store.GetDerivedKPIMetricsByProjectIdAndDisplayCategory(projectID, source, includeDerivedKPIs)...)

		config["metrics"] = rMetrics
		configs = append(configs, config)
	}
	return configs
}
