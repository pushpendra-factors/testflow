package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	MetaStepTimeInfo = "MetaStepTimeInfo"
)

func (store *MemSQL) RunFunnelQuery(projectId uint64, query model.Query) (*model.QueryResult, int, string) {
	logFields := log.Fields{
		"project_id": projectId,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	if !isValidFunnelQuery(&query) {
		return nil, http.StatusBadRequest, model.ErrMsgMaxFunnelStepsExceeded
	}

	if C.SkipEventNameStepByProjectID(projectId) {
		store.fillEventNameIDs(projectId, &query)
	}

	groupIds := make([]int, 0)
	for i := range query.EventsWithProperties {
		if C.IsEventsFunnelsGroupSupportEnabled(projectId) && U.IsGroupEventName(query.EventsWithProperties[i].Name) {
			groupName := U.GetGroupNameFromGroupEventName(query.EventsWithProperties[i].Name)
			group, status := store.GetGroup(projectId, groupName)
			if status != http.StatusFound {
				return nil, http.StatusBadRequest, "group with the given groupName not found in the project"
			}
			groupIds = append(groupIds, group.ID)
		} else {
			groupIds = append(groupIds, 0)
		}
	}

	stmnt, params, err := BuildFunnelQuery(projectId, query, groupIds)
	if err != nil {
		log.WithError(err).Error(model.ErrMsgQueryProcessingFailure)
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}

	logCtx := log.WithFields(logFields)
	if stmnt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	result, err := store.ExecQuery(stmnt, params)
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog1")
	}

	if len(query.GroupByProperties) > 0 {
		// $no_group comes as the last record for MemSQL query. Put back as first.
		noGroupRow := result.Rows[len(result.Rows)-1]
		result.Rows = append([][]interface{}{noGroupRow}, result.Rows[0:len(result.Rows)-1]...)
	}
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog2")
	}

	// should be done before translation of group keys
	translateNullToZeroOnFunnelResult(result)
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog3")
	}

	if len(query.EventsWithProperties) > 1 {
		err = addStepTimeToMeta(result, logCtx)
		if err != nil {
			logCtx.WithError(err).Error("Failed adding funnel step time to meta.")
			return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
		}
	}
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog4")
	}

	err = addStepConversionPercentageToFunnel(result)
	if err != nil {
		logCtx.WithError(err).Error("Failed adding funnel step conversion percentage.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog5")
	}

	err = translateGroupKeysIntoColumnNames(result, query.GroupByProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed translating group keys on result.")
		return nil, http.StatusInternalServerError, model.ErrMsgQueryProcessingFailure
	}
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog6")
	}

	sanitizeNumericalBucketRanges(result, &query)
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog7")
	}

	if model.HasGroupByDateTypeProperties(query.GroupByProperties) {
		model.SanitizeDateTypeRows(result, &query)
	}
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog8")
	}

	addQueryToResultMeta(result, query)
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog9")
	}

	updatedMetaStepTimeInfoHeaders(result)
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog10")
	}
	model.SanitizeStringSumToNumeric(result)
	if projectId == 659 {
		logCtx.WithFields(log.Fields{"result": result, "err": err}).Info("SenseHQDebugLog11")
	}
	return result, http.StatusOK, "Successfully executed query"
}

// updatedMetaStepTimeInfoHeaders updates meta rows to match the result rows
func updatedMetaStepTimeInfoHeaders(result *model.QueryResult) {
	logFields := log.Fields{
		"result": result,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// Update the row headers in MetaStepTimeInfo using result original group count
	for idx := range result.Meta.MetaMetrics {
		if result.Meta.MetaMetrics[idx].Title == MetaStepTimeInfo {
			for groupByKeyIdx, column := range result.Meta.MetaMetrics[idx].Headers {
				if strings.HasPrefix(column, model.GroupKeyPrefix) {
					for rowIdx := range result.Meta.MetaMetrics[idx].Rows {
						result.Meta.MetaMetrics[idx].Rows[rowIdx][groupByKeyIdx] = result.Rows[rowIdx][groupByKeyIdx]
					}
				}
			}
			// Only one meta step time info, hence returning
			return
		}
	}
}

// addStepTimeToMeta adds step time in result's meta metrics
func addStepTimeToMeta(result *model.QueryResult, logCtx *log.Entry) error {
	logFields := log.Fields{
		"result":  result,
		"log_ctx": logCtx,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	headers := make([]string, 0, 0)
	var groupKeyIndexes []int
	var stepTimeIndexes []int
	stepTimeStartIndex := -1
	for index, h := range result.Headers {
		if strings.HasPrefix(h, model.GroupKeyPrefix) {
			headers = append(headers, h)
			groupKeyIndexes = append(groupKeyIndexes, index)
		}
		if strings.HasSuffix(h, model.FunnelTimeSuffix) && strings.HasPrefix(h, model.StepPrefix) {
			headers = append(headers, h)
			stepTimeIndexes = append(stepTimeIndexes, index)
			if stepTimeStartIndex == -1 {
				stepTimeStartIndex = index
			}
		}
	}

	//result.Meta.MetaMetrics
	var rows [][]interface{}
	for i := range result.Rows {
		var row []interface{}
		for _, ci := range groupKeyIndexes {
			row = append(row, result.Rows[i][ci])
		}
		for _, ci := range stepTimeIndexes {
			val := 0.0
			key, ok := result.Rows[i][ci].(int)
			if ok {
				val = float64(key)
			} else {
				key, ok := result.Rows[i][ci].(string)
				if ok && key != "" {
					if valFloat, err := strconv.ParseFloat(key, 64); err == nil {
						value, err := U.FloatRoundOffWithPrecision(valFloat, 2)
						if err != nil {
							// add log but don't fail the query
							logCtx.WithError(err).Error("Failed to round off time value")
							value = 0.0
						}
						val = value
					}
				}
			}
			row = append(row, val)
		}
		rows = append(rows, row)
	}

	metaMetricsStepTime := model.HeaderRows{Title: MetaStepTimeInfo, Headers: headers, Rows: rows}
	result.Meta.MetaMetrics = append(result.Meta.MetaMetrics, metaMetricsStepTime)

	// Removing step time count
	result.Headers = result.Headers[:stepTimeStartIndex]
	for i := range result.Rows {
		result.Rows[i] = result.Rows[i][:stepTimeStartIndex]
	}
	return nil
}

func BuildFunnelQuery(projectId uint64, query model.Query, groupIds []int) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"query":      query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	addIndexToGroupByProperties(&query)

	if query.EventsCondition == model.QueryTypeEventsOccurrence {
		return "", nil, errors.New("funnel on events occurrence is not supported")
	}

	return buildUniqueUsersFunnelQuery(projectId, query, groupIds)
}

func translateNullToZeroOnFunnelResult(result *model.QueryResult) {
	logFields := log.Fields{
		"result": result,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var percentageIndexes []int

	var index int
	for _, h := range result.Headers {
		if strings.HasPrefix(h, model.FunnelConversionPrefix) || strings.HasPrefix(h, model.StepPrefix) {
			percentageIndexes = append(percentageIndexes, index)
		}
		index++
	}

	for i := range result.Rows {
		for _, ci := range percentageIndexes {
			if result.Rows[i][ci] == nil {
				result.Rows[i][ci] = 0
			}
		}
	}
}

func addStepConversionPercentageToFunnel(result *model.QueryResult) error {
	logFields := log.Fields{
		"result": result,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(result.Rows) == 0 {
		return errors.New("invalid funnel result")
	}

	stepIndexes := make([]int, 0, 0)
	nonStepIndexes := make([]int, 0, 0)
	for i, header := range result.Headers {
		if strings.HasPrefix(header, model.StepPrefix) {
			stepIndexes = append(stepIndexes, i)
		} else {
			nonStepIndexes = append(nonStepIndexes, i)
		}
	}

	headers := make([]string, 0, 0)

	for _, nsi := range nonStepIndexes {
		headers = append(headers, result.Headers[nsi])
	}

	for _, si := range stepIndexes {
		headers = append(headers, result.Headers[si])
		if si == stepIndexes[0] {
			continue
		}

		headers = append(headers, fmt.Sprintf("%s%s_%s",
			model.FunnelConversionPrefix, result.Headers[si-1], result.Headers[si]))
	}

	headers = append(headers, fmt.Sprintf("%s%s", model.FunnelConversionPrefix, "overall"))
	result.Headers = headers // headers with conversion.

	for ri := range result.Rows {
		row := make([]interface{}, 0, 0)

		for _, ci := range nonStepIndexes {
			row = append(row, result.Rows[ri][ci])
		}

		for _, ci := range stepIndexes {
			row = append(row, result.Rows[ri][ci])

			if ci == stepIndexes[0] {
				continue
			}

			prevCount, err := U.GetAggrAsFloat64(result.Rows[ri][ci-1])
			if err != nil {
				return err
			}

			curCount, err := U.GetAggrAsFloat64(result.Rows[ri][ci])
			if err != nil {
				return err
			}

			row = append(row, getConversionPercentageAsString(prevCount, curCount))
		}

		// add overall conversion.
		firstStepCount, err := U.GetAggrAsFloat64(result.Rows[ri][stepIndexes[0]])
		if err != nil {
			return err
		}

		lastIndex := stepIndexes[len(stepIndexes)-1]
		lastStepCount, err := U.GetAggrAsFloat64(result.Rows[ri][lastIndex])
		if err != nil {
			return err
		}
		row = append(row, getConversionPercentageAsString(firstStepCount, lastStepCount))

		result.Rows[ri] = row // row with conversion.
	}

	return nil
}

func getConversionPercentageAsString(prevCount float64, curCount float64) string {
	logFields := log.Fields{
		"prev_count": prevCount,
		"cur_count":  curCount,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var conversion float64

	if prevCount == 0 {
		conversion = float64(0)
	} else {
		conversion = (curCount / prevCount) * 100
	}

	// Percentage with one decimal point.
	return fmt.Sprintf("%0.1f", conversion)
}

/*
buildUniqueUsersFunnelQuery

/*
WITH
	step_0_names AS (
		SELECT id, project_id, name FROM event_names WHERE project_id='1' AND name='View Project'
	),
	step_0 AS (
		SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id)
		as coal_user_id, events.user_id, events.timestamp, 1 as step_0 FROM events JOIN users ON events.user_id=users.id
		WHERE events.project_id='1' AND timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN
		(SELECT id FROM step_0_names WHERE project_id='1' AND name='View Project') ORDER BY coal_user_id, events.timestamp ASC
		),
	step_1_names AS (
		SELECT id, project_id, name FROM event_names WHERE project_id='1' AND name='Fund Project'
	),
	step_1 AS (
		SELECT COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as step_1
		FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='1' AND timestamp>='1393612200' AND
		timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='1'
		AND name='Fund Project') ORDER BY coal_user_id, events.timestamp ASC
	),
	step_1_step_0_users AS (
		SELECT DISTINCT ON(step_1.coal_user_id) step_1.coal_user_id, step_1.user_id,step_1.timestamp, step_1
		FROM step_0 LEFT JOIN step_1 ON step_0.coal_user_id = step_1.coal_user_id WHERE step_1.timestamp > step_0.timestamp
		ORDER BY step_1.coal_user_id, timestamp ASC
		),
	step_2_names AS (
		SELECT id, project_id, name FROM event_names WHERE project_id='1' AND name='run_query'
	),
	step_2 AS (
		SELECT COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as step_2
		FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='1' AND timestamp>='1393612200' AND
		timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM step_2_names WHERE project_id='1' AND
		name='run_query') ORDER BY coal_user_id, events.timestamp ASC
		),
	step_2_step_1_users AS (
		SELECT DISTINCT ON(step_2.coal_user_id) step_2.coal_user_id, step_2.user_id,step_2.timestamp, step_2 FROM
		step_1_step_0_users LEFT JOIN step_2 ON step_1_step_0_users.coal_user_id = step_2.coal_user_id WHERE
		step_2.timestamp > step_1_step_0_users.timestamp ORDER BY step_2.coal_user_id, timestamp ASC
		),
	funnel AS (
		SELECT step_0, step_1, step_2 FROM step_0 LEFT JOIN users on step_0.user_id=users.id LEFT JOIN
		step_1_step_0_users ON step_0.coal_user_id=step_1_step_0_users.coal_user_id LEFT JOIN step_2_step_1_users
		ON step_1_step_0_users.coal_user_id=step_2_step_1_users.coal_user_id
		)

	SELECT SUM(step_0) AS step_0, SUM(step_1) AS step_1, SUM(step_2) AS step_2 FROM funnel
*/
/*
buildFunnelQuery with session analysis
WITH
	step_0_names AS (
		SELECT id, project_id, name FROM event_names WHERE project_id='1' AND name='View Project'
	),
	step_0 AS (
		SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id)
		as coal_user_id, events.user_id, events.timestamp, 1 as step_0, events.session_id as session_id FROM events JOIN users ON events.user_id=users.id
		WHERE events.project_id='1' AND timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN
		(SELECT id FROM step_0_names WHERE project_id='1' AND name='View Project') ORDER BY coal_user_id, events.timestamp ASC
		),
	step_1_names AS (
		SELECT id, project_id, name FROM event_names WHERE project_id='1' AND name='Fund Project'
	),
	step_1 AS (
		SELECT COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as step_1,
		events.session_id as session_id FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='1' AND
		timestamp>='1393612200' AND timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM step_1_names WHERE
		project_id='1' AND name='Fund Project') ORDER BY coal_user_id, events.timestamp ASC
	),
	step_1_step_0_users AS (
		SELECT DISTINCT ON(step_1.coal_user_id) step_1.coal_user_id, step_1.user_id,step_1.timestamp, step_1, step_1.session_id,
		FROM step_0 LEFT JOIN step_1 ON step_0.coal_user_id = step_1.coal_user_id WHERE step_1.timestamp > step_0.timestamp
		and step_1.session_id = step_0.session_id ORDER BY step_1.coal_user_id, timestamp ASC
		),
	step_2_names AS (
		SELECT id, project_id, name FROM event_names WHERE project_id='1' AND name='run_query'
	),
	step_2 AS (
		SELECT COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as step_2,
		 events.session_id as session_id
		FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='1' AND timestamp>='1393612200' AND
		timestamp<='1396290599' AND events.event_name_id IN (SELECT id FROM step_2_names WHERE project_id='1' AND
		name='run_query') ORDER BY coal_user_id, events.timestamp ASC
		),
	step_2_step_1_users AS (
		SELECT DISTINCT ON(step_2.coal_user_id) step_2.coal_user_id, step_2.user_id,step_2.timestamp, step_2, step_2.session_id
		FROM step_1_step_0_users LEFT JOIN step_2 ON step_1_step_0_users.coal_user_id = step_2.coal_user_id WHERE
		step_2.timestamp > step_1_step_0_users.timestamp AND step_2.session_id = step_1_step_0_users.session_id ORDER BY step_2.coal_user_id, timestamp ASC
		),
	funnel AS (
		SELECT step_0, step_1, step_2 FROM step_0 LEFT JOIN users on step_0.user_id=users.id LEFT JOIN
		step_1_step_0_users ON step_0.coal_user_id=step_1_step_0_users.coal_user_id LEFT JOIN step_2_step_1_users
		ON step_1_step_0_users.coal_user_id=step_2_step_1_users.coal_user_id
		)

	SELECT SUM(step_0) AS step_0, SUM(step_1) AS step_1, SUM(step_2) AS step_2 FROM funnel


*/
func isSessionAnalysisReq(start int64, end int64) bool {
	logFields := log.Fields{
		"start": start,
		"end":   end,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if start != 0 && end != 0 && start < end {
		return true
	}
	return false
}

func buildStepXToYJoin(stepName string, prevStepName string, previousCombinedUsersStepName string,
	isSessionAnalysisReqBool bool, q model.Query, i int) string {
	logFields := log.Fields{
		"step_name":                        stepName,
		"prev_step_name":                   prevStepName,
		"previous_combined_user_step_name": previousCombinedUsersStepName,
		"is_session_analysis_req_bool":     isSessionAnalysisReqBool,
		"q":                                q,
		"i":                                i,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	comparisonSymbol := ">="
	if q.EventsWithProperties[i].Name == q.EventsWithProperties[i-1].Name {
		comparisonSymbol = ">"
	}
	stepXToYJoin := fmt.Sprintf("LEFT JOIN %s ON %s.coal_user_id = %s.coal_user_id WHERE %s.timestamp %s %s.timestamp",
		stepName, previousCombinedUsersStepName, stepName, stepName, comparisonSymbol, previousCombinedUsersStepName)
	if i == 1 {
		stepXToYJoin = fmt.Sprintf("LEFT JOIN %s ON %s.coal_user_id = %s.coal_user_id WHERE %s.timestamp %s %s.timestamp",
			stepName, prevStepName, stepName, stepName, comparisonSymbol, prevStepName)
	}

	if isSessionAnalysisReqBool && i >= int(q.SessionStartEvent) && i < int(q.SessionEndEvent) {
		if i == 1 {
			stepXToYJoin = fmt.Sprintf("%s and %s.session_id = %s.session_id",
				stepXToYJoin, stepName, prevStepName)
		} else {
			stepXToYJoin = fmt.Sprintf("%s and %s.session_id = %s.session_id",
				stepXToYJoin, stepName, previousCombinedUsersStepName)
		}
	}
	return stepXToYJoin
}

func buildStepXToY(stepXToYSelect string, prevStepName string, previousCombinedUsersStepName string,
	stepXToYJoin string, stepName string, i int) string {
	logFields := log.Fields{
		"step_name":                        stepName,
		"prev_step_name":                   prevStepName,
		"previous_combined_user_step_name": previousCombinedUsersStepName,
		"step_x_to_y_join":                 stepXToYJoin,
		"step_x_to_y_select":               stepXToYSelect,
		"i":                                i,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	stepXToY := fmt.Sprintf("SELECT %s FROM %s %s GROUP BY %s.coal_user_id", stepXToYSelect, previousCombinedUsersStepName, stepXToYJoin, stepName)
	if i == 1 {
		stepXToY = fmt.Sprintf("SELECT %s FROM %s %s GROUP BY %s.coal_user_id", stepXToYSelect, prevStepName, stepXToYJoin, stepName)
	}
	return stepXToY
}
func buildAddSelect(stepName string, i int) string {
	logFields := log.Fields{
		"step_name": stepName,
		"i":         i,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	addSelect := fmt.Sprintf("COALESCE(users.customer_user_id,events.user_id) as coal_user_id, FIRST(events.user_id, FROM_UNIXTIME(events.timestamp)) as user_id,"+
		" FIRST(events.timestamp, FROM_UNIXTIME(events.timestamp)) as timestamp, 1 as %s", stepName)

	if i > 0 {
		addSelect = fmt.Sprintf("COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as %s", stepName)
	}
	return addSelect
}

func buildAddSelectForGroup(stepName string, i int) string {
	logFields := log.Fields{
		"step_name": stepName,
		"i":         i,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	addSelect := fmt.Sprintf("COALESCE(users.customer_user_id,users.id) as coal_user_id, FIRST(users.id, FROM_UNIXTIME(events.timestamp)) as user_id,"+
		" FIRST(events.timestamp, FROM_UNIXTIME(events.timestamp)) as timestamp, 1 as %s", stepName)

	if i > 0 {
		addSelect = fmt.Sprintf("COALESCE(users.customer_user_id,users.id) as coal_user_id, events.user_id, events.timestamp, 1 as %s", stepName)
	}
	return addSelect
}

func removePresentPropertiesGroupBys(groupBys []model.QueryGroupByProperty) []model.QueryGroupByProperty {
	logFields := log.Fields{
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	filteredProps := make([]model.QueryGroupByProperty, 0)
	for _, prop := range groupBys {
		if prop.EventNameIndex == 0 && prop.EventName == model.UserPropertyGroupByPresent {
			// For $present, event name index is not set and is default 0.
			continue
		}
		filteredProps = append(filteredProps, prop)
	}
	return filteredProps
}

func buildNoneHandledGroupKeys(groupProps []model.QueryGroupByProperty) string {
	logFields := log.Fields{
		"group_props": groupProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	groupKeys := ""

	// Empty handling and null case handling on funnel join.
	for i, v := range groupProps {
		gKey := groupKeyByIndex(v.Index)
		groupSelect := fmt.Sprintf("CASE WHEN %s IS NULL THEN '%s' WHEN %s = '' THEN '%s' ELSE %s END AS %s",
			gKey, model.PropertyValueNone, gKey, model.PropertyValueNone, gKey, gKey)

		groupKeys = appendStatement(groupKeys, groupSelect)
		if i < len(groupProps)-1 {
			groupKeys = groupKeys + ", "
		}
	}

	return groupKeys
}

/*
Funner Query for:
Events:
	$session
	View Project
Group By:
	event_property -> 1. $session -> $day_of_week (categorical)
	user_property -> $present -> $session_count (numerical)
Query:
WITH
step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name=''),

step_0 AS (SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id)
as coal_user_id, events.user_id, events.timestamp, 1 as step_0, CASE WHEN events.properties->>'' IS NULL THEN '$none'
WHEN events.properties->>'' = '' THEN '$none' ELSE events.properties->>'' END AS _group_key_0 FROM events JOIN users
ON events.user_id=users.id WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599' AND
events.event_name_id IN (SELECT id FROM step_0_names WHERE project_id='3' AND name='') ORDER BY coal_user_id, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step_1 AS (SELECT COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as step_1
FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='3' AND timestamp>='1393612200' AND timestamp<='1396290599'
AND events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='3' AND name='View Project') ORDER BY coal_user_id,
events.timestamp ASC), step_1_step_0_users AS (SELECT DISTINCT ON(step_1.coal_user_id) step_1.coal_user_id,
step_1.user_id,step_1.timestamp, step_1 FROM step_0 LEFT JOIN step_1 ON step_0.coal_user_id = step_1.coal_user_id WHERE
step_1.timestamp > step_0.timestamp ORDER BY step_1.coal_user_id, timestamp ASC),

funnel AS (SELECT step_0, step_1, CASE WHEN user_properties.properties->>'' IS NULL THEN '$none' WHEN
user_properties.properties->>'' = '' THEN '$none' ELSE user_properties.properties->>'' END AS _group_key_1,
CASE WHEN _group_key_0 IS NULL THEN '$none' WHEN _group_key_0 = '' THEN '$none' ELSE _group_key_0 END AS
_group_key_0 FROM step_0 LEFT JOIN users on step_0.user_id=users.id LEFT JOIN user_properties on
users.properties_id=user_properties.id  LEFT JOIN step_1_step_0_users ON step_0.coal_user_id=step_1_step_0_users.coal_user_id),

_group_key_1_bounds AS (SELECT percentile_disc(0.02) WITHIN GROUP(ORDER BY _group_key_1::numeric) + 0.00001 AS lbound,
percentile_disc(0.98) WITHIN GROUP(ORDER BY _group_key_1::numeric) AS ubound FROM funnel WHERE _group_key_1 != '$none'),

bucketed AS (SELECT _group_key_0, COALESCE(NULLIF(_group_key_1, '$none'), 'NaN') AS _group_key_1, CASE
WHEN _group_key_1 = '$none' THEN -1 ELSE width_bucket(_group_key_1::numeric, _group_key_1_bounds.lbound::numeric,
COALESCE(NULLIF(_group_key_1_bounds.ubound, _group_key_1_bounds.lbound), _group_key_1_bounds.ubound+1)::numeric, 8)
END AS _group_key_1_bucket, step_0, step_1 FROM funnel, _group_key_1_bounds)

SELECT '$no_group' AS _group_key_0,'$no_group' AS _group_key_1, SUM(step_0) AS step_0, SUM(step_1) AS step_1 FROM
funnel UNION ALL SELECT * FROM ( SELECT _group_key_0, COALESCE(NULLIF(concat(round(min(_group_key_1::numeric), 1),
' - ', round(max(_group_key_1::numeric), 1)), 'NaN - NaN'), '$none') AS _group_key_1, SUM(step_0) AS step_0,
SUM(step_1) AS step_1 FROM bucketed GROUP BY _group_key_0, _group_key_1_bucket ORDER BY _group_key_1_bucket LIMIT 100 ) AS group_funnel
*/
func buildUniqueUsersFunnelQuery(projectId uint64, q model.Query, groupIds []int) (string, []interface{}, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"q":          q,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("invalid no.of events for funnel query")
	}

	funnelSteps := make([]string, 0, 0)
	previousCombinedUsersStepName := ""

	var qStmnt string
	var qParams []interface{}
	// Init joinTimeSelect with step_0 time.
	joinTimeSelect := "step_0.timestamp AS step_0_timestamp"
	for i := range q.EventsWithProperties {
		var addParams []interface{}
		stepName := stepNameByIndex(i)

		isSessionAnalysisReqBool := isSessionAnalysisReq(q.SessionStartEvent, q.SessionEndEvent)
		// Unique users from events filter.
		var addSelect string
		if groupIds[i] != 0 {
			addSelect = buildAddSelectForGroup(stepName, i)
		} else {
			addSelect = buildAddSelect(stepName, i)
		}
		if isSessionAnalysisReqBool && i >= int(q.SessionStartEvent)-1 && i < int(q.SessionEndEvent) {
			if q.EventsWithProperties[i].Name != "$session" {
				addSelect = addSelect + ", events.session_id as session_id"
			} else {
				addSelect = addSelect + ", events.id as session_id"
			}
		}
		egSelect, egParams, egGroupKeys, _ := buildGroupKeyForStepForFunnel(
			projectId, &q.EventsWithProperties[i], q.GroupByProperties, i+1, q.Timezone)
		if egSelect != "" {
			addSelect = joinWithComma(addSelect, egSelect)
		}
		addParams = egParams
		var addJoinStatement string
		if groupIds[i] != 0 {
			addJoinStatement = fmt.Sprintf("JOIN users AS groups ON events.user_id=groups.id JOIN users ON users.group_%d_user_id = groups.id AND users.project_id = ? ", groupIds[i])
		} else {
			addJoinStatement = "JOIN users ON events.user_id=users.id AND users.project_id = ? "
		}
		addJoinStatement = addJoinStatement + getUsersFilterJoinStatement(projectId, q.GlobalUserProperties)
		addParams = append(addParams, projectId)

		var groupBy string
		if i == 0 {
			groupBy = "coal_user_id"
		} else {
			groupBy = "coal_user_id, timestamp"
		}
		addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[i], q.From, q.To,
			"", stepName, addSelect, addParams, addJoinStatement, groupBy, "", q.GlobalUserProperties)
		if len(q.EventsWithProperties) > 1 && i == 0 {
			qStmnt = qStmnt + ", "
		}

		if i == 0 {
			funnelSteps = append(funnelSteps, stepName)
			continue
		}

		prevStepName := stepNameByIndex(i - 1)

		// step_x_to_y - Join users who did step_x after step_y.
		stepXToYName := fmt.Sprintf("%s_%s_users", stepName, prevStepName)

		stepXToYSelect := fmt.Sprintf("%s.coal_user_id, FIRST(%s.user_id, FROM_UNIXTIME(%s.timestamp)) as user_id, FIRST(%s.timestamp, FROM_UNIXTIME(%s.timestamp)) as timestamp, %s", stepName, stepName, stepName, stepName, stepName, stepName)
		if isSessionAnalysisReqBool && i >= int(q.SessionStartEvent)-1 && i < int(q.SessionEndEvent) {
			stepXToYSelect = fmt.Sprintf("%s.coal_user_id, FIRST(%s.user_id, FROM_UNIXTIME(%s.timestamp)) as user_id, FIRST(%s.timestamp, FROM_UNIXTIME(%s.timestamp)) as timestamp,"+
				" FIRST(%s.session_id, FROM_UNIXTIME(%s.timestamp)) as session_id, FIRST(%s, FROM_UNIXTIME(%s.timestamp)) as %s",
				stepName, stepName, stepName, stepName, stepName, stepName, stepName, stepName, stepName, stepName)
		}

		if egGroupKeys != "" {
			stepXToYSelect = joinWithComma(stepXToYSelect, egGroupKeys)
		}
		joinTimeSelect = joinWithComma(joinTimeSelect, fmt.Sprintf("FIRST(%s.timestamp, FROM_UNIXTIME(%s.timestamp)) AS %s_timestamp", stepName, stepName, stepName))
		stepXToYSelect = joinWithComma(stepXToYSelect, joinTimeSelect)
		// re-init joinTimeSelect
		joinTimeSelect = ""

		previousCombinedUsersStepName = prevStepName + "_" + stepNameByIndex(i-2) + "_users"
		stepXToYJoin := buildStepXToYJoin(stepName, prevStepName, previousCombinedUsersStepName, isSessionAnalysisReqBool, q, i)

		stepXToY := buildStepXToY(stepXToYSelect, prevStepName, previousCombinedUsersStepName, stepXToYJoin, stepName, i)

		qStmnt = joinWithComma(qStmnt, as(stepXToYName, stepXToY))

		funnelSteps = append(funnelSteps, stepXToYName)

		if i < len(q.EventsWithProperties)-1 {
			qStmnt = qStmnt + ", "
		}
	}

	funnelCountAliases := make([]string, 0, 0)
	funnelCountTimeAliases := make([]string, 0, 0)
	for i := range q.EventsWithProperties {
		funnelCountAliases = append(funnelCountAliases, fmt.Sprintf("step_%d", i))
		if len(q.EventsWithProperties) > 1 {
			funnelCountTimeAliases = append(funnelCountTimeAliases, fmt.Sprintf("step_%d_timestamp", i))
		}
	}

	var stepsJoinStmnt string
	for i, fs := range funnelSteps {
		if i > 0 {
			// builds "LEFT JOIN step2 on step0_users.coal_user_id=step_0_step_1_users.coal_user_id"
			stepsJoinStmnt = appendStatement(stepsJoinStmnt,
				fmt.Sprintf("LEFT JOIN %s ON %s.coal_user_id=%s.coal_user_id", fs, funnelSteps[i-1], fs))
		}
	}

	userGroupProps := filterGroupPropsByType(q.GroupByProperties, model.PropertyEntityUser)
	userGroupProps = removeEventSpecificUserGroupBys(userGroupProps)
	ugSelect, ugParams, _ := buildGroupKeys(projectId, userGroupProps, q.Timezone)

	propertiesJoinStmnt := ""
	if hasGroupEntity(q.GroupByProperties, model.PropertyEntityUser) {
		propertiesJoinStmnt = fmt.Sprintf("LEFT JOIN users on %s.user_id=users.id", funnelSteps[0])
		// Using string format for project_id condition, as the value is from internal system.
		propertiesJoinStmnt = propertiesJoinStmnt + " AND " + fmt.Sprintf("users.project_id = %d", projectId)
	}

	stepFunnelName := "funnel"
	// select step counts, user properties and event properties group_keys.
	stepFunnelSelect := joinWithComma(funnelCountAliases...)
	if len(q.EventsWithProperties) > 1 {
		for _, str := range funnelCountTimeAliases {
			stepFunnelSelect = joinWithComma(stepFunnelSelect, str)
		}
	}
	stepFunnelSelect = joinWithComma(stepFunnelSelect, ugSelect)
	eventGroupProps := removePresentPropertiesGroupBys(q.GroupByProperties)
	egGroupKeys := buildNoneHandledGroupKeys(eventGroupProps)
	if egGroupKeys != "" {
		stepFunnelSelect = joinWithComma(stepFunnelSelect, egGroupKeys)
	}

	funnelStmnt := "SELECT" + " " + stepFunnelSelect + " " + "FROM" + " " + funnelSteps[0] +
		" " + propertiesJoinStmnt + " " + stepsJoinStmnt
	qStmnt = joinWithComma(qStmnt, as(stepFunnelName, funnelStmnt))
	qParams = append(qParams, ugParams...)

	var aggregateSelectKeys, aggregateFromName, aggregateGroupBys, aggregateOrderBys string
	aggregateFromName = stepFunnelName
	if isGroupByTypeWithBuckets(q.GroupByProperties) {
		stepTimeSelect := ""
		if len(q.EventsWithProperties) > 1 {
			for _, str := range funnelCountTimeAliases {
				if stepTimeSelect == "" {
					stepTimeSelect = str
				} else {
					stepTimeSelect = joinWithComma(stepTimeSelect, str)
				}
			}
		}
		bucketedFromName, bucketedSelectKeys, bucketedGroupBys, bucketedOrderBys :=
			appendNumericalBucketingSteps(&qStmnt, &qParams, q.GroupByProperties,
				stepFunnelName, stepTimeSelect, false,
				strings.Join(funnelCountAliases, ", "))
		aggregateSelectKeys = bucketedSelectKeys
		aggregateFromName = bucketedFromName
		aggregateGroupBys = strings.Join(bucketedGroupBys, ", ")
		aggregateOrderBys = strings.Join(bucketedOrderBys, ", ")
	} else {
		_, _, groupKeys := buildGroupKeys(projectId, q.GroupByProperties, q.Timezone)
		aggregateSelectKeys = groupKeys + ", "
		aggregateFromName = stepFunnelName
		aggregateGroupBys = groupKeys
		aggregateOrderBys = funnelCountAliases[0] + " DESC"
	}

	// builds "SUM(step1) AS step1, SUM(step1) AS step2".
	var rawCountSelect string
	for _, fca := range funnelCountAliases {
		rawCountSelect = joinWithComma(rawCountSelect, fmt.Sprintf("SUM(%s) AS %s", fca, fca))
	}

	avgStepTimeSelect := make([]string, 0, 0)
	if len(q.EventsWithProperties) > 1 {
		for i := 1; i < len(q.EventsWithProperties); i++ {
			avgStepTimeSelect = append(avgStepTimeSelect,
				fmt.Sprintf("AVG(step_%d_timestamp-step_%d_timestamp) AS step_%d_%d%s", i, i-1, i-1, i, model.FunnelTimeSuffix))
		}
	}

	if len(avgStepTimeSelect) > 0 {
		avgStepTimeSelectStmt := joinWithComma(avgStepTimeSelect...)
		rawCountSelect = joinWithComma(rawCountSelect, avgStepTimeSelectStmt)
	}

	var termStmnt string
	if len(q.GroupByProperties) == 0 {
		termStmnt = "SELECT" + " " + rawCountSelect + " " + "FROM" + " " + stepFunnelName
	} else {
		// builds UNION ALL with overall conversion and grouped conversion.
		noGroupAlias := "$no_group"
		var groupKeysPlaceholder string
		for i, group := range q.GroupByProperties {
			groupKeysPlaceholder = groupKeysPlaceholder + fmt.Sprintf("'%s' AS %s", noGroupAlias, groupKeyByIndex(group.Index))
			if i < len(q.GroupByProperties)-1 {
				groupKeysPlaceholder = groupKeysPlaceholder + ","
			}
		}
		noGroupSelect := "SELECT" + " " + joinWithComma(groupKeysPlaceholder, rawCountSelect) +
			" " + "FROM" + " " + stepFunnelName

		limitedGroupBySelect := "SELECT" + " " + aggregateSelectKeys + rawCountSelect + " " +
			"FROM" + " " + aggregateFromName + " " + "GROUP BY" + " " + aggregateGroupBys + " " +
			// order and limit by last step of funnel.
			fmt.Sprintf("ORDER BY %s LIMIT %d", aggregateOrderBys, model.ResultsLimit)

		// wrapped with select to apply limits only to grouped select rows.
		groupBySelect := fmt.Sprintf("SELECT * FROM ( %s ) AS group_funnel", limitedGroupBySelect)

		termStmnt = groupBySelect + " " + "UNION ALL" + " " + noGroupSelect
	}

	qStmnt = appendStatement(qStmnt, termStmnt)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

// buildGroupKeyForStep moved to memsql/event_analytics.go

func buildGroupKeyForStepForFunnel(projectID uint64, eventWithProperties *model.QueryEventWithProperties,
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
	groupSelect, groupSelectParams, groupKeys := "", make([]interface{}, 0), ""
	if ewpIndex == 1 {
		groupSelect, groupSelectParams, groupKeys = buildGroupKeysWithFirst(projectID, groupPropsByStep, timezoneString)
	} else {
		groupSelect, groupSelectParams, groupKeys = buildGroupKeys(projectID, groupPropsByStep, timezoneString)
	}
	return groupSelect, groupSelectParams, groupKeys, groupByUserProperties
}
