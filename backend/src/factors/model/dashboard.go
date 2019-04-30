package model

import (
	C "factors/config"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Dashboard struct {
	// Composite primary key, id + project_id.
	ID        string `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	Name      string `gorm:"not null" json:"name"`
	Type      string `gorm:"not null" json:"type"`
	// User should be able to mark it as primary dashboard.
	// Primary   bool   `json:"primary"`
	Deleted   bool      `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	TypePersonal = "tp"
	TypeSharable = "ts"
)

const defaultNamePersonalDashboard = "My Dashboard"

func createDashboard(projectId uint64, dashboard *Dashboard) (*Dashboard, int) {
	db := C.GetServices().Db

	if projectId == 0 || dashboard.Name == "" {
		return nil, http.StatusBadRequest
	}

	dashboard.ProjectId = projectId
	if err := db.Create(dashboard).Error; err != nil {
		log.WithFields(log.Fields{"dashboard": dashboard,
			"project_id": projectId}).WithError(err).Error("Failed to create dashboard.")
		return nil, http.StatusInternalServerError
	}

	return dashboard, http.StatusCreated
}

func CreatePersonalDashboard(projectId uint64) (*Dashboard, int) {
	return createDashboard(projectId, &Dashboard{Name: defaultNamePersonalDashboard, Type: TypePersonal})
}

func CreateSharableDashboard(projectId uint64, dashboard *Dashboard) (*Dashboard, int) {
	dashboard.Type = TypeSharable
	return createDashboard(projectId, dashboard)
}

// Todo: Manage ACLs for dashboards and return dashboards
// to which the requesting agent has permissions to.
func GetDashboards(projectId uint64) ([]Dashboard, int) {
	db := C.GetServices().Db

	var dashboards []Dashboard
	if projectId == 0 {
		log.Error("Failed to get dashboards. Invalid project_id.")
		return dashboards, http.StatusInternalServerError
	}

	err := db.Order("created_at DESC").Where("project_id = ?", projectId).Find(&dashboards).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboards.")
		return dashboards, http.StatusInternalServerError
	}

	return dashboards, http.StatusFound
}
