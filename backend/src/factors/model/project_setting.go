package model

import (
	C "factors/config"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
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
}

func GetProjectSetting(projectId uint64) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(projectId); !valid {
		return nil, http.StatusBadRequest
	}

	var projectSetting ProjectSetting
	if err := db.Where("project_id = ?", projectId).First(&projectSetting).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &projectSetting, http.StatusFound
}

func createProjectSetting(ps *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if valid := isValidProjectScope(ps.ProjectId); !valid {
		return nil, http.StatusBadRequest
	}

	if err := db.Create(ps).Error; err != nil {
		log.WithFields(log.Fields{"ProjectSetting": ps}).WithError(err).Error("Failed creating ProjectSetting.")
		return nil, http.StatusInternalServerError
	}

	return ps, http.StatusCreated
}

func UpdateProjectSettings(projectId uint64, settings *ProjectSetting) (*ProjectSetting, int) {
	db := C.GetServices().Db

	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	if settings.IntAdwordsCustomerAccountId != nil {
		// clean adwords customer_account_id before update.
		adwordsCustomerAccountId := strings.Replace(
			*settings.IntAdwordsCustomerAccountId, "-", "", -1)
		settings.IntAdwordsCustomerAccountId = &adwordsCustomerAccountId
	}

	var updatedProjectSetting ProjectSetting
	if err := db.Model(&updatedProjectSetting).Where("project_id = ?",
		projectId).Updates(settings).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"ProjectSetting": settings}).WithError(err).Error("Failed updating ProjectSettings.")
		return nil, http.StatusInternalServerError
	}

	return &updatedProjectSetting, http.StatusAccepted
}

func IsPSettingsIntSegmentEnabled(projectId uint64) bool {
	settings, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectId, "err_code": errCode}).Error("Failed fetching project settings.")
		return false
	}

	return *settings.IntSegment
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
}

var developmentShopifyInfo = map[string]shopifyInfoStruct{
	"aravind-test123.myshopify.com": shopifyInfoStruct{
		projectId: 2,
		apiKey:    "93f0ecd1ff038bb0de72ec1f4dcf34b3aecf2a2f15f1f531dbd89bfecb546b1e",
	},
}
var stagingShopifyInfo = map[string]shopifyInfoStruct{}
var productionShopifyInfo = map[string]shopifyInfoStruct{}

func GetProjectIdAndSecretByShopifyDomain(
	shopifyDomain string) (uint64, string, int) {
	var shopifyInfo map[string]shopifyInfoStruct
	if C.IsDevelopment() {
		shopifyInfo = developmentShopifyInfo
	} else if C.IsStaging() {
		shopifyInfo = stagingShopifyInfo
	} else if C.IsProduction() {
		shopifyInfo = productionShopifyInfo
	}
	if info, found := shopifyInfo[shopifyDomain]; found {
		return info.projectId, info.apiKey, http.StatusFound
	} else {
		log.Error(fmt.Sprintf("Unknown shopify domain - %s", shopifyDomain))
	}
	return 0, "", http.StatusInternalServerError
}
