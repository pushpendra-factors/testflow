package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func (store *MemSQL) RunProfilesGroupQuery(queriesOriginal []model.ProfileQuery,
	projectID int64, enableOptimisedFilter bool) (model.ResultGroup, int) {

	logFields := log.Fields{
		"queries_original": queriesOriginal,
		"project_id":       projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	queries := make([]model.ProfileQuery, 0)
	U.DeepCopy(&queriesOriginal, &queries)

	var resultGroup model.ResultGroup
	resultGroup.Results = make([]model.QueryResult, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(queries), AllowedGoroutines))
	for index, query := range queries {
		count++
		go store.runSingleProfilesQuery(projectID, query, &resultGroup.Results[index], &waitGroup, index, enableOptimisedFilter)
		if count%AllowedGoroutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, AllowedGoroutines))
		}
	}
	waitGroup.Wait()
	for _, result := range resultGroup.Results {
		if result.Headers[0] == model.AliasError {
			return resultGroup, http.StatusPartialContent
		}
	}
	return resultGroup, http.StatusOK
}

func (store *MemSQL) runSingleProfilesQuery(projectID int64, query model.ProfileQuery,
	resultHolder *model.QueryResult, waitGroup *sync.WaitGroup, queryIndex int, enableOptimisedFilter bool) {
	logFields := log.Fields{
		"query":         query,
		"project_id":    projectID,
		"result_holder": resultHolder,
		"wait_group":    waitGroup,
		"query_index":   queryIndex,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	defer waitGroup.Done()
	result, errCode, errMsg := store.ExecuteProfilesQuery(projectID, query, enableOptimisedFilter)
	if errCode != http.StatusOK {
		errorResult := buildErrorResult(errMsg)
		*resultHolder = *errorResult
	} else {
		model.AddQueryIndexToResult(result, queryIndex)
		*resultHolder = *result
	}
}

func (store *MemSQL) ExecuteProfilesQuery(projectID int64, query model.ProfileQuery,
	enableOptimisedFilter bool) (*model.QueryResult, int, string) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if model.IsValidUserSource(query.Type) {
		return store.ExecuteAllUsersProfilesQuery(projectID, query, enableOptimisedFilter)
	} else {
		return &model.QueryResult{}, http.StatusBadRequest,
			fmt.Sprintf("Invalid QueryType or GroupName for profiles. QueryType: %s, GroupName: %s", query.Type, query.GroupAnalysis)
	}
}

func (store *MemSQL) ExecuteAllUsersProfilesQuery(projectID int64, query model.ProfileQuery,
	enableOptimisedFilter bool) (*model.QueryResult, int, string) {

	// Profile query with wrap is doing repartitioning and then filtering.
	// If From and To used are of bigger time range, it is repartitioning huge data.
	if (query.To-query.From >= 10*U.SECONDS_IN_YEAR) && enableOptimisedFilter {
		enableOptimisedFilter = false
	}

	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	allowProfilesGroupSupport := C.IsProfileGroupSupportEnabled(projectID)
	if allowProfilesGroupSupport && model.IsValidProfileQueryGroupName(query.GroupAnalysis) && query.GroupAnalysis != model.USERS {
		group, status := store.GetGroup(projectID, query.GroupAnalysis)
		if status == http.StatusFound {
			query.GroupId = group.ID
		}
	}
	query = model.TransformProfilesQuery(query)

	var statement string
	var params []interface{}
	var err error
	if enableOptimisedFilter {
		statement, params, err = buildAllUsersQueryV2(projectID, query)
	} else {
		statement, params, err = buildAllUsersQuery(projectID, query)
	}
	if err != nil {
		log.WithError(err).Error(model.ErrMsgQueryProcessingFailure)
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	logCtx := log.WithFields(logFields)
	if statement == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	result, err, reqID := store.ExecQuery(statement, params)
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	startComputeTime := time.Now()
	err = SanitizeQueryResultProfiles(result, &query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to sanitize query results.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	U.LogComputeTimeWithQueryRequestID(startComputeTime, reqID, &logFields)
	return result, http.StatusOK, ""
}

func buildAllUsersQuery(projectID int64, query model.ProfileQuery) (string, []interface{}, error) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var params []interface{}
	var groupBySelectParams []interface{}
	selectKeys := make([]string, 0)
	groupByKeys := make([]string, 0)
	groupByStmnt := ""
	_, selectKeysForProfile := getSelectKeysForProfile(query, model.USERS)
	selectKeys = append(selectKeys, selectKeysForProfile)

	for _, groupBy := range query.GroupBys {
		gKey := groupKeyByIndex(groupBy.Index)
		groupBySelect, groupByParams := getNoneHandledGroupBySelectForProfiles(projectID, groupBy, gKey, query.Timezone)
		selectKeys = append(selectKeys, groupBySelect)
		groupByKeys = append(groupByKeys, gKey)
		groupBySelectParams = append(groupBySelectParams, groupByParams...)
	}
	if len(groupByKeys) != 0 {
		groupByStmnt = "GROUP BY " + joinWithComma(groupByKeys...)
	}

	selectStmnt := joinWithComma(selectKeys...)
	// Using 0 as profile queries are not time bound. The additional properties table will
	// not be used till we migrate all data and remove timestamp condition.
	filterStmnt, filterParams, err := buildWhereFromProperties(projectID, query.Filters, 0)
	if filterStmnt != "" {
		filterStmnt = " AND " + filterStmnt
	}
	if err != nil {
		return "", make([]interface{}, 0), err
	}

	allowSupportForSourceColumnInUsers := C.IsProfileQuerySourceSupported(projectID)
	allowProfilesGroupSupport := C.IsProfileGroupSupportEnabled(projectID)

	var stepSqlStmnt string
	stepSqlStmnt = fmt.Sprintf(
		"SELECT %s FROM users WHERE users.project_id = ? %s AND join_timestamp>=? AND join_timestamp<=?", selectStmnt, filterStmnt)
	params = append(params, groupBySelectParams...)
	params = append(params, projectID)
	params = append(params, filterParams...)
	params = append(params, query.From)
	params = append(params, query.To)
	if allowSupportForSourceColumnInUsers {
		if model.UserSourceMap[query.Type] == model.UserSourceWeb {
			stepSqlStmnt = fmt.Sprintf("%s AND (source=? OR source IS NULL)", stepSqlStmnt)
		} else {
			stepSqlStmnt = fmt.Sprintf("%s AND source=?", stepSqlStmnt)
		}
		params = append(params, model.GetSourceFromQueryTypeOrGroupName(query))
	}

	if !allowProfilesGroupSupport || (allowProfilesGroupSupport && query.GroupAnalysis == model.USERS) {
		stepSqlStmnt = fmt.Sprintf("%s AND (is_group_user=0 or is_group_user IS NULL)", stepSqlStmnt)
	} else {
		stepSqlStmnt = fmt.Sprintf("%s AND (is_group_user=1 AND group_%d_id IS NOT NULL)", stepSqlStmnt, query.GroupId)
	}

	stepSqlStmnt = fmt.Sprintf("%s %s ORDER BY %s", stepSqlStmnt, groupByStmnt, model.AliasAggr)
	if !query.LimitNotApplicable {
		stepSqlStmnt = fmt.Sprintf("%s LIMIT %d", stepSqlStmnt, model.ProfilesLimit)
	}

	finalSQLStmnt := ""
	if isGroupByTypeWithBuckets(query.GroupBys) {
		selectAliases := model.AliasAggr
		sqlStmnt := "WITH step_0 AS (" + stepSqlStmnt + ")"
		isAggregateOnProperty := false
		if query.AggregateProperty != "" && query.AggregateProperty != "1" {
			isAggregateOnProperty = true
		}
		bucketedStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys := appendNumericalBucketingSteps(isAggregateOnProperty, &sqlStmnt, &params, query.GroupBys, "step_0", "", false, selectAliases)
		selectAliases = selectAliases + ", " + aggregateSelectKeys[:len(aggregateSelectKeys)-2]
		finalGroupBy := model.AliasAggr + ", " + strings.Join(aggregateGroupBys, ",")
		finalOrderBy := model.AliasAggr + ", " + strings.Join(aggregateOrderBys, ",")
		finalSQLStmnt = fmt.Sprintf("%s SELECT %s FROM %s GROUP BY %s ORDER BY %s", sqlStmnt, selectAliases, bucketedStepName, finalGroupBy, finalOrderBy)

		if !query.LimitNotApplicable {
			finalSQLStmnt = fmt.Sprintf("%s LIMIT %d", finalSQLStmnt, model.ProfilesLimit)
		}

	} else {
		finalSQLStmnt = stepSqlStmnt
	}

	return finalSQLStmnt, params, nil
}

// buildAllUsersQueryV2 - buildAllUsersQuery with JSON Filter wrapper support for optimised performance.
func buildAllUsersQueryV2(projectID int64, query model.ProfileQuery) (string, []interface{}, error) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var params []interface{}
	var groupBySelectParams []interface{}
	selectKeys := make([]string, 0)
	groupByKeys := make([]string, 0)
	groupByStmnt := ""
	selectKeys = append(selectKeys)

	for _, groupBy := range query.GroupBys {
		gKey := groupKeyByIndex(groupBy.Index)
		groupBySelect, groupByParams := getNoneHandledGroupBySelectForProfiles(projectID, groupBy, gKey, query.Timezone)
		selectKeys = append(selectKeys, groupBySelect)
		groupByKeys = append(groupByKeys, gKey)
		groupBySelectParams = append(groupBySelectParams, groupByParams...)
	}

	if len(groupByKeys) != 0 {
		groupByStmnt = "GROUP BY " + joinWithComma(groupByKeys...)
	}

	// Primary select as sub query with limit and apply JSON filter on the result.
	wrapperViewName := "wrap_" + U.RandomLowerAphaNumString(5)
	selectKeysOriginaColumns, selectKeysWithAggregate := getSelectKeysForProfile(query, wrapperViewName)
	selectKeys = append(selectKeys, selectKeysOriginaColumns)

	selectStmnt := joinWithComma(selectKeys...)
	// Using 0 as profile queries are not time bound. The additional properties table will
	// not be used till we migrate all data and remove timestamp condition.
	filterStmnt, filterParams, err := buildWhereFromProperties(projectID, query.Filters, 0)
	if err != nil {
		return "", make([]interface{}, 0), err
	}

	allowSupportForSourceColumnInUsers := C.IsProfileQuerySourceSupported(projectID)
	allowProfilesGroupSupport := C.IsProfileGroupSupportEnabled(projectID)

	var stepSqlStmnt string
	stepSqlStmnt = fmt.Sprintf(
		"SELECT %s FROM users WHERE users.project_id = ? AND join_timestamp>=? AND join_timestamp<=?",
		selectStmnt,
	)

	params = append(params, groupBySelectParams...)
	params = append(params, projectID)
	params = append(params, query.From)
	params = append(params, query.To)

	if allowSupportForSourceColumnInUsers {
		if model.UserSourceMap[query.Type] == model.UserSourceWeb {
			stepSqlStmnt = fmt.Sprintf("%s AND (source=? OR source IS NULL)", stepSqlStmnt)
		} else {
			stepSqlStmnt = fmt.Sprintf("%s AND source=?", stepSqlStmnt)
		}
		params = append(params, model.GetSourceFromQueryTypeOrGroupName(query))
	}

	if !allowProfilesGroupSupport || (allowProfilesGroupSupport && query.GroupAnalysis == model.USERS) {
		stepSqlStmnt = fmt.Sprintf("%s AND (is_group_user=0 or is_group_user IS NULL)", stepSqlStmnt)
	} else {
		stepSqlStmnt = fmt.Sprintf("%s AND (is_group_user=1 AND group_%d_id IS NOT NULL)", stepSqlStmnt, query.GroupId)
	}

	wrapperSelectStmnt := joinWithComma(selectKeysWithAggregate, joinWithComma(groupByKeys...))
	wrappedSqlStmnt := fmt.Sprintf("SELECT %s FROM (%s LIMIT %d) %s", wrapperSelectStmnt, stepSqlStmnt, model.FilterOptLimit, wrapperViewName)
	if filterStmnt != "" {
		wrappedSqlStmnt = wrappedSqlStmnt + " WHERE " + strings.ReplaceAll(filterStmnt, "users.", wrapperViewName+".")
	}
	params = append(params, filterParams...)

	wrappedSqlStmnt = fmt.Sprintf("%s %s ORDER BY %s", wrappedSqlStmnt, groupByStmnt, model.AliasAggr)
	if !query.LimitNotApplicable {
		wrappedSqlStmnt = fmt.Sprintf("%s LIMIT %d", wrappedSqlStmnt, model.ProfilesLimit)
	}

	finalSQLStmnt := ""
	if isGroupByTypeWithBuckets(query.GroupBys) {
		selectAliases := model.AliasAggr
		sqlStmnt := "WITH step_0 AS (" + wrappedSqlStmnt + ")"
		isAggregateOnProperty := false
		if query.AggregateProperty != "" && query.AggregateProperty != "1" {
			isAggregateOnProperty = true
		}
		bucketedStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys := appendNumericalBucketingSteps(isAggregateOnProperty, &sqlStmnt, &params, query.GroupBys, "step_0", "", false, selectAliases)
		selectAliases = selectAliases + ", " + aggregateSelectKeys[:len(aggregateSelectKeys)-2]
		finalGroupBy := model.AliasAggr + ", " + strings.Join(aggregateGroupBys, ",")
		finalOrderBy := model.AliasAggr + ", " + strings.Join(aggregateOrderBys, ",")
		finalSQLStmnt = fmt.Sprintf("%s SELECT %s FROM %s GROUP BY %s ORDER BY %s", sqlStmnt, selectAliases, bucketedStepName, finalGroupBy, finalOrderBy)
		if !query.LimitNotApplicable {
			wrappedSqlStmnt = fmt.Sprintf("%s LIMIT %d", wrappedSqlStmnt, model.ProfilesLimit)
		}
	} else {
		finalSQLStmnt = wrappedSqlStmnt
	}

	return finalSQLStmnt, params, nil
}

func getSelectKeysForProfile(query model.ProfileQuery, tableViewName string) (string, string) {
	if query.AggregateProperty == "1" || query.AggregateProperty == "" || query.AggregateFunction == model.UniqueAggregateFunction { // Generally count is only used againt them.
		return "users.customer_user_id, users.id, users.properties", fmt.Sprintf("COUNT(DISTINCT(COALESCE(%s.customer_user_id, %s.id))) as %s", tableViewName, tableViewName, model.AliasAggr)
	} else {
		return "users.properties", fmt.Sprintf("%s(CASE WHEN JSON_EXTRACT_STRING(%s.properties, '%s') IS NULL THEN 0 ELSE JSON_EXTRACT_STRING(%s.properties, '%s') END ) as %s", query.AggregateFunction,
			tableViewName, query.AggregateProperty, tableViewName, query.AggregateProperty, model.AliasAggr)
	}
}

func getNoneHandledGroupBySelectForProfiles(projectID int64, groupProp model.QueryGroupByProperty, groupKey string, timezoneString string) (string, []interface{}) {
	logFields := log.Fields{
		"group_prop":      groupProp,
		"project_id":      projectID,
		"group_key":       groupKey,
		"timezone_string": timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var groupSelect string
	groupSelectParams := make([]interface{}, 0)
	if groupProp.Property == U.UP_USER_ID {
		groupSelect = fmt.Sprintf("COALESCE(customer_user_id, id) AS %s", groupKey)
	} else if groupProp.Type != U.PropertyTypeDateTime {
		groupSelect = fmt.Sprintf("CASE WHEN JSON_EXTRACT_STRING(%s, ?) IS NULL THEN '%s' WHEN JSON_EXTRACT_STRING(%s, ?) = '' THEN '%s' ELSE JSON_EXTRACT_STRING(%s, ?) END AS %s",
			"properties", model.PropertyValueNone, "properties", model.PropertyValueNone, "properties", groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	} else {
		propertyName := "JSON_EXTRACT_STRING(" + "properties" + ", ?)"
		timestampStr := getSelectTimestampByTypeAndPropertyName(groupProp.Granularity, propertyName, timezoneString)
		groupSelect = fmt.Sprintf("CASE WHEN JSON_EXTRACT_STRING(%s, ?) IS NULL THEN '%s' WHEN JSON_EXTRACT_STRING(%s, ?) = '' THEN '%s' WHEN JSON_EXTRACT_STRING(%s, ?) = '0' THEN '%s' ELSE %s END AS %s",
			"properties", model.PropertyValueNone, "properties", model.PropertyValueNone, "properties", model.PropertyValueNone, timestampStr, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property, groupProp.Property}
	}
	return groupSelect, groupSelectParams
}

func SanitizeQueryResultProfiles(result *model.QueryResult, query *model.ProfileQuery) error {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// Replace group keys with real column names. should be last step.
	// of sanitization.
	if err := translateGroupKeysIntoColumnNames(result, query.GroupBys); err != nil {
		return err
	}

	if isGroupByTypeWithBuckets(query.GroupBys) {
		sanitizeNumericalBucketRangesProfiles(result, query)
	}

	if model.HasGroupByDateTypeProperties(query.GroupBys) {
		model.SanitizeDateTypeRowsProfiles(result, query)
	}
	return nil
}

func sanitizeNumericalBucketRangesProfiles(result *model.QueryResult, query *model.ProfileQuery) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	headerIndexMap := make(map[string][]int)
	for index, header := range result.Headers {
		// If same group by is added twice, it will appear twice in headers.
		// Keep as a list to sanitize both indexes.
		headerIndexMap[header] = append(headerIndexMap[header], index)
	}

	sanitizedProperties := make(map[string]bool)
	for _, gbp := range query.GroupBys {
		if isGroupByTypeWithBuckets([]model.QueryGroupByProperty{gbp}) {
			if _, sanitizedAlready := sanitizedProperties[gbp.Property]; sanitizedAlready {
				continue
			}
			indexesToSanitize := headerIndexMap[gbp.Property]
			for _, indexToSanitize := range indexesToSanitize {
				sanitizeNumericalBucketRangeProfiles(query, result.Rows, indexToSanitize)
			}
			sanitizedProperties[gbp.Property] = true
		}
	}
}

func sanitizeNumericalBucketRangeProfiles(query *model.ProfileQuery, rows [][]interface{}, indexToSanitize int) {
	logFields := log.Fields{
		"rows":              rows,
		"query":             query,
		"index_to_sanitize": indexToSanitize,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, row := range rows {
		// Remove trailing .0 in start and end value of range.
		row[indexToSanitize] = trailingZeroRegex.ReplaceAllString(row[indexToSanitize].(string), "")

		// Change range with same start and end ex: 2 - 2 to just 2.
		if row[indexToSanitize] != model.PropertyValueNone {
			rowSplit := strings.Split(row[indexToSanitize].(string), model.NumericalGroupBySeparator)
			if rowSplit[0] == rowSplit[1] {
				row[indexToSanitize] = model.GetBucketRangeForStartAndEnd(rowSplit[0], rowSplit[1])
			}
		}
	}
}
