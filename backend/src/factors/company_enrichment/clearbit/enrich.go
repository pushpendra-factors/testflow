package clearbit

import (
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

// IsEligible is a method on CustomerClearbit to check the eligibility of enrichment by customer clearbit API.
func (cb *CustomerClearbit) IsEligible(projectSettings *model.ProjectSetting) (bool, error) {

	projectId := projectSettings.ProjectId
	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_CLEARBIT, false)
	if err != nil {
		log.Error("Failed to fetch feature flag")
		return false, err
	}
	customerClearbitKey := projectSettings.ClearbitKey
	intCustomerClearbit := projectSettings.IntClearBit

	eligible := (featureFlag && *intCustomerClearbit && (customerClearbitKey != ""))

	return eligible, nil
}

// Enrich is a method on CustomerClearbit to enrich the company identification user properties.
func (cb *CustomerClearbit) Enrich(projectSettings *model.ProjectSetting, userProperties *U.PropertiesMap, userId, clientIP string) {

	projectId := projectSettings.ProjectId
	customerClearbitAPIKey := projectSettings.ClearbitKey

	FillClearbitUserProperties(projectId, customerClearbitAPIKey, userProperties, userId, clientIP)
	(*userProperties)[U.ENRICHMENT_SOURCE] = API_CLEARBIT
}

func FillClearbitUserProperties(projectId int64, clearbitKey string,
	userProperties *U.PropertiesMap, UserId string, clientIP string) (string, int) {

	logCtx := log.WithField("project_id", projectId)

	resultChannel := make(chan clear_bit.ResultChannel)
	var res clear_bit.ResultChannel
	clearBitExists, _ := clear_bit.GetClearbitCacheResult(projectId, UserId, clientIP)
	if !clearBitExists {

		go clear_bit.ExecuteClearBitEnrichV1(projectId, clearbitKey, userProperties, clientIP, resultChannel, logCtx)

		select {
		case res = <-resultChannel:
			if res.ExecuteStatus == 1 {

				clear_bit.SetClearBitCacheResult(projectId, UserId, clientIP)
			} else {
				logCtx.Warn("ExecuteClearbit failed in Track call")
			}
		case <-time.After(U.TimeoutOneSecond):
			logCtx.Warn("clear_bit enrichment timed out in Track call")
		}
	}
	return res.Domain, res.ExecuteStatus
}
