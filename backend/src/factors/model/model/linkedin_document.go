package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// LinkedinDocument ...
type LinkedinDocument struct {
	ProjectID           int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAdAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_ad_account_id"`
	TypeAlias           string          `gorm:"-" json:"type_alias"`
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Timestamp           int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                  string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	CampaignID          string          `json:"campaign_id"`
	CampaignGroupID     string          `json:"campaign_group_id"`
	CreativeID          string          `json:"creative_id"`
	Value               *postgres.Jsonb `json:"value"`
	IsBackfilled        bool            `json:"is_backfilled"`
	SyncStatus          int             `gorm:"default:0" json:"sync_status"`
	IsGroupUserCreated  bool            `json:"is_group_user_created"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type LinkedinLastSyncInfoPayload struct {
	ProjectID           string `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
}
type LinkedinLastSyncInfo struct {
	ProjectID             int64  `json:"project_id"`
	CustomerAdAccountID   string `json:"customer_ad_account_id"`
	DocumentType          int    `json:"document_type"`
	DocumentTypeAlias     string `json:"type_alias"`
	LastTimestamp         int64  `json:"last_timestamp"`
	LastBackfillTimestamp int64  `json:"last_backfill_timestamp"`
	SyncType              int    `json:"sync_type"`
}

type LinkedinDeleteDocumentsPayload struct {
	ProjectID           int64  `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
	Timestamp           int64  `json:"timestamp"`
	TypeAlias           string `json:"type_alias"`
}

type LinkedinCampaignGroupInfoRequestPayload struct {
	ProjectID           int64  `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
	StartTimestamp      string `json:"start_timestamp"`
	EndTimestamp        string `json:"end_timestamp"`
}

type LinkedinValidationRequestPayload struct {
	ProjectID           int64  `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
	StartTimestamp      string `json:"start_timestamp"`
	EndTimestamp        string `json:"end_timestamp"`
	SyncStatus          int    `json:"sync_status"`
}
type DomainDataResponse struct {
	ProjectID           int64  `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
	Timestamp           string `json:"timestamp"`
	ID                  string `json:"id"`
	VanityName          string `json:"vanity_name"`
	LocalizedName       string `json:"localized_name"`
	Domain              string `json:"domain"`
	HeadQuarters        string `json:"headquarters"`
	PreferredCountry    string `json:"preferred_country"`
	CampaignGroupID     string `json:"campaign_group_id"`
	CampaignGroupName   string `json:"campaign_group_name"`
	Impressions         int64  `json:"impressions"`
	Clicks              int64  `json:"clicks"`
	CampaignID          string `json:"campaign_id"`
	CampaignName        string `json:"campaign_name"`
	RawDomain           string `json:"raw_domain"`
	OrgID               string `json:"org_id"`
}
type ValueForEventLookupMap struct {
	EventID       string `json:"event_id"`
	UserID        string `json:"user_id"`
	PropertyValue int64  `json:"p_value"`
}

type LinkedinCappingDataSet struct {
	OrgID           string `json:"org_id"`
	CompanyDomain   string `json:"company_domain"`
	CompanyName     string `json:"company_name"`
	CampaignGroupID string `json:"campaign_group_id"`
	CampaignID      string `json:"campaign_id"`
	CampaignName    string `json:"campaign_name"`
	Impressions     int64  `json:"impressions"`
	Clicks          int64  `json:"clicks"`
}

const (
	LinkedinCampaignGroup         = "campaign_group"
	LinkedinCampaign              = "campaign"
	LinkedinCreative              = "creative"
	LinkedinStringColumn          = "linkedin"
	LinkedinCompany               = "company"
	LinkedInMemberCompany         = "member_company"
	LinkedInMemberCompanyInsights = "member_company_insights"
)

var ObjectsForLinkedin = []string{AdwordsCampaign, AdwordsAdGroup}
var ObjectsForLinkedinCompany = []string{AdwordsCampaign, AdwordsAdGroup, LinkedinCompany}
var ObjectToDisplayCategoryForLinkedin = map[string]string{
	AdwordsCampaign: "Campaign Group",
	AdwordsAdGroup:  "Campaign",
	LinkedinCompany: "Company",
}

var CompanySyncJobTypeMap = map[string]int{
	"daily": 0,
	"t8":    1,
	"t22":   2,
}

var ObjectToValueInLinkedinJobsMapping = map[string]string{
	"campaign_group:name": "campaign_group_name",
	"campaign:name":       "campaign_group_name",
	"ad_group:name":       "campaign_name",
	"campaign_group:id":   "campaign_group_id",
	"campaign:id":         "campaign_id",
	"creative:id":         "creative_id",
}
var ObjectAndKeyInLinkedinToPropertyMapping = map[string]string{
	"campaign:name": "campaign_group_name",
	"ad_group:name": "campaign_name",
}
var LinkedinExternalRepresentationToInternalRepresentation = map[string]string{
	"name":              "name",
	"id":                "id",
	"impressions":       "impressions",
	"clicks":            "clicks",
	"spend":             "spend",
	"conversion":        "conversionValueInLocalCurrency",
	"campaign":          "campaign_group",
	"ad_group":          "campaign",
	"ad":                "creative",
	"channel":           "channel",
	"company":           "member_company_insights",
	"vanity_name":       "vanityName",
	"localized_name":    "localizedName",
	"headquarters":      "companyHeadquarters",
	"domain":            "localizedWebsite",
	"preferred_country": "preferredCountry",
}

var LinkedinInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions":                               "impressions",
	"clicks":                                    "clicks",
	"spend":                                     "spend",
	"conversions":                               "conversion",
	"campaign_group:name":                       "campaign_name",
	"campaign:name":                             "ad_group_name",
	"campaign_group:id":                         "campaign_id",
	"campaign:id":                               "ad_group_id",
	"creative:id":                               "ad_id",
	"channel:name":                              "channel_name",
	"member_company_insights:vanity_name":       "company_vanity_name",
	"member_company_insights:localized_name":    "company_localized_name",
	"member_company_insights:headquarters":      "company_headquarters",
	"member_company_insights:domain":            "company_domain",
	"member_company_insights:preferred_country": "company_preferred_country",
}
var LinkedinInternalGroupByRepresentation = map[string]string{
	"impressions":                               "impressions",
	"clicks":                                    "clicks",
	"spend":                                     "spend",
	"conversions":                               "conversion",
	"campaign_group:name":                       "campaign_name",
	"campaign:name":                             "ad_group_name",
	"campaign_group:id":                         "campaign_group_id",
	"campaign:id":                               "campaign_id",
	"creative:id":                               "creative_id",
	"channel:name":                              "channel_name",
	"member_company_insights:vanity_name":       "company_vanity_name",
	"member_company_insights:localized_name":    "company_localized_name",
	"member_company_insights:headquarters":      "company_headquarters",
	"member_company_insights:domain":            "company_domain",
	"member_company_insights:preferred_country": "company_preferred_country",
}
var LinkedinObjectMapForSmartProperty = map[string]string{
	"campaign_group": "campaign",
	"campaign":       "ad_group",
}

const (
	LinkedinSpecificError = "Failed in linkedin with the error."
)
