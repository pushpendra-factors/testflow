package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type ProjectSetting struct {
	// Foreign key constraint project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId         uint64          `gorm:"primary_key:true" json:"project_id,omitempty"`
	AttributionConfig *postgres.Jsonb `json:"attribution_config"`

	// Using pointers to avoid update by default value.
	// omit empty to avoid nil(filelds not updated) on resp json.
	AutoTrack            *bool `gorm:"not null;default:false" json:"auto_track,omitempty"`
	AutoTrackSPAPageView *bool `gorm:"not null;default:false" json:"auto_track_spa_page_view"`
	AutoFormCapture      *bool `gorm:"not null;default:false" json:"auto_form_capture,omitempty"`
	ExcludeBot           *bool `gorm:"not null;default:false" json:"exclude_bot,omitempty"`
	// Segment integration settings.
	IntSegment *bool `gorm:"not null;default:false" json:"int_segment,omitempty"`
	// Adwords integration settings.
	// Foreign key constraint int_adwords_enabled_agent_uuid -> agents(uuid)
	// Todo: Set int_adwords_enabled_agent_uuid, int_adwords_customer_account_id to NULL
	// for disabling adwords integration for the project.
	IntAdwordsEnabledAgentUUID  *string `json:"int_adwords_enabled_agent_uuid,omitempty"`
	IntAdwordsCustomerAccountId *string `json:"int_adwords_customer_account_id,omitempty"`
	// for disabling google search console integration for the project.
	IntGoogleOrganicEnabledAgentUUID *string `json:"int_google_organic_enabled_agent_uuid,omitempty"`
	IntGoogleOrganicURLPrefixes      *string `json:"int_google_organic_url_prefixes,omitempty"`
	// Hubspot integration settings.
	IntHubspot                *bool           `gorm:"not null;default:false" json:"int_hubspot,omitempty"`
	IntHubspotApiKey          string          `json:"int_hubspot_api_key,omitempty"`
	IntHubspotFirstTimeSynced bool            `json:"int_hubspot_first_time_synced,omitempty"`
	IntHubspotPortalID        *int            `json:"int_hubspot_portal_id,omitempty"`
	IntHubspotSyncInfo        *postgres.Jsonb `json:"int_hubspot_sync_info,omitempty" `
	CreatedAt                 time.Time       `json:"created_at"`
	UpdatedAt                 time.Time       `json:"updated_at"`
	//Facebook settings
	IntFacebookEmail       string  `json:"int_facebook_email,omitempty"`
	IntFacebookAccessToken string  `json:"int_facebook_access_token,omitempty"`
	IntFacebookAgentUUID   *string `json:"int_facebook_agent_uuid,omitempty"`
	IntFacebookUserID      string  `json:"int_facebook_user_id,omitempty"`
	IntFacebookAdAccount   string  `json:"int_facebook_ad_account,omitempty"`
	IntFacebookTokenExpiry int64   `json:"int_facebook_token_expiry"`
	// Archival related fields.
	ArchiveEnabled  *bool `gorm:"default:false" json:"archive_enabled"`
	BigqueryEnabled *bool `gorm:"default:false" json:"bigquery_enabled"`
	//Salesforce settings
	IntSalesforceEnabledAgentUUID *string `json:"int_salesforce_enabled_agent_uuid,omitempty"`
	//Linkedin related fields
	IntLinkedinAdAccount          string          `json:"int_linkedin_ad_account"`
	IntLinkedinAccessToken        string          `json:"int_linkedin_access_token"`
	IntLinkedinRefreshToken       string          `json:"int_linkedin_refresh_token"`
	IntLinkedinRefreshTokenExpiry int64           `json:"int_linkedin_refresh_token_expiry"`
	IntLinkedinAccessTokenExpiry  int64           `json:"int_linkedin_access_token_expiry"`
	IntLinkedinAgentUUID          *string         `json:"int_linkedin_agent_uuid"`
	IntDrift                      *bool           `gorm:"not null;default:false" json:"int_drift,omitempty"`
	IntGoogleIngestionTimezone    string          `json:"int_google_ingestion_timezone"`
	IntClearBit                   *bool           `gorm:"not null;default:false" json:"int_clear_bit,omitempty"`
	IntAdwordsClientManagerMap    *postgres.Jsonb `json:"int_adwords_client_manager_map"`
}

type AttributionConfig struct {
	KpisToAttribute   AttributionKpis `json:"kpis_to_attribute"`
	AttributionWindow int64           `json:"attribution_window"`
	Enabled           *bool           `json:"enabled"`
}

type AttributionKpis struct {
	UserKpi []*postgres.Jsonb `json:"user_kpi"`
	HsKpi   []*postgres.Jsonb `json:"hs_kpi"`
	SfKpi   []*postgres.Jsonb `json:"sf_kpi"`
}

const ProjectSettingKeyToken = "token"
const ProjectSettingKeyPrivateToken = "private_token"

var projectSettingKeys = [...]string{
	ProjectSettingKeyToken,
	ProjectSettingKeyPrivateToken,
}

type AdwordsProjectSettings struct {
	ProjectId                  uint64
	CustomerAccountId          string
	AgentUUID                  string
	RefreshToken               string
	IntGoogleIngestionTimezone string
	IntAdwordsClientManagerMap *postgres.Jsonb
}

type GoogleOrganicProjectSettings struct {
	ProjectID    uint64
	URLPrefix    string
	AgentUUID    string
	RefreshToken string
}
type HubspotProjectSettings struct {
	ProjectId         uint64          `json:"-"`
	APIKey            string          `json:"api_key"`
	IsFirstTimeSynced bool            `json:"is_first_time_synced"`
	SyncInfo          *postgres.Jsonb `json:"sync_info"`
}

type FacebookProjectSettings struct {
	ProjectId              uint64 `json:"project_id"`
	Timezone               string `json:"timezone" gorm:"column:int_google_ingestion_timezone"`
	IntFacebookUserId      string `json:"int_facebook_user_id"`
	IntFacebookAccessToken string `json:"int_facebook_access_token"`
	IntFacebookAdAccount   string `json:"int_facebook_ad_account"`
	IntFacebookEmail       string `json:"int_facebook_email"`
	IntFacebookTokenExpiry int64  `json:"int_facebook_token_expiry"`
}

type LinkedinProjectSettings struct {
	ProjectId                     string `json:"project_id"`
	IntLinkedinAdAccount          string `json:"int_linkedin_ad_account"`
	IntLinkedinRefreshToken       string `json:"int_linkedin_refresh_token"`
	IntLinkedinRefreshTokenExpiry int64  `json:"int_linkedin_refresh_token_expiry"`
	IntLinkedinAccessToken        string `json:"int_linkedin_access_token"`
}

// SalesforceProjectSettings contains refresh_token and instance_url for enabled projects
type SalesforceProjectSettings struct {
	ProjectID    uint64 `json:"-"`
	RefreshToken string `json:"refresh_token"`
	InstanceURL  string `json:"instance_url"`
}

// Identification types
const (
	IdentificationTypePhone = "phone"
	IdentificationTypeEmail = "email"
)

var standardOrderedIdentifiers = []string{IdentificationTypeEmail, IdentificationTypePhone}

var overWriteIndentificationOrderByProjectID = map[uint64][]string{
	483: {IdentificationTypePhone, IdentificationTypeEmail},
}

// GetIdentifierPrecendenceOrderByProjectID return ordered identifier type by projectID
func GetIdentifierPrecendenceOrderByProjectID(projectID uint64) []string {
	if orderedIndentifiers, exists := overWriteIndentificationOrderByProjectID[projectID]; exists {
		return orderedIndentifiers
	}

	return standardOrderedIdentifiers
}
