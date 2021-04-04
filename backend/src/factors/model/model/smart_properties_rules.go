package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const (
	CREATED = "created"
	UPDATED = "updated"
	DELETED = "deleted"
)

type SmartPropertiesRules struct {
	UUID        string          `gorm:"primary_key:true;type:varchar(255)" json:"uuid"`
	ProjectID   uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	TypeAlias   string          `json:"type_alias"`
	Type        int             `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rules       *postgres.Jsonb `json:"rules"`
	IsDeleted   bool            `json:"is_deleted"`
	Picked      bool            `json:"picked"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Rule struct {
	Value   string            `json:"value"`
	Source  string            `json:"source"`
	Filters []ChannelFilterV1 `json:"filters"`
}

type SmartPropertiesRulesConfig struct {
	Name    string   `json:"name"`
	Sources []Source `json:"sources"`
}

type Source struct {
	Name                 string                       `json:"name"`
	ObjectsAndProperties []ChannelObjectAndProperties `json:"objects_and_properties"`
}

type ChannelDocumentsWithFields struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	AdGroupID    string `json:"ad_group_id"`
	AdGroupName  string `json:"ad_group_name"`
}
