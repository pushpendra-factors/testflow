package model

import (
	U "factors/util"
	"time"
)

// CRMProperty interface for crm_properties table
type CRMProperty struct {
	ID               string      `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID        int64       `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source           U.CRMSource `gorm:"primary_key:true;auto_increment:false" json:"source"`
	Type             int         `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Name             string      `gorm:"primary_key:true;auto_increment:false" json:"name"`
	Timestamp        int64       `json:"timestamp"`
	ExternalDataType string      `json:"external_data_type"`
	MappedDataType   string      `json:"mapped_data_type"`
	Label            string      `json:"label"`
	Synced           bool        `json:"synced"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

func IsValidCRMMappedDataType(dataType string) bool {
	if dataType != U.PropertyTypeDateTime && dataType != U.PropertyTypeNumerical {
		return false
	}

	return true
}
