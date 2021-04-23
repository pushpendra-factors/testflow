package model

import(
	"time"
)

const (
	DisplayNameEventEntityType = 1
	DisplayNameEventPropertyEntityType = 2
	DisplayNameUserPropertyEntityType = 3
	DisplayNameObjectEntityType = 4
)

type DisplayName struct {
	// Composite primary key with project_id, event_name_id,key .
	ProjectID   uint64  `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"project_id"`
	EventName   string `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"event_name"`
	PropertyName   string `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"property_name"`
	Tag   string `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"tag"`
	Entity      int     `gorm:"not null" json:"entity"`
	Group        string  `gorm:"not null" json:"group"`
	GroupObjectName        string  `gorm:"not null" json:"group_object_name"`
	DisplayName        string  `gorm:"not null" json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}