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
	Entity string `json:"entity"`
	// Type: categorical or numerical
	Type     string `json:"type"`
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type QueryGroupByProperty struct {
	// Entity: user or event.
	Entity   string `json:"entity"`
	Property string `json:"property"`
	Index    int    `json:"index"`
}

type QueryEventWithProperties struct {
	Name       string          `json:"name"`
	Properties []QueryProperty `json:"properties"`
}

type Query struct {
	Type                 string                     `json:"type"`
	EventsCondition      string                     `json:"eventsCondition"` // all or any
	EventsWithProperties []QueryEventWithProperties `json:"eventsWithProperties"`
	GroupByProperties    []QueryGroupByProperty     `json:"groupByProperties"`
	GroupByTimestamp     bool                       `json:"groupByTimestamp"`
	Timezone             string                     `json:"timezone"`
	From                 int64                      `json:"from"`
	To                   int64                      `json:"to"`
}

const (
	PropertyEntityUser  = "user"
	PropertyEntityEvent = "event"

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
	DefaultTimezone = "UTC"
)

const (
	EqualsOpStr = "equals"
	EqualsOp    = "="
)

var queryOps = map[string]string{
	EqualsOpStr:          EqualsOp,
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
	properties []QueryProperty) (rStmnt string, rParams []interface{}) {

	pLen := len(properties)

	if pLen == 0 {
		return rStmnt, rParams
	}

	rParams = make([]interface{}, 0, 0)
	for i, p := range properties {
		var pStr string
		if p.Type == U.PropertyTypeNumerical {
			pStr = "(%s->>?)::numeric%s?"
		} else {
			pStr = "%s->>?%s?"
		}
		pStmnt := fmt.Sprintf(pStr, getPropertyEntityField(p.Entity), getOp(p.Operator))

		if i == 0 {
			rStmnt = pStmnt
		} else {
			rStmnt = fmt.Sprintf("%s %s %s", rStmnt, mergeCond, pStmnt)
		}

		rParams = append(rParams, p.Property, p.Value)
	}

	return rStmnt, rParams
}

// Alias for group by properties gk_1, gk_2.
func groupKeyByIndex(i int) string {
	return fmt.Sprintf("%s%d", GroupKeyPrefix, i)
}

// groupBySelect: user_properties.properties->'age' as gk_1, events.properties->'category' as gk_2
// groupByKeys: gk_1, gk_2
// How to use?
// select user_properties.properties->'age' as gk_1, events.properties->'category' as gk_2 from events
// group by gk_1, gk_2
func buildGroupKeys(groupProps []QueryGroupByProperty) (groupSelect string,
	groupSelectParams []interface{}, groupKeys string) {

	groupSelectParams = make([]interface{}, 0, 0)

	for i, v := range groupProps {
		// Order of group is preserved as received.
		gKey := groupKeyByIndex(v.Index)
		groupSelect = groupSelect + fmt.Sprintf("%s->? as %s",
			getPropertyEntityField(v.Entity), gKey)
		groupKeys = groupKeys + gKey
		if i < len(groupProps)-1 {
			groupSelect = groupSelect + ", "
			groupKeys = groupKeys + ", "
		}
		groupSelectParams = append(groupSelectParams, v.Property)
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
	wStmnt, wParams := buildWhereFromProperties("AND", qep.Properties)
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

func appendOrderByLastGroupKey(qStmnt string, groupProps []QueryGroupByProperty) string {
	if len(groupProps) == 0 {
		return qStmnt
	}

	maxIndex := -1
	for _, v := range groupProps {
		if v.Index > maxIndex {
			maxIndex = v.Index
		}
	}

	return fmt.Sprintf("%s ORDER BY %s", qStmnt, groupKeyByIndex(maxIndex))
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

func appendGroupByTimestamp(qStmnt string, groupByTimestamp bool, groupKeys ...string) string {
	// Added groups with timestamp.
	groups := make([]string, 0, 0)
	if groupByTimestamp {
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

func BuildAnyGivenEventFilterQuery(projectId uint64, q Query) (string, []interface{}, string, string) {
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
BuildUniqueUsersWithAllGivenEventsQuery builds a query like below,
Group by: user_properties.

WITH
    e1 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
        AND events.properties->>'category'='Sports' AND user_properties.properties->>'gender'='M'
    ),
    e1_e2 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id FROM events
        JOIN e1 ON events.user_id=e1.event_user_id
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='Fund Project')
        AND events.properties->>'category'='Sports' AND user_properties.properties->>'gender'='M'
    ),
    events_with_latest_user_props as (
        SELECT user_properties.properties->'$region' as group_prop1, COALESCE(users.customer_user_id, event_user_id) as real_user_id from e1_e2
        left join users on e1_e2.event_user_id=users.id
        left join user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id
    )
    SELECT group_prop1, count(distinct(real_user_id)) FROM events_with_latest_user_props GROUP BY group_prop1 order by group_prop1;
*/
func BuildUniqueUsersWithAllGivenEventsQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		return "", nil, errors.New(ErrUnsupportedGroupByEventPropertyOnUserQuery)
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	refStepName := addFilterEventsWithPropsForUsersQuery(projectId, q, &qStmnt, &qParams)
	qStmnt = qStmnt + ","

	// Todo(Dinesh): Remove events_with_latest_user_props step add the logic
	// to final select itself, check BuildUniqueUsersWithAnyGivenEventsQuery.
	stepEventsWithLatestUserProps := "events_with_latest_user_props"
	aliasRealUserID := "real_user_id"

	addSelect := SelectCoalesceCustomerUserIDAndUserID + " as " + aliasRealUserID
	addSelect = appendSelectTimestampColIfRequired(addSelect, q.GroupByTimestamp)
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName,
		stepEventsWithLatestUserProps, &qStmnt, &qParams, addSelect)

	// select.
	var qSelect string
	qSelect = appendSelectTimestampColIfRequired(qSelect, q.GroupByTimestamp)
	qSelect = joinWithComma(qSelect, groupKeys, fmt.Sprintf("COUNT(DISTINCT(%s))", aliasRealUserID))

	termStmnt := "SELECT " + qSelect + " FROM " + stepEventsWithLatestUserProps
	termStmnt = appendGroupByTimestamp(termStmnt, q.GroupByTimestamp, groupKeys)

	qStmnt = appendStatement(qStmnt, termStmnt)
	qStmnt = appendOrderByLastGroupKey(qStmnt, q.GroupByProperties)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
BuildUniqueUsersWithAnyGivenEventsQuery
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
	SELECT user_properties.properties->'gender' as gk_0, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id)))
	FROM any_event LEFT JOIN users ON any_event.event_user_id=users.id
	LEFT JOIN user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id GROUP BY gk_0 order by gk_0;

*/
func BuildUniqueUsersWithAnyGivenEventsQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		return "", nil, errors.New(ErrUnsupportedGroupByEventPropertyOnUserQuery)
	}

	// init with any given event filter query.
	qStmnt, qParams, _, refStepName := BuildAnyGivenEventFilterQuery(projectId, q)

	// select
	var termSelect string
	termSelect = appendSelectTimestampColIfRequired(termSelect, q.GroupByTimestamp)
	termSelect = joinWithComma(termSelect, fmt.Sprintf("COUNT(DISTINCT(%s))", SelectCoalesceCustomerUserIDAndUserID))

	// group by
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName,
		"", &qStmnt, &qParams, termSelect)
	qStmnt = appendGroupByTimestamp(qStmnt, q.GroupByTimestamp, groupKeys)
	qStmnt = appendOrderByLastGroupKey(qStmnt, q.GroupByProperties)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
BuildUniqueUsersSingleEventQuery
Group By: user_properties

WITH
    step0 AS (
		SELECT distinct(events.id) as event_id, events.user_id as event_user_id FROM events WHERE events.project_id='2'
		AND timestamp>='1552450157' AND timestamp<='1553054957'
		AND events.event_name_id IN ( SELECT id FROM event_names WHERE project_id='2'AND name='View Project' )
    )
	SELECT user_properties.properties->'gender' as gk_0, COUNT(DISTINCT(COALESCE(users.customer_user_id, event_user_id)))
	FROM step0 LEFT JOIN users ON step0.event_user_id=users.id
	LEFT JOIN user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id GROUP BY gk_0 order by gk_0;
*/
func BuildUniqueUsersSingleEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
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
	filterSelect = joinWithComma(filterSelect, fmt.Sprintf("COUNT(DISTINCT(%s))", SelectCoalesceCustomerUserIDAndUserID))

	// group by
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName,
		"", &qStmnt, &qParams, filterSelect)
	qStmnt = appendGroupByTimestamp(qStmnt, q.GroupByTimestamp, groupKeys)
	qStmnt = appendOrderByLastGroupKey(qStmnt, q.GroupByProperties)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
BuildEventsOccurrenceWithAnyGivenEventQuery builds query for any given event and single event query,
Group by: user_properties, event_properties.

WITH
    e1 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id, events.properties->'category' as group_prop1 FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
        AND user_properties.properties->>'gender'='M'
    ),
    e2 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id, events.properties->'category' as group_prop1 FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='Fund Project')
        AND user_properties.properties->>'gender'='M'
    ),
    any_event AS (
        SELECT event_id, event_user_id, group_prop1 FROM e1 UNION ALL SELECT event_id, event_user_id, group_prop1 FROM e2
    )
    SELECT user_properties.properties->'$region' as group_prop2, group_prop1, count(*) from any_event
    left join users on any_event.event_user_id=users.id
    left join user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id
    group by group_prop1, group_prop2 order by group_prop2;
*/
func BuildEventsOccurrenceWithAnyGivenEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	// init with any given event filter query.
	qStmnt, qParams, egKeys, refStepName := BuildAnyGivenEventFilterQuery(projectId, q)

	// count.
	userGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	_, _, groupKeys := buildGroupKeys(q.GroupByProperties)

	// select
	tSelect := joinWithComma(ugSelect, egKeys)
	tSelect = appendSelectTimestampColIfRequired(tSelect, q.GroupByTimestamp)
	tSelect = joinWithComma(tSelect, "COUNT(*)") // aggregator.

	termStmnt := "SELECT " + tSelect + " FROM " + refStepName
	// join lateset user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id" +
			" " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	termStmnt = appendGroupByTimestamp(termStmnt, q.GroupByTimestamp, groupKeys)

	qParams = append(qParams, ugSelectParams...)
	qStmnt = appendStatement(qStmnt, termStmnt)
	qStmnt = appendOrderByLastGroupKey(qStmnt, q.GroupByProperties)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

/*
BuildEventsOccurrenceWithAnyGivenEventQuery builds query for any given event and single event query,
Group by: user_properties, event_properties.

* Without group by user_property

WITH
	SELECT COUNT(*), events.properties->'category' as group_prop1 FROM events
	LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
	WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
	AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
	AND user_properties.properties->>'gender'='M'


* With group by user_property

WITH
    e1 AS (
        SELECT distinct(events.id) as event_id, events.user_id as event_user_id, events.properties->'category' as group_prop1 FROM events
        LEFT JOIN user_properties ON events.user_properties_id=user_properties.id
		WHERE events.project_id=2 AND events.timestamp >= 1393632004 AND events.timestamp <= 1396310325
		AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id='2' AND name='View Project')
        AND user_properties.properties->>'gender'='M'
    )
    SELECT user_properties.properties->'$region' as group_prop2, group_prop1, count(*) from e1
    left join users on e1.event_user_id=users.id
    left join user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id
    group by group_prop1, group_prop2 order by group_prop2;
*/

func BuildEventsOccurrenceSingleEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) != 1 {
		return "", nil, errors.New("invalid no.of events for single event query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityUser) {
		// Using any given event query, which handles groups already.
		return BuildEventsOccurrenceWithAnyGivenEventQuery(projectId, q)
	}

	var qStmnt string
	var qParams []interface{}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egSelect, egSelectParams, egKeys := buildGroupKeys(eventGroupProps)

	var qSelect string
	qSelect = appendSelectTimestampIfRequired(qSelect, q.Timezone, q.GroupByTimestamp)
	qSelect = joinWithComma(qSelect, egSelect, "COUNT(*)")

	addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[0], q.From, q.To,
		"", qSelect, egSelectParams, "")

	qStmnt = appendGroupByTimestamp(qStmnt, q.GroupByTimestamp, egKeys)
	qStmnt = appendOrderByLastGroupKey(qStmnt, q.GroupByProperties)

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

// Validates and returns errMsg which is used as response.
func validateQuery(query Query) string {
	if query.Type != QueryTypeEventsOccurrence &&
		query.Type != QueryTypeUniqueUsers {
		return "Unknown query type given"
	}

	if query.EventsCondition != EventCondAllGivenEvent &&
		query.EventsCondition != EventCondAnyGivenEvent {
		return "Unknown events condition given"
	}

	if len(query.EventsWithProperties) == 0 {
		return "No events to process"
	}

	if query.From == 0 || query.To == 0 {
		return "Invalid query time range"
	}

	return ""
}

// BuildQuery - Dispatches corresponding build method based on attributes.
func BuildQuery(projectId uint64, query Query) (string, []interface{}, error) {
	if query.Type == QueryTypeEventsOccurrence {
		if len(query.EventsWithProperties) == 1 {
			return BuildEventsOccurrenceSingleEventQuery(projectId, query)
		}

		if query.EventsCondition == EventCondAnyGivenEvent {
			return BuildEventsOccurrenceWithAnyGivenEventQuery(projectId, query)
		}

		return "", nil, errors.New("all given events condition is not possible on events query")
	}

	if query.Type == QueryTypeUniqueUsers {
		if len(query.EventsWithProperties) == 1 {
			return BuildUniqueUsersSingleEventQuery(projectId, query)
		}

		if query.EventsCondition == EventCondAnyGivenEvent {
			return BuildUniqueUsersWithAnyGivenEventsQuery(projectId, query)
		}

		return BuildUniqueUsersWithAllGivenEventsQuery(projectId, query)
	}

	return "", nil, errors.New("invalid query")
}

func Analyze(projectId uint64, query Query) ([]string, map[int][]interface{}, int, string) {
	db := C.GetServices().Db

	errMsg := validateQuery(query)
	if errMsg != "" {
		return nil, nil, http.StatusBadRequest, errMsg
	}

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

	// Todo(Dinesh): Reomve RowsToMap(inefficient) and use specific struct.
	resultCols, resulRows, err := U.DBRowsToMap(rows)
	if err != nil {
		logCtx.WithError(err).Error("Failed processing SQL query result")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	realCols, err := getColumnNamesForGroupKeys(resultCols, query.GroupByProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed mapping real column names on SQL query result")
		return nil, nil, http.StatusInternalServerError, ErrMsgQueryProcessingFailure
	}

	return realCols, resulRows, http.StatusOK, "Successfully executed query"
}
