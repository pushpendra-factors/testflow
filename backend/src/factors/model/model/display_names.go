package model

import (
	"time"
)

const (
	DisplayNameEventEntityType         = 1
	DisplayNameEventPropertyEntityType = 2
	DisplayNameUserPropertyEntityType  = 3
	DisplayNameObjectEntityType        = 4
)

type DisplayName struct {
	// Composite primary key with project_id, event_name_id,key .
	ID              string    `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID       int64     `gorm:"primary_key:true;unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"project_id"`
	EventName       string    `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"event_name"`
	PropertyName    string    `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"property_name"`
	Tag             string    `gorm:"unique_index:display_names_project_id_event_name_property_name_tag_unique_idx" json:"tag"`
	EntityType      int       `gorm:"not null" json:"entity_type"`
	GroupName       string    `gorm:"not null" json:"group_name"`
	GroupObjectName string    `gorm:"not null" json:"group_object_name"`
	DisplayName     string    `gorm:"not null" json:"display_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
