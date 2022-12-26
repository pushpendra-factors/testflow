package model

import (
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
	Title          string          `json:"ti"`
	Query          FactorsGoalRule `json:"fr"`
	StartTimestamp int64           `json:"sts"`
	EndTimestamp   int64           `json:"ets"`
}

type ExplainV2EntityInfo struct {
	Id             string         `json:"id"`
	Title          string         `json:"title"`
	Status         string         `json:"status"`
	CreatedBy      string         `json:"created_by"`
	Date           time.Time      `json:"date"`
	ExplainV2Query ExplainV2Query `json:"query"`
	ModelID        uint64         `json:"mid"`
}
