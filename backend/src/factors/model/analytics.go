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

	"github.com/jinzhu/gorm/dialects/postgres"

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
	EventName      string `json:"ena"`
	EventNameIndex int    `json:"eni"`
}

type QueryEventWithProperties struct {
	Name       string          `json:"na"`
	Properties []QueryProperty `json:"pr"`
}

// BaseQuery Base query interface for all query classes.
type BaseQuery interface {
	GetClass() string
	GetQueryDateRange() (int64, int64)
	SetQueryDateRange(from, to int64)
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
	OverridePeriod    bool  `json:"ovp"`
	SessionStartEvent int64 `json:"sse"`
	SessionEndEvent   int64 `json:"see"`
}

type QueryResultMeta struct {
	Query    Query  `json:"query"`
	Currency string `json:"currency"` //Currency field is used for Attribution query response.
}

func (q *Query) GetClass() string {
	return q.Class
}

func (q *Query) GetQueryDateRange() (from, to int64) {
	return q.From, q.To
}

func (q *Query) SetQueryDateRange(from, to int64) {
	q.From, q.To = from, to
}

type QueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	// Todo(Dinesh): Change it to interface and use
	// the query result struct across different queries.
	Meta QueryResultMeta `json:"meta"`
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

	QueryClassInsights    = "insights"
	QueryClassFunnel      = "funnel"
	QueryClassChannel     = "channel"
	QueryClassAttribution = "attribution"
	QueryClassWeb         = "web"

	QueryTypeEventsOccurrence = "events_occurrence"
	QueryTypeUniqueUsers      = "unique_users"

	ErrUnsupportedGroupByEventPropertyOnUserQuery = "group by event property is not supported for user query"
	ErrMsgQueryProcessingFailure                  = "Failed processing query"
	ErrMsgMaxFunnelStepsExceeded                  = "Max funnel steps exceeded"

	SelectDefaultEventFilter              = "events.id as event_id, events.user_id as event_user_id"
	SelectDefaultEventFilterWithDistinct  = "DISTINCT(events.id) as event_id, events.user_id as event_user_id"
	SelectDefaultEventFilterByAlias       = "event_id, event_user_id, event_name"
	SelectCoalesceCustomerUserIDAndUserID = "COALESCE(users.customer_user_id, event_user_id)"

	GroupKeyPrefix  = "_group_key_"
	AliasDateTime   = "datetime"
	AliasAggr       = "count"
	DefaultTimezone = "UTC"
	ResultsLimit    = 100
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

// UserPropertyGroupByPresent Sent from frontend for breakdown on latest user property.
const UserPropertyGroupByPresent string = "$present"

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
	eventLevelGroupBys, _ := separateEventLevelGroupBys(groupProps)

	for _, gp := range eventLevelGroupBys {
		groupKey := fmt.Sprintf("%s.%s", stepNameByIndex(gp.EventNameIndex-1),
			groupKeyByIndex(gp.Index))
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

func removeEventSpecificUserGroupBys(groupBys []QueryGroupByProperty) []QueryGroupByProperty {
	filteredProps := make([]QueryGroupByProperty, 0)
	for _, prop := range groupBys {
		if isEventLevelGroupBy(prop) {
			// For $present, event name index is not set and is default 0.
			continue
		}
		filteredProps = append(filteredProps, prop)
	}
	return filteredProps
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

func appendOrderByAggr(qStmnt string) string {
	return fmt.Sprintf("%s ORDER BY %s DESC", qStmnt, AliasAggr)
}
func appendOrderByEventNameAndAggr(qStmnt string) string {
	return fmt.Sprintf("%s ORDER BY event_name, %s DESC", qStmnt, AliasAggr)
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

func addEventFilterStepsForUniqueUsersQuery(projectID uint64, q *Query,
	qStmnt *string, qParams *[]interface{}) []string {

	var commonSelect string
	var stepOrderBy string

	if q.GetGroupByTimestamp() != "" {
		selectTimestamp := getSelectTimestampByType(q.GetGroupByTimestamp(), q.Timezone)
		// select and order by with datetime.
		commonSelect = fmt.Sprintf("DISTINCT ON(COALESCE(users.customer_user_id,events.user_id), %s)", selectTimestamp) +
			fmt.Sprintf(" COALESCE(users.customer_user_id,events.user_id) as coal_user_id, %s as %s,", selectTimestamp, AliasDateTime) +
			" events.user_id as event_user_id"

		stepOrderBy = fmt.Sprintf("coal_user_id, %s, events.timestamp ASC", AliasDateTime)
	} else {
		// default select.
		commonSelect = "DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id)" +
			" as coal_user_id, events.user_id as event_user_id"
	}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egAnySelect, egAnyParams, _ := buildGroupKeys(eventGroupProps)

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		if stepOrderBy == "" {
			// Using first occurred event_properites after distinct on user_id.
			stepOrderBy = "coal_user_id, events.timestamp ASC"
		}
	}

	steps := make([]string, 0, 0)
	for i, ewp := range q.EventsWithProperties {
		refStepName := stepNameByIndex(i)
		steps = append(steps, refStepName)

		var stepSelect string
		var stepParams []interface{}
		var groupByUserProperties bool
		if q.EventsCondition == EventCondAllGivenEvent {
			stepSelect, stepParams, _, groupByUserProperties = buildGroupKeyForStep(
				&q.EventsWithProperties[i], q.GroupByProperties, i+1)
			stepSelect = joinWithComma(commonSelect, stepSelect)
		} else {
			stepSelect = joinWithComma(commonSelect, egAnySelect)
			stepParams = egAnyParams
		}

		addJoinStmnt := "JOIN users ON events.user_id=users.id"
		if groupByUserProperties {
			addJoinStmnt += " JOIN user_properties on events.user_properties_id=user_properties.id"
		}
		addFilterEventsWithPropsQuery(projectID, qStmnt, qParams, ewp, q.From, q.To,
			"", refStepName, stepSelect, stepParams, addJoinStmnt, "", stepOrderBy)

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
	eventLevelGroupBys, otherGroupBys := separateEventLevelGroupBys(query.GroupByProperties)
	var egKeys string
	if query.EventsCondition == EventCondAllGivenEvent {
		_, _, egKeys = buildGroupKeys(eventLevelGroupBys)
	} else {
		eventGroupProps := filterGroupPropsByType(otherGroupBys, PropertyEntityEvent)
		_, _, egKeys = buildGroupKeys(eventGroupProps)
	}

	// select
	userGroupProps := filterGroupPropsByType(otherGroupBys, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	*qParams = append(*qParams, ugSelectParams...)
	// order of group keys changes here if users and event
	// group by used together, but translated correctly.
	termSelect := joinWithComma(ugSelect, egKeys)

	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	termSelect = appendSelectTimestampColIfRequired(termSelect, isGroupByTimestamp)
	termSelect = joinWithComma(termSelect, fmt.Sprintf("COUNT(DISTINCT(%s.event_user_id)) AS %s", refStep, AliasAggr))

	// group by
	termStmnt := "SELECT " + termSelect + " FROM " + refStep
	// join latest user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStep + ".event_user_id=users.id"
		termStmnt = termStmnt + " " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	_, _, groupKeys := buildGroupKeys(query.GroupByProperties)
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, isGroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, query.GroupByProperties, isGroupByTimestamp)

	*qStmnt = appendStatement(*qStmnt, termStmnt)
}

func separateEventLevelGroupBys(allGroupBys []QueryGroupByProperty) (
	eventLevelGroupBys, otherGroupBys []QueryGroupByProperty) {
	for _, groupby := range allGroupBys {
		if isEventLevelGroupBy(groupby) {
			eventLevelGroupBys = append(eventLevelGroupBys, groupby)
		} else {
			// This will also have $present event.
			otherGroupBys = append(otherGroupBys, groupby)
		}
	}
	return
}

// isEventLevelGroupBy Checks if the groupBy is for a particular event in query.ewp.
func isEventLevelGroupBy(groupBy QueryGroupByProperty) bool {
	return groupBy.EventName != "" && groupBy.EventNameIndex != 0
}

/*
buildUniqueUsersWithAllGivenEventsQuery builds a query like below,
Group by: user_properties, event_properties

Example: Query without date and with group by properties.

Sample query with ewp:
	View Project
	$session
gbp:
	user_property -> $present -> $city
	event_property -> 2. $session -> $browser
	user_property -> 2. $session -> $platform
	user_property -> 1. View Project -> $user_agent

WITH
step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step_0 AS (SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
events.user_id as event_user_id, CASE WHEN user_properties.properties->>'$user_agent' IS NULL THEN '$none'
WHEN user_properties.properties->>'$user_agent' = '' THEN '$none' ELSE user_properties.properties->>'$user_agent' END AS _group_key_3
FROM events JOIN users ON events.user_id=users.id JOIN user_properties on events.user_properties_id=user_properties.id
WHERE events.project_id='3' AND timestamp>='1595788200' AND timestamp<='1596392999' AND events.event_name_id IN
(SELECT id FROM step_0_names WHERE project_id='3' AND name='View Project') ORDER BY coal_user_id, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='$session'),

step_1 AS (SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
events.user_id as event_user_id, CASE WHEN events.properties->>'$browser' IS NULL THEN '$none' WHEN events.properties->>'$browser' = '' THEN '$none'
ELSE events.properties->>'$browser' END AS _group_key_1, CASE WHEN user_properties.properties->>'$platform' IS NULL THEN '$none'
WHEN user_properties.properties->>'$platform' = '' THEN '$none' ELSE user_properties.properties->>'$platform' END AS _group_key_2 FROM events
JOIN users ON events.user_id=users.id JOIN user_properties on events.user_properties_id=user_properties.id WHERE events.project_id='3'
AND timestamp>='1595788200' AND timestamp<='1596392999' AND events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='3' AND name='$session')
ORDER BY coal_user_id, events.timestamp ASC),

users_intersect AS (SELECT step_0.event_user_id as event_user_id, step_1._group_key_1, step_1._group_key_2, step_0._group_key_3 FROM step_0
JOIN step_1 ON step_1.coal_user_id = step_0.coal_user_id)

SELECT CASE WHEN user_properties.properties->>'$city' IS NULL THEN '$none'
WHEN user_properties.properties->>'$city' = '' THEN '$none' ELSE user_properties.properties->>'$city' END AS _group_key_0,
_group_key_1, _group_key_2, _group_key_3, COUNT(DISTINCT(users_intersect.event_user_id)) AS count FROM users_intersect
LEFT JOIN users ON users_intersect.event_user_id=users.id LEFT JOIN user_properties ON users.id=user_properties.user_id
AND user_properties.id=users.properties_id GROUP BY _group_key_0, _group_key_1, _group_key_2, _group_key_3 ORDER BY count DESC LIMIT 100000

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

Sample query with ewp:
	View Project
	$session
gbp:
	event_property -> $browser
	user_property -> $city

WITH
step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step_0 AS (SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
events.user_id as event_user_id, CASE WHEN events.properties->>'$browser' IS NULL THEN '$none' WHEN events.properties->>'$browser' = '' THEN '$none'
ELSE events.properties->>'$browser' END AS _group_key_0 FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='3'
AND timestamp>='1595788200' AND timestamp<='1596392999' AND events.event_name_id IN (SELECT id FROM step_0_names WHERE project_id='3'
AND name='View Project') ORDER BY coal_user_id, events.timestamp ASC),

step_1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name=''),

step_1 AS (SELECT DISTINCT ON(COALESCE(users.customer_user_id,events.user_id)) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
events.user_id as event_user_id, CASE WHEN events.properties->>'$browser' IS NULL THEN '$none' WHEN events.properties->>'$browser' = '' THEN '$none'
ELSE events.properties->>'$browser' END AS _group_key_0 FROM events JOIN users ON events.user_id=users.id WHERE events.project_id='3'
AND timestamp>='1595788200' AND timestamp<='1596392999' AND events.event_name_id IN (SELECT id FROM step_1_names WHERE project_id='3'
AND name='') ORDER BY coal_user_id, events.timestamp ASC),

users_union AS (SELECT step_0.event_user_id as event_user_id, _group_key_0 FROM step_0 UNION ALL
SELECT step_1.event_user_id as event_user_id, _group_key_0 FROM step_1)

SELECT CASE WHEN user_properties.properties->>'$city' IS NULL THEN '$none' WHEN user_properties.properties->>'$city' = '' THEN '$none'
ELSE user_properties.properties->>'$city' END AS _group_key_1, _group_key_0, COUNT(DISTINCT(users_union.event_user_id)) AS count
FROM users_union LEFT JOIN users ON users_union.event_user_id=users.id LEFT JOIN user_properties ON users.id=user_properties.user_id
AND user_properties.id=users.properties_id GROUP BY _group_key_0, _group_key_1 ORDER BY count DESC LIMIT 100000

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

Sample query for ewp:
	View Project
	$session
gbp:
	event_property -> $browser
	user_property -> $city

WITH
step0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='View Project'),

step0 AS (SELECT events.id as event_id, events.user_id as event_user_id, CASE WHEN events.properties->>'$browser' IS NULL THEN
'$none' WHEN events.properties->>'$browser' = '' THEN '$none' ELSE events.properties->>'$browser' END AS _group_key_0, 'View Project'::text
AS event_name  FROM events  WHERE events.project_id='3' AND timestamp>='1595788200' AND timestamp<='1596392999' AND
events.event_name_id IN (SELECT id FROM step0_names WHERE project_id='3' AND name='View Project')),

step1_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='3' AND name='$session'),

step1 AS (SELECT events.id as event_id, events.user_id as event_user_id, CASE WHEN events.properties->>'$session' IS NULL THEN '$none'
WHEN events.properties->>'$session' = '' THEN '$none' ELSE events.properties->>'$session' END AS _group_key_0, '$session'::text AS event_name FROM events
WHERE events.project_id='3' AND timestamp>='1595788200' AND timestamp<='1596392999' AND
events.event_name_id IN (SELECT id FROM step1_names WHERE project_id='3' AND name='')),

any_event AS ( SELECT event_id, event_user_id, event_name, _group_key_0 FROM step0
UNION ALL SELECT event_id, event_user_id, event_name, _group_key_0 FROM step1)

SELECT event_name, CASE WHEN user_properties.properties->>'$city' IS NULL THEN '$none' WHEN user_properties.properties->>'$city' = '' THEN '$none'
ELSE user_properties.properties->>'$city' END AS _group_key_1, _group_key_0, COUNT(*) AS count FROM any_event LEFT JOIN users
ON any_event.event_user_id=users.id LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id
GROUP BY event_name, _group_key_0, _group_key_1 ORDER BY event_name, count DESC LIMIT 100000
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
		eventNameSelect := "'" + ewp.Name + "'" + "::text" + " AS event_name "
		filterSelect := joinWithComma(filterSelect, eventNameSelect)
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

	eventNameSelect := "event_name"
	groupKeys = joinWithComma(eventNameSelect, groupKeys)

	// select
	tSelect := joinWithComma(eventNameSelect, ugSelect, egKeys)
	tSelect = appendSelectTimestampColIfRequired(tSelect, isGroupByTimestamp)
	tSelect = joinWithComma(tSelect, fmt.Sprintf("COUNT(*) AS %s", AliasAggr)) // aggregator.

	termStmnt := "SELECT " + tSelect + " FROM " + refStepName
	// join lateset user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id" +
			" " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, isGroupByTimestamp, groupKeys)
	if len(q.EventsWithProperties) > 1 {
		termStmnt = appendOrderByEventNameAndAggr(termStmnt)
	} else {
		termStmnt = appendOrderByAggr(termStmnt)
	}
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
		if groupByUserProperties {
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

		_, _, groupKeys := buildGroupKeys(q.GroupByProperties)
		limitedGroupBySelect := "SELECT" + " " + joinWithComma(groupKeys, rawCountSelect) + " " +
			"FROM" + " " + stepFunnelName + " " + "GROUP BY" + " " + groupKeys + " " +
			// order and limit by last step of funnel.
			fmt.Sprintf("ORDER BY %s DESC LIMIT %d", funnelCountAliases[0], ResultsLimit)

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
		if strings.HasPrefix(col, GroupKeyPrefix) || col == "event_name" {
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

func DecodeQueryForClass(queryJSON postgres.Jsonb, queryClass string) (BaseQuery, error) {
	var baseQuery BaseQuery
	var err error
	switch queryClass {
	case QueryClassFunnel, QueryClassInsights:
		var query Query
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassAttribution:
		var query AttributionQueryUnit
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassChannel:
		var query ChannelQueryUnit
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	default:
		return baseQuery, fmt.Errorf("query class %s not supported", queryClass)
	}

	return baseQuery, err
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
