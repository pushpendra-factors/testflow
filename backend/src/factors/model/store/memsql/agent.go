package memsql

import (
	C "factors/config"
	"factors/model/model"
	"fmt"
	"net/http"
	"strings"
	"time"

	U "factors/util"

	billing "factors/billing/chargebee"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) satisfiesAgentForeignConstraints(agent model.Agent) int {
	logFields := log.Fields{
		"agent": agent,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if agent.InvitedBy != nil && *agent.InvitedBy != "" {
		_, errCode := store.GetAgentByUUID(*agent.InvitedBy)
		if errCode != http.StatusFound {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

func (store *MemSQL) createAgent(agent *model.Agent) (*model.Agent, int) {
	logFields := log.Fields{
		"agent": agent,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if agent.Email == "" {
		log.Error("CreateAgent Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	agent.Email = strings.ToLower(agent.Email)
	// Unique (email) constraint.
	if _, errCode := store.GetAgentByEmail(agent.Email); errCode == http.StatusFound {
		return nil, http.StatusBadRequest
	}

	// Adding random string as salt before create.
	if agent.Salt == "" {
		agent.Salt = U.RandomString(model.AgentSaltLength)
	}

	if agent.UUID == "" {
		agent.UUID = U.GetUUID()
	}

	if errCode := store.satisfiesAgentForeignConstraints(*agent); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	db := C.GetServices().Db
	if err := db.Create(agent).Error; err != nil {
		log.WithError(err).Error("CreateAgent Failed")
		return nil, http.StatusInternalServerError
	}

	return agent, http.StatusCreated
}

func (store *MemSQL) CreateAgentWithDependencies(params *model.CreateAgentParams) (*model.CreateAgentResponse, int) {
	logFields := log.Fields{
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if params == nil || params.PlanCode == "" || params.Agent == nil || params.Agent.Email == "" {
		return nil, http.StatusBadRequest
	}

	resp := &model.CreateAgentResponse{}

	invitedBy := params.Agent.InvitedBy

	if strings.HasSuffix(params.Agent.Email, "factors.ai") && *invitedBy == "" {

		customer, status, err := billing.CreateChargebeeCustomer(*params.Agent)
		if err != nil || status != http.StatusCreated {
			return nil, http.StatusInternalServerError

		}
		params.Agent.BillingCustomerID = customer.Id
	}

	agent, errCode := store.createAgent(params.Agent)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	resp.Agent = agent

	billingAccount, errCode := store.createBillingAccount(params.PlanCode, agent.UUID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	resp.BillingAccount = billingAccount

	return resp, http.StatusCreated
}

func (store *MemSQL) GetAgentByEmail(email string) (*model.Agent, int) {
	logFields := log.Fields{
		"email": email,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func (store *MemSQL) GetAgentByUUID(uuid string) (*model.Agent, int) {
	logFields := log.Fields{
		"uuid": uuid,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetAgentsByUUIDs(uuids []string) ([]*model.Agent, int) {
	logFields := log.Fields{
		"uuids": uuids,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetAgentInfo(uuid string) (*model.AgentInfo, int) {
	logFields := log.Fields{
		"uuid": uuid,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	agent, errCode := store.GetAgentByUUID(uuid)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	agentInfo := model.CreateAgentInfo(agent)
	return agentInfo, errCode
}

func (store *MemSQL) UpdateAgentIntAdwordsRefreshToken(uuid, refreshToken string) int {
	logFields := log.Fields{
		"uuid":          uuid,
		"refresh_token": refreshToken,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if uuid == "" || refreshToken == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentAdwordsRefreshToken failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntAdwordsRefreshToken(refreshToken))
}
func (store *MemSQL) UpdateAgentIntGoogleOrganicRefreshToken(uuid, refreshToken string) int {
	logFields := log.Fields{
		"uuid":          uuid,
		"refresh_token": refreshToken,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if uuid == "" || refreshToken == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentGSCRefreshToken failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntGSCRefreshToken(refreshToken))
}

func (store *MemSQL) UpdateAgentIntSalesforce(uuid, refreshToken string, instanceUrl string) int {
	logFields := log.Fields{
		"uuid":          uuid,
		"refresh_token": refreshToken,
		"instance_url":  instanceUrl,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if uuid == "" || refreshToken == "" || instanceUrl == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentIntSalesforce failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntSalesforceRefreshToken(refreshToken), model.IntSalesforceInstanceURL(instanceUrl))
}

func (store *MemSQL) UpdateAgentSalesforceInstanceURL(uuid, instanceUrl string) int {
	logFields := log.Fields{
		"uuid":         uuid,
		"instance_url": instanceUrl,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if uuid == "" || instanceUrl == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentInstanceURL failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, model.IntSalesforceInstanceURL(instanceUrl))
}

func (store *MemSQL) UpdateAgentPassword(uuid, plainTextPassword string, passUpdatedAt time.Time) int {
	logFields := log.Fields{
		"uuid":                uuid,
		"plain_text_password": plainTextPassword,
		"pass_updated_at":     passUpdatedAt,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func (store *MemSQL) UpdateAgentLastLoginInfo(agentUUID string, ts time.Time) int {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
		"ts":         ts,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if agentUUID == "" {
		log.Error("UpdateAgentLastLoginInfo Failed. Missing params")
		return http.StatusBadRequest
	}

	return updateAgent(agentUUID, model.LastLoggedInAtAndIncrLoginCount(ts))
}
func (store *MemSQL) UpdateLastLoggedOut(agentUUID string, timestamp int64) int {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
		"timestamp":  timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if agentUUID == "" {
		log.Error("Update Last Logged out Failed. Missing params")
		return http.StatusBadRequest
	}

	return updateAgent(agentUUID, model.LastLoggedOut(timestamp))
}

func (store *MemSQL) UpdateAgentVerificationDetails(agentUUID, password, firstName,
	lastName string, verified bool, passUpdatedAt time.Time) int {
	logFields := log.Fields{
		"agent_uuid":      agentUUID,
		"password":        password,
		"first_name":      firstName,
		"last_name":       lastName,
		"verified":        verified,
		"pass_updated_at": passUpdatedAt,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func (store *MemSQL) UpdateAgentVerificationDetailsFromAuth0(agentUUID, firstName, lastName string, verified bool, value *postgres.Jsonb) int {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
		"first_name": firstName,
		"last_name":  lastName,
		"verified":   verified,
		"auth0":      true,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

func (store *MemSQL) UpdateAgentEmailVerificationDetails(agentUUID string, isVerfied bool) int {
	options := make([]model.Option, 0)
	options = append(options, model.IsEmailVerified(isVerfied))
	return updateAgent(agentUUID, options...)
}

func (store *MemSQL) UpdateAgentInformation(agentUUID, firstName, lastName, phone string, isOnboardingFlowSeen *bool, isFormFilled *bool) int {
	logFields := log.Fields{
		"agent_uuid":              agentUUID,
		"first_name":              firstName,
		"last_name":               lastName,
		"phone":                   phone,
		"in_onboarding_flow_seen": isOnboardingFlowSeen,
		"is_form_filled":          isFormFilled,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	if isFormFilled != nil {
		updateParams = append(updateParams, model.IsFormFilled(*isFormFilled))
	}
	return updateAgent(agentUUID, updateParams...)
}

func updateAgent(agentUUID string, options ...model.Option) int {
	logFields := log.Fields{
		"agent_uuid": agentUUID,
		"options":    options,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetPrimaryAgentOfProject(projectId int64) (uuid string, errCode int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) IsSlackIntegratedForProject(projectID int64, agentUUID string) (bool, int) {
	agent, errCode := store.GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		return false, errCode
	}
	if agent.SlackAccessTokens == nil {
		return false, http.StatusOK
	}
	var authToken model.SlackAuthTokens
	isEmpty := U.IsEmptyPostgresJsonb(agent.SlackAccessTokens)
	if isEmpty {
		return false, http.StatusOK
	}
	err := U.DecodePostgresJsonbToStructType(agent.SlackAccessTokens, &authToken)
	if err != nil {
		log.WithError(err).Error("Failed to decode slack auth tokens")
		return false, http.StatusInternalServerError
	}
	if SlackAccessTokens, ok := authToken[projectID]; ok {
		// check if this is a valid token
		if SlackAccessTokens.BotAccessToken != "" && SlackAccessTokens.UserAccessToken != "" {
			return true, http.StatusOK
		}

	}
	return false, http.StatusOK
}
func (store *MemSQL) IsTeamsIntegratedForProject(projectID int64, agentUUID string) (bool, int) {
	agent, errCode := store.GetAgentByUUID(agentUUID)
	if errCode != http.StatusFound {
		return false, errCode
	}
	if agent.TeamsAccessTokens == nil {
		return false, http.StatusOK
	}
	var authToken model.TeamsAuthTokens
	isEmpty := U.IsEmptyPostgresJsonb(agent.TeamsAccessTokens)
	if isEmpty {
		return false, http.StatusOK
	}
	err := U.DecodePostgresJsonbToStructType(agent.TeamsAccessTokens, &authToken)
	if err != nil {
		log.WithError(err).Error("Failed to decode team auth tokens")
		return false, http.StatusInternalServerError
	}
	if teamsAccessTokens, ok := authToken[projectID]; ok {
		// check if this is a valid token
		if teamsAccessTokens.AccessToken != "" && teamsAccessTokens.RefreshToken != "" {
			return true, http.StatusOK
		}

	}
	return false, http.StatusOK
}

func (store *MemSQL) IsTeamsIntegrated(projectID int64) (bool, int) {

	var isIntegrated int64
	queryStmnt := "select  count(*) > 0  from agents where JSON_EXTRACT_STRING(JSON_EXTRACT_STRING(teams_access_tokens, '%d' ) , 'refresh_token') is not null "

	db := C.GetServices().Db

	row := db.Raw(fmt.Sprintf(queryStmnt, projectID)).Row()

	err := row.Scan(&isIntegrated)
	if err != nil {
		log.WithFields(log.Fields{"projectID": projectID, "error": err}).Error("Failed getting team integration info")
		return false, http.StatusInternalServerError
	}

	return isIntegrated == 1, http.StatusOK

}
