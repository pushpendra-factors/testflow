package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"sort"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type DashboardUnit struct {
	// Composite primary key, id + project_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboard_units(project_id) ref projects(id).
	ProjectID    uint64    `gorm:"primary_key:true" json:"project_id"`
	DashboardId  uint64    `gorm:"primary_key:true" json:"dashboard_id"`
	Description  string    `json:"description"`
	Presentation string    `gorm:"type:varchar(5);not null" json:"presentation"`
	IsDeleted    bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	QueryId      uint64    `gorm:"not null" json:"query_id"`
}

type DashboardUnitRequestPayload struct {
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	QueryId      uint64 `json:"query_id"`
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
	TimeZone      U.TimeZoneString
}

const (
	CachingUnitNormal            = 0
	CachingUnitWebAnalytics      = 1
	CachingUnitStatusFailed      = -1
	CachingUnitStatusTimeout     = -2
	CachingUnitStatusNotComputed = 0
	CachingUnitStatusPassed      = 1
)

const QueryNotFoundError = "Failed to fetch query from query_id"

type CachingUnitReport struct {
	UnitType     int // CachingUnitNormal=1 or CachingUnitWebAnalytics=1
	ProjectId    uint64
	DashboardID  uint64
	UnitID       uint64
	QueryID      uint64
	QueryClass   string
	Query        interface{}
	From, To     int64
	QueryRange   string
	Status       int // CachingUnitStatusFailed=-1 or CachingUnitStatusNotComputed=0 or CachingUnitStatusPassed=1
	TimeTaken    int64
	TimeTakenStr string
}

func GetCachingUnitReportUniqueKey(report CachingUnitReport) string {
	return fmt.Sprintf("%v-%d-%d-%d-%v-%d-%d", report.UnitType, report.ProjectId, report.DashboardID, report.UnitID,
		report.QueryClass, report.From, report.To)
}

type CachingProjectReport struct {
	ProjectId    uint64
	ProjectName  string
	TotalRuntime string
}

type FailedDashboardUnitReport struct {
	DashboardID uint64
	UnitID      uint64
	QueryClass  string
	QueryRange  string
}

func getDashboardUnitQueryResultCacheKey(projectID, dashboardID, unitID uint64, from, to int64, timezoneString U.TimeZoneString) (*cacheRedis.Key, error) {
	prefix := "dashboard:query"
	var suffix string
	if U.IsStartOfTodaysRangeIn(from, timezoneString) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:now", dashboardID, unitID, from)
	} else {
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:%d", dashboardID, unitID, from, to)
	}
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

func getDashboardCacheAnalyticsCacheKey(projectID, dashboardID, unitID uint64, from, to int64, timezoneString U.TimeZoneString, preset string) (*cacheRedis.Key, error) {
	prefix := "dashboard:analytics"
	var suffix string
	if U.IsStartOfTodaysRangeIn(from, timezoneString) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:now:preset:%v", dashboardID, unitID, from, preset)
	} else {
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:%d:preset:%v", dashboardID, unitID, from, to, preset)
	}
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

var DashboardUnitPresentations = [...]string{
	HorizontalBar,
	PresentationScatterPlot,
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
	if dashboardUnit.Presentation != "" {
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

	}

	// Todo(Dinesh): Validate query based on query class here.
	return true, ""
}

func GetNSlowestUnits(cacheReports []CachingUnitReport, n int) []CachingUnitReport {

	var units []CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	return units[0:U.MinInt(n, len(units))]
}

func GetTotalFailedComputedNotComputed(cacheReports []CachingUnitReport) (int, int, int) {

	statusFailed := 0
	statusNotComputed := 0
	statusPassed := 0

	for _, unit := range cacheReports {
		switch unit.Status {
		case CachingUnitStatusFailed:
			statusFailed++
		case CachingUnitStatusPassed:
			statusPassed++
		case CachingUnitStatusNotComputed:
			statusNotComputed++
		}
	}
	return statusFailed, statusPassed, statusNotComputed
}

func GetFailedUnitsByProject(cacheReports []CachingUnitReport) map[uint64][]FailedDashboardUnitReport {

	var units []CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectFailedUnits := make(map[uint64][]FailedDashboardUnitReport)
	for _, unit := range cacheReports {
		if unit.Status == CachingUnitStatusFailed {
			failedUnit := FailedDashboardUnitReport{
				DashboardID: unit.DashboardID,
				UnitID:      unit.UnitID,
				QueryClass:  unit.QueryClass,
				QueryRange:  unit.QueryRange,
			}
			if value, exists := projectFailedUnits[unit.ProjectId]; exists {
				projectFailedUnits[unit.ProjectId] = append(value, failedUnit)
			} else {
				failedUnits := []FailedDashboardUnitReport{failedUnit}
				projectFailedUnits[unit.ProjectId] = failedUnits
			}
		}
	}
	return projectFailedUnits
}

func GetTimedOutUnitsByProject(cacheReports []CachingUnitReport) map[uint64][]FailedDashboardUnitReport {

	var units []CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectTimedOutUnits := make(map[uint64][]FailedDashboardUnitReport)
	for _, unit := range cacheReports {
		if unit.Status == CachingUnitStatusTimeout {
			timedOutUnit := FailedDashboardUnitReport{
				DashboardID: unit.DashboardID,
				UnitID:      unit.UnitID,
				QueryClass:  unit.QueryClass,
				QueryRange:  unit.QueryRange,
			}
			if value, exists := projectTimedOutUnits[unit.ProjectId]; exists {
				projectTimedOutUnits[unit.ProjectId] = append(value, timedOutUnit)
			} else {
				failedUnits := []FailedDashboardUnitReport{timedOutUnit}
				projectTimedOutUnits[unit.ProjectId] = failedUnits
			}
		}
	}
	return projectTimedOutUnits
}

func GetNSlowestProjects(cacheReports []CachingUnitReport, n int) []CachingProjectReport {

	var units []CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectTotalTime := make(map[uint64]int64)
	for _, unit := range cacheReports {
		projectTotalTime[unit.ProjectId] = projectTotalTime[unit.ProjectId] + unit.TimeTaken
	}

	var projects []CachingProjectReport
	for key, value := range projectTotalTime {
		projects = append(projects, CachingProjectReport{ProjectId: key,
			ProjectName: "", TotalRuntime: U.SecondsToHMSString(value)})
	}
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].TotalRuntime > projects[j].TotalRuntime
	})

	return projects[0:U.MinInt(n, len(projects))]

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
			now := time.Now().Unix()
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
	now := time.Now().Unix()
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
