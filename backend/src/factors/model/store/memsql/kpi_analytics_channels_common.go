package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) ExecuteKPIQueryForChannels(projectID int64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
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
	err = sanitizeChannelQueryResult(&queryResult, kpiQuery)
	if err != nil {
		logCtx.Warn(err)
		return queryResults, http.StatusBadRequest
	}
	queryResults = append(queryResults, queryResult)
	return queryResults, statusCode
}

//NOTE: The following methods can be merged with methods from event_analytics later on
func sanitizeChannelQueryResult(result *model.QueryResult, query model.KPIQuery) error {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	hasAnyGroupByTimestamp := (query.GroupByTimestamp != "")
	if hasAnyGroupByTimestamp {
		aggrIndex, timeIndex, err := GetTimstampAndAggregateIndexOnQueryResult(result.Headers)
		if err != nil {
			return err
		}
		if len(query.GroupBy) == 0 {
			err = addMissingTimestampsOnChannelResultWithoutGroupByProps(result, &query, aggrIndex, timeIndex, true)
		} else {
			err = addMissingTimestampsOnChannelResultWithGroupByProps(result, &query, aggrIndex, timeIndex, true)
		}

		if err != nil {
			return err
		}

		sortResultRowsByTimestamp(result.Rows, timeIndex)
	}

	return nil
}
