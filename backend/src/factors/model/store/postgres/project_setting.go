package postgres

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/metrics"
	"factors/model/model"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

func (pg *Postgres) GetProjectSetting(projectId uint64) (*model.ProjectSetting, int) {
	db := C.GetServices().Db
	logCtx := log.WithField("project_id", projectId)

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

type ProjectSettingChannelResponse struct {
	Setting   *model.ProjectSetting
	ErrorCode int
}

// GetProjectSettingByKeyWithTimeout - Get project_settings from db based on key,
// gets timedout and returns StatusInternalServerError, if the query takes more than
// the given duration. Returns default project_settings immediately, if the
// config/flag use_default_project_setting_for_sdk is set to true.
func (pg *Postgres) GetProjectSettingByKeyWithTimeout(key, value string, timeout time.Duration) (*model.ProjectSetting, int) {
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
func (pg *Postgres) EnableBigqueryArchivalForProject(projectID uint64) int {
	db := C.GetServices().Db
	logCtx := log.WithField("project_id", projectID)

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

// getProjectSettingByKey - Get project settings by a column on projects.
func getProjectSettingByKey(key, value string) (*model.ProjectSetting, int) {
	if key == "" || value == "" {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithField("key", key).WithField("value", value)

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
	// table_name:column_name
	prefix := fmt.Sprintf("%s:%s", "project_settings", tokenKey)
	return cacheRedis.NewKeyWithProjectUID(tokenValue, prefix, "")
}

func getCacheProjectSetting(tokenKey, tokenValue string) (*model.ProjectSetting, int) {
	logCtx := log.WithField("token_value", tokenValue)

	if tokenValue == "" {
		return nil, http.StatusBadRequest
	}

	key, err := getProjectSettingCacheKey(tokenKey, tokenValue)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on getCacheProjectSetting")
		return nil, http.StatusInternalServerError
	}

	settingsJson, err := cacheRedis.Get(key)
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

	return &settings, http.StatusFound
}

func setCacheProjectSetting(tokenKey, tokenValue string, settings *model.ProjectSetting) int {
	logCtx := log.WithField("token_value", tokenValue)

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
	err = cacheRedis.Set(key, string(settingsJson), expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache on setCacheProjectSetting")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func delCacheProjectSetting(tokenKey, tokenValue string) int {
	logCtx := log.WithField("token_key", tokenKey)

	if tokenValue == "" {
		return http.StatusBadRequest
	}

	key, err := getProjectSettingCacheKey(tokenKey, tokenValue)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get project settings by token cache key on delCacheProjectSetting")
		return http.StatusInternalServerError
	}

	err = cacheRedis.Del(key)
	if err != nil && err != redis.ErrNil {
		logCtx.WithError(err).Error("Failed to del cache on delCacheProjectSetting")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func getProjectSettingDefault() *model.ProjectSetting {
	enabled := true
	disabled := false
	return &model.ProjectSetting{
		AutoTrack:       &enabled,
		AutoFormCapture: &enabled,
		ExcludeBot:      &enabled,
		IntSegment:      &enabled,
		IntDrift:        &disabled,
		IntClearBit:     &disabled,
	}
}

// getProjectSettingByKeyWithDefault - Get from cache or db, if not use default.
func (pg *Postgres) getProjectSettingByKeyWithDefault(tokenKey, tokenValue string) (*model.ProjectSetting, int) {
	settings, errCode := getCacheProjectSetting(tokenKey, tokenValue)
	if errCode == http.StatusFound {
		return settings, http.StatusFound
	}

	settings, errCode = pg.GetProjectSettingByKeyWithTimeout(tokenKey, tokenValue, time.Millisecond*50)
	if errCode != http.StatusFound {
		// Use default settings, if db failure.
		// Do not cache default.
		return getProjectSettingDefault(), http.StatusFound
	}

	// add to cache.
	setCacheProjectSetting(tokenKey, tokenValue, settings)

	return settings, http.StatusFound
}

func (pg *Postgres) GetProjectSettingByTokenWithCacheAndDefault(token string) (*model.ProjectSetting, int) {
	return pg.getProjectSettingByKeyWithDefault(model.ProjectSettingKeyToken, token)
}

func (pg *Postgres) GetProjectSettingByPrivateTokenWithCacheAndDefault(
	privateToken string) (*model.ProjectSetting, int) {

	return pg.getProjectSettingByKeyWithDefault(
		model.ProjectSettingKeyPrivateToken, privateToken)
}

func createProjectSetting(ps *model.ProjectSetting) (*model.ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(ps.ProjectId); !valid {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(ps).Error; err != nil {
		log.WithFields(log.Fields{"model.ProjectSetting": ps}).WithError(
			err).Error("Failed creating model.ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return ps, http.StatusCreated
}

func (pg *Postgres) delAllProjectSettingsCacheForProject(projectId uint64) {
	project, errCode := pg.GetProject(projectId)
	if errCode != http.StatusFound {
		log.Error("Failed to get project on delAllProjectSettingsCacheKeys.")
	}

	// delete all project setting cache keys by respective
	// token key and value.
	delCacheProjectSetting(model.ProjectSettingKeyToken, project.Token)
	delCacheProjectSetting(model.ProjectSettingKeyPrivateToken, project.PrivateToken)
}

func (pg *Postgres) UpdateProjectSettings(projectId uint64, settings *model.ProjectSetting) (*model.ProjectSetting, int) {
	db := C.GetServices().Db

	if projectId == 0 || settings == nil {
		return nil, http.StatusBadRequest
	}

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

	if settings.IntHubspotApiKey != "" {
		existingSettings, status := pg.GetProjectSetting(projectId)
		if status != http.StatusFound {
			return nil, http.StatusInternalServerError
		}

		if existingSettings.IntHubspotApiKey != settings.IntHubspotApiKey {
			hubspotIntegrationAccount, err := model.GetHubspotIntegrationAccount(settings.IntHubspotApiKey)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectId}).WithError(
					err).Error("Error fetching hubspot account details on integration.") // log error but still allow integration
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

	pg.delAllProjectSettingsCacheForProject(projectId)

	return &updatedProjectSetting, http.StatusAccepted
}

func (pg *Postgres) IsPSettingsIntShopifyEnabled(projectId uint64) bool {
	return true
}

func (pg *Postgres) GetIntAdwordsRefreshTokenForProject(projectId uint64) (string, int) {
	settings, errCode := pg.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return "", errCode
	}

	if settings.IntAdwordsEnabledAgentUUID == nil || *settings.IntAdwordsEnabledAgentUUID == "" {
		return "", http.StatusNotFound
	}

	logCtx := log.WithField("agent_uuid",
		*settings.IntAdwordsEnabledAgentUUID).WithField("project_id", projectId)

	agent, errCode := pg.GetAgentByUUID(*settings.IntAdwordsEnabledAgentUUID)
	if errCode != http.StatusFound {
		logCtx.Error("Adwords enabled agent not found on agents table.")
		return "", errCode
	}

	refreshToken := agent.IntAdwordsRefreshToken
	if refreshToken == "" {
		logCtx.Error("Adwords enabled agent refresh token is empty.")
		return "", http.StatusInternalServerError
	}

	return refreshToken, http.StatusFound
}
func (pg *Postgres) GetIntGoogleOrganicRefreshTokenForProject(projectId uint64) (string, int) {
	settings, errCode := pg.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return "", errCode
	}

	if settings.IntGoogleOrganicEnabledAgentUUID == nil || *settings.IntGoogleOrganicEnabledAgentUUID == "" {
		return "", http.StatusNotFound
	}

	logCtx := log.WithField("agent_uuid",
		*settings.IntGoogleOrganicEnabledAgentUUID).WithField("project_id", projectId)

	agent, errCode := pg.GetAgentByUUID(*settings.IntGoogleOrganicEnabledAgentUUID)
	if errCode != http.StatusFound {
		logCtx.Error("GoogleOrganic enabled agent not found on agents table.")
		return "", errCode
	}

	refreshToken := agent.IntGoogleOrganicRefreshToken
	if refreshToken == "" {
		logCtx.Error("GoogleOrganic enabled agent refresh token is empty.")
		return "", http.StatusInternalServerError
	}

	return refreshToken, http.StatusFound
}

func (pg *Postgres) GetIntAdwordsProjectSettingsForProjectID(projectID uint64) ([]model.AdwordsProjectSettings, int) {

	queryStr := "SELECT project_settings.project_id, project_settings.int_adwords_customer_account_id as customer_account_id," +
		" " + "project_settings.int_google_ingestion_timezone as int_google_ingestion_timezone," +
		" " + "agents.int_adwords_refresh_token as refresh_token, project_settings.int_adwords_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_adwords_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.project_id = ?" +
		" " + "AND project_settings.int_adwords_customer_account_id IS NOT NULL" +
		" " + "AND project_settings.int_adwords_enabled_agent_uuid IS NOT NULL "
	params := []interface{}{projectID}

	return pg.getIntAdwordsProjectSettings(queryStr, params)
}
func (pg *Postgres) GetAllIntAdwordsProjectSettings() ([]model.AdwordsProjectSettings, int) {

	queryStr := "SELECT project_settings.project_id, project_settings.int_adwords_customer_account_id as customer_account_id," +
		" " + "project_settings.int_google_ingestion_timezone as int_google_ingestion_timezone," +
		" " + "agents.int_adwords_refresh_token as refresh_token, project_settings.int_adwords_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_adwords_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.int_adwords_customer_account_id IS NOT NULL" +
		" " + "AND project_settings.int_adwords_enabled_agent_uuid IS NOT NULL"
	params := make([]interface{}, 0, 0)

	return pg.getIntAdwordsProjectSettings(queryStr, params)
}

func (pg *Postgres) getIntAdwordsProjectSettings(query string, params []interface{}) ([]model.AdwordsProjectSettings, int) {
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

func (pg *Postgres) GetIntGoogleOrganicProjectSettingsForProjectID(projectID uint64) ([]model.GoogleOrganicProjectSettings, int) {

	queryStr := "SELECT project_settings.project_id, project_settings.int_google_organic_url_prefixes as url_prefix," +
		" " + "agents.int_google_organic_refresh_token as refresh_token, project_settings.int_google_organic_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_google_organic_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.project_id = ?" +
		" " + "AND project_settings.int_google_organic_url_prefixes IS NOT NULL" +
		" " + "AND project_settings.int_google_organic_enabled_agent_uuid IS NOT NULL "
	params := []interface{}{projectID}

	return pg.getIntGoogleOrganicProjectSettings(queryStr, params)
}

func (pg *Postgres) GetAllIntGoogleOrganicProjectSettings() ([]model.GoogleOrganicProjectSettings, int) {

	queryStr := "SELECT project_settings.project_id, project_settings.int_google_organic_url_prefixes as url_prefix," +
		" " + "agents.int_google_organic_refresh_token as refresh_token, project_settings.int_google_organic_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_google_organic_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.int_google_organic_url_prefixes IS NOT NULL" +
		" " + "AND project_settings.int_google_organic_enabled_agent_uuid IS NOT NULL "
	params := make([]interface{}, 0, 0)

	return pg.getIntGoogleOrganicProjectSettings(queryStr, params)
}
func (pg *Postgres) getIntGoogleOrganicProjectSettings(query string, params []interface{}) ([]model.GoogleOrganicProjectSettings, int) {
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

func (pg *Postgres) GetAllHubspotProjectSettings() ([]model.HubspotProjectSettings, int) {
	var hubspotProjectSettings []model.HubspotProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_hubspot='true' AND int_hubspot_api_key IS NOT NULL ").Select(
		"project_id, int_hubspot_api_key as api_key, int_hubspot_first_time_synced as is_first_time_synced," +
			"int_hubspot_sync_info as sync_info").
		Find(&hubspotProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot project_settings.")
		return hubspotProjectSettings, http.StatusInternalServerError
	}

	return hubspotProjectSettings, http.StatusFound
}

func (pg *Postgres) GetAllHubspotProjectSettingsForProjectID(projectID uint64) ([]model.HubspotProjectSettings, int) {
	var hubspotProjectSettings []model.HubspotProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_hubspot='true' AND int_hubspot_api_key IS NOT NULL AND project_id IN (?)", projectID).Select(
		"project_id, int_hubspot_api_key as api_key, int_hubspot_first_time_synced as is_first_time_synced," +
			"int_hubspot_sync_info as sync_info").
		Find(&hubspotProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get hubspot project_settings.")
		return hubspotProjectSettings, http.StatusInternalServerError
	}

	return hubspotProjectSettings, http.StatusFound
}

type shopifyInfoStruct struct {
	apiKey    string
	projectId uint64
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

func (pg *Postgres) GetProjectDetailsByShopifyDomain(
	shopifyDomain string) (uint64, string, bool, int) {
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

func (pg *Postgres) GetFacebookEnabledIDsAndProjectSettings() ([]uint64, []model.FacebookProjectSettings, int) {
	db := C.GetServices().Db

	facebookProjectSettings := make([]model.FacebookProjectSettings, 0, 0)
	facebookIDs := make([]uint64, 0, 0)

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

func (pg *Postgres) GetFacebookEnabledIDsAndProjectSettingsForProject(projectIDs []uint64) ([]uint64, []model.FacebookProjectSettings, int) {
	db := C.GetServices().Db

	facebookProjectSettings := make([]model.FacebookProjectSettings, 0, 0)
	facebookIDs := make([]uint64, 0, 0)

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

func (pg *Postgres) GetLinkedinEnabledProjectSettings() ([]model.LinkedinProjectSettings, int) {
	db := C.GetServices().Db

	linkedinProjectSettings := make([]model.LinkedinProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_linkedin_refresh_token IS NOT NULL AND int_linkedin_refresh_token != ''").Find(&linkedinProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get linkedin enabled project settings for sync info.")
		return linkedinProjectSettings, http.StatusInternalServerError
	}
	return linkedinProjectSettings, http.StatusOK
}
func (pg *Postgres) GetLinkedinEnabledProjectSettingsForProjects(projectIDs []string) ([]model.LinkedinProjectSettings, int) {
	db := C.GetServices().Db

	linkedinProjectSettings := make([]model.LinkedinProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_linkedin_refresh_token IS NOT NULL AND int_linkedin_refresh_token != '' AND project_id IN (?)", projectIDs).Find(&linkedinProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get linkedin enabled project settings for sync info.")
		return linkedinProjectSettings, http.StatusInternalServerError
	}
	return linkedinProjectSettings, http.StatusOK
}

// GetArchiveEnabledProjectIDs Returns list of project ids which have archive enabled.
func (pg *Postgres) GetArchiveEnabledProjectIDs() ([]uint64, int) {
	var projectIDs []uint64
	db := C.GetServices().Db

	rows, err := db.Model(&model.ProjectSetting{}).Where("archive_enabled = true").Select("project_id").Rows()
	if err != nil {
		log.WithError(err).Error("Query failed for GetArchiveEnabledProjectIDs")
		return projectIDs, http.StatusInternalServerError
	}

	for rows.Next() {
		var projectID uint64
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
func (pg *Postgres) GetBigqueryEnabledProjectIDs() ([]uint64, int) {
	var projectIDs []uint64
	db := C.GetServices().Db

	rows, err := db.Model(&model.ProjectSetting{}).Where("bigquery_enabled = true").Select("project_id").Rows()
	if err != nil {
		log.WithError(err).Error("Query failed for GetBigqueryEnabledProjectIDs")
		return projectIDs, http.StatusInternalServerError
	}

	for rows.Next() {
		var projectID uint64
		err = rows.Scan(&projectID)
		if err != nil {
			log.WithError(err).Error("Error while scanning")
			continue
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, http.StatusFound
}

// GetAllSalesforceProjectSettings return list of all enabled salesforce projects and their meta data
func (pg *Postgres) GetAllSalesforceProjectSettings() ([]model.SalesforceProjectSettings, int) {
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

func (pg *Postgres) GetAllSalesforceProjectSettingsForProject(projectID uint64) ([]model.SalesforceProjectSettings, int) {
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

func (pg *Postgres) GetAdwordsEnabledProjectIDAndCustomerIDsFromProjectSettings() (map[uint64][]string, error) {
	db := C.GetServices().Db

	projectSettings := make([]model.ProjectSetting, 0, 0)
	mapOfProjectToCustomerIds := make(map[uint64][]string)

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
func (pg *Postgres) IsBingIntegrationAvailable(projectID uint64) bool {
	ftMapping, err := pg.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		return false
	}
	if ftMapping.ConnectorID == "" {
		return false
	}
	return true
}
func (pg *Postgres) DeleteChannelIntegration(projectID uint64, channelName string) (int, error) {
	if projectID == 0 {
		return http.StatusBadRequest, errors.New("invalid projectID")
	}
	switch channelName {
	case "facebook":
		return pg.DeleteFacebookIntegration(projectID)
	case "linkedin":
		return pg.DeleteLinkedinIntegration(projectID)
	case "adwords":
		return pg.DeleteAdwordsIntegration(projectID)
	case "google_organic":
		return pg.DeleteGoogleOrganicIntegration(projectID)
	default:
		return http.StatusBadRequest, errors.New("invalid channel name")
	}
}
