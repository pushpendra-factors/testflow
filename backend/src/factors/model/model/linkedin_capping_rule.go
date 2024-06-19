package model

import (
	"encoding/json"
	U "factors/util"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

var LINKEDIN_FREQUENCY_CAPPING_OBJECTS = [3]string{LINKEDIN_ACCOUNT, LINKEDIN_CAMPAIGN_GROUP, LINKEDIN_CAMPAIGN}

type LinkedinCappingRule struct {
	ID                    string          `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	ProjectID             int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ObjectType            string          `gorm:"primary_key:true" json:"object_type"` //account, campaign, campaign_group
	Name                  string          `gorm:"primary_key:true" json:"name"`
	DisplayName           string          `json:"display_name"`
	Status                string          `json:"status"`
	Description           string          `json:"description"`
	ObjectIDs             *postgres.Jsonb `json:"object_ids"`
	Granularity           string          `json:"granularity"`
	ImpressionThreshold   int64           `json:"impression_threshold"`
	ClickThreshold        int64           `json:"click_threshold"`
	IsAdvancedRuleEnabled bool            `json:"is_advanced_rule_enabled"`
	AdvancedRuleType      string          `json:"advanced_rule_type"` //account or segment
	AdvancedRules         *postgres.Jsonb `json:"advanced_rules"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type LinkedinCappingRuleWithDecodedValues struct {
	Rule          LinkedinCappingRule
	ObjectIDs     []string
	AdvancedRules []AdvancedRuleFilters
}

type AdvancedRuleFilters struct {
	Filters             []QueryProperty `json:"filters"`
	ImpressionThreshold int64           `json:"impression_threshold"`
	ClickThreshold      int64           `json:"click_threshold"`
}

type RuleMatchedDataSet struct {
	CappingData       LinkedinCappingDataSet
	PropertiesMatched map[string]interface{}
	Rule              LinkedinCappingRule
	Campaigns         []CampaignsIDName
}

type GroupRelatedData struct {
	UserID       string
	GroupUsers   []User
	DecodedProps []map[string]interface{}
}

const (
	LINKEDIN_ACCOUNT        = "account"
	LINKEDIN_CAMPAIGN_GROUP = "campaign_group"
	LINKEDIN_CAMPAIGN       = "campaign"
	LINKEDIN_STATUS_ACTIVE  = "active"
	LINKEDIN_STATUS_PAUSED  = "paused"
	LINKEDIN_STATUS_DELETED = "deleted"
)

var SampleCappingRule = LinkedinCappingRule{
	ID:                    "1234abcd",
	ProjectID:             2,
	ObjectType:            LINKEDIN_CAMPAIGN_GROUP,
	Name:                  "sample_rule",
	DisplayName:           "Sample Rule",
	Status:                LINKEDIN_STATUS_ACTIVE,
	Description:           "something",
	ObjectIDs:             &postgres.Jsonb{RawMessage: json.RawMessage(`["cg123"]`)},
	IsAdvancedRuleEnabled: true,
	AdvancedRuleType:      "account",
	AdvancedRules:         &postgres.Jsonb{RawMessage: SampleAdvancedRule},
	CreatedAt:             time.Time{},
	UpdatedAt:             time.Time{},
}

var SampleAdvancedRule, _ = json.Marshal([]AdvancedRuleFilters{{
	Filters: []QueryProperty{
		{
			Type:      U.PropertyTypeCategorical,
			Property:  U.DP_ENGAGEMENT_LEVEL,
			Operator:  EqualsOp,
			Value:     ENGAGEMENT_LEVEL_HOT,
			LogicalOp: LOGICAL_OP_AND,
			Entity:    PropertyEntityUserGlobal,
		},
	},
	ImpressionThreshold: 1000,
	ClickThreshold:      100,
},
})

type LinkedinCappingConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
	Type    string `json:"type"`
}

var SampleCampaignGroupConfig = []LinkedinCappingConfig{
	{
		ID:      "cg123",
		Name:    "CG1",
		Deleted: false,
		Type:    "campaign_group",
	},
	{
		ID:      "cg234",
		Name:    "CG2",
		Deleted: true,
		Type:    "campaign_group",
	},
}

var SampleCampaignConfig = []LinkedinCappingConfig{
	{
		ID:      "c123",
		Name:    "C1",
		Deleted: false,
		Type:    "campaign",
	},
	{
		ID:      "c234",
		Name:    "C2",
		Deleted: true,
		Type:    "campaign",
	},
}

func GenerateNameFromDisplayName(displayName string) string {
	lowerName := strings.ToLower(displayName)

	return regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(lowerName, "_")
}
