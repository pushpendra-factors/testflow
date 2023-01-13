package memsql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var trailingZeroRegex = regexp.MustCompile(`\.0\b`)

var queryOps = map[string]string{
	model.EqualsOpStr:             model.EqualsOp,
	model.NotEqualOpStr:           model.NotEqualOp,
	model.GreaterThanOpStr:        ">",
	model.LesserThanOpStr:         "<",
	model.GreaterThanOrEqualOpStr: ">=",
	model.LesserThanOrEqualOpStr:  "<=",
	model.ContainsOpStr:           "RLIKE",
	model.NotContainsOpStr:        "NOT RLIKE",
}

func with(stmnt string) string {
	logFields := log.Fields{
		"stmnt": stmnt,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return fmt.Sprintf("WITH %s", stmnt)
}

func getOp(OpStr string, typeStr string) string {
	logFields := log.Fields{
		"op_str":   OpStr,
		"type_str": typeStr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if typeStr == U.PropertyTypeDateTime {
		return queryOps[model.EqualsOpStr]
	}
	v, ok := queryOps[OpStr]
	if !ok {
		log.Errorf("invalid query operator %s, using default", OpStr)
		return model.EqualsOp
	}

	return v
}

func getPropertyEntityField(projectID int64, groupProp model.QueryGroupByProperty) string {
	logFields := log.Fields{
		"project_id": projectID,
		"group_prop": groupProp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if groupProp.Entity == model.PropertyEntityUser {
		// Use event level user properties for event level group by.
		if isEventLevelGroupBy(groupProp) {
			return "events.user_properties"
		}

		return "users.properties"
	} else if groupProp.Entity == model.PropertyEntityEvent {
		return "events.properties"
	}

	return ""
}

// returns statement and parameters to pull content group values from event table
func getSelectSQLStmtForContentGroup(contentGroupNamesToDummyNamesMap map[string]string) (rStmnt string, rParams []interface{}, err error) {

	caseSelectStmt := "CASE WHEN JSON_EXTRACT_STRING(sessions.properties, ?) IS NULL THEN ? " +
		" WHEN JSON_EXTRACT_STRING(sessions.properties, ?) = '' THEN ? ELSE JSON_EXTRACT_STRING(sessions.properties, ?) END"
	for name, dummyName := range contentGroupNamesToDummyNamesMap {
		rStmnt = rStmnt + caseSelectStmt + " AS " + dummyName + " ,"
		qParams := []interface{}{name, model.PropertyValueNone, name, model.PropertyValueNone, name}
		rParams = append(rParams, qParams...)
	}

	return rStmnt, rParams, nil
}

func as(asName, asQuery string) string {
	logFields := log.Fields{
		"as_name":  asName,
		"as_query": asQuery,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return fmt.Sprintf("%s AS (%s)", asName, asQuery)
}

func appendStatement(x, y string) string {
	logFields := log.Fields{
		"x": x,
		"y": y,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return fmt.Sprintf("%s %s", x, y)
}

func isValidLogicalOp(op string) bool {
	logFields := log.Fields{
		"op": op,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return op == "AND" || op == "OR"
}

func buildWhereFromProperties(projectID int64, properties []model.QueryProperty,
	fromTimestamp int64) (rStmnt string, rParams []interface{}, err error) {
	logFields := log.Fields{
		"project_id":     projectID,
		"properties":     properties,
		"from_timestamp": fromTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	pLen := len(properties)
	if pLen == 0 {
		return rStmnt, rParams, nil
	}

	rParams = make([]interface{}, 0, 0)
	propertyToHasNoneFilter := model.GetPropertyToHasNoneFilter(properties)
	groupedProperties := model.GetPropertiesGrouped(properties)

	for indexOfGroup, currentGroupedProperties := range groupedProperties {
		var currentGroupStmnt, pStmnt string
		for indexOfProperty, p := range currentGroupedProperties {

			if p.LogicalOp == "" {
				p.LogicalOp = "AND"
			}

			if !isValidLogicalOp(p.LogicalOp) {
				return rStmnt, rParams, errors.New("invalid logical op on where condition")
			}
			pStmnt = ""
			hasNoneFilter := model.CheckIfMapHasNoneFilter(propertyToHasNoneFilter, p)
			propertyEntity := GetPropertyEntityFieldForFilter(p.Entity, fromTimestamp)
			propertyOp := getOp(p.Operator, p.Type)

			if p.Value != model.PropertyValueNone {
				if p.Type == U.PropertyTypeDateTime {
					var pParams []interface{}
					pStmnt, pParams, err = GetDateFilter(p, propertyEntity, p.Property)
					if err != nil {
						return pStmnt, pParams, err
					}
					rParams = append(rParams, pParams...)
				} else if p.Type == U.PropertyTypeNumerical {
					// convert to float for numerical properties.
					pStmnt = fmt.Sprintf("CASE WHEN JSON_GET_TYPE(JSON_EXTRACT_JSON(%s, ?)) = 'double' THEN  JSON_EXTRACT_DOUBLE(%s, ?) %s ? ELSE false END", propertyEntity, propertyEntity, propertyOp)
					rParams = append(rParams, p.Property, p.Property, p.Value)
				} else {
					// categorical property type.
					pValue := p.Value
					if p.Operator == model.ContainsOpStr {
						pStmnt = fmt.Sprintf("JSON_EXTRACT_STRING(%s, ?) %s ?", propertyEntity, propertyOp)
						rParams = append(rParams, p.Property, pValue)
					} else if !hasNoneFilter && p.Operator == model.NotContainsOpStr {
						pStmnt1 := fmt.Sprintf(" ( JSON_EXTRACT_STRING(%s, ?) %s ? ", propertyEntity, propertyOp)
						rParams = append(rParams, p.Property, pValue)
						pStmnt2 := fmt.Sprintf(" OR JSON_EXTRACT_STRING(%s, ?) = '' ", propertyEntity)
						rParams = append(rParams, p.Property)
						pStmnt3 := fmt.Sprintf(" OR JSON_EXTRACT_STRING(%s, ?) IS NULL ) ", propertyEntity)
						rParams = append(rParams, p.Property)
						pStmnt = pStmnt1 + pStmnt2 + pStmnt3
					} else if !hasNoneFilter && p.Operator == model.NotEqualOpStr {
						// PR: 2342 - This change is to allow empty ('') or NULL values during a filter of != value
						// ex: JSON_EXTRACT_STRING(events.properties, '$source') != 'google' OR JSON_EXTRACT_STRING(events.properties, '$source') = '' OR JSON_EXTRACT_STRING(events.properties, '$source') IS NULL
						pStmnt1 := fmt.Sprintf(" ( JSON_EXTRACT_STRING(%s, ?) %s ? ", propertyEntity, propertyOp)
						rParams = append(rParams, p.Property, pValue)
						pStmnt2 := fmt.Sprintf(" OR JSON_EXTRACT_STRING(%s, ?) = '' ", propertyEntity)
						rParams = append(rParams, p.Property)
						pStmnt3 := fmt.Sprintf(" OR JSON_EXTRACT_STRING(%s, ?) IS NULL ) ", propertyEntity)
						rParams = append(rParams, p.Property)
						pStmnt = pStmnt1 + pStmnt2 + pStmnt3
					} else {
						pStmnt = fmt.Sprintf("JSON_EXTRACT_STRING(%s, ?) %s ?", propertyEntity, propertyOp)
						rParams = append(rParams, p.Property, pValue)
					}
				}
			} else {
				// where condition for $none value.
				// var pStmnt string
				if propertyOp == model.EqualsOp || propertyOp == model.RLikeOp {
					// i.e: (NOT jsonb_exists(events.properties, 'property_name') OR events.properties->>'property_name'='')
					pStmnt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) IS NULL OR JSON_EXTRACT_STRING(%s, ?)='')", propertyEntity, propertyEntity)
				} else if propertyOp == model.NotEqualOp || propertyOp == model.NotRLikeOp {
					// i.e: (jsonb_exists(events.properties, 'property_name') AND events.properties->>'property_name'!='')
					pStmnt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) IS NOT NULL AND JSON_EXTRACT_STRING(%s, ?)!='')", propertyEntity, propertyEntity)
				} else {
					return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
				}
				rParams = append(rParams, p.Property, p.Property)
			}
			if indexOfProperty == 0 {
				currentGroupStmnt = pStmnt
			} else {
				currentGroupStmnt = fmt.Sprintf("%s %s %s", currentGroupStmnt, p.LogicalOp, pStmnt)
			}
		}
		if indexOfGroup == 0 {
			rStmnt = fmt.Sprintf("(%s)", currentGroupStmnt)
		} else {
			rStmnt = fmt.Sprintf("%s AND (%s)", rStmnt, currentGroupStmnt)
		}

	}

	return rStmnt, rParams, nil
}

// from = t1, to = t2
// before = < t2
// since = > t1
// between = t1 to t2
// notBetween = ~(t1 to t2)
// inPrev == inLast = t1 to t2
// notInPrev == notInLast = ~(t1 to t2)
// inCurr = >t1
// notinCurr = <t1
func GetDateFilter(qP model.QueryProperty, propertyEntity string, property string) (string, []interface{}, error) {
	logFields := log.Fields{
		"q_p":             qP,
		"property_entity": propertyEntity,
		"property":        property,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var stmt string
	var resultParams []interface{}
	dateTimeValue, err := model.DecodeDateTimePropertyValue(qP.Value)
	if err != nil {
		log.WithError(err).Error("Failed reading timestamp on user join query.")
		return "", nil, err
	}
	if qP.Operator == model.BeforeStr {
		stmt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) < ?)", propertyEntity)
		resultParams = append(resultParams, property, dateTimeValue.To)
	} else if qP.Operator == model.NotInCurrent {
		stmt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) < ?)", propertyEntity)
		resultParams = append(resultParams, property, dateTimeValue.From)
	} else if qP.Operator == model.SinceStr || qP.Operator == model.InCurrent {
		stmt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) >= ?)", propertyEntity)
		resultParams = append(resultParams, property, dateTimeValue.From)
	} else if qP.Operator == model.EqualsOpStr || qP.Operator == model.BetweenStr || qP.Operator == model.InPrevious || qP.Operator == model.InLastStr { // equals - Backward Compatible of Between
		stmt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) BETWEEN ? AND ?)", propertyEntity)
		resultParams = append(resultParams, property, dateTimeValue.From, dateTimeValue.To)
	} else if qP.Operator == model.NotInBetweenStr || qP.Operator == model.NotInPrevious || qP.Operator == model.NotInLastStr {
		stmt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s, ?) NOT BETWEEN ? AND ?)", propertyEntity)
		resultParams = append(resultParams, property, dateTimeValue.From, dateTimeValue.To)
	}
	return stmt, resultParams, nil
}

func getEventsFilterJoinStatement(projectID int64,
	eventLevelProperties []model.QueryProperty, fromTimestamp int64) string {
	logFields := log.Fields{
		"project_id":             projectID,
		"event_level_properties": eventLevelProperties,
		"from_timestamp":         fromTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(eventLevelProperties) == 0 {
		return ""
	}

	if !C.UseEventsFilterPropertiesOptimisedLogic(fromTimestamp) {
		return ""
	}

	joinStmnt := " " + "LEFT JOIN event_properties_json ON events.id = event_properties_json.id AND events.user_id = event_properties_json.user_id"
	joinStmnt = joinStmnt + " " + fmt.Sprintf("AND events.project_id=%d", projectID)
	return joinStmnt
}

func getUsersFilterJoinStatement(projectID int64,
	globalUserProperties []model.QueryProperty) string {
	logFields := log.Fields{
		"project_id":             projectID,
		"global_user_properties": globalUserProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(globalUserProperties) == 0 {
		return ""
	}

	if !C.UseUsersFilterPropertiesOptimisedLogic() {
		return ""
	}

	joinStmnt := " " + "LEFT JOIN user_properties_json ON users.id = user_properties_json.id"
	joinStmnt = joinStmnt + " " + fmt.Sprintf("AND users.project_id = %d", projectID)
	return joinStmnt
}

// returns SQL query condition to address conditions only on events.properties
func getFilterSQLStmtForEventProperties(projectID int64, properties []model.QueryProperty,
	fromTimestamp int64) (rStmnt string, rParams []interface{}, joinStmnt string, err error) {
	logFields := log.Fields{
		"project_id":      projectID,
		"properties":      properties,
		"from_time_stamp": fromTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var filteredProperty []model.QueryProperty
	for _, p := range properties {
		if p.Entity == model.PropertyEntityEvent {
			filteredProperty = append(filteredProperty, p)
		}
	}

	wStmt, wParams, err := buildWhereFromProperties(projectID, filteredProperty, fromTimestamp)
	if err != nil {
		return "", nil, "", err
	}

	return wStmt, wParams, getEventsFilterJoinStatement(projectID, filteredProperty, fromTimestamp), nil
}

// returns SQL query condition to address conditions for Users.properties
func getFilterSQLStmtForLatestUserProperties(projectID int64,
	properties []model.QueryProperty, fromTimestamp int64) (
	rStmnt string, rParams []interface{}, err error) {
	logFields := log.Fields{
		"project_id":      projectID,
		"properties":      properties,
		"from_time_stamp": fromTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var filteredProperty []model.QueryProperty
	for _, p := range properties {
		if p.Entity == model.PropertyEntityUserGlobal {
			filteredProperty = append(filteredProperty, p)
		}
	}

	wStmt, wParams, err := buildWhereFromProperties(projectID, filteredProperty, fromTimestamp)
	if err != nil {
		return "", nil, err
	}

	return wStmt, wParams, nil
}

// returns SQL query condition to address conditions only on user_properties.properties
func getFilterSQLStmtForUserProperties(projectID int64,
	properties []model.QueryProperty, fromTimestamp int64) (
	rStmnt string, rParams []interface{}, joinStmnt string, err error) {
	logFields := log.Fields{
		"project_id":      projectID,
		"properties":      properties,
		"from_time_stamp": fromTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var filteredProperty []model.QueryProperty
	for _, p := range properties {
		if p.Entity == model.PropertyEntityUser {
			filteredProperty = append(filteredProperty, p)
		}
	}

	wStmt, wParams, err := buildWhereFromProperties(projectID, filteredProperty, fromTimestamp)
	if err != nil {
		return "", nil, "", err
	}

	return wStmt, wParams, getEventsFilterJoinStatement(projectID, filteredProperty, fromTimestamp), nil
}

// Alias for group by properties gk_1, gk_2.
func groupKeyByIndex(i int) string {
	logFields := log.Fields{
		"i": i,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return fmt.Sprintf("%s%d", model.GroupKeyPrefix, i)
}

func stepNameByIndex(i int) string {
	logFields := log.Fields{
		"i": i,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return fmt.Sprintf("%s%d", model.StepPrefix, i)
}

func GetPropertyEntityFieldForFilter(entityName string, fromTimestamp int64) string {
	logFields := log.Fields{
		"entity_name":     entityName,
		"from_time_stamp": fromTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	switch entityName {

	case model.PropertyEntityUser:
		if !C.UseEventsFilterPropertiesOptimisedLogic(fromTimestamp) {
			return model.GetPropertyEntityFieldForFilter(entityName)
		}

		return "event_properties_json.user_properties_json"
	case model.PropertyEntityEvent:
		if !C.UseEventsFilterPropertiesOptimisedLogic(fromTimestamp) {
			return model.GetPropertyEntityFieldForFilter(entityName)
		}

		return "event_properties_json.properties_json"
	case model.PropertyEntityUserGlobal:
		if !C.UseUsersFilterPropertiesOptimisedLogic() {
			return model.GetPropertyEntityFieldForFilter(entityName)
		}

		return "user_properties_json.properties_json"
	}

	return ""
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
func getNoneHandledGroupBySelect(projectID int64, groupProp model.QueryGroupByProperty, groupKey string, timezoneString string) (string, []interface{}) {
	logFields := log.Fields{
		"project_id":      projectID,
		"group_prop":      groupProp,
		"group_key":       groupKey,
		"timezone_string": timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	entityField := getPropertyEntityField(projectID, groupProp)
	var groupSelect string
	groupSelectParams := make([]interface{}, 0)
	if groupProp.Type != U.PropertyTypeDateTime {
		groupSelect = fmt.Sprintf("CASE WHEN JSON_EXTRACT_STRING(%s, ?) IS NULL THEN '%s' WHEN JSON_EXTRACT_STRING(%s, ?) = '' THEN '%s' ELSE JSON_EXTRACT_STRING(%s, ?) END AS %s",
			entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, entityField, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	} else {
		propertyName := "JSON_EXTRACT_STRING(" + entityField + ", ?)"
		timestampStr := getSelectTimestampByTypeAndPropertyName(groupProp.Granularity, propertyName, timezoneString)
		groupSelect = fmt.Sprintf("CASE WHEN JSON_EXTRACT_STRING(%s, ?) IS NULL THEN '%s' WHEN JSON_EXTRACT_STRING(%s, ?) = '' THEN '%s' WHEN JSON_EXTRACT_STRING(%s, ?) = '0' THEN '%s' ELSE %s END AS %s",
			entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, timestampStr, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property, groupProp.Property}
	}
	return groupSelect, groupSelectParams
}
func getNoneHandledGroupBySelectWithFirst(projectID int64, groupProp model.QueryGroupByProperty, groupKey string, timezoneString string) (string, []interface{}) {
	logFields := log.Fields{
		"project_id":      projectID,
		"group_prop":      groupProp,
		"group_key":       groupKey,
		"timezone_string": timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	entityField := getPropertyEntityField(projectID, groupProp)
	var groupSelect string
	groupSelectParams := make([]interface{}, 0)
	if groupProp.Type != U.PropertyTypeDateTime {
		groupSelect = fmt.Sprintf("CASE WHEN JSON_EXTRACT_STRING(FIRST(%s, FROM_UNIXTIME(events.timestamp)), ?) IS NULL THEN '%s' WHEN JSON_EXTRACT_STRING(FIRST(%s, FROM_UNIXTIME(events.timestamp)), ?) = '' THEN '%s' ELSE JSON_EXTRACT_STRING(FIRST(%s, FROM_UNIXTIME(events.timestamp)), ?) END AS %s",
			entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, entityField, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	} else {
		propertyName := "JSON_EXTRACT_STRING(FIRST(" + entityField + ", FROM_UNIXTIME(events.timestamp)), ?)"
		timestampStr := getSelectTimestampByTypeAndPropertyName(groupProp.Granularity, propertyName, timezoneString)
		groupSelect = fmt.Sprintf("CASE WHEN JSON_EXTRACT_STRING(FIRST(%s, FROM_UNIXTIME(events.timestamp)), ?) IS NULL THEN '%s' WHEN JSON_EXTRACT_STRING(FIRST(%s, FROM_UNIXTIME(events.timestamp)), ?) = '' THEN '%s' WHEN JSON_EXTRACT_STRING(FIRST(%s, FROM_UNIXTIME(events.timestamp)), ?) = '0' THEN '%s' ELSE %s END AS %s",
			entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, timestampStr, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property, groupProp.Property}
	}
	return groupSelect, groupSelectParams
}

// groupBySelect: user_properties.properties->>'age' as gk_1, events.properties->>'category' as gk_2
// groupByKeys: gk_1, gk_2
// How to use?
// select user_properties.properties->>'age' as gk_1, events.properties->>'category' as gk_2 from events
// group by gk_1, gk_2
func buildGroupKeys(projectID int64, groupProps []model.QueryGroupByProperty, timezoneString string) (groupSelect string,
	groupSelectParams []interface{}, groupKeys string) {
	logFields := log.Fields{
		"project_id":      projectID,
		"group_props":     groupProps,
		"timezone_string": timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	groupSelectParams = make([]interface{}, 0)

	for i, v := range groupProps {
		// Order of group is preserved as received.
		gKey := groupKeyByIndex(v.Index)
		noneHandledSelect, noneHandledSelectParams := getNoneHandledGroupBySelect(projectID, v, gKey, timezoneString)
		groupSelect = groupSelect + noneHandledSelect
		groupKeys = groupKeys + gKey
		if i < len(groupProps)-1 {
			groupSelect = groupSelect + ", "
			groupKeys = groupKeys + ", "
		}
		groupSelectParams = append(groupSelectParams, noneHandledSelectParams...)
	}

	return groupSelect, groupSelectParams, groupKeys
}
func buildGroupKeysWithFirst(projectID int64, groupProps []model.QueryGroupByProperty, timezoneString string) (groupSelect string,
	groupSelectParams []interface{}, groupKeys string) {
	logFields := log.Fields{
		"project_id":      projectID,
		"group_props":     groupProps,
		"timezone_string": timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	groupSelectParams = make([]interface{}, 0)

	for i, v := range groupProps {
		// Order of group is preserved as received.
		gKey := groupKeyByIndex(v.Index)
		noneHandledSelect, noneHandledSelectParams := getNoneHandledGroupBySelectWithFirst(projectID, v, gKey, timezoneString)
		groupSelect = groupSelect + noneHandledSelect
		groupKeys = groupKeys + gKey
		if i < len(groupProps)-1 {
			groupSelect = groupSelect + ", "
			groupKeys = groupKeys + ", "
		}
		groupSelectParams = append(groupSelectParams, noneHandledSelectParams...)
	}

	return groupSelect, groupSelectParams, groupKeys
}

func isGroupByTypeWithBuckets(groupProps []model.QueryGroupByProperty) bool {
	logFields := log.Fields{
		"group_props": groupProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, groupByProp := range groupProps {
		if groupByProp.Type == U.PropertyTypeNumerical {
			if groupByProp.GroupByType == "" || groupByProp.GroupByType == model.GroupByTypeWithBuckets {
				// Empty condition for backward compatibility as existing queries will not have GroupByType.
				return true
			}
		}
	}
	return false
}

func appendNumericalBucketingSteps(isAggregateOnProperty bool, qStmnt *string, qParams *[]interface{}, groupProps []model.QueryGroupByProperty, refStepName, eventNameSelect string,
	isGroupByTimestamp bool, additionalSelectKeys string) (bucketedStepName, aggregateSelectKeys string, aggregateGroupBys, aggregateOrderBys []string) {
	logFields := log.Fields{
		"q_stmnt":                qStmnt,
		"q_params":               qParams,
		"group_props":            groupProps,
		"ref_step_name":          refStepName,
		"event_name_select":      eventNameSelect,
		"is_group_by_timestamp":  isGroupByTimestamp,
		"additional_select_keys": additionalSelectKeys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	bucketedStepName = "bucketed"
	bucketedSelect := "SELECT "
	if eventNameSelect != "" {
		bucketedSelect = bucketedSelect + eventNameSelect + ", "
	}
	if isAggregateOnProperty {
		bucketedSelect = bucketedSelect + model.AliasAggr + ", "
	}

	var boundStepNames []string
	var bucketedNumericValueFilter []string
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
		boundsStatement := fmt.Sprintf("SELECT percentile_disc(%.2f) WITHIN GROUP(ORDER BY CONVERT(%s, DECIMAL(20,2))) + 0.00001 AS lbound, "+
			"percentile_disc(%.2f) WITHIN GROUP(ORDER BY CONVERT(%s, DECIMAL(20,2))) AS ubound FROM %s "+
			"WHERE %s != '%s' AND %s != '' AND %s RLIKE ? ", model.NumericalLowerBoundPercentile, groupKey, model.NumericalUpperBoundPercentile,
			groupKey, refStepName, groupKey, model.PropertyValueNone, groupKey, groupKey)
		*qParams = append(*qParams, model.NumericalValuePostgresRegex)
		boundsStatement = as(boundsStepName, boundsStatement)
		*qStmnt = joinWithComma(*qStmnt, boundsStatement)

		// Preparing 'bucketed' step with changing $none to NaN for float conversion.
		noneToNaN := fmt.Sprintf("COALESCE(NULLIF(COALESCE(NULLIF(%s, '%s'), ''), ''), 'NaN') AS %s, ",
			groupKey, model.PropertyValueNone, groupKey)

		// Adding width_bucket for each record, keeping -1 for $none.
		// TODO(prateek): Add a UDF if required in case NTILE is working well.
		bucketKey := groupKey + "_bucket"
		stepBucket := fmt.Sprintf("CASE WHEN %s = '%s' THEN -1 WHEN %s = '' THEN -1 WHEN CONVERT(%s, DECIMAL(20,2)) < %s.lbound THEN 0"+
			" WHEN CONVERT(%s, DECIMAL(20,2)) >= %s.ubound THEN %d ELSE NTILE(%d) OVER (ORDER BY CONVERT(%s, DECIMAL(20,2)) ASC) END AS %s, ",
			groupKey, model.PropertyValueNone, groupKey, groupKey, boundsStepName,
			groupKey, boundsStepName, model.NumericalGroupByBuckets-1, model.NumericalGroupByBuckets-2, groupKey, bucketKey)

		// Creating bucket string to be used in group by. Also, replacing NaN-Nan to $none.
		aggregateSelectKeys = aggregateSelectKeys + fmt.Sprintf(
			"COALESCE(NULLIF(concat(CASE WHEN %s = 'NaN' THEN 'NaN' ELSE round(min(CONVERT(%s, DECIMAL(20,2))), 1) END, '%s', CASE WHEN %s = 'NaN' THEN 'NaN' ELSE round(max(CONVERT(%s, DECIMAL(20,2))), 1) END), 'NaN%sNaN'), '%s') AS %s, ",
			groupKey, groupKey, model.NumericalGroupBySeparator, groupKey, groupKey, model.NumericalGroupBySeparator, model.PropertyValueNone, groupKey)
		bucketedSelect = bucketedSelect + noneToNaN + stepBucket
		boundStepNames = append(boundStepNames, boundsStepName)
		aggregateGroupBys = append(aggregateGroupBys, bucketKey)
		aggregateOrderBys = append(aggregateOrderBys, bucketKey)
		bucketedNumericValueFilter = append(bucketedNumericValueFilter,
			fmt.Sprintf("%s RLIKE ?", groupKey))
		*qParams = append(*qParams, model.NumericalValuePostgresRegex)
	}

	bucketedSelect = bucketedSelect + additionalSelectKeys
	if isGroupByTimestamp {
		bucketedSelect = joinWithComma(bucketedSelect, model.AliasDateTime)
	}
	bucketedSelect = bucketedSelect + " FROM " + refStepName
	if len(boundStepNames) > 0 {
		bucketedSelect = bucketedSelect + ", " + strings.Join(boundStepNames, ", ")
	}
	bucketedSelect = bucketedSelect + " WHERE " + strings.Join(bucketedNumericValueFilter, " AND ")

	*qStmnt = joinWithComma(*qStmnt, as(bucketedStepName, bucketedSelect))
	return
}

// builds group keys with step of corresponding user given event name.
// i.e step_0.__group_key_0, step_1.group_key_1
func buildEventGroupKeysWithStep(groupProps []model.QueryGroupByProperty,
	ewps []model.QueryEventWithProperties) (groupKeys string) {
	logFields := log.Fields{
		"group_props": groupProps,
		"ewps":        ewps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	eventLevelGroupBys, _ := separateEventLevelGroupBys(groupProps)

	for _, gp := range eventLevelGroupBys {
		groupKey := fmt.Sprintf("%s.%s", stepNameByIndex(gp.EventNameIndex-1),
			groupKeyByIndex(gp.Index))
		groupKeys = joinWithComma(groupKeys, groupKey)
	}

	return groupKeys
}

func replaceTableWithViewForEventsAndUsers(stmnt, viewName string) string {
	stmnt = strings.ReplaceAll(stmnt, "users.", viewName+".")
	stmnt = strings.ReplaceAll(stmnt, "events.", viewName+".")
	return stmnt
}

// addFilterEventsWithPropsQuery Adds a step of events filter with QueryEventWithProperties.
// WITH step_0_names AS (SELECT id, project_id, name FROM event_names WHERE project_id='20426' AND name='$session') ,
// step_0 AS (SELECT DISTINCT ON(coal_user_id) COALESCE(users.customer_user_id,events.user_id) as coal_user_id,
// events.user_id as event_user_id , '0_$session'::text AS event_name FROM events JOIN users ON events.user_id=users.id
// AND users.project_id = '20426' WHERE events.project_id='20426' AND timestamp>='1583001004' AND timestamp<='1585679399'
// AND events.event_name_id IN (SELECT id FROM step_0_names WHERE project_id='20426' AND name='$session') AND
// ( (( JSON_EXTRACT_STRING(events.properties, '$source') != 'google' OR JSON_EXTRACT_STRING(events.properties, '$source') = '' OR JSON_EXTRACT_STRING(events.properties, '$source') IS NULL ) AND
// ( JSON_EXTRACT_STRING(events.properties, '$source') != 'facebook' OR JSON_EXTRACT_STRING(events.properties, '$source') = '' OR JSON_EXTRACT_STRING(events.properties, '$source') IS NULL )) )
// ORDER BY coal_user_id, events.timestamp ASC) , each_users_union AS (SELECT step_0.event_user_id, step_0.coal_user_id, step_0.event_name, FROM step_0)
// SELECT event_name, _group_key_0, _group_key_1, COUNT(DISTINCT(coal_user_id)) AS count FROM each_users_union GROUP BY event_name ,
// _group_key_0, _group_key_1 ORDER BY count DESC LIMIT 100000;
func addFilterEventsWithPropsQuery(projectId int64, qStmnt *string, qParams *[]interface{},
	qep model.QueryEventWithProperties, from int64, to int64, fromStr string,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string, groupBy string, orderBy string, globalUserFilter []model.QueryProperty) error {

	logFields := log.Fields{
		"project_id":         projectId,
		"q_stmnt":            qStmnt,
		"q_params":           qParams,
		"que":                qep,
		"from":               from,
		"to":                 to,
		"from_str":           fromStr,
		"step_name":          stepName,
		"add_selec_stmnt":    addSelecStmnt,
		"add_select_params":  addSelectParams,
		"add_joint_stmnt":    addJoinStmnt,
		"group_by":           groupBy,
		"order_by":           orderBy,
		"global_user_filter": globalUserFilter,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if (from == 0 && fromStr == "") || to == 0 {
		return errors.New("invalid timerange on events filter")
	}

	if addSelecStmnt == "" {
		return errors.New("invalid select on events filter")
	}

	eventFilterJoinStmnt := ""
	eventFilterJoinStmnt = getEventsFilterJoinStatement(projectId, qep.Properties, from)

	rStmnt := "SELECT " + addSelecStmnt + " FROM events" + " " + eventFilterJoinStmnt + " " + addJoinStmnt
	var fromTimestamp string
	if from > 0 {
		fromTimestamp = "?"
	} else if fromStr != "" {
		fromTimestamp = fromStr // allows from_timestamp from another step.
	}

	var eventNamesCacheStmnt string
	eventNamesRef := "event_names"
	skipEventNameStep := C.SkipEventNameStepByProjectID(projectId)
	if stepName != "" && !skipEventNameStep {
		eventNamesRef = fmt.Sprintf("%s_names", stepName)
		eventNamesCacheStmnt = as(eventNamesRef, "SELECT id, project_id, name FROM event_names WHERE project_id=? AND name=?")
		*qParams = append(*qParams, projectId, qep.Name)
	}

	whereCond := fmt.Sprintf("WHERE events.project_id=? AND timestamp>=%s AND timestamp<=?", fromTimestamp)
	// select id of event_names from names step.
	if !skipEventNameStep {
		whereCond = whereCond + fmt.Sprintf(" "+"AND events.event_name_id IN (SELECT id FROM %s WHERE project_id=? AND name=?)", eventNamesRef)
	} else {

		eventNameIDsFilter := "events.event_name_id = ?"
		if len(qep.EventNameIDs) > 1 {
			for range qep.EventNameIDs[1:] {
				eventNameIDsFilter += " " + "OR" + " " + "events.event_name_id = ?"
			}
		}

		whereCond = whereCond + " " + "AND" + " " + " ( " + eventNameIDsFilter + " ) "
	}

	rStmnt = appendStatement(rStmnt, whereCond)

	// adds params in order of '?'.
	if addSelecStmnt != "" && addSelectParams != nil {
		*qParams = append(*qParams, addSelectParams...)
	}
	*qParams = append(*qParams, projectId)
	if from > 0 {
		*qParams = append(*qParams, from)
	}

	*qParams = append(*qParams, to)
	if !skipEventNameStep {
		*qParams = append(*qParams, projectId, qep.Name)
	} else {
		*qParams = append(*qParams, qep.EventNameIDs...)
	}

	// applying global user filter
	gupStmt := ""
	var gupParam []interface{}
	var err error
	if globalUserFilter != nil && len(globalUserFilter) != 0 {
		// add user filter
		gupStmt, gupParam, err = getFilterSQLStmtForLatestUserProperties(
			projectId, globalUserFilter, from)

		if err != nil {
			return errors.New("invalid user properties for global filter")
		}
		rStmnt = rStmnt + " AND " + gupStmt

		*qParams = append(*qParams, gupParam...)
	}

	// mergeCond for whereProperties can also be 'OR'.
	wStmnt, wParams, err := buildWhereFromProperties(projectId, qep.Properties, from)
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

	if !skipEventNameStep && eventNamesCacheStmnt != "" {
		rStmnt = joinWithComma(eventNamesCacheStmnt, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)

	return nil
}

// addFilterEventsWithPropsQuery Adds a step of events filter with QueryEventWithProperties.
// step_0 AS (SELECT COALESCE(step_0_event_users.customer_user_id,step_0_event_users.user_id) as coal_user_id,
// FIRST(step_0_event_users.user_id, FROM_UNIXTIME(step_0_event_users.timestamp)) as user_id,
// FIRST(step_0_event_users.timestamp, FROM_UNIXTIME(step_0_event_users.timestamp)) as timestamp, 1 as step_0
// FROM (SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp,
// events.properties as event_properties, events.user_properties as event_user_properties,
// users.customer_user_id, users.properties as global_user_properties
// FROM events JOIN users ON events.user_id=users.id AND users.project_id = 2  WHERE events.project_id=2 AND
// timestamp>=1656844268 AND timestamp<=1657708269 AND  ( events.event_name_id = 4862cb39-8749-40c6-988a-c0a08a2d796b )
// LIMIT 10000000000) step_0_event_users WHERE JSON_EXTRACT_STRING(step_0_event_users.event_properties, '$source') = 'google'
// GROUP BY coal_user_id)
func addFilterEventsWithPropsQueryV2(projectId int64, qStmnt *string, qParams *[]interface{},
	qep model.QueryEventWithProperties, from int64, to int64, fromStr string,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string, groupBy string, orderBy string, globalUserFilter []model.QueryProperty) error {

	logFields := log.Fields{
		"project_id":         projectId,
		"q_stmnt":            qStmnt,
		"q_params":           qParams,
		"que":                qep,
		"from":               from,
		"to":                 to,
		"from_str":           fromStr,
		"step_name":          stepName,
		"add_selec_stmnt":    addSelecStmnt,
		"add_select_params":  addSelectParams,
		"add_joint_stmnt":    addJoinStmnt,
		"group_by":           groupBy,
		"order_by":           orderBy,
		"global_user_filter": globalUserFilter,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if (from == 0 && fromStr == "") || to == 0 {
		return errors.New("invalid timerange on events filter")
	}

	if addSelecStmnt == "" {
		return errors.New("invalid select on events filter")
	}

	eventsWrapSelect := "events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp"
	if strings.Contains(addSelecStmnt, "events.session_id") {
		eventsWrapSelect = joinWithComma(eventsWrapSelect, "events.session_id")
	}

	eventsWrapSelect = joinWithComma(eventsWrapSelect, "events.properties as event_properties, events.user_properties as event_user_properties")
	if addJoinStmnt != "" {
		eventsWrapSelect = joinWithComma(eventsWrapSelect, "users.customer_user_id, users.properties as global_user_properties")
		// For the special case of groups join, where events.user_id
		// cannot be used as replacement of users.id.
		if strings.Contains(addSelecStmnt, "users.users_user_id") {
			eventsWrapSelect = joinWithComma(eventsWrapSelect, "users.id as users_user_id")
		}
	}

	eventsWrapStmnt := "SELECT " + eventsWrapSelect + " FROM events" + " " + addJoinStmnt

	var fromTimestamp string
	if from > 0 {
		fromTimestamp = "?"
	} else if fromStr != "" {
		fromTimestamp = fromStr // allows from_timestamp from another step.
	}
	// Non-JSON where conditions except event_names.
	eventsWrapWhereCondition := fmt.Sprintf("WHERE events.project_id=? AND timestamp>=%s AND timestamp<=?", fromTimestamp)

	// Building event_names condition and appending.
	var eventNamesCacheStmnt string
	eventNamesRef := "event_names"
	skipEventNameStep := C.SkipEventNameStepByProjectID(projectId)
	if stepName != "" && !skipEventNameStep {
		eventNamesRef = fmt.Sprintf("%s_names", stepName)
		eventNamesCacheStmnt = as(eventNamesRef, "SELECT id, project_id, name FROM event_names WHERE project_id=? AND name=?")
		*qParams = append(*qParams, projectId, qep.Name)
	}
	// select id of event_names from names step.
	if !skipEventNameStep {
		eventsWrapWhereCondition = eventsWrapWhereCondition + fmt.Sprintf(" "+"AND events.event_name_id IN (SELECT id FROM %s WHERE project_id=? AND name=?)", eventNamesRef)
	} else {
		eventNameIDsFilter := "events.event_name_id = ?"
		if len(qep.EventNameIDs) > 1 {
			for range qep.EventNameIDs[1:] {
				eventNameIDsFilter += " " + "OR" + " " + "events.event_name_id = ?"
			}
		}

		eventsWrapWhereCondition = eventsWrapWhereCondition + " " + "AND" + " " + " ( " + eventNameIDsFilter + " ) "
	}
	eventsWrapStmnt = appendStatement(eventsWrapStmnt, eventsWrapWhereCondition)

	eventsWrapViewName := fmt.Sprintf("%s_event_users_view", stepName)
	// Builds `(SELECT xxx FROM events [JOIN users] WHERE <Non-JSON filters...> LIMIT 1000000000) step_0_event_users_view`
	eventsWrapView := fmt.Sprintf("(%s LIMIT %d) %s", eventsWrapStmnt, model.FilterOptLimit, eventsWrapViewName)

	// Change properties column for the view.
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "users.properties", eventsWrapViewName+".global_user_properties")
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "events.properties", eventsWrapViewName+".event_properties")
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "events.user_properties", eventsWrapViewName+".event_user_properties")
	// Use view instead of events and users table.
	addSelecStmnt = replaceTableWithViewForEventsAndUsers(addSelecStmnt, eventsWrapViewName)

	rStmnt := "SELECT " + addSelecStmnt + " FROM " + " " + eventsWrapView

	// adds params in order of '?'.
	if addSelecStmnt != "" && addSelectParams != nil {
		*qParams = append(*qParams, addSelectParams...)
	}
	*qParams = append(*qParams, projectId)
	if from > 0 {
		*qParams = append(*qParams, from)
	}

	*qParams = append(*qParams, to)
	if !skipEventNameStep {
		*qParams = append(*qParams, projectId, qep.Name)
	} else {
		*qParams = append(*qParams, qep.EventNameIDs...)
	}

	// Applying global user filter
	latestUserPropFilterStmnt := ""
	var latestUserPropFilterParam []interface{}
	var err error

	isWhereAdded := false
	hasGlobalUserProperties := globalUserFilter != nil && len(globalUserFilter) != 0
	if hasGlobalUserProperties {
		// add user filter
		latestUserPropFilterStmnt, latestUserPropFilterParam, err = getFilterSQLStmtForLatestUserProperties(
			projectId, globalUserFilter, from)
		if err != nil {
			return errors.New("invalid user properties for global filter")
		}

		latestUserPropFilterStmnt = strings.ReplaceAll(latestUserPropFilterStmnt,
			"users.properties", eventsWrapViewName+".global_user_properties")

		rStmnt = rStmnt + " WHERE " + latestUserPropFilterStmnt
		isWhereAdded = true
		*qParams = append(*qParams, latestUserPropFilterParam...)
	}

	// mergeCond for whereProperties can also be 'OR'.
	wStmnt, wParams, err := buildWhereFromProperties(projectId, qep.Properties, from)
	if err != nil {
		return err
	}

	if wStmnt != "" {
		conditionKeyword := "WHERE"
		if isWhereAdded {
			conditionKeyword = "AND"
		}

		wStmnt = strings.ReplaceAll(wStmnt, "events.properties", eventsWrapViewName+".event_properties")
		wStmnt = strings.ReplaceAll(wStmnt, "events.user_properties", eventsWrapViewName+".event_user_properties")

		rStmnt = rStmnt + " " + conditionKeyword + " " + fmt.Sprintf("( %s )", wStmnt)
		*qParams = append(*qParams, wParams...)
	}

	if groupBy != "" {
		groupBy = replaceTableWithViewForEventsAndUsers(groupBy, eventsWrapViewName)
		rStmnt = fmt.Sprintf("%s GROUP BY %s", rStmnt, groupBy)
	}

	if orderBy != "" {
		orderBy = replaceTableWithViewForEventsAndUsers(orderBy, eventsWrapViewName)
		rStmnt = fmt.Sprintf("%s ORDER BY %s", rStmnt, orderBy)
	}

	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
	}

	if !skipEventNameStep && eventNamesCacheStmnt != "" {
		rStmnt = joinWithComma(eventNamesCacheStmnt, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)

	return nil
}

// addFilterEventsWithPropsQueryV3 Adds a step of events filter with QueryEventWithProperties.
// step_0 AS (SELECT COALESCE(step_0_event_users.customer_user_id,step_0_event_users.user_id) as coal_user_id,
// FIRST(step_0_event_users.user_id, FROM_UNIXTIME(step_0_event_users.timestamp)) as user_id,
// FIRST(step_0_event_users.timestamp, FROM_UNIXTIME(step_0_event_users.timestamp)) as timestamp, 1 as step_0
// FROM (SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp,
// events.properties as event_properties, events.user_properties as event_user_properties,
// users.customer_user_id, users.properties as global_user_properties
// FROM events JOIN users ON events.user_id=users.id AND users.project_id = 2  WHERE events.project_id=2 AND
// timestamp>=1656844268 AND timestamp<=1657708269 AND  ( events.event_name_id = 4862cb39-8749-40c6-988a-c0a08a2d796b )
// LIMIT 10000000000) step_0_event_users WHERE JSON_EXTRACT_STRING(step_0_event_users.event_properties, '$source') = 'google'
// GROUP BY coal_user_id)
func addFilterEventsWithPropsQueryV3(projectId int64, qStmnt *string, qParams *[]interface{},
	qep model.QueryEventWithProperties, from int64, to int64, fromStr string,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string, groupBy string, orderBy string, globalUserFilter []model.QueryProperty) error {

	logFields := log.Fields{
		"project_id":         projectId,
		"q_stmnt":            qStmnt,
		"q_params":           qParams,
		"que":                qep,
		"from":               from,
		"to":                 to,
		"from_str":           fromStr,
		"step_name":          stepName,
		"add_selec_stmnt":    addSelecStmnt,
		"add_select_params":  addSelectParams,
		"add_joint_stmnt":    addJoinStmnt,
		"group_by":           groupBy,
		"order_by":           orderBy,
		"global_user_filter": globalUserFilter,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if (from == 0 && fromStr == "") || to == 0 {
		return errors.New("invalid timerange on events filter")
	}

	if addSelecStmnt == "" {
		return errors.New("invalid select on events filter")
	}

	eventsWrapSelect := "events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp"
	if strings.Contains(addSelecStmnt, "events.session_id") {
		eventsWrapSelect = joinWithComma(eventsWrapSelect, "events.session_id")
	}

	userGroupColumn := ""
	usersUserGroupColumn := ""
	eventsWrapSelect = joinWithComma(eventsWrapSelect, "events.properties as event_properties, events.user_properties as event_user_properties")
	if addJoinStmnt != "" {
		if !strings.Contains(addJoinStmnt, "user_groups") && !strings.Contains(addJoinStmnt, "group_users") {
			eventsWrapSelect = joinWithComma(eventsWrapSelect, "users.customer_user_id, users.properties as global_user_properties")
			// For the special case of groups join, where events.user_id
			// cannot be used as replacement of users.id.
			if strings.Contains(addSelecStmnt, "users.users_user_id") {
				eventsWrapSelect = joinWithComma(eventsWrapSelect, "users.id as users_user_id")
			}
		} else {
			if strings.Contains(addJoinStmnt, "user_groups") {
				userGroupColumn = model.GetQueryGroupUserID(addSelecStmnt)
				eventsWrapSelect = joinWithComma(eventsWrapSelect, fmt.Sprintf("%s as group_user_id", userGroupColumn))
				usersUserGroupColumn = model.GetQueryUserGroupUserID(addSelecStmnt)
				eventsWrapSelect = joinWithComma(eventsWrapSelect, fmt.Sprintf("%s as user_group_user_id", usersUserGroupColumn))
			}

			if strings.Contains(addJoinStmnt, "group_users") {
				eventsWrapSelect = joinWithComma(eventsWrapSelect, "group_users.properties as group_properties")
			}
		}
	}

	eventsWrapStmnt := "SELECT " + eventsWrapSelect + " FROM events" + " " + addJoinStmnt

	var fromTimestamp string
	if from > 0 {
		fromTimestamp = "?"
	} else if fromStr != "" {
		fromTimestamp = fromStr // allows from_timestamp from another step.
	}
	// Non-JSON where conditions except event_names.
	eventsWrapWhereCondition := fmt.Sprintf("WHERE events.project_id=? AND timestamp>=%s AND timestamp<=?", fromTimestamp)

	if userGroupColumn != "" {
		eventsWrapWhereCondition = eventsWrapWhereCondition + " AND " + " ( group_user_id IS NOT NULL OR user_group_user_id IS NOT NULL )"
	}

	// Building event_names condition and appending.
	var eventNamesCacheStmnt string
	eventNamesRef := "event_names"
	skipEventNameStep := C.SkipEventNameStepByProjectID(projectId)
	if stepName != "" && !skipEventNameStep {
		eventNamesRef = fmt.Sprintf("%s_names", stepName)
		eventNamesCacheStmnt = as(eventNamesRef, "SELECT id, project_id, name FROM event_names WHERE project_id=? AND name=?")
		*qParams = append(*qParams, projectId, qep.Name)
	}
	// select id of event_names from names step.
	if !skipEventNameStep {
		eventsWrapWhereCondition = eventsWrapWhereCondition + fmt.Sprintf(" "+"AND events.event_name_id IN (SELECT id FROM %s WHERE project_id=? AND name=?)", eventNamesRef)
	} else {
		eventNameIDsFilter := "events.event_name_id = ?"
		if len(qep.EventNameIDs) > 1 {
			for range qep.EventNameIDs[1:] {
				eventNameIDsFilter += " " + "OR" + " " + "events.event_name_id = ?"
			}
		}

		eventsWrapWhereCondition = eventsWrapWhereCondition + " " + "AND" + " " + " ( " + eventNameIDsFilter + " ) "
	}
	eventsWrapStmnt = appendStatement(eventsWrapStmnt, eventsWrapWhereCondition)

	eventsWrapViewName := fmt.Sprintf("%s_event_users_view", stepName)
	// Builds `(SELECT xxx FROM events [JOIN users] WHERE <Non-JSON filters...> LIMIT 1000000000) step_0_event_users_view`
	eventsWrapView := fmt.Sprintf("(%s LIMIT %d) %s", eventsWrapStmnt, model.FilterOptLimit, eventsWrapViewName)

	if userGroupColumn != "" {
		addSelecStmnt = strings.ReplaceAll(addSelecStmnt, userGroupColumn, eventsWrapViewName+".group_user_id")
		addSelecStmnt = strings.ReplaceAll(addSelecStmnt, usersUserGroupColumn, eventsWrapViewName+".user_group_user_id")
	}

	// Change properties column for the view.
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "users.properties", eventsWrapViewName+".global_user_properties")
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "events.properties", eventsWrapViewName+".event_properties")
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "events.user_properties", eventsWrapViewName+".event_user_properties")
	addSelecStmnt = strings.ReplaceAll(addSelecStmnt, "group_users.properties", eventsWrapViewName+".group_properties")
	// Use view instead of events and users table.
	addSelecStmnt = replaceTableWithViewForEventsAndUsers(addSelecStmnt, eventsWrapViewName)

	rStmnt := "SELECT " + addSelecStmnt + " FROM " + " " + eventsWrapView

	// adds params in order of '?'.
	if addSelecStmnt != "" && addSelectParams != nil {
		*qParams = append(*qParams, addSelectParams...)
	}
	*qParams = append(*qParams, projectId)
	if from > 0 {
		*qParams = append(*qParams, from)
	}

	*qParams = append(*qParams, to)
	if !skipEventNameStep {
		*qParams = append(*qParams, projectId, qep.Name)
	} else {
		*qParams = append(*qParams, qep.EventNameIDs...)
	}

	// Applying global user filter
	latestUserPropFilterStmnt := ""
	var latestUserPropFilterParam []interface{}
	var err error

	isWhereAdded := false
	hasGlobalPropertiesFilter := globalUserFilter != nil && len(globalUserFilter) != 0
	hasGlobalGroupProperties := hasGlobalPropertiesFilter && strings.Contains(addJoinStmnt, "group_users")
	if hasGlobalPropertiesFilter {

		if !hasGlobalGroupProperties {
			// add user filter
			latestUserPropFilterStmnt, latestUserPropFilterParam, err = getFilterSQLStmtForLatestUserProperties(
				projectId, globalUserFilter, from)
			if err != nil {
				return errors.New("invalid user properties for global filter")
			}
			latestUserPropFilterStmnt = strings.ReplaceAll(latestUserPropFilterStmnt,
				"users.properties", eventsWrapViewName+".global_user_properties")

		} else {
			// add group properties filter
			latestUserPropFilterStmnt, latestUserPropFilterParam, err = getFilterSQLStmtForLatestUserProperties(
				projectId, globalUserFilter, from)
			if err != nil {
				return errors.New("invalid group properties for global filter")
			}
			latestUserPropFilterStmnt = strings.ReplaceAll(latestUserPropFilterStmnt,
				"users.properties", eventsWrapViewName+".group_properties")
		}

		rStmnt = rStmnt + " WHERE " + latestUserPropFilterStmnt
		isWhereAdded = true
		*qParams = append(*qParams, latestUserPropFilterParam...)
	}

	// mergeCond for whereProperties can also be 'OR'.
	wStmnt, wParams, err := buildWhereFromProperties(projectId, qep.Properties, from)
	if err != nil {
		return err
	}

	if wStmnt != "" {
		conditionKeyword := "WHERE"
		if isWhereAdded {
			conditionKeyword = "AND"
		}

		wStmnt = strings.ReplaceAll(wStmnt, "events.properties", eventsWrapViewName+".event_properties")
		wStmnt = strings.ReplaceAll(wStmnt, "events.user_properties", eventsWrapViewName+".event_user_properties")

		rStmnt = rStmnt + " " + conditionKeyword + " " + fmt.Sprintf("( %s )", wStmnt)
		*qParams = append(*qParams, wParams...)
	}

	if groupBy != "" {
		groupBy = replaceTableWithViewForEventsAndUsers(groupBy, eventsWrapViewName)
		rStmnt = fmt.Sprintf("%s GROUP BY %s", rStmnt, groupBy)
	}

	if orderBy != "" {
		orderBy = replaceTableWithViewForEventsAndUsers(orderBy, eventsWrapViewName)
		rStmnt = fmt.Sprintf("%s ORDER BY %s", rStmnt, orderBy)
	}

	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
	}

	if !skipEventNameStep && eventNamesCacheStmnt != "" {
		rStmnt = joinWithComma(eventNamesCacheStmnt, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)

	return nil
}

func hasWhereEntity(ewp model.QueryEventWithProperties, entity string) bool {
	logFields := log.Fields{
		"ewp":    ewp,
		"entity": entity,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, p := range ewp.Properties {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func joinWithComma(x ...string) string {
	logFields := log.Fields{
		"x": x,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return joinWithWordInBetween(",", x...)
}

func joinWithWordInBetween(word string, x ...string) string {
	logFields := log.Fields{
		"word": word,
		"x":    x,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var y string
	for i, v := range x {
		if v != "" {
			if i == 0 || y == "" {
				y = v
			} else {
				y = fmt.Sprintf("%s %s %s", y, word, v)
			}
		}
	}

	return y
}

func appendSuffix(suffix string, x ...string) []string {
	logFields := log.Fields{
		"suffix": suffix,
		"x":      x,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var y []string
	for _, v := range x {
		y = append(y, fmt.Sprintf("%s %s ", v, suffix))
	}

	return y
}

func hasGroupEntity(props []model.QueryGroupByProperty, entity string) bool {
	logFields := log.Fields{
		"props":  props,
		"entity": entity,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, p := range props {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func addJoinLatestUserPropsQuery(projectID int64, groupProps []model.QueryGroupByProperty,
	refStepName string, stepName string, qStmnt *string, qParams *[]interface{}, addSelect string, timeZone string) string {
	logFields := log.Fields{
		"project_id":    projectID,
		"step_name":     stepName,
		"q_stmnt":       qStmnt,
		"q_params":      qParams,
		"group_props":   groupProps,
		"ref_step_name": refStepName,
		"add_select":    addSelect,
		"time_zone":     timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	groupSelect, gSelectParams, gKeys := buildGroupKeys(projectID, groupProps, timeZone)

	rStmnt := "SELECT " + joinWithComma(groupSelect, addSelect) + " from " + refStepName +
		" " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id"
	// Using string format for project_id condition, as the value is from internal system.
	rStmnt = rStmnt + " AND " + fmt.Sprintf("users.project_id = %d", projectID)
	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)
	*qParams = append(*qParams, gSelectParams...)

	return gKeys
}

func filterGroupPropsByType(gp []model.QueryGroupByProperty, entity string) []model.QueryGroupByProperty {
	logFields := log.Fields{
		"gp":     gp,
		"entity": entity,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	groupProps := make([]model.QueryGroupByProperty, 0)

	for _, v := range gp {
		if v.Entity == entity {
			groupProps = append(groupProps, v)
		}
	}
	return groupProps
}

func removeEventSpecificUserGroupBys(groupBys []model.QueryGroupByProperty) []model.QueryGroupByProperty {
	logFields := log.Fields{
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	filteredProps := make([]model.QueryGroupByProperty, 0)
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
	logFields := log.Fields{
		"q_stmnt": qStmnt,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return fmt.Sprintf("%s ORDER BY %s DESC", qStmnt, model.AliasAggr)
}

func appendOrderByAggrAndGroupKeys(projectID int64, qStmnt string, groupBys []model.QueryGroupByProperty, timeZone string) string {
	logFields := log.Fields{
		"q_stmnt":    qStmnt,
		"project_id": projectID,
		"group_bys":  groupBys,
		"time_zone":  timeZone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, _, groupKeys := buildGroupKeys(projectID, groupBys, timeZone)
	return joinWithComma(fmt.Sprintf("%s ORDER BY %s DESC", qStmnt, model.AliasAggr), groupKeys)
}

//func appendOrderByEventNameAndAggr(qStmnt string) string {
//	return fmt.Sprintf("%s ORDER BY event_name, %s DESC", qStmnt, model.AliasAggr)
//}

func appendSelectTimestampColIfRequired(stmnt string, isRequired bool) string {
	logFields := log.Fields{
		"stmnt":       stmnt,
		"is_required": isRequired,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if !isRequired {
		return stmnt
	}

	return joinWithComma(stmnt, model.AliasDateTime)
}

func getSelectTimestampByType(timestampType, timezone string) string {
	logFields := log.Fields{
		"timestamp_type": timestampType,
		"timezone":       timezone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectTz string

	if timezone == "" {
		selectTz = model.DefaultTimezone
	} else {
		selectTz = timezone
	}

	var selectStr string
	if timestampType == model.GroupByTimestampHour {
		selectStr = fmt.Sprintf("date_trunc('hour', CONVERT_TZ(FROM_UNIXTIME(timestamp), 'UTC', '%s'))", selectTz)
	} else if timestampType == model.GroupByTimestampWeek {
		// default week is Monday to Sunday for memsql, updating it to Sunday to Saturday
		selectStr = fmt.Sprintf("date_trunc('week', CONVERT_TZ(FROM_UNIXTIME(timestamp + (24*60*60)), 'UTC', '%s')) - INTERVAL 1 day", selectTz)
	} else if timestampType == model.GroupByTimestampMonth {
		selectStr = fmt.Sprintf("date_trunc('month', CONVERT_TZ(FROM_UNIXTIME(timestamp), 'UTC', '%s'))", selectTz)
	} else if timestampType == model.GroupByTimestampQuarter {
		selectStr = fmt.Sprintf("date_trunc('quarter', CONVERT_TZ(FROM_UNIXTIME(timestamp), 'UTC', '%s'))", selectTz)
	} else {
		// defaults to GroupByTimestampDate.
		selectStr = fmt.Sprintf("date_trunc('day', CONVERT_TZ(FROM_UNIXTIME(timestamp), 'UTC', '%s'))", selectTz)
	}

	return selectStr
}

func getSelectTimestampByTypeAndPropertyName(timestampType, propertyName, timezone string) string {
	logFields := log.Fields{
		"timestamp_type": timestampType,
		"timezone":       timezone,
		"property_name":  propertyName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectTz string

	if timezone == "" {
		selectTz = model.DefaultTimezone
	} else {
		selectTz = timezone
	}

	propertyToNum := "CONVERT(SUBSTRING(" + propertyName + ",1,10), DECIMAL(10))"
	var selectStr string

	// Note: Second is used as granularity only in profiles which is called from attribution.
	if timestampType == model.GroupByTimestampSecond {
		selectStr = fmt.Sprintf("date_trunc('second', CONVERT_TZ(FROM_UNIXTIME("+propertyToNum+"), 'UTC', '%s'))", selectTz)
	} else if timestampType == model.GroupByTimestampHour {
		selectStr = fmt.Sprintf("date_trunc('hour', CONVERT_TZ(FROM_UNIXTIME("+propertyToNum+"), 'UTC', '%s'))", selectTz)
	} else if timestampType == model.GroupByTimestampWeek {
		// default week is Monday to Sunday for memsql, updating it to Sunday to Saturday
		selectStr = fmt.Sprintf("date_trunc('week', CONVERT_TZ(FROM_UNIXTIME("+propertyToNum+" + (24*60*60)), 'UTC', '%s')) - INTERVAL 1 day", selectTz)
	} else if timestampType == model.GroupByTimestampMonth {
		selectStr = fmt.Sprintf("date_trunc('month', CONVERT_TZ(FROM_UNIXTIME("+propertyToNum+"), 'UTC', '%s'))", selectTz)
	} else if timestampType == model.GroupByTimestampQuarter {
		selectStr = fmt.Sprintf("date_trunc('quarter', CONVERT_TZ(FROM_UNIXTIME("+propertyToNum+"), 'UTC', '%s'))", selectTz)
	} else {
		// defaults to GroupByTimestampDate.
		selectStr = fmt.Sprintf("date_trunc('day', CONVERT_TZ(FROM_UNIXTIME("+propertyToNum+"), 'UTC', '%s'))", selectTz)
	}

	return selectStr
}

func appendSelectTimestampIfRequired(stmnt string, groupByTimestamp string, timezone string) string {
	logFields := log.Fields{
		"stmnt":               stmnt,
		"timezone":            timezone,
		"group_by_time_stamp": groupByTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if groupByTimestamp == "" {
		return stmnt
	}

	return joinWithComma(stmnt, fmt.Sprintf("%s as %s",
		getSelectTimestampByType(groupByTimestamp, timezone), model.AliasDateTime))
}

func appendGroupByTimestampIfRequired(qStmnt string, isRequired bool, groupKeys ...string) string {
	logFields := log.Fields{
		"q_stmnt":     qStmnt,
		"is_required": isRequired,
		"group_keys":  groupKeys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Added groups with timestamp.
	groups := make([]string, 0)
	if isRequired {
		groups = append(groups, model.AliasDateTime)
	}
	groups = append(groups, groupKeys...)
	qStmnt = appendGroupBy(qStmnt, groups...)
	return qStmnt
}

func appendGroupBy(qStmnt string, gKeys ...string) string {
	logFields := log.Fields{
		"q_stmnt": qStmnt,
		"g_keys":  gKeys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(gKeys) == 0 || (len(gKeys) == 1 && gKeys[0] == "") {
		return qStmnt
	}

	return fmt.Sprintf("%s GROUP BY %s", qStmnt, joinWithComma(gKeys...))
}

func appendLimitByCondition(qStmnt string, groupProps []model.QueryGroupByProperty, groupByTimestamp bool) string {
	logFields := log.Fields{
		"q_stmnt":            qStmnt,
		"group_props":        groupProps,
		"group_by_timestamp": groupByTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Limited with max limit on SQL. Limited on server side.
	return fmt.Sprintf("%s LIMIT %d", qStmnt, model.MaxResultsLimit)
}

func separateEventLevelGroupBys(allGroupBys []model.QueryGroupByProperty) (
	eventLevelGroupBys, otherGroupBys []model.QueryGroupByProperty) {
	logFields := log.Fields{
		"all_groups_bys": allGroupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func isEventLevelGroupBy(groupBy model.QueryGroupByProperty) bool {
	logFields := log.Fields{
		"group_by": groupBy,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return groupBy.EventName != "" && groupBy.EventNameIndex != 0
}

// TranslateGroupKeysIntoColumnNames - Replaces groupKeys on result
// headers with real column names.
func translateGroupKeysIntoColumnNames(result *model.QueryResult,
	groupProps []model.QueryGroupByProperty) error {
	logFields := log.Fields{
		"group_props": groupProps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rcols := make([]string, 0)

	indexLookup := make(map[int]model.QueryGroupByProperty)
	for _, v := range groupProps {
		indexLookup[v.Index] = v
	}

	for i := range result.Headers {
		if strings.HasPrefix(result.Headers[i], model.GroupKeyPrefix) {
			gIndexStr := strings.TrimPrefix(result.Headers[i], model.GroupKeyPrefix)
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
func addIndexToGroupByProperties(query *model.Query) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for i := range query.GroupByProperties {
		query.GroupByProperties[i].Index = i
	}
}

func getGroupKeyIndexesForSlicing(cols []string) (int, int, error) {
	logFields := log.Fields{
		"cols": cols,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	start := -1
	end := -1

	index := 0
	for _, col := range cols {
		if strings.HasPrefix(col, model.GroupKeyPrefix) || col == "event_name" {
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
	logFields := log.Fields{
		"cols": cols,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"group_by_timestamp": groupByTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if groupByTimestamp == "" {
		return true
	}

	for _, gbtType := range model.GroupByTimestampTypes {
		if gbtType == groupByTimestamp {
			return true
		}
	}

	return false
}

// IsValidEventsQuery Validates query for Events class and returns corresponding errMsg.
func IsValidEventsQuery(query *model.Query) (bool, string) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if query.Class != model.QueryClassEvents {
		return false, "Invalid query class given"
	}

	if query.Type != model.QueryTypeEventsOccurrence &&
		query.Type != model.QueryTypeUniqueUsers {
		return false, "Invalid query type given"
	}

	if query.EventsCondition != model.EventCondAllGivenEvent &&
		query.EventsCondition != model.EventCondAnyGivenEvent &&
		query.EventsCondition != model.EventCondEachGivenEvent {
		return false, "Invalid events condition given"
	}

	errMsg, hasError := validateQueryProps(query)
	if hasError {
		return false, errMsg
	}
	return true, ""
}

// IsValidQuery Validates and returns errMsg which is used as response.
func IsValidQuery(query *model.Query) (bool, string) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if query.Type != model.QueryTypeEventsOccurrence &&
		query.Type != model.QueryTypeUniqueUsers {
		return false, "Invalid query type given"
	}

	if query.EventsCondition != model.EventCondAllGivenEvent &&
		query.EventsCondition != model.EventCondAnyGivenEvent &&
		query.EventsCondition != model.EventCondEachGivenEvent &&
		query.EventsCondition != model.EventCondFunnelAnyGivenEvent {
		return false, "Invalid events condition given"
	}

	errMsg, hasError := validateQueryProps(query)
	if hasError {
		return false, errMsg
	}
	return true, ""
}

func getQueryCacheRedisKeySuffix(hashString string, from, to int64, timezoneString U.TimeZoneString) string {
	logFields := log.Fields{
		"hash_string":     hashString,
		"from":            from,
		"to":              to,
		"timezone_string": timezoneString,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if to-from == model.DateRangePreset2MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, model.DateRangePreset2MinLabel)
	} else if to-from == model.DateRangePreset30MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, model.DateRangePreset30MinLabel)
	} else if U.IsStartOfTodaysRangeIn(from, timezoneString) {
		return fmt.Sprintf("%s:from:%d", hashString, from)
	}
	return fmt.Sprintf("%s:from:%d:to:%d", hashString, from, to)
}

// GetQueryResultFromCache To get value from cache for a particular query payload.
// resultContainer to be passed by reference.
func GetQueryResultFromCache(projectID int64, query model.BaseQuery, resultContainer interface{}) (model.QueryCacheResult, int) {
	logFields := log.Fields{
		"projected_id":     projectID,
		"query":            query,
		"resukt_container": resultContainer,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	var queryResult model.QueryCacheResult
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return queryResult, http.StatusInternalServerError
	}

	// Using persistent redis for this.
	value, exists, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if err != nil {
		logCtx.WithError(err).Error("Error getting value from redis")
		return queryResult, http.StatusInternalServerError
	}
	if !exists {
		return queryResult, http.StatusNotFound
	} else if value == model.QueryCacheInProgressPlaceholder {
		return queryResult, http.StatusAccepted
	}

	err = json.Unmarshal([]byte(value), &queryResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal cache result to result container")
		return queryResult, http.StatusInternalServerError
	}

	err = json.Unmarshal([]byte(value), resultContainer)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal cache result to result container")
		return queryResult, http.StatusInternalServerError
	}

	return queryResult, http.StatusFound
}

func validateQueryProps(query *model.Query) (string, bool) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"cols": cols,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	timeIndex := -1
	aggrIndex := -1

	for i, c := range cols {
		if c == model.AliasDateTime {
			timeIndex = i
		}

		if c == model.AliasAggr {
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

// TODO (Anil) update this for v1/ each users count query
func sanitizeNumericalBucketRange(query *model.Query, rows [][]interface{}, indexToSanitize int) {
	logFields := log.Fields{
		"query":             query,
		"rows":              rows,
		"index_to_sanitize": indexToSanitize,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for index, row := range rows {
		if query.Class == model.QueryClassFunnel && index == 0 {
			// For funnel queries, first row is $no_group query. Skip sanitization.
			continue
		}

		// Remove trailing .0 in start and end value of range.
		row[indexToSanitize] = trailingZeroRegex.ReplaceAllString(row[indexToSanitize].(string), "")

		// Change range with same start and end ex: 2 - 2 to just 2.
		if row[indexToSanitize] != model.PropertyValueNone {
			rowSplit := strings.Split(row[indexToSanitize].(string), model.NumericalGroupBySeparator)
			if rowSplit[0] == rowSplit[1] {
				row[indexToSanitize] = model.GetBucketRangeForStartAndEnd(rowSplit[0], rowSplit[1])
			}
		}
	}
}

// sanitizeNumericalBucketRanges Removes any .0 added to bucket ranges wherever possible.
func sanitizeNumericalBucketRanges(result *model.QueryResult, query *model.Query) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	headerIndexMap := make(map[string][]int)
	for index, header := range result.Headers {
		// If same group by is added twice, it will appear twice in headers.
		// Keep as a list to sanitize both indexes.
		headerIndexMap[header] = append(headerIndexMap[header], index)
	}

	sanitizedProperties := make(map[string]bool)
	for _, gbp := range query.GroupByProperties {
		if isGroupByTypeWithBuckets([]model.QueryGroupByProperty{gbp}) {
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

// ExecQueryWithContext Executes raw query with context. Useful to kill
// queries on program exit or crash.
func (store *MemSQL) ExecQueryWithContext(stmnt string, params []interface{}) (*sql.Rows, *sql.Tx, error, string) {
	reqID := U.GetUniqueQueryRequestID()

	var db *gorm.DB
	if C.IsDBConnectionPool2Enabled() {
		db = C.GetServices().Db2
	} else {
		db = C.GetServices().Db
	}

	tx, err := db.DB().Begin()
	if err != nil {
		log.WithError(err).Error("Failed to begin DB transaction.")
		return nil, nil, err, reqID
	}

	// For query: ...where id in ($1) where $1 is passed as a slice, convert to pq.Array()
	stmnt, params = model.ExpandArrayWithIndividualValues(stmnt, params)

	logFields := log.Fields{
		"anaytics":       true,
		"expanded_query": U.DBDebugPreparedStatement(C.GetConfig().Env, stmnt, params),
		// Limit statement and params length.
		"original_query": U.TrimQueryString(C.GetConfig().Env, stmnt),
		"params":         U.TrimQueryParams(C.GetConfig().Env, params),
		"req_id":         reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// Prefix application name for in comment for debugging.
	stmnt = fmt.Sprintf("/*!%s-%s*/ ", C.GetConfig().AppName, reqID) + stmnt

	// Set resource pool before query.
	if !C.IsDBConnectionPool2Enabled() {
		if usePool, poolName := C.UseResourcePoolForAnalytics(); usePool {
			C.SetMemSQLResourcePoolQueryCallbackUsingSQLTx(tx, poolName)
		}
	}
	startExecTime := time.Now()
	rows, err := tx.QueryContext(*C.GetServices().DBContext, stmnt, params...)
	U.LogExecutionTimeWithQueryRequestID(startExecTime, reqID, &logFields)
	if err != nil {
		log.WithError(err).WithFields(logFields).Error("Failed to exec query with context.")
	}

	log.WithError(err).WithFields(logFields).Info("Exec query with context.")

	return rows, tx, err, reqID
}

func (store *MemSQL) ExecQuery(stmnt string, params []interface{}) (*model.QueryResult, error, string) {
	logFields := log.Fields{
		"stmnt":  stmnt,
		"params": params,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	rows, tx, err, reqID := store.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err, reqID
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows, tx, reqID)
	if err != nil {
		return nil, err, reqID
	}

	result := &model.QueryResult{Headers: resultHeaders, Rows: resultRows}
	return result, nil, reqID
}

func addQueryToResultMeta(result *model.QueryResult, query model.Query) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	result.Meta.Query = query
}

func isValidFunnelQuery(query *model.Query) bool {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return len(query.EventsWithProperties) <= 10
}

func (store *MemSQL) IsValidFunnelGroupQueryIfExists(projectID int64, query *model.Query, groupIds []int) (int, bool) {

	if query.GroupAnalysis == "" || model.IsFunnelQueryGroupNameUser(query.GroupAnalysis) {
		return 0, true
	}

	scopeGroupID, valid := store.IsValidFunnelGroupQuery(projectID, query, groupIds)
	return scopeGroupID, valid
}

func (store *MemSQL) IsValidFunnelGroupQuery(projectID int64, query *model.Query, groupIds []int) (int, bool) {
	if !model.IsValidFunnelQueryGroupName(query.GroupAnalysis) {
		log.WithFields(log.Fields{"query": query}).Error("Invalid funnel group query.")
		return 0, false
	}

	scopeGroup, status := store.GetGroup(projectID, query.GroupAnalysis)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"query": query}).Error("Failed to get group id from funnel groupe query")
		}
		return 0, false
	}

	for i := range groupIds {

		// group id 0 is for user
		if groupIds[i] == 0 {
			continue
		}

		if groupIds[i] != scopeGroup.ID {
			log.WithFields(log.Fields{"query": query}).Error("Invalid funnel group query. Different groups selected.")
			return 0, false
		}
	}

	return scopeGroup.ID, true
}

func (store *MemSQL) Analyze(projectId int64, queryOriginal model.Query, enableFilterOpt bool, enableFunnelV2 bool) (*model.QueryResult, int, string) {
	logFields := log.Fields{
		"project_id":     projectId,
		"query_original": queryOriginal,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var query model.Query
	U.DeepCopy(&queryOriginal, &query)

	if valid, errMsg := IsValidQuery(&query); !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	if query.Class == model.QueryClassFunnel {
		return store.RunFunnelQuery(projectId, query, enableFilterOpt, enableFunnelV2)
	}
	return store.RunInsightsQuery(projectId, query, enableFilterOpt)
}

func (store *MemSQL) IsGroupEventNameByQueryEventWithProperties(projectID int64, ewp model.QueryEventWithProperties) (string, int) {
	eventNameID := ""
	if len(ewp.EventNameIDs) > 0 {
		eventNameID = U.GetPropertyValueAsString(ewp.EventNameIDs[0])
	}

	return store.IsGroupEventName(projectID, ewp.Name, eventNameID)
}

func addMissingTimestampsOnResultWithoutGroupByProps(result *model.QueryResult,
	query *model.Query, aggrIndex int, timestampIndex int) error {
	logFields := log.Fields{
		"query":           query,
		"aggr_index":      aggrIndex,
		"timestamp_index": timestampIndex,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	filledResult := make([][]interface{}, 0, 0)
	rowsByTimestamp := make(map[string]bool, 0)
	for _, row := range result.Rows {
		timestampString := U.GetTimestampAsStrWithTimezone(
			row[timestampIndex].(time.Time), query.Timezone)

		rowsByTimestamp[timestampString] = true
		filledResult = append(filledResult, row)
	}

	timestamps, offsets := getAllTimestampsAndOffsetBetweenByType(query.From, query.To,
		query.GetGroupByTimestamp(), query.Timezone)

	// range over timestamps between given from and to.
	// uses timestamp string for comparison.
	for index, ts := range timestamps {

		if _, exists := rowsByTimestamp[U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index])]; !exists {
			newRow := make([]interface{}, 3, 3)
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
func addMissingTimestampsOnResultWithGroupByProps(result *model.QueryResult,
	query *model.Query, aggrIndex int, timestampIndex int) error {
	logFields := log.Fields{
		"query":           query,
		"aggr_index":      aggrIndex,
		"timestamp_index": timestampIndex,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
		encCols = append(encCols, U.GetTimestampAsStrWithTimezone(row[timestampIndex].(time.Time), query.Timezone))
		encKey := getEncodedKeyForCols(encCols)

		rowsByGroupAndTimestamp[encKey] = true
		filledResult = append(filledResult, row)
	}

	timestamps, offsets := getAllTimestampsAndOffsetBetweenByType(query.From, query.To,
		query.GetGroupByTimestamp(), query.Timezone)

	for _, row := range result.Rows {
		for index, ts := range timestamps {
			encCols := make([]interface{}, 0, 0)
			encCols = append(encCols, row[gkStart:gkEnd]...)
			// encoded key with generated timestamp.
			encCols = append(encCols, U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index]))
			encKey := getEncodedKeyForCols(encCols)

			if _, exists := rowsByGroupAndTimestamp[encKey]; !exists {
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

// fills empty values for dates which data is not present. Eg-> query date range: 1 jan to 3 jan. Data in DB:= 1/01 -> 100, 3/01 -> 150
// After going through following method, final data: 1/01 -> 100, 2/01 -> 0, 3/01 -> 150

func addMissingTimestampsOnChannelResultWithoutGroupByProps(result *model.QueryResult,
	query *model.KPIQuery, aggrIndex int, timestampIndex int, isTimezoneEnabled bool) error {
	logFields := log.Fields{
		"query":               query,
		"aggr_index":          aggrIndex,
		"timestamp_index":     timestampIndex,
		"is_timezone_enabled": isTimezoneEnabled,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rowsByTimestamp := make(map[string][]interface{}, 0)
	for _, row := range result.Rows {

		ts, tErr := U.GetTimeFromParseTimeStrWithErrorFromInterface(row[timestampIndex])
		if tErr != nil {
			return tErr
		}
		rowsByTimestamp[U.GetTimestampAsStrWithTimezone(ts, query.Timezone)] = row
	}

	timestamps, offsets := getAllTimestampsAndOffsetBetweenByType(query.From, query.To,
		query.GroupByTimestamp, query.Timezone)

	filledResult := make([][]interface{}, 0, 0)
	// range over timestamps between given from and to.
	// uses timestamp string for comparison.
	for index, ts := range timestamps {
		timestampWithTimezone := U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index])
		if row, exists := rowsByTimestamp[U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index])]; exists {
			// overrides timestamp with user timezone as sql results doesn't
			// return timezone used to query.
			row[timestampIndex] = timestampWithTimezone
			filledResult = append(filledResult, row)
		} else {
			newRow := make([]interface{}, 2, 2)
			newRow[timestampIndex] = timestampWithTimezone
			newRow[aggrIndex] = 0
			filledResult = append(filledResult, newRow)
		}
	}

	result.Rows = filledResult
	return nil
}

// Need a separate method for this because group by keys are involved and we have fill data for each key.
// query -> group by camapign_name, date -> 1 jan to 2 jan. DB data [[1/01, a, 100], [2/01, b, 50]]
// Final data -> [[1/01, a, 100],[1/01, b, 0],[2/01, a, 0], [2/01, b, 50]]
func addMissingTimestampsOnChannelResultWithGroupByProps(result *model.QueryResult,
	query *model.KPIQuery, aggrIndex int, timestampIndex int, isTimezoneEnabled bool) error {
	logFields := log.Fields{
		"query":               query,
		"aggr_index":          aggrIndex,
		"timestamp_index":     timestampIndex,
		"is_timezone_enabled": isTimezoneEnabled,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	gkStart, gkEnd, err := getChannelGroupKeyIndexesForSlicing(result.Headers)
	if err != nil {
		return err
	}

	filledResult := make([][]interface{}, 0)
	rowsByGroupAndTimestamp := make(map[string]bool, 0)
	for _, row := range result.Rows {
		encCols := make([]interface{}, 0)
		encCols = append(encCols, row[gkStart:gkEnd]...)

		ts, tErr := U.GetTimeFromParseTimeStrWithErrorFromInterface(row[timestampIndex])
		if tErr != nil {
			return tErr
		}
		timestampWithTimezone := U.GetTimestampAsStrWithTimezone(ts, query.Timezone)
		// encoded key with group values and timestamp from db row.
		encCols = append(encCols, timestampWithTimezone)
		encKey := getEncodedKeyForCols(encCols)

		rowsByGroupAndTimestamp[encKey] = true
		filledResult = append(filledResult, row)
	}

	timestamps, offsets := getAllTimestampsAndOffsetBetweenByType(query.From, query.To,
		query.GroupByTimestamp, query.Timezone)

	for _, row := range result.Rows {
		for index, ts := range timestamps {
			encCols := make([]interface{}, 0)
			encCols = append(encCols, row[gkStart:gkEnd]...)
			// encoded key with generated timestamp.
			timestampWithTimezone := U.GetTimestampAsStrWithTimezoneGivenOffset(ts, offsets[index])
			encCols = append(encCols, timestampWithTimezone)
			encKey := getEncodedKeyForCols(encCols)

			if _, exists := rowsByGroupAndTimestamp[encKey]; !exists {
				// create new row with group values and missing date
				// for those group combination and aggr 0.
				rowLen := len(result.Headers)
				newRow := make([]interface{}, rowLen)
				groupValues := row[gkStart:gkEnd]

				for i := 0; i < rowLen; {
					if i == gkStart {
						for _, gv := range groupValues {
							newRow[i] = gv
							if i == timestampIndex {
								newRow[i] = timestampWithTimezone
							}
							i++
						}
					}

					if i == aggrIndex {
						newRow[i] = 0
						i++
					}

					if i == timestampIndex {
						newRow[i] = timestampWithTimezone
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

func getChannelGroupKeyIndexesForSlicing(cols []string) (int, int, error) {
	logFields := log.Fields{
		"cols": cols,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	start := -1
	end := -1

	index := 0
	for _, col := range cols {
		if col != "datetime" && col != "aggregate" {
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

var (
	conversionTimeRegex = regexp.MustCompile(`^(\d+)(D|H|M)$`)
)

func getConversionTimeJoinCondition(q model.Query, i int) (string, int) {
	if !conversionTimeRegex.Match([]byte(q.ConversionTime)) {
		log.WithFields(log.Fields{"query": q}).Error("Invalid conversion time on funnel query.")
		return "", 0
	}

	if q.Timezone == "" {
		log.WithFields(log.Fields{"query": q}).Error("Invalid timezone on funnel query conversion time.")
		return "", 0
	}

	substrings := conversionTimeRegex.FindStringSubmatch(q.ConversionTime)
	numStr, precision := substrings[1], substrings[2]
	num, err := U.GetPropertyValueAsFloat64(numStr)
	if err != nil {
		log.WithError(err).Error("Failed to convert conversion time to count. Continuing with 0.")
	}

	if precision == "D" {
		// From current day(0) till nth day in timezone midnight
		stmnt := fmt.Sprintf(" AND timestampdiff(DAY, DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', '%s')), "+
			"DATE(CONVERT_TZ(FROM_UNIXTIME(step_%d_timestamp), 'UTC', '%s'))) <= ? ", q.Timezone, i, q.Timezone)
		return stmnt, int(num)
	}

	if precision == "H" {
		stmnt := fmt.Sprintf(" AND (step_%d_timestamp - step_0_timestamp)  <= ? ", i)
		return stmnt, int(num * 3600)
	}

	stmnt := fmt.Sprintf(" AND (step_%d_timestamp - step_0_timestamp)  <= ? ", i)
	return stmnt, int(num * 60)

}
