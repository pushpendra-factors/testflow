package postgres

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"strings"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// TODO: Make index name a constant and read it
const error_Duplicate_email_error = "pq: duplicate key value violates unique constraint \"agent_email_unique_idx\""

func createAgent(agent *model.Agent) (*model.Agent, int) {
	if agent.Email == "" {
		log.Error("CreateAgent Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	agent.Email = strings.ToLower(agent.Email)

	// Adding random string as salt before create.
	if agent.Salt == "" {
		agent.Salt = U.RandomString(model.AgentSaltLength)
	}

	db := C.GetServices().Db
	if err := db.Create(agent).Error; err != nil {
		if err.Error() == error_Duplicate_email_error {
			log.WithError(err).Error("CreateAgent Failed, duplicate email")
			return nil, http.StatusBadRequest
		}
		log.WithError(err).Error("CreateAgent Failed")
		return nil, http.StatusInternalServerError
	}

	return agent, http.StatusCreated
}

func (pg *Postgres) CreateAgentWithDependencies(params *model.CreateAgentParams) (*model.CreateAgentResponse, int) {
	if params == nil || params.PlanCode == "" || params.Agent == nil || params.Agent.Email == "" {
		return nil, http.StatusBadRequest
	}

	resp := &model.CreateAgentResponse{}

	agent, errCode := createAgent(params.Agent)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	resp.Agent = agent

	billingAccount, errCode := pg.createBillingAccount(params.PlanCode, agent.UUID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	resp.BillingAccount = billingAccount

	return resp, http.StatusCreated
}

func (pg *Postgres) GetAgentByEmail(email string) (*model.Agent, int) {

	if email == "" {
		log.Error("GetAgentByEmail Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	email = strings.ToLower(email)

	db := C.GetServices().Db

	var agent model.Agent
	if err := db.Limit(1).Where("email = ?", email).Find(&agent).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &agent, http.StatusFound
}

func (pg *Postgres) GetAgentByUUID(uuid string) (*model.Agent, int) {
	if uuid == "" {
		log.Error("GetAgentByUUID Failed. UUID not provided.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var agent model.Agent

	if err := db.Limit(1).Where("uuid = ?", uuid).Find(&agent).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("GetAgentByUUID Failed.")
		return nil, http.StatusInternalServerError
	}

	return &agent, http.StatusFound
}

func (pg *Postgres) GetAgentsByUUIDs(uuids []string) ([]*model.Agent, int) {
	if len(uuids) == 0 {
		log.Error("No uuids for agents")
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db

	agents := make([]*model.Agent, 0, 0)

	if err := db.Limit(len(uuids)).Where("uuid IN (?)", uuids).Find(&agents).Error; err != nil {
		log.Error("could not get agents for given agentUUIDs", err)
		return nil, http.StatusInternalServerError
	}

	if len(agents) == 0 {
		log.Error("No agents are found for given agentUUID")
		return nil, http.StatusNotFound
	}

	return agents, http.StatusFound
}

func (pg *Postgres) GetAgentInfo(uuid string) (*model.AgentInfo, int) {
	agent, errCode := pg.GetAgentByUUID(uuid)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	agentInfo := model.CreateAgentInfo(agent)
	return agentInfo, errCode
}

func (pg *Postgres) UpdateAgentIntAdwordsRefreshToken(uuid, refreshToken string) int {
	if uuid == "" || refreshToken == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentAdwordsRefreshToken failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntAdwordsRefreshToken(refreshToken))
}
func (pg *Postgres) UpdateAgentIntGoogleOrganicRefreshToken(uuid, refreshToken string) int {
	if uuid == "" || refreshToken == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentGSCRefreshToken failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntGSCRefreshToken(refreshToken))
}

func (pg *Postgres) UpdateAgentIntSalesforce(uuid, refreshToken string, instanceUrl string) int {
	if uuid == "" || refreshToken == "" || instanceUrl == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentIntSalesforce failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntSalesforceRefreshToken(refreshToken), model.IntSalesforceInstanceURL(instanceUrl))
}

func (pg *Postgres) UpdateAgentSalesforceInstanceURL(uuid, instanceUrl string) int {
	if uuid == "" || instanceUrl == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentInstanceURL failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntSalesforceInstanceURL(instanceUrl))
}

func (pg *Postgres) UpdateAgentPassword(uuid, plainTextPassword string, passUpdatedAt time.Time) int {

	if uuid == "" || plainTextPassword == "" {
		log.Error("UpdateAgentPassword Failed. Missing params")
		return http.StatusBadRequest
	}

	hashedPassword, err := model.HashPassword(plainTextPassword)
	if err != nil {
		return http.StatusInternalServerError
	}

	return updateAgent(uuid, model.PasswordAndPasswordCreatedAt(hashedPassword, passUpdatedAt),
		model.Salt(U.RandomString(model.AgentSaltLength)))
}

func (pg *Postgres) UpdateAgentLastLoginInfo(agentUUID string, ts time.Time) int {
	if agentUUID == "" {
		log.Error("UpdateAgentLastLoginInfo Failed. Missing params")
		return http.StatusBadRequest
	}

	return updateAgent(agentUUID, model.LastLoggedInAtAndIncrLoginCount(ts))
}

func (pg *Postgres) UpdateAgentVerificationDetails(agentUUID, password, firstName,
	lastName string, verified bool, passUpdatedAt time.Time) int {

	if agentUUID == "" {
		log.Error("UpdateAgentVerificationDetails Failed. Missing params")
		return http.StatusBadRequest
	}

	hashedPassword, err := model.HashPassword(password)
	if err != nil {
		return http.StatusInternalServerError
	}

	options := make([]model.Option, 0)
	if firstName != "" {
		options = append(options, model.Firstname(firstName))
	}
	if lastName != "" {
		options = append(options, model.Lastname(lastName))
	}
	options = append(options, model.IsEmailVerified(verified))
	options = append(options, model.PasswordAndPasswordCreatedAt(hashedPassword, passUpdatedAt))
	return updateAgent(agentUUID, options...)
}

func (pg *Postgres) UpdateAgentVerificationDetailsFromAuth0(agentUUID, firstName, lastName string, verified bool, value *postgres.Jsonb) int {

	if agentUUID == "" {
		log.Error("UpdateAgentVerificationDetails Failed. Missing params")
		return http.StatusBadRequest
	}

	options := make([]model.Option, 0)
	if firstName != "" {
		options = append(options, model.Firstname(firstName))
	}
	if lastName != "" {
		options = append(options, model.Lastname(lastName))
	}
	options = append(options, model.IsEmailVerified(verified))
	options = append(options, model.IsAuth0User(true))
	options = append(options, model.Auth0Value(value))
	return updateAgent(agentUUID, options...)
}

func (pg *Postgres) UpdateAgentInformation(agentUUID, firstName, lastName, phone string, isOnboardingFlowSeen *bool) int {
	if agentUUID == "" {
		return http.StatusBadRequest
	}
	updateParams := []model.Option{}
	if firstName != "" {
		updateParams = append(updateParams, model.Firstname(firstName))
	}
	if lastName != "" {
		updateParams = append(updateParams, model.Lastname(lastName))
	}
	if phone != "" {
		updateParams = append(updateParams, model.Phone(phone))
	}
	if isOnboardingFlowSeen != nil {
		updateParams = append(updateParams, model.IsOnboardingFlowSeen(*isOnboardingFlowSeen))
	}
	return updateAgent(agentUUID, updateParams...)
}

func updateAgent(agentUUID string, options ...model.Option) int {
	if agentUUID == "" {
		return http.StatusBadRequest
	}

	fields := model.FieldsToUpdate{}

	for _, option := range options {
		option(fields)
	}

	if len(fields) == 0 {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	db = db.Model(&model.Agent{}).Where("uuid = ?", agentUUID).Updates(fields)

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateAgent Failed")
		return http.StatusInternalServerError
	}
	if db.RowsAffected == 0 {
		return http.StatusNoContent
	}
	return http.StatusAccepted
}

func (pg *Postgres) GetPrimaryAgentOfProject(projectId uint64) (uuid string, errCode int) {
	db := C.GetServices().Db

	var projectAgentMappings []model.ProjectAgentMapping
	err := db.Limit(1).Order("created_at ASC").
		Where("project_id = ?", projectId).Find(&projectAgentMappings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get primary agent of project.")
		return "", http.StatusInternalServerError
	}

	if len(projectAgentMappings) == 0 {
		return "", http.StatusNotFound
	}

	return projectAgentMappings[0].AgentUUID, http.StatusFound
}
