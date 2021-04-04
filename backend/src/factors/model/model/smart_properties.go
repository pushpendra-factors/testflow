package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type SmartProperties struct {
	ProjectID      uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ObjectType     int             `gorm:"primary_key:true;auto_increment:false" json:"object_type"`
	ObjectID       string          `gorm:"primary_key:true;auto_increment:false" json:"object_id"`
	ObjectProperty *postgres.Jsonb `json:"object_property"`
	Properties     *postgres.Jsonb `json:"properties"`
	RulesRef       *postgres.Jsonb `json:"rules_ref"`
	Source         string          `gorm:"primary_key:true;auto_increment:false" json:"source"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
