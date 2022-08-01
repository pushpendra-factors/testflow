package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type OTPRule struct {
	// Composite primary key with project_id and uuid.
	ID                string         `gorm:"primary_key:true;type:uuid" json:"id"`
	ProjectID         int64          `gorm:"primary_key:true" json:"project_id"`
	RuleType          string         `gorm:"not null" json:"rule_type"`
	TouchPointTimeRef string         `gorm:"not null" json:"touch_point_time_ref"`
	Filters           postgres.Jsonb `gorm:"not null" json:"filters"`
	PropertiesMap     postgres.Jsonb `json:"properties_map"`
	IsDeleted         bool           `gorm:"not null;default:false" json:"is_deleted"`
	CreatedBy         string         `json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type TouchPointFilter struct {
	Property string `json:"pr"`
	// Entity: user or event.
	Entity string `json:"en"`
	// Type: categorical or numerical
	Type      string `json:"ty"`
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}

type TouchPointPropertyValue struct {
	Type  string `json:"ty"`
	Value string `json:"va"`
}

const (
	LastModifiedTimeRef               = "LAST_MODIFIED_TIME_REF"
	TouchPointPropertyValueAsProperty = "Property"
	TouchPointPropertyValueAsConstant = "Constant"
	TouchPointRuleTypeEmails          = "Emails"
	TouchPointRuleTypeMeetings        = "Meetings"
	TouchPointRuleTypeCalls           = "Calls"
	TouchPointRuleTypeForms           = "Form_Submissions"
	TouchPointRuleTypeHSNormal        = "HS_normal"
	TouchPointRuleTypeSFNormal        = "SF_normal"
)
