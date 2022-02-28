package postgres

import (
	"factors/model/model"
	"net/http"
	"strconv"
	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) GetKPIConfigsForBingAds(projectID uint64, reqID string) (map[string]interface{}, int) {
	projectIDInString := []string{strconv.FormatUint(projectID, 10)}
	isBingAdsIntegrationDone := pg.IsBingIntegrationAvailable(projectID)
	if !isBingAdsIntegrationDone{
		log.WithField("projectId", projectIDInString).Warn("Bingads integration not available.")
		return nil, http.StatusOK
	}
	config := model.GetKPIConfigsForBingAds()
	BingadsObjectsAndProperties := pg.buildObjectAndPropertiesForBingAds(projectID, model.ObjectsForBingads)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(BingadsObjectsAndProperties)
	return config, http.StatusOK
}
