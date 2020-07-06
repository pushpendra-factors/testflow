package model

import (
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type DashboardUnit struct {
	// Composite primary key, id + project_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboard_units(project_id) ref projects(id).
	ProjectId    uint64         `gorm:"primary_key:true" json:"project_id"`
	DashboardId  uint64         `gorm:"primary_key:true" json:"dashboard_id"`
	Title        string         `gorm:"not null" json:"title"`
	Query        postgres.Jsonb `gorm:"not null" json:"query"`
	Presentation string         `gorm:"type:varchar(5);not null" json:"presentation"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type DashboardCacheResult struct {
	Result      interface{} `json:"result"`
	From        int64       `json:"from"`
	To          int64       `json:"tom"`
	RefreshedAt int64       `json:"refreshed_at"`
}

const (
	PresentationLine   = "pl"
	PresentationBar    = "pb"
	PresentationTable  = "pt"
	PresentationCard   = "pc"
	PresentationFunnel = "pf"
)

var presentations = [...]string{PresentationLine, PresentationBar,
	PresentationTable, PresentationCard, PresentationFunnel}

const (
	UnitCard  = "card"
	UnitChart = "chart"
)

var UnitTypes = [...]string{UnitCard, UnitChart}

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

func CreateDashboardUnit(projectId uint64, agentUUID string, dashboardUnit *DashboardUnit) (*DashboardUnit, int, string) {
	db := C.GetServices().Db

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest, "Invalid request"
	}

	valid, errMsg := isValidDashboardUnit(dashboardUnit)
	if !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	hasAccess, dashboard := HasAccessToDashboard(projectId, agentUUID, dashboardUnit.DashboardId)
	if !hasAccess {
		return nil, http.StatusForbidden, "Unauthorized to access dashboard"
	}

	dashboardUnit.ProjectId = projectId
	if err := db.Create(dashboardUnit).Error; err != nil {
		errMsg := "Falied to create dashboard unit."
		log.WithFields(log.Fields{"dashboard_unit": dashboardUnit,
			"project_id": projectId}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}

	errCode := addUnitPositionOnDashboard(projectId, agentUUID, dashboardUnit.DashboardId,
		dashboardUnit.ID, GetUnitType(dashboardUnit.Presentation), dashboard.UnitsPosition)
	if errCode != http.StatusAccepted {
		errMsg := "Failed add position for new dashboard unit."
		log.WithFields(log.Fields{"project_id": projectId,
			"dashboardUnitId": dashboardUnit.ID}).Error(errMsg)
		return nil, http.StatusInternalServerError, ""
	}

	return dashboardUnit, http.StatusCreated, ""
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
	return dashboardUnits, http.StatusFound
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

	return dashboardUnits, http.StatusFound
}

func DeleteDashboardUnit(projectId uint64, agentUUID string, dashboardId uint64, id uint64) int {
	db := C.GetServices().Db

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to delete dashboard unit. Invalid scope ids.")
		return http.StatusBadRequest
	}

	hasAccess, dashboard := HasAccessToDashboard(projectId, agentUUID, dashboardId)
	if !hasAccess {
		return http.StatusForbidden
	}

	err := db.Where("id = ? AND project_id = ? AND dashboard_id = ?",
		id, projectId, dashboardId).Delete(&DashboardUnit{}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "dashboard_id": dashboardId,
			"unit_id": id}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}

	errCode := removeUnitPositionOnDashboard(projectId, agentUUID, dashboardId, id, dashboard.UnitsPosition)
	if errCode != http.StatusAccepted {
		errMsg := "Failed remove position for unit on dashboard."
		log.WithFields(log.Fields{"project_id": projectId, "unitId": id}).Error(errMsg)
		// log error and continue to delete dashboard unit.
		// To avoid improper experience.
		return http.StatusAccepted
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
		startTime := U.TimeNowUnix()
		unitsCount := CacheDashboardUnitsForProjectID(projectID, numRoutines)

		timeTaken := U.TimeNowUnix() - startTime
		timeTakenString := U.SecondsToHMSString(timeTaken)
		logCtx.WithFields(log.Fields{"TimeTaken": timeTaken}).Infof("Time taken for caching %d dashboard units %s", unitsCount, timeTakenString)
	}
	return
}

// CacheDashboardUnitsForProjectID Caches all the dashboard units for the given `projectID`.
func CacheDashboardUnitsForProjectID(projectID uint64, numRoutines int) int {
	logCtx := log.WithFields(log.Fields{
		"Method":    "CacheDashboardUnitsForProjectID",
		"ProjectID": projectID,
	})
	if numRoutines == 0 {
		numRoutines = 1
	}

	dashboardUnits, errCode := GetDashboardUnitsForProjectID(projectID)
	if errCode != http.StatusFound || len(dashboardUnits) == 0 {
		return 0
	}

	var waitGroup sync.WaitGroup
	count := 0
	for _, dashboardUnit := range dashboardUnits {
		logCtx = logCtx.WithFields(log.Fields{
			"UnitID":      dashboardUnit.ID,
			"DashboardID": dashboardUnit.DashboardId,
		})

		waitGroup.Add(1)
		count++
		go CacheDashboardUnit(projectID, dashboardUnit, &waitGroup)
		if count%numRoutines == 0 {
			waitGroup.Wait()
		}
	}
	waitGroup.Wait()
	return len(dashboardUnits)
}

// CacheDashboardUnit Caches query for given dashboard unit for default date range presets.
func CacheDashboardUnit(projectID uint64, dashboardUnit DashboardUnit, waitGroup *sync.WaitGroup) {
	logCtx := log.WithFields(log.Fields{
		"Method":      "CacheDashboardUnit",
		"ProjectID":   projectID,
		"DashboardID": dashboardUnit.DashboardId,
		"UnitID":      dashboardUnit.ID,
	})
	defer waitGroup.Done()

	var query Query
	if err := U.DecodePostgresJsonbToStructType(&dashboardUnit.Query, &query); err != nil {
		logCtx.WithError(err).Errorf("Failed to decode jsonb query")
		return
	}

	for preset, rangeFunction := range U.QueryDateRangePresets {
		query.From, query.To = rangeFunction()
		logCtx = logCtx.WithFields(log.Fields{"Preset": preset, "From": query.From, "To": query.To})

		if isDashboardUnitAlreadyCachedForRange(projectID, dashboardUnit.DashboardId, dashboardUnit.ID, query.From, query.To) {
			continue
		}

		result, errCode, errMsg := Analyze(projectID, query)
		if errCode != http.StatusOK {
			logCtx.Errorf("Error while running query %s", errMsg)
			return
		}
		SetCacheResultByDashboardIdAndUnitId(result, projectID, dashboardUnit.DashboardId, dashboardUnit.ID, query.To, query.From)
	}
	return
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
	exists, err := cacheRedis.Exists(cacheKey)
	if err != nil {
		log.WithError(err).Errorf("Redis error on exists")
		return false
	}
	return exists
}
