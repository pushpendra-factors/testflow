package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/metrics"
)

type ProjectSetting struct {
	// Foreign key constraint project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId uint64 `gorm:"primary_key:true" json:"-"` // exclude on JSON response.
	// Using pointers to avoid update by default value.
	// omit empty to avoid nil(filelds not updated) on resp json.
	AutoTrack       *bool `gorm:"not null;default:false" json:"auto_track,omitempty"`
	AutoFormCapture *bool `gorm:"not null;default:false" json:"auto_form_capture,omitempty"`
	ExcludeBot      *bool `gorm:"not null;default:false" json:"exclude_bot,omitempty"`
	// Segment integration settings.
	IntSegment *bool `gorm:"not null;default:false" json:"int_segment,omitempty"`
	// Adwords integration settings.
	// Foreign key constraint int_adwords_enabled_agent_uuid -> agents(uuid)
	// Todo: Set int_adwords_enabled_agent_uuid, int_adwords_customer_account_id to NULL
	// for disabling adwords integration for the project.
	IntAdwordsEnabledAgentUUID  *string `json:"int_adwords_enabled_agent_uuid,omitempty"`
	IntAdwordsCustomerAccountId *string `json:"int_adwords_customer_account_id,omitempty"`
	// Hubspot integration settings.
	IntHubspot       *bool     `gorm:"not null;default:false" json:"int_hubspot,omitempty"`
	IntHubspotApiKey string    `json:"int_hubspot_api_key,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	//Facebook settings
	IntFacebookEmail       string  `json:"int_facebook_email,omitempty"`
	IntFacebookAccessToken string  `json:"int_facebook_access_token,omitempty"`
	IntFacebookAgentUUID   *string `json:"int_facebook_agent_uuid,omitempty"`
	IntFacebookUserID      string  `json:"int_facebook_user_id,omitempty"`
	IntFacebookAdAccount   string  `json:"int_facebook_ad_account,omitempty"`
	// Archival related fields.
	ArchiveEnabled  *bool `gorm:"default:false" json:"archive_enabled"`
	BigqueryEnabled *bool `gorm:"default:false" json:"bigquery_enabled"`
	//Salesforce settings
	IntSalesforceEnabledAgentUUID *string `json:"int_salesforce_enabled_agent_uuid,omitempty"`
}

const ProjectSettingKeyToken = "token"
const ProjectSettingKeyPrivateToken = "private_token"

var projectSettingKeys = [...]string{
	ProjectSettingKeyToken,
	ProjectSettingKeyPrivateToken,
}

// Salesforce required fields per project
var (
	designcafeLeadsAllowedFields = map[string]bool{
		"Id":                                 true,
		"LastName":                           true,
		"Salutation":                         true,
		"Company":                            true,
		"Street":                             true,
		"City":                               true,
		"State":                              true,
		"Phone":                              true,
		"MobilePhone":                        true,
		"Email":                              true,
		"LeadSource":                         true,
		"Status":                             true,
		"ConvertedDate":                      true,
		"CreatedDate":                        true,
		"CreatedById":                        true,
		"LastModifiedDate":                   true,
		"LastModifiedById":                   true,
		"LastActivityDate":                   true,
		"DC_Lead_Status__c":                  true,
		"Alternate_Contact_Number__c":        true,
		"Approx_Budget__c":                   true,
		"Call_Center_Agent__c":               true,
		"Call_Count__c":                      true,
		"Call_Stage__c":                      true,
		"CMM_Name__c":                        true,
		"Channel__c":                         true,
		"DC_Lead_Source__c":                  true,
		"DSA__c":                             true,
		"Designer__c":                        true,
		"First_Date_of_Contact__c":           true,
		"Follow_Up_Count__c":                 true,
		"Follow_Up_Date_Time__c":             true,
		"GClid__c":                           true,
		"Home_Type__c":                       true,
		"Interior_work_needed_for__c":        true,
		"Is_Designer_Assigned__c":            true,
		"Last_Source__c":                     true,
		"Lead_Allocation_Time__c":            true,
		"Lead_Owner_Name__c":                 true,
		"Lead_Owner_Region__c":               true,
		"Lead_Owner_Team__c":                 true,
		"Lead_Qualified_Date__c":             true,
		"CMM_Team__c":                        true,
		"Has_Designer_Accepted__c":           true,
		"Meeting_Scheduled_on_First_Call__c": true,
		"Meeting_Type__c":                    true,
		"Page_URL__c":                        true,
		"Pre_Qualified_Date__c":              true,
		"Project_Name__c":                    true,
		"Region__c":                          true,
		"Requirement_Details__c":             true,
		"Source_Journey__c":                  true,
		"Source__c":                          true,
		"Lead_Owner_Role__c":                 true,
		"Property_Type__c":                   true,
		"Willingness_For_Meeting__c":         true,
		"Ad_Group__c":                        true,
		"Ad_Name__c":                         true,
		"Ad_Network__c":                      true,
		"Budget__c":                          true,
		"Campagin__c":                        true,
		"DC_Campaign_Source__c":              true,
		"Device_Type__c":                     true,
		"EC_Location__c":                     true,
		"Enquiry_ID__c":                      true,
		"IP_Address__c":                      true,
		"Meeting_Venue__c":                   true,
		"Page_URL_1__c":                      true,
		"Property_Possession_Date__c":        true,
		"Qualified_Status__c":                true,
		"Recontacted__c":                     true,
		"Affiliate_Name__c":                  true,
		"first_date_of_contact_to_Qualified_c__c": true,
		"Keyword__c":                        true,
		"Custom_Lead_ID__c":                 true,
		"Pre_Qualified_FollowUp_Date__c":    true,
		"Date_When_Meeting_is_Scheduled__c": true,
		"Mobile_Number_External_Field__c":   true,
		"Country_Code__c":                   true,
		"OTP_Verified__c":                   true,
		"User_Mobile__c":                    true,
		"Messaging_Source__c":               true,
		"Entry_Url__c":                      true,
		"First_Visit_Time__c":               true,
		"IPAddress__c":                      true,
		"Page__c":                           true,
		"Time_On_Last_page__c":              true,
		"User_Browser__c":                   true,
		"User_Last_Url__c":                  true,
		"User_OS__c":                        true,
		"Lockdown_Survey__c":                true,
		"Civil_Work__c":                     true,
		"Whatsapp_Opt_IN__c":                true,
		"MobileYM__c":                       true,
		"Customer_WhatsApp_OptIN__c":        true,
	}

	designcafeOpportunityAllowedFields = map[string]bool{
		"Id":                             true, // require for identification purpose
		"Name":                           true,
		"StageName":                      true,
		"Amount":                         true,
		"ExpectedRevenue":                true,
		"CloseDate":                      true,
		"Type":                           true,
		"LeadSource":                     true,
		"IsWon":                          true,
		"CreatedDate":                    true,
		"LastModifiedDate":               true,
		"Budget_Confirmed__c":            true,
		"Affiliate_Name__c":              true,
		"Loss_Reason__c":                 true,
		"Call_Center_Agent__c":           true,
		"Call_Stage__c":                  true,
		"CMM_Team__c":                    true,
		"Channel__c":                     true,
		"Client_s_Budget__c":             true,
		"Signup_Amount__c":               true,
		"DSA__c":                         true,
		"Designer__c":                    true,
		"Home_Type__c":                   true,
		"Lead_Stage__c":                  true,
		"Lead_Status__c":                 true,
		"Meeting_Date_and_Time__c":       true,
		"Meeting_Scheduled_Date_Time__c": true,
		"Meeting_Type__c":                true,
		"Meeting_Venue__c":               true,
		"Offer_and_discounts__c":         true,
		"Possession_Status__c":           true,
		"Project_Name__c":                true,
		"Property_Address__c":            true,
		"Proposed_Budget__c":             true,
		"CMM_Name__c":                    true,
		"Region__c":                      true,
		"Source__c":                      true,
		"Total_Amount__c":                true,
		"Approx_Budget__c":               true,
		"Ad_Group__c":                    true,
		"Ad_Name__c":                     true,
		"Campaign__c":                    true,
		"Enquiry_ID__c":                  true,
		"Overall_Sales_Duration__c":      true,
		"Campaign_Source__c":             true,
		"DC_Lead_Source__c":              true,
		"Reason_for_Loss__c":             true,
		"Closing_Offer__c":               true,
		"Mobile__c":                      true,
		"Wohoo_Card__c":                  true,
		"Customer_ID__c":                 true,
		"Phone__c":                       true,
		"Messaging_Source__c":            true,
		"Payment_Mode__c":                true,
		"Packages__c":                    true,
		"MobileYM__c":                    true,
	}

	designcafeAllowedObjects = map[string]map[string]bool{
		SalesforceDocumentTypeNameLead:        designcafeLeadsAllowedFields,
		SalesforceDocumentTypeNameOpportunity: designcafeOpportunityAllowedFields,
	}

	SalesforceProjectStore = map[uint64]map[string]map[string]bool{
		483: designcafeAllowedObjects,
	}
)

// GetSalesforceAllowedObjects return allowed object type for a project
func GetSalesforceAllowedObjects(projectID uint64) []int {
	var docTypes []int
	if projectID == 0 {
		return docTypes
	}

	if objects, exist := SalesforceProjectStore[projectID]; exist {
		for name := range objects {
			docType := GetSalesforceDocTypeByAlias(name)
			docTypes = append(docTypes, docType)
		}
		return docTypes
	} else {
		docTypes = SalesforceStandardDocumentType
	}
	return docTypes
}

// GetSalesforceAllowedfiedsByObject return list of allowed field for a project
func GetSalesforceAllowedfiedsByObject(projectID uint64, objectName string) map[string]bool {
	if projectID == 0 {
		return nil
	}

	if objects, exist := SalesforceProjectStore[projectID]; exist {
		if _, objExist := objects[objectName]; objExist {
			return objects[objectName]
		}
	}

	return nil
}

func GetProjectSetting(projectId uint64) (*ProjectSetting, int) {
	db := C.GetServices().Db
	logCtx := log.WithField("project_id", projectId)

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest
	}

	var projectSetting ProjectSetting
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
	Setting   *ProjectSetting
	ErrorCode int
}

// GetProjectSettingByKeyWithTimeout - Get project_settings from db based on key,
// gets timedout and returns StatusInternalServerError, if the query takes more than
// the given duration. Returns default project_settings immediately, if the
// config/flag use_default_project_setting_for_sdk is set to true.
func GetProjectSettingByKeyWithTimeout(key, value string, timeout time.Duration) (*ProjectSetting, int) {
	if C.GetConfig().UseDefaultProjectSettingForSDK {
		return getProjectSettingDefault(), http.StatusFound
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

// getProjectSettingByKey - Get project settings by a column on projects.
func getProjectSettingByKey(key, value string) (*ProjectSetting, int) {
	if key == "" || value == "" {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithField("key", key).WithField("value", value)

	var setting ProjectSetting
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

func getCacheProjectSetting(tokenKey, tokenValue string) (*ProjectSetting, int) {
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

	var settings ProjectSetting
	err = json.Unmarshal([]byte(settingsJson), &settings)
	if err != nil {
		log.WithError(err).Error(
			"Failed to unmarshal cached project settings on getCacheProjectSetting")
		return nil, http.StatusInternalServerError
	}

	return &settings, http.StatusFound
}

func setCacheProjectSetting(tokenKey, tokenValue string, settings *ProjectSetting) int {
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

func getProjectSettingDefault() *ProjectSetting {
	enabled := true
	return &ProjectSetting{
		AutoTrack:       &enabled,
		AutoFormCapture: &enabled,
		ExcludeBot:      &enabled,
		IntSegment:      &enabled,
	}
}

// getProjectSettingByKeyWithDefault - Get from cache or db, if not use default.
func getProjectSettingByKeyWithDefault(tokenKey, tokenValue string) (*ProjectSetting, int) {
	settings, errCode := getCacheProjectSetting(tokenKey, tokenValue)
	if errCode == http.StatusFound {
		return settings, http.StatusFound
	}

	settings, errCode = GetProjectSettingByKeyWithTimeout(tokenKey, tokenValue, time.Millisecond*30)
	if errCode != http.StatusFound {
		// Use default settings, if db failure.
		// Do not cache default.
		return getProjectSettingDefault(), http.StatusFound
	}

	// add to cache.
	setCacheProjectSetting(tokenKey, tokenValue, settings)

	return settings, http.StatusFound
}

func GetProjectSettingByTokenWithCacheAndDefault(token string) (*ProjectSetting, int) {
	return getProjectSettingByKeyWithDefault(ProjectSettingKeyToken, token)
}

func GetProjectSettingByPrivateTokenWithCacheAndDefault(
	privateToken string) (*ProjectSetting, int) {

	return getProjectSettingByKeyWithDefault(
		ProjectSettingKeyPrivateToken, privateToken)
}

func createProjectSetting(ps *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(ps.ProjectId); !valid {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(ps).Error; err != nil {
		log.WithFields(log.Fields{"ProjectSetting": ps}).WithError(
			err).Error("Failed creating ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return ps, http.StatusCreated
}

func delAllProjectSettingsCacheForProject(projectId uint64) {
	project, errCode := GetProject(projectId)
	if errCode != http.StatusFound {
		log.Error("Failed to get project on delAllProjectSettingsCacheKeys.")
	}

	// delete all project setting cache keys by respective
	// token key and value.
	delCacheProjectSetting(ProjectSettingKeyToken, project.Token)
	delCacheProjectSetting(ProjectSettingKeyPrivateToken, project.PrivateToken)
}

func UpdateProjectSettings(projectId uint64, settings *ProjectSetting) (*ProjectSetting, int) {
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

	var updatedProjectSetting ProjectSetting
	if err := db.Model(&updatedProjectSetting).Where("project_id = ?",
		projectId).Updates(settings).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"ProjectSetting": settings}).WithError(
			err).Error("Failed updating ProjectSettings.")
		return nil, http.StatusInternalServerError
	}

	delAllProjectSettingsCacheForProject(projectId)

	return &updatedProjectSetting, http.StatusAccepted
}

func IsPSettingsIntShopifyEnabled(projectId uint64) bool {
	return true
}

func GetIntAdwordsRefreshTokenForProject(projectId uint64) (string, int) {
	settings, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return "", errCode
	}

	if settings.IntAdwordsEnabledAgentUUID == nil || *settings.IntAdwordsEnabledAgentUUID == "" {
		return "", http.StatusNotFound
	}

	logCtx := log.WithField("agent_uuid",
		*settings.IntAdwordsEnabledAgentUUID).WithField("project_id", projectId)

	agent, errCode := GetAgentByUUID(*settings.IntAdwordsEnabledAgentUUID)
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

type AdwordsProjectSettings struct {
	ProjectId         uint64
	CustomerAccountId string
	AgentUUID         string
	RefreshToken      string
}

func GetAllIntAdwordsProjectSettings() ([]AdwordsProjectSettings, int) {
	db := C.GetServices().Db

	adwordsProjectSettings := make([]AdwordsProjectSettings, 0, 0)

	queryStr := "SELECT project_settings.project_id, project_settings.int_adwords_customer_account_id as customer_account_id," +
		" " + "agents.int_adwords_refresh_token as refresh_token, project_settings.int_adwords_enabled_agent_uuid as agent_uuid" +
		" " + "FROM project_settings LEFT JOIN agents ON project_settings.int_adwords_enabled_agent_uuid = agents.uuid" +
		" " + "WHERE project_settings.int_adwords_customer_account_id IS NOT NULL" +
		" " + "AND project_settings.int_adwords_enabled_agent_uuid IS NOT NULL"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get all adwords project settings.")
		return adwordsProjectSettings, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var adwordsSettings AdwordsProjectSettings
		if err := db.ScanRows(rows, &adwordsSettings); err != nil {
			log.WithError(err).Error("Failed to scan get all adwords project settings.")
			return adwordsProjectSettings, http.StatusInternalServerError
		}

		adwordsProjectSettings = append(adwordsProjectSettings, adwordsSettings)
	}

	return adwordsProjectSettings, http.StatusOK
}

type HubspotProjectSettings struct {
	ProjectId uint64 `json:"-"`
	APIKey    string `json:"api_key"`
}

func GetAllHubspotProjectSettings() ([]HubspotProjectSettings, int) {
	var hubspotProjectSettings []HubspotProjectSettings

	db := C.GetServices().Db
	err := db.Table("project_settings").Where(
		"int_hubspot='true' AND int_hubspot_api_key IS NOT NULL ").Select(
		"project_id, int_hubspot_api_key as api_key").Find(
		&hubspotProjectSettings).Error
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

func GetProjectDetailsByShopifyDomain(
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

type FacebookProjectSettings struct {
	ProjectId              string `json:"project_id"`
	IntFacebookUserId      string `json:"int_facebook_user_id"`
	IntFacebookAccessToken string `json:"int_facebook_access_token"`
	IntFacebookAdAccount   string `json:"int_facebook_ad_account"`
	IntFacebookEmail       string `json:"int_facebook_email"`
}

func GetFacebookEnabledProjectSettings() ([]FacebookProjectSettings, int) {
	db := C.GetServices().Db

	facebookProjectSettings := make([]FacebookProjectSettings, 0, 0)

	err := db.Table("project_settings").Where("int_facebook_access_token IS NOT NULL AND int_facebook_access_token != ''").Find(&facebookProjectSettings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get facebook enabled project settings for sync info.")
		return facebookProjectSettings, http.StatusInternalServerError
	}
	return facebookProjectSettings, http.StatusOK
}

// GetArchiveEnabledProjectIDs Returns list of project ids which have archive enabled.
func GetArchiveEnabledProjectIDs() ([]uint64, int) {
	var projectIDs []uint64
	db := C.GetServices().Db

	rows, err := db.Model(&ProjectSetting{}).Where("archive_enabled = true").Select("project_id").Rows()
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
func GetBigqueryEnabledProjectIDs() ([]uint64, int) {
	var projectIDs []uint64
	db := C.GetServices().Db

	rows, err := db.Model(&ProjectSetting{}).Where("bigquery_enabled = true").Select("project_id").Rows()
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

// SalesforceProjectSettings contains refresh_token and instance_url for enabled projects
type SalesforceProjectSettings struct {
	ProjectID    uint64 `json:"-"`
	RefreshToken string `json:"refresh_token"`
	InstanceURL  string `json:"instance_url"`
}

// GetAllSalesforceProjectSettings return list of all enabled salesforce projects and their meta data
func GetAllSalesforceProjectSettings() ([]SalesforceProjectSettings, int) {
	var salesforceProjectSettings []SalesforceProjectSettings

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
