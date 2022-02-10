package memsql

import (
	C "factors/config"
	"factors/model/model"
	"net/http"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// statusCode need to be clear on http.StatusOk or http.StatusAccepted or something else.
// TODO handle errors and kpiFunction statusCode.
func (store *MemSQL) ExecuteKPIQueryGroup(projectID uint64, reqID string, kpiQueryGroup model.KPIQueryGroup) ([]model.QueryResult, int) {
	var queryResults []model.QueryResult
	finalStatusCode := http.StatusOK
	isTimezoneEnabled := false
	kpiTimezoneString := string(kpiQueryGroup.GetTimeZone())
	if C.IsMultipleProjectTimezoneEnabled(projectID) {
		isTimezoneEnabled = true
	}
	for _, query := range kpiQueryGroup.Queries {
		query.Filters = append(query.Filters, kpiQueryGroup.GlobalFilters...)
		query.GroupBy = kpiQueryGroup.GlobalGroupBy
		kpiFunction := store.kpiQueryFunctionDeciderBasedOnCategory(query.Category, query)
		result, statusCode := kpiFunction(projectID, reqID, query)
		if statusCode != http.StatusOK {
			finalStatusCode = statusCode
		}
		queryResults = append(queryResults, result...)
	}
	if finalStatusCode != http.StatusOK {
		log.WithField("kpiQueryGroup", kpiQueryGroup).WithField("queryResults", queryResults).Error("Failed in executing following KPI Query with status Not Ok.")
		return []model.QueryResult{model.QueryResult{}, model.QueryResult{}}, finalStatusCode
	}

	gbtRelatedQueryResults, nonGbtRelatedQueryResults, gbtRelatedQueries, nonGbtRelatedQueries := model.SplitQueryResultsIntoGBTAndNonGBT(queryResults, kpiQueryGroup, finalStatusCode)
	finalQueryResult := make([]model.QueryResult, 0)
	gbtRelatedMergedResults := model.MergeQueryResults(gbtRelatedQueryResults, gbtRelatedQueries, kpiTimezoneString, finalStatusCode, isTimezoneEnabled)
	nonGbtRelatedMergedResults := model.MergeQueryResults(nonGbtRelatedQueryResults, nonGbtRelatedQueries, kpiTimezoneString, finalStatusCode, isTimezoneEnabled)
	if (!reflect.DeepEqual(model.QueryResult{}, gbtRelatedMergedResults)) {
		finalQueryResult = append(finalQueryResult, gbtRelatedMergedResults)
	}
	if (!reflect.DeepEqual(model.QueryResult{}, nonGbtRelatedMergedResults)) {
		finalQueryResult = append(finalQueryResult, nonGbtRelatedMergedResults)
	}
	return finalQueryResult, finalStatusCode
}

func (store *MemSQL) kpiQueryFunctionDeciderBasedOnCategory(category string, query model.KPIQuery) func(uint64, string, model.KPIQuery) ([]model.QueryResult, int) {
	var result func(uint64, string, model.KPIQuery) ([]model.QueryResult, int)
	if category == model.ChannelCategory {
		result = store.ExecuteKPIQueryForChannels
	} else if category == model.EventCategory {
		result = store.ExecuteKPIQueryForEvents
		// Wait and remove by may 2022.
		// else if category == model.EventCategory && !U.ContainsStringInArray([]string{model.HubspotContactsDisplayCategory, model.SalesforceUsersDisplayCategory}, query.DisplayCategory) {
		// 	result = store.ExecuteKPIQueryForEvents
		// } else if U.ContainsStringInArray([]string{model.HubspotContactsDisplayCategory, model.SalesforceUsersDisplayCategory}, query.DisplayCategory) &&
		// 	U.ContainsStringInArray([]string{model.CountOfContactsCreated, model.CountOfContactsUpdated, model.CountOfLeadsCreated, model.CountOfLeadsUpdated}, query.Metrics[0]) {
		// 	result = store.ExecuteKPIQueryForEvents
	} else {
		result = store.ExecuteKPIQueryForProfiles
	}
	return result
}
