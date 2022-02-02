package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForAdwords(projectID uint64, reqID string) (map[string]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id": reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	adwordsSettings, errCode := store.GetIntAdwordsProjectSettingsForProjectID(projectID)
	if errCode != http.StatusOK {
		return nil, http.StatusOK
	}
	if len(adwordsSettings) == 0 {
		return nil, http.StatusOK
	}
	config := model.GetKPIConfigsForAdwords()
	adwordsObjectsAndProperties := store.buildObjectAndPropertiesForAdwords(projectID, model.ObjectsForAdwords)
	config["properties"] = model.TransformChannelsPropertiesConfigToKpiPropertiesConfig(adwordsObjectsAndProperties)
	return config, http.StatusOK
}

func (store *MemSQL) ExecuteKPIQueryForChannels(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id": reqID,
		"kpi_query": kpiQuery,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
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
