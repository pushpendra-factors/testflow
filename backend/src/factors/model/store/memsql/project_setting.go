package memsql

import (
	"database/sql"
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/metrics"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

const (
	sqlNull string = "NULL"
)

func (store *MemSQL) satisfiesProjectSettingForeignConstraints(setting model.ProjectSetting) int {
	logFields := log.Fields{
		"setting": setting,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(setting.ProjectId)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}

	if setting.IntAdwordsEnabledAgentUUID != nil && *setting.IntAdwordsEnabledAgentUUID != "" {
		_, agentErrCode := store.GetAgentByUUID(*setting.IntAdwordsEnabledAgentUUID)
		if agentErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}

	if setting.IntFacebookAgentUUID != nil && *setting.IntFacebookAgentUUID != "" {
		_, agentErrCode := store.GetAgentByUUID(*setting.IntFacebookAgentUUID)
		if agentErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}

	if setting.IntLinkedinAgentUUID != nil && *setting.IntLinkedinAgentUUID != "" {
		_, agentErrCode := store.GetAgentByUUID(*setting.IntLinkedinAgentUUID)
		if agentErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}

	if setting.IntSalesforceEnabledAgentUUID != nil && *setting.IntSalesforceEnabledAgentUUID != "" {
		_, agentErrCode := store.GetAgentByUUID(*setting.IntSalesforceEnabledAgentUUID)
		if agentErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

func (store *MemSQL) GetProjectSetting(projectId int64) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectId).First(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting Project setting failed")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &projectSetting, http.StatusFound
}

func (store *MemSQL) IsClearbitIntegratedByProjectID(projectID int64) (bool, int) {
	db := C.GetServices().Db
	logFields := log.Fields{"project_id": projectID}
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectID); !valid {
		return false, http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectID).Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting project_setting failed - clearbit integration check")

		if gorm.IsRecordNotFoundError(err) {
			return false, http.StatusNotFound
		}
		return false, http.StatusInternalServerError
	}
	return *projectSetting.IntClearBit, http.StatusFound
}

func (store *MemSQL) GetClearbitKeyFromProjectSetting(projectId int64) (string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return "", http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectId).Select("clearbit_key").Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting clear_bit key from project_setting failed")

		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError
	}

	return projectSetting.ClearbitKey, http.StatusFound
}

// The integration to 6signal is made by either of 2 ways - 6 signal by 6signal or Factors Deanonimisation
// on integration page of factors.
func (store *MemSQL) IsSixSignalIntegratedByEitherWay(projectID int64) (bool, int) {
	db := C.GetServices().Db
	logFields := log.Fields{"project_id": projectID}
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectID); !valid {
		return false, http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectID).Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting project_setting failed - 6 signal integration check")

		if gorm.IsRecordNotFoundError(err) {
			return false, http.StatusNotFound
		}
		return false, http.StatusInternalServerError
	}
	return *projectSetting.IntClientSixSignalKey || *projectSetting.IntFactorsSixSignalKey, http.StatusFound
}

func (store *MemSQL) GetClient6SignalKeyFromProjectSetting(projectId int64) (string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return "", http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id= ?", projectId).Select("client6_signal_key").Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting client6_signal_key from project_setting failed")

		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError

	}
	return projectSetting.Client6SignalKey, http.StatusFound

}

func (store *MemSQL) GetFactors6SignalKeyFromProjectSetting(projectId int64) (string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return "", http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id= ?", projectId).Select("factors6_signal_key").Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting factors6_signal_key from project_setting failed")

		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError

	}
	return projectSetting.Factors6SignalKey, http.StatusFound

}

func (store *MemSQL) GetSixsignalEmailListFromProjectSetting(projectId int64) (string, int) {
	db := C.GetServices().Db
	logFields := log.Fields{"project_id": projectId}
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return "", http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectId).Select("sixsignal_email_list").Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting project_setting failed in GetSixsignalEmailListFromProjectSetting.")

		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError
	}
	return projectSetting.SixSignalEmailList, http.StatusFound
}

func (store *MemSQL) AddSixsignalEmailList(projectId int64, emailIds string) int {
	logFields := log.Fields{
		"project_id": projectId,
		"emailIds":   emailIds,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if err := db.Table("project_settings").Where("project_id = ?", projectId).Update("sixsignal_email_list", emailIds).Error; err != nil {
		logCtx.WithError(err).Error("Setting sixsignal_email_list from project_setting failed")
		return http.StatusInternalServerError
	}
	return http.StatusCreated
}

func (store *MemSQL) GetIntegrationBitsFromProjectSetting(projectId int64) (string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectId); !valid {
		return "", http.StatusBadRequest
	}

	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectId).Select("integration_bits").Find(&projectSetting).Error; err != nil {
		logCtx.WithError(err).Error("Getting clear_bit key from project_setting failed")

		if gorm.IsRecordNotFoundError(err) {
			return "", http.StatusNotFound
		}
		return "", http.StatusInternalServerError
	}

	return projectSetting.IntegrationBits, http.StatusFound
}

func (store *MemSQL) SetIntegrationBits(projectID int64, integrationBits string) int {
	logFields := log.Fields{
		"project_id": projectID,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if err := db.Table("project_settings").Where("project_id = ?", projectID).Update("integration_bits", integrationBits).Error; err != nil {
		logCtx.WithError(err).Error("Setting integration_bits from project_setting failed")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

type ProjectSettingChannelResponse struct {
	Setting   *model.ProjectSetting
	ErrorCode int
}

// GetProjectSettingByKeyWithTimeout - Get project_settings from db based on key,
// gets timedout and returns StatusInternalServerError, if the query takes more than
// the given duration. Returns default project_settings immediately, if the
// config/flag use_default_project_setting_for_sdk is set to true.
func (store *MemSQL) GetProjectSettingByKeyWithTimeout(key, value string, timeout time.Duration) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"key":     key,
		"value":   value,
		"timeout": timeout,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if C.GetConfig().UseDefaultProjectSettingForSDK {
		// Returning not_modified to avoid caching default.
		return getProjectSettingDefault(), http.StatusNotModified
	}

	// TODO(Dinesh): Use gorm db.WithContext and context.WithTimeout
	// once gorm v2 is production ready and upgraded.
	// Ref: https://gorm.io/docs/context.html
	responseChannel := make(chan ProjectSettingChannelResponse, 1)
	go func() {
		settings, errCode := getProjectSettingByKey(key, value)
		responseChannel <- ProjectSettingChannelResponse{
			Setting:   settings,
			ErrorCode: errCode,
		}
	}()

	select {
	case response := <-responseChannel:
		return response.Setting, response.ErrorCode
	case <-time.After(timeout):
		// Tracking the info log on a chart.
		log.WithField("tag", "get_settings_timeout").
			WithField("project_key", key).
			WithField("value", value).
			Info("Get project_settings has timedout.")
		metrics.Increment(metrics.IncrSDKGetSettingsTimeout)
		return nil, http.StatusInternalServerError
	}
}

// EnableBigqueryArchivalForProject To enable archival and bigquery in project_settings.
func (store *MemSQL) EnableBigqueryArchivalForProject(projectID int64) int {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if valid := isValidProjectScope(projectID); !valid {
		return http.StatusBadRequest
	}

	if err := db.Model(model.ProjectSetting{}).Where("project_id = ?", projectID).
		Updates(map[string]interface{}{"archive_enabled": true, "bigquery_enabled": true}).Error; err != nil {
		logCtx.WithError(err).Error("Failed to update project_settings for bigquery archival")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) UpdateSegmentMarkerLastRun(projectID int64, lastRunTime time.Time) int {
	logFields := log.Fields{
		"project_id":    projectID,
		"last_run_time": lastRunTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if err := db.Model(model.ProjectSetting{}).Where("project_id = ?", projectID).
		Update("segment_marker_last_run", lastRunTime).Error; err != nil {
		logCtx.WithError(err).Error("Failed to update project_settings for segment_marker_last_run")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
func (store *MemSQL) UpdateSegmentMarkerLastRunForAllAccounts(projectID int64, lastRunTime time.Time) int {
	logFields := log.Fields{
		"project_id":    projectID,
		"last_run_time": lastRunTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	if err := db.Model(model.ProjectSetting{}).Where("project_id = ?", projectID).
		Update("marker_last_run_all_accounts", lastRunTime).Error; err != nil {
		logCtx.WithError(err).Error("Failed to update project_settings for marker_last_run_all_accounts")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// getProjectSettingByKey - Get project settings by a column on projects.
func getProjectSettingByKey(key, value string) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"key":   key,
		"value": value,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if key == "" || value == "" {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithFields(logFields)

	var setting model.ProjectSetting
	db := C.GetServices().Db
	whereKey := fmt.Sprintf("%s = ?", key)
	err := db.Table("projects").Select("project_settings.*").Limit(1).Where(whereKey, value).Joins(
		"LEFT JOIN project_settings ON projects.id=project_settings.project_id").Find(&setting).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get project settings by token.")
		return nil, http.StatusInternalServerError
	}

	return &setting, http.StatusFound
}

func getProjectSettingCacheKey(tokenKey, tokenValue string) (*cacheRedis.Key, error) {
	logFields := log.Fields{
		"token_key":   tokenKey,
		"token_value": tokenValue,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// table_name:column_name
	prefix := fmt.Sprintf("%s:%s", "project_settings", tokenKey)
	return cacheRedis.NewKeyWithProjectUID(tokenValue, prefix, "")
}

func getCacheProjectSetting(tokenKey, tokenValue string) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"token_key":   tokenKey,
		"token_value": tokenValue,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if tokenValue == "" {
		return nil, http.StatusBadRequest
	}

	key, err := getProjectSettingCacheKey(tokenKey, tokenValue)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on getCacheProjectSetting")
		return nil, http.StatusInternalServerError
	}

	var settingsJson string
	if C.UseQueueRedis() {
		settingsJson, err = cacheRedis.GetQueueRedis(key)
	} else {
		settingsJson, err = cacheRedis.Get(key)
	}

	if err != nil {
		if err == redis.ErrNil {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed to get key from cache on getCacheProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	var settings model.ProjectSetting
	err = json.Unmarshal([]byte(settingsJson), &settings)
	if err != nil {
		log.WithError(err).Error(
			"Failed to unmarshal cached project settings on getCacheProjectSetting")
		return nil, http.StatusInternalServerError
	}

	defaultIntegrationValue := false
	if settings.IntRudderstack == nil {
		settings.IntRudderstack = &defaultIntegrationValue
	}

	return &settings, http.StatusFound
}

func setCacheProjectSetting(tokenKey, tokenValue string, settings *model.ProjectSetting) int {
	logFields := log.Fields{
		"token_key":   tokenKey,
		"token_value": tokenValue,
		"settings":    settings,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if tokenValue == "" || settings == nil {
		return http.StatusBadRequest
	}

	settingsJson, err := json.Marshal(settings)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to marshal project settings on setCacheProjectSetting.")
		return http.StatusInternalServerError
	}

	key, err := getProjectSettingCacheKey(tokenKey, tokenValue)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on setCacheProjectSetting")
		return http.StatusInternalServerError
	}

	var expiryInSecs float64 = 60 * 60
	if C.UseQueueRedis() {
		err = cacheRedis.SetQueueRedis(key, string(settingsJson), expiryInSecs)
	} else {
		err = cacheRedis.Set(key, string(settingsJson), expiryInSecs)
	}
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache on setCacheProjectSetting")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func delCacheProjectSetting(tokenKey, tokenValue string) int {
	logFields := log.Fields{
		"token_key":   tokenKey,
		"token_value": tokenValue,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	if tokenValue == "" {
		return http.StatusBadRequest
	}

	key, err := getProjectSettingCacheKey(tokenKey, tokenValue)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on delCacheProjectSetting")
		return http.StatusInternalServerError
	}

	if C.UseQueueRedis() {
		err = cacheRedis.DelQueueRedis(key)
	} else {
		err = cacheRedis.Del(key)
	}

	if err != nil && err != redis.ErrNil {
		logCtx.WithError(err).Error("Failed to del cache on delCacheProjectSetting")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func getProjectSettingDefault() *model.ProjectSetting {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	enabled := true
	disabled := false
	return &model.ProjectSetting{
		AutoTrack:              &enabled,
		AutoFormCapture:        &enabled,
		AutoCaptureFormFills:   &disabled,
		AutoTrackSPAPageView:   &disabled,
		ExcludeBot:             &enabled,
		IntSegment:             &enabled,
		IntRudderstack:         &disabled,
		IntDrift:               &disabled,
		IntClearBit:            &disabled,
		IntFactorsSixSignalKey: &enabled,
		AutoClickCapture:       &model.AutoClickCaptureDefault,
	}
}

// getProjectSettingByKeyWithDefault - Get from cache or db, if not use default.
func (store *MemSQL) getProjectSettingByKeyWithDefault(tokenKey, tokenValue string) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"token_key":   tokenKey,
		"token_value": tokenValue,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	settings, errCode := getCacheProjectSetting(tokenKey, tokenValue)
	if errCode == http.StatusFound {
		return settings, http.StatusFound
	}

	settings, errCode = store.GetProjectSettingByKeyWithTimeout(tokenKey, tokenValue, time.Millisecond*200)
	if errCode != http.StatusFound {
		// Use default settings, if db failure.
		// Do not cache default.
		return getProjectSettingDefault(), http.StatusFound
	}

	// add to cache.
	setCacheProjectSetting(tokenKey, tokenValue, settings)

	return settings, http.StatusFound
}

func (store *MemSQL) GetProjectSettingByTokenWithCacheAndDefault(token string) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"token": token,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.getProjectSettingByKeyWithDefault(model.ProjectSettingKeyToken, token)
}

func (store *MemSQL) GetProjectSettingByPrivateTokenWithCacheAndDefault(
	privateToken string) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"private_token": privateToken,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	return store.getProjectSettingByKeyWithDefault(
		model.ProjectSettingKeyPrivateToken, privateToken)
}

func (store *MemSQL) createProjectSetting(ps *model.ProjectSetting) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"ps": ps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if valid := isValidProjectScope(ps.ProjectId); !valid {
		return nil, http.StatusBadRequest
	}

	if errCode := store.satisfiesProjectSettingForeignConstraints(*ps); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	if err := db.Create(ps).Error; err != nil {
		log.WithFields(log.Fields{"model.ProjectSetting": ps}).WithError(
			err).Error("Failed creating model.ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return ps, http.StatusCreated
}

func (store *MemSQL) delAllProjectSettingsCacheForProject(projectId int64) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	project, errCode := store.GetProject(projectId)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get project on delAllProjectSettingsCacheKeys.")
	}

	// delete all project setting cache keys by respective
	// token key and value.
	delCacheProjectSetting(model.ProjectSettingKeyToken, project.Token)
	delCacheProjectSetting(model.ProjectSettingKeyPrivateToken, project.PrivateToken)
}

func (store *MemSQL) UpdateProjectSettings(projectId int64, settings *model.ProjectSetting) (*model.ProjectSetting, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"settings":   settings,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if projectId == 0 || settings == nil {
		return nil, http.StatusBadRequest
	}
	settings.ProjectId = projectId

	if settings.IntAdwordsCustomerAccountId != nil {
		var cleanAdwordsAccountIds []string
		adwordsAccoundIds := strings.Split(*settings.IntAdwordsCustomerAccountId, ",")
		for _, accountId := range adwordsAccoundIds {
			adwordsCustomerAccountId := strings.Replace(
				accountId, "-", "", -1)
			adwordsCustomerAccountId = strings.TrimSpace(adwordsCustomerAccountId)
			cleanAdwordsAccountIds = append(cleanAdwordsAccountIds, adwordsCustomerAccountId)
		}
		*settings.IntAdwordsCustomerAccountId = strings.Join(cleanAdwordsAccountIds, ",")
	}

	if settings.IntGoogleOrganicURLPrefixes != nil {
		*settings.IntGoogleOrganicURLPrefixes = strings.ReplaceAll(*settings.IntGoogleOrganicURLPrefixes, " ", "")
	}

	if errCode := store.satisfiesProjectSettingForeignConstraints(*settings); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}
	if settings.IntHubspotApiKey != "" || settings.IntHubspotRefreshToken != "" {
		existingSettings, status := store.GetProjectSetting(projectId)
		if status != http.StatusFound {
			return nil, http.StatusInternalServerError
		}

		if (settings.IntHubspotApiKey != "" && existingSettings.IntHubspotApiKey != settings.IntHubspotApiKey) ||
			(settings.IntHubspotRefreshToken != "" && existingSettings.IntHubspotRefreshToken != settings.IntHubspotRefreshToken) {
			hubspotIntegrationAccount, err := model.GetHubspotIntegrationAccount(projectId, settings.IntHubspotApiKey,
				settings.IntHubspotRefreshToken, C.GetHubspotAppID(), C.GetHubspotAppSecret())
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectId}).WithError(
					err).Error("Failed to fetch hubspot account details on integration.") // log error but still allow integration
			}

			if existingSettings.IntHubspotPortalID != nil {
				if existingSettings.IntHubspotApiKey != "" &&
					hubspotIntegrationAccount.PortalID != *existingSettings.IntHubspotPortalID {
					log.WithFields(log.Fields{"project_id": projectId, "previous_portal_id": *existingSettings.IntHubspotPortalID,
						"new_portal_id": hubspotIntegrationAccount.PortalID}).Error("Portal id mismatch on hubspot re integration.")
					settings.IntHubspotPortalID = &hubspotIntegrationAccount.PortalID
				}
			} else {
				settings.IntHubspotPortalID = &hubspotIntegrationAccount.PortalID
			}

		}
	}

	// validate ip
	if settings.FilterIps != nil {
		var filterIps model.FilterIps
		err := U.DecodePostgresJsonbToStructType(settings.FilterIps, &filterIps)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "setting": settings}).WithError(
				err).Error("Failed decoding json. Aborting.")
			return nil, http.StatusInternalServerError
		}
		if !IsValidIP(filterIps) {
			log.WithFields(log.Fields{"project_id": projectId, "filter_ips": filterIps}).
				Error("Invalid IP Address entered. Aborting")
			return nil, http.StatusBadRequest
		}
	}

	var updatedProjectSetting model.ProjectSetting
	if err := db.Model(&updatedProjectSetting).Where("project_id = ?",
		projectId).Updates(settings).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"model.ProjectSetting": settings}).WithError(
			err).Error("Failed updating ProjectSettings.")
		return nil, http.StatusInternalServerError
	}
	store.delAllProjectSettingsCacheForProject(projectId)

	return &updatedProjectSetting, http.StatusAccepted
}

func IsValidIP(filterIps model.FilterIps) bool {
	for _, ip := range filterIps.BlockIps {
		ip = strings.TrimSpace(ip)
		if net.ParseIP(ip) == nil {
			return false
		}
	}
	return true
}

func IsBlockedIP(checkString0 string, checkString1 string, filterIps model.FilterIps) bool {
	checkString0 = strings.TrimSpace(checkString0)
	checkString1 = strings.TrimSpace(checkString1)
	for _, blockedIP := range filterIps.BlockIps {
		ip := strings.TrimSpace(blockedIP)
		if checkString0 == ip || checkString1 == ip {
			return true
		}
	}
	return false
}

func (store *MemSQL) IsPSettingsIntShopifyEnabled(projectId int64) bool {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return true
}

func (store *MemSQL) GetIntAdwordsRefreshTokenForProject(projectId int64) (string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	settings, errCode := store.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return "", errCode
	}

	if settings.IntAdwordsEnabledAgentUUID == nil || *settings.IntAdwordsEnabledAgentUUID == "" {
		return "", http.StatusNotFound
	}

	logCtx := log.WithFields(logFields)

	agent, errCode := store.GetAgentByUUID(*settings.IntAdwordsEnabledAgentUUID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Adwords enabled agent not found on agents table.")
		return "", errCode
	}

	refreshToken := agent.IntAdwordsRefreshToken
	if refreshToken == "" {
		logCtx.Error("Adwords enabled agent refresh token is empty.")
		return "", http.StatusInternalServerError
	}

	return refreshToken, http.StatusFound
}

func (store *MemSQL) GetIntGoogleOrganicRefreshTokenForProject(projectId int64) (string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	settings, errCode := store.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return "", errCode
	}

	if settings.IntGoogleOrganicEnabledAgentUUID == nil || *settings.IntGoogleOrganicEnabledAgentUUID == "" {
		return "", http.StatusNotFound
	}

	logCtx := log.WithFields(logFields)

	agent, errCode := store.GetAgentByUUID(*settings.IntGoogleOrganicEnabledAgentUUID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("GoogleOrganic enabled agent not found on agents table.")
		return "", errCode
	}

	refreshToken := agent.IntGoogleOrganicRefreshToken
	if refreshToken == "" {
		logCtx.Error("GoogleOrganic enabled agent refresh token is empty.")
		return "", http.StatusInternalServerError
	}

	return refreshToken, http.StatusFound
}

func (store *MemSQL) GetIntAdwordsProjectSettingsForProjectID(projectID int64) ([]model.AdwordsProjectSettings, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	queryStr := "SELECT project_settings.project_id, project_settings.int_adwords_customer_account_id as customer_account_id," +
		" " + "project_settings.int_google_ingestion_timezone as int_google_ingestion_timezone," +
		" " + "project_settings.int_adwords_client_manager_map as int_adwords_client_manager_map," +
		" " + "agents.int_adwords_refresh_token as refresh_token, project_settings.int_adwords_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_adwords_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.project_id = ?" +
		" " + "AND project_settings.int_adwords_customer_account_id IS NOT NULL" +
		" " + "AND project_settings.int_adwords_enabled_agent_uuid IS NOT NULL "
	params := []interface{}{projectID}

	return store.getIntAdwordsProjectSettings(queryStr, params)
}

func (store *MemSQL) GetAllIntAdwordsProjectSettings() ([]model.AdwordsProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	queryStr := "SELECT project_settings.project_id, project_settings.int_adwords_customer_account_id as customer_account_id," +
		" " + "project_settings.int_google_ingestion_timezone as int_google_ingestion_timezone," +
		" " + "project_settings.int_adwords_client_manager_map as int_adwords_client_manager_map," +
		" " + "agents.int_adwords_refresh_token as refresh_token, project_settings.int_adwords_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_adwords_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.int_adwords_customer_account_id IS NOT NULL" +
		" " + "AND project_settings.int_adwords_enabled_agent_uuid IS NOT NULL"
	params := make([]interface{}, 0, 0)

	return store.getIntAdwordsProjectSettings(queryStr, params)
}

func (store *MemSQL) getIntAdwordsProjectSettings(query string, params []interface{}) ([]model.AdwordsProjectSettings, int) {
	logFields := log.Fields{
		"query":  query,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	adwordsProjectSettings := make([]model.AdwordsProjectSettings, 0, 0)
	rows, err := db.Raw(query, params).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all adwords project settings.")
		return adwordsProjectSettings, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var adwordsSettings model.AdwordsProjectSettings
		if err := db.ScanRows(rows, &adwordsSettings); err != nil {
			log.WithError(err).Error("Failed to scan get all adwords project settings.")
			return adwordsProjectSettings, http.StatusInternalServerError
		}

		adwordsProjectSettings = append(adwordsProjectSettings, adwordsSettings)
	}

	return adwordsProjectSettings, http.StatusOK
}

func (store *MemSQL) GetIntGoogleOrganicProjectSettingsForProjectID(projectID int64) ([]model.GoogleOrganicProjectSettings, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	queryStr := "SELECT project_settings.project_id, project_settings.int_google_organic_url_prefixes as url_prefix," +
		" " + "agents.int_google_organic_refresh_token as refresh_token, project_settings.int_google_organic_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_google_organic_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.project_id = ?" +
		" " + "AND project_settings.int_google_organic_url_prefixes IS NOT NULL AND project_settings.int_google_organic_url_prefixes != ''" +
		" " + "AND project_settings.int_google_organic_enabled_agent_uuid IS NOT NULL "
	params := []interface{}{projectID}

	return store.getIntGoogleOrganicProjectSettings(queryStr, params)
}

func (store *MemSQL) GetAllIntGoogleOrganicProjectSettings() ([]model.GoogleOrganicProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	queryStr := "SELECT project_settings.project_id, project_settings.int_google_organic_url_prefixes as url_prefix," +
		" " + "agents.int_google_organic_refresh_token as refresh_token, project_settings.int_google_organic_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_google_organic_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.int_google_organic_url_prefixes IS NOT NULL AND project_settings.int_google_organic_url_prefixes != ''" +
		" " + "AND project_settings.int_google_organic_enabled_agent_uuid IS NOT NULL "
	params := make([]interface{}, 0, 0)

	return store.getIntGoogleOrganicProjectSettings(queryStr, params)
}
func (store *MemSQL) getIntGoogleOrganicProjectSettings(query string, params []interface{}) ([]model.GoogleOrganicProjectSettings, int) {
	logFields := log.Fields{
		"query":  query,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	gscProjectSettings := make([]model.GoogleOrganicProjectSettings, 0, 0)
	rows, err := db.Raw(query, params).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all gsc project settings.")
		return gscProjectSettings, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var gscSettings model.GoogleOrganicProjectSettings
		if err := db.ScanRows(rows, &gscSettings); err != nil {
			log.WithError(err).Error("Failed to scan get all gsc project settings.")
			return gscProjectSettings, http.StatusInternalServerError
		}

		gscProjectSettings = append(gscProjectSettings, gscSettings)
	}

	return gscProjectSettings, http.StatusOK
}

func (store *MemSQL) GetAllHubspotProjectSettings() ([]model.HubspotProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	var hubspotProjectSettings []model.HubspotProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_hubspot=true AND ( int_hubspot_api_key IS NOT NULL OR int_hubspot_refresh_token is not NULL )").Select(
		"project_id, int_hubspot_api_key as api_key, int_hubspot_first_time_synced as is_first_time_synced," +
			"int_hubspot_sync_info as sync_info, int_hubspot_refresh_token as refresh_token").
		Find(&hubspotProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot project_settings.")
		return hubspotProjectSettings, http.StatusInternalServerError
	}

	return hubspotProjectSettings, http.StatusFound
}

func (store *MemSQL) IsHubspotIntegrationAvailable(projectID int64) bool {
	hubspotProjectSettings, errCode := store.GetAllHubspotProjectSettingsForProjectID(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).Warn(" Failed in getting hubspot project settings.")
		return false
	}
	if len(hubspotProjectSettings) == 0 {
		log.WithField("projectId", projectID).Warn("Hubspot integration is not available.")
		return false
	}
	return true
}

func (store *MemSQL) GetAllHubspotProjectSettingsForProjectID(projectID int64) ([]model.HubspotProjectSettings, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var hubspotProjectSettings []model.HubspotProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_hubspot=true AND ( int_hubspot_api_key IS NOT NULL OR int_hubspot_refresh_token is not NULL ) AND project_id IN (?)", projectID).Select(
		"project_id, int_hubspot_api_key as api_key, int_hubspot_first_time_synced as is_first_time_synced," +
			"int_hubspot_sync_info as sync_info, int_hubspot_refresh_token as refresh_token").
		Find(&hubspotProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot project_settings.")
		return hubspotProjectSettings, http.StatusInternalServerError
	}

	return hubspotProjectSettings, http.StatusFound
}

type shopifyInfoStruct struct {
	apiKey    string
	projectId int64
	hashEmail bool
}

var developmentShopifyInfo = map[string]shopifyInfoStruct{
	"aravind-test123.myshopify.com": shopifyInfoStruct{
		projectId: 2,
		apiKey:    "93f0ecd1ff038bb0de72ec1f4dcf34b3aecf2a2f15f1f531dbd89bfecb546b1e",
		hashEmail: false,
	},
}
var stagingShopifyInfo = map[string]shopifyInfoStruct{
	"aravind-test123.myshopify.com": shopifyInfoStruct{
		projectId: 21,
		apiKey:    "93f0ecd1ff038bb0de72ec1f4dcf34b3aecf2a2f15f1f531dbd89bfecb546b1e",
		hashEmail: true,
	},
}
var productionShopifyInfo = map[string]shopifyInfoStruct{
	"aravind-test123.myshopify.com": shopifyInfoStruct{
		projectId: 395,
		apiKey:    "93f0ecd1ff038bb0de72ec1f4dcf34b3aecf2a2f15f1f531dbd89bfecb546b1e",
		hashEmail: false,
	},
	"quirksmith.myshopify.com": shopifyInfoStruct{
		projectId: 401,
		apiKey:    "8dd75ec8aded049912dffe8ecab9591606ac3b0ee389cf2a76b26be88854fff4",
		hashEmail: true,
	},
	"flatheads.myshopify.com": shopifyInfoStruct{
		projectId: 380,
		apiKey:    "844e9cc7c7e673a9513827a8a89613970be27ef8ec67ab4da2e3c9202f1ec7d8",
		hashEmail: true,
	},
	"azani.myshopify.com": shopifyInfoStruct{
		projectId: 410,
		apiKey:    "7fb11f1eacd53c1e8d0d254c5d2b23d2",
		hashEmail: true,
	},
}

func (store *MemSQL) GetProjectDetailsByShopifyDomain(
	shopifyDomain string) (int64, string, bool, int) {
	logFields := log.Fields{
		"shopify_domain": shopifyDomain,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var shopifyInfo map[string]shopifyInfoStruct
	if C.IsDevelopment() {
		shopifyInfo = developmentShopifyInfo
	} else if C.IsStaging() {
		shopifyInfo = stagingShopifyInfo
	} else if C.IsProduction() {
		shopifyInfo = productionShopifyInfo
	}
	if info, found := shopifyInfo[shopifyDomain]; found {
		return info.projectId, info.apiKey, info.hashEmail, http.StatusFound
	} else {
		log.Error(fmt.Sprintf("Unknown shopify domain - %s", shopifyDomain))
	}
	return 0, "", false, http.StatusInternalServerError
}

func (store *MemSQL) GetFacebookEnabledIDsAndProjectSettings() ([]int64, []model.FacebookProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	facebookProjectSettings := make([]model.FacebookProjectSettings, 0, 0)
	facebookIDs := make([]int64, 0, 0)

	err := db.Table("project_settings").Where("int_facebook_access_token IS NOT NULL AND int_facebook_access_token != ''").Find(&facebookProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get facebook enabled project settings for sync info.")
		return facebookIDs, facebookProjectSettings, http.StatusInternalServerError
	}
	for _, facebookProjectSetting := range facebookProjectSettings {
		facebookIDs = append(facebookIDs, facebookProjectSetting.ProjectId)
	}
	return facebookIDs, facebookProjectSettings, http.StatusOK
}

func (store *MemSQL) GetFacebookEnabledIDsAndProjectSettingsForProject(projectIDs []int64) ([]int64, []model.FacebookProjectSettings, int) {
	logFields := log.Fields{
		"project_ids": projectIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	facebookProjectSettings := make([]model.FacebookProjectSettings, 0, 0)
	facebookIDs := make([]int64, 0, 0)

	err := db.Table("project_settings").Where("int_facebook_access_token IS NOT NULL AND int_facebook_access_token != '' AND project_id IN (?)", projectIDs).Find(&facebookProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get facebook enabled project settings for sync info.")
		return facebookIDs, facebookProjectSettings, http.StatusInternalServerError
	}
	for _, facebookProjectSetting := range facebookProjectSettings {
		facebookIDs = append(facebookIDs, facebookProjectSetting.ProjectId)
	}
	return facebookIDs, facebookProjectSettings, http.StatusOK
}

func (store *MemSQL) GetLinkedinEnabledProjectSettings() ([]model.LinkedinProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	linkedinProjectSettings := make([]model.LinkedinProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_linkedin_refresh_token IS NOT NULL AND int_linkedin_refresh_token != ''").Find(&linkedinProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get linkedin enabled project settings for sync info.")
		return linkedinProjectSettings, http.StatusInternalServerError
	}
	return linkedinProjectSettings, http.StatusOK
}
func (store *MemSQL) GetLinkedinEnabledProjectSettingsForProjects(projectIDs []string) ([]model.LinkedinProjectSettings, int) {
	logFields := log.Fields{
		"project_ids": projectIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	linkedinProjectSettings := make([]model.LinkedinProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_linkedin_refresh_token IS NOT NULL AND int_linkedin_refresh_token != '' AND project_id IN (?)", projectIDs).Find(&linkedinProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get linkedin enabled project settings for sync info.")
		return linkedinProjectSettings, http.StatusInternalServerError
	}
	return linkedinProjectSettings, http.StatusOK
}

func (store *MemSQL) GetG2EnabledProjectSettings() ([]model.G2ProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	g2ProjectSettings := make([]model.G2ProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_g2_api_key IS NOT NULL AND int_g2_api_key != '' AND int_g2 = true").Find(&g2ProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get g2 enabled project settings for sync info.")
		return g2ProjectSettings, http.StatusInternalServerError
	}
	return g2ProjectSettings, http.StatusOK
}

func (store *MemSQL) GetG2EnabledProjectSettingsForProjects(projectIDs []int64) ([]model.G2ProjectSettings, int) {
	logFields := log.Fields{
		"project_ids": projectIDs,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	g2ProjectSettings := make([]model.G2ProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_g2_api_key IS NOT NULL AND int_g2_api_key != '' AND int_g2 = true AND project_id IN (?)", projectIDs).Find(&g2ProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get g2 enabled project settings for sync info.")
		return g2ProjectSettings, http.StatusInternalServerError
	}
	return g2ProjectSettings, http.StatusOK
}

// GetArchiveEnabledProjectIDs Returns list of project ids which have archive enabled.
func (store *MemSQL) GetArchiveEnabledProjectIDs() ([]int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	var projectIDs []int64
	db := C.GetServices().Db

	rows, err := db.Model(&model.ProjectSetting{}).Where("archive_enabled = true").Select("project_id").Rows()
	if err != nil {
		log.WithError(err).Error("Query failed for GetArchiveEnabledProjectIDs")
		return projectIDs, http.StatusInternalServerError
	}

	for rows.Next() {
		var projectID int64
		err = rows.Scan(&projectID)
		if err != nil {
			log.WithError(err).Error("Error while scanning")
			continue
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, http.StatusFound
}

// GetBigqueryEnabledProjectIDs Returns list of project ids which have bigquery enabled.
func (store *MemSQL) GetBigqueryEnabledProjectIDs() ([]int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	var projectIDs []int64
	db := C.GetServices().Db

	rows, err := db.Model(&model.ProjectSetting{}).Where("bigquery_enabled = true").Select("project_id").Rows()
	if err != nil {
		log.WithError(err).Error("Query failed for GetBigqueryEnabledProjectIDs")
		return projectIDs, http.StatusInternalServerError
	}

	for rows.Next() {
		var projectID int64
		err = rows.Scan(&projectID)
		if err != nil {
			log.WithError(err).Error("Error while scanning")
			continue
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, http.StatusFound
}

func (store *MemSQL) IsSalesforceIntegrationAvailable(projectID int64) bool {
	salesforceProjectSettings, errCode := store.GetAllSalesforceProjectSettingsForProject(projectID)
	if errCode != http.StatusFound && errCode != http.StatusOK {
		log.WithField("projectId", projectID).Warn(" Failed in getting salesforce project settings.")
		return false
	}
	if len(salesforceProjectSettings) == 0 {
		log.WithField("projectId", projectID).Warn("Salesforce integration is not available.")
		return false
	}
	return true
}

// GetAllSalesforceProjectSettings return list of all enabled salesforce projects and their meta data
func (store *MemSQL) GetAllSalesforceProjectSettings() ([]model.SalesforceProjectSettings, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	var salesforceProjectSettings []model.SalesforceProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_salesforce_enabled_agent_uuid != '' AND int_salesforce_enabled_agent_uuid IS NOT NULL ").Joins(" left join agents on project_settings.int_salesforce_enabled_agent_uuid = agents.uuid").Select(
		"project_id, int_salesforce_refresh_token as refresh_token, int_salesforce_instance_url as instance_url").Find(
		&salesforceProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get salesforce project_settings.")
		return salesforceProjectSettings, http.StatusInternalServerError
	}

	return salesforceProjectSettings, http.StatusFound
}

func (store *MemSQL) GetAllSalesforceProjectSettingsForProject(projectID int64) ([]model.SalesforceProjectSettings, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var salesforceProjectSettings []model.SalesforceProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_salesforce_enabled_agent_uuid != '' AND int_salesforce_enabled_agent_uuid IS NOT NULL AND project_id IN (?)", projectID).Joins(" left join agents on project_settings.int_salesforce_enabled_agent_uuid = agents.uuid").Select(
		"project_id, int_salesforce_refresh_token as refresh_token, int_salesforce_instance_url as instance_url").Find(
		&salesforceProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get salesforce project_settings.")
		return salesforceProjectSettings, http.StatusInternalServerError
	}

	return salesforceProjectSettings, http.StatusFound
}

func (store *MemSQL) GetAdwordsEnabledProjectIDAndCustomerIDsFromProjectSettings() (map[int64][]string, error) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db

	projectSettings := make([]model.ProjectSetting, 0, 0)
	mapOfProjectToCustomerIds := make(map[int64][]string)

	err := db.Table("project_settings").Where("int_adwords_enabled_agent_uuid IS NOT NULL AND int_adwords_enabled_agent_uuid != ''").Find(&projectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get facebook enabled project settings for sync info.")
		return mapOfProjectToCustomerIds, err
	}
	for _, projectSetting := range projectSettings {
		projectID := projectSetting.ProjectId
		if projectSetting.IntAdwordsCustomerAccountId != nil {
			var cleanAdwordsAccountIds []string
			adwordsAccoundIDs := strings.Split(*projectSetting.IntAdwordsCustomerAccountId, ",")
			for _, accountID := range adwordsAccoundIDs {
				adwordsCustomerAccountID := strings.Replace(accountID, "-", "", -1)
				adwordsCustomerAccountID = strings.TrimSpace(adwordsCustomerAccountID)
				cleanAdwordsAccountIds = append(cleanAdwordsAccountIds, adwordsCustomerAccountID)
			}
			mapOfProjectToCustomerIds[projectID] = cleanAdwordsAccountIds
		}
	}
	return mapOfProjectToCustomerIds, nil
}
func (store *MemSQL) IsBingIntegrationAvailable(projectID int64) bool {
	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		return false
	}
	if ftMapping.ConnectorID == "" {
		return false
	}
	return true
}
func (store *MemSQL) DeleteChannelIntegration(projectID int64, channelName string) (int, error) {
	logFields := log.Fields{
		"project_id":   projectID,
		"channel_name": channelName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 {
		return http.StatusBadRequest, errors.New("invalid projectID")
	}
	switch channelName {
	case "facebook":
		return store.DeleteFacebookIntegration(projectID)
	case "linkedin":
		return store.DeleteLinkedinIntegration(projectID)
	case "adwords":
		return store.DeleteAdwordsIntegration(projectID)
	case "google_organic":
		return store.DeleteGoogleOrganicIntegration(projectID)
	default:
		return http.StatusBadRequest, errors.New("invalid channel name")
	}
}

func (store *MemSQL) IsMarketoIntegrationAvailable(projectID int64) bool {
	connectorId, err := store.GetFiveTranMapping(projectID, model.MarketoIntegration)
	if err != nil {
		return false
	}
	if connectorId == "" {
		return false
	}
	return true
}

func (store *MemSQL) GetAllLeadSquaredEnabledProjects() (map[int64]model.LeadSquaredConfig, error) {

	db := C.GetServices().Db
	result := make(map[int64]model.LeadSquaredConfig)
	projectSettings := make([]model.ProjectSetting, 0, 0)
	err := db.Table("project_settings").Where("lead_squared_config IS NOT NULL").Find(&projectSettings).Error

	for _, setting := range projectSettings {
		var leadSquaredConfig model.LeadSquaredConfig
		err = U.DecodePostgresJsonbToStructType(setting.LeadSquaredConfig, &leadSquaredConfig)
		if err != nil {
			return nil, errors.New("failed to decode jsonb to lead squared setting")
		}
		result[setting.ProjectId] = leadSquaredConfig
	}
	return result, nil
}

func (store *MemSQL) GetAllAdsImportEnabledProjects() (map[int64]map[string]model.LastProcessedAdsImport, error) {
	db := C.GetServices().Db
	result := make(map[int64]map[string]model.LastProcessedAdsImport)
	adsImport := make([]model.AdsImport, 0, 0)
	err := db.Table("ads_import").Where("status = true").Find(&adsImport).Error
	if err != nil {
		log.WithError(err).Error("Failed to get ads_import data")
		return result, err
	}
	for _, data := range adsImport {
		var lastProcessed map[string]model.LastProcessedAdsImport
		err = U.DecodePostgresJsonbToStructType(data.LastProcessedIndex, &lastProcessed)
		if err != nil {
			log.WithError(err).Error("Decode failed")
			return nil, errors.New("failed to decode jsonb to last processed")
		}
		result[data.ProjectID] = lastProcessed
	}
	return result, nil
}

func (store *MemSQL) UpdateLastProcessedAdsData(updatedFields map[string]model.LastProcessedAdsImport, projectId int64) int {
	db := C.GetServices().Db
	updateFieldsJson, _ := U.EncodeStructTypeToPostgresJsonb(updatedFields)
	if err := db.Table("ads_import").
		Where("project_id = ?", projectId).
		Update("last_processed_index", updateFieldsJson).Error; err != nil {
		log.WithError(err).Error("Updating last processed index config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) GetCustomAdsSourcesByProject(projectID int64) ([]string, int) {
	db := C.GetServices().Db
	adsImport := make([]string, 0, 0)
	dbQuery := "SELECT DISTINCT(source) FROM integration_documents where project_id = ? and source != 'bingads' GROUP BY source order by min(created_at)"
	rows, err := db.Raw(dbQuery, projectID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all sources.")
		return adsImport, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		if err = rows.Scan(&source); err != nil {
			log.WithError(err).Error("Failed to get source. Scanning failed.")
			return adsImport, http.StatusInternalServerError
		}
		adsImport = append(adsImport, source)
	}
	return adsImport, http.StatusOK
}

func (store *MemSQL) GetCustomAdsAccountsByProject(projectID int64, source []string) ([]string, int) {
	db := C.GetServices().Db
	adsImport := make([]string, 0, 0)
	rows, err := db.Raw("SELECT DISTINCT customer_account_id FROM integration_documents where project_id = ? AND source IN (?)", projectID, source).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all sources.")
		return adsImport, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var account string
		if err = rows.Scan(&account); err != nil {
			log.WithError(err).Error("Failed to get source. Scanning failed.")
			return adsImport, http.StatusInternalServerError
		}
		adsImport = append(adsImport, account)
	}
	return adsImport, http.StatusOK
}

func (store *MemSQL) IsCustomAdsAvailable(projectID int64) bool {
	db := C.GetServices().Db
	adsImport := make([]model.AdsImport, 0, 0)
	err := db.Table("ads_import").Where("status = true AND project_id = ?", projectID).Find(&adsImport).Error
	if err != nil {
		return false
	}
	if len(adsImport) > 0 {
		return true
	}
	return false
}

func (store *MemSQL) UpdateLeadSquaredFirstTimeSyncStatus(projectId int64) int {
	db := C.GetServices().Db
	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectId).First(&projectSetting).Error; err != nil {
		log.WithError(err).Error("Getting Project setting failed")
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	}
	var leadSquaredConfig model.LeadSquaredConfig
	err := U.DecodePostgresJsonbToStructType(projectSetting.LeadSquaredConfig, &leadSquaredConfig)
	if err != nil {
		log.WithError(err).Error("decoding postgres json failed")
		return http.StatusInternalServerError
	}
	leadSquaredConfig.FirstTimeSync = true
	leadSquaredConfigJsonb, err := U.EncodeStructTypeToPostgresJsonb(leadSquaredConfig)
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("lead_squared_config", leadSquaredConfigJsonb).Error; err != nil {
		log.WithError(err).Error("Updating leadsquared config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) UpdateLeadSquaredConfig(projectId int64, accessKey string, host string, secretkey string) int {
	db := C.GetServices().Db
	var projectSetting model.ProjectSetting
	if err := db.Where("project_id = ?", projectId).First(&projectSetting).Error; err != nil {
		log.WithError(err).Error("Getting Project setting failed")
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	}
	if accessKey == "" || host == "" || secretkey == "" {
		if err := db.Model(&model.ProjectSetting{}).
			Where("project_id = ?", projectId).
			Update("lead_squared_config", nil).Error; err != nil {
			log.WithError(err).Error("Updating leadsquared config failed")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}
	var leadSquaredConfig model.LeadSquaredConfig
	if projectSetting.LeadSquaredConfig != nil {
		err := U.DecodePostgresJsonbToStructType(projectSetting.LeadSquaredConfig, &leadSquaredConfig)
		if err != nil {
			log.WithError(err).Error("decoding postgres json failed")
			return http.StatusInternalServerError
		}
	}
	leadSquaredConfig.FirstTimeSync = false
	leadSquaredConfig.AccessKey = accessKey
	leadSquaredConfig.SecretKey = secretkey
	leadSquaredConfig.Host = host
	leadSquaredConfig.BigqueryDataset = fmt.Sprintf("%v_%v_%v", "lead_squared", projectId, U.TimeNowUnix())
	leadSquaredConfigJsonb, _ := U.EncodeStructTypeToPostgresJsonb(leadSquaredConfig)
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("lead_squared_config", leadSquaredConfigJsonb).Error; err != nil {
		log.WithError(err).Error("Updating leadsquared config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) GetAllWeeklyInsightsEnabledProjects() ([]int64, error) {

	db := C.GetServices().Db
	result := make([]int64, 0)
	projectSettings := make([]model.ProjectSetting, 0, 0)
	_ = db.Table("project_settings").Where("is_weekly_insights_enabled = true").Find(&projectSettings).Error

	for _, setting := range projectSettings {
		result = append(result, setting.ProjectId)
	}
	return result, nil
}

func (store *MemSQL) GetAllPathAnalysisEnabledProjects() ([]int64, error) {

	db := C.GetServices().Db
	result := make([]int64, 0)
	projectSettings := make([]model.ProjectSetting, 0, 0)
	_ = db.Table("project_settings").Where("is_path_analysis_enabled = true").Find(&projectSettings).Error

	for _, setting := range projectSettings {
		result = append(result, setting.ProjectId)
	}
	return result, nil
}

func (store *MemSQL) GetAllExplainEnabledProjects() ([]int64, error) {

	db := C.GetServices().Db
	result := make([]int64, 0)
	projectSettings := make([]model.ProjectSetting, 0, 0)
	_ = db.Table("project_settings").Where("is_explain_enabled = true").Find(&projectSettings).Error

	for _, setting := range projectSettings {
		result = append(result, setting.ProjectId)
	}
	return result, nil
}

func (store *MemSQL) EnableWeeklyInsights(projectId int64) int {
	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("is_weekly_insights_enabled", true).Error; err != nil {
		log.WithError(err).Error("Updating is_weekly_insights_enabled config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) EnablePathAnalysis(projectId int64) int {
	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("is_path_analysis_enabled", true).Error; err != nil {
		log.WithError(err).Error("Updating is_path_analysis_enabled config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) EnableExplain(projectId int64) int {
	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("is_explain_enabled", true).Error; err != nil {
		log.WithError(err).Error("Updating is_explain_enabled config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) DisableWeeklyInsights(projectId int64) int {
	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("is_weekly_insights_enabled", false).Error; err != nil {
		log.WithError(err).Error("Updating is_weekly_insights_enabled config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) DisablePathAnalysis(projectId int64) int {
	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("is_path_analysis_enabled", false).Error; err != nil {
		log.WithError(err).Error("Updating is_path_analysis_enabled config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (store *MemSQL) DisableExplain(projectId int64) int {
	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("is_explain_enabled", false).Error; err != nil {
		log.WithError(err).Error("Updating is_explain_enabled config failed")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// fetch segment_marker_last_run for given project_id
func (store *MemSQL) GetSegmentMarkerLastRunTime(projectID int64) (time.Time, int) {
	db := C.GetServices().Db
	var projectSettings model.ProjectSetting
	err := db.Table("project_settings").Select("segment_marker_last_run").Where("project_id=?", projectID).Find(&projectSettings).Error
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Warn("Failed to fetch segment_marker_last_run from project_settings.")
		return time.Time{}, http.StatusInternalServerError
	}
	if projectSettings.SegmentMarkerLastRun.IsZero() {
		return time.Time{}, http.StatusNotFound
	}

	return projectSettings.SegmentMarkerLastRun, http.StatusFound
}

// fetch marker_last_run_all_accounts for given project_id
func (store *MemSQL) GetMarkerLastForAllAccounts(projectID int64) (time.Time, int) {
	db := C.GetServices().Db
	var projectSettings model.ProjectSetting
	err := db.Table("project_settings").Select("marker_last_run_all_accounts").Where("project_id=?", projectID).Find(&projectSettings).Error
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Warn("Failed to fetch marker_last_run_all_accounts from project_settings.")
		return time.Time{}, http.StatusInternalServerError
	}

	return projectSettings.MarkerLastRunAllAccounts, http.StatusFound
}

// fetch marker_last_run_all_accounts for given project_id
func (store *MemSQL) ProjectCountToRunAllMarkerFor() (int, int) {
	db := C.GetServices().Db
	var totalProjects int

	allDomainsRunHours := C.TimeRangeForAllDomains()
	defaultTime := U.DefaultTime()
	createTime := U.TimeNowZ().Add(time.Duration(-allDomainsRunHours) * time.Hour)

	// count of greater than default time to 24 hours
	err := db.Model(&model.ProjectSetting{}).Where("marker_last_run_all_accounts > ?", defaultTime).
		Or("created_at >= ?", createTime).Count(&totalProjects).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return totalProjects, http.StatusNotFound
		}
		log.WithError(err).Error("Error fetching total number of projects present")
		return totalProjects, http.StatusInternalServerError
	}

	return totalProjects, http.StatusFound
}

// fetch project_ids to run all marker for
func (store *MemSQL) GetProjectIDsListForMarker(limit int) ([]int64, int) {
	db := C.GetServices().Db

	// greater than default time to 24 hours
	query := fmt.Sprintf(`SELECT project_id FROM project_settings 
	WHERE (marker_last_run_all_accounts > ? OR created_at >= ?)
	AND marker_last_run_all_accounts <= ?
	ORDER BY marker_last_run_all_accounts ASC
	LIMIT %d;`, limit)

	defaultTime := U.DefaultTime()
	allDomainsRunHours := C.TimeRangeForAllDomains()
	createTime := U.TimeNowZ().Add(time.Duration(-allDomainsRunHours) * time.Hour)

	queryParams := []interface{}{defaultTime, createTime, createTime}

	rows, err := db.Raw(query, queryParams...).Rows()

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return []int64{}, http.StatusNotFound
		}
		log.WithError(err).Error("Error fetching top %d least updated projects", limit)
		return []int64{}, http.StatusInternalServerError
	}

	defer rows.Close()

	var projectIDs []int64

	for rows.Next() {
		var project_id int64
		err = rows.Scan(&project_id)
		if err != nil {
			log.WithError(err).Error("Error fetching rows")
			return []int64{}, http.StatusInternalServerError
		}
		projectIDs = append(projectIDs, project_id)
	}

	return projectIDs, http.StatusFound
}

// define a db method to fetch all the rowsp
func (store *MemSQL) GetFormFillEnabledProjectIDWithToken() (*map[int64]string, int) {
	type idWithToken struct {
		ProjectID int64
		Token     string
	}

	idsWithToken := make([]idWithToken, 0, 0)
	idWithTokenMap := map[int64]string{}

	db := C.GetServices().Db
	err := db.Table("project_settings").Select("projects.id as project_id, projects.token").
		Joins("LEFT JOIN projects ON project_settings.project_id = projects.id").
		Where("project_settings.auto_capture_form_fills = true").Find(&idsWithToken).Error
	if err != nil {
		log.WithError(err).Error("Fetching form fills enabled project_id failed")
		return &idWithTokenMap, http.StatusInternalServerError
	}

	if len(idsWithToken) == 0 {
		return &idWithTokenMap, http.StatusNotFound
	}

	for i := range idsWithToken {
		idWithTokenMap[idsWithToken[i].ProjectID] = idsWithToken[i].Token
	}

	return &idWithTokenMap, http.StatusFound
}

func (store *MemSQL) GetTimelinesConfig(projectID int64) (model.TimelinesConfig, error) {
	db := C.GetServices().Db
	var projectSettings model.ProjectSetting
	err := db.Table("project_settings").Select("timelines_config").Where("project_id=?", projectID).Find(&projectSettings).Error
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Failed to fetch timelines_config from project_settings.")
		return model.TimelinesConfig{}, err
	}
	if projectSettings.TimelinesConfig == nil {
		return model.TimelinesConfig{}, nil
	}
	timelinesConfig := projectSettings.TimelinesConfig.RawMessage
	tlConfigDecoded := model.TimelinesConfig{}
	err = json.Unmarshal(timelinesConfig, &tlConfigDecoded)
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Timeline Config Decode Failed")
		return model.TimelinesConfig{}, err
	}
	return tlConfigDecoded, nil
}

func (store *MemSQL) UpdateAccScoreWeights(projectId int64, weights model.AccWeights) error {

	weightjsonb, err := U.EncodeStructTypeToPostgresJsonb(weights)
	if err != nil {
		log.WithError(err).Error("Error creating postgres jsonb in acc score")
		return err
	}

	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("acc_score_weights", weightjsonb).Error; err != nil {
		log.WithError(err).Error("Updating is_explain_enabled config failed")
		return err
	}
	return err
}

func (store *MemSQL) GetWeightsByProject(project_id int64) (*model.AccWeights, int) {
	logFields := log.Fields{
		"project_id": project_id,
	}
	logCtx := log.WithFields(logFields)

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	var project_settings model.ProjectSetting
	var weights model.AccWeights
	if err := db.Table("project_settings").Limit(1).Where("project_id = ?", project_id).Find(&project_settings).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logCtx.WithField("project_id", project_id).WithError(err).Error("Unable to fetch weights from DB")
			return nil, http.StatusNotFound
		}
		logCtx.WithField("project_id", project_id).WithError(err).Error("Unable to fetch weights from DB")
		return &model.AccWeights{}, http.StatusInternalServerError
	}

	if project_settings.Acc_score_weights == nil {
		return nil, http.StatusBadRequest
	} else {
		err := U.DecodePostgresJsonbToStructType(project_settings.Acc_score_weights, &weights)
		if err != nil {
			logCtx.WithError(err).Error("Unable to decode weights")
			return nil, http.StatusInternalServerError
		}
	}

	return &weights, http.StatusFound
}

func (store *MemSQL) UpdateEngagementLevel(projectId int64, engagementBuckets model.BucketRanges) error {

	bucketsjsonb, err := U.EncodeStructTypeToPostgresJsonb(engagementBuckets)
	if err != nil {
		log.WithError(err).Error("Error creating postgres jsonb in acc score")
		return err
	}

	db := C.GetServices().Db
	if err := db.Model(&model.ProjectSetting{}).
		Where("project_id = ?", projectId).
		Update("custom_engagement_buckets", bucketsjsonb).Error; err != nil {
		log.WithError(err).Error("Updating is_explain_enabled config failed")
		return err
	}
	return err
}

func (store *MemSQL) GetEngagementLevelsByProject(project_id int64) (model.BucketRanges, int) {
	logFields := log.Fields{
		"project_id": project_id,
	}
	logCtx := log.WithFields(logFields)

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	var project_settings model.ProjectSetting
	var buckets model.BucketRanges
	if err := db.Table("project_settings").Limit(1).Where("project_id = ?", project_id).Find(&project_settings).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			logCtx.WithField("project_id", project_id).WithError(err).Error("Unable to fetch engagement buckets from DB")
			return model.BucketRanges{}, http.StatusNotFound
		}
		logCtx.WithField("project_id", project_id).WithError(err).Error("Unable to fetch engagement buckets from DB")
		return model.BucketRanges{}, http.StatusInternalServerError
	}

	if project_settings.CustomEngagementBuckets == nil {
		return model.BucketRanges{}, http.StatusNotFound
	} else {
		err := U.DecodePostgresJsonbToStructType(project_settings.CustomEngagementBuckets, &buckets)
		if err != nil {
			logCtx.WithError(err).Error("Unable to decode engagement buckets")
			return model.BucketRanges{}, http.StatusInternalServerError
		}
	}

	return buckets, http.StatusFound
}

func getNumberOfProjectsWithParagonEnabled() (int64, error) {
	var count int64
	db := C.GetServices().Db

	err := db.Model(&model.ProjectSetting{}).Not("int_paragon_token = ?", sqlNull).Count(&count).Error
	if err != nil {
		log.WithError(err).Error("error getting paragon enabled projects count")
		return count, err
	}

	return count, nil
}

func (store *MemSQL) GetParagonTokenFromProjectSetting(projectID int64) (string, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("invalid parameter")
		return "", http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	setting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		errMsg := "failed to fetch project settings"
		logCtx.Error(errMsg)
		return "", errCode, fmt.Errorf("failed to fetch project settings")
	}

	token := setting.IntParagonToken
	if token == "" {
		logCtx.Error("paragon not enabled yet")
		return "", http.StatusNotFound, fmt.Errorf("paragon not enabled yet")
	}

	return token, http.StatusFound, nil
}

func (store *MemSQL) GetParagonEnabledProjectsCount(projectID int64) (int64, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("invalid parameter")
		return 0, http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	var count int64
	db := C.GetServices().Db
	err := db.Model(&model.ProjectSetting{}).
		Not("int_paragon_token = ?", "NULL").Not("int_paragon_token = ?", "").Count(&count).Error
	if err != nil {
		errMsg := "failed to fetch project settings"
		logCtx.Error(errMsg)
		return 0, http.StatusInternalServerError, fmt.Errorf("failed to fetch paragon token count from project settings")
	}

	return count, http.StatusOK, nil
}

func (store *MemSQL) AddParagonTokenAndEnablingAgentToProjectSetting(projectID int64, agentID, token string) (int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"agent_id":   agentID,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("invalid parameter")
		return http.StatusBadRequest, fmt.Errorf("invalid parameter")
	}

	db := C.GetServices().Db
	if err := db.Model(model.ProjectSetting{}).Where("project_id = ?", projectID).
		Updates(map[string]interface{}{"int_paragon_token": token, "in_paragon_enabling_agent_id": agentID}).Error; err != nil {
		logCtx.WithError(err).Error("failed to update project_settings for paragon settings")
		return http.StatusInternalServerError, err
	}

	return http.StatusAccepted, nil
}

func (store *MemSQL) GetIntegrationState(projectID int64, docType string, isCRMTypeDoc bool) (model.IntegrationStatus, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"document_type": docType,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("invalid parameter")
		return model.IntegrationStatus{}, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var stmt string
	var params []interface{}
	if isCRMTypeDoc {
		if docType == model.LEADSQUARED || docType == model.MARKETO {

			stmt = fmt.Sprintf(`with 
			step_1 as (select synced, MAX(created_at) created_at  from crm_users where project_id = ? and source = %d group by synced),
			step_2 as (select synced,MAX(created_at) created_at  from crm_activities where project_id = ? and source = %d group by synced),
			step_3 as (select synced,MAX(created_at) created_at from crm_groups where project_id = ? and source = %d group by synced),
			step_4 as (select * from step_1
			union all
			select * from step_2
			union all
			select * from step_3)
			select synced,MAX(created_at) last_at from step_4 group by synced`, U.SourceCRM[docType], U.SourceCRM[docType], U.SourceCRM[docType])
			params = append(params, projectID, projectID, projectID)
		} else {
			stmt = fmt.Sprintf("SELECT synced, MAX(created_at) last_at from %s_documents where project_id= ? group by synced", docType)
			params = append(params, projectID)
		}
	} else {
		stmt = fmt.Sprintf("SELECT MAX(created_at) last_at from %s_documents where project_id= ? ", docType)
		params = append(params, projectID)
	}
	rows, err := db.Raw(stmt, params...).Rows()
	if err != nil {
		return model.IntegrationStatus{}, http.StatusInternalServerError
	}
	defer rows.Close()

	var lastSyncedAt, lastPulledAt int64

	for rows.Next() {
		var lastAtNull sql.NullTime
		var synced int

		var err error
		if isCRMTypeDoc {
			err = rows.Scan(&synced, &lastAtNull)
		} else {
			err = rows.Scan(&lastAtNull)
			synced = 1
		}

		if err != nil {
			logCtx.WithError(err).Error("Failed to get status for document type")
		}

		if lastAtNull.Valid {
			if synced == 1 {
				lastPulledAt = lastAtNull.Time.Unix()
			} else {
				lastSyncedAt = lastAtNull.Time.Unix()
			}
		}

	}

	result := model.IntegrationStatus{}

	a := U.TimeNowUnix()
	if (a-lastSyncedAt < 2*U.SECONDS_IN_A_DAY) && isCRMTypeDoc {
		result.State = model.SYNC_PENDING
	} else if a-lastPulledAt < 2*U.SECONDS_IN_A_DAY {
		result.State = model.PULL_DELAYED
	} else {
		result.State = model.SYNCED
	}

	if isCRMTypeDoc {
		result.LastSyncedAt = lastSyncedAt
	} else {
		result.LastSyncedAt = lastPulledAt
	}

	return result, http.StatusOK

}
