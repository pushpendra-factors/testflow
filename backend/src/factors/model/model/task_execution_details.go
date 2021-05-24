package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type TaskExecutionDetails struct {
	ID          string          `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	ExecutionID uint64          `gorm:"primary_key:true;auto_increment:false" json:"execution_id"`
	TaskID      uint64          `json:"task_id"`
	ProjectID   uint64          `json:"project_id"`
	Delta       uint64          `json:"delta"`
	Metadata    *postgres.Jsonb `json:"metadata"`
	IsCompleted bool            `json:"is_completed"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
