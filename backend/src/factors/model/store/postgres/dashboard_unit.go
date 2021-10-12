package postgres

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// CreateDashboardUnitForMultipleDashboards creates multiple dashboard units each for given
// list of dashboards
func (pg *Postgres) CreateDashboardUnitForMultipleDashboards(dashboardIds []uint64, projectId uint64,
	agentUUID string, unitPayload model.DashboardUnitRequestPayload) ([]*model.DashboardUnit, int, string) {

	var dashboardUnits []*model.DashboardUnit
	for _, dashboardId := range dashboardIds {
		dashboardUnit, errCode, errMsg := pg.CreateDashboardUnit(projectId, agentUUID,
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
func (pg *Postgres) CreateMultipleDashboardUnits(requestPayload []model.DashboardUnitRequestPayload, projectId uint64,
	agentUUID string, dashboardId uint64) ([]*model.DashboardUnit, int, string) {
	var dashboardUnits []*model.DashboardUnit
	for _, payload := range requestPayload {

		// query should have been created before the dashboard unit
		if payload.QueryId == 0 {
			return dashboardUnits, http.StatusBadRequest, "invalid queryID. empty queryID."
		}
		dashboardUnit, errCode, errMsg := pg.CreateDashboardUnit(projectId, agentUUID,
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

func (pg *Postgres) CreateDashboardUnit(projectId uint64, agentUUID string, dashboardUnit *model.DashboardUnit) (*model.DashboardUnit, int, string) {
	return pg.CreateDashboardUnitForDashboardClass(projectId, agentUUID, dashboardUnit, model.DashboardClassUserCreated)
}

func (pg *Postgres) CreateDashboardUnitForDashboardClass(projectId uint64, agentUUID string, dashboardUnit *model.DashboardUnit,
	dashboardClass string) (*model.DashboardUnit, int, string) {

	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"dashboard_unit": dashboardUnit, "project_id": projectId})
	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest, "Invalid request"
	}

	pg.updateDashboardUnitPresentation(dashboardUnit)

	valid, errMsg := model.IsValidDashboardUnit(dashboardUnit)
	if !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	hasAccess, dashboard := pg.HasAccessToDashboard(projectId, agentUUID, dashboardUnit.DashboardId)
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
func (pg *Postgres) updateDashboardUnitPresentation(unit *model.DashboardUnit) {
	logCtx := log.WithFields(log.Fields{
		"Method":    "updateDashboardUnitSettingsAndPresentation",
		"ProjectID": unit.ProjectID,
	})

	queryInfo, errC := pg.GetQueryWithQueryId(unit.ProjectID, unit.QueryId)
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
func (pg *Postgres) GetDashboardUnitsForProjectID(projectID uint64) ([]model.DashboardUnit, int) {
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

func (pg *Postgres) GetDashboardUnits(projectId uint64, agentUUID string, dashboardId uint64) ([]model.DashboardUnit, int) {
	db := C.GetServices().Db

	var dashboardUnits []model.DashboardUnit
	if projectId == 0 || dashboardId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id or agent_id")
		return dashboardUnits, http.StatusBadRequest
	}

	if hasAccess, _ := pg.HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
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
func (pg *Postgres) GetDashboardUnitByUnitID(projectID, unitID uint64) (*model.DashboardUnit, int) {
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

func (pg *Postgres) GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID, dashboardID uint64, types []string) ([]model.DashboardUnit, int) {
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

func (pg *Postgres) DeleteDashboardUnit(projectId uint64, agentUUID string, dashboardId uint64, id uint64) int {

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to delete dashboard unit. Invalid scope ids.")
		return http.StatusBadRequest
	}

	hasAccess, dashboard := pg.HasAccessToDashboard(projectId, agentUUID, dashboardId)
	if !hasAccess {
		return http.StatusForbidden
	}

	errCode := pg.removeUnitPositionOnDashboard(projectId, agentUUID, dashboardId, id, dashboard.UnitsPosition)
	if errCode != http.StatusAccepted {
		errMsg := "Failed remove position for unit on dashboard."
		log.WithFields(log.Fields{"project_id": projectId, "unitId": id}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
	}
	return pg.deleteDashboardUnit(projectId, dashboardId, id)
}

// DeleteMultipleDashboardUnits deletes multiple dashboard units for given dashboard
func (pg *Postgres) DeleteMultipleDashboardUnits(projectID uint64, agentUUID string, dashboardID uint64,
	dashboardUnitIDs []uint64) (int, string) {

	for _, dashboardUnitID := range dashboardUnitIDs {
		errCode := pg.DeleteDashboardUnit(projectID, agentUUID, dashboardID, dashboardUnitID)
		if errCode != http.StatusAccepted {
			errMsg := "Failed delete unit on dashboard."
			log.WithFields(log.Fields{"project_id": projectID,
				"dashboard_id": dashboardID, "unit_id": dashboardUnitID}).Error(errMsg)
			return errCode, errMsg
		}
	}
	return http.StatusAccepted, ""
}

func (pg *Postgres) deleteDashboardUnit(projectID uint64, dashboardID uint64, ID uint64) int {
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

func (pg *Postgres) UpdateDashboardUnit(projectId uint64, agentUUID string,
	dashboardId uint64, id uint64, unit *model.DashboardUnit) (*model.DashboardUnit, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "agentUUID": agentUUID, "dashboard_id": dashboardId})

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to update dashboard unit. Invalid scope ids.")
		return nil, http.StatusBadRequest
	}

	if hasAccess, _ := pg.HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
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
func (pg *Postgres) CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int) {
	logCtx := log.WithFields(log.Fields{
		"Method": "CacheDashboardUnitsForProjects",
	})

	projectIDs := pg.GetProjectsToRunForIncludeExcludeString(stringProjectsIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		logCtx = logCtx.WithFields(log.Fields{"ProjectID": projectID})
		logCtx.Info("Starting to cache units for the project")
		startTime := U.TimeNowUnix()
		unitsCount := pg.CacheDashboardUnitsForProjectID(projectID, numRoutines)

		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Time taken for caching %d dashboard units", unitsCount)
	}
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func (pg *Postgres) CacheDashboardUnitsForProjectID(projectID uint64, numRoutines int) int {
	if numRoutines == 0 {
		numRoutines = 1
	}

	dashboardUnits, errCode := pg.GetDashboardUnitsForProjectID(projectID)
	if errCode != http.StatusFound || len(dashboardUnits) == 0 {
		return 0
	}

	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(dashboardUnits), numRoutines))
	for i := range dashboardUnits {
		count++
		go pg.CacheDashboardUnit(dashboardUnits[i], &waitGroup)
		if count%numRoutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
		}
	}
	waitGroup.Wait()
	return len(dashboardUnits)
}

// GetQueryAndClassFromDashboardUnit returns query and query-class of dashboard unit.
func (pg *Postgres) GetQueryAndClassFromDashboardUnit(dashboardUnit *model.DashboardUnit) (queryClass string, queryInfo *model.Queries, errMsg string) {
	projectID := dashboardUnit.ProjectID
	savedQuery, errCode := pg.GetQueryWithQueryId(projectID, dashboardUnit.QueryId)
	if errCode != http.StatusFound {
		errMsg = fmt.Sprintf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
		return "", nil, errMsg
	}

	queryClass, errMsg = pg.GetQueryClassFromQueries(*savedQuery)
	if errMsg != "" {
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return "", nil, errMsg
	}
	return queryClass, savedQuery, ""
}

func (pg *Postgres) GetQueryClassFromQueries(query model.Queries) (queryClass, errMsg string) {
	var tempQuery model.Query
	var queryGroup model.QueryGroup
	// try decoding for Query
	U.DecodePostgresJsonbToStructType(&query.Query, &tempQuery)
	if tempQuery.Class == "" {
		// if fails, try decoding for QueryGroup
		err1 := U.DecodePostgresJsonbToStructType(&query.Query, &queryGroup)
		if err1 != nil {
			errMsg = fmt.Sprintf("Failed to decode jsonb query")
			return "", errMsg
		}
		queryClass = queryGroup.GetClass()
	} else {
		queryClass = tempQuery.Class
	}
	return queryClass, ""
}

// CacheDashboardUnit Caches query for given dashboard unit for default date range presets.
func (pg *Postgres) CacheDashboardUnit(dashboardUnit model.DashboardUnit, waitGroup *sync.WaitGroup) {

	logCtx := log.WithFields(log.Fields{
		"ProjectID":       dashboardUnit.ProjectID,
		"DashboardID":     dashboardUnit.DashboardId,
		"DashboardUnitID": dashboardUnit.ID,
	})
	defer waitGroup.Done()
	queryClass, _, errMsg := pg.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
	if errMsg != "" {
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}
	// excluding 'Web' class dashboard units
	if queryClass == model.QueryClassWeb {
		return
	}
	timezoneString, statusCode := pg.GetTimezoneForProject(dashboardUnit.ProjectID)
	if statusCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed to get project Timezone for %d", dashboardUnit.ProjectID)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}
	var unitWaitGroup sync.WaitGroup
	unitWaitGroup.Add(len(U.QueryDateRangePresets))
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to, errCode := rangeFunction(timezoneString)
		if errCode != nil {
			errMsg := fmt.Sprintf("Failed to get proper project Timezone for %d", dashboardUnit.ProjectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}

		queryInfo, errC := pg.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
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
		go pg.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup)
	}
	unitWaitGroup.Wait()
}

// CacheDashboardUnitForDateRange To cache a dashboard unit for the given range.
func (pg *Postgres) CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload) (int, string) {
	// Catches any panic in query execution and logs as an error. Prevents jobs from crashing.
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	timezoneString := baseQuery.GetTimeZone()
	from, to := baseQuery.GetQueryDateRange()
	logCtx := log.WithFields(log.Fields{
		"Method":          "CacheDashboardUnitForDateRange",
		"ProjectID":       projectID,
		"DashboardID":     dashboardID,
		"DashboardUnitID": dashboardUnitID,
		"FromTo":          fmt.Sprintf("%d-%d", from, to),
	})
	if !model.ShouldRefreshDashboardUnit(projectID, dashboardID, dashboardUnitID, from, to, timezoneString, false) {
		return http.StatusOK, ""
	}
	logCtx.Info("Starting to cache unit for date range")
	startTime := U.TimeNowUnix()

	var result interface{}
	var err error
	var errCode int
	var errMsg string
	if baseQuery.GetClass() == model.QueryClassFunnel || baseQuery.GetClass() == model.QueryClassInsights {
		analyticsQuery := baseQuery.(*model.Query)
		result, errCode, errMsg = pg.Analyze(projectID, *analyticsQuery)
	} else if baseQuery.GetClass() == model.QueryClassAttribution {
		attributionQuery := baseQuery.(*model.AttributionQueryUnit)
		result, err = pg.ExecuteAttributionQuery(projectID, attributionQuery.Query)
		if err != nil && !model.IsIntegrationNotFoundError(err) {
			errCode = http.StatusInternalServerError
		} else {
			errCode = http.StatusOK
		}
	} else if baseQuery.GetClass() == model.QueryClassChannel {
		channelQuery := baseQuery.(*model.ChannelQueryUnit)
		result, errCode = pg.ExecuteChannelQuery(projectID, channelQuery.Query)
	} else if baseQuery.GetClass() == model.QueryClassChannelV1 {
		groupQuery := baseQuery.(*model.ChannelGroupQueryV1)
		result, errCode = pg.RunChannelGroupQuery(projectID, groupQuery.Queries, "")
	} else if baseQuery.GetClass() == model.QueryClassEvents {
		groupQuery := baseQuery.(*model.QueryGroup)
		result, errCode = pg.RunEventsGroupQuery(groupQuery.Queries, projectID)
	} else if baseQuery.GetClass() == model.QueryClassKPI {
		groupQuery := baseQuery.(*model.KPIQueryGroup)
		result, errCode = pg.ExecuteKPIQueryGroup(projectID, "", *groupQuery)
	}
	if errCode != http.StatusOK {
		logCtx.Info("failed to run the query for dashboard caching")
		return http.StatusInternalServerError, fmt.Sprintf("Error while running query %s", errMsg)
	}

	timeTaken := U.TimeNowUnix() - startTime
	timeTakenString := U.SecondsToHMSString(timeTaken)
	logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
		Info("Done caching unit for range")
	model.SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardID, dashboardUnitID, from, to, timezoneString)

	// Set in query cache result as well in case someone runs the same query from query handler.
	model.SetQueryCacheResult(projectID, baseQuery, result)
	return http.StatusOK, ""
}

func (pg *Postgres) cacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
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
	errCode, errMsg := pg.CacheDashboardUnitForDateRange(cachePayload)
	if errCode != http.StatusOK {
		logCtx.Errorf("Error while running query %s", errMsg)
	}
}

// CacheDashboardsForMonthlyRange To cache monthly dashboards for the project id.
func (pg *Postgres) CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int) {
	projectIDsToRun := pg.GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs)
	for _, projectID := range projectIDsToRun {
		logCtx := log.WithFields(log.Fields{
			"Method":    "CacheDashboardUnit",
			"ProjectID": projectID,
		})
		timezoneString, statusCode := pg.GetTimezoneForProject(projectID)
		if statusCode != http.StatusFound {
			errMsg := fmt.Sprintf("Failed to get project Timezone for %d", projectID)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			continue
		}
		monthlyRanges := U.GetMonthlyQueryRangesTuplesZ(numMonths, timezoneString)
		dashboardUnits, errCode := pg.GetDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			return
		}

		for _, dashboardUnit := range dashboardUnits {
			queryInfo, errC := pg.GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
			if errC != http.StatusFound {
				logCtx.Errorf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
				continue
			}
			queryClass, errMsg := pg.GetQueryClassFromQueries(*queryInfo)
			if errMsg != "" {
				C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
				return
			}

			// excluding 'Web' class dashboard units
			if queryClass == model.QueryClassWeb {
				continue
			}

			var waitGroup sync.WaitGroup
			waitGroup.Add(U.MinInt(len(monthlyRanges), numRoutines))
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
				go pg.cacheDashboardUnitForDateRange(cachePayload, &waitGroup)
				if count%numRoutines == 0 {
					waitGroup.Wait()
					waitGroup.Add(U.MinInt(len(monthlyRanges)-count, numRoutines))
				}
			}
			waitGroup.Wait()
		}
	}
}
