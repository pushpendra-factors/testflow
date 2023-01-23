package model

import "github.com/jinzhu/gorm/dialects/postgres"

type SegmentResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SegmentPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       Query  `json:"query"`
	Type        string `json:"type"`
}

type Segment struct {
	ProjectID   int64           `gorm:"primary_key:true" json:"project_id"`
	Id          string          `gorm:"primary_key:true" json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Query       *postgres.Jsonb `json:"query"`
	Type        string          `json:"type"`
}
