package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const (
	// Status
	ACTIVE   = "active"
	BUILDING = "building"
	SAVED    = "saved"

	// Event type
	STARTSWITH = "startswith"
	ENDSWITH   = "endswith"
	BuildLimit = 10
)

type PathAnalysis struct {
	ID                string          `gorm:"column:id; type:uuid; default:uuid_generate_v4()" json:"id"`
	ProjectID         int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	Title             string          `gorm:"column:title; not null" json:"title"`
	PathAnalysisQuery *postgres.Jsonb `gorm:"column:query" json:"query"`
	Status            string          `gorm:"column:status" json:"status"`
	IsDeleted         bool            `gorm:"column:is_deleted; not null; default:false" json:"is_deleted"`
	CreatedBy         string          `gorm:"column:created_by" json:"created_by"`
	CreatedAt         time.Time       `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
}

type PathAnalysisQuery struct {
	Title               string              `json:"title"`
	EventType           string              `json:"event_type"`
	Group 				string 				`json:"group"`
	Event               PathAnalysisEvent   `json:"event"`
	NumberOfSteps       int                 `json:"steps"`
	IncludeEvents       []PathAnalysisEvent `json:"include_events"`
	ExcludeEvents       []PathAnalysisEvent `json:"exclude_events"`
	StartTimestamp      int64               `json:"starttimestamp"`
	EndTimestamp        int64               `json:"endtimestamp"`
	AvoidRepeatedEvents bool                `json:"avoid_repeated_events"`
	Filter              []QueryProperty     `json:"filter"`
}

type PathAnalysisEvent struct {
	Filter []QueryProperty `json:"filter"`
	Alias  string          `json:"alias"`
	Label  string          `json:"label"`
	Group  string          `json:"group"`
}

type PathAnalysisEntityInfo struct {
	Id                string            `json:"id"`
	Title             string            `json:"title"`
	Status            string            `json:"status"`
	CreatedBy         string            `json:"created_by"`
	Date              time.Time         `json:"date"`
	PathAnalysisQuery PathAnalysisQuery `json:"query"`
}

// type Tabler interface {
// 	TableName() string
// }

// TableName overrides the table name used by PathAnalysis from `path_analyses` to `pathanalysis`
func (PathAnalysis) TableName() string {
	return "pathanalysis"
}
