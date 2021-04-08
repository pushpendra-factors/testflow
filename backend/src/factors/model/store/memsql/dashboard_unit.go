package memsql

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// CreateDashboardUnitForMultipleDashboards creates multiple dashboard units each for given
// list of dashboards
func (store *MemSQL) CreateDashboardUnitForMultipleDashboards(dashboardIds []uint64, projectId uint64,
	agentUUID string, unitPayload model.DashboardUnitRequestPayload) ([]*model.DashboardUnit, int, string) {

	var dashboardUnits []*model.DashboardUnit
	for _, dashboardId := range dashboardIds {
		dashboardUnit, errCode, errMsg := store.CreateDashboardUnit(projectId, agentUUID,
			&model.DashboardUnit{
				DashboardId:  dashboardId,
				Description:  unitPayload.Description,
				Query:        postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
				Title:        unitPayload.Title,
				Presentation: unitPayload.Presentation,
				QueryId:      unitPayload.QueryId,
				Settings:     *unitPayload.Settings,
			}, model.DashboardUnitWithQueryID)
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
				Query:        postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
				Title:        payload.Title,
				Presentation: payload.Presentation,
				QueryId:      payload.QueryId,
				Settings:     *payload.Settings,
			}, model.DashboardUnitWithQueryID)
		if errCode != http.StatusCreated {
			return nil, errCode, errMsg
		}
		dashboardUnits = append(dashboardUnits, dashboardUnit)
	}
	return dashboardUnits, http.StatusCreated, ""
}

func (store *MemSQL) CreateDashboardUnit(projectId uint64, agentUUID string, dashboardUnit *model.DashboardUnit,
	queryType string) (*model.DashboardUnit, int, string) {
	return store.CreateDashboardUnitForDashboardClass(projectId, agentUUID, dashboardUnit, queryType, model.DashboardClassUserCreated)
}

func (store *MemSQL) CreateDashboardUnitForDashboardClass(projectId uint64, agentUUID string, dashboardUnit *model.DashboardUnit,
	queryType, dashboardClass string) (*model.DashboardUnit, int, string) {
	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"dashboard_unit": dashboardUnit, "project_id": projectId})
	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest, "Invalid request"
	}

	updateDashboardUnitSettingsAndPresentation(dashboardUnit)

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

	// Todo (Anil) remove this query creation after we move to new UI completely.
	if dashboardUnit.QueryId == 0 {
		query, errCode, errMsg := store.CreateQuery(projectId,
			&model.Queries{
				Query: dashboardUnit.Query,
				Title: dashboardUnit.Title,
				Type:  model.QueryTypeDashboardQuery,
			})
		if errCode != http.StatusCreated {
			logCtx.Error(errMsg)
			return nil, errCode, errMsg
		}
		dashboardUnit.QueryId = query.ID
	} else {
		// Todo (Anil) for new UI requests, fill up Query using queryId for backward compatibility
		query, errCode := store.GetQueryWithQueryId(projectId, dashboardUnit.QueryId)
		// skip if error exists
		if errCode == http.StatusFound {
			queryJsonb, err := U.EncodeStructTypeToPostgresJsonb((*query).Query)
			if err == nil {
				dashboardUnit.Query = *queryJsonb
			}
		} else {
			errMsg = fmt.Sprintf("Failed to get query with id %d", dashboardUnit.QueryId)
			logCtx.Error(errMsg)
			return nil, errCode, errMsg
		}
	}
	dashboardUnit.ProjectID = projectId
	if err := db.Create(dashboardUnit).Error; err != nil {
		errMsg := "Failed to create dashboard unit."
		log.WithFields(log.Fields{"dashboard_unit": dashboardUnit,
			"project_id": projectId}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}

	// Todo (Anil) remove this DashboardUnitForNoQueryID based UnitPosition updating
	// ... after we move to new UI completely. todo
	if queryType == model.DashboardUnitForNoQueryID {
		errCode := store.addUnitPositionOnDashboard(projectId, agentUUID, dashboardUnit.DashboardId,
			dashboardUnit.ID, model.GetUnitType(dashboardUnit.Presentation), dashboard.UnitsPosition)
		if errCode != http.StatusAccepted {
			errMsg := "Failed add position for new dashboard unit."
			log.WithFields(log.Fields{"project_id": projectId,
				"dashboardUnitId": dashboardUnit.ID}).Error(errMsg)
			return nil, http.StatusInternalServerError, ""
		}
	}
	return dashboardUnit, http.StatusCreated, ""
}

// updateDashboardUnitSettingsAndPresentation updates Settings or Presentation for
// dashboard Unit using the other's value.
func updateDashboardUnitSettingsAndPresentation(unit *model.DashboardUnit) {

	if unit.Presentation != "" {
		// request is received from old UI updating Settings
		s := make(map[string]string)
		s["chart"] = unit.Presentation
		settings, err := json.Marshal(s)
		if err != nil {
			log.WithFields(log.Fields{"project_id": unit.ProjectID,
				"dashboardUnitId": unit.ID}).Error("failed to update settings for given presentation")
			return
		}
		unit.Settings = postgres.Jsonb{RawMessage: settings}
	} else {
		// request is received from new UI updating Presentation
		settings := make(map[string]string)
		err := json.Unmarshal(unit.Settings.RawMessage, &settings)
		if err != nil {
			log.WithFields(log.Fields{"project_id": unit.ProjectID,
				"dashboardUnitId": unit.ID}).Error("failed to update presentation for given settings")
			return
		}
		unit.Presentation = settings["chart"]
	}
}

// GetDashboardUnitsForProjectID Returns all dashboard units for the given projectID.
func (store *MemSQL) GetDashboardUnitsForProjectID(projectID uint64) ([]model.DashboardUnit, int) {
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

	dashboardUnits = store.fillQueryInDashboardUnits(dashboardUnits)

	return dashboardUnits, http.StatusFound
}

// Todo (Anil) Remove: Adding query using queryId for units from new UI
// fillQueryInDashboardUnits updates unit.Query by fetching query using queryId
func (store *MemSQL) fillQueryInDashboardUnits(units []model.DashboardUnit) []model.DashboardUnit {

	for i, unit := range units {
		query, errCode := store.GetQueryWithQueryId(unit.ProjectID, unit.QueryId)
		if errCode == http.StatusFound {
			queryJsonb, err := U.EncodeStructTypeToPostgresJsonb((*query).Query)
			if err == nil {
				units[i].Query = *queryJsonb
			}
		}
	}
	return units
}

func (store *MemSQL) GetDashboardUnits(projectId uint64, agentUUID string, dashboardId uint64) ([]model.DashboardUnit, int) {
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

	dashboardUnits = store.fillQueryInDashboardUnits(dashboardUnits)

	return dashboardUnits, http.StatusFound
}

// GetDashboardUnitByUnitID To get a dashboard unit by project id and unit id.
func (store *MemSQL) GetDashboardUnitByUnitID(projectID, unitID uint64) (*model.DashboardUnit, int) {
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

	dashboardUnits = store.fillQueryInDashboardUnits(dashboardUnits)

	return dashboardUnits, http.StatusFound
}

func (store *MemSQL) DeleteDashboardUnit(projectId uint64, agentUUID string, dashboardId uint64, id uint64) int {

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
	db := C.GetServices().Db
	// Required for getting query_id.
	dashboardUnit, errCode := store.GetDashboardUnitByUnitID(projectID, ID)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError
	}

	err := db.Model(&model.DashboardUnit{}).Where("id = ? AND project_id = ? AND dashboard_id = ?",
		ID, projectID, dashboardID).Update(map[string]interface{}{"is_deleted": true}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID,
			"unit_id": ID}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}

	// Removing dashboard saved query.
	errCode, errMsg := store.DeleteDashboardQuery(projectID, dashboardUnit.QueryId)
	if errCode != http.StatusAccepted {
		log.WithFields(log.Fields{"project_id": projectID, "unitId": ID}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
	}
	return http.StatusAccepted
}

func (store *MemSQL) UpdateDashboardUnit(projectId uint64, agentUUID string,
	dashboardId uint64, id uint64, unit *model.DashboardUnit) (*model.DashboardUnit, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "agentUUID": agentUUID, "dashboard_id": dashboardId})

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
	if unit.Title != "" {
		updateFields["title"] = unit.Title
	}
	if unit.Description != "" {
		updateFields["description"] = unit.Description
	}
	if unit.Presentation != "" {
		updateFields["presentation"] = unit.Presentation
	}
	if !U.IsEmptyPostgresJsonb(&unit.Settings) {
		updateFields["settings"] = unit.Settings
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
	// update query table
	var dashboardUnit model.DashboardUnit
	err = db.Model(&model.DashboardUnit{}).Where("id = ? AND project_id = ? AND dashboard_id = ? AND is_deleted = ?",
		id, projectId, dashboardId, false).Find(&dashboardUnit).Error
	_, errCode := store.UpdateSavedQuery(projectId, dashboardUnit.QueryId, &model.Queries{Title: unit.Title, Type: model.QueryTypeDashboardQuery})
	if errCode != http.StatusAccepted {
		logCtx.WithError(err).Error("updatedDashboardUnitFields failed at UpdateSavedQuery in queries.go")
		return nil, errCode
	}

	// returns only updated fields, avoid using it on model.DashboardUnit API.
	return &updatedDashboardUnitFields, http.StatusAccepted
}

// CacheDashboardUnitsForProjects Runs for all the projectIDs passed as comma separated.
func (store *MemSQL) CacheDashboardUnitsForProjects(stringProjectsIDs, excludeProjectIDs string, numRoutines int) {
	logCtx := log.WithFields(log.Fields{
		"Method": "CacheDashboardUnitsForProjects",
	})

	projectIDs := store.GetProjectsToRunForIncludeExcludeString(stringProjectsIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		logCtx = logCtx.WithFields(log.Fields{"ProjectID": projectID})
		logCtx.Info("Starting to cache units for the project")
		startTime := U.TimeNowUnix()
		unitsCount := store.CacheDashboardUnitsForProjectID(projectID, numRoutines)

		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Time taken for caching %d dashboard units", unitsCount)
	}
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func (store *MemSQL) CacheDashboardUnitsForProjectID(projectID uint64, numRoutines int) int {
	if numRoutines == 0 {
		numRoutines = 1
	}

	dashboardUnits, errCode := store.GetDashboardUnitsForProjectID(projectID)
	if errCode != http.StatusFound || len(dashboardUnits) == 0 {
		return 0
	}

	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(dashboardUnits), numRoutines))
	for i := range dashboardUnits {
		count++
		go store.CacheDashboardUnit(dashboardUnits[i], &waitGroup)
		if count%numRoutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
		}
	}
	waitGroup.Wait()
	return len(dashboardUnits)
}

// GetQueryAndClassFromDashboardUnit Fill query and returns query class of dashboard unit.
func (store *MemSQL) GetQueryAndClassFromDashboardUnit(dashboardUnit *model.DashboardUnit) (queryClass, errMsg string) {
	projectID := dashboardUnit.ProjectID
	savedQuery, errCode := store.GetQueryWithQueryId(projectID, dashboardUnit.QueryId)
	if errCode != http.StatusFound {
		errMsg = fmt.Sprintf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
		return
	}
	dashboardUnit.Query = savedQuery.Query

	var query model.Query
	var queryGroup model.QueryGroup
	// try decoding for Query
	U.DecodePostgresJsonbToStructType(&savedQuery.Query, &query)
	if query.Class == "" {
		// if fails, try decoding for QueryGroup
		err1 := U.DecodePostgresJsonbToStructType(&savedQuery.Query, &queryGroup)
		if err1 != nil {
			errMsg = fmt.Sprintf("Failed to decode jsonb query, query_id %d", dashboardUnit.QueryId)
			return
		}
		queryClass = queryGroup.GetClass()
	} else {
		queryClass = query.Class
	}
	return
}

// CacheDashboardUnit Caches query for given dashboard unit for default date range presets.
func (store *MemSQL) CacheDashboardUnit(dashboardUnit model.DashboardUnit, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	queryClass, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
	if errMsg != "" {
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}

	// excluding 'Web' class dashboard units
	if queryClass == model.QueryClassWeb {
		return
	}

	var unitWaitGroup sync.WaitGroup
	unitWaitGroup.Add(len(U.QueryDateRangePresets))
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := model.DecodeQueryForClass(dashboardUnit.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		baseQuery.SetQueryDateRange(from, to)
		cachePayload := model.DashboardUnitCachePayload{
			DashboardUnit: dashboardUnit,
			BaseQuery:     baseQuery,
		}
		go store.cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup)
	}
	unitWaitGroup.Wait()
}

// CacheDashboardUnitForDateRange To cache a dashboard unit for the given range.
func (store *MemSQL) CacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload) (int, string) {
	dashboardUnit := cachePayload.DashboardUnit
	baseQuery := cachePayload.BaseQuery
	projectID := dashboardUnit.ProjectID
	dashboardID := dashboardUnit.DashboardId
	dashboardUnitID := dashboardUnit.ID
	from, to := baseQuery.GetQueryDateRange()
	logCtx := log.WithFields(log.Fields{
		"Method":          "CacheDashboardUnitForDateRange",
		"ProjectID":       projectID,
		"DashboardID":     dashboardID,
		"DashboardUnitID": dashboardUnitID,
		"FromTo":          fmt.Sprintf("%d-%d", from, to),
	})
	if !model.ShouldRefreshDashboardUnit(projectID, dashboardID, dashboardUnitID, from, to, false) {
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
		result, errCode, errMsg = store.Analyze(projectID, *analyticsQuery)
	} else if baseQuery.GetClass() == model.QueryClassAttribution {
		attributionQuery := baseQuery.(*model.AttributionQueryUnit)
		result, err = store.ExecuteAttributionQuery(projectID, attributionQuery.Query)
		if err != nil {
			errCode = http.StatusInternalServerError
		} else {
			errCode = http.StatusOK
		}
	} else if baseQuery.GetClass() == model.QueryClassChannel {
		channelQuery := baseQuery.(*model.ChannelQueryUnit)
		result, errCode = store.ExecuteChannelQuery(projectID, channelQuery.Query)
	} else if baseQuery.GetClass() == model.QueryClassChannelV1 {
		groupQuery := baseQuery.(*model.ChannelGroupQueryV1)
		result, errCode = store.RunChannelGroupQuery(projectID, groupQuery.Queries, "")
	} else if baseQuery.GetClass() == model.QueryClassEvents {
		groupQuery := baseQuery.(*model.QueryGroup)
		result, errCode = store.RunEventsGroupQuery(groupQuery.Queries, projectID)
	}
	if errCode != http.StatusOK {
		return http.StatusInternalServerError, fmt.Sprintf("Error while running query %s", errMsg)
	}

	timeTaken := U.TimeNowUnix() - startTime
	timeTakenString := U.SecondsToHMSString(timeTaken)
	logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
		Info("Done caching unit for range")
	model.SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardID, dashboardUnitID, from, to)

	// Set in query cache result as well in case someone runs the same query from query handler.
	model.SetQueryCacheResult(projectID, baseQuery, result)
	return http.StatusOK, ""
}

func (store *MemSQL) cacheDashboardUnitForDateRange(cachePayload model.DashboardUnitCachePayload,
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
	errCode, errMsg := store.CacheDashboardUnitForDateRange(cachePayload)
	if errCode != http.StatusOK {
		logCtx.Errorf("Error while running query %s", errMsg)
	}
}

// CacheDashboardsForMonthlyRange To cache monthly dashboards for the project id.
func (store *MemSQL) CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs string, numMonths, numRoutines int) {
	projectIDsToRun := store.GetProjectsToRunForIncludeExcludeString(projectIDs, excludeProjectIDs)
	monthlyRanges := U.GetMonthlyQueryRangesTuplesIST(numMonths)
	for _, projectID := range projectIDsToRun {
		dashboardUnits, errCode := store.GetDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound || len(dashboardUnits) == 0 {
			return
		}

		for _, dashboardUnit := range dashboardUnits {
			queryClass, errMsg := store.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg != "" {
				C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
				continue
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
				baseQuery, err := model.DecodeQueryForClass(dashboardUnit.Query, queryClass)
				if err != nil {
					errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
					C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
					return
				}
				baseQuery.SetQueryDateRange(from, to)
				cachePayload := model.DashboardUnitCachePayload{
					DashboardUnit: dashboardUnit,
					BaseQuery:     baseQuery,
				}
				go store.cacheDashboardUnitForDateRange(cachePayload, &waitGroup)
				if count%numRoutines == 0 {
					waitGroup.Wait()
					waitGroup.Add(U.MinInt(len(monthlyRanges)-count, numRoutines))
				}
			}
			waitGroup.Wait()
		}
	}
}
