package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Queries struct {
	// Composite primary key, id + project_id.
	ID int64 `gorm:"primary_key:true;auto_increment:false" json:"id"`
	// Foreign key queries(project_id) ref projects(id).
	ProjectID                  int64          `gorm:"primary_key:true" json:"project_id"`
	Title                      string         `gorm:"not null" json:"title"`
	Query                      postgres.Jsonb `gorm:"not null" json:"query"`
	Type                       int            `gorm:"not null; primary_key:true" json:"type"`
	IsDeleted                  bool           `gorm:"not null;default:false" json:"is_deleted"`
	CreatedBy                  string         `gorm:"type:varchar(255)" json:"created_by"`
	CreatedByName              string         `gorm:"-" json:"created_by_name"`
	CreatedByEmail             string         `gorm:"-" json:"created_by_email"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	Settings                   postgres.Jsonb `json:"settings"`
	IdText                     string         `json:"id_text"`
	Converted                  bool
	LockedForCacheInvalidation bool
	IsDashboardQuery           bool `gorm:"-" json:"is_dashboard_query"`
}

type QueriesString struct {
	// Composite primary key, id + project_id.
	ID string `gorm:"primary_key:true;auto_increment:false" json:"id"`
	// Foreign key queries(project_id) ref projects(id).
	ProjectID        int64          `gorm:"primary_key:true" json:"project_id"`
	Title            string         `gorm:"not null" json:"title"`
	Query            postgres.Jsonb `gorm:"not null" json:"query"`
	Type             int            `gorm:"not null; primary_key:true" json:"type"`
	IsDeleted        bool           `gorm:"not null;default:false" json:"is_deleted"`
	CreatedBy        string         `gorm:"type:varchar(255)" json:"created_by"`
	CreatedByName    string         `gorm:"-" json:"created_by_name"`
	CreatedByEmail   string         `gorm:"-" json:"created_by_email"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Settings         postgres.Jsonb `json:"settings"`
	IdText           string         `json:"id_text"`
	Converted        bool
	IsDashboardQuery bool `gorm:"-" json:"is_dashboard_query"`
}

const (
	QueryTypeAllQueries         = 0
	QueryTypeDashboardQuery     = 1
	QueryTypeSavedQuery         = 2
	QueryTypeAttributionV1Query = 3
	QueryTypeSixSignalQuery     = 4
)
