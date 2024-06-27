package demandbase

import (
	"factors/config"
	"factors/integration/demandbase"
	"factors/model/model"
	"factors/model/store"
	"time"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const API_DEMANDBASE = "API_Demandbase"

type CustomerDemandbase struct {
}

/*
Customer Demandbase enrichment flow:
	1. IsEligible method checks for the eligibility creteria. If eligible,
	2. Enrich method fetches the API Key and call the method to fill the company identification props.
	3. FillClearbitUserProperties fill the props on the basis of the API Key.
*/

// IsEligible method checks the eligibility creteria for enrichment via Demandbase integration.
func (cd *CustomerDemandbase) IsEligible(projectSettings *model.ProjectSetting, logCtx *log.Entry) (bool, error) {

	projectId := projectSettings.ProjectId

	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_DEMANDBASE, false)
	if err != nil {
		log.Error("Failed to fetch feature flag for demandbase feature")
		return false, err
	}
	if !featureFlag {
		return false, nil
	}
	customerDemandbaseKey := projectSettings.ClientDemandbaseKey
	intClientDemandbase := projectSettings.IntClientDemandbase

	eligible := (featureFlag && *intClientDemandbase && (customerDemandbaseKey != ""))
	if config.IsEnrichmentDebugLogsEnabled(projectId) && !eligible {
		logCtx.Warn("Eligibility check failed for customer demandbase")
	}

	return eligible, nil

}

// Enrich is a method on CustomerDemandbase to enrich the company identification user properties.
func (cd *CustomerDemandbase) Enrich(projectSettings *model.ProjectSetting,
	userProperties *U.PropertiesMap, userId, clientIP string, logCtx *log.Entry) (string, int) {

	projectId := projectSettings.ProjectId
	customerDemandbaseKey := projectSettings.ClientDemandbaseKey

	if config.IsEnrichmentDebugLogsEnabled(projectId) {
		logCtx.Info("Enrichment using customer demandbase.")
	}

	domain, status := FillDemandbaseUserProperties(projectId, customerDemandbaseKey, userProperties, userId, clientIP, logCtx)

	(*userProperties)[U.ENRICHMENT_SOURCE] = API_DEMANDBASE
	return domain, status
}

// FillDemandbaseUserProperties checks if the cache exists, if not it executes the enrich method using goroutine.
func FillDemandbaseUserProperties(projectId int64, apiKey string, userProperties *U.PropertiesMap,
	UserId, clientIP string, logCtx *log.Entry) (string, int) {

	resultChannel := make(chan demandbase.ResultChannel)
	var res demandbase.ResultChannel
	demandbaseExists, _ := GetDemandbaseRedisCacheResult(projectId, UserId, clientIP)

	if !demandbaseExists {
		go demandbase.ExecuteDemandbaseEnrich(projectId, apiKey, userProperties, clientIP, resultChannel, logCtx)
		select {
		case res = <-resultChannel:
			if res.ExecuteStatus == 1 {
				if config.IsEnrichmentDebugLogsEnabled(projectId) {
					logCtx.Info("ExecuteDemandbase suceeded in track call")
				}

			} else {
				if config.IsEnrichmentDebugLogsEnabled(projectId) {
					logCtx.Warn("ExecuteDemandbase failed in track call")
				}
			}
		case <-time.After(U.TimeoutOneSecond):
			if config.IsEnrichmentDebugLogsEnabled(projectId) {
				logCtx.Warn("Demandbase enrichment timed out in Track call")
			}

		}
	} else if config.IsEnrichmentDebugLogsEnabled(projectId) && demandbaseExists {
		logCtx.Info("Getting the enrichment data from user props.")
	}

	return res.Domain, res.ExecuteStatus
}
