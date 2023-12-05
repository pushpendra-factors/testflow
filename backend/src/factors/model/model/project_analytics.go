package model

type ProjectAnalytics struct {
	Date                  string `json:"date"`
	ProjectID             string `json:"project_id"`
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
	DailyLoginCount       int64  `json:"daily_login_count"`
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
	"Daily Login Count",
}

var GlobalDataProjectAnalyticsColumnsName = []string{
	"User Count",
	"Alerts Count",
	"People Segments Count",
	"Accounts Segments Count",
	"Dashboard Count",
	"Webhooks Count",
	"Report Count",
	"SDK Integration Completed",
	"Identified Count",
	"Login Count",
}

var GlobalDataIntegrationListColumnsName = []string{
	"Integration Connected",
	"Integration Disconnected",
}

var GlobalDataIntegrationListColumnsNameToJsonKeys = map[string]string{
	"Integration Connected":    "connected",
	"Integration Disconnected": "disconnected",
}

var GlobalDataProjectAnalyticsColumnsNameToJsonKeys = map[string]string{
	"User Count":                "user_count",
	"Alerts Count":              "alerts_count",
	"People Segments Count":     "segment_count_user",
	"Accounts Segments Count":   "segment_count_account",
	"Dashboard Count":           "dashboard_count",
	"Webhooks Count":            "webhooks_count",
	"Report Count":              "report_count",
	"Login Count":               "login_count",
	"SDK Integration Completed": "sdk_int_completed",
	"Identified Count":          "identified_count",
	"Integration Connected":     "integration_connected",
	"Integration Disconnected":  "integration_disconnected",
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
	"Daily Login Count":               "daily_login_count",
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

var LoginCountQueryStmnt = `
        {
            "cl": "events",
            "ty": "unique_users",
            "grpa": "users",
            "ewp": [
                {
                    "an": "",
                    "na": "app.factors.ai",
                    "grpa": "Page Views",
                    "pr": [
                        {
                            "en": "user",
                            "grpn": "user",
                            "lop": "AND",
                            "op": "notEqual",
                            "pr": "email",
                            "ty": "categorical",
                            "va": "$none"
                        }
                    ]
                }
            ],
            "gup": [
				{
				  "en": "user_g",
				  "grpn": "user",
				  "lop": "AND",
				  "op": "equals",
				  "pr": "project_id",
				  "ty": "categorical",
				  "va": "%v"
				}
			  ],
            "gbt": "",
            "gbp": [
                {
                    "pr": "project_id",
                    "en": "user",
                    "pty": "categorical",
                    "grpn": "OTHERS",
                    "ena": "$present"
                }
            ],
            "ec": "each_given_event",
            "tz": "%v",
            "fr": %v,
            "to": %v
        }
    `

var DailyLoginCountQueryStmnt = `
	{
		"cl": "events",
		"ty": "unique_users",
		"grpa": "users",
		"ewp": [
			{
				"an": "",
				"na": "app.factors.ai",
				"grpa": "Page Views",
				"pr": [
					{
						"en": "user",
						"grpn": "user",
						"lop": "AND",
						"op": "notEqual",
						"pr": "email",
						"ty": "categorical",
						"va": "$none"
					}
				]
			}
		],
		"gup": [
			{
			  "en": "user_g",
			  "grpn": "user",
			  "lop": "AND",
			  "op": "equals",
			  "pr": "project_id",
			  "ty": "categorical",
			  "va": "%v"
			}
		  ],
		"gbt": "date",
		"gbp": [
			{
				"pr": "project_id",
				"en": "user",
				"pty": "categorical",
				"grpn": "OTHERS",
				"ena": "$present"
			}
		],
		"ec": "each_given_event",
		"tz": "%v",
		"fr": %v,
		"to": %v
	}
`

var ProjectAnalyticsEventSingleQueryStmnt = map[string]string{
	"login_count":       LoginCountQueryStmnt,
	"daily_login_count": DailyLoginCountQueryStmnt,
}
