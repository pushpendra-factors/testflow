package factors_deanon

import (
	"errors"
	"factors/config"
	"factors/model/model"
	"fmt"
	"net/http"

	cacheRedis "factors/cache/redis"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const ACCOUNT_LIMIT_PARTIAL_EXCEEDED = "75_percent_limit_exceeded"
const ACCOUNT_LIMIT_FULLY_EXCEEDED = "100_percent_limit_exceeded"
const MAILMODO_75_PERCENT_EXCEEDED_CAMPAIGN_ID = "7b6f7b1d-b87c-54f1-87ae-14d6911b1a3e"
const MAILMODO_100_PERCENT_EXCEEDED_CAMPAIGN_ID = "a7b3aef4-1d15-538f-b01c-55868cfd5c9c"
const EMAIL_TYPE_TRANSACTIONAL = "transactional"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
type PartialAccountLimitExceeded struct {
	Client HTTPClient
}

type FullAccountLimitExceeded struct {
	Client HTTPClient
}

// Match method on struct of PartialAccountLimitExceeded checks the cases if the acount limit exceeds 75% or not and alert is not sent.
func (p *PartialAccountLimitExceeded) Match(projectId int64, percentageExhausted float64, limit int64, timeZone U.TimeZoneString, logCtx *log.Entry) (bool, error) {

	sendAlert := false
	var err error
	if percentageExhausted >= 75 && percentageExhausted < 100 {
		sendAlert, err = ShouldSendAccountLimitAlert(projectId, limit, ACCOUNT_LIMIT_PARTIAL_EXCEEDED, timeZone, logCtx)
		if err != nil {
			return false, err
		}
	}

	return sendAlert, nil
}

// Match method on struct of FullAccountLimitExceeded checks the cases if the acount limit exceeds 100% or not and alert is not sent.
func (f *FullAccountLimitExceeded) Match(projectId int64, percentageExhausted float64, limit int64, timeZone U.TimeZoneString, logCtx *log.Entry) (bool, error) {

	alertNotSent := false
	var err error
	if percentageExhausted >= 100 {
		alertNotSent, err = ShouldSendAccountLimitAlert(projectId, limit, ACCOUNT_LIMIT_FULLY_EXCEEDED, timeZone, logCtx)
		if err != nil {
			return false, err
		}
	}

	return alertNotSent, nil
}

// Execute method on struct of PartialAccountLimitExceeded forms the request for and gets the response from Mailmodo if account limit exceeds 75%.
func (p *PartialAccountLimitExceeded) Execute(projectId int64, payloadJSON []byte, logCtx *log.Entry) error {

	request, err := model.FormMailmodoTriggerCampaignRequest(MAILMODO_75_PERCENT_EXCEEDED_CAMPAIGN_ID, payloadJSON)
	if err != nil {
		return err
	}

	response, err := p.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("failed sending account limit email alert")
	}

	if config.IsEnrichmentDebugLogsEnabled(projectId) {
		logCtx.Info("Partial Account limit exceeded email alert sent successfully.")
	}
	return nil
}

// Execute method on struct of FullAccountLimitExceeded forms the request for and gets the response from Mailmodo if account limit exceeds 100%.
func (f *FullAccountLimitExceeded) Execute(projectId int64, payloadJSON []byte, logCtx *log.Entry) error {

	request, err := model.FormMailmodoTriggerCampaignRequest(MAILMODO_100_PERCENT_EXCEEDED_CAMPAIGN_ID, payloadJSON)
	if err != nil {
		return err
	}

	response, err := f.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("failed sending account limit email alert")
	}

	if config.IsEnrichmentDebugLogsEnabled(projectId) {
		logCtx.Info("Full Account limit exceeded email alert sent successfully.")
	}

	return nil
}

// ShouldSendAccountLimitAlert checks if the alert has already been sent or not.
func ShouldSendAccountLimitAlert(projectId int64, limit int64, exhaustType string, timeZone U.TimeZoneString, logCtx *log.Entry) (bool, error) {

	key, err := GetAccountLimitEmailAlertCacheKey(projectId, limit, exhaustType, timeZone, logCtx)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for account limit email alert")
		return false, err
	}
	exists, err := cacheRedis.ExistsPersistent(key)
	if err != nil {
		logCtx.WithError(err).Error("Failed to check existence of cache key for account limit email alert")
		return false, err
	}
	return !exists, nil
}

// GetAccountLimitEmailAlertCacheKey gets the cache key for the account limit email alert.
func GetAccountLimitEmailAlertCacheKey(projectId int64, limit int64, exhaustType string, timeZone U.TimeZoneString, logCtx *log.Entry) (*cacheRedis.Key, error) {

	monthYear := U.GetCurrentMonthYear(timeZone)
	prefix := fmt.Sprintf("alert:%s:limit:%v", exhaustType, limit)
	key, err := cacheRedis.NewKey(projectId, prefix, monthYear)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for account limit email alert")
		return nil, err
	}

	return key, nil
}

// SetAccountLimitEmailAlertCacheKey sets the value of the cache key to true, which indicates the alert has been sent.
func SetAccountLimitEmailAlertCacheKey(projectId int64, limit int64, exhaustType string, timeZone U.TimeZoneString, logCtx *log.Entry) error {

	key, err := GetAccountLimitEmailAlertCacheKey(projectId, limit, exhaustType, timeZone, logCtx)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key for account limit email alert")
		return err
	}
	// three month
	var expiry float64 = 3 * 31 * 24 * 60 * 60
	err = cacheRedis.SetPersistent(key, "true", expiry)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache key for account limit email alert")
		return err
	}
	return nil
}
