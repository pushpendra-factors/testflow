package postgres

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
	log "github.com/sirupsen/logrus"
)

const (
	DashboardCacheInvalidationDuration14DaysInSecs = 14 * 24 * 60 * 60
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
func (pg *Postgres) CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map) {
	logFields := log.Fields{
		"string_projects_ids": stringProjectsIDs,
		"exclude_project_ids": excludeProjectIDs,
		"num_routines":        numRoutines,
		"report_collector":    reportCollector,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	projectIDs := pg.GetProjectsToRunForIncludeExcludeString(stringProjectsIDs, excludeProjectIDs)
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

		dashboardUnits, errCode := pg.GetDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			logCtx.Info("not running caching for the project - units not found")
			continue
		}

		filterDashboardUnits := make([]model.DashboardUnit, 0)
		filterDashboardUnitQueryClass := make([]string, 0)
		for _, dashboardUnit := range dashboardUnits {

			queryClass, _, errMsg := pg.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
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
			unitsCount = pg.CacheDashboardUnitsForProjectID(projectID, validDashboardUnitIDs, validDashboardUnitQueryClass, numRoutines, reportCollector)

		} else {

			log.WithFields(log.Fields{"project_id": projectID, "total_units": len(filterDashboardUnits), "accessed_units": len(filterDashboardUnits)}).Info("Project Report - normal run")
			unitsCount = pg.CacheDashboardUnitsForProjectID(projectID, filterDashboardUnits, filterDashboardUnitQueryClass, numRoutines, reportCollector)
		}
		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		log.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Project Report: Time taken for caching %d dashboard units", unitsCount)
	}
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func (pg *Postgres) CacheDashboardUnitsForProjectID(projectID uint64, dashboardUnits []model.DashboardUnit,
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
			go pg.CacheDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i])
			if count%numRoutines == 0 {
				waitGroup.Wait()
				waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
			}
		} else {
			pg.CacheDashboardUnit(dashboardUnits[i], &waitGroup, reportCollector, dashboardUnitQueryClass[i])
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Wait()
	}
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

// GetQueryAndClassFromDashboardUnit returns query and query-class of dashboard unit.
func (pg *Postgres) GetQueryAndClassFromQueryIdString(queryIdString string, projectId uint64) (queryClass string, queryInfo *model.Queries, errMsg string) {
	savedQuery, errCode := pg.GetQueryWithQueryIdString(projectId, queryIdString)
	if errCode != http.StatusFound {
		errMsg = fmt.Sprintf("Failed to fetch query from query_id %v", queryIdString)
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
func (pg *Postgres) CacheDashboardUnit(dashboardUnit model.DashboardUnit, waitGroup *sync.WaitGroup, reportCollector *sync.Map, queryClass string) {
	logCtx := log.WithFields(log.Fields{
		"ProjectID":       dashboardUnit.ProjectID,
		"DashboardID":     dashboardUnit.DashboardId,
		"DashboardUnitID": dashboardUnit.ID,
	})
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
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
	for preset, rangeFunction := range U.QueryDateRangePresets {

		fr, t, errCode := rangeFunction(timezoneString)
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
			go pg.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector)
		} else {
			pg.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup, reportCollector)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		unitWaitGroup.Wait()
	}
}

// CacheDashboardUnitForDateRange To cache a dashboard unit for the given range.
func (pg *Postgres) CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload) (int, string, model.CachingUnitReport) {
	// Catches any panic in query execution and logs as an error. Prevent jobs from crashing.
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	timezoneString := baseQuery.GetTimeZone()
	from, to := baseQuery.GetQueryDateRange()

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

	logCtx := log.WithFields(log.Fields{
		"Method":          "CacheDashboardUnitForDateRange",
		"ProjectID":       projectID,
		"DashboardID":     dashboardID,
		"DashboardUnitID": dashboardUnitID,
		"FromTo":          fmt.Sprintf("%d-%d", from, to),
	})
	if !model.ShouldRefreshDashboardUnit(projectID, dashboardID, dashboardUnitID, from, to, timezoneString, false) {
		return http.StatusOK, "", unitReport
	}
	logCtx.Info("Starting to cache unit for date range")
	startTime := U.TimeNowUnix()

	var result interface{}
	var err error
	var errCode int
	var errMsg string
	if baseQuery.GetClass() == model.QueryClassFunnel || baseQuery.GetClass() == model.QueryClassInsights {
		analyticsQuery := baseQuery.(*model.Query)
		unitReport.Query = analyticsQuery
		result, errCode, errMsg = pg.Analyze(projectID, *analyticsQuery)
	} else if baseQuery.GetClass() == model.QueryClassAttribution {
		attributionQuery := baseQuery.(*model.AttributionQueryUnit)
		unitReport.Query = attributionQuery
		var debugQueryKey string
		result, err = pg.ExecuteAttributionQuery(projectID, attributionQuery.Query, debugQueryKey)
		logCtx.WithFields(log.Fields{"Query": attributionQuery.Query, "ErrCode": err}).Info("Got attribution result")
		if err != nil && !model.IsIntegrationNotFoundError(err) {
			errCode = http.StatusInternalServerError
		} else {
			errCode = http.StatusOK
		}
	} else if baseQuery.GetClass() == model.QueryClassChannel {
		channelQuery := baseQuery.(*model.ChannelQueryUnit)
		unitReport.Query = channelQuery
		result, errCode = pg.ExecuteChannelQuery(projectID, channelQuery.Query)
	} else if baseQuery.GetClass() == model.QueryClassChannelV1 {
		groupQuery := baseQuery.(*model.ChannelGroupQueryV1)
		unitReport.Query = groupQuery
		result, errCode = pg.RunChannelGroupQuery(projectID, groupQuery.Queries, "")
	} else if baseQuery.GetClass() == model.QueryClassEvents {
		groupQuery := baseQuery.(*model.QueryGroup)
		unitReport.Query = groupQuery
		result, errCode = pg.RunEventsGroupQuery(groupQuery.Queries, projectID)
	} else if baseQuery.GetClass() == model.QueryClassKPI {
		groupQuery := baseQuery.(*model.KPIQueryGroup)
		unitReport.Query = groupQuery
		result, errCode = pg.ExecuteKPIQueryGroup(projectID, "", *groupQuery)
	}
	if errCode != http.StatusOK {
		logCtx.WithField("QueryClass", baseQuery.GetClass()).WithField("Query", unitReport.Query).Info("failed to run the query for dashboard caching")
		unitReport.Status = model.CachingUnitStatusFailed
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

func (pg *Postgres) cacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
	waitGroup *sync.WaitGroup, reportCollector *sync.Map) {
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
	errCode, errMsg, report := pg.CacheDashboardUnitForDateRange(cachePayload)
	reportCollector.Store(model.GetCachingUnitReportUniqueKey(report), report)
	if errCode != http.StatusOK {
		logCtx.Errorf("Error while running query %s", errMsg)
		return
	}
	logCtx.Info("Completed caching for Dashboard unit")
}

// CacheDashboardsForMonthlyRange To cache monthly dashboards for the project id.
func (pg *Postgres) CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map) {
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
					go pg.cacheDashboardUnitForDateRange(cachePayload, &waitGroup, reportCollector)
					if count%numRoutines == 0 {
						waitGroup.Wait()
						waitGroup.Add(U.MinInt(len(monthlyRanges)-count, numRoutines))
					}
				} else {
					pg.cacheDashboardUnitForDateRange(cachePayload, &waitGroup, reportCollector)
				}

			}
			if C.GetIsRunningForMemsql() == 0 {
				waitGroup.Wait()
			}
		}
	}
}
