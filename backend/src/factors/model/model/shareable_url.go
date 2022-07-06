package model

import (
	"time"
)

type ShareableURL struct {
	ID         string `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	QueryID    string `gorm:"not null" json:"query_id"`
	EntityType int    `gorm:"not null" json:"entity_type"`
	ShareType  int    `gorm:"not null" json:"share_type"`
	// AllowedUsers   string    `gorm:"type:varchar" json:"allowed_users"`
	EntityID  int64     `gorm:"not null" json:"entity_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `gorm:"not null;default:false" json:"is_deleted"`
	ExpiresAt int64     `json:"expires_at"`
	ProjectID int64     `gorm:"not null" json:"project_id"`
	CreatedBy string    `gorm:"not null;type:varchar(255)" json:"created_by"`
}

// Keep them ordered in ascending order with respect to scope of the share type
const (
	ShareableURLShareTypePublic int = iota + 1
	// ShareableURLShareTypeAllProjectUsers
	// ShareableURLShareTypeAllowedUsers
)

const (
	ShareableURLEntityTypeQuery int = iota + 1
	ShareableURLEntityTypeTemplate
	ShareableURLEntityTypeDashboard
)

var ValidShareTypes = map[int]bool{
	ShareableURLShareTypePublic: true,
}

var ValidShareEntityTypes = map[int]bool{
	ShareableURLEntityTypeQuery:     true,
	ShareableURLEntityTypeTemplate:  true,
	ShareableURLEntityTypeDashboard: true,
}
