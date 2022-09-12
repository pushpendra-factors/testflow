package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"reflect"
	"strings"

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
		var result []model.QueryResult
		var statusCode int
		var errMsg string
		var hashCode string

		if query.QueryType == model.KpiDerivedQueryType {
			result, statusCode, errMsg = store.ExecuteDerivedKPIQuery(projectID, reqID, query, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			if statusCode != http.StatusOK {
				finalStatusCode = statusCode
				log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).WithField("queryResults", queryResults).Error(errMsg)
				break
			}
		} else {
			result, statusCode, hashCode, errMsg = store.ExecuteNonDerivedQuery(projectID, reqID, query, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

			if statusCode != http.StatusOK {
				finalStatusCode = statusCode
				log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).WithField("queryResults", queryResults).Error(errMsg)
				break
			} else {
				if hashCode != "" {
					hashMapOfQueryToResult[hashCode] = result
				}
			}
		}
		queryResults = append(queryResults, result...)
	}
	if finalStatusCode != http.StatusOK {
		log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("queryResults", queryResults).Error("Failed in executing following KPI Query with status Not Ok.")
		return []model.QueryResult{{}, {}}, finalStatusCode
	}

	for index, query := range kpiQueryGroup.Queries {
		if query.Category == model.ProfileCategory && query.GroupByTimestamp == "" {
			hashCode, err := query.GetQueryCacheHashString()
			if err != nil {
				log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).Error("Failed while generating hashString for kpi 2.")
				return []model.QueryResult{{}, {}}, http.StatusBadRequest
			}
			if resultsWithGbt, exists := hashMapOfQueryToResult[hashCode]; exists {
				queryResults[index] = model.GetNonGBTResultsFromGBTResults(resultsWithGbt, query)[0]
			} else {
				log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("queryResults", queryResults).Error("Query group doesnt contain all the gbt and non gbt pair of query.")
				return []model.QueryResult{{}, {}}, http.StatusBadRequest
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

func (store *MemSQL) ExecuteNonDerivedQuery(projectID int64, reqID string,
	query model.KPIQuery, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int, string, string) {

	result := make([]model.QueryResult, 0)
	statusCode := http.StatusOK
	hashCode := ""
	if query.Category == model.ProfileCategory {
		if query.GroupByTimestamp != "" {
			var err error
			result, statusCode = store.ExecuteKPIQueryForProfiles(projectID, reqID,
				query, enableOptimisedFilterOnProfileQuery)

			query.GroupByTimestamp = ""
			hashCode, err = query.GetQueryCacheHashString()
			if err != nil {
				return result, http.StatusInternalServerError, "", "Failed while generating hashString for kpi."
			}
		} else {
			result = make([]model.QueryResult, 1)
		}
	} else if query.Category == model.ChannelCategory || query.Category == model.CustomChannelCategory {
		result, statusCode = store.ExecuteKPIQueryForChannels(projectID, reqID, query)
	} else if query.Category == model.EventCategory {
		result, statusCode = store.ExecuteKPIQueryForEvents(projectID, reqID, query, enableOptimisedFilterOnEventUserQuery)
	}
	return result, statusCode, hashCode, ""
}

func (store *MemSQL) ExecuteDerivedKPIQuery(projectID int64, reqID string, baseQuery model.KPIQuery,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int, string) {
	queryResults := make([]model.QueryResult, 0)
	finalStatusCode := http.StatusOK
	hashMapOfQueryToResult := make(map[string][]model.QueryResult)
	mapOfFormulaVariableToQueryResult := make(map[string]model.QueryResult)

	derivedMetric, errMsg, statusCode := store.GetDerivedMetricsByName(projectID, baseQuery.Metrics[0])
	if statusCode != http.StatusFound {
		return queryResults, statusCode, errMsg
	}
	kpiQueryGroup := model.KPIQueryGroup{}
	err := U.DecodePostgresJsonbToStructType(derivedMetric.Transformations, &kpiQueryGroup)
	if err != nil {
		return queryResults, http.StatusInternalServerError, "Failed during decode of derived kpi transformations."
	}
	for index, query := range kpiQueryGroup.Queries {
		kpiQueryGroup.Queries[index].Filters = append(query.Filters, baseQuery.Filters...)
		kpiQueryGroup.Queries[index].GroupBy = baseQuery.GroupBy
		kpiQueryGroup.Queries[index].From = baseQuery.From
		kpiQueryGroup.Queries[index].To = baseQuery.To
		kpiQueryGroup.Queries[index].Timezone = baseQuery.Timezone
		kpiQueryGroup.Queries[index].GroupByTimestamp = baseQuery.GroupByTimestamp
	}

	for _, query := range kpiQueryGroup.Queries {
		var result []model.QueryResult
		var statusCode int
		var hashCode string
		result, statusCode, hashCode, errMsg = store.ExecuteNonDerivedQuery(projectID, reqID, query, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

		if statusCode != http.StatusOK {
			finalStatusCode = statusCode
			break
		} else {
			if hashCode != "" {
				hashMapOfQueryToResult[hashCode] = result
			}
		}

		mapOfFormulaVariableToQueryResult[query.Name] = result[0]
	}

	if finalStatusCode != http.StatusOK {
		return make([]model.QueryResult, 0), finalStatusCode, errMsg
	} else {
		result := EvaluateKPIExpressionWithBraces(mapOfFormulaVariableToQueryResult, kpiQueryGroup.Queries[0].Timezone, kpiQueryGroup.Formula)
		queryResults = append(queryResults, result)
		return queryResults, http.StatusOK, ""
	}
}

func EvaluateKPIExpressionWithBraces(mapOfFormulaVariableToQueryResult map[string]model.QueryResult, timezone string, formula string) model.QueryResult {
	valueStack := make([]model.QueryResult, 0)
	operatorStack := make([]string, 0)

	for _, currentVariable := range formula {
		currentFormulaVariable := string(currentVariable)
		if currentFormulaVariable == "(" {
			operatorStack = append(operatorStack, currentFormulaVariable)
		} else if strings.Contains(U.Alpha, strings.ToLower(currentFormulaVariable)) {
			valueStack = append(valueStack, mapOfFormulaVariableToQueryResult[currentFormulaVariable])
		} else if currentFormulaVariable == ")" {
			for len(operatorStack) != 0 && operatorStack[len(operatorStack)-1] != "(" {
				v1 := valueStack[len(valueStack)-1]
				valueStack = valueStack[:len(valueStack)-1]
				v2 := valueStack[len(valueStack)-1]
				valueStack = valueStack[:len(valueStack)-1]
				results := make([]*model.QueryResult, 0)
				results = append(results, &v2)
				results = append(results, &v1)
				op := operatorStack[len(operatorStack)-1]
				ops := make([]string, 0)
				ops = append(ops, op)
				operatorStack = operatorStack[:len(operatorStack)-1]
				valueStack = append(valueStack, model.HandlingEventResultsByApplyingOperations(results, ops, timezone, true)) // apply operations and return result
			}
			if len(operatorStack) != 0 {
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
		} else {
			for len(operatorStack) != 0 && U.Precedence(operatorStack[len(operatorStack)-1]) >= U.Precedence(currentFormulaVariable) {
				v1 := valueStack[len(valueStack)-1]
				valueStack = valueStack[:len(valueStack)-1]
				v2 := valueStack[len(valueStack)-1]
				results := make([]*model.QueryResult, 0)
				results = append(results, &v2)
				results = append(results, &v1)
				valueStack = valueStack[:len(valueStack)-1]
				op := operatorStack[len(operatorStack)-1]
				ops := make([]string, 0)
				ops = append(ops, op)
				operatorStack = operatorStack[:len(operatorStack)-1]
				valueStack = append(valueStack, model.HandlingEventResultsByApplyingOperations(results, ops, timezone, true)) // apply operations and return result
			}
			operatorStack = append(operatorStack, currentFormulaVariable)
		}
	}

	for len(operatorStack) != 0 {
		v1 := valueStack[len(valueStack)-1]
		valueStack = valueStack[:len(valueStack)-1]
		v2 := valueStack[len(valueStack)-1]
		results := make([]*model.QueryResult, 0)
		results = append(results, &v2)
		results = append(results, &v1)
		valueStack = valueStack[:len(valueStack)-1]
		op := operatorStack[len(operatorStack)-1]
		ops := make([]string, 0)
		ops = append(ops, op)
		operatorStack = operatorStack[:len(operatorStack)-1]
		valueStack = append(valueStack, model.HandlingEventResultsByApplyingOperations(results, ops, timezone, true))
	}
	return valueStack[len(valueStack)-1]
}
