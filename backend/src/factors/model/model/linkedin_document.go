package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// LinkedinDocument ...
type LinkedinDocument struct {
	ProjectID           uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAdAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_ad_account_id"`
	TypeAlias           string          `gorm:"-" json:"type_alias"`
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Timestamp           int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                  string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	CampaignID          string          `json:"campaign_id"`
	CampaignGroupID     string          `json:"campaign_group_id"`
	CreativeID          string          `json:"creative_id"`
	Value               *postgres.Jsonb `json:"value"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type LinkedinLastSyncInfoPayload struct {
	ProjectID           string `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
}
type LinkedinLastSyncInfo struct {
	ProjectID           uint64 `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_account_id"`
	DocumentType        int    `json:"document_type"`
	DocumentTypeAlias   string `json:"type_alias"`
	LastTimestamp       int64  `json:"last_timestamp"`
}

const (
	LinkedinCampaignGroup = "campaign_group"
	LinkedinCampaign      = "campaign"
	LinkedinCreative      = "creative"
	LinkedinStringColumn  = "linkedin"
)

var ObjectsForLinkedin = []string{AdwordsCampaign, AdwordsAdGroup}

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
	"name":        "name",
	"id":          "id",
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "spend",
	"conversion":  "conversionValueInLocalCurrency",
	"campaign":    "campaign_group",
	"ad_group":    "campaign",
	"ad":          "creative",
	"channel":     "channel",
}

var LinkedinInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions":         "impressions",
	"clicks":              "clicks",
	"spend":               "spend",
	"conversions":         "conversion",
	"campaign_group:name": "campaign_name",
	"campaign:name":       "ad_group_name",
	"campaign_group:id":   "campaign_id",
	"campaign:id":         "ad_group_id",
	"creative:id":         "ad_id",
	"channel:name":        "channel_name",
}
var LinkedinInternalGroupByRepresentation = map[string]string{
	"impressions":         "impressions",
	"clicks":              "clicks",
	"spend":               "spend",
	"conversions":         "conversion",
	"campaign_group:name": "campaign_name",
	"campaign:name":       "ad_group_name",
	"campaign_group:id":   "campaign_group_id",
	"campaign:id":         "campaign_id",
	"creative:id":         "creative_id",
	"channel:name":        "channel_name",
}
var LinkedinObjectMapForSmartProperty = map[string]string{
	"campaign_group": "campaign",
	"campaign":       "ad_group",
}

const (
	LinkedinSpecificError = "Failed in linkedin with the error."
)
