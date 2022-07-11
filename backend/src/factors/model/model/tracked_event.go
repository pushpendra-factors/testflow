package model

import "time"

// TrackedEvent - DB model for table: tracked_events
type FactorsTrackedEvent struct {
	ID            uint64     `gorm:"primary_key:true;" json:"id"`
	ProjectID     int64      `json:"project_id"`
	EventNameID   string     `json:"event_name_id"`
	Type          string     `gorm:"not null;type:varchar(2)" json:"type"`
	CreatedBy     *string    `json:"created_by"`
	LastTrackedAt *time.Time `json:"last_tracked_at"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type FactorsTrackedEventInfo struct {
	ID            uint64     `gorm:"primary_key:true;" json:"id"`
	ProjectID     int64      `json:"project_id"`
	EventNameID   string     `json:"event_name_id"`
	Type          string     `gorm:"not null;type:varchar(2)" json:"type"`
	CreatedBy     *string    `json:"created_by"`
	LastTrackedAt *time.Time `json:"last_tracked_at"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	Name          string     `json:"name"`
}
