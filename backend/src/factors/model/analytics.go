package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type QueryProperty struct {
	// Entity: user or event.
	Entity string `json:"en"`
	// Type: categorical or numerical
	Type     string `json:"ty"`
	Property string `json:"pr"`
	Operator string `json:"op"`
	Value    string `json:"va"`
}

type QueryGroupByProperty struct {
	// Entity: user or event.
	Entity   string `json:"en"`
	Property string `json:"pr"`
	Index    int    `json:"in"`
}

type QueryEventWithProperties struct {
	Name       string          `json:"na"`
	Properties []QueryProperty `json:"pr"`
}

type Query struct {
	Type                 string                     `json:"ty"`
	EventsCondition      string                     `json:"ec"` // all or any
	EventsWithProperties []QueryEventWithProperties `json:"ewp"`
	GroupByProperties    []QueryGroupByProperty     `json:"gbp"`
	GroupByTimestamp     bool                       `json:"gbt"`
	Timezone             string                     `json:"tz"`
	OverridePeriod       bool                       `json:"ovp"`
	From                 int64                      `json:"fr"`
	To                   int64                      `json:"to"`
}

const (
	PropertyEntityUser  = "user"
	PropertyEntityEvent = "event"

	PropertyValueNone = "$none"

	EventCondAnyGivenEvent = "any_given_event"
	EventCondAllGivenEvent = "all_given_event"

	QueryTypeEventsOccurrence = "events_occurrence"
	QueryTypeUniqueUsers      = "unique_users"

	ErrUnsupportedGroupByEventPropertyOnUserQuery = "group by event property is not supported for user query"
	ErrMsgQueryProcessingFailure                  = "Failed processing query"

	SelectDefaultEventFilter              = "DISTINCT(events.id) as event_id, events.user_id as event_user_id"
	SelectDefaultEventFilterByAlias       = "event_id, event_user_id"
	SelectCoalesceCustomerUserIDAndUserID = "COALESCE(users.customer_user_id, event_user_id)"

	GroupKeyPrefix  = "_group_key_"
	AliasDate       = "date"
	AliasAggr       = "count"
	DefaultTimezone = "UTC"
	ResultsLimit    = 20
	MaxResultsLimit = 100000
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

func buildWhereFromProperties(mergeCond string,
	properties []QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	pLen := len(properties)
	if pLen == 0 {
		return rStmnt, rParams, nil
	}

	rParams = make([]interface{}, 0, 0)
	for i, p := range properties {
		propertyEntity := getPropertyEntityField(p.Entity)
		propertyOp := getOp(p.Operator)

		if p.Value != PropertyValueNone {
			pStmnt := fmt.Sprintf("%s->>?%s?", propertyEntity, propertyOp)
			if i == 0 {
				rStmnt = pStmnt
			} else {
				rStmnt = fmt.Sprintf("%s %s %s", rStmnt, mergeCond, pStmnt)
			}

			rParams = append(rParams, p.Property, p.Value)
			continue
		}

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
			rStmnt = fmt.Sprintf("%s %s %s", rStmnt, mergeCond, whereCond)
		}

		rParams = append(rParams, p.Property, p.Property)
	}

	return rStmnt, rParams, nil
}

// Alias for group by properties gk_1, gk_2.
func groupKeyByIndex(i int) string {
	return fmt.Sprintf("%s%d", GroupKeyPrefix, i)
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

// Adds initial events with props filtering query.
func addFilterEventsWithPropsQuery(projectId uint64, qStmnt *string,
	qParams *[]interface{}, qep QueryEventWithProperties, from int64, to int64,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string) error {

	if from == 0 || to == 0 {
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
	rStmnt = appendStatement(
		rStmnt,
		"WHERE events.project_id=? AND timestamp>=? AND timestamp<=?"+
			" "+"AND events.event_name_id IN ( SELECT id FROM event_names WHERE project_id=? AND name=? )",
	)

	if addSelecStmnt != "" && addSelectParams != nil {
		*qParams = append(*qParams, addSelectParams...)
	}
	*qParams = append(*qParams, projectId, from, to, projectId, qep.Name)

	// mergeCond for whereProperties can also be 'OR'.
	wStmnt, wParams, err := buildWhereFromProperties("AND", qep.Properties)
	if err != nil {
		return err
	}

	if wStmnt != "" {
		rStmnt = rStmnt + " AND " + fmt.Sprintf("( %s )", wStmnt)
		*qParams = append(*qParams, wParams...)
	}

	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
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

// Adds additional join to the events filter,  for considering users only from prev
// step and with diff step names.
func addFilterEventsWithPropsForUsersQuery(projectId uint64, query Query,
	qStmnt *string, qParams *[]interface{}) string {

	var rStmnt string
	var rParams []interface{}

	stepName := ""
	lastEwpIndex := len(query.EventsWithProperties) - 1
	for i, ewp := range query.EventsWithProperties {
		// initially prevStepName will be empty.
		prevStepName := stepName

		if stepName == "" {
			stepName = fmt.Sprintf("step%d", i)
		} else {
			stepName = fmt.Sprintf("%s_step%d", stepName, i)
		}

		var usersFromLastStep string
		if i > 0 {
			usersFromLastStep = "JOIN " + prevStepName + " ON events.user_id=" + prevStepName + ".event_user_id"
		}

		filterSelect := SelectDefaultEventFilter
		filterSelect = appendSelectTimestampIfRequired(filterSelect, query.Timezone, query.GroupByTimestamp)
		addFilterEventsWithPropsQuery(projectId, &rStmnt, &rParams, ewp, query.From, query.To,
			stepName, filterSelect, nil, usersFromLastStep)

		if i != lastEwpIndex {
			rStmnt = rStmnt + ", "
		}
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)
	*qParams = append(*qParams, rParams...)

	// last step name for ref.
	return stepName
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

	return joinWithComma(stmnt, AliasDate)
}

func getSelectTimestamp(timezone string) string {
	var selectTz string

	if timezone == "" {
		selectTz = DefaultTimezone
	} else {
		selectTz = timezone
	}

	return fmt.Sprintf("(to_timestamp(timestamp) AT TIME ZONE '%s')::date as %s",
		selectTz, AliasDate)
}

func appendSelectTimestampIfRequired(stmnt string, timezone string, isRequired bool) string {
	if !isRequired {
		return stmnt
	}

	return joinWithComma(stmnt, getSelectTimestamp(timezone))
}

func appendGroupByTimestampIfRequired(qStmnt string, isRequired bool, groupKeys ...string) string {
	// Added groups with timestamp.
	groups := make([]string, 0, 0)
	if isRequired {
		groups = append(groups, AliasDate)
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

func buildAnyGivenEventFilterQuery(projectId uint64, q Query) (string, []interface{}, string, string) {
	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	groupSelect, groupParams, groupKeys := buildGroupKeys(eventGroupProps)

	var filterSelect string
	if q.Type == QueryTypeUniqueUsers {
		filterSelect = "DISTINCT(events.user_id) as event_user_id"
	} else {
		filterSelect = SelectDefaultEventFilter
	}

	filterSelect = joinWithComma(filterSelect, groupSelect)
	filterSelect = appendSelectTimestampIfRequired(filterSelect, q.Timezone, q.GroupByTimestamp)

	refStepName := ""
	filters := make([]string, 0)
	for i, ewp := range q.EventsWithProperties {
		refStepName = fmt.Sprintf("step%d", i)
		filters = append(filters, refStepName)
		addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, ewp, q.From, q.To,
			refStepName, filterSelect, groupParams, "")
		if len(q.EventsWithProperties) > 1 {
			qStmnt = qStmnt + ", "
		}
	}

	// union.
	if len(filters) > 1 {
		var unionType string
		if q.Type == QueryTypeUniqueUsers {
			// event_user_id is not unique after filtering.
			unionType = "UNION"
		} else {
			// event_id is already unique.
			unionType = "UNION ALL"
		}

		unionStepName := "any_event"
		unionStmnt := ""
		for _, filter := range filters {
			if unionStmnt != "" {
				unionStmnt = appendStatement(unionStmnt, unionType)
			}

			var qSelect string
			if q.Type == QueryTypeUniqueUsers {
				qSelect = "event_user_id"
			} else {
				qSelect = SelectDefaultEventFilterByAlias
			}

			qSelect = appendSelectTimestampColIfRequired(qSelect, q.GroupByTimestamp)
			qSelect = joinWithComma(qSelect, groupKeys)
			unionStmnt = unionStmnt + " SELECT " + qSelect + " FROM " + filter
		}
		unionStmnt = as(unionStepName, unionStmnt)
		qStmnt = appendStatement(qStmnt, unionStmnt)

		refStepName = unionStepName
	}

	return qStmnt, qParams, groupKeys, refStepName
}

/*
buildUniqueUsersWithAllGivenEventsQuery builds a query like below,
Group by: user_properties.

WITH
    e1 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=1 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='View Project')
        AND events.properties->>'category'='Sports' AND user_properties.properties->>'gender'='M'
    ),
    e1_e2 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id FROM events
        JOIN e1 ON events.user_id=e1.event_user_id
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=1 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='1' AND name='Fund Project')
        AND events.properties->>'category'='Sports' AND user_properties.properties->>'gender'='M'
    )
    SELECT user_properties.properties->>'$region' as group_prop1, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id))) from e1_e2
    left join users on e1_e2.event_user_id=users.id
    left join user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id
    GROUP BY group_prop1 order by count desc;
*/
func buildUniqueUsersWithAllGivenEventsQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		return "", nil, errors.New(ErrUnsupportedGroupByEventPropertyOnUserQuery)
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	refStepName := addFilterEventsWithPropsForUsersQuery(projectId, q, &qStmnt, &qParams)

	// select
	var termSelect string
	termSelect = appendSelectTimestampColIfRequired(termSelect, q.GroupByTimestamp)
	termSelect = joinWithComma(termSelect, fmt.Sprintf("COUNT(DISTINCT(%s)) AS %s", SelectCoalesceCustomerUserIDAndUserID, AliasAggr))

	// group by
	var termStmnt string
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName,
		"", &termStmnt, &qParams, termSelect)
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, q.GroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, q.GroupByProperties, q.GroupByTimestamp)

	qStmnt = appendStatement(qStmnt, termStmnt)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersWithAnyGivenEventsQuery
Group By: user_properties

WITH
	e1 AS (
        SELECT  DISTINCT(events.user_id) as event_user_id FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
        AND events.properties->>'category'='Sports'
    ),
    e2 AS (
        SELECT DISTINCT(events.user_id) as event_user_id FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='Fund Project')
        AND events.properties->>'category'='Sports'
    ),
	any_event AS (
        SELECT event_user_id FROM e1 UNION SELECT event_user_id FROM e2
	)
	SELECT user_properties.properties->>'gender' as gk_0, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id)))
	FROM any_event LEFT JOIN users ON any_event.event_user_id=users.id
	LEFT JOIN user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id GROUP BY gk_0 order by gk_0;

*/
func buildUniqueUsersWithAnyGivenEventsQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		return "", nil, errors.New(ErrUnsupportedGroupByEventPropertyOnUserQuery)
	}

	// init with any given event filter query.
	qStmnt, qParams, _, refStepName := buildAnyGivenEventFilterQuery(projectId, q)

	// select
	var termSelect string
	termSelect = appendSelectTimestampColIfRequired(termSelect, q.GroupByTimestamp)
	termSelect = joinWithComma(termSelect, fmt.Sprintf("COUNT(DISTINCT(%s)) AS %s", SelectCoalesceCustomerUserIDAndUserID, AliasAggr))

	// group by
	var termStmnt string
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName,
		"", &termStmnt, &qParams, termSelect)
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, q.GroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, q.GroupByProperties, q.GroupByTimestamp)

	qStmnt = appendStatement(qStmnt, termStmnt)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersSingleEventQuery
Group By: user_properties

WITH
    step0 AS (
		SELECT distinct(events.id) as event_id, events.user_id as event_user_id FROM events WHERE events.project_id='2'
		AND timestamp>='1552450157' AND timestamp<='1553054957'
		AND events.event_name_id IN ( SELECT id FROM event_names WHERE project_id='2'AND name='View Project' )
    )
	SELECT user_properties.properties->>'gender' as gk_0, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id)))
	FROM step0 LEFT JOIN users ON step0.event_user_id=users.id
	LEFT JOIN user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id GROUP BY gk_0 order by gk_0;
*/
func buildUniqueUsersSingleEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) != 1 {
		return "", nil, errors.New("invalid no.of events for single event query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		return "", nil, errors.New(ErrUnsupportedGroupByEventPropertyOnUserQuery)
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	refStepName := addFilterEventsWithPropsForUsersQuery(projectId, q, &qStmnt, &qParams)

	// select
	var filterSelect string
	filterSelect = appendSelectTimestampColIfRequired(filterSelect, q.GroupByTimestamp)
	filterSelect = joinWithComma(filterSelect, fmt.Sprintf("COUNT(DISTINCT(%s)) AS %s", SelectCoalesceCustomerUserIDAndUserID, AliasAggr))

	// group by
	var termStmnt string
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName,
		"", &termStmnt, &qParams, filterSelect)
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, q.GroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, q.GroupByProperties, q.GroupByTimestamp)

	qStmnt = appendStatement(qStmnt, termStmnt)

	// enclosed by 'with'.
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

	// init with any given event filter query.
	qStmnt, qParams, egKeys, refStepName := buildAnyGivenEventFilterQuery(projectId, q)

	// count.
	userGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	_, _, groupKeys := buildGroupKeys(q.GroupByProperties)

	// select
	tSelect := joinWithComma(ugSelect, egKeys)
	tSelect = appendSelectTimestampColIfRequired(tSelect, q.GroupByTimestamp)
	tSelect = joinWithComma(tSelect, fmt.Sprintf("COUNT(*) AS %s", AliasAggr)) // aggregator.

	termStmnt := "SELECT " + tSelect + " FROM " + refStepName
	// join lateset user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id" +
			" " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	termStmnt = appendGroupByTimestampIfRequired(termStmnt, q.GroupByTimestamp, groupKeys)
	termStmnt = appendOrderByAggr(termStmnt)
	termStmnt = appendLimitByCondition(termStmnt, q.GroupByProperties, q.GroupByTimestamp)

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

	var qSelect string
	qSelect = appendSelectTimestampIfRequired(qSelect, q.Timezone, q.GroupByTimestamp)
	qSelect = joinWithComma(qSelect, egSelect, fmt.Sprintf("COUNT(*) AS %s", AliasAggr))

	addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[0], q.From, q.To,
		"", qSelect, egSelectParams, "")

	qStmnt = appendGroupByTimestampIfRequired(qStmnt, q.GroupByTimestamp, egKeys)
	qStmnt = appendOrderByAggr(qStmnt)
	qStmnt = appendLimitByCondition(qStmnt, q.GroupByProperties, q.GroupByTimestamp)

	return qStmnt, qParams, nil
}

// getColumnNamesForGroupKeys - Replaces groupKeys with real column names.
func getColumnNamesForGroupKeys(cols []string, groupProps []QueryGroupByProperty) ([]string, error) {
	rcols := make([]string, 0, 0)

	indexLookup := make(map[int]QueryGroupByProperty, 0)
	for _, v := range groupProps {
		indexLookup[v.Index] = v
	}

	for i := range cols {
		if strings.HasPrefix(cols[i], GroupKeyPrefix) {
			gIndexStr := strings.TrimPrefix(cols[i], GroupKeyPrefix)
			if gIndex, err := strconv.Atoi(gIndexStr); err != nil {
				log.WithField("group_key", gIndexStr).Error(
					"Invalid group key index. Failed translating group key to real name.")
				return nil, errors.New("invalid group key index")
			} else {
				rcols = append(rcols, indexLookup[gIndex].Property)
			}
		} else {
			rcols = append(rcols, cols[i])
		}
	}

	return rcols, nil
}

// Preserves order of group by properties on query as
// given by the user.
func addIndexToGroupByProperties(query *Query) {
	for i := range query.GroupByProperties {
		query.GroupByProperties[i].Index = i
	}
}

// Returns beginning index of group keys on result.
func getGroupKeyStartIndex(cols []string) (int, error) {
	index := 0
	for _, col := range cols {
		if strings.HasPrefix(col, GroupKeyPrefix) {
			return index, nil
		}
		index++
	}

	return index, errors.New("no group keys found")
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

	return true, ""
}

// BuildQuery - Dispatches corresponding build method based on attributes.
func BuildQuery(projectId uint64, query Query) (string, []interface{}, error) {
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

// Limits results by left and right keys. Assumes result is already
// sorted by count and all group keys are used on SQL group by (makes all three group by
// values together as unique). Limited set dimension = ResultLimit * ResultLimit.
func limitMultiGroupByPropertiesQueryResult(groupPropsLen int, groupByTimestamp bool,
	resultCols []string, resultRows [][]interface{}) ([][]interface{}, error) {

	limitedResult := make([][]interface{}, 0, 0)

	start, err := getGroupKeyStartIndex(resultCols)
	if err != nil {
		return limitedResult, err
	}
	end := groupPropsLen - 1

	// Lookup based on left key (encoded key of all group key values excluding last)
	// right key (last group key value) ie: g1, g2, g3 -> map[c1:g1_c2:g2]map[g3]bool
	keyLookup := make(map[string]map[interface{}]bool, 0)
	for _, row := range resultRows {
		leftKey := getEncodedKeyForCols(row[start:end])

		_, leftKeyExists := keyLookup[leftKey]
		// Limits no.of left keys to ResultsLimit.
		if !leftKeyExists && len(keyLookup) < ResultsLimit {
			keyLookup[leftKey] = make(map[interface{}]bool, 0)
			leftKeyExists = true
		}

		var rightKeyExists bool
		if leftKeyExists {
			// Limits no.of right keys to ResultsLimit.
			_, rightKeyExits := keyLookup[leftKey][row[end]]
			if !rightKeyExits && len(keyLookup[leftKey]) < ResultsLimit {
				keyLookup[leftKey][row[end]] = true
				rightKeyExists = true
			}
		}

		if leftKeyExists && rightKeyExists {
			limitedResult = append(limitedResult, row)
		}
	}

	return limitedResult, nil
}

// Limits top results and makes sure same group key combination available on different
// datetime, if exists on SQL result. Assumes result is sorted by count.
func limitGroupByTimestampQueryResult(groupPropsLen int,
	groupByTimestamp bool, resultCols []string, resultRows [][]interface{}) ([][]interface{}, error) {

	limitedResult := make([][]interface{}, 0, 0)

	start, err := getGroupKeyStartIndex(resultCols)
	if err != nil {
		return limitedResult, err
	}
	end := groupPropsLen

	keyLookup := make(map[string]bool, 0)
	for _, row := range resultRows {
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

	return limitedResult, nil
}

func LimitQueryResults(groupPropsLen int, groupByTimestamp bool,
	resultCols []string, resultRows [][]interface{}) ([][]interface{}, error) {

	if groupPropsLen > 0 && groupByTimestamp {
		return limitGroupByTimestampQueryResult(groupPropsLen, groupByTimestamp, resultCols, resultRows)
	}

	if groupPropsLen > 1 {
		return limitMultiGroupByPropertiesQueryResult(groupPropsLen, groupByTimestamp, resultCols, resultRows)
	}

	// limited already on SQL.
	return resultRows, nil
}

func Analyze(projectId uint64, query Query) ([]string, [][]interface{}, int, string) {
	db := C.GetServices().Db

	valid, errMsg := IsValidQuery(&query)
	if !valid {
		return nil, nil, http.StatusBadRequest, errMsg
	}

	addIndexToGroupByProperties(&query)

	stmnt, params, err := BuildQuery(projectId, query)
	if err != nil {
		log.WithError(err).Error(ErrMsgQueryProcessingFailure)
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	logCtx := log.WithFields(log.Fields{"analytics_query": query, "statement": stmnt, "params": params})
	if stmnt == "" || len(params) == 0 {
		logCtx.Error("Failed generating SQL query from Analyze query")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	rows, err := db.Raw(stmnt, params...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed executing SQL query generated")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	// Todo(Dinesh): Remove ReadRows(inefficient) and use specific struct.
	resultCols, resultRows, err := U.ReadRows(rows)
	if err != nil {
		logCtx.WithError(err).Error("Failed processing SQL query result")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	groupPropsLen := len(query.GroupByProperties)

	resultRows, err = LimitQueryResults(groupPropsLen, query.GroupByTimestamp, resultCols, resultRows)
	if err != nil {
		logCtx.WithError(err).Error("Failed processing query results for limiting")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	realCols, err := getColumnNamesForGroupKeys(resultCols, query.GroupByProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed mapping real column names on SQL query result")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	return realCols, resultRows, http.StatusOK, "Successfully executed query"
}
