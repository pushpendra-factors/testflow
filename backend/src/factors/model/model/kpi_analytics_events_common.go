package model

import (
	U "factors/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TO Change.
func ValidateKPIQuery(kpiQuery KPIQuery) bool {
	if kpiQuery.DisplayCategory == WebsiteSessionDisplayCategory {
		return ValidateKPISessions(kpiQuery)
	} else if kpiQuery.DisplayCategory == PageViewsDisplayCategory {
		return ValidateKPIPageView(kpiQuery)
	} else if kpiQuery.DisplayCategory == FormSubmissionsDisplayCategory {
		return ValidateKPIFormSubmissions(kpiQuery)
		// } else if kpiQuery.DisplayCategory == HubspotContactsDisplayCategory {
		// 	return ValidateKPIHubspotContacts(kpiQuery)
		// } else if kpiQuery.DisplayCategory == HubspotCompaniesDisplayCategory {
		// 	return ValidateKPIHubspotCompanies(kpiQuery)
		// } else if kpiQuery.DisplayCategory == SalesforceUsersDisplayCategory {
		// 	return ValidateKPISalesforceUsers(kpiQuery)
		// } else if kpiQuery.DisplayCategory == SalesforceAccountsDisplayCategory {
		// 	return ValidateKPISalesforceAccounts(kpiQuery)
		// } else if kpiQuery.DisplayCategory == SalesforceOpportunitiesDisplayCategory {
		// 	return ValidateKPISalesforceOpportunities(kpiQuery)
	} else {
		return false
	}
}

func GetDirectDerviableQueryPropsFromKPI(kpiQuery KPIQuery) Query {
	var query Query
	query.Class = "events"
	query.GroupByTimestamp = kpiQuery.GroupByTimestamp
	query.EventsCondition = "each_given_event"
	query.Timezone = kpiQuery.Timezone
	query.From = kpiQuery.From
	query.To = kpiQuery.To
	return query
}

func BuildFiltersAndGroupByBasedOnKPIQuery(query Query, kpiQuery KPIQuery, metric string) Query {
	objectType := GetObjectTypeForQueryExecute(kpiQuery.DisplayCategory, metric, kpiQuery.PageUrl)
	query.EventsWithProperties, query.GlobalUserProperties = getFilterEventsForEventAnalytics(kpiQuery.Filters, objectType)
	query.GroupByProperties = getGroupByEventsForEventsAnalytics(kpiQuery.GroupBy, objectType)
	return query
}

func GetObjectTypeForQueryExecute(displayCategory string, metric string, pageUrl string) string {
	metricsData := MapOfMetricsToData[displayCategory][metric]
	var objectType string
	if displayCategory != PageViewsDisplayCategory {
		objectType = metricsData["object_type"]
	} else if displayCategory == PageViewsDisplayCategory && U.ContainsStringInArray([]string{Entrances, Exits}, metric) {
		objectType = U.EVENT_NAME_SESSION
	} else {
		objectType = pageUrl
	}
	return objectType
}

func GetObjectTypeForFilterValues(displayCategory string, metric string) string {
	var objectType string
	if displayCategory == WebsiteSessionDisplayCategory {
		objectType = U.EVENT_NAME_SESSION
	} else if displayCategory == FormSubmissionsDisplayCategory {
		objectType = U.EVENT_NAME_FORM_SUBMITTED
	} else if U.ContainsStringInArray([]string{HubspotContactsDisplayCategory, HubspotCompaniesDisplayCategory, SalesforceUsersDisplayCategory,
		SalesforceAccountsDisplayCategory, SalesforceOpportunitiesDisplayCategory}, displayCategory) {
		metricsData := MapOfMetricsToData[displayCategory][metric]
		objectType = metricsData["object_type"]
	} else { // pageViews case as default.
		objectType = displayCategory
	}

	return objectType
}

func getFilterEventsForEventAnalytics(filters []KPIFilter, objectType string) ([]QueryEventWithProperties, []QueryProperty) {
	var filterForEventEventAnalytics QueryEventWithProperties
	var filterForUserPropertiesEventAnalytics []QueryProperty
	var currentFilterProperties QueryProperty
	filterForEventEventAnalytics.Name = objectType
	if len(filters) == 0 {
		return []QueryEventWithProperties{filterForEventEventAnalytics}, filterForUserPropertiesEventAnalytics
	}

	for _, filter := range filters {
		currentFilterProperties.Entity = filter.Entity
		currentFilterProperties.Type = filter.PropertyDataType
		currentFilterProperties.Property = filter.PropertyName
		currentFilterProperties.Operator = filter.Condition
		currentFilterProperties.Value = filter.Value
		currentFilterProperties.LogicalOp = filter.LogicalOp
		filterForEventEventAnalytics.Properties = append(filterForEventEventAnalytics.Properties, currentFilterProperties)
	}
	return []QueryEventWithProperties{filterForEventEventAnalytics}, filterForUserPropertiesEventAnalytics
}

func getGroupByEventsForEventsAnalytics(groupBys []KPIGroupBy, objectType string) []QueryGroupByProperty {
	var groupBysForEventAnalytics []QueryGroupByProperty
	var currentGroupByProperty QueryGroupByProperty

	for _, kpiGroupBy := range groupBys {
		currentGroupByProperty = QueryGroupByProperty{}
		currentGroupByProperty.Property = kpiGroupBy.PropertyName
		currentGroupByProperty.Type = kpiGroupBy.PropertyDataType
		currentGroupByProperty.GroupByType = kpiGroupBy.GroupByType //Raw or bucketed
		// currentGroupByProperty.Index = index
		currentGroupByProperty.EventName = objectType
		currentGroupByProperty.EventNameIndex = 1
		currentGroupByProperty.Granularity = kpiGroupBy.Granularity
		currentGroupByProperty.Entity = kpiGroupBy.Entity

		groupBysForEventAnalytics = append(groupBysForEventAnalytics, currentGroupByProperty)
	}
	return groupBysForEventAnalytics
}

func SplitKPIQueryToInternalKPIQueries(query Query, kpiQuery KPIQuery, metric string, transformations []TransformQueryi) []Query {
	var finalResultantQueries []Query
	for _, metricTransformation := range transformations {
		currentQuery := query
		if metricTransformation.Metrics.Entity == EventEntity {
			currentQuery.Type = "events_occurrence"
		} else {
			currentQuery.Type = "unique_users"
		}
		currentQuery.AggregateFunction = metricTransformation.Metrics.Aggregation
		currentQuery.AggregateProperty = metricTransformation.Metrics.Property
		currentQuery.AggregateEntity = metricTransformation.Metrics.Entity
		currentQuery.AggregatePropertyType = metricTransformation.Metrics.GroupByType
		currentQuery.EventsWithProperties = prependEventFiltersBasedOnInternalTransformation(metricTransformation.Filters, query.EventsWithProperties, kpiQuery, metric)
		currentQuery.GlobalUserProperties = prependUserFiltersBasedOnInternalTransformation(metricTransformation.Filters, query.GlobalUserProperties, kpiQuery, metric)
		finalResultantQueries = append(finalResultantQueries, currentQuery)
	}
	return finalResultantQueries
}

func prependEventFiltersBasedOnInternalTransformation(filters []QueryProperty, eventsWithProperties []QueryEventWithProperties, kpiQuery KPIQuery, metric string) []QueryEventWithProperties {
	resultantEventsWithProperties := make([]QueryEventWithProperties, 1)
	var filtersBasedOnMetric []QueryProperty
	if kpiQuery.DisplayCategory == PageViewsDisplayCategory && U.ContainsStringInArray([]string{Entrances, Exits}, metric) {
		for _, filter := range filters {
			filtersBasedOnMetric = append(filtersBasedOnMetric, QueryProperty{
				Entity:    filter.Entity,
				Type:      filter.Type,
				Property:  filter.Property,
				Operator:  filter.Operator,
				LogicalOp: filter.LogicalOp,
				Value:     kpiQuery.PageUrl,
			})
		}
	} else {
		filtersBasedOnMetric = filters
	}
	resultantEventsWithProperties[0].Name = eventsWithProperties[0].Name
	resultantEventsWithProperties[0].AliasName = eventsWithProperties[0].AliasName
	resultantEventsWithProperties[0].Properties = append(filtersBasedOnMetric, eventsWithProperties[0].Properties...)
	return resultantEventsWithProperties
}

func prependUserFiltersBasedOnInternalTransformation(filters []QueryProperty, userProperties []QueryProperty, kpiQuery KPIQuery, metric string) []QueryProperty {
	return make([]QueryProperty, 0)
}

// Functions supporting transforming eventResults to KPIresults
// Note: Considering the format to be generally... event_index, event_name,..., count.
func TransformResultsToKPIResults(results []*QueryResult, hasGroupByTimestamp bool, hasAnyGroupBy bool, displayCategory string, timezoneString string) []*QueryResult {
	resultantResults := make([]*QueryResult, 0)
	for _, result := range results {
		var tmpResult *QueryResult
		tmpResult = &QueryResult{}

		tmpResult.Headers = getTransformedHeaders(result.Headers, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		tmpResult.Rows = GetTransformedRows(tmpResult.Headers, result.Rows, hasGroupByTimestamp, hasAnyGroupBy, len(result.Headers), timezoneString)
		resultantResults = append(resultantResults, tmpResult)
	}
	return resultantResults
}

func getTransformedHeaders(headers []string, hasGroupByTimestamp bool, hasAnyGroupBy bool, displayCategory string) []string {
	currentHeaders := make([]string, 0)
	if hasAnyGroupBy && hasGroupByTimestamp {
		currentHeaders = append(headers[1:2], headers[3:]...)
	} else if !hasAnyGroupBy && hasGroupByTimestamp {
		headers[1] = AliasAggr
		currentHeaders = headers
	} else {
		currentHeaders = headers[2:]
	}
	return currentHeaders
}

func GetTransformedRows(headers []string, rows [][]interface{}, hasGroupByTimestamp bool, hasAnyGroupBy bool, headersLen int, timezoneString string) [][]interface{} {
	var currentRows [][]interface{}
	currentRows = make([][]interface{}, 0)
	if len(rows) == 0 {
		return currentRows
	}

	for _, row := range rows {
		var currentRow []interface{}
		if len(row) == 0 {
			currentRow = make([]interface{}, headersLen)
			for index := range currentRow[:headersLen-1] {
				currentRow[index] = ""
			}
			currentRow[headersLen-1] = 0
		} else {
			currentRow = row
		}
		if hasAnyGroupBy && hasGroupByTimestamp {
			currentRow = append(currentRow[1:2], currentRow[3:]...)
			currentRows = append(currentRows, currentRow)
		} else if !hasAnyGroupBy && hasGroupByTimestamp {
			currentRows = append(currentRows, currentRow)
		} else {
			currentRows = append(currentRows, currentRow[2:])
		}
	}

	currentRows = TransformDateTypeValueForEventsKPI(headers, currentRows, hasGroupByTimestamp, timezoneString)
	return currentRows
}

func TransformDateTypeValueForEventsKPI(headers []string, rows [][]interface{}, groupByTimestampPresent bool, timezoneString string) [][]interface{} {
	indexForDateTime := -1
	if !groupByTimestampPresent {
		return rows
	}
	for index, header := range headers {
		if header == "datetime" {
			indexForDateTime = index
			break
		}
	}

	for index, row := range rows {
		currentValueInTimeFormat, _ := row[indexForDateTime].(time.Time)
		loc, _ := time.LoadLocation(timezoneString)
		currentValueInTimeFormat = currentValueInTimeFormat.In(loc)
		rows[index][indexForDateTime] = U.GetTimestampAsStrWithTimezone(currentValueInTimeFormat, timezoneString)
	}

	return rows
}

// Each KPI metric is internally converted to event analytics.
// Considering all rows to be equal in size because of analytics response.
// resultAsMap - key with groupByColumns, value as row.
func HandlingEventResultsByApplyingOperations(results []*QueryResult, transformations []TransformQueryi, timezone string, isTimezoneEnabled bool) QueryResult {
	resultKeys := getAllKeysFromResults(results)
	var finalResult QueryResult
	finalResultRows := make([][]interface{}, 0)
	for index, result := range results {
		if index == 0 {
			resultKeys = addValuesToHashMap(resultKeys, result.Rows)
		} else {
			for _, row := range result.Rows {
				key := U.GetkeyFromRow(row)
				value1 := resultKeys[key]
				value2 := row[len(row)-1]
				operator := transformations[index-1].Metrics.Operator
				result := getValueFromValuesAndOperator(value1, value2, operator)
				resultKeys[key] = result
			}
		}
	}

	for key, value := range resultKeys {
		row := make([]interface{}, 0)
		columns := strings.Split(key, ":;")
		for _, column := range columns[:len(columns)-1] {
			if strings.HasPrefix(column, "dat$") {
				unixValue, _ := strconv.ParseInt(strings.TrimPrefix(column, "dat$"), 10, 64)
				columnValue, _ := U.GetTimeFromUnixTimestampWithZone(unixValue, timezone, isTimezoneEnabled)
				row = append(row, columnValue)
			} else {
				row = append(row, column)
			}
		}
		row = append(row, value)
		finalResultRows = append(finalResultRows, row)
	}
	finalResultRows = U.GetSorted2DArrays(finalResultRows)
	finalResult.Headers = results[0].Headers
	finalResult.Rows = finalResultRows
	return finalResult
}

func getAllKeysFromResults(results []*QueryResult) map[string]interface{} {
	resultKeys := make(map[string]interface{}, 0)
	var key string
	for _, result := range results {
		for _, row := range result.Rows {
			key = U.GetkeyFromRow(row)
			resultKeys[key] = 0
		}
	}
	return resultKeys
}

func addValuesToHashMap(resultKeys map[string]interface{}, rows [][]interface{}) map[string]interface{} {
	for _, row := range rows {
		key := U.GetkeyFromRow(row)
		resultKeys[key] = row[len(row)-1]
	}
	return resultKeys
}

func getValueFromValuesAndOperator(value1 interface{}, value2 interface{}, operator string) float64 {
	var result float64
	value1InFloat := U.SafeConvertToFloat64(value1)
	value2InFloat := U.SafeConvertToFloat64(value2)
	if operator == "Division" {
		if value2InFloat == 0 {
			result = 0
		} else {
			result = value1InFloat / value2InFloat
		}
	}
	return result
}

func makeHashWithKeyAsGroupBy(rows [][]interface{}) map[string][]interface{} {
	var hashMap map[string][]interface{} = make(map[string][]interface{})
	for _, row := range rows {
		key := U.GetkeyFromRow(row)
		hashMap[key] = row
	}
	return hashMap
}

func SplitQueryResultsIntoGBTAndNonGBT(queryResults []QueryResult, kpiQueryGroup KPIQueryGroup, finalStatusCode int) ([]QueryResult, []QueryResult, []KPIQuery, []KPIQuery) {
	gbtRelatedQueryResults := make([]QueryResult, 0)
	gbtRelatedQueries := make([]KPIQuery, 0)
	nonGbtRelatedQueryResults := make([]QueryResult, 0)
	nonGbtRelatedQueries := make([]KPIQuery, 0)
	for index, kpiQuery := range kpiQueryGroup.Queries {
		if kpiQuery.GroupByTimestamp != "" {
			gbtRelatedQueryResults = append(gbtRelatedQueryResults, queryResults[index])
			gbtRelatedQueries = append(gbtRelatedQueries, kpiQuery)
		} else {
			nonGbtRelatedQueryResults = append(nonGbtRelatedQueryResults, queryResults[index])
			nonGbtRelatedQueries = append(nonGbtRelatedQueries, kpiQuery)
		}
	}
	if len(nonGbtRelatedQueries) == 0 {
		nonGbtRelatedQueryResults = nil
	}
	if len(gbtRelatedQueries) == 0 {
		gbtRelatedQueryResults = nil
	}
	return gbtRelatedQueryResults, nonGbtRelatedQueryResults, gbtRelatedQueries, nonGbtRelatedQueries
}

func MergeQueryResults(queryResults []QueryResult, queries []KPIQuery, timezoneString string, finalStatusCode int, isTimezoneEnabled bool) QueryResult {
	if finalStatusCode != http.StatusOK || len(queryResults) == 0 {
		queryResult := QueryResult{}
		return queryResult
	}

	queryResult := QueryResult{}
	queryResult.Headers = TransformColumnResultGroup(queryResults, queries, timezoneString)
	queryResult.Rows = TransformRowsResultGroup(queryResults, timezoneString, isTimezoneEnabled)
	return queryResult
}

// NOTE: Basing on single metric being sent per query.
func TransformColumnResultGroup(queryResults []QueryResult, queries []KPIQuery, timezoneString string) []string {
	finalResultantColumns := make([]string, 0)
	for index, queryResult := range queryResults {
		if index == 0 {
			finalResultantColumns = append(queryResult.Headers[:len(queryResult.Headers)-1], queries[index].Metrics...)
		} else {
			finalResultantColumns = append(finalResultantColumns, queries[index].Metrics...)
		}
	}
	return finalResultantColumns
}

// Form Map with key as combination of columns and values.
// Steps involved are as follows.
// 1. Make an empty hashMap with key and value as array of 0's as prefixed values.
// 2. Add the values to hashMap. Here keys are contextual to kpi and will not be duplicate.
// 3. Convert Map to 2d Array and then sort.
func TransformRowsResultGroup(queryResults []QueryResult, timezoneString string, isTimezoneEnabled bool) [][]interface{} {
	resultAsMap := make(map[string][]interface{})
	numberOfQueryResults := len(queryResults)

	// Step 1
	for _, queryResult := range queryResults {
		for _, row := range queryResult.Rows {
			key := U.GetkeyFromRow(row)
			emptyValues := make([]interface{}, numberOfQueryResults)
			for index := range emptyValues {
				emptyValues[index] = 0
			}
			resultAsMap[key] = emptyValues
		}
	}

	// Step 2
	for queryIndex, queryResult := range queryResults {
		for _, row := range queryResult.Rows {
			key := U.GetkeyFromRow(row)
			val := row[len(row)-1]
			resultAsMap[key][queryIndex] = val
		}
	}

	// Step 3
	finalResultantRows := make([][]interface{}, 0, 0)
	for key, value := range resultAsMap {
		currentRow := make([]interface{}, 0)
		columns := strings.Split(key, ":;")
		for _, column := range columns[:len(columns)-1] {
			if strings.HasPrefix(column, "dat$") {
				unixValue, _ := strconv.ParseInt(strings.TrimPrefix(column, "dat$"), 10, 64)
				columnValue, _ := U.GetTimeFromUnixTimestampWithZone(unixValue, timezoneString, isTimezoneEnabled)
				currentRow = append(currentRow, columnValue)
			} else {
				currentRow = append(currentRow, column)
			}
		}
		currentRow = append(currentRow, value...)

		finalResultantRows = append(finalResultantRows, currentRow)
	}
	finalResultantRows = U.GetSorted2DArrays(finalResultantRows)
	return finalResultantRows
}
