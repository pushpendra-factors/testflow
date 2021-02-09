package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type AdwordsDocument struct {
	ProjectID         uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_acc_id"`
	TypeAlias         string          `gorm:"-" json:"type_alias"`
	Type              int             `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Timestamp         int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	CampaignID        int64           `json:"-"`
	AdGroupID         int64           `json:"-"`
	AdID              int64           `json:"-"`
	KeywordID         int64           `json:"-"`
	Value             *postgres.Jsonb `json:"value"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type CampaignInfo struct {
	AdgroupName  string
	CampaignName string
	AdID         string
}

type AdwordsLastSyncInfo struct {
	ProjectId         uint64 `json:"project_id"`
	CustomerAccountId string `json:"customer_acc_id"`
	RefreshToken      string `json:"refresh_token"`
	DocumentType      int    `json:"-"`
	DocumentTypeAlias string `json:"doc_type_alias"`
	LastTimestamp     int64  `json:"last_timestamp"`
}

const (
	CampaignPerformanceReport = "campaign_performance_report"
	AdGroupPerformanceReport  = "ad_group_performance_report"
	AdPerformanceReport       = "ad_performance_report"
	KeywordPerformanceReport  = "keyword_performance_report"
	AdwordsCampaign           = "campaign"
	AdwordsAdGroup            = "ad_group"
	AdwordsAd                 = "ad"
	AdwordsKeyword            = "keyword"
	AdwordsStringColumn       = "adwords"
)

// AdwordsDocumentTypeAlias ...
var AdwordsDocumentTypeAlias = map[string]int{
	"campaigns":                   1,
	"ads":                         2,
	"ad_groups":                   3,
	"click_performance_report":    4,
	CampaignPerformanceReport:     5,
	AdPerformanceReport:           6,
	"search_performance_report":   7,
	KeywordPerformanceReport:      8,
	"customer_account_properties": 9,
	AdGroupPerformanceReport:      10,
}
