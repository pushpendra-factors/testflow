package model

import "time"

type FivetranMappings struct {
	ID          string     `gorm:"primary_key:true;" json:"id"`
	ProjectID   uint64     `json:"project_id"`
	Integration string     `json:"integration"`
	ConnectorID string     `json:"connector_id"`
	SchemaID    string     `json:"schema_id"`
	Accounts    string     `json:"accounts"`
	Status      bool       `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

var BingAdsIntegration = "bingads"
var MarketoIntegration = "marketo"
