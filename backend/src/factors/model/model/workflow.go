package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type WorkflowTemplateConfig struct {
	Description  string                 `json:"description"`
	Integrations []string               `json:"integrations"`
	Payload      map[string]interface{} `json:"payload"`
	Published    string                 `json:"published"`
	Title        string                 `json:"title"`
	Trigger      WorkflowAlertBody      `json:"trigger"`
}

type WorkflowStaticDataFields struct {
	ShortDescription string   `json:"short_description"`
	LongDescription  string   `json:"long_description"`
	ImageAddress     string   `json:"image_address"`
	Category         []string `json:"category"`
	Tags             []string `json:"tags"`
	Integrations     []string `json:"integrations"`
}

type Workflow struct {
	ID             string          `json:"id"`
	ProjectID      int64           `json:"project_id"`
	Name           string          `json:"name"`
	AlertBody      *postgres.Jsonb `json:"alert_body"`
	InternalStatus string            `json:"internal_status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	CreatedBy      string          `json:"created_by"`
	IsDeleted      bool            `gorm:"not null;default:false" json:"is_deleted"`
}

type WorkflowDisplayableInfo struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	AlertBody *postgres.Jsonb `json:"alert_body"`
	Status    string            `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

type WorkflowAlertBody struct {
	Title               string          `json:"title"`
	Description         string          `json:"description"`
	TemplateTitle       string          `json:"template_title"`
	TemplateID          int             `json:"template_id"`
	TemplateDescription string          `json:"template_description"`
	EventLevel          string          `json:"event_level"`
	ActionPerformed     string          `json:"action_performed"`
	Event               string          `json:"event"`
	Filters             []QueryProperty `json:"filters"`
	DontRepeatAlerts    bool            `json:"repeat_alerts"`
	CoolDownTime        int64           `json:"cool_down_time"`
	BreakdownProperties []QueryProperty `json:"breakdown_properties"`
	SetAlertLimit       bool            `json:"notifications"`
	AlertLimit          int64           `json:"alert_limit"`
	MessageProperties   *postgres.Jsonb `json:"message_properties"`
}

type WorkflowPropertiesMapping struct {
	Factors string `json:"factors"`
	Others  string `json:"others"`
}

type WorkflowMappingDetails map[string]string

// Different config to be used in the payloads

type WorkflowPayloadProperties map[string]interface{}

type WorkflowParagonPayload struct {
	MandatoryPropsCompany  map[string]interface{} `json:"mandatory_props_companykey"`
	AdditionalPropsCompany map[string]interface{} `json:"additional_props_company"`
}

type ApolloWorkflowConfig struct {
	ApiKey            string   `json:"api_key"`
	PersonLocations   []string `json:"person_locations"`
	PersonSeniorities []string `json:"person_seniorities"`
	PersonTitles      []string `json:"person_titles"`
	MaxContacts       int64    `json:"max_contacts"`
	AccountLabedID    string   `json:"account_label_id"`
	SequenceID        string   `json:"sequence_id"`
	ContactLabelName  string   `json:"contact_label_name"`
	EmailID           string   `json:"email_id"`
}

type LinkedInAudienceCreationWorkflowConfig struct {
	AudienceName string `json:"audience_name"`
	AccountID    string `json:"account_id"`
	ProjectID    int64  `json:"project_id"`
}

type LinkedInWorkflowConfig struct {
	AudienceID     int64                  `json:"audience_id"`
	FacilitatorKey string                 `json:"facilitator_key"`
	Properties     map[string]interface{} `json:"properties"`
}
