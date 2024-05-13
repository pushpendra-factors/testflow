package memsql

import (
	"factors/model/model"
	U "factors/util"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Currently keeping it to only KPI and statically assigning few values.
func (store *MemSQL) ExecuteWidgetGroup(projectID int64, widgetGroup model.WidgetGroup, segmentID string, reqID string, requestParams model.RequestSegmentKPI) ([]model.QueryResult, int) {
	results := make([]model.QueryResult, len(widgetGroup.DecodedWidgets))
	finalStatusCode := http.StatusOK

	segment, statusCode := store.GetSegmentById(projectID, segmentID)
	if statusCode != http.StatusFound {
		return results, http.StatusBadRequest
	}

	lastRunTime, lastRunStatusCode := store.GetMarkerLastForAllAccounts(projectID)
	// for case - segment is updated but all_run for the day is yet to run
	if lastRunStatusCode != http.StatusFound || segment.UpdatedAt.After(lastRunTime) {
		return results, http.StatusUnprocessableEntity
	}

	for index, widget := range widgetGroup.DecodedWidgets {

		if widget.QueryType == model.QueryClassKPI {
			cKPIQueryGroup, errMsg, statusCode := store.buildKPIQuery(projectID, widget, segmentID, requestParams)
			if statusCode != http.StatusOK {
				log.WithField("widget", widget).WithField("errMsg", errMsg).Warn("Failed in executing widget group")
				results[index] = model.QueryResult{}
				continue
			}
			kpiResult, statusCode := store.ExecuteKPIQueryGroup(projectID, reqID, cKPIQueryGroup, true, true)
			if statusCode != http.StatusOK {
				results[index] = model.QueryResult{}
				finalStatusCode = statusCode
			} else {
				results[index].Rows = kpiResult[1].Rows
				results[index].Headers = kpiResult[1].Headers
			}

		} else if widget.QueryType == model.QueryClassAccounts {
			accountAnalyticsQuery := store.BuildAccountAnalytics(projectID, widget, segmentID, requestParams)
			analyticsResult, statusCode := store.ExecuteAccountAnalyticsQuery(projectID, reqID, accountAnalyticsQuery)
			if statusCode != http.StatusOK {
				results[index] = model.QueryResult{}
				finalStatusCode = statusCode
			} else {
				results[index].Rows = analyticsResult.Rows
				results[index].Headers = analyticsResult.Headers
			}
		}
	}

	return results, finalStatusCode
}

func (store *MemSQL) buildKPIQuery(projectID int64, widget model.Widget, segmentID string, requestParams model.RequestSegmentKPI) (model.KPIQueryGroup, string, int) {

	kpiQueryGroup := model.KPIQueryGroup{}
	kpiQueryGroup.Class = model.QueryClassKPI
	kpiQueryGroup.GlobalFilters = make([]model.KPIFilter, 0)
	kpiQueryGroup.GlobalGroupBy = make([]model.KPIGroupBy, 0)
	kpiQueryGroup.SegmentID = segmentID

	customMetric, errMsg, statusCode := store.GetKpiRelatedCustomMetricsByName(projectID, widget.QueryMetric)
	if statusCode != http.StatusFound {
		return kpiQueryGroup, errMsg, statusCode
	}
	kpiQueryGroup.DisplayResultAs = customMetric.DisplayResultAs

	kpiQuery := model.KPIQuery{}
	kpiQuery.Category = model.ProfileCategory
	kpiQuery.DisplayCategory = customMetric.SectionDisplayCategory
	kpiQuery.Metrics = []string{widget.QueryMetric}
	kpiQuery.From = requestParams.From
	kpiQuery.To = requestParams.To
	kpiQuery.GroupByTimestamp = model.GroupByTimestampQuarter
	kpiQuery.Timezone = requestParams.Timezone
	kpiQuery.SegmentID = segmentID

	if customMetric.TypeOfQuery == model.DerivedQueryType {
		kpiQuery.QueryType = "derived"
	} else {
		kpiQuery.QueryType = "custom"
	}

	kpiQuery1 := model.KPIQuery{}
	U.DeepCopy(&kpiQuery, &kpiQuery1)
	kpiQuery1.GroupByTimestamp = ""

	kpiQueryGroup.Queries = []model.KPIQuery{kpiQuery, kpiQuery1}

	return kpiQueryGroup, "", http.StatusOK
}

// TODO LATER
// var MapOfAccountAnalyticsMetricToExpression = map[string][]string{
// 	TotalAccountsMetric:       []string{model.CountAggregateFunction, "1"},
// 	HighEngagedAccountsMetric: []string{model.CountAggregateFunction, "1"},
// }

func (store *MemSQL) BuildAccountAnalytics(projectID int64, widget model.Widget, segmentID string, requestParams model.RequestSegmentKPI) model.AccountAnalyticsQuery {

	filters, _ := model.MapOfAccountAnalyticsMetricToFilters[widget.QueryMetric]
	analyticsQuery := model.AccountAnalyticsQuery{
		AggregateFunction: model.CountAggregateFunction,
		AggregateProperty: "1",
		Metric:            widget.QueryMetric,
		From:              requestParams.From,
		To:                requestParams.To,
		Timezone:          requestParams.Timezone,
		Filters:           filters,
		SegmentID:         segmentID,
	}
	return analyticsQuery
}
