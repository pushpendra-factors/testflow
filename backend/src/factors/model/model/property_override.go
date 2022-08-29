package model

import "time"

type PropertyOverride struct {
	ProjectID    int64     `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	PropertyName string    `json:"property_name"`
	OverrideType int       `json:"override_type"`
	Entity       int       `gorm:"not null" json:"entity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
