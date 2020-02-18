package model

import (
	"encoding/json"
	C "factors/config"
	"net/http"
	"sort"
	"time"

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
	Type          string          `gorm:"type:varchar(5);not null" json:"type"`
	UnitsPosition *postgres.Jsonb `json:"units_position"` // map[string]map[uint64]int -> map[unit_type]unit_id:unit_position
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type UpdatableDashboard struct {
	Name          string                     `json:"name"`
	UnitsPosition *map[string]map[uint64]int `json:"units_position"`
}

const (
	DashboardTypePrivate        = "pr"
	DashboardTypeProjectVisible = "pv"
)

var types = []string{DashboardTypePrivate, DashboardTypeProjectVisible}

const AgentProjectPersonalDashboardName = "My Dashboard"

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
		&Dashboard{Name: AgentProjectPersonalDashboardName, Type: DashboardTypePrivate})
}

func GetDashboards(projectId uint64, agentUUID string) ([]Dashboard, int) {
	db := C.GetServices().Db

	var dashboards []Dashboard
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboards. Invalid project_id.")
		return dashboards, http.StatusBadRequest
	}

	err := db.Order("created_at ASC").Where("project_id = ? AND (type = ? OR agent_uuid = ?)",
		projectId, DashboardTypeProjectVisible, agentUUID).Find(&dashboards).Error
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

	if err := db.Where("project_id = ? AND id = ? AND (type = ? OR agent_uuid = ?)", projectId, id,
		DashboardTypeProjectVisible, agentUUID).First(&dashboard).Error; err != nil {
		logCtx.WithError(err).Error(
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

	// if maxPos exists, increament the maxPos by one for newPos.
	// else start positions from 0.
	if maxPos > -1 {
		newPos = maxPos + 1
	}
	currentPosition[unitType][unitId] = newPos

	return UpdateDashboard(projectId, agentUUID, id, &UpdatableDashboard{UnitsPosition: &currentPosition})
}

func removeAndRebalanceUnitsPositionByType(positions *map[string]map[uint64]int, unitId uint64, unitType string) {
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

func isValidUnitsPosition(positions *map[string]map[uint64]int) bool {
	if positions == nil {
		return false
	}

	idsByTypeMap := make(map[string]map[uint64]bool, 0)

	for _, typ := range UnitTypes {
		if _, exists := idsByTypeMap[typ]; !exists {
			idsByTypeMap[typ] = make(map[uint64]bool, 0)
		}

		if posMap, exists := (*positions)[typ]; exists && len(posMap) > 0 {
			actualPos := make([]int, 0, 0)

			for id, pos := range posMap {
				actualPos = append(actualPos, pos)
				// populate idsByType for id dup check.
				idsByTypeMap[typ][id] = true
			}

			// validates positions.
			sort.Sort(sort.IntSlice(actualPos))
			// positions should be 0 to len(actualPos) - 1
			// after sort, which is equivalent to array indexes.
			for index, pos := range actualPos {
				if pos != index {
					return false
				}
			}
		}
	}

	// id should be unique across different unit types.
	for idType, ids := range idsByTypeMap {
		for _, typ := range UnitTypes {
			if idType != typ {
				for id := range ids {
					if _, exists := idsByTypeMap[typ][id]; exists {
						return false
					}
				}
			}
		}
	}

	return true
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

		if !isValidUnitsPosition(dashboard.UnitsPosition) {
			logCtx.Error("Invalid units position.")
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

	// nothing to update.
	if len(updateFields) == 0 {
		return http.StatusBadRequest
	}

	err := db.Model(&Dashboard{}).Where("project_id = ? AND id = ?", projectId, id).Update(updateFields).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"update": updateFields}).WithError(err).Error("Failed to update dashboard.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
