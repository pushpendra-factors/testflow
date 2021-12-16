package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type ContentGroup struct {
	ID                      string          `gorm:"primary_key:true;" json:"id"`
	ProjectID               uint64          `json:"project_id"`
	ContentGroupName        string          `json:"content_group_name"`
	ContentGroupDescription string          `json:"content_group_description"`
	Rule                    *postgres.Jsonb `json:"rule,omitempty"`
	CreatedBy               string          `json:"created_by"`
	IsDeleted               bool            `json:"is_deleted"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

type ContentGroupRule struct {
	ContentGroupValue string                  `json:"content_group_value"`
	Rule              ContentGroupRuleFilters `json:"rule,omitempty"`
}

type ContentGroupRuleFilters []ContentGroupValue

type ContentGroupValue struct {
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}
