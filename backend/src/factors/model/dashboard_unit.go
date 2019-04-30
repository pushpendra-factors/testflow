package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type DashboardUnit struct {
	// Composite primary key, id + project_id.
	ID        uint64 `gorm:"primary_key:true" json:"id"`
	ProjectId uint64 `gorm:"primary_key:true" json:"project_id"`
	// Foreign key dashboard(id).
	DashboardId  uint64         `json:"dashboard_id"`
	Title        string         `gorm:"not null" json:"title"`
	Query        postgres.Jsonb `gorm:"not null" json:"query"`
	Presentation string         `gorm:"not null" json:"presentation"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

const (
	presentationLine  = "pl"
	presentationBar   = "pb"
	presentationTable = "pt"
)
