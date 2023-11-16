package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// FactorsGoal - DB model for table - goals
type FactorsGoal struct {
	ID            uint64         `gorm:"primary_key:true;" json:"id"`
	ProjectID     int64          `json:"project_id"`
	Name          string         `json:"name"`
	Rule          postgres.Jsonb `json:"rule,omitempty"`
	Type          string         `gorm:"not null;type:varchar(2)" json:"type"`
	CreatedBy     *string        `json:"created_by"`
	LastTrackedAt *time.Time     `json:"last_tracked_at"`
	IsActive      bool           `json:"is_active"`
	CreatedAt     *time.Time     `json:"created_at"`
	UpdatedAt     *time.Time     `json:"updated_at"`
}

// FactorsGoalRule - object structure
type FactorsGoalRule struct {
	StartEvent string            `json:"st_en"`
	EndEvent   string            `json:"en_en"`
	Rule       FactorsGoalFilter `json:"rule"`
	Visited    bool              `json:"vs"`
}

// FactorsGoalFilter - rule object
type FactorsGoalFilter struct {
	StartEnEventFitler      []KeyValueTuple `json:"st_en_ft"`
	EndEnEventFitler        []KeyValueTuple `json:"en_en_ft"`
	StartEnUserFitler       []KeyValueTuple `json:"st_us_ft"`
	EndEnUserFitler         []KeyValueTuple `json:"en_us_ft"`
	GlobalFilters           []KeyValueTuple `json:"ft"`
	StartEnGroup            string          `json:"st_en_gp"`
	EndEnGroup              string          `json:"en_en_gp"`
	IncludedEvents          []string        `json:"in_en"`
	IncludedEventProperties []string        `json:"in_epr"`
	IncludedUserProperties  []string        `json:"in_upr"`
}

type ExplainV3Event struct {
	Alias  string          `json:"alias"`
	Label  string          `json:"label"`
	Group  string          `json:"group"`
	Filter []QueryProperty `json:"filter"`
}

type ExplainV3GoalRule struct {
	StartEvent     ExplainV3Event   `json:"st_en"`
	EndEvent       ExplainV3Event   `json:"en_en"`
	IncludedEvents []ExplainV3Event `json:"in_en"`
	Visited        bool             `json:"vs"`
}

// KeyValueTuple - key value pair
type KeyValueTuple struct {
	Key        string  `json:"key"`
	Value      string  `json:"vl"`
	Operator   bool    `json:"operator"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Type       string  `json:"property_type"`
}
