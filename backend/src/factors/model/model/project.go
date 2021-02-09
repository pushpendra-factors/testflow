package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Project struct {
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `gorm:"not null;" json:"name"`
	// An index created on token.
	Token string `gorm:"size:32" json:"token"`
	// An index created on private_token.
	PrivateToken string          `gorm:"size:32" json:"private_token"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ProjectURI   string          `json:"project_uri"`
	TimeFormat   string          `json:"time_format"`
	DateFormat   string          `json:"date_format"`
	TimeZone     string          `json:"time_zone"`
	JobsMetadata *postgres.Jsonb `json:"-"`
}

const (
	JobsMetadataKeyNextSessionStartTimestamp = "next_session_start_timestamp"
	JobsMetadataColumnName                   = "jobs_metadata"
)

const DefaultProjectName = "My Project"
