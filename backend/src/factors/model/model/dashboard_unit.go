package model

import (
	"errors"
	cacheRedis "factors/cache/redis"
	"factors/model/store"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type DashboardUnit struct {
	// Composite primary key, id + project_id.
	ID int64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboard_units(project_id) ref projects(id).
	ProjectID    int64     `gorm:"primary_key:true" json:"project_id"`
	DashboardId  int64     `gorm:"primary_key:true" json:"dashboard_id"`
	Description  string    `json:"description"`
	Presentation string    `gorm:"type:varchar(5);not null" json:"presentation"`
	IsDeleted    bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	QueryId      int64     `gorm:"not null" json:"query_id"`
}

type DashboardUnitString struct {
	// Composite primary key, id + project_id.
	ID string `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboard_units(project_id) ref projects(id).
	ProjectID    int64     `gorm:"primary_key:true" json:"project_id"`
	DashboardId  string    `gorm:"primary_key:true" json:"dashboard_id"`
	Description  string    `json:"description"`
	Presentation string    `gorm:"type:varchar(5);not null" json:"presentation"`
	IsDeleted    bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	QueryId      string    `gorm:"not null" json:"query_id"`
}

type DashboardUnitRequestPayload struct {
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	QueryId      int64  `json:"query_id,string"`
}

type DashboardUnitRequestPayloadString struct {
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	QueryId      string `json:"query_id"`
}

// DashboardUnitCachePayload Payload for dashboard caching method.
type DashboardUnitCachePayload struct {
	DashboardUnit DashboardUnit
	BaseQuery     BaseQuery
	Preset        string
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
	ProjectId    int64
	DashboardID  int64
	UnitID       int64
	QueryID      int64
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
	ProjectId    int64
	ProjectName  string
	TotalRuntime string
}

type FailedDashboardUnitReport struct {
	DashboardID int64
	UnitID      int64
	QueryClass  string
	QueryRange  string
	From        string
	To          string
	Preset      string
}

func GetDashboardUnitQueryResultCacheKeyWithPreset(projectID int64, dashboardID, unitID int64, preset string, from, to int64, timezoneString U.TimeZoneString) (*cacheRedis.Key, error) {
	prefix := "dashboard:query"

	var suffix string

	if U.IsStartOfTodaysRangeIn(from, timezoneString) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:now", dashboardID, unitID, from)
	} else {
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:%d", dashboardID, unitID, from, to)
	}
	if preset != "" {
		s1 := strings.Split(suffix, ":from")
		a := fmt.Sprintf(":preset:%s:from", preset)
		if len(s1) < 2 {
			return nil, errors.New("invalid cache key")
		}
		suffix = s1[0] + a + s1[1]
	}

	return cacheRedis.NewKey(projectID, prefix, suffix)
}

func GetDashboardUnitQueryResultCacheKey(projectID int64, dashboardID, unitID int64, from, to int64, timezoneString U.TimeZoneString) (*cacheRedis.Key, error) {
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

func GetDashboardCacheAnalyticsCacheKey(projectID int64, dashboardID, unitID int64, from, to int64, timezoneString U.TimeZoneString, preset string) (*cacheRedis.Key, error) {
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

var SearchKeyPreset = map[string][]string{
	"CURRENT_WEEK":  {"CURRENT_WEEK"},
	"LAST_WEEK":     {"LAST_WEEK", "CURRENT_WEEK"},
	"CURRENT_MONTH": {"CURRENT_MONTH"},
	"LAST_MONTH":    {"LAST_MONTH", "CURRENT_MONTH"},
	"YESTERDAY":     {"YESTERDAY", "TODAY"},
	"TODAY":         {"TODAY"},
}

// GetDashboardUnitQueryLastComputedResultCacheKey return last computed cachekey
func GetDashboardUnitQueryLastComputedResultCacheKey(projectID int64, dashboardID, unitID int64, preset string, from, to int64, timezoneString U.TimeZoneString) (*cacheRedis.Key, error) {

	logCtx := log.WithFields(log.Fields{
		"CacheKey": fmt.Sprintf("PID:%d:DID:%d:DUID:%d:PRESET:%s", projectID, dashboardID, unitID, preset),
	})
	logCtx.Info("fetching Last computed")
	var cacheKeys []*cacheRedis.Key
	var err error

	for _, pre := range SearchKeyPreset[preset] {
		pattern := fmt.Sprintf("dashboard:query:pid:%d:did:%d:duid:%d:preset:%s:from:%d:to:*", projectID, dashboardID, unitID, pre, from)
		cacheKey, err := cacheRedis.Scan(pattern, MaxNumberPerScanCount, MaxNumberPerScanCount)
		cacheKeys = append(cacheKeys, cacheKey...)
		if err != nil {
			log.WithError(err).Error("Failed to get cache key")

		}
	}

	var latestComputedAt int64 = 0
	var latestComputedKey *cacheRedis.Key

	for _, key := range cacheKeys {
		//get latest
		stringKey, _ := key.Key()
		LatestTo := strings.Split(stringKey, ":to:")
		latestTo, err := strconv.ParseInt(LatestTo[1], 10, 64)

		if err != nil {
			continue
		}
		_, limit, _ := U.QueryDateRangePresets[preset](timezoneString)
		if latestTo > limit {
			continue
		}
		if latestComputedAt < latestTo {
			latestComputedAt = latestTo
			latestComputedKey = key
		}

	}

	log.WithFields(log.Fields{"latest_key": latestComputedKey, "len_cacheKeys": len(cacheKeys)}).Info("Last computed cache key")

	if latestComputedKey == nil {
		queryStartTime := time.Now().UTC().Unix()
		cacheKey, err := GetDashboardUnitQueryResultCacheKeyWithPreset(projectID, dashboardID, unitID, preset, from, to, timezoneString)
		log.WithFields(log.Fields{"preset": preset}).Info("Failed to find cache key")
		logCtx.WithFields(log.Fields{"TimePassedInMins": float64(time.Now().UTC().Unix()-queryStartTime) / 60}).Info("218-find latest key took time")

		return cacheKey, err
	}

	return latestComputedKey, err
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

func GetFailedUnitsByProject(cacheReports []CachingUnitReport) map[int64][]FailedDashboardUnitReport {

	var units []CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectFailedUnits := make(map[int64][]FailedDashboardUnitReport)
	for _, unit := range cacheReports {
		timezone, _ := store.GetStore().GetTimezoneForProject(unit.ProjectId)
		if unit.Status == CachingUnitStatusFailed {
			failedUnit := FailedDashboardUnitReport{
				DashboardID: unit.DashboardID,
				UnitID:      unit.UnitID,
				QueryClass:  unit.QueryClass,
				QueryRange:  unit.QueryRange,
				From:        U.GetDateOnlyFormatFromTimestampAndTimezone(unit.From, timezone),
				To:          U.GetDateOnlyFormatFromTimestampAndTimezone(unit.To, timezone),
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

func GetTimedOutUnitsByProject(cacheReports []CachingUnitReport) map[int64][]FailedDashboardUnitReport {

	var units []CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectTimedOutUnits := make(map[int64][]FailedDashboardUnitReport)
	for _, unit := range cacheReports {
		timezone, _ := store.GetStore().GetTimezoneForProject(unit.ProjectId)
		if unit.Status == CachingUnitStatusTimeout {
			timedOutUnit := FailedDashboardUnitReport{
				DashboardID: unit.DashboardID,
				UnitID:      unit.UnitID,
				QueryClass:  unit.QueryClass,
				QueryRange:  unit.QueryRange,
				From:        U.GetDateOnlyFormatFromTimestampAndTimezone(unit.From, timezone),
				To:          U.GetDateOnlyFormatFromTimestampAndTimezone(unit.To, timezone),
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

	projectTotalTime := make(map[int64]int64)
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

func ShouldCacheUnitForTimeRange(queryClass, preset string, from, to int64, onlyAttribution, skipAttribution int, skipPreset bool) (bool, int64, int64) {

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

		if preset == U.DateRangePresetYesterday && !skipPreset {
			return true, from, to
		}

		if (preset == U.DateRangePresetLastWeek || preset == U.DateRangePresetLastMonth) && !skipPreset {
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
