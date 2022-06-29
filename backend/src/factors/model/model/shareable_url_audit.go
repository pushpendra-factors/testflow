package model

import (
	"time"
)

type ShareableURLAudit struct {
	ID         string `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID  uint64 `gorm:"not null" json:"project_id"`
	ShareID    string `gorm:"type:uuid;not null" json:"share_id"`
	QueryID    string `json:"query_id"`
	EntityID   int64  `json:"entity_id"`
	EntityType int    `json:"entity_type"`
	ShareType  int    `json:"share_type"`
	// AllowedUsers   string    `gorm:"type:varchar" json:"allowed_users"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsDeleted  bool      `gorm:"not null;default:false" json:"is_deleted"`
	ExpiresAt  int64     `json:"expires_at"`
	AccessedBy string    `gorm:"type:varchar(255)" json:"accessed_by"`
}
