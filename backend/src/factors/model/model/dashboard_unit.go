package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type DashboardUnit struct {
	// Composite primary key, id + project_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboard_units(project_id) ref projects(id).
	ProjectID    uint64    `gorm:"primary_key:true" json:"project_id"`
	DashboardId  uint64    `gorm:"primary_key:true" json:"dashboard_id"`
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `json:"description"`
	Presentation string    `gorm:"type:varchar(5);not null" json:"presentation"`
	IsDeleted    bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// TODO (Anil) remove this field once we move to saved queries
	Query    postgres.Jsonb `gorm:"not null" json:"query"`
	QueryId  uint64         `gorm:"not null" json:"query_id"`
	Settings postgres.Jsonb `json:"settings"`
}

type DashboardUnitRequestPayload struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	// TODO (Anil) remove this field once we move to saved queries
	Query    *postgres.Jsonb `json:"query"`
	QueryId  uint64          `json:"query_id"`
	Settings *postgres.Jsonb `json:"settings"`
}

func getDashboardUnitQueryResultCacheKey(projectID, dashboardID, unitID uint64, from, to int64) (*cacheRedis.Key, error) {
	prefix := "dashboard:query"
	var suffix string
	if U.IsStartOfTodaysRange(from, U.TimeZoneStringIST) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:now", dashboardID, unitID, from)
	} else {
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:%d", dashboardID, unitID, from, to)
	}
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

func IsDashboardUnitAlreadyCachedForRange(projectID, dashboardID, unitID uint64, from, to int64) bool {
	if U.IsStartOfTodaysRange(from, U.TimeZoneStringIST) {
		// If from time is of today's beginning, refresh today's everytime a request is received.
		return false
	}
	cacheKey, err := getDashboardUnitQueryResultCacheKey(projectID, dashboardID, unitID, from, to)
	if err != nil {
		log.WithError(err).Errorf("Failed to get cache key")
		return false
	}
	exists, err := cacheRedis.ExistsPersistent(cacheKey)
	if err != nil {
		log.WithError(err).Errorf("Redis error on exists")
		return false
	}
	return exists
}

const (
	DashboardUnitForNoQueryID = "NoQueryID"
	DashboardUnitWithQueryID  = "WithQueryID"
)

var DashboardUnitPresentations = [...]string{
	PresentationLine,
	PresentationBar,
	PresentationTable,
	PresentationCard,
	PresentationFunnel,
}

const (
	UnitCard  = "card"
	UnitChart = "chart"
)

var UnitPresentationTypes = [...]string{UnitCard, UnitChart}

func GetUnitType(presentation string) string {
	if presentation == PresentationCard {
		return UnitCard
	}

	return UnitChart
}
