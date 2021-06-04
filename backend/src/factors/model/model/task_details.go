package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const (
	Stateless = 1
	Hourly    = 2
	Daily     = 3
	Weekly    = 4
	Monthly   = 5
	Quarterly = 6
)

type TaskDetails struct {
	ID                       string          `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	TaskID                   uint64          `gorm:"primary_key:true;auto_increment:false" json:"task_id"`
	TaskName                 string          `json:"task_name"`
	Source                   string          `json:"source"`
	Frequency                int             `json:"frequency"`
	FrequencyInterval        int             `json:"frequency_interval"`
	SkipStartIndex           int             `json:"skip_start_index"`
	SkipEndIndex             int             `json:"skip_end_index"`
	OffsetStartMinutes       int             `json:"offset_start_minutes"`
	Recurrence               bool            `json:"recurrence"`
	Metadata                 *postgres.Jsonb `json:"metadata"`
	IsProjectEnabled         bool            `json:"is_project_enabled"`
	DelayAlertThresholdHours uint64          `json:"delay_alert_threshold_hours"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}
