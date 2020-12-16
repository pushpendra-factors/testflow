package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type DashboardUnit struct {
	// Composite primary key, id + project_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboard_units(project_id) ref projects(id).
	ProjectID    uint64    `gorm:"primary_key:true" json:"project_id"`
	DashboardId  uint64    `gorm:"primary_key:true" json:"dashboard_id"`
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `json:"description"`
	Presentation string    `gorm:"type:varchar(5);not null" json:"presentation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// TODO (Anil) remove this field once we move to saved queries
	Query    postgres.Jsonb `gorm:"not null" json:"query"`
	QueryId  uint64         `gorm:"not null" json:"query_id"`
	Settings postgres.Jsonb `json:"settings"`
}

type DashboardUnitRequestPayload struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	// TODO (Anil) remove this field once we move to saved queries
	Query    *postgres.Jsonb `json:"query"`
	QueryId  uint64          `json:"query_id"`
	Settings *postgres.Jsonb `json:"settings"`
}

type DashboardCacheResult struct {
	Result      interface{} `json:"result"`
	From        int64       `json:"from"`
	To          int64       `json:"tom"`
	RefreshedAt int64       `json:"refreshed_at"`
}

// DashboardUnitCachePayload Payload for dashboard caching method.
type DashboardUnitCachePayload struct {
	dashboardUnit DashboardUnit
	baseQuery     BaseQuery
}

const (
	PresentationLine   = "pl"
	PresentationBar    = "pb"
	PresentationTable  = "pt"
	PresentationCard   = "pc"
	PresentationFunnel = "pf"

	DashboardUnitForNoQueryID = "NoQueryID"
	DashboardUnitWithQueryID  = "WithQueryID"
)

var presentations = [...]string{PresentationLine, PresentationBar,
	PresentationTable, PresentationCard, PresentationFunnel}

const (
	UnitCard  = "card"
	UnitChart = "chart"
)

var UnitPresentationTypes = [...]string{UnitCard, UnitChart}

func isValidDashboardUnit(dashboardUnit *DashboardUnit) (bool, string) {
	if dashboardUnit.DashboardId == 0 {
		return false, "Invalid dashboard"
	}

	if dashboardUnit.Title == "" {
		return false, "Invalid title"
	}

	validPresentation := false
	for _, p := range presentations {
		if p == dashboardUnit.Presentation {
			validPresentation = true
			break
		}
	}
	if !validPresentation {
		return false, "Invalid presentation"
	}

	// Todo(Dinesh): Validate query based on query class here.

	return true, ""
}

func GetUnitType(presentation string) string {
	if presentation == PresentationCard {
		return UnitCard
	}

	return UnitChart
}

// CreateDashboardUnitForMultipleDashboards creates multiple dashboard units each for given
// list of dashboards
func CreateDashboardUnitForMultipleDashboards(dashboardIds []uint64, projectId uint64,
	agentUUID string, unitPayload DashboardUnitRequestPayload) ([]*DashboardUnit, int, string) {

	var dashboardUnits []*DashboardUnit
	for _, dashboardId := range dashboardIds {
		dashboardUnit, errCode, errMsg := CreateDashboardUnit(projectId, agentUUID,
			&DashboardUnit{
				DashboardId:  dashboardId,
				Description:  unitPayload.Description,
				Query:        postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
				Title:        unitPayload.Title,
				Presentation: unitPayload.Presentation,
				QueryId:      unitPayload.QueryId,
				Settings:     *unitPayload.Settings,
			}, DashboardUnitWithQueryID)
		if errCode != http.StatusCreated {
			return nil, errCode, errMsg
		}
		dashboardUnits = append(dashboardUnits, dashboardUnit)
	}
	return dashboardUnits, http.StatusCreated, ""
}

// CreateMultipleDashboardUnits creates multiple dashboard units for list of queries for single dashboard
func CreateMultipleDashboardUnits(requestPayload []DashboardUnitRequestPayload, projectId uint64,
	agentUUID string, dashboardId uint64) ([]*DashboardUnit, int, string) {
	var dashboardUnits []*DashboardUnit
	for _, payload := range requestPayload {

		// query should have been created before the dashboard unit
		if payload.QueryId == 0 {
			return dashboardUnits, http.StatusBadRequest, "invalid queryID. empty queryID."
		}
		dashboardUnit, errCode, errMsg := CreateDashboardUnit(projectId, agentUUID,
			&DashboardUnit{
				DashboardId:  dashboardId,
				Description:  payload.Description,
				Query:        postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
				Title:        payload.Title,
				Presentation: payload.Presentation,
				QueryId:      payload.QueryId,
				Settings:     *payload.Settings,
			}, DashboardUnitWithQueryID)
		if errCode != http.StatusCreated {
			return nil, errCode, errMsg
		}
		dashboardUnits = append(dashboardUnits, dashboardUnit)
	}
	return dashboardUnits, http.StatusCreated, ""
}

func CreateDashboardUnit(projectId uint64, agentUUID string, dashboardUnit *DashboardUnit, queryType string) (*DashboardUnit, int, string) {
	db := C.GetServices().Db

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest, "Invalid request"
	}

	updateDashboardUnitSettingsAndPresentation(dashboardUnit)

	valid, errMsg := isValidDashboardUnit(dashboardUnit)
	if !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	hasAccess, dashboard := HasAccessToDashboard(projectId, agentUUID, dashboardUnit.DashboardId)
	if !hasAccess {
		return nil, http.StatusForbidden, "Unauthorized to access dashboard"
	}
	// Todo (Anil) remove this query creation after we move to new UI completely.
	if dashboardUnit.QueryId == 0 {
		query, errCode, errMsg := CreateQuery(projectId,
			&Queries{
				Query: dashboardUnit.Query,
				Title: dashboardUnit.Title,
				Type:  QueryTypeDashboardQuery,
			})
		if errCode != http.StatusCreated {
			log.WithFields(log.Fields{"dashboard_unit": dashboardUnit,
				"project_id": projectId}).Error(errMsg)
			return nil, errCode, errMsg
		}
		dashboardUnit.QueryId = query.ID
	} else {
		// Todo (Anil) for new UI requests, fill up Query using queryId for backward compatibility
		query, errCode := GetQueryWithQueryId(dashboardUnit.ProjectID, dashboardUnit.QueryId)
		// skip if error exists
		if errCode == http.StatusFound {
			queryJsonb, err := U.EncodeStructTypeToPostgresJsonb((*query).Query)
			if err == nil {
				dashboardUnit.Query = *queryJsonb
			}
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
	if queryType == DashboardUnitForNoQueryID {
		errCode := addUnitPositionOnDashboard(projectId, agentUUID, dashboardUnit.DashboardId,
			dashboardUnit.ID, GetUnitType(dashboardUnit.Presentation), dashboard.UnitsPosition)
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
func updateDashboardUnitSettingsAndPresentation(unit *DashboardUnit) {

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
		unit.Settings = postgres.Jsonb{settings}
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
func GetDashboardUnitsForProjectID(projectID uint64) ([]DashboardUnit, int) {
	db := C.GetServices().Db

	var dashboardUnits []DashboardUnit
	if projectID == 0 {
		log.Errorf("Invalid project id %d", projectID)
		return dashboardUnits, http.StatusBadRequest
	} else if err := db.Where("project_id = ?", projectID).Find(&dashboardUnits).Error; err != nil {
		log.WithError(err).Errorf("Failed to get dashboard units for projectID %d", projectID)
		return dashboardUnits, http.StatusInternalServerError
	}

	dashboardUnits = fillQueryInDashboardUnits(dashboardUnits)

	return dashboardUnits, http.StatusFound
}

// Todo (Anil) Remove: Adding query using queryId for units from new UI
// fillQueryInDashboardUnits updates unit.Query by fetching query using queryId
func fillQueryInDashboardUnits(units []DashboardUnit) []DashboardUnit {

	for i, unit := range units {
		query, errCode := GetQueryWithQueryId(unit.ProjectID, unit.QueryId)
		if errCode == http.StatusFound {
			queryJsonb, err := U.EncodeStructTypeToPostgresJsonb((*query).Query)
			if err == nil {
				units[i].Query = *queryJsonb
			}
		}
	}
	return units
}

func GetDashboardUnits(projectId uint64, agentUUID string, dashboardId uint64) ([]DashboardUnit, int) {
	db := C.GetServices().Db

	var dashboardUnits []DashboardUnit
	if projectId == 0 || dashboardId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id or agent_id")
		return dashboardUnits, http.StatusBadRequest
	}

	if hasAccess, _ := HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
		return nil, http.StatusForbidden
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ?",
		projectId, dashboardId).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	dashboardUnits = fillQueryInDashboardUnits(dashboardUnits)

	return dashboardUnits, http.StatusFound
}

func getDashboardUnitQueryResultCacheKey(projectID, dashboardID, unitID uint64, from, to int64) (*cacheRedis.Key, error) {
	prefix := "dashboard:query"
	var suffix string
	if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
		// Query for today's dashboard. Use to as 'now'.
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:now", dashboardID, unitID, from)
	} else {
		suffix = fmt.Sprintf("did:%d:duid:%d:from:%d:to:%d", dashboardID, unitID, from, to)
	}
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

// GetDashboardUnitByUnitID To get a dashboard unit by project id and unit id.
func GetDashboardUnitByUnitID(projectID, unitID uint64) (*DashboardUnit, int) {
	db := C.GetServices().Db
	var dashboardUnit DashboardUnit
	if err := db.Model(&DashboardUnit{}).Where("project_id = ? AND id=?",
		projectID, unitID).Find(&dashboardUnit).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &dashboardUnit, http.StatusFound
}

func GetDashboardUnitsByProjectIDAndDashboardIDAndTypes(projectID, dashboardID uint64, types []string) ([]DashboardUnit, int) {
	db := C.GetServices().Db

	var dashboardUnits []DashboardUnit
	if projectID == 0 || dashboardID == 0 {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id ")
		return dashboardUnits, http.StatusBadRequest
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ? ",
		projectID, dashboardID).Where("presentation IN (?)", types).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectID).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	if len(dashboardUnits) == 0 {
		return dashboardUnits, http.StatusNotFound
	}

	dashboardUnits = fillQueryInDashboardUnits(dashboardUnits)

	return dashboardUnits, http.StatusFound
}

func DeleteDashboardUnit(projectId uint64, agentUUID string, dashboardId uint64, id uint64) int {

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to delete dashboard unit. Invalid scope ids.")
		return http.StatusBadRequest
	}

	hasAccess, dashboard := HasAccessToDashboard(projectId, agentUUID, dashboardId)
	if !hasAccess {
		return http.StatusForbidden
	}

	errCode := removeUnitPositionOnDashboard(projectId, agentUUID, dashboardId, id, dashboard.UnitsPosition)
	if errCode != http.StatusAccepted {
		errMsg := "Failed remove position for unit on dashboard."
		log.WithFields(log.Fields{"project_id": projectId, "unitId": id}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
	}
	return deleteDashboardUnit(projectId, dashboardId, id)
}

// DeleteMultipleDashboardUnits deletes multiple dashboard units for given dashboard
func DeleteMultipleDashboardUnits(projectID uint64, agentUUID string, dashboardID uint64,
	dashboardUnitIDs []uint64) (int, string) {

	for _, dashboardUnitID := range dashboardUnitIDs {
		errCode := DeleteDashboardUnit(projectID, agentUUID, dashboardID, dashboardUnitID)
		if errCode != http.StatusAccepted {
			errMsg := "Failed delete unit on dashboard."
			log.WithFields(log.Fields{"project_id": projectID,
				"dashboard_id": dashboardID, "unit_id": dashboardUnitID}).Error(errMsg)
			return errCode, errMsg
		}
	}
	return http.StatusAccepted, ""
}

func deleteDashboardUnit(projectID uint64, dashboardID uint64, ID uint64) int {
	db := C.GetServices().Db
	// Required for getting query_id.
	dashboardUnit, errCode := GetDashboardUnitByUnitID(projectID, ID)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError
	}

	err := db.Where("id = ? AND project_id = ? AND dashboard_id = ?",
		ID, projectID, dashboardID).Delete(&dashboardUnit).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID, "dashboard_id": dashboardID,
			"unit_id": ID}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}

	// removing dashboard saved query
	errCode, errMsg := DeleteDashboardQuery(projectID, dashboardUnit.QueryId)
	if errCode != http.StatusAccepted {
		log.WithFields(log.Fields{"project_id": projectID, "unitId": ID}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
	}
	return http.StatusAccepted
}

func UpdateDashboardUnit(projectId uint64, agentUUID string,
	dashboardId uint64, id uint64, unit *DashboardUnit) (*DashboardUnit, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "agentUUID": agentUUID, "dashboard_id": dashboardId})

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to update dashboard unit. Invalid scope ids.")
		return nil, http.StatusBadRequest
	}

	if hasAccess, _ := HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
		return nil, http.StatusForbidden
	}

	db := C.GetServices().Db

	// update allowed fields.
	updateFields := make(map[string]interface{}, 0)
	if unit.Title != "" {
		updateFields["title"] = unit.Title
	}

	// nothing to update.
	if len(updateFields) == 0 {
		return nil, http.StatusBadRequest
	}
	var updatedDashboardUnitFields DashboardUnit
	err := db.Model(&updatedDashboardUnitFields).Where("id = ? AND project_id = ? AND dashboard_id = ?",
		id, projectId, dashboardId).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error("updatedDashboardUnitFields failed at UpdateDashboardUnit in dashboard_unit.go")
		return nil, http.StatusInternalServerError
	}
	// update query table
	var dashboardUnit DashboardUnit
	err = db.Model(&DashboardUnit{}).Where("id = ? AND project_id = ? AND dashboard_id = ?",
		id, projectId, dashboardId).Find(&dashboardUnit).Error
	_, errCode := UpdateSavedQuery(projectId, dashboardUnit.QueryId, &Queries{Title: unit.Title, Type: QueryTypeDashboardQuery})
	if errCode != http.StatusAccepted {
		logCtx.WithError(err).Error("updatedDashboardUnitFields failed at UpdateSavedQuery in queries.go")
		return nil, errCode
	}

	// returns only updated fields, avoid using it on DashboardUnit API.
	return &updatedDashboardUnitFields, http.StatusAccepted
}

// CacheDashboardUnitsForProjects Runs for all the projectIDs passed as comma separated.
func CacheDashboardUnitsForProjects(stringProjectsIDs string, numRoutines int) {
	logCtx := log.WithFields(log.Fields{
		"Method": "CacheDashboardUnitsForProjects",
	})

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(stringProjectsIDs, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	if allProjects {
		var errCode int
		projectIDs, errCode = GetAllProjectIDs()
		if errCode != http.StatusFound {
			return
		}
	}

	for _, projectID := range projectIDs {
		logCtx = logCtx.WithFields(log.Fields{"ProjectID": projectID})
		logCtx.Info("Starting to cache units for the project")
		startTime := U.TimeNowUnix()
		unitsCount := CacheDashboardUnitsForProjectID(projectID, numRoutines)

		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
			Infof("Time taken for caching %d dashboard units", unitsCount)
	}
	return
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func CacheDashboardUnitsForProjectID(projectID uint64, numRoutines int) int {
	if numRoutines == 0 {
		numRoutines = 1
	}

	dashboardUnits, errCode := GetDashboardUnitsForProjectID(projectID)
	if errCode != http.StatusFound || len(dashboardUnits) == 0 {
		return 0
	}

	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(dashboardUnits), numRoutines))
	for i := range dashboardUnits {
		count++
		go CacheDashboardUnit(dashboardUnits[i], &waitGroup)
		if count%numRoutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(dashboardUnits)-count, numRoutines))
		}
	}
	waitGroup.Wait()
	return len(dashboardUnits)
}

// CacheDashboardUnit Caches query for given dashboard unit for default date range presets.
func CacheDashboardUnit(dashboardUnit DashboardUnit, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	projectID := dashboardUnit.ProjectID
	savedQuery, errCode := GetQueryWithQueryId(projectID, dashboardUnit.QueryId)
	if errCode != http.StatusFound {
		errMsg := fmt.Sprintf("Failed to fetch query from query_id %d", dashboardUnit.QueryId)
		C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
		return
	}

	var queryClass string
	var query Query
	var queryGroup QueryGroup
	// try decoding for Query
	U.DecodePostgresJsonbToStructType(&savedQuery.Query, &query)
	if query.Class == "" {
		// if fails, try decoding for QueryGroup
		err1 := U.DecodePostgresJsonbToStructType(&savedQuery.Query, &queryGroup)
		if err1 != nil {
			errMsg := fmt.Sprintf("Failed to decode jsonb query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		queryClass = queryGroup.GetClass()
	} else {
		queryClass = query.Class
	}

	// excluding 'Web' class dashboard units
	if queryClass == QueryClassWeb {
		return
	}

	var unitWaitGroup sync.WaitGroup
	unitWaitGroup.Add(len(U.QueryDateRangePresets))
	for _, rangeFunction := range U.QueryDateRangePresets {
		from, to := rangeFunction()
		// Create a new baseQuery instance every time to avoid overwriting from, to values in routines.
		baseQuery, err := DecodeQueryForClass(dashboardUnit.Query, queryClass)
		if err != nil {
			errMsg := fmt.Sprintf("Error decoding query, query_id %d", dashboardUnit.QueryId)
			C.PingHealthcheckForFailure(C.HealthcheckDashboardCachingPingID, errMsg)
			return
		}
		baseQuery.SetQueryDateRange(from, to)
		cachePayload := DashboardUnitCachePayload{
			dashboardUnit: dashboardUnit,
			baseQuery:     baseQuery,
		}
		go cacheDashboardUnitForDateRange(cachePayload, &unitWaitGroup)
	}
	unitWaitGroup.Wait()
}

func cacheDashboardUnitForDateRange(cachePayload DashboardUnitCachePayload, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	dashboardUnit := cachePayload.dashboardUnit
	baseQuery := cachePayload.baseQuery
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
	if isDashboardUnitAlreadyCachedForRange(projectID, dashboardID, dashboardUnitID, from, to) {
		return
	}
	logCtx.Info("Starting to cache unit for date range")
	startTime := U.TimeNowUnix()

	var result interface{}
	var err error
	var errCode int
	var errMsg string
	if baseQuery.GetClass() == QueryClassFunnel || baseQuery.GetClass() == QueryClassInsights {
		analyticsQuery := baseQuery.(*Query)
		result, errCode, errMsg = Analyze(projectID, *analyticsQuery)
	} else if baseQuery.GetClass() == QueryClassAttribution {
		attributionQuery := baseQuery.(*AttributionQueryUnit)
		result, err = ExecuteAttributionQuery(projectID, attributionQuery.Query)
		if err != nil {
			errCode = http.StatusInternalServerError
		} else {
			errCode = http.StatusOK
		}
	} else if baseQuery.GetClass() == QueryClassChannel {
		channelQuery := baseQuery.(*ChannelQueryUnit)
		result, errCode = ExecuteChannelQuery(projectID, channelQuery.Query)
	} else if baseQuery.GetClass() == QueryClassEvents {
		groupQuery := baseQuery.(*QueryGroup)
		result, errCode = RunEventsGroupQuery(groupQuery.Queries, projectID)
	}
	if errCode != http.StatusOK {
		logCtx.Errorf("Error while running query %s", errMsg)
		return
	}

	timeTaken := U.TimeNowUnix() - startTime
	timeTakenString := U.SecondsToHMSString(timeTaken)
	logCtx.WithFields(log.Fields{"TimeTaken": timeTaken, "TimeTakenString": timeTakenString}).
		Info("Done caching unit for range")
	SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardID, dashboardUnitID, from, to)
}

func isDashboardUnitAlreadyCachedForRange(projectID, dashboardID, unitID uint64, from, to int64) bool {
	if from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) {
		// If from time is of today's beginning, refresh today's everytime a request is received.
		return false
	}
	cacheKey, err := getDashboardUnitQueryResultCacheKey(projectID, dashboardID, unitID, from, to)
	if err != nil {
		log.WithError(err).Errorf("Failed to get cache key")
		return false
	}
	exists, err := cacheRedis.ExistsPersistent(cacheKey)
	if err != nil {
		log.WithError(err).Errorf("Redis error on exists")
		return false
	}
	return exists
}
