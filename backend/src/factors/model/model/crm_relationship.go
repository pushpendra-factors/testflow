package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CRMRelationship struct {
	ID        string      `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID int64       `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source    U.CRMSource `gorm:"primary_key:true;auto_increment:false" json:"source"`
	FromType  int         `gorm:"primary_key:true;auto_increment:false" json:"from_type"`
	FromID    string      `gorm:"primary_key:true;auto_increment:false" json:"from_id"`
	ToType    int         `gorm:"primary_key:true;auto_increment:false" json:"to_type"`
	ToID      string      `gorm:"primary_key:true;auto_increment:false" json:"to_id"`
	Timestamp int64       `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	// Store external_relationship_name and external_relationship_id - if the relationship is defined by another object.
	ExternalRelationshipName string `json:"external_relationship_name"`
	ExternalRelationshipID   string `json:"external_relationship_id"`
	// Properties stores information related to the relatonship if required for special processing
	Properties  *postgres.Jsonb `json:"properties"`
	SkipProcess bool            `json:"skip_process"` // set skip_process if it should not be processed
	Synced      bool            `json:"synced"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
