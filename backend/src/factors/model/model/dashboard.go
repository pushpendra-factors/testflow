package model

import (
	"encoding/json"
	"errors"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	cacheRedis "factors/cache/redis"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

type Dashboard struct {
	// Composite primary key, id + project_id + agent_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboards(project_id) ref projects(id).
	ProjectId     uint64          `gorm:"primary_key:true" json:"project_id"`
	AgentUUID     string          `gorm:"primary_key:true" json:"-"`
	Name          string          `gorm:"not null" json:"name"`
	Description   string          `json:"description"`
	Type          string          `gorm:"type:varchar(5);not null" json:"type"`
	Class         string          `json:"class"`
	UnitsPosition *postgres.Jsonb `json:"units_position"` // map[string]map[uint64]int -> map[unit_type]unit_id:unit_position
	IsDeleted     bool            `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type UpdatableDashboard struct {
	Name          string                     `json:"name"`
	Type          string                     `json:"type"`
	Description   string                     `json:"description"`
	UnitsPosition *map[string]map[uint64]int `json:"units_position"`
}

type DashboardCacheResult struct {
	Result      interface{} `json:"result"`
	From        int64       `json:"from"`
	To          int64       `json:"tom"`
	Timezone    string      `json:"timezone"`
	RefreshedAt int64       `json:"refreshed_at"`
}

const DashboardCachingDurationInSeconds = 32 * 24 * 60 * 60 // 32 days.

const (
	DashboardTypePrivate        = "pr"
	DashboardTypeProjectVisible = "pv"

	DashboardClassUserCreated      = "user_created"
	DashboardClassWebsiteAnalytics = "web"
)

var DashboardTypes = []string{DashboardTypePrivate, DashboardTypeProjectVisible}

const AgentProjectPersonalDashboardName = "My Dashboard"
const AgentProjectPersonalDashboardDescription = "No Description"

func GetCacheResultByDashboardIdAndUnitId(reqId string, projectId, dashboardId, unitId uint64, from, to int64, timezoneString U.TimeZoneString) (*DashboardCacheResult, int, error) {
	var cacheResult *DashboardCacheResult
	logCtx := log.WithFields(log.Fields{
		"reqId":    reqId,
		"Method":   "GetCacheResultByDashboardIdAndUnitId",
		"CacheKey": fmt.Sprintf("PID:%d:DID:%d:DUID:%d", projectId, dashboardId, unitId),
	})
	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		return cacheResult, http.StatusBadRequest, errors.New("invalid scope ids")
	}

	cacheKey, err := getDashboardUnitQueryResultCacheKey(projectId, dashboardId, unitId, from, to, timezoneString)
	if err != nil {
		return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to fetch cache key - " + err.Error())
	}

	result, status, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if status == false {
		if err == nil {
			return cacheResult, http.StatusNotFound, nil
		}
		return cacheResult, http.StatusInternalServerError, errors.New("Dashboard Cache: Failed to get data from cache - " + err.Error())
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

func SetCacheResultByDashboardIdAndUnitId(result interface{}, projectId uint64, dashboardId uint64, unitId uint64, from, to int64, timezoneString U.TimeZoneString) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"dashboard_id": dashboardId, "dashboard_unit_id": unitId,
	})

	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		logCtx.Error("Invalid scope ids.")
		return
	}

	cacheKey, err := getDashboardUnitQueryResultCacheKey(projectId, dashboardId, unitId, from, to, timezoneString)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}

	dashboardCacheResult := DashboardCacheResult{
		Result:      result,
		From:        from,
		To:          to,
		Timezone:    string(timezoneString),
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
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

// ShouldRefreshDashboardUnit Whether to force refresh dashboard unit irrespective of the cache and expiry.
func ShouldRefreshDashboardUnit(projectID, dashboardID, dashboardUnitID uint64, from, to int64, timezoneString U.TimeZoneString, isWebAnalytics bool) bool {
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
