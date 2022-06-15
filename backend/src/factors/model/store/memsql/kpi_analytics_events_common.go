package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// We convert kpi Query to eventQueries by applying transformation.
func (store *MemSQL) ExecuteKPIQueryForEvents(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"req_id":     reqID,
		"kpi_query":  kpiQuery,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	queryResults := make([]model.QueryResult, len(kpiQuery.Metrics))
	isValid := store.ValidateKPIQuery(projectID, kpiQuery)
	if !isValid {
		return queryResults, http.StatusBadRequest
	}
	return store.transformToAndExecuteEventAnalyticsQueries(projectID, kpiQuery)
}

// query is being mutated. So, waitGroup can side effects.
func (store *MemSQL) transformToAndExecuteEventAnalyticsQueries(projectID uint64, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"kpi_query":  kpiQuery,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	for index, result := range queryResults {
		if result.Headers == nil || result.Headers[0] == model.AliasError {
			log.WithField("kpiQuery", kpiQuery).WithField("queryResults", queryResults).WithField("index", index).Error("Failed in executing following KPI Query.")
			return queryResults, http.StatusBadRequest
		}
	}
	return queryResults, http.StatusOK
}

func (store *MemSQL) ValidateKPIQuery(projectID uint64, kpiQuery model.KPIQuery) bool {
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
func (store *MemSQL) ExecuteForSingleKPIMetric(projectID uint64, query model.Query, kpiQuery model.KPIQuery,
	kpiMetric string, result *model.QueryResult, waitGroup *sync.WaitGroup) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"kpi_query":  kpiQuery,
		"kpi_metric": kpiMetric,
		"result":     result,
		"wait_group": waitGroup,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()
	finalResult := model.QueryResult{}

	finalResult = store.wrappedExecuteForResult(projectID, query, kpiQuery, kpiMetric)
	*result = finalResult
}

func (store *MemSQL) wrappedExecuteForResult(projectID uint64, query model.Query, kpiQuery model.KPIQuery,
	kpiMetric string) model.QueryResult {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"kpi_query":  kpiQuery,
		"kpi_metric": kpiMetric,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	transformations := model.TransformationOfKPIMetricsToEventAnalyticsQuery[kpiQuery.DisplayCategory][kpiMetric]
	currentQuery := model.BuildFiltersAndGroupByBasedOnKPIQuery(query, kpiQuery, kpiMetric)
	currentQueries := model.SplitKPIQueryToInternalKPIQueries(currentQuery, kpiQuery, kpiMetric, transformations)
	finalResult := store.executeForResults(projectID, currentQueries, kpiQuery, transformations)
	return finalResult
}

func (store *MemSQL) executeForResults(projectID uint64, queries []model.Query, kpiQuery model.KPIQuery, transformations []model.TransformQueryi) model.QueryResult {
	logFields := log.Fields{
		"project_id":     projectID,
		"queries":        queries,
		"kpi_query":      kpiQuery,
		"transformation": transformations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	results := make([]*model.QueryResult, len(queries))
	hasGroupByTimestamp := false
	displayCategory := kpiQuery.DisplayCategory
	var finalResult *model.QueryResult
	isTimezoneEnabled := false
	if C.IsMultipleProjectTimezoneEnabled(projectID) {
		isTimezoneEnabled = true
	}
	if kpiQuery.GroupByTimestamp != "" {
		hasGroupByTimestamp = true
	}
	var statusCode int
	var errMsg string
	if len(queries) == 1 {
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results[0], statusCode, errMsg = store.RunInsightsQuery(projectID, queries[0])
		if statusCode != http.StatusOK {
			finalResult = buildErrorResult(errMsg)
			return *finalResult
		}
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory, kpiQuery.Timezone)
		finalResult = results[0]
	} else {
		for i, query := range queries {
			results[i], statusCode, errMsg = store.RunInsightsQuery(projectID, query)
			if statusCode != http.StatusOK {
				finalResult = buildErrorResult(errMsg)
				return *finalResult
			}
		}
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory, kpiQuery.Timezone)
		*finalResult = model.HandlingEventResultsByApplyingOperations(results, transformations, kpiQuery.Timezone, isTimezoneEnabled)
	}
	return *finalResult
}
