package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"sync"
)

// statusCode need to be clear on http.StatusOk or http.StatusAccepted or something else.
// TODO handle errors and kpiFunction statusCode.
func (store *MemSQL) ExecuteKPIQueryGroup(projectID uint64, reqID string, kpiQueryGroup model.KPIQueryGroup) ([]model.QueryResult, int) {
	var queryResults []model.QueryResult
	finalStatusCode := http.StatusOK
	for _, query := range kpiQueryGroup.Queries {
		query.Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		query.GroupBy = kpiQueryGroup.GlobalGroupBy
		kpiFunction := store.kpiQueryFunctionDeciderBasedOnCategory(query.Category)
		result, statusCode := kpiFunction(projectID, reqID, query)
		if statusCode != http.StatusOK {
			finalStatusCode = statusCode
		}
		queryResults = append(queryResults, result...)
	}
	return queryResults, finalStatusCode
}

func (store *MemSQL) kpiQueryFunctionDeciderBasedOnCategory(category string) func(uint64, string, model.KPIQuery) ([]model.QueryResult, int) {
	var result func(uint64, string, model.KPIQuery) ([]model.QueryResult, int)
	if category == model.ChannelCategory {
		result = store.ExecuteKPIQueryForChannels
	} else {
		result = store.ExecuteKPIQueryForEvents
	}
	return result
}

// We convert kpi Query to eventQueries by applying transformation.
func (store *MemSQL) ExecuteKPIQueryForEvents(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	queryResults := make([]model.QueryResult, len(kpiQuery.Metrics))
	isValid := model.ValidateKPIQuery(kpiQuery)
	if !isValid {
		return queryResults, http.StatusPartialContent
	}
	return store.transformToAndExecuteEventAnalyticsQueries(projectID, kpiQuery)
}

func (store *MemSQL) transformToAndExecuteEventAnalyticsQueries(projectID uint64, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	var query model.Query
	var queryResults []model.QueryResult
	queryResults = make([]model.QueryResult, len(kpiQuery.Metrics))
	query = model.GetDirectDerviableQueryPropsFromKPI(kpiQuery)

	var waitGroup sync.WaitGroup
	count := 0
	actualRoutineLimit := U.MinInt(len(kpiQuery.Metrics), AllowedGoroutines)
	waitGroup.Add(actualRoutineLimit)
	for index, kpiMetric := range kpiQuery.Metrics {
		count++
		go store.ExecuteForSingleKPIMetric(projectID, query, kpiQuery, kpiMetric, &queryResults[index], &waitGroup)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(kpiQuery.Metrics)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	for _, result := range queryResults {
		if result.Headers == nil || result.Headers[0] == model.AliasError {
			return queryResults, http.StatusPartialContent
		}
	}
	return queryResults, http.StatusOK
}

// Each KPI Metric is mapped to array of operations containing metrics and aggregates, filters.
func (store *MemSQL) ExecuteForSingleKPIMetric(projectID uint64, query model.Query, kpiQuery model.KPIQuery,
	kpiMetric string, result *model.QueryResult, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	finalResult := model.QueryResult{}

	finalResult = store.wrappedExecuteForResult(projectID, query, kpiQuery, kpiMetric)
	*result = finalResult
}

func (store *MemSQL) wrappedExecuteForResult(projectID uint64, query model.Query, kpiQuery model.KPIQuery,
	kpiMetric string) model.QueryResult {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	transformations := model.TransformationOfKPIMetricsToEventAnalyticsQuery[kpiQuery.DisplayCategory][kpiMetric]
	currentQuery := model.BuildFiltersAndGroupByBasedOnKPIQuery(query, kpiQuery, kpiMetric)
	currentQueries := model.SplitKPIQueryToInternalKPIQueries(currentQuery, kpiQuery, kpiMetric, transformations)
	finalResult := store.executeForResults(projectID, currentQueries, kpiQuery, transformations)
	return finalResult
}

func (store *MemSQL) executeForResults(projectID uint64, queries []model.Query, kpiQuery model.KPIQuery, transformations []model.TransformQueryi) model.QueryResult {
	results := make([]*model.QueryResult, len(queries))
	hasGroupByTimestamp := false
	displayCategory := kpiQuery.DisplayCategory
	var finalResult model.QueryResult
	if kpiQuery.GroupByTimestamp != "" {
		hasGroupByTimestamp = true
	}
	if len(queries) == 1 {
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results[0], _, _ = store.RunInsightsQuery(projectID, queries[0])
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		finalResult = *results[0]
	} else {
		for i, query := range queries {
			results[i], _, _ = store.RunInsightsQuery(projectID, query)
		}
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		finalResult = model.HandlingEventResultsByApplyingOperations(results, transformations)
	}
	return finalResult
}
