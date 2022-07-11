package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type TaskExecutionDetails struct {
	ID          string          `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	TaskID      uint64          `json:"task_id"`
	ProjectID   int64           `json:"project_id"`
	Delta       uint64          `json:"delta"`
	Metadata    *postgres.Jsonb `json:"metadata"`
	IsCompleted bool            `json:"is_completed"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
