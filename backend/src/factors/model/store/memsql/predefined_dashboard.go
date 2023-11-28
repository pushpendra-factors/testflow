package memsql

import (
	"factors/model/model"
)

const (
	DateTruncateInURlPWA 		= "UNIX_TIMESTAMP(date_trunc('%s', FROM_UNIXTIME(timestamp_at_day)))"
	DateTruncateInURlForWeekPWA = "UNIX_TIMESTAMP(date_trunc('week', FROM_UNIXTIME(timestamp_at_day + (24*60*60) )) - INTERVAL 1 day )"
)

var mapOfGroupByTimestampToDateTrunc = map[string]string {
	model.GroupByTimestampDate: "day",
}

func (store *MemSQL) CreatePredefinedDashboards(projectID int64, agentUUID string) int {
	statusCode := store.CreatePredefinedWebsiteAggregation(projectID, agentUUID)
	return statusCode
}
