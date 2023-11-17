package sixsignal

import (
	"factors/integration/six_signal"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_6SIGNAL = "API_6Sense"

type CustomerSixSignal struct {
}

func (ss *CustomerSixSignal) IsEligible(projectSettings *model.ProjectSetting) (bool, error) {

	projectId := projectSettings.ProjectId
	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_SIX_SIGNAL, false)
	if err != nil {
		log.Error("Failed to fetch feature flag")
		return false, err
	}
	customerSixSignalKey := projectSettings.Client6SignalKey
	intCustomerSixSignal := projectSettings.IntClientSixSignalKey

	eligible := (featureFlag && *intCustomerSixSignal && (customerSixSignalKey != ""))

	return eligible, nil
}

func (ss *CustomerSixSignal) Enrich(projectSettings *model.ProjectSetting,
	userProperties *U.PropertiesMap, userId, clientIP string) {

	projectId := projectSettings.ProjectId
	customerSixSignalAPIKey := projectSettings.Client6SignalKey

	FillSixSignalUserProperties(projectId, customerSixSignalAPIKey, userProperties, userId, clientIP)
	(*userProperties)[U.ENRICHMENT_SOURCE] = API_6SIGNAL
}

func FillSixSignalUserProperties(projectId int64, apiKey string, userProperties *U.PropertiesMap,
	UserId, clientIP string) (string, int) {

	logCtx := log.WithField("project_id", projectId)
	resultChannel := make(chan six_signal.ResultChannel)
	var res six_signal.ResultChannel
	sixSignalExists, _ := model.GetSixSignalCacheResult(projectId, UserId, clientIP)
	if !sixSignalExists {

		go six_signal.ExecuteSixSignalEnrichV1(projectId, apiKey, userProperties, clientIP, resultChannel)
		select {
		case res := <-resultChannel:
			if res.ExecuteStatus == 1 {
				model.SetSixSignalCacheResult(projectId, UserId, clientIP)
				logCtx.Info("ExecuteSixSignal suceeded in track call")

			} else {
				logCtx.Warn("ExecuteSixSignal failed in track call")
			}
		case <-time.After(U.TimeoutOneSecond):
			logCtx.Warn("six_Signal enrichment timed out in Track call")

		}
	}

	return res.Domain, res.ExecuteStatus
}
