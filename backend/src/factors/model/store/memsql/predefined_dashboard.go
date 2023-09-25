package memsql

func (store *MemSQL) CreatePredefinedDashboards(projectID int64, agentUUID string) int {
	statusCode := store.CreatePredefinedWebsiteAggregation(projectID, agentUUID)
	return statusCode
}
