package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// type WorkflowTemplate struct {
// 	ID                       string          `json:"id"`
// 	Name                     string          `json:"name"`
// 	WorkflowStaticDataFields *postgres.Jsonb `json:"workflow_static_data_fields"`
// 	AlertBody                *postgres.Jsonb `json:"alert_body"`
// 	CreatedAt                int64           `json:"created_at"`
// 	UpdatedAt                int64           `json:"updated_at"`
// 	IsDeleted                bool            `json:"is_deleted"`
// }

type Workflow struct {
	ID        string          `json:"id"`
	ProjectID int64           `json:"project_id"`
	Name      string          `json:"name"`
	AlertBody *postgres.Jsonb `json:"alert_body"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	CreatedBy string          `json:"created_by"`
	IsDeleted bool            `gorm:"not null;default:false" json:"is_deleted"`
}

type WorkflowStaticDataFields struct {
	ShortDescription string   `json:"short_description"`
	LongDescription  string   `json:"long_description"`
	ImageAddress     string   `json:"image_address"`
	Category         []string `json:"category"`
	Tags             []string `json:"tags"`
	Integrations     []string `json:"integrations"`
}

type WorkflowAlertBody struct {
	Title                 string                 `json:"title"`
	Event                 string                 `json:"event"`
	EventLevel            string                 `json:"event_level"`
	Filters               []QueryProperty        `json:"filters"`
	BreakdownProperties   []QueryProperty        `json:"breakdown_properties"`
	PayloadMappings       map[string]interface{} `json:"payload_mappings"`
	DeliveryConfiguration string                 `json:"delivery_configuration"`
}