package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CRMActivity struct {
	ID         string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID  uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source     CRMSource       `gorm:"primary_key:true;auto_increment:false" json:"source"`
	Name       string          `gorm:"primary_key:true;auto_increment:false" json:"name"`
	Type       int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	ActorType  int             `gorm:"primary_key:true;auto_increment:false" json:"actor_type"`
	ActorID    string          `gorm:"primary_key:true;auto_increment:false" json:"actor_id"`
	Timestamp  int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	Properties *postgres.Jsonb `json:"properties"`
	Synced     bool            `json:"synced"`
	SyncID     string          `json:"sync_id"`
	UserID     string          `json:"user_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
