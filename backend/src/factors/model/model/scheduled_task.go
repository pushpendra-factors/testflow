package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// ScheduledTask Entity to store the details for scheduled tasks.
type ScheduledTask struct {
	ID            string              `gorm:"primary_key:true;uuid;default:uuid_generate_v4()" json:"id"` // Id for the current task.
	JobID         string              `gorm:"not null" json:"job_id"`                                     // Id for the parent task.
	ProjectID     int64               `gorm:"not null" json:"project_id"`
	TaskType      ScheduledTaskType   `gorm:"not null" json:"task_type"`
	TaskStatus    ScheduledTaskStatus `gorm:"not null" json:"task_status"`
	TaskStartTime int64               `gorm:"default:null" json:"task_start_time"` // Time when tast run started.
	TaskEndTime   int64               `gorm:"default:null" json:"task_end_time"`   // Time when task run ended.
	TaskDetails   *postgres.Jsonb     `json:"task_details"`                        // Metadata for the task.
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type ScheduledTaskStatus string
type ScheduledTaskType string

const (
	TASK_TYPE_EVENTS_ARCHIVAL ScheduledTaskType = "EVENTS_ARCHIVAL"
	TASK_TYPE_BIGQUERY_UPLOAD ScheduledTaskType = "BIGQUERY_UPLOAD"
)

const (
	TASK_STATUS_IN_PROGRESS ScheduledTaskStatus = "IN_PROGRESS"
	TASK_STATUS_SUCCESS     ScheduledTaskStatus = "SUCCESS"
	TASK_STATUS_FAILED      ScheduledTaskStatus = "FAILED"
)

// EventArchivalTaskDetails To store metadata for individual task run.
type EventArchivalTaskDetails struct {
	FromTimestamp int64  `json:"from_timestamp"`
	ToTimestamp   int64  `json:"to_timestamp"`
	EventCount    int64  `json:"event_count"`
	FileCreated   bool   `json:"file_created"`
	FilePath      string `json:"filepath"`
	UsersFilePath string `json:"users_filepath"`
	BucketName    string `json:"bucket_name"`
}

// BigqueryUploadTaskDetails To store metadata for bigquery upload tasks.
type BigqueryUploadTaskDetails struct {
	FromTimestamp     int64           `json:"from_timestamp"`
	ToTimestamp       int64           `json:"to_timestamp"`
	BigqueryProjectID string          `json:"bq_project_id"`
	BigqueryDataset   string          `json:"bq_dataset"`
	BigqueryTable     string          `json:"bq_table"`
	ArchivalTaskID    string          `json:"archival_task_id"`
	UploadStats       *postgres.Jsonb `json:"upload_stats"`
	UsersUploadStats  *postgres.Jsonb `json:"users_supload_stats"`
}
