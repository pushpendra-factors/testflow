package model

import (
	U "factors/util"
	"strconv"
	"strings"
)

const TimestampHeader = "datetime"

var mapOfKPIToProfileType = map[string]string{
	HubspotContactsDisplayCategory:         UserSourceHubspotString,
	HubspotCompaniesDisplayCategory:        UserSourceHubspotString,
	HubspotDealsDisplayCategory:            UserSourceHubspotString,
	SalesforceUsersDisplayCategory:         UserSourceSalesforceString,
	SalesforceAccountsDisplayCategory:      UserSourceSalesforceString,
	SalesforceOpportunitiesDisplayCategory: UserSourceSalesforceString,
}

var mapOfKPICategoryToProfileGroupAnalysis = map[string]string{
	HubspotContactsDisplayCategory:         USERS,
	HubspotCompaniesDisplayCategory:        GROUP_NAME_HUBSPOT_COMPANY,
	HubspotDealsDisplayCategory:            GROUP_NAME_HUBSPOT_DEAL,
	SalesforceUsersDisplayCategory:         USERS,
	SalesforceAccountsDisplayCategory:      GROUP_NAME_SALESFORCE_ACCOUNT,
	SalesforceOpportunitiesDisplayCategory: GROUP_NAME_SALESFORCE_OPPORTUNITY,
}

// Setting and getting Time for profiles query is 0,0. Need to understand.
func GetDirectDerivableProfileQueryFromKPI(kpiQuery KPIQuery) ProfileQueryGroup {
	var profileQueryGroup ProfileQueryGroup
	profileQueryGroup.Class = ProfileQueryClass
	profileQueryGroup.From = kpiQuery.From
	profileQueryGroup.To = kpiQuery.To
	profileQueryGroup.Timezone = kpiQuery.Timezone
	profileQueryGroup.GlobalFilters = transformFiltersKPIToProfiles(kpiQuery.Filters)
	profileQueryGroup.GlobalGroupBys = transformGroupByKPIToProfiles(kpiQuery.GroupBy)
	profileQueryGroup.GroupAnalysis = mapOfKPICategoryToProfileGroupAnalysis[kpiQuery.DisplayCategory]
	return profileQueryGroup
}

func transformFiltersKPIToProfiles(filters []KPIFilter) []QueryProperty {
	var resultantFilters []QueryProperty
	var currentFilters QueryProperty
	for _, filter := range filters {
		currentFilters.Entity = filter.Entity
		currentFilters.Type = filter.PropertyDataType
		currentFilters.Property = filter.PropertyName
		currentFilters.Operator = filter.Condition
		currentFilters.Value = filter.Value
		currentFilters.LogicalOp = filter.LogicalOp
		resultantFilters = append(resultantFilters, currentFilters)
	}
	return resultantFilters
}

func transformGroupByKPIToProfiles(groupBys []KPIGroupBy) []QueryGroupByProperty {
	var resultantGroupBys []QueryGroupByProperty
	var currentGroupByProperty QueryGroupByProperty
	for _, kpiGroupBy := range groupBys {
		currentGroupByProperty = QueryGroupByProperty{}
		currentGroupByProperty.Entity = kpiGroupBy.Entity
		currentGroupByProperty.Property = kpiGroupBy.PropertyName
		currentGroupByProperty.Type = kpiGroupBy.PropertyDataType
		currentGroupByProperty.GroupByType = kpiGroupBy.GroupByType
		currentGroupByProperty.Granularity = kpiGroupBy.Granularity
		resultantGroupBys = append(resultantGroupBys, currentGroupByProperty)
	}
	return resultantGroupBys
}

func GetProfileGroupByFromDateField(dateField string, groupByTimestamp string) QueryGroupByProperty {
	var currentGroupByProperty QueryGroupByProperty
	currentGroupByProperty = QueryGroupByProperty{}
	currentGroupByProperty.Entity = PropertyEntityEvent
	currentGroupByProperty.Property = dateField
	currentGroupByProperty.Type = AliasDateTime
	currentGroupByProperty.Granularity = groupByTimestamp
	return currentGroupByProperty
}

func AddCustomMetricsTransformationsToProfileQuery(profileQueryGroup ProfileQueryGroup, kpiMetric string, customMetric CustomMetric, transformation CustomMetricTransformation, kpiQuery KPIQuery) []ProfileQuery {
	resultantProfileQueries := make([]ProfileQuery, 0)

	if transformation.AggregateFunction == "AVG" {
		currentProfileQuery1 := GetProfileQueriesOnCustomMetric(profileQueryGroup, transformation, SumAggregateFunction, transformation.AggregateProperty, transformation.AggregatePropertyType, kpiQuery)
		currentProfileQuery2 := GetProfileQueriesOnCustomMetric(profileQueryGroup, transformation, Count, "1", "categorical", kpiQuery)
		resultantProfileQueries = append([]ProfileQuery{currentProfileQuery1}, currentProfileQuery2)
	} else {
		currentProfileQuery := GetProfileQueriesOnCustomMetric(profileQueryGroup, transformation, transformation.AggregateFunction, transformation.AggregateProperty, transformation.AggregatePropertyType, kpiQuery)
		resultantProfileQueries = append(resultantProfileQueries, currentProfileQuery)
	}
	return resultantProfileQueries
}

func GetProfileQueriesOnCustomMetric(profileQueryGroup ProfileQueryGroup, transformation CustomMetricTransformation,
	aggregateFunction string, AggregateProperty string, AggregatePropertyType string, kpiQuery KPIQuery) ProfileQuery {
	profileQuery := ProfileQuery{}

	profileCategory, exists := mapOfKPIToProfileType[kpiQuery.DisplayCategory]
	if !exists {
		profileCategory = ""
	}
	profileQuery.AggregateFunction = aggregateFunction
	profileQuery.AggregateProperty = transformation.AggregateProperty
	profileQuery.AggregatePropertyType = transformation.AggregatePropertyType
	profileQuery.From = profileQueryGroup.From
	profileQuery.To = profileQueryGroup.To
	profileQuery.Timezone = profileQueryGroup.Timezone
	profileQuery.Type = profileCategory
	profileQuery.GroupAnalysis = profileQueryGroup.GroupAnalysis

	profileQuery.Filters = append(transformFiltersKPIToProfiles(transformation.Filters), profileQueryGroup.GlobalFilters...)
	if kpiQuery.GroupByTimestamp == "" {
		profileQuery.GroupBys = profileQueryGroup.GlobalGroupBys
	} else {
		profileQuery.GroupBys = append([]QueryGroupByProperty{GetProfileGroupByFromDateField(transformation.DateField, kpiQuery.GroupByTimestamp)}, profileQueryGroup.GlobalGroupBys...)
	}
	for i, _ := range profileQuery.GroupBys {
		profileQuery.GroupBys[i].Index = i
	}
	profileQuery.DateField = transformation.DateField
	return profileQuery
}

// Post Execution Transformations
func TransformProfileResultsToKPIResults(results []QueryResult, hasGroupByTimestamp bool, hasAnyGroupBys bool) []QueryResult {
	resultantResults := make([]QueryResult, 0)
	for _, result := range results {
		var tmpResult QueryResult
		tmpResult = QueryResult{}

		tmpResult.Headers = getTransformedHeadersForProfileResults(result.Headers, hasGroupByTimestamp, hasAnyGroupBys)
		tmpResult.Rows = getTransformedRowsForProfileResults(result.Rows, hasGroupByTimestamp, hasAnyGroupBys, len(result.Headers))
		resultantResults = append(resultantResults, tmpResult)
	}
	return resultantResults
}

func getTransformedHeadersForProfileResults(headers []string, hasGroupByTimestamp bool, hasAnyGroupBys bool) []string {
	if hasGroupByTimestamp {
		finalResultantHeaders := append(headers[2:], headers[1])
		finalResultantHeaders[0] = TimestampHeader
		return finalResultantHeaders
	} else if !hasGroupByTimestamp && hasAnyGroupBys {
		finalResultantHeaders := append(headers[2:], headers[1])
		return finalResultantHeaders
	} else {
		finalResultantHeaders := []string{headers[1]}
		return finalResultantHeaders
	}
}

func getTransformedRowsForProfileResults(rows [][]interface{}, hasGroupByTimestamp bool, hasAnyGroupBys bool, headersLen int) [][]interface{} {
	var currentRows [][]interface{}
	currentRows = make([][]interface{}, 0)
	if len(rows) == 0 {
		return currentRows
	}

	for _, row := range rows {
		var currentRow []interface{}

		if len(row) == 0 {
			currentRow = make([]interface{}, headersLen)
			for index := range currentRow {
				currentRow[index] = ""
			}
			currentRow[0] = 0
			currentRow[1] = 0
		} else {
			currentRow = row
		}
		if hasGroupByTimestamp || (!hasGroupByTimestamp && hasAnyGroupBys) {
			currentRow = append(currentRow[2:], currentRow[1])
			currentRows = append(currentRows, currentRow)
		} else {
			currentRows = append(currentRows, currentRow[1:])
		}
	}
	return currentRows
}

// Here we are considering only one transformation
func HandlingProfileResultsByApplyingOperations(results []QueryResult, operator string, timezone string, isTimezoneEnabled bool) QueryResult {
	resultKeys := getAllKeysFromResultsArray(results)
	var finalResult QueryResult
	finalResultRows := make([][]interface{}, 0)
	resultKeys = addValuesToHashMap(resultKeys, results[0].Rows)

	for _, row := range results[1].Rows {
		key := U.GetkeyFromRow(row)
		value1 := resultKeys[key]
		value2 := row[len(row)-1]
		result := getValueFromValuesAndOperator(value1, value2, operator)
		resultKeys[key] = result
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

func getAllKeysFromResultsArray(results []QueryResult) map[string]interface{} {
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
