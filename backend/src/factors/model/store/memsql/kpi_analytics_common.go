package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// statusCode need to be clear on http.StatusOk or http.StatusAccepted or something else.
// Below function relies on fact that each query has only one metric.
func (store *MemSQL) ExecuteKPIQueryGroup(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int) {

	var queryResults []model.QueryResult
	finalStatusCode := http.StatusOK
	isTimezoneEnabled := false
	kpiTimezoneString := string(kpiQueryGroup.GetTimeZone())
	hashMapOfQueryToResult := make(map[string][]model.QueryResult)
	if C.IsMultipleProjectTimezoneEnabled(projectID) {
		isTimezoneEnabled = true
	}
	for index, query := range kpiQueryGroup.Queries {
		kpiQueryGroup.Queries[index].Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		kpiQueryGroup.Queries[index].GroupBy = kpiQueryGroup.GlobalGroupBy
	}
	for _, query := range kpiQueryGroup.Queries {
		if query.Category == model.ProfileCategory {
			if query.GroupByTimestamp != "" {
				result, statusCode := store.ExecuteKPIQueryForProfiles(projectID, reqID,
					query, enableOptimisedFilterOnProfileQuery)

				if statusCode != http.StatusOK {
					finalStatusCode = statusCode
				}
				queryResults = append(queryResults, result...)

				query.GroupByTimestamp = ""
				hashCode, err := query.GetQueryCacheHashString()
				if err != nil {
					log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).
						WithField("query", query).Error("Failed while generating hashString for kpi.")
				}
				hashMapOfQueryToResult[hashCode] = result
			} else {
				result := make([]model.QueryResult, 1)
				queryResults = append(queryResults, result...)
			}
		} else {
			var result []model.QueryResult
			var statusCode int
			if query.Category == model.ChannelCategory {
				result, statusCode = store.ExecuteKPIQueryForChannels(projectID, reqID, query)
			} else if query.Category == model.EventCategory {
				result, statusCode = store.ExecuteKPIQueryForEvents(projectID, reqID, query, enableOptimisedFilterOnEventUserQuery)
			}
			if statusCode != http.StatusOK {
				finalStatusCode = statusCode
			}

			queryResults = append(queryResults, result...)
		}
	}
	if finalStatusCode != http.StatusOK {
		log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("queryResults", queryResults).Error("Failed in executing following KPI Query with status Not Ok.")
		return []model.QueryResult{model.QueryResult{}, model.QueryResult{}}, finalStatusCode
	}

	for index, query := range kpiQueryGroup.Queries {
		if query.Category == model.ProfileCategory && query.GroupByTimestamp == "" {
			hashCode, err := query.GetQueryCacheHashString()
			if err != nil {
				log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).Error("Failed while generating hashString for kpi 2.")
				return []model.QueryResult{model.QueryResult{}, model.QueryResult{}}, http.StatusBadRequest
			}
			if resultsWithGbt, exists := hashMapOfQueryToResult[hashCode]; exists {
				queryResults[index] = model.GetNonGBTResultsFromGBTResults(resultsWithGbt, query)[0]
			} else {
				log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("queryResults", queryResults).Error("Query group doesnt contain all the gbt and non gbt pair of query.")
				return []model.QueryResult{model.QueryResult{}, model.QueryResult{}}, http.StatusBadRequest
			}
		}
	}

	gbtRelatedQueryResults, nonGbtRelatedQueryResults, gbtRelatedQueries, nonGbtRelatedQueries := model.SplitQueryResultsIntoGBTAndNonGBT(queryResults, kpiQueryGroup, finalStatusCode)
	finalQueryResult := make([]model.QueryResult, 0)
	gbtRelatedMergedResults := model.MergeQueryResults(gbtRelatedQueryResults, gbtRelatedQueries, kpiTimezoneString, finalStatusCode, isTimezoneEnabled)
	nonGbtRelatedMergedResults := model.MergeQueryResults(nonGbtRelatedQueryResults, nonGbtRelatedQueries, kpiTimezoneString, finalStatusCode, isTimezoneEnabled)
	if (!reflect.DeepEqual(model.QueryResult{}, gbtRelatedMergedResults)) {
		finalQueryResult = append(finalQueryResult, gbtRelatedMergedResults)
	}
	if (!reflect.DeepEqual(model.QueryResult{}, nonGbtRelatedMergedResults)) {
		finalQueryResult = append(finalQueryResult, nonGbtRelatedMergedResults)
	}
	return finalQueryResult, finalStatusCode
}
