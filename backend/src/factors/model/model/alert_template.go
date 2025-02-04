package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type AlertTemplate struct {
	V                 string          `gorm:"not null" json:"v"`
	ID                int             `gorm:"primary_key:true" json:"id"`
	Title             string          `gorm:"not null" json:"title"`
	Alert             *postgres.Jsonb `json:"alert"`
	TemplateConstants *postgres.Jsonb `json:"template_constants"`
	WorkflowConfig    *postgres.Jsonb `json:"workflow_config"`
	IsWorkflow        bool            `json:"is_workflow"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	IsDeleted         bool            `gorm:"not null;default:false" json:"is_deleted"`
}
