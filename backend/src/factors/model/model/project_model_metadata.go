package model

import "time"

type ProjectModelMetadata struct {
	ID        string    `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectId int64     `json:"project_id"`
	ModelId   uint64    `json:"model_id"`
	ModelType string    `json:"model_type"`
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
	Chunks    string    `json:"chunks"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
