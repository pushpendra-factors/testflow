package factors_deanon

import (
	"encoding/json"
	"errors"
	"factors/company_enrichment/clearbit"
	"factors/company_enrichment/sixsignal"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const FACTORS_6SIGNAL = "factors_6sense"
const FACTORS_CLEARBIT = "factors_clearbit"

var defaultFactorsDeanonConfig = model.FactorsDeanonConfig{
	Clearbit:  model.DeanonVendorConfig{TrafficFraction: 0.0},
	SixSignal: model.DeanonVendorConfig{TrafficFraction: 1.0}}

type FactorsDeanon struct {
}

/*
Factors Deanonymisation enrichment and metering flow:
	1. IsEligible method checks for the eligibility creteria. If eligible,
	2. Enrich method fetches the factors deanon config and call the method to fill the company identification props.
	3. FillFactorsDeanonUserProps fill the props on the basis of the config and return the domain enriched.
	4. Meter method meters the unique domain enrichment count on monthly basis, daily successful enrichment count, daily total API calls.
*/

// IsEligible method checks the eligibility creteria for enrichment via factors deanonymisation .
func (fd *FactorsDeanon) IsEligible(projectSettings *model.ProjectSetting, isoCode, pageURL string) (bool, error) {

	projectId := projectSettings.ProjectId
	logCtx := log.WithField("project_id", projectId)

	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_FACTORS_DEANONYMISATION, false)
	if err != nil {
		logCtx.Error("Failed to fetch feature flag")
		return false, err
	}
	if !featureFlag {
		return false, nil
	}

	isDeanonQuotaAvailable, err := CheckingFactorsDeanonQuotaLimit(projectId)
	if err != nil {
		logCtx.Error("Error in checking deanon quota exhausted.")
		return false, err
	}
	if !isDeanonQuotaAvailable {
		return false, nil
	}

	factorDeanonRulesJson := projectSettings.SixSignalConfig
	isFactorsDeanonRulesValid, err := ApplyFactorsDeanonRules(factorDeanonRulesJson, isoCode, pageURL, projectId)
	if err != nil {
		logCtx.Error("Error in checking deanon enrichment rules")
		return false, err
	}
	if !isFactorsDeanonRulesValid {
		return false, nil
	}

	intFactorsDeanon := projectSettings.IntFactorsSixSignalKey

	eligible := featureFlag && isDeanonQuotaAvailable && isFactorsDeanonRulesValid && *intFactorsDeanon

	return eligible, nil
}

// Enrich method fetches the factors deanon config and calls the method
// to enrich the company identification props on basis of the config.
func (fd *FactorsDeanon) Enrich(projectSettings *model.ProjectSetting,
	userProperties *U.PropertiesMap, eventProperties *U.PropertiesMap, userId, clientIP string) (string, int) {

	projectId := projectSettings.ProjectId
	var factorsDeanonConfig model.FactorsDeanonConfig
	if projectSettings.FactorsDeanonConfig != nil {
		err := json.Unmarshal(projectSettings.FactorsDeanonConfig.RawMessage, &factorsDeanonConfig)
		if err != nil {
			log.WithField("deanon_enrich_config", projectSettings.FactorsDeanonConfig).WithError(err).Error("Failed to decode deanon enrich config")
		}
	} else {
		factorsDeanonConfig = defaultFactorsDeanonConfig
	}

	domain, status := FillFactorsDeanonUserProperties(projectId, factorsDeanonConfig, projectSettings, userProperties, eventProperties, userId, clientIP)
	return domain, status
}

// Meter method meters the count of unique domain enrichment for the calendar month
// and successful domain enrichment count and total API calls count for each day.
func (fd *FactorsDeanon) Meter(projectId int64, domain string) {

	timeZone, statusCode := store.GetStore().GetTimezoneForProject(projectId)
	if statusCode != http.StatusFound {
		timeZone = U.TimeZoneStringIST
	}

	//Unique Domain metering for calendar month
	err := model.SetSixSignalMonthlyUniqueEnrichmentCount(projectId, domain, timeZone)
	if err != nil {
		log.Error("SetSixSignalMonthlyUniqueEnrichmentCount Failed.")
	}

	if domain != "" {
		// Successful domain enrichment count for each day
		model.SetSixSignalAPICountCacheResult(projectId, U.TimeZoneStringIST)
	}

	// Total successful API calls for a day
	model.SetSixSignalAPITotalHitCountCacheResult(projectId, U.TimeZoneStringIST)

}

/*
HandleAccountLimitAlert handles the email alerts for account limit if it exceeds 75% or 100% to notify the client.

	Returns :
		- http.StatusOK, nil : Successful Match and Execute.
		- http.BadRequest, error : Successful Match but Execute Failed.
		- http.StatusForbidden, error : Match Failed.
		- http.StatusInternalServerError, error: error in getting or setting internal data.
*/
func (fd *FactorsDeanon) HandleAccountLimitAlert(projectId int64, client HTTPClient) (int, error) {

	logCtx := log.WithField("project_id", projectId)

	project, status := store.GetStore().GetProject(projectId)
	if status != http.StatusFound {
		logCtx.Error("failed to get project.")
		return http.StatusInternalServerError, errors.New("failed to get project")
	}

	percentageExhausted, limit, err := GetAccountLimitAndPercentageExhausted(projectId)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	partialLimitExceeded := PartialAccountLimitExceeded{
		Client: client,
	}
	fullLimitExceeded := FullAccountLimitExceeded{
		Client: client,
	}

	email, errCode := store.GetStore().GetProjectAgentLatestAdminEmailByProjectId(projectId)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError, errors.New("failed fetching admin email by projectId")
	}

	payloadJSON, err := json.Marshal(model.MailmodoTriggerCampaignRequestPayload{ReceiverEmail: email, Data: map[string]interface{}{"Project Name": project.Name}})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if partialLimitMatch, _ := partialLimitExceeded.Match(projectId, percentageExhausted, limit, U.TimeZoneString(project.TimeZone)); partialLimitMatch {

		err := partialLimitExceeded.Execute(projectId, payloadJSON)
		if err != nil {
			logCtx.Error(err)
			return http.StatusBadRequest, err
		}

		err = SetAccountLimitEmailAlertCacheKey(projectId, limit, ACCOUNT_LIMIT_PARTIAL_EXCEEDED, U.TimeZoneString(project.TimeZone))
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil

	}

	if fullLimitMatch, _ := fullLimitExceeded.Match(projectId, percentageExhausted, limit, U.TimeZoneString(project.TimeZone)); fullLimitMatch {

		err := fullLimitExceeded.Execute(projectId, payloadJSON)
		if err != nil {
			logCtx.Error(err)
			return http.StatusBadRequest, err
		}

		err = SetAccountLimitEmailAlertCacheKey(projectId, limit, ACCOUNT_LIMIT_FULLY_EXCEEDED, U.TimeZoneString(project.TimeZone))
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	logCtx.Info("No account limit alerts sent.")
	return http.StatusForbidden, nil
}

// FillFactorsDeanonUserProperties calls the respective method for clearbit and sixsignal enrichment
// on basis of factors deanon config.
func FillFactorsDeanonUserProperties(projectId int64, factorsDeanonConfig model.FactorsDeanonConfig,
	projectSettings *model.ProjectSetting, userProperties *U.PropertiesMap, eventProperties *U.PropertiesMap, userId, clientIP string) (string, int) {

	logCtx := log.WithField("project_id", projectId)

	domain := ""
	status := 0
	count, limit, err := GetFactorsDeanonCountAndLimit(projectId)
	if err != nil {
		logCtx.Error("Error while fetching deanon count and limit")
		return domain, status
	}

	/*
		The factors clearbit key are being created manually for now. The below logic is to handle the case where
		the factors clearbit key is not present, then we don't check the traffic fraction and goes through sixsignal
		else we will check the fraction.
	*/
	if projectSettings.FactorsClearbitKey == "" {
		factors6SignalKey := C.GetFactorsSixSignalAPIKey()
		domain, status = sixsignal.FillSixSignalUserProperties(projectId, factors6SignalKey, userProperties, userId, clientIP)
		(*userProperties)[U.ENRICHMENT_SOURCE] = FACTORS_6SIGNAL
		if status == 1 {
			(*eventProperties)[U.EP_COMPANY_ENRICHED] = FACTORS_6SIGNAL
		}

	} else {
		if count < int64(float64(limit)*factorsDeanonConfig.Clearbit.TrafficFraction) {

			domain, status = clearbit.FillClearbitUserProperties(projectId, projectSettings.FactorsClearbitKey, userProperties, userId, clientIP)
			(*userProperties)[U.ENRICHMENT_SOURCE] = FACTORS_CLEARBIT
			if status == 1 {
				(*eventProperties)[U.EP_COMPANY_ENRICHED] = FACTORS_CLEARBIT
			}

		} else {
			factors6SignalKey := C.GetFactorsSixSignalAPIKey()
			domain, status = sixsignal.FillSixSignalUserProperties(projectId, factors6SignalKey, userProperties, userId, clientIP)
			(*userProperties)[U.ENRICHMENT_SOURCE] = FACTORS_6SIGNAL
			if status == 1 {
				(*eventProperties)[U.EP_COMPANY_ENRICHED] = FACTORS_6SIGNAL
			}

		}
	}
	return domain, status
}
