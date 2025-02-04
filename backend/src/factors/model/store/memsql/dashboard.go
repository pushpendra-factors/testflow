package memsql

import (
	"encoding/json"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesDashboardForeignConstraints(dashboard model.Dashboard) int {
	logFields := log.Fields{
		"dashboard": dashboard,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, projectErrCode := store.GetProject(dashboard.ProjectId)
	_, agentErrCode := store.GetAgentByUUID(dashboard.AgentUUID)
	if projectErrCode != http.StatusFound || agentErrCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func isValidDashboard(dashboard *model.Dashboard) bool {
	logFields := log.Fields{
		"dashboard": dashboard,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if dashboard.Name == "" {
		return false
	}

	validType := false
	for _, t := range model.DashboardTypes {
		if t == dashboard.Type {
			validType = true
			break
		}
	}

	return validType
}

func (store *MemSQL) CreateDashboard(projectId int64, agentUUID string, dashboard *model.Dashboard) (*model.Dashboard, int) {
	logFields := log.Fields{
		"dashboard":  dashboard,
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	if !isValidDashboard(dashboard) {
		return nil, http.StatusBadRequest
	}

	if dashboard.Class == "" {
		dashboard.Class = model.DashboardClassUserCreated
	}

	allBoardFolder, errCode := store.GetAllBoardsDashboardFolder(projectId)
	if errCode != http.StatusFound {
		log.WithFields(log.Fields{"dashboard": dashboard, "project_id": projectId}).Error("Failed to create dashboard. All Boards error.")
		return nil, http.StatusInternalServerError
	}

	dashboard.FolderID = allBoardFolder.Id
	dashboard.ProjectId = projectId
	dashboard.AgentUUID = agentUUID
	if errCode := store.satisfiesDashboardForeignConstraints(*dashboard); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	if err := db.Create(dashboard).Error; err != nil {
		log.WithFields(log.Fields{"dashboard": dashboard,
			"project_id": projectId}).WithError(err).Error("Failed to create dashboard.")
		return nil, http.StatusInternalServerError
	}

	return dashboard, http.StatusCreated
}

func (store *MemSQL) CreateAgentPersonalDashboardForProject(projectId int64, agentUUID string) (*model.Dashboard, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.CreateDashboard(projectId, agentUUID,
		&model.Dashboard{Name: model.AgentProjectPersonalDashboardName,
			Description: model.AgentProjectPersonalDashboardDescription,
			Type:        model.DashboardTypePrivate,
		})
}

func (store *MemSQL) existsDashboardByID(projectID int64, dashboardID int64) bool {
	logFields := log.Fields{
		"dashboard_id": dashboardID,
		"project_id":   projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboard model.Dashboard
	err := db.Limit(1).Where("project_id = ? AND id = ?", projectID, dashboardID).Select("id").Find(&dashboard).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithField("project_id", projectID).WithField("id", dashboardID).Error("Failed to check dashboard by id")
		}
		return false
	}
	if dashboard.ID != 0 {
		return true
	}
	return false
}

func (store *MemSQL) existsDashboardByInternalID(projectID int64, internalID int64) bool {
	logFields := log.Fields{
		"dashboard_id": internalID,
		"project_id":   projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboard model.Dashboard
	err := db.Limit(1).Where("project_id = ? AND internal_id = ?", projectID, internalID).Select("id").Find(&dashboard).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithField("project_id", projectID).WithField("internal_id", internalID).Error("Failed to check dashboard by internal id")
			return true
		}
		return false
	}
	if dashboard.ID != 0 {
		return true
	}
	return false
}

func (store *MemSQL) GetDashboards(projectId int64, agentUUID string) ([]model.Dashboard, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboards []model.Dashboard
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboards. Invalid project_id.")
		return dashboards, http.StatusBadRequest
	}

	err := db.Order("created_at ASC").Where("project_id = ? AND (type = ? OR type = ? OR agent_uuid = ?) AND is_deleted = ?",
		projectId, model.DashboardTypeProjectVisible, model.DashboardTypeAttributionV1, agentUUID, false).Find(&dashboards).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboards.")
		return dashboards, http.StatusInternalServerError
	}

	return dashboards, http.StatusFound
}

func (store *MemSQL) GetDashboard(projectId int64, agentUUID string, id int64) (*model.Dashboard, int) {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	logCtx := log.WithFields(logFields)

	var dashboard model.Dashboard
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard. Invalid project_id or agent_id")
		return nil, http.StatusBadRequest
	}

	if err := db.Where("project_id = ? AND id = ? AND (type = ? OR type = ? OR agent_uuid = ?) AND is_deleted = ?",
		projectId, id, model.DashboardTypeProjectVisible, model.DashboardTypeAttributionV1, agentUUID, false).First(&dashboard).Error; err != nil {
		logCtx.WithError(err).WithField("dashboardID", id).Error(
			"Getting dashboard failed in GetDashboard")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	return &dashboard, http.StatusFound
}

// GetAttributionV1DashboardByDashboardName returns attribution v1 dashboard for given project id and dashboard name
func (store *MemSQL) GetAttributionV1DashboardByDashboardName(projectId int64, dashboardName string) (*model.Dashboard, int) {
	logFields := log.Fields{
		"project_id":     projectId,
		"dashboard_name": dashboardName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	logCtx := log.WithFields(logFields)

	var dashboard model.Dashboard
	if dashboardName == "" {
		log.Error("Failed to get dashboard by name. Invalid dashboard_name")
		return nil, http.StatusBadRequest
	}

	if err := db.Where("project_id = ? AND name = ? AND type = ? AND is_deleted = ?",
		projectId, dashboardName, model.DashboardTypeAttributionV1, false).First(&dashboard).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		logCtx.WithError(err).WithField("dashboardName", dashboardName).Error(
			"Getting dashboard failed in GetDashboard")
		return nil, http.StatusInternalServerError
	}

	return &dashboard, http.StatusFound
}

// HasAccessToDashboard validates access to dashboard.
func (store *MemSQL) HasAccessToDashboard(projectId int64, agentUUID string, id int64) (bool, *model.Dashboard) {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	dashboard, errCode := store.GetDashboard(projectId, agentUUID, id)
	if errCode != http.StatusFound {
		return false, nil
	}

	return true, dashboard
}

// Adds a position to the given unit on dashboard by unit_type.
func (store *MemSQL) addUnitPositionOnDashboard(projectId int64, agentUUID string,
	id int64, unitId int64, unitType string, currentUnitsPos *postgres.Jsonb) int {
	logFields := log.Fields{
		"id":               id,
		"project_id":       projectId,
		"agent_uuid":       agentUUID,
		"unit_id":          unitId,
		"unit_type":        unitType,
		"current_unit_pos": currentUnitsPos,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || agentUUID == "" || id == 0 || unitId == 0 {
		return http.StatusBadRequest
	}

	var currentPosition map[string]map[int64]int
	newPos := 0
	if currentUnitsPos != nil {
		err := json.Unmarshal((*currentUnitsPos).RawMessage, &currentPosition)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "id": id,
				"unit_position": currentPosition}).WithError(err).Error("Failed decoding current units position.")
			return http.StatusInternalServerError
		}
	} else {
		currentPosition = make(map[string]map[int64]int, 0)
	}

	if _, typeExists := currentPosition[unitType]; !typeExists {
		currentPosition[unitType] = make(map[int64]int, 0)
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

	return store.UpdateDashboard(projectId, agentUUID, id,
		&model.UpdatableDashboard{UnitsPosition: &currentPosition})
}

func removeAndRebalanceUnitsPositionByType(positions *map[string]map[int64]int,
	unitId int64, unitType string) {
	logFields := log.Fields{
		"positions": positions,
		"unit_id":   unitId,
		"unit_type": unitType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	removedPos := (*positions)[unitType][unitId]
	delete((*positions)[unitType], unitId)

	// reposition units positioned after removed unit.
	for id, pos := range (*positions)[unitType] {
		if pos > removedPos {
			(*positions)[unitType][id] = pos - 1
		}
	}
}

func (store *MemSQL) removeUnitPositionOnDashboard(projectId int64, agentUUID string,
	id int64, unitId int64, currentUnitsPos *postgres.Jsonb) int {
	logFields := log.Fields{
		"id":               id,
		"project_id":       projectId,
		"agent_uuid":       agentUUID,
		"unit_id":          unitId,
		"current_unit_pos": currentUnitsPos,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || agentUUID == "" || id == 0 ||
		unitId == 0 || currentUnitsPos == nil {
		return http.StatusBadRequest
	}

	var currentPositions map[string]map[int64]int
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

	return store.UpdateDashboard(projectId, agentUUID, id,
		&model.UpdatableDashboard{UnitsPosition: &currentPositions})
}

func isValidUnitsPosition(positions *map[string]map[int64]int) (bool, error) {
	logFields := log.Fields{
		"positions": positions,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if positions == nil {
		return false, errors.New("nil position map")
	}

	for _, typ := range model.UnitPresentationTypes {
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

// UpdateDashboard updates name,description,type,settings, folderId. The update also impacts the updated_at column value with each update.
func (store *MemSQL) UpdateDashboard(projectId int64, agentUUID string, id int64, dashboard *model.UpdatableDashboard) int {
	logFields := log.Fields{
		"id":         id,
		"project_id": projectId,
		"agent_uuid": agentUUID,
		"dashboard":  dashboard,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectId == 0 || agentUUID == "" || id == 0 {
		log.Error("Failed to update dashboard. Invalid scope ids.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	// use HasAccessToDashboard maintain consistency on checking accessibility.
	if hasAccess, _ := store.HasAccessToDashboard(projectId, agentUUID, id); !hasAccess {
		// do not use http.StatusUnauthorised, breaks UI.
		return http.StatusForbidden
	}

	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)
	if dashboard.UnitsPosition != nil {
		logCtx := log.WithFields(logFields)

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

	if dashboard.Settings != nil && !U.IsEmptyPostgresJsonb(dashboard.Settings) {
		updateFields["settings"] = dashboard.Settings
	}
	if dashboard.Type != "" {
		updateFields["type"] = dashboard.Type
	}

	if dashboard.FolderID != "" {
		updateFields["folder_id"] = dashboard.FolderID
	}

	// nothing to update.
	if len(updateFields) == 0 {
		return http.StatusBadRequest
	}

	err := db.Model(&model.Dashboard{}).Where("project_id = ? AND id = ? AND is_deleted = ?", projectId, id, false).
		Update(updateFields).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"update": updateFields}).WithError(err).Error("Failed to update dashboard.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// DeleteDashboard To delete a dashboard by id.
func (store *MemSQL) DeleteDashboard(projectID int64, agentUUID string, dashboardID int64) int {
	logFields := log.Fields{
		"project_id":   projectID,
		"agent_uuid":   agentUUID,
		"dashboard_id": dashboardID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	if projectID == 0 || agentUUID == "" ||
		dashboardID == 0 {

		log.Error("Failed to delete dashboard. Invalid scope ids.")
		return http.StatusBadRequest
	}

	hasAccess, _ := store.HasAccessToDashboard(projectID, agentUUID, dashboardID)
	if !hasAccess {
		return http.StatusForbidden
	}

	dashboardUnits, errCode := store.GetDashboardUnits(projectID, agentUUID, dashboardID)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("failed to fetch dashboard units for delete dashboard")
		return http.StatusBadRequest
	}

	// Delete dashboard units for the the given dashboard first.
	for _, dashboardUnit := range dashboardUnits {
		errCode := store.deleteDashboardUnit(projectID, dashboardID, dashboardUnit.ID)
		if errCode != http.StatusAccepted {
			// continue
			log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID,
				"dashboard_uint_id": dashboardUnit.ID, "err_code": errCode}).Error("failed to delete dashboard unit.")
		}
	}

	// Delete the dashboard itself.
	err := db.Model(&model.Dashboard{}).Where("id= ? AND project_id=?", dashboardID, projectID).
		Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID}).
			WithError(err).Error("Failed to delete dashboard.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func (store *MemSQL) createDefaultDashboardsForProject(projectId int64, agentUUID string) int {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.CreateWebAnalyticsDefaultDashboardWithUnits(projectId, agentUUID)
}

func (store *MemSQL) GetDashboardsByFolderId(projectId int64, folderId string) ([]model.Dashboard, int) {

	logCtx := log.WithFields(log.Fields{"folder_id": folderId, "project_id": projectId})
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db
	var dashboards []model.Dashboard

	err := db.Order("created_at ASC").Where("project_id = ? AND folder_id= ? AND is_deleted = ?", projectId, folderId, false).Find(&dashboards).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return dashboards, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get dashboards.")
		return dashboards, http.StatusInternalServerError
	}

	return dashboards, http.StatusFound

}

func (store *MemSQL) UpdateFolderIdForMultipleDashboards(projectId int64, dashboards []model.Dashboard, folderId string) int {

	logCtx := log.WithFields(log.Fields{"dashboards": dashboards, "project_id": projectId, "folder_id": folderId})
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logCtx.Data)
	db := C.GetServices().Db

	updateFields := make(map[string]interface{}, 0)
	updateFields["folder_id"] = folderId
	var dashboardIds []int64
	for _, dashboard := range dashboards {
		dashboardIds = append(dashboardIds, dashboard.ID)
	}

	err := db.Model(&model.Dashboard{}).Where("project_id = ? AND id IN (?) AND is_deleted = ?", projectId, dashboardIds, false).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update folder id for multiple dashboard.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted

}
