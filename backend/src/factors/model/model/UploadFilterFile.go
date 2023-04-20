package model

import "time"

type UploadFilterFiles struct {
	FileReference string    `gorm:"primary_key:true;" json:"file_reference"`
	ProjectID     int64     `json:"project_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
