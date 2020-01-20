package model

import (
	"encoding/json"
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type QueryProperty struct {
	// Entity: user or event.
	Entity string `json:"en"`
	// Type: categorical or numerical
	Type      string `json:"ty"`
	Property  string `json:"pr"`
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}

type QueryGroupByProperty struct {
	// Entity: user or event.
	Entity   string `json:"en"`
	Property string `json:"pr"`
	Index    int    `json:"in"`
	// group by specific event name.
	EventName string `json:"ena"`
}

type QueryEventWithProperties struct {
	Name       string          `json:"na"`
	Properties []QueryProperty `json:"pr"`
}

type Query struct {
	Class                string                     `json:"cl"`
	Type                 string                     `json:"ty"`
	EventsCondition      string                     `json:"ec"` // all or any
	EventsWithProperties []QueryEventWithProperties `json:"ewp"`
	GroupByProperties    []QueryGroupByProperty     `json:"gbp"`
	GroupByTimestamp     interface{}                `json:"gbt"`
	Timezone             string                     `json:"tz"`
	From                 int64                      `json:"fr"`
	To                   int64                      `json:"to"`
	// Deprecated: Keeping it for old dashboard units.
	OverridePeriod bool `json:"ovp"`
}

type QueryResutlMeta struct {
	Query Query `json:"query"`
}

type QueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Meta    QueryResutlMeta `json:"meta"`
}

type DateTimePropertyValue struct {
	From           int64 `json:"fr"`
	To             int64 `json:"to"`
	OverridePeriod bool  `json:"ovp"`
}

const (
	PropertyEntityUser  = "user"
	PropertyEntityEvent = "event"

	PropertyValueNone = "$none"

	EventCondAnyGivenEvent = "any_given_event"
	EventCondAllGivenEvent = "all_given_event"

	QueryClassInsights = "insights"
	QueryClassFunnel   = "funnel"
	QueryClassChannel  = "channel"

	QueryTypeEventsOccurrence = "events_occurrence"
	QueryTypeUniqueUsers      = "unique_users"

	ErrUnsupportedGroupByEventPropertyOnUserQuery = "group by event property is not supported for user query"
	ErrMsgQueryProcessingFailure                  = "Failed processing query"
	ErrMsgMaxFunnelStepsExceeded                  = "Max funnel steps exceeded"

	SelectDefaultEventFilter              = "events.id as event_id, events.user_id as event_user_id"
	SelectDefaultEventFilterWithDistinct  = "DISTINCT(events.id) as event_id, events.user_id as event_user_id"
	SelectDefaultEventFilterByAlias       = "event_id, event_user_id"
	SelectCoalesceCustomerUserIDAndUserID = "COALESCE(users.customer_user_id, event_user_id)"

	GroupKeyPrefix  = "_group_key_"
	AliasDateTime   = "datetime"
	AliasAggr       = "count"
	DefaultTimezone = "UTC"
	ResultsLimit    = 20
	MaxResultsLimit = 100000

	StepPrefix             = "step_"
	FunnelConversionPrefix = "conversion_"
)

const (
	EqualsOpStr   = "equals"
	EqualsOp      = "="
	NotEqualOpStr = "notEqual"
	NotEqualOp    = "!="
)

var queryOps = map[string]string{
	EqualsOpStr:          EqualsOp,
	NotEqualOpStr:        NotEqualOp,
	"greaterThan":        ">",
	"lesserThan":         "<",
	"greaterThanOrEqual": ">=",
	"lesserThanOrEqual":  "<=",
	"contains":           "LIKE",
	"notContains":        "NOT LIKE",
}

const (
	GroupByTimestampHour = "hour"
	GroupByTimestampDate = "date"
)

var groupByTimestampTypes = []string{
	GroupByTimestampDate,
	GroupByTimestampHour,
}

func (query *Query) GetGroupByTimestamp() string {
	switch query.GroupByTimestamp.(type) {
	case bool:
		// For query objects on old dashboard units,
		// with GroupByTimestamp as bool and true, to work.
		if query.GroupByTimestamp.(bool) {
			windowInSecs := query.To - query.From
			if windowInSecs <= 86400 {
				return GroupByTimestampHour
			}
			return GroupByTimestampDate
		}

		return ""
	case string:
		return query.GroupByTimestamp.(string)
	default:
		return ""
	}
}

func with(stmnt string) string {
	return fmt.Sprintf("WITH %s", stmnt)
}

func getOp(OpStr string) string {
	v, ok := queryOps[OpStr]
	if !ok {
		log.Errorf("invalid query operator %s, using default", OpStr)
		return EqualsOp
	}

	return v
}

func getPropertyEntityField(entityName string) string {
	if entityName == PropertyEntityUser {
		return "user_properties.properties"
	} else if entityName == PropertyEntityEvent {
		return "events.properties"
	}

	return ""
}

func as(asName, asQuery string) string {
	return fmt.Sprintf("%s AS (%s)", asName, asQuery)
}

func appendStatement(x, y string) string {
	return fmt.Sprintf("%s %s", x, y)
}

func DecodeDateTimePropertyValue(dateTimeJson string) (*DateTimePropertyValue, error) {
	var dateTimeProperty DateTimePropertyValue
	err := json.Unmarshal([]byte(dateTimeJson), &dateTimeProperty)
	if err != nil {
		return &dateTimeProperty, err
	}

	return &dateTimeProperty, nil
}

func isValidLogicalOp(op string) bool {
	return op == "AND" || op == "OR"
}

func buildWhereFromProperties(properties []QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	pLen := len(properties)
	if pLen == 0 {
		return rStmnt, rParams, nil
	}

	rParams = make([]interface{}, 0, 0)
	for i, p := range properties {
		// defaults logic op if not given.
		if p.LogicalOp == "" {
			p.LogicalOp = "AND"
		}

		if !isValidLogicalOp(p.LogicalOp) {
			return rStmnt, rParams, errors.New("invalid logical op on where condition")
		}

		propertyEntity := getPropertyEntityField(p.Entity)
		propertyOp := getOp(p.Operator)

		if p.Value != PropertyValueNone {
			var pStmnt string
			if p.Type == U.PropertyTypeDateTime {
				pStmnt = fmt.Sprintf("(%s->>?>=? AND %s->>?<=?)", propertyEntity, propertyEntity)

				dateTimeValue, err := DecodeDateTimePropertyValue(p.Value)
				if err != nil {
					log.WithError(err).Error("Failed reading timestamp on user join query.")
					return "", nil, err
				}
				rParams = append(rParams, p.Property, dateTimeValue.From, p.Property, dateTimeValue.To)
			} else if p.Type == U.PropertyTypeNumerical {
				// convert to float for numerical properties.
				pStmnt = fmt.Sprintf("(%s->>?)::float %s ?", propertyEntity, propertyOp)
				rParams = append(rParams, p.Property, p.Value)
			} else {
				// categorical property type.
				var pValue string
				if p.Operator == "contains" || p.Operator == "notContains" {
					pValue = fmt.Sprintf("%%%s%%", p.Value)
				} else {
					pValue = p.Value
				}

				pStmnt = fmt.Sprintf("%s->>? %s ?", propertyEntity, propertyOp)
				rParams = append(rParams, p.Property, pValue)
			}

			if i == 0 {
				rStmnt = pStmnt
			} else {
				rStmnt = fmt.Sprintf("%s %s %s", rStmnt, p.LogicalOp, pStmnt)
			}

			continue
		}

		// where condition for $none value.
		var whereCond string
		if propertyOp == EqualsOp {
			// i.e: (NOT jsonb_exists(events.properties, 'property_name') OR events.properties->>'property_name'='')
			whereCond = fmt.Sprintf("(NOT jsonb_exists(%s, ?) OR %s->>?='')", propertyEntity, propertyEntity)
		} else if propertyOp == NotEqualOp {
			// i.e: (jsonb_exists(events.properties, 'property_name') AND events.properties->>'property_name'!='')
			whereCond = fmt.Sprintf("(jsonb_exists(%s, ?) AND %s->>?!='')", propertyEntity, propertyEntity)
		} else {
			return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
		}

		if i == 0 {
			rStmnt = whereCond
		} else {
			rStmnt = fmt.Sprintf("%s %s %s", rStmnt, p.LogicalOp, whereCond)
		}

		rParams = append(rParams, p.Property, p.Property)
	}

	return rStmnt, rParams, nil
}

// Alias for group by properties gk_1, gk_2.
func groupKeyByIndex(i int) string {
	return fmt.Sprintf("%s%d", GroupKeyPrefix, i)
}

func stepNameByIndex(i int) string {
	return fmt.Sprintf("%s%d", StepPrefix, i)
}

// Translates empty and null group by property values as $none on select.
// CASE WHEN events.properties->>'x' IS NULL THEN '$none' WHEN events.properties->>'x' = '' THEN '$none'
// ELSE events.properties->>'x' END as _group_key_0
func getNoneHandledGroupBySelect(groupProp QueryGroupByProperty, groupKey string) (string, []interface{}) {
	entityField := getPropertyEntityField(groupProp.Entity)
	groupSelect := fmt.Sprintf("CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' ELSE %s->>? END AS %s",
		entityField, PropertyValueNone, entityField, PropertyValueNone, entityField, groupKey)
	groupSelectParams := []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	return groupSelect, groupSelectParams
}

// groupBySelect: user_properties.properties->>'age' as gk_1, events.properties->>'category' as gk_2
// groupByKeys: gk_1, gk_2
// How to use?
// select user_properties.properties->>'age' as gk_1, events.properties->>'category' as gk_2 from events
// group by gk_1, gk_2
func buildGroupKeys(groupProps []QueryGroupByProperty) (groupSelect string,
	groupSelectParams []interface{}, groupKeys string) {

	groupSelectParams = make([]interface{}, 0, 0)

	for i, v := range groupProps {
		// Order of group is preserved as received.
		gKey := groupKeyByIndex(v.Index)
		noneSelect, noneParams := getNoneHandledGroupBySelect(v, gKey)
		groupSelect = groupSelect + noneSelect
		groupKeys = groupKeys + gKey
		if i < len(groupProps)-1 {
			groupSelect = groupSelect + ", "
			groupKeys = groupKeys + ", "
		}
		groupSelectParams = append(groupSelectParams, noneParams...)
	}

	return groupSelect, groupSelectParams, groupKeys
}

// builds group keys with step of corresponding user given event name.
// i.e step_0.__group_key_0, step_1.group_key_1
func buildEventGroupKeysWithStep(groupProps []QueryGroupByProperty,
	ewps []QueryEventWithProperties) (groupKeys string) {

	eventGroupProps := filterGroupPropsByType(groupProps, PropertyEntityEvent)
	stepIndexByEventName := make(map[string]int, 0)
	for i, ewp := range ewps {
		stepIndexByEventName[ewp.Name] = i
	}

	for _, gp := range eventGroupProps {
		var groupKey string

		if gp.EventName != "" {
			if stepIndex, exists := stepIndexByEventName[gp.EventName]; exists {
				groupKey = fmt.Sprintf("%s.%s", stepNameByIndex(stepIndex),
					groupKeyByIndex(gp.Index))
			}
		}

		if groupKey == "" {
			// if group by property not provided with an event name.
			// refer to the first step(step_0) by default.
			groupKey = fmt.Sprintf("%s.%s", stepNameByIndex(0),
				groupKeyByIndex(gp.Index))
		}

		groupKeys = joinWithComma(groupKeys, groupKey)
	}

	return groupKeys
}

// Adds a step of events filter with QueryEventWithProperties.
func addFilterEventsWithPropsQuery(projectId uint64, qStmnt *string, qParams *[]interface{},
	qep QueryEventWithProperties, from int64, to int64, fromStr string,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string, groupBy string, orderBy string) error {

	if (from == 0 && fromStr == "") || to == 0 {
		return errors.New("invalid timerange on events filter")
	}

	if addSelecStmnt == "" {
		return errors.New("invalid select on events filter")
	}

	rStmnt := "SELECT " + addSelecStmnt + " FROM events" + " " + addJoinStmnt

	// join user property, if user_property present on event with properties list.
	if hasWhereEntity(qep, PropertyEntityUser) {
		rStmnt = appendStatement(rStmnt, "LEFT JOIN user_properties ON events.user_properties_id=user_properties.id")
	}

	var fromTimestamp string
	if from > 0 {
		fromTimestamp = "?"
	} else if fromStr != "" {
		fromTimestamp = fromStr // allows from_timestamp from another step.
	}

	var eventNamesCacheStmnt string
	eventNamesRef := "event_names"
	if stepName != "" {
		eventNamesRef = fmt.Sprintf("%s_names", stepName)
		eventNamesCacheStmnt = as(eventNamesRef, "SELECT id, project_id, name FROM event_names WHERE project_id=? AND name=?")
		*qParams = append(*qParams, projectId, qep.Name)
	}

	whereCond := fmt.Sprintf("WHERE events.project_id=? AND timestamp>=%s AND timestamp<=?"+
		// select id of event_names from names step.
		" "+"AND events.event_name_id IN (SELECT id FROM %s WHERE project_id=? AND name=?)", fromTimestamp, eventNamesRef)
	rStmnt = appendStatement(rStmnt, whereCond)

	// adds params in order of '?'.
	if addSelecStmnt != "" && addSelectParams != nil {
		*qParams = append(*qParams, addSelectParams...)
	}
	*qParams = append(*qParams, projectId)
	if from > 0 {
		*qParams = append(*qParams, from)
	}
	*qParams = append(*qParams, to, projectId, qep.Name)

	// mergeCond for whereProperties can also be 'OR'.
	wStmnt, wParams, err := buildWhereFromProperties(qep.Properties)
	if err != nil {
		return err
	}

	if wStmnt != "" {
		rStmnt = rStmnt + " AND " + fmt.Sprintf("( %s )", wStmnt)
		*qParams = append(*qParams, wParams...)
	}

	if groupBy != "" {
		rStmnt = fmt.Sprintf("%s GROUP BY %s", rStmnt, groupBy)
	}

	if orderBy != "" {
		rStmnt = fmt.Sprintf("%s ORDER BY %s", rStmnt, orderBy)
	}

	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
	}

	if eventNamesCacheStmnt != "" {
		rStmnt = joinWithComma(eventNamesCacheStmnt, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)

	return nil
}

func hasWhereEntity(ewp QueryEventWithProperties, entity string) bool {
	for _, p := range ewp.Properties {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func joinWithComma(x ...string) string {
	var y string
	for i, v := range x {
		if v != "" {
			if i == 0 || y == "" {
				y = v
			} else {
				y = fmt.Sprintf("%s, %s", y, v)
			}
		}
	}

	return y
}

func hasGroupEntity(props []QueryGroupByProperty, entity string) bool {
	for _, p := range props {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func addJoinLatestUserPropsQuery(groupProps []QueryGroupByProperty, refStepName string, stepName string,
	qStmnt *string, qParams *[]interface{}, addSelect string) string {

	groupSelect, gSelectParams, gKeys := buildGroupKeys(groupProps)

	rStmnt := "SELECT " + joinWithComma(groupSelect, addSelect) + " from " + refStepName +
		" " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id"

	if hasGroupEntity(groupProps, PropertyEntityUser) {
		rStmnt = rStmnt + " " + " LEFT JOIN user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id"
	}

	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)
	*qParams = append(*qParams, gSelectParams...)

	return gKeys
}

func filterGroupPropsByType(gp []QueryGroupByProperty, entity string) []QueryGroupByProperty {
	groupProps := make([]QueryGroupByProperty, 0)
	for _, v := range gp {
		if v.Entity == entity {
			groupProps = append(groupProps, v)
		}
	}
	return groupProps
}

func appendOrderByAggr(qStmnt string) string {
	return fmt.Sprintf("%s ORDER BY %s DESC", qStmnt, AliasAggr)
}

func appendSelectTimestampColIfRequired(stmnt string, isRequired bool) string {
	if !isRequired {
		return stmnt
	}

	return joinWithComma(stmnt, AliasDateTime)
}

func getSelectTimestampByType(timestampType, timezone string) string {
	var selectTz string

	if timezone == "" {
		selectTz = DefaultTimezone
	} else {
		selectTz = timezone
	}

	var selectStr string
	if timestampType == GroupByTimestampHour {
		selectStr = fmt.Sprintf("date_trunc('hour', to_timestamp(timestamp) AT TIME ZONE '%s')", selectTz)
	} else {
		// defaults to GroupByTimestampDate.
		selectStr = fmt.Sprintf("date_trunc('day', to_timestamp(timestamp) AT TIME ZONE '%s')", selectTz)
	}

	return selectStr
}

func appendSelectTimestampIfRequired(stmnt string, groupByTimestamp string, timezone string) string {
	if groupByTimestamp == "" {
		return stmnt
	}

	return joinWithComma(stmnt, fmt.Sprintf("%s as %s",
		getSelectTimestampByType(groupByTimestamp, timezone), AliasDateTime))
}

func appendGroupByTimestampIfRequired(qStmnt string, isRequired bool, groupKeys ...string) string {
	// Added groups with timestamp.
	groups := make([]string, 0, 0)
	if isRequired {
		groups = append(groups, AliasDateTime)
	}
	groups = append(groups, groupKeys...)
	qStmnt = appendGroupBy(qStmnt, groups...)
	return qStmnt
}

func appendGroupBy(qStmnt string, gKeys ...string) string {
	if len(gKeys) == 0 || (len(gKeys) == 1 && gKeys[0] == "") {
		return qStmnt
	}

	return fmt.Sprintf("%s GROUP BY %s", qStmnt, joinWithComma(gKeys...))
}

func appendLimitByCondition(qStmnt string, groupProps []QueryGroupByProperty, groupByTimestamp bool) string {
	if len(groupProps) == 1 && !groupByTimestamp {
		return fmt.Sprintf("%s LIMIT %d", qStmnt, ResultsLimit)
	}

	// Limited with max limit on SQL. Limited on server side.
	return fmt.Sprintf("%s LIMIT %d", qStmnt, MaxResultsLimit)
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

func addEventFilterStepsForUniqueUsersQuery(projectId uint64, q *Query,
	qStmnt *string, qParams *[]interface{}) []string {

	var stepSelect string
	var stepOrderBy string

	if q.GetGroupByTimestamp() != "" {
		selectTimestamp := getSelectTimestampByType(q.GetGroupByTimestamp(), q.Timezone)
		// select and order by with datetime.
		stepSelect = fmt.Sprintf("DISTINCT ON(events.user_id, %s)", selectTimestamp) +
			" " + joinWithComma("events.user_id as event_user_id",
			fmt.Sprintf("%s as %s", selectTimestamp, AliasDateTime))

		stepOrderBy = fmt.Sprintf("events.user_id, %s, events.timestamp ASC", AliasDateTime)
	} else {
		// default select.
		stepSelect = "DISTINCT ON(events.user_id) events.user_id as event_user_id"
	}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egSelect, egParams, _ := buildGroupKeys(eventGroupProps)

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		stepSelect = joinWithComma(stepSelect, egSelect)

		if stepOrderBy == "" {
			// Using first occurred event_properites after distinct on user_id.
			stepOrderBy = "events.user_id, events.timestamp ASC"
		}
	}

	steps := make([]string, 0, 0)
	for i, ewp := range q.EventsWithProperties {
		refStepName := stepNameByIndex(i)
		steps = append(steps, refStepName)

		addFilterEventsWithPropsQuery(projectId, qStmnt, qParams, ewp, q.From, q.To,
			"", refStepName, stepSelect, egParams, "", "", stepOrderBy)

		if i < len(q.EventsWithProperties)-1 {
			*qStmnt = *qStmnt + ","
		}
	}

	return steps
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
	eventGroupProps := filterGroupPropsByType(query.GroupByProperties, PropertyEntityEvent)
	_, _, egKeys := buildGroupKeys(eventGroupProps)

	// select
	userGroupProps := filterGroupPropsByType(query.GroupByProperties, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	*qParams = append(*qParams, ugSelectParams...)
	// order of group keys changes here if users and event
	// group by used together, but translated correctly.
	termSelect := joinWithComma(ugSelect, egKeys)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	termSelect = appendSelectTimestampColIfRequired(termSelect, isGroupByTimestamp)
	termSelect = joinWithComma(termSelect, fmt.Sprintf("COUNT(DISTINCT(%s)) AS %s",
		SelectCoalesceCustomerUserIDAndUserID, AliasAggr))

	// group by
	termStmnt := "SELECT " + termSelect + " FROM " + refStep
	termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStep + ".event_user_id=users.id"
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	_, _, groupKeys := buildGroupKeys(query.GroupByProperties)
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, isGroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, query.GroupByProperties, isGroupByTimestamp)

	*qStmnt = appendStatement(*qStmnt, termStmnt)
}

/*
buildUniqueUsersWithAllGivenEventsQuery builds a query like below,
Group by: user_properties, event_properties

Example: Query without date and with group by properties.

WITH
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
),
users_intersect AS (
    SELECT step0.event_user_id as event_user_id, step0.group_prop2 as group_prop2 from step0
    JOIN step1 ON step0.event_user_id = step1.event_user_id
)
SELECT user_properties.properties->>'$region' as group_prop1, group_prop2,
COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))) FROM users_intersect
LEFT JOIN users ON users_intersect.event_user_id=users.id
LEFT JOIN user_properties ON users.id=user_properties.user_id and user_properties.id=users.properties_id
GROUP BY group_prop1, group_prop2 ORDER BY count DESC LIMIT 10000;

-- Expanded replacement for COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))):
real_users AS (SELECT date, COALESCE(users.customer_user_id, event_user_id) as real_user_id FROM
users_intersect LEFT JOIN users ON users_intersect.event_user_id=users.id GROUP BY date, real_user_id)
SELECT date, COUNT(real_user_id) AS count from real_users GROUP BY date ORDER BY count DESC
--

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
func buildUniqueUsersWithAllGivenEventsQuery(projectId uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps := addEventFilterStepsForUniqueUsersQuery(projectId, &query, &qStmnt, &qParams)

	// users intersection
	intersectSelect := fmt.Sprintf("%s.event_user_id as event_user_id", steps[0])
	if query.GetGroupByTimestamp() != "" {
		intersectSelect = joinWithComma(intersectSelect,
			fmt.Sprintf("%s.%s as %s", steps[0], AliasDateTime, AliasDateTime))
	}

	eventGroupProps := filterGroupPropsByType(query.GroupByProperties, PropertyEntityEvent)
	// adds group by event property with user selected event (step).
	eventGroupKeysWithStep := buildEventGroupKeysWithStep(eventGroupProps,
		query.EventsWithProperties)
	intersectSelect = joinWithComma(intersectSelect, eventGroupKeysWithStep)

	var intersectJoin string
	for i := range steps {
		if i > 0 {
			intersectJoin = intersectJoin + " " + fmt.Sprintf("JOIN %s ON %s.event_user_id = %s.event_user_id",
				steps[i], steps[i], steps[i-1])

			// include date also intersection condition on
			// group by timestamp.
			if query.GetGroupByTimestamp() != "" {
				intersectJoin = intersectJoin + " " + fmt.Sprintf("AND %s.%s = %s.%s",
					steps[i], AliasDateTime, steps[i-1], AliasDateTime)
			}
		}
	}
	stepUsersIntersect := "users_intersect"
	qStmnt = joinWithComma(qStmnt, as(stepUsersIntersect,
		fmt.Sprintf("SELECT %s FROM %s %s", intersectSelect, steps[0], intersectJoin)))

	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, stepUsersIntersect)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersWithAnyGivenEventsQuery
Group By: user_properties, event_properties

Example: Query without date and with group by properties.

WITH
	step0 AS (
		-- Using DISTINCT ON user_id for getting unique users and event properties on first
		occurrence of the event done by the user --
		SELECT DISTINCT ON(events.user_id) events.user_id as event_user_id,
		events.properties->>'category' as gk_1 FROM events
		LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='View Project')
		ORDER BY events.user_id, events.timestamp ASC
	),
	step1 AS (
		SELECT DISTINCT ON(events.user_id) events.user_id as event_user_id,
		events.properties->>'category' as gk_1 FROM events
		LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='Fund Project')
		ORDER BY events.user_id, events.timestamp ASC
	),
	users_union AS (
		-- Using union all instead of union for avoiding unwanted
		dedeuplication involving event_property from steps --
		SELECT step0.event_user_id as event_user_id, gk_1 FROM step0 UNION ALL
		SELECT step1.event_user_id as event_user_id, gk_1 FROM step1
	)
	SELECT user_properties.properties->>'gender' as gk_0, gk_1,
	COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))) FROM users_union
	LEFT JOIN users ON users_union.event_user_id=users.id
	LEFT JOIN user_properties ON users.id=user_properties.user_id and user_properties.id=users.properties_id
	GROUP BY gk_0, gk_1 ORDER BY count DESC LIMIT 10000;

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
func buildUniqueUsersWithAnyGivenEventsQuery(projectId uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps := addEventFilterStepsForUniqueUsersQuery(projectId, &query, &qStmnt, &qParams)

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

	stepUsersUnion := "users_union"
	qStmnt = joinWithComma(qStmnt, as(stepUsersUnion, unionStmnt))

	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, stepUsersUnion)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersSingleEventQuery
Group By: user_properties, event_properties

WITH
    step0 AS (
		SELECT DISTINCT ON(events.user_id, (to_timestamp(timestamp) AT TIME ZONE 'UTC')::date) events.user_id as event_user_id,
		(to_timestamp(timestamp) AT TIME ZONE 'UTC')::date as date FROM events
		WHERE events.project_id='1' AND timestamp>='1393632004' AND timestamp<='1396310325'
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='View Project')
		ORDER BY events.user_id, date, events.timestamp ASC
    )
	SELECT COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))), date
	FROM step0 LEFT JOIN users ON step0.event_user_id=users.id
	GROUP BY date, gk_0 order by gk_0;
*/
func buildUniqueUsersSingleEventQuery(projectId uint64, query Query) (string, []interface{}, error) {
	if len(query.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	steps := addEventFilterStepsForUniqueUsersQuery(projectId, &query, &qStmnt, &qParams)
	addUniqueUsersAggregationQuery(&query, &qStmnt, &qParams, steps[0])
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildEventsOccurrenceWithGivenEventQuery builds query for any given event and single event query,
Group by: user_properties, event_properties.

WITH
    e1 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id, events.properties->>'category' as group_prop1 FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
        AND user_properties.properties->>'gender'='M'
    ),
    e2 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id, events.properties->>'category' as group_prop1 FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='Fund Project')
        AND user_properties.properties->>'gender'='M'
    ),
    any_event AS (
        SELECT event_id, event_user_id, group_prop1 FROM e1 UNION ALL SELECT event_id, event_user_id, group_prop1 FROM e2
    )
    SELECT user_properties.properties->>'$region' as group_prop2, group_prop1, count(*) from any_event
    left join users on any_event.event_user_id=users.id
    left join user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id
    group by group_prop1, group_prop2 order by group_prop2;
*/
func buildEventsOccurrenceWithGivenEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
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
		refStepName = fmt.Sprintf("step%d", i)
		filters = append(filters, refStepName)
		addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, ewp, q.From, q.To, "",
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

	// select
	tSelect := joinWithComma(ugSelect, egKeys)
	tSelect = appendSelectTimestampColIfRequired(tSelect, isGroupByTimestamp)
	tSelect = joinWithComma(tSelect, fmt.Sprintf("COUNT(*) AS %s", AliasAggr)) // aggregator.

	termStmnt := "SELECT " + tSelect + " FROM " + refStepName
	// join lateset user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id" +
			" " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, isGroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, q.GroupByProperties, isGroupByTimestamp)

	qParams = append(qParams, ugSelectParams...)
	qStmnt = appendStatement(qStmnt, termStmnt)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
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

	if hasGroupEntity(q.GroupByProperties, PropertyEntityUser) {
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
buildUniqueUsersFunnelQuery

WITH
	step1 AS (
		SELECT DISTINCT ON(events.user_id) events.user_id as user_id, 1 as step1, events.timestamp as step1_timestamp,
        events.properties->>'presentation' as group_key_1 from events where events.project_id = 1
		and event_name_id IN (select id from event_names where project_id=1 and name = 'localhost:3000/#/core')
		and timestamp between 1393632004 and 1560231260 order by events.user_id, events.timestamp ASC
	),
	step2 as (
		SELECT DISTINCT ON(events.user_id) events.user_id as user_id, 1 as step2, events.timestamp as step2_timestamp,
        events.properties->>'presentation' as group_key_1 from events
		LEFT JOIN step1 on events.user_id=step1.user_id where events.project_id = 1
		and event_name_id IN (select id from event_names where project_id=1 and name = 'run_query')
		and timestamp between step1.step1_timestamp and 1560231260 order by events.user_id, events.timestamp ASC
	),
	step3 as (
		SELECT DISTINCT ON(events.user_id) events.user_id as user_id, 1 as step3, events.timestamp as step3_timestamp,
        events.properties->>'presentation' as group_key_1  from events
		LEFT JOIN step2 on events.user_id=step2.user_id where events.project_id = 1
		and event_name_id IN (select id from event_names where project_id=1 and name = 'localhost:3000/#/dashboard')
		and timestamp between step2.step2_timestamp and 1560231260 order by events.user_id, events.timestamp ASC
	),
	funnel as (
		SELECT DISTINCT ON(COALESCE(users.customer_user_id, step1.user_id))
        COALESCE(users.customer_user_id, step1.user_id) AS real_user_id,
        CASE WHEN user_properties.properties->>'$city' IS NULL THEN '$none'
        WHEN user_properties.properties->>'$city' = '' THEN '$none'
        ELSE user_properties.properties->>'$city' END AS group_key_0,
        step1, step2, step3, step2.group_key_1 as group_key_1 from step1
        LEFT JOIN users on step1.user_id=users.id
        LEFT JOIN user_properties on users.properties_id=user_properties.id
		LEFT JOIN step2 on step1.user_id=step2.user_id
		LEFT JOIN step3 on step2.user_id=step3.user_id
        -- ORDER BY real_user_id, step1, step2, step3 makes sure to count every step he did being anonymous, after COALESCE --
        ORDER BY real_user_id, step1, step2, step3
	)
	SELECT '$no_group' as group_key_0, '$no_group' as group_key_1, SUM(step1) as step1,
    SUM(step2) as step2, SUM(step3) as step3 from funnel
	UNION ALL
	SELECT * FROM (
		SELECT group_key_0, group_key_1, SUM(step1) as step1, SUM(step2) as step2, SUM(step3) as step3 from funnel group by group_key_0, group_key_1 order by step1 desc limit 10
	) AS group_funnel
*/
func buildUniqueUsersFunnelQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("invalid no.of events for funnel query")
	}

	var qStmnt string
	var qParams []interface{}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egSelect, egParams, _ := buildGroupKeys(eventGroupProps)

	var funnelSteps []string
	for i := range q.EventsWithProperties {
		var addParams []interface{}
		stepName := stepNameByIndex(i)
		// builds DISTINCT ON(events.user_id) events.user_id as user_id, 1 as step1, events.timestamp as step1_timestamp
		addSelect := "DISTINCT ON(events.user_id) events.user_id as user_id, 1 as " + stepName + ", events.timestamp as " + fmt.Sprintf("%s_timestamp", stepName)
		// adds event properties to select
		addSelect = joinWithComma(addSelect, egSelect)
		addParams = egParams

		var from int64
		var fromStr string
		// use actual from for first step,
		// and min timestamp from prev step
		// for next step's from.
		if i == 0 {
			from = q.From
		} else {
			// builds "step_2.step_2_timestamp"
			prevStepName := stepNameByIndex(i - 1)
			fromStr = fmt.Sprintf("%s.%s_timestamp", prevStepName, prevStepName)
		}

		var usersJoinStmnt string
		if i > 0 {
			// builds "LEFT JOIN step_0 on events.user_id=step_0.user_id"
			prevStepName := stepNameByIndex(i - 1)
			usersJoinStmnt = fmt.Sprintf("LEFT JOIN %s on events.user_id=%s.user_id", prevStepName, prevStepName)
		}

		// ordered to get first occurrence on DISTINCT ON(user_id).
		// occurrence: e0, e0, e1, e0, e2 funnel: e0 -> e1 -> e2.
		// Should consider first occurrence of e0 followed by e1 and e2.
		orderBy := "events.user_id, events.timestamp ASC"

		err := addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[i], from, q.To,
			fromStr, stepName, addSelect, addParams, usersJoinStmnt, "", orderBy)
		if err != nil {
			return "", nil, err
		}

		if i < len(q.EventsWithProperties)-1 {
			qStmnt = qStmnt + ", "
		}

		funnelSteps = append(funnelSteps, stepName)
	}

	var funnelStepsForSelect string
	for i, fs := range funnelSteps {
		// builds "step1, step2, step3"
		funnelStepsForSelect = funnelStepsForSelect + fs
		if i != len(funnelSteps)-1 {
			funnelStepsForSelect = funnelStepsForSelect + ","
		}
	}

	userGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityUser)
	ugSelect, ugParams, _ := buildGroupKeys(userGroupProps)

	// Join user properties if group by propertie given.
	// builds "LEFT JOIN users on step1.user_id=users.id LEFT JOIN user_properties on users.properties_id=user_properties.id"
	properitesJoinStmnt := fmt.Sprintf("LEFT JOIN users on %s.user_id=users.id", funnelSteps[0])
	if hasGroupEntity(q.GroupByProperties, PropertyEntityUser) {
		properitesJoinStmnt = properitesJoinStmnt + " " + "LEFT JOIN user_properties on users.properties_id=user_properties.id"
	}

	var stepsJoinStmnt string
	for i, fs := range funnelSteps {
		if i > 0 {
			// builds "LEFT JOIN step2 on step1.user_id=step2.user_id"
			stepsJoinStmnt = appendStatement(stepsJoinStmnt,
				fmt.Sprintf("LEFT JOIN %s ON %s.user_id=%s.user_id", fs, funnelSteps[i-1], fs))
		}
	}
	// adds group by event property with user selected event (step).
	eventGroupKeysWithStep := buildEventGroupKeysWithStep(eventGroupProps,
		q.EventsWithProperties)

	// builds ORDER BY real_user_id, step_1, step_2 ASC
	funnelOrderBy := "real_user_id"
	for _, fs := range funnelSteps {
		funnelOrderBy = joinWithComma(funnelOrderBy, fs)
	}

	funnelStepName := "funnel"
	coalesceRealUser := fmt.Sprintf("COALESCE(users.customer_user_id, %s.user_id)", funnelSteps[0])
	coalesceUserSelect := fmt.Sprintf("DISTINCT ON(%s) %s AS real_user_id", coalesceRealUser, coalesceRealUser)
	funnelStmnt := "SELECT" + " " + joinWithComma(coalesceUserSelect, ugSelect, eventGroupKeysWithStep, funnelStepsForSelect) + " " + "FROM" + " " +
		funnelSteps[0] + " " + properitesJoinStmnt + " " + stepsJoinStmnt + " " + "ORDER BY" + " " + funnelOrderBy
	funnelStmnt = as(funnelStepName, funnelStmnt)
	qStmnt = joinWithComma(qStmnt, funnelStmnt)
	qParams = append(qParams, ugParams...)

	// builds "SUM(step1) AS step1, SUM(step1) AS step2",
	var rawCountSelect string
	for _, fs := range funnelSteps {
		rawCountSelect = joinWithComma(rawCountSelect, fmt.Sprintf("SUM(%s) AS %s", fs, fs))
	}

	var termStmnt string
	if len(q.GroupByProperties) == 0 {
		termStmnt = "SELECT" + " " + rawCountSelect + " " + "FROM" + " " + funnelStepName
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
			" " + "FROM" + " " + funnelStepName

		_, _, groupKeys := buildGroupKeys(q.GroupByProperties)
		limitedGroupBySelect := "SELECT" + " " + joinWithComma(groupKeys, rawCountSelect) + " " +
			"FROM" + " " + funnelStepName + " " + "GROUP BY" + " " + groupKeys + " " +
			// limits result by count of step_1 to ResultsLimit.
			fmt.Sprintf("ORDER BY %s DESC LIMIT %d", funnelSteps[0], ResultsLimit)

		// wrapped with select to apply limits only to grouped select rows.
		groupBySelect := fmt.Sprintf("SELECT * FROM ( %s ) AS group_funnel", limitedGroupBySelect)

		termStmnt = noGroupSelect + " " + "UNION ALL" + " " + groupBySelect
	}

	qStmnt = appendStatement(qStmnt, termStmnt)
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

// TranslateGroupKeysIntoColumnNames - Replaces groupKeys on result
// headers with real column names.
func translateGroupKeysIntoColumnNames(result *QueryResult,
	groupProps []QueryGroupByProperty) error {

	rcols := make([]string, 0, 0)

	indexLookup := make(map[int]QueryGroupByProperty, 0)
	for _, v := range groupProps {
		indexLookup[v.Index] = v
	}

	for i := range result.Headers {
		if strings.HasPrefix(result.Headers[i], GroupKeyPrefix) {
			gIndexStr := strings.TrimPrefix(result.Headers[i], GroupKeyPrefix)
			if gIndex, err := strconv.Atoi(gIndexStr); err != nil {
				log.WithField("group_key", gIndexStr).Error(
					"Invalid group key index. Failed translating group key to real name.")
				return errors.New("invalid group key index")
			} else {
				rcols = append(rcols, indexLookup[gIndex].Property)
			}
		} else {
			rcols = append(rcols, result.Headers[i])
		}
	}

	result.Headers = rcols
	return nil
}

// Preserves order of group by properties on query as
// given by the user.
func addIndexToGroupByProperties(query *Query) {
	for i := range query.GroupByProperties {
		query.GroupByProperties[i].Index = i
	}
}

func getGroupKeyIndexesForSlicing(cols []string) (int, int, error) {
	start := -1
	end := -1

	index := 0
	for _, col := range cols {
		if strings.HasPrefix(col, GroupKeyPrefix) {
			if start == -1 {
				start = index
			} else {
				end = index
			}
		}
		index++
	}

	// single element.
	if start > -1 && end == -1 {
		end = start
	}

	if start == -1 {
		return start, end, errors.New("no group keys found")
	}

	// end index + 1 reads till end index on slice.
	end = end + 1

	return start, end, nil
}

// Creates encoded key by cols position and value.
func getEncodedKeyForCols(cols []interface{}) string {
	var key string
	for i, col := range cols {
		enc := fmt.Sprintf("c%d:%s", i, col)
		if i == 0 {
			key = enc
			continue
		}
		key = key + "_" + enc
	}
	return key
}

func isValidGroupByTimestamp(groupByTimestamp string) bool {
	if groupByTimestamp == "" {
		return true
	}

	for _, gbtType := range groupByTimestampTypes {
		if gbtType == groupByTimestamp {
			return true
		}
	}

	return false
}

// IsValidQuery Validates and returns errMsg which is used as response.
func IsValidQuery(query *Query) (bool, string) {
	if query.Type != QueryTypeEventsOccurrence &&
		query.Type != QueryTypeUniqueUsers {
		return false, "Invalid query type given"
	}

	if query.EventsCondition != EventCondAllGivenEvent &&
		query.EventsCondition != EventCondAnyGivenEvent {
		return false, "Invalid events condition given"
	}

	if len(query.EventsWithProperties) == 0 {
		return false, "No events to process"
	}

	if query.From == 0 || query.To == 0 {
		return false, "Invalid query time range"
	}

	if !isValidGroupByTimestamp(query.GetGroupByTimestamp()) {
		return false, "Invalid group by timestamp"
	}

	return true, ""
}

// BuildInsightsQuery - Dispatches corresponding build method for insights.
func BuildInsightsQuery(projectId uint64, query Query) (string, []interface{}, error) {
	addIndexToGroupByProperties(&query)

	if query.Type == QueryTypeEventsOccurrence {
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

		return buildUniqueUsersWithAllGivenEventsQuery(projectId, query)
	}

	return "", nil, errors.New("invalid query")
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

func getConversionPercentage(prevCount float64, curCount float64) float64 {
	if prevCount == 0 {
		return float64(0)
	}

	return (curCount / prevCount) * 100
}

func addStepConversionPercentageToFunnel(result *QueryResult) error {
	if len(result.Rows) == 0 {
		return errors.New("invalid funnel result")
	}

	stepIndexes := make([]int, 0, 0)
	for i, header := range result.Headers {
		if strings.HasPrefix(header, StepPrefix) {
			stepIndexes = append(stepIndexes, i)
		}
	}

	for ri := range result.Rows {
		// add step conversions.
		conversions := make([]interface{}, 0, 0)
		for _, ci := range stepIndexes {
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

			conversion := getConversionPercentage(prevCount, curCount)
			conversions = append(conversions, fmt.Sprintf("%0.0f", conversion))
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

		conversion := getConversionPercentage(firstStepCount, lastStepCount)
		conversions = append(conversions, fmt.Sprintf("%0.0f", conversion))
		result.Rows[ri] = append(result.Rows[ri], conversions...)
	}

	// add conversion headers.
	conversionStrs := make([]string, 0, 0)
	for _, si := range stepIndexes {
		if si > stepIndexes[0] {
			conversionStrs = append(conversionStrs, fmt.Sprintf("%s%s_%s",
				FunnelConversionPrefix, result.Headers[si-1], result.Headers[si]))
		}
	}
	conversionStrs = append(conversionStrs, fmt.Sprintf("%s%s", FunnelConversionPrefix, "overall"))
	result.Headers = append(result.Headers, conversionStrs...)

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

func GetTimstampAndAggregateIndexOnQueryResult(cols []string) (int, int, error) {
	timeIndex := -1
	aggrIndex := -1

	for i, c := range cols {
		if c == AliasDateTime {
			timeIndex = i
		}

		if c == AliasAggr {
			aggrIndex = i
		}
	}

	var err error
	if timeIndex == -1 {
		err = errors.New("invalid result without timestamp")
	}
	if aggrIndex == -1 {
		err = errors.New("invalid result without aggregate")
	}

	return aggrIndex, timeIndex, err
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
	// uses timstamp string for comparison.
	for _, ts := range timestamps {
		if row, exists := rowsByTimestamp[U.GetTimestampAsStrWithTimezone(ts, query.Timezone)]; exists {
			// overrides timestamp with user timezone as sql results doesn't
			// return timezone used to query.
			row[timestampIndex] = ts
			filledResult = append(filledResult, row)
		} else {
			newRow := make([]interface{}, 2, 2)
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

func sanitizeGroupByTimestampResult(result *QueryResult, query *Query) error {
	aggrIndex, timeIndex, err := GetTimstampAndAggregateIndexOnQueryResult(result.Headers)
	if err != nil {
		return err
	}

	// Todo: Supports only date as timestamp, add support for hour and month.
	if len(query.GroupByProperties) == 0 {
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

// Converts DB results into plottable query results.
func SanitizeQueryResult(result *QueryResult, query *Query) error {
	if query.GetGroupByTimestamp() != "" {
		return sanitizeGroupByTimestampResult(result, query)
	}

	// Replace group keys with real column names. should be last step.
	// of sanitization.
	return translateGroupKeysIntoColumnNames(result, query.GroupByProperties)
}

func ExecQuery(stmnt string, params []interface{}) (*QueryResult, error) {
	db := C.GetServices().Db

	rows, err := db.Raw(stmnt, params...).Rows()
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}

	result := &QueryResult{Headers: resultHeaders, Rows: resultRows}
	return result, nil
}

func addMetaToQueryResult(result *QueryResult, query Query) {
	result.Meta.Query = query
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

	addMetaToQueryResult(result, query)

	return result, http.StatusOK, "Successfully executed query"
}

func isValidFunnelQuery(query *Query) bool {
	return len(query.EventsWithProperties) <= 4
}

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

	addMetaToQueryResult(result, query)

	return result, http.StatusOK, "Successfully executed query"
}

func Analyze(projectId uint64, query Query) (*QueryResult, int, string) {
	if valid, errMsg := IsValidQuery(&query); !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	if query.Class == QueryClassFunnel {
		return RunFunnelQuery(projectId, query)
	}

	return RunInsightsQuery(projectId, query)
}
