package model

import (
	"encoding/json"
	"errors"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	cacheRedis "factors/cache/redis"

	"github.com/gomodule/redigo/redis"
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
	RefreshedAt int64       `json:"refreshed_at"`
}

const DashboardCachingDurationInSeconds = 32 * 24 * 60 * 60 // 32 days.

const (
	DashboardTypePrivate        = "pr"
	DashboardTypeProjectVisible = "pv"
)

var DashboardTypes = []string{DashboardTypePrivate, DashboardTypeProjectVisible}

const AgentProjectPersonalDashboardName = "My Dashboard"
const AgentProjectPersonalDashboardDescription = "No Description"

func GetCacheResultByDashboardIdAndUnitId(projectId, dashboardId, unitId uint64, from, to int64) (*DashboardCacheResult, int, error) {
	var cacheResult *DashboardCacheResult
	logCtx := log.WithFields(log.Fields{
		"Method":   "GetCacheResultByDashboardIdAndUnitId",
		"CacheKey": fmt.Sprintf("PID:%d:DID:%d:DUID:%d", projectId, dashboardId, unitId),
	})
	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		return cacheResult, http.StatusBadRequest, errors.New("invalid scope ids")
	}

	cacheKey, err := getDashboardUnitQueryResultCacheKey(projectId, dashboardId, unitId, from, to)
	if err != nil {
		return cacheResult, http.StatusBadRequest, err
	}

	result, err := cacheRedis.GetPersistent(cacheKey)
	if err == redis.ErrNil {
		return cacheResult, http.StatusNotFound, nil
	} else if err != nil {
		logCtx.WithError(err).Error("error doing Get from redis")
		return cacheResult, http.StatusInternalServerError, err
	}

	err = json.Unmarshal([]byte(result), &cacheResult)
	if err != nil {
		logCtx.WithError(err).Errorf("Error decoding redis result %v", result)
		return cacheResult, http.StatusInternalServerError, err
	}

	if cacheResult.RefreshedAt == 0 {
		cacheResult.RefreshedAt = U.TimeNowIn(U.TimeZoneStringIST).Unix()
	}
	return cacheResult, http.StatusFound, nil
}

func SetCacheResultByDashboardIdAndUnitId(result interface{}, projectId uint64, dashboardId uint64, unitId uint64, from, to int64) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"dashboard_id": dashboardId, "dashboard_unit_id": unitId,
	})

	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		logCtx.Error("Invalid scope ids.")
		return
	}

	cacheKey, err := getDashboardUnitQueryResultCacheKey(projectId, dashboardId, unitId, from, to)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}

	dashboardCacheResult := DashboardCacheResult{
		Result:      result,
		From:        from,
		To:          to,
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
	}

	enDashboardCacheResult, err := json.Marshal(dashboardCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult.")
		return
	}

	err = cacheRedis.SetPersistent(cacheKey, string(enDashboardCacheResult), U.GetDashboardCacheResultExpiryInSeconds(from, to))
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for channel query")
		return
	}
}
