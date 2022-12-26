package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type ProjectSetting struct {
	// Foreign key constraint project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId         int64           `gorm:"primary_key:true" json:"project_id,omitempty"`
	AttributionConfig *postgres.Jsonb `json:"attribution_config"`
	TimelinesConfig   *postgres.Jsonb `json:"timelines_config"`

	// Using pointers to avoid update by default value.
	// omit empty to avoid nil(filelds not updated) on resp json.
	AutoTrack            *bool `gorm:"not null;default:false" json:"auto_track,omitempty"`
	AutoTrackSPAPageView *bool `gorm:"not null;default:false" json:"auto_track_spa_page_view"`
	AutoFormCapture      *bool `gorm:"not null;default:false" json:"auto_form_capture,omitempty"`
	AutoClickCapture     *bool `gorm:"not null;default:false" json:"auto_click_capture,omitempty"`
	AutoCaptureFormFills *bool `gorm:"not null;default:false" json:"auto_capture_form_fills"`
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
	//Cache settings preset
	CacheSettings             *postgres.Jsonb `json:"cache_settings"`
	IntHubspot                *bool           `gorm:"not null;default:false" json:"int_hubspot,omitempty"`
	IntHubspotApiKey          string          `json:"int_hubspot_api_key,omitempty"`
	IntHubspotRefreshToken    string          `json:"int_hubspot_refresh_token,omitempty"`
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
	IntFacebookIngestionTimezone  string          `json:"int_facebook_ingestion_timezone"`
	IntClearBit                   *bool           `gorm:"not null;default:false" json:"int_clear_bit,omitempty"`
	IntClientSixSignalKey         *bool           `gorm:"not null;default:false" json:"int_client_six_signal_key,omitempty"`
	IntFactorsSixSignalKey        *bool           `gorm:"not null;default:false" json:"int_factors_six_signal_key,omitempty"`
	IntAdwordsClientManagerMap    *postgres.Jsonb `json:"int_adwords_client_manager_map"`
	ClearbitKey                   string          `json:"clearbit_key"`
	Client6SignalKey              string          `json:"client6_signal_key"`
	Factors6SignalKey             string          `json:"factors6_signal_key"`
	LeadSquaredConfig             *postgres.Jsonb `json:"lead_squared_config"`
	IsWeeklyInsightsEnabled       bool            `json:"is_weekly_insights_enabled"`
	IsExplainEnabled              bool            `json:"is_explain_enabled"`
	IntegrationBits               string          `json: "-"`
	// Rudderstack integration settings.
	IntRudderstack *bool `gorm:"not null;default:false" json:"int_rudderstack,omitempty"`
}

/* Sample Attribution Setting
{
  "kpis_to_attribute": {
    "user_kpi": [
      {
        "u_kpi_1": "u_val_1"
      },
      {
        "u_kpi_2": "u_val_2"
      }
    ],
    "hs_kpi": [
      {
        "hs_kpi_1": "hs_val_1"
      },
      {
        "hs_kpi_2": "hs_val_2"
      }
    ],
    "sf_kpi": [
      {
        "sf_kpi_1": "sf_val_1"
      },
      {
        "sf_kpi_2": "sf_val_2"
      }
    ]
  },
  "attribution_window": 100,
  "user_kpi": true,
  "hubspot_deals": false,
  "salesforce_opportunities": false,
  "hubspot_companies": true,
  "salesforce_accounts": true,
  "pre_compute_enabled": false
}
*/

type TimelinesConfig struct {
	DisabledEvents []string      `json:"disabled_events"`
	UserConfig     UserConfig    `json:"user_config"`
	AccountConfig  AccountConfig `json:"account_config"`
}

type UserConfig struct {
	PropsToShow []string `json:"props_to_show"`
}

type AccountConfig struct {
	AccountPropsToShow []string `json:"account_props_to_show"`
	UserPropToShow     string   `json:"user_prop_to_show"`
}

type AttributionConfig struct {
	KpisToAttribute                   AttributionKpis `json:"kpis_to_attribute"`
	AttributionWindow                 int64           `json:"attribution_window"`
	AnalyzeTypeUserKPI                bool            `json:"user_kpi"`
	AnalyzeTypeHSDealsEnabled         bool            `json:"hubspot_deals"`
	AnalyzeTypeSFOpportunitiesEnabled bool            `json:"salesforce_opportunities"`
	AnalyzeTypeHSCompaniesEnabled     bool            `json:"hubspot_companies"`
	AnalyzeTypeSFAccountsEnabled      bool            `json:"salesforce_accounts"`
	PreComputeEnabled                 bool            `json:"pre_compute_enabled"`
}

type AttributionKpis struct {
	UserKpi *postgres.Jsonb `json:"user_kpi"`
	HsKpi   *postgres.Jsonb `json:"hs_kpi"`
	SfKpi   *postgres.Jsonb `json:"sf_kpi"`
}

type LeadSquaredConfig struct {
	Host            string `json:"host"`
	AccessKey       string `json:"access_key"`
	SecretKey       string `json:"secret_key"`
	BigqueryDataset string `json:"bigquery_dataset"`
	FirstTimeSync   bool   `json:"first_time_sync"`
}

const ProjectSettingKeyToken = "token"
const ProjectSettingKeyPrivateToken = "private_token"

var AutoClickCaptureDefault = false
var AutoCaptureFormFillsDefault = false

const DEFAULT_STRING_WITH_ZEROES_32BIT = "00000000000000000000000000000000"

var projectSettingKeys = [...]string{
	ProjectSettingKeyToken,
	ProjectSettingKeyPrivateToken,
}

type AdwordsProjectSettings struct {
	ProjectId                  int64
	CustomerAccountId          string
	AgentUUID                  string
	RefreshToken               string
	IntGoogleIngestionTimezone string
	IntAdwordsClientManagerMap *postgres.Jsonb
}

type GoogleOrganicProjectSettings struct {
	ProjectID    int64
	URLPrefix    string
	AgentUUID    string
	RefreshToken string
}
type HubspotProjectSettings struct {
	ProjectId         int64           `json:"-"`
	APIKey            string          `json:"api_key"`
	RefreshToken      string          `json:"refresh_token"`
	IsFirstTimeSynced bool            `json:"is_first_time_synced"`
	SyncInfo          *postgres.Jsonb `json:"sync_info"`
}

type FacebookProjectSettings struct {
	ProjectId              int64  `json:"project_id"`
	Timezone               string `json:"timezone" gorm:"column:int_facebook_ingestion_timezone"`
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
	ProjectID    int64  `json:"-"`
	RefreshToken string `json:"refresh_token"`
	InstanceURL  string `json:"instance_url"`
}

type CacheSettings struct {
	AttributionCachePresets map[string]bool `json:"attribution_cache_presets"`
}

// Identification types
const (
	IdentificationTypePhone = "phone"
	IdentificationTypeEmail = "email"
)

var standardOrderedIdentifiers = []string{IdentificationTypeEmail, IdentificationTypePhone}

var overWriteIndentificationOrderByProjectID = map[int64][]string{
	483: {IdentificationTypePhone, IdentificationTypeEmail},
}

// GetIdentifierPrecendenceOrderByProjectID return ordered identifier type by projectID
func GetIdentifierPrecendenceOrderByProjectID(projectID int64) []string {
	if orderedIndentifiers, exists := overWriteIndentificationOrderByProjectID[projectID]; exists {
		return orderedIndentifiers
	}

	return standardOrderedIdentifiers
}

// HubspotCustomIdentificationFieldByProjectID Hubspot projects with custom field for identification
var HubspotCustomIdentificationFieldByProjectID = map[int64]string{
	// use raw property name
	2251799836000005: "user_id",
}

// GetHubspotCustomIdentificationFieldByProjectID use to get custom field for hubspot custom identification enabled project
func GetHubspotCustomIdentificationFieldByProjectID(projectID int64) string {
	return HubspotCustomIdentificationFieldByProjectID[projectID]
}
