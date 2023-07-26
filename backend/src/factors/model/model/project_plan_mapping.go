package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type ProjectPlanMapping struct {
	ProjectID     int64           `gorm:"column:project_id" json:"project_id"`
	PlanID        int64           `gorm:"column:plan_id" json:"plan_id"`
	OverWrite     *postgres.Jsonb `gorm:"column:over_write" json:"over_write"` //OverWrite type from plan_details model
	LastRenewedOn time.Time       `gorm:"column:last_renewed_on" json:"last_renewed_on"`
}
