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

	// domain group does not exist and ProcessOnlyAllAccountsSegments set to true, so aborting
	if (status != http.StatusFound && domainGroup != nil) && C.ProcessOnlyAllAccountsSegments() {
		return http.StatusOK
	}

	var users []model.User

	var endTime, timeTaken int64
	startTime := time.Now().Unix()

	domainsList, allUsersRun, status, lookBack := GetDomainsToRunMarkerFor(projectID, domainGroup, status)

	// fetching list of all the users with last updated_at in last x hour
	if !C.ProcessOnlyAllAccountsSegments() {
		users, status = store.GetStore().GetNonGroupUsersUpdatedAtGivenHour(projectID, lookBack)
		endTime = time.Now().Unix()
		timeTaken = endTime - startTime
		log.WithFields(log.Fields{"project_id": projectID, "no_of_users": len(users)}).
			Info("Total no.of users pulled time(sec) ", timeTaken)
	}

	if (len(domainsList) <= 0 && len(users) <= 0) || status != http.StatusOK {
		if status == http.StatusNotFound {
			// because the job run should still be successful even if data is not found
			return http.StatusOK
		}
		return status
	}

	endTime = time.Now().Unix()
	timeTaken = endTime - startTime

	// Total no.of records pulled.
	// Time taken to pull.

	log.WithFields(log.Fields{"project_id": projectID, "no_of_domains": len(domainsList)}).Info("Total no.of domains pulled time(sec) ", timeTaken)

	if len(users) >= 250000 {
		log.WithFields(log.Fields{"project_id": projectID}).Warn("Total records exceeded 250k")
	}

	if len(domainsList) == 2000000 {
		log.WithFields(log.Fields{"project_id": projectID}).Error("No.of domains hitting max limit.")
	} else if len(domainsList) >= 1000000 {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Total domains exceeded 1M")
	}

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
	eventNameIDsList := make(map[string]bool)

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
			segmentQuery.From = U.TimeNowZ().AddDate(0, 0, -90).Unix()
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
	eventNameIDsMap := make(map[string]string)

	// adding ids for all the event_names
	if len(eventNameIDsList) > 0 {
		eventNameIDsMap, status = store.GetStore().GetEventNameIdsWithGivenNames(projectID, eventNameIDsList)
		if status != http.StatusFound {
			log.WithField("project_id", projectID).Error("Error fetching event_names for the project")
			return status
		}
	}

	// if !C.ProcessOnlyAllAccountsSegments() {
	// process user based segments
	// 	userProfileSegmentsProcessing(projectID, users, allSegmentsMap, decodedSegmentRulesMap, eventNameIDsMap)
	// }

	var errCode int
	// check if there is no $domains type segment in the project
	if _, exists := allSegmentsMap[model.GROUP_NAME_DOMAINS]; !exists {
		if allUsersRun {
			// updating marker_last_run_all_accounts in project settings after completing the job
			errCode = store.GetStore().UpdateSegmentMarkerLastRunForAllAccounts(projectID, U.TimeNowZ())
		} else {
			errCode = store.GetStore().UpdateSegmentMarkerLastRun(projectID, U.TimeNowZ())
		}
		if errCode != http.StatusAccepted {
			log.WithField("project_id", projectID).Error("Unable to update last_run in project_settings.")
			return errCode
		}

		return http.StatusOK
	}

	allAccountsDecodedSegmentRule, ruleExists := decodedSegmentRulesMap[model.GROUP_NAME_DOMAINS]
	if ruleExists && (domainGroup != nil && domainGroup.ID > 0) {

		// process the domains list
		status, err := processDomainsInBatches(projectID, domainsList, allSegmentsMap[model.GROUP_NAME_DOMAINS], allAccountsDecodedSegmentRule, domainGroup.ID, eventNameIDsMap)

		if status != http.StatusOK || err != nil {
			log.WithField("project_id", projectID).Error("Unable to update associated_segments to the domain user.")
			return status
		}
	}

	if allUsersRun {
		// updating marker_last_run_all_accounts in project settings after completing the job
		errCode = store.GetStore().UpdateSegmentMarkerLastRunForAllAccounts(projectID, U.TimeNowZ())
	} else {
		// updating segment_marker_last_run in project settings after completing the job
		errCode = store.GetStore().UpdateSegmentMarkerLastRun(projectID, U.TimeNowZ())
	}

	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectID).Error("Unable to update last_run in project_settings.")
		return errCode
	}

	return http.StatusOK
}

func GetDomainsToRunMarkerFor(projectID int64, domainGroup *model.Group, domainGroupStatus int) ([]string, bool, int, time.Time) {

	allUsersRun := false
	allDomainsRunHours := C.TimeRangeForAllDomains()

	// fetch lastRunTime
	lastRunTime, lastRunStatusCode := store.GetStore().GetMarkerLastForAllAccounts(projectID)
	if lastRunStatusCode == http.StatusFound && C.AllAccountsRuntMarker(projectID) {
		timeNow := time.Now().UTC()
		timeDifference := timeNow.Sub(lastRunTime)
		if timeDifference >= time.Duration(allDomainsRunHours)*time.Hour && domainGroupStatus == http.StatusFound {

			domainIDList, status := store.GetStore().GetAllDomainsByProjectID(projectID, domainGroup.ID)

			if len(domainIDList) > 0 {
				allUsersRun = true
				return domainIDList, allUsersRun, http.StatusOK, time.Time{}
			}

			if status != http.StatusFound {
				return domainIDList, allUsersRun, status, time.Time{}
			}
		}
	}

	var lookBack time.Time

	if !C.UseLookbackForSegmentMarker() {
		lookBack, _ = store.GetStore().GetSegmentMarkerLastRunTime(projectID)
	}

	if lookBack.IsZero() || C.UseLookbackForSegmentMarker() {
		hour := C.LookbackForSegmentMarker()
		lookBack = U.TimeNowZ().Add(time.Duration(-hour) * time.Hour)
	}

	// list of domains and their associated users with last updated_at in last x hour
	domainIDList, statusCode := store.GetStore().GetLatestUpatedDomainsByProjectID(projectID, domainGroup.ID, lookBack)

	if statusCode == http.StatusInternalServerError {
		log.WithFields(log.Fields{"project_id": projectID, "look_back": lookBack, "status_code": statusCode}).
			Error("Server error, couldn't find updated records")
		return []string{}, allUsersRun, statusCode, lookBack
	} else if statusCode == http.StatusNotFound || len(domainIDList) == 0 {
		log.WithFields(log.Fields{"project_id": projectID, "look_back": lookBack, "status_code": statusCode}).
			Warn("Couldn't find updated records in last given hours for this project ")
		return []string{}, allUsersRun, http.StatusOK, lookBack
	}

	return domainIDList, allUsersRun, http.StatusOK, time.Time{}
}

func processDomainsInBatches(projectID int64, domainIDList []string, segments []model.Segment, segmentsRulesArr []model.Query,
	domainGroupId int, eventNameIDsMap map[string]string) (int, error) {
	// Batch size
	batchSize := 100

	statusArr := make([]bool, 0)

	startTime := time.Now().Unix()
	userCount := new(int64)

	domainIDChunks := U.GetStringListAsBatch(domainIDList, batchSize)

	for ci := range domainIDChunks {
		var wg sync.WaitGroup

		wg.Add(len(domainIDChunks[ci]))
		for _, domID := range domainIDChunks[ci] {
			go fetchAndProcessFromDomainList(projectID, domID, segments, segmentsRulesArr, domainGroupId,
				eventNameIDsMap, &wg, &statusArr, userCount)
		}
		wg.Wait()
	}

	endTime := time.Now().Unix()
	timeTaken := endTime - startTime

	// Time taken to process domains.
	log.WithFields(log.Fields{"project_id": projectID, "no_of_domains": len(domainIDList), "user_count": userCount}).Info("Processing Time for domains in sec ", timeTaken)

	if len(statusArr) > 0 {
		log.WithField("project_id", projectID).Error("failures while running segment_markup for following number of batches of domain ", len(statusArr))
	}

	return http.StatusOK, nil
}

func fetchAndProcessFromDomainList(projectID int64, domainID string, segments []model.Segment, segmentsRulesArr []model.Query,
	domainGroupId int, eventNameIDsMap map[string]string, waitGroup *sync.WaitGroup, statusArr *[]bool, userCount *int64) {

	logFields := log.Fields{
		"project_id":      projectID,
		"domain_group_id": domainGroupId,
		"domain_id":       domainID,
		"segments":        segments,
		"wait_group":      waitGroup,
	}

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()

	users, status := store.GetStore().GetUsersAssociatedToDomainList(projectID, domainGroupId, domainID)

	if len(users) == 0 || status != http.StatusFound {
		log.WithField("project_id", projectID).Error("Unable to find users for domain ", domainID)
		*statusArr = append(*statusArr, false)
		return
	}

	*userCount = *userCount + int64(len(users))

	status, err := domainUsersProcessingWithErrcode(projectID, domainID, users, segments,
		segmentsRulesArr, eventNameIDsMap)

	if status != http.StatusOK || err != nil {
		log.WithField("project_id", projectID).Error("Unable to update associated_segments to the domain user.")
	}

	if status != http.StatusOK || err != nil {
		*statusArr = append(*statusArr, false)
	}
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
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]string) {
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
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]string, waitGroup *sync.WaitGroup, statusMap *map[string]int) {
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
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]string) int {
	userAssociatedSegments := make(map[string]model.AssociatedSegments)

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
		if user.Source != nil && (model.GroupUserSource[groupName] == *user.Source || model.UserSourceMap[groupName] == *user.Source) {
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
	domainGroupId int, eventNameIDsMap map[string]string) (int, error) {
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
	eventNameIDsMap map[string]string, waitGroup *sync.WaitGroup, statusMap *map[string]int) {
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

func domainUsersProcessingWithErrcode(projectID int64, domId string, usersArray []model.User, segments []model.Segment, segmentsRulesArr []model.Query, eventNameIDsMap map[string]string) (int, error) {
	associatedSegments := make(map[string]model.AssociatedSegments)
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
	segmentId string, domId string, eventNameIDsMap map[string]string) bool {
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

func isRuleMatched(projectID int64, segment model.Segment, decodedProperties *map[string]interface{}, eventNameIDsMap map[string]string, userID string) bool {
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

func updateSegmentMap(matched bool, user model.User, userPartOfSegments map[string]model.AssociatedSegments,
	segmentId string) map[string]model.AssociatedSegments {
	segmentMap := userPartOfSegments
	var updatedMap model.AssociatedSegments

	if matched {
		updatedMap.UpdatedAt = model.FormatTimeToString(user.UpdatedAt)
		if user.LastEventAt != nil {
			updatedMap.LastEventAt = model.FormatTimeToString(*user.LastEventAt)
		}
		updatedMap.V = 0
		segmentMap[segmentId] = updatedMap
	}
	return segmentMap
}

func updateAllAccountsSegmentMap(matched bool, userArr []model.User, userPartOfSegments map[string]model.AssociatedSegments, segmentId string) map[string]model.AssociatedSegments {
	segmentMap := userPartOfSegments
	var updatedMap model.AssociatedSegments
	if matched {
		maxUpadtedAt := userArr[0].UpdatedAt
		maxLastEventAt := time.Time{}
		for _, user := range userArr {
			if (user.UpdatedAt).After(maxUpadtedAt) {
				maxUpadtedAt = user.UpdatedAt
			}
			if user.LastEventAt != nil {
				if (*user.LastEventAt).After(maxLastEventAt) {
					maxLastEventAt = *user.LastEventAt
				}
			}
		}
		updatedMap.UpdatedAt = model.FormatTimeToString(maxUpadtedAt)
		if (maxLastEventAt.Compare(time.Time{})) != 0 {
			updatedMap.LastEventAt = model.FormatTimeToString(maxLastEventAt)
		}
		updatedMap.V = 0
		segmentMap[segmentId] = updatedMap
	}

	return segmentMap
}

func performedEventsCheck(projectID int64, segmentID string, eventNameIDsMap map[string]string,
	segmentQuery *model.Query, userID string, userArr []model.User) bool {

	isPerformedEvent, isAllAccounts := false, false

	if segmentQuery.Caller != model.USER_PROFILES {
		isAllAccounts = true
	}
	isPerformedEvent = memsql.EventsPerformedCheck(projectID, segmentID, eventNameIDsMap, segmentQuery, userID, isAllAccounts, userArr)

	return isPerformedEvent
}
