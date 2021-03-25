package model

type ProjectAnalytics struct {
	ProjectID         uint64 `json:"project_id"`
	ProjectName       string `json:"project_name"`
	TotalEvents       uint64 `json:"total_events"`
	TotalUniqueEvents uint64 `json:"total_unique_events"`
	TotalUniqueUsers  uint64 `json:"total_unique_users"`
}