package model

import (
	"errors"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	MetaEachEventCountMetrics = "EachEventCount"
	AllowedGoroutines         = 4
)

type ResultGroup struct {
	Results []QueryResult `json:"result_group"`
}

func RunEventsGroupQuery(queries []Query, projectId uint64) (ResultGroup, int) {

	var resultGroup ResultGroup
	resultGroup.Results = make([]QueryResult, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(queries), AllowedGoroutines))
	for index, query := range queries {
		count++
		go runSingleEventsQuery(projectId, query, &resultGroup.Results[index], &waitGroup)
		if count%AllowedGoroutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, AllowedGoroutines))
		}
	}
	waitGroup.Wait()
	return resultGroup, http.StatusOK
}

func runSingleEventsQuery(projectId uint64, query Query, resultHolder *QueryResult, waitGroup *sync.WaitGroup) {

	defer waitGroup.Done()
	result, errCode, errMsg := ExecuteEventsQuery(projectId, query)
	if errCode != http.StatusOK {
		errorResult := buildErrorResult(errMsg)
		*resultHolder = *errorResult
	} else {
		*resultHolder = *result
	}
}

func ExecuteEventsQuery(projectId uint64, query Query) (*QueryResult, int, string) {

	if valid, errMsg := IsValidEventsQuery(&query); !valid {
		return nil, http.StatusBadRequest, errMsg
	}
	return RunInsightsQuery(projectId, query)
}

func RunInsightsQuery(projectId uint64, query Query) (*QueryResult, int, string) {
	stmnt, params, err := BuildInsightsQuery(projectId, query)
	if err != nil {
		log.WithError(err).Error(ErrMsgQueryProcessingFailure)
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	logCtx := log.WithFields(log.Fields{"analytics_query": query,
		"statement": stmnt, "params": params})
	if stmnt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}
	result, err := ExecQuery(stmnt, params)
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	groupPropsLen := len(query.GroupByProperties)

	err = LimitQueryResult(result, groupPropsLen, query.GetGroupByTimestamp() != "")
	if err != nil {
		logCtx.WithError(err).Error("Failed processing query results for limiting.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	err = SanitizeQueryResult(result, &query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to sanitize query results.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	// if and only if breakdown is by datetime (condition for both events/users count for each event.)
	if len(query.GroupByProperties) == 0 &&
		query.EventsCondition == EventCondEachGivenEvent &&
		query.GroupByTimestamp != nil && query.GroupByTimestamp.(string) != "" {

		result, err = transformResultsForEachEventQuery(result, query)
		addEventMetricsMetaToQueryResult(result)

		if err != nil {
			logCtx.WithError(err).Error("Failed to transform query results.")
			return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
		}
	} else if query.EventsCondition == EventCondEachGivenEvent &&
		!strings.Contains(strings.Join(result.Headers, ","), AliasEventName) {
		// for data consistency: single event result has no event_name column, appending here
		if len(query.EventsWithProperties) == 1 {
			result.Headers = append(result.Headers, AliasEventName)
			for i, _ := range result.Rows {
				result.Rows[i] = append(result.Rows[i], query.EventsWithProperties[0].Name)
			}
		}
	}

	addQueryToResultMeta(result, query)

	return result, http.StatusOK, "Successfully executed query"
}

// buildErrorResult takes the failure msg and wraps it into a QueryResult object
func buildErrorResult(errMsg string) *QueryResult {
	errMsg = "Query failed:" + " - " + errMsg
	headers := []string{AliasError}
	rows := make([][]interface{}, 0, 0)
	row := make([]interface{}, 0, 0)
	row = append(row, errMsg)
	rows = append(rows, row)
	errorResult := &QueryResult{Headers: headers, Rows: rows}
	return errorResult
}

// transformResultsForEachEventQuery transforms QueryResult with new header as datetime and events
func transformResultsForEachEventQuery(oldResult *QueryResult, query Query) (*QueryResult, error) {

	// for single event, oldResult has no 'event_name' column but only 2 columns: {datetime, count}
	// adding  'event_name' with row values for standard transformation
	if len(oldResult.Headers) == 2 {
		if len(query.EventsWithProperties) > 0 {
			oldResult.Headers = append(oldResult.Headers, AliasEventName)
			for i, _ := range oldResult.Rows {
				oldResult.Rows[i] = append(oldResult.Rows[i], query.EventsWithProperties[0].Name)
			}
		}
	}
	eventNameIndex := 0
	dateIndex := 0
	countIndex := 0
	for i, key := range oldResult.Headers {
		if key == AliasEventName {
			eventNameIndex = i
		}
		if key == AliasDateTime {
			dateIndex = i
		}
		if key == AliasAggr {
			countIndex = i
		}
	}

	eventsHeaderIndexMap := make(map[string]int)
	for _, row := range oldResult.Rows {
		eventName := ""
		if row[eventNameIndex] == nil && len(query.EventsWithProperties) == 1 {
			eventName = query.EventsWithProperties[0].Name
		} else {
			eventName = row[eventNameIndex].(string)
		}
		// initial header value = 1
		eventsHeaderIndexMap[eventName] = 1
	}

	// headers : datetime, event1, event2, ...
	newResultHeaders := []string{AliasDateTime}
	newResultRows := make([][]interface{}, 0, 0)

	// skipping 0 as it is index is for 'datetime' header
	headerIndex := 1
	for name, _ := range eventsHeaderIndexMap {
		newResultHeaders = append(newResultHeaders, name)
		eventsHeaderIndexMap[name] = headerIndex
		headerIndex++
	}

	datetimeToNewResultRowNoMap := make(map[int64]int)
	datetimeEncountered := make(map[int64]bool)
	rowNo := 0
	for _, row := range oldResult.Rows {
		// headers : datetime, event1, event2, ...
		dateTime := row[dateIndex].(time.Time).Unix()
		eventName := ""
		if row[eventNameIndex] == nil && len(query.EventsWithProperties) == 1 {
			eventName = query.EventsWithProperties[0].Name
		} else {
			eventName = row[eventNameIndex].(string)
		}
		eventNameHeaderIndex := eventsHeaderIndexMap[eventName]

		if _, exists := datetimeEncountered[dateTime]; exists && datetimeEncountered[dateTime] == true {
			newResultRowNo := datetimeToNewResultRowNoMap[dateTime]
			newResultRows[newResultRowNo][eventNameHeaderIndex] = row[countIndex]
		} else {
			newRow := make([]interface{}, len(newResultHeaders), len(newResultHeaders))
			newRow[0] = row[dateIndex]
			newRow[eventNameHeaderIndex] = row[countIndex]
			newResultRows = append(newResultRows, newRow)
			datetimeToNewResultRowNoMap[dateTime] = rowNo
			datetimeEncountered[dateTime] = true
			rowNo++
		}
	}

	newResult := &QueryResult{Headers: newResultHeaders, Rows: newResultRows}

	return newResult, nil
}

// addEventMetricsMetaToQueryResult adds meta metrics in query result based on query type, event
// condition and group by inputs
func addEventMetricsMetaToQueryResult(result *QueryResult) {

	metaMetricsEventCount := HeaderRows{}
	metaMetricsEventCount.Title = MetaEachEventCountMetrics
	// headers : event1, event2, ...
	metaMetricsEventCount.Headers = []string{}
	headerIndexToEventName := make(map[int]string)
	for i, key := range result.Headers {
		// skipping datetime and including all event names as header
		if key != AliasDateTime {
			metaMetricsEventCount.Headers = append(metaMetricsEventCount.Headers, key)
			headerIndexToEventName[i] = key
		}
	}
	eventCount := make(map[string]int64)
	for _, row := range result.Rows {
		for i := 1; i < len(result.Headers); i++ {
			count, ok := row[i].(int64)
			if !ok {
				count = 0 // 0s are stored as .(int)
			}
			eventCount[headerIndexToEventName[i]] = eventCount[headerIndexToEventName[i]] + count
		}
	}
	rows := make([][]interface{}, 0, 0)
	singleCountRow := make([]interface{}, len(metaMetricsEventCount.Headers), len(metaMetricsEventCount.Headers))
	for i := 0; i < len(metaMetricsEventCount.Headers); i++ {
		singleCountRow[i] = eventCount[metaMetricsEventCount.Headers[i]]
	}
	rows = append(rows, singleCountRow)

	metaMetricsEventCount.Rows = rows
	result.Meta.MetaMetrics = append(result.Meta.MetaMetrics, metaMetricsEventCount)
}

// BuildInsightsQuery - Dispatches corresponding build method for insights.
func BuildInsightsQuery(projectId uint64, query Query) (string, []interface{}, error) {
	addIndexToGroupByProperties(&query)

	if query.Type == QueryTypeEventsOccurrence {
		if query.EventsCondition == EventCondEachGivenEvent {
			return buildEventCountForEachGivenEventsQueryNEW(projectId, query)
		}

		if len(query.EventsWithProperties) == 1 {
			return buildEventsOccurrenceSingleEventQuery(projectId, query)
		}

		return buildEventsOccurrenceWithGivenEventQuery(projectId, query)
	}

	if query.Type == QueryTypeUniqueUsers {
		if len(query.EventsWithProperties) == 1 {
			return buildUniqueUsersSingleEventQuery(projectId, query)
		}

		if query.EventsCondition == EventCondAnyGivenEvent {
			return buildUniqueUsersWithAnyGivenEventsQuery(projectId, query)
		}

		if query.EventsCondition == EventCondAllGivenEvent {
			return buildUniqueUsersWithAllGivenEventsQuery(projectId, query)
		}

		if query.EventsCondition == EventCondEachGivenEvent {
			return buildUniqueUsersWithEachGivenEventsQuery(projectId, query)
		}
	}

	return "", nil, errors.New("invalid query")
}

func LimitQueryResult(result *QueryResult, groupPropsLen int, groupByTimestamp bool) error {
	if groupPropsLen > 0 && groupByTimestamp {
		return limitGroupByTimestampResult(result, groupByTimestamp)
	}

	if groupPropsLen > 1 {
		return limitMultiGroupByPropertiesResult(result, groupByTimestamp)
	}

	// Others limited on SQL Query.
	return nil
}

// Limits top results and makes sure same group key combination available on different
// datetime, if exists on SQL result. Assumes result is sorted by count. Preserves all
// datetime for the limited combination of group keys.
func limitGroupByTimestampResult(result *QueryResult, groupByTimestamp bool) error {

	limitedResult := make([][]interface{}, 0, 0)

	start, end, err := getGroupKeyIndexesForSlicing(result.Headers)
	if err != nil {
		return err
	}

	// map[gk1:gk2] -> true
	keyLookup := make(map[string]bool, 0)
	for _, row := range result.Rows {
		// all group keys used as enc key.
		key := getEncodedKeyForCols(row[start:end])

		_, keyExists := keyLookup[key]
		// Limits no.of keys to ResultsLimit.
		if !keyExists && len(keyLookup) < ResultsLimit {
			keyLookup[key] = true
			keyExists = true
		}

		if keyExists {
			limitedResult = append(limitedResult, row)
		}
	}

	result.Rows = limitedResult
	return nil
}

// Limits results by left and right keys. Assumes result is already
// sorted by count and all group keys are used on SQL group by (makes all three group by
// values together as unique). Limited set dimension = ResultLimit * ResultLimit.
func limitMultiGroupByPropertiesResult(result *QueryResult, groupByTimestamp bool) error {

	limitedResult := make([][]interface{}, 0, 0)

	start, end, err := getGroupKeyIndexesForSlicing(result.Headers)
	if err != nil {
		return err
	}

	// Lookup based on left key (encoded key of all group key values excluding last)
	// right key (last group key value) ie: g1, g2, g3 -> map[c1:g1_c2:g2]map[g3]bool
	leftKeyEnd := end - 1
	keyLookup := make(map[string]map[interface{}]bool, 0)
	for _, row := range result.Rows {
		leftKey := getEncodedKeyForCols(row[start:leftKeyEnd])

		_, leftKeyExists := keyLookup[leftKey]
		// Limits no.of left keys to ResultsLimit.
		if !leftKeyExists && len(keyLookup) < ResultsLimit {
			keyLookup[leftKey] = make(map[interface{}]bool, 0)
			leftKeyExists = true
		}

		var rightKeyExists bool
		if leftKeyExists {
			// Limits no.of right keys to ResultsLimit.
			_, rightKeyExits := keyLookup[leftKey][row[leftKeyEnd]]
			if !rightKeyExits && len(keyLookup[leftKey]) < ResultsLimit {
				keyLookup[leftKey][row[leftKeyEnd]] = true
				rightKeyExists = true
			}
		}

		if leftKeyExists && rightKeyExists {
			limitedResult = append(limitedResult, row)
		}
	}

	result.Rows = limitedResult
	return nil
}

// Converts DB results into plottable query results.
func SanitizeQueryResult(result *QueryResult, query *Query) error {
	if query.GetGroupByTimestamp() != "" {
		return sanitizeGroupByTimestampResult(result, query)
	}

	// Replace group keys with real column names. should be last step.
	// of sanitization.
	if err := translateGroupKeysIntoColumnNames(result, query.GroupByProperties); err != nil {
		return err
	}

	if hasNumericalGroupBy(query.GroupByProperties) {
		sanitizeNumericalBucketRanges(result, query)
	}
	return nil
}

func sanitizeGroupByTimestampResult(result *QueryResult, query *Query) error {
	aggrIndex, timeIndex, err := GetTimstampAndAggregateIndexOnQueryResult(result.Headers)
	if err != nil {
		return err
	}

	// Todo: Supports only date as timestamp, add support for hour and month.
	if len(query.GroupByProperties) == 0 && len(query.EventsWithProperties) < 2 {
		err = addMissingTimestampsOnResultWithoutGroupByProps(result, query, aggrIndex, timeIndex)
	} else {
		err = addMissingTimestampsOnResultWithGroupByProps(result, query, aggrIndex, timeIndex)
	}

	if err != nil {
		return err
	}

	sortResultRowsByTimestamp(result.Rows, timeIndex)

	return nil
}

func sortResultRowsByTimestamp(resultRows [][]interface{}, timestampIndex int) {
	sort.Slice(resultRows, func(i, j int) bool {
		return (resultRows[i][timestampIndex].(time.Time)).Unix() <
			(resultRows[j][timestampIndex].(time.Time)).Unix()
	})
}

func getAllTimestampsBetweenByType(from, to int64, typ, timezone string) []time.Time {
	if typ == GroupByTimestampDate {
		return U.GetAllDatesAsTimestamp(from, to, timezone)
	}

	if typ == GroupByTimestampHour {
		return U.GetAllHoursAsTimestamp(from, to, timezone)
	}

	if typ == GroupByTimestampWeek {
		return U.GetAllWeeksAsTimestamp(from, to, timezone)
	}

	if typ == GroupByTimestampMonth {
		return U.GetAllMonthsAsTimestamp(from, to, timezone)
	}
	return []time.Time{}
}

func addMissingTimestampsOnResultWithoutGroupByProps(result *QueryResult, query *Query,
	aggrIndex int, timestampIndex int) error {

	rowsByTimestamp := make(map[string][]interface{}, 0)
	for _, row := range result.Rows {
		ts := row[timestampIndex].(time.Time)
		rowsByTimestamp[U.GetTimestampAsStrWithTimezone(ts, query.Timezone)] = row
	}

	timestamps := getAllTimestampsBetweenByType(query.From, query.To,
		query.GetGroupByTimestamp(), query.Timezone)

	filledResult := make([][]interface{}, 0, 0)
	// range over timestamps between given from and to.
	// uses timestamp string for comparison.
	for _, ts := range timestamps {
		if row, exists := rowsByTimestamp[U.GetTimestampAsStrWithTimezone(ts, query.Timezone)]; exists {
			// overrides timestamp with user timezone as sql results doesn't
			// return timezone used to query.
			row[timestampIndex] = ts
			filledResult = append(filledResult, row)
		} else {
			newRow := make([]interface{}, 3, 3)
			newRow[timestampIndex] = ts
			newRow[aggrIndex] = 0
			filledResult = append(filledResult, newRow)
		}
	}

	result.Rows = filledResult
	return nil
}

// Fills missing timestamp between given from and to timestamp for all group key combinations,
// on the limited result.
func addMissingTimestampsOnResultWithGroupByProps(result *QueryResult, query *Query,
	aggrIndex int, timestampIndex int) error {

	gkStart, gkEnd, err := getGroupKeyIndexesForSlicing(result.Headers)
	if err != nil {
		return err
	}

	filledResult := make([][]interface{}, 0, 0)

	rowsByGroupAndTimestamp := make(map[string]bool, 0)
	for _, row := range result.Rows {
		encCols := make([]interface{}, 0, 0)
		encCols = append(encCols, row[gkStart:gkEnd]...)

		timestampWithTimezone := U.GetTimestampAsStrWithTimezone(
			row[timestampIndex].(time.Time), query.Timezone)
		// encoded key with group values and timestamp from db row.
		encCols = append(encCols, timestampWithTimezone)
		encKey := getEncodedKeyForCols(encCols)
		rowsByGroupAndTimestamp[encKey] = true

		// overrides timestamp with user timezone as sql results doesn't
		// return timezone used to query.
		row[timestampIndex] = U.GetTimeFromTimestampStr(timestampWithTimezone)
		filledResult = append(filledResult, row)
	}

	timestamps := getAllTimestampsBetweenByType(query.From, query.To,
		query.GetGroupByTimestamp(), query.Timezone)

	for _, row := range result.Rows {
		for _, ts := range timestamps {
			encCols := make([]interface{}, 0, 0)
			encCols = append(encCols, row[gkStart:gkEnd]...)
			// encoded key with generated timestamp.
			encCols = append(encCols, U.GetTimestampAsStrWithTimezone(ts, query.Timezone))
			encKey := getEncodedKeyForCols(encCols)

			_, exists := rowsByGroupAndTimestamp[encKey]
			if !exists {
				// create new row with group values and missing date
				// for those group combination and aggr 0.
				rowLen := len(result.Headers)
				newRow := make([]interface{}, rowLen, rowLen)
				groupValues := row[gkStart:gkEnd]

				for i := 0; i < rowLen; {
					if i == gkStart {
						for _, gv := range groupValues {
							newRow[i] = gv
							i++
						}
					}

					if i == aggrIndex {
						newRow[i] = 0
						i++
					}

					if i == timestampIndex {
						newRow[i] = ts
						i++
					}
				}
				rowsByGroupAndTimestamp[encKey] = true
				filledResult = append(filledResult, newRow)
			}
		}
	}

	result.Rows = filledResult
	return nil
}

/*
addEventFilterStepsForUniqueUsersQuery - Builds and adds events filter steps
for unique users queries with group by properties and date.

Example:

step0 AS (
	-- Using DISTINCT ON user_id for getting unique users and event properties on
	first occurrence of the event done by the user --
    SELECT DISTINCT ON(events.user_id) events.user_id as event_user_id,
    events.properties->>'category' as group_prop2 FROM events
    LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
    WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
    AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='View Project')
    ORDER BY events.user_id, events.timestamp ASC
),
step1 AS (
    SELECT DISTINCT ON(events.user_id) events.user_id as event_user_id,
    events.properties->>'category' as group_prop2 FROM events
    LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
    WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
    AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='Fund Project')
    ORDER BY events.user_id, events.timestamp ASC
)
*/

func addEventFilterStepsForUniqueUsersQuery(projectID uint64, q *Query,
	qStmnt *string, qParams *[]interface{}) ([]string, map[string][]string) {

	var commonSelect string
	var commonOrderBy string
	stepsToKeysMap := make(map[string][]string)
	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egAnySelect, egAnyParams, egAnyGroupKeys := buildGroupKeys(eventGroupProps)

	if q.GetGroupByTimestamp() != "" {
		selectTimestamp := getSelectTimestampByType(q.GetGroupByTimestamp(), q.Timezone)
		// select and order by with datetime.
		commonSelect = fmt.Sprintf("DISTINCT ON(coal_user_id%%, %s)", selectTimestamp) +
			fmt.Sprintf(" COALESCE(users.customer_user_id,events.user_id) as coal_user_id%%, %s as %s,", selectTimestamp, AliasDateTime) +
			" events.user_id as event_user_id"

		commonSelect = strings.ReplaceAll(commonSelect, "%", "%s")

		commonOrderBy = fmt.Sprintf("coal_user_id%%, %s, events.timestamp ASC", AliasDateTime)
		commonOrderBy = strings.ReplaceAll(commonOrderBy, "%", "%s")
	} else {
		// default select.
		commonSelect = "DISTINCT ON(coal_user_id%s) COALESCE(users.customer_user_id,events.user_id)" +
			" as coal_user_id%s, events.user_id as event_user_id"
	}

	if len(q.GroupByProperties) > 0 && commonOrderBy == "" {
		// Using first occurred event_properites after distinct on user_id.
		commonOrderBy = "coal_user_id%s, events.timestamp ASC"
	}

	steps := make([]string, 0, 0)
	for i, ewp := range q.EventsWithProperties {
		refStepName := stepNameByIndex(i)
		steps = append(steps, refStepName)

		var stepSelect, stepOrderBy string
		var stepParams []interface{}
		var groupByUserProperties bool
		if q.EventsCondition == EventCondAllGivenEvent || q.EventsCondition == EventCondEachGivenEvent {
			var stepGroupSelect, stepGroupKeys string
			var stepGroupParams []interface{}
			stepGroupSelect, stepGroupParams, stepGroupKeys, groupByUserProperties = buildGroupKeyForStep(
				&q.EventsWithProperties[i], q.GroupByProperties, i+1)

			eventSelect := commonSelect
			if q.EventsCondition == EventCondEachGivenEvent {
				eventNameSelect := "'" + ewp.Name + "'" + "::text" + " AS event_name "
				eventSelect = joinWithComma(eventSelect, eventNameSelect)
			}
			if stepGroupSelect != "" {
				stepSelect = fmt.Sprintf(eventSelect, ", "+stepGroupKeys, ", "+stepGroupSelect)
				stepOrderBy = fmt.Sprintf(commonOrderBy, ", "+stepGroupKeys)
				stepParams = stepGroupParams
				stepsToKeysMap[refStepName] = strings.Split(stepGroupKeys, ",")
			} else {
				stepSelect = fmt.Sprintf(eventSelect, "", "")
				if commonOrderBy != "" {
					stepOrderBy = fmt.Sprintf(commonOrderBy, "")
				}
			}
		} else {
			if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
				stepSelect = fmt.Sprintf(commonSelect, ", "+egAnyGroupKeys, ", "+egAnySelect)
				stepParams = egAnyParams
				stepOrderBy = fmt.Sprintf(commonOrderBy, ", "+egAnyGroupKeys)
			} else {
				stepSelect = fmt.Sprintf(commonSelect, "", "")
				if commonOrderBy != "" {
					stepOrderBy = fmt.Sprintf(commonOrderBy, "")
				}
			}
		}

		addJoinStmnt := "JOIN users ON events.user_id=users.id"
		if groupByUserProperties && !hasWhereEntity(ewp, PropertyEntityUser) {
			// If event has filter on user property, JOIN on user_properties is added in next step.
			// Skip addding here to avoid duplication.
			addJoinStmnt += " JOIN user_properties on events.user_properties_id=user_properties.id"
		}
		addFilterEventsWithPropsQuery(projectID, qStmnt, qParams, ewp, q.From, q.To,
			"", refStepName, stepSelect, stepParams, addJoinStmnt, "", stepOrderBy)

		if i < len(q.EventsWithProperties)-1 {
			*qStmnt = *qStmnt + ","
		}
	}

	return steps, stepsToKeysMap
}

/*
addUniqueUsersAggregationQuery - Builds and adds final aggregation query for Unique Users queries
with group by properties and date.

Example:

SELECT user_properties.properties->>'gender' as gk_0, gk_1,
COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))) FROM users_union
LEFT JOIN users ON users_union.event_user_id=users.id
LEFT JOIN user_properties ON users.id=user_properties.user_id and user_properties.id=users.properties_id
GROUP BY gk_0, gk_1 ORDER BY count DESC LIMIT 10000;
*/
func addUniqueUsersAggregationQuery(query *Query, qStmnt *string, qParams *[]interface{}, refStep string) {
	eventLevelGroupBys, otherGroupBys := separateEventLevelGroupBys(query.GroupByProperties)
	var egKeys string
	var unionStepName string

	if query.EventsCondition == EventCondAllGivenEvent {
		_, _, egKeys = buildGroupKeys(eventLevelGroupBys)
		unionStepName = "all_users_intersect"
	} else if query.EventsCondition == EventCondEachGivenEvent {
		_, _, egKeys = buildGroupKeys(eventLevelGroupBys)
		unionStepName = "each_users_union"
	} else {
		eventGroupProps := filterGroupPropsByType(otherGroupBys, PropertyEntityEvent)
		_, _, egKeys = buildGroupKeys(eventGroupProps)
		unionStepName = "any_users_union"
	}

	// select
	userGroupProps := filterGroupPropsByType(otherGroupBys, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	*qParams = append(*qParams, ugSelectParams...)
	// order of group keys changes here if users and event
	// group by used together, but translated correctly.
	termSelect := ""
	if query.EventsCondition == EventCondEachGivenEvent {
		termSelect = fmt.Sprintf(" %s.event_name, ", refStep)
	}
	termSelect = termSelect + joinWithComma(ugSelect, egKeys)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	termSelect = appendSelectTimestampColIfRequired(termSelect, isGroupByTimestamp)

	termStmnt := fmt.Sprintf("SELECT %s.event_user_id, ", refStep) + termSelect + " FROM " + refStep
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStep + ".event_user_id=users.id"
		termStmnt = termStmnt + " " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}

	_, _, groupKeys := buildGroupKeys(query.GroupByProperties)

	termStmnt = as(unionStepName, termStmnt)
	var aggregateFromStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys string
	if hasNumericalGroupBy(query.GroupByProperties) {
		eventName := ""
		if query.EventsCondition == EventCondEachGivenEvent {
			eventName = AliasEventName
		}
		bucketedStepName, bucketedSelectKeys, bucketedGroupBys, bucketedOrderBys := appendNumericalBucketingSteps(
			&termStmnt, query.GroupByProperties, unionStepName, eventName, isGroupByTimestamp, "event_user_id")
		aggregateFromStepName = bucketedStepName
		aggregateSelectKeys = bucketedSelectKeys
		aggregateGroupBys = strings.Join(bucketedGroupBys, ", ")
		aggregateOrderBys = strings.Join(bucketedOrderBys, ", ")
		*qStmnt = appendStatement(*qStmnt, ", "+termStmnt)
	} else {

		if groupKeys != "" {
			// Order by count, which will added later.
			aggregateFromStepName = unionStepName
			aggregateSelectKeys = groupKeys + ", "
			aggregateGroupBys = groupKeys
			*qStmnt = appendStatement(*qStmnt, ", "+termStmnt)
		} else {
			// No group by clause added. Use previous step and rest all leave empty.
			aggregateFromStepName = refStep
		}
	}

	aggregateSelect := "SELECT "
	if isGroupByTimestamp {
		aggregateSelect = aggregateSelect + AliasDateTime + ", "
		aggregateGroupBys = joinWithComma(aggregateGroupBys, AliasDateTime)
	}
	// adding select event_name and group by event_name for each event-user count
	if query.EventsCondition == EventCondEachGivenEvent {
		aggregateSelect = aggregateSelect + AliasEventName + ", "
		aggregateGroupBys = joinWithComma(AliasEventName, aggregateGroupBys)
	}
	aggregateSelect = aggregateSelect + aggregateSelectKeys + fmt.Sprintf("COUNT(DISTINCT(event_user_id)) AS %s FROM %s",
		AliasAggr, aggregateFromStepName)

	aggregateSelect = appendGroupBy(aggregateSelect, aggregateGroupBys)
	if aggregateOrderBys != "" {
		aggregateSelect = aggregateSelect + " ORDER BY " + aggregateOrderBys
	} else {
		aggregateSelect = appendOrderByAggr(aggregateSelect)
	}
	aggregateSelect = appendLimitByCondition(aggregateSelect, query.GroupByProperties, isGroupByTimestamp)
	*qStmnt = appendStatement(*qStmnt, aggregateSelect)
}

/*
buildEventsOccurrenceWithAnyGivenEventQuery builds query for any given event and single event query,
Group by: user_properties, event_properties.

* Without group by user_property

WITH
	SELECT COUNT(*), events.properties->>'category' as group_prop1 FROM events
	LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
	WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
	AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
	AND user_properties.properties->>'gender'='M'


* With group by user_property

WITH
    e1 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id, events.properties->>'category' as group_prop1 FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
        AND user_properties.properties->>'gender'='M'
    )
    SELECT user_properties.properties->>'$region' as group_prop2, group_prop1, count(*) from e1
    left join users on e1.event_user_id=users.id
    left join user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id
    group by group_prop1, group_prop2 order by group_prop2;
*/

func buildEventsOccurrenceSingleEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) != 1 {
		return "", nil, errors.New("invalid no.of events for single event query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityUser) || hasNumericalGroupBy(q.GroupByProperties) {
		// Using any given event query, which handles groups already.
		return buildEventsOccurrenceWithGivenEventQuery(projectId, q)
	}

	var qStmnt string
	var qParams []interface{}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egSelect, egSelectParams, egKeys := buildGroupKeys(eventGroupProps)
	isGroupByTimestamp := q.GetGroupByTimestamp() != ""

	var qSelect string
	qSelect = appendSelectTimestampIfRequired(qSelect, q.GetGroupByTimestamp(), q.Timezone)
	qSelect = joinWithComma(qSelect, egSelect, fmt.Sprintf("COUNT(*) AS %s", AliasAggr))

	addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[0], q.From, q.To,
		"", "", qSelect, egSelectParams, "", "", "")

	qStmnt = appendGroupByTimestampIfRequired(qStmnt, isGroupByTimestamp, egKeys)
	qStmnt = appendOrderByAggr(qStmnt)
	qStmnt = appendLimitByCondition(qStmnt, q.GroupByProperties, isGroupByTimestamp)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersWithEachGivenEventsQuery computes user count for each event with given filter and
breakdown.

Example: Query with date.
Group By: user_properties, event_properties

Sample query with ewp:
	$session
		"en": "event", "pr": "$source", "op": "equals", "va": "google", "ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
	MagazineViews
		"en": "event","pr": "$source","op": "equals","va": "google","ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
	www.livspace.com/in/hire-a-designer
		"en": "event","pr": "$source","op": "equals","va": "google","ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
gbp: [
	"pr": "$source","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "MagazineViews","eni": 2
	"pr": "$city","en": "user","pty": "categorical","ena": "www.livspace.com/in/hire-a-designer","eni": 3
	"pr": "$country","en": "user","pty": "categorical"
]
gbt: "date"

QUERY:

WITH

step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND name='$session'),

step_0 AS (SELECT DISTINCT ON(coal_user_id, _group_key_0, _group_key_1, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta')) COALESCE(users.customer_user_id,events.user_id)
as coal_user_id, CASE WHEN events.properties->>'$source' IS NULL THEN '$none' WHEN
events.properties->>'$source' = '' THEN '$none' ELSE events.properties->>'$source' END AS
_group_key_0, CASE WHEN events.properties->>'$campaign' IS NULL THEN '$none' WHEN
events.properties->>'$campaign' = '' THEN '$none' ELSE events.properties->>'$campaign'
END AS _group_key_1, date_trunc('day', to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta')
as datetime, events.user_id as event_user_id, '$session'::text AS event_name  FROM events JOIN
users ON events.user_id=users.id LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
WHERE events.project_id='204' AND timestamp>='1583001000' AND timestamp<='1585679399' AND
events.event_name_id IN (SELECT id FROM step_0_names WHERE project_id='204' AND name='$session')
AND ( events.properties->>'$source' = 'google' AND user_properties.properties->>'$country' = 'India' )
ORDER BY coal_user_id, _group_key_0, _group_key_1, datetime, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND
name='MagazineViews'), step_1 AS (SELECT DISTINCT ON(coal_user_id, _group_key_2, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta')) COALESCE(users.customer_user_id,
events.user_id) as coal_user_id, CASE WHEN events.properties->>'$campaign' IS NULL THEN '$none'
WHEN events.properties->>'$campaign' = '' THEN '$none' ELSE events.properties->>'$campaign'
END AS _group_key_2, date_trunc('day', to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta')
as datetime, events.user_id as event_user_id, 'MagazineViews'::text AS event_name  FROM events
JOIN users ON events.user_id=users.id LEFT JOIN user_properties ON
events.user_properties_id=user_properties.id WHERE events.project_id='204' AND timestamp>='1583001000'
AND timestamp<='1585679399' AND events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='204'
AND name='MagazineViews') AND ( events.properties->>'$source' = 'google' AND
user_properties.properties->>'$country' = 'India' ) ORDER BY coal_user_id, _group_key_2, datetime,
events.timestamp ASC),

step_2_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND
name='www.livspace.com/in/hire-a-designer'),

step_2 AS (SELECT DISTINCT ON(coal_user_id, _group_key_3, date_trunc('day', to_timestamp(timestamp)
AT TIME ZONE 'Asia/Calcutta')) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
CASE WHEN user_properties.properties->>'$city' IS NULL THEN '$none' WHEN
user_properties.properties->>'$city' = '' THEN '$none' ELSE user_properties.properties->>'$city'
END AS _group_key_3, date_trunc('day', to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta') as
datetime, events.user_id as event_user_id, 'www.livspace.com/in/hire-a-designer'::text AS
event_name  FROM events JOIN users ON events.user_id=users.id LEFT JOIN user_properties ON
events.user_properties_id=user_properties.id WHERE events.project_id='204' AND timestamp>='1583001000'
AND timestamp<='1585679399' AND events.event_name_id IN (SELECT id FROM step_2_names WHERE
project_id='204' AND name='www.livspace.com/in/hire-a-designer') AND ( events.properties->>'$source'
= 'google' AND user_properties.properties->>'$country' = 'India' ) ORDER BY coal_user_id, _group_key_3,
datetime, events.timestamp ASC),


each_events_union AS (SELECT step_0.event_name as event_name, step_0.coal_user_id as coal_user_id,
step_0.event_user_id as event_user_id, datetime , _group_key_0 as _group_key_0,  _group_key_1 as
_group_key_1,  ''  as _group_key_2,  ''  as _group_key_3 FROM step_0 UNION ALL SELECT
step_1.event_name as event_name, step_1.coal_user_id as coal_user_id, step_1.event_user_id as
event_user_id, datetime ,  ''  as _group_key_0,  ''  as  _group_key_1, _group_key_2 as _group_key_2,
''  as _group_key_3 FROM step_1 UNION ALL SELECT step_2.event_name as event_name, step_2.coal_user_id
as coal_user_id, step_2.event_user_id as event_user_id, datetime ,  ''  as _group_key_0,  ''
as  _group_key_1,  ''  as _group_key_2, _group_key_3 as _group_key_3 FROM step_2) ,

each_users_union AS (SELECT each_events_union.event_user_id,  each_events_union.event_name,
CASE WHEN user_properties.properties->>'$country' IS NULL THEN '$none' WHEN
user_properties.properties->>'$country' = '' THEN '$none' ELSE user_properties.properties->>'$country'
END AS _group_key_4, _group_key_0, _group_key_1, _group_key_2, _group_key_3, datetime FROM each_events_union
LEFT JOIN users ON each_events_union.event_user_id=users.id LEFT JOIN user_properties ON
users.id=user_properties.user_id AND user_properties.id=users.properties_id)

SELECT datetime, event_name, _group_key_0, _group_key_1, _group_key_2, _group_key_3, _group_key_4,
COUNT(DISTINCT(event_user_id)) AS count FROM each_users_union GROUP BY event_name, _group_key_0,
_group_key_1, _group_key_2, _group_key_3, _group_key_4, datetime ORDER BY count DESC LIMIT 100000
*/
func buildUniqueUsersWithEachGivenEventsQuery(projectID uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps, stepsToKeysMap := addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams)
	totalGroupKeys := 0
	for _, val := range stepsToKeysMap {
		totalGroupKeys = totalGroupKeys + len(val)
	}
	// union each event
	stepUsersUnion := "each_events_union"
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	var unionStmnt string

	for i, step := range steps {
		selectStr := fmt.Sprintf("%s.event_name as event_name, %s.coal_user_id as coal_user_id, %s.event_user_id as event_user_id", step, step, step)
		selectStr = appendSelectTimestampColIfRequired(selectStr, isGroupByTimestamp)
		egKeysForStep := getKeysForStep(step, steps, stepsToKeysMap, totalGroupKeys)
		if egKeysForStep != "" {
			selectStr = selectStr + " , " + egKeysForStep
		}
		selectStmnt := fmt.Sprintf("SELECT %s FROM %s", selectStr, step)
		if i == 0 {
			unionStmnt = selectStmnt
		} else {
			unionStmnt = unionStmnt + " UNION ALL " + selectStmnt
		}
	}

	qStmnt = joinWithComma(qStmnt, as(stepUsersUnion, unionStmnt))
	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, stepUsersUnion)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/* getKeysForStep returns column keys for select query for the given step with values and
  empty string ('') for all other step's breakdowns
  for ex:
	breakdown for e1: k0
	breakdown for e2: k1, k2
	breakdown for e3: k3, k4
	key for e1: gk0, '', '',  '', ''
	key for e2: '', gk1, gk2, '', ''
    key for e3: '', '',  '',  gk3, gk4
*/
func getKeysForStep(step string, steps []string, keysMap map[string][]string, totalKeys int) string {

	keys := ""
	keyCnt := 0
	for _, s := range steps {
		if s == step {
			for i := 0; i < len(keysMap[s]); i++ {
				keys = keys + keysMap[s][i]
				keyCnt++
				if keyCnt != totalKeys {
					keys += ", "
				}
			}
		} else {
			for i := 0; i < len(keysMap[s]); i++ {
				keys += " '' " + " as " + keysMap[s][i]
				keyCnt++
				if keyCnt != totalKeys {
					keys += ", "
				}
			}
		}
	}
	return keys
}

/*
buildUniqueUsersWithAllGivenEventsQuery builds a query like below,
Group by: user_properties, event_properties

Example: Query without date and with group by properties.

Sample query with ewp:
	View Project
	Fund Project
gbp:
	user_property -> $present -> $session_count (numerical)
	event_property -> 2. $session -> $hour_of_day (numerical)
	user_property -> 2. $session -> $platform
	user_property -> 1. View Project -> $user_agent

WITH
step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step_0 AS (SELECT DISTINCT ON(coal_user_id, _group_key_3) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
CASE WHEN user_properties.properties->>'' IS NULL THEN '$none' WHEN user_properties.properties->>'' = '' THEN '$none'
ELSE user_properties.properties->>'' END AS _group_key_3, events.user_id as event_user_id FROM events JOIN users ON
events.user_id=users.id JOIN user_properties on events.user_properties_id=user_properties.id WHERE events.project_id='3'
AND timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM step_0_names
WHERE project_id='3' AND name='View Project') ORDER BY coal_user_id, _group_key_3, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='Fund Project'),
step_1 AS (SELECT DISTINCT ON(coal_user_id, _group_key_1, _group_key_2) COALESCE(users.customer_user_id,events.user_id)
as coal_user_id, CASE WHEN events.properties->>'' IS NULL THEN '$none' WHEN events.properties->>'' = '' THEN '$none'
ELSE events.properties->>'' END AS _group_key_1, CASE WHEN user_properties.properties->>'' IS NULL THEN '$none'
WHEN user_properties.properties->>'' = '' THEN '$none' ELSE user_properties.properties->>'' END AS _group_key_2,
events.user_id as event_user_id FROM events JOIN users ON events.user_id=users.id JOIN user_properties on
events.user_properties_id=user_properties.id WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599'
AND events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='3' AND name='Fund Project')
ORDER BY coal_user_id, _group_key_1, _group_key_2, events.timestamp ASC),

events_intersect AS (SELECT step_0.event_user_id as event_user_id, step_1._group_key_1, step_1._group_key_2,
step_0._group_key_3 FROM step_0  JOIN step_1 ON step_1.coal_user_id = step_0.coal_user_id) ,

all_users_intersect AS (SELECT events_intersect.event_user_id, CASE WHEN user_properties.properties->>'' IS NULL THEN
'$none' WHEN user_properties.properties->>'' = '' THEN '$none' ELSE user_properties.properties->>'' END AS _group_key_0,
_group_key_1, _group_key_2, _group_key_3 FROM events_intersect LEFT JOIN users ON events_intersect.event_user_id=users.id
LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id),

_group_key_0_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_0::numeric) + 0.00001 AS lbound, percentile_disc(0.98)
 WITHIN GROUP(ORDER BY _group_key_0::numeric) AS ubound FROM all_users_intersect WHERE _group_key_0 != '$none'),

_group_key_1_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_1::numeric) + 0.00001 AS lbound, percentile_disc(0.98)
WITHIN GROUP(ORDER BY _group_key_1::numeric) AS ubound FROM all_users_intersect WHERE _group_key_1 != '$none'),

bucketed AS (SELECT COALESCE(NULLIF(_group_key_0, '$none'), 'NaN') AS _group_key_0, CASE WHEN _group_key_0 = '$none' THEN -1
ELSE width_bucket(_group_key_0::numeric, _group_key_0_bounds.lbound::numeric, COALESCE(NULLIF(_group_key_0_bounds.ubound,
_group_key_0_bounds.lbound), _group_key_0_bounds.ubound+1)::numeric, 8) END AS _group_key_0_bucket, COALESCE(NULLIF(_group_key_1, '$none'),
'NaN') AS _group_key_1, CASE WHEN _group_key_1 = '$none' THEN -1 ELSE width_bucket(_group_key_1::numeric, _group_key_1_bounds.lbound::numeric,
COALESCE(NULLIF(_group_key_1_bounds.ubound, _group_key_1_bounds.lbound), _group_key_1_bounds.ubound+1)::numeric, 8) END
AS _group_key_1_bucket, _group_key_2, _group_key_3, event_user_id FROM all_users_intersect, _group_key_0_bounds, _group_key_1_bounds)

SELECT COALESCE(NULLIF(concat(round(min(_group_key_0::numeric), 1), ' - ', round(max(_group_key_0::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_0,
COALESCE(NULLIF(concat(round(min(_group_key_1::numeric), 1), ' - ', round(max(_group_key_1::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_1,
_group_key_2, _group_key_3,  COUNT(DISTINCT(event_user_id)) AS count FROM bucketed GROUP BY _group_key_0_bucket, _group_key_1_bucket,
_group_key_2, _group_key_3 ORDER BY _group_key_0_bucket, _group_key_1_bucket LIMIT 100000

Example: Query with date

WITH
step0 AS (
	-- DISTINCT ON user_id, date preserves users on each date --
	SELECT DISTINCT ON(events.user_id, (to_timestamp(timestamp) AT TIME ZONE 'UTC')::date) events.user_id as event_user_id,
	(to_timestamp(timestamp) AT TIME ZONE 'UTC')::date as date FROM events
	WHERE events.project_id='1' AND timestamp>='1561091973' AND timestamp<='1561178373'
	AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='localhost:3000/#/core')
	ORDER BY events.user_id, date, events.timestamp ASC
	-- Order by user_id, timestamp is not possible as we need to preserve unique user with date using DISTINCT ON (user_id, date) --
),
step1 AS (
	SELECT DISTINCT ON(events.user_id, (to_timestamp(timestamp) AT TIME ZONE 'UTC')::date) events.user_id as event_user_id,
	(to_timestamp(timestamp) AT TIME ZONE 'UTC')::date as date FROM events
	WHERE events.project_id='1' AND timestamp>='1561091973' AND timestamp<='1561178373'
	AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='run_query')
	ORDER BY events.user_id, date, events.timestamp ASC
),
users_intersect AS (
	-- Users who have done all the steps on each date. Join by user, date. --
	SELECT step0.event_user_id as event_user_id, step0.date as date FROM step0
	JOIN step1 ON step0.event_user_id = step1.event_user_id AND step0.date = step1.date
)
SELECT date, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))) AS count FROM users_intersect
LEFT JOIN users ON users_intersect.event_user_id=users.id GROUP BY date ORDER BY count DESC LIMIT 100000;

*/
func buildUniqueUsersWithAllGivenEventsQuery(projectID uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps, _ := addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams)

	// users intersection
	intersectSelect := fmt.Sprintf("%s.event_user_id as event_user_id", steps[0])
	if query.GetGroupByTimestamp() != "" {
		intersectSelect = joinWithComma(intersectSelect,
			fmt.Sprintf("%s.%s as %s", steps[0], AliasDateTime, AliasDateTime))
	}

	// adds group by event property with user selected event (step).
	eventGroupKeysWithStep := buildEventGroupKeysWithStep(query.GroupByProperties,
		query.EventsWithProperties)
	intersectSelect = joinWithComma(intersectSelect, eventGroupKeysWithStep)

	var intersectJoin string
	for i := range steps {
		if i > 0 {
			intersectJoin = intersectJoin + " " + fmt.Sprintf("JOIN %s ON %s.coal_user_id = %s.coal_user_id",
				steps[i], steps[i], steps[i-1])

			// include date also intersection condition on
			// group by timestamp.
			if query.GetGroupByTimestamp() != "" {
				intersectJoin = intersectJoin + " " + fmt.Sprintf("AND %s.%s = %s.%s",
					steps[i], AliasDateTime, steps[i-1], AliasDateTime)
			}
		}
	}
	stepEventsIntersect := "events_intersect"
	qStmnt = joinWithComma(qStmnt, as(stepEventsIntersect,
		fmt.Sprintf("SELECT %s FROM %s %s", intersectSelect, steps[0], intersectJoin)))

	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, stepEventsIntersect)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersWithAnyGivenEventsQuery
Group By: user_properties, event_properties

Example: Query without date and with group by properties.

Sample query with ewp:
	View Project
	Fund Project
gbp:
	event_property -> $browser
	event_property -> $hour_of_day (numerical)
	user_property -> $session_count (numerical)

WITH
step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step_0 AS (SELECT DISTINCT ON(coal_user_id, _group_key_0, _group_key_1) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
CASE WHEN events.properties->>'' IS NULL THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END
AS _group_key_0, CASE WHEN events.properties->>'' IS NULL THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE
events.properties->>'' END AS _group_key_1, events.user_id as event_user_id FROM events JOIN users ON events.user_id=users.id
WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN (SELECT id
FROM step_0_names WHERE project_id='3' AND name='View Project') ORDER BY coal_user_id, _group_key_0, _group_key_1, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='Fund Project'),

step_1 AS (SELECT DISTINCT ON(coal_user_id, _group_key_0, _group_key_1) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
CASE WHEN events.properties->>'' IS NULL THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END
AS _group_key_0, CASE WHEN events.properties->>'' IS NULL THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE
events.properties->>'' END AS _group_key_1, events.user_id as event_user_id FROM events JOIN users ON events.user_id=users.id
WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM
step_1_names WHERE project_id='3' AND name='Fund Project') ORDER BY coal_user_id, _group_key_0, _group_key_1, events.timestamp ASC),

events_union AS (SELECT step_0.event_user_id as event_user_id, _group_key_0, _group_key_1 FROM step_0 UNION ALL
SELECT step_1.event_user_id as event_user_id, _group_key_0, _group_key_1 FROM step_1) ,

any_users_union AS (SELECT events_union.event_user_id, CASE WHEN user_properties.properties->>'' IS NULL THEN '$none'
WHEN user_properties.properties->>'' = '' THEN '$none' ELSE user_properties.properties->>'' END AS _group_key_2, _group_key_0,
_group_key_1 FROM events_union LEFT JOIN users ON events_union.event_user_id=users.id LEFT JOIN user_properties
ON users.id=user_properties.user_id AND user_properties.id=users.properties_id),

_group_key_1_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_1::numeric) + 0.00001 AS lbound, percentile_disc(0.98)
WITHIN GROUP(ORDER BY _group_key_1::numeric) AS ubound FROM any_users_union WHERE _group_key_1 != '$none'),

_group_key_2_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_2::numeric) + 0.00001 AS lbound, percentile_disc(0.98)
WITHIN GROUP(ORDER BY _group_key_2::numeric) AS ubound FROM any_users_union WHERE _group_key_2 != '$none'),

bucketed AS (SELECT _group_key_0, COALESCE(NULLIF(_group_key_1, '$none'), 'NaN') AS _group_key_1, CASE WHEN _group_key_1 = '$none'
THEN -1 ELSE width_bucket(_group_key_1::numeric, _group_key_1_bounds.lbound::numeric, COALESCE(NULLIF(_group_key_1_bounds.ubound,
_group_key_1_bounds.lbound), _group_key_1_bounds.ubound+1)::numeric, 8) END AS _group_key_1_bucket, COALESCE(NULLIF(_group_key_2, '$none'),
'NaN') AS _group_key_2, CASE WHEN _group_key_2 = '$none' THEN -1 ELSE width_bucket(_group_key_2::numeric, _group_key_2_bounds.lbound::numeric,
COALESCE(NULLIF(_group_key_2_bounds.ubound, _group_key_2_bounds.lbound), _group_key_2_bounds.ubound+1)::numeric, 8) END
AS _group_key_2_bucket, event_user_id FROM any_users_union, _group_key_1_bounds, _group_key_2_bounds)

SELECT _group_key_0, COALESCE(NULLIF(concat(round(min(_group_key_1::numeric), 1), ' - ', round(max(_group_key_1::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_1,
COALESCE(NULLIF(concat(round(min(_group_key_2::numeric), 1), ' - ', round(max(_group_key_2::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_2,
COUNT(DISTINCT(event_user_id)) AS count FROM bucketed GROUP BY _group_key_0, _group_key_1_bucket, _group_key_2_bucket
ORDER BY _group_key_1_bucket, _group_key_2_bucket LIMIT 100000

Example: Query with date.

WITH
	step0 AS (
		-- DISTINCT ON user_id, date preserves users on each date --
		SELECT DISTINCT ON(events.user_id, (to_timestamp(timestamp) AT TIME ZONE 'UTC')::date) events.user_id as event_user_id,
		(to_timestamp(timestamp) AT TIME ZONE 'UTC')::date as date FROM events
		WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='View Project')
		ORDER BY events.user_id, date, events.timestamp ASC
		-- Order by user_id, timestamp is not possible as we need to preserve unique user with date using DISTINCT ON (user_id, date) --
	),
	step1 AS (
		SELECT DISTINCT ON(events.user_id, (to_timestamp(timestamp) AT TIME ZONE 'UTC')::date) events.user_id as event_user_id,
		(to_timestamp(timestamp) AT TIME ZONE 'UTC')::date as date FROM events
		WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='Fund Project')
		ORDER BY events.user_id, date, events.timestamp ASC
	),
	users_union AS (
		SELECT step0.event_user_id as event_user_id, step0.date as date FROM step0 UNION ALL
    	SELECT step1.event_user_id as event_user_id, step1.date as date FROM step1
	)
	SELECT date, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))) AS count FROM users_union
	LEFT JOIN users ON users_union.event_user_id=users.id GROUP BY date ORDER BY count DESC LIMIT 100000;
*/
func buildUniqueUsersWithAnyGivenEventsQuery(projectID uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps, _ := addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams)

	eventGroupProps := filterGroupPropsByType(query.GroupByProperties, PropertyEntityEvent)
	_, _, egKeys := buildGroupKeys(eventGroupProps)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	var unionStmnt string
	for i, step := range steps {
		selectStr := fmt.Sprintf("%s.event_user_id as event_user_id", step)
		selectStr = appendSelectTimestampColIfRequired(selectStr, isGroupByTimestamp)
		selectStr = joinWithComma(selectStr, egKeys)

		selectStmnt := fmt.Sprintf("SELECT %s FROM %s", selectStr, step)
		if i == 0 {
			unionStmnt = selectStmnt
		} else {
			unionStmnt = unionStmnt + " UNION ALL " + selectStmnt
		}
	}

	stepUsersUnion := "events_union"
	qStmnt = joinWithComma(qStmnt, as(stepUsersUnion, unionStmnt))

	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, stepUsersUnion)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersSingleEventQuery
Sample query for ewp ALL:
	View Project
Group By:
	event_property -> 1. View Project -> $hour_of_day
	user_property -> $present -> $campaign

WITH
step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step_0 AS (SELECT DISTINCT ON(coal_user_id, _group_key_0) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
CASE WHEN events.properties->>'' IS NULL THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>''
END AS _group_key_0, events.user_id as event_user_id FROM events JOIN users ON events.user_id=users.id WHERE
events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM
step_0_names WHERE project_id='3' AND name='View Project') ORDER BY coal_user_id, _group_key_0, events.timestamp ASC),

all_users_intersect AS (SELECT step_0.event_user_id, CASE WHEN user_properties.properties->>'' IS NULL THEN '$none'
WHEN user_properties.properties->>'' = '' THEN '$none' ELSE user_properties.properties->>'' END AS _group_key_1,
_group_key_0 FROM step_0 LEFT JOIN users ON step_0.event_user_id=users.id LEFT JOIN user_properties ON
users.id=user_properties.user_id AND user_properties.id=users.properties_id),

_group_key_0_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_0::numeric) + 0.00001 AS lbound, percentile_disc(0.98)
WITHIN GROUP(ORDER BY _group_key_0::numeric) AS ubound FROM all_users_intersect WHERE _group_key_0 != '$none'),

bucketed AS (SELECT COALESCE(NULLIF(_group_key_0, '$none'), 'NaN') AS _group_key_0, CASE WHEN _group_key_0 = '$none' THEN -1
ELSE width_bucket(_group_key_0::numeric, _group_key_0_bounds.lbound::numeric, COALESCE(NULLIF(_group_key_0_bounds.ubound,
_group_key_0_bounds.lbound), _group_key_0_bounds.ubound+1)::numeric, 8) END AS _group_key_0_bucket, _group_key_1,
event_user_id FROM all_users_intersect, _group_key_0_bounds)

SELECT COALESCE(NULLIF(concat(round(min(_group_key_0::numeric), 1), ' - ', round(max(_group_key_0::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_0,
_group_key_1,  COUNT(DISTINCT(event_user_id)) AS count FROM bucketed GROUP BY _group_key_0_bucket, _group_key_1
ORDER BY _group_key_0_bucket LIMIT 100000
*/
func buildUniqueUsersSingleEventQuery(projectID uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps, _ := addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams)
	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, steps[0])
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildEventsOccurrenceWithGivenEventQuery builds query for any given event and single event query,
Group by: user_properties, event_properties.

Sample query for ewp:
	View Project
	Fund Project
gbp:
	event_property -> $hour_of_day (numeric)
	event_property -> $day_of_week (categoric)
	user_property -> $session_count (numeric)

WITH
step0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step0 AS (SELECT events.id as event_id, events.user_id as event_user_id, CASE WHEN events.properties->>'' IS NULL THEN '$none'
WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END AS _group_key_0, CASE WHEN events.properties->>'' IS NULL
THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END AS _group_key_1, 'View Project'::text
AS event_name  FROM events  WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599' AND
events.event_name_id IN (SELECT id FROM step0_names WHERE project_id='3' AND name='View Project')),

step1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='Fund Project'),

step1 AS (SELECT events.id as event_id, events.user_id as event_user_id, CASE WHEN events.properties->>'' IS NULL THEN '$none'
WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END AS _group_key_0, CASE WHEN events.properties->>'' IS NULL
THEN '$none' WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END AS _group_key_1, 'Fund Project'::text
AS event_name  FROM events  WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599' AND
events.event_name_id IN (SELECT id FROM step1_names WHERE project_id='3' AND name='Fund Project')),

any_event AS ( SELECT event_id, event_user_id, event_name, _group_key_0, _group_key_1 FROM step0 UNION ALL
SELECT event_id, event_user_id, event_name, _group_key_0, _group_key_1 FROM step1),

users_any_event AS (SELECT event_name, CASE WHEN user_properties.properties->>'' IS NULL THEN '$none' WHEN
user_properties.properties->>'' = '' THEN '$none' ELSE user_properties.properties->>'' END AS _group_key_2,
_group_key_0, _group_key_1, event_user_id FROM any_event LEFT JOIN users ON any_event.event_user_id=users.id
LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id),

_group_key_0_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_0::numeric) + 0.00001 AS lbound,
percentile_disc(0.98) WITHIN GROUP(ORDER BY _group_key_0::numeric) AS ubound FROM users_any_event WHERE _group_key_0 != '$none'),

_group_key_2_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_2::numeric) + 0.00001 AS lbound,
percentile_disc(0.98) WITHIN GROUP(ORDER BY _group_key_2::numeric) AS ubound FROM users_any_event WHERE _group_key_2 != '$none'),

bucketed AS (SELECT event_name, COALESCE(NULLIF(_group_key_0, '$none'), 'NaN') AS _group_key_0, CASE WHEN _group_key_0 = '$none'
THEN -1 ELSE width_bucket(_group_key_0::numeric, _group_key_0_bounds.lbound::numeric, COALESCE(NULLIF(_group_key_0_bounds.ubound,
_group_key_0_bounds.lbound), _group_key_0_bounds.ubound+1)::numeric, 8) END AS _group_key_0_bucket, _group_key_1,
COALESCE(NULLIF(_group_key_2, '$none'), 'NaN') AS _group_key_2, CASE WHEN _group_key_2 = '$none' THEN -1 ELSE
width_bucket(_group_key_2::numeric, _group_key_2_bounds.lbound::numeric, COALESCE(NULLIF(_group_key_2_bounds.ubound,
_group_key_2_bounds.lbound), _group_key_2_bounds.ubound+1)::numeric, 8) END AS _group_key_2_bucket, event_user_id
FROM users_any_event, _group_key_0_bounds, _group_key_2_bounds)

SELECT event_name, COALESCE(NULLIF(concat(round(min(_group_key_0::numeric), 1), ' - ', round(max(_group_key_0::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_0,
_group_key_1, COALESCE(NULLIF(concat(round(min(_group_key_2::numeric), 1), ' - ', round(max(_group_key_2::numeric), 1)), 'NaN-NaN'), '$none') AS _group_key_2,
COUNT(*) AS count FROM bucketed GROUP BY _group_key_0_bucket, _group_key_1, _group_key_2_bucket, event_name ORDER BY event_name,
_group_key_0_bucket, _group_key_2_bucket, count DESC LIMIT 100000
*/
func buildEventsOccurrenceWithGivenEventQuery(projectID uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egSelect, egParams, egKeys := buildGroupKeys(eventGroupProps)
	isGroupByTimestamp := q.GetGroupByTimestamp() != ""

	filterSelect := joinWithComma(SelectDefaultEventFilter, egSelect)
	filterSelect = appendSelectTimestampIfRequired(filterSelect, q.GetGroupByTimestamp(), q.Timezone)

	refStepName := ""
	filters := make([]string, 0)
	for i, ewp := range q.EventsWithProperties {
		eventNameSelect := "'" + ewp.Name + "'" + "::text" + " AS event_name "
		filterSelect := joinWithComma(filterSelect, eventNameSelect)
		refStepName = fmt.Sprintf("step%d", i)
		filters = append(filters, refStepName)
		addFilterEventsWithPropsQuery(projectID, &qStmnt, &qParams, ewp, q.From, q.To, "",
			refStepName, filterSelect, egParams, "", "", "")
		if len(q.EventsWithProperties) > 1 {
			qStmnt = qStmnt + ", "
		}
	}

	// union.
	if len(filters) > 1 {
		// event_id is already unique.
		unionStepName := "any_event"
		unionStmnt := ""
		for _, filter := range filters {
			if unionStmnt != "" {
				unionStmnt = appendStatement(unionStmnt, "UNION ALL")
			}

			qSelect := appendSelectTimestampColIfRequired(SelectDefaultEventFilterByAlias, isGroupByTimestamp)
			qSelect = joinWithComma(qSelect, egKeys)
			unionStmnt = unionStmnt + " SELECT " + qSelect + " FROM " + filter
		}
		unionStmnt = as(unionStepName, unionStmnt)
		qStmnt = appendStatement(qStmnt, unionStmnt)

		refStepName = unionStepName
	}

	// count.
	userGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	_, _, groupKeys := buildGroupKeys(q.GroupByProperties)

	eventNameSelect := "event_name"
	groupKeys = joinWithComma(eventNameSelect, groupKeys)

	// select
	tSelect := joinWithComma(eventNameSelect, ugSelect, egKeys, "event_user_id")
	tSelect = appendSelectTimestampColIfRequired(tSelect, isGroupByTimestamp)

	termStmnt := "SELECT " + tSelect + " FROM " + refStepName
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id" +
			" " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}

	withUsersStepName := "users_any_event"
	qStmnt = joinWithComma(qStmnt, as(withUsersStepName, termStmnt))

	aggregateSelect := "SELECT "
	if isGroupByTimestamp {
		aggregateSelect = aggregateSelect + AliasDateTime + ", "
	}
	if hasNumericalGroupBy(q.GroupByProperties) {
		bucketedStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys := appendNumericalBucketingSteps(
			&qStmnt, q.GroupByProperties, withUsersStepName, eventNameSelect, isGroupByTimestamp, "event_user_id")
		aggregateGroupBys = append(aggregateGroupBys, eventNameSelect)
		aggregateSelectKeys = eventNameSelect + ", " + aggregateSelectKeys
		aggregateSelect = aggregateSelect + aggregateSelectKeys
		aggregateSelect = appendStatement(aggregateSelect, fmt.Sprintf("COUNT(*) AS %s FROM %s", AliasAggr, bucketedStepName))
		aggregateSelect = appendGroupByTimestampIfRequired(aggregateSelect, isGroupByTimestamp, strings.Join(aggregateGroupBys, ", "))
		aggregateSelect = aggregateSelect + fmt.Sprintf(" ORDER BY %s, %s", eventNameSelect, strings.Join(aggregateOrderBys, ", "))
	} else {
		aggregateSelect = aggregateSelect + groupKeys
		aggregateSelect = joinWithComma(aggregateSelect, fmt.Sprintf("COUNT(*) AS %s FROM %s", AliasAggr, withUsersStepName))
		aggregateSelect = appendGroupByTimestampIfRequired(aggregateSelect, isGroupByTimestamp, groupKeys)
		aggregateSelect = aggregateSelect + fmt.Sprintf(" ORDER BY %s", eventNameSelect)
	}

	aggregateSelect = aggregateSelect + fmt.Sprintf(", %s DESC", AliasAggr)
	aggregateSelect = appendLimitByCondition(aggregateSelect, q.GroupByProperties, isGroupByTimestamp)

	qParams = append(qParams, ugSelectParams...)
	qStmnt = appendStatement(qStmnt, aggregateSelect)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildEventCountForEachGivenEventsQueryNEW computes event count for each event with given filter and
breakdown.

Example: Query with date.
Group By: user_properties, event_properties

Sample query with ewp:
	$session
		"en": "event", "pr": "$source", "op": "equals", "va": "google", "ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
	MagazineViews
		"en": "event","pr": "$source","op": "equals","va": "google","ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
	www.livspace.com/in/hire-a-designer
		"en": "event","pr": "$source","op": "equals","va": "google","ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
gbp: [
	"pr": "$source","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "MagazineViews","eni": 2
	"pr": "$city","en": "user","pty": "categorical","ena": "www.livspace.com/in/hire-a-designer","eni": 3
	"pr": "$country","en": "user","pty": "categorical"
]
gbt: "date"

QUERY:

WITH

step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND name='$session'),

step_0 AS (SELECT events.id as event_id, events.user_id as event_user_id, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta') as datetime, '$session'::text AS event_name ,
CASE WHEN events.properties->>'$source' IS NULL THEN '$none' WHEN events.properties->>'$source' = ''
THEN '$none' ELSE events.properties->>'$source' END AS _group_key_0, CASE WHEN
events.properties->>'$campaign' IS NULL THEN '$none' WHEN events.properties->>'$campaign' = ''
THEN '$none' ELSE events.properties->>'$campaign' END AS _group_key_1 FROM events JOIN users ON
events.user_id=users.id LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
WHERE events.project_id='204' AND timestamp>='1583001000' AND timestamp<='1585679399' AND
events.event_name_id IN (SELECT id FROM step_0_names WHERE project_id='204' AND name='$session')
AND ( events.properties->>'$source' = 'google' AND user_properties.properties->>'$country' = 'India' )
ORDER BY event_id, _group_key_0, _group_key_1, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND name='MagazineViews'),

step_1 AS (SELECT events.id as event_id, events.user_id as event_user_id, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta') as datetime, 'MagazineViews'::text AS event_name ,
CASE WHEN events.properties->>'$campaign' IS NULL THEN '$none' WHEN events.properties->>'$campaign' = ''
THEN '$none' ELSE events.properties->>'$campaign' END AS _group_key_2 FROM events JOIN users ON
events.user_id=users.id LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
WHERE events.project_id='204' AND timestamp>='1583001000' AND timestamp<='1585679399' AND
events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='204' AND name='MagazineViews')
AND ( events.properties->>'$source' = 'google' AND user_properties.properties->>'$country' = 'India' )
ORDER BY event_id, _group_key_2, events.timestamp ASC),


step_2_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND
name='www.livspace.com/in/hire-a-designer'),

step_2 AS (SELECT events.id as event_id, events.user_id as event_user_id, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta') as datetime,
'www.livspace.com/in/hire-a-designer'::text AS event_name , CASE WHEN user_properties.properties->>'$city'
IS NULL THEN '$none' WHEN user_properties.properties->>'$city' = '' THEN '$none' ELSE
user_properties.properties->>'$city' END AS _group_key_3 FROM events JOIN users ON events.user_id=users.id
LEFT JOIN user_properties ON events.user_properties_id=user_properties.id WHERE events.project_id='204'
AND timestamp>='1583001000' AND timestamp<='1585679399' AND events.event_name_id IN (SELECT id FROM
step_2_names WHERE project_id='204' AND name='www.livspace.com/in/hire-a-designer') AND
( events.properties->>'$source' = 'google' AND user_properties.properties->>'$country' = 'India' )
ORDER BY event_id, _group_key_3, events.timestamp ASC),

each_events_union AS (SELECT step_0.event_name as event_name, step_0.event_id as event_id,
step_0.event_user_id as event_user_id, datetime , _group_key_0 as _group_key_0,  _group_key_1 as
_group_key_1,  ''  as _group_key_2,  ''  as _group_key_3 FROM step_0 UNION ALL SELECT step_1.event_name
as event_name, step_1.event_id as event_id, step_1.event_user_id as event_user_id, datetime ,
''  as _group_key_0,  ''  as  _group_key_1, _group_key_2 as _group_key_2,  ''  as _group_key_3
FROM step_1 UNION ALL SELECT step_2.event_name as event_name, step_2.event_id as event_id,
step_2.event_user_id as event_user_id, datetime ,  ''  as _group_key_0,  ''  as  _group_key_1,  ''
as _group_key_2, _group_key_3 as _group_key_3 FROM step_2) ,

each_users_union AS (SELECT each_events_union.event_user_id, each_events_union.event_id,
each_events_union.event_name, CASE WHEN user_properties.properties->>'$country' IS NULL THEN '$none'
WHEN user_properties.properties->>'$country' = '' THEN '$none' ELSE user_properties.properties->>'$country'
END AS _group_key_4, _group_key_0, _group_key_1, _group_key_2, _group_key_3, datetime FROM
each_events_union LEFT JOIN users ON each_events_union.event_user_id=users.id LEFT JOIN user_properties
ON users.id=user_properties.user_id AND user_properties.id=users.properties_id)

SELECT datetime, event_name, _group_key_0, _group_key_1, _group_key_2, _group_key_3, _group_key_4,
COUNT(event_id) AS count FROM each_users_union GROUP BY event_name, _group_key_0, _group_key_1,
_group_key_2, _group_key_3, _group_key_4, datetime ORDER BY count DESC LIMIT 100000
*/
func buildEventCountForEachGivenEventsQueryNEW(projectID uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps, stepsToKeysMap := addEventFilterStepsForEventCountQuery(projectID, &query, &qStmnt, &qParams)
	totalGroupKeys := 0
	for _, val := range stepsToKeysMap {
		totalGroupKeys = totalGroupKeys + len(val)
	}
	// union each event
	stepUsersUnion := "each_events_union"
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	var unionStmnt string

	for i, step := range steps {
		selectStr := fmt.Sprintf("%s.event_name as event_name, %s.event_id as event_id, %s.event_user_id as event_user_id", step, step, step)
		selectStr = appendSelectTimestampColIfRequired(selectStr, isGroupByTimestamp)
		egKeysForStep := getKeysForStep(step, steps, stepsToKeysMap, totalGroupKeys)
		if egKeysForStep != "" {
			selectStr = selectStr + " , " + egKeysForStep
		}
		selectStmnt := fmt.Sprintf("SELECT %s FROM %s", selectStr, step)
		if i == 0 {
			unionStmnt = selectStmnt
		} else {
			unionStmnt = unionStmnt + " UNION ALL " + selectStmnt
		}
	}

	qStmnt = joinWithComma(qStmnt, as(stepUsersUnion, unionStmnt))
	addEventCountAggregationQuery(&query, &qStmnt, &qParams, stepUsersUnion)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/* addEventFilterStepsForEventCountQuery builds step queries for each event including their filter
 and breakdown
	for ex:
Sample query with ewp:
	$session
		"en": "event", "pr": "$source", "op": "equals", "va": "google", "ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
	MagazineViews
		"en": "event","pr": "$source","op": "equals","va": "google","ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
gbp: [
	"pr": "$source","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "MagazineViews","eni": 2
	"pr": "$country","en": "user","pty": "categorical"
]
gbt: "date"

Steps returned:

step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND name='$session'),

step_0 AS (SELECT events.id as event_id, events.user_id as event_user_id, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta') as datetime, '$session'::text AS event_name ,
CASE WHEN events.properties->>'$source' IS NULL THEN '$none' WHEN events.properties->>'$source' = ''
THEN '$none' ELSE events.properties->>'$source' END AS _group_key_0, CASE WHEN
events.properties->>'$campaign' IS NULL THEN '$none' WHEN events.properties->>'$campaign' = ''
THEN '$none' ELSE events.properties->>'$campaign' END AS _group_key_1 FROM events JOIN users ON
events.user_id=users.id LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
WHERE events.project_id='204' AND timestamp>='1583001000' AND timestamp<='1585679399' AND
events.event_name_id IN (SELECT id FROM step_0_names WHERE project_id='204' AND name='$session')
AND ( events.properties->>'$source' = 'google' AND user_properties.properties->>'$country' = 'India' )
ORDER BY event_id, _group_key_0, _group_key_1, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='204' AND name='MagazineViews'),

step_1 AS (SELECT events.id as event_id, events.user_id as event_user_id, date_trunc('day',
to_timestamp(timestamp) AT TIME ZONE 'Asia/Calcutta') as datetime, 'MagazineViews'::text AS event_name ,
CASE WHEN events.properties->>'$campaign' IS NULL THEN '$none' WHEN events.properties->>'$campaign' = ''
THEN '$none' ELSE events.properties->>'$campaign' END AS _group_key_2 FROM events JOIN users ON
events.user_id=users.id LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
WHERE events.project_id='204' AND timestamp>='1583001000' AND timestamp<='1585679399' AND
events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='204' AND name='MagazineViews')
AND ( events.properties->>'$source' = 'google' AND user_properties.properties->>'$country' = 'India' )
ORDER BY event_id, _group_key_2, events.timestamp ASC),
*/
func addEventFilterStepsForEventCountQuery(projectID uint64, q *Query,
	qStmnt *string, qParams *[]interface{}) ([]string, map[string][]string) {

	var commonSelect string
	var commonOrderBy string
	stepsToKeysMap := make(map[string][]string)

	commonSelect = SelectDefaultEventFilter
	commonSelect = appendSelectTimestampIfRequired(commonSelect, q.GetGroupByTimestamp(), q.Timezone)

	if len(q.GroupByProperties) > 0 && commonOrderBy == "" {
		commonOrderBy = "event_id%s, events.timestamp ASC"
	}

	steps := make([]string, 0, 0)
	for i, ewp := range q.EventsWithProperties {
		refStepName := stepNameByIndex(i)
		steps = append(steps, refStepName)

		var stepSelect, stepOrderBy string
		var stepParams []interface{}
		var groupByUserProperties bool
		var stepGroupSelect, stepGroupKeys string
		var stepGroupParams []interface{}
		stepGroupSelect, stepGroupParams, stepGroupKeys, groupByUserProperties = buildGroupKeyForStep(
			&q.EventsWithProperties[i], q.GroupByProperties, i+1)

		eventSelect := commonSelect
		eventNameSelect := "'" + ewp.Name + "'" + "::text" + " AS event_name "
		eventSelect = joinWithComma(eventSelect, eventNameSelect)
		if stepGroupSelect != "" {
			stepSelect = eventSelect + ", " + stepGroupSelect
			stepOrderBy = fmt.Sprintf(commonOrderBy, ", "+stepGroupKeys)
			stepParams = stepGroupParams
			stepsToKeysMap[refStepName] = strings.Split(stepGroupKeys, ",")
		} else {
			stepSelect = eventSelect
			if commonOrderBy != "" {
				stepOrderBy = fmt.Sprintf(commonOrderBy, "")
			}
		}

		addJoinStmnt := "JOIN users ON events.user_id=users.id"
		if groupByUserProperties && !hasWhereEntity(ewp, PropertyEntityUser) {
			// If event has filter on user property, JOIN on user_properties is added in next step.
			// Skip addding here to avoid duplication.
			addJoinStmnt += " JOIN user_properties on events.user_properties_id=user_properties.id"
		}
		addFilterEventsWithPropsQuery(projectID, qStmnt, qParams, ewp, q.From, q.To,
			"", refStepName, stepSelect, stepParams, addJoinStmnt, "", stepOrderBy)

		if i < len(q.EventsWithProperties)-1 {
			*qStmnt = *qStmnt + ","
		}
	}

	return steps, stepsToKeysMap
}

/*
	addEventCountAggregationQuery applies global breakdown of user properties on each event query
	and presents query for selecting the final set of columns (data) from derived queries.
	for ex:
Sample query with ewp:
	$session
		"en": "event", "pr": "$source", "op": "equals", "va": "google", "ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
	MagazineViews
		"en": "event","pr": "$source","op": "equals","va": "google","ty": "categorical","lop": "AND"
		"en": "user","pr": "$country","op": "equals","va": "India","ty": "categorical","lop": "AND"
gbp: [
	"pr": "$source","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "$session","eni": 1
	"pr": "$campaign","en": "event","pty": "categorical","ena": "MagazineViews","eni": 2
	"pr": "$country","en": "user","pty": "categorical"
]
gbt: "date"

The union query and final select columns query are:

each_users_union AS (SELECT each_events_union.event_user_id, each_events_union.event_id,
each_events_union.event_name, CASE WHEN user_properties.properties->>'$country' IS NULL THEN '$none'
WHEN user_properties.properties->>'$country' = '' THEN '$none' ELSE user_properties.properties->>'$country'
END AS _group_key_4, _group_key_0, _group_key_1, _group_key_2, _group_key_3, datetime FROM
each_events_union LEFT JOIN users ON each_events_union.event_user_id=users.id LEFT JOIN user_properties
ON users.id=user_properties.user_id AND user_properties.id=users.properties_id)

SELECT datetime, event_name, _group_key_0, _group_key_1, _group_key_2, _group_key_3, _group_key_4,
COUNT(event_id) AS count FROM each_users_union GROUP BY event_name, _group_key_0, _group_key_1,
_group_key_2, _group_key_3, _group_key_4, datetime ORDER BY count DESC LIMIT 100000

*/
func addEventCountAggregationQuery(query *Query, qStmnt *string, qParams *[]interface{}, refStep string) {
	eventLevelGroupBys, otherGroupBys := separateEventLevelGroupBys(query.GroupByProperties)
	var egKeys string
	var unionStepName string

	_, _, egKeys = buildGroupKeys(eventLevelGroupBys)
	unionStepName = "each_users_union"

	// select
	userGroupProps := filterGroupPropsByType(otherGroupBys, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	*qParams = append(*qParams, ugSelectParams...)
	termSelect := ""
	termSelect = fmt.Sprintf(" %s.event_name, ", refStep)
	termSelect = termSelect + joinWithComma(ugSelect, egKeys)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	termSelect = appendSelectTimestampColIfRequired(termSelect, isGroupByTimestamp)

	termStmnt := fmt.Sprintf("SELECT %s.event_user_id, %s.event_id, ", refStep, refStep) + termSelect + " FROM " + refStep
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStep + ".event_user_id=users.id"
		termStmnt = termStmnt + " " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}

	_, _, groupKeys := buildGroupKeys(query.GroupByProperties)

	termStmnt = as(unionStepName, termStmnt)
	var aggregateFromStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys string
	if hasNumericalGroupBy(query.GroupByProperties) {
		eventName := AliasEventName
		bucketedStepName, bucketedSelectKeys, bucketedGroupBys, bucketedOrderBys := appendNumericalBucketingSteps(
			&termStmnt, query.GroupByProperties, unionStepName, eventName, isGroupByTimestamp, "event_id")
		aggregateFromStepName = bucketedStepName
		aggregateSelectKeys = bucketedSelectKeys
		aggregateGroupBys = strings.Join(bucketedGroupBys, ", ")
		aggregateOrderBys = strings.Join(bucketedOrderBys, ", ")
		*qStmnt = appendStatement(*qStmnt, ", "+termStmnt)
	} else {

		if groupKeys != "" {
			// Order by count, which will added later.
			aggregateFromStepName = unionStepName
			aggregateSelectKeys = groupKeys + ", "
			aggregateGroupBys = groupKeys
			*qStmnt = appendStatement(*qStmnt, ", "+termStmnt)
		} else {
			// No group by clause added. Use previous step and rest all leave empty.
			aggregateFromStepName = refStep
		}
	}

	aggregateSelect := "SELECT "
	if isGroupByTimestamp {
		aggregateSelect = aggregateSelect + AliasDateTime + ", "
		aggregateGroupBys = joinWithComma(aggregateGroupBys, AliasDateTime)
	}
	// adding select event_name and group by event_name for each event-user count
	if query.EventsCondition == EventCondEachGivenEvent {
		aggregateSelect = aggregateSelect + AliasEventName + ", "
		aggregateGroupBys = joinWithComma(AliasEventName, aggregateGroupBys)
	}
	aggregateSelect = aggregateSelect + aggregateSelectKeys + fmt.Sprintf("COUNT(event_id) AS %s FROM %s",
		AliasAggr, aggregateFromStepName)

	aggregateSelect = appendGroupBy(aggregateSelect, aggregateGroupBys)
	if aggregateOrderBys != "" {
		aggregateSelect = aggregateSelect + " ORDER BY " + aggregateOrderBys
	} else {
		aggregateSelect = appendOrderByAggr(aggregateSelect)
	}
	aggregateSelect = appendLimitByCondition(aggregateSelect, query.GroupByProperties, isGroupByTimestamp)
	*qStmnt = appendStatement(*qStmnt, aggregateSelect)
}
