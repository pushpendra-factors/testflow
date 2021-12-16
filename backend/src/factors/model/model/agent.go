package model

import (
	"time"

	"github.com/jinzhu/gorm"
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

	InvitedBy             *string   `gorm:"type:uuid" json:"invited_by"`
	IsOnboardingFlowSeen bool      `json:"is_onboarding_flow_seen"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	IsDeleted             bool      `json:"is_deleted"`

	LastLoggedInAt *time.Time `json:"last_logged_in_at"`
	LoginCount     uint64     `json:"login_count"`

	IntAdwordsRefreshToken       string `json:"-"`
	IntSalesforceInstanceURL     string `json:"int_salesforce_instance_url"`
	IntSalesforceRefreshToken    string `json:"int_salesforce_refresh_token"`
	CompanyURL                   string `json:"company_url"`
	SubscribeNewsletter          bool   `json:"subscribe_newsletter"`
	IntGoogleOrganicRefreshToken string `json:"int_google_organic_refresh_token"`
}

type CreateAgentParams struct {
	Agent    *Agent
	PlanCode string
}

type CreateAgentResponse struct {
	Agent          *Agent
	BillingAccount *BillingAccount
}

type AgentInfo struct {
	UUID            string     `json:"uuid"`
	Email           string     `json:"email"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoggedIn    *time.Time `json:"last_logged_in"`
	Phone           string     `json:"phone"`
	IsOnboardingFlowSeen bool `json:"is_onboarding_flow_seen"`
}

const (
	AgentSaltLength = 32
)

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

type FieldsToUpdate map[string]interface{}

type Option func(FieldsToUpdate)

func Firstname(firstName string) Option {
	return func(fields FieldsToUpdate) {
		fields["first_name"] = firstName
	}
}

func Lastname(lastName string) Option {
	return func(fields FieldsToUpdate) {
		fields["last_name"] = lastName
	}
}

func Phone(phone string) Option {
	return func(fields FieldsToUpdate) {
		fields["phone"] = phone
	}
}

func IsOnboardingFlowSeen(status bool) Option{
	return func(fields FieldsToUpdate) {
		fields["is_onboarding_flow_seen"] = status
	}
}
func PasswordAndPasswordCreatedAt(password string, ts time.Time) Option {
	return func(fields FieldsToUpdate) {
		fields["password"] = password
		fields["password_created_at"] = ts
	}
}

func Salt(salt string) Option {
	return func(fields FieldsToUpdate) {
		fields["salt"] = salt
	}
}

func LastLoggedInAtAndIncrLoginCount(time time.Time) Option {
	return func(fields FieldsToUpdate) {
		fields["last_logged_in_at"] = time
		fields["login_count"] = gorm.Expr("login_count + ? ", 1)
	}
}

func IntAdwordsRefreshToken(refreshToken string) Option {
	return func(fields FieldsToUpdate) {
		fields["int_adwords_refresh_token"] = refreshToken
	}
}
func IntGSCRefreshToken(refreshToken string) Option {
	return func(fields FieldsToUpdate) {
		fields["int_google_organic_refresh_token"] = refreshToken
	}
}

func IntSalesforceRefreshToken(refreshToken string) Option {
	return func(fields FieldsToUpdate) {
		fields["int_salesforce_refresh_token"] = refreshToken
	}
}

func IntSalesforceInstanceURL(instanceUrl string) Option {
	return func(fields FieldsToUpdate) {
		fields["int_salesforce_instance_url"] = instanceUrl
	}
}

func IsEmailVerified(verified bool) Option {
	return func(fields FieldsToUpdate) {
		fields["is_email_verified"] = verified
	}
}
