package model

import (
	C "factors/config"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Dashboard struct {
	// Composite primary key, id + project_id + agent_id.
	ID uint64 `gorm:"primary_key:true" json:"id"`
	// Foreign key dashboards(project_id) ref projects(id).
	ProjectId uint64 `gorm:"primary_key:true" json:"project_id"`
	AgentUUID string `gorm:"primary_key:true" json:"-"`
	Name      string `gorm:"not null" json:"name"`
	Type      string `gorm:"type:varchar(5);not null" json:"type"`
	// User should be able to mark it as primary dashboard.
	// Primary   bool   `json:"primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	DashboardTypePrivate        = "pr"
	DashboardTypeProjectVisible = "pv"
)

var types = []string{DashboardTypePrivate, DashboardTypeProjectVisible}

const AgentProjectPersonalDashboardName = "My Dashboard"

func isValidDashboard(dashboard *Dashboard) bool {
	if dashboard.Name == "" {
		return false
	}

	validType := false
	for _, t := range types {
		if t == dashboard.Type {
			validType = true
			break
		}
	}

	return validType
}

func CreateDashboard(projectId uint64, agentUUID string, dashboard *Dashboard) (*Dashboard, int) {
	db := C.GetServices().Db

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	if !isValidDashboard(dashboard) {
		return nil, http.StatusBadRequest
	}

	dashboard.ProjectId = projectId
	dashboard.AgentUUID = agentUUID

	if err := db.Create(dashboard).Error; err != nil {
		log.WithFields(log.Fields{"dashboard": dashboard,
			"project_id": projectId}).WithError(err).Error("Failed to create dashboard.")
		return nil, http.StatusInternalServerError
	}

	return dashboard, http.StatusCreated
}

func CreateAgentPersonalDashboardForProject(projectId uint64, agentUUID string) (*Dashboard, int) {
	return CreateDashboard(projectId, agentUUID,
		&Dashboard{Name: AgentProjectPersonalDashboardName, Type: DashboardTypePrivate})
}

func GetDashboards(projectId uint64, agentUUID string) ([]Dashboard, int) {
	db := C.GetServices().Db

	var dashboards []Dashboard
	if projectId == 0 || agentUUID == "" {
		log.Error("Failed to get dashboards. Invalid project_id.")
		return dashboards, http.StatusBadRequest
	}

	err := db.Order("created_at ASC").Where("project_id = ? AND (type = ? OR agent_uuid = ?)",
		projectId, DashboardTypeProjectVisible, agentUUID).Find(&dashboards).Error
	if err != nil {
		log.WithField("project_id", projectId).WithError(err).Error("Failed to get dashboards.")
		return dashboards, http.StatusInternalServerError
	}

	return dashboards, http.StatusFound
}

// HasAccessToDashboard validates access to dashboard by project_id,
// agent_id considering type.
func HasAccessToDashboard(projectId uint64, agentUUID string, id uint64) bool {
	dashboards, errCode := GetDashboards(projectId, agentUUID)
	if errCode != http.StatusFound {
		return false
	}

	hasAccess := false
	for _, dashboard := range dashboards {
		if dashboard.ID == id {
			hasAccess = true
			break
		}
	}

	return hasAccess
}
