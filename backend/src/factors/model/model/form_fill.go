package model

import (
	"time"
)

type CaptureFormFillResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SDKFormFillPayload struct {
	FormId           string `json:"form_id"`
	Value            string `json:"value"`
	TimeSpent        uint64 `json:"time_spent_on_field"`
	FirstUpdatedTime int64  `json:"first_updated_time"`
	LastUpdatedTime  int64  `json:"last_updated_time"`
	FieldId          string `json:"field_id"`
}

type FormFill struct {
	ProjectID        int64     `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	FormId           string    `json:"form_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Id               string    `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Value            string    `json:"value"`
	TimeSpentOnField uint64    `json:"time_spent_on_field"`
	FirstUpdatedTime int64     `json:"first_updated_time"`
	LastUpdatedTime  int64     `json:"last_updated_time"`
	FieldId          string    `json:"field_id"`
}
