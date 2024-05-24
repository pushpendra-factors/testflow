package clearbit

import (
	"factors/config"
	"factors/integration/clear_bit"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_CLEARBIT = "clearbit_api"

type CustomerClearbit struct {
}

/*
Customer clearbit enrichment and metering flow:
	1. IsEligible method checks for the eligibility creteria. If eligible,
	2. Enrich method fetches the API Key and call the method to fill the company identification props.
	3. FillClearbitUserProperties fill the props on the basis of the API Key.
*/

// IsEligible is a method on CustomerClearbit to check the eligibility of enrichment by customer clearbit API.
func (cb *CustomerClearbit) IsEligible(projectSettings *model.ProjectSetting, logCtx *log.Entry) (bool, error) {

	projectId := projectSettings.ProjectId

	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_CLEARBIT, false)
	if err != nil {
		logCtx.Error("Failed to fetch feature flag")
		return false, err
	}
	if !featureFlag {
		return false, nil
	}
	customerClearbitKey := projectSettings.ClearbitKey
	intCustomerClearbit := projectSettings.IntClearBit

	eligible := (featureFlag && *intCustomerClearbit && (customerClearbitKey != ""))
	if config.IsEnrichmentDebugLogsEnabled(projectId) && !eligible {
		logCtx.Warn("Eligibility check failed for customer clearbit")
	}

	return eligible, nil
}

// Enrich is a method on CustomerClearbit to enrich the company identification user properties.
func (cb *CustomerClearbit) Enrich(projectSettings *model.ProjectSetting, userProperties *U.PropertiesMap, userId, clientIP string, logCtx *log.Entry) (string, int) {

	projectId := projectSettings.ProjectId
	customerClearbitAPIKey := projectSettings.ClearbitKey

	if config.IsEnrichmentDebugLogsEnabled(projectId) {
		logCtx.Info("Enrichment using customer clearbit.")
	}

	domain, status := FillClearbitUserProperties(projectId, customerClearbitAPIKey, userProperties, userId, clientIP, logCtx)
	(*userProperties)[U.ENRICHMENT_SOURCE] = API_CLEARBIT
	return domain, status
}

// FillClearbitUserProperties checks the cache and if not present it calls the goroutine func for enrichment.
func FillClearbitUserProperties(projectId int64, clearbitKey string,
	userProperties *U.PropertiesMap, UserId string, clientIP string, logCtx *log.Entry) (string, int) {

	resultChannel := make(chan clear_bit.ResultChannel)
	var res clear_bit.ResultChannel
	clearBitExists, _ := clear_bit.GetClearbitCacheResult(projectId, UserId, clientIP)
	if !clearBitExists {

		go clear_bit.ExecuteClearBitEnrichV1(projectId, clearbitKey, userProperties, clientIP, resultChannel, logCtx)

		select {
		case res = <-resultChannel:
			if res.ExecuteStatus == 1 {
				clear_bit.SetClearBitCacheResult(projectId, UserId, clientIP)
				if config.IsEnrichmentDebugLogsEnabled(projectId) {
					logCtx.Info("ExecuteClearbit success in Track Call.")
				}
			} else {
				if config.IsEnrichmentDebugLogsEnabled(projectId) {
					logCtx.Warn("ExecuteClearbit failed in Track call")
				}

			}
		case <-time.After(U.TimeoutOneSecond):
			if config.IsEnrichmentDebugLogsEnabled(projectId) {
				logCtx.Warn("clear_bit enrichment timed out in Track call")
			}
		}
	} else if config.IsEnrichmentDebugLogsEnabled(projectId) && clearBitExists {
		logCtx.Info("Getting the enrichment data from user props.")
	}
	return res.Domain, res.ExecuteStatus
}
