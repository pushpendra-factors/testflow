package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type FacebookDocument struct {
	ProjectID           uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAdAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_ad_account_id"`
	Platform            string          `gorm:"primary_key:true;auto_increment:false" json:"platform"`
	TypeAlias           string          `gorm:"-" json:"type_alias"`
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Timestamp           int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                  string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	CampaignID          string          `json:"-"`
	AdSetID             string          `json:"-"`
	AdID                string          `json:"-"`
	Value               *postgres.Jsonb `json:"value"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// FacebookLastSyncInfo ...
type FacebookLastSyncInfo struct {
	ProjectID           uint64 `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_acc_id"`
	Platform            string `json:"platform"`
	DocumentType        int    `json:"-"`
	DocumentTypeAlias   string `json:"type_alias"`
	LastTimestamp       int64  `json:"last_timestamp"`
}

// FacebookLastSyncInfoPayload ...
type FacebookLastSyncInfoPayload struct {
	ProjectId           string `json:"project_id"`
	CustomerAdAccountId string `json:"account_id"`
}

var ObjectToValueInFacebookJobsMapping = map[string]string{
	"campaign:name": "campaign_name",
	"ad_set:name":   "adset_name",
	"campaign:id":   "campaign_id",
	"ad_set:id":     "ad_set_id",
	"ad:id":         "ad_id",
}
var ObjectAndKeyInFacebookToPropertyMapping = map[string]string{
	"campaign:name": "campaign_name",
	"ad_group:name": "adset_name",
}

var FacebookExternalRepresentationToInternalRepresentation = map[string]string{
	"name":        "name",
	"id":          "id",
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "spend",
	"conversion":  "conversions",
	"campaign":    "campaign",
	"ad_group":    "ad_set",
	"ad":          "ad",
}

var FacebookInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions":   "impressions",
	"clicks":        "clicks",
	"spend":         "spend",
	"conversions":   "conversion",
	"campaign:name": "campaign_name",
	"ad_set:name":   "ad_group_name",
	"campaign:id":   "campaign_id",
	"ad_set:id":     "ad_group_id",
	"ad:id":         "ad_id",
}
var FacebookObjectMapForSmartProperty = map[string]string{
	"campaign": "campaign",
	"ad_set":   "ad_group",
}

const (
	FacebookSpecificError = "Failed in facebook with the following error."
)
