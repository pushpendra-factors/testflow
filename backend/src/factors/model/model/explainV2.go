package model

import (
	"fmt"
	"math"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// id text NOT NULL,
// task_id bigint NOT NULL,
// project_id bigint NOT NULL,
// query json,
// status bigint NOT NULL,
// is_deleted boolean,
// created_at timestamp(6) NOT NULL,
// updated_at timestamp(6) NOT NULL,
// Primary key (id)

const (
	TASK_QUEUED  = 0
	TASK_RUNNING = 1
	TASK_ERROR   = 2
	TASK_DELETED = 3
	TASK_DONE    = 4
)

// type QueryExplainV2 struct {
// 	ID         string          `gorm:"primary_key:true" json:"id"`
// 	Task_Id    int64           `json:"tid"`
// 	Project_id int64           `json:"pid"`
// 	Query      *postgres.Jsonb `json:"qry"`
// 	Status     int64           `json:"sts"`
// 	Is_deleted bool            `json:"del"`
// 	Created_at time.Time       `json:"cdt"`
// 	Updated_at time.Time       `json:"udt"`
// }

type ExplainV2 struct {
	ID             string          `gorm:"column:id; type:uuid; default:uuid_generate_v4()" json:"id"`
	ProjectID      int64           `gorm:"column:project_id; primary_key:true" json:"project_id"`
	Title          string          `gorm:"column:title; not null" json:"title"`
	ExplainV2Query *postgres.Jsonb `gorm:"column:query" json:"query"`
	Status         string          `gorm:"column:status" json:"status"`
	IsDeleted      bool            `gorm:"column:is_deleted; not null; default:false" json:"is_deleted"`
	CreatedBy      string          `gorm:"column:created_by" json:"created_by"`
	CreatedAt      time.Time       `gorm:"column:created_at; autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at; autoUpdateTime" json:"updated_at"`
	ModelID        uint64          `gorm:"column:model_id" json:"model_id"`
}

type ExplainV2Query struct {
	Title          string            `json:"name"`
	Query          FactorsGoalRule   `json:"rule"`
	StartTimestamp int64             `json:"sts"`
	EndTimestamp   int64             `json:"ets"`
	Raw_query      string            `json:"rw"`
	QueryV3        ExplainV3GoalRule `json:"rulev3"`
	IsV3           bool              `json:"v3"`
}

type ExplainV2EntityInfo struct {
	Id             string          `json:"id"`
	Title          string          `json:"title"`
	Status         string          `json:"status"`
	CreatedBy      string          `json:"created_by"`
	Date           time.Time       `json:"date"`
	ExplainV2Query FactorsGoalRule `json:"query"`
	ModelID        uint64          `json:"mid"`
	Raw_query      string          `json:"rq"`
}

type ExplainV3EntityInfo struct {
	Id             string            `json:"id"`
	Title          string            `json:"title"`
	Status         string            `json:"status"`
	CreatedBy      string            `json:"created_by"`
	Date           time.Time         `json:"date"`
	ExplainV3Query ExplainV3GoalRule `json:"query"`
	StartTimestamp int64             `json:"sts"`
	EndTimestamp   int64             `json:"ets"`
}

func ConvertFactorsGoalRuleToExplainV3GoalRule(query FactorsGoalRule) ExplainV3GoalRule {

	var queryV2 ExplainV3GoalRule
	if query.StartEvent != "" {
		queryV2.StartEvent = ExplainV3Event{Label: query.StartEvent}
	}
	if query.EndEvent != "" {
		queryV2.EndEvent = ExplainV3Event{Label: query.EndEvent}
	}
	for _, evName := range query.Rule.IncludedEvents {
		queryV2.IncludedEvents = append(queryV2.IncludedEvents, ExplainV3Event{Label: evName})
	}
	queryV2.Visited = query.Visited
	var startFilters = make([]QueryProperty, 0)
	for _, filter := range query.Rule.StartEnEventFitler {
		startFilters = append(startFilters, ReverseMapProperty(filter, "event"))
	}
	for _, filter := range query.Rule.StartEnUserFitler {
		startFilters = append(startFilters, ReverseMapProperty(filter, "user"))
	}
	queryV2.StartEvent.Filter = startFilters
	var endFilters = make([]QueryProperty, 0)
	for _, filter := range query.Rule.EndEnEventFitler {
		endFilters = append(endFilters, ReverseMapProperty(filter, "event"))
	}
	for _, filter := range query.Rule.EndEnUserFitler {
		endFilters = append(endFilters, ReverseMapProperty(filter, "user"))
	}
	queryV2.EndEvent.Filter = endFilters
	return queryV2
}

func ReverseMapProperty(ip KeyValueTuple, entity string) QueryProperty {
	op := QueryProperty{}
	op.Entity = entity
	op.Type = ip.Type
	op.Property = ip.Key
	if ip.Type == "categorical" {
		op.Value = ip.Value
		if ip.Operator {
			op.Operator = "equals"
		} else {
			op.Operator = "notEqual"
		}
	}
	if ip.Type == "numerical" {
		if ip.LowerBound != -math.MaxFloat64 {
			op.Value = fmt.Sprintf("%f", ip.LowerBound)
			op.Operator = "lowerThan"
		} else {
			op.Value = fmt.Sprintf("%f", ip.UpperBound)
			op.Operator = "greaterThan"
		}
	}
	return op
}
