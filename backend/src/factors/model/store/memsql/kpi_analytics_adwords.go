package memsql

import (
	"factors/model/model"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigsForAdwords(projectID uint64, reqID string) (map[string]interface{}, int) {
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
	logCtx := log.WithField("projectId", projectID).WithField("reqId", reqID)
	channelsV1Query, err := model.TransformKPIQueryToChannelsV1Query(kpiQuery)
	queryResults := make([]model.QueryResult, 0)
	var queryResult model.QueryResult
	if err != nil {
		logCtx.Warn(err)
		return queryResults, http.StatusBadRequest
	}
	resultHolder, statusCode := store.ExecuteChannelQueryV1(projectID, &channelsV1Query, reqID)
	queryResult.Headers = model.GetTransformedHeadersForChannels(resultHolder.Headers)
	queryResult.Rows = resultHolder.Rows
	queryResults = append(queryResults, queryResult)
	return queryResults, statusCode
}
