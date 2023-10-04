package memsql

const (
	DateTruncateInURlPWA = "UNIX_TIMESTAMP(date_trunc('%s', FROM_UNIXTIME(timestamp_at_day)))"
)

func (store *MemSQL) CreatePredefinedDashboards(projectID int64, agentUUID string) int {
	statusCode := store.CreatePredefinedWebsiteAggregation(projectID, agentUUID)
	return statusCode
}
