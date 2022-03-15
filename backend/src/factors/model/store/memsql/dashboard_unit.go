package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
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
func (store *MemSQL) CreateDashboardUnitForMultipleDashboards(dashboardIds []uint64, projectId uint64,
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
func (store *MemSQL) CreateMultipleDashboardUnits(requestPayload []model.DashboardUnitRequestPayload, projectId uint64,
	agentUUID string, dashboardId uint64) ([]*model.DashboardUnit, int, string) {
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

func (store *MemSQL) CreateDashboardUnit(projectId uint64, agentUUID string, dashboardUnit *model.DashboardUnit) (*model.DashboardUnit, int, string) {
	logFields := log.Fields{
		"dashboard_unit": dashboardUnit,
		"project_id":     projectId,
		"agent_uuid":     agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.CreateDashboardUnitForDashboardClass(projectId, agentUUID, dashboardUnit, model.DashboardClassUserCreated)
}

func (store *MemSQL) CreateDashboardUnitForDashboardClass(projectId uint64, agentUUID string, dashboardUnit *model.DashboardUnit,
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
		logCtx.Errorf("Failed to fetch query from query_id %d", unit.QueryId)
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
func (store *MemSQL) GetDashboardUnitsForProjectID(projectID uint64) ([]model.DashboardUnit, int) {
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

func (store *MemSQL) GetDashboardUnits(projectId uint64, agentUUID string, dashboardId uint64) ([]model.DashboardUnit, int) {
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

// GetDashboardUnitByUnitID To get a dashboard unit by project id and unit id.
func (store *MemSQL) GetDashboardUnitByUnitID(projectID, unitID uint64) (*model.DashboardUnit, int) {
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

func (store *MemSQL) GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID, dashboardID uint64, types []string) ([]model.DashboardUnit, int) {
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

func (store *MemSQL) DeleteDashboardUnit(projectId uint64, agentUUID string, dashboardId uint64, id uint64) int {
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

	hasAccess, dashboard := store.HasAccessToDashboard(projectId, agentUUID, dashboardId)
	if !hasAccess {
		return http.StatusForbidden
	}

	errCode := store.removeUnitPositionOnDashboard(projectId, agentUUID, dashboardId, id, dashboard.UnitsPosition)
	if errCode != http.StatusAccepted {
		errMsg := "Failed remove position for unit on dashboard."
		log.WithFields(log.Fields{"project_id": projectId, "unitId": id}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
	}
	return store.deleteDashboardUnit(projectId, dashboardId, id)
}

// DeleteMultipleDashboardUnits deletes multiple dashboard units for given dashboard
func (store *MemSQL) DeleteMultipleDashboardUnits(projectID uint64, agentUUID string, dashboardID uint64,
	dashboardUnitIDs []uint64) (int, string) {
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

func (store *MemSQL) deleteDashboardUnit(projectID uint64, dashboardID uint64, ID uint64) int {
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

func (store *MemSQL) UpdateDashboardUnit(projectId uint64, agentUUID string,
	dashboardId uint64, id uint64, unit *model.DashboardUnit) (*model.DashboardUnit, int) {
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
func (store *MemSQL) CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map) {
	logFields := log.Fields{
		"string_projects_ids": stringProjectsIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_routines":        numRoutines,
		"report_collector":    reportCollector,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	projectIDs := store.GetProjectsToRunForIncludeExcludeString(stringProjectsIDs, excludeProjectIDs)
	var mapOfValidDashboardUnits map[uint64]map[uint64]bool
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
			logCtx.Error("not running caching for the project - units not found")
			continue
		}

		filterDashboardUnits := make([]model.DashboardUnit, 0)
		filterDashboardUnitQueryClass := make([]string, 0)
		for _, dashboardUnit := range dashboardUnits {

			queryClass, _, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg != "" {
				log.WithFields(logFields).Error("failed to get query class")
				continue
			}
			// skip web analytics here
			if queryClass == model.QueryClassWeb {
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
			unitsCount = store.CacheDashboardUnitsForProjectID(projectID, validDashboardUnitIDs, validDashboardUnitQueryClass, numRoutines, reportCollector)

		} else {

			log.WithFields(log.Fields{"project_id": projectID, "total_units": len(filterDashboardUnits), "accessed_units": len(filterDashboardUnits)}).Info("Project Report - normal run")
			unitsCount = store.CacheDashboardUnitsForProjectID(projectID, filterDashboardUnits, filterDashboardUnitQueryClass, numRoutines, reportCollector)
		}
		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		log.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Project Report: Time taken for caching %d dashboard units", unitsCount)
	}
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func (store *MemSQL) CacheDashboardUnitsForProjectID(projectID uint64, dashboardUnits []model.DashboardUnit,
	dashboardUnitQueryClass []string, numRoutines int, reportCollector *sync.Map) int {
	logFields := log.Fields{
		"project_id":         projectID,
		"num_routines":       numRoutines,
		"report_collector":   reportCollector,
		"dashboard_unit_ids": dashboardUnits,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
			go store.CacheDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i])
			if count%numRoutines == 0 {
				waitGroup.Wait()
				waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
			}
		} else {
			store.CacheDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i])
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
		errMsg = fmt.Sprintf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
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
func (store *MemSQL) GetQueryAndClassFromQueryIdString(queryIdString string, projectId uint64) (queryClass string, queryInfo *model.Queries, errMsg string) {
	logFields := log.Fields{
		"query_id_string": queryIdString,
		"project_id":      projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	savedQuery, errCode := store.GetQueryWithQueryIdString(projectId, queryIdString)
	if errCode != http.StatusFound {
		errMsg = fmt.Sprintf("Failed to fetch query from query_id %v", queryIdString)
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
func (store *MemSQL) CacheDashboardUnit(dashboardUnit model.DashboardUnit, waitGroup *sync.WaitGroup, reportCollector *sync.Map, queryClass string) {
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
			logCtx.Errorf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
			continue
		}

		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := model.DecodeQueryForClass(queryInfo.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}

		// Filtering queries on type and range for attribution query
		shouldCache, from, to := model.ShouldCacheUnitForTimeRange(queryClass, preset, fr, t, C.GetOnlyAttributionDashboardCaching(), C.GetSkipAttributionDashboardCaching())
		if !shouldCache {
			continue
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
			unitWaitGroup.Add(1)
			go store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector)
		} else {
			store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

// CacheDashboardUnitForDateRange To cache a dashboard unit for the given range.
func (store *MemSQL) CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload) (int, string, model.CachingUnitReport) {
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

		result, errCode, errMsg = store.Analyze(projectID, *analyticsQuery)

		channel := make(chan Result)

		go store.runFunnelAndInsightsUnit(projectID, *analyticsQuery, channel)

		select {
		case response := <-channel:
			result = response.res
			err = response.err
			errCode = response.errCode
			errMsg = response.errMsg
			if err != nil {
				logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the FunnelORInsights unit")
				errCode = http.StatusInternalServerError
			} else {
				logCtx.WithFields(log.Fields{"Query": *analyticsQuery, "ErrCode": "UnitRunTimeOut"}).Info("Success for the FunnelORInsights unit")
				errCode = http.StatusOK
			}
		case <-time.After(16 * 60 * time.Second):
			queryTimedOut = true
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
			if err != nil && !model.IsIntegrationNotFoundError(response.err) {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut", "Error": response.err}).Info("Failed for the attribution unit")
				errCode = http.StatusInternalServerError
			} else {
				logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut"}).Info("Success for the attribution unit")
				errCode = http.StatusOK
			}
		case <-time.After(20 * 60 * time.Second):
			queryTimedOut = true
			logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": "UnitRunTimeOut"}).Info("Timeout for the attribution unit")
		}

	} else if baseQuery.GetClass() == model.QueryClassChannel {
		channelQuery := baseQuery.(*model.ChannelQueryUnit)
		unitReport.Query = channelQuery
		result, errCode = store.ExecuteChannelQuery(projectID, channelQuery.Query)
	} else if baseQuery.GetClass() == model.QueryClassChannelV1 {
		groupQuery := baseQuery.(*model.ChannelGroupQueryV1)
		unitReport.Query = groupQuery
		reqID := xid.New().String()
		result, errCode = store.RunChannelGroupQuery(projectID, groupQuery.Queries, reqID)
	} else if baseQuery.GetClass() == model.QueryClassEvents {
		groupQuery := baseQuery.(*model.QueryGroup)
		unitReport.Query = groupQuery
		result, errCode = store.RunEventsGroupQuery(groupQuery.Queries, projectID)
	} else if baseQuery.GetClass() == model.QueryClassKPI {
		groupQuery := baseQuery.(*model.KPIQueryGroup)
		unitReport.Query = groupQuery
		result, errCode = store.ExecuteKPIQueryGroup(projectID, "", *groupQuery)
	} else if baseQuery.GetClass() == model.QueryClassProfiles {
		groupQuery := baseQuery.(*model.ProfileQueryGroup)
		unitReport.Query = groupQuery
		result, errCode = store.RunProfilesGroupQuery(groupQuery.Queries, projectID)
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
	model.SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardID, dashboardUnitID, from, to, timezoneString)

	// Set in query cache result as well in case someone runs the same query from query handler.
	model.SetQueryCacheResult(projectID, baseQuery, result)
	unitReport.Status = model.CachingUnitStatusPassed
	unitReport.TimeTaken = timeTaken
	unitReport.TimeTakenStr = timeTakenString
	return http.StatusOK, "", unitReport
}

type Result struct {
	res     *model.QueryResult
	err     error
	errCode int
	errMsg  string
}

func (store *MemSQL) runFunnelAndInsightsUnit(projectID uint64, queryOriginal model.Query, c chan Result) {

	r, eCode, eMsg := store.Analyze(projectID, queryOriginal)
	result := Result{res: r, errCode: eCode, errMsg: eMsg}
	c <- result
}

func (store *MemSQL) runAttributionUnit(projectID uint64, queryOriginal *model.AttributionQuery, c chan Result) {

	r, err := store.ExecuteAttributionQuery(projectID, queryOriginal)
	result := Result{res: r, err: err, errMsg: ""}
	c <- result
}

func (store *MemSQL) cacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map) {
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
	errCode, errMsg, report := store.CacheDashboardUnitForDateRange(cachePayload)
	reportCollector.Store(model.GetCachingUnitReportUniqueKey(report), report)
	if errCode != http.StatusOK {
		logCtx.Errorf("Error while running query %s", errMsg)
		return
	}
	logCtx.Info("Completed caching for Dashboard unit")
}

// CacheDashboardsForMonthlyRange To cache monthly dashboards for the project id.
func (store *MemSQL) CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map) {
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
				logCtx.Errorf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
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
					go store.cacheDashboardUnitForDateRange(cachePayload, &waitGroup, reportCollector)
					if count%numRoutines == 0 {
						waitGroup.Wait()
						waitGroup.Add(U.MinInt(len(monthlyRanges)-count, numRoutines))
					}
				} else {
					store.cacheDashboardUnitForDateRange(cachePayload, &waitGroup, reportCollector)
				}
			}
			if C.GetIsRunningForMemsql() == 0 {
				waitGroup.Wait()
			}
		}
	}
}
