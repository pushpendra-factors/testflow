package model

import "github.com/jinzhu/gorm/dialects/postgres"

type SegmentResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SegmentPayload struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Query       SegmentQuery `json:"query"`
	Type        string       `json:"type"`
}

type Segment struct {
	ProjectID   int64           `gorm:"primary_key:true" json:"project_id"`
	Id          string          `gorm:"primary_key:true" json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Query       *postgres.Jsonb `json:"query"`
	Type        string          `json:"type"`
}

type SegmentQuery struct {
	EventsWithProperties []EventWithProperty `json:"ewp"`
	GlobalProperties     []QueryProperty     `json:"gp"`
}

type EventWithProperty struct {
	Names           string          `json:"na"`
	Properties      []QueryProperty `json:"pr"`
	LogicalOperator string          `json:"lop"`
}
