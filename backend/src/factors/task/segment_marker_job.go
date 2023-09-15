package task

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	C "factors/config"

	log "github.com/sirupsen/logrus"

	"factors/model/model"

	"factors/model/store"

	U "factors/util"
)

func SegmentMarker(projectID int64) int {

	domainGroup, status := store.GetStore().GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if status != http.StatusFound {
		log.Error("Domain group not enabled")
		return status
	}

	var lookBack time.Time

	if !C.UseLookbackForSegmentMarker() {
		lookBack, _ = store.GetStore().GetSegmentMarkerLastRunTime(projectID)
	} else if lookBack.IsZero() {
		hour := C.LookbackForSegmentMarker()
		lookBack = U.TimeNowZ().Add(time.Duration(-hour) * time.Hour)
	}

	// list of 1000 domains their associated users with last updated_at in last 1 hour
	users, statusCode := store.GetStore().GetUsersUpdatedAtGivenHour(projectID, lookBack, domainGroup.ID)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("Couldn't find updated records in last one hour")
		return statusCode
	}

	// list of all segments
	allSegmentsMap, statusCode := store.GetStore().GetAllSegments(projectID)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("No segment found for this project")
		return statusCode
	}

	// set Query Timezone
	timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("Failed to get Timezone.")
		return http.StatusBadRequest
	}

	// decoding all json segment rules
	decodedSegmentRulesMap := make(map[string][]model.Query)
	for groupName, segmentArr := range allSegmentsMap {
		decodedSegmentRulesMap[groupName] = make([]model.Query, 0)
		for _, segment := range segmentArr {
			segmentQuery := model.Query{}
			err := U.DecodePostgresJsonbToStructType(segment.Query, &segmentQuery)
			if err != nil {
				log.WithField("project_id", projectID).Error("Unable to decode segment query")
				return http.StatusInternalServerError
			}
			// datetime conversion
			err = segmentQuery.TransformDateTypeFilters()
			if err != nil {
				log.WithField("project_id", projectID).Error("Failed to transform segment filters.")
				return http.StatusBadRequest
			}
			segmentQuery.Timezone = string(timezoneString)
			segmentQuery.Source = segment.Type
			if segmentQuery.Caller == model.USER_PROFILES {
				segmentQuery.GroupAnalysis = model.FILTER_TYPE_USERS
			} else if segmentQuery.Caller == model.ACCOUNT_PROFILES {
				segmentQuery.GroupAnalysis = segment.Type
			}
			decodedSegmentRulesMap[groupName] = append(decodedSegmentRulesMap[groupName], segmentQuery)
		}
	}

	statusMap := make(map[string]int, 0)

	var waitGroup sync.WaitGroup
	actualRoutineLimit := U.MinInt(len(users), C.AllowedGoRoutinesSegmentMarker())
	waitGroup.Add(actualRoutineLimit)
	count := 0
	for _, user := range users {
		count++
		go usersProcessing(projectID, user, allSegmentsMap, decodedSegmentRulesMap, &waitGroup, &statusMap)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(users)-count, actualRoutineLimit))
		}
	}

	waitGroup.Wait()

	if len(statusMap) > 0 {
		log.WithField("project_id", projectID).Error("failures while running segment_markup for following users ", statusMap)
	}

	// check if there is no All type segment in the project
	if _, exists := allSegmentsMap["All"]; exists {
		status, err := allAccountsSegmentMarkup(projectID, users, allSegmentsMap["All"], decodedSegmentRulesMap["All"], domainGroup.ID)
		if status != http.StatusOK || err != nil {
			log.WithField("project_id", projectID).Error("Unable to update associated_segments to the domain user.")
			return status
		}
	}

	// updating segment_marker_last_run in project settings after completing the job
	_, errCode := store.GetStore().UpdateProjectSettings(projectID,
		&model.ProjectSetting{SegmentMarkerLastRun: U.TimeNowZ()})

	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectID).Error("Unable to update segment_marker_last_run in project_settings.")
		return errCode
	}

	return http.StatusOK
}

func usersProcessing(projectID int64, user model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query, waitGroup *sync.WaitGroup, statusMap *map[string]int) {
	logFields := log.Fields{
		"project_id":                projectID,
		"user":                      user,
		"all_segments_map":          allSegmentsMap,
		"decoded_segment_rules_map": decodedSegmentRulesMap,
		"wait_group":                waitGroup,
	}

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()

	status := userProcessingWithErrcode(projectID, user, allSegmentsMap, decodedSegmentRulesMap)

	if status != http.StatusOK {
		(*statusMap)[user.ID] = status
	}

}

func userProcessingWithErrcode(projectID int64, user model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query) int {
	userAssociatedSegments := make(map[string]interface{})

	// decoding user properties col
	decodedProps, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		log.WithField("project_id", projectID).Error("Unable to decode user properties.")
		return http.StatusInternalServerError
	}

	matched := false
	for groupName, segmentArray := range allSegmentsMap {
		if model.IsDomainGroup(groupName) {
			continue
		}
		if model.GroupUserSource[groupName] == *user.Source || model.UserSourceMap[groupName] == *user.Source {
			for index, segment := range segmentArray {
				segmentQuery := decodedSegmentRulesMap[groupName][index]

				// check if grpa == source
				if (segmentQuery.GroupAnalysis == model.SourceGroupUser[*user.Source] && (user.IsGroupUser != nil && *user.IsGroupUser)) ||
					(segmentQuery.GroupAnalysis == model.FILTER_TYPE_USERS && (user.IsGroupUser == nil || !*user.IsGroupUser)) {
					// apply segment rule on the user
					matched = isRuleMatched(segment, decodedProps)

					// update associated_segments map on the basis of segment rule applied
					userAssociatedSegments = updateSegmentMap(matched, user,
						userAssociatedSegments, segment.Id)
				}
			}
		}
	}

	// decoding associated_segments col
	userPartOfSegments, err := U.DecodePostgresJsonb(&user.AssociatedSegments)
	if err != nil {
		log.WithField("project_id", projectID).Error("Unable to decode associated_segments.")
		return http.StatusInternalServerError
	}

	// do not update in db if no associated segment
	if len(*userPartOfSegments) == 0 && len(userAssociatedSegments) == 0 {
		return http.StatusOK
	}

	// update associated_segments in db
	status, _ := store.GetStore().UpdateAssociatedSegments(projectID, user.ID,
		userAssociatedSegments)
	if status != http.StatusOK {
		log.WithField("project_id", projectID).Error("Unable to update associated_segments to the user.")
		return status
	}
	return http.StatusOK
}

func allAccountsSegmentMarkup(projectID int64, users []model.User, segments []model.Segment, segmentsRulesArr []model.Query,
	domainGroupId int) (int, error) {
	if len(segments) == 0 {
		return http.StatusOK, nil
	}

	domainUsersMap := make(map[string][]model.User)
	for _, user := range users {
		groupUserId, err := findUserGroupByID(user, domainGroupId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		domainUsersMap[groupUserId] = append(domainUsersMap[groupUserId], user)
	}

	for domId, usersArray := range domainUsersMap {
		associatedSegments := make(map[string]interface{})
		decodedPropsArr := make([]map[string]interface{}, 0)
		for _, user := range usersArray {
			// decoding user properties col
			decodedProps, err := U.DecodePostgresJsonb(&user.Properties)
			if err != nil {
				log.WithField("project_id", projectID).Error("Unable to decode user properties.")
				return http.StatusInternalServerError, err
			}
			decodedPropsArr = append(decodedPropsArr, *decodedProps)
		}

		for index, segmentRule := range segmentsRulesArr {
			// apply segment rule on the user
			matched := isRuleMatchedAllAccounts(segmentRule, decodedPropsArr, usersArray)

			// update associated_segments map on the basis of segment rule applied
			associatedSegments = updateAllAccountsSegmentMap(matched, usersArray,
				associatedSegments, segments[index].Id)
		}

		// update associated_segments in db
		status, err := store.GetStore().UpdateAssociatedSegments(projectID, domId,
			associatedSegments)
		if status != http.StatusOK {
			log.WithField("project_id", projectID).Error("Unable to update associated_segments to the user.")
			return status, err
		}
	}

	return http.StatusOK, nil
}

func findUserGroupByID(u model.User, id int) (string, error) {
	switch id {
	case 1:
		return u.Group1UserID, nil
	case 2:
		return u.Group2UserID, nil
	case 3:
		return u.Group3UserID, nil
	case 4:
		return u.Group4UserID, nil
	case 5:
		return u.Group5UserID, nil
	case 6:
		return u.Group6UserID, nil
	case 7:
		return u.Group7UserID, nil
	case 8:
		return u.Group8UserID, nil
	default:
		return "", fmt.Errorf("no matching group for ID %d", id)
	}
}

func isRuleMatchedAllAccounts(segment model.Query, decodedProperties []map[string]interface{}, userArr []model.User) bool {
	// isMatched = all rules matched (a or b) AND (c or d)
	isMatched := false

	if segment.GlobalUserProperties == nil || len(segment.GlobalUserProperties) == 0 {
		// currently, only support for global properties
		log.Error("No GlobalUserProperties for the segment found.")
		return isMatched
	}

	groupedProperties := model.GetPropertiesGrouped(segment.GlobalUserProperties)
	for index, currentGroupedProperties := range groupedProperties {
		// validity for each group like (a or b ) AND (c or d)
		groupedPropsMatched := false
		for _, p := range currentGroupedProperties {
			isValueFound := checkPropertyInAllUsers(segment.GroupAnalysis, p, decodedProperties, userArr)
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
	return isMatched
}

func isRuleMatched(segment model.Segment, decodedProperties *map[string]interface{}) bool {
	// isMatched = all rules matched (a or b) AND (c or d)
	isMatched := false
	segmentQuery := &model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, segmentQuery)
	if err != nil {
		log.Error("Unable to decode segment query")
		return isMatched
	}
	if segmentQuery.GlobalUserProperties == nil || len(segmentQuery.GlobalUserProperties) == 0 {
		// currently, only support for global properties
		log.Error("No GlobalUserProperties for the segment found.")
		return isMatched
	}

	// If UI presents filters in "(a or b) AND (c or D)" order, Request has it as "a or b AND c or D"
	// Using AND as a separation between lines and execution order to achieve the same as above.
	groupedProperties := model.GetPropertiesGrouped(segmentQuery.GlobalUserProperties)
	for index, currentGroupedProperties := range groupedProperties {
		// validity for each group like (a or b ) AND (c or d)
		groupedPropsMatched := false
		for idx, p := range currentGroupedProperties {
			// for groups, support is only added for group property filters
			if p.Entity != model.PropertyEntityUserGlobal {
				continue
			}
			isValueFound := checkPropertyOfGivenType(p, decodedProperties)
			if idx == 0 {
				groupedPropsMatched = isValueFound
			} else {
				groupedPropsMatched = isValueFound || groupedPropsMatched
			}
		}
		if index == 0 {
			isMatched = groupedPropsMatched
			continue
		}
		isMatched = groupedPropsMatched && isMatched
	}
	return isMatched
}

func checkPropertyInAllUsers(grpa string, p model.QueryProperty, decodedProperties []map[string]interface{}, userArr []model.User) bool {
	isValueFound := false
	for index, user := range userArr {
		// skip for group user if entity is user_group
		if p.Entity == model.PropertyEntityUserGroup && (user.IsGroupUser != nil && *user.IsGroupUser) {
			continue
		}
		isValueFound = checkPropertyOfGivenType(p, &decodedProperties[index])

		// check for negative filters
		if (p.Operator == model.NotContainsOpStr && p.Value != model.PropertyValueNone) ||
			(p.Operator == model.ContainsOpStr && p.Value == model.PropertyValueNone) ||
			(p.Operator == model.NotEqualOpStr && p.Value != model.PropertyValueNone) ||
			(p.Operator == model.EqualsOpStr && p.Value == model.PropertyValueNone) {
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

func checkPropertyOfGivenType(p model.QueryProperty, decodedProperties *map[string]interface{}) bool {
	isValueFound := false
	if p.Value != model.PropertyValueNone {
		if p.Type == U.PropertyTypeDateTime {
			isValueFound = checkDateTypeProperty(p, decodedProperties)
		} else if p.Type == U.PropertyTypeNumerical {
			isValueFound = checkNumericalTypeProperty(p, decodedProperties)
		} else {
			isValueFound = checkCategoricalTypeProperty(p, decodedProperties)
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

func updateSegmentMap(matched bool, user model.User, userPartOfSegments map[string]interface{},
	segmentId string) map[string]interface{} {
	segmentMap := userPartOfSegments
	if matched {
		segmentMap[segmentId] = user.UpdatedAt
	}
	return segmentMap
}

func updateAllAccountsSegmentMap(matched bool, userArr []model.User, userPartOfSegments map[string]interface{}, segmentId string) map[string]interface{} {
	segmentMap := userPartOfSegments
	if matched {
		maxUpadtedAt := userArr[0].UpdatedAt
		for _, user := range userArr {
			if (user.UpdatedAt).After(maxUpadtedAt) {
				maxUpadtedAt = user.UpdatedAt
			}
		}

		segmentMap[segmentId] = maxUpadtedAt
	}

	return segmentMap
}

func checkNumericalTypeProperty(segmentRule model.QueryProperty, properties *map[string]interface{}) bool {
	if _, exists := (*properties)[segmentRule.Property]; !exists {
		return false
	}
	var propertyExists bool
	checkValue, _ := strconv.ParseFloat(segmentRule.Value, 64)
	propertyValue := (*properties)[segmentRule.Property].(float64)
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

func checkCategoricalTypeProperty(segmentRule model.QueryProperty, properties *map[string]interface{}) bool {
	if (segmentRule.Operator == model.NotContainsOpStr && segmentRule.Value != model.PropertyValueNone) ||
		(segmentRule.Operator == model.ContainsOpStr && segmentRule.Value == model.PropertyValueNone) ||
		(segmentRule.Operator == model.NotEqualOpStr && segmentRule.Value != model.PropertyValueNone) ||
		(segmentRule.Operator == model.EqualsOpStr && segmentRule.Value == model.PropertyValueNone) {
		if _, exists := (*properties)[segmentRule.Property]; !exists {
			return true
		}
	}
	if _, exists := (*properties)[segmentRule.Property]; !exists {
		return false
	}
	var propertyExists bool
	checkValue := segmentRule.Value
	propertyValue := (*properties)[segmentRule.Property].(string)
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
	propertyValue, err := strconv.ParseInt((*properties)[segmentRule.Property].(string), 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed reading timestamp on user join query.")
	}
	checkValue, err := model.DecodeDateTimePropertyValue(segmentRule.Value)
	if err != nil {
		log.WithError(err).Error("Failed reading timestamp on user join query.")
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
