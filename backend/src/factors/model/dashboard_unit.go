package model

import (
	C "factors/config"
	"net/http"
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
		return nil, http.StatusUnauthorized, "Unauthorized to access dashboard"
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

func GetDashboardUnits(projectId uint64, agentUUID string, dashboardId uint64) ([]DashboardUnit, int) {
	db := C.GetServices().Db

	var dashboardUnits []DashboardUnit
	if projectId == 0 || dashboardId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id or agent_id")
		return dashboardUnits, http.StatusBadRequest
	}

	if hasAccess, _ := HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
		return nil, http.StatusUnauthorized
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ?",
		projectId, dashboardId).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	return dashboardUnits, http.StatusFound
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
		return http.StatusUnauthorized
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
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func UpdateDashboardUnit(projectId uint64, agentUUID string,
	dashboardId uint64, id uint64, unit *DashboardUnit) (*DashboardUnit, int) {

	if projectId == 0 || agentUUID == "" ||
		dashboardId == 0 || id == 0 {

		log.Error("Failed to update dashboard unit. Invalid scope ids.")
		return nil, http.StatusBadRequest
	}

	if hasAccess, _ := HasAccessToDashboard(projectId, agentUUID, dashboardId); !hasAccess {
		return nil, http.StatusUnauthorized
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
		return nil, http.StatusInternalServerError
	}

	// returns only updated fields, avoid using it on DashboardUnit API.
	return &updatedDashboardUnitFields, http.StatusAccepted
}
