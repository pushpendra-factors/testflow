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
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"-"`
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
	LinkedinSpecificError = "Failed in linkedin with the error."
)
