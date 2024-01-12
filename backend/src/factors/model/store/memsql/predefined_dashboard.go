package memsql

import (
	"factors/model/model"
	"net/http"
	log "github.com/sirupsen/logrus"
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

// TODO check if agentUUID should be passed as param.
func (store *MemSQL) CreatePredefWebAggDashboardIfNotExists(projectID int64) int {
	if store.IsPredefWebAggDashboardExists(projectID, "") {
		return http.StatusFound
	}

	agentUUID, statusCode := store.GetPrimaryAgentOfProject(projectID)
	if statusCode != http.StatusFound {
		log.WithField("projectID", projectID).Warn("Failed in getting primary agent")
		return http.StatusInternalServerError
	}
	return store.CreatePredefinedWebsiteAggregation(projectID, agentUUID)
}