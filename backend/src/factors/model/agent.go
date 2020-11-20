package model

import (
	C "factors/config"
	"net/http"
	"strings"
	"time"

	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	SALT_LEN = 32
)

type Agent struct {
	UUID string `gorm:"primary_key:true;type:varchar(255);default:uuid_generate_v4()" json:"uuid"`

	FirstName string `gorm:"type:varchar(100)" json:"first_name"`
	LastName  string `gorm:"type:varchar(100)" json:"last_name"`

	Email           string `gorm:"type:varchar(100)" json:"email"`
	IsEmailVerified bool   `json:"is_email_verified"`

	Phone string `gorm:"type:varchar(100)" json:"phone"`

	Salt              string     `gorm:"type:varchar(100)" json:"-"` // Should we add a unique on salt ?
	Password          string     `gorm:"type:varchar(100)" json:"-"`
	PasswordCreatedAt *time.Time `json:"password_created_at"`

	InvitedBy *string `gorm:"type:uuid" json:"invited_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`

	LastLoggedInAt *time.Time `json:"last_logged_in_at"`
	LoginCount     uint64     `json:"login_count"`

	IntAdwordsRefreshToken    string `json:"-"`
	IntSalesforceInstanceURL  string `json:"int_salesforce_instance_url"`
	IntSalesforceRefreshToken string `json:"int_salesforce_refresh_token"`
	CompanyURL                string `json:"company_url"`
	SubscribeNewsletter       bool   `json:"subscribe_newsletter"`
}

// AgentInfo - Exposable Info.
type AgentInfo struct {
	UUID            string     `json:"uuid"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoggedIn    *time.Time `json:"last_logged_in"`
	Phone           string     `json:"phone"`
}

func (a *Agent) BeforeCreate(scope *gorm.Scope) error {
	a.Salt = U.RandomString(SALT_LEN)
	return nil
}

// TODO: Make index name a constant and read it
const error_Duplicate_email_error = "pq: duplicate key value violates unique constraint \"agent_email_unique_idx\""

func createAgent(agent *Agent) (*Agent, int) {
	if agent.Email == "" {
		log.Error("CreateAgent Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	agent.Email = strings.ToLower(agent.Email)

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

type CreateAgentParams struct {
	Agent    *Agent
	PlanCode string
}

type CreateAgentResponse struct {
	Agent          *Agent
	BillingAccount *BillingAccount
}

func CreateAgentWithDependencies(params *CreateAgentParams) (*CreateAgentResponse, int) {

	if params == nil || params.PlanCode == "" || params.Agent == nil || params.Agent.Email == "" {
		return nil, http.StatusBadRequest
	}

	resp := &CreateAgentResponse{}

	agent, errCode := createAgent(params.Agent)
	if errCode != http.StatusCreated {
		return nil, errCode
	}
	resp.Agent = agent

	billingAccount, errCode := createBillingAccount(params.PlanCode, agent.UUID)
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	resp.BillingAccount = billingAccount

	return resp, http.StatusCreated
}

func GetAgentByEmail(email string) (*Agent, int) {

	if email == "" {
		log.Error("GetAgentByEmail Failed. Email not provided.")
		return nil, http.StatusBadRequest
	}

	email = strings.ToLower(email)

	db := C.GetServices().Db

	var agent Agent
	if err := db.Limit(1).Where("email = ?", email).Find(&agent).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &agent, http.StatusFound
}

func GetAgentByUUID(uuid string) (*Agent, int) {
	if uuid == "" {
		log.Error("GetAgentByUUID Failed. UUID not provided.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var agent Agent

	if err := db.Limit(1).Where("uuid = ?", uuid).Find(&agent).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithError(err).Error("GetAgentByUUID Failed.")
		return nil, http.StatusInternalServerError
	}

	return &agent, http.StatusFound
}

func GetAgentsByUUIDs(uuids []string) ([]*Agent, int) {
	if len(uuids) == 0 {
		return nil, http.StatusBadRequest
	}
	db := C.GetServices().Db

	agents := make([]*Agent, 0, 0)

	if err := db.Limit(len(uuids)).Where("uuid IN (?)", uuids).Find(&agents).Error; err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(agents) == 0 {
		return nil, http.StatusNotFound
	}

	return agents, http.StatusFound
}

func GetAgentInfo(uuid string) (*AgentInfo, int) {
	agent, errCode := GetAgentByUUID(uuid)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	agentInfo := CreateAgentInfo(agent)
	return agentInfo, errCode
}

func CreateAgentInfo(agent *Agent) *AgentInfo {
	if agent == nil {
		return nil
	}
	return &AgentInfo{
		FirstName:       agent.FirstName,
		LastName:        agent.LastName,
		Email:           agent.Email,
		UUID:            agent.UUID,
		IsEmailVerified: agent.IsEmailVerified,
		LastLoggedIn:    agent.LastLoggedInAt,
		Phone:           agent.Phone,
	}
}

func CreateAgentInfos(agents []*Agent) []*AgentInfo {
	agentInfos := make([]*AgentInfo, 0, 0)
	for _, agent := range agents {
		agentInfos = append(agentInfos, CreateAgentInfo(agent))
	}
	return agentInfos
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func IsPasswordAndHashEqual(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func UpdateAgentIntAdwordsRefreshToken(uuid, refreshToken string) int {
	if uuid == "" || refreshToken == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentAdwordsRefreshToken failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, IntAdwordsRefreshToken(refreshToken))
}

func UpdateAgentIntSalesforce(uuid, refreshToken string, instanceUrl string) int {
	if uuid == "" || refreshToken == "" || instanceUrl == "" {
		log.WithField("agent_uuid", uuid).Error(
			"UpdateAgentIntSalesforce failed. Invalid params.")
		return http.StatusBadRequest
	}

	return updateAgent(uuid, IntSalesforceRefreshToken(refreshToken), IntSalesforceInstanceURL(instanceUrl))
}

func UpdateAgentPassword(uuid, plainTextPassword string, passUpdatedAt time.Time) int {

	if uuid == "" || plainTextPassword == "" {
		log.Error("UpdateAgentPassword Failed. Missing params")
		return http.StatusBadRequest
	}

	hashedPassword, err := HashPassword(plainTextPassword)
	if err != nil {
		return http.StatusInternalServerError
	}

	return updateAgent(uuid, PasswordAndPasswordCreatedAt(hashedPassword, passUpdatedAt),
		Salt(U.RandomString(SALT_LEN)))
}

func UpdateAgentLastLoginInfo(agentUUID string, ts time.Time) int {
	if agentUUID == "" {
		log.Error("UpdateAgentLastLoginInfo Failed. Missing params")
		return http.StatusBadRequest
	}

	return updateAgent(agentUUID, LastLoggedInAtAndIncrLoginCount(ts))
}

func UpdateAgentVerificationDetails(agentUUID, password, firstName,
	lastName string, verified bool, passUpdatedAt time.Time) int {

	if agentUUID == "" {
		log.Error("UpdateAgentVerificationDetails Failed. Missing params")
		return http.StatusBadRequest
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return http.StatusInternalServerError
	}

	options := make([]Option, 0)
	if firstName != "" {
		options = append(options, Firstname(firstName))
	}
	if lastName != "" {
		options = append(options, Lastname(lastName))
	}
	options = append(options, IsEmailVerified(verified))
	options = append(options, PasswordAndPasswordCreatedAt(hashedPassword, passUpdatedAt))
	return updateAgent(agentUUID, options...)
}

func UpdateAgentInformation(agentUUID, firstName, lastName, phone string) int {
	if agentUUID == "" {
		return http.StatusBadRequest
	}
	updateParams := []Option{}
	if firstName != "" {
		updateParams = append(updateParams, Firstname(firstName))
	}
	if lastName != "" {
		updateParams = append(updateParams, Lastname(lastName))
	}
	if phone != "" {
		updateParams = append(updateParams, Phone(phone))
	}
	return updateAgent(agentUUID, updateParams...)
}

type fieldsToUpdate map[string]interface{}

type Option func(fieldsToUpdate)

func Firstname(firstName string) Option {
	return func(fields fieldsToUpdate) {
		fields["first_name"] = firstName
	}
}

func Lastname(lastName string) Option {
	return func(fields fieldsToUpdate) {
		fields["last_name"] = lastName
	}
}

func Phone(phone string) Option {
	return func(fields fieldsToUpdate) {
		fields["phone"] = phone
	}
}

func PasswordAndPasswordCreatedAt(password string, ts time.Time) Option {
	return func(fields fieldsToUpdate) {
		fields["password"] = password
		fields["password_created_at"] = ts
	}
}

func Salt(salt string) Option {
	return func(fields fieldsToUpdate) {
		fields["salt"] = salt
	}
}

func LastLoggedInAtAndIncrLoginCount(time time.Time) Option {
	return func(fields fieldsToUpdate) {
		fields["last_logged_in_at"] = time
		fields["login_count"] = gorm.Expr("login_count + ? ", 1)
	}
}

func IntAdwordsRefreshToken(refreshToken string) Option {
	return func(fields fieldsToUpdate) {
		fields["int_adwords_refresh_token"] = refreshToken
	}
}

func IntSalesforceRefreshToken(refreshToken string) Option {
	return func(fields fieldsToUpdate) {
		fields["int_salesforce_refresh_token"] = refreshToken
	}
}

func IntSalesforceInstanceURL(instanceUrl string) Option {
	return func(fields fieldsToUpdate) {
		fields["int_salesforce_instance_url"] = instanceUrl
	}
}

func IsEmailVerified(verified bool) Option {
	return func(fields fieldsToUpdate) {
		fields["is_email_verified"] = verified
	}
}

func updateAgent(agentUUID string, options ...Option) int {
	if agentUUID == "" {
		return http.StatusBadRequest
	}

	fields := fieldsToUpdate{}

	for _, option := range options {
		option(fields)
	}

	if len(fields) == 0 {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db

	db = db.Model(&Agent{}).Where("uuid = ?", agentUUID).Updates(fields)

	if db.Error != nil {
		log.WithError(db.Error).Error("UpdateAgent Failed")
		return http.StatusInternalServerError
	}
	if db.RowsAffected == 0 {
		return http.StatusNoContent
	}
	return http.StatusAccepted
}
