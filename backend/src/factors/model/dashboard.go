package model

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
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

const (
	DashboardTypePrivate        = "pr"
	DashboardTypeProjectVisible = "pv"
)

var types = []string{DashboardTypePrivate, DashboardTypeProjectVisible}

const AgentProjectPersonalDashboardName = "My Dashboard"
const AgentProjectPersonalDashboardDescription = "No Description"

func isValidDashboard(dashboard *Dashboard) bool {
	if dashboard.Name == "" {
		return false
	}

	validType := false
	for _, t := range types {
		if t == dashboard.Type {
			validType = true
			break
		}
	}

	return validType
}

func CreateDashboard(projectId uint64, agentUUID string, dashboard *Dashboard) (*Dashboard, int) {
	db := C.GetServices().Db

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	if !isValidDashboard(dashboard) {
		return nil, http.StatusBadRequest
	}

	dashboard.ProjectId = projectId
	dashboard.AgentUUID = agentUUID

	if err := db.Create(dashboard).Error; err != nil {
		log.WithFields(log.Fields{"dashboard": dashboard,
			"project_id": projectId}).WithError(err).Error("Failed to create dashboard.")
		return nil, http.StatusInternalServerError
	}

	return dashboard, http.StatusCreated
}

func CreateAgentPersonalDashboardForProject(projectId uint64, agentUUID string) (*Dashboard, int) {
	return CreateDashboard(projectId, agentUUID,
		&Dashboard{Name: AgentProjectPersonalDashboardName, Description: AgentProjectPersonalDashboardDescription, Type: DashboardTypePrivate})
}

func GetDashboards(projectId uint64, agentUUID string) ([]Dashboard, int) {
	db := C.GetServices().Db

	var dashboards []Dashboard
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboards. Invalid project_id.")
		return dashboards, http.StatusBadRequest
	}

	err := db.Order("created_at ASC").Where("project_id = ? AND (type = ? OR agent_uuid = ?) AND is_deleted = ?",
		projectId, DashboardTypeProjectVisible, agentUUID, false).Find(&dashboards).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboards.")
		return dashboards, http.StatusInternalServerError
	}

	return dashboards, http.StatusFound
}

func GetDashboard(projectId uint64, agentUUID string, id uint64) (*Dashboard, int) {
	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "agentUUID": agentUUID})

	var dashboard Dashboard
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard. Invalid project_id or agent_id")
		return nil, http.StatusBadRequest
	}

	if err := db.Where("project_id = ? AND id = ? AND (type = ? OR agent_uuid = ?) AND is_deleted = ?", projectId, id,
		DashboardTypeProjectVisible, agentUUID, false).First(&dashboard).Error; err != nil {
		logCtx.WithError(err).WithField("dashboardID", id).Error(
			"Getting dashboard failed in GetDashboard")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	return &dashboard, http.StatusFound
}

// HasAccessToDashboard validates access to dashboard.
func HasAccessToDashboard(projectId uint64, agentUUID string, id uint64) (bool, *Dashboard) {
	dashboard, errCode := GetDashboard(projectId, agentUUID, id)
	if errCode != http.StatusFound {
		return false, nil
	}

	return true, dashboard
}

// Adds a position to the given unit on dashboard by unit_type.
func addUnitPositionOnDashboard(projectId uint64, agentUUID string,
	id uint64, unitId uint64, unitType string, currentUnitsPos *postgres.Jsonb) int {

	if projectId == 0 || agentUUID == "" || id == 0 || unitId == 0 {
		return http.StatusBadRequest
	}

	var currentPosition map[string]map[uint64]int
	newPos := 0
	if currentUnitsPos != nil {
		err := json.Unmarshal((*currentUnitsPos).RawMessage, &currentPosition)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "id": id,
				"unit_position": currentPosition}).WithError(err).Error("Failed decoding current units position.")
			return http.StatusInternalServerError
		}
	} else {
		currentPosition = make(map[string]map[uint64]int, 0)
	}

	if _, typeExists := currentPosition[unitType]; !typeExists {
		currentPosition[unitType] = make(map[uint64]int, 0)
	}

	maxPos := -1
	for _, pos := range currentPosition[unitType] {
		if pos > maxPos {
			maxPos = pos
		}
	}

	// if maxPos exists, increment the maxPos by one for newPos.
	// else start positions from 0.
	if maxPos > -1 {
		newPos = maxPos + 1
	}
	currentPosition[unitType][unitId] = newPos

	return UpdateDashboard(projectId, agentUUID, id, &UpdatableDashboard{UnitsPosition: &currentPosition})
}

func removeAndRebalanceUnitsPositionByType(positions *map[string]map[uint64]int,
	unitId uint64, unitType string) {

	removedPos := (*positions)[unitType][unitId]
	delete((*positions)[unitType], unitId)

	// reposition units positioned after removed unit.
	for id, pos := range (*positions)[unitType] {
		if pos > removedPos {
			(*positions)[unitType][id] = pos - 1
		}
	}
}

func removeUnitPositionOnDashboard(projectId uint64, agentUUID string,
	id uint64, unitId uint64, currentUnitsPos *postgres.Jsonb) int {

	if projectId == 0 || agentUUID == "" || id == 0 ||
		unitId == 0 || currentUnitsPos == nil {
		return http.StatusBadRequest
	}

	var currentPositions map[string]map[uint64]int
	err := json.Unmarshal((*currentUnitsPos).RawMessage, &currentPositions)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"unit_position": currentUnitsPos}).WithError(err).Error("Failed decoding current units position.")
		return http.StatusInternalServerError
	}

	var sourceUnitType string
	for typ := range currentPositions {
		for id := range currentPositions[typ] {
			if id == unitId {
				sourceUnitType = typ
				break
			}
		}
	}

	if sourceUnitType == "" {
		return http.StatusBadRequest
	}

	removeAndRebalanceUnitsPositionByType(&currentPositions, unitId, sourceUnitType)

	return UpdateDashboard(projectId, agentUUID, id, &UpdatableDashboard{UnitsPosition: &currentPositions})
}

func isValidUnitsPosition(positions *map[string]map[uint64]int) (bool, error) {
	if positions == nil {
		return false, errors.New("nil position map")
	}

	for _, typ := range UnitPresentationTypes {
		if posMap, exists := (*positions)[typ]; exists && len(posMap) > 0 {
			actualPos := make([]int, 0, 0)

			for _, pos := range posMap {
				actualPos = append(actualPos, pos)
			}

			// validates positions.
			sort.Sort(sort.IntSlice(actualPos))
			// sorted positions should be unique and incremented.
			for i := range actualPos {
				for futureIndex := i + 1; futureIndex < len(actualPos)-1; futureIndex++ {
					if actualPos[i] == actualPos[futureIndex] {
						return false, errors.New("duplicate position")
					}
				}
			}
		}
	}

	// Todo: Add duplicate id across different unit types.
	// Now frontend uses the position by existing dashboard units.
	// So no duplicates possible.

	return true, nil
}

func UpdateDashboard(projectId uint64, agentUUID string, id uint64, dashboard *UpdatableDashboard) int {
	if projectId == 0 || agentUUID == "" || id == 0 {
		log.Error("Failed to update dashboard. Invalid scope ids.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	// use HasAccessToDashboard maintain consistency on checking accessibility.
	if hasAccess, _ := HasAccessToDashboard(projectId, agentUUID, id); !hasAccess {
		// do not use http.StatusUnauthorised, breaks UI.
		return http.StatusForbidden
	}

	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)
	if dashboard.UnitsPosition != nil {
		logCtx := log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"positions": dashboard.UnitsPosition})

		if valid, err := isValidUnitsPosition(dashboard.UnitsPosition); !valid {
			logCtx.WithError(err).Error("Invalid units position.")
			return http.StatusBadRequest
		}

		currentPositionBytes, err := json.Marshal(dashboard.UnitsPosition)
		if err != nil {
			logCtx.WithError(err).Error("Failed to JSON encode new units position.")
			return http.StatusInternalServerError
		}
		currentPositionJsonb := &postgres.Jsonb{RawMessage: json.RawMessage(currentPositionBytes)}
		updateFields["units_position"] = currentPositionJsonb
	}

	if dashboard.Name != "" {
		updateFields["name"] = dashboard.Name
	}

	if dashboard.Description != "" {
		updateFields["description"] = dashboard.Description
	}

	if dashboard.Type != "" {
		updateFields["type"] = dashboard.Type
	}

	// nothing to update.
	if len(updateFields) == 0 {
		return http.StatusBadRequest
	}

	err := db.Model(&Dashboard{}).Where("project_id = ? AND id = ? AND is_deleted = ?", projectId, id, false).
		Update(updateFields).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"update": updateFields}).WithError(err).Error("Failed to update dashboard.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// DeleteDashboard To delete a dashboard by id.
func DeleteDashboard(projectID uint64, agentUUID string, dashboardID uint64) int {
	db := C.GetServices().Db

	if projectID == 0 || agentUUID == "" ||
		dashboardID == 0 {

		log.Error("Failed to delete dashboard. Invalid scope ids.")
		return http.StatusBadRequest
	}

	hasAccess, _ := HasAccessToDashboard(projectID, agentUUID, dashboardID)
	if !hasAccess {
		return http.StatusForbidden
	}

	dashboardUnits, errCode := GetDashboardUnits(projectID, agentUUID, dashboardID)
	if errCode != http.StatusFound {
		log.Error("failed to fetch dashboard units for delete dashboard")
		return http.StatusBadRequest
	}

	// Delete dashboard units for the the given dashboard first.
	for _, dashboardUnit := range dashboardUnits {
		errCode := deleteDashboardUnit(projectID, dashboardID, dashboardUnit.ID)
		if errCode != http.StatusAccepted {
			// continue
			log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID,
				"dashboard_uint_id": dashboardUnit.ID}).Error("failed to delete dashboard unit.")
		}
	}

	// Removing any reports attached with this dashboard id.
	errCode = DeleteReportByDashboardID(projectID, dashboardID)
	if errCode != http.StatusAccepted {
		log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID}).
			Error("failed to report for dashboard")
	}

	// Delete the dashboard itself.
	err := db.Model(&Dashboard{}).Where("id= ? AND project_id=?", dashboardID, projectID).
		Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID}).
			WithError(err).Error("Failed to delete dashboard.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

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

func createDefaultDashboardsForProject(projectId uint64, agentUUID string) int {
	return CreateWebAnalyticsDefaultDashboardWithUnits(projectId, agentUUID)
}
