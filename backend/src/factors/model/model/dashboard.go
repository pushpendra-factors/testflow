package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"

	"strconv"
	"strings"
)

type Dashboard struct {
	// Composite primary key, id + project_id + agent_id.
	ID int64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboards(project_id) ref projects(id).
	ProjectId     int64           `gorm:"primary_key:true" json:"project_id"`
	AgentUUID     string          `gorm:"primary_key:true" json:"-"`
	Name          string          `gorm:"not null" json:"name"`
	Description   string          `json:"description"`
	Type          string          `gorm:"type:varchar(5);not null" json:"type"`
	Settings      postgres.Jsonb  `json:"settings"`
	Class         string          `json:"class"`
	UnitsPosition *postgres.Jsonb `json:"units_position"` // map[string]map[uint64]int -> map[unit_type]unit_id:unit_position
	IsDeleted     bool            `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type DashboardString struct {
	// Composite primary key, id + project_id + agent_id.
	ID string `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboards(project_id) ref projects(id).
	ProjectId     int64           `gorm:"primary_key:true" json:"project_id"`
	AgentUUID     string          `gorm:"primary_key:true" json:"-"`
	Name          string          `gorm:"not null" json:"name"`
	Description   string          `json:"description"`
	Type          string          `gorm:"type:varchar(5);not null" json:"type"`
	Settings      postgres.Jsonb  `json:"settings"`
	Class         string          `json:"class"`
	UnitsPosition *postgres.Jsonb `json:"units_position"` // map[string]map[uint64]int -> map[unit_type]unit_id:unit_position
	IsDeleted     bool            `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type UpdatableDashboard struct {
	Name          string                    `json:"name"`
	Type          string                    `json:"type"`
	Description   string                    `json:"description"`
	UnitsPosition *map[string]map[int64]int `json:"units_position"`
	Settings      *postgres.Jsonb           `json:"settings"`
}

type DashboardCacheResult struct {
	Result      interface{} `json:"result"`
	RefreshedAt int64       `json:"refreshed_at"`
	CacheMeta   interface{} `json:"cache_meta"`
	Timezone    string      `json:"timezone"`
}

type DashQueryResult struct {
	// Composite primary key, id + project_id.
	ID              string         `gorm:"not null" json:"id"`
	ProjectID       int64          `gorm:"not null" json:"project_id"`
	DashboardID     int64          `gorm:"not null" json:"dashboard_id"`
	DashboardUnitID int64          `gorm:"not null" json:"dashboard_unit_id"`
	QueryID         int64          `gorm:"not null" json:"query_id"`
	FromT           int64          `gorm:"not null" json:"from_t"`
	ToT             int64          `gorm:"not null" json:"to_t"`
	Result          postgres.Jsonb `gorm:"not null" json:"result"`
	IsDeleted       bool           `gorm:"not null;default:false" json:"is_deleted"`
	ComputedAt      int64          `gorm:"not null" json:"computed_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

const DashboardCachingDurationInSeconds = 32 * 24 * 60 * 60              // 32 days.
const DashboardCacheInvalidationDuration14DaysInSecs = 14 * 24 * 60 * 60 // 14 days.
const MaxNumberOfDashboardUnitCacheAccessedIn14Days = 50000
const MaxNumberPerScanCount = 50000

const (
	DashboardTypePrivate        = "pr"
	DashboardTypeProjectVisible = "pv"
	DashboardTypeAttributionV1  = "av1"

	DashboardClassUserCreated      = "user_created"
	DashboardClassWebsiteAnalytics = "web"

	AttributionV1Description = ""
	AttributionV1Name        = "Attribution Reporting"
)

var DashboardTypes = []string{DashboardTypePrivate, DashboardTypeProjectVisible, DashboardTypeAttributionV1}

const AgentProjectPersonalDashboardName = "My Dashboard"
const AgentProjectPersonalDashboardDescription = "No Description"

func GetCacheResultByDashboardIdAndUnitId(reqId string, projectId int64, dashboardId, unitId int64, from, to int64, timezoneString U.TimeZoneString) (*DashboardCacheResult, int, error) {
	var cacheResult *DashboardCacheResult

	logCtx := log.WithFields(log.Fields{
		"reqId":    reqId,
		"Method":   "GetCacheResultByDashboardIdAndUnitId",
		"CacheKey": fmt.Sprintf("PID:%d:DID:%d:DUID:%d", projectId, dashboardId, unitId),
	})
	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		return cacheResult, http.StatusBadRequest, errors.New("invalid scope ids")
	}

	cacheKey, err := GetDashboardUnitQueryResultCacheKey(projectId, dashboardId, unitId, from, to, timezoneString)

	if err != nil {
		return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to fetch cache key - " + err.Error())
	}
	result, status, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if !status {
		if err != nil {
			return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to get data from cache - " + err.Error())
		}

		return cacheResult, http.StatusNotFound, nil
	}
	err = json.Unmarshal([]byte(result), &cacheResult)

	if err != nil {
		logCtx.WithError(err).Errorf("Error decoding redis result %v", result)
		return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to unmarshal cache response - " + err.Error())
	}

	if cacheResult.RefreshedAt == 0 {
		cacheResult.RefreshedAt = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	}
	return cacheResult, http.StatusFound, nil
}

func GetCacheResultByDashboardIdAndUnitIdWithPreset(reqId string, projectId int64, dashboardId, unitId int64, preset string, from, to int64, timezoneString U.TimeZoneString) (*DashboardCacheResult, int, error) {
	var cacheResult *DashboardCacheResult

	logCtx := log.WithFields(log.Fields{
		"reqId":  reqId,
		"Method": "GetCacheResultByDashboardIdAndUnitId",
		"preset": preset, "from": from, "to": to,
		"CacheKey": fmt.Sprintf("PID:%d:DID:%d:DUID:%d:PRESET:%s", projectId, dashboardId, unitId, preset),
	})
	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		return cacheResult, http.StatusBadRequest, errors.New("invalid scope ids")
	}

	cacheKey, err := GetDashboardUnitQueryLastComputedResultCacheKey(projectId, dashboardId, unitId, preset, from, to, timezoneString)
	if err != nil {
		return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to fetch cache key - " + err.Error())
	}
	result, status, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if !status {
		if err != nil {
			return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to get data from cache - " + err.Error())
		}

		return cacheResult, http.StatusNotFound, nil
	}
	err = json.Unmarshal([]byte(result), &cacheResult)

	if err != nil {
		logCtx.WithError(err).Errorf("Error decoding redis result %v", result)
		return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to unmarshal cache response - " + err.Error())
	}

	if cacheResult.RefreshedAt == 0 {
		cacheResult.RefreshedAt = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	}
	return cacheResult, http.StatusFound, nil
}

func SetCacheResultByDashboardIdAndUnitIdWithPreset(result interface{}, projectId int64, dashboardId int64, unitId int64, preset string, from, to int64, timezoneString U.TimeZoneString, meta interface{}) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"dashboard_id": dashboardId, "dashboard_unit_id": unitId,
		"preset": preset, "from": from, "to": to,
	})

	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		logCtx.Error("Invalid scope ids.")
		return
	}

	cacheKey, err := GetDashboardUnitQueryResultCacheKeyWithPreset(projectId, dashboardId, unitId, preset, from, to, timezoneString)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}
	if meta == nil {
		meta = CacheMeta{
			From:           from,
			To:             to,
			RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
			Timezone:       string(timezoneString),
			Preset:         preset,
			LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		}
	}

	dashboardCacheResult := DashboardCacheResult{
		Result:      result,
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Timezone:    string(timezoneString),
		CacheMeta:   meta,
	}

	enDashboardCacheResult, err := json.Marshal(dashboardCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult.")
		return
	}

	expiryInSecs := float64(0)
	if C.IsProjectAllowedForLongerExpiry(projectId) {
		// Approx 3 months for any query less than 3 months
		if to-from < (15 * U.SECONDS_IN_A_DAY) {
			expiryInSecs = float64(92 * U.SECONDS_IN_A_DAY)
		}
		// Approx 1 year for any query more than a month
		if to-from > (27 * U.SECONDS_IN_A_DAY) {
			expiryInSecs = float64(365 * U.SECONDS_IN_A_DAY)
		}
		// for anything between 15
		expiryInSecs = float64(U.CacheExpiryDefaultInSeconds)
	} else {
		expiryInSecs = U.GetDashboardCacheResultExpiryInSeconds(from, to, timezoneString)
	}

	err = cacheRedis.SetPersistent(cacheKey, string(enDashboardCacheResult), expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for channel query")
		return
	}
}

func SetCacheResultByDashboardIdAndUnitId(result interface{}, projectId int64, dashboardId int64, unitId int64, from, to int64, timezoneString U.TimeZoneString, meta interface{}) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"dashboard_id": dashboardId, "dashboard_unit_id": unitId,
	})

	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		logCtx.Error("Invalid scope ids.")
		return
	}

	cacheKey, err := GetDashboardUnitQueryResultCacheKey(projectId, dashboardId, unitId, from, to, timezoneString)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}
	if meta == nil {
		meta = CacheMeta{
			From:           from,
			To:             to,
			RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
			Timezone:       string(timezoneString),
			Preset:         "",
			LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		}
	}

	dashboardCacheResult := DashboardCacheResult{
		Result:      result,
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Timezone:    string(timezoneString),
		CacheMeta:   meta,
	}

	enDashboardCacheResult, err := json.Marshal(dashboardCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult.")
		return
	}

	err = cacheRedis.SetPersistent(cacheKey, string(enDashboardCacheResult), U.GetDashboardCacheResultExpiryInSeconds(from, to, timezoneString))
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for channel query")
		return
	}
}

// GetDashboardCacheAnalyticsValidityMap returns a map of all ProjectID-dashboardunitID pairs that have been accessed in the last 14 days
func GetDashboardCacheAnalyticsValidityMap() (map[int64]map[int64]bool, int64, error) {
	logCtx := log.WithFields(log.Fields{"method": "GetDashboardCacheAnalyticsValidityMap"})

	cacheKeys, err := cacheRedis.ScanPersistent("dashboard:analytics:*", MaxNumberOfDashboardUnitCacheAccessedIn14Days, MaxNumberOfDashboardUnitCacheAccessedIn14Days)

	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return nil, 0, err
	}

	mapOfValidDashboardUnits := map[int64]map[int64]bool{}
	totalValidUnits := int64(0)
	for _, cacheKey := range cacheKeys {
		var cacheResult *DashboardCacheResult
		result, status, err := cacheRedis.GetIfExistsPersistent(cacheKey)
		if !status && err != nil {
			logCtx.WithError(err).Error("Failed to get result from cache for key %v", cacheKey)
			continue
		}
		err = json.Unmarshal([]byte(result), &cacheResult)
		if err != nil {
			logCtx.WithError(err).Errorf("Error decoding redis result %v for key %v", result, cacheKey)
			continue
		}
		projectId := cacheKey.ProjectID
		if _, exists := mapOfValidDashboardUnits[projectId]; !exists {
			mapOfValidDashboardUnits[projectId] = map[int64]bool{}
		}
		keyValues := strings.Split(cacheKey.Suffix, ":")
		dashboardUnitId, _ := strconv.ParseInt(keyValues[3], 10, 64)
		timeDifference := U.TimeNowIn(U.TimeZoneStringIST).Unix() - cacheResult.RefreshedAt

		if timeDifference < DashboardCacheInvalidationDuration14DaysInSecs && timeDifference >= 0 {
			mapOfValidDashboardUnits[projectId][dashboardUnitId] = true
			totalValidUnits = totalValidUnits + 1
		}
	}
	logCtx.WithFields(log.Fields{"total_valid_units": totalValidUnits}).Info("Total valid unit count")
	return mapOfValidDashboardUnits, totalValidUnits, err
}

// SetDashboardCacheAnalytics Sets the result in cache after generating a cacheKey to store against
func SetDashboardCacheAnalytics(projectId int64, dashboardId int64, unitId int64, from, to int64, timezoneString U.TimeZoneString) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"dashboard_id": dashboardId, "dashboard_unit_id": unitId,
		"from": from, "to": to,
	})

	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		logCtx.Error("Invalid scope ids.")
		return
	}
	preset := U.GetPresetNameByFromAndTo(from, to, timezoneString)

	cacheKey, err := GetDashboardCacheAnalyticsCacheKey(projectId, dashboardId, unitId, from, to, timezoneString, preset)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}

	dashboardCacheResult := DashboardCacheResult{
		Result: nil,
		CacheMeta: CacheMeta{
			Preset:         preset,
			From:           from,
			To:             to,
			RefreshedAt:    U.TimeNowIn(timezoneString).Unix(),
			LastComputedAt: U.TimeNowIn(timezoneString).Unix(),
			Timezone:       string(timezoneString),
		},
		Timezone: string(timezoneString),

		RefreshedAt: U.TimeNowIn(timezoneString).Unix(), // This represents the time when dashboard unit ID was requested
	}

	enDashboardCacheResult, err := json.Marshal(dashboardCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult.")
		return
	}

	err = cacheRedis.SetPersistent(cacheKey, string(enDashboardCacheResult), DashboardCacheInvalidationDuration14DaysInSecs)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for dashboard query")
		return
	}
}

// ShouldRefreshDashboardUnit Whether to force refresh dashboard unit irrespective of the cache and expiry.
func ShouldRefreshDashboardUnit(projectID int64, dashboardID, dashboardUnitID int64, from, to int64, timezoneString U.TimeZoneString, isWebAnalytics bool) bool {

	// since all the ranges are pre-defined & back dated, skip all checks for Longer Expiry
	if C.IsProjectAllowedForLongerExpiry(projectID) {
		return true
	}

	// If today's range or last 30 minutes window, refresh on every trigger.
	if U.IsStartOfTodaysRangeIn(from, timezoneString) || U.Is30MinutesTimeRange(from, to) {
		return true
	}

	var refreshedAt int64
	if isWebAnalytics {
		result, errCode := GetCacheResultForWebAnalyticsDashboard(projectID, dashboardID, from, to, timezoneString)
		if errCode != http.StatusFound {
			return true
		}
		refreshedAt = result.RefreshedAt
	} else {
		result, errCode, _ := GetCacheResultByDashboardIdAndUnitId("", projectID, dashboardID, dashboardUnitID, from, to, timezoneString)
		if errCode != http.StatusFound || result == nil {
			return true
		}
		refreshedAt = result.RefreshedAt
	}

	// Needs change.
	toStartOfDay := U.GetBeginningOfDayTimestampIn(to, timezoneString)
	nowStartOfDay := U.GetBeginningOfDayTimestampIn(U.TimeNowUnix(), timezoneString)

	// If in last 2 days (mutable data range), check the RefreshedAt in Cache result.
	// Skip, if RefreshedAt is for the same day to restrict force cache only once a day.
	if nowStartOfDay > toStartOfDay && nowStartOfDay-toStartOfDay <= U.ImmutableDataEndDateBufferInSeconds {
		refreshedAtDate := U.GetDateAsStringIn(refreshedAt, timezoneString)
		todaysDate := U.GetDateAsStringIn(U.TimeNowUnix(), timezoneString)
		if refreshedAtDate < todaysDate {
			return true
		}
	}
	return false
}
