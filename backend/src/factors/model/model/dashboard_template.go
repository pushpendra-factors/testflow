package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type UnitInfo struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	// dashboard details
	Position int `json:"position"`
	Size     int `json:"size"`
	// query details
	QueryType     int            `json:"query_type"`
	QuerySettings postgres.Jsonb `json:"query_settings"`
	Query         postgres.Jsonb `json:"query"`
}

type DashboardTemplate struct {
	ID                 string          `gorm:"primary_key:true" json:"id"`
	Title              string          `gorm:"not null" json:"title"`
	Description        string          `json:"description"`
	Dashboard          *postgres.Jsonb `json:"dashboard"`
	Units              *postgres.Jsonb `json:"units"`
	IsDeleted          bool            `gorm:"not null;default:false" json:"is_deleted"`
	SimilarTemplateIds *postgres.Jsonb `json:"similar_template_ids"`
	Tags               *postgres.Jsonb `json:"tags"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
