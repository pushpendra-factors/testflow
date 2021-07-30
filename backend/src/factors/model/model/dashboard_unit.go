package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
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

// DashboardUnitCachePayload Payload for dashboard caching method.
type DashboardUnitCachePayload struct {
	DashboardUnit DashboardUnit
	BaseQuery     BaseQuery
}

type BeamDashboardUnitCachePayload struct {
	DashboardUnit DashboardUnit
	QueryClass    string
	Query         postgres.Jsonb
	From, To      int64
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
	PresentationStack,
	PresentationArea,
	PresentationHorizontalBar,
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

func IsValidDashboardUnit(dashboardUnit *DashboardUnit) (bool, string) {
	if dashboardUnit.DashboardId == 0 {
		return false, "Invalid dashboard"
	}

	if dashboardUnit.Title == "" {
		return false, "Invalid title"
	}

	validPresentation := false
	for _, p := range DashboardUnitPresentations {
		if p == dashboardUnit.Presentation {
			validPresentation = true
			break
		}
	}
	if !validPresentation {
		return false, "Invalid presentation"
	}
	// Todo(Dinesh): Validate query based on query class here.
	return true, ""
}

func ShouldCacheUnitForTimeRange(queryClass, preset string, from, to int64, onlyAttribution, skipAttribution int) (bool, int64, int64) {

	if queryClass == QueryClassAttribution {
		// Rule 1: Skip attribution class queries if skipAttribution = 1
		if skipAttribution == 1 {
			return false, 0, 0
		}
	} else {
		// Rule 2: Skip other class queries if onlyAttribution = 1
		if onlyAttribution == 1 {
			return false, 0, 0
		}
	}
	// Using one minute as a buffer time in taking out difference
	epsilonSeconds := int64(60)
	if queryClass == QueryClassAttribution {

		// Rule 2: Skip for Today or Yesterday for attribution class queries
		if preset == U.DateRangePresetToday || preset == U.DateRangePresetYesterday {
			return false, 0, 0
		}

		if preset == U.DateRangePresetLastWeek || preset == U.DateRangePresetLastMonth {
			// Rule 2': If last week/last month is well before one day in past, compute for entire range
			now := U.TimeNow().Unix()
			if (to + U.SECONDS_IN_A_DAY) <= (now - epsilonSeconds) {
				return true, from, to
			}
		}

		// Cases for this week, this month, last week, last month with no complete data
		// For other presets, we skip computing yesterday's data, hence effective to = to - SECONDS_IN_A_DAY
		effectiveTo := to - U.SECONDS_IN_A_DAY

		// Rule 3: If the computation data is Equal to or more than 1 day, we should run attribution else skip it.
		// Since From/To is always start/end of the day, if the diff is > 0, it is effectively a day
		if (effectiveTo - from) > 0 {
			return true, from, effectiveTo
		}

		// Cases for Sunday, Monday of ThisWeek & 1st, 2nd of ThisMonth, it shouldn't cache.
		return false, 0, 0
	}

	// For other units, caching should run without changing `to`.
	return true, from, to
}

func GetEffectiveTimeRangeForDashboardUnitAttributionQuery(from, to int64) (int64, int64) {

	epsilonSeconds := int64(60)
	now := U.TimeNow().Unix()
	if (to + U.SECONDS_IN_A_DAY) <= (now - epsilonSeconds) {
		return from, to
	}

	effectiveTo := to - U.SECONDS_IN_A_DAY
	// Since From/To is always start/end of the day, if the diff is > 0, it is effectively a day
	if (effectiveTo - from) > 0 {
		return from, effectiveTo
	}
	return 0, 0
}
