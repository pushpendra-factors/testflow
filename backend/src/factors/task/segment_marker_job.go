package task

import (
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

func SegmentMarker(projectID int64, projectIdListAllRun map[int64]bool, latestSegmentsMap map[int64][]string) int {

	if C.DisableAccountsRuntMarker(projectID) {
		return http.StatusOK
	}

	domainGroup, status := store.GetStore().GetGroup(projectID, model.GROUP_NAME_DOMAINS)

	// domain group does not exist and ProcessOnlyAllAccountsSegments set to true, so aborting
	if (domainGroup == nil || status != http.StatusFound) && C.ProcessOnlyAllAccountsSegments() {
		return http.StatusOK
	}

	var users []model.User

	var endTime, timeTaken int64
	startTime := time.Now().Unix()

	domainsList, allUsersRun, status, lookBack := GetDomainsToRunMarkerFor(projectID, domainGroup, status, projectIdListAllRun, latestSegmentsMap)

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

	// set Query Timezone
	timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("Failed to get Timezone.")
		return http.StatusBadRequest
	}

	// list of segments
	allSegmentsMap, statusCode := ListOfSegmentsToRunMarkerFor(projectID, allUsersRun, latestSegmentsMap)
	if statusCode != http.StatusFound {
		log.WithField("project_id", projectID).Warn("No segment found for this project")
		return http.StatusOK
	}

	// list of all event_names and all ids for it
	eventNameIDsList := make(map[string]bool)

	// map of all the inList and notInList files values
	fileValuesMap := make(map[string]map[string]bool)

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
			if segment.Type == model.GROUP_NAME_DOMAINS {
				segmentQuery.Caller = model.PROFILE_TYPE_ACCOUNT
				segmentQuery.GroupAnalysis = model.GROUP_NAME_DOMAINS
			} else {
				segmentQuery.Caller = model.PROFILE_TYPE_USER
				segmentQuery.GroupAnalysis = model.FILTER_TYPE_USERS
			}
			segmentQuery.GlobalUserProperties = model.TransformPayloadForInProperties(segmentQuery.GlobalUserProperties)
			segmentQuery.From = U.TimeNowZ().AddDate(0, 0, -90).Unix()
			segmentQuery.To = U.TimeNowZ().Unix()
			for _, eventProp := range segmentQuery.EventsWithProperties {
				if eventProp.Name != "" {
					eventNameIDsList[eventProp.Name] = true
				}
			}
			decodedSegmentRulesMap[groupName] = append(decodedSegmentRulesMap[groupName], segmentQuery)

			// map for inList and NotInList
			fileValuesMap = memsql.GetFileValues(projectID, segmentQuery.GlobalUserProperties, fileValuesMap)
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
		status, err := processDomainsInBatches(projectID, domainsList, allSegmentsMap[model.GROUP_NAME_DOMAINS],
			allAccountsDecodedSegmentRule, domainGroup.ID, eventNameIDsMap, fileValuesMap)

		if status != http.StatusOK || err != nil {
			log.WithField("project_id", projectID).Error("Unable to update associated_segments to the domain user.")
			return status
		}
	}

	var runForGivenSegment bool
	var segmentIds []string

	// to run marker for 100k domains for selected recently updated/created segments
	if !allUsersRun && C.EnableLatestSegmentsMarkerRun(projectID) {
		segmentIds, runForGivenSegment = latestSegmentsMap[projectID]
	}

	if allUsersRun {
		// updating marker_last_run_all_accounts in project settings after completing the job
		errCode = store.GetStore().UpdateSegmentMarkerLastRunForAllAccounts(projectID, U.TimeNowZ())
	} else if runForGivenSegment && len(segmentIds) > 0 {
		// updating marker_run_segment in segments table after completing the job
		errCode = store.GetStore().UpdateMarkerRunSegment(projectID, segmentIds, U.TimeNowZ())
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

func GetDomainsToRunMarkerFor(projectID int64, domainGroup *model.Group, domainGroupStatus int,
	projectIdListAllRun map[int64]bool, latestSegmentsMap map[int64][]string) ([]string, bool, int, time.Time) {

	allUsersRun := false

	limitVal := C.MarkerDomainLimitForAllRun()

	_, isAllRun := projectIdListAllRun[projectID]

	var runForGivenSegment bool

	// to run marker for 100k domains for selected recently updated/created segments
	if !isAllRun && C.EnableLatestSegmentsMarkerRun(projectID) {
		_, runForGivenSegment = latestSegmentsMap[projectID]
	}

	if C.AllAccountsRuntMarker(projectID) {

		if domainGroupStatus == http.StatusFound && (isAllRun || runForGivenSegment) {

			domainIDList, status := store.GetStore().GetAllDomainsByProjectID(projectID, domainGroup.ID, limitVal)

			if len(domainIDList) > 0 {
				allUsersRun = isAllRun
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
	domainIDList, statusCode := store.GetStore().GetLatestUpatedDomainsByProjectID(projectID, domainGroup.ID, lookBack, limitVal)

	if statusCode == http.StatusInternalServerError {
		log.WithFields(log.Fields{"project_id": projectID, "look_back": lookBack, "status_code": statusCode}).
			Error("Server error, couldn't find updated records")
		return []string{}, allUsersRun, statusCode, lookBack
	} else if statusCode == http.StatusNotFound || len(domainIDList) == 0 {
		return []string{}, allUsersRun, http.StatusOK, lookBack
	}

	return domainIDList, allUsersRun, http.StatusOK, time.Time{}
}

func ListOfSegmentsToRunMarkerFor(projectID int64, allUsersRun bool, latestSegmentsMap map[int64][]string) (map[string][]model.Segment, int) {

	var runForGivenSegment bool

	var segmentIds []string

	// to run marker for 100k domains for selected recently updated/created segments
	if !allUsersRun && C.EnableLatestSegmentsMarkerRun(projectID) {
		segmentIds, runForGivenSegment = latestSegmentsMap[projectID]
	}

	if runForGivenSegment && len(segmentIds) > 0 {
		// list of given segments
		return store.GetStore().GetSegmentByGivenIds(projectID, segmentIds)
	}

	// list of all segments
	return store.GetStore().GetAllSegments(projectID)
}

func processDomainsInBatches(projectID int64, domainIDList []string, segments []model.Segment, segmentsRulesArr []model.Query,
	domainGroupId int, eventNameIDsMap map[string]string, fileValuesMap map[string]map[string]bool) (int, error) {
	// Batch size
	batchSize := C.BatchSizeSegmentMarker()

	statusArr := make([]bool, 0)

	startTime := time.Now().Unix()
	userCount := new(int64)
	domainUpdateCount := new(int64)

	domainIDChunks := U.GetStringListAsBatch(domainIDList, batchSize)

	for ci := range domainIDChunks {
		var wg sync.WaitGroup

		wg.Add(len(domainIDChunks[ci]))
		for _, domID := range domainIDChunks[ci] {
			go fetchAndProcessFromDomainList(projectID, domID, segments, segmentsRulesArr, domainGroupId,
				eventNameIDsMap, &wg, &statusArr, userCount, domainUpdateCount, fileValuesMap)
		}
		wg.Wait()
	}

	endTime := time.Now().Unix()
	timeTaken := endTime - startTime

	// Time taken to process domains.
	log.WithFields(log.Fields{"project_id": projectID, "no_of_domains": len(domainIDList), "user_count": userCount,
		"no_of_updates": *domainUpdateCount}).Info("Processing Time for domains in sec ", timeTaken)

	if len(statusArr) > 0 {
		log.WithField("project_id", projectID).Error("failures while running segment_markup for following number of batches of domain ", len(statusArr))
	}

	return http.StatusOK, nil
}

func fetchAndProcessFromDomainList(projectID int64, domainID string, segments []model.Segment, segmentsRulesArr []model.Query,
	domainGroupId int, eventNameIDsMap map[string]string, waitGroup *sync.WaitGroup, statusArr *[]bool, userCount *int64,
	domainUpdateCount *int64, fileValuesMap map[string]map[string]bool) {

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

	users, status := store.GetStore().GetAllPropertiesForDomain(projectID, domainGroupId, domainID, userCount)
	if len(users) == 0 || status != http.StatusFound {
		*statusArr = append(*statusArr, false)
		return
	}

	status, err := domainUsersProcessingWithErrcode(projectID, domainID, users, segments,
		segmentsRulesArr, eventNameIDsMap, domainUpdateCount, fileValuesMap)

	if status != http.StatusOK || err != nil {
		log.WithField("project_id", projectID).Error("Unable to update associated_segments to the domain user.")
		*statusArr = append(*statusArr, false)
	}
}

func userProfileSegmentsProcessing(projectID int64, users []model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]string, fileValuesMap map[string]map[string]bool) {
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
		go usersProcessing(projectID, user, allSegmentsMap, decodedSegmentRulesMap, eventNameIDsMap, &waitGroup,
			&statusMap, fileValuesMap)
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
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]string, waitGroup *sync.WaitGroup,
	statusMap *map[string]int, fileValuesMap map[string]map[string]bool) {
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

	status := userProcessingWithErrcode(projectID, user, allSegmentsMap, decodedSegmentRulesMap, eventNameIDsMap, fileValuesMap)

	if status != http.StatusOK {
		(*statusMap)[user.ID] = status
	}

}

func userProcessingWithErrcode(projectID int64, user model.User, allSegmentsMap map[string][]model.Segment,
	decodedSegmentRulesMap map[string][]model.Query, eventNameIDsMap map[string]string, fileValuesMap map[string]map[string]bool) int {
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
					matched = isRuleMatched(projectID, segment, decodedProps, eventNameIDsMap, user.ID, fileValuesMap)

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

func domainUsersProcessingWithErrcode(projectID int64, domId string, usersArray []model.User,
	segments []model.Segment, segmentsRulesArr []model.Query, eventNameIDsMap map[string]string,
	domainUpdateCount *int64, fileValuesMap map[string]map[string]bool) (int, error) {
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
		matched := memsql.IsRuleMatchedAllAccounts(projectID, segmentRule, decodedPropsArr, usersArray, segments[index].Id,
			domId, eventNameIDsMap, fileValuesMap)

		// update associated_segments map on the basis of segment rule applied
		associatedSegments = updateAllAccountsSegmentMap(matched, usersArray,
			associatedSegments, segments[index].Id)
	}

	// check whether associated_segments need to be updated
	existingAssociatedSegment, status := store.GetStore().GetAssociatedSegmentForUser(projectID, domId)

	if status != http.StatusFound {
		if len(associatedSegments) == 0 {
			return http.StatusOK, nil
		}
	}

	updateAssociatedSegment := IsMapsNotMatching(projectID, domId, associatedSegments, existingAssociatedSegment)

	if !updateAssociatedSegment {
		return http.StatusOK, nil
	}

	// update associated_segments in db
	status, err := store.GetStore().UpdateAssociatedSegments(projectID, domId,
		associatedSegments)
	*domainUpdateCount++
	if status != http.StatusOK {
		log.WithFields(log.Fields{"project_id": projectID, "domain_id": domId}).Error("Unable to update associated_segments to the user")
		return status, err
	}

	return http.StatusOK, nil
}

func IsMapsNotMatching(projectID int64, domId string, associatedSegments map[string]model.AssociatedSegments, oldAssociatedSegments map[string]interface{}) bool {
	isDifferent := false

	// length check
	if len(associatedSegments) != len(oldAssociatedSegments) {
		isDifferent = true
	}

	// checking new map, if previously computed segment does not exist
	for segID := range associatedSegments {
		if _, exists := oldAssociatedSegments[segID]; !exists {
			log.WithFields(log.Fields{"project_id": projectID, "domain_id": domId, "segment_id": segID}).
				Info("Entering the segment")
			isDifferent = true

			// alert for entering segment
			store.GetStore().FindAndCacheAlertForCurrentSegment(projectID, segID, domId, model.ACTION_SEGMENT_ENTRY)
		}
	}

	// checking old map, if newly computed segment does not exist
	for segID := range oldAssociatedSegments {
		if _, exists := associatedSegments[segID]; !exists {
			log.WithFields(log.Fields{"project_id": projectID, "domain_id": domId, "segment_id": segID}).
				Info("Leaving the segment")

			isDifferent = true

			// alert for exiting segment
			store.GetStore().FindAndCacheAlertForCurrentSegment(projectID, segID, domId, model.ACTION_SEGMENT_EXIT)
		}
	}

	return isDifferent
}

func isRuleMatched(projectID int64, segment model.Segment, decodedProperties *map[string]interface{},
	eventNameIDsMap map[string]string, userID string, fileValuesMap map[string]map[string]bool) bool {
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
		isMatched = memsql.PerformedEventsCheck(projectID, segment.Id, eventNameIDsMap, segmentQuery, userID, userArr)
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
			isValueFound := memsql.CheckPropertyOfGivenType(projectID, p, decodedProperties, fileValuesMap)
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
		isMatched = memsql.PerformedEventsCheck(projectID, segment.Id, eventNameIDsMap, segmentQuery, userID, userArr)
	}

	return isMatched
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
		associatedSegUpdatedAt := time.Now().UTC()
		maxLastEventAt := time.Time{}
		for _, user := range userArr {
			if user.LastEventAt != nil && (*user.LastEventAt).After(maxLastEventAt) {
				maxLastEventAt = *user.LastEventAt
			}
		}
		updatedMap.UpdatedAt = model.FormatTimeToString(associatedSegUpdatedAt)
		if (maxLastEventAt.Compare(time.Time{})) != 0 {
			updatedMap.LastEventAt = model.FormatTimeToString(maxLastEventAt)
		}
		updatedMap.V = 0
		segmentMap[segmentId] = updatedMap
	}

	return segmentMap
}
