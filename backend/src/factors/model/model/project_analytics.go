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
