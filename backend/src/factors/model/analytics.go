package model

import (
	"encoding/json"
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
	Type     string `json:"pty"` // Property type categorical / numerical.
	// group by specific event name.
	EventName      string `json:"ena"`
	EventNameIndex int    `json:"eni"`
	Granularity    string `json:"grn"` // currently used only for datetime - year/month/week/day/hour
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

type QueryGroup struct {
	Queries []Query `json:"query_group"`
}

func (q *QueryGroup) GetClass() string {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to belong to same class
		return q.Queries[0].Class
	}
	return ""
}

func (q *QueryGroup) GetQueryDateRange() (from, to int64) {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to run for same time range
		return q.Queries[0].From, q.Queries[0].To
	}
	return 0, 0
}

func (q *QueryGroup) SetQueryDateRange(from, to int64) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].From, q.Queries[i].To = from, to
	}
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
	Query       Query        `json:"query"`
	Currency    string       `json:"currency"` //Currency field is used for Attribution query response.
	MetaMetrics []HeaderRows `json:"metrics"`
}

type HeaderRows struct {
	Title   string          `json:"title"`
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
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
	// Todo(Dinesh): Use Generic query result
	// for meta as interface{}.
	Meta QueryResultMeta `json:"meta"`
}

// GenericQueryResult - Common query result
// structure with meta.
type GenericQueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Meta    interface{}     `json:"meta"`
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

	EventCondAnyGivenEvent  = "any_given_event"
	EventCondAllGivenEvent  = "all_given_event"
	EventCondEachGivenEvent = "each_given_event"

	QueryClassEvents      = "events"
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

	SelectDefaultEventFilter        = "events.id as event_id, events.user_id as event_user_id"
	SelectDefaultUserFilter         = "events.user_id as event_user_id"
	SelectDefaultEventFilterByAlias = "event_id, event_user_id, event_name"
	SelectDefaultUserFilterByAlias  = "coal_user_id, event_user_id, event_name"

	GroupKeyPrefix  = "_group_key_"
	AliasEventName  = "event_name"
	AliasDateTime   = "datetime"
	AliasAggr       = "count"
	AliasError      = "error"
	DefaultTimezone = "UTC"
	ResultsLimit    = 100
	MaxResultsLimit = 100000

	StepPrefix             = "step_"
	FunnelConversionPrefix = "conversion_"

	NumericalGroupByBuckets       = 10
	NumericalGroupBySeparator     = " - "
	NumericalLowerBoundPercentile = 0.02
	NumericalUpperBoundPercentile = 0.98
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

var trailingZeroRegex = regexp.MustCompile(`\.0\b`)

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
				pStmnt = fmt.Sprintf("CASE WHEN json_typeof(%s::json->?) = 'number' THEN  (%s->>?)::float %s ? ELSE false END", propertyEntity, propertyEntity, propertyOp)
				rParams = append(rParams, p.Property, p.Property, p.Value)
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

// returns SQL query condition to address conditions only on events.properties
func getFilterSQLStmtForEventProperties(properties []QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	var filteredProperty []QueryProperty
	for _, p := range properties {

		propertyEntity := getPropertyEntityField(p.Entity)
		if propertyEntity == "events.properties" {
			filteredProperty = append(filteredProperty, p)
		}
	}
	wStmt, wParams, err := buildWhereFromProperties(filteredProperty)
	if err != nil {
		return "", nil, err
	}
	return wStmt, wParams, nil
}

// returns SQL query condition to address conditions only on user_properties.properties
func getFilterSQLStmtForUserProperties(properties []QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	var filteredProperty []QueryProperty
	for _, p := range properties {

		propertyEntity := getPropertyEntityField(p.Entity)
		if propertyEntity == "user_properties.properties" {
			filteredProperty = append(filteredProperty, p)
		}
	}
	wStmt, wParams, err := buildWhereFromProperties(filteredProperty)
	if err != nil {
		return "", nil, err
	}
	return wStmt, wParams, nil
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
/* SAMPLE QUERY FOR date time property
SELECT CASE WHEN events.properties->>'check_Timestamp' IS NULL THEN '$none' WHEN events.properties->>'check_Timestamp' = '' THEN '$none' ELSE
CASE WHEN jsonb_typeof(events.properties->'check_Timestamp') = 'number' THEN date_trunc('day', to_timestamp((events.properties->'check_Timestamp')::numeric)::timestamp)::text
ELSE date_trunc('day', to_timestamp((events.properties->>'check_Timestamp')::numeric)::timestamp)::text END END AS _group_key_0, COUNT(*) AS count FROM events
WHERE events.project_id='1' AND timestamp>='1602527400' AND timestamp<='1602576868' AND events.event_name_id IN
(SELECT id FROM event_names WHERE project_id='1' AND name='factors-dev.com:3000/#/settings') GROUP BY _group_key_0 ORDER BY count DESC LIMIT 100
*/
func getNoneHandledGroupBySelect(groupProp QueryGroupByProperty, groupKey string) (string, []interface{}) {
	entityField := getPropertyEntityField(groupProp.Entity)
	var groupSelect string
	groupSelectParams := make([]interface{}, 0)
	if groupProp.Type != U.PropertyTypeDateTime {
		groupSelect = fmt.Sprintf("CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' ELSE %s->>? END AS %s",
			entityField, PropertyValueNone, entityField, PropertyValueNone, entityField, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	} else {
		groupSelect = fmt.Sprintf("CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' WHEN %s->>? = '0' THEN '%s' ELSE  date_trunc('%s', to_timestamp(to_number((%s->>?)::text,'9999999999'))::timestamp)::text END AS %s",
			entityField, PropertyValueNone, entityField, PropertyValueNone, entityField, PropertyValueNone, groupProp.Granularity, entityField, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property, groupProp.Property}
	}
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

func hasNumericalGroupBy(groupProps []QueryGroupByProperty) bool {
	for _, groupByProp := range groupProps {
		if groupByProp.Type == U.PropertyTypeNumerical {
			return true
		}
	}
	return false
}

func appendNumericalBucketingSteps(qStmnt *string, groupProps []QueryGroupByProperty, refStepName, eventNameSelect string,
	isGroupByTimestamp bool, additionalSelectKeys string) (bucketedStepName, aggregateSelectKeys string, aggregateGroupBys, aggregateOrderBys []string) {
	bucketedStepName = "bucketed"
	bucketedSelect := "SELECT "
	if eventNameSelect != "" {
		bucketedSelect = bucketedSelect + eventNameSelect + ", "
	}
	var boundStepNames []string
	for _, gbp := range groupProps {
		groupKey := groupKeyByIndex(gbp.Index)
		if gbp.Type != U.PropertyTypeNumerical {
			bucketedSelect = bucketedSelect + groupKey + ", "
			aggregateGroupBys = append(aggregateGroupBys, groupKey)
			aggregateSelectKeys = aggregateSelectKeys + groupKey + ", "
			continue
		}

		// Adding _group_key_x_bounds step.
		// Buckets are formed using lower and upper bound as below:
		//		Bucket 1 as < lbound
		// 		Bucket 2 as >= lbound && < (lbound + diff / 8)
		// 		....
		// 		Bucket 10 as >= ubound
		// Mostly data is skewed towards first value which would mostly be default.
		// And since first bucket range is < lbound, first bucket mostly end up getting 0 values.
		// Adding very small 0.00001 to the lower bound so that while bucketing lbound comes in
		// first bucket instead of 2nd bucket.
		boundsStepName := groupKey + "_bounds"
		boundsStatement := fmt.Sprintf("SELECT percentile_disc(%.2f) WITHIN GROUP(ORDER BY %s::numeric) + 0.00001 AS lbound, "+
			"percentile_disc(%.2f) WITHIN GROUP(ORDER BY %s::numeric) AS ubound FROM %s "+
			"WHERE %s != '%s' AND %s != '' ", NumericalLowerBoundPercentile, groupKey, NumericalUpperBoundPercentile,
			groupKey, refStepName, groupKey, PropertyValueNone, groupKey)
		boundsStatement = as(boundsStepName, boundsStatement)
		*qStmnt = joinWithComma(*qStmnt, boundsStatement)

		// Preparing 'bucketed' step with changing $none to NaN for float conversion.
		noneToNaN := fmt.Sprintf("COALESCE(NULLIF(COALESCE(NULLIF(%s, '%s'), ''), ''), 'NaN') AS %s, ", groupKey, PropertyValueNone, groupKey)

		// Adding width_bucket for each record, keeping -1 for $none.
		bucketKey := groupKey + "_bucket"
		stepBucket := fmt.Sprintf("CASE WHEN %s = '%s' THEN -1 WHEN %s = '' THEN -1 ELSE width_bucket(%s::numeric, %s.lbound::numeric, "+
			"COALESCE(NULLIF(%s.ubound, %s.lbound), %s.ubound+1)::numeric, %d) END AS %s, ",
			groupKey, PropertyValueNone, groupKey, groupKey, boundsStepName, boundsStepName,
			boundsStepName, boundsStepName, NumericalGroupByBuckets-2, bucketKey)

		// Creating bucket string to be used in group by. Also, replacing NaN-Nan to $none.
		aggregateSelectKeys = aggregateSelectKeys + fmt.Sprintf(
			"COALESCE(NULLIF(concat(round(min(%s::numeric), 1), '%s', round(max(%s::numeric), 1)), 'NaN%sNaN'), '%s') AS %s, ",
			groupKey, NumericalGroupBySeparator, groupKey, NumericalGroupBySeparator, PropertyValueNone, groupKey)
		bucketedSelect = bucketedSelect + noneToNaN + stepBucket
		boundStepNames = append(boundStepNames, boundsStepName)
		aggregateGroupBys = append(aggregateGroupBys, bucketKey)
		aggregateOrderBys = append(aggregateOrderBys, bucketKey)
	}

	bucketedSelect = bucketedSelect + additionalSelectKeys
	if isGroupByTimestamp {
		bucketedSelect = joinWithComma(bucketedSelect, AliasDateTime)
	}
	bucketedSelect = bucketedSelect + " FROM " + refStepName
	if len(boundStepNames) > 0 {
		bucketedSelect = bucketedSelect + ", " + strings.Join(boundStepNames, ", ")
	}

	*qStmnt = joinWithComma(*qStmnt, as(bucketedStepName, bucketedSelect))
	return
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

// IsValidEventsQuery Validates query for Events class and returns corresponding errMsg.
func IsValidEventsQuery(query *Query) (bool, string) {
	if query.Class != QueryClassEvents {
		return false, "Invalid query class given"
	}

	if query.Type != QueryTypeEventsOccurrence &&
		query.Type != QueryTypeUniqueUsers {
		return false, "Invalid query type given"
	}

	if query.EventsCondition != EventCondAllGivenEvent &&
		query.EventsCondition != EventCondAnyGivenEvent &&
		query.EventsCondition != EventCondEachGivenEvent {
		return false, "Invalid events condition given"
	}

	errMsg, hasError := validateQueryProps(query)
	if hasError {
		return false, errMsg
	}
	return true, ""
}

// IsValidQuery Validates and returns errMsg which is used as response.
func IsValidQuery(query *Query) (bool, string) {
	if query.Type != QueryTypeEventsOccurrence &&
		query.Type != QueryTypeUniqueUsers {
		return false, "Invalid query type given"
	}

	if query.EventsCondition != EventCondAllGivenEvent &&
		query.EventsCondition != EventCondAnyGivenEvent &&
		query.EventsCondition != EventCondEachGivenEvent {
		return false, "Invalid events condition given"
	}

	errMsg, hasError := validateQueryProps(query)
	if hasError {
		return false, errMsg
	}
	return true, ""
}

func validateQueryProps(query *Query) (string, bool) {
	if len(query.EventsWithProperties) == 0 {
		return "No events to process", true
	}

	if query.From == 0 || query.To == 0 {
		return "Invalid query time range", true
	}

	if !isValidGroupByTimestamp(query.GetGroupByTimestamp()) {
		return "Invalid group by timestamp", true
	}
	return "", false
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

// GetBucketRangeForStartAndEnd Converts 2 - 2 range types to 2.
func GetBucketRangeForStartAndEnd(rangeStart, rangeEnd interface{}) string {
	if rangeStart == rangeEnd {
		return fmt.Sprintf("%v", rangeStart)
	}
	return fmt.Sprintf("%v%s%v", rangeStart, NumericalGroupBySeparator, rangeEnd)
}

// TODO (Anil) update this for v1/ each users count query
func sanitizeNumericalBucketRange(query *Query, rows [][]interface{}, indexToSanitize int) {
	for index, row := range rows {
		if query.Class == QueryClassFunnel && index == 0 {
			// For funnel queries, first row is $no_group query. Skip sanitization.
			continue
		}

		// Remove trailing .0 in start and end value of range.
		row[indexToSanitize] = trailingZeroRegex.ReplaceAllString(row[indexToSanitize].(string), "")

		// Change range with same start and end ex: 2 - 2 to just 2.
		if row[indexToSanitize] != PropertyValueNone {
			rowSplit := strings.Split(row[indexToSanitize].(string), NumericalGroupBySeparator)
			if rowSplit[0] == rowSplit[1] {
				row[indexToSanitize] = GetBucketRangeForStartAndEnd(rowSplit[0], rowSplit[1])
			}
		}
	}
}

// sanitizeNumericalBucketRanges Removes any .0 added to bucket ranges wherever possible.
func sanitizeNumericalBucketRanges(result *QueryResult, query *Query) {
	headerIndexMap := make(map[string][]int)
	for index, header := range result.Headers {
		// If same group by is added twice, it will appear twice in headers.
		// Keep as a list to sanitize both indexes.
		headerIndexMap[header] = append(headerIndexMap[header], index)
	}

	sanitizedProperties := make(map[string]bool)
	for _, gbp := range query.GroupByProperties {
		if gbp.Type == U.PropertyTypeNumerical {
			if _, sanitizedAlready := sanitizedProperties[gbp.Property]; sanitizedAlready {
				continue
			}
			indexesToSanitize := headerIndexMap[gbp.Property]
			for _, indexToSanitize := range indexesToSanitize {
				sanitizeNumericalBucketRange(query, result.Rows, indexToSanitize)
			}
			sanitizedProperties[gbp.Property] = true
		}
	}
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

func addQueryToResultMeta(result *QueryResult, query Query) {
	result.Meta.Query = query
}

func isValidFunnelQuery(query *Query) bool {
	return len(query.EventsWithProperties) <= 4
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
	case QueryClassEvents:
		var query QueryGroup
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
