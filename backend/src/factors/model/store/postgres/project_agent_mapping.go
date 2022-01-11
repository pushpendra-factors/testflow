package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

const error_Duplicate_project_agent_mapping_error = "pq: duplicate key value violates unique constraint \"project_agent_mappings_pkey\""

// Add Check
// Project should not have more than 100 Agents
func createProjectAgentMapping(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int) {
	if pam == nil || pam.AgentUUID == "" || pam.ProjectID == 0 {
		return nil, http.StatusBadRequest
	}

	if pam.Role == 0 {
		pam.Role = model.AGENT
	}

	db := C.GetServices().Db

	if err := db.Create(pam).Error; err != nil {
		if err.Error() == error_Duplicate_project_agent_mapping_error {
			return nil, http.StatusFound
		}
		log.WithError(err).Error("CreateProjectAgentMapping Failed.")
		return nil, http.StatusInternalServerError
	}

	return pam, http.StatusCreated
}

func (pg *Postgres) CreateProjectAgentMappingWithDependencies(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int) {
	cPam, errCode := createProjectAgentMapping(pam)
	if errCode != http.StatusCreated {
		return cPam, errCode
	}

	// dependencies.
	_, errCode = pg.CreateAgentPersonalDashboardForProject(pam.ProjectID, pam.AgentUUID)
	if errCode != http.StatusCreated {
		// Should not fail agent creation if failed. log and continue.
		// User will be able to create a dashboard himself.
		log.WithFields(log.Fields{"project_id": pam.ProjectID,
			"agent_uuid": pam.AgentUUID}).Error("Failed to create agent's personal dashboard for project.")
	}

	return cPam, http.StatusCreated
}
func (pg *Postgres) CreateProjectAgentMappingWithDependenciesWithoutDashboard(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int) {
	cPam, errCode := createProjectAgentMapping(pam)
	if errCode != http.StatusCreated {
		return cPam, errCode
	}
	return cPam, http.StatusCreated
}
func (pg *Postgres) GetProjectAgentMapping(projectId uint64, agentUUID string) (*model.ProjectAgentMapping, int) {

	if projectId == 0 || agentUUID == "" {
		return nil, http.StatusBadRequest
	}

	pam := &model.ProjectAgentMapping{}
	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ?", projectId).Where("agent_uuid = ?", agentUUID).Find(pam).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return pam, http.StatusFound
}

func (pg *Postgres) GetProjectAgentMappingsByProjectId(projectId uint64) ([]model.ProjectAgentMapping, int) {
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var pam []model.ProjectAgentMapping
	if err := db.Limit(model.MAX_AGENTS_PER_PROJECT).Where("project_id = ?", projectId).Find(&pam).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(pam) == 0 {
		return nil, http.StatusNotFound
	}

	return pam, http.StatusFound
}

func (pg *Postgres) GetProjectAgentMappingsByProjectIds(projectIds []uint64) ([]model.ProjectAgentMapping, int) {
	if len(projectIds) == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var pam []model.ProjectAgentMapping
	if err := db.Where("project_id IN (?)", projectIds).Find(&pam).Error; err != nil {
		log.WithError(err).Error("Finding model.ProjectAgentMapping failed using ProjectIds")
		return nil, http.StatusInternalServerError
	}

	if len(pam) == 0 {
		return nil, http.StatusNotFound
	}

	return pam, http.StatusFound
}

func (pg *Postgres) GetProjectAgentMappingsByAgentUUID(agentUUID string) ([]model.ProjectAgentMapping, int) {
	if agentUUID == "" {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db
	var pam []model.ProjectAgentMapping
	if err := db.Where("agent_uuid = ?", agentUUID).Find(&pam).Error; err != nil {
		log.WithError(err).Error("Finding model.ProjectAgentMapping failed using AgentUUID")
		return nil, http.StatusInternalServerError
	}

	if len(pam) == 0 {
		return nil, http.StatusNotFound
	}

	return pam, http.StatusFound
}

func (pg *Postgres) DoesAgentHaveProject(agentUUID string) int {
	if agentUUID == "" {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	var pam model.ProjectAgentMapping
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

func (pg *Postgres) DeleteProjectAgentMapping(projectId uint64, agentUUIDToRemove string) int {
	if projectId == 0 || agentUUIDToRemove == "" {
		return http.StatusBadRequest
	}
	db := C.GetServices().Db

	db = db.Where("project_id = ?", projectId).Where("agent_uuid = ? ", agentUUIDToRemove).Delete(&model.ProjectAgentMapping{})

	if db.Error != nil {
		log.WithFields(log.Fields{"projectId": projectId, "agentUUID": agentUUIDToRemove}).WithError(db.Error).Error(
			"Deleting model.ProjectAgentMapping failed.")
		return http.StatusInternalServerError
	}

	if db.RowsAffected == 0 {
		return http.StatusNotFound
	}

	return http.StatusAccepted
}

func (pg *Postgres) EditProjectAgentMapping(projectId uint64, agentUUIDToEdit string, role int64) int {
	if projectId == 0 || agentUUIDToEdit == "" || role == 0 {
		return http.StatusBadRequest
	}
	db := C.GetServices().Db

	updateFields := make(map[string]interface{}, 0)
	updateFields["role"] = role

	err := db.Model(&model.ProjectAgentMapping{}).Where("project_id = ? AND agent_uuid = ?", projectId, agentUUIDToEdit).Update(updateFields).Error

	if err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "agentUUID": agentUUIDToEdit}).WithError(err).Error(
			"Deleting model.ProjectAgentMapping failed.")
		return http.StatusInternalServerError
	}
	return 0
}
