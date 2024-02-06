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

// KPI Query Group Execution Flow:
// Validate KPI Query Group
// Get all property mappings used in global filters and groupbys using fetchPropertyMappingsForKPIQueryGroupGlobals
// Add global filters and global groupby to each query
// Get internal kpi query group of derived kpis and store them in GBT and NonGBT maps using buildInternalQueryGroupForDerivedKPIs
// buildInternalQueryGroupForDerivedKPIs fetches the derived kpi transformation, stores it as a KPIQueryGroup and adds the filters and group bys of the derived query to each query of this group

// ExecuteKPIQueriesAndGetResultsAsMap:
//		- Run each query using runSingleKPIQuery:
//			- Non derived execution (ExecuteNonDerivedKPIQuery)
//				- Resolve property mappings used in the query by resolvePropertyMappingFiltersAndGroupBysToInternalProperties
//				- Based on category execute query
//				- replaceInternalPropertyHeadersWithPropertyMappingHeaders reverts result headers to property mappings
//				- Compute hashstring
//				- Return results and hashstring
//			- Derived execution (ExecuteDerivedKPIQuery)
//				- Get internal query group using the maps created at buildInternalQueryGroupForDerivedKPIs
//				- Execute each query as non derived using ExecuteNonDerivedKPIQuery to get results and hashstring to create a mapOfInternalQueryToResult
//				- Compute hashstring for derived query
//				- Return mapOfInternalQueryToResult and hashstring
//			- Add results to various maps (GBT NonGBT GBTDerived NonGBTDerived) with hashstring as key
//			- Return these maps

// GetNonGBTResultsFromGBTResultsAndMaps
// GetFinalResultantResultsForKPI
// SplitQueryResultsIntoGBTAndNonGBT
// MergeQueryResults for GBT
// MergeQueryResults for NonGBT
// Merge them to form finalQueryResult

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

	if ok, errMsg := kpiQueryGroup.IsValid(); !ok {
		log.WithField("project_id", projectID).Error(errMsg)
		return []model.QueryResult{{}, {}}, http.StatusBadRequest
	}

	// Fetch all property mappings from global filters and groupbys from db and store in a map
	mapOfPropertyMappingNameToDisplayCategoryToProperty, statusCode := store.fetchPropertyMappingsForKPIQueryGroupGlobals(kpiQueryGroup, projectID)
	if statusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, statusCode
	}

	// Check if the property mappings used are having display categories of each query
	if !kpiQueryGroup.AreDisplayCategoriesPresentInPropertyMapping(mapOfPropertyMappingNameToDisplayCategoryToProperty, kpiQueryGroup.GlobalFilters, kpiQueryGroup.GlobalGroupBy) {
		log.WithField("project_id", projectID).Error("Display category not present in property mapping used")
		return []model.QueryResult{{}, {}}, http.StatusBadRequest
	}

	// Add global filters and groupbys to each query
	for index, query := range kpiQueryGroup.Queries {
		kpiQueryGroup.Queries[index].Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		kpiQueryGroup.Queries[index].GroupBy = kpiQueryGroup.GlobalGroupBy
	}

	// Store the mapping of external query to internal queries of all derived kpi
	externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, statusCode := store.buildInternalQueryGroupForDerivedKPIs(kpiQueryGroup, mapOfPropertyMappingNameToDisplayCategoryToProperty, projectID)
	if statusCode != http.StatusOK {
		return []model.QueryResult{{}, {}}, statusCode
	}

	finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults,
		mapOfGBTKPINormalQueryToResults := store.ExecuteKPIQueriesAndGetResultsAsMap(projectID,
		reqID, kpiQueryGroup, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, mapOfPropertyMappingNameToDisplayCategoryToProperty, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
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
	if gbtRelatedQueryResults == nil || len(gbtRelatedQueryResults) == 0 || len(gbtRelatedQueryResults[0].Headers) == 0 ||
		nonGbtRelatedQueryResults == nil || len(nonGbtRelatedQueryResults) == 0 || len(nonGbtRelatedQueryResults[0].Headers) == 0 {
		return []model.QueryResult{{}, {}}, finalStatusCode
	}

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

// Fetches all the property mappings used in the global filters and groupbys of the KPI Query Group
// returns a map of property mapping name to display category to properties map
// This map is useful for resolving property mappings at single query level
func (store *MemSQL) fetchPropertyMappingsForKPIQueryGroupGlobals(kpiQueryGroup model.KPIQueryGroup, projectID int64) (map[string]map[string]model.Property, int) {
	mapOfPropertyMappingNameToDisplayCategoryToProperty := make(map[string]map[string]model.Property)
	for _, filter := range kpiQueryGroup.GlobalFilters {
		// Check if filter is a property mapping and not already present in the map
		if filter.IsPropertyMapping {
			if _, ok := mapOfPropertyMappingNameToDisplayCategoryToProperty[filter.PropertyName]; !ok {
				displayCategoryToPropertiesMap, errMsg, statusCode := store.GetDisplayCategoryToPropertiesByProjectIDAndPropertyMappingName(projectID, filter.PropertyName)
				if statusCode != http.StatusOK {
					log.WithField("project_id", projectID).Error("Failed while retrieving Property Mapping Error: ", errMsg)
					return mapOfPropertyMappingNameToDisplayCategoryToProperty, statusCode
				}
				mapOfPropertyMappingNameToDisplayCategoryToProperty[filter.PropertyName] = displayCategoryToPropertiesMap
			}
		}
	}

	for _, groupBy := range kpiQueryGroup.GlobalGroupBy {
		if groupBy.IsPropertyMapping {
			if _, ok := mapOfPropertyMappingNameToDisplayCategoryToProperty[groupBy.PropertyName]; !ok {
				displayCategoryToPropertiesMap, errMsg, statusCode := store.GetDisplayCategoryToPropertiesByProjectIDAndPropertyMappingName(projectID, groupBy.PropertyName)
				if statusCode != http.StatusOK {
					log.WithField("project_id", projectID).WithField("err_code", statusCode).Error("Failed while retrieving Property Mapping Error: ", errMsg)
					return mapOfPropertyMappingNameToDisplayCategoryToProperty, statusCode
				}
				mapOfPropertyMappingNameToDisplayCategoryToProperty[groupBy.PropertyName] = displayCategoryToPropertiesMap
			}
		}
	}
	return mapOfPropertyMappingNameToDisplayCategoryToProperty, http.StatusOK
}

// Build internal query group of all derived kpis in the query group
func (store *MemSQL) buildInternalQueryGroupForDerivedKPIs(kpiQueryGroup model.KPIQueryGroup, mapOfPropertyMappingNameToDisplayCategoryToProperty map[string]map[string]model.Property, projectID int64) (map[string]model.KPIQueryGroup, map[string]model.KPIQueryGroup, int) {
	externalGBTQueryToInternalQueries := make(map[string]model.KPIQueryGroup)
	externalNonGBTQueryToInternalQueries := make(map[string]model.KPIQueryGroup)
	for _, query := range kpiQueryGroup.Queries {
		if query.QueryType == model.KpiDerivedQueryType {
			internalKPIQueryGroup := model.KPIQueryGroup{}

			derivedMetric, errMsg, statusCode := store.GetDerivedCustomMetricByProjectIdName(projectID, query.Metrics[0])
			if statusCode != http.StatusFound {
				log.WithField("project_id", projectID).WithField("err_code", statusCode).Error("Failed while retrieving derived metric: ", errMsg)
				return externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, http.StatusInternalServerError
			}

			err := U.DecodePostgresJsonbToStructType(derivedMetric.Transformations, &internalKPIQueryGroup)
			if err != nil {
				log.WithError(err).WithField("project_id", projectID).Error("Failed while decoding transformations: ", err)
				return externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, http.StatusInternalServerError
			}

			// Check if the display categories of internal queries are present in the property mappings used in the global filters and groupbys of the KPI Query Group
			if !internalKPIQueryGroup.AreDisplayCategoriesPresentInPropertyMapping(mapOfPropertyMappingNameToDisplayCategoryToProperty, kpiQueryGroup.GlobalFilters, kpiQueryGroup.GlobalGroupBy) {
				log.WithField("project_id", projectID).Error("Display category of internal queries not present in property mappings")
				return externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, http.StatusBadRequest
			}

			internalKPIQueryGroup.DisplayResultAs = derivedMetric.DisplayResultAs
			for internalIndex, internalQuery := range internalKPIQueryGroup.Queries {
				internalKPIQueryGroup.Queries[internalIndex].Filters = append(internalQuery.Filters, query.Filters...)
				internalKPIQueryGroup.Queries[internalIndex].GroupBy = query.GroupBy
				internalKPIQueryGroup.Queries[internalIndex].From = query.From
				internalKPIQueryGroup.Queries[internalIndex].To = query.To
				internalKPIQueryGroup.Queries[internalIndex].Timezone = query.Timezone
				internalKPIQueryGroup.Queries[internalIndex].GroupByTimestamp = query.GroupByTimestamp
				internalKPIQueryGroup.Queries[internalIndex].LimitNotApplicable = query.LimitNotApplicable
			}

			hashCode, err := query.GetQueryCacheHashString()
			if err != nil {
				log.WithError(err).WithField("project_id", projectID).Error("Failed while generating hashcode for derived query: ", err)
				return externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, http.StatusInternalServerError
			}
			if query.GroupByTimestamp == "" {
				externalNonGBTQueryToInternalQueries[hashCode] = internalKPIQueryGroup
			} else {
				externalGBTQueryToInternalQueries[hashCode] = internalKPIQueryGroup
			}
		}
	}
	return externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, http.StatusOK
}

// ExecuteKPIQueriesAndGetResultsAsMap executes the queries in the KPIQueryGroup
// It parallelises the execution of queries with upto model.AllowedGoroutinesForKPI goroutines
// Uses runSingleKPIQuery to execute a single query
// Returns the final status code, map of nonGBT derived KPI to internal KPI to results, map of GBT derived KPI to internal KPI to results,
// map of nonGBT normal query to results, map of GBT normal query to results
func (store *MemSQL) ExecuteKPIQueriesAndGetResultsAsMap(projectID int64, reqID string, kpiQueryGroup model.KPIQueryGroup,
	externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries map[string]model.KPIQueryGroup,
	mapOfPropertyMappingNameToDisplayCategoryToProperty map[string]map[string]model.Property,
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
		go store.runSingleKPIQuery(projectID, reqID, kpiQueryGroup, query, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, mapOfPropertyMappingNameToDisplayCategoryToProperty,
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
	mapOfPropertyMappingNameToDisplayCategoryToProperty map[string]map[string]model.Property,
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
		internalQueryToQueryResult, statusCode, derivedKPIHashCode, errMsg = store.ExecuteDerivedKPIQuery(projectID, reqID, query, externalGBTQueryToInternalQueries, externalNonGBTQueryToInternalQueries, mapOfPropertyMappingNameToDisplayCategoryToProperty, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		finalStatus.CheckAndSetStatus(statusCode)
		if statusCode != http.StatusOK {
			log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("err_code", statusCode).WithField("query", query).Error(errMsg)

		} else {
			if query.GroupByTimestamp == "" {
				mapOfNonGBTDerivedKPIToInternalKPIToResults[derivedKPIHashCode] = internalQueryToQueryResult
			} else {
				mapOfGBTDerivedKPIToInternalKPIToResults[derivedKPIHashCode] = internalQueryToQueryResult
			}
		}
	} else {
		var hashCode string
		result, statusCode, hashCode, errMsg = store.ExecuteNonDerivedKPIQuery(projectID, reqID, query, mapOfPropertyMappingNameToDisplayCategoryToProperty, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
		finalStatus.CheckAndSetStatus(statusCode)
		if statusCode != http.StatusOK {
			log.WithField("reqID", reqID).WithField("kpiQueryGroup", kpiQueryGroup).WithField("query", query).
				WithField("result", result).WithField("err_code", statusCode).Error(errMsg)
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
	mapOfPropertyMappingNameToDisplayCategoryToProperty map[string]map[string]model.Property,
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
		result, statusCode, hashCode, errMsg := store.ExecuteNonDerivedKPIQuery(projectID, reqID, query, mapOfPropertyMappingNameToDisplayCategoryToProperty, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

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
	query model.KPIQuery, mapOfPropertyMappingNameToDisplayCategoryToProperty map[string]map[string]model.Property,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool) ([]model.QueryResult, int, string, string) {

	result := make([]model.QueryResult, 0)
	statusCode := http.StatusOK
	hashCode, _ := query.GetQueryCacheHashString()

	tempQuery, internalToExternalGroupByHeadersForPropertyMappings, statusCode, errMsg := transformPropertyMappingFiltersAndGroupBysToInternalProperties(query, mapOfPropertyMappingNameToDisplayCategoryToProperty)
	if statusCode != http.StatusOK {
		return result, statusCode, "", errMsg
	}

	if tempQuery.Category == model.ProfileCategory {
		if tempQuery.GroupByTimestamp != "" {
			result, statusCode = store.ExecuteKPIQueryForProfiles(projectID, reqID,
				tempQuery, enableOptimisedFilterOnProfileQuery)
		} else {
			return make([]model.QueryResult, 1), http.StatusOK, hashCode, ""
		}
	} else if tempQuery.Category == model.ChannelCategory || tempQuery.Category == model.CustomChannelCategory {
		result, statusCode = store.ExecuteKPIQueryForChannels(projectID, reqID, tempQuery)
	} else if tempQuery.Category == model.EventCategory {
		result, statusCode = store.ExecuteKPIQueryForEvents(projectID, reqID, tempQuery, enableOptimisedFilterOnEventUserQuery)
	}

	if statusCode != http.StatusOK {
		return result, statusCode, hashCode, ""
	}

	updatedResult := replaceInternalPropertyHeadersWithPropertyMappingHeaders(result, internalToExternalGroupByHeadersForPropertyMappings)
	return updatedResult, statusCode, hashCode, ""
}

// If property mapping is used in filters or group by, it will be resolved to the internal property
// In case of property mappings in group by, headers of results will be internal property names.
// Hence we need to maintain a map of internal property name to property mapping name.
// Assuming normal properties are not available when property mapping is being used in Global GroupBy
func transformPropertyMappingFiltersAndGroupBysToInternalProperties(query model.KPIQuery, mapOfPropertyMappingNameToDisplayCategoryToProperty map[string]map[string]model.Property) (model.KPIQuery, map[int]string, int, string) {
	var tempQuery model.KPIQuery
	U.DeepCopy(&query, &tempQuery)

	internalToExternalGroupByHeadersForPropertyMappings := make(map[int]string)

	// Only one property per display category is allowed in a property mapping
	for index, filter := range tempQuery.Filters {
		if filter.IsPropertyMapping {
			internalPropertyForCurrentCategory := mapOfPropertyMappingNameToDisplayCategoryToProperty[filter.PropertyName][tempQuery.DisplayCategory]
			tempQuery.Filters[index].PropertyName = internalPropertyForCurrentCategory.Name
			tempQuery.Filters[index].Entity = internalPropertyForCurrentCategory.Entity
			tempQuery.Filters[index].ObjectType = internalPropertyForCurrentCategory.ObjectType
			tempQuery.Filters[index].PropertyDataType = internalPropertyForCurrentCategory.DataType
		}
	}

	// for GBT i.e. GroupByTimestamp = "date", 1st header of result is "datetime"
	// Hence all group by headers will be shifted by 1, if GBT is present
	indexIncrement := 0
	if tempQuery.GroupByTimestamp != "" {
		indexIncrement = 1
	}
	for index, groupby := range tempQuery.GroupBy {
		if groupby.IsPropertyMapping {
			internalPropertyForCurrentCategory := mapOfPropertyMappingNameToDisplayCategoryToProperty[groupby.PropertyName][tempQuery.DisplayCategory]
			tempQuery.GroupBy[index].PropertyName = internalPropertyForCurrentCategory.Name
			tempQuery.GroupBy[index].Entity = internalPropertyForCurrentCategory.Entity
			tempQuery.GroupBy[index].ObjectType = internalPropertyForCurrentCategory.ObjectType
			tempQuery.GroupBy[index].PropertyDataType = internalPropertyForCurrentCategory.DataType
			internalToExternalGroupByHeadersForPropertyMappings[index+indexIncrement] = groupby.PropertyName
		}
	}
	return tempQuery, internalToExternalGroupByHeadersForPropertyMappings, http.StatusOK, ""
}

// Replace internal group by headers with property mapping headers using the map created in resolvePropertyMappingFiltersAndGroupBysToInternalProperties
func replaceInternalPropertyHeadersWithPropertyMappingHeaders(result []model.QueryResult, internalToExternalGroupByHeadersForPropertyMappings map[int]string) []model.QueryResult {
	for index, _ := range result {
		for headerIndex, value := range internalToExternalGroupByHeadersForPropertyMappings {
			result[index].Headers[headerIndex] = value
		}
	}
	return result
}
