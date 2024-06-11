package memsql

import (
	"database/sql"
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func EventsPerformedCheck(projectID int64, segmentId string, eventNameIDsMap map[string]string,
	segmentQuery *model.Query, userID string, isAllAccounts bool, userArray []model.User) bool {

	// default value to be set true
	for index := range segmentQuery.EventsWithProperties {
		if segmentQuery.EventsWithProperties[index].FrequencyOperator != "" {
			continue
		}
		segmentQuery.EventsWithProperties[index].IsEventPerformed = true
	}

	isMatched := false

	var didEventPresent, didNotDoEventPresent bool
	var didEventCount, didNotEventCount int

	for _, event := range segmentQuery.EventsWithProperties {
		if event.IsEventPerformed {
			didEventCount = didEventCount + 1
		} else {
			didNotEventCount = didNotEventCount + 1
		}
	}

	if didEventCount > 0 {
		didEventPresent = true
	}
	if didNotEventCount > 0 {
		didNotDoEventPresent = true
	}

	if didEventPresent {
		isMatched = didEventQuery(projectID, segmentId, eventNameIDsMap, segmentQuery,
			userID, isAllAccounts, userArray, didEventCount)

		if segmentQuery.EventsCondition == model.EventCondAnyGivenEvent && isMatched {
			// if any event condition satisfies, return
			return isMatched
		} else if segmentQuery.EventsCondition == model.EventCondAllGivenEvent && !isMatched {
			// if any event condition does not satisfies, return
			return isMatched
		}
	}

	// return if not exist
	if !didNotDoEventPresent {
		return isMatched
	}

	// already returned in case of any (positive) and all (negative) case
	isMatched = didNotEventQuery(projectID, segmentId, eventNameIDsMap, segmentQuery,
		userID, isAllAccounts, userArray, didNotEventCount)

	return isMatched
}

func didEventQuery(projectID int64, segmentId string, eventNameIDsMap map[string]string,
	segmentQuery *model.Query, userID string, isAllAccounts bool, userArray []model.User, didEventCount int) bool {
	isMatched := false

	var userIDString string
	params := []interface{}{projectID}
	didEvent := true

	if isAllAccounts {
		userIDS := userIdsList(*segmentQuery, userArray, didEvent)
		userIDS = append(userIDS, userID)
		if len(userIDS) == 0 {
			return isMatched
		}
		userIDString = "user_id IN (?)"
		params = append(params, userIDS)
	} else {
		userIDString = "user_id=?"
		params = append(params, userID)
	}

	var query string
	var queryParams []interface{}
	var err error

	if C.MarkerEnableOptimisedEventsQuery(projectID) {
		query, queryParams, err = eventsQueryOptimised(projectID, eventNameIDsMap, segmentQuery, userID, userIDString, params, didEvent)
	} else {
		query, queryParams, err = eventsQuery(projectID, eventNameIDsMap, segmentQuery, userID, userIDString, params, didEvent)
	}

	if err != nil {
		return isMatched
	}

	startTime := time.Now().UnixMilli()

	result, status := GetStore().CheckIfUserPerformedGivenEvents(projectID, userID, query, queryParams)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "segment_id": segmentId}).
				Error("Error while validating for performed events")
		}
		return isMatched
	}

	endTime := time.Now().UnixMilli()
	timeTaken := endTime - startTime

	if C.MarkerEnableOptimisedEventsQuery(projectID) && timeTaken > 100 {
		log.WithFields(log.Fields{"project_id": projectID, "domain_id": userID, "segment_id": segmentId}).
			Warn("Time taken for execution of optimized events query ", timeTaken)
	}

	if segmentQuery.EventsCondition == model.EventCondAllGivenEvent {
		if len(result) == didEventCount {
			isMatched = true
		}
	} else {
		if len(result) > 0 {
			isMatched = true
		}
	}

	eventConditionMatched := false

	for index, event := range segmentQuery.EventsWithProperties {
		if event.FrequencyOperator == "" || event.Frequency == "" {
			continue
		}

		checkValue, err := strconv.ParseFloat(event.Frequency, 64)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "domain_id": userID, "segment_id": segmentId}).
				WithError(err).Error("Failed to convert property value to float type")
			isMatched = false
			continue
		}
		count := result[eventNameIDsMap[event.Name]]
		eventConditionMatched = NumericalPropCheck(event.FrequencyOperator, float64(count), checkValue)

		if index == 0 {
			isMatched = eventConditionMatched
			continue
		}

		if segmentQuery.EventsCondition == model.EventCondAnyGivenEvent {
			isMatched = isMatched || eventConditionMatched
			// if any event condition satisfies, break
			if isMatched {
				return isMatched
			}
		} else if segmentQuery.EventsCondition == model.EventCondAllGivenEvent {
			isMatched = isMatched && eventConditionMatched
			// if any event condition does not satisfies, break
			if !isMatched {
				return isMatched
			}
		}
	}

	return isMatched
}
func didNotEventQuery(projectID int64, segmentId string, eventNameIDsMap map[string]string,
	segmentQuery *model.Query, userID string, isAllAccounts bool, userArray []model.User, didNotEventCount int) bool {
	isMatched := false

	var userIDString string
	params := []interface{}{projectID}
	didEvent := false

	if isAllAccounts {
		userIDS := userIdsList(*segmentQuery, userArray, didEvent)
		userIDS = append(userIDS, userID)
		if len(userIDS) == 0 {
			return isMatched
		}
		userIDString = "user_id IN (?)"
		params = append(params, userIDS)
	} else {
		userIDString = "user_id=?"
		params = append(params, userID)
	}

	query, queryParams, err := eventsQuery(projectID, eventNameIDsMap, segmentQuery, userID, userIDString, params, didEvent)

	if err != nil {
		return isMatched
	}

	result, status := GetStore().CheckIfUserPerformedGivenEvents(projectID, userID, query, queryParams)
	if status == http.StatusInternalServerError {
		log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "segment_id": segmentId}).
			Error("Error while validating for did not performed events")
	}

	// if any record found, then return false for all_events
	// if len of result less than didNotEventCount, return true for any_events
	if (segmentQuery.EventsCondition == model.EventCondAnyGivenEvent && len(result) < didNotEventCount) ||
		(segmentQuery.EventsCondition == model.EventCondAllGivenEvent && len(result) == 0) {
		isMatched = true
	}

	return isMatched
}

// SELECT COUNT(event_name_id), event_name_id FROM events
// WHERE project_id = '15000001' AND user_id='4cf0f3f9-b708-4c32-a8ca-b90af412fa55'
// AND ((event_name_id='34a0c60e-1d83-498b-b2ab-5a112a6d812d' AND
// (JSON_EXTRACT_STRING(events.user_properties, '$country') RLIKE 'India')) OR
// (event_name_id='f82d3153-c091-471a-9021-29277dc03647') OR
// (event_name_id='dd092c62-b2f0-4f2a-b780-ffc158d0f4b9') ) GROUP BY event_name_id;

func eventsQuery(projectID int64, eventNameIDsMap map[string]string, segmentQuery *model.Query, userID string,
	userIDString string, qParams []interface{}, didEvent bool) (string, []interface{}, error) {
	query := fmt.Sprintf(`SELECT COUNT(event_name_id), event_name_id
	FROM events
	WHERE project_id = ? AND %s
	  AND (`, userIDString)

	params := qParams

	var eventStr string
	firstEventFound := true
	for _, event := range segmentQuery.EventsWithProperties {
		if event.IsEventPerformed != didEvent {
			continue
		}

		queryStr := "(event_name_id=?"
		params = append(params, eventNameIDsMap[event.Name])

		// support for in last x days
		if event.Range > 0 {
			queryStr = queryStr + " AND timestamp >= ?"
			params = append(params, event.Range)
		}

		if len(event.Properties) > 0 {
			whereCond, queryParams, err := buildWhereFromProperties(projectID, event.Properties, 0)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "user_id": userID}).WithError(err).
					Error("Failed to build where condition for performed events check")
				return query, []interface{}{}, err
			}
			if len(queryParams) > 0 {
				queryStr = queryStr + " AND " + whereCond + ")"
				params = append(params, queryParams...)

			}
		} else {
			queryStr = queryStr + ")"
		}

		if firstEventFound {
			eventStr = queryStr
			firstEventFound = !firstEventFound
			continue
		}
		eventStr = eventStr + " OR " + queryStr
	}

	query = query + eventStr + " ) GROUP BY event_name_id;"

	return query, params, nil
}

// WITH selected_events AS
//
//	(SELECT event_name_id, JSON_INCLUDE_MASK(properties, '{"$hubspot_engagement_from":1}' ) as properties,
//	JSON_INCLUDE_MASK(user_properties, '{"$hubspot_company_created":1}' ) as user_properties
//	FROM events WHERE project_id='8000024'
//	AND user_id IN ('378f220c-6311-40e8-bad0-88e939395ad1','778f9212-b0f2-4a5d-aa55-7db6af30938e',
//	'ca1197ac-f5ce-4bb9-9599-a989b4c3d5bc')
//	AND event_name_id IN ('6cc5ed17-09ff-4ee0-bdf1-844e97cfb27f','9fed79d3-1322-4393-a644-73c3e56826da')
//	ORDER BY timestamp DESC LIMIT 100000)
//
// SELECT COUNT(event_name_id), event_name_id FROM selected_events
// WHERE (event_name_id='6cc5ed17-09ff-4ee0-bdf1-844e97cfb27f' AND
// (JSON_EXTRACT_STRING(selected_events.properties, '$hubspot_engagement_from') = 'Somewhere')) OR
// (event_name_id='9fed79d3-1322-4393-a644-73c3e56826da' AND
// (JSON_EXTRACT_STRING(selected_events.user_properties, '$hubspot_company_created') = 'Heyflow' OR
// JSON_EXTRACT_STRING(selected_events.user_properties, '$hubspot_company_created') = 'ChargeBee'))
// GROUP BY event_name_id;

func eventsQueryOptimised(projectID int64, eventNameIDsMap map[string]string, segmentQuery *model.Query, userID string,
	userIDString string, params []interface{}, didEvent bool) (string, []interface{}, error) {
	query := `WITH selected_events AS 
	(SELECT event_name_id`

	var eventStr string
	eventNameIDsParams := make([]interface{}, 0)
	qParams := make([]interface{}, 0)
	eventPropertiesFilter := map[string]int{}
	userPropertiesFilter := map[string]int{}

	timestampRequired := false
	firstEventFound := true
	for _, event := range segmentQuery.EventsWithProperties {
		if didEvent != event.IsEventPerformed {
			continue
		}

		queryStr := "(event_name_id=?"
		qParams = append(qParams, eventNameIDsMap[event.Name])

		// support for in last x days
		if event.Range > 0 {
			queryStr = queryStr + " AND timestamp >=?"
			qParams = append(qParams, event.Range)
			timestampRequired = true
		}
		if len(event.Properties) > 0 {
			whereCond, queryParams, err := buildWhereFromProperties(projectID, event.Properties, 0)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "user_id": userID}).WithError(err).
					Error("Failed to build where condition for performed events check")
				return query, []interface{}{}, err
			}
			if len(queryParams) > 0 {
				queryStr = queryStr + " AND " + whereCond + ")"
				qParams = append(qParams, queryParams...)

			}
		} else {
			queryStr = queryStr + ")"
		}

		eventNameIDsParams = append(eventNameIDsParams, eventNameIDsMap[event.Name])

		for _, eventFilter := range event.Properties {
			if eventFilter.Entity == model.PropertyEntityEvent {
				eventPropertiesFilter[eventFilter.Property] = 1
			} else if eventFilter.Entity == model.PropertyEntityUser {
				userPropertiesFilter[eventFilter.Property] = 1
			}
		}

		if firstEventFound {
			eventStr = queryStr
			firstEventFound = !firstEventFound
			continue
		}

		eventStr = eventStr + " OR " + queryStr
	}

	// add timestamp column if required
	if timestampRequired {
		query = query + ", timestamp"
	}

	// event props support
	if len(eventPropertiesFilter) > 0 {
		fieldsInclude, err := json.Marshal(eventPropertiesFilter)
		if err != nil {
			return query, params, err
		}

		query = query + fmt.Sprintf(", JSON_INCLUDE_MASK(properties, '%s' ) as properties", string(fieldsInclude))
	}

	// user props support
	if len(userPropertiesFilter) > 0 {
		fieldsInclude, err := json.Marshal(userPropertiesFilter)
		if err != nil {
			return query, params, err
		}

		query = query + fmt.Sprintf(", JSON_INCLUDE_MASK(user_properties, '%s' ) as user_properties", string(fieldsInclude))
	}

	eventStr = strings.ReplaceAll(eventStr, "events.", "selected_events.")

	params = append(params, eventNameIDsParams)

	query = query + ` FROM events WHERE project_id=? 
	AND user_id IN (?) 
	AND event_name_id IN (?)
	ORDER BY timestamp DESC LIMIT 100000)
	SELECT COUNT(event_name_id), event_name_id FROM selected_events 
	WHERE ` + eventStr +
		` GROUP BY event_name_id;`

	params = append(params, qParams...)

	return query, params, nil
}

func IsRuleMatchedAllAccounts(projectID int64, segment model.Query, decodedProperties []map[string]interface{}, userArr []model.User,
	segmentId string, domId string, eventNameIDsMap map[string]string, fileValuesMap map[string]map[string]bool) bool {
	// isMatched = all rules matched (a or b) AND (c or d)
	isMatched := false

	if (segment.GlobalUserProperties == nil || len(segment.GlobalUserProperties) == 0) &&
		(segment.EventsWithProperties != nil && len(segment.EventsWithProperties) > 0) {
		isMatched = PerformedEventsCheck(projectID, segmentId, eventNameIDsMap, &segment, domId, userArr)
		return isMatched
	}

	groupedProperties := model.GetPropertiesGrouped(segment.GlobalUserProperties)
	for index, currentGroupedProperties := range groupedProperties {
		// validity for each group like (a or b ) AND (c or d)
		groupedPropsMatched := false
		for _, p := range currentGroupedProperties {
			isValueFound := CheckPropertyInAllUsers(projectID, p, decodedProperties, userArr, fileValuesMap)
			if isValueFound {
				groupedPropsMatched = true
				break
			}
		}
		if index == 0 {
			isMatched = groupedPropsMatched
			continue
		}
		isMatched = groupedPropsMatched && isMatched
	}

	if isMatched && (segment.EventsWithProperties != nil && len(segment.EventsWithProperties) > 0) {
		isMatched = PerformedEventsCheck(projectID, segmentId, eventNameIDsMap, &segment, domId, userArr)
	}
	return isMatched
}

func PerformedEventsCheck(projectID int64, segmentID string, eventNameIDsMap map[string]string,
	segmentQuery *model.Query, userID string, userArr []model.User) bool {

	isPerformedEvent, isAllAccounts := false, false

	isAllAccounts = (segmentQuery.Caller == model.PROFILE_TYPE_ACCOUNT)
	isPerformedEvent = EventsPerformedCheck(projectID, segmentID, eventNameIDsMap, segmentQuery, userID, isAllAccounts, userArr)

	return isPerformedEvent
}

func CheckPropertyInAllUsers(projectId int64, p model.QueryProperty, decodedProperties []map[string]interface{},
	userArr []model.User, fileValuesMap map[string]map[string]bool) bool {
	isValueFound := false
	for index, user := range userArr {
		// skip for group user if entity is user_group or entity is user_g and is not a group user
		if (p.Entity == model.PropertyEntityUserGroup && (user.IsGroupUser != nil && *user.IsGroupUser)) ||
			(p.Entity == model.PropertyEntityUserGlobal && (user.IsGroupUser == nil || !*user.IsGroupUser)) {
			continue
		}

		// group based filtering for engagement properties
		if p.GroupName == U.GROUP_NAME_DOMAINS && user.Source != nil && *user.Source != model.UserSourceDomains {
			continue
		}

		isValueFound = CheckPropertyOfGivenType(projectId, p, &decodedProperties[index], fileValuesMap)

		// check for negative filters
		if (p.Operator == model.NotContainsOpStr && p.Value != model.PropertyValueNone) ||
			(p.Operator == model.ContainsOpStr && p.Value == model.PropertyValueNone) ||
			(p.Operator == model.NotEqualOpStr && p.Value != model.PropertyValueNone) ||
			(p.Operator == model.EqualsOpStr && p.Value == model.PropertyValueNone) ||
			(p.Operator == model.NotInList && p.Value != model.PropertyValueNone) {
			if !isValueFound {
				return isValueFound
			}
			continue
		}

		if isValueFound {
			return isValueFound
		}
	}
	return isValueFound
}

func (store *MemSQL) CheckIfUserPerformedGivenEvents(projectID int64, userID string, queryStr string, params []interface{}) (map[string]int, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"user_id":    userID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	results := make(map[string]int)

	db := C.GetServices().Db
	rows, err := db.Raw(queryStr, params...).Rows()

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return map[string]int{}, http.StatusNotFound
		}
		log.WithFields(logFields).WithError(err).Error("Error fetching records")
		return map[string]int{}, http.StatusInternalServerError
	}

	for rows.Next() {
		var event_name_id string
		var count int
		err = rows.Scan(&count, &event_name_id)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("Error fetching rows")
			return map[string]int{}, http.StatusInternalServerError
		}
		results[event_name_id] = count
	}

	if len(results) == 0 {
		return map[string]int{}, http.StatusNotFound
	}

	return results, http.StatusFound
}

func userIdsList(query model.Query, userArray []model.User, didEventsList bool) []string {
	groupIDs := GetEventGroupIds(query, didEventsList)
	userIDsMap := userIdMap(userArray)

	isGroupAdded := make(map[int]bool)
	var userIdsList []string
	for _, groupID := range groupIDs {
		if isGroupAdded[groupID] {
			continue
		}
		userIdsList = append(userIdsList, userIDsMap[groupID]...)
		isGroupAdded[groupID] = true
	}
	return userIdsList
}

func userIdMap(userArray []model.User) map[int][]string {
	userIdMap := make(map[int][]string)
	// grouping all users together, and all groups depending on their source
	// here 0 indicated user
	source := 0
	for _, user := range userArray {
		if user.IsGroupUser != nil && *user.IsGroupUser {
			source = *user.Source
		} else {
			source = 0
		}
		userIdMap[source] = append(userIdMap[source], user.ID)
	}

	return userIdMap
}

func GetEventGroupIds(query model.Query, didEventsList bool) []int {
	groupIds := make([]int, 0)
	for _, event := range query.EventsWithProperties {
		if event.IsEventPerformed != didEventsList {
			continue
		}
		groupName := U.GetGroupNameFromGroupEventName(event.Name)
		if groupName != "" {
			groupIds = append(groupIds, model.GroupUserSource[groupName])
		} else {
			groupIds = append(groupIds, 0)
		}
	}
	return groupIds
}

// SELECT id, is_group_user, group_4_user_id, group_1_id, group_4_id, associated_segments,
// customer_user_id FROM users WHERE project_id=15000011 LIMIT 50
func (store *MemSQL) FetchAssociatedSegmentsFromUsers(projectID int64) (int, []model.User, []map[string]interface{}) {
	if projectID == 0 {
		log.WithField("project_id", projectID).Error("invalid project")
		return http.StatusBadRequest, nil, nil
	}

	query := `SELECT id, is_group_user, group_4_user_id, group_1_id, group_4_id, associated_segments,
	customer_user_id FROM users WHERE project_id=? LIMIT 50`
	db := C.GetServices().Db

	rows, err := db.Raw(query, projectID).Rows()
	if err != nil {
		return http.StatusInternalServerError, nil, nil
	}

	users := make([]model.User, 0)
	associatedSegmentsList := make([]map[string]interface{}, 0)

	for rows.Next() {
		var id string
		var is_group_user_null sql.NullBool
		var group_4_user_id_null sql.NullString
		var group_2_id_null sql.NullString
		var group_4_id_null sql.NullString
		var associated_segments *postgres.Jsonb
		var customer_user_id_null sql.NullString

		if err = rows.Scan(&id,
			&is_group_user_null, &group_4_user_id_null,
			&group_2_id_null, &group_4_id_null, &associated_segments, &customer_user_id_null); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return http.StatusBadRequest, nil, nil
		}

		var associatedSegmentsMap map[string]interface{}
		if associated_segments != nil {
			associatedSegmentsBytes, err := associated_segments.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal associated_segments")
				return http.StatusInternalServerError, nil, nil
			}

			err = json.Unmarshal(associatedSegmentsBytes.([]byte), &associatedSegmentsMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal associated_segments")
				return http.StatusInternalServerError, nil, nil
			}
		}

		is_group_user := U.IfThenElse(is_group_user_null.Valid, is_group_user_null.Bool, false).(bool)
		group_4_user_id := U.IfThenElse(group_4_user_id_null.Valid, group_4_user_id_null.String, "").(string)
		group_2_id := U.IfThenElse(group_2_id_null.Valid, group_2_id_null.String, "").(string)
		group_4_id := U.IfThenElse(group_4_id_null.Valid, group_4_id_null.String, "").(string)
		customer_user_id := U.IfThenElse(customer_user_id_null.Valid, customer_user_id_null.String, "").(string)

		user := model.User{
			ID:             id,
			IsGroupUser:    &is_group_user,
			Group2ID:       group_2_id,
			Group4UserID:   group_4_user_id,
			Group4ID:       group_4_id,
			CustomerUserId: customer_user_id,
		}

		users = append(users, user)
		associatedSegmentsList = append(associatedSegmentsList, associatedSegmentsMap)
	}

	return http.StatusFound, users, associatedSegmentsList
}

func CheckPropertyOfGivenType(projectId int64, p model.QueryProperty, decodedProperties *map[string]interface{},
	fileValuesMap map[string]map[string]bool) bool {
	isValueFound := false
	if p.Value != model.PropertyValueNone {
		if p.Type == U.PropertyTypeDateTime {
			isValueFound = checkDateTypeProperty(p, decodedProperties)
		} else if p.Type == U.PropertyTypeNumerical {
			isValueFound = checkNumericalTypeProperty(p, decodedProperties)
		} else {
			isValueFound = checkCategoricalTypeProperty(projectId, p, decodedProperties, fileValuesMap)
		}
	} else {
		// where condition for $none value.
		propertyValue := (*decodedProperties)[p.Property]
		if p.Operator == model.EqualsOpStr || p.Operator == model.ContainsOpStr {
			if propertyValue == nil || propertyValue == "" {
				isValueFound = true
			}
		} else if p.Operator == model.NotEqualOpStr || p.Operator == model.NotContainsOpStr {
			if propertyValue != nil && propertyValue != "" {
				isValueFound = true
			}
		}
	}
	return isValueFound
}

func checkNumericalTypeProperty(segmentRule model.QueryProperty, properties *map[string]interface{}) bool {
	if _, exists := (*properties)[segmentRule.Property]; !exists {
		return false
	}
	var propertyExists bool
	checkValue, err := strconv.ParseFloat(segmentRule.Value, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert payload value to float type")
	}
	var propertyValue float64
	if floatVal, ok := (*properties)[segmentRule.Property].(float64); ok {
		propertyValue = floatVal
	} else if stringVal, ok := (*properties)[segmentRule.Property].(string); ok {
		if stringVal != "" {
			propertyValue, err = strconv.ParseFloat(stringVal, 64)
			if err != nil {
				log.WithError(err).Error("Failed to convert property value to float type")
			}
		}
	} else if intVal, ok := (*properties)[segmentRule.Property].(int64); ok {
		propertyValue = float64(intVal)
	}

	propertyExists = NumericalPropCheck(segmentRule.Operator, propertyValue, checkValue)

	return propertyExists
}

func NumericalPropCheck(operator string, propertyValue float64, checkValue float64) bool {
	var propertyExists bool
	switch operator {
	case model.EqualsOpStr:
		if propertyValue == checkValue {
			propertyExists = true
		}
	case model.NotEqualOpStr:
		if propertyValue != checkValue {
			propertyExists = true
		}
	case model.GreaterThanOpStr:
		if propertyValue > checkValue {
			propertyExists = true
		}
	case model.LesserThanOpStr:
		if propertyValue < checkValue {
			propertyExists = true
		}
	case model.GreaterThanOrEqualOpStr:
		if propertyValue >= checkValue {
			propertyExists = true
		}
	case model.LesserThanOrEqualOpStr:
		if propertyValue <= checkValue {
			propertyExists = true
		}
	default:
		propertyExists = false
	}

	return propertyExists
}

func checkCategoricalTypeProperty(projectId int64, segmentRule model.QueryProperty, properties *map[string]interface{},
	fileValuesMap map[string]map[string]bool) bool {

	if (segmentRule.Operator == model.NotContainsOpStr && segmentRule.Value != model.PropertyValueNone) ||
		(segmentRule.Operator == model.ContainsOpStr && segmentRule.Value == model.PropertyValueNone) ||
		(segmentRule.Operator == model.NotEqualOpStr && segmentRule.Value != model.PropertyValueNone) ||
		(segmentRule.Operator == model.EqualsOpStr && segmentRule.Value == model.PropertyValueNone) ||
		(segmentRule.Operator == model.NotInList && segmentRule.Value != model.PropertyValueNone) {
		if _, exists := (*properties)[segmentRule.Property]; !exists {
			return true
		}
	}

	if _, exists := (*properties)[segmentRule.Property]; !exists {
		return false
	}
	var propertyExists bool
	checkValue := segmentRule.Value
	var propertyValue string
	if stringVal, ok := (*properties)[segmentRule.Property].(string); ok {
		if stringVal != "" {
			propertyValue = stringVal
		}
	} else if floatVal, ok := (*properties)[segmentRule.Property].(float64); ok {
		propertyValue = fmt.Sprintf("%v", floatVal)
	} else if intVal, ok := (*properties)[segmentRule.Property].(int64); ok {
		propertyValue = fmt.Sprintf("%d", intVal)
	} else if boolVal, ok := (*properties)[segmentRule.Property].(bool); ok {
		propertyValue = strconv.FormatBool(boolVal)
	}
	switch segmentRule.Operator {
	case model.ContainsOpStr:
		if strings.Contains(propertyValue, checkValue) {
			propertyExists = true
		}
	case model.NotContainsOpStr:
		if !strings.Contains(propertyValue, checkValue) {
			propertyExists = true
		}
	case model.EqualsOpStr:
		if propertyValue == checkValue {
			propertyExists = true
		}
	case model.NotEqualOpStr:
		if propertyValue != checkValue {
			propertyExists = true
		}
	case model.InList:

		propertyExists = false
		if fileMap, mapExists := fileValuesMap[segmentRule.Value]; mapExists {
			if _, valExists := fileMap[propertyValue]; valExists {
				propertyExists = true
			}
		}

	case model.NotInList:

		propertyExists = true
		if fileMap, mapExists := fileValuesMap[segmentRule.Value]; mapExists {
			if _, valExists := fileMap[propertyValue]; valExists {
				propertyExists = false
			}
		}

	default:
		propertyExists = false
	}
	return propertyExists
}

func checkDateTypeProperty(segmentRule model.QueryProperty, properties *map[string]interface{}) bool {
	if _, exists := (*properties)[segmentRule.Property]; !exists {
		return false
	}
	var propertyExists bool
	checkValue, err := model.DecodeDateTimePropertyValue(segmentRule.Value)
	if err != nil {
		log.WithError(err).Error("Failed to convert datetime property from payload.")
	}
	var propertyValue int64
	if floatVal, ok := (*properties)[segmentRule.Property].(float64); ok {
		propertyValue = int64(floatVal)
	} else if stringVal, ok := (*properties)[segmentRule.Property].(string); ok {
		if stringVal != "" {
			propertyValue, err = strconv.ParseInt(stringVal, 10, 64)
			if err != nil {
				log.WithError(err).Error("Failed to convert datetime property from properties.")
			}
		}
	} else if intVal, ok := (*properties)[segmentRule.Property].(int64); ok {
		propertyValue = intVal
	}
	switch segmentRule.Operator {
	case model.BeforeStr:
		if propertyValue < checkValue.To {
			propertyExists = true
		}
	case model.NotInCurrent:
		if propertyValue < checkValue.From {
			propertyExists = true
		}
	case model.SinceStr, model.InCurrent:
		if propertyValue >= checkValue.From {
			propertyExists = true
		}
	case model.EqualsOpStr, model.BetweenStr, model.InPrevious, model.InLastStr:
		if checkValue.From <= propertyValue && propertyValue <= checkValue.To {
			propertyExists = true
		}
	case model.NotInBetweenStr, model.NotInPrevious, model.NotInLastStr:
		if !(checkValue.From <= propertyValue && propertyValue <= checkValue.To) {
			propertyExists = true
		}
	default:
		propertyExists = false
	}
	return propertyExists
}

func (store *MemSQL) GetAllPropertiesForDomain(projectID int64, domainGroupId int,
	domainID string, userCount *int64) ([]model.User, int) {

	userStmnt := "AND (is_group_user IS NULL OR is_group_user=0) ORDER BY properties_updated_timestamp DESC LIMIT 100"
	grpUserStmnt := "AND is_group_user=1 ORDER BY properties_updated_timestamp DESC LIMIT 50"

	// fetching top 100 non group users
	users, status := store.GetUsersAssociatedToDomainList(projectID, domainGroupId, domainID, userStmnt)

	if status == http.StatusInternalServerError {
		log.WithField("project_id", projectID).Error("Unable to find users for domain ", domainID)
		return []model.User{}, status
	}

	// fetching top 50 group users
	grpUsers, status := store.GetUsersAssociatedToDomainList(projectID, domainGroupId, domainID, grpUserStmnt)

	if status == http.StatusInternalServerError || (len(users) == 0 && len(grpUsers) == 0) {
		log.WithField("project_id", projectID).Error("Unable to find users for domain ", domainID)
		return []model.User{}, status
	}

	if len(grpUsers) > 0 {
		users = append(users, grpUsers...)
	}

	*userCount = *userCount + int64(len(users))

	// appending domain details to process domain based filters
	domDetails, status := store.GetDomainDetailsByID(projectID, domainID, domainGroupId)

	if status != http.StatusFound {
		log.WithField("project_id", projectID).Error("Unable to find details for domain %s", domainID)
	} else {
		users = append(users, domDetails)
	}

	return users, http.StatusFound
}

func (store *MemSQL) GetAllGroupPropertiesForDomain(projectID int64, domainGroupId int,
	domainID string) ([]model.User, int) {

	grpUserStmnt := "AND is_group_user=1 ORDER BY properties_updated_timestamp DESC LIMIT 50"
	// fetching top 50 group users
	users, status := store.GetUsersAssociatedToDomainList(projectID, domainGroupId, domainID, grpUserStmnt)

	if status == http.StatusInternalServerError || len(users) == 0 {
		log.WithField("project_id", projectID).Error("Unable to find users for domain ", domainID)
		return []model.User{}, status
	}

	// appending domain details to process domain based filters
	domDetails, status := store.GetDomainDetailsByID(projectID, domainID, domainGroupId)

	if status != http.StatusFound {
		log.WithField("project_id", projectID).Errorf("Unable to find details for domain %s", domainID)
	} else {
		users = append(users, domDetails)
	}

	return users, http.StatusFound
}

func GetFileValues(projectID int64, globalUserProperties []model.QueryProperty,
	fileValues map[string]map[string]bool) map[string]map[string]bool {
	allFilesValuesMap := fileValues

	for _, filter := range globalUserProperties {
		if filter.Operator != model.InList && filter.Operator != model.NotInList {
			continue
		}

		// file values already exists in the map
		if _, exists := allFilesValuesMap[filter.Value]; exists {
			continue
		}

		fileValueMap := make(map[string]bool)

		valueListString := GetValueListFromFile(projectID, filter)
		valueList := strings.Split(valueListString, " , ")
		for _, val := range valueList {
			valString := strings.ReplaceAll(U.TrimSingleQuotes(strings.TrimSpace(val)), "\\", "")
			fileValueMap[valString] = true
		}

		allFilesValuesMap[filter.Value] = fileValueMap
	}

	return allFilesValuesMap
}

func (store *MemSQL) GetNewlyUpdatedAndCreatedSegments() (map[int64][]string, error) {
	query := `SELECT segments.project_id AS project_id, segments.id AS id FROM segments
	JOIN project_settings 
	ON segments.project_id = project_settings.project_id
	WHERE segments.updated_at > project_settings.marker_last_run_all_accounts
	AND project_settings.marker_last_run_all_accounts > ?
	AND segments.updated_at > segments.marker_run_segment LIMIT 5000000;`

	var segments []model.Segment
	db := C.GetServices().Db
	err := db.Raw(query, U.DefaultTime()).Scan(&segments).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return map[int64][]string{}, nil
		}
		log.WithError(err).Error("Failed to get newly updated and created segments")
		return map[int64][]string{}, err
	}

	segmentsMap := make(map[int64][]string)

	for _, segment := range segments {
		if _, exists := segmentsMap[segment.ProjectID]; !exists {
			segmentsMap[segment.ProjectID] = make([]string, 0)
		}
		segmentsMap[segment.ProjectID] = append(segmentsMap[segment.ProjectID], segment.Id)
	}

	return segmentsMap, nil
}
