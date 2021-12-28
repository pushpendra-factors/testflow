package postgres

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

func (pg *Postgres) RunProfilesGroupQuery(queriesOriginal []model.ProfileQuery, projectID uint64) (model.ResultGroup, int) {
	queries := make([]model.ProfileQuery, 0)
	U.DeepCopy(&queriesOriginal, &queries)

	var resultGroup model.ResultGroup
	resultGroup.Results = make([]model.QueryResult, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(queries), AllowedGoroutines))
	for index, query := range queries {
		count++
		go pg.runSingleProfilesQuery(projectID, query, &resultGroup.Results[index], &waitGroup, index)
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

func (pg *Postgres) runSingleProfilesQuery(projectID uint64, query model.ProfileQuery,
	resultHolder *model.QueryResult, waitGroup *sync.WaitGroup, queryIndex int) {

	defer waitGroup.Done()
	result, errCode, errMsg := pg.ExecuteProfilesQuery(projectID, query)
	if errCode != http.StatusOK {
		errorResult := buildErrorResult(errMsg)
		*resultHolder = *errorResult
	} else {
		model.AddQueryIndexToResult(result, queryIndex)
		*resultHolder = *result
	}
}

func (pg *Postgres) ExecuteProfilesQuery(projectID uint64, query model.ProfileQuery) (*model.QueryResult, int, string) {
	switch query.Type {
	case "all_users":
		return pg.ExecuteAllUsersProfilesQuery(projectID, query)
	default:
		return &model.QueryResult{}, http.StatusBadRequest, "Invalid query type for profiles"
	}
}

func (pg *Postgres) ExecuteAllUsersProfilesQuery(projectID uint64, query model.ProfileQuery) (*model.QueryResult, int, string) {
	query = model.TransformProfilesQuery(query)
	sql, params, err := buildAllUsersQuery(projectID, query)
	if err != nil {
		log.WithError(err).Error(model.ErrMsgQueryProcessingFailure)
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "profiles_query": query,
		"statement": sql, "params": params})
	if sql == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	result, err := pg.ExecQuery(sql, params)
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	err = SanitizeQueryResultProfiles(result, &query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to sanitize query results.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	return result, http.StatusOK, ""
}

func buildAllUsersQuery(projectID uint64, query model.ProfileQuery) (string, []interface{}, error) {
	var params []interface{}
	var groupBySelectParams []interface{}
	selectKeys := make([]string, 0)
	groupByKeys := make([]string, 0)
	groupByStmnt := ""
	selectKeys = append(selectKeys, model.DefaultSelectForAllUsers)

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
	filterStmnt, filterParams, err := buildWhereFromProperties(projectID, query.Filters)
	if filterStmnt != "" {
		filterStmnt = " AND " + filterStmnt
	}
	if err != nil {
		return "", make([]interface{}, 0), err
	}

	allowSupportForSourceColumnInUsers := C.IsProfileQuerySourceSupported(projectID)

	var stepSqlStmnt string
	stepSqlStmnt = fmt.Sprintf("SELECT %s FROM users WHERE project_id = ? %s AND join_timestamp>=? AND join_timestamp<=?", selectStmnt, filterStmnt)
	params = append(params, groupBySelectParams...)
	params = append(params, projectID)
	params = append(params, filterParams...)
	params = append(params, query.From)
	params = append(params, query.To)
	if allowSupportForSourceColumnInUsers && model.IsValidUserSource(query.Source) {
		stepSqlStmnt = fmt.Sprintf("%s AND source=?", stepSqlStmnt)
		params = append(params, query.Source)
	}
	stepSqlStmnt = fmt.Sprintf("%s %s ORDER BY all_users LIMIT 10000", stepSqlStmnt, groupByStmnt)

	finalSQLStmnt := ""
	if isGroupByTypeWithBuckets(query.GroupBys) {
		selectAliases := "all_users"
		sqlStmnt := "WITH step_0 AS (" + stepSqlStmnt + ")"
		bucketedStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys := appendNumericalBucketingSteps(&sqlStmnt, &params, query.GroupBys, "step_0", "", false, selectAliases)
		selectAliases = aggregateSelectKeys + selectAliases
		finalGroupBy := "all_users, " + strings.Join(aggregateGroupBys, ",")
		finalOrderBy := "all_users, " + strings.Join(aggregateOrderBys, ",")
		finalSQLStmnt = fmt.Sprintf("%s SELECT %s FROM %s GROUP BY %s ORDER BY %s LIMIT 1000", sqlStmnt, selectAliases, bucketedStepName, finalGroupBy, finalOrderBy)
	} else {
		finalSQLStmnt = stepSqlStmnt
	}

	return finalSQLStmnt, params, nil
}

func getNoneHandledGroupBySelectForProfiles(projectID uint64, groupProp model.QueryGroupByProperty, groupKey string, timezoneString string) (string, []interface{}) {
	var groupSelect string
	groupSelectParams := make([]interface{}, 0)
	if groupProp.Type != U.PropertyTypeDateTime {
		groupSelect = fmt.Sprintf(`CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' ELSE %s->>? END AS %s`,
			"properties", model.PropertyValueNone, "properties", model.PropertyValueNone, "properties", groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	} else {
		timestampStr := getSelectTimestampByTypeAndPropertyName(groupProp.Granularity, "properties"+"->>?", timezoneString)
		groupSelect = fmt.Sprintf(`CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' WHEN %s->>? = '0' THEN '%s' ELSE %s END AS %s`,
			"properties", model.PropertyValueNone, "properties", model.PropertyValueNone, "properties", model.PropertyValueNone, timestampStr, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property, groupProp.Property}
	}
	return groupSelect, groupSelectParams
}

func SanitizeQueryResultProfiles(result *model.QueryResult, query *model.ProfileQuery) error {

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
