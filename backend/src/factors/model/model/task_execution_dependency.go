package model

import (
	"time"
)

type TaskExecutionDependencyDetails struct {
	ID               string    `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	TaskID           uint64    `json:"task_id"`
	DependentTaskID  uint64    `json:"dependent_task_id"`
	DependencyOffset int       `json:"dependency_offset"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
