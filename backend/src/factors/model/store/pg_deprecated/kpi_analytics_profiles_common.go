package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) ExecuteKPIQueryForProfiles(projectID uint64, reqID string, kpiQuery model.KPIQuery) ([]model.QueryResult, int) {
	return pg.TransformToAndExecuteProfileAnalyticsQueries(projectID, kpiQuery, reqID)
}

// Check statusCode
func (pg *Postgres) TransformToAndExecuteProfileAnalyticsQueries(projectID uint64, kpiQuery model.KPIQuery, reqID string) ([]model.QueryResult, int) {
	var profileQueryGroup model.ProfileQueryGroup
	var queryResults []model.QueryResult
	queryResults = make([]model.QueryResult, len(kpiQuery.Metrics))
	profileQueryGroup = model.GetDirectDerivableProfileQueryFromKPI(kpiQuery)
	var waitGroup sync.WaitGroup
	count := 0
	actualRoutineLimit := U.MinInt(len(kpiQuery.Metrics), AllowedGoroutines)
	waitGroup.Add(actualRoutineLimit)
	for index, kpiMetric := range kpiQuery.Metrics {
		count++
		go pg.transformAndExecuteForSingleKPIMetricProfile(projectID, profileQueryGroup, kpiQuery, kpiMetric, &queryResults[index], &waitGroup)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(kpiQuery.Metrics)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	for index, result := range queryResults {
		if result.Headers == nil || result.Headers[0] == model.AliasError {
			log.WithField("kpiquery", kpiQuery).WithField("query result", queryResults).WithField("index", index).Warn("Failed in executing following KPI profile query.")
			return queryResults, http.StatusPartialContent
		}
	}
	return queryResults, http.StatusOK
}

func (pg *Postgres) transformAndExecuteForSingleKPIMetricProfile(projectID uint64, profileQueryGroup model.ProfileQueryGroup, kpiQuery model.KPIQuery,
	kpiMetric string, result *model.QueryResult, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	finalResult := model.QueryResult{}

	finalResult = pg.wrappedExecuteForResultProfile(projectID, profileQueryGroup, kpiQuery, kpiMetric)
	*result = finalResult
}

func (pg *Postgres) wrappedExecuteForResultProfile(projectID uint64, profileQueryGroup model.ProfileQueryGroup, kpiQuery model.KPIQuery,
	kpiMetric string) model.QueryResult {
	// Execute Profiles Query For Single KPI.
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	hasGroupByTimestamp := (kpiQuery.GroupByTimestamp != "")
	hasAnyGroupBys := (len(kpiQuery.GroupBy) > 0)
	finalResult := model.QueryResult{}
	isTimezoneEnabled := false
	if C.IsMultipleProjectTimezoneEnabled(projectID) {
		isTimezoneEnabled = true
	}

	var transformation model.CustomMetricTransformation
	customMetric, err, statusCode := pg.GetCustomMetricsByName(projectID, kpiMetric)
	if statusCode != http.StatusFound {
		finalResult.Headers = append(finalResult.Headers, model.AliasError)
		return finalResult
	}
	err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
	if err1 != nil {
		log.WithField("customMetric", customMetric).WithField("err", err).Warn("Failed in decoding custom Metric")
	}
	currentQueries := model.AddCustomMetricsTransformationsToProfileQuery(profileQueryGroup, kpiMetric, customMetric, transformation, kpiQuery)
	resultGroup, errCode := pg.RunProfilesGroupQuery(currentQueries, projectID)
	if errCode != http.StatusOK {
		// Log or not.
		finalResult.Headers = append(finalResult.Headers, model.AliasError)
		return finalResult
	}
	// Transformation of Profiles Results of Single KPI.
	if len(currentQueries) == 1 {
		results := model.TransformProfileResultsToKPIResults(resultGroup.Results, hasGroupByTimestamp, hasAnyGroupBys)
		finalResult = results[0]
		return finalResult
	} else {
		results := model.TransformProfileResultsToKPIResults(resultGroup.Results, hasGroupByTimestamp, hasAnyGroupBys)
		finalResult = model.HandlingProfileResultsByApplyingOperations(results, "Division", kpiQuery.Timezone, isTimezoneEnabled)
	}
	return finalResult
}
