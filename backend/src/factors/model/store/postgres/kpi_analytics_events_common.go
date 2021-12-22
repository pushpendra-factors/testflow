package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

// statusCode need to be clear on http.StatusOk or http.StatusAccepted or something else.
// TODO handle errors and kpiFunction statusCode.
func (pg *Postgres) ExecuteKPIQueryGroup(projectID uint64, reqID string, kpiQueryGroup model.KPIQueryGroup) ([]model.QueryResult, int) {
	var queryResults []model.QueryResult
	finalStatusCode := http.StatusOK
	for _, query := range kpiQueryGroup.Queries {
		query.Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		query.GroupBy = kpiQueryGroup.GlobalGroupBy
		kpiFunction := pg.kpiQueryFunctionDeciderBasedOnCategory(query.Category)
		result, statusCode := kpiFunction(projectID, reqID, query)
		if statusCode != http.StatusOK {
			finalStatusCode = statusCode
		}
		queryResults = append(queryResults, result...)
	}
	return queryResults, finalStatusCode
}

func (pg *Postgres) kpiQueryFunctionDeciderBasedOnCategory(category string) func(uint64, string, model.KPIQuery) ([]model.QueryResult, int) {
	var result func(uint64, string, model.KPIQuery) ([]model.QueryResult, int)
	if category == model.ChannelCategory {
		result = pg.ExecuteKPIQueryForChannels
	} else {
		result = pg.ExecuteKPIQueryForEvents
	}
	return result
}

// We convert kpi Query to eventQueries by applying transformation.
func (pg *Postgres) ExecuteKPIQueryForEvents(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	queryResults := make([]model.QueryResult, len(kpiQuery.Metrics))
	isValid := model.ValidateKPIQuery(kpiQuery)
	if !isValid {
		return queryResults, http.StatusBadRequest
	}
	return pg.transformToAndExecuteEventAnalyticsQueries(projectID, kpiQuery)
}

func (pg *Postgres) transformToAndExecuteEventAnalyticsQueries(projectID uint64, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
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
		go pg.ExecuteForSingleKPIMetric(projectID, query, kpiQuery, kpiMetric, &queryResults[index], &waitGroup)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(kpiQuery.Metrics)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	for index, result := range queryResults {
		if result.Headers == nil || result.Headers[0] == model.AliasError {
			log.WithField("kpiQuery", kpiQuery).WithField("queryResults", queryResults).WithField("index", index).Error("Failed in executing following KPI Query.")
			return queryResults, http.StatusPartialContent
		}
	}
	return queryResults, http.StatusOK
}

// Each KPI Metric is mapped to array of operations containing metrics and aggregates, filters.
func (pg *Postgres) ExecuteForSingleKPIMetric(projectID uint64, query model.Query, kpiQuery model.KPIQuery,
	kpiMetric string, result *model.QueryResult, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	finalResult := model.QueryResult{}

	finalResult = pg.wrappedExecuteForResult(projectID, query, kpiQuery, kpiMetric)
	*result = finalResult
}

func (pg *Postgres) wrappedExecuteForResult(projectID uint64, query model.Query, kpiQuery model.KPIQuery, kpiMetric string) model.QueryResult {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	transformations := model.TransformationOfKPIMetricsToEventAnalyticsQuery[kpiQuery.DisplayCategory][kpiMetric]
	currentQuery := model.BuildFiltersAndGroupByBasedOnKPIQuery(query, kpiQuery, kpiMetric)
	currentQueries := model.SplitKPIQueryToInternalKPIQueries(currentQuery, kpiQuery, kpiMetric, transformations)
	finalResult := pg.executeForResults(projectID, currentQueries, kpiQuery, transformations)
	return finalResult
}

func (pg *Postgres) executeForResults(projectID uint64, queries []model.Query, kpiQuery model.KPIQuery, transformations []model.TransformQueryi) model.QueryResult {
	results := make([]*model.QueryResult, len(queries))
	hasGroupByTimestamp := false
	displayCategory := kpiQuery.DisplayCategory
	var finalResult model.QueryResult
	isTimezoneEnabled := false
	if C.IsMultipleProjectTimezoneEnabled(projectID) {
		isTimezoneEnabled = true
	}
	if kpiQuery.GroupByTimestamp != "" {
		hasGroupByTimestamp = true
	}
	if len(queries) == 1 {
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results[0], _, _ = pg.RunInsightsQuery(projectID, queries[0])
		if results[0].Headers == nil || results[0].Headers[0] == model.AliasError {
			finalResult = model.QueryResult{}
			finalResult.Headers = results[0].Headers
			return finalResult
		}
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		finalResult = *results[0]
	} else {
		for i, query := range queries {
			results[i], _, _ = pg.RunInsightsQuery(projectID, query)
			if results[i].Headers == nil || results[i].Headers[0] == model.AliasError {
				finalResult = model.QueryResult{}
				finalResult.Headers = results[i].Headers
				return finalResult
			}
		}
		hasAnyGroupBy := len(queries[0].GroupByProperties) != 0
		results = model.TransformResultsToKPIResults(results, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		finalResult = model.HandlingEventResultsByApplyingOperations(results, transformations, kpiQuery.Timezone, isTimezoneEnabled)
	}
	return finalResult
}
