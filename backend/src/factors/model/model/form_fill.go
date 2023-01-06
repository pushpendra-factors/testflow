package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type CaptureFormFillResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SDKFormFillPayload struct {
	UserId          string          `json:"user_id"`
	FieldId         string          `json:"field_id"`
	FormId          string          `json:"form_id"`
	Value           string          `json:"value"`
	UpdatedAt       *time.Time      `json:"-"`
	EventProperties *postgres.Jsonb `json:"event_properties"`
}

type FormFill struct {
	ProjectID       int64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	FormId          string          `json:"form_id"`
	Id              string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Value           string          `json:"value"`
	FieldId         string          `json:"field_id"`
	UserId          string          `json:"user_id"`
	EventProperties *postgres.Jsonb `json:"event_properties"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
