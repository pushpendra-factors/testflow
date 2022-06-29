package model

import (
	"time"
)

type WeeklyInsightsMetadata struct {
	ID                  string    `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectId           uint64    `json:"project_id"`
	QueryId             int64     `json:"query_id"`
	InsightType         string    `json:"insight_type"`
	BaseStartTime       int64     `json:"base_start_time"`
	BaseEndTime         int64     `json:"base_end_time"`
	ComparisonStartTime int64     `json:"comparison_start_time"`
	ComparisonEndTime   int64     `json:"comparison_end_time"`
	InsightId           uint64    `json:"insight_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
