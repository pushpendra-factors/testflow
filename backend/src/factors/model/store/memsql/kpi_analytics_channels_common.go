package memsql

import (
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) ExecuteKPIQueryForChannels(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	logCtx := log.WithField("projectId", projectID).WithField("reqId", reqID)
	channelsV1Query, err := model.TransformKPIQueryToChannelsV1Query(kpiQuery)
	queryResults := make([]model.QueryResult, 0)
	hasAnyGroupByTimestamp := (kpiQuery.GroupByTimestamp != "")
	hasAnyGroupBy := len(kpiQuery.GroupBy) != 0
	var queryResult model.QueryResult
	if err != nil {
		logCtx.Warn(err)
		return queryResults, http.StatusBadRequest
	}
	resultHolder, statusCode := store.ExecuteChannelQueryV1(projectID, &channelsV1Query, reqID)
	queryResult.Headers = model.GetTransformedHeadersForChannels(resultHolder.Headers, hasAnyGroupByTimestamp, hasAnyGroupBy)
	queryResult.Rows = model.TransformDateTypeValueForChannels(resultHolder.Headers, resultHolder.Rows, hasAnyGroupByTimestamp, hasAnyGroupBy, channelsV1Query.Timezone)
	queryResults = append(queryResults, queryResult)
	return queryResults, statusCode
}
