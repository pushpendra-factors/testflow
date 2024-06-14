package model

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type LinkedinExclusion struct {
	ID                    string          `json:"id"`
	ProjectID             int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	OrgID                 string          `gorm:"primary_key:true;auto_increment:false" json:"org_id"`
	Timestamp             int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"` //format yyyymmdd
	Campaigns             *postgres.Jsonb `gorm:"primary_key:true;auto_increment:false" json:"campaigns"`
	CampaignsCount        int             `json:"campaigns_count"`
	CompanyName           string          `json:"company_name"`
	IsPushedToLinkedin    bool            `json:"is_pushed_to_linkedin"`
	IsRemovedFromLinkedin bool            `json:"is_removed_from_linkedin"`
	RuleID                string          `json:"rule_id"`
	RuleObjectType        string          `json:"rule_object_type"`
	RuleSnapshot          *postgres.Jsonb `json:"rule_snapshot"`
	PropertiesSnapshot    *postgres.Jsonb `json:"properties_snapshot"`
	ExactSubruleMatched   *postgres.Jsonb `json:"exact_subrule_matched"`
	LinkedinData          *postgres.Jsonb `json:"linkedin_data"`
	ImpressionsSaved      int64           `json:"impressions_saved"`
	ClicksSaved           int64           `json:"clicks_saved"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}
type CampaignsIDName struct {
	CampaignID   string
	CampaignName string
}

type CampaignNameTargetingCriteria struct {
	CampaignName      string
	CampaignGroupID   string
	AdAccountID       string
	TargetingCriteria map[string]bool
}

var sampleRule, _ = json.Marshal(SampleCappingRule)
var sampleCampaigns, _ = json.Marshal(CampaignsIDName{CampaignID: "789", CampaignName: "some campaign"})
var SampleExclusion = LinkedinExclusion{
	ProjectID:          2,
	OrgID:              "123",
	Timestamp:          202405,
	Campaigns:          &postgres.Jsonb{RawMessage: sampleCampaigns},
	CampaignsCount:     1,
	CompanyName:        "some company",
	RuleID:             "1234abcd",
	RuleSnapshot:       &postgres.Jsonb{RawMessage: sampleRule},
	RuleObjectType:     LINKEDIN_CAMPAIGN_GROUP,
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
