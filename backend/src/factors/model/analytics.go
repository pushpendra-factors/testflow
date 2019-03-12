package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"

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
	From                 int64                      `json:"from"`
	To                   int64                      `json:"to"`
}

const (
	PropertyEntityUser  = "user"
	PropertyEntityEvent = "event"
)

var queryOps = map[string]string{
	"equals":              "=",
	"greaterThan":         ">",
	"lesserThan":          "<",
	"greaterThanOrEquals": ">=",
	"lessertThanOrEquals": "<=",
}

func with(stmnt string) string {
	return fmt.Sprintf("with %s", stmnt)
}

func getOp(OpStr string) string {
	v, ok := queryOps[OpStr]
	if !ok {
		log.Fatalf("invalid query operator %s", OpStr)
	}

	return v
}

func getPropertyEntityField(entityName string) string {
	if entityName == PropertyEntityUser {
		return "user_properties.properties"
	} else if entityName == PropertyEntityEvent {
		return "events.properties"
	} else {
		log.Fatalf("invalid property type %s", entityName)
	}

	return ""
}

func as(asName, asQuery string) string {
	return fmt.Sprintf("%s as (%s)", asName, asQuery)
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
		pStmnt := fmt.Sprintf("%s->>?%s?",
			getPropertyEntityField(p.Entity), getOp(p.Operator))

		if i == 0 {
			rStmnt = pStmnt
		} else {
			rStmnt = fmt.Sprintf("%s %s %s", rStmnt, mergeCond, pStmnt)
		}

		// value is kept string always as ? on gorm, prepares
		// values with '' even for integers.
		rParams = append(rParams, p.Property, p.Value)
	}

	return rStmnt, rParams
}

// Alias for group by properties gk_1, gk_2.
func groupKeyByIndex(i int) string {
	return fmt.Sprintf("gk_%d", i)
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
	stepName string, addSelecStmnt string, addSelectParams []interface{}, addJoinStmnt string) error {

	if from == 0 || to == 0 {
		return errors.New("invalid time range for events filtering")
	}

	rStmnt := ""
	selectStmnt := "distinct(events.id) as event_id, events.user_id as event_user_id"
	rStmnt = appendStatement(
		rStmnt,
		"SELECT "+joinWithComma(selectStmnt, addSelecStmnt)+" FROM events"+" "+addJoinStmnt,
	)
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

	rStmnt = as(stepName, rStmnt)
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

func joinWithComma(x, y string) string {
	if x != "" && y != "" {
		return fmt.Sprintf("%s, %s", x, y)
	}

	if x != "" {
		return x
	}

	if y != "" {
		return y
	}

	return ""
}

func hasGroupEntity(props []QueryGroupByProperty, entity string) bool {
	for _, p := range props {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func addJoinLatestUserPropsQuery(groupProps []QueryGroupByProperty, refStepName string,
	qStmnt *string, qParams *[]interface{}, addSelect string) string {

	groupSelect, gSelectParams, gKeys := buildGroupKeys(groupProps)

	rStmnt := "SELECT " + joinWithComma(groupSelect, addSelect) + " from " + refStepName +
		" " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id"

	if hasGroupEntity(groupProps, PropertyEntityUser) {
		rStmnt = rStmnt + " " + " LEFT JOIN user_properties on users.id=user_properties.user_id and user_properties.id=users.properties_id"
	}

	rStmnt = as("events_with_latest_user_props", rStmnt)

	*qStmnt = appendStatement(*qStmnt, rStmnt)
	*qParams = append(*qParams, gSelectParams...)

	return gKeys
}

// Adds additional join to the events filter,  for considering users only from prev
// step and with diff step names.
func addFilterEventsWithPropsForUsersQuery(projectId uint64, eventsWithProperties []QueryEventWithProperties,
	from int64, to int64, qStmnt *string, qParams *[]interface{}) string {

	var rStmnt string
	var rParams []interface{}

	stepName := ""
	for i, ewp := range eventsWithProperties {
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

		addFilterEventsWithPropsQuery(projectId, &rStmnt, &rParams, ewp, from, to,
			stepName, "", nil, usersFromLastStep)
		rStmnt = rStmnt + ", "
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
    SELECT group_prop1, count(distinct(real_user_id)) FROM events_with_latest_user_props GROUP BY group_prop1;
*/
func BuildUniqueUsersWithAllGivenEventsQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	qStmnt := ""
	qParams := make([]interface{}, 0, 0)

	refStepName := addFilterEventsWithPropsForUsersQuery(projectId, q.EventsWithProperties, q.From, q.To, &qStmnt, &qParams)

	colasceUser := "COALESCE(users.customer_user_id, event_user_id) as real_user_id"
	groupKeys := addJoinLatestUserPropsQuery(q.GroupByProperties, refStepName, &qStmnt, &qParams, colasceUser)

	// Final unique users.
	if groupKeys == "" {
		qStmnt = appendStatement(qStmnt, "SELECT count(distinct(real_user_id)) FROM events_with_latest_user_props;")
	} else {
		qStmnt = appendStatement(qStmnt, fmt.Sprintf("SELECT %s count(distinct(real_user_id)) FROM events_with_latest_user_props GROUP BY %s;", groupKeys+" ,", groupKeys))
	}

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
    group by group_prop1, group_prop2;
*/
func BuildEventsOccurrenceWithAnyGivenEventQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("zero events on the query")
	}

	var qStmnt string
	var qParams []interface{}

	eventGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityEvent)
	egSelect, egSelectParams, egKeys := buildGroupKeys(eventGroupProps)

	// event filters.
	filters := make([]string, 0)
	refStepName := ""
	for i, ewp := range q.EventsWithProperties {
		refStepName = fmt.Sprintf("step%d", i)
		filters = append(filters, refStepName)
		addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, ewp, q.From, q.To,
			refStepName, egSelect, egSelectParams, "")
		if len(q.EventsWithProperties) > 1 {
			qStmnt = qStmnt + ", "
		}
	}

	var eventsSrc string
	if len(q.EventsWithProperties) == 1 {
		// one event. no union required.
		eventsSrc = refStepName
	} else {
		// union.
		unionStmnt := ""
		for _, filter := range filters {
			if unionStmnt != "" {
				unionStmnt = unionStmnt + " UNION ALL "
			}

			groupKeys := ""
			if egKeys != "" {
				groupKeys = ", " + egKeys
			}
			unionStmnt = unionStmnt + " SELECT event_id, event_user_id " + groupKeys + " FROM " + filter
		}
		unionStmnt = as("any_event", unionStmnt)
		qStmnt = appendStatement(qStmnt, unionStmnt)
		eventsSrc = "any_event"
	}

	// count.
	userGroupProps := filterGroupPropsByType(q.GroupByProperties, PropertyEntityUser)
	ugSelect, ugSelectParams, _ := buildGroupKeys(userGroupProps)
	_, _, groupKeys := buildGroupKeys(q.GroupByProperties)

	groupSelect := joinWithComma(ugSelect, egKeys)
	if groupSelect != "" {
		groupSelect = groupSelect + ","
	}

	termStmnt := "SELECT " + groupSelect + " count(*) FROM " + eventsSrc
	// join lateset user_properties, only if group by user property present.
	if ugSelect != "" {
		termStmnt = termStmnt + " " + "LEFT JOIN users ON " + eventsSrc + ".event_user_id=users.id" +
			" " + "LEFT JOIN user_properties ON users.id=user_properties.user_id AND user_properties.id=users.properties_id"
	}
	// add group by if group by conditions present.
	if groupKeys != "" {
		termStmnt = termStmnt + " " + "GROUP BY" + " " + groupKeys
	}

	qParams = append(qParams, ugSelectParams...)
	qStmnt = appendStatement(qStmnt, termStmnt)

	// enclosed by 'with'.
	qStmnt = with(qStmnt)

	return qStmnt, qParams, nil
}

func Analyze(projectId uint64, query Query) ([]string, map[int][]interface{}, error) {
	db := C.GetServices().Db

	if query.From == 0 || query.To == 0 {
		return nil, nil, errors.New("invalid time range on query")
	}

	var stmnt string
	var params []interface{}
	var err error

	// use switch
	if query.Type == "unique_users" {
		stmnt, params, err = BuildUniqueUsersWithAllGivenEventsQuery(projectId, query)
	} else if query.Type == "events_occurrence" {
		stmnt, params, err = BuildEventsOccurrenceWithAnyGivenEventQuery(projectId, query)
	} else {
		return nil, nil, errors.New("unknown analyze query type")
	}

	if err != nil {
		return nil, nil, err
	}

	if stmnt == "" || len(params) == 0 {
		log.WithFields(log.Fields{}).Error("Invalid query statement and params generated.")
		return nil, nil, errors.New("invalid analyze query statement or params generated")
	}

	rows, err := db.Raw(stmnt, params...).Rows()
	if err != nil {
		return nil, nil, err
	}

	// Todo(Dinesh): Reomve RowsToMap(inefficient) and use specific struct.
	return U.DBRowsToMap(rows)
}
