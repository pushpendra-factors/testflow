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
	// Foreign keys projects(id) and dashboards(id).
	ProjectId    uint64         `gorm:"primary_key:true" json:"project_id"`
	DashboardId  uint64         `gorm:"primary_key:true" json:"dashboard_id"`
	Title        string         `gorm:"not null" json:"title"`
	Query        postgres.Jsonb `gorm:"not null" json:"query"`
	Presentation string         `gorm:"not null" json:"presentation"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

const (
	presentationLine  = "pl"
	presentationBar   = "pb"
	presentationTable = "pt"
	presentationCard  = "pc"
)

var presentations = [...]string{presentationLine, presentationBar, presentationTable, presentationCard}

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

func CreateDashboardUnit(projectId uint64, dashboardUnit *DashboardUnit) (*DashboardUnit, int, string) {
	db := C.GetServices().Db

	if projectId == 0 {
		return nil, http.StatusBadRequest, "Invalid project id"
	}

	valid, errMsg := isValidDashboardUnit(dashboardUnit)
	if !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	dashboardUnit.ProjectId = projectId
	if err := db.Create(dashboardUnit).Error; err != nil {
		errMsg := "Falied to create dashboard unit."
		log.WithFields(log.Fields{"dashboard_unit": dashboardUnit,
			"project_id": projectId}).WithError(err).Error(errMsg)
		return nil, http.StatusInternalServerError, errMsg
	}

	return dashboardUnit, http.StatusCreated, ""
}

// Todo: Manage ACLs for dashboards and return dashboards_units
// to which the requesting agent has permissions by dashboard ACL.
func GetDashboardUnits(projectId uint64, dashboardId uint64) ([]DashboardUnit, int) {
	db := C.GetServices().Db

	var dashboardUnits []DashboardUnit
	if projectId == 0 || dashboardId == 0 {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id.")
		return dashboardUnits, http.StatusBadRequest
	}

	err := db.Order("created_at DESC").Where("project_id = ? AND dashboard_id = ?",
		projectId, dashboardId).Find(&dashboardUnits).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboard units.")
		return dashboardUnits, http.StatusInternalServerError
	}

	return dashboardUnits, http.StatusFound
}

func DeleteDashboardUnit(projectId uint64, dashbordId uint64, id uint64) int {
	db := C.GetServices().Db

	if projectId == 0 || dashbordId == 0 || id == 0 {
		log.Error("Failed to get dashboard units. Invalid project_id or dashboard_id or unit_id")
		return http.StatusBadRequest
	}

	err := db.Where("id = ? AND project_id = ? AND dashboard_id = ?", id, projectId, dashbordId).Delete(&DashboardUnit{}).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "dashboard_id": dashbordId,
			"unit_id": id}).WithError(err).Error("Failed to delete dashboard unit.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
