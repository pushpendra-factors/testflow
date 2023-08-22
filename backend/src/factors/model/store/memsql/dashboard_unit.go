package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesDashboardUnitForeignConstraints(dashboardUnit model.DashboardUnit) int {
	logFields := log.Fields{
		"dash_board_unit": dashboardUnit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(dashboardUnit.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	} else {
		if exists := store.existsDashboardByID(dashboardUnit.ProjectID, dashboardUnit.DashboardId); !exists {
			return http.StatusBadRequest
		}
		if _, errCode := store.getQueryWithQueryID(
			dashboardUnit.ProjectID, dashboardUnit.QueryId, model.QueryTypeAllQueries); errCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

// CreateDashboardUnitForMultipleDashboards creates multiple dashboard units each for given
// list of dashboards
func (store *MemSQL) CreateDashboardUnitForMultipleDashboards(dashboardIds []int64, projectId int64,
	agentUUID string, unitPayload model.DashboardUnitRequestPayload) ([]*model.DashboardUnit, int, string) {
	logFields := log.Fields{
		"dash_board_ids": dashboardIds,
		"project_id":     projectId,
		"agent_uuid":     agentUUID,
		"unit_payload":   unitPayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var dashboardUnits []*model.DashboardUnit
	for _, dashboardId := range dashboardIds {
		dashboardUnit, errCode, errMsg := store.CreateDashboardUnit(projectId, agentUUID,
			&model.DashboardUnit{
				DashboardId:  dashboardId,
				Description:  unitPayload.Description,
				Presentation: unitPayload.Presentation,
				QueryId:      unitPayload.QueryId,
			})
		if errCode != http.StatusCreated {
			return nil, errCode, errMsg
		}
		dashboardUnits = append(dashboardUnits, dashboardUnit)
	}
	return dashboardUnits, http.StatusCreated, ""
}

// CreateMultipleDashboardUnits creates multiple dashboard units for list of queries for single dashboard
func (store *MemSQL) CreateMultipleDashboardUnits(requestPayload []model.DashboardUnitRequestPayload, projectId int64,
	agentUUID string, dashboardId int64) ([]*model.DashboardUnit, int, string) {
	logFields := log.Fields{
		"request_payload": requestPayload,
		"project_id":      projectId,
		"agent_uuid":      agentUUID,
		"dashboard_id":    dashboardId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var dashboardUnits []*model.DashboardUnit
	for _, payload := range requestPayload {

		// query should have been created before the dashboard unit
		if payload.QueryId == 0 {
			return dashboardUnits, http.StatusBadRequest, "invalid queryID. empty queryID."
		}
		dashboardUnit, errCode, errMsg := store.CreateDashboardUnit(projectId, agentUUID,
			&model.DashboardUnit{
				DashboardId:  dashboardId,
				Description:  payload.Description,
				Presentation: payload.Presentation,
				QueryId:      payload.QueryId,
			})
		if errCode != http.StatusCreated {
			return nil, errCode, errMsg
		}
		dashboardUnits = append(dashboardUnits, dashboardUnit)
	}
	return dashboardUnits, http.StatusCreated, ""
}

func (store *MemSQL) CreateDashboardUnit(projectId int64, agentUUID string, dashboardUnit *model.DashboardUnit) (*model.DashboardUnit, int, string) {
	logFields := log.Fields{
		"dashboard_unit": dashboardUnit,
		"project_id":     projectId,
		"agent_uuid":     agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.CreateDashboardUnitForDashboardClass(projectId, agentUUID, dashboardUnit, model.DashboardClassUserCreated)
}

func (store *MemSQL) CreateDashboardUnitForDashboardClass(projectId int64, agentUUID string, dashboardUnit *model.DashboardUnit,
	dashboardClass string) (*model.DashboardUnit, int, string) {
	logFields := log.Fields{
		"dashboard_unit":  dashboardUnit,
		"project_id":      projectId,
		"agent_uuid":      agentUUID,
		"dashboard_class": dashboardClass,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	logCtx := log.WithFields(logFields)
	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest, "Invalid request"
	}

	store.updateDashboardUnitPresentation(dashboardUnit)

	valid, errMsg := model.IsValidDashboardUnit(dashboardUnit)
	if !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	hasAccess, dashboard := store.HasAccessToDashboard(projectId, agentUUID, dashboardUnit.DashboardId)
	if !hasAccess {
		return nil, http.StatusForbidden, "Unauthorized to access dashboard"
	}

	if dashboard.Class != dashboardClass {
		return nil, http.StatusForbidden, fmt.Sprintf("Restricted access to dashboard class '%s'", dashboard.Class)
	}

	dashboardUnit.ProjectID = projectId
	if err := db.Create(dashboardUnit).Error; err != nil {
		errMsg := "Failed to create dashboard unit."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}

	return dashboardUnit, http.StatusCreated, ""
}

// updateDashboardUnitPresentation updates Presentation for dashboard Unit using corresponding query settings
func (store *MemSQL) updateDashboardUnitPresentation(unit *model.DashboardUnit) {
	logFields := log.Fields{
		"unit": unit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	queryInfo, errC := store.GetQueryWithQueryId(unit.ProjectID, unit.QueryId)
	if errC != http.StatusFound {
		logCtx.WithField("err_code", errC).
			WithField("query_id", unit.QueryId).
			Error("Failed to fetch query from query_id")
	}
	// request is received from new UI updating Presentation
	settings := make(map[string]string)
	err := json.Unmarshal(queryInfo.Settings.RawMessage, &settings)
	if err != nil {
		log.WithFields(log.Fields{"project_id": unit.ProjectID,
			"dashboardUnitId": unit.ID}).Error("failed to update presentation for given settings")
		return
	}
	unit.Presentation = settings["chart"]
}

// GetDashboardUnitsForProjectID Returns all dashboard units for the given projectID.
func (store *MemSQL) GetDashboardUnitsForProjectID(projectID int64) ([]model.DashboardUnit, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if projectID == 0 {
		log.Errorf("Invalid project id %d", projectID)
		return dashboardUnits, http.StatusBadRequest
	} else if err := db.Where("project_id = ? AND is_deleted = ?", projectID, false).
		Find(&dashboardUnits).Error; err != nil {
		log.WithError(err).Errorf("Failed to get dashboard units for projectID %d", projectID)
		return dashboardUnits, http.StatusInternalServerError
	}

	return dashboardUnits, http.StatusFound
}

// GetAttributionDashboardUnitsForProjectID Returns all attribution V1 dashboard units for the given projectID.
func (store *MemSQL) GetAttributionDashboardUnitsForProjectID(projectID int64) ([]model.DashboardUnit, int) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var dashboardUnits []model.DashboardUnit
	if projectID == 0 {
		log.Errorf("Invalid project id %d", projectID)
		return dashboardUnits, http.StatusBadRequest
	}

	dashboard, errCode := store.GetAttributionV1DashboardByDashboardName(projectID, model.AttributionV1Name)
	if errCode != http.StatusFound || dashboard == nil {
		log.WithFields(log.Fields{"method": "GetAttributionV1DashboardByDashboardName"}).Info("Failed to get Attribution dashboard")
		return dashboardUnits, errCode
	}

	dashboardUnits, errCode = store.GetDashboardUnitByDashboardID(projectID, dashboard.ID)

	if errCode != http.StatusFound || len(dashboardUnits) == 0 {
		log.WithFields(log.Fields{"method": "GetAttributionV1DashboardByDashboardName", "dashboard": dashboard}).Info("Failed to get dashboard units for Attribution V1 dashboard")
		return dashboardUnits, http.StatusInternalServerError
	}

	return dashboardUnits, http.StatusFound
}

func (store *MemSQL) GetDashboardUnits(projectId int64, agentUUID string, dashboardId int64) ([]model.DashboardUnit, int) {
	logFields := log.Fields{
		"project_id":   projectId,
		"agent_uuid":   agentUUID,
		"dashboard_id": dashboardId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if projectId == 0 || dashboardId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id or agent_id")
		return dashboardUnits, http.StatusBadRequest
	}

	if hasAccess, _ := store.HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
		return nil, http.StatusForbidden
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ? AND is_deleted = ?",
		projectId, dashboardId, false).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	return dashboardUnits, http.StatusFound
}

// GetDashboardUnitByDashboardID To get a dashboard unit by project id and dashboard id.
func (store *MemSQL) GetDashboardUnitByDashboardID(projectId int64, dashboardId int64) ([]model.DashboardUnit, int) {
	logFields := log.Fields{
		"project_id":   projectId,
		"dashboard_id": dashboardId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if projectId == 0 || dashboardId == 0 {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id or agent_id")
		return dashboardUnits, http.StatusBadRequest
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ? AND is_deleted = ?",
		projectId, dashboardId, false).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	return dashboardUnits, http.StatusFound
}

func (store *MemSQL) GetQueryFromUnitID(projectID int64, unitID int64) (queryClass string, queryInfo *model.Queries, errMsg string) {
	dashboardUnit, statusCode := store.GetDashboardUnitByUnitID(projectID, unitID)
	if statusCode != http.StatusFound {
		return "", nil, "Failed to fetch dashboard unit from unit ID"
	}
	return store.GetQueryAndClassFromDashboardUnit(dashboardUnit)
}

// GetDashboardUnitByUnitID To get a dashboard unit by project id and unit id.
func (store *MemSQL) GetDashboardUnitByUnitID(projectID int64, unitID int64) (*model.DashboardUnit, int) {
	logFields := log.Fields{
		"unit_id":    unitID,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var dashboardUnit model.DashboardUnit
	if err := db.Model(&model.DashboardUnit{}).Where("project_id = ? AND id=? AND is_deleted = ?",
		projectID, unitID, false).Find(&dashboardUnit).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &dashboardUnit, http.StatusFound
}

func (store *MemSQL) GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID int64, dashboardID int64, types []string) ([]model.DashboardUnit, int) {
	logFields := log.Fields{
		"project_id":   projectID,
		"types":        types,
		"dashboard_id": dashboardID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if projectID == 0 || dashboardID == 0 {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id ")
		return dashboardUnits, http.StatusBadRequest
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ? AND is_deleted = ? ",
		projectID, dashboardID, false).Where("presentation IN (?)", types).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	if len(dashboardUnits) == 0 {
		return dashboardUnits, http.StatusNotFound
	}

	return dashboardUnits, http.StatusFound
}

// NOTE: This can cause more latency when more dashboardUnits are there.
// Currently preload of related structs is not possible. Hence using loop.
func (store *MemSQL) GetDashboardUnitNamesByProjectIdTypeAndName(projectID int64, reqID string, typeOfQuery string, nameOfQuery string) ([]string, int) {
	rDashboardUnitNames := make([]string, 0)
	dashboardUnits, statusCode := store.GetDashboardUnitsForProjectID(projectID)
	if statusCode != http.StatusFound {
		log.WithField("projectID", projectID).Warn("Failed in getting dashboardUnits - GetDashboardUnitNamesByProjectIdTypeAndName")
		return rDashboardUnitNames, statusCode
	}
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	for _, dashboardUnit := range dashboardUnits {
		queryClass, queryInfo, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
		if errMsg != "" && !strings.Contains(errMsg, model.QueryNotFoundError) {
			logCtx.WithField("errMsg", errMsg).WithField("dashboardUnit", dashboardUnit).Warn("Error during decode of Dashboard unit - GetDashboardUnitNamesByProjectIdTypeAndName")
			return rDashboardUnitNames, http.StatusInternalServerError
		} else if strings.Contains(errMsg, model.QueryNotFoundError) {
			continue
		}

		if queryClass == typeOfQuery {
			baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
			if err != nil {
				errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
				logCtx.WithField("dashboardUnit", dashboardUnit).WithField("errMsg", errMsg).WithField("err", err).Warn("GetDashboardUnitNamesByProjectIdTypeAndName")
				return rDashboardUnitNames, http.StatusInternalServerError
			}

			isPresent := baseQuery.CheckIfNameIsPresent(nameOfQuery)
			if isPresent {
				rDashboardUnitNames = append(rDashboardUnitNames, queryInfo.Title)
			}
		}
	}
	return rDashboardUnitNames, http.StatusFound
}

func (store *MemSQL) GetDashboardUnitNamesByProjectIdTypeAndPropertyMappingName(projectID int64, reqID, nameOfPropertyMappings string) ([]string, int) {
	rDashboardUnitNames := make([]string, 0)
	dashboardUnits, statusCode := store.GetDashboardUnitsForProjectID(projectID)
	logCtx := log.WithField("reqID", reqID).WithField("projectID", projectID)
	if statusCode != http.StatusFound {
		logCtx.Warn("Failed in getting dashboardUnits - GetDashboardUnitNamesByProjectIdTypeAndName")
		return rDashboardUnitNames, statusCode
	}
	for _, dashboardUnit := range dashboardUnits {
		queryClass, queryInfo, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
		if errMsg != "" && !strings.Contains(errMsg, model.QueryNotFoundError) {
			logCtx.WithField("errMsg", errMsg).WithField("dashboardUnit", dashboardUnit).Warn("Error during decode of Dashboard unit - GetDashboardUnitNamesByProjectIdTypeAndName")
			return rDashboardUnitNames, http.StatusInternalServerError
		} else if strings.Contains(errMsg, model.QueryNotFoundError) {
			continue
		}

		// TODO: Are other query classes also required?
		if queryClass == model.QueryClassKPI {
			// Decode query from jsonb to KPIQueryGroup struct
			var query model.KPIQueryGroup
			if err := U.DecodePostgresJsonbToStructType(&queryInfo.Query, &query); err != nil {
				errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
				logCtx.WithField("dashboardUnit", dashboardUnit).WithField("errMsg", errMsg).WithField("err", err).Warn("GetDashboardUnitNamesByProjectIdTypeAndName")
				return rDashboardUnitNames, http.StatusInternalServerError
			}

			// Check if property mapping name is present in Global Filters and Group By of KPI Query
			if query.CheckIfPropertyMappingNameIsPresent(nameOfPropertyMappings) {
				rDashboardUnitNames = append(rDashboardUnitNames, queryInfo.Title)
			}
		}
	}
	if len(rDashboardUnitNames) == 0 {
		return rDashboardUnitNames, http.StatusNotFound
	}
	return rDashboardUnitNames, http.StatusFound
}

func (store *MemSQL) DeleteDashboardUnit(projectId int64, agentUUID string, dashboardId int64, id int64) int {
	logFields := log.Fields{
		"project_id":   projectId,
		"agent_uuid":   agentUUID,
		"dashboard_id": dashboardId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to delete dashboard unit. Invalid scope ids.")
		return http.StatusBadRequest
	}

	hasAccess, _ := store.HasAccessToDashboard(projectId, agentUUID, dashboardId)
	if !hasAccess {
		return http.StatusForbidden
	}

	/*errCode := store.removeUnitPositionOnDashboard(projectId, agentUUID, dashboardId, id, dashboard.UnitsPosition)
	if errCode != http.StatusAccepted {
		errMsg := "Failed remove position for unit on dashboard."
		log.WithFields(log.Fields{"project_id": projectId, "unitId": id}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
	}*/
	return store.deleteDashboardUnit(projectId, dashboardId, id)
}

// DeleteMultipleDashboardUnits deletes multiple dashboard units for given dashboard
func (store *MemSQL) DeleteMultipleDashboardUnits(projectID int64, agentUUID string, dashboardID int64,
	dashboardUnitIDs []int64) (int, string) {
	logFields := log.Fields{
		"dashboard_unit_ids": dashboardUnitIDs,
		"project_id":         projectID,
		"agent_uuid":         agentUUID,
		"dashboard_id":       dashboardID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	for _, dashboardUnitID := range dashboardUnitIDs {
		errCode := store.DeleteDashboardUnit(projectID, agentUUID, dashboardID, dashboardUnitID)
		if errCode != http.StatusAccepted {
			errMsg := "Failed delete unit on dashboard."
			log.WithFields(log.Fields{"project_id": projectID,
				"dashboard_id": dashboardID, "unit_id": dashboardUnitID}).Error(errMsg)
			return errCode, errMsg
		}
	}
	return http.StatusAccepted, ""
}

func (store *MemSQL) deleteDashboardUnit(projectID int64, dashboardID int64, ID int64) int {
	logFields := log.Fields{
		"project_id":   projectID,
		"id":           ID,
		"dashboard_id": dashboardID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	err := db.Model(&model.DashboardUnit{}).Where("id = ? AND project_id = ? AND dashboard_id = ?",
		ID, projectID, dashboardID).Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID,
			"unit_id": ID}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func (store *MemSQL) UpdateDashboardUnit(projectId int64, agentUUID string,
	dashboardId int64, id int64, unit *model.DashboardUnit) (*model.DashboardUnit, int) {
	logFields := log.Fields{
		"id":           id,
		"unit":         unit,
		"project_id":   projectId,
		"agent_uuid":   agentUUID,
		"dashboard_id": dashboardId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to update dashboard unit. Invalid scope ids.")
		return nil, http.StatusBadRequest
	}

	if hasAccess, _ := store.HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
		return nil, http.StatusForbidden
	}

	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)
	if unit.Description != "" {
		updateFields["description"] = unit.Description
	}
	if unit.Presentation != "" {
		updateFields["presentation"] = unit.Presentation
	}

	// nothing to update.
	if len(updateFields) == 0 {
		return nil, http.StatusBadRequest
	}
	var updatedDashboardUnitFields model.DashboardUnit
	db := C.GetServices().Db

	err := db.Model(&updatedDashboardUnitFields).Where("id = ? AND project_id = ? AND dashboard_id = ? AND is_deleted = ?",
		id, projectId, dashboardId, false).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error("updatedDashboardUnitFields failed at UpdateDashboardUnit in dashboard_unit.go")
		return nil, http.StatusInternalServerError
	}
	// returns only updated fields, avoid using it on model.DashboardUnit API.
	return &updatedDashboardUnitFields, http.StatusAccepted
}

// CacheDashboardUnitsForProjects Runs for all the projectIDs passed as comma separated.
func (store *MemSQL) CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string,
	numRoutines int, reportCollector *sync.Map, enableFilterOpt bool, startTimeForCache int64) {

	logFields := log.Fields{
		"string_projects_ids": stringProjectsIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_routines":        numRoutines,
		"report_collector":    reportCollector,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	projectIDs := store.GetProjectsToRunForIncludeExcludeString(stringProjectsIDs, excludeProjectIDs)
	var mapOfValidDashboardUnits map[int64]map[int64]bool
	var err error
	validUnitCount := int64(0)
	if C.GetUsageBasedDashboardCaching() == 1 {
		mapOfValidDashboardUnits, validUnitCount, err = model.GetDashboardCacheAnalyticsValidityMap()
		if err != nil {
			logCtx.WithError(err).Error("Failed to pull Dashboard Cached Units in last 14 days")
			return
		}
		logCtx.WithFields(log.Fields{"total_valid_units": validUnitCount}).Info("Total of units accessed in last 14 days - cache")
	}

	for _, projectID := range projectIDs {
		logCtx = logCtx.WithFields(log.Fields{"ProjectID": projectID})
		logCtx.Info("Starting to cache units for the project")
		startTime := U.TimeNowUnix()
		unitsCount := 0

		dashboardUnits, errCode := store.GetDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			logCtx.Info("not running caching for the project - units not found")
			continue
		}

		filterDashboardUnits := make([]model.DashboardUnit, 0)
		filterDashboardUnitQueryClass := make([]string, 0)
		for _, dashboardUnit := range dashboardUnits {

			// skip caching the dashboard if not in the list
			if !C.IsDashboardAllowedForCaching(dashboardUnit.DashboardId) {
				continue
			}
			queryClass, queryInfo, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg != "" {
				log.WithFields(logFields).Error("failed to get query class")
				continue
			}
			// skip web analytics here
			if queryClass == model.QueryClassWeb {
				continue
			}
			// skip attribution v1 query
			if queryInfo.Type == model.QueryTypeAttributionV1Query {
				continue
			}
			// filtering attribution query for attribution run
			if queryClass != model.QueryClassAttribution && C.GetOnlyAttributionDashboardCaching() == 1 {
				continue
			}
			// skip attribution query when skip is set = 1
			if queryClass == model.QueryClassAttribution && C.GetSkipAttributionDashboardCaching() == 1 {
				continue
			}
			// filtering kpi query for attribution run
			if queryClass != model.QueryClassKPI && C.GetOnlyKPICachingCaching() == 1 {
				continue
			}
			// skip kpi query for skip enabled run
			if queryClass == model.QueryClassKPI && C.GetSkipKPICachingCaching() == 1 {
				continue
			}

			filterDashboardUnits = append(filterDashboardUnits, dashboardUnit)
			filterDashboardUnitQueryClass = append(filterDashboardUnitQueryClass, queryClass)
		}
		if len(filterDashboardUnits) == 0 || (len(filterDashboardUnits) != len(filterDashboardUnitQueryClass)) {
			logCtx.WithFields(log.Fields{
				"finalDashboardUnits":          len(filterDashboardUnits),
				"finalDashboardUnitQueryClass": len(filterDashboardUnitQueryClass),
			}).Info("not running caching for project ")
			continue
		}

		if C.GetUsageBasedDashboardCaching() == 1 {

			var validDashboardUnitIDs []model.DashboardUnit
			var validDashboardUnitQueryClass []string
			if _, exists := mapOfValidDashboardUnits[projectID]; exists {
				for idx, dashboardUnit := range filterDashboardUnits {
					if value, ex := mapOfValidDashboardUnits[projectID][dashboardUnit.ID]; ex {
						if value {
							validDashboardUnitIDs = append(validDashboardUnitIDs, dashboardUnit)
							validDashboardUnitQueryClass = append(validDashboardUnitQueryClass, filterDashboardUnitQueryClass[idx])
						}
					} else {
						log.WithFields(log.Fields{"dashboardUnit": dashboardUnit}).Info("skipping caching unit as not accessed")
					}
				}
			}
			log.WithFields(log.Fields{"project_id": projectID, "total_units": len(filterDashboardUnits), "accessed_units": len(validDashboardUnitIDs)}).Info("Project Report - last 14 days")
			unitsCount = store.CacheDashboardUnitsForProjectID(projectID, validDashboardUnitIDs, validDashboardUnitQueryClass, numRoutines, reportCollector, enableFilterOpt, startTimeForCache)

		} else {
			log.WithFields(log.Fields{"project_id": projectID, "total_units": len(filterDashboardUnits), "accessed_units": len(filterDashboardUnits)}).Info("Project Report - normal run")
			unitsCount = store.CacheDashboardUnitsForProjectID(projectID, filterDashboardUnits, filterDashboardUnitQueryClass, numRoutines, reportCollector, enableFilterOpt, startTimeForCache)
		}
		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		log.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Project Report: Time taken for caching %d dashboard units", unitsCount)
	}
}

// DBCacheAttributionDashboardUnitsForProjects Runs for all the projectIDs passed as comma separated. ###1
func (store *MemSQL) DBCacheAttributionDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string,
	numRoutines int, reportCollector *sync.Map, enableFilterOpt bool, startTimeForCache int64) {

	logFields := log.Fields{
		"string_projects_ids": stringProjectsIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_routines":        numRoutines,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	projectIDs := store.GetProjectsToRunForIncludeExcludeString(stringProjectsIDs, excludeProjectIDs)
	var mapOfValidDashboardUnits map[int64]map[int64]bool
	var err error
	validUnitCount := int64(0)
	if C.GetUsageBasedDashboardCaching() == 1 {
		mapOfValidDashboardUnits, validUnitCount, err = model.GetDashboardCacheAnalyticsValidityMap()
		if err != nil {
			logCtx.WithError(err).Error("Failed to pull Dashboard Cached Units in last 14 days")
			return
		}
		logCtx.WithFields(log.Fields{"total_valid_units": validUnitCount}).Info("Total of units accessed in last 14 days - cache")
	}

	logCtx.WithFields(log.Fields{"all_eligible_units": mapOfValidDashboardUnits}).Info("dashboard eligible units")

	for _, projectID := range projectIDs {
		logCtx = logCtx.WithFields(log.Fields{"ProjectID": projectID})
		logCtx.Info("Starting to cache units for the project")
		startTime := U.TimeNowUnix()
		unitsCount := 0

		dashboardUnits, errCode := store.GetAttributionDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			logCtx.Info("not running caching for the project - units not found")
			continue
		}

		filterDashboardUnits := make([]model.DashboardUnit, 0)
		filterDashboardUnitQueryClass := make([]string, 0)
		for _, dashboardUnit := range dashboardUnits {

			// skip caching the dashboard if not in the list
			if !C.IsDashboardAllowedForCaching(dashboardUnit.DashboardId) {
				continue
			}
			// skip caching the dashboard unit if not in the list
			if !C.IsDashboardUnitAllowedForCaching(dashboardUnit.ID) {
				continue
			}
			queryClass, queryInfo, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg != "" {
				log.WithFields(logFields).Error("failed to get query class")
				continue
			}
			// Overriding it because we need all attribution queries on QueryClassAttributionV1
			// TODO after all queries are moved QueryClassAttributionV1
			queryClass = model.QueryClassAttributionV1
			// skip all other queries than attribution v1 query
			if queryInfo.Type != model.QueryTypeAttributionV1Query {
				continue
			}

			filterDashboardUnits = append(filterDashboardUnits, dashboardUnit)
			filterDashboardUnitQueryClass = append(filterDashboardUnitQueryClass, queryClass)
		}
		if len(filterDashboardUnits) == 0 || (len(filterDashboardUnits) != len(filterDashboardUnitQueryClass)) {
			logCtx.WithFields(log.Fields{
				"finalDashboardUnits":          len(filterDashboardUnits),
				"finalDashboardUnitQueryClass": len(filterDashboardUnitQueryClass),
			}).Info("not running caching for project ")
			continue
		}

		if C.GetUsageBasedDashboardCaching() == 1 {

			var validDashboardUnitIDs []model.DashboardUnit
			var validDashboardUnitQueryClass []string
			if _, exists := mapOfValidDashboardUnits[projectID]; exists {
				for idx, dashboardUnit := range filterDashboardUnits {
					if value, ex := mapOfValidDashboardUnits[projectID][dashboardUnit.ID]; ex {
						if value {
							validDashboardUnitIDs = append(validDashboardUnitIDs, dashboardUnit)
							validDashboardUnitQueryClass = append(validDashboardUnitQueryClass, filterDashboardUnitQueryClass[idx])
						}
					} else {
						log.WithFields(log.Fields{"dashboardUnit": dashboardUnit}).Info("skipping caching unit as not accessed")
					}
				}
			}
			log.WithFields(log.Fields{"project_id": projectID, "total_units": len(filterDashboardUnits), "accessed_units": len(validDashboardUnitIDs)}).Info("Project Report - last 14 days")
			unitsCount = store.CacheAttributionDashboardUnitsForProjectID(projectID, validDashboardUnitIDs, validDashboardUnitQueryClass, numRoutines, reportCollector, enableFilterOpt, startTimeForCache)
		} else {
			log.WithFields(log.Fields{"project_id": projectID, "total_units": len(filterDashboardUnits), "accessed_units": len(filterDashboardUnits)}).Info("Project Report - normal run")
			unitsCount = store.CacheAttributionDashboardUnitsForProjectID(projectID, filterDashboardUnits, filterDashboardUnitQueryClass, numRoutines, reportCollector, enableFilterOpt, startTimeForCache)
		}
		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		log.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Project Report: Time taken for caching %d dashboard units", unitsCount)
	}
}

// CacheAttributionDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`. ###2
func (store *MemSQL) CacheAttributionDashboardUnitsForProjectID(projectID int64, dashboardUnits []model.DashboardUnit,
	dashboardUnitQueryClass []string, numRoutines int, reportCollector *sync.Map, enableFilterOpt bool, startTimeForCache int64) int {
	logFields := log.Fields{
		"project_id":         projectID,
		"num_routines":       numRoutines,
		"report_collector":   reportCollector,
		"dashboard_unit_ids": dashboardUnits,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	if numRoutines == 0 {
		numRoutines = 1
	}

	var waitGroup sync.WaitGroup
	count := 0
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Add(U.MinInt(len(dashboardUnits), numRoutines))
	}
	for i := range dashboardUnits {
		count++
		if C.GetIsRunningForMemsql() == 0 {
			go store.CacheAttributionDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i], enableFilterOpt, startTimeForCache)
			if count%numRoutines == 0 {
				waitGroup.Wait()
				waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
			}
		} else {
			store.CacheAttributionDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i], enableFilterOpt, startTimeForCache)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Wait()
	}
	return len(dashboardUnits)
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func (store *MemSQL) CacheDashboardUnitsForProjectID(projectID int64, dashboardUnits []model.DashboardUnit,
	dashboardUnitQueryClass []string, numRoutines int, reportCollector *sync.Map, enableFilterOpt bool, startTimeForCache int64) int {
	logFields := log.Fields{
		"project_id":         projectID,
		"num_routines":       numRoutines,
		"report_collector":   reportCollector,
		"dashboard_unit_ids": dashboardUnits,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	if numRoutines == 0 {
		numRoutines = 1
	}

	var waitGroup sync.WaitGroup
	count := 0
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Add(U.MinInt(len(dashboardUnits), numRoutines))
	}
	for i := range dashboardUnits {
		count++
		if C.GetIsRunningForMemsql() == 0 {
			go store.CacheDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i], enableFilterOpt, startTimeForCache)
			if count%numRoutines == 0 {
				waitGroup.Wait()
				waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
			}
		} else {
			store.CacheDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i], enableFilterOpt, startTimeForCache)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Wait()
	}
	return len(dashboardUnits)
}

// GetQueryAndClassFromDashboardUnit returns query and query-class of dashboard unit.
func (store *MemSQL) GetQueryAndClassFromDashboardUnit(dashboardUnit *model.DashboardUnit) (queryClass string, queryInfo *model.Queries, errMsg string) {
	logFields := log.Fields{
		"dashboard_unit": dashboardUnit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	projectID := dashboardUnit.ProjectID
	savedQuery, errCode := store.GetQueryWithQueryId(projectID, dashboardUnit.QueryId)
	if errCode != http.StatusFound {
		errMsg = fmt.Sprintf("%s %d", model.QueryNotFoundError, dashboardUnit.QueryId)
		return "", nil, errMsg
	}

	queryClass, errMsg = store.GetQueryClassFromQueries(*savedQuery)
	if errMsg != "" {
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return "", nil, errMsg
	}
	return queryClass, savedQuery, ""
}

// GetQueryAndClassFromDashboardUnit returns query and query-class of dashboard unit.
func (store *MemSQL) GetQueryAndClassFromQueryIdString(queryIdString string, projectId int64) (queryClass string, queryInfo *model.Queries, errMsg string) {
	logFields := log.Fields{
		"query_id_string": queryIdString,
		"project_id":      projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	savedQuery, errCode := store.GetQueryWithQueryIdString(projectId, queryIdString)
	if errCode != http.StatusFound {
		errMsg = fmt.Sprintf("Failed to fetch query from query_id", queryIdString)
		return "", nil, errMsg
	}

	queryClass, errMsg = store.GetQueryClassFromQueries(*savedQuery)
	if errMsg != "" {
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return "", nil, errMsg
	}
	return queryClass, savedQuery, ""
}

func (store *MemSQL) GetQueryClassFromQueries(query model.Queries) (queryClass, errMsg string) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var temp_query model.Query
	var queryGroup model.QueryGroup
	// try decoding for Query
	U.DecodePostgresJsonbToStructType(&query.Query, &temp_query)
	if temp_query.Class == "" {
		// if fails, try decoding for QueryGroup
		err1 := U.DecodePostgresJsonbToStructType(&query.Query, &queryGroup)
		if err1 != nil {
			errMsg = fmt.Sprintf("Failed to decode jsonb query")
			return "", errMsg
		}
		queryClass = queryGroup.GetClass()
	} else {
		queryClass = temp_query.Class
	}
	return queryClass, ""
}

// CacheDashboardUnit Caches query for given dashboard unit for default date range presets.
func (store *MemSQL) CacheDashboardUnit(dashboardUnit model.DashboardUnit,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map, queryClass string, enableFilterOpt bool, startTimeForCache int64) {

	logFields := log.Fields{
		"dashboard_unit":   dashboardUnit,
		"wait_group":       waitGroup,
		"report_collector": reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}

	// excluding 'Web' class dashboard units
	if queryClass == model.QueryClassWeb {
		return
	}

	timezoneString, statusCode := store.GetTimezoneForProject(dashboardUnit.ProjectID)
	if statusCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed to get project Timezone for %d", dashboardUnit.ProjectID)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}

	// 1607755924 is Saturday, 12 December 2020 06:52:04! Kept it randomly to avoid running job if date range is not right
	if startTimeForCache != -1 && startTimeForCache > 1607755924 {
		store.RunCachingToBackFillRanges(dashboardUnit, startTimeForCache, timezoneString, logCtx, queryClass, reportCollector, enableFilterOpt)

		// Checking if we are running for custom date range
	} else if C.GetConfig().CustomDateStart != -1 && C.GetConfig().CustomDateStart != 0 &&
		C.GetConfig().CustomDateEnd != -1 && C.GetConfig().CustomDateEnd != 0 {

		store.RunCustomQueryRangeCaching(dashboardUnit, timezoneString, logCtx, queryClass, reportCollector, enableFilterOpt)
	} else {

		store.RunEverydayCaching(dashboardUnit, timezoneString, logCtx, queryClass, reportCollector, enableFilterOpt)
	}

}

// CacheAttributionDashboardUnit Caches query for given dashboard unit for default date range presets. ###3
func (store *MemSQL) CacheAttributionDashboardUnit(dashboardUnit model.DashboardUnit,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map, queryClass string, enableFilterOpt bool, startTimeForCache int64) {

	logFields := log.Fields{
		"dashboard_unit":   dashboardUnit,
		"wait_group":       waitGroup,
		"report_collector": reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}

	timezoneString, statusCode := store.GetTimezoneForProject(dashboardUnit.ProjectID)
	if statusCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed to get project Timezone for %d", dashboardUnit.ProjectID)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}

	// Running standrad caching attribution
	store.RunEverydayCachingForAttribution(dashboardUnit, timezoneString, logCtx, queryClass, reportCollector, enableFilterOpt)

	// Todo (Anil) backfilled 3 months version of attribution run
	store.RunCachingForLast3MonthsAttribution(dashboardUnit, timezoneString, logCtx, queryClass, reportCollector, enableFilterOpt)
}

// GetLast3MonthStoredQueriesFromAndTo fetches all date ranges computed and stored in DB
func (store *MemSQL) GetLast3MonthStoredQueriesFromAndTo(projectID, dashboardID, dashboardUnitID, queryID int64) (*[]model.DashQueryResult, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"query_id":   queryID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var dashQueryResults []model.DashQueryResult
	err := db.Table("dash_query_results").Select("from_t,to_t").Where("project_id=? AND dashboard_id=? AND dashboard_unit_id=? AND query_id=?",
		projectID, dashboardID, dashboardUnitID, queryID).Find(&dashQueryResults).Error

	if err != nil {
		log.WithFields(logFields).WithFields(log.Fields{"err": err}).Error("Error in executing query in GetLast3MonthStoredQueriesFromAndTo")
		return &[]model.DashQueryResult{}, http.StatusNotFound
	}
	log.WithFields(log.Fields{"dash_query_results_pulled": len(dashQueryResults)}).Info("no of dashQueryResults pulled")
	return &dashQueryResults, http.StatusFound
}

// RunCachingForLast3MonthsAttribution runs for all date ranges for last 3 months which are not yet computed
func (store *MemSQL) RunCachingForLast3MonthsAttribution(dashboardUnit model.DashboardUnit,
	timezoneString U.TimeZoneString, logCtx *log.Entry, queryClass string, reportCollector *sync.Map, enableFilterOpt bool) {

	var unitWaitGroup sync.WaitGroup

	_threeMonthsBackTime := time.Now().Unix() - int64(125*U.SECONDS_IN_A_DAY)
	startTimeForCache := U.SanitizeWeekStart(_threeMonthsBackTime, timezoneString)
	if startTimeForCache == 0 {
		errMsg := fmt.Sprintf("Error in getting the begining time from  %d", startTimeForCache)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardDBAttributionPingID, errMsg)
		return
	}
	resultsComputed, _ := store.GetLast3MonthStoredQueriesFromAndTo(dashboardUnit.ProjectID, dashboardUnit.DashboardId, dashboardUnit.ID, dashboardUnit.QueryId)

	// get from and to for all the time ranges!
	monthRange := U.GetAllMonthFromTo(startTimeForCache, timezoneString)
	weeksRange := U.GetAllWeeksFromStartTime(startTimeForCache, timezoneString)

	allRange := append(monthRange, weeksRange...)

	log.WithFields(log.Fields{"projectID": dashboardUnit.ProjectID,
		"allRange": allRange,
		"Method":   "RunCachingForLast3MonthsAttribution"}).Info("Attribution V1 caching debug")

	for _, queryRange := range allRange {

		from := queryRange.From
		to := queryRange.To

		// check if the result exists in DB
		for _, resultEntry := range *resultsComputed {
			if resultEntry.FromT == from && resultEntry.ToT == to {
				// already computed hence skip computing
				logCtx.Info("Already computed unit, skipping")
				continue
			}
		}

		queryInfo, errC := store.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
		if errC != http.StatusFound {
			logCtx.Errorf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
			continue
		}

		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardDBAttributionPingID, errMsg)
			return
		}

		projectSettingsJSON, statusCodeProjectSettings := store.GetProjectSetting(dashboardUnit.ProjectID)
		var cacheSettings model.CacheSettings

		if projectSettingsJSON == nil || statusCodeProjectSettings != http.StatusFound {
			log.WithField("projectID", dashboardUnit.ProjectID).WithField("statusCodeProjectSettings",
				statusCodeProjectSettings).Warn("errored in fetching project Settings")
			continue
		}

		if projectSettingsJSON.CacheSettings != nil && !U.IsEmptyPostgresJsonb(projectSettingsJSON.CacheSettings) {
			err = json.Unmarshal(projectSettingsJSON.CacheSettings.RawMessage, &cacheSettings)
		}

		if err != nil {
			continue
		}

		// Filtering queries on type and range for attribution query
		allowedPreset := cacheSettings.AttributionCachePresets
		shouldCache, from, to := model.ShouldCacheUnitForTimeRangeDashboardV1(queryClass, "", from, to,
			C.GetOnlyAttributionDashboardCaching(), C.GetSkipAttributionDashboardCaching(), allowedPreset[""])
		if !shouldCache {
			continue
		}

		baseQuery.SetQueryDateRange(from, to)
		baseQuery.SetTimeZone(timezoneString)
		baseQuery.SetDefaultGroupByTimestamp()
		err = baseQuery.TransformDateTypeFilters()
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query Value, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardDBAttributionPingID, errMsg)
			return
		}
		cachePayload := model.DashboardUnitCachePayload{
			DashboardUnit: dashboardUnit,
			BaseQuery:     baseQuery,
			Preset:        "",
		}
		if C.GetIsRunningForMemsql() == 0 {
			unitWaitGroup.Add(1)
			go store._cacheAttributionDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		} else {
			store._cacheAttributionDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

func (store *MemSQL) RunEverydayCachingForAttribution(dashboardUnit model.DashboardUnit, timezoneString U.TimeZoneString,
	logCtx *log.Entry, queryClass string, reportCollector *sync.Map, enableFilterOpt bool) {

	var unitWaitGroup sync.WaitGroup

	for preset, rangeFunction := range U.QueryDateRangePresets {

		fr, t, errCode := rangeFunction(timezoneString)
		if errCode != nil {
			errMsg := fmt.Sprintf("Failed to get proper project Timezone for %d", dashboardUnit.ProjectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardDBAttributionPingID, errMsg)
			return
		}

		queryInfo, errC := store.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
		if errC != http.StatusFound {
			logCtx.WithField("err_code", errC).Errorf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
			continue
		}

		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardDBAttributionPingID, errMsg)
			return
		}

		projectSettingsJSON, statusCodeProjectSettings := store.GetProjectSetting(dashboardUnit.ProjectID)
		var cacheSettings model.CacheSettings

		if projectSettingsJSON == nil || statusCodeProjectSettings != http.StatusFound {
			log.WithField("projectID", dashboardUnit.ProjectID).WithField("statusCodeProjectSettings",
				statusCodeProjectSettings).Warn("errored in fetching project Settings")
			continue
		}

		if projectSettingsJSON.CacheSettings != nil && !U.IsEmptyPostgresJsonb(projectSettingsJSON.CacheSettings) {
			err = json.Unmarshal(projectSettingsJSON.CacheSettings.RawMessage, &cacheSettings)
		}

		if err != nil {
			continue
		}

		// Filtering queries on type and range for attribution query
		allowedPreset := cacheSettings.AttributionCachePresets
		shouldCache, from, to := model.ShouldCacheUnitForTimeRange(queryClass, preset, fr, t,
			C.GetOnlyAttributionDashboardCaching(), C.GetSkipAttributionDashboardCaching(), allowedPreset[preset])
		if !shouldCache {
			continue
		}

		baseQuery.SetQueryDateRange(from, to)
		baseQuery.SetTimeZone(timezoneString)
		baseQuery.SetDefaultGroupByTimestamp()
		err = baseQuery.TransformDateTypeFilters()
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query Value, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		cachePayload := model.DashboardUnitCachePayload{
			DashboardUnit: dashboardUnit,
			BaseQuery:     baseQuery,
			Preset:        preset,
		}
		if C.GetIsRunningForMemsql() == 0 {
			unitWaitGroup.Add(1)
			go store._cacheAttributionDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		} else {
			store._cacheAttributionDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		}
	}

	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

func (store *MemSQL) RunCustomQueryRangeCaching(dashboardUnit model.DashboardUnit, timezoneString U.TimeZoneString,
	logCtx *log.Entry, queryClass string, reportCollector *sync.Map, enableFilterOpt bool) {

	var unitWaitGroup sync.WaitGroup

	from := C.GetConfig().CustomDateStart
	to := C.GetConfig().CustomDateEnd

	queryInfo, errC := store.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
	if errC != http.StatusFound {
		logCtx.WithField("err_code", errC).
			WithField("query_id", dashboardUnit.QueryId).
			Error("Failed to fetch query from query_id")
		return
	}
	if queryInfo.LockedForCacheInvalidation {
		logCtx.WithField("query_id", dashboardUnit.QueryId).Error("Didnt run caching because of lock on query.")
		return
	}

	// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
	baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
	if err != nil {
		errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}

	projectSettingsJSON, statusCodeProjectSettings := store.GetProjectSetting(dashboardUnit.ProjectID)
	var cacheSettings model.CacheSettings

	if projectSettingsJSON == nil || statusCodeProjectSettings != http.StatusFound {
		log.WithField("projectID", dashboardUnit.ProjectID).WithField("statusCodeProjectSettings", statusCodeProjectSettings).Warn("errored in fetching project Settings")
		return
	}

	if projectSettingsJSON.CacheSettings != nil && !U.IsEmptyPostgresJsonb(projectSettingsJSON.CacheSettings) {
		err = json.Unmarshal(projectSettingsJSON.CacheSettings.RawMessage, &cacheSettings)
	}

	if err != nil {
		return
	}

	baseQuery.SetQueryDateRange(from, to)
	baseQuery.SetTimeZone(timezoneString)
	baseQuery.SetDefaultGroupByTimestamp()
	err = baseQuery.TransformDateTypeFilters()
	if err != nil {
		errMsg := fmt.Sprintf("Error decoding query Value, query_id %d", dashboardUnit.QueryId)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}
	cachePayload := model.DashboardUnitCachePayload{
		DashboardUnit: dashboardUnit,
		BaseQuery:     baseQuery,
		Preset:        "Custom",
	}
	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Add(1)
		go store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
	} else {
		store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
	}

	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

func (store *MemSQL) RunEverydayCaching(dashboardUnit model.DashboardUnit, timezoneString U.TimeZoneString, logCtx *log.Entry,
	queryClass string, reportCollector *sync.Map, enableFilterOpt bool) {

	var unitWaitGroup sync.WaitGroup

	for preset, rangeFunction := range U.QueryDateRangePresets {

		fr, t, errCode := rangeFunction(timezoneString)
		if errCode != nil {
			errMsg := fmt.Sprintf("Failed to get proper project Timezone for %d", dashboardUnit.ProjectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}

		queryInfo, errC := store.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
		if errC != http.StatusFound {
			logCtx.WithField("err_code", errC).
				WithField("query", dashboardUnit.QueryId).
				Error("Failed to fetch query from query_id", dashboardUnit.QueryId)
			continue
		}

		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}

		projectSettingsJSON, statusCodeProjectSettings := store.GetProjectSetting(dashboardUnit.ProjectID)
		var cacheSettings model.CacheSettings

		if projectSettingsJSON == nil || statusCodeProjectSettings != http.StatusFound {
			log.WithField("projectID", dashboardUnit.ProjectID).WithField("statusCodeProjectSettings", statusCodeProjectSettings).Warn("errored in fetching project Settings")
			continue
		}

		if projectSettingsJSON.CacheSettings != nil && !U.IsEmptyPostgresJsonb(projectSettingsJSON.CacheSettings) {
			err = json.Unmarshal(projectSettingsJSON.CacheSettings.RawMessage, &cacheSettings)
		}

		if err != nil {
			continue
		}

		// Filtering queries on type and range for attribution query
		allowedPreset := cacheSettings.AttributionCachePresets
		shouldCache, from, to := model.ShouldCacheUnitForTimeRange(queryClass, preset, fr, t,
			C.GetOnlyAttributionDashboardCaching(), C.GetSkipAttributionDashboardCaching(), allowedPreset[preset])
		if !shouldCache {
			continue
		}

		baseQuery.SetQueryDateRange(from, to)
		baseQuery.SetTimeZone(timezoneString)
		baseQuery.SetDefaultGroupByTimestamp()
		err = baseQuery.TransformDateTypeFilters()
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query Value, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		cachePayload := model.DashboardUnitCachePayload{
			DashboardUnit: dashboardUnit,
			BaseQuery:     baseQuery,
			Preset:        preset,
		}
		if C.GetIsRunningForMemsql() == 0 {
			unitWaitGroup.Add(1)
			go store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		} else {
			store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		}
	}

	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

func (store *MemSQL) RunCachingToBackFillRanges(dashboardUnit model.DashboardUnit, startTimeForCache int64,
	timezoneString U.TimeZoneString, logCtx *log.Entry, queryClass string, reportCollector *sync.Map, enableFilterOpt bool) {

	var unitWaitGroup sync.WaitGroup

	// get from and to for all the time ranges!
	monthRange := U.GetAllMonthFromTo(startTimeForCache, timezoneString)
	weeksRange := U.GetAllWeeksFromStartTime(startTimeForCache, timezoneString)

	allRange := append(monthRange, weeksRange...)

	for _, queryRange := range allRange {

		from := queryRange.From
		to := queryRange.To

		queryInfo, errC := store.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
		if errC != http.StatusFound {
			logCtx.WithField("err_code", errC).
				WithField("query_id", dashboardUnit.QueryId).Errorf("Failed to fetch query from query_id")
			continue
		}

		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}

		projectSettingsJSON, statusCodeProjectSettings := store.GetProjectSetting(dashboardUnit.ProjectID)
		var cacheSettings model.CacheSettings

		if projectSettingsJSON == nil || statusCodeProjectSettings != http.StatusFound {
			log.WithField("projectID", dashboardUnit.ProjectID).WithField("statusCodeProjectSettings",
				statusCodeProjectSettings).Warn("errored in fetching project Settings")
			continue
		}

		if projectSettingsJSON.CacheSettings != nil && !U.IsEmptyPostgresJsonb(projectSettingsJSON.CacheSettings) {
			err = json.Unmarshal(projectSettingsJSON.CacheSettings.RawMessage, &cacheSettings)
		}

		if err != nil {
			continue
		}

		// Filtering queries on type and range for attribution query
		allowedPreset := cacheSettings.AttributionCachePresets
		shouldCache, from, to := model.ShouldCacheUnitForTimeRange(queryClass, "", from, to,
			C.GetOnlyAttributionDashboardCaching(), C.GetSkipAttributionDashboardCaching(), allowedPreset[""])
		if !shouldCache {
			continue
		}

		baseQuery.SetQueryDateRange(from, to)
		baseQuery.SetTimeZone(timezoneString)
		baseQuery.SetDefaultGroupByTimestamp()
		err = baseQuery.TransformDateTypeFilters()
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query Value, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		cachePayload := model.DashboardUnitCachePayload{
			DashboardUnit: dashboardUnit,
			BaseQuery:     baseQuery,
			Preset:        "",
		}
		if C.GetIsRunningForMemsql() == 0 {
			unitWaitGroup.Add(1)
			go store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		} else {
			store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector, enableFilterOpt)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

// CacheAttributionDashboardUnitForDateRange To cache a dashboard unit for the given range. ###5
func (store *MemSQL) CacheAttributionDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	enableFilterOpt bool) (int, string, model.CachingUnitReport) {

	logFields := log.Fields{
		"cache_payload": cachePayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Catches any panic in query execution and logs as an error. Prevent jobs from crashing.
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	from, to := baseQuery.GetQueryDateRange()
	timezoneString := baseQuery.GetTimeZone()
	preset := cachePayload.Preset
	if preset == "" {
		preset = U.GetPresetNameByFromAndTo(from, to, timezoneString)
	}

	unitReport := model.CachingUnitReport{
		UnitType:    model.CachingUnitNormal,
		ProjectId:   projectID,
		DashboardID: dashboardID,
		UnitID:      dashboardUnitID,
		QueryID:     dashboardUnit.QueryId,
		QueryClass:  baseQuery.GetClass(),
		Status:      model.CachingUnitStatusNotComputed,
		From:        from,
		To:          to,
		QueryRange:  U.SecondsToHMSString(to - from),
	}

	logCtx := log.WithFields(log.Fields{
		"cache_payload": cachePayload,
		"from":          from,
		"to":            to,
	}).WithFields(log.Fields{"PreUnitReport": unitReport})

	if !model.ShouldRefreshDashboardUnit(projectID, dashboardID, dashboardUnitID, from, to, timezoneString, false) {
		return http.StatusOK, "", unitReport
	}
	logCtx.Info("Starting to cache attribution v1 unit for date range")
	startTime := U.TimeNowUnix()

	var result interface{}
	var err error
	var errCode int
	var errMsg string
	queryTimedOut := false

	// Both attribution V0 and V1 has same base class model.QueryClassAttribution
	if baseQuery.GetClass() == model.QueryClassAttribution {

		attributionQuery := baseQuery.(*model.AttributionQueryUnitV1)
		unitReport.Query = attributionQuery

		channel := make(chan Result)
		logCtx.Info("Running attribution V1 caching")
		go store.runAttributionUnitV1(projectID, attributionQuery.Query, channel)

		select {
		case response := <-channel:
			result = response.res
			err = response.err
			errCode = response.errCode
			errMsg = response.errMsg
			if errCode == 0 {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "PanicOccured", "Error": ""}).Info("Failed for the attribution unit")
				errCode = http.StatusInternalServerError
			} else if err != nil && !model.IsIntegrationNotFoundError(response.err) {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the attribution unit")
				errCode = http.StatusInternalServerError
			} else if reflect.ValueOf(result).IsNil() {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the attribution unit - Result is nil")
				errCode = http.StatusInternalServerError
			} else {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "Success"}).Info("Success for the attribution unit")
				errCode = http.StatusOK
			}
		case <-time.After(180 * 60 * time.Second):
			queryTimedOut = true
			errCode = http.StatusInternalServerError
			logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut"}).Info("Timeout for the attribution unit")
		}

	}

	if errCode != http.StatusOK {
		if queryTimedOut {
			logCtx.WithField("QueryClass", baseQuery.GetClass()).WithField("Query", unitReport.Query).Info("query timed out - dashboard caching")
			unitReport.Status = model.CachingUnitStatusTimeout
		} else {
			unitReport.Status = model.CachingUnitStatusFailed
			logCtx.WithField("QueryClass", baseQuery.GetClass()).WithField("Query", unitReport.Query).Info("failed to run the query for dashboard caching")
		}
		unitReport.TimeTaken = U.TimeNowUnix() - startTime
		unitReport.TimeTakenStr = U.SecondsToHMSString(unitReport.TimeTaken)
		return http.StatusInternalServerError, fmt.Sprintf("Error while running query %s", errMsg), unitReport
	}

	timeTaken := U.TimeNowUnix() - startTime
	timeTakenString := U.SecondsToHMSString(timeTaken)
	logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).Info("Done caching for attribution dashbaord unit for range")

	meta := model.CacheMeta{
		Timezone:       string(timezoneString),
		From:           from,
		To:             to,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         preset,
	}

	model.SetCacheResultByDashboardIdAndUnitIdWithPreset(result, projectID, dashboardID, dashboardUnitID, preset,
		from, to, timezoneString, meta)
	errCode, errMsg = store.CreateResultInDB(result, projectID, dashboardID, dashboardUnitID, dashboardUnit.QueryId, preset,
		from, to, timezoneString, meta)
	if errCode != http.StatusCreated {
		logCtx.WithFields(log.Fields{"ErrorCode": errCode, "ErrorMsg": errMsg}).Error("Failed to crease database entry")
	} else {
		logCtx.WithFields(log.Fields{"ErrorCode": errCode}).Info("Added result in DB")
	}
	// Set in query cache result as well in case someone runs the same query from query handler.
	model.SetQueryCacheResult(projectID, baseQuery, result)
	unitReport.Status = model.CachingUnitStatusPassed
	unitReport.TimeTaken = timeTaken
	unitReport.TimeTakenStr = timeTakenString
	return http.StatusOK, "", unitReport
}

// CreateResultInDB inserts the computed dashboard query into DB under table DashQueryResult
func (store *MemSQL) CreateResultInDB(result interface{}, projectId int64, dashboardId int64, unitId int64, queryId int64,
	preset string, from, to int64, timezoneString U.TimeZoneString, meta interface{}) (int, string) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"dashboard_id": dashboardId, "dashboard_unit_id": unitId,
		"preset": preset, "from": from, "to": to,
	})

	if projectId == 0 || dashboardId == 0 || unitId == 0 {
		logCtx.Error("Invalid scope ids.")
		return http.StatusInternalServerError, "Invalid scope Ids"
	}
	db := C.GetServices().Db

	resMarshalled, err := json.Marshal(result)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode dashboardCacheResult.")
		return http.StatusInternalServerError, "Failed to encode dashboardCacheResult."
	}
	// resJson := &postgres.Jsonb{json.RawMessage(resMarshalled)}

	resultWarpper := model.DashQueryResult{
		ID:              U.GetUUID(),
		ProjectID:       projectId,
		DashboardID:     dashboardId,
		DashboardUnitID: unitId,
		QueryID:         queryId,
		FromT:           from,
		ToT:             to,
		Result:          resMarshalled,
		IsDeleted:       false,
		ComputedAt:      U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		//Timezone:    string(timezoneString)
	}

	if err := db.Create(&resultWarpper).Error; err != nil {
		errMsg := "Failed to insert result."
		logCtx.WithError(err).Error(errMsg)
		return http.StatusInternalServerError, errMsg
	}
	return http.StatusCreated, ""

}

// CacheDashboardUnitForDateRange To cache a dashboard unit for the given range.
func (store *MemSQL) CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	enableFilterOpt bool) (int, string, model.CachingUnitReport) {

	logFields := log.Fields{
		"cache_payload": cachePayload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Catches any panic in query execution and logs as an error. Prevent jobs from crashing.
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	from, to := baseQuery.GetQueryDateRange()
	timezoneString := baseQuery.GetTimeZone()
	preset := cachePayload.Preset
	if preset == "" {
		preset = U.GetPresetNameByFromAndTo(from, to, timezoneString)
	}

	unitReport := model.CachingUnitReport{
		UnitType:    model.CachingUnitNormal,
		ProjectId:   projectID,
		DashboardID: dashboardID,
		UnitID:      dashboardUnitID,
		QueryClass:  baseQuery.GetClass(),
		Status:      model.CachingUnitStatusNotComputed,
		From:        from,
		To:          to,
		QueryRange:  U.SecondsToHMSString(to - from),
	}

	logCtx := log.WithFields(logFields).WithFields(log.Fields{"PreUnitReport": unitReport})
	if !model.ShouldRefreshDashboardUnit(projectID, dashboardID, dashboardUnitID, from, to, timezoneString, false) {
		return http.StatusOK, "", unitReport
	}
	logCtx.Info("Starting to cache unit for date range")
	startTime := U.TimeNowUnix()

	var result interface{}
	var err error
	var errCode int
	var errMsg string
	queryTimedOut := false
	if baseQuery.GetClass() == model.QueryClassFunnel || baseQuery.GetClass() == model.QueryClassInsights {
		analyticsQuery := baseQuery.(*model.Query)
		unitReport.Query = analyticsQuery

		hashCode1, _ := analyticsQuery.GetQueryCacheHashString()
		channel := make(chan Result)

		// log.WithField("projectID", projectID).WithField("analyticsQuery", analyticsQuery).WithField("hashCode", hashCode).Warn("Before running Funnel or analytics query.")
		beforeRunningAnalyticsQuery := *analyticsQuery
		go store.runFunnelAndInsightsUnit(projectID, *analyticsQuery, channel, enableFilterOpt)
		hashCode2, _ := analyticsQuery.GetQueryCacheHashString()
		if hashCode2 != hashCode1 {
			log.WithField("projectID", projectID).WithField("analyticsQuery", analyticsQuery).WithField("beforeRunningAnalyticsQuery", beforeRunningAnalyticsQuery).
				WithField("hashCode1", hashCode1).WithField("hashCode2", hashCode2).Warn("Query is being modified.")
		}

		select {
		case response := <-channel:
			result = response.res
			err = response.err
			errCode = response.errCode
			errMsg = response.errMsg
			if errCode == 0 {
				logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "PanicOccured", "Error": ""}).Info("Failed for the FunnelORInsights unit")
				errCode = http.StatusInternalServerError
			} else if err != nil {
				logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the FunnelORInsights unit")
				errCode = http.StatusInternalServerError
			} else if reflect.ValueOf(result).IsNil() {
				logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the FunnelORInsights unit - Result is nil")
				errCode = http.StatusInternalServerError
			} else {
				logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "UnitRunTimeOut"}).Info("Success for the FunnelORInsights unit")
				errCode = http.StatusOK
			}
		case <-time.After(16 * 60 * time.Second):
			queryTimedOut = true
			errCode = http.StatusInternalServerError
			logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "UnitRunTimeOut"}).Info("Timeout for the FunnelORInsights unit")
		}

	} else if baseQuery.GetClass() == model.QueryClassAttribution {

		attributionQuery := baseQuery.(*model.AttributionQueryUnit)
		unitReport.Query = attributionQuery

		channel := make(chan Result)
		go store.runAttributionUnit(projectID, attributionQuery.Query, channel)

		select {
		case response := <-channel:
			result = response.res
			err = response.err
			errCode = response.errCode
			errMsg = response.errMsg
			if errCode == 0 {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "PanicOccured", "Error": ""}).Info("Failed for the attribution unit")
				errCode = http.StatusInternalServerError
			} else if err != nil && !model.IsIntegrationNotFoundError(response.err) {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the attribution unit")
				errCode = http.StatusInternalServerError
			} else if reflect.ValueOf(result).IsNil() {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the attribution unit - Result is nil")
				errCode = http.StatusInternalServerError
			} else {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut"}).Info("Success for the attribution unit")
				errCode = http.StatusOK
			}
		case <-time.After(180 * 60 * time.Second):
			queryTimedOut = true
			errCode = http.StatusInternalServerError
			logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut"}).Info("Timeout for the attribution unit")
		}

	} else if baseQuery.GetClass() == model.QueryClassChannelV1 {
		groupQuery := baseQuery.(*model.ChannelGroupQueryV1)
		unitReport.Query = groupQuery
		reqID := xid.New().String()
		result, errCode = store.RunChannelGroupQuery(projectID, groupQuery.Queries, reqID)
	} else if baseQuery.GetClass() == model.QueryClassEvents {
		groupQuery := baseQuery.(*model.QueryGroup)
		unitReport.Query = groupQuery
		hashCode1, _ := groupQuery.GetQueryCacheHashString()
		QueryiesBeforeRunning := groupQuery.Queries

		// log.WithField("projectID", projectID).WithField("groupQuery", groupQuery).WithField("hashCode", hashCode).Warn("Before analytics v1 query.")
		result, errCode = store.RunEventsGroupQuery(groupQuery.Queries, projectID, C.EnableOptimisedFilterOnEventUserQuery())
		hashCode2, _ := groupQuery.GetQueryCacheHashString()
		// log.WithField("projectID", projectID).WithField("groupQuery", groupQuery).WithField("hashCode", hashCode).Warn("After analytics v1 query.")
		if hashCode2 != hashCode1 {
			log.WithField("projectID", projectID).
				WithField("groupQuery", groupQuery).WithField("QueryiesBeforeRunning", QueryiesBeforeRunning).
				WithField("hashCode1", hashCode1).WithField("hashCode2", hashCode2).Warn("Events Query is being modified.")
		}

	} else if baseQuery.GetClass() == model.QueryClassKPI {
		groupQuery := baseQuery.(*model.KPIQueryGroup)
		unitReport.Query = groupQuery
		result, errCode = store.ExecuteKPIQueryGroup(projectID, "", *groupQuery, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	} else if baseQuery.GetClass() == model.QueryClassProfiles {
		groupQuery := baseQuery.(*model.ProfileQueryGroup)
		unitReport.Query = groupQuery
		result, errCode = store.RunProfilesGroupQuery(groupQuery.Queries, projectID, C.EnableOptimisedFilterOnProfileQuery())
	}
	if errCode != http.StatusOK {
		if queryTimedOut {
			logCtx.WithField("QueryClass", baseQuery.GetClass()).WithField("Query", unitReport.Query).Info("query timed out - dashboard caching")
			unitReport.Status = model.CachingUnitStatusTimeout
		} else {
			unitReport.Status = model.CachingUnitStatusFailed
			logCtx.WithField("QueryClass", baseQuery.GetClass()).WithField("Query", unitReport.Query).Info("failed to run the query for dashboard caching")
		}
		unitReport.TimeTaken = U.TimeNowUnix() - startTime
		unitReport.TimeTakenStr = U.SecondsToHMSString(unitReport.TimeTaken)
		return http.StatusInternalServerError, fmt.Sprintf("Error while running query %s", errMsg), unitReport
	}

	timeTaken := U.TimeNowUnix() - startTime
	timeTakenString := U.SecondsToHMSString(timeTaken)
	logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).Info("Done caching unit for range")
	//model.SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardID, dashboardUnitID, preset, from, to, timezoneString, nil)
	meta := model.CacheMeta{
		Timezone:       string(timezoneString),
		From:           from,
		To:             to,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         preset,
	}
	if C.IsLastComputedWhitelisted(projectID) {
		model.SetCacheResultByDashboardIdAndUnitIdWithPreset(result, projectID, dashboardID, dashboardUnitID, preset,
			from, to, timezoneString, meta)
	} else {
		model.SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardID, dashboardUnitID,
			from, to, timezoneString, meta)
	}
	// Set in query cache result as well in case someone runs the same query from query handler.
	model.SetQueryCacheResult(projectID, baseQuery, result)
	unitReport.Status = model.CachingUnitStatusPassed
	unitReport.TimeTaken = timeTaken
	unitReport.TimeTakenStr = timeTakenString
	return http.StatusOK, "", unitReport
}

type Result struct {
	res            *model.QueryResult
	err            error
	errCode        int
	errMsg         string
	lastComputedAt int64
}

func (store *MemSQL) runFunnelAndInsightsUnit(projectID int64, queryOriginal model.Query, c chan Result, enableFilterOpt bool) {
	r, eCode, eMsg := store.WrapperForAnalyze(projectID, queryOriginal, enableFilterOpt)
	result := Result{res: r, errCode: eCode, errMsg: eMsg, lastComputedAt: U.TimeNowUnix()}
	c <- result
}

func (store *MemSQL) WrapperForAnalyze(projectID int64, queryOriginal model.Query, enableFilterOpt bool) (*model.QueryResult, int, string) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	return store.Analyze(projectID, queryOriginal, enableFilterOpt, false) // disable dashboard caching for funnelv2
}

func (store *MemSQL) runAttributionUnitV1(projectID int64, queryOriginal *model.AttributionQueryV1, c chan Result) {
	attributionQueryUnitPayload := model.AttributionQueryUnitV1{
		Class: model.QueryClassAttribution,
		Query: queryOriginal,
	}
	queryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectID)
	debugQueryKey := model.GetStringKeyFromCacheRedisKey(queryKey)
	var r *model.QueryResult
	var err error
	var result Result
	r, err = store.WrapperForExecuteAttributionQueryV1(projectID, queryOriginal, debugQueryKey)
	if err != nil {
		result = Result{res: r, err: err, errMsg: "", lastComputedAt: U.TimeNowUnix(), errCode: http.StatusInternalServerError}
	} else {
		result = Result{res: r, err: err, errMsg: "", lastComputedAt: U.TimeNowUnix(), errCode: http.StatusOK}
	}

	c <- result
}

func (store *MemSQL) WrapperForExecuteAttributionQueryV1(projectID int64, queryOriginal *model.AttributionQueryV1, debugQueryKey string) (*model.QueryResult, error) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	return store.ExecuteAttributionQueryV1(projectID, queryOriginal, debugQueryKey,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
}

func (store *MemSQL) runAttributionUnit(projectID int64, queryOriginal *model.AttributionQuery, c chan Result) {
	attributionQueryUnitPayload := model.AttributionQueryUnit{
		Class: model.QueryClassAttribution,
		Query: queryOriginal,
	}
	QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectID)
	debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
	var r *model.QueryResult
	var err error
	var result Result
	r, err = store.WrapperForExecuteAttributionQueryV0(projectID, queryOriginal, debugQueryKey)
	if err != nil {
		result = Result{res: r, err: err, errMsg: "", lastComputedAt: U.TimeNowUnix(), errCode: http.StatusInternalServerError}
	} else {
		result = Result{res: r, err: err, errMsg: "", lastComputedAt: U.TimeNowUnix(), errCode: http.StatusOK}
	}

	c <- result
}

func (store *MemSQL) WrapperForExecuteAttributionQueryV0(projectID int64, queryOriginal *model.AttributionQuery, debugQueryKey string) (*model.QueryResult, error) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	return store.ExecuteAttributionQueryV0(projectID, queryOriginal, debugQueryKey,
		C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
}

// _cacheAttributionDashboardUnitForDateRange acts as collector to the core query caching method for Attribution V1 ###8
func (store *MemSQL) _cacheAttributionDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map, enableFilterOpt bool) {
	logFields := log.Fields{
		"cache_payload":    cachePayload,
		"wait_group":       waitGroup,
		"report_collector": reportCollector,
	}
	log.WithFields(logFields).Info("Attribution V1 caching debug _cacheAttributionDashboardUnitForDateRange")
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	from, to := baseQuery.GetQueryDateRange()
	logCtx := log.WithFields(log.Fields{
		"Method":          "_cacheAttributionDashboardUnitForDateRange",
		"ProjectID":       projectID,
		"DashboardID":     dashboardID,
		"DashboardUnitID": dashboardUnitID,
		"from":            from,
		"to":              to,
		"FromTo":          fmt.Sprintf("%d-%d", from, to),
	})

	errCode, errMsg, report := store.CacheAttributionDashboardUnitForDateRange(cachePayload, enableFilterOpt)
	reportCollector.Store(model.GetCachingUnitReportUniqueKey(report), report)
	if errCode != http.StatusOK {
		logCtx.WithField("err_code", errCode).Errorf("Error while running attribution v1 query %s", errMsg)
		return
	}
	logCtx.Info("Completed caching for AttributionV1 Dashboard unit")
}

func (store *MemSQL) cacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map, enableFilterOpt bool) {
	logFields := log.Fields{
		"cache_payload":    cachePayload,
		"wait_group":       waitGroup,
		"report_collector": reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	from, to := baseQuery.GetQueryDateRange()
	logCtx := log.WithFields(log.Fields{
		"Method":          "cacheDashboardUnitForDateRange",
		"ProjectID":       projectID,
		"DashboardID":     dashboardID,
		"DashboardUnitID": dashboardUnitID,
		"FromTo":          fmt.Sprintf("%d-%d", from, to),
	})
	errCode, errMsg, report := store.CacheDashboardUnitForDateRange(cachePayload, enableFilterOpt)
	reportCollector.Store(model.GetCachingUnitReportUniqueKey(report), report)
	if errCode != http.StatusOK {
		logCtx.WithField("err_code", errCode).Errorf("Error while running query %s", errMsg)
		return
	}
	logCtx.Info("Completed caching for Dashboard unit")
}

// CacheDashboardsForMonthlyRange To cache monthly dashboards for the project id.
func (store *MemSQL) CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string,
	numMonths, numRoutines int, reportCollector *sync.Map, enableFilterOpt bool) {
	logFields := log.Fields{
		"project_ids":         projectIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_months":          numMonths,
		"num_routines":        numRoutines,
		"report_collector":    reportCollector,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	projectIDsToRun := store.GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs)
	for _, projectID := range projectIDsToRun {
		logCtx := log.WithFields(logFields)
		timezoneString, statusCode := store.GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			errMsg := fmt.Sprintf("Failed to get project Timezone for %d", projectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			continue
		}
		monthlyRanges := U.GetMonthlyQueryRangesTuplesZ(numMonths, timezoneString)
		dashboardUnits, errCode := store.GetDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			return
		}

		for _, dashboardUnit := range dashboardUnits {
			queryInfo, errC := store.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
			if errC != http.StatusFound {
				logCtx.WithField("err_code", errC).
					WithField("query_id", dashboardUnit.QueryId).
					Error("Failed to fetch query from query_id")
				continue
			}

			queryClass, errMsg := store.GetQueryClassFromQueries(*queryInfo)
			if errMsg != "" {
				C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
				return
			}

			// excluding 'Web' class dashboard units
			if queryClass == model.QueryClassWeb {
				continue
			}

			var waitGroup sync.WaitGroup
			if C.GetIsRunningForMemsql() == 0 {
				waitGroup.Add(U.MinInt(len(monthlyRanges), numRoutines))
			}
			count := 0
			for _, monthlyRange := range monthlyRanges {
				count++
				from, to := monthlyRange.First, monthlyRange.Second
				// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
				baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
				if err != nil {
					errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
					C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
					return
				}
				baseQuery.SetQueryDateRange(from, to)
				baseQuery.SetTimeZone(timezoneString)
				err = baseQuery.TransformDateTypeFilters()
				if err != nil {
					errMsg := fmt.Sprintf("Error decoding query Value, query_id %d", dashboardUnit.QueryId)
					C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
					return
				}
				cachePayload := model.DashboardUnitCachePayload{
					DashboardUnit: dashboardUnit,
					BaseQuery:     baseQuery,
				}
				if C.GetIsRunningForMemsql() == 0 {
					go store.cacheDashboardUnitForDateRange(cachePayload, &waitGroup, reportCollector, enableFilterOpt)
					if count%numRoutines == 0 {
						waitGroup.Wait()
						waitGroup.Add(U.MinInt(len(monthlyRanges)-count, numRoutines))
					}
				} else {
					store.cacheDashboardUnitForDateRange(cachePayload, &waitGroup, reportCollector, enableFilterOpt)
				}
			}
			if C.GetIsRunningForMemsql() == 0 {
				waitGroup.Wait()
			}
		}
	}
}

func (store *MemSQL) GetFailedUnitsByProject(cacheReports []model.CachingUnitReport) map[int64][]model.FailedDashboardUnitReport {

	var units []model.CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectFailedUnits := make(map[int64][]model.FailedDashboardUnitReport)
	for _, unit := range cacheReports {
		timezone, _ := store.GetTimezoneForProject(unit.ProjectId)
		if unit.Status == model.CachingUnitStatusFailed {
			failedUnit := model.FailedDashboardUnitReport{
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
				failedUnits := []model.FailedDashboardUnitReport{failedUnit}
				projectFailedUnits[unit.ProjectId] = failedUnits
			}
		}
	}
	return projectFailedUnits
}

func (store *MemSQL) GetTimedOutUnitsByProject(cacheReports []model.CachingUnitReport) map[int64][]model.FailedDashboardUnitReport {

	var units []model.CachingUnitReport
	U.DeepCopy(&cacheReports, &units)

	sort.Slice(units, func(i, j int) bool {
		return units[i].TimeTaken > units[j].TimeTaken
	})

	projectTimedOutUnits := make(map[int64][]model.FailedDashboardUnitReport)
	for _, unit := range cacheReports {
		timezone, _ := store.GetTimezoneForProject(unit.ProjectId)
		if unit.Status == model.CachingUnitStatusTimeout {
			timedOutUnit := model.FailedDashboardUnitReport{
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
				failedUnits := []model.FailedDashboardUnitReport{timedOutUnit}
				projectTimedOutUnits[unit.ProjectId] = failedUnits
			}
		}
	}
	return projectTimedOutUnits
}
