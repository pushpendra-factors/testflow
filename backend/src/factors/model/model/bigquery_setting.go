package model

import "time"

type BigquerySetting struct {
	ID                      string    `gorm:"primary_key:true;auto_increment:false" json:"id"`
	ProjectID               int64     `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	BigqueryProjectID       string    `gorm:"column:bq_project_id" json:"bq_project_id"`
	BigqueryDatasetName     string    `gorm:"column:bq_dataset_name" json:"bq_dataset_name"`
	BigqueryCredentialsJSON string    `gorm:"column:bq_credentials_json" json:"bq_credentials_json"`
	LastRunAt               int64     `json:"last_run_at"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
