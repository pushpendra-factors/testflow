package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

const (
	CategoryWebAnalytics        = "Web Analytics"
	CategoryPaidMarketing       = "Paid Marketing"
	CategoryOrganicPerformance  = "Organic Performance"
	CategoryLandingPageAnalysis = "Landing Page analysis"
	CategoryCRMInsights         = "CRM Insights"
	CategoryFullFunnelMarketing = "Full Funnel Marketing"
	CategoryExecutiveDashboards = "Executive Dashboards"

	IntegrationWebsiteSDK          = "website_sdk"
	IntegrationSegment             = "segment"
	IntegrationMarketo             = "marketo"
	IntegrationHubspot             = "hubspot"
	IntegrationSalesforce          = "salesforce"
	IntegrationAdwords             = "adwords"
	IntegrationFacebook            = "facebook"
	IntegrationLinkedin            = "linkedin"
	IntegrationGoogleSearchConsole = "google_Search_console"
	IntegrationBing                = "bing"
	IntegrationClearbit            = "clearbit"
	IntegrationLeadsquared         = "leadsquared"
	Integration6Signal             = "6signal"
)

type UnitInfo struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Presentation string `json:"presentation"`
	// dashboard details
	Position int `json:"position"`
	Size     int `json:"size"`
	// query details
	QueryType     int            `json:"query_type"`
	QuerySettings postgres.Jsonb `json:"query_settings"`
	Query         postgres.Jsonb `json:"query"`
}

type DashboardTemplate struct {
	ID                   string          `gorm:"primary_key:true" json:"id"`
	Title                string          `gorm:"not null" json:"title"`
	Description          string          `json:"description"`
	Dashboard            *postgres.Jsonb `json:"dashboard"`
	Units                *postgres.Jsonb `json:"units"`
	SimilarTemplateIds   *postgres.Jsonb `json:"similar_template_ids"`
	Tags                 *postgres.Jsonb `json:"tags"`
	Categories           *postgres.Jsonb `json:"categories"`
	RequiredIntegrations *postgres.Jsonb `json:"required_integrations"`
	IsDeleted            bool            `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
