package model

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type LinkedinExclusion struct {
	ProjectID             int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	AdAccountID           string          `gorm:"primary_key:true;auto_increment:false" json:"ad_account_id"`
	OrgID                 string          `gorm:"primary_key:true;auto_increment:false" json:"org_id"`
	Timestamp             int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"` //format yyyymmdd
	CampaignID            string          `gorm:"primary_key:true;auto_increment:false" json:"campaign_id"`
	CompanyName           string          `json:"company_name"`
	CampaignName          string          `json:"campaign_name"`
	IsPushedToLinkedin    bool            `json:"is_pushed_to_linkedin"`
	IsRemovedFromLinkedin bool            `json:"is_removed_from_linkedin"`
	RuleID                string          `json:"rule_id"`
	RuleSnapshot          *postgres.Jsonb `json:"rule_snapshot"`
	PropertiesSnapshot    *postgres.Jsonb `json:"properties_snapshot"`
	ImpressionsSaved      int64           `json:"impressions_saved"`
	ClicksSaved           int64           `json:"clicks_saved"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type CampaignNameTargetingCriteria struct {
	CampaignName      string
	CampaignGroupID   string
	AdAccountID       string
	TargetingCriteria map[string]bool
}

var sampleRule, _ = json.Marshal(SampleCappingRule)
var SampleExclusion = LinkedinExclusion{
	ProjectID:          2,
	OrgID:              "123",
	Timestamp:          202405,
	CampaignID:         "789",
	CompanyName:        "some company",
	CampaignName:       "some campaign",
	RuleID:             "1234abcd",
	RuleSnapshot:       &postgres.Jsonb{sampleRule},
	PropertiesSnapshot: &postgres.Jsonb{RawMessage: json.RawMessage(`{"property1": 100}`)},
	ImpressionsSaved:   1000,
	ClicksSaved:        100,
	CreatedAt:          time.Time{},
	UpdatedAt:          time.Time{},
}

type ExclusionDashboardMetric struct {
	Name       string `json:"name"`
	Value      int64  `json:"value"`
	MetricType string `json:"metric_type"`
}

var SampleExclusionDashboard = [4]ExclusionDashboardMetric{
	{
		Name:       "metric1",
		Value:      100,
		MetricType: "numerical",
	},
	{
		Name:       "metric2",
		Value:      100,
		MetricType: "numerical",
	},
	{
		Name:       "metric3",
		Value:      100,
		MetricType: "numerical",
	},
	{
		Name:       "metric4",
		Value:      100,
		MetricType: "percentage",
	},
}
