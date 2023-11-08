package task

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	C "factors/config"

	log "github.com/sirupsen/logrus"

	"factors/model/model"
	"factors/model/store/memsql"

	"factors/model/store"

	U "factors/util"
)

func SegmentMarker(projectID int64) int {

	domainGroup, status := store.GetStore().GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if status != http.StatusFound {
		log.WithField("project_id", projectID).Info("Domain group not enabled")
	}

	var lookBack time.Time

	if !C.UseLookbackForSegmentMarker() {
		lookBack, _ = store.GetStore().GetSegmentMarkerLastRunTime(projectID)
	}

	if lookBack.IsZero() || C.UseLookbackForSegmentMarker() {
		hour := C.LookbackForSegmentMarker()
		lookBack = U.TimeNowZ().Add(time.Duration(-hour) * time.Hour)
	}

	var users []model.User
	var statusCode int
	startTime := time.Now().Unix()

	if status != http.StatusFound {
		// domains not enabled, so fetching list of all the users with last updated_at in last x hour
		users, statusCode = store.GetStore().GetNonGroupUsersUpdatedAtGivenHour(projectID, lookBack)
	} else {
		// list of domains and their associated users with last updated_at in last x hour
		users, statusCode = store.GetStore().GetUsersUpdatedAtGivenHour(projectID, lookBack, domainGroup.ID)
	}
	if statusCode != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectID, "look_back": lookBack}).
			Warn("Couldn't find updated records in last given hours with given statuscode for this project", statusCode)
		return http.StatusOK
	}
	endTime := time.Now().Unix()
	timeTaken := endTime - startTime

	// Total no.of records pulled.
	// Time taken to pull.
	log.WithFields(log.Fields{"project_id": projectID, "no_of_records": len(users)}).Info("Total no.of records pulled time(sec) ", timeTaken)

	// list of all segments
	allSegmentsMap, statusCode := store.GetStore().GetAllSegments(projectID)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Warn("No segment found for this project")
		return http.StatusOK
	}

	// set Query Timezone
	timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("Failed to get Timezone.")
		return http.StatusBadRequest
	}

	// list of all event_names and all ids for it
	eventNameIDsList := make(map[string]interface{})

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
			segmentQuery.GlobalUserProperties = transformPayloadForInProperties(segmentQuery.GlobalUserProperties)
			segmentQuery.From = U.TimeNowZ().AddDate(0, 0, -28).Unix()
			segmentQuery.To = U.TimeNowZ().Unix()
			for _, eventProp := range segmentQuery.EventsWithProperties {
				if eventProp.Name != "" {
					eventNameIDsList[eventProp.Name] = true
				}
			}
			decodedSegmentRulesMap[groupName] = append(decodedSegmentRulesMap[groupName], segmentQuery)
		}
	}

	// map for all event_names and all ids for it
	eventNameIDsMap := make(map[string]interface{})

	// adding ids for all the event_names
	if len(eventNameIDsList) == 0 {
		log.WithField("project_id", projectID).Info("No segments with performed events for this project")
	} else {
		eventNameIDsMap, status = store.GetStore().GetEventNameIdsWithGivenNames(projectID, eventNameIDsList)
		if status != http.StatusFound {
			log.WithField("project_id", projectID).Error("Error fetching event_names for the project")
			return status
		}
	}

	if !C.ProcessOnlyAllAccountsSegments() {
		// process user based segments
		userProfileSegmentsProcessing(projectID, users, allSegmentsMap, decodedSegmentRulesMap, eventNameIDsMap)
	}

	// check if there is no All type segment in the project
	var domainsGroupName string
	if _, exists := allSegmentsMap["All"]; exists {
		domainsGroupName = "All"
	} else if _, exists := allSegmentsMap[model.GROUP_NAME_DOMAINS]; exists {
		domainsGroupName = model.GROUP_NAME_DOMAINS
	}

	if domainsGroupName != "" {
		allAccountsDecodedSegmentRule, exists := decodedSegmentRulesMap[domainsGroupName]
		if exists {
			status, err := allAccountsSegmentMarkup(projectID, users, allSegmentsMap[domainsGroupName], allAccountsDecodedSegmentRule, domainGroup.ID, eventNameIDsMap)
			if status != http.StatusOK || err != nil {
				log.WithField("project_id", projectID).Error("Unable to update associated_segments to the domain user.")
				return status
			}
		}
	}

	// updating segment_marker_last_run in project settings after completing the job
	errCode := store.GetStore().UpdateSegmentMarkerLastRun(projectID, U.TimeNowZ())

	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectID).Error("Unable to update segment_marker_last_run in project_settings.")
		return errCode
	}

	return http.StatusOK
}

func transformPayloadForInProperties(globalUserProperties []model.QueryProperty) []model.QueryProperty {
	for i, p := range globalUserProperties {
		if v, exist := model.IN_PROPERTIES_DEFAULT_QUERY_MAP[p.Property]; exist {
			v.LogicalOp = p.LogicalOp
			if p.Value == "true" {
				globalUserProperties[i] = v
			} else if p.Value == "false" || p.Value == "$none" {
				v.Operator = model.EqualsOpStr
				v.Value = "$none"
				globalUserProperties[i] = v
			}
		}
	}
	return globalUserProperties
}

func userProfileSegmentsProcessing(projectID int64, users []model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]interface{}) {
	statusMap := make(map[string]int, 0)

	var waitGroup sync.WaitGroup
	actualRoutineLimit := U.MinInt(len(users), C.AllowedGoRoutinesSegmentMarker())
	waitGroup.Add(actualRoutineLimit)
	count := 0

	usersProcessingStartTime := time.Now().Unix()
	var totalUsers, totalAccounts int64
	for _, user := range users {

		if user.IsGroupUser == nil || !*user.IsGroupUser {
			totalUsers++
		} else if *user.IsGroupUser {
			totalAccounts++
		}

		count++
		go usersProcessing(projectID, user, allSegmentsMap, decodedSegmentRulesMap, eventNameIDsMap, &waitGroup, &statusMap)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(users)-count, actualRoutineLimit))
		}
	}

	waitGroup.Wait()

	usersProcessingEndTime := time.Now().Unix()

	// Total time taken to process.
	// Total user records.
	// Total account records.
	log.WithFields(log.Fields{"project_id": projectID, "total_accounts_processed": totalAccounts,
		"total_users_processed": totalUsers, "total_time_for_processing_sec": (usersProcessingEndTime - usersProcessingStartTime)}).Info("Analysing informartion for users processing")
	if len(statusMap) > 0 {
		log.WithField("project_id", projectID).Error("failures while running segment_markup for following users ", statusMap)
	}
}

func usersProcessing(projectID int64, user model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]interface{}, waitGroup *sync.WaitGroup, statusMap *map[string]int) {
	logFields := log.Fields{
		"project_id":                projectID,
		"user":                      user,
		"all_segments_map":          allSegmentsMap,
		"decoded_segment_rules_map": decodedSegmentRulesMap,
		"events_name_id_map":        eventNameIDsMap,
		"wait_group":                waitGroup,
	}

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()

	status := userProcessingWithErrcode(projectID, user, allSegmentsMap, decodedSegmentRulesMap, eventNameIDsMap)

	if status != http.StatusOK {
		(*statusMap)[user.ID] = status
	}

}

func userProcessingWithErrcode(projectID int64, user model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]interface{}) int {
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
					matched = isRuleMatched(projectID, segment, decodedProps, eventNameIDsMap, user.ID)

					// update associated_segments map on the basis of segment rule applied
					userAssociatedSegments = updateSegmentMap(matched, user,
						userAssociatedSegments, segment.Id)
				}
			}
		}
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
	domainGroupId int, eventNameIDsMap map[string]interface{}) (int, error) {
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

	statusMap := make(map[string]int, 0)

	startTime := time.Now().Unix()
	var waitGroup sync.WaitGroup
	actualRoutineLimit := U.MinInt(len(domainUsersMap), C.AllowedGoRoutinesSegmentMarker())
	waitGroup.Add(actualRoutineLimit)
	count := 0
	for domId, usersArray := range domainUsersMap {
		count++
		go domainusersProcessing(projectID, domId, usersArray, segments, segmentsRulesArr, eventNameIDsMap, &waitGroup, &statusMap)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(domainUsersMap)-count, actualRoutineLimit))
		}
	}

	waitGroup.Wait()

	endTime := time.Now().Unix()
	timeTaken := endTime - startTime

	// Time taken to process domains.
	log.WithFields(log.Fields{"project_id": projectID, "no_of_domains": len(domainUsersMap)}).Info("Processing Time for domains in sec ", timeTaken)

	if len(statusMap) > 0 {
		log.WithField("project_id", projectID).Error("failures while running segment_markup for following domains ", statusMap)
	}

	return http.StatusOK, nil
}

func domainusersProcessing(projectID int64, domId string, users []model.User, segments []model.Segment, segmentsRulesArr []model.Query,
	eventNameIDsMap map[string]interface{}, waitGroup *sync.WaitGroup, statusMap *map[string]int) {
	logFields := log.Fields{
		"project_id": projectID,
		"domain_id":  domId,
		"users":      users,
		"segments":   segments,
		"wait_group": waitGroup,
	}

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()

	status, err := domainUsersProcessingWithErrcode(projectID, domId, users, segments, segmentsRulesArr, eventNameIDsMap)

	if status != http.StatusOK || err != nil {
		(*statusMap)[domId] = status
	}

}

func domainUsersProcessingWithErrcode(projectID int64, domId string, usersArray []model.User, segments []model.Segment, segmentsRulesArr []model.Query, eventNameIDsMap map[string]interface{}) (int, error) {
	associatedSegments := make(map[string]interface{})
	decodedPropsArr := make([]map[string]interface{}, 0)
	for _, user := range usersArray {
		// decoding user properties col
		decodedProps, err := U.DecodePostgresJsonb(&user.Properties)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": user.ID}).Error("Unable to decode user properties for user.")
			return http.StatusInternalServerError, err
		}
		decodedPropsArr = append(decodedPropsArr, *decodedProps)
	}

	for index, segmentRule := range segmentsRulesArr {
		// apply segment rule on the user
		matched := isRuleMatchedAllAccounts(projectID, segmentRule, decodedPropsArr, usersArray, segments[index].Id, domId, eventNameIDsMap)

		// update associated_segments map on the basis of segment rule applied
		associatedSegments = updateAllAccountsSegmentMap(matched, usersArray,
			associatedSegments, segments[index].Id)
	}

	// update associated_segments in db
	status, err := store.GetStore().UpdateAssociatedSegments(projectID, domId,
		associatedSegments)
	if status != http.StatusOK {
		log.WithFields(log.Fields{"project_id": projectID, "domain_id": domId}).Error("Unable to update associated_segments to the user")
		return status, err
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

func isRuleMatchedAllAccounts(projectID int64, segment model.Query, decodedProperties []map[string]interface{}, userArr []model.User,
	segmentId string, domId string, eventNameIDsMap map[string]interface{}) bool {
	// isMatched = all rules matched (a or b) AND (c or d)
	isMatched := false

	if (segment.GlobalUserProperties == nil || len(segment.GlobalUserProperties) == 0) &&
		(segment.EventsWithProperties != nil && len(segment.EventsWithProperties) > 0) {
		isMatched = performedEventsCheck(projectID, segmentId, eventNameIDsMap, &segment, domId, userArr)
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

	if isMatched && (segment.EventsWithProperties != nil && len(segment.EventsWithProperties) > 0) {
		isMatched = performedEventsCheck(projectID, segmentId, eventNameIDsMap, &segment, domId, userArr)
	}
	return isMatched
}

func isRuleMatched(projectID int64, segment model.Segment, decodedProperties *map[string]interface{}, eventNameIDsMap map[string]interface{}, userID string) bool {
	// isMatched = all rules matched (a or b) AND (c or d)
	isMatched := false
	segmentQuery := &model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, segmentQuery)
	if err != nil {
		log.WithField("segment_id", segment.Id).Error("Unable to decode segment query")
		return isMatched
	}
	if (segmentQuery.GlobalUserProperties == nil || len(segmentQuery.GlobalUserProperties) == 0) &&
		(segmentQuery.EventsWithProperties != nil && len(segmentQuery.EventsWithProperties) > 0) {
		var userArr []model.User
		isMatched = performedEventsCheck(projectID, segment.Id, eventNameIDsMap, segmentQuery, userID, userArr)
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
			isValueFound := memsql.CheckPropertyOfGivenType(p, decodedProperties)
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

	if isMatched && (segmentQuery.EventsWithProperties != nil && len(segmentQuery.EventsWithProperties) > 0) {
		var userArr []model.User
		isMatched = performedEventsCheck(projectID, segment.Id, eventNameIDsMap, segmentQuery, userID, userArr)
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
		isValueFound = memsql.CheckPropertyOfGivenType(p, &decodedProperties[index])

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

func updateSegmentMap(matched bool, user model.User, userPartOfSegments map[string]interface{},
	segmentId string) map[string]interface{} {
	segmentMap := userPartOfSegments
	if matched {
		updatedAt := model.FormatTimeToString(user.UpdatedAt)
		segmentMap[segmentId] = updatedAt
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
		updatedAt := model.FormatTimeToString(maxUpadtedAt)
		segmentMap[segmentId] = updatedAt
	}

	return segmentMap
}

func performedEventsCheck(projectID int64, segmentID string, eventNameIDsMap map[string]interface{},
	segmentQuery *model.Query, userID string, userArr []model.User) bool {

	isPerformedEvent, isAllAccounts := false, false

	if segmentQuery.Caller != model.USER_PROFILES {
		isAllAccounts = true
	}
	isPerformedEvent = memsql.EventsPerformedCheck(projectID, segmentID, eventNameIDsMap, segmentQuery, userID, isAllAccounts, userArr)

	return isPerformedEvent
}
