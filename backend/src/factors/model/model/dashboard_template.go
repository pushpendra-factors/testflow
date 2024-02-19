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
	IntegrationGoogleSearchConsole = "google_search_console"
	IntegrationBing                = "bingads"
	IntegrationClearbit            = "clearbit"
	IntegrationLeadsquared         = "leadsquared"
	Integration6Signal             = "6signal"
	IntegrationG2                  = "g2"
)

// Please avoid changing these constants. May cause impact across project.
// Helps in fetching id of templates on FE.
const (
	ALL_PAID_MARKETING                 = "ALL_PAID_MARKETING"
	HUBSPOT_CONTACT_ATTRIBUTION        = "HUBSPOT_CONTACT_ATTRIBUTION"
	GOOGLE_ADWORDS                     = "GOOGLE_ADWORDS"
	PAID_SEARCH_MARKETING              = "PAID_SEARCH_MARKETING"
	PAID_SEARCH_TRAFFIC                = "PAID_SEARCH_TRAFFIC"
	PAID_SOCIAL_MARKETING              = "PAID_SOCIAL_MARKETING"
	ORGANIC_PERFORMANCE                = "ORGANIC_PERFORMANCE"
	OVERALL_REPORTING                  = "OVERALL_REPORTING"
	WEBSITE_VISITOR_IDENTIFICATION     = "WEBSITE_VISITOR_IDENTIFICATION"
	WEB_ANALYTICS                      = "WEB_ANALYTICS"
	WEB_KPIS_AND_OVERVIEW              = "WEB_KPIS_AND_OVERVIEW"
	HUBSPOT_INSIGHTS                   = "HUBSPOT_INSIGHTS"
	GOOGLE_SEARCH_CONSOLE              = "GOOGLE_SEARCH_CONSOLE"
	G2_INFLLUENCE_SALESFORCE           = "G2_INFLLUENCE_SALESFORCE"
	G2_INFLUENCE_HUBSPOT               = "G2_INFLUENCE_HUBSPOT"
	LINKEDIN_INFLUENCE_SALESFORCE      = "LINKEDIN_INFLUENCE_SALESFORCE"
	LINKEDIN_INFLUENCE_HUBSPOT         = "LINKEDIN_INFLUENCE_HUBSPOT"
	BLOG_CONVERSION_SUMMARY_SALESFORCE = "BLOG_CONVERSION_SUMMARY_SALESFORCE"
	BLOG_CONVERSION_SUMMARY_HUBSPOT    = "BLOG_CONVERSION_SUMMARY_HUBSPOT"
	BLOGS_TRAFFIC_SUMMARY              = "BLOGS_TRAFFIC_SUMMARY"
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
	Type                 string          `gorm:"not null" json:"type"`
	IsDeleted            bool            `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}
