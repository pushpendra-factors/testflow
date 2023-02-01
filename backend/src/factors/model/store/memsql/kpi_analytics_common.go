package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"net/http"
	"reflect"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// statusCode need to be clear on http.StatusOk or http.StatusAccepted or something else.
// Below function relies on fact that each query has only one metric.
// Note: All of the hash functions use the query without GBT to form keys.
func (store *MemSQL) ExecuteKPIQueryGroup(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      kpiQueryGroup,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	kpiTimezoneString := string(kpiQueryGroup.GetTimeZone())
	var finalResultantResults []model.QueryResult

	// Get all property mappings from filters and groupbys
	propertyMappingNameToDisplayCategoryPropertiesMap := make(map[string]map[string]model.Property)
	for _, filter := range kpiQueryGroup.GlobalFilters {
		// Check if filter is a property mapping and not already present in the map
		if filter.IsPropertyMapping {
			// Validation for unused fields in property mapping filter from payload
			if filter.ObjectType != "" || filter.Entity != "" {
				log.WithField("project_id", projectID).Error("Invalid request for property mapping filter: ", filter)
				return []model.QueryResult{{}, {}}, http.StatusBadRequest
			}
			if _, ok := propertyMappingNameToDisplayCategoryPropertiesMap[filter.PropertyName]; !ok {
				displayCategoryToPropertiesMap, errMsg, statusCode := store.GetDisplayCategoryToPropertiesByProjectIDAndPropertyMappingName(projectID, filter.PropertyName)
				if statusCode != http.StatusOK {
					log.WithField("project_id", projectID).Error("Failed while retrieving Property Mapping Error: ", errMsg)
					return []model.QueryResult{{}, {}}, statusCode
				}
				propertyMappingNameToDisplayCategoryPropertiesMap[filter.PropertyName] = displayCategoryToPropertiesMap
			}
		}
	}

	for _, groupBy := range kpiQueryGroup.GlobalGroupBy {
		if groupBy.IsPropertyMapping {
			// Validation for unused fields in property mapping groupby from payload
			if groupBy.ObjectType != "" || groupBy.Entity != "" || groupBy.GroupByType != "" {
				log.WithField("project_id", projectID).Error("Invalid request for property mapping groupby: ", groupBy)
				return []model.QueryResult{{}, {}}, http.StatusBadRequest
			}
			if _, ok := propertyMappingNameToDisplayCategoryPropertiesMap[groupBy.PropertyName]; !ok {
				displayCategoryToPropertiesMap, errMsg, statusCode := store.GetDisplayCategoryToPropertiesByProjectIDAndPropertyMappingName(projectID, groupBy.PropertyName)
				if statusCode != http.StatusOK {
					log.WithField("project_id", projectID).Error("Failed while retrieving Property Mapping Error: ", errMsg)
					return []model.QueryResult{{}, {}}, statusCode
				}
				propertyMappingNameToDisplayCategoryPropertiesMap[groupBy.PropertyName] = displayCategoryToPropertiesMap
			}
		}
	}

	// Stores the mapping of external query to internal queries in case of derived kpi
	externalGBTQueryToInternalQueries := make(map[string]model.KPIQueryGroup)
	externalNonGBTQueryToInternalQueries := make(map[string]model.KPIQueryGroup)
	for index, query := range kpiQueryGroup.Queries {
		kpiQueryGroup.Queries[index].Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		kpiQueryGroup.Queries[index].GroupBy = kpiQueryGroup.GlobalGroupBy
	}

	// Build internal queries and reuse further
	for _, query := range kpiQueryGroup.Queries {
		if query.QueryType == model.KpiDerivedQueryType {
			internalKPIQueryGroup := model.KPIQueryGroup{}

			derivedMetric, errMsg, statusCode := store.GetDerivedCustomMetricByProjectIdName(projectID, query.Metrics[0])
			// Previously these errors did not stop execution of ExecuteKPIQueryGroup
			if statusCode != http.StatusFound {
				log.WithField("project_id", projectID).Error("Failed while retrieving derived metric: ", errMsg)
				return []model.QueryResult{{}, {}}, http.StatusInternalServerError
			}

			err := U.DecodePostgresJsonbToStructType(derivedMetric.Transformations, &internalKPIQueryGroup)
			if err != nil {
				log.WithField("project_id", projectID).Error("Failed while decoding transformations: ", err)
				return []model.QueryResult{{}, {}}, http.StatusInternalServerError
			}

			internalKPIQueryGroup.DisplayResultAs = derivedMetric.DisplayResultAs
			for internalIndex, internalQuery := range internalKPIQueryGroup.Queries {
				internalKPIQueryGroup.Queries[internalIndex].Filters = append(internalQuery.Filters, query.Filters...)
				internalKPIQueryGroup.Queries[internalIndex].GroupBy = query.GroupBy
				internalKPIQueryGroup.Queries[internalIndex].From = query.From
				internalKPIQueryGroup.Queries[internalIndex].To = query.To
				internalKPIQueryGroup.Queries[internalIndex].Timezone = query.Timezone
				internalKPIQueryGroup.Queries[internalIndex].GroupByTimestamp = query.GroupByTimestamp
			}

			hashCode, err := query.GetQueryCacheHashString()
			if err != nil {
				log.WithField("project_id", projectID).Error("Failed while generating hashcode for derived query: ", err)
				return []model.QueryResult{{}, {}}, http.StatusInternalServerError
			}
			if query.GroupByTimestamp == "" {
				externalNonGBTQueryToInternalQueries[hashCode] = internalKPIQueryGroup
			} else {
				externalGBTQueryToInternalQueries[hashCode] = internalKPIQueryGroup
			}

		}
	}

	finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults,
		mapOfGBTKPINormalQueryToResults := store.ExecuteKPIQueriesAndGetResultsAsMap(projectID,
		reqID, kpiQueryGroup, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, propertyMappingNameToDisplayCategoryPropertiesMap, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
	if finalStatusCode == 2 {
		return []model.QueryResult{{}, {}}, http.StatusInternalServerError
	}
	if finalStatusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, finalStatusCode
	}
	finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults = model.GetNonGBTResultsFromGBTResultsAndMaps(reqID, kpiQueryGroup,
		mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults, mapOfGBTKPINormalQueryToResults, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries)
	if finalStatusCode == 2 {
		return []model.QueryResult{{}, {}}, http.StatusInternalServerError
	}
	if finalStatusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, finalStatusCode
	}

	finalResultantResults, finalStatusCode = model.GetFinalResultantResultsForKPI(reqID, kpiQueryGroup, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults,
		mapOfNonGBTKPINormalQueryToResults, mapOfGBTKPINormalQueryToResults, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries)
	if finalStatusCode == 2 {
		return []model.QueryResult{{}, {}}, http.StatusInternalServerError
	}
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

// ExecuteKPIQueriesAndGetResultsAsMap executes the queries in the KPIQueryGroup
// It parallelises the execution of queries with upto model.AllowedGoroutinesForKPI goroutines
// Uses runSingleKPIQuery to execute a single query
// Returns the final status code, map of nonGBT derived KPI to internal KPI to results, map of GBT derived KPI to internal KPI to results,
// map of nonGBT normal query to results, map of GBT normal query to results
func (store *MemSQL) ExecuteKPIQueriesAndGetResultsAsMap(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
	externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries map[string]model.KPIQueryGroup,
	propertyMappingNameToDisplayCategoryPropertiesMap map[string]map[string]model.Property,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (int, map[string]map[string][]model.QueryResult,
	map[string]map[string][]model.QueryResult, map[string][]model.QueryResult, map[string][]model.QueryResult) {
	finalStatus := model.KPIStatus{Status: 2}
	logFields := log.Fields{
		"project_id": projectID,
		"query":      kpiQueryGroup,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	mapOfGBTDerivedKPIToInternalKPIToResults := make(map[string]map[string][]model.QueryResult)
	mapOfNonGBTDerivedKPIToInternalKPIToResults := make(map[string]map[string][]model.QueryResult)

	mapOfGBTKPINormalQueryToResults := make(map[string][]model.QueryResult)
	mapOfNonGBTKPINormalQueryToResults := make(map[string][]model.QueryResult)

	var waitGroup sync.WaitGroup
	count := 0
	actualRoutineLimit := U.MinInt(len(kpiQueryGroup.Queries), model.AllowedGoroutinesForKPI)
	waitGroup.Add(actualRoutineLimit)
	for _, query := range kpiQueryGroup.Queries {
		count++
		go store.runSingleKPIQuery(projectID, reqID, kpiQueryGroup, query, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, propertyMappingNameToDisplayCategoryPropertiesMap,
			enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery,
			mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTKPINormalQueryToResults,
			mapOfNonGBTKPINormalQueryToResults, &finalStatus, &waitGroup)

		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(kpiQueryGroup.Queries)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	if finalStatus.Status != http.StatusOK {
		mapOfGBTDerivedKPIToInternalKPIToResults = make(map[string]map[string][]model.QueryResult)
		mapOfNonGBTDerivedKPIToInternalKPIToResults = make(map[string]map[string][]model.QueryResult)

		mapOfGBTKPINormalQueryToResults = make(map[string][]model.QueryResult)
		mapOfNonGBTKPINormalQueryToResults = make(map[string][]model.QueryResult)
	}

	return finalStatus.Status, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults, mapOfGBTKPINormalQueryToResults
}

// runSingleKPIQuery runs a single KPI query
// It branches execution into 2 cases:
// 1. Normal KPI query using ExecuteNonDerivedKPIQuery
// 2. Derived KPI query using ExecuteDerivedKPIQuery
// Adds the result to the map of results
// internalKPIQuery, internalQueryToQueryResult, statusCode, derivedKPIHashCode, errMsg
func (store *MemSQL) runSingleKPIQuery(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup, query model.KPIQuery,
	externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries map[string]model.KPIQueryGroup,
	propertyMappingNameToDisplayCategoryPropertiesMap map[string]map[string]model.Property,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool,
	mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTDerivedKPIToInternalKPIToResults map[string]map[string][]model.QueryResult,
	mapOfGBTKPINormalQueryToResults, mapOfNonGBTKPINormalQueryToResults map[string][]model.QueryResult,
	finalStatus *model.KPIStatus, waitGroup *sync.WaitGroup) {

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer waitGroup.Done()
	var result []model.QueryResult
	var statusCode int
	var errMsg string
	internalQueryToQueryResult := make(map[string][]model.QueryResult)

	if query.QueryType == model.KpiDerivedQueryType {
		var derivedKPIHashCode string
		internalQueryToQueryResult, statusCode, derivedKPIHashCode, errMsg = store.ExecuteDerivedKPIQuery(projectID, reqID, query, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, propertyMappingNameToDisplayCategoryPropertiesMap, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		finalStatus.CheckAndSetStatus(statusCode)
		if statusCode != http.StatusOK {
			log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).Error(errMsg)

		} else {
			if query.GroupByTimestamp == "" {
				mapOfNonGBTDerivedKPIToInternalKPIToResults[derivedKPIHashCode] = internalQueryToQueryResult
			} else {
				mapOfGBTDerivedKPIToInternalKPIToResults[derivedKPIHashCode] = internalQueryToQueryResult
			}
		}
	} else {
		var hashCode string
		result, statusCode, hashCode, errMsg = store.ExecuteNonDerivedKPIQuery(projectID, reqID, query, propertyMappingNameToDisplayCategoryPropertiesMap, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		finalStatus.CheckAndSetStatus(statusCode)
		if statusCode != http.StatusOK {
			log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).WithField("result", result).Error(errMsg)
		} else {
			if query.GroupByTimestamp == "" {
				mapOfNonGBTKPINormalQueryToResults[hashCode] = result
			} else {
				mapOfGBTKPINormalQueryToResults[hashCode] = result
			}
		}
	}
}

// Query and hashCode from it are not always same, but follows same logic. Query without GBT is considered for hashCode.
// ExecuteDerivedKPIQuery executes all the internal queries of a derived kpi individually using ExecuteNonDerivedKPIQuery
// and returns the result of all the internal queries
// internalQueryToQueryResult, statusCode, derivedKPIHashCode, errMsg
func (store *MemSQL) ExecuteDerivedKPIQuery(projectID int64, reqID string, baseQuery model.KPIQuery,
	externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries map[string]model.KPIQueryGroup,
	propertyMappingNameToDisplayCategoryPropertiesMap map[string]map[string]model.Property,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) (map[string][]model.QueryResult, int, string, string) {

	queryResults := make([]model.QueryResult, 0)
	mapOfInternalQueryToResult := make(map[string][]model.QueryResult)

	derivedQueryHashCode, err := baseQuery.GetQueryCacheHashString()
	if err != nil {
		return mapOfInternalQueryToResult, http.StatusInternalServerError, "", "Failed while generating hashString for kpi."
	}

	var internalKPIQueryGroup model.KPIQueryGroup
	if baseQuery.GroupByTimestamp == "" {
		internalKPIQueryGroup = externalNonGBTQueryToInternalQueries[derivedQueryHashCode]
	} else {
		internalKPIQueryGroup = externalGBTQueryToInternalQueries[derivedQueryHashCode]
	}

	for _, query := range internalKPIQueryGroup.Queries {
		var result []model.QueryResult
		var hashCode string
		result, statusCode, hashCode, errMsg := store.ExecuteNonDerivedKPIQuery(projectID, reqID, query, propertyMappingNameToDisplayCategoryPropertiesMap, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

		if statusCode != http.StatusOK {
			mapOfInternalQueryToResult := make(map[string][]model.QueryResult)
			return mapOfInternalQueryToResult, statusCode, hashCode, errMsg
		}

		mapOfInternalQueryToResult[hashCode] = result

		queryResults = append(queryResults, result...)
	}

	return mapOfInternalQueryToResult, http.StatusOK, derivedQueryHashCode, ""
}

// ExecuteNonDerivedKPIQuery executes a non derived kpi query and returns the result.
// If property mapping is used in filters or group by, it will be resolved to the internal property
// Results are obtained by executing the query based on its QueryType
// Result headers are converted back to property mapping names if any property mapping is group by
// QueryResult, statusCode, hashCode, errMsg
func (store *MemSQL) ExecuteNonDerivedKPIQuery(projectID int64, reqID string,
	query model.KPIQuery, propertyMappingNameToDisplayCategoryPropertiesMap map[string]map[string]model.Property,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int, string, string) {

	result := make([]model.QueryResult, 0)
	statusCode := http.StatusOK
	hashCode := ""

	var tempQuery model.KPIQuery
	U.DeepCopy(&query, &tempQuery)

	// Only one property per display category is allowed in a property mapping
	for index, filter := range tempQuery.Filters {
		if filter.IsPropertyMapping {
			var internalPropertyForCurrentCategory model.Property
			if displayCategoryToPropertiesMap, ok := propertyMappingNameToDisplayCategoryPropertiesMap[filter.PropertyName]; ok {
				if internalPropertyForCurrentCategory, ok = displayCategoryToPropertiesMap[tempQuery.DisplayCategory]; !ok {
					return result, http.StatusBadRequest, hashCode, "Invalid display category " + tempQuery.DisplayCategory + " for property mapping " + filter.PropertyName
				}
			} else {
				return result, http.StatusBadRequest, hashCode, "Invalid property mapping name " + filter.PropertyName
			}
			tempQuery.Filters[index].PropertyName = internalPropertyForCurrentCategory.Name
			tempQuery.Filters[index].Entity = internalPropertyForCurrentCategory.Entity
			tempQuery.Filters[index].ObjectType = internalPropertyForCurrentCategory.ObjectType
			tempQuery.Filters[index].PropertyDataType = internalPropertyForCurrentCategory.DataType
		}
	}

	// In case of property mappings in group by, headers of results will be internal property names.
	// Hence we need to maintain a map of internal property name to property mapping name.
	// Assuming normal properties are not available when property mapping is being used in Global GroupBy
	internalToExternalGroupByHeadersForPropertyMappings := make(map[string]string)

	for index, groupby := range tempQuery.GroupBy {
		if groupby.IsPropertyMapping {
			var internalPropertyForCurrentCategory model.Property
			if displayCategoryToPropertiesMap, ok := propertyMappingNameToDisplayCategoryPropertiesMap[groupby.PropertyName]; ok {
				if internalPropertyForCurrentCategory, ok = displayCategoryToPropertiesMap[tempQuery.DisplayCategory]; !ok {
					return result, http.StatusBadRequest, hashCode, "Invalid display category " + tempQuery.DisplayCategory + " for property mapping " + groupby.PropertyName
				}
			} else {
				return result, http.StatusBadRequest, hashCode, "Invalid property mapping name " + groupby.PropertyName
			}
			tempQuery.GroupBy[index].PropertyName = internalPropertyForCurrentCategory.Name
			tempQuery.GroupBy[index].Entity = internalPropertyForCurrentCategory.Entity
			tempQuery.GroupBy[index].ObjectType = internalPropertyForCurrentCategory.ObjectType
			tempQuery.GroupBy[index].PropertyDataType = internalPropertyForCurrentCategory.DataType
			tempQuery.GroupBy[index].GroupByType = internalPropertyForCurrentCategory.GroupByType
			internalToExternalGroupByHeadersForPropertyMappings[internalPropertyForCurrentCategory.Name] = groupby.PropertyName
		}
	}

	if tempQuery.Category == model.ProfileCategory {
		if tempQuery.GroupByTimestamp != "" {
			result, statusCode = store.ExecuteKPIQueryForProfiles(projectID, reqID,
				tempQuery, enableOptimisedFilterOnProfileQuery)
		} else {
			result = make([]model.QueryResult, 1)
		}
	} else if tempQuery.Category == model.ChannelCategory || tempQuery.Category == model.CustomChannelCategory {
		result, statusCode = store.ExecuteKPIQueryForChannels(projectID, reqID, tempQuery)
	} else if tempQuery.Category == model.EventCategory {
		result, statusCode = store.ExecuteKPIQueryForEvents(projectID, reqID, tempQuery, enableOptimisedFilterOnEventUserQuery)
	}

	// Replace internal group by headers with property mapping headers
	for index, res := range result {
		newResultHeaders := make([]string, 0)
		for _, header := range res.Headers {
			// if header is present in map then replace else add as it is in newResultHeaders
			if externalHeader, ok := internalToExternalGroupByHeadersForPropertyMappings[header]; ok {
				newResultHeaders = append(newResultHeaders, externalHeader)
			} else {
				newResultHeaders = append(newResultHeaders, header)
			}
		}
		result[index].Headers = newResultHeaders
	}

	hashCode, _ = query.GetQueryCacheHashString()

	return result, statusCode, hashCode, ""
}
