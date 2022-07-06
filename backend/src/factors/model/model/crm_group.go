package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CRMGroup struct {
	ID         string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID  int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source     U.CRMSource     `gorm:"primary_key:true;auto_increment:false" json:"source"`
	Type       int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Timestamp  int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	Action     CRMAction       `gorm:"auto_increment:false;not null" json:"action"`
	Metadata   *postgres.Jsonb `json:"metadata"`
	Properties *postgres.Jsonb `gorm:"not null" json:"properties"`
	Synced     bool            `gorm:"default:false" json:"synced"`
	SyncID     string          `gorm:"default:null" json:"sync_id"`
	UserID     string          `gorm:"default:null" json:"user_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
