package model

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

var ISOCODES = map[string]int{"AF": 1, "AX": 1, "AL": 1, "DZ": 1, "AS": 1, "AD": 1, "AO": 1, "AI": 1, "AQ": 1, "AG": 1, "AR": 1, "AM": 1, "AW": 1, "AU": 1, "AT": 1, "AZ": 1, "BS": 1, "BH": 1, "BD": 1, "BB": 1, "BY": 1, "BE": 1, "BZ": 1, "BJ": 1, "BM": 1, "BT": 1, "BO": 1, "BQ": 1, "BA": 1, "BW": 1, "BV": 1, "BR": 1, "IO": 1, "BN": 1, "BG": 1, "BF": 1, "BI": 1, "KH": 1, "CM": 1, "CA": 1, "CV": 1, "KY": 1, "CF": 1, "TD": 1, "CL": 1, "CN": 1, "CX": 1, "CC": 1, "CO": 1, "KM": 1, "CG": 1, "CD": 1, "CK": 1, "CR": 1, "CI": 1, "HR": 1, "CU": 1, "CW": 1, "CY": 1, "CZ": 1, "DK": 1, "DJ": 1, "DM": 1, "DO": 1, "EC": 1, "EG": 1, "SV": 1, "GQ": 1, "ER": 1, "EE": 1, "ET": 1, "FK": 1, "FO": 1, "FJ": 1, "FI": 1, "FR": 1, "GF": 1, "PF": 1, "TF": 1, "GA": 1, "GM": 1, "GE": 1, "DE": 1, "GH": 1, "GI": 1, "GR": 1, "GL": 1, "GD": 1, "GP": 1, "GU": 1, "GT": 1, "GG": 1, "GN": 1, "GW": 1, "GY": 1, "HT": 1, "HM": 1, "VA": 1, "HN": 1, "HK": 1, "HU": 1, "IS": 1, "IN": 1, "ID": 1, "IR": 1, "IQ": 1, "IE": 1, "IM": 1, "IL": 1, "IT": 1, "JM": 1, "JP": 1, "JE": 1, "JO": 1, "KZ": 1, "KE": 1, "KI": 1, "KP": 1, "KR": 1, "KW": 1, "KG": 1, "LA": 1, "LV": 1, "LB": 1, "LS": 1, "LR": 1, "LY": 1, "LI": 1, "LT": 1, "LU": 1, "MO": 1, "MK": 1, "MG": 1, "MW": 1, "MY": 1, "MV": 1, "ML": 1, "MT": 1, "MH": 1, "MQ": 1, "MR": 1, "MU": 1, "YT": 1, "MX": 1, "FM": 1, "MD": 1, "MC": 1, "MN": 1, "ME": 1, "MS": 1, "MA": 1, "MZ": 1, "MM": 1, "NA": 1, "NR": 1, "NP": 1, "NL": 1, "NC": 1, "NZ": 1, "NI": 1, "NE": 1, "NG": 1, "NU": 1, "NF": 1, "MP": 1, "NO": 1, "OM": 1, "PK": 1, "PW": 1, "PS": 1, "PA": 1, "PG": 1, "PY": 1, "PE": 1, "PH": 1, "PN": 1, "PL": 1, "PT": 1, "PR": 1, "QA": 1, "RE": 1, "RO": 1, "RU": 1, "RW": 1, "BL": 1, "SH": 1, "KN": 1, "LC": 1, "MF": 1, "PM": 1, "VC": 1, "WS": 1, "SM": 1, "ST": 1, "SA": 1, "SN": 1, "RS": 1, "SC": 1, "SL": 1, "SG": 1, "SX": 1, "SK": 1, "SI": 1, "SB": 1, "SO": 1, "ZA": 1, "GS": 1, "SS": 1, "ES": 1, "LK": 1, "SD": 1, "SR": 1, "SJ": 1, "SZ": 1, "SE": 1, "CH": 1, "SY": 1, "TW": 1, "TJ": 1, "TZ": 1, "TH": 1, "TL": 1, "TG": 1, "TK": 1, "TO": 1, "TT": 1, "TN": 1, "TR": 1, "TM": 1, "TC": 1, "TV": 1, "UG": 1, "UA": 1, "AE": 1, "GB": 1, "US": 1, "UM": 1, "UY": 1, "UZ": 1, "VU": 1, "VE": 1, "VN": 1, "VG": 1, "VI": 1, "WF": 1, "EH": 1, "YE": 1, "ZM": 1, "ZW": 1, "XK": 1}
var countryMap = map[string]string{"afghanistan": "AF", "albania": "AL", "algeria": "DZ", "american samoa": "AS", "andorra": "AD", "angola": "AO", "anguilla": "AI", "antarctica": "AQ", "antigua and barbuda": "AG", "argentina": "AR", "armenia": "AM", "aruba": "AW", "australia": "AU", "austria": "AT", "azerbaijan": "AZ", "bahamas": "BS", "bahrain": "BH", "bangladesh": "BD", "barbados": "BB", "belarus": "BY", "belgium": "BE", "belize": "BZ", "benin": "BJ", "bermuda": "BM", "bhutan": "BT", "bolivarian republic of venezuela": "VE", "bolivia": "BO", "bolivia (plurinational state of)": "BO", "bonaire, sint eustatius and saba": "BQ", "bonaire, sint eustatius, and saba": "BQ", "bosnia and herzegovina": "BA", "botswana": "BW", "bouvet island": "BV", "brazil": "BR", "british indian ocean territory": "IO", "british virgin islands": "IO", "brunei": "BN", "brunei darussalam": "BN", "bulgaria": "BG", "burkina faso": "BF", "burundi": "BI", "cabo verde": "CV", "cambodia": "KH", "cameroon": "CM", "canada": "CA", "cayman islands": "KY", "central african republic": "CF", "chad": "TD", "chile": "CL", "china": "CN", "christmas island": "CX", "cocos (keeling) islands": "CC", "colombia": "CO", "comoros": "KM", "congo": "CG", "congo republic": "CD", "congo, democratic republic of the congo": "CD", "cook islands": "CK", "costa rica": "CR", "croatia": "HR", "cuba": "CU", "curaao": "CW", "curaçao": "CW", "cyprus": "CY", "czechia": "CZ", "côte d'ivoire": "CI", "democratic republic of timor-leste": "TL", "denmark": "DK", "djibouti": "DJ", "dominica": "DM", "dominican republic": "DO", "dr congo": "CD", "east timor": "TL", "ecuador": "EC", "egypt": "EG", "el salvador": "SV", "equatorial guinea": "GQ", "eritrea": "ER", "estonia": "EE", "eswatini": "SZ", "ethiopia": "ET", "falkland islands": "FK", "falkland islands (malvinas)": "FK", "faroe islands": "FO", "federated states of micronesia": "FM", "fiji": "FJ", "finland": "FI", "france": "FR", "french guiana": "GF", "french polynesia": "PF", "french southern territories": "TF", "gabon": "GA", "gambia": "GM", "georgia": "GE", "germany": "DE", "ghana": "GH", "gibraltar": "GI", "greece": "GR", "greenland": "GL", "grenada": "GD", "guadeloupe": "GP", "guam": "GU", "guatemala": "GT", "guernsey": "GG", "guinea": "GN", "guinea-bissau": "GW", "guyana": "GY", "haiti": "HT", "hashemite kingdom of jordan": "JO", "heard island and mcdonald islands": "HM", "holy see": "VA", "honduras": "HN", "hong kong": "HK", "hungary": "HU", "iceland": "IS", "india": "IN", "indonesia": "ID", "iran": "IR", "iran (islamic republic of)": "IR", "iraq": "IQ", "ireland": "IE", "isle of man": "IM", "israel": "IL", "italy": "IT", "ivory coast": "CI", "jamaica": "JM", "japan": "JP", "jersey": "JE", "jordan": "JO", "kazakhstan": "KZ", "kenya": "KE", "kiribati": "KI", "korea (democratic people's republic of)": "KP", "korea, republic of": "KR", "kuwait": "KW", "kyrgyzstan": "KG", "lao people's democratic republic": "LA", "laos": "LA", "latvia": "LV", "lebanon": "LB", "lesotho": "LS", "liberia": "LR", "libya": "LY", "liechtenstein": "LI", "lithuania": "LT", "luxembourg": "LU", "macao": "MO", "macedonia": "MK", "madagascar": "MG", "malawi": "MW", "malaysia": "MY", "maldives": "MV", "mali": "ML", "malta": "MT", "marshall islands": "MH", "martinique": "MQ", "mauritania": "MR", "mauritius": "MU", "mayotte": "YT", "mexico": "MX", "micronesia (federated states of)": "FM", "moldova": "MD", "moldova, republic of": "MD", "monaco": "MC", "mongolia": "MN", "montenegro": "ME", "montserrat": "MS", "morocco": "MA", "mozambique": "MZ", "myanmar": "MM", "namibia": "NA", "nauru": "NR", "nepal": "NP", "netherlands": "NL", "new caledonia": "NC", "new zealand": "NZ", "nicaragua": "NI", "niger": "NE", "nigeria": "NG", "niue": "NU", "norfolk island": "NF", "north macedonia": "MK", "northern mariana islands": "MP", "norway": "NO", "oman": "OM", "pakistan": "PK", "palau": "PW", "palestine": "PS", "palestine, state of": "PS", "panama": "PA", "papua new guinea": "PG", "paraguay": "PY", "peru": "PE", "philippines": "PH", "pitcairn": "PN", "poland": "PL", "portugal": "PT", "principality of monaco": "MC", "puerto rico": "PR", "qatar": "QA", "republic of korea": "KR", "republic of lithuania": "LT", "republic of moldova": "MD", "republic of the congo": "CD", "romania": "RO", "runion": "RE", "russia": "RU", "russian federation": "RU", "rwanda": "RW", "réunion": "RE", "saint barthlemy": "BL", "saint barthélemy": "BL", "saint helena, ascension and tristan da cunha": "SH", "saint kitts and nevis": "KN", "saint lucia": "LC", "saint martin": "SX", "saint martin (french part)": "MF", "saint pierre and miquelon": "PM", "saint vincent and the grenadines": "VC", "samoa": "WS", "san marino": "SM", "sao tome and principe": "ST", "saudi arabia": "SA", "senegal": "SN", "serbia": "RS", "seychelles": "SC", "sierra leone": "SL", "singapore": "SG", "sint maarten": "SX", "sint maarten (dutch part)": "SX", "slovakia": "SK", "slovenia": "SI", "so tom and prncipe": "ST", "solomon islands": "SB", "somalia": "SO", "south africa": "ZA", "south georgia and the south sandwich islands": "GS", "south korea": "KR", "south sudan": "SS", "spain": "ES", "sri lanka": "LK", "st kitts and nevis": "KN", "st vincent and grenadines": "VC", "sudan": "SD", "suriname": "SR", "svalbard and jan mayen": "SJ", "sweden": "SE", "switzerland": "CH", "syria": "SY", "syrian arab republic": "SY", "são tomé and príncipe": "ST", "taiwan": "TW", "taiwan, province of china": "TW", "tajikistan": "TJ", "tanzania": "TZ", "tanzania, united republic of": "TZ", "thailand": "TH", "timor-leste": "TL", "togo": "TG", "tokelau": "TK", "tonga": "TO", "trinidad and tobago": "TT", "tunisia": "TN", "turkey": "TR", "turkmenistan": "TM", "turks and caicos islands": "TC", "tuvalu": "TV", "uganda": "UG", "ukraine": "UA", "united arab emirates": "AE", "united kingdom": "GB", "united states": "US", "uruguay": "UY", "us virgin islands": "VI", "uzbekistan": "UZ", "vanuatu": "VU", "vatican city": "VA", "venezuela": "VE", "viet nam": "VN", "vietnam": "VN", "virgin islands, u.s.": "VI", "wallis and futuna": "WF", "wallis and futuna islands": "WF", "western sahara": "EH", "yemen": "YE", "zambia": "ZM", "zimbabwe": "ZW", "åland islands": "AX"}

type ProjectSetting struct {
	// Foreign key constraint project_id -> projects(id)
	// Used project_id as primary key also, becase of 1-1 relationship.
	ProjectId         int64           `gorm:"primary_key:true" json:"project_id,omitempty"`
	AttributionConfig *postgres.Jsonb `json:"attribution_config"`
	SixSignalConfig   *postgres.Jsonb `json:"six_signal_config"`
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
	SegmentMarkerLastRun      time.Time       `json:"segment_marker_last_run"`
	FilterIps                 *postgres.Jsonb `json:"filter_ips,omitempty"`
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
	IntRudderstack        *bool           `gorm:"not null;default:false" json:"int_rudderstack,omitempty"`
	ProjectCurrency       string          `json:"currency"`
	IsPathAnalysisEnabled bool            `json:"is_path_analysis_enabled"`
	Acc_score_weights     *postgres.Jsonb `json:"acc_score_weights"`
	// onboarding flow steps
	IsDeanonymizationRequested bool            `json:"is_deanonymization_requested"`
	IsOnboardingCompleted      bool            `json:"is_onboarding_completed"`
	SixSignalEmailList         string          `gorm:"column:sixsignal_email_list" json:"sixsignal_email_list"`
	IntG2ApiKey                string          `json:"int_g2_api_key"`
	IntG2                      *bool           `gorm:"not null;default:false" json:"int_g2,omitempty"`
	OnboardingSteps            *postgres.Jsonb `json:"onboarding_steps"`
	SDKAPIURL                  string          `gorm:"-" json:"sdk_api_url"`
	SDKAssetURL                string          `gorm:"-" json:"sdk_asset_url"`
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

/*
{
  "api_limit": 60,
  "country_include": [
    {
      "value": "India",
      "type": "equals"
    },
    {
      "value": "USA",
      "type": "equals"
    },
    {
      "value": "Nepal",
      "type": "notEqual"
    },
    {
      "value": "SriLanka",
      "type": "notEqual"
    }
  ],
  "country_exclude": [],
  "pages_include": [
    {
      "value": "https://www.factors.ai/",
      "type": "equals"
    },
    {
      "value": "demo",
      "type": "contains"
    },
    {
      "value": "https://www.factors.ai/terms",
      "type": "notEqual"
    }
  ],
  "pages_exclude": []
}
*/

type SixSignalConfig struct {
	//APILimit       int               `json:"api_limit"`
	CountryInclude []SixSignalFilter `json:"country_include"`
	CountryExclude []SixSignalFilter `json:"country_exclude"`
	PagesInclude   []SixSignalFilter `json:"pages_include"`
	PagesExclude   []SixSignalFilter `json:"pages_exclude"`
}

type SixSignalFilter struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type TimelinesConfig struct {
	DisabledEvents []string      `json:"disabled_events"`
	UserConfig     UserConfig    `json:"user_config"`
	AccountConfig  AccountConfig `json:"account_config"`
}

type FilterIps struct {
	BlockIps []string `json:"block_ips"`
}
type UserConfig struct {
	Milestones    []string `json:"milestones"`
	TableProps    []string `json:"table_props"`
	LeftpaneProps []string `json:"leftpane_props"`
}

type AccountConfig struct {
	Milestones    []string `json:"milestones"`
	TableProps    []string `json:"table_props"`
	LeftpaneProps []string `json:"leftpane_props"`
	UserProp      string   `json:"user_prop"`
}

type AttributionConfig struct {
	KpisToAttribute                   *postgres.Jsonb `json:"kpis_to_attribute"`
	AttributionWindow                 int             `json:"attribution_window"`
	AnalyzeTypeUserKPI                bool            `json:"user_kpi"`
	AnalyzeTypeHSDealsEnabled         bool            `json:"hubspot_deals"`
	AnalyzeTypeSFOpportunitiesEnabled bool            `json:"salesforce_opportunities"`
	AnalyzeTypeHSCompaniesEnabled     bool            `json:"hubspot_companies"`
	AnalyzeTypeSFAccountsEnabled      bool            `json:"salesforce_accounts"`
	PreComputeEnabled                 bool            `json:"pre_compute_enabled"`
	QueryType                         string          `json:"query_type"`
}

/*type AttributionKpis struct {
	UserKpi *postgres.Jsonb `json:"user_kpi"`
	HsKpi   *postgres.Jsonb `json:"hs_kpi"`
	SfKpi   *postgres.Jsonb `json:"sf_kpi"`
}*/

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

type G2ProjectSettings struct {
	ProjectID   int64  `json:"project_id"`
	IntG2APIKey string `json:"int_g2_api_key"`
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

func GetProjectSDKAPIAndAssetURL(projectID int64) (string, string) {
	// Todo: The selected URL for each project has to be persisted.
	// Any change to availableHashes will change the SDK code for a project, if not persisted.
	availableHashes := []string{
		"b3mxnuvcer",
		"dyh8ken8pc",
	}

	index := (int(projectID) % len(availableHashes))
	hash := availableHashes[index]

	assetURL := "https://asset." + hash + ".com/" + hash + ".js"
	apiURL := "https://api." + hash + ".com"

	return assetURL, apiURL
}

func ReplaceCountryNameWithIsoCodeInSixSignalConfig(filter []SixSignalFilter) ([]SixSignalFilter, bool) {
	var failure bool
	if len(filter) > 0 {
		for i := range filter {
			if v, exists := countryMap[strings.ToLower(filter[i].Value)]; exists {
				filter[i].Value = v
			} else {
				failure = true
			}
		}
	}

	return filter, failure

}
