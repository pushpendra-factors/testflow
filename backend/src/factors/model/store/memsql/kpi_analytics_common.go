package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// statusCode need to be clear on http.StatusOk or http.StatusAccepted or something else.
// Below function relies on fact that each query has only one metric.
// Note: All of the hash functions use the query without GBT to form keys.
func (store *MemSQL) ExecuteKPIQueryGroup(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int) {

	kpiTimezoneString := string(kpiQueryGroup.GetTimeZone())
	var finalResultantResults []model.QueryResult

	for index, query := range kpiQueryGroup.Queries {
		kpiQueryGroup.Queries[index].Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		kpiQueryGroup.Queries[index].GroupBy = kpiQueryGroup.GlobalGroupBy
	}

	finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults,
		mapOfGBTKPINormalQueryToResults, externalQueryToInternalQueries := store.ExecuteKPIQueriesAndGetResultsAsMap(projectID,
		reqID, kpiQueryGroup, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
	if finalStatusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, finalStatusCode
	}

	finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults = model.GetNonGBTResultsFromGBTResultsAndMaps(reqID, kpiQueryGroup,
		mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults, mapOfGBTKPINormalQueryToResults, externalQueryToInternalQueries)
	if finalStatusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, finalStatusCode
	}

	finalResultantResults, finalStatusCode = model.GetFinalResultantResultsForKPI(reqID, kpiQueryGroup, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults,
		mapOfNonGBTKPINormalQueryToResults, mapOfGBTKPINormalQueryToResults, externalQueryToInternalQueries)
	if finalStatusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, finalStatusCode
	}

	gbtRelatedQueryResults, nonGbtRelatedQueryResults, gbtRelatedQueries, nonGbtRelatedQueries := model.SplitQueryResultsIntoGBTAndNonGBT(finalResultantResults, kpiQueryGroup, finalStatusCode)
	finalQueryResult := make([]model.QueryResult, 0)
	gbtRelatedMergedResults := model.MergeQueryResults(gbtRelatedQueryResults, gbtRelatedQueries, kpiTimezoneString, finalStatusCode)
	nonGbtRelatedMergedResults := model.MergeQueryResults(nonGbtRelatedQueryResults, nonGbtRelatedQueries, kpiTimezoneString, finalStatusCode)
	if (!reflect.DeepEqual(model.QueryResult{}, gbtRelatedMergedResults)) {
		finalQueryResult = append(finalQueryResult, gbtRelatedMergedResults)
	}
	if (!reflect.DeepEqual(model.QueryResult{}, nonGbtRelatedMergedResults)) {
		finalQueryResult = append(finalQueryResult, nonGbtRelatedMergedResults)
	}
	return finalQueryResult, finalStatusCode
}

func (store *MemSQL) ExecuteKPIQueriesAndGetResultsAsMap(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (int, map[string]map[string][]model.QueryResult,
	map[string]map[string][]model.QueryResult, map[string][]model.QueryResult, map[string][]model.QueryResult, map[string]model.KPIQueryGroup) {
	finalStatusCode := http.StatusOK

	mapOfGBTDerivedKPIToInternalKPIToResults := make(map[string]map[string][]model.QueryResult)
	mapOfNonGBTDerivedKPIToInternalKPIToResults := make(map[string]map[string][]model.QueryResult)

	mapOfGBTKPINormalQueryToResults := make(map[string][]model.QueryResult)
	mapOfNonGBTKPINormalQueryToResults := make(map[string][]model.QueryResult)
	externalQueryToInternalQueries := make(map[string]model.KPIQueryGroup)

	for _, query := range kpiQueryGroup.Queries {
		var result []model.QueryResult
		var statusCode int
		var errMsg string
		internalKPIQuery := model.KPIQueryGroup{}
		internalQueryToQueryResult := make(map[string][]model.QueryResult)

		if query.QueryType == model.KpiDerivedQueryType {
			var derivedKPIHashCode string
			internalKPIQuery, internalQueryToQueryResult, statusCode, derivedKPIHashCode, errMsg = store.ExecuteDerivedKPIQuery(projectID, reqID, query, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			if statusCode != http.StatusOK {
				finalStatusCode = statusCode
				mapOfGBTDerivedKPIToInternalKPIToResults = make(map[string]map[string][]model.QueryResult)
				mapOfNonGBTDerivedKPIToInternalKPIToResults = make(map[string]map[string][]model.QueryResult)

				mapOfGBTKPINormalQueryToResults = make(map[string][]model.QueryResult)
				mapOfNonGBTKPINormalQueryToResults = make(map[string][]model.QueryResult)
				log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).Error(errMsg)
				break
			} else {
				if query.GroupByTimestamp == "" {
					mapOfNonGBTDerivedKPIToInternalKPIToResults[derivedKPIHashCode] = internalQueryToQueryResult
				} else {
					mapOfGBTDerivedKPIToInternalKPIToResults[derivedKPIHashCode] = internalQueryToQueryResult
				}

				externalQueryToInternalQueries[derivedKPIHashCode] = internalKPIQuery
			}
		} else {
			var hashCode string
			result, statusCode, hashCode, errMsg = store.ExecuteNonDerivedQuery(projectID, reqID, query, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			if statusCode != http.StatusOK {
				finalStatusCode = statusCode
				mapOfGBTDerivedKPIToInternalKPIToResults = make(map[string]map[string][]model.QueryResult)
				mapOfNonGBTDerivedKPIToInternalKPIToResults = make(map[string]map[string][]model.QueryResult)

				mapOfGBTKPINormalQueryToResults = make(map[string][]model.QueryResult)
				mapOfNonGBTKPINormalQueryToResults = make(map[string][]model.QueryResult)
				log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).WithField("result", result).Error(errMsg)
				break
			} else {
				if query.GroupByTimestamp == "" {
					mapOfNonGBTKPINormalQueryToResults[hashCode] = result
				} else {
					mapOfGBTKPINormalQueryToResults[hashCode] = result
				}
			}
		}
	}

	return finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults, mapOfGBTKPINormalQueryToResults, externalQueryToInternalQueries
}

func (store *MemSQL) ExecuteDerivedKPIQuery(projectID int64, reqID string, baseQuery model.KPIQuery,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (model.KPIQueryGroup, map[string][]model.QueryResult, int, string, string) {
	internalKPIQueryGroup := model.KPIQueryGroup{}

	queryResults := make([]model.QueryResult, 0)
	mapOfInternalQueryToResult := make(map[string][]model.QueryResult)

	derivedMetric, errMsg, statusCode := store.GetDerivedMetricsByName(projectID, baseQuery.Metrics[0])
	if statusCode != http.StatusFound {
		return internalKPIQueryGroup, mapOfInternalQueryToResult, statusCode, "", errMsg
	}

	err := U.DecodePostgresJsonbToStructType(derivedMetric.Transformations, &internalKPIQueryGroup)
	if err != nil {
		return internalKPIQueryGroup, mapOfInternalQueryToResult, http.StatusInternalServerError, "", "Failed during decode of derived kpi transformations."
	}

	for index, query := range internalKPIQueryGroup.Queries {
		internalKPIQueryGroup.Queries[index].Filters = append(query.Filters, baseQuery.Filters...)
		internalKPIQueryGroup.Queries[index].GroupBy = baseQuery.GroupBy
		internalKPIQueryGroup.Queries[index].From = baseQuery.From
		internalKPIQueryGroup.Queries[index].To = baseQuery.To
		internalKPIQueryGroup.Queries[index].Timezone = baseQuery.Timezone
		internalKPIQueryGroup.Queries[index].GroupByTimestamp = baseQuery.GroupByTimestamp
	}

	for _, query := range internalKPIQueryGroup.Queries {
		var result []model.QueryResult
		var hashCode string
		result, statusCode, hashCode, errMsg = store.ExecuteNonDerivedQuery(projectID, reqID, query, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

		if statusCode != http.StatusOK {
			mapOfInternalQueryToResult := make(map[string][]model.QueryResult)
			return internalKPIQueryGroup, mapOfInternalQueryToResult, statusCode, hashCode, errMsg
		}

		mapOfInternalQueryToResult[hashCode] = result

		queryResults = append(queryResults, result...)
	}

	baseQuery.GroupByTimestamp = ""
	derivedQueryHashCode, err := baseQuery.GetQueryCacheHashString()
	if err != nil {
		return internalKPIQueryGroup, mapOfInternalQueryToResult, http.StatusInternalServerError, "", "Failed while generating hashString for kpi."
	}

	for index := range internalKPIQueryGroup.Queries {
		internalKPIQueryGroup.Queries[index].GroupByTimestamp = ""
	}

	return internalKPIQueryGroup, mapOfInternalQueryToResult, http.StatusOK, derivedQueryHashCode, ""
}

func (store *MemSQL) ExecuteNonDerivedQuery(projectID int64, reqID string,
	query model.KPIQuery, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int, string, string) {

	result := make([]model.QueryResult, 0)
	statusCode := http.StatusOK
	hashCode := ""
	if query.Category == model.ProfileCategory {
		if query.GroupByTimestamp != "" {
			result, statusCode = store.ExecuteKPIQueryForProfiles(projectID, reqID,
				query, enableOptimisedFilterOnProfileQuery)
		} else {
			result = make([]model.QueryResult, 1)
		}
	} else if query.Category == model.ChannelCategory || query.Category == model.CustomChannelCategory {
		result, statusCode = store.ExecuteKPIQueryForChannels(projectID, reqID, query)
	} else if query.Category == model.EventCategory {
		result, statusCode = store.ExecuteKPIQueryForEvents(projectID, reqID, query, enableOptimisedFilterOnEventUserQuery)
	}

	query.GroupByTimestamp = ""
	hashCode, _ = query.GetQueryCacheHashString()

	return result, statusCode, hashCode, ""
}
