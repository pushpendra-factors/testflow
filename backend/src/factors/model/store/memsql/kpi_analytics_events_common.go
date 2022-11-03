package memsql

import (
	"factors/model/model"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// We convert kpi Query to eventQueries by applying transformation.
func (store *MemSQL) ExecuteKPIQueryForEvents(projectID int64, reqID string,
	kpiQuery model.KPIQuery, enableFilterOpt bool) ([]model.QueryResult, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
		"kpi_query":  kpiQuery,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	queryResults := make([]model.QueryResult, len(kpiQuery.Metrics))
	isValid := store.ValidateKPIQuery(projectID, kpiQuery)
	if !isValid {
		return queryResults, http.StatusPartialContent
	}
	return store.transformToAndExecuteEventAnalyticsQueries(projectID, kpiQuery, enableFilterOpt)
}

func (store *MemSQL) transformToAndExecuteEventAnalyticsQueries(projectID int64,
	kpiQuery model.KPIQuery, enableFilterOpt bool) ([]model.QueryResult, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"kpi_query":  kpiQuery,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var statusCode, finalStatusCode int
	var query model.Query
	var queryResults []model.QueryResult
	queryResults = make([]model.QueryResult, len(kpiQuery.Metrics))
	query = model.GetDirectDerviableQueryPropsFromKPI(kpiQuery)

	for index, kpiMetric := range kpiQuery.Metrics {
		queryResults[index], statusCode = store.ExecuteEventsForSingleKPIMetric(projectID, query, kpiQuery, kpiMetric, enableFilterOpt)
		finalStatusCode = statusCode
		if statusCode != http.StatusOK {
			queryResults = make([]model.QueryResult, len(kpiQuery.Metrics))
			return queryResults, finalStatusCode
		}
	}

	return queryResults, finalStatusCode
}

func (store *MemSQL) ValidateKPIQuery(projectID int64, kpiQuery model.KPIQuery) bool {
	if kpiQuery.DisplayCategory == model.WebsiteSessionDisplayCategory {
		return store.ValidateKPISessions(projectID, kpiQuery)
	} else if kpiQuery.DisplayCategory == model.PageViewsDisplayCategory {
		return model.ValidateKPIPageView(kpiQuery)
	} else if kpiQuery.DisplayCategory == model.FormSubmissionsDisplayCategory {
		return model.ValidateKPIFormSubmissions(kpiQuery)
	} else {
		return false
	}
}

// Each KPI Metric is mapped to array of operations containing metrics and aggregates, filters.
func (store *MemSQL) ExecuteEventsForSingleKPIMetric(projectID int64, query model.Query, kpiQuery model.KPIQuery,
	kpiMetric string, enableFilterOpt bool) (model.QueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"kpi_query":  kpiQuery,
		"kpi_metric": kpiMetric,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	currentQueries, transformations := model.ConvertKPIQueryToInternalEventQueriesAndTransformations(projectID, query, kpiQuery, kpiMetric, enableFilterOpt)
	return store.executeForResults(projectID, currentQueries, kpiQuery, transformations, enableFilterOpt)
}

func (store *MemSQL) executeForResults(projectID int64, queries []model.Query, kpiQuery model.KPIQuery,
	transformations []model.TransformQueryi, enableFilterOpt bool) (model.QueryResult, int) {
	logFields := log.Fields{
		"project_id":     projectID,
		"queries":        queries,
		"kpi_query":      kpiQuery,
		"transformation": transformations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	results := make([]*model.QueryResult, len(queries))
	hasGroupByTimestamp := (kpiQuery.GroupByTimestamp != "")
	displayCategory := kpiQuery.DisplayCategory
	var statusCode, finalStatusCode int
	var finalResult model.QueryResult

	if len(queries) == 1 {
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results[0], statusCode, _ = store.RunInsightsQuery(projectID, queries[0], enableFilterOpt)
		finalStatusCode = statusCode
		if results[0].Headers == nil || results[0].Headers[0] == model.AliasError || statusCode != http.StatusOK {
			finalResult = model.QueryResult{}
			finalResult.Headers = make([]string, 0)
			return finalResult, finalStatusCode
		}
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory, kpiQuery.Timezone)
		finalResult = *results[0]
	} else {
		for i, query := range queries {
			results[i], statusCode, _ = store.RunInsightsQuery(projectID, query, enableFilterOpt)
			finalStatusCode = statusCode
			if results[i].Headers == nil || results[i].Headers[0] == model.AliasError || statusCode != http.StatusOK {
				finalResult = model.QueryResult{}
				finalResult.Headers = make([]string, 0)
				return finalResult, finalStatusCode
			}
		}
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory, kpiQuery.Timezone)
		operations := make([]string, 0)
		for _, transformation := range transformations {
			operations = append(operations, transformation.Metrics.Operator)
		}
		finalResult = model.HandlingEventResultsByApplyingOperations(results, operations, kpiQuery.Timezone)
	}
	return finalResult, finalStatusCode
}
