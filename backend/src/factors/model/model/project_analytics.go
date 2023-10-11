package model

type ProjectAnalytics struct {
	Date                  string `json:"date"`
	ProjectID             int64  `json:"project_id"`
	ProjectName           string `json:"project_name"`
	AdwordsEvents         uint64 `json:"adwords_events"`
	FacebookEvents        uint64 `json:"facebook_events"`
	HubspotEvents         uint64 `json:"hubspot_events"`
	LinkedinEvents        uint64 `json:"linkedin_events"`
	SalesforceEvents      uint64 `json:"salesforce_events"`
	TotalEvents           uint64 `json:"total_events"`
	TotalUniqueEvents     uint64 `json:"total_unique_events"`
	TotalUniqueUsers      uint64 `json:"total_unique_users"`
	SixSignalAPIHits      uint64 `json:"six_signal_api_hits"`
	SixSignalAPITotalHits uint64 `json:"six_signal_api_total_hits"`
}

var ProjectAnalyticsColumnsName = []string{
	"Date",
	"Project Name",
	"Adwords",
	"Facebook",
	"Hubspot",
	"Linkedin",
	"Salesforce",
	"Total Events",
	"Total Unique Events",
	"Total Unique Users",
	"6Signal Domain Enrichment Count",
	"6Signal Total API Hits",
}

var AllProjectAnalyticsColumnsName = []string{
	"Project ID",
	"Project Name",
	"Adwords",
	"Facebook",
	"Hubspot",
	"Linkedin",
	"Salesforce",
	"Total Events",
	"Total Unique Events",
	"Total Unique Users",
	"6Signal Domain Enrichment Count",
	"6Signal Total API Hits",
}

var ProjectAnalyticsColumnsNameToJsonKeys = map[string]string{

	"Date":                            "date",
	"Project Name":                    "project_name",
	"Adwords":                         "adwords_events",
	"Facebook":                        "facebook_events",
	"Hubspot":                         "hubspot_events",
	"Linkedin":                        "linkedin_events",
	"Salesforce":                      "salesforce_events",
	"Total Events":                    "total_events",
	"Total Unique Events":             "total_unique_events",
	"Total Unique Users":              "total_unique_users",
	"6Signal Domain Enrichment Count": "six_signal_api_hits",
	"6Signal Total API Hits":          "six_signal_api_total_hits",
	"Project ID":                      "project_id",
}

var CrmStatusColumnsName = []string{
	"Document Type",
	"Action",
	"Total Pulled",
	"Total Enriched",
	"Yet To Be Enriched",
}

var CrmStatusColumnsNameToJsonKeys = map[string]string{
	"Document Type":      "document_type",
	"Action":             "action",
	"Total Pulled":       "total_pulled",
	"Total Enriched":     "total_enriched",
	"Yet To Be Enriched": "yet_to_be_enriched",
}
