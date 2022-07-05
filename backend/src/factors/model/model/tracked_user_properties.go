package model

import "time"

// FactorsTrackedUserProperty - DB model for table: tracked_events
type FactorsTrackedUserProperty struct {
	ID               uint64     `gorm:"primary_key:true;" json:"id"`
	ProjectID        int64      `json:"project_id"`
	UserPropertyName string     `json:"user_property_name"`
	Type             string     `gorm:"not null;type:varchar(2)" json:"type"`
	CreatedBy        *string    `json:"created_by;default:null"`
	LastTrackedAt    *time.Time `json:"last_tracked_at"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}
