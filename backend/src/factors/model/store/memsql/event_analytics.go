package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	MetaEachEventCountMetrics = "EachEventCount"
	MetaEventInfo             = "MetaEventInfo"
	AllowedGoroutines         = 4
)

type ResultGroup struct {
	Results []model.QueryResult `json:"result_group"`
}

func (store *MemSQL) RunEventsGroupQuery(queriesOriginal []model.Query,
	projectId int64, enableFilterOpt bool) (model.ResultGroup, int) {

	logFields := log.Fields{
		"queries_original": queriesOriginal,
		"project_id":       projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	queries := make([]model.Query, 0, 0)
	U.DeepCopy(&queriesOriginal, &queries)

	var resultGroup model.ResultGroup
	resultGroup.Results = make([]model.QueryResult, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	actualRoutineLimit := U.MinInt(len(queries), AllowedGoroutines)
	waitGroup.Add(actualRoutineLimit)
	for index, query := range queries {
		count++
		go store.runSingleEventsQuery(projectId, query, &resultGroup.Results[index], &waitGroup, enableFilterOpt)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	for _, result := range resultGroup.Results {
		if result.Headers == nil {
			return resultGroup, http.StatusInternalServerError
		}
		if result.Headers[0] == model.AliasError {
			return resultGroup, http.StatusPartialContent
		}
	}
	return resultGroup, http.StatusOK
}

func (store *MemSQL) runSingleEventsQuery(projectId int64, query model.Query,
	resultHolder *model.QueryResult, waitGroup *sync.WaitGroup, enableFilterOpt bool) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectId,
		"wait_group": waitGroup,
	}
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	defer waitGroup.Done()
	result, errCode, errMsg := store.ExecuteEventsQuery(projectId, query, enableFilterOpt)
	if errCode != http.StatusOK {
		errorResult := buildErrorResult(errMsg)
		*resultHolder = *errorResult
	} else {
		*resultHolder = *result
	}
	return
}

func (store *MemSQL) ExecuteEventsQuery(projectId int64, query model.Query, enableFilterOpt bool) (*model.QueryResult, int, string) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	if valid, errMsg := IsValidEventsQuery(&query); !valid {
		return nil, http.StatusBadRequest, errMsg
	}
	return store.RunInsightsQuery(projectId, query, enableFilterOpt)
}

func (store *MemSQL) fillEventNameIDs(projectID int64, query *model.Query) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for i := range query.EventsWithProperties {
		eventNames, status := store.GetEventNamesByNames(projectID, []string{query.EventsWithProperties[i].Name})
		if status != http.StatusFound {
			log.WithFields(log.Fields{"project_id": projectID, "event_name": query.EventsWithProperties[i].Name}).
				Error("Failed to get event names in fillEventNameIDs. Continuing with empty uuid.")
			query.EventsWithProperties[i].EventNameIDs = []interface{}{""}
			continue
		}

		for j := range eventNames {
			query.EventsWithProperties[i].EventNameIDs = append(query.EventsWithProperties[i].EventNameIDs, eventNames[j].ID)
		}
	}

}

func logDifferenceIfAny(hash1, hash2 string, query model.Query, logMessage string) {
	if hash1 != hash2 {
		log.WithField("query", query).Warn(logMessage)
	}
}

func (store *MemSQL) RunInsightsQuery(projectId int64, query model.Query, enableFilterOpt bool) (*model.QueryResult, int, string) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if C.SkipEventNameStepByProjectID(projectId) {
		store.fillEventNameIDs(projectId, &query)
	}

	hash1, _ := query.GetQueryCacheHashString()
	stmnt, params, err := store.BuildInsightsQuery(projectId, query, enableFilterOpt)
	hash2, _ := query.GetQueryCacheHashString()
	if err != nil {
		log.WithError(err).Error(model.ErrMsgQueryProcessingFailure)
		return &model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	logCtx := log.WithFields(logFields)

	if stmnt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return &model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	hash1, _ = query.GetQueryCacheHashString()
	result, err, reqID := store.ExecQuery(stmnt, params)
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return &model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, query, "Query change - 02")


	startComputeTime := time.Now()
	groupPropsLen := len(query.GroupByProperties)
	err = LimitQueryResult(projectId, result, groupPropsLen, query.GetGroupByTimestamp() != "")
	if err != nil {
		logCtx.WithError(err).Error("Failed processing query results for limiting.")
		return &model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	model.AddMissingEventNamesInResult(result, &query)
	hash1, _ = query.GetQueryCacheHashString()
	err = SanitizeQueryResult(result, &query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to sanitize query results.")
		return &model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, query, "Query change - 03")

	hash1, _ = query.GetQueryCacheHashString()
	// Replace the event_name with alias, if the event condition is each_given_event
	if query.EventsCondition == model.EventCondEachGivenEvent {
		model.AddAliasNameOnEventCondEachGivenEventQueryResult(result, query)
	}
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, query, "Query change - 04")

	hash1, _ = query.GetQueryCacheHashString()
	// if and only if breakdown is by datetime (condition for both events/users count for each event.)
	if len(query.GroupByProperties) == 0 &&
		query.EventsCondition == model.EventCondEachGivenEvent &&
		query.GroupByTimestamp != nil && query.GroupByTimestamp.(string) != "" {

		result, err = transformResultsForEachEventQuery(result, query)
		addEventMetricsMetaToQueryResult(result)

		if err != nil {
			logCtx.WithError(err).Error("Failed to transform query results.")
			return &model.QueryResult{}, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
		}
	} else if query.EventsCondition == model.EventCondEachGivenEvent &&
		!strings.Contains(strings.Join(result.Headers, ","), model.AliasEventName) {
		// for data consistency: single event result has no event_name column, appending here
		if len(query.EventsWithProperties) == 1 {
			result.Headers = append(result.Headers, model.AliasEventName)
			for i := range result.Rows {
				result.Rows[i] = append(result.Rows[i], query.EventsWithProperties[0].Name)
			}
		}
	} else if query.EventsCondition == model.EventCondEachGivenEvent {
		// add event name header and fill rows
		addEventNameIndexInResult(result)

	} else {
		// removing index from event name for old queries.
		eventNameIndex := -1
		for i, key := range result.Headers {
			if key == model.AliasEventName {
				eventNameIndex = i
			}
		}
		if eventNameIndex != -1 {
			for i, row := range result.Rows {
				eventName := row[eventNameIndex].(string)
				splitPos := strings.Index(eventName, "_")
				result.Rows[i][eventNameIndex] = eventName[splitPos+1:]
			}
		}
	}
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, query, "Query change - 05")

	addQueryToResultMeta(result, query)
	U.LogComputeTimeWithQueryRequestID(startComputeTime, reqID, &logFields)

	return result, http.StatusOK, "Successfully executed query"
}

// buildErrorResult takes the failure msg and wraps it into a model.QueryResult object
func buildErrorResult(errMsg string) *model.QueryResult {
	logFields := log.Fields{
		"err_msg": errMsg,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	errMsg = "Query failed:" + " - " + errMsg
	headers := []string{model.AliasError}
	rows := make([][]interface{}, 0, 0)
	row := make([]interface{}, 0, 0)
	row = append(row, errMsg)
	rows = append(rows, row)
	errorResult := &model.QueryResult{Headers: headers, Rows: rows}
	return errorResult
}

// updateEventNameInHeaderAndAddMeta makes header from 0_$session to $session
// and adds event's index, name and headerIndex in meta
func updateEventNameInHeaderAndAddMeta(result *model.QueryResult) {
	logFields := log.Fields{}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var rows [][]interface{}
	for i, key := range result.Headers {
		if key != model.AliasDateTime && key != model.AliasEventIndex && key != model.AliasAggr {

			splitPos := strings.Index(key, "_")
			eventIndex, err := strconv.Atoi(key[0:splitPos])
			eventName := key[splitPos+1:]
			row := []interface{}{i, eventIndex, eventName}
			rows = append(rows, row)
			if err == nil {
				result.Headers[i] = eventName
			} else {
				log.WithError(err).Error("failed to convert string to integer")
			}
		}
	}
	metaMetricsEventMeta := model.HeaderRows{Title: MetaEventInfo,
		Headers: []string{"HeaderIndex", "EventIndex", "EventName"}, Rows: rows}
	result.Meta.MetaMetrics = append(result.Meta.MetaMetrics, metaMetricsEventMeta)
}

// addEventNameIndexInResult adds event_index and fills up the rows accordingly
func addEventNameIndexInResult(result *model.QueryResult) {
	logFields := log.Fields{}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	eventNameIndex := -1
	for i, key := range result.Headers {
		if key == model.AliasEventName {
			eventNameIndex = i
		}
	}
	for i, row := range result.Rows {
		// modifying event_name value
		eventName := row[eventNameIndex].(string)
		splitPos := strings.Index(eventName, "_")
		row[eventNameIndex] = eventName[splitPos+1:]

		// adding event_index value
		tempRow := []interface{}{0}
		tempRow = append(tempRow, result.Rows[i]...)
		result.Rows[i] = tempRow
		indexStr := eventName[0:splitPos]
		eventIndex, err := strconv.Atoi(indexStr)
		if err == nil {
			result.Rows[i][0] = eventIndex
		} else {
			log.WithError(err).Error("failed to convert string to integer")
		}
	}

	// adding event_index column
	newHeader := []string{model.AliasEventIndex}
	newHeader = append(newHeader, result.Headers...)
	result.Headers = newHeader
}

// transformResultsForEachEventQuery transforms model.QueryResult with new header as datetime and events
func transformResultsForEachEventQuery(oldResult *model.QueryResult, query model.Query) (*model.QueryResult, error) {
	logFields := log.Fields{
		"query":      query,
		"old_result": oldResult,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// for single event, oldResult has no 'event_name' column but only 2 columns: {datetime, count}
	// adding  'event_name' with row values for standard transformation
	if len(oldResult.Headers) == 2 {
		if len(query.EventsWithProperties) > 0 {
			tempHeader := []string{model.AliasEventName}
			tempHeader = append(tempHeader, oldResult.Headers...)
			oldResult.Headers = tempHeader
			oldResult.Headers = append(oldResult.Headers, model.AliasEventName)
			for i := range oldResult.Rows {
				tempRow := []interface{}{"0_" + query.EventsWithProperties[0].Name}
				tempRow = append(tempRow, oldResult.Rows[i])
				oldResult.Rows[i] = tempRow
			}
		}
	}

	eventNameIndex := 0
	dateIndex := 0
	countIndex := 0
	for i, key := range oldResult.Headers {
		if key == model.AliasEventName {
			eventNameIndex = i
		}
		if key == model.AliasDateTime {
			dateIndex = i
		}
		if key == model.AliasAggr {
			countIndex = i
		}
	}

	eventsHeaderIndexMap := make(map[string]int)
	for _, row := range oldResult.Rows {
		eventName := ""
		if row[eventNameIndex] == nil && len(query.EventsWithProperties) == 1 {
			if query.EventsWithProperties[0].AliasName != "" {
				eventName = "0_" + query.EventsWithProperties[0].AliasName
			} else {
				eventName = "0_" + query.EventsWithProperties[0].Name
			}
		} else {
			eventName = row[eventNameIndex].(string)
		}
		// initial header value = 1
		eventsHeaderIndexMap[eventName] = 1
	}

	// headers : datetime, event1, event2, ...
	newResultHeaders := []string{model.AliasDateTime}
	newResultRows := make([][]interface{}, 0, 0)

	// skipping 0 as it is index is for 'datetime' header
	headerIndex := 1
	for name := range eventsHeaderIndexMap {
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
			eventName = "0_" + query.EventsWithProperties[0].Name
			if query.EventsWithProperties[0].AliasName != "" {
				eventName = "0_" + query.EventsWithProperties[0].AliasName
			}
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

	newResult := &model.QueryResult{Headers: newResultHeaders, Rows: newResultRows}

	updateEventNameInHeaderAndAddMeta(newResult)

	// Below piece of transformation re-orders the result w.r.t to the query order of events.
	// This uses MetaEventInfo, which is a prior meta of result stating the order of events in query and in result.
	metaMetricsEventMeta := model.HeaderRows{}
	for _, val := range newResult.Meta.MetaMetrics {
		if val.Title == MetaEventInfo {
			metaMetricsEventMeta = val
		}
	}
	if metaMetricsEventMeta.Title == MetaEventInfo {

		queryIndexToName := make(map[int]string)
		minEventIndex := 100000
		for _, row := range metaMetricsEventMeta.Rows {
			// "HeaderIndex(current res order)", "EventIndex(query event order)", "EventName"
			queryIndexToName[row[1].(int)] = row[2].(string)
			if minEventIndex > row[0].(int) {
				minEventIndex = row[0].(int)
			}
		}

		finalIndexToOldIndexMap := make(map[int]int)
		// Copy non-event index which will remain same
		for i := 0; i < minEventIndex; i++ {
			finalIndexToOldIndexMap[i] = i
		}
		for _, row := range metaMetricsEventMeta.Rows {
			finalIndexToOldIndexMap[row[1].(int)+minEventIndex] = row[0].(int)
		}

		var finalResultHeaders []string
		finalResultRows := make([][]interface{}, 0, 0)
		// transform the headers w.r.t the events order in query
		for idx, _ := range newResultHeaders {
			finalResultHeaders = append(finalResultHeaders, newResultHeaders[finalIndexToOldIndexMap[idx]])
		}

		// copying the data to finalResultRows
		for _, row := range newResultRows {
			newRow := make([]interface{}, 0, 0)
			for _, colValue := range row {
				newRow = append(newRow, colValue)
			}
			finalResultRows = append(finalResultRows, newRow)
		}

		// Rearranging the data to final Result
		for rowNo, row := range newResultRows {
			for colNo, _ := range row {
				finalResultRows[rowNo][colNo] = newResultRows[rowNo][finalIndexToOldIndexMap[colNo]]
			}
		}
		finalResult := &model.QueryResult{Headers: finalResultHeaders, Rows: finalResultRows, Meta: newResult.Meta}
		return finalResult, nil
	}

	return newResult, nil
}

// addEventMetricsMetaToQueryResult adds meta metrics in query result based on query type, event
// condition and group by inputs
func addEventMetricsMetaToQueryResult(result *model.QueryResult) {
	logFields := log.Fields{}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	metaMetricsEventCount := model.HeaderRows{}
	metaMetricsEventCount.Title = MetaEachEventCountMetrics
	// headers : event1, event2, ...
	metaMetricsEventCount.Headers = []string{}
	headerIndexToEventName := make(map[int]string)
	for i, key := range result.Headers {
		// skipping datetime and including all event names as header
		if key != model.AliasDateTime {
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
func (store *MemSQL) BuildInsightsQuery(projectId int64, query model.Query, enableFilterOpt bool) (string, []interface{}, error) {
	logFields := log.Fields{
		"query":      query,
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	addIndexToGroupByProperties(&query)

	if query.Type == model.QueryTypeEventsOccurrence {
		if query.EventsCondition == model.EventCondEachGivenEvent {
			return store.buildEventCountForEachGivenEventsQueryNEW(projectId, query, enableFilterOpt)
		}

		if len(query.EventsWithProperties) == 1 {
			return buildEventsOccurrenceSingleEventQuery(projectId, query, enableFilterOpt)
		}

		return buildEventsOccurrenceWithGivenEventQuery(projectId, query, enableFilterOpt)
	}

	if query.Type == model.QueryTypeUniqueUsers {
		if len(query.EventsWithProperties) == 1 {
			return store.buildUniqueUsersSingleEventQuery(projectId, query, enableFilterOpt)
		}

		if query.EventsCondition == model.EventCondAnyGivenEvent {
			return store.buildUniqueUsersWithAnyGivenEventsQuery(projectId, query, enableFilterOpt)
		}

		if query.EventsCondition == model.EventCondAllGivenEvent {
			return store.buildUniqueUsersWithAllGivenEventsQuery(projectId, query, enableFilterOpt)
		}

		if query.EventsCondition == model.EventCondEachGivenEvent {
			return store.buildUniqueUsersWithEachGivenEventsQuery(projectId, query, enableFilterOpt)
		}
	}

	return "", nil, errors.New("invalid query")
}

func LimitQueryResult(projectID int64, result *model.QueryResult, groupPropsLen int, groupByTimestamp bool) error {
	logFields := log.Fields{
		"group_props_len": groupPropsLen,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if groupPropsLen > 0 && groupByTimestamp {
		return limitGroupByTimestampResult(projectID, result, groupByTimestamp)
	}

	if groupPropsLen > 1 {
		return limitMultiGroupByPropertiesResult(projectID, result, groupByTimestamp)
	}

	// Others limited on SQL Query.
	return nil
}

// Limits top results and makes sure same group key combination available on different
// datetime, if exists on SQL result. Assumes result is sorted by count. Preserves all
// datetime for the limited combination of group keys.
func limitGroupByTimestampResult(projectID int64, result *model.QueryResult, groupByTimestamp bool) error {
	logFields := log.Fields{
		"group_by_timestamp": groupByTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
		maxLimit := model.ResultsLimit
		if C.IsKPILimitIncreaseAllowedForProject(projectID) {
			maxLimit = model.MaxResultsLimit
		}

		if !keyExists && len(keyLookup) < maxLimit {
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
func limitMultiGroupByPropertiesResult(projectID int64, result *model.QueryResult, groupByTimestamp bool) error {
	logFields := log.Fields{
		"group_by_timestamp": groupByTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
		maxLimit := model.ResultsLimit
		if C.IsKPILimitIncreaseAllowedForProject(projectID) {
			maxLimit = model.MaxResultsLimit
		}

		if !leftKeyExists && len(keyLookup) < maxLimit {
			keyLookup[leftKey] = make(map[interface{}]bool, 0)
			leftKeyExists = true
		}

		var rightKeyExists bool
		if leftKeyExists {
			// Limits no.of right keys to ResultsLimit.
			_, rightKeyExits := keyLookup[leftKey][row[leftKeyEnd]]

			if !rightKeyExits && len(keyLookup[leftKey]) < maxLimit {
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

// SanitizeQueryResult Converts DB results into plottable query results.
func SanitizeQueryResult(result *model.QueryResult, query *model.Query) error {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if query.GetGroupByTimestamp() != "" {
		if err := sanitizeGroupByTimestampResult(result, query); err != nil {
			return err
		}
	}

	// Replace group keys with real column names. should be last step.
	// of sanitization.
	if err := translateGroupKeysIntoColumnNames(result, query.GroupByProperties); err != nil {
		return err
	}

	if isGroupByTypeWithBuckets(query.GroupByProperties) {
		sanitizeNumericalBucketRanges(result, query)
	}

	if model.HasGroupByDateTypeProperties(query.GroupByProperties) {
		model.SanitizeDateTypeRows(result, query)
	}
	return nil
}

func sanitizeGroupByTimestampResult(result *model.QueryResult, query *model.Query) error {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	aggrIndex, timeIndex, err := GetTimstampAndAggregateIndexOnQueryResult(result.Headers)
	if err != nil {
		return err
	}
	transformTimeValueInResults(result, timeIndex, query.Timezone)

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

func transformTimeValueInResults(result *model.QueryResult, timeIndex int, timezone string) {

	for index := range result.Rows {

		timestampWithTimezone := U.GetTimestampAsStrWithTimezone(
			result.Rows[index][timeIndex].(time.Time), timezone)

		// overrides timestamp with user timezone as sql results doesn't
		// return timezone used to query.
		// when broken by monthly, data is present for 2022-04-01T00:00:00+11:00
		// when broken by day, data is present for 2022-04-03T00:00:00+10:00, where the time got changed at 2:00.
		// Hence converting it using offsets.
		ts := U.GetTimeFromTimestampStr(timestampWithTimezone)
		offset := U.GetTimezoneOffsetFromString(ts, timezone)
		currTimeStrFromOffset := U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offset)
		currTimeFromOffset := U.GetTimeFromParseTimeStr(currTimeStrFromOffset)

		result.Rows[index][timeIndex] = currTimeFromOffset
	}
}

func sortResultRowsByTimestamp(resultRows [][]interface{}, timestampIndex int) {
	logFields := log.Fields{
		"result_rows":     resultRows,
		"timestamp_index": timestampIndex,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	sort.Slice(resultRows, func(i, j int) bool {
		return (resultRows[i][timestampIndex].(time.Time)).Unix() <
			(resultRows[j][timestampIndex].(time.Time)).Unix()
	})
}

func sortChannelResultRowsByTimestamp(resultRows [][]interface{}, timestampIndex int) {
	logFields := log.Fields{
		"result_rows":     resultRows,
		"timestamp_index": timestampIndex,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	sort.Slice(resultRows, func(i, j int) bool {
		ts1, _ := U.GetTimeFromParseTimeStrWithErrorFromInterface(resultRows[i][timestampIndex])
		ts2, _ := U.GetTimeFromParseTimeStrWithErrorFromInterface(resultRows[i][timestampIndex])
		return ts1.Unix() < ts2.Unix()
	})
}

// In day light savings, the timezone gets changed at 1:00AM or 2:00AM. Offsets are relied upon the start of Timestamp consideration.
func GetAllTimestampsAndOffsetBetweenByType(from, to int64, typ, timezone string) ([]time.Time, []string) {
	logFields := log.Fields{
		"from":     from,
		"to":       to,
		"typ":      typ,
		"timezone": timezone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if typ == model.GroupByTimestampDate {
		return U.GetAllDatesAndOffsetAsTimestamp(from, to, timezone)
	}

	if typ == model.GroupByTimestampHour {
		return U.GetAllHoursAsTimestamp(from, to, timezone)
	}

	if typ == model.GroupByTimestampWeek {
		return U.GetAllWeeksAsTimestamp(from, to, timezone)
	}

	if typ == model.GroupByTimestampMonth {
		return U.GetAllMonthsAsTimestamp(from, to, timezone)
	}

	if typ == model.GroupByTimestampQuarter {
		return U.GetAllQuartersAsTimestamp(from, to, timezone)
	}

	return []time.Time{}, []string{}
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

func (store *MemSQL) addEventFilterStepsForUniqueUsersQuery(projectID int64, q *model.Query,
	qStmnt *string, qParams *[]interface{}, enableFilterOpt bool) ([]string, map[string][]string) {
	logFields := log.Fields{
		"project_id": projectID,
		"q":          q,
		"q_stmnt":    qStmnt,
		"q_params":   qParams,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var commonSelect string
	var commonOrderBy string
	var commonGroupBy string
	stepsToKeysMap := make(map[string][]string)

	var aggregatePropertyDetails model.QueryGroupByProperty
	aggregateKey := model.AliasAggr
	aggregatePropertyDetails.Property = q.AggregateProperty
	aggregatePropertyDetails.Entity = q.AggregateEntity
	var aggregateSelect string
	var aggregateParams []interface{}

	if q.Caller == model.USER_PROFILE_CALLER {
		commonSelect = fmt.Sprintf("COALESCE(users.customer_user_id, events.user_id) as coal_user_id%%, users.updated_at as last_activity, ISNULL(users.customer_user_id) AS is_anonymous, users.properties as properties")
		commonSelect = strings.ReplaceAll(commonSelect, "%", "%s")
	} else if q.Caller == model.ACCOUNT_PROFILE_CALLER {
		commonSelect = fmt.Sprintf("events.user_id as identity%%, users.updated_at as last_activity, users.properties as properties")
		commonSelect = strings.ReplaceAll(commonSelect, "%", "%s")
	} else {
		if q.GetGroupByTimestamp() != "" {
			selectTimestamp := getSelectTimestampByType(q.GetGroupByTimestamp(), q.Timezone)
			// select and order by with datetime.
			commonSelect = fmt.Sprintf("COALESCE(users.customer_user_id, events.user_id) as coal_user_id%%, %s as %s,", selectTimestamp, model.AliasDateTime) +
				" FIRST(events.user_id, FROM_UNIXTIME(events.timestamp)) as event_user_id"

			commonSelect = strings.ReplaceAll(commonSelect, "%", "%s")

			commonOrderBy = fmt.Sprintf("coal_user_id%%, %s, events.timestamp ASC", model.AliasDateTime)
			commonOrderBy = strings.ReplaceAll(commonOrderBy, "%", "%s")
			commonGroupBy = "datetime, coal_user_id"
		} else {
			// default select.
			commonSelect = "COALESCE(users.customer_user_id, events.user_id)" +
				" as coal_user_id%s, FIRST(events.user_id, FROM_UNIXTIME(events.timestamp)) as event_user_id"
			commonGroupBy = "coal_user_id"

		}
	}

	var commonSelectArr []string
	for i := range q.EventsWithProperties {
		_, status := store.IsGroupEventNameByQueryEventWithProperties(projectID, q.EventsWithProperties[i])
		stepCommonSelect := ""
		if status == http.StatusFound {
			stepCommonSelect = strings.ReplaceAll(commonSelect, "events.user_id", "users.users_user_id")
		} else {
			stepCommonSelect = commonSelect
		}

		commonSelectArr = append(commonSelectArr, stepCommonSelect)
	}

	if q.AggregateProperty != "" && q.AggregateProperty != "1" {
		for i := range commonSelectArr {
			aggregateSelect, aggregateParams = getNoneHandledGroupBySelect(projectID, aggregatePropertyDetails, aggregateKey, q.Timezone)
			commonSelectArr[i] = commonSelectArr[i] + ", " + aggregateSelect
		}
		*qParams = append(*qParams, aggregateParams...)
	}

	if len(q.GroupByProperties) > 0 && commonOrderBy == "" {
		// Using first occurred event_properties after distinct on user_id.
		commonOrderBy = "coal_user_id%s, events.timestamp ASC"
	}

	steps := make([]string, 0, 0)
	// Adding source string
	var addSourceStmt, addColsString string
	if IsCallerProfiles(q.Caller) {
		addSourceStmt, addColsString = store.addSourceFilterForSegments(projectID, q.Source, q.Caller)
	}
	for i, ewp := range q.EventsWithProperties {
		refStepName := stepNameByIndex(i)
		steps = append(steps, refStepName)

		var stepSelect, stepOrderBy, stepGroupBy string
		var stepParams []interface{}
		var stepGroupSelect, stepGroupKeys string
		var stepGroupParams []interface{}
		stepGroupSelect, stepGroupParams, stepGroupKeys, _ = buildGroupKeyForStep(
			projectID, &q.EventsWithProperties[i], q.GroupByProperties, i+1, q.Timezone)

		eventSelect := commonSelectArr[i]
		eventParam := ""
		if q.EventsCondition == model.EventCondEachGivenEvent && !IsCallerProfiles(q.Caller) {
			eventNameSelect := fmt.Sprintf("? AS event_name ")
			eventParam = fmt.Sprintf("%s_%s", strconv.Itoa(i), ewp.Name)
			eventSelect = joinWithComma(eventSelect, eventNameSelect)
		}
		if stepGroupSelect != "" {
			stepSelect = fmt.Sprintf(eventSelect, ", "+stepGroupSelect)
			stepParams = append(stepParams, stepGroupParams...)
			stepOrderBy = fmt.Sprintf(commonOrderBy, ", "+stepGroupKeys)
			stepGroupBy = joinWithComma(commonGroupBy, stepGroupKeys)
			if q.EventsCondition == model.EventCondEachGivenEvent {
				stepParams = append(stepParams, eventParam)
			}
			stepsToKeysMap[refStepName] = strings.Split(stepGroupKeys, ",")
		} else {
			stepSelect = fmt.Sprintf(eventSelect, "")
			if q.EventsCondition == model.EventCondEachGivenEvent && !IsCallerProfiles(q.Caller) {
				stepParams = append(stepParams, eventParam)
			}
			if commonOrderBy != "" {
				stepOrderBy = fmt.Sprintf(commonOrderBy, "")
			}
			stepGroupBy = commonGroupBy
		}

		// Default join statement for users.
		addJoinStmnt := "JOIN users ON events.user_id=users.id AND users.project_id = ?"

		// Join support for original users of group.
		groupName, status := store.IsGroupEventNameByQueryEventWithProperties(projectID, q.EventsWithProperties[i])
		if status == http.StatusFound {
			group, status := store.GetGroup(projectID, groupName)
			if status == http.StatusFound {
				if q.Caller == model.ACCOUNT_PROFILE_CALLER {
					addJoinStmnt = "LEFT JOIN users ON events.user_id=users.id AND users.project_id = ?"
				} else {
					addJoinStmnt = fmt.Sprintf("LEFT JOIN users ON events.user_id=users.group_%d_user_id AND users.project_id = ? ", group.ID)
				}
			} else {
				log.WithField("project_id", projectID).WithField("group", groupName).
					Error("Failed to find group on analytical query execution.")
			}
		}

		stepParams = append(stepParams, projectID)

		addFilterFunc := addFilterEventsWithPropsQuery
		if enableFilterOpt {
			addFilterFunc = addFilterEventsWithPropsQueryV2
		}
		addFilterFunc(projectID, qStmnt, qParams, ewp, q.From, q.To,
			"", refStepName, stepSelect, stepParams, addJoinStmnt, stepGroupBy, stepOrderBy, q.GlobalUserProperties)

		// adding source check
		if IsCallerProfiles(q.Caller) {
			if C.EnableOptimisedFilterOnEventUserQuery() {
				if i == 0 {
					addSourceStmt = strings.ReplaceAll(addSourceStmt, "_event_users_view.", fmt.Sprintf("%s_event_users_view.", refStepName))
				} else {
					addSourceStmt = strings.ReplaceAll(addSourceStmt, fmt.Sprintf("step_%d_event_users_view.", i-1), fmt.Sprintf("%s_event_users_view.", refStepName))

				}
				if len(q.GlobalUserProperties) == 0 && len(ewp.Properties) == 0 {
					*qStmnt = strings.TrimSuffix(*qStmnt, ")") + " WHERE" + addSourceStmt + ")"
				} else {
					*qStmnt = strings.TrimSuffix(*qStmnt, ")") + " AND" + addSourceStmt + ")"
				}
			} else {
				*qStmnt = strings.TrimSuffix(*qStmnt, ")") + " AND" + addSourceStmt + ")"
			}
		}
		if i < len(q.EventsWithProperties)-1 {
			*qStmnt = *qStmnt + ","
		}
	}

	// adding source columns
	if IsCallerProfiles(q.Caller) && C.EnableOptimisedFilterOnEventUserQuery() {
		qStmtSplit := strings.Split(*qStmnt, "(SELECT")
		result := qStmtSplit[0] + "(SELECT" + qStmtSplit[1]
		for idx := 2; idx < len(qStmtSplit); idx++ {
			if idx%2 == 0 {
				result = result + "(SELECT " + addColsString + ", " + qStmtSplit[idx]
			} else {
				result = result + "(SELECT " + qStmtSplit[idx]
			}
		}
		*qStmnt = result
	}

	return steps, stepsToKeysMap
}

func IsCallerProfiles(caller string) bool {
	return (caller == model.USER_PROFILE_CALLER || caller == model.ACCOUNT_PROFILE_CALLER)
}

// Adds source string, Example
// WITH  step_0 AS (SELECT users.id as identity, users.updated_at as last_activity, users.properties as properties
// FROM events  JOIN users ON events.user_id=users.id AND users.project_id = '2000009' WHERE events.project_id='2000009'
// AND timestamp='1671686385' AND timestamp='1674105588' AND  ( events.event_name_id = '061b66b9-ea69-4aae-bd09-551e868ba320' )
// AND ( (JSON_EXTRACT_STRING(events.properties, '$salesforce_campaign_type') = 'Some Salesforce Type') )
// AND (users.is_group_user=1) AND users.group_1_id IS NOT NULL), step_1 AS (SELECT users.id as identity,
// users.updated_at as last_activity, users.properties as properties FROM events  JOIN users ON
// events.user_id=users.id AND users.project_id = '2000009' WHERE events.project_id='2000009' AND timestamp='1671686385'
// AND timestamp='1674105588' AND  ( events.event_name_id = '061b66b9-ea69-4aae-bd09-551e868ba320' )  AND
// ( (JSON_EXTRACT_STRING(events.properties, '$channel') = 'ChannelName1') ) AND (users.is_group_user=1) AND
// users.group_1_id IS NOT NULL)
func (store *MemSQL) addSourceFilterForSegments(projectID int64,
	source string, caller string) (string, string) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var addSourceStmt string
	addColString := " " + "users.updated_at, users.is_group_user,"
	var selectVal string
	if C.EnableOptimisedFilterOnEventUserQuery() {
		selectVal = "_event_users_view"
	} else {
		selectVal = "users"
	}
	if caller == model.USER_PROFILE_CALLER {
		addSourceStmt = " " + fmt.Sprintf("(%s.is_group_user=0 OR %s.is_group_user IS NULL)", selectVal, selectVal)
		if model.UserSourceMap[source] == model.UserSourceWeb {
			addSourceStmt = addSourceStmt + " " + fmt.Sprintf("AND (%s.source="+strconv.Itoa(model.UserSourceMap[source])+" OR %s.source IS NULL)", selectVal, selectVal)
		} else if source == "All" {
			addSourceStmt = addSourceStmt + ""
		} else {
			addSourceStmt = addSourceStmt + " " + fmt.Sprintf("AND %s.source=", selectVal) + strconv.Itoa(model.UserSourceMap[source])
		}
		addColString = addColString + " " + "users.source"
	} else if caller == model.ACCOUNT_PROFILE_CALLER {
		groups, errCode := store.GetGroups(projectID)
		if errCode != http.StatusFound {
			log.WithField("status", errCode).Error("Failed to get groups while adding group info.")
		}
		groupNameIDMap := make(map[string]int)
		if len(groups) > 0 {
			for _, group := range groups {
				if group.Name == model.GROUP_NAME_HUBSPOT_COMPANY || group.Name == model.GROUP_NAME_SALESFORCE_ACCOUNT {
					groupNameIDMap[group.Name] = group.ID
				}
			}
		}
		hubspotID, hubspotExists := groupNameIDMap[model.GROUP_NAME_HUBSPOT_COMPANY]
		salesforceID, salesforceExists := groupNameIDMap[model.GROUP_NAME_SALESFORCE_ACCOUNT]

		if !hubspotExists && !salesforceExists {
			log.WithFields(logFields).Error("No CRMs Enabled for this project.")
		}
		if source == model.GROUP_NAME_HUBSPOT_COMPANY && !hubspotExists {
			log.WithFields(logFields).Error("Hubspot Not Enabled for this project.")
		}
		if source == model.GROUP_NAME_SALESFORCE_ACCOUNT && !salesforceExists {
			log.WithFields(logFields).Error("Salesforce Not Enabled for this project.")
		}
		addSourceStmt = " " + fmt.Sprintf("(%s.is_group_user=1)", selectVal)
		if source == "All" && hubspotExists && salesforceExists {
			addSourceStmt = addSourceStmt + " " + fmt.Sprintf("AND (%s.group_%d_id IS NOT NULL OR %s.group_%d_id IS NOT NULL)", selectVal, hubspotID, selectVal, salesforceID)
			addColString = addColString + " " + fmt.Sprintf("users.group_%d_id, users.group_%d_id", hubspotID, salesforceID)
		} else if (source == "All" || source == model.GROUP_NAME_HUBSPOT_COMPANY) && hubspotExists {
			addSourceStmt = addSourceStmt + " " + fmt.Sprintf("AND %s.group_%d_id IS NOT NULL", selectVal, hubspotID)
			addColString = addColString + " " + fmt.Sprintf("users.group_%d_id", hubspotID)
		} else if (source == "All" || source == model.GROUP_NAME_SALESFORCE_ACCOUNT) && salesforceExists {
			addSourceStmt = addSourceStmt + " " + fmt.Sprintf("AND %s.group_%d_id IS NOT NULL", selectVal, salesforceID)
			addColString = addColString + " " + fmt.Sprintf("users.group_%d_id", salesforceID)
		}
	}
	return addSourceStmt, addColString
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
func addUniqueUsersAggregationQuery(projectID int64, query *model.Query, qStmnt *string,
	qParams *[]interface{}, refStep string) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"q_stmnt":    qStmnt,
		"q_params":   qParams,
		"ref_step":   refStep,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	eventLevelGroupBys, otherGroupBys := separateEventLevelGroupBys(query.GroupByProperties)
	var egKeys string
	var unionStepName string

	var aggregatePropertyDetails model.QueryGroupByProperty
	aggregateKey := model.AliasAggr
	aggregatePropertyDetails.Property = query.AggregateProperty
	aggregatePropertyDetails.Entity = query.AggregateEntity
	var aggregateSelect string
	var aggregateParams []interface{}

	_, _, egKeys = buildGroupKeys(projectID, eventLevelGroupBys, query.Timezone)
	if query.EventsCondition == model.EventCondAllGivenEvent {
		unionStepName = "all_users_intersect"
	} else if query.EventsCondition == model.EventCondEachGivenEvent {
		unionStepName = "each_users_union"
	} else {
		unionStepName = "any_users_union"
	}

	// select
	userGroupProps := filterGroupPropsByType(otherGroupBys, model.PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(projectID, userGroupProps, query.Timezone)
	*qParams = append(*qParams, ugSelectParams...)
	// order of group keys changes here if users and event
	// group by used together, but translated correctly.
	termSelect := ""
	if query.EventsCondition == model.EventCondEachGivenEvent {
		termSelect = fmt.Sprintf(" %s.event_name, ", refStep)
	}
	termSelect = termSelect + joinWithComma(ugSelect, egKeys)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	termSelect = appendSelectTimestampColIfRequired(termSelect, isGroupByTimestamp)
	termStmnt := ""

	if query.AggregateProperty != "" && query.AggregateProperty != "1" {
		if query.AggregatePropertyType == U.PropertyTypeNumerical {
			noneSelectCase := fmt.Sprintf("CASE WHEN %s.%s = '%s' THEN 0.0 ", refStep, aggregatePropertyDetails.Property, model.PropertyValueNone)
			emptySelectCase := fmt.Sprintf("WHEN %s.%s = '' THEN 0.0 ", refStep, aggregatePropertyDetails.Property)
			defaultCase := fmt.Sprintf("ELSE %s.%s END AS %s ", refStep, aggregatePropertyDetails.Property, model.AliasAggr)
			aggregateSelect = noneSelectCase + emptySelectCase + defaultCase
		} else {
			aggregateSelect, aggregateParams = getNoneHandledGroupBySelect(projectID, aggregatePropertyDetails, aggregateKey, query.Timezone)
		}
		termStmnt = termStmnt + ", " + aggregateSelect
		*qParams = append(*qParams, aggregateParams...)
	}

	if termSelect != "" {
		termStmnt = fmt.Sprintf("SELECT %s.event_user_id, %s.coal_user_id, ", refStep, refStep) + termSelect + " FROM " + refStep
	} else {
		termStmnt = fmt.Sprintf("SELECT %s.event_user_id, %s.coal_user_id ", refStep, refStep) + termSelect + " FROM " + refStep
	}
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStep + ".event_user_id=users.id"
		// Using string format for project_id condition, as the value is from internal system.
		termStmnt = termStmnt + " AND " + fmt.Sprintf("users.project_id = %d", projectID)
	}

	_, _, groupKeys := buildGroupKeys(projectID, query.GroupByProperties, query.Timezone)

	termStmnt = as(unionStepName, termStmnt)
	var aggregateFromStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys string
	if isGroupByTypeWithBuckets(query.GroupByProperties) {
		eventName := ""
		if query.EventsCondition == model.EventCondEachGivenEvent {
			eventName = model.AliasEventName
		}
		isAggregateOnProperty := false
		if query.AggregateProperty != "" && query.AggregateProperty != "1" {
			isAggregateOnProperty = true
		}
		bucketedStepName, bucketedSelectKeys, bucketedGroupBys, bucketedOrderBys := appendNumericalBucketingSteps(isAggregateOnProperty,
			&termStmnt, qParams, query.GroupByProperties, unionStepName, eventName, isGroupByTimestamp, "event_user_id, coal_user_id")
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

	aggregateSelect = "SELECT "
	if isGroupByTimestamp {
		aggregateSelect = aggregateSelect + model.AliasDateTime + ", "
		aggregateGroupBys = joinWithComma(aggregateGroupBys, model.AliasDateTime)
	}
	// adding select event_name and group by event_name for each event-user count
	if query.EventsCondition == model.EventCondEachGivenEvent {
		aggregateSelect = aggregateSelect + model.AliasEventName + ", "
		aggregateGroupBys = joinWithComma(model.AliasEventName, aggregateGroupBys)
	}
	if query.AggregateProperty != "" && query.AggregateProperty != "1" {
		aggregateSelect = aggregateSelect + aggregateSelectKeys + fmt.Sprintf("%s(%s) as %s FROM %s",
			query.AggregateFunction, model.AliasAggr, model.AliasAggr, aggregateFromStepName)
	} else {
		aggregateSelect = aggregateSelect + aggregateSelectKeys + query.GetAggregateFunction() + fmt.Sprintf("(DISTINCT(coal_user_id)) AS %s FROM %s",
			model.AliasAggr, aggregateFromStepName)
	}

	aggregateSelect = appendGroupBy(aggregateSelect, aggregateGroupBys)
	if aggregateOrderBys != "" {
		aggregateSelect = aggregateSelect + " ORDER BY " + aggregateOrderBys
	} else {
		aggregateSelect = appendOrderByAggr(aggregateSelect)
	}

	if query.Caller == model.USER_PROFILE_CALLER {
		aggregateSelect = fmt.Sprintf("SELECT DISTINCT(coal_user_id) as identity, is_anonymous, last_activity, properties FROM %s", aggregateFromStepName)
	} else if query.Caller == model.ACCOUNT_PROFILE_CALLER {
		aggregateSelect = fmt.Sprintf("SELECT DISTINCT(identity) as identity, last_activity, properties FROM %s", aggregateFromStepName)
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

func buildEventsOccurrenceSingleEventQuery(projectId int64,
	q model.Query, enableFilterOpt bool) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"q":          q,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(q.EventsWithProperties) != 1 {
		return "", nil, errors.New("invalid no.of events for single event query")
	}

	if hasGroupEntity(q.GroupByProperties, model.PropertyEntityUser) || isGroupByTypeWithBuckets(q.GroupByProperties) {
		// Using any given event query, which handles groups already.
		return buildEventsOccurrenceWithGivenEventQuery(projectId, q, enableFilterOpt)
	}

	var qStmnt string
	var qParams []interface{}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, model.PropertyEntityEvent)
	egSelect, egSelectParams, egKeys := buildGroupKeys(projectId, eventGroupProps, q.Timezone)
	isGroupByTimestamp := q.GetGroupByTimestamp() != ""

	var qSelect string
	qSelect = appendSelectTimestampIfRequired(qSelect, q.GetGroupByTimestamp(), q.Timezone)
	qSelect = joinWithComma(qSelect, egSelect, fmt.Sprintf("COUNT(*) AS %s", model.AliasAggr))

	addFilterFunc := addFilterEventsWithPropsQuery
	if enableFilterOpt {
		addFilterFunc = addFilterEventsWithPropsQueryV2
	}
	addFilterFunc(projectId, &qStmnt, &qParams, q.EventsWithProperties[0], q.From, q.To,
		"", "", qSelect, egSelectParams, "", "", "", q.GlobalUserProperties)

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
func (store *MemSQL) buildUniqueUsersWithEachGivenEventsQuery(projectID int64,
	query model.Query, enableFilterOpt bool) (string, []interface{}, error) {

	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0)

	steps, stepsToKeysMap := store.addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams, enableFilterOpt)
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
		if query.Caller == model.USER_PROFILE_CALLER {
			selectStr = fmt.Sprintf("%s.coal_user_id as coal_user_id, %s.is_anonymous, %s.last_activity, %s.properties", step, step, step, step)
		} else if query.Caller == model.ACCOUNT_PROFILE_CALLER {
			selectStr = fmt.Sprintf("%s.identity, %s.last_activity, %s.properties", step, step, step)
		}
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
	addUniqueUsersAggregationQuery(projectID, &query, &qStmnt, &qParams, stepUsersUnion)
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
	logFields := log.Fields{
		"step":       step,
		"steps":      steps,
		"keys_map":   keysMap,
		"total_keys": totalKeys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

events_intersect AS (SELECT step_0.event_user_id as event_user_id, step_0.coal_user_id as coal_user_id, step_1._group_key_1, step_1._group_key_2,
step_0._group_key_3 FROM step_0  JOIN step_1 ON step_1.coal_user_id = step_0.coal_user_id) ,

all_users_intersect AS (SELECT events_intersect.event_user_id, events_intersect.coal_user_id, CASE WHEN user_properties.properties->>'' IS NULL THEN
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
_group_key_2, _group_key_3,  COUNT(DISTINCT(coal_user_id)) AS count FROM bucketed GROUP BY _group_key_0_bucket, _group_key_1_bucket,
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
func (store *MemSQL) buildUniqueUsersWithAllGivenEventsQuery(projectID int64,
	query model.Query, enableFilterOpt bool) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0)

	steps, _ := store.addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams, enableFilterOpt)

	// users intersection
	var intersectSelect string
	if query.Caller == model.USER_PROFILE_CALLER {
		intersectSelect = fmt.Sprintf("%s.coal_user_id as coal_user_id, %s.is_anonymous, %s.last_activity, %s.properties", steps[0], steps[0], steps[0], steps[0])
	} else if query.Caller == model.ACCOUNT_PROFILE_CALLER {
		intersectSelect = fmt.Sprintf("%s.identity, %s.last_activity, %s.properties", steps[0], steps[0], steps[0])
	} else {
		intersectSelect = fmt.Sprintf("%s.event_user_id as event_user_id, %s.coal_user_id as coal_user_id", steps[0], steps[0])
	}
	if query.GetGroupByTimestamp() != "" {
		intersectSelect = joinWithComma(intersectSelect,
			fmt.Sprintf("%s.%s as %s", steps[0], model.AliasDateTime, model.AliasDateTime))
	}

	// adds group by event property with user selected event (step).
	eventGroupKeysWithStep := buildEventGroupKeysWithStep(query.GroupByProperties,
		query.EventsWithProperties)
	intersectSelect = joinWithComma(intersectSelect, eventGroupKeysWithStep)

	var intersectJoin string
	for i := range steps {
		if i > 0 {
			if query.Caller == model.ACCOUNT_PROFILE_CALLER {
				intersectJoin = intersectJoin + " " + fmt.Sprintf("JOIN %s ON %s.identity = %s.identity",
					steps[i], steps[i], steps[i-1])
			} else {
				intersectJoin = intersectJoin + " " + fmt.Sprintf("JOIN %s ON %s.coal_user_id = %s.coal_user_id",
					steps[i], steps[i], steps[i-1])
			}

			// include date also intersection condition on
			// group by timestamp.
			if query.GetGroupByTimestamp() != "" {
				intersectJoin = intersectJoin + " " + fmt.Sprintf("AND %s.%s = %s.%s",
					steps[i], model.AliasDateTime, steps[i-1], model.AliasDateTime)
			}
		}
	}
	stepEventsIntersect := "events_intersect"
	qStmnt = joinWithComma(qStmnt, as(stepEventsIntersect,
		fmt.Sprintf("SELECT %s FROM %s %s", intersectSelect, steps[0], intersectJoin)))

	addUniqueUsersAggregationQuery(projectID, &query, &qStmnt, &qParams, stepEventsIntersect)
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
func (store *MemSQL) buildUniqueUsersWithAnyGivenEventsQuery(projectID int64,
	query model.Query, enableFilterOpt bool) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0)

	steps, stepsToKeysMap := store.addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams, enableFilterOpt)
	totalGroupKeys := 0
	for _, val := range stepsToKeysMap {
		totalGroupKeys = totalGroupKeys + len(val)
	}

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	var unionStmnt string
	for i, step := range steps {
		selectStr := fmt.Sprintf("%s.event_user_id as event_user_id, %s.coal_user_id as coal_user_id", step, step)
		selectStr = appendSelectTimestampColIfRequired(selectStr, isGroupByTimestamp)

		if query.Caller == model.USER_PROFILE_CALLER {
			selectStr = fmt.Sprintf("%s.coal_user_id as coal_user_id, %s.is_anonymous, %s.last_activity, %s.properties", step, step, step, step)
		} else if query.Caller == model.ACCOUNT_PROFILE_CALLER {
			selectStr = fmt.Sprintf("%s.identity, %s.last_activity, %s.properties", step, step, step)
		}
		egKeysForStep := getKeysForStep(step, steps, stepsToKeysMap, totalGroupKeys)
		selectStr = joinWithComma(selectStr, egKeysForStep)

		selectStmnt := fmt.Sprintf("SELECT %s FROM %s", selectStr, step)
		if i == 0 {
			unionStmnt = selectStmnt
		} else {
			unionStmnt = unionStmnt + " UNION ALL " + selectStmnt
		}
	}

	stepUsersUnion := "events_union"
	qStmnt = joinWithComma(qStmnt, as(stepUsersUnion, unionStmnt))

	addUniqueUsersAggregationQuery(projectID, &query, &qStmnt, &qParams, stepUsersUnion)
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
func (store *MemSQL) buildUniqueUsersSingleEventQuery(projectID int64,
	query model.Query, enableFilterOpt bool) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0)

	steps, _ := store.addEventFilterStepsForUniqueUsersQuery(projectID, &query, &qStmnt, &qParams, enableFilterOpt)
	addUniqueUsersAggregationQuery(projectID, &query, &qStmnt, &qParams, steps[0])
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
func buildEventsOccurrenceWithGivenEventQuery(projectID int64, q model.Query,
	enableFilterOpt bool) (string, []interface{}, error) {

	logFields := log.Fields{
		"project_id": projectID,
		"q":          q,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0)

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, model.PropertyEntityEvent)
	egSelect, egParams, egKeys := buildGroupKeys(projectID, eventGroupProps, q.Timezone)
	isGroupByTimestamp := q.GetGroupByTimestamp() != ""

	filterSelect := joinWithComma(model.SelectDefaultEventFilter, egSelect)
	filterSelect = appendSelectTimestampIfRequired(filterSelect, q.GetGroupByTimestamp(), q.Timezone)

	refStepName := ""
	filters := make([]string, 0)
	for i, ewp := range q.EventsWithProperties {
		eventNameSelect := "'" + strconv.Itoa(i) + "_" + ewp.Name + "'" + " AS event_name "
		filterSelect := joinWithComma(filterSelect, eventNameSelect)
		refStepName = fmt.Sprintf("step%d", i)
		filters = append(filters, refStepName)

		addFilterFunc := addFilterEventsWithPropsQuery
		if enableFilterOpt {
			addFilterFunc = addFilterEventsWithPropsQueryV2
		}
		addFilterFunc(projectID, &qStmnt, &qParams, ewp, q.From, q.To, "",
			refStepName, filterSelect, egParams, "", "", "", q.GlobalUserProperties)
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

			qSelect := appendSelectTimestampColIfRequired(model.SelectDefaultEventFilterByAlias, isGroupByTimestamp)
			qSelect = joinWithComma(qSelect, egKeys)
			unionStmnt = unionStmnt + " SELECT " + qSelect + " FROM " + filter
		}
		unionStmnt = as(unionStepName, unionStmnt)
		qStmnt = appendStatement(qStmnt, unionStmnt)

		refStepName = unionStepName
	}

	// count.
	userGroupProps := filterGroupPropsByType(q.GroupByProperties, model.PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(projectID, userGroupProps, q.Timezone)
	_, _, groupKeys := buildGroupKeys(projectID, q.GroupByProperties, q.Timezone)
	qParams = append(qParams, ugSelectParams...)

	eventNameSelect := "event_name"
	groupKeys = joinWithComma(eventNameSelect, groupKeys)

	// select
	tSelect := joinWithComma(eventNameSelect, ugSelect, egKeys, "event_user_id")
	tSelect = appendSelectTimestampColIfRequired(tSelect, isGroupByTimestamp)

	termStmnt := "SELECT " + tSelect + " FROM " + refStepName
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id"
		// Using string format for project_id condition, as the value is from internal system.
		termStmnt = termStmnt + " AND " + fmt.Sprintf("users.project_id = %d", projectID)
	}

	withUsersStepName := "users_any_event"
	qStmnt = joinWithComma(qStmnt, as(withUsersStepName, termStmnt))

	aggregateSelect := "SELECT "
	if isGroupByTimestamp {
		aggregateSelect = aggregateSelect + model.AliasDateTime + ", "
	}
	if isGroupByTypeWithBuckets(q.GroupByProperties) {
		isAggregateOnProperty := false
		if q.AggregateProperty != "" && q.AggregateProperty != "1" {
			isAggregateOnProperty = true
		}
		bucketedStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys := appendNumericalBucketingSteps(isAggregateOnProperty,
			&qStmnt, &qParams, q.GroupByProperties, withUsersStepName, eventNameSelect, isGroupByTimestamp, "event_user_id")
		aggregateGroupBys = append(aggregateGroupBys, eventNameSelect)
		aggregateSelectKeys = eventNameSelect + ", " + aggregateSelectKeys
		aggregateSelect = aggregateSelect + aggregateSelectKeys
		aggregateSelect = appendStatement(aggregateSelect, fmt.Sprintf("COUNT(*) AS %s FROM %s", model.AliasAggr, bucketedStepName))
		aggregateSelect = appendGroupByTimestampIfRequired(aggregateSelect, isGroupByTimestamp, strings.Join(aggregateGroupBys, ", "))
		aggregateSelect = aggregateSelect + fmt.Sprintf(" ORDER BY %s, %s", eventNameSelect, strings.Join(aggregateOrderBys, ", "))
	} else {
		aggregateSelect = aggregateSelect + groupKeys
		aggregateSelect = joinWithComma(aggregateSelect, fmt.Sprintf("COUNT(*) AS %s FROM %s", model.AliasAggr, withUsersStepName))
		aggregateSelect = appendGroupByTimestampIfRequired(aggregateSelect, isGroupByTimestamp, groupKeys)
		aggregateSelect = aggregateSelect + fmt.Sprintf(" ORDER BY %s", eventNameSelect)
	}

	aggregateSelect = aggregateSelect + fmt.Sprintf(", %s DESC", model.AliasAggr)
	aggregateSelect = appendLimitByCondition(aggregateSelect, q.GroupByProperties, isGroupByTimestamp)

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
func (store *MemSQL) buildEventCountForEachGivenEventsQueryNEW(projectID int64, query model.Query, enableFilterOpt bool) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	hash1, _ := query.GetQueryCacheHashString()
	steps, stepsToKeysMap, err := store.addEventFilterStepsForEventCountQuery(projectID, &query, &qStmnt, &qParams, enableFilterOpt)
	hash2, _ := query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, query, "Query change - 2")

	if err != nil {
		return qStmnt, qParams, err
	}
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
		if query.AggregateProperty != "" && query.AggregateProperty != "1" {
			if query.AggregatePropertyType == U.PropertyTypeNumerical {
				noneSelectCase := fmt.Sprintf("CASE WHEN %s.%s = '%s' THEN 0.0 ", step, model.AliasAggr, model.PropertyValueNone)
				emptySelectCase := fmt.Sprintf("WHEN %s.%s = '' THEN 0.0 ", step, model.AliasAggr)
				defaultCase := fmt.Sprintf("ELSE %s.%s END as %s ", step, model.AliasAggr, model.AliasAggr)
				selectStr = selectStr + ", " + noneSelectCase + emptySelectCase + defaultCase
			} else {
				selectStr = selectStr + ", " + fmt.Sprintf("%s.%s as %s", step, model.AliasAggr, model.AliasAggr)
			}
		}
		selectStmnt := fmt.Sprintf("SELECT %s FROM %s", selectStr, step)
		if i == 0 {
			unionStmnt = selectStmnt
		} else {
			unionStmnt = unionStmnt + " UNION ALL " + selectStmnt
		}
	}

	qStmnt = joinWithComma(qStmnt, as(stepUsersUnion, unionStmnt))
	hash1, _ = query.GetQueryCacheHashString()
	addEventCountAggregationQuery(projectID, &query, &qStmnt, &qParams, stepUsersUnion)
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, query, "Query change - 3")

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
func (store *MemSQL) addEventFilterStepsForEventCountQuery(projectID int64, q *model.Query, qStmnt *string,
	qParams *[]interface{}, enableFilterOpt bool) ([]string, map[string][]string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"q":          q,
		"q_stmnt":    qStmnt,
		"q_params":   qParams,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var aggregatePropertyDetails model.QueryGroupByProperty
	aggregateKey := model.AliasAggr
	aggregatePropertyDetails.Property = q.AggregateProperty
	aggregatePropertyDetails.Entity = q.AggregateEntity
	var aggregateSelect string
	var aggregateParams []interface{}
	var commonParams []interface{}

	var commonSelectArr []string
	var commonOrderBy string
	stepsToKeysMap := make(map[string][]string)

	for i := range q.EventsWithProperties {
		_, status := store.IsGroupEventNameByQueryEventWithProperties(projectID, q.EventsWithProperties[i])
		if status == http.StatusFound && (len(q.GlobalUserProperties) > 0 ||
			model.IsQueryGroupByLatestUserProperty(q.GroupByProperties)) {
			commonSelect := model.SelectDefaultGroupEventFilter
			commonSelect = appendSelectTimestampIfRequired(commonSelect, q.GetGroupByTimestamp(), q.Timezone)
			commonSelectArr = append(commonSelectArr, commonSelect)
			continue
		}

		commonSelect := model.SelectDefaultEventFilter
		commonSelect = appendSelectTimestampIfRequired(commonSelect, q.GetGroupByTimestamp(), q.Timezone)
		commonSelectArr = append(commonSelectArr, commonSelect)
	}

	if q.AggregateProperty != "" && q.AggregateProperty != "1" {
		for i := range commonSelectArr {
			aggregateSelect, aggregateParams = getNoneHandledGroupBySelect(projectID, aggregatePropertyDetails, aggregateKey, q.Timezone)
			commonSelectArr[i] = commonSelectArr[i] + ", " + aggregateSelect
		}
		commonParams = aggregateParams
	}

	if len(q.GroupByProperties) > 0 && commonOrderBy == "" {
		commonOrderBy = "event_id%s, events.timestamp ASC"
	}

	steps := make([]string, 0, 0)
	for i, ewp := range q.EventsWithProperties {
		refStepName := stepNameByIndex(i)
		steps = append(steps, refStepName)

		var stepSelect, stepOrderBy string
		var stepParams []interface{}
		var stepGroupSelect, stepGroupKeys string
		var stepGroupParams []interface{}

		hash1, _ := q.GetQueryCacheHashString()
		stepGroupSelect, stepGroupParams, stepGroupKeys, _ = buildGroupKeyForStep(projectID,
			&q.EventsWithProperties[i], q.GroupByProperties, i+1, q.Timezone)
		hash2, _ := q.GetQueryCacheHashString()
		logDifferenceIfAny(hash1, hash2, *q, "Query change - 2 - 1")

		eventSelect := commonSelectArr[i]
		stepParams = append(stepParams, commonParams...)
		eventParam := ""
		eventNameSelect := fmt.Sprintf("? AS event_name ")
		eventParam = fmt.Sprintf("%s_%s", strconv.Itoa(i), ewp.Name)
		eventSelect = joinWithComma(eventSelect, eventNameSelect)
		if stepGroupSelect != "" {
			stepSelect = eventSelect + ", " + stepGroupSelect
			stepOrderBy = fmt.Sprintf(commonOrderBy, ", "+stepGroupKeys)
			stepParams = append(stepParams, eventParam)
			stepParams = append(stepParams, stepGroupParams...)
			stepsToKeysMap[refStepName] = strings.Split(stepGroupKeys, ",")
		} else {
			stepSelect = eventSelect
			stepParams = append(stepParams, eventParam)
			if commonOrderBy != "" {
				stepOrderBy = fmt.Sprintf(commonOrderBy, "")
			}
		}

		addJoinStmnt := ""
		if C.SkipUserJoinInEventQueryByProjectID(projectID) {

			groupName, status := store.IsGroupEventNameByQueryEventWithProperties(projectID, q.EventsWithProperties[i])
			if status == http.StatusFound && (model.IsQueryGroupByLatestUserProperty(q.GroupByProperties) || len(q.GlobalUserProperties) > 0) {
				group, status := store.GetGroup(projectID, groupName)
				if status != http.StatusFound {
					if status != http.StatusNotFound {
						log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName}).Error("Failed to get group details. Using users join.")
					} else {
						log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName}).Error("Group not found. Using users join.")
					}
					addJoinStmnt = "JOIN users ON events.user_id=users.id AND users.project_id = ?"
				} else {
					addJoinStmnt = fmt.Sprintf("JOIN users ON events.user_id=users.group_%d_user_id AND users.project_id = ?", group.ID)
				}
				stepParams = append(stepParams, projectID)
			} else if len(q.GlobalUserProperties) > 0 {
				addJoinStmnt = "JOIN users ON events.user_id=users.id AND users.project_id = ?"
				stepParams = append(stepParams, projectID)
			}
		} else {
			addJoinStmnt = "JOIN users ON events.user_id=users.id AND users.project_id = ?"
			stepParams = append(stepParams, projectID)
		}

		addFilterFunc := addFilterEventsWithPropsQuery
		if enableFilterOpt {
			addFilterFunc = addFilterEventsWithPropsQueryV2
		}
		err := addFilterFunc(projectID, qStmnt, qParams, ewp, q.From, q.To,
			"", refStepName, stepSelect, stepParams, addJoinStmnt, "", stepOrderBy, q.GlobalUserProperties)
		if err != nil {
			return steps, stepsToKeysMap, err
		}

		if i < len(q.EventsWithProperties)-1 {
			*qStmnt = *qStmnt + ","
		}
	}

	return steps, stepsToKeysMap, nil
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
func addEventCountAggregationQuery(projectID int64, query *model.Query, qStmnt *string,
	qParams *[]interface{}, refStep string) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"q_stmnt":    qStmnt,
		"q_params":   qParams,
		"ref_step":   refStep,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	hash1, _ := query.GetQueryCacheHashString()
	eventLevelGroupBys, otherGroupBys := separateEventLevelGroupBys(query.GroupByProperties)
	var egKeys string
	var unionStepName string
	var termStmnt string
	hash2, _ := query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, *query, "Query change - 3-1")

	_, _, egKeys = buildGroupKeys(projectID, eventLevelGroupBys, query.Timezone)
	unionStepName = "each_users_union"

	// select
	userGroupProps := filterGroupPropsByType(otherGroupBys, model.PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(projectID, userGroupProps, query.Timezone)
	*qParams = append(*qParams, ugSelectParams...)
	termSelect := ""
	termSelect = fmt.Sprintf(" %s.event_name, ", refStep)
	termSelect = termSelect + joinWithComma(ugSelect, egKeys)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	termSelect = appendSelectTimestampColIfRequired(termSelect, isGroupByTimestamp)

	if query.AggregateProperty != "" && query.AggregateProperty != "1" {
		aggregateColumnSelect := fmt.Sprintf("%s.%s as %s", refStep, model.AliasAggr, model.AliasAggr)
		termStmnt = fmt.Sprintf("SELECT %s.event_user_id, %s.event_id, ", refStep, refStep) + aggregateColumnSelect +
			", " + termSelect + " FROM " + refStep
	} else {
		termStmnt = fmt.Sprintf("SELECT %s.event_user_id, %s.event_id, ", refStep, refStep) + termSelect +
			" FROM " + refStep
	}
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStep + ".event_user_id=users.id"
		termStmnt = termStmnt + " "

		// Using string format for project_id condition, as the value is from internal system.
		termStmnt = termStmnt + " AND " + fmt.Sprintf("users.project_id = %d", projectID)
	}

	hash1, _ = query.GetQueryCacheHashString()
	_, _, groupKeys := buildGroupKeys(projectID, query.GroupByProperties, query.Timezone)
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, *query, "Query change - 3-2")

	hash1, _ = query.GetQueryCacheHashString()
	termStmnt = as(unionStepName, termStmnt)
	var aggregateFromStepName, aggregateSelectKeys, aggregateGroupBys, aggregateOrderBys string
	if isGroupByTypeWithBuckets(query.GroupByProperties) {
		eventName := model.AliasEventName
		isAggregateOnProperty := false
		if query.AggregateProperty != "" && query.AggregateProperty != "1" {
			isAggregateOnProperty = true
		}
		bucketedStepName, bucketedSelectKeys, bucketedGroupBys, bucketedOrderBys := appendNumericalBucketingSteps(isAggregateOnProperty,
			&termStmnt, qParams, query.GroupByProperties, unionStepName, eventName, isGroupByTimestamp, "event_id")
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
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, *query, "Query change - 3-3")


	aggregateSelect := "SELECT "
	if isGroupByTimestamp {
		aggregateSelect = aggregateSelect + model.AliasDateTime + ", "
		aggregateGroupBys = joinWithComma(aggregateGroupBys, model.AliasDateTime)
	}
	// adding select event_name and group by event_name for each event-user count
	if query.EventsCondition == model.EventCondEachGivenEvent {
		aggregateSelect = aggregateSelect + model.AliasEventName + ", "
		aggregateGroupBys = joinWithComma(model.AliasEventName, aggregateGroupBys)
	}
	if query.AggregateProperty != "" && query.AggregateProperty != "1" {
		aggregateSelect = aggregateSelect + aggregateSelectKeys + fmt.Sprintf("%s(CASE WHEN %s IS NULL THEN 0 ELSE %s END) as %s FROM %s",
			query.AggregateFunction, model.AliasAggr, model.AliasAggr, model.AliasAggr, aggregateFromStepName)
	} else {
		aggregateSelect = aggregateSelect + aggregateSelectKeys + fmt.Sprintf("COUNT(event_id) AS %s FROM %s",
			model.AliasAggr, aggregateFromStepName)
	}

	aggregateSelect = appendGroupBy(aggregateSelect, aggregateGroupBys)
	if aggregateOrderBys != "" {
		aggregateSelect = aggregateSelect + " ORDER BY " + aggregateOrderBys
	} else {
		aggregateSelect = appendOrderByAggr(aggregateSelect)
	}
	hash1, _ = query.GetQueryCacheHashString()
	aggregateSelect = appendLimitByCondition(aggregateSelect, query.GroupByProperties, isGroupByTimestamp)
	hash2, _ = query.GetQueryCacheHashString()
	logDifferenceIfAny(hash1, hash2, *query, "Query change - 3-4")

	*qStmnt = appendStatement(*qStmnt, aggregateSelect)
}

// builds group keys for event properties for given step (event_with_properties).
func buildGroupKeyForStep(projectID int64, eventWithProperties *model.QueryEventWithProperties,
	groupProps []model.QueryGroupByProperty, ewpIndex int, timezoneString string) (string, []interface{}, string, bool) {
	logFields := log.Fields{
		"project_id":            projectID,
		"event_with_properties": eventWithProperties,
		"group_props":           groupProps,
		"ewp_index":             ewpIndex,
		"timezone_string":       timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	groupPropsByStep := make([]model.QueryGroupByProperty, 0, 0)
	groupByUserProperties := false
	for i := range groupProps {
		if groupProps[i].EventNameIndex == ewpIndex &&
			groupProps[i].EventName == eventWithProperties.Name {
			groupPropsByStep = append(groupPropsByStep, groupProps[i])
			if groupProps[i].Entity == model.PropertyEntityUser {
				groupByUserProperties = true
			}
		}

	}
	groupSelect, groupSelectParams, groupKeys := buildGroupKeys(projectID, groupPropsByStep, timezoneString)
	return groupSelect, groupSelectParams, groupKeys, groupByUserProperties
}
