package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
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
	model.ContainsOpStr:           "ILIKE",
	model.NotContainsOpStr:        "NOT ILIKE",
}

// Query cache related constants.
const (
	QueryCacheInProgressPlaceholder string = "QUERY_CACHE_IN_PROGRESS"

	DateRangePreset2MinLabel      string = "2MIN"
	DateRangePreset30MinLabel     string = "30MIN"
	DateRangePreset2MinInSeconds  int64  = 2 * 60
	DateRangePreset30MinInSeconds int64  = 30 * 60

	QueryCachePlaceholderExpirySeconds     float64 = 5 * 60            // 5 Minutes.
	QueryCacheImmutableResultExpirySeconds float64 = 30 * 24 * 60 * 60 // 30 Days.
	QueryCacheMutableResultExpirySeconds   float64 = 10 * 60           // 10 Minutes.

	QueryCacheRequestSleepHeader       string = "QuerySleepSeconds"
	QueryCacheResponseFromCacheHeader  string = "Fromcache"
	QueryCacheResponseCacheRefreshedAt string = "Refreshedat"
	QueryCacheRedisKeyPrefix           string = "query:cache"
)

// UserPropertyGroupByPresent Sent from frontend for breakdown on latest user property.
const UserPropertyGroupByPresent string = "$present"

// NumericalValuePostgresRegex Used to remove non numerical values in numerical bucketing.
const NumericalValuePostgresRegex string = "\\$none|^-?[0-9]+\\.?[0-9]*$"

func with(stmnt string) string {
	return fmt.Sprintf("WITH %s", stmnt)
}

func getOp(OpStr string) string {
	v, ok := queryOps[OpStr]
	if !ok {
		log.Errorf("invalid query operator %s, using default", OpStr)
		return model.EqualsOp
	}

	return v
}

func getPropertyEntityField(projectID uint64, groupProp model.QueryGroupByProperty) string {
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

func getPropertyEntityFieldForFilter(projectID uint64, entityName string) string {
	if entityName == model.PropertyEntityUser {
		// Filtering is supported only with event level user_properties.
		return "events.user_properties"
	} else if entityName == model.PropertyEntityEvent {
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

func isValidLogicalOp(op string) bool {
	return op == "AND" || op == "OR"
}

func buildWhereFromProperties(projectID uint64, properties []model.QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	pLen := len(properties)
	if pLen == 0 {
		return rStmnt, rParams, nil
	}

	groupedProperties := make(map[string][]model.QueryProperty)
	propertyKeys := make([]string, 0, 0)
	for _, p := range properties {
		// Use Entity.Property as key to distinguish same property filter on user and event entity.
		propertyKey := p.Entity + "." + p.Property
		if _, exists := groupedProperties[propertyKey]; !exists {
			propertyKeys = append(propertyKeys, propertyKey)
		}
		groupedProperties[propertyKey] = append(groupedProperties[propertyKey], p)
	}

	rParams = make([]interface{}, 0, 0)
	groupIndex := 0
	for _, propertyKey := range propertyKeys {
		var groupStmnt string
		for i, p := range groupedProperties[propertyKey] {
			// defaults logic op if not given.
			if p.LogicalOp == "" {
				p.LogicalOp = "AND"
			}

			if !isValidLogicalOp(p.LogicalOp) {
				return rStmnt, rParams, errors.New("invalid logical op on where condition")
			}

			propertyEntity := getPropertyEntityFieldForFilter(projectID, p.Entity)
			propertyOp := getOp(p.Operator)

			if p.Value != model.PropertyValueNone {
				var pStmnt string
				if p.Type == U.PropertyTypeDateTime {
					pStmnt = fmt.Sprintf("(%s->>?>=? AND %s->>?<=?)", propertyEntity, propertyEntity)

					dateTimeValue, err := model.DecodeDateTimePropertyValue(p.Value)
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
					if p.Operator == model.ContainsOpStr || p.Operator == model.NotContainsOpStr {
						pValue = fmt.Sprintf("%%%s%%", p.Value)
					} else {
						pValue = p.Value
					}

					pStmnt = fmt.Sprintf("%s->>? %s ?", propertyEntity, propertyOp)
					rParams = append(rParams, p.Property, pValue)
				}

				if i == 0 {
					groupStmnt = pStmnt
				} else {
					groupStmnt = fmt.Sprintf("%s %s %s", groupStmnt, p.LogicalOp, pStmnt)
				}

				continue
			}

			// where condition for $none value.
			var whereCond string
			if propertyOp == model.EqualsOp {
				// i.e: (NOT jsonb_exists(events.properties, 'property_name') OR events.properties->>'property_name'='')
				whereCond = fmt.Sprintf("(NOT jsonb_exists(%s, ?) OR %s->>?='')", propertyEntity, propertyEntity)
			} else if propertyOp == model.NotEqualOp {
				// i.e: (jsonb_exists(events.properties, 'property_name') AND events.properties->>'property_name'!='')
				whereCond = fmt.Sprintf("(jsonb_exists(%s, ?) AND %s->>?!='')", propertyEntity, propertyEntity)
			} else {
				return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
			}

			if i == 0 {
				groupStmnt = whereCond
			} else {
				groupStmnt = fmt.Sprintf("%s %s %s", groupStmnt, p.LogicalOp, whereCond)
			}

			rParams = append(rParams, p.Property, p.Property)
		}
		if groupIndex == 0 {
			rStmnt = fmt.Sprintf("(%s)", groupStmnt)
		} else {
			logicalOp := groupedProperties[propertyKey][0].LogicalOp
			if logicalOp == "" {
				logicalOp = "AND"
			}
			rStmnt = fmt.Sprintf("%s %s (%s)", rStmnt, logicalOp, groupStmnt)
		}
		groupIndex++
	}

	return rStmnt, rParams, nil
}

// returns SQL query condition to address conditions only on events.properties
func getFilterSQLStmtForEventProperties(projectID uint64, properties []model.QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	var filteredProperty []model.QueryProperty
	for _, p := range properties {

		propertyEntity := getPropertyEntityFieldForFilter(projectID, p.Entity)
		if propertyEntity == "events.properties" {
			filteredProperty = append(filteredProperty, p)
		}
	}
	wStmt, wParams, err := buildWhereFromProperties(projectID, filteredProperty)
	if err != nil {
		return "", nil, err
	}
	return wStmt, wParams, nil
}

// returns SQL query condition to address conditions only on user_properties.properties
func getFilterSQLStmtForUserProperties(projectID uint64, properties []model.QueryProperty) (rStmnt string, rParams []interface{}, err error) {

	var filteredProperty []model.QueryProperty
	for _, p := range properties {
		propertyEntity := getPropertyEntityFieldForFilter(projectID, p.Entity)
		if propertyEntity == "events.user_properties" {
			filteredProperty = append(filteredProperty, p)
		}
	}
	wStmt, wParams, err := buildWhereFromProperties(projectID, filteredProperty)
	if err != nil {
		return "", nil, err
	}
	return wStmt, wParams, nil
}

// Alias for group by properties gk_1, gk_2.
func groupKeyByIndex(i int) string {
	return fmt.Sprintf("%s%d", model.GroupKeyPrefix, i)
}

func stepNameByIndex(i int) string {
	return fmt.Sprintf("%s%d", model.StepPrefix, i)
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
func getNoneHandledGroupBySelect(projectID uint64, groupProp model.QueryGroupByProperty, groupKey string) (string, []interface{}) {
	entityField := getPropertyEntityField(projectID, groupProp)
	var groupSelect string
	groupSelectParams := make([]interface{}, 0)
	if groupProp.Type != U.PropertyTypeDateTime {
		groupSelect = fmt.Sprintf("CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' ELSE %s->>? END AS %s",
			entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, entityField, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property}
	} else {
		groupSelect = fmt.Sprintf("CASE WHEN %s->>? IS NULL THEN '%s' WHEN %s->>? = '' THEN '%s' WHEN %s->>? = '0' THEN '%s' ELSE  date_trunc('%s', timezone('%s', to_timestamp(to_number((%s->>?)::text,'9999999999'))))::text END AS %s",
			entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, entityField, model.PropertyValueNone, groupProp.Granularity, U.TimeZoneStringIST, entityField, groupKey)
		groupSelectParams = []interface{}{groupProp.Property, groupProp.Property, groupProp.Property, groupProp.Property}
	}
	return groupSelect, groupSelectParams
}

// groupBySelect: user_properties.properties->>'age' as gk_1, events.properties->>'category' as gk_2
// groupByKeys: gk_1, gk_2
// How to use?
// select user_properties.properties->>'age' as gk_1, events.properties->>'category' as gk_2 from events
// group by gk_1, gk_2
func buildGroupKeys(projectID uint64, groupProps []model.QueryGroupByProperty) (groupSelect string,
	groupSelectParams []interface{}, groupKeys string) {

	groupSelectParams = make([]interface{}, 0)

	for i, v := range groupProps {
		// Order of group is preserved as received.
		gKey := groupKeyByIndex(v.Index)
		noneHandledSelect, noneHandledSelectParams := getNoneHandledGroupBySelect(projectID, v, gKey)
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

func appendNumericalBucketingSteps(qStmnt *string, qParams *[]interface{}, groupProps []model.QueryGroupByProperty, refStepName, eventNameSelect string,
	isGroupByTimestamp bool, additionalSelectKeys string) (bucketedStepName, aggregateSelectKeys string, aggregateGroupBys, aggregateOrderBys []string) {
	bucketedStepName = "bucketed"
	bucketedSelect := "SELECT "
	if eventNameSelect != "" {
		bucketedSelect = bucketedSelect + eventNameSelect + ", "
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
		boundsStatement := fmt.Sprintf("SELECT percentile_disc(%.2f) WITHIN GROUP(ORDER BY %s::numeric) + 0.00001 AS lbound, "+
			"percentile_disc(%.2f) WITHIN GROUP(ORDER BY %s::numeric) AS ubound FROM %s "+
			"WHERE %s != '%s' AND %s != '' AND %s ~ ? ", model.NumericalLowerBoundPercentile, groupKey, model.NumericalUpperBoundPercentile,
			groupKey, refStepName, groupKey, model.PropertyValueNone, groupKey, groupKey)
		*qParams = append(*qParams, NumericalValuePostgresRegex)
		boundsStatement = as(boundsStepName, boundsStatement)
		*qStmnt = joinWithComma(*qStmnt, boundsStatement)

		// Preparing 'bucketed' step with changing $none to NaN for float conversion.
		noneToNaN := fmt.Sprintf("COALESCE(NULLIF(COALESCE(NULLIF(%s, '%s'), ''), ''), 'NaN') AS %s, ",
			groupKey, model.PropertyValueNone, groupKey)

		// Adding width_bucket for each record, keeping -1 for $none.
		bucketKey := groupKey + "_bucket"
		stepBucket := fmt.Sprintf("CASE WHEN %s = '%s' THEN -1 WHEN %s = '' THEN -1 ELSE width_bucket(%s::numeric, %s.lbound::numeric, "+
			"COALESCE(NULLIF(%s.ubound, %s.lbound), %s.ubound+1)::numeric, %d) END AS %s, ",
			groupKey, model.PropertyValueNone, groupKey, groupKey, boundsStepName, boundsStepName,
			boundsStepName, boundsStepName, model.NumericalGroupByBuckets-2, bucketKey)

		// Creating bucket string to be used in group by. Also, replacing NaN-Nan to $none.
		aggregateSelectKeys = aggregateSelectKeys + fmt.Sprintf(
			"COALESCE(NULLIF(concat(round(min(%s::numeric), 1), '%s', round(max(%s::numeric), 1)), 'NaN%sNaN'), '%s') AS %s, ",
			groupKey, model.NumericalGroupBySeparator, groupKey, model.NumericalGroupBySeparator, model.PropertyValueNone, groupKey)
		bucketedSelect = bucketedSelect + noneToNaN + stepBucket
		boundStepNames = append(boundStepNames, boundsStepName)
		aggregateGroupBys = append(aggregateGroupBys, bucketKey)
		aggregateOrderBys = append(aggregateOrderBys, bucketKey)
		bucketedNumericValueFilter = append(bucketedNumericValueFilter,
			fmt.Sprintf("%s ~ ?", groupKey))
		*qParams = append(*qParams, NumericalValuePostgresRegex)
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
	qep model.QueryEventWithProperties, from int64, to int64, fromStr string,
	stepName string, addSelecStmnt string, addSelectParams []interface{},
	addJoinStmnt string, groupBy string, orderBy string) error {

	if (from == 0 && fromStr == "") || to == 0 {
		return errors.New("invalid timerange on events filter")
	}

	if addSelecStmnt == "" {
		return errors.New("invalid select on events filter")
	}

	rStmnt := "SELECT " + addSelecStmnt + " FROM events" + " " + addJoinStmnt

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
	wStmnt, wParams, err := buildWhereFromProperties(projectId, qep.Properties)
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

func hasWhereEntity(ewp model.QueryEventWithProperties, entity string) bool {
	for _, p := range ewp.Properties {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func joinWithComma(x ...string) string {
	return joinWithWordInBetween(",", x...)
}

func joinWithWordInBetween(word string, x ...string) string {
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
	var y []string
	for _, v := range x {
		y = append(y, fmt.Sprintf("%s %s ", v, suffix))
	}

	return y
}

func hasGroupEntity(props []model.QueryGroupByProperty, entity string) bool {
	for _, p := range props {
		if p.Entity == entity {
			return true
		}
	}

	return false
}

func addJoinLatestUserPropsQuery(projectID uint64, groupProps []model.QueryGroupByProperty,
	refStepName string, stepName string, qStmnt *string, qParams *[]interface{}, addSelect string) string {

	groupSelect, gSelectParams, gKeys := buildGroupKeys(projectID, groupProps)

	rStmnt := "SELECT " + joinWithComma(groupSelect, addSelect) + " from " + refStepName +
		" " + "LEFT JOIN users ON " + refStepName + ".event_user_id=users.id"

	if stepName != "" {
		rStmnt = as(stepName, rStmnt)
	}

	*qStmnt = appendStatement(*qStmnt, rStmnt)
	*qParams = append(*qParams, gSelectParams...)

	return gKeys
}

func filterGroupPropsByType(gp []model.QueryGroupByProperty, entity string) []model.QueryGroupByProperty {
	groupProps := make([]model.QueryGroupByProperty, 0)

	for _, v := range gp {
		if v.Entity == entity {
			groupProps = append(groupProps, v)
		}
	}
	return groupProps
}

func removeEventSpecificUserGroupBys(groupBys []model.QueryGroupByProperty) []model.QueryGroupByProperty {
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
	return fmt.Sprintf("%s ORDER BY %s DESC", qStmnt, model.AliasAggr)
}
func appendOrderByEventNameAndAggr(qStmnt string) string {
	return fmt.Sprintf("%s ORDER BY event_name, %s DESC", qStmnt, model.AliasAggr)
}

func appendSelectTimestampColIfRequired(stmnt string, isRequired bool) string {
	if !isRequired {
		return stmnt
	}

	return joinWithComma(stmnt, model.AliasDateTime)
}

func getSelectTimestampByType(timestampType, timezone string) string {
	var selectTz string

	if timezone == "" {
		selectTz = model.DefaultTimezone
	} else {
		selectTz = timezone
	}

	var selectStr string
	if timestampType == model.GroupByTimestampHour {
		selectStr = fmt.Sprintf("date_trunc('hour', to_timestamp(timestamp) AT TIME ZONE '%s')", selectTz)
	} else if timestampType == model.GroupByTimestampWeek {
		// default week is Monday to Sunday for postgres, updating it to Sunday to Saturday
		selectStr = fmt.Sprintf("date_trunc('week', to_timestamp(timestamp + (24*60*60)) AT TIME ZONE '%s') - INTERVAL '1 day' ", selectTz)
	} else if timestampType == model.GroupByTimestampMonth {
		selectStr = fmt.Sprintf("date_trunc('month', to_timestamp(timestamp) AT TIME ZONE '%s')", selectTz)
	} else if timestampType == model.GroupByTimestampQuarter {
		selectStr = fmt.Sprintf("date_trunc('quarter', to_timestamp(timestamp) AT TIME ZONE '%s')", selectTz)
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
		getSelectTimestampByType(groupByTimestamp, timezone), model.AliasDateTime))
}

func appendGroupByTimestampIfRequired(qStmnt string, isRequired bool, groupKeys ...string) string {
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
	if len(gKeys) == 0 || (len(gKeys) == 1 && gKeys[0] == "") {
		return qStmnt
	}

	return fmt.Sprintf("%s GROUP BY %s", qStmnt, joinWithComma(gKeys...))
}

func appendLimitByCondition(qStmnt string, groupProps []model.QueryGroupByProperty, groupByTimestamp bool) string {
	if len(groupProps) == 1 && !groupByTimestamp {
		return fmt.Sprintf("%s LIMIT %d", qStmnt, model.ResultsLimit)
	}

	// Limited with max limit on SQL. Limited on server side.
	return fmt.Sprintf("%s LIMIT %d", qStmnt, model.MaxResultsLimit)
}

func separateEventLevelGroupBys(allGroupBys []model.QueryGroupByProperty) (
	eventLevelGroupBys, otherGroupBys []model.QueryGroupByProperty) {
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
	return groupBy.EventName != "" && groupBy.EventNameIndex != 0
}

// TranslateGroupKeysIntoColumnNames - Replaces groupKeys on result
// headers with real column names.
func translateGroupKeysIntoColumnNames(result *model.QueryResult,
	groupProps []model.QueryGroupByProperty) error {

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
	for i := range query.GroupByProperties {
		query.GroupByProperties[i].Index = i
	}
}

func getGroupKeyIndexesForSlicing(cols []string) (int, int, error) {
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

	for _, gbtType := range model.GroupByTimestampTypes {
		if gbtType == groupByTimestamp {
			return true
		}
	}

	return false
}

// IsValidEventsQuery Validates query for Events class and returns corresponding errMsg.
func IsValidEventsQuery(query *model.Query) (bool, string) {
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

func getQueryCacheRedisKeySuffix(hashString string, from, to int64) string {
	if to-from == DateRangePreset2MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, DateRangePreset2MinLabel)
	} else if to-from == DateRangePreset30MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, DateRangePreset30MinLabel)
	} else if U.IsStartOfTodaysRange(from, U.TimeZoneStringIST) {
		return fmt.Sprintf("%s:from:%d", hashString, from)
	}
	return fmt.Sprintf("%s:from:%d:to:%d", hashString, from, to)
}

// GetQueryResultFromCache To get value from cache for a particular query payload.
// resultContainer to be passed by reference.
func GetQueryResultFromCache(projectID uint64, query model.BaseQuery, resultContainer interface{}) (model.QueryCacheResult, int) {
	logCtx := log.WithFields(log.Fields{
		"ProjectID": projectID,
	})

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
	} else if value == QueryCacheInProgressPlaceholder {
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

// ExecQueryWithContext Executes raw query with context. Useful to kill queries on program exit or crash.
func (pg *Postgres) ExecQueryWithContext(stmnt string, params []interface{}) (*sql.Rows, error) {
	db := C.GetServices().Db
	debugQuery := U.DBDebugPreparedStatement(stmnt, params)

	// For query: ...where id in ($1) where $1 is passed as a slice, convert to pq.Array()
	stmnt, params = model.ExpandArrayWithIndividualValues(stmnt, params)

	// Change ? in the query to to $1, $2 format.
	stmnt = model.TransformQueryPlaceholdersForContext(stmnt)

	rows, err := db.DB().QueryContext(*C.GetServices().DBContext, stmnt, params...)
	if C.GetConfig().Env == C.DEVELOPMENT || C.GetConfig().Env == C.TEST || err != nil {
		log.WithField("Query", debugQuery).Info("Exec query with context")
	}
	return rows, err
}

func (pg *Postgres) ExecQuery(stmnt string, params []interface{}) (*model.QueryResult, error) {
	rows, err := pg.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}

	result := &model.QueryResult{Headers: resultHeaders, Rows: resultRows}
	return result, nil
}

func addQueryToResultMeta(result *model.QueryResult, query model.Query) {
	result.Meta.Query = query
}

func isValidFunnelQuery(query *model.Query) bool {
	return len(query.EventsWithProperties) <= 6
}

// NOTE: TODO to different to queryGroup
func DecodeQueryForClass(queryJSON postgres.Jsonb, queryClass string) (model.BaseQuery, error) {
	var baseQuery model.BaseQuery
	var err error
	switch queryClass {
	case model.QueryClassFunnel, model.QueryClassInsights:
		var query model.Query
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case model.QueryClassAttribution:
		var query model.AttributionQueryUnit
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case model.QueryClassChannel:
		var query model.ChannelQueryUnit
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case model.QueryClassChannelV1:
		var query model.ChannelGroupQueryV1
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case model.QueryClassEvents:
		var query model.QueryGroup
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case model.QueryClassWeb:
		var query model.DashboardUnitsWebAnalyticsQuery
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	default:
		return baseQuery, fmt.Errorf("query class %s not supported", queryClass)
	}

	return baseQuery, err
}

func (pg *Postgres) Analyze(projectId uint64, queryOriginal model.Query) (*model.QueryResult, int, string) {
	var query model.Query
	U.DeepCopy(&queryOriginal, &query)

	if valid, errMsg := IsValidQuery(&query); !valid {
		return nil, http.StatusBadRequest, errMsg
	}

	if query.Class == model.QueryClassFunnel {
		return pg.RunFunnelQuery(projectId, query)
	}
	return pg.RunInsightsQuery(projectId, query)
}
