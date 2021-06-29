package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type AdwordsDocument struct {
	ProjectID         uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_acc_id"`
	TypeAlias         string          `gorm:"-" json:"type_alias"`
	Type              int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
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

type AdwordsLastSyncInfo struct {
	ProjectId         uint64 `json:"project_id"`
	Timezone          string `json:"timezone"`
	CustomerAccountId string `json:"customer_acc_id"`
	RefreshToken      string `json:"refresh_token"`
	DocumentType      int    `json:"-"`
	DocumentTypeAlias string `json:"doc_type_alias"`
	LastTimestamp     int64  `json:"last_timestamp"`
}

type AdwordsLastSyncInfoPayload struct {
	ProjectId uint64 `json:"project_id"`
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
	AdwordsSpecificError      = "Failed in adwords with following error."
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

const (
	ApprovalStatus         = "approval_status"
	MatchType              = "match_type"
	FirstPositionCpc       = "first_position_cpc"
	FirstPageCpc           = "first_page_cpc"
	IsNegative             = "is_negative"
	TopOfPageCpc           = "top_of_page_cpc"
	QualityScore           = "quality_score"
	AdvertisingChannelType = "advertising_channel_type"
	Impressions            = "impressions"

	Clicks                                     = "clicks"
	ClickThroughRate                           = "click_through_rate"
	Conversion                                 = "conversion"
	ConversionRate                             = "conversion_rate"
	CostPerClick                               = "cost_per_click"
	CostPerConversion                          = "cost_per_conversion"
	SearchImpressionShare                      = "search_impression_share"
	SearchClickShare                           = "search_click_share"
	SearchTopImpressionShare                   = "search_top_impression_share"
	SearchAbsoluteTopImpressionShare           = "search_absolute_top_impression_share"
	SearchBudgetLostAbsoluteTopImpressionShare = "search_budget_lost_absolute_top_impression_share"
	SearchBudgetLostImpressionShare            = "search_budget_lost_impression_share"
	SearchBudgetLostTopImpressionShare         = "search_budget_lost_top_impression_share"
	SearchRankLostAbsoluteTopImpressionShare   = "search_rank_lost_absolute_top_impression_share"
	SearchRankLostImpressionShare              = "search_rank_lost_impression_share"
	SearchRankLostTopImpressionShare           = "search_rank_lost_top_impression_share"
	TotalSearchImpression                      = "total_search_impression"
	TotalSearchClick                           = "total_search_click"
	TotalSearchTopImpression                   = "total_search_top_impression"
	TotalSearchAbsoluteTopImpression           = "total_search_absolute_top_impression"
	TotalSearchBudgetLostAbsoluteTopImpression = "total_search_budget_lost_absolute_top_impression"
	TotalSearchBudgetLostImpression            = "total_search_budget_lost_impression"
	TotalSearchBudgetLostTopImpression         = "total_search_budget_lost_top_impression"
	TotalSearchRankLostAbsoluteTopImpression   = "total_search_rank_lost_absolute_top_impression"
	TotalSearchRankLostImpression              = "total_search_rank_lost_impression"
	TotalSearchRankLostTopImpression           = "total_search_rank_lost_top_impression"
	AdwordsSmartProperty                       = "smart_properties"
	SearchPerformanceReport                    = "search_performance_report"
)

/*
	Map from request Params to Internal Representation is needed, so that validation of params and operating within adwords context becomes easy.
	Map from Internal Representation to Representation within Report/Job as field values can vary.
	Map from Internal Representation to External Representation is needed to expose right column names and also to perform clear operations like union or join.
		We can follow the same representation of external even during cte formation, though used in internal context.
	We might all above complicated transformations in api if we merge all document types i.e.facebook, linkedin etc...
*/
var AdwordsExtToInternal = map[string]string{
	"campaign":                       "campaign",
	"ad_group":                       "ad_group",
	"ad":                             "ad",
	"name":                           "name",
	"keyword":                        "keyword",
	"id":                             "id",
	"status":                         "status",
	ApprovalStatus:                   ApprovalStatus,
	MatchType:                        MatchType,
	FirstPositionCpc:                 FirstPositionCpc,
	FirstPageCpc:                     FirstPageCpc,
	IsNegative:                       IsNegative,
	TopOfPageCpc:                     TopOfPageCpc,
	QualityScore:                     QualityScore,
	AdvertisingChannelType:           AdvertisingChannelType,
	Impressions:                      Impressions,
	Clicks:                           Clicks,
	"spend":                          "cost",
	Conversion:                       "conversions",
	ClickThroughRate:                 ClickThroughRate,
	ConversionRate:                   ConversionRate,
	CostPerClick:                     CostPerClick,
	CostPerConversion:                CostPerConversion,
	SearchImpressionShare:            SearchImpressionShare,
	SearchClickShare:                 SearchClickShare,
	SearchTopImpressionShare:         SearchTopImpressionShare,
	SearchAbsoluteTopImpressionShare: SearchAbsoluteTopImpressionShare,
	SearchBudgetLostAbsoluteTopImpressionShare: SearchBudgetLostAbsoluteTopImpressionShare,
	SearchBudgetLostImpressionShare:            SearchBudgetLostImpressionShare,
	SearchBudgetLostTopImpressionShare:         SearchBudgetLostTopImpressionShare,
	SearchRankLostAbsoluteTopImpressionShare:   SearchRankLostAbsoluteTopImpressionShare,
	SearchRankLostImpressionShare:              SearchRankLostImpressionShare,
	SearchRankLostTopImpressionShare:           SearchRankLostTopImpressionShare,
}

var AdwordsInternalPropertiesToJobsInternal = map[string]string{
	"campaign:id":                       "id",
	"campaign:name":                     "name",
	"campaign:status":                   "status",
	"campaign:advertising_channel_type": AdvertisingChannelType,
	"ad_group:id":                       "id",
	"ad_group:name":                     "name",
	"ad_group:status":                   "status",
	"ad:id":                             "ad_id",
	"keyword:id":                        "id",
	"keyword:name":                      "criteria",
	"keyword:status":                    "status",
	"keyword:approval_status":           ApprovalStatus,
	"keyword:match_type":                "keyword_match_type",
	"keyword:first_position_cpc":        FirstPositionCpc,
	"keyword:first_page_cpc":            FirstPageCpc,
	"keyword:is_negative":               IsNegative,
	"keyword:top_of_page_cpc":           TopOfPageCpc,
	"keyword:quality_score":             QualityScore,
}

var AdwordsInternalPropertiesToReportsInternal = map[string]string{
	"campaign:id":                       "campaign_id",
	"campaign:name":                     "campaign_name",
	"campaign:status":                   "campaign_status",
	"campaign:advertising_channel_type": AdvertisingChannelType,
	"ad_group:id":                       "ad_group_id",
	"ad_group:name":                     "ad_group_name",
	"ad_group:status":                   "ad_group_status",
	"ad:id":                             "ad_id",
	"keyword:id":                        "keyword_id",
	"keyword:name":                      "criteria",
	"keyword:status":                    "status",
	"keyword:approval_status":           ApprovalStatus,
	"keyword:match_type":                "keyword_match_type",
	"keyword:first_position_cpc":        FirstPositionCpc,
	"keyword:first_page_cpc":            FirstPageCpc,
	"keyword:is_negative":               IsNegative,
	"keyword:top_of_page_cpc":           TopOfPageCpc,
	"keyword:quality_score":             QualityScore,
}
