package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesPAMForeignConstraints(pam model.ProjectAgentMapping) int {
	logFields := log.Fields{
		"pam": pam,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, agentErrCode := store.GetAgentByUUID(pam.AgentUUID)
	_, projectErrCode := store.GetProject(pam.ProjectID)
	if agentErrCode != http.StatusFound || projectErrCode != http.StatusFound {
		return http.StatusBadRequest
	}

	if pam.InvitedBy != nil && *pam.InvitedBy != "" {
		_, invitedByErrCode := store.GetAgentByUUID(*pam.InvitedBy)
		if invitedByErrCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

// Add Check
// Project should not have more than 100 Agents
func createProjectAgentMapping(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"pam": pam,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if pam == nil || pam.AgentUUID == "" || pam.ProjectID == 0 {
		return nil, http.StatusBadRequest
	}

	if pam.Role == 0 {
		pam.Role = model.AGENT
	}

	db := C.GetServices().Db

	if err := db.Create(pam).Error; err != nil {
		if IsDuplicateRecordError(err) {
			return nil, http.StatusFound
		}
		log.WithError(err).Error("CreateProjectAgentMapping Failed.")
		return nil, http.StatusInternalServerError
	}

	return pam, http.StatusCreated
}

func (store *MemSQL) CreateProjectAgentMappingWithDependencies(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"pam": pam,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if errCode := store.satisfiesPAMForeignConstraints(*pam); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	cPam, errCode := createProjectAgentMapping(pam)
	if errCode != http.StatusCreated {
		return cPam, errCode
	}

	// dependencies.
	_, errCode = store.CreateAgentPersonalDashboardForProject(pam.ProjectID, pam.AgentUUID)
	if errCode != http.StatusCreated {
		// Should not fail agent creation if failed. log and continue.
		// User will be able to create a dashboard himself.
		log.WithFields(log.Fields{"project_id": pam.ProjectID,
			"agent_uuid": pam.AgentUUID}).Error("Failed to create agent's personal dashboard for project.")
	}

	return cPam, http.StatusCreated
}
func (store *MemSQL) CreateProjectAgentMappingWithDependenciesWithoutDashboard(pam *model.ProjectAgentMapping) (*model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"pam": pam,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if errCode := store.satisfiesPAMForeignConstraints(*pam); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	cPam, errCode := createProjectAgentMapping(pam)
	if errCode != http.StatusCreated {
		return cPam, errCode
	}
	return cPam, http.StatusCreated
}
func (store *MemSQL) GetProjectAgentMapping(projectId uint64, agentUUID string) (*model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func (store *MemSQL) GetProjectAgentMappingsByProjectId(projectId uint64) ([]model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetProjectAgentMappingsByProjectIds(projectIds []uint64) ([]model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"project_ids": projectIds,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetProjectAgentMappingsByAgentUUID(agentUUID string) ([]model.ProjectAgentMapping, int) {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) DoesAgentHaveProject(agentUUID string) int {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) DeleteProjectAgentMapping(projectId uint64, agentUUIDToRemove string) int {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid_to_remove": agentUUIDToRemove,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) EditProjectAgentMapping(projectId uint64, agentUUIDToEdit string, role int64) int {
	logFields := log.Fields{
		"project_id": projectId,
		"agent_uuid_to_edit": agentUUIDToEdit,
		"role": role,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
