package model

import (
	C "factors/config"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

const (
	AGENT                  = 1
	ADMIN                  = 2
	MAX_AGENTS_PER_PROJECT = 500
)

type ProjectAgentMapping struct {

	// Composite primary key with project_id and agent_uuid
	AgentUUID string `gorm:"primary_key:true;type:varchar(255)" json:"agent_uuid"`
	ProjectID uint64 `gorm:"primary_key:true" json:"project_id"`

	// Foreign key constraints added in creation script
	// project_id -> projects(id)
	// agent_uuid -> agents(uuid)
	// invited_by -> agents(uuid)

	Role uint64 `json:"role"`

	// Created as pointer to allow storing NULL in db
	InvitedBy *string `gorm:"type:varchar(255)" json:"invited_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Add Check
// Project should not have more than 100 Agents
func createProjectAgentMapping(pam *ProjectAgentMapping) (*ProjectAgentMapping, int) {
	if pam == nil {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	if err := db.Create(pam).Error; err != nil {
		log.WithError(err).Error("CreateProjectAgentMapping Failed.")
		// TODO(Ankit): check if is duplicate error
		return nil, http.StatusInternalServerError
	}

	return pam, http.StatusCreated
}

func CreateProjectAgentMappingWithDependencies(pam *ProjectAgentMapping) (*ProjectAgentMapping, int) {
	cPam, errCode := createProjectAgentMapping(pam)
	if errCode != http.StatusCreated {
		return cPam, errCode
	}

	// dependencies.
	_, errCode = CreateAgentPersonalDashboardForProject(pam.ProjectID, pam.AgentUUID)
	if errCode != http.StatusCreated {
		// Should not fail agent creation if failed. log and continue.
		// User will be able to create a dashboard himself.
		log.WithFields(log.Fields{"project_id": pam.ProjectID,
			"agent_uuid": pam.AgentUUID}).Error("Failed to create agent's personal dashboard for project.")
	}

	return cPam, http.StatusCreated
}

func GetProjectAgentMapping(projectId uint64, agentUUID string) (*ProjectAgentMapping, int) {

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	pam := &ProjectAgentMapping{}
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ?", projectId).Where("agent_uuid = ?", agentUUID).Find(pam).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return pam, http.StatusFound
}

func GetProjectAgentMappingsByProjectId(projectId uint64) ([]ProjectAgentMapping, int) {
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var pam []ProjectAgentMapping
	if err := db.Limit(MAX_AGENTS_PER_PROJECT).Where("project_id = ?", projectId).Find(&pam).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(pam) == 0 {
		return nil, http.StatusNotFound
	}

	return pam, http.StatusFound
}

func GetProjectAgentMappingsByAgentUUID(agentUUID string) ([]ProjectAgentMapping, int) {
	if agentUUID == "" {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var pam []ProjectAgentMapping
	if err := db.Where("agent_uuid = ?", agentUUID).Find(&pam).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(pam) == 0 {
		return nil, http.StatusNotFound
	}

	return pam, http.StatusFound
}

func DoesAgentHaveProject(agentUUID string) int {
	if agentUUID == "" {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	var pam ProjectAgentMapping
	if err := db.Where("agent_uuid = ?", agentUUID).First(&pam).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		log.WithField("agent_uuid",
			agentUUID).WithError(err).Error("Failed to check does agent have project.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}
