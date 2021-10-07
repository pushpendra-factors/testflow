package model

import "time"

// Group is an interface for groups table
type Group struct {
	ProjectID uint64    `gorm:"primary_key:true;" json:"project_id"`
	ID        int       `gorm:"not null" json:"id"`
	Name      string    `gorm:"primary_key:true;" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const GROUP_NAME_HUBSPOT_COMPANY = "$hubspot_company"

// AllowedGroupNames list of allowed group names
var AllowedGroupNames = map[string]bool{
	GROUP_NAME_HUBSPOT_COMPANY: true,
}

// AllowedGroups total groups allowed per project
var AllowedGroups = 4
