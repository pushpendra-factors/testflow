package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) GetKPIConfigFromStandardUserProperties(projectID int64) []map[string]string {
	var resultantKPIConfigProperties []map[string]string
	var tempKPIConfigProperty map[string]string
	propertiesToDisplayNames := store.GetStandardUserPropertiesBasedOnIntegration(projectID)
	for userProperty, userDisplayPropertyName := range propertiesToDisplayNames {

		tempKPIConfigProperty = map[string]string{
			"name":         userProperty,
			"display_name": userDisplayPropertyName,
			"data_type":    U.GetPropertyTypeByName(userProperty),
			"entity":       model.UserEntity,
		}
		resultantKPIConfigProperties = append(resultantKPIConfigProperties, tempKPIConfigProperty)
	}
	if resultantKPIConfigProperties == nil {
		return make([]map[string]string, 0)
	}
	return resultantKPIConfigProperties
}

func (store *MemSQL) ExecuteKPIQueryForProfiles(projectID int64, reqID string,
	kpiQuery model.KPIQuery, enableOptimisedFilter bool) ([]model.QueryResult, int) {
	return store.TransformToAndExecuteProfileAnalyticsQueries(projectID, kpiQuery, reqID, enableOptimisedFilter)
}

func (store *MemSQL) TransformToAndExecuteProfileAnalyticsQueries(projectID int64, kpiQuery model.KPIQuery,
	reqID string, enableOptimisedFilter bool) ([]model.QueryResult, int) {
	var profileQueryGroup model.ProfileQueryGroup
	var statusCode, finalStatusCode int
	var queryResults []model.QueryResult
	queryResults = make([]model.QueryResult, len(kpiQuery.Metrics))
	profileQueryGroup = model.GetDirectDerivableProfileQueryFromKPI(kpiQuery)

	for index, kpiMetric := range kpiQuery.Metrics {
		queryResults[index], statusCode = store.ExecuteForSingleKPIMetricProfile(projectID, profileQueryGroup, kpiQuery, kpiMetric, enableOptimisedFilter)
		finalStatusCode = statusCode
		if statusCode != http.StatusOK {
			queryResults = make([]model.QueryResult, len(kpiQuery.Metrics))
			return queryResults, finalStatusCode
		}
	}
	return queryResults, finalStatusCode
}

// TODO Later - Generalising the transformation of external computation to internal computations.
// Eg - representing avg in terms of count(property)/count(*) with division operator.
func (store *MemSQL) ExecuteForSingleKPIMetricProfile(projectID int64, profileQueryGroup model.ProfileQueryGroup,
	kpiQuery model.KPIQuery, kpiMetric string, enableOptimisedFilter bool) (model.QueryResult, int) {
	// Execute Profiles Query For Single KPI.
	hasGroupByTimestamp := (kpiQuery.GroupByTimestamp != "")
	hasAnyGroupBys := (len(kpiQuery.GroupBy) > 0)
	finalResult := model.QueryResult{}

	var transformation model.CustomMetricTransformation
	customMetric, err, statusCode := store.GetProfileCustomMetricByProjectIdName(projectID, kpiMetric)
	if statusCode != http.StatusFound {
		finalResult.Headers = append(finalResult.Headers, model.AliasError)
		return finalResult, statusCode
	}
	err1 := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &transformation)
	if err1 != nil {
		log.WithField("customMetric", customMetric).WithField("err", err).Warn("Failed in decoding custom Metric")
	}
	currentQueries := model.AddCustomMetricsTransformationsToProfileQuery(profileQueryGroup, kpiMetric, customMetric, transformation, kpiQuery)
	resultGroup, statusCode2 := store.RunProfilesGroupQuery(currentQueries, projectID, enableOptimisedFilter)
	if statusCode2 != http.StatusOK {
		// Log or not.
		finalResult.Headers = append(finalResult.Headers, model.AliasError)
		return finalResult, statusCode2
	}
	// Transformation of Profiles Results of Single KPI.
	if len(currentQueries) == 1 {
		results := model.TransformProfileResultsToKPIResults(resultGroup.Results, hasGroupByTimestamp, hasAnyGroupBys)
		finalResult = results[0]
		return finalResult, http.StatusOK
	} else {
		results := model.TransformProfileResultsToKPIResults(resultGroup.Results, hasGroupByTimestamp, hasAnyGroupBys)
		finalResult = model.HandlingProfileResultsByApplyingOperations(results, currentQueries, kpiQuery.Timezone)
	}
	return finalResult, http.StatusOK
}
