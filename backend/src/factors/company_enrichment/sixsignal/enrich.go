package sixsignal

import (
	"factors/config"
	"factors/integration/six_signal"
	v3 "factors/integration/six_signal/v3"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_6SIGNAL = "API_6Sense"

type CustomerSixSignal struct {
}

/*
Customer clearbit enrichment and metering flow:
	1. IsEligible method checks for the eligibility creteria. If eligible,
	2. Enrich method fetches the API Key and call the method to fill the company identification props.
	3. FillSixSignalUserProperties fill the props on the basis of the API Key.
*/

// IsEligible checks the eligibilty of enrichment via customer sixsignal.
func (ss *CustomerSixSignal) IsEligible(projectSettings *model.ProjectSetting, logCtx *log.Entry) (bool, error) {

	projectId := projectSettings.ProjectId

	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_SIX_SIGNAL, false)
	if err != nil {
		log.Error("Failed to fetch feature flag")
		return false, err
	}
	if !featureFlag {
		return false, nil
	}
	customerSixSignalKey := projectSettings.Client6SignalKey
	intCustomerSixSignal := projectSettings.IntClientSixSignalKey

	eligible := (featureFlag && *intCustomerSixSignal && (customerSixSignalKey != ""))
	if config.IsEnrichmentDebugLogsEnabled(projectId) && !eligible {
		logCtx.Warn("Eligibility check failed for customer sixsignal")
	}

	return eligible, nil
}

// Enrich method fetches the customer sixsignal API Key and calls the method for enrichment via sixsignal.
func (ss *CustomerSixSignal) Enrich(projectSettings *model.ProjectSetting,
	userProperties *U.PropertiesMap, userId, clientIP string, logCtx *log.Entry) (string, int) {

	projectId := projectSettings.ProjectId
	customerSixSignalAPIKey := projectSettings.Client6SignalKey

	if config.IsEnrichmentDebugLogsEnabled(projectId) {
		logCtx.Info("Enrichment using customer sixsignal.")
	}

	domain, status := FillSixSignalUserProperties(projectId, customerSixSignalAPIKey, userProperties, userId, clientIP, logCtx)

	(*userProperties)[U.ENRICHMENT_SOURCE] = API_6SIGNAL
	return domain, status
}

// FillSixSignalUserProperties checks if the cache exists, if not it executes the enrich method using goroutine.
func FillSixSignalUserProperties(projectId int64, apiKey string, userProperties *U.PropertiesMap,
	UserId, clientIP string, logCtx *log.Entry) (string, int) {

	resultChannel := make(chan six_signal.ResultChannel)
	var res six_signal.ResultChannel
	sixSignalExists, _ := model.GetSixSignalCacheResult(projectId, UserId, clientIP)
	if !sixSignalExists {

		if config.IsSixSignalV3Enabled(projectId) {
			logCtx.Info("Enrichment by 6Signal v3.")
			go v3.ExecuteSixSignalEnrichV3(projectId, apiKey, userProperties, clientIP, resultChannel, logCtx)
		} else {
			go six_signal.ExecuteSixSignalEnrichV1(projectId, apiKey, userProperties, clientIP, resultChannel, logCtx)
		}
		select {
		case res = <-resultChannel:
			if res.ExecuteStatus == 1 {
				model.SetSixSignalCacheResult(projectId, UserId, clientIP)
				if config.IsEnrichmentDebugLogsEnabled(projectId) {
					logCtx.Info("ExecuteSixSignal suceeded in track call")
				}

			} else {
				if config.IsEnrichmentDebugLogsEnabled(projectId) {
					logCtx.Warn("ExecuteSixSignal failed in track call")
				}
			}
		case <-time.After(U.TimeoutOneSecond):
			if config.IsEnrichmentDebugLogsEnabled(projectId) {
				logCtx.Warn("six_Signal enrichment timed out in Track call")
			}

		}
	} else if config.IsEnrichmentDebugLogsEnabled(projectId) && sixSignalExists {
		logCtx.Info("Getting enrichment data from user props.")
	}

	return res.Domain, res.ExecuteStatus
}
