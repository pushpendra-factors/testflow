package model

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func RunFunnelQuery(projectId uint64, query Query) (*QueryResult, int, string) {
	if !isValidFunnelQuery(&query) {
		return nil, http.StatusBadRequest, ErrMsgMaxFunnelStepsExceeded
	}

	stmnt, params, err := BuildFunnelQuery(projectId, query)
	if err != nil {
		log.WithError(err).Error(ErrMsgQueryProcessingFailure)
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	logCtx := log.WithFields(log.Fields{"analytics_query": query, "statement": stmnt, "params": params})
	if stmnt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from analytics query.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	result, err := ExecQuery(stmnt, params)
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	// should be done before translation of group keys
	translateNullToZeroOnFunnelResult(result)

	err = addStepConversionPercentageToFunnel(result)
	if err != nil {
		logCtx.WithError(err).Error("Failed adding funnel step conversion percentage.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	err = translateGroupKeysIntoColumnNames(result, query.GroupByProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed translating group keys on result.")
		return nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	sanitizeNumericalBucketRanges(result, &query)

	addQueryToResultMeta(result, query)

	return result, http.StatusOK, "Successfully executed query"
}

func BuildFunnelQuery(projectId uint64, query Query) (string, []interface{}, error) {
	addIndexToGroupByProperties(&query)

	if query.EventsCondition == QueryTypeEventsOccurrence {
		return "", nil, errors.New("funnel on events occurrence is not supported")
	}

	return buildUniqueUsersFunnelQuery(projectId, query)
}

func translateNullToZeroOnFunnelResult(result *QueryResult) {
	var percentageIndexes []int
	var index int
	for _, h := range result.Headers {
		if strings.HasPrefix(h, FunnelConversionPrefix) || strings.HasPrefix(h, StepPrefix) {
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

func addStepConversionPercentageToFunnel(result *QueryResult) error {
	if len(result.Rows) == 0 {
		return errors.New("invalid funnel result")
	}

	stepIndexes := make([]int, 0, 0)
	nonStepIndexes := make([]int, 0, 0)
	for i, header := range result.Headers {
		if strings.HasPrefix(header, StepPrefix) {
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
			FunnelConversionPrefix, result.Headers[si-1], result.Headers[si]))
	}

	headers = append(headers, fmt.Sprintf("%s%s", FunnelConversionPrefix, "overall"))
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

			prevCount, err := getAggrAsFloat64(result.Rows[ri][ci-1])
			if err != nil {
				return err
			}

			curCount, err := getAggrAsFloat64(result.Rows[ri][ci])
			if err != nil {
				return err
			}

			row = append(row, getConversionPercentageAsString(prevCount, curCount))
		}

		// add overall conversion.
		firstStepCount, err := getAggrAsFloat64(result.Rows[ri][stepIndexes[0]])
		if err != nil {
			return err
		}

		lastIndex := stepIndexes[len(stepIndexes)-1]
		lastStepCount, err := getAggrAsFloat64(result.Rows[ri][lastIndex])
		if err != nil {
			return err
		}
		row = append(row, getConversionPercentageAsString(firstStepCount, lastStepCount))

		result.Rows[ri] = row // row with conversion.
	}

	return nil
}

func getConversionPercentageAsString(prevCount float64, curCount float64) string {
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
	if start != 0 && end != 0 && start < end {
		return true
	}
	return false
}

func buildStepXToYJoin(stepName string, prevStepName string, previousCombinedUsersStepName string,
	isSessionAnalysisReqBool bool, q Query, i int) string {

	stepXToYJoin := fmt.Sprintf("LEFT JOIN %s ON %s.coal_user_id = %s.coal_user_id WHERE %s.timestamp > %s.timestamp",
		stepName, previousCombinedUsersStepName, stepName, stepName, previousCombinedUsersStepName)
	if i == 1 {
		stepXToYJoin = fmt.Sprintf("LEFT JOIN %s ON %s.coal_user_id = %s.coal_user_id WHERE %s.timestamp > %s.timestamp",
			stepName, prevStepName, stepName, stepName, prevStepName)
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

	stepXToY := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY %s.coal_user_id, timestamp ASC", stepXToYSelect, previousCombinedUsersStepName, stepXToYJoin, stepName)
	if i == 1 {
		stepXToY = fmt.Sprintf("SELECT %s FROM %s %s ORDER BY %s.coal_user_id, timestamp ASC", stepXToYSelect, prevStepName, stepXToYJoin, stepName)
	}
	return stepXToY
}
func buildAddSelect(stepName string, i int) string {
	addSelect := fmt.Sprintf("DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as %s", stepName)

	if i > 0 {
		addSelect = fmt.Sprintf("COALESCE(users.customer_user_id,events.user_id) as coal_user_id, events.user_id, events.timestamp, 1 as %s", stepName)
	}
	return addSelect
}

func removePresentPropertiesGroupBys(groupBys []QueryGroupByProperty) []QueryGroupByProperty {
	filteredProps := make([]QueryGroupByProperty, 0)
	for _, prop := range groupBys {
		if prop.EventNameIndex == 0 && prop.EventName == UserPropertyGroupByPresent {
			// For $present, event name index is not set and is default 0.
			continue
		}
		filteredProps = append(filteredProps, prop)
	}
	return filteredProps
}

func buildNoneHandledGroupKeys(groupProps []QueryGroupByProperty) string {
	groupKeys := ""

	// Empty handling and null case handling on funnel join.
	for i, v := range groupProps {
		gKey := groupKeyByIndex(v.Index)
		groupSelect := fmt.Sprintf("CASE WHEN %s IS NULL THEN '%s' WHEN %s = '' THEN '%s' ELSE %s END AS %s",
			gKey, PropertyValueNone, gKey, PropertyValueNone, gKey, gKey)

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

_group_key_1_bounds AS (SELECT percentile_disc(0.1) WITHIN GROUP(ORDER BY _group_key_1::numeric desc) AS ubound,
percentile_disc(0.9) WITHIN GROUP(ORDER BY _group_key_1::numeric desc) AS lbound FROM funnel WHERE _group_key_1 != '$none'),

bucketed AS (SELECT _group_key_0, COALESCE(NULLIF(_group_key_1, '$none'), 'NaN') AS _group_key_1, CASE
WHEN _group_key_1 = '$none' THEN -1 ELSE width_bucket(_group_key_1::numeric, _group_key_1_bounds.lbound::numeric,
COALESCE(NULLIF(_group_key_1_bounds.ubound, _group_key_1_bounds.lbound), _group_key_1_bounds.ubound+1)::numeric, 8)
END AS _group_key_1_bucket, step_0, step_1 FROM funnel, _group_key_1_bounds)

SELECT '$no_group' AS _group_key_0,'$no_group' AS _group_key_1, SUM(step_0) AS step_0, SUM(step_1) AS step_1 FROM
funnel UNION ALL SELECT * FROM ( SELECT _group_key_0, COALESCE(NULLIF(concat(round(min(_group_key_1::numeric), 1),
' - ', round(max(_group_key_1::numeric), 1)), 'NaN - NaN'), '$none') AS _group_key_1, SUM(step_0) AS step_0,
SUM(step_1) AS step_1 FROM bucketed GROUP BY _group_key_0, _group_key_1_bucket ORDER BY _group_key_1_bucket LIMIT 100 ) AS group_funnel
*/
func buildUniqueUsersFunnelQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("invalid no.of events for funnel query")
	}

	funnelSteps := make([]string, 0, 0)
	previousCombinedUsersStepName := ""

	var qStmnt string
	var qParams []interface{}
	for i := range q.EventsWithProperties {
		var addParams []interface{}
		stepName := stepNameByIndex(i)

		isSessionAnalysisReqBool := isSessionAnalysisReq(q.SessionStartEvent, q.SessionEndEvent)
		// Unique users from events filter.
		addSelect := buildAddSelect(stepName, i)
		if isSessionAnalysisReqBool && i >= int(q.SessionStartEvent)-1 && i < int(q.SessionEndEvent) {
			if q.EventsWithProperties[i].Name != "$session" {
				addSelect = addSelect + ", events.session_id as session_id"
			} else {
				addSelect = addSelect + ", events.id::text as session_id"
			}
		}
		egSelect, egParams, egGroupKeys, groupByUserProperties := buildGroupKeyForStep(
			&q.EventsWithProperties[i], q.GroupByProperties, i+1)
		if egSelect != "" {
			addSelect = joinWithComma(addSelect, egSelect)
		}
		addParams = egParams
		addJoinStatement := "JOIN users ON events.user_id=users.id"
		if groupByUserProperties && !hasWhereEntity(q.EventsWithProperties[i], PropertyEntityUser) {
			// If event has filter on user property, JOIN on user_properties is added in next step.
			// Skip addding here to avoid duplication.
			addJoinStatement += " JOIN user_properties on events.user_properties_id=user_properties.id"
		}
		addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[i], q.From, q.To,
			"", stepName, addSelect, addParams, addJoinStatement, "", "coal_user_id, events.timestamp ASC")

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

		stepXToYSelect := fmt.Sprintf("DISTINCT ON(%s.coal_user_id) %s.coal_user_id, %s.user_id,%s.timestamp, %s", stepName, stepName, stepName, stepName, stepName)
		if isSessionAnalysisReqBool && i >= int(q.SessionStartEvent) && i < int(q.SessionEndEvent) {
			stepXToYSelect = fmt.Sprintf("DISTINCT ON(%s.coal_user_id) %s.coal_user_id, %s.user_id,%s.timestamp, %s.session_id, %s", stepName, stepName, stepName, stepName, stepName, stepName)
		}

		if egGroupKeys != "" {
			stepXToYSelect = joinWithComma(stepXToYSelect, egGroupKeys)
		}

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
	for i := range q.EventsWithProperties {
		funnelCountAliases = append(funnelCountAliases, fmt.Sprintf("step_%d", i))
	}

	var stepsJoinStmnt string
	for i, fs := range funnelSteps {
		if i > 0 {
			// builds "LEFT JOIN step2 on step0_users.coal_user_id=step_0_step_1_useres.coal_user_id"
			stepsJoinStmnt = appendStatement(stepsJoinStmnt,
				fmt.Sprintf("LEFT JOIN %s ON %s.coal_user_id=%s.coal_user_id", fs, funnelSteps[i-1], fs))
		}
	}

	userGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityUser)
	userGroupProps = removeEventSpecificUserGroupBys(userGroupProps)
	ugSelect, ugParams, _ := buildGroupKeys(userGroupProps)

	properitesJoinStmnt := ""
	if hasGroupEntity(q.GroupByProperties, PropertyEntityUser) {
		properitesJoinStmnt = fmt.Sprintf("LEFT JOIN users on %s.user_id=users.id", funnelSteps[0])
		properitesJoinStmnt = properitesJoinStmnt + " " + "LEFT JOIN user_properties on users.properties_id=user_properties.id"
	}

	stepFunnelName := "funnel"
	// select step counts, user properties and event properties group_keys.
	stepFunnelSelect := joinWithComma(funnelCountAliases...)
	stepFunnelSelect = joinWithComma(stepFunnelSelect, ugSelect)
	eventGroupProps := removePresentPropertiesGroupBys(q.GroupByProperties)
	egGroupKeys := buildNoneHandledGroupKeys(eventGroupProps)
	if egGroupKeys != "" {
		stepFunnelSelect = joinWithComma(stepFunnelSelect, egGroupKeys)
	}

	funnelStmnt := "SELECT" + " " + stepFunnelSelect + " " + "FROM" + " " + funnelSteps[0] +
		" " + properitesJoinStmnt + " " + stepsJoinStmnt
	qStmnt = joinWithComma(qStmnt, as(stepFunnelName, funnelStmnt))
	qParams = append(qParams, ugParams...)

	var aggregateSelectKeys, aggregateFromName, aggregateGroupBys, aggregateOrderBys string
	aggregateFromName = stepFunnelName
	if hasNumericalGroupBy(q.GroupByProperties) {
		bucketedFromName, bucketedSelectKeys, bucketedGroupBys, bucketedOrderBys :=
			appendNumericalBucketingSteps(&qStmnt, q.GroupByProperties, stepFunnelName, "", false,
				strings.Join(funnelCountAliases, ", "))
		aggregateSelectKeys = bucketedSelectKeys
		aggregateFromName = bucketedFromName
		aggregateGroupBys = strings.Join(bucketedGroupBys, ", ")
		aggregateOrderBys = strings.Join(bucketedOrderBys, ", ")
	} else {
		_, _, groupKeys := buildGroupKeys(q.GroupByProperties)
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
			fmt.Sprintf("ORDER BY %s LIMIT %d", aggregateOrderBys, ResultsLimit)

		// wrapped with select to apply limits only to grouped select rows.
		groupBySelect := fmt.Sprintf("SELECT * FROM ( %s ) AS group_funnel", limitedGroupBySelect)

		termStmnt = noGroupSelect + " " + "UNION ALL" + " " + groupBySelect
	}

	qStmnt = appendStatement(qStmnt, termStmnt)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

// builds group keys for event properties for given step (event_with_properties).
func buildGroupKeyForStep(eventWithProperties *QueryEventWithProperties,
	groupProps []QueryGroupByProperty, ewpIndex int) (string, []interface{}, string, bool) {

	groupPropsByStep := make([]QueryGroupByProperty, 0, 0)
	groupByUserProperties := false
	for i := range groupProps {
		if groupProps[i].EventNameIndex == ewpIndex &&
			groupProps[i].EventName == eventWithProperties.Name {
			groupPropsByStep = append(groupPropsByStep, groupProps[i])
			if groupProps[i].Entity == PropertyEntityUser {
				groupByUserProperties = true
			}
		}
	}

	groupSelect, groupSelectParams, groupKeys := buildGroupKeys(groupPropsByStep)
	return groupSelect, groupSelectParams, groupKeys, groupByUserProperties
}
