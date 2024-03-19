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

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func EventsPerformedCheck(projectID int64, segmentId string, eventNameIDsMap map[string]string,
	segmentQuery *model.Query, userID string, isAllAccounts bool, userArray []model.User) bool {

	isMatched := false
	var userIDString string
	params := []interface{}{projectID}

	if isAllAccounts {
		userIDS := userIdsList(*segmentQuery, userArray)
		if len(userIDS) == 0 {
			return isMatched
		}
		userIDString = "user_id IN (?)"
		params = append(params, userIDS)
	} else {
		userIDString = "user_id=?"
		params = append(params, userID)
	}

	query := fmt.Sprintf(`SELECT COUNT(event_name_id)
	FROM events
	WHERE project_id = ? AND %s
	  AND (`, userIDString)

	var eventStr string
	for index, event := range segmentQuery.EventsWithProperties {

		queryStr := "(event_name_id=?"
		params = append(params, eventNameIDsMap[event.Name])
		if len(event.Properties) > 0 {
			whereCond, queryParams, err := buildWhereFromProperties(projectID, event.Properties, 0)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectID, "user_id": userID}).WithError(err).
					Error("Failed to build where condition for performed events check")
				return isMatched
			}
			if len(queryParams) > 0 {
				queryStr = queryStr + " AND " + whereCond + ")"
				params = append(params, queryParams...)

			}
		} else {
			queryStr = queryStr + ")"
		}

		if index == 0 {
			eventStr = queryStr
			continue
		}
		eventStr = eventStr + " OR " + queryStr
	}

	query = query + eventStr + " ) GROUP BY event_name_id;"

	result, status := GetStore().CheckIfUserPerformedGivenEvents(query, params)
	if status != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "segment_id": segmentId}).
			Error("Error while validating for performed events")
	}

	if segmentQuery.EventsCondition == model.EventCondAllGivenEvent {
		if len(result) == len(segmentQuery.EventsWithProperties) {
			isMatched = true
		}
	} else {
		if len(result) > 0 {
			isMatched = true
		}
	}

	return isMatched
}

func IsRuleMatchedAllAccounts(projectID int64, segment model.Query, decodedProperties []map[string]interface{}, userArr []model.User,
	segmentId string, domId string, eventNameIDsMap map[string]string) bool {
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
			isValueFound := CheckPropertyInAllUsers(projectID, p, decodedProperties, userArr)
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

func CheckPropertyInAllUsers(projectId int64, p model.QueryProperty, decodedProperties []map[string]interface{}, userArr []model.User) bool {
	isValueFound := false
	for index, user := range userArr {
		// skip for group user if entity is user_group or entity is user_g and is not a group user
		if (p.Entity == model.PropertyEntityUserGroup && (user.IsGroupUser != nil && *user.IsGroupUser)) ||
			(p.Entity == model.PropertyEntityUserGlobal && (user.IsGroupUser == nil || !*user.IsGroupUser)) {
			continue
		}
		isValueFound = CheckPropertyOfGivenType(projectId, p, &decodedProperties[index])

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

// SELECT COUNT(event_name_id) FROM events
// WHERE project_id = '15000001' AND user_id='4cf0f3f9-b708-4c32-a8ca-b90af412fa55'
// AND ((event_name_id='34a0c60e-1d83-498b-b2ab-5a112a6d812d' AND
// (JSON_EXTRACT_STRING(events.user_properties, '$country') RLIKE 'India')) OR
// (event_name_id='f82d3153-c091-471a-9021-29277dc03647') OR
// (event_name_id='dd092c62-b2f0-4f2a-b780-ffc158d0f4b9') ) GROUP BY event_name_id;
func (store *MemSQL) CheckIfUserPerformedGivenEvents(queryStr string, params []interface{}) ([]int, int) {
	db := C.GetServices().Db

	var result []int
	if err := db.Raw(queryStr, params...).Scan(&result).Error; err != nil {
		return result, http.StatusInternalServerError
	}

	return result, http.StatusFound
}

func userIdsList(query model.Query, userArray []model.User) []string {
	groupIDs := GetEventGroupIds(query)
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

func GetEventGroupIds(query model.Query) []int {
	groupIds := make([]int, 0)
	for _, event := range query.EventsWithProperties {
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

func CheckPropertyOfGivenType(projectId int64, p model.QueryProperty, decodedProperties *map[string]interface{}) bool {
	isValueFound := false
	if p.Value != model.PropertyValueNone {
		if p.Type == U.PropertyTypeDateTime {
			isValueFound = checkDateTypeProperty(p, decodedProperties)
		} else if p.Type == U.PropertyTypeNumerical {
			isValueFound = checkNumericalTypeProperty(p, decodedProperties)
		} else {
			isValueFound = checkCategoricalTypeProperty(projectId, p, decodedProperties)
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
	switch segmentRule.Operator {
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

func checkCategoricalTypeProperty(projectId int64, segmentRule model.QueryProperty, properties *map[string]interface{}) bool {

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
		valueListString := GetValueListFromFile(projectId, segmentRule)
		valueList := strings.Split(valueListString, " , ")
		for _, val := range valueList {
			if strings.ReplaceAll(U.TrimSingleQuotes(strings.TrimSpace(val)), "\\", "") == propertyValue {
				propertyExists = true
			}
		}

	case model.NotInList:

		propertyExists = true
		valueListString := GetValueListFromFile(projectId, segmentRule)
		valueList := strings.Split(valueListString, " , ")
		for _, val := range valueList {
			if strings.ReplaceAll(U.TrimSingleQuotes(strings.TrimSpace(val)), "\\", "") == propertyValue {
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
		if !(checkValue.From <= propertyValue) && !(propertyValue <= checkValue.To) {
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
