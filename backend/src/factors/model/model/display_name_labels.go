package model

import (
	"time"
)

type DisplayNameLabel struct {
	ID          string    `gorm:"primary_key:true;type:varchar(255)" json:"id"`
	ProjectID   int64     `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	Source      string    `json:"source"`
	PropertyKey string    `json:"key"`
	Value       string    `json:"value"`
	Label       string    `json:"label"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
