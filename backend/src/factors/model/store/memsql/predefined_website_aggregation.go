package memsql

import (
	"encoding/json"
	// b64 "encoding/base64"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) IsPredefWebAggDashboardExists(projectID int64, agentUUID string) bool {
	return store.existsDashboardByInternalID(projectID, 1)
}

func (store *MemSQL) CreatePredefinedWebsiteAggregation(projectID int64, agentUUID string) int {

	predefinedDashboard := model.PredefinedDashboards[0]

	currentPositionBytes, err := json.Marshal(model.PredefinedDashboardUnitsPosition)
	if err != nil {
		log.WithField("projectID", projectID).WithError(err).Error("Failed to JSON encode new units position.")
		return http.StatusInternalServerError
	}
	currentPositionJsonb := &postgres.Jsonb{RawMessage: json.RawMessage(currentPositionBytes)}

	dashboard := model.Dashboard{
		ProjectId:     projectID,
		InternalID:    1,
		AgentUUID:     agentUUID,
		Name:          predefinedDashboard.Name,
		Description:   predefinedDashboard.DisplayName,
		Type:          model.DashboardTypeProjectVisible,
		Class:         model.DashboardClassPredefined,
		UnitsPosition: currentPositionJsonb,
	}

	_, statusCode := store.CreateDashboard(projectID, agentUUID, &dashboard)
	if statusCode != http.StatusCreated {
		log.WithField("projectID", projectID).WithField("internalID", 1).Warn("Failed in predefined dashboard: website aggregation")
	}
	return statusCode
}

func (store *MemSQL) ExecuteQueryGroupForPredefinedWebsiteAggregation(projectID int64, request model.PredefWebsiteAggregationQueryGroup, useTestTable bool) ([]model.QueryResult, int, string) {
	resResults := make([]model.QueryResult, len(request.Queries))
	for index, query := range request.Queries {
		result, statusCode, errMsg := store.ExecuteSingleWebAggregationQuery(projectID, query, useTestTable)
		resResults[index] = result
		if statusCode != http.StatusOK {
			return resResults, statusCode, errMsg
		}
	}

	return resResults, http.StatusOK, ""
}

// Check if SanitizeQueryResult is required.
// I think group by doesnt have index. We can proceed with normal it normally.
// Eventname double encoding. Required later.
// IMP - For the query with no groupby, we are expecting only one metric.
func (store *MemSQL) ExecuteSingleWebAggregationQuery(projectID int64, q model.PredefWebsiteAggregationQuery, useTestTable bool) (model.QueryResult, int, string) {

	var dupQ model.PredefWebsiteAggregationQuery
	U.DeepCopy(&q, &dupQ)
	dupQ.From = U.ConvertEqualTimeFromOtherTimezoneToUTC(q.From, U.TimeZoneString(q.Timezone))
	dupQ.To = U.ConvertEqualTimeFromOtherTimezoneToUTC(q.To, U.TimeZoneString(q.Timezone))
	dupQ.Timezone = "UTC"

	stmnt, params := buildPredefinedWebsiteAggregationQuery(projectID, dupQ, useTestTable)
	result, err, reqID := store.ExecQuery(stmnt, params)
	if err != nil {
		log.WithField("projectID", projectID).WithField("query", q).WithField("reqID", reqID).
			WithError(err).Error("Failed during predefined query execute.")
		return model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	transformedResult := transformPWAResultMetrics(dupQ, *result)
	if dupQ.GetGroupByTimestamp() == "" {
		if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
			emptyRow := make([]interface{}, 0)

			if (q.GroupBy != model.PredefinedGroupBy{}) {
				emptyRow = append(emptyRow, model.PropertyValueNone)
			}
			for range q.Metrics {
				emptyRow = append(emptyRow, 0)
			}
			emptyRows := make([][]interface{}, 1)
			emptyRows[0] = emptyRow
			transformedResult.Rows = emptyRows
		}
	} else {
		transformedResult = transformPWAResultDateValuesToInt(*result)
		transformedResult = addMissingColumnsAndTimestampPWAResult(transformedResult, dupQ)

		transformedResult = transformPWAResultDateValuesToDateFormat(transformedResult, q.Timezone)
	}

	transformedResult.Query = q
	return transformedResult, http.StatusOK, "Successfully executed."
}

func buildPredefinedWebsiteAggregationQuery(projectID int64, q model.PredefWebsiteAggregationQuery, useTestTable bool) (string, []interface{}) {

	internalEventType := q.InternalEventType
	resMetricsToTransformations := make(map[string][]model.PredefWebsiteAggregationMetricTransform)
	resMetrics := make([]string, 0)
	for _, metric := range q.Metrics {
		resMetrics = append(resMetrics, metric.Name)
		transformation := model.PredefWebMetricToInternalTransformations[metric.Name]
		resMetricsToTransformations[metric.Name] = transformation
	}

	resGroupBy := ""
	internalGroupBy := ""
	if internalGroupKey, exists := model.MapOfPredefWebsiteAggregGroupByExternalToInternal[q.GroupBy.Name]; exists {
		resGroupBy = q.GroupBy.Name
		internalGroupBy = internalGroupKey
	} else {
		resGroupBy = q.GroupBy.Name
		internalGroupBy = q.GroupBy.Name
	}

	resSelectStatement := getPredefinedSelectStmnt(resMetrics, resMetricsToTransformations, resGroupBy, internalGroupBy, q.GetGroupByTimestamp())

	resFromStmnt := ""
	if useTestTable {
		resFromStmnt = fmt.Sprintf("%s %s ", model.DBFrom, model.WebsiteAggregationTable)
	} else {
		resFromStmnt = fmt.Sprintf("%s %s ", model.DBFrom, model.WebsiteAggregationTestTable)
	}
	resFromStmnt = fmt.Sprintf("%s %s ", model.DBFrom, model.WebsiteAggregationTable)

	resFilterStmnt, resFilterParams := getPredefinedFilterStmnt(projectID, q.Filters, internalEventType, q.From, q.To, q.Timezone)

	resGroupByStmnt := getGroupByStmnt(resGroupBy, q.GetGroupByTimestamp())

	resOrderByStmnt := fmt.Sprintf("%s %s DESC ", model.DBOrderBy, joinWithComma(resMetrics...))

	resLimitStmnt := fmt.Sprintf("%s %v", model.DBLimit, model.PredefinedWebAggrLimit)

	resStmnt := resSelectStatement + resFromStmnt + resFilterStmnt + resGroupByStmnt + resOrderByStmnt + resLimitStmnt

	resParams := resFilterParams

	return resStmnt, resParams
}

func getPredefSelectTimestampByType(groupByTimestamp string) string {
	var groupByTimestampSql string
	if groupByTimestamp == model.GroupByTimestampWeek {
		groupByTimestampSql = DateTruncateInURlForWeekPWA
	} else {
		if dateTrunVal, ok := mapOfGroupByTimestampToDateTrunc[groupByTimestamp]; ok {
			groupByTimestampSql = fmt.Sprintf(DateTruncateInURlPWA, dateTrunVal)
		} else {
			groupByTimestampSql = fmt.Sprintf(DateTruncateInURlPWA, groupByTimestamp)
		}
	}
	return groupByTimestampSql
}

// This is used to form all transformations and groupBy. This cant be reusable.
func getPredefinedSelectStmnt(metrics []string, metricsToTransformations map[string][]model.PredefWebsiteAggregationMetricTransform,
	externalGroupBy, internalGroupBy, groupByTimestamp string) string {

	resSelectStmnt := model.DBSelect
	SelectExpressions := make([]string, 0)

	if groupByTimestamp != "" {
		groupByTimestampSql := getPredefSelectTimestampByType(groupByTimestamp)
		SelectExpressions = append(SelectExpressions, fmt.Sprintf("%s as %s", groupByTimestampSql, model.AliasDateTime))
	}

	if internalGroupBy != "" {
		SelectExpressions = append(SelectExpressions, fmt.Sprintf("%s as %s", internalGroupBy, externalGroupBy))
	}
	for _, metric := range metrics {
		transformations := metricsToTransformations[metric]
		currentExpression := ""
		for index, transformation := range transformations {
			if index == 0 {
				currentExpression = "("
				currentExpression += fmt.Sprintf("CAST(%s(%s) AS UNSIGNED) ", transformation.Operation, transformation.InternalProperty)
				expression := model.MapOfOperatorToExpression[transformation.ArithmeticOperator]
				currentExpression += expression + " "
			} else {
				currentExpression += fmt.Sprintf("CAST(%s(%s) AS UNSIGNED) ", transformation.Operation, transformation.InternalProperty)
				expression := model.MapOfOperatorToExpression[transformation.ArithmeticOperator]
				currentExpression += expression + " "
			}
		}
		currentExpression += fmt.Sprintf(") as %s ", metric)

		SelectExpressions = append(SelectExpressions, currentExpression)
	}
	resSelectStmnt += joinWithComma(SelectExpressions...)

	return resSelectStmnt
}

func getPredefinedFilterStmnt(projectID int64, filters []model.PredefinedFilter, internalEventType string, from, to int64, timezone string) (string, []interface{}) {

	rStmnt := ""
	rParams := make([]interface{}, 0)

	whereStmnt := model.DBWhere + fmt.Sprintf("project_id = %v AND event_type = '%v' AND timestamp_at_day BETWEEN %v AND %v ",
		projectID, internalEventType, from, to)

	groupedFilters := getPredefinedFiltersGrouped(filters)
	for grpIndex, currentFilterGrp := range groupedFilters {
		var currentGrpStmnt, pStmnt string
		for filterIndex, filter := range currentFilterGrp {

			if filter.LogicalOp == "" {
				filter.LogicalOp = "AND"
			}
			pStmnt = ""
			propertyOp := getOp(filter.Condition, "categorical")
			pValue := filter.Value

			pStmnt = fmt.Sprintf("%s %s ?", filter.PropertyName, propertyOp)
			rParams = append(rParams, pValue)

			if filterIndex == 0 {
				currentGrpStmnt = pStmnt
			} else {
				currentGrpStmnt = fmt.Sprintf("%s %s %s", currentGrpStmnt, filter.LogicalOp, pStmnt)
			}
		}
		if grpIndex == 0 {
			rStmnt = fmt.Sprintf("(%s)", currentGrpStmnt)
		} else {
			rStmnt = fmt.Sprintf("%s AND (%s)", rStmnt, currentGrpStmnt)
		}
	}
	rStmnt = joinWithWordInBetween("AND", whereStmnt, rStmnt)
	rStmnt += " "
	return rStmnt, rParams
}

func getPredefinedFiltersGrouped(filters []model.PredefinedFilter) [][]model.PredefinedFilter {
	groups := make([][]model.PredefinedFilter, 0)
	currentGroup := make([]model.PredefinedFilter, 0)

	for index, filter := range filters {
		if index == 0 || filter.LogicalOp != "AND" {
			currentGroup = append(currentGroup, filter)
		} else {
			groups = append(groups, currentGroup)

			currentGroup = make([]model.PredefinedFilter, 0)
			currentGroup = append(currentGroup, filter)
		}
	}

	if len(currentGroup) != 0 {
		groups = append(groups, currentGroup)
	}
	return groups
}

func getGroupByStmnt(groupBy, groupByTimestamp string) string {
	groupBys := make([]string, 0)
	if groupByTimestamp != "" {
		groupBys = append(groupBys, model.AliasDateTime)
	}
	if groupBy != "" {
		groupBys = append(groupBys, groupBy)
	}
	if len(groupBys) == 0 {
		return ""
	}
	groupByString := joinWithComma(groupBys...)
	return fmt.Sprintf("%s %s ", model.DBGroupByConst, groupByString)
}

// Predefined Website Aggregation Result metrics are in string. Converting to int.
func transformPWAResultMetrics(q model.PredefWebsiteAggregationQuery, result model.QueryResult) model.QueryResult {
	if len(result.Rows) == 0 {
		return result
	}
	rowLen := len(result.Rows[0])
	for rowIndex, _ := range result.Rows {
		for metricIndex, _ := range q.Metrics {
			value := result.Rows[rowIndex][rowLen-1-metricIndex]
			switch valueType := value.(type) {
			case float64:
			case string:
				valueInString := value.(string)
				result.Rows[rowIndex][rowLen-1-metricIndex] = U.SafeConvertToFloat64(valueInString)
			default:
				log.WithField("value", value).Info("Unsupported type used on GetSortWeightFromAnyType %+v", valueType)
				result.Rows[rowIndex][rowLen-1-metricIndex] = 0
			}
		}
	}
	return result
}

func transformPWAResultDateValuesToInt(result model.QueryResult) model.QueryResult {

	for rowIndex, row := range result.Rows {
		var valueInInt int64
		switch row[0].(type) {
		case float64:
			valueInInt = int64(row[0].(float64))
		case string:
			valueInString := row[0].(string)
			valueInInt = int64(U.SafeConvertToFloat64(valueInString))
		}
		result.Rows[rowIndex][0] = valueInInt
	}
	return result
}

func transformPWAResultDateValuesToDateFormat(result model.QueryResult, timezone string) model.QueryResult {
	for rowIndex, row := range result.Rows {
		valueInInt := row[0].(int64)
		valueInUTC := time.Unix(valueInInt, 0).UTC()
		valueWithOffset := U.GetTimestampAsStrWithTimezone(valueInUTC, timezone)
		result.Rows[rowIndex][0] = valueWithOffset
	}
	return result
}

// if represents state when there is no group by and else represents state where there is group by.
// Number of metrics in result can be multiple. We cant assume a single. Number of max groupby is 1.
func addMissingColumnsAndTimestampPWAResult(result model.QueryResult, query model.PredefWebsiteAggregationQuery) model.QueryResult {
	resultantRows := make([][]interface{}, 0)

	if (query.GroupBy == model.PredefinedGroupBy{}) {
		mapOfAllColumValuesToResult := make(map[string][]interface{})
		for _, row := range result.Rows {

			key := U.GetkeyFromRowTillEnd(row, 1)

			mapOfAllColumValuesToResult[key] = row
		}

		timestampsInTime, _ := GetAllTimestampsAndOffsetBetweenByType(query.From, query.To,
			query.GetGroupByTimestamp(), query.Timezone)

		for _, timestampInTime := range timestampsInTime {
			timestampInEpoch := timestampInTime.Unix()
			timestampKey := fmt.Sprintf("%v:;", timestampInEpoch)
			if row, exists := mapOfAllColumValuesToResult[timestampKey]; exists {
				resultantRows = append(resultantRows, row)
			} else {
				emptyRow := make([]interface{}, 1)
				emptyRow[0] = timestampInEpoch
				for range query.Metrics {
					emptyRow = append(emptyRow, 0)
				}
				resultantRows = append(resultantRows, emptyRow)
			}
		}
	} else {
		mapOfGroup := make(map[string]interface{})
		mapOfAllColumValuesToResult := make(map[string][]interface{})
		groupIndex := 1

		for _, row := range result.Rows {
			groupKey := fmt.Sprintf("%v:;", row[groupIndex])

			key := U.GetkeyFromRowTillEnd(row, 2)
			mapOfGroup[groupKey] = row[groupIndex]
			mapOfAllColumValuesToResult[key] = row
		}

		timestampsInTime, _ := GetAllTimestampsAndOffsetBetweenByType(query.From, query.To,
			query.GroupByTimestamp, query.Timezone)

		for _, timestampInTime := range timestampsInTime {
			timestampInEpoch := timestampInTime.Unix()
			timestampKey := fmt.Sprintf("%v", timestampInEpoch)

			if len(mapOfGroup) == 0 {
				emptyRow := make([]interface{}, 1)
				emptyRow[0] = timestampInEpoch
				emptyRow = append(emptyRow, model.PropertyValueNone)
				for range query.Metrics {
					emptyRow = append(emptyRow, 0)
				}
				resultantRows = append(resultantRows, emptyRow)
			} else {
				for groupKey, groupValue := range mapOfGroup {
					key := fmt.Sprintf("%v:;%v", timestampKey, groupKey)
					if row, exists := mapOfAllColumValuesToResult[key]; exists {
						resultantRows = append(resultantRows, row)
					} else {
						emptyRow := make([]interface{}, 1)
						emptyRow[0] = timestampInEpoch
						emptyRow = append(emptyRow, groupValue)
						for range query.Metrics {
							emptyRow = append(emptyRow, 0)
						}
						resultantRows = append(resultantRows, emptyRow)
					}
				}
			}
		}
	}

	result.Rows = resultantRows
	return result
}
