package model

import (
	U "factors/util"
	"fmt"
)

func ValidateKPIQuery(kpiQuery KPIQuery) bool {
	if kpiQuery.DisplayCategory == WebsiteSessionDisplayCategory {
		return ValidateKPISessions(kpiQuery)
	} else if kpiQuery.DisplayCategory == PageViewsDisplayCategory {
		return ValidateKPIPageView(kpiQuery)
	} else if kpiQuery.DisplayCategory == FormSubmissionsDisplayCategory {
		return ValidateKPIFormSubmissions(kpiQuery)
	} else if kpiQuery.DisplayCategory == HubspotContactsDisplayCategory {
		return ValidateKPIHubspotContacts(kpiQuery)
	} else if kpiQuery.DisplayCategory == HubspotCompaniesDisplayCategory {
		return ValidateKPIHubspotCompanies(kpiQuery)
	} else if kpiQuery.DisplayCategory == SalesforceUsersDisplayCategory {
		return ValidateKPISalesforceUsers(kpiQuery)
	} else if kpiQuery.DisplayCategory == SalesforceAccountsDisplayCategory {
		return ValidateKPISalesforceAccounts(kpiQuery)
	} else if kpiQuery.DisplayCategory == SalesforceOpportunitiesDisplayCategory {
		return ValidateKPISalesforceOpportunities(kpiQuery)
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
		currentQuery.EventsWithProperties = prependFiltersBasedOnInternalTransformation(metricTransformation.Filters, query.EventsWithProperties)
		finalResultantQueries = append(finalResultantQueries, currentQuery)
	}
	return finalResultantQueries
}

func prependFiltersBasedOnInternalTransformation(filters []QueryProperty, eventsWithProperties []QueryEventWithProperties) []QueryEventWithProperties {
	var filtersBasedOnMetric []QueryProperty
	filtersBasedOnMetric = append(filtersBasedOnMetric, filters...)
	eventsWithProperties[0].Properties = append(filtersBasedOnMetric, eventsWithProperties[0].Properties...)
	return eventsWithProperties
}

// Functions supporting transforming eventResults to KPIresults
// Note: Considering the format to be generally... event_index, event_name,..., count.
func TransformResultsToKPIResults(results []*QueryResult, hasGroupByTimestamp bool, hasAnyGroupBy bool, displayCategory string) []*QueryResult {
	resultantResults := make([]*QueryResult, 0)
	for _, result := range results {
		var tmpResult *QueryResult
		tmpResult = &QueryResult{}

		tmpResult.Headers = getTransformedHeaders(result.Headers, hasGroupByTimestamp, hasAnyGroupBy, displayCategory)
		tmpResult.Rows = getTransformedRows(result.Rows, hasGroupByTimestamp, hasAnyGroupBy)
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

// append(row[1:2], row[3:]...))
// TODO: validate if rows are there or not.
func getTransformedRows(rows [][]interface{}, hasGroupByTimestamp bool, hasAnyGroupBy bool) [][]interface{} {
	var currentRows [][]interface{}
	currentRows = make([][]interface{}, 0)
	for _, row := range rows {
		if hasAnyGroupBy && hasGroupByTimestamp {
			currentRow := append(row[1:2], row[3:]...)
			currentRows = append(currentRows, currentRow)
		} else if !hasAnyGroupBy && hasGroupByTimestamp {
			currentRows = append(currentRows, row)
		} else {
			currentRows = append(currentRows, row[2:])
		}
	}
	return currentRows
}

// Each KPI metric is internally converted to event analytics.
// Considering all rows to be equal in size because of analytics response.
// resultAsMap - key with groupByColumns, value as row.
func HandlingEventResultsByApplyingOperations(results []*QueryResult, transformations []TransformQueryi) QueryResult {
	resultAsMap := make(map[string][]interface{})
	finalResultRows := make([][]interface{}, 0)
	var finalResult QueryResult
	for index, result := range results {
		if index == 0 {
			resultAsMap = makeHashWithKeyAsGroupBy(result.Rows)
		} else {
			intermediateResultsAsMap := make(map[string][]interface{})
			for _, row := range result.Rows {
				key := getkeyFromRow(row)
				value1 := resultAsMap[key][len(row)-1]
				value2 := row[len(row)-1]
				operator := transformations[index-1].Metrics.Operator
				result := getValueFromValuesAndOperator(value1, value2, operator)
				row[len(row)-1] = result
				intermediateResultsAsMap[key] = row
			}
			resultAsMap = intermediateResultsAsMap
		}
	}
	for _, value := range resultAsMap {
		finalResultRows = append(finalResultRows, value)
	}
	finalResult.Headers = results[0].Headers
	finalResult.Rows = finalResultRows
	return finalResult
}

// TODO: Decide value when divided by 0.
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
		key := getkeyFromRow(row)
		hashMap[key] = row
	}
	return hashMap
}

func getkeyFromRow(row []interface{}) string {
	if len(row) <= 1 {
		return "1"
	}
	var key string
	for _, value := range row[:len(row)-1] {
		key = fmt.Sprintf("%v", value) + ":"
	}
	return key
}
