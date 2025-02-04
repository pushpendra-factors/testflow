package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// NOTE: We are not supporting group analytics as a part of event analysis.
// Group analysis of events is not yet added to KPI.
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
	if kpiQuery.QueryType == model.KpiCustomQueryType {
		for _, kpiMetric := range kpiQuery.Metrics {
			if _, _, statusCode := store.GetEventBasedCustomMetricByProjectIdName(projectID, kpiMetric); statusCode != http.StatusFound {
				return false
			}
		}
		return true
	} else {
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

	if kpiQuery.QueryType == model.KpiCustomQueryType {
		var transformation model.CustomMetricTransformation
		customMetric, err, statusCode := store.GetEventBasedCustomMetricByProjectIdName(projectID, kpiMetric)
		if statusCode != http.StatusFound {
			return model.QueryResult{Headers: []string{model.AliasError}}, statusCode
		}
		err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
		if err1 != nil {
			log.WithField("customMetric", customMetric).WithField("err", err).Warn("Failed in decoding custom Metric")
		}
		currentQueries, operations := model.ConvertCustomKPIQueryToInternalEventQueriesAndTransformationOperations(projectID, query, kpiQuery, kpiMetric, transformation, enableFilterOpt)
		return store.executeForResults(projectID, currentQueries, kpiQuery, operations, enableFilterOpt)
	} else {
		currentQueries, operations := model.ConvertStaticKPIQueryToInternalEventQueriesAndTransformationOperations(projectID, query, kpiQuery, kpiMetric, enableFilterOpt)
		return store.executeForResults(projectID, currentQueries, kpiQuery, operations, enableFilterOpt)
	}
}

func (store *MemSQL) executeForResults(projectID int64, queries []model.Query, kpiQuery model.KPIQuery,
	operations []string, enableFilterOpt bool) (model.QueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"queries":    queries,
		"kpi_query":  kpiQuery,
		"operations": operations,
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
		finalResult = model.HandlingEventResultsByApplyingOperations(results, operations, kpiQuery.Timezone)
	}
	return finalResult, finalStatusCode
}
