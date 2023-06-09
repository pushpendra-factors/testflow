package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type G2Document struct {
	ProjectID int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Timestamp int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias string          `gorm:"-" json:"type_alias"`
	Value     *postgres.Jsonb `json:"value"`
	Synced    bool            `json:"synced"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// G2LastSyncInfo doc type last sync info
type G2LastSyncInfo struct {
	ProjectID int64  `json:"project_id"`
	Type      int    `json:"type"`
	TypeAlias string `json:"type_alias"`
	Timestamp int64  `json:"timestamp"`
}
