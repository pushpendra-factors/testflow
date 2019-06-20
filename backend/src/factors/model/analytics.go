package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
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
	Class                string                     `json:"cl"`
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

type QueryResutlMeta struct {
	EventsWithProperties []QueryEventWithProperties `json:"ewp"`
	GroupByProperties    []QueryGroupByProperty     `json:"gbp"`
}

type QueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Meta    QueryResutlMeta `json:"meta"`
}

const (
	PropertyEntityUser  = "user"
	PropertyEntityEvent = "event"

	PropertyValueNone = "$none"

	EventCondAnyGivenEvent = "any_given_event"
	EventCondAllGivenEvent = "all_given_event"

	QueryClassInsights = "insights"
	QueryClassFunnel   = "funnel"

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
	AliasDate       = "date"
	AliasAggr       = "count"
	DefaultTimezone = "UTC"
	ResultsLimit    = 20
	MaxResultsLimit = 100000

	FunnelConversionPrefix = "conversion"
	FunnelStepPrefix       = "step"
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

func getIntervalFromRangeStrAsUnix(rangeStr string) (int64, int64, error) {
	splitedRange := strings.Split(rangeStr, ":")

	startAsInt, err := strconv.ParseInt(splitedRange[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	endAsInt, err := strconv.ParseInt(splitedRange[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	actStart := now.New(time.Unix(startAsInt, 0).UTC()).BeginningOfDay().Unix()
	actEnd := now.New(time.Unix(endAsInt, 0).UTC()).EndOfDay().Unix()

	return actStart, actEnd, nil
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
			var pStmnt string
			if p.Type == U.PropertyTypeDateTime {
				pStmnt = fmt.Sprintf("(%s->>?>=? AND %s->>?<=?)", propertyEntity, propertyEntity)

				start, end, err := getIntervalFromRangeStrAsUnix(p.Value)
				if err != nil {
					log.WithError(err).Error("Failed reading timestamp on user join query.")
					return "", nil, err
				}
				rParams = append(rParams, p.Property, start, p.Property, end)
			} else {
				pStmnt = fmt.Sprintf("%s->>?%s?", propertyEntity, propertyOp)
				rParams = append(rParams, p.Property, p.Value)
			}

			if i == 0 {
				rStmnt = pStmnt
			} else {
				rStmnt = fmt.Sprintf("%s %s %s", rStmnt, mergeCond, pStmnt)
			}

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

// Adds a step of events filter with QueryEventWithProperties.
func addFilterEventsWithPropsQuery(projectId uint64, qStmnt *string, qParams *[]interface{},
	qep QueryEventWithProperties, from int64, to int64, fromStr string,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string, filterGroupBy string) error {

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

	whereCond := fmt.Sprintf("WHERE events.project_id=? AND timestamp>=%s AND timestamp<=?"+
		" "+"AND events.event_name_id IN ( SELECT id FROM event_names WHERE project_id=? AND name=? )", fromTimestamp)
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
	wStmnt, wParams, err := buildWhereFromProperties("AND", qep.Properties)
	if err != nil {
		return err
	}

	if wStmnt != "" {
		rStmnt = rStmnt + " AND " + fmt.Sprintf("( %s )", wStmnt)
		*qParams = append(*qParams, wParams...)
	}

	if filterGroupBy != "" {
		rStmnt = fmt.Sprintf("%s group by %s", rStmnt, filterGroupBy)
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

		var filterSelect string
		var usersFromLastStep string
		if i > 0 {
			filterSelect = SelectDefaultEventFilterWithDistinct
			usersFromLastStep = "JOIN " + prevStepName + " ON events.user_id=" + prevStepName + ".event_user_id"
		} else {
			filterSelect = SelectDefaultEventFilter
		}

		filterSelect = appendSelectTimestampIfRequired(filterSelect, query.Timezone, query.GroupByTimestamp)
		addFilterEventsWithPropsQuery(projectId, &rStmnt, &rParams, ewp, query.From, query.To, "",
			stepName, filterSelect, nil, usersFromLastStep, "")

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
		addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, ewp, q.From, q.To, "",
			refStepName, filterSelect, groupParams, "", "")
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
		"", "", qSelect, egSelectParams, "", "")

	qStmnt = appendGroupByTimestampIfRequired(qStmnt, q.GroupByTimestamp, egKeys)
	qStmnt = appendOrderByAggr(qStmnt)
	qStmnt = appendLimitByCondition(qStmnt, q.GroupByProperties, q.GroupByTimestamp)

	return qStmnt, qParams, nil
}

/*
buildUniqueUsersFunnelQuery

WITH
	step1 AS (
		SELECT events.user_id, 1 as step1, min(timestamp) as step1_timestamp from events
		LEFT JOIN users on events.user_id=users.id where events.project_id = 1
		and event_name_id IN (select id from event_names where project_id=1 and name = 'localhost:3000/#/core')
		and timestamp between 1393632004 and 1560231260 group by events.user_id
	),
	step2 as (
		SELECT events.user_id, 1 as step2, min(timestamp) as step2_timestamp from events
		LEFT JOIN step1 on events.user_id=step1.user_id where events.project_id = 1
		and event_name_id IN (select id from event_names where project_id=1 and name = 'run_query')
		and timestamp between step1.step1_timestamp and 1560231260 group by events.user_id -- from timestamp, after step1 --
	),
	step3 as (
		SELECT events.user_id, 1 as step3 from events
		LEFT JOIN step2 on events.user_id=step2.user_id where events.project_id = 1
		and event_name_id IN (select id from event_names where project_id=1 and name = 'localhost:3000/#/dashboard')
		and timestamp between step2.step2_timestamp and 1560231260 group by events.user_id
	),
	funnel as (
		SELECT CASE WHEN user_properties.properties->>'$city' IS NULL THEN '$none'
        WHEN user_properties.properties->>'$city' = '' THEN '$none'
        ELSE user_properties.properties->>'$city' END AS group_key_0,
        step1, step2, step3 from step1
        LEFT JOIN users on step1.user_id=users.id -- join users and user_properties only if group by available --
        LEFT JOIN user_properties on users.properties_id=user_properties.id
		LEFT JOIN step2 on step1.user_id=step2.user_id
		LEFT JOIN step3 on step2.user_id=step3.user_id
	)
	SELECT '$no_group' as group_key_0, SUM(step1) as step1, SUM(step2) as step2, SUM(step3) as step3,
	ROUND((SUM(step2)::DECIMAL/SUM(step1)::DECIMAL) * 100) as step1_step2_conversion,
	ROUND((SUM(step3)::DECIMAL/SUM(step2)::DECIMAL) * 100) as step2_step3_conversion,
	ROUND((SUM(step3)::DECIMAL/SUM(step1)::DECIMAL) * 100) as overall_conversion from funnel
    UNION ALL
	SELECT group_key_0, SUM(step1) as step1, SUM(step2) as step2, SUM(step3) as step3,
	ROUND((SUM(step2)::DECIMAL/SUM(step1)::DECIMAL) * 100) as step1_step2_conversion,
	ROUND((SUM(step3)::DECIMAL/SUM(step2)::DECIMAL) * 100) as step2_step3_conversion,
	ROUND((SUM(step3)::DECIMAL/SUM(step1)::DECIMAL) * 100) as overall_conversion from funnel
    group by group_key_0;
*/
func buildUniqueUsersFunnelQuery(projectId uint64, q Query) (string, []interface{}, error) {
	if len(q.EventsWithProperties) == 0 {
		return "", nil, errors.New("invalid no.of events for funnel query")
	}

	if hasGroupEntity(q.GroupByProperties, PropertyEntityEvent) {
		return "", nil, errors.New("user funnel doesn't support group by event property")
	}

	var qStmnt string
	var qParams []interface{}

	var funnelSteps []string
	for i := range q.EventsWithProperties {
		var addParams []interface{}
		stepName := fmt.Sprintf("%s_%d", FunnelStepPrefix, i)
		// builds "events.user_id, 1 AS step1, min(timestamp) AS step1_timestamp"
		addSelect := "events.user_id, 1 AS " + stepName + ", min(timestamp) AS " + fmt.Sprintf("%s_timestamp", stepName)

		var from int64
		var fromStr string
		// use actual from for first step,
		// and min timestamp from prev step
		// for next step's from.
		if i == 0 {
			from = q.From
		} else {
			// builds "step_2.step_2_timestamp"
			prevStepName := fmt.Sprintf("%s_%d", FunnelStepPrefix, i-1)
			fromStr = fmt.Sprintf("%s.%s_timestamp", prevStepName, prevStepName)
		}

		var usersJoinStmnt string
		if i == 0 {
			usersJoinStmnt = "LEFT JOIN users ON events.user_id=users.id"
		} else {
			// builds "LEFT JOIN step_0 on events.user_id=step_0.user_id"
			prevStepName := fmt.Sprintf("%s_%d", FunnelStepPrefix, i-1)
			usersJoinStmnt = fmt.Sprintf("LEFT JOIN %s on events.user_id=%s.user_id", prevStepName, prevStepName)
		}

		err := addFilterEventsWithPropsQuery(projectId, &qStmnt, &qParams, q.EventsWithProperties[i],
			from, q.To, fromStr, stepName, addSelect, addParams, usersJoinStmnt, "events.user_id")
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

	groupSelect, groupParams, groupKeys := buildGroupKeys(q.GroupByProperties)
	qParams = append(qParams, groupParams...)

	// Join user properties if group by propertie given.
	var properitesJoinStmnt string
	if len(q.GroupByProperties) > 0 {
		// builds "LEFT JOIN users on step1.user_id=users.id LEFT JOIN user_properties on users.properties_id=user_properties.id"
		properitesJoinStmnt = fmt.Sprintf("LEFT JOIN users on %s.user_id=users.id LEFT JOIN user_properties on users.properties_id=user_properties.id", funnelSteps[0])
	}

	var stepsJoinStmnt string
	for i, fs := range funnelSteps {
		if i > 0 {
			// builds "LEFT JOIN step2 on step1.user_id=step2.user_id"
			stepsJoinStmnt = appendStatement(stepsJoinStmnt,
				fmt.Sprintf("LEFT JOIN %s ON %s.user_id=%s.user_id", fs, funnelSteps[i-1], fs))
		}
	}

	funnelStepName := "funnel"
	funnelStmnt := "SELECT" + " " + joinWithComma(groupSelect, funnelStepsForSelect) + " " + "FROM" + " " +
		funnelSteps[0] + " " + properitesJoinStmnt + " " + stepsJoinStmnt
	funnelStmnt = as(funnelStepName, funnelStmnt)
	qStmnt = joinWithComma(qStmnt, funnelStmnt)

	// builds "SUM(step1) AS step1, SUM(step1) AS step2",
	var rawCountSelect string
	for _, fs := range funnelSteps {
		rawCountSelect = joinWithComma(rawCountSelect, fmt.Sprintf("SUM(%s) AS %s", fs, fs))
	}

	var aggrSelect string
	for i, fs := range funnelSteps {
		if i > 0 {
			aggrSelect = appendStatement(aggrSelect,
				fmt.Sprintf("ROUND((SUM(%s)::DECIMAL/SUM(%s)::DECIMAL) * 100) AS %s_%s_%s,",
					fs, funnelSteps[i-1], FunnelConversionPrefix, funnelSteps[i-1], fs))
		}
	}
	// append overall conversion.
	aggrSelect = appendStatement(aggrSelect,
		fmt.Sprintf("ROUND((SUM(%s)::DECIMAL/SUM(%s)::DECIMAL) * 100) AS %s_overall",
			funnelSteps[len(funnelSteps)-1], funnelSteps[0], FunnelConversionPrefix))

	aggrSelect = joinWithComma(rawCountSelect, aggrSelect)

	var termStmnt string
	if len(q.GroupByProperties) == 0 {
		termStmnt = "SELECT" + " " + aggrSelect + " " + "FROM" + " " + funnelStepName
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
		noGroupSelect := "SELECT" + " " + joinWithComma(groupKeysPlaceholder, aggrSelect) +
			" " + "FROM" + " " + funnelStepName
		groupBySelect := "SELECT" + " " + joinWithComma(groupKeys, aggrSelect) + " " +
			"FROM" + " " + funnelStepName + " " + "GROUP BY" + " " + groupKeys
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
		if strings.HasPrefix(h, FunnelConversionPrefix) || strings.HasPrefix(h, FunnelStepPrefix) {
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

func getTimstampAndAggregateIndexOnResult(resultCols []string) (int, int, error) {
	timeIndex := -1
	aggrIndex := -1

	for i, c := range resultCols {
		if c == AliasDate {
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

func addMissingTimestampsOnResultWithoutGroupByProps(result *QueryResult, query *Query,
	aggrIndex int, timestampIndex int) error {

	rowsByTimestamp := make(map[string][]interface{}, 0)
	for _, row := range result.Rows {
		ts := row[timestampIndex].(time.Time)
		rowsByTimestamp[U.GetDateOnlyTimestampStr(ts)] = row
	}

	timestamps := U.GetAllDatesAsTimestamp(query.From, query.To)

	filledResult := make([][]interface{}, 0, 0)
	// range over timestamps between given from and to.
	// uses timstamp string for comparison.
	for _, ts := range timestamps {
		if row, exists := rowsByTimestamp[U.GetDateOnlyTimestampStr(ts)]; exists {
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
		// encoded key with group values and timestamp from db row.
		encCols = append(encCols, U.GetDateOnlyTimestampStr(row[timestampIndex].(time.Time)))
		encKey := getEncodedKeyForCols(encCols)

		rowsByGroupAndTimestamp[encKey] = true
		filledResult = append(filledResult, row)
	}

	timestamps := U.GetAllDatesAsTimestamp(query.From, query.To)

	for _, row := range result.Rows {
		for _, ts := range timestamps {
			encCols := make([]interface{}, 0, 0)
			encCols = append(encCols, row[gkStart:gkEnd]...)
			// encoded key with generated timestamp.
			encCols = append(encCols, U.GetDateOnlyTimestampStr(ts))
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
	aggrIndex, timeIndex, err := getTimstampAndAggregateIndexOnResult(result.Headers)
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
	if query.GroupByTimestamp {
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
	result.Meta.EventsWithProperties = query.EventsWithProperties
	result.Meta.GroupByProperties = query.GroupByProperties
}

func RunInsightsQuery(projectId uint64, query Query) (*QueryResult, int, string) {
	stmnt, params, err := BuildInsightsQuery(projectId, query)
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

	groupPropsLen := len(query.GroupByProperties)

	err = LimitQueryResult(result, groupPropsLen, query.GroupByTimestamp)
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
