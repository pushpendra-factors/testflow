package memsql

import (
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Duplicate code below exists. Please change in GetSQLAndParamsForSegmentsAssociatedAccounts if required.
func (store *MemSQL) GetMarkedDomainsListByProjectId(projectID int64, payload model.TimelinePayload, downloadLimitGiven bool) (model.AccountsProfileQueryResponsePayload, int, string) {
	logFields := log.Fields{
		"project_id": projectID,
		"payload":    payload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	isPreview := false

	if projectID == 0 {
		return model.AccountsProfileQueryResponsePayload{}, http.StatusBadRequest, "Project Id is Invalid"
	}

	// return if source is not $domains
	if payload.Query.Source != model.GROUP_NAME_DOMAINS {
		logCtx.Error("Query failed. invalid source")
		return model.AccountsProfileQueryResponsePayload{}, http.StatusBadRequest, "Failed to fetch source."
	}

	// set Query Timezone
	timezoneString, statusCode := store.GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		logCtx.Error("Query failed. Failed to get Timezone.")
		return model.AccountsProfileQueryResponsePayload{}, statusCode, "Failed to fetch project timezone."
	}
	payload.Query.Timezone = string(timezoneString)

	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return model.AccountsProfileQueryResponsePayload{}, http.StatusInternalServerError, "failed to get group while adding group info"
	}

	var profiles []model.Profile
	var errMsg string

	payload.Query.Caller = model.PROFILE_TYPE_ACCOUNT

	if payload.SegmentId == "" {
		// new preview flow
		profiles, isPreview, statusCode, errMsg = store.GetPreviewDomainsListByProjectId(projectID, payload,
			domainGroup.ID, downloadLimitGiven)
	} else {
		// fetching accounts for segments
		profiles, isPreview, statusCode, errMsg = store.GetDomainsListFromMarker(projectID, payload,
			domainGroup.ID, downloadLimitGiven)
	}

	// if status is not found or profiles are empty, redirecting to old flow
	if statusCode == http.StatusInternalServerError {
		return model.AccountsProfileQueryResponsePayload{}, statusCode, errMsg
	}

	// datetime conversion
	err := payload.Query.TransformDateTypeFilters()
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to transform segment filters.")
		return model.AccountsProfileQueryResponsePayload{}, statusCode, "Failed to get segment"
	}
	groupedFilters := GroupFiltersByGroupName(payload.Query.GlobalUserProperties)

	// set TableProps
	if len(payload.Query.TableProps) == 0 {
		payload.Query.TableProps = store.GetTablePropsFromConfig(projectID, model.PROFILE_TYPE_ACCOUNT)
	}

	// Get merged properties for all accounts
	profiles, statusCode = store.AccountPropertiesForDomainsEnabled(projectID, profiles, groupedFilters, payload.Query.TableProps)
	if statusCode != http.StatusOK {
		return model.AccountsProfileQueryResponsePayload{}, statusCode, "Query Transformation Failed."
	}

	// Get Return Table Content
	returnData, err := FormatProfilesStruct(projectID, profiles, model.PROFILE_TYPE_ACCOUNT, payload.Query.TableProps, payload.Query.Source)
	if err != nil {
		logCtx.WithError(err).WithField("status", err).Error("Failed to filter properties from profiles.")
		return model.AccountsProfileQueryResponsePayload{}, http.StatusInternalServerError, "Query formatting failed."
	}

	response := model.AccountsProfileQueryResponsePayload{
		Profiles:  returnData,
		IsPreview: isPreview,
	}

	if !isPreview && payload.SegmentId != "" {

		accountCount, statusCode := store.GetAccountAssociatedToSegmentCount(projectID, domainGroup.ID, payload.SegmentId)
		if statusCode == http.StatusFound {
			response.Count = accountCount
		}

	}

	return response, http.StatusFound, ""
}

// Duplicated code from above. Please look for a change in GetMarkedDomainsListByProjectId.
func (store *MemSQL) GetSQLAndParamsForSegmentsAssociatedAccounts(projectID int64, segmentID string) (string, []interface{}, *model.Group, string) {
	rStmnt := ""
	rParams := make([]interface{}, 0)
	domainGroup, errCode := store.GetGroup(projectID, model.GROUP_NAME_DOMAINS)
	if errCode != http.StatusFound || domainGroup == nil {
		return rStmnt, rParams, domainGroup, "failed to get group while adding group info"
	}

	sql := " SELECT id as identity, group_%d_id as host_name FROM  users " +
		"WHERE project_id = ? AND JSON_EXTRACT_STRING(associated_segments, ?) IS NOT NULL " +
		"AND is_group_user = 1 AND source = ? AND group_%d_id IS NOT NULL  "
	rStmnt = fmt.Sprintf(sql, domainGroup.ID, domainGroup.ID)
	rParams = append(rParams, projectID, segmentID, model.UserSourceDomains)
	return rStmnt, rParams, domainGroup, ""
}

func CompareFilters(segmentQuery model.Query, payloadQuery model.Query) bool {
	additionalFiltersExist := false

	if len(segmentQuery.GlobalUserProperties) != len(payloadQuery.GlobalUserProperties) ||
		len(segmentQuery.EventsWithProperties) != len(payloadQuery.EventsWithProperties) {
		return true
	}

	// comparing event conditions
	if (segmentQuery.EventsCondition != payloadQuery.EventsCondition) &&
		((segmentQuery.EventsCondition == "" && payloadQuery.EventsCondition != model.EventCondAnyGivenEvent) ||
			(segmentQuery.EventsCondition == model.EventCondAnyGivenEvent && payloadQuery.EventsCondition != "")) {
		return true
	}

	// map of property and array of model.QueryProperty
	payloadGup := make(map[string][]model.QueryProperty)
	for _, gup := range payloadQuery.GlobalUserProperties {
		if _, exists := payloadGup[gup.Property]; !exists {
			payloadGup[gup.Property] = make([]model.QueryProperty, 0)
		}
		payloadGup[gup.Property] = append(payloadGup[gup.Property], gup)
	}

	// map of property and array of model.QueryEventWithProperties
	payloadEwp := make(map[string][]model.QueryEventWithProperties)
	for _, ewp := range payloadQuery.EventsWithProperties {
		if _, exists := payloadGup[ewp.Name]; !exists {
			payloadEwp[ewp.Name] = make([]model.QueryEventWithProperties, 0)
		}
		payloadEwp[ewp.Name] = append(payloadEwp[ewp.Name], ewp)
	}

	// check for gup
	for _, query := range segmentQuery.GlobalUserProperties {
		if _, exists := payloadGup[query.Property]; !exists {
			additionalFiltersExist = true
			break
		}
		filterExists := false
		for _, mapGupArr := range payloadGup[query.Property] {
			if reflect.DeepEqual(query, mapGupArr) {
				filterExists = true
				break
			}
		}
		if !filterExists {
			additionalFiltersExist = true
			break
		}
	}

	// gup not matched
	if additionalFiltersExist {
		return additionalFiltersExist
	}

	// check for ewp
	for _, query := range segmentQuery.EventsWithProperties {
		if _, exists := payloadEwp[query.Name]; !exists {
			additionalFiltersExist = true
			break
		}
		filterExists := false
		for _, mapEwpArr := range payloadEwp[query.Name] {
			if reflect.DeepEqual(query, mapEwpArr) {
				filterExists = true
				break
			}
		}
		if !filterExists {
			additionalFiltersExist = true
			break
		}
	}

	return additionalFiltersExist
}

func (store *MemSQL) GetDomainsListFromMarker(projectID int64, payload model.TimelinePayload,
	domainGroupID int, downloadLimitGiven bool) ([]model.Profile, bool, int, string) {

	logFields := log.Fields{
		"project_id": projectID,
		"payload":    payload,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

	isPreview := false
	segment, statusCode := store.GetSegmentById(projectID, payload.SegmentId)
	if statusCode != http.StatusFound {
		logCtx.Error("Segment not found.")
		return []model.Profile{}, isPreview, statusCode, "Failed to get segment"
	}

	lastRunTime, lastRunStatusCode := store.GetMarkerLastForAllAccounts(projectID)

	// all run for given segment is completed or not if recently updated/created
	runForGivenSegmentComplete := segment.MarkerRunSegment.After(segment.UpdatedAt)

	// Code duplicated to Memsql/segment analytics. Change in here should reflect there as well.
	// for case - segment is updated but all_run for the day is yet to run
	if lastRunStatusCode != http.StatusFound || (segment.UpdatedAt.After(lastRunTime) && !runForGivenSegmentComplete) {
		if C.IsMarkerPreviewEnabled(projectID) {
			return store.GetPreviewDomainsListByProjectId(projectID, payload, domainGroupID, downloadLimitGiven)
		}
		return []model.Profile{}, isPreview, lastRunStatusCode, ""
	}

	// if segment.UpdatedAt
	segmentQuery := model.Query{}
	err := U.DecodePostgresJsonbToStructType(segment.Query, &segmentQuery)
	if err != nil {
		log.WithField("project_id", projectID).Error("Unable to decode segment query")
		return []model.Profile{}, isPreview, statusCode, "Failed to get segment"
	}

	additionalFiltersExist := CompareFilters(segmentQuery, payload.Query)

	if additionalFiltersExist {
		if C.IsMarkerPreviewEnabled(projectID) {
			return store.GetPreviewDomainsListByProjectId(projectID, payload, domainGroupID, downloadLimitGiven)
		}
		return []model.Profile{}, isPreview, http.StatusOK, ""
	}

	// check for search filter
	whereForSearchFilters, searchFiltersParams := SearchFilterForAllAccounts(payload.SearchFilter, domainGroupID)

	var domLimit int
	if downloadLimitGiven {
		domLimit = 10000
	} else {
		domLimit = 1000
	}

	query := fmt.Sprintf(`WITH step_0 as (
		SELECT 
		  id as identity, 
		  group_%d_id as domain_name 
		FROM 
		  users 
		WHERE 
		  project_id = ? 
		  AND JSON_EXTRACT_STRING(
			associated_segments, ?
		  ) IS NOT NULL 
		  AND is_group_user = 1 
		  AND source = ? 
		  AND group_%d_id IS NOT NULL %s
		  AND is_deleted = false 
		LIMIT 
		  100000
	  ) 
	  SELECT 
		identity, 
		domain_name, 
		MAX(users.last_event_at) as last_activity 
	  FROM 
		step_0 
		JOIN (
		  SELECT 
			last_event_at, 
			group_%d_user_id 
		  FROM 
			users 
		  WHERE 
			users.project_id = ? 
			AND users.source != ? 
			AND last_event_at IS NOT NULL
			AND is_deleted = false
		) AS users ON step_0.identity = users.group_%d_user_id 
	  GROUP BY 
		identity 
	  ORDER BY 
		last_activity DESC 
	  LIMIT 
		%d;`, domainGroupID, domainGroupID, whereForSearchFilters,
		domainGroupID, domainGroupID, domLimit)

	params := []interface{}{projectID, payload.SegmentId, model.UserSourceDomains}

	if whereForSearchFilters != "" {
		params = append(params, searchFiltersParams...)
	}

	params = append(params, projectID, model.UserSourceDomains)

	var profiles []model.Profile
	db := C.GetServices().Db
	err = db.Raw(query, params...).Scan(&profiles).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return []model.Profile{}, isPreview, http.StatusNotFound, ""
		}
		return []model.Profile{}, isPreview, http.StatusInternalServerError, "Error in fetching rows for all accounts"
	}

	return profiles, isPreview, http.StatusOK, ""
}

func (store *MemSQL) GetPreviewDomainsListByProjectId(projectID int64, payload model.TimelinePayload,
	domainGroupID int, downloadLimitGiven bool) ([]model.Profile, bool, int, string) {

	runLimit := C.RunNumberToProcessDomainsForPreview()

	limitAcc := C.AccountLimitForPreview()
	userCount := new(int64)
	profilesList := make([]model.Profile, 0)

	// calculate limit to fetch total number of domains
	limitVal := C.DomainsToProcessForPreview() * runLimit

	isPreview := true

	// set listing limit to 1000 in case of all accounts listing
	if len(payload.Query.EventsWithProperties) == 0 && len(payload.Query.GlobalUserProperties) == 0 {
		// setting isPreview as false when no filters applied
		isPreview = false
		if downloadLimitGiven {
			limitAcc, limitVal = 10000, 10000
		} else {
			limitAcc, limitVal = 1000, 1000
		}
	}

	payload, fileValuesMap, eventNameIDsList, status, errMsg := store.transformPayload(projectID, payload)

	if status != http.StatusOK || errMsg != "" {
		return []model.Profile{}, isPreview, status, errMsg
	}

	// making all filter's operator as "OR" and skip domain level filters
	filtersForAllAccounts := modifyFiltersForAllAccounts(payload.Query.GlobalUserProperties)

	totalUserCount, status := store.GetAccountAssociatedUserCountByProjectID(projectID, domainGroupID)

	if status != http.StatusFound {
		return []model.Profile{}, isPreview, status, "Users Not Found"
	}

	if totalUserCount >= 1000000 {
		// not pruning query by filters if total user count more than 1M
		filtersForAllAccounts = make([]model.QueryProperty, 0)
	}

	// increase limit if no relevant filter to apply
	if len(filtersForAllAccounts) == 0 &&
		(len(payload.Query.GlobalUserProperties) > 0 || len(payload.Query.EventsWithProperties) > 0) {
		limitVal = 50000
	}

	startTime := time.Now().Unix()

	//Ashhar check this
	domainIDs, status := store.GetAllDomainsForPreviewByProjectID(projectID, domainGroupID, limitVal,
		filtersForAllAccounts, payload.SearchFilter)

	// return if no domains found
	if status != http.StatusFound || len(domainIDs) <= 0 {
		return []model.Profile{}, isPreview, status, "Failed to get domains list"
	}

	// breaking total domains list in small batches and running one at a time
	domainIDsList := U.GetStringListAsBatch(domainIDs, C.DomainsToProcessForPreview())

	for li := range domainIDsList {

		profiles, status, errMsg := store.GetPreviewDomainsListByProjectIdPerRun(projectID, payload,
			domainGroupID, eventNameIDsList, userCount, domainIDsList[li], limitAcc, fileValuesMap)

		if status != http.StatusOK || errMsg != "" {
			return []model.Profile{}, isPreview, status, errMsg
		}
		profilesList = append(profilesList, profiles...)

		if len(profiles) >= limitAcc {
			break
		}
	}

	// if profiles more than limitAcc, sort them in descending order and send top limitAcc profiles
	if len(profilesList) > limitAcc {
		sort.Sort(structSorter(profilesList))
		profilesList = profilesList[:limitAcc]
	}

	endTime := time.Now().Unix()
	timeTaken := endTime - startTime
	log.WithFields(log.Fields{"project_id": projectID, "user_count": *userCount}).Info("Time taken for preview query to compute results: ", timeTaken)

	return profilesList, isPreview, http.StatusOK, ""
}

func (store *MemSQL) GetPreviewDomainsListByProjectIdPerRun(projectID int64, payload model.TimelinePayload,
	domainGroupID int, eventNameIDsMap map[string]string, userCount *int64,
	domainIDList []string, limitAcc int, fileValuesMap map[string]map[string]bool) ([]model.Profile, int, string) {

	profiles := make([]model.Profile, 0)

	batchSize := C.BatchSizePreviewtMarker()

	domainIDChunks := U.GetStringListAsBatch(domainIDList, batchSize)
	var mu sync.Mutex

	for ci := range domainIDChunks {
		var wg sync.WaitGroup

		wg.Add(len(domainIDChunks[ci]))
		for _, domID := range domainIDChunks[ci] {
			go store.processDomain(projectID, payload, domainGroupID, userCount, domID, eventNameIDsMap,
				&wg, &mu, &profiles, limitAcc, fileValuesMap)
		}
		wg.Wait()

		if len(profiles) >= limitAcc {
			break
		}
	}

	return profiles, http.StatusOK, ""
}

func (store *MemSQL) transformPayload(projectID int64, payload model.TimelinePayload) (model.TimelinePayload,
	map[string]map[string]bool, map[string]string, int, string) {

	var status int
	// map for all event_names and all ids for it
	eventNameIDsMap := make(map[string]string)

	// map of all the inList and notInList files values
	fileValuesMap := make(map[string]map[string]bool)

	// datetime conversion
	err := payload.Query.TransformDateTypeFilters()
	if err != nil {
		log.WithField("project_id", projectID).Error("Failed to transform segment filters.")
		return payload, fileValuesMap, eventNameIDsMap, status, "Failed to transform segment filters"
	}

	// list of all event_names and all ids for it
	eventNameIDsList := make(map[string]bool)

	payload.Query.GlobalUserProperties = model.TransformPayloadForInProperties(payload.Query.GlobalUserProperties)
	payload.Query.From = U.TimeNowZ().AddDate(0, 0, -90).Unix()
	payload.Query.To = U.TimeNowZ().Unix()
	for index, eventProp := range payload.Query.EventsWithProperties {
		if eventProp.Name != "" {
			eventNameIDsList[eventProp.Name] = true
		}
		if eventProp.Range > int64(0) {
			days := int(eventProp.Range)
			timestampForRange := U.TimeNowIn(payload.Query.GetTimeZone()).AddDate(0, 0, -days).Unix()
			payload.Query.EventsWithProperties[index].Range = timestampForRange
		}
	}

	// adding ids for all the event_names
	if len(eventNameIDsList) > 0 {
		eventNameIDsMap, status = store.GetEventNameIdsWithGivenNames(projectID, eventNameIDsList)
		if status != http.StatusFound {
			log.WithField("project_id", projectID).Error("Error fetching event_names for the project")
			return payload, fileValuesMap, eventNameIDsMap, status, "Error fetching event_names for the project"
		}
	}

	// map for inList and NotInList
	fileValuesMap = GetFileValues(projectID, payload.Query.GlobalUserProperties, fileValuesMap)

	return payload, fileValuesMap, eventNameIDsMap, http.StatusOK, ""
}

func (store *MemSQL) processDomain(projectID int64, payload model.TimelinePayload, domainGroupID int,
	userCount *int64, domID string, eventNameIDsMap map[string]string, waitGroup *sync.WaitGroup, mu *sync.Mutex,
	profiles *[]model.Profile, limitAcc int, fileValuesMap map[string]map[string]bool) {
	logFields := log.Fields{
		"project_id":      projectID,
		"domain_group_id": domainGroupID,
		"domain_id":       domID,
		"wait_group":      waitGroup,
	}

	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()

	profile, isMatched, status := store.processDomainWithErr(projectID, payload, domainGroupID,
		userCount, domID, eventNameIDsMap, fileValuesMap)
	if status != http.StatusOK {
		log.WithFields(log.Fields{"project_id": projectID, "domID": domID}).Error("Error while computing for given payload")
		return
	}

	// returning if not matched
	if !isMatched {
		return
	}

	// locking before modifying the array
	addProfile(mu, profiles, limitAcc, profile)
}

func addProfile(mu *sync.Mutex, profiles *[]model.Profile, limitAcc int, profile model.Profile) {
	mu.Lock()
	defer mu.Unlock()

	*profiles = append(*profiles, profile)
}

func (store *MemSQL) processDomainWithErr(projectID int64, payload model.TimelinePayload, domainGroupID int,
	userCount *int64, domID string, eventNameIDsMap map[string]string, fileValuesMap map[string]map[string]bool) (model.Profile, bool, int) {

	isMatched := false

	users, status := store.GetAllPropertiesForDomain(projectID, domainGroupID, domID, userCount)
	if len(users) == 0 || status != http.StatusFound {
		return model.Profile{}, isMatched, status
	}

	var profile model.Profile
	var isLastEventAtFound bool
	if len(payload.Query.GlobalUserProperties) == 0 && len(payload.Query.EventsWithProperties) == 0 {
		profile, isLastEventAtFound = profileValues(projectID, users, domID, domainGroupID)
		return profile, isLastEventAtFound, http.StatusOK
	}

	decodedPropsArr := make([]map[string]interface{}, 0)

	for _, user := range users {
		// decoding user properties col
		decodedProps, err := U.DecodePostgresJsonb(&user.Properties)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": user.ID}).Error("Unable to decode user properties for user.")
			return model.Profile{}, isMatched, http.StatusInternalServerError
		}
		decodedPropsArr = append(decodedPropsArr, *decodedProps)
	}

	isMatched = IsRuleMatchedAllAccounts(projectID, payload.Query, decodedPropsArr, users,
		"", domID, eventNameIDsMap, fileValuesMap)

	if !isMatched {
		return model.Profile{}, isMatched, http.StatusOK
	}

	profile, isLastEventAtFound = profileValues(projectID, users, domID, domainGroupID)

	// if last_event_at does not exist, don't append show that profile
	isMatched = isMatched && isLastEventAtFound

	return profile, isMatched, http.StatusOK
}

func profileValues(projectID int64, users []model.User, domID string, domainGroupID int) (model.Profile, bool) {
	var profile model.Profile
	profile.Identity = domID

	maxLastEventAt := time.Time{}

	// set last_activity
	var domUser model.User
	for _, user := range users {
		// storing domain details
		if user.Source != nil && *user.Source == model.UserSourceDomains {
			domUser = user
			continue
		}
		if user.LastEventAt != nil && (*user.LastEventAt).After(maxLastEventAt) {
			maxLastEventAt = *user.LastEventAt
		}
	}

	// set hostName
	domainName, err := model.FindUserGroupByID(domUser, domainGroupID)
	if err != nil {
		domainName, err = model.ConvertDomainIdToHostName(domID)
		if err != nil || domainName == "" {
			log.WithFields(log.Fields{"project_id": projectID, "dom_id": domID}).
				WithError(err).Error("Couldn't translate ID to Hostname")
		}
	}
	profile.LastActivity = maxLastEventAt
	profile.DomainName = domainName

	isLastEventAtFound := !maxLastEventAt.IsZero()

	return profile, isLastEventAtFound
}

func modifyFiltersForAllAccounts(globalFilters []model.QueryProperty) []model.QueryProperty {
	modifiedFilters := make([]model.QueryProperty, 0)

	// Limiting number of filters used for initial pruning
	limitFilters := 5

	filterCount := 0
	for _, filter := range globalFilters {
		// using only given initial number of filters for pruning
		if filterCount == limitFilters {
			break
		}

		// not including negative filters and domain type filters
		if filter.GroupName == model.GROUP_NAME_DOMAINS || model.IsNegativeFilter(filter) {
			continue
		}
		filter.LogicalOp = "OR"
		modifiedFilters = append(modifiedFilters, filter)
		filterCount += 1
	}

	return modifiedFilters
}

type structSorter []model.Profile

func (a structSorter) Len() int {
	return len(a)
}
func (a structSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a structSorter) Less(i, j int) bool {
	return a[i].LastActivity.After(a[j].LastActivity)
}
