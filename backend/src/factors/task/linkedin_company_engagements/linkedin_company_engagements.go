package linkedin_company_engagements

import (
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type syncWorkerStatus struct {
	HasFailure bool
	ErrMsg     string
	StatusCode int
	Lock       sync.Mutex
}

func CreateGroupUserAndEventsV2(linkedinProjectSetting model.LinkedinProjectSettings, batchSize int) (string, int) {
	projectID, err := strconv.ParseInt(linkedinProjectSetting.ProjectId, 10, 64)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}
	log.Info("Starting processing for project ", projectID)

	eventNameViewedAD, eventNameClickedAD, errMsg, errCode := createDependentLinkedInEventNames(projectID)
	if errCode != http.StatusOK {
		return errMsg, errCode
	}
	timeZone, errCode := store.GetStore().GetTimezoneForProject(projectID)
	if errCode != http.StatusFound {
		return "Failed to get timezone", errCode
	}
	location, err := time.LoadLocation(string(timeZone))
	if err != nil {
		return "Failed to load location via timezone", http.StatusInternalServerError
	}

	distinctTimestamps, errCode := store.GetStore().GetDistinctTimestampsForEventCreationFromLinkedinDocs(linkedinProjectSetting.ProjectId)
	if errCode != http.StatusOK {
		return "Failed to get distinct timestamps for event creation from linkedin", errCode
	}
	for _, timestamp := range distinctTimestamps {
		domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
		if errCode != http.StatusOK {
			return "Failed to get domain data from linkedin", errCode
		}
		log.WithFields(log.Fields{"count": len(domainDataSet), "timestamp": timestamp}).Info("Number of documents to process")
		timestampStr := strconv.FormatInt(timestamp, 10)
		timestampForEventLookup, err := time.ParseInLocation("20060102", timestampStr, location)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		unixTimestampForEventLookup := timestampForEventLookup.Unix()
		// Campaign here means adgroup for adwords. Using campaign_group/campaign terminology since it's linkedin data
		isCampaignData := checkIfIncomingDataHasCampaigns(domainDataSet)
		if isCampaignData {
			return "Campaign flag not enabled for this project", http.StatusBadRequest
		}

		imprEventsMapWithCampaign, clicksEventsMapWithCampaign, err := store.GetStore().GetLinkedinEventFieldsBasedOnTimestampV2(projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		for len(domainDataSet) > 0 {
			var syncStatus syncWorkerStatus
			batchedDomainData := getBatchOfDomainDataV2(projectID, domainDataSet, batchSize)
			for _, domainDataBatch := range batchedDomainData {
				var wg sync.WaitGroup
				for _, domainData := range domainDataBatch {
					wg.Add(1)
					go createGroupUserAndEventsForGivenDomainDataBatchV2(projectID, eventNameViewedAD,
						eventNameClickedAD, location, domainData, imprEventsMapWithCampaign, clicksEventsMapWithCampaign, &wg, &syncStatus)
				}
				wg.Wait()
				if syncStatus.HasFailure {
					return syncStatus.ErrMsg, syncStatus.StatusCode
				}
			}
			domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
			if errCode != http.StatusOK {
				return "Failed to get domain data from linkedin", errCode
			}
		}
	}
	userIDsWithMismatch, errMsg, errCode := getUserIDsWithDataMismatch(projectID)
	if errMsg != "" {
		return errMsg, errCode
	}
	if userIDsWithMismatch != "" {
		log.WithFields(log.Fields{"projectID": projectID, "userIDs": userIDsWithMismatch}).Info("Mismatched data userIDs")
	}
	log.Info("Ended processing for project ", projectID)
	return "", http.StatusOK
}

func createDependentLinkedInEventNames(projectID int64) (*model.EventName, *model.EventName, string, int) {
	eventNameViewedAD, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: projectID,
		Name:      U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD,
	})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return nil, nil, "Failed in creating viewed ad event name", errCode
	}
	eventNameClickedAD, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: projectID,
		Name:      U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD,
	})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return nil, nil, "Failed in creating clicked ad event name", errCode
	}

	_, status := store.GetStore().CreateOrGetDomainsGroup(projectID)
	if status != http.StatusFound && status != http.StatusCreated {
		return nil, nil, "Failed to create GroupDomain", http.StatusInternalServerError
	}

	_, status = store.GetStore().CreateGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY, model.AllowedGroupNames)
	if status != http.StatusCreated && status != http.StatusConflict {
		return nil, nil, "Failed to create Group", http.StatusInternalServerError
	}

	return eventNameViewedAD, eventNameClickedAD, "", http.StatusOK
}

func createGroupUserAndEventsForGivenDomainDataBatchV2(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, groupDomainData []model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]map[string]interface{},
	clicksEventsMapWithCampaign map[int64]map[string]map[string]map[string]interface{}, wg *sync.WaitGroup, syncStatus *syncWorkerStatus) {
	defer wg.Done()
	for _, domainData := range groupDomainData {
		errMsg, errCode := createGroupUserAndEventsForDomainDataV2(projectID, eventNameViewedAD,
			eventNameClickedAD, location, domainData, imprEventsMapWithCampaign, clicksEventsMapWithCampaign)

		syncStatus.Lock.Lock()
		if errCode != http.StatusOK {
			syncStatus.HasFailure = true
			syncStatus.ErrMsg = errMsg
			syncStatus.StatusCode = errCode
			syncStatus.Lock.Unlock()
			break
		}
		syncStatus.Lock.Unlock()
	}
}
func createGroupUserAndEventsForDomainDataV2(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, domainData model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]map[string]interface{},
	clicksEventsMapWithCampaign map[int64]map[string]map[string]map[string]interface{}) (string, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"doument":    domainData,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if (domainData.Domain == "" || domainData.Domain == "$none") || (domainData.Impressions == 0 && domainData.Clicks == 0) {
		err := store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}
		return "", http.StatusOK
	}

	properties := U.PropertiesMap{
		U.LI_DOMAIN:            domainData.Domain,
		U.LI_HEADQUARTER:       domainData.HeadQuarters,
		U.LI_LOCALIZED_NAME:    domainData.LocalizedName,
		U.LI_VANITY_NAME:       domainData.VanityName,
		U.LI_PREFERRED_COUNTRY: domainData.PreferredCountry,
		U.LI_ORGANIZATION_ID:   domainData.ID,
	}

	timestamp, err := time.ParseInLocation("20060102", domainData.Timestamp, location)
	if err != nil {
		return err.Error(), http.StatusInternalServerError

	}

	unixTimestamp := timestamp.Unix()
	userID, errCode := SDK.TrackGroupWithDomain(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, domainData.Domain, properties, unixTimestamp)
	if errCode == http.StatusNotImplemented {
		err = store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}
		return "", http.StatusOK
	} else if errCode != http.StatusOK {
		logCtx.Fatal("Failed in TrackGroupWithDomain")
		return "Failed in TrackGroupWithDomain", errCode
	}

	// creating/updating impr event
	errMsg, errCode := createOrUpdateEventFromDomainDataV2(projectID, userID, eventNameViewedAD.ID, domainData, U.LI_AD_VIEW_COUNT, domainData.Impressions, unixTimestamp, imprEventsMapWithCampaign)
	if errMsg != "" {
		errMsg += " - impression event"
		logCtx.Error(errMsg)
		return errMsg, errCode
	}
	// creating/updating click event
	errMsg, errCode = createOrUpdateEventFromDomainDataV2(projectID, userID, eventNameClickedAD.ID, domainData, U.LI_AD_CLICK_COUNT, domainData.Clicks, unixTimestamp+1, clicksEventsMapWithCampaign)
	if errMsg != "" {
		errMsg += " - click event"
		logCtx.Error(errMsg)
		return errMsg, errCode
	}
	existingImprCount := getExistingPropertyValue(domainData.Domain, unixTimestamp, domainData.CampaignGroupID, imprEventsMapWithCampaign)
	existingClickCount := getExistingPropertyValue(domainData.Domain, unixTimestamp+1, domainData.CampaignGroupID, clicksEventsMapWithCampaign)
	impr_diff := float64(domainData.Impressions) - existingImprCount
	clicks_diff := float64(domainData.Clicks) - existingClickCount

	userIDToUpdate := getUserIDFromEventsForUpdatingGroupUser(userID, domainData.Domain, unixTimestamp, domainData.CampaignGroupID, imprEventsMapWithCampaign)
	if userID != userIDToUpdate {
		log.WithFields(log.Fields{"projectID": projectID, "userID": userID, "userIDToUpdate": userIDToUpdate}).Info("Different user ID updated")
	}
	groupID := U.GetDomainGroupDomainName(projectID, domainData.Domain)
	if groupID == "" {
		return "Failed to get domain for raw domain", http.StatusNotImplemented
	}
	errMsg, errCode = updateAccountLevelPropertiesForGroupUser(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, groupID, userIDToUpdate, impr_diff, clicks_diff)
	if errMsg != "" {
		logCtx.Error(errMsg)
		return errMsg, errCode
	}

	err = store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
	if err != nil {
		logCtx.WithError(err).Error("Failed in updating user creation details after event and user update")
		return "Failed in updating user creation details after event and user update", http.StatusInternalServerError
	}
	return "", http.StatusOK
}

func createOrUpdateEventFromDomainDataV2(projectID int64, userID string, eventNameID string,
	domainData model.DomainDataResponse, propertyName string, propertyValue int64, timestamp int64,
	eventLookupMap map[int64]map[string]map[string]map[string]interface{}) (string, int) {

	event := model.Event{
		EventNameId: eventNameID,
		Timestamp:   timestamp,
		ProjectId:   projectID,
		UserId:      userID,
	}
	eventPropertiesMap := U.PropertiesMap{
		propertyName:      propertyValue,
		U.EP_CAMPAIGN:     domainData.CampaignGroupName,
		U.EP_CAMPAIGN_ID:  domainData.CampaignGroupID,
		U.EP_SKIP_SESSION: U.PROPERTY_VALUE_TRUE,
	}
	isEventReq, eventID, eventUserID := checkIfEventCreationReqV2(propertyValue, domainData.Domain, timestamp, domainData.CampaignGroupID, eventLookupMap)
	if isEventReq {
		eventPropertiesJsonB, err := U.EncodeStructTypeToPostgresJsonb(&eventPropertiesMap)
		if err != nil {
			return "Failed in encoding properties to JSONb", http.StatusInternalServerError
		}
		event.Properties = *eventPropertiesJsonB
		_, errCode := store.GetStore().CreateEvent(&event)
		if errCode != http.StatusCreated {
			return "Failed in creating event", errCode
		}
	} else if eventID != "" {
		errCode := store.GetStore().UpdateEventProperties(projectID, eventID, eventUserID, &eventPropertiesMap, timestamp, nil)
		if errCode != http.StatusAccepted {
			log.WithFields(log.Fields{"projectID": projectID, "eventID": eventID, "userID": eventUserID, "timestamp": timestamp, "props": eventPropertiesMap}).Error("Failed in updating event")
			return "Failed in updating event", errCode
		}
	}

	return "", http.StatusOK
}

func updateAccountLevelPropertiesForGroupUser(projectID int64, groupName string, groupID string, groupUserID string, impr_diff float64, clicks_diff float64) (string, int) {

	groupUser, errCode := store.GetStore().GetGroupUserByGroupID(projectID, groupName, groupID)
	if errCode != http.StatusFound {
		log.WithFields(log.Fields{"project_id": projectID, "group_name": groupName, "group_id": groupID, "groupUserID": groupUserID}).Error("Ashhar_1 - Failed to get existing group user")
		return "Failed to get existing group user", errCode
	}
	existingProperties, err := U.DecodePostgresJsonb(&groupUser.Properties)
	if err != nil {
		return "Failed to decode user properties on UpdateUserGroupProperties.", http.StatusInternalServerError
	}
	newProperties := make(U.PropertiesMap)
	if value, exists := (*existingProperties)[U.LI_TOTAL_AD_VIEW_COUNT]; exists {
		newProperties[U.LI_TOTAL_AD_VIEW_COUNT] = value.(float64) + impr_diff
	} else {
		newProperties[U.LI_TOTAL_AD_VIEW_COUNT] = impr_diff
	}
	if value, exists := (*existingProperties)[U.LI_TOTAL_AD_CLICK_COUNT]; exists {
		newProperties[U.LI_TOTAL_AD_CLICK_COUNT] = value.(float64) + clicks_diff
	} else {
		newProperties[U.LI_TOTAL_AD_CLICK_COUNT] = clicks_diff
	}

	propertiesMap := map[string]interface{}(newProperties)
	source := model.GetGroupUserSourceNameByGroupName(groupName)

	currTimestamp := time.Now().Unix()
	// check any case if any group user is created
	// no way to differentiate
	_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID, groupUserID, &propertiesMap, currTimestamp, currTimestamp, source)
	if err != nil {
		return "Failed to create or update group user on updateAccountLevelPropertiesForGroupUser.", http.StatusInternalServerError
	}
	return "", http.StatusOK
}

func CreateGroupUserAndEventsV3(linkedinProjectSetting model.LinkedinProjectSettings, batchSize int) (string, int) {
	projectID, err := strconv.ParseInt(linkedinProjectSetting.ProjectId, 10, 64)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}
	log.Info("Starting processing for project ", projectID)

	// can be moved to a separate function. line 313
	eventNameViewedAD, eventNameClickedAD, errMsg, errCode := createDependentLinkedInEventNames(projectID)
	if errCode != http.StatusOK {
		return errMsg, errCode
	}
	timeZone, errCode := store.GetStore().GetTimezoneForProject(projectID)
	if errCode != http.StatusFound {
		return "Failed to get timezone", errCode
	}
	location, err := time.LoadLocation(string(timeZone))
	if err != nil {
		return "Failed to load location via timezone", http.StatusInternalServerError
	}
	// to line 324

	distinctTimestamps, errCode := store.GetStore().GetDistinctTimestampsForEventCreationFromLinkedinDocs(linkedinProjectSetting.ProjectId)
	if errCode != http.StatusOK {
		return "Failed to get distinct timestamps for event creation from linkedin", errCode
	}
	for _, timestamp := range distinctTimestamps {
		domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
		if errCode != http.StatusOK {
			return "Failed to get domain data from linkedin", errCode
		}
		log.WithFields(log.Fields{"count": len(domainDataSet), "timestamp": timestamp}).Info("Number of documents to process")
		timestampStr := strconv.FormatInt(timestamp, 10)
		timestampForEventLookup, err := time.ParseInLocation("20060102", timestampStr, location)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		unixTimestampForEventLookup := timestampForEventLookup.Unix()

		isPartialDataPresent := checkIfPartialDataIsPresent(projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID)
		if isPartialDataPresent {
			log.WithFields(log.Fields{"projectID": projectID, "timestamp": unixTimestampForEventLookup}).Error("Partial data present")
			return "Partial data present", http.StatusInternalServerError
		}

		// Campaign here means adgroup for adwords. Using campaign_group/campaign terminology since it's linkedin data
		isCampaignData := checkIfIncomingDataHasCampaigns(domainDataSet)

		if !isCampaignData {
			errMsg, errCode := deleteEventsAndUpdateAccountPropertiesBasedOnType(projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID, true, batchSize)
			if errCode != http.StatusAccepted {
				return errMsg, errCode
			}

			imprEventsMapWithCampaignGroup, clicksEventsMapWithCampaignGroup, err := store.GetStore().GetLinkedinEventFieldsBasedOnTimestampV2(projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID)
			if err != nil {
				return err.Error(), http.StatusInternalServerError
			}
			for len(domainDataSet) > 0 {
				var syncStatus syncWorkerStatus
				batchedDomainData := getBatchOfDomainDataV2(projectID, domainDataSet, batchSize)
				for _, domainDataBatch := range batchedDomainData {
					var wg sync.WaitGroup
					for _, domainData := range domainDataBatch {
						wg.Add(1)
						go createGroupUserAndEventsForGivenDomainDataBatchV2(projectID, eventNameViewedAD,
							eventNameClickedAD, location, domainData, imprEventsMapWithCampaignGroup, clicksEventsMapWithCampaignGroup, &wg, &syncStatus)
					}
					wg.Wait()
					if syncStatus.HasFailure {
						return syncStatus.ErrMsg, syncStatus.StatusCode
					}
				}
				domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
				if errCode != http.StatusOK {
					return "Failed to get domain data from linkedin", errCode
				}
			}
		} else {
			errMsg, errCode := deleteEventsAndUpdateAccountPropertiesBasedOnType(projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID, false, batchSize)
			if errCode != http.StatusAccepted {
				return errMsg, errCode
			}
			imprEventsMapWithCampaign, clicksEventsMapWithCampaign, err := store.GetStore().GetLinkedinEventFieldsBasedOnTimestampV3(projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID)
			if err != nil {
				return err.Error(), http.StatusInternalServerError
			}
			for len(domainDataSet) > 0 {
				var syncStatus syncWorkerStatus
				batchedDomainData := getBatchOfDomainDataV2(projectID, domainDataSet, batchSize)
				for _, domainDataBatch := range batchedDomainData {
					var wg sync.WaitGroup
					for _, domainData := range domainDataBatch {
						wg.Add(1)
						go createGroupUserAndEventsForGivenDomainDataBatchV3(projectID, eventNameViewedAD,
							eventNameClickedAD, location, domainData, imprEventsMapWithCampaign, clicksEventsMapWithCampaign, &wg, &syncStatus)
					}
					wg.Wait()
					if syncStatus.HasFailure {
						return syncStatus.ErrMsg, syncStatus.StatusCode
					}
				}
				domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
				if errCode != http.StatusOK {
					return "Failed to get domain data from linkedin", errCode
				}
			}
		}
	}

	userIDsWithMismatch, errMsg, errCode := getUserIDsWithDataMismatch(projectID)
	if errMsg != "" {
		return errMsg, errCode
	}
	if userIDsWithMismatch != "" {
		log.WithFields(log.Fields{"projectID": projectID, "userIDs": userIDsWithMismatch}).Info("Mismatched data userIDs")
	}
	log.Info("Ended processing for project ", projectID)
	return "", http.StatusOK
}

func deleteEventsAndUpdateAccountPropertiesBasedOnType(projectID int64, timestamp int64, imprEventNameID, clickEventNameID string, deleteCampaignType bool,
	batchSize int) (string, int) {

	userIDToUserInfoForDeleteAndUpdate, err := getUserInfoDeleteAndUpdate(projectID, timestamp, imprEventNameID, clickEventNameID, deleteCampaignType)
	if err != nil {
		log.WithFields(log.Fields{"projectID": projectID, "timestamp": timestamp, "isDeleteCampaign": deleteCampaignType}).Info("Failed in getting affected ")
		return err.Error(), http.StatusInternalServerError
	}
	allUserIds := make([]string, 0)
	for key := range userIDToUserInfoForDeleteAndUpdate {
		allUserIds = append(allUserIds, key)
	}
	log.WithFields(log.Fields{"projectID": projectID, "timestamp": timestamp, "isDeleteCampaign": deleteCampaignType, "userIDs": allUserIds}).Info("Deleting events for given users IDs")

	batchedUserInfo, errMsg := buildBatchOfGroupUsersBasedOnUserIDs(projectID, userIDToUserInfoForDeleteAndUpdate, batchSize)
	if errMsg != "" {
		return errMsg, http.StatusInternalServerError
	}

	// var syncStatus syncWorkerStatus
	// for _, batch := range batchedUserInfo {
	// 	var wg sync.WaitGroup
	// 	for _, user := range batch {
	// 		wg.Add(1)
	// 		go deleteEventsAndUpdateAccountPropertiesForUser(projectID, user, userIDToUserInfoForDeleteAndUpdate[user.ID], imprEventNameID, clickEventNameID, &wg, &syncStatus)
	// 	}
	// 	wg.Wait()
	// 	if syncStatus.HasFailure {
	// 		return syncStatus.ErrMsg, syncStatus.StatusCode
	// 	}
	// }
	for _, batch := range batchedUserInfo {
		for _, user := range batch {
			errMsg, errCode := deleteEventsAndUpdateAccountProperties(projectID, user, userIDToUserInfoForDeleteAndUpdate[user.ID], imprEventNameID, clickEventNameID)
			if errMsg != "" {
				return errMsg, errCode
			}
		}
	}

	return "", http.StatusAccepted
}

func deleteEventsAndUpdateAccountPropertiesForUser(projectID int64, user model.User, userInfo UserInfoForDeleteAndUpdate, imprEventNameID, clickEventNameID string, wg *sync.WaitGroup, syncStatus *syncWorkerStatus) {
	defer wg.Done()
	errMsg, errCode := deleteEventsAndUpdateAccountProperties(projectID, user, userInfo, imprEventNameID, clickEventNameID)
	if errCode != http.StatusAccepted {
		log.WithFields(log.Fields{"projectID": projectID, "userID": user.ID, "errMsg": errMsg}).Error("Failed in deletion of events and users updation")
	}

	syncStatus.Lock.Lock()
	if errCode != http.StatusAccepted {
		syncStatus.HasFailure = true
		syncStatus.ErrMsg = errMsg
		syncStatus.StatusCode = errCode
		syncStatus.Lock.Unlock()
	}
	syncStatus.Lock.Unlock()
}

func createGroupUserAndEventsForGivenDomainDataBatchV3(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, groupDomainData []model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]model.ValueForEventLookupMap,
	clicksEventsMapWithCampaign map[int64]map[string]map[string]model.ValueForEventLookupMap, wg *sync.WaitGroup, syncStatus *syncWorkerStatus) {
	defer wg.Done()
	for _, domainData := range groupDomainData {
		errMsg, errCode := createGroupUserAndEventsForDomainDataV3(projectID, eventNameViewedAD,
			eventNameClickedAD, location, domainData, imprEventsMapWithCampaign, clicksEventsMapWithCampaign)
		if errCode != http.StatusOK {
			log.WithFields(log.Fields{"projectID": projectID, "domainData": domainData, "errMsg": errMsg}).Error("Failed in user and event creation")
		}

		syncStatus.Lock.Lock()
		if errCode != http.StatusOK {
			syncStatus.HasFailure = true
			syncStatus.ErrMsg = errMsg
			syncStatus.StatusCode = errCode
			syncStatus.Lock.Unlock()
			break
		}
		syncStatus.Lock.Unlock()
	}
}
func createGroupUserAndEventsForDomainDataV3(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, domainData model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]model.ValueForEventLookupMap,
	clicksEventsMapWithCampaign map[int64]map[string]map[string]model.ValueForEventLookupMap) (string, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"doument":    domainData,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if (domainData.Domain == "" || domainData.Domain == "$none") || (domainData.Impressions == 0 && domainData.Clicks == 0) {
		err := store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}
		return "", http.StatusOK
	}

	properties := U.PropertiesMap{
		U.LI_DOMAIN:            domainData.Domain,
		U.LI_HEADQUARTER:       domainData.HeadQuarters,
		U.LI_LOCALIZED_NAME:    domainData.LocalizedName,
		U.LI_VANITY_NAME:       domainData.VanityName,
		U.LI_PREFERRED_COUNTRY: domainData.PreferredCountry,
		U.LI_ORGANIZATION_ID:   domainData.OrgID,
	}

	timestamp, err := time.ParseInLocation("20060102", domainData.Timestamp, location)
	if err != nil {
		return err.Error(), http.StatusInternalServerError

	}

	unixTimestamp := timestamp.Unix()
	userID, errCode := SDK.TrackGroupWithDomain(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, domainData.Domain, properties, unixTimestamp)
	if errCode == http.StatusNotImplemented {
		// StatusNotImplemented means that we cannot parse the domain, assuming domain is invalid. Hence moving on
		err = store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}
		return "", http.StatusOK
	} else if errCode != http.StatusOK {
		logCtx.Fatal("Failed in TrackGroupWithDomain")
		return "Failed in TrackGroupWithDomain", errCode
	}

	// creating/updating impr event
	errMsg, errCode := createOrUpdateEventFromDomainDataV3(projectID, userID, eventNameViewedAD.ID, domainData, U.LI_AD_VIEW_COUNT, domainData.Impressions, unixTimestamp, imprEventsMapWithCampaign)
	if errMsg != "" {
		errMsg += " - impression event"
		logCtx.Error(errMsg)
		return errMsg, errCode
	}
	// creating/updating click event
	errMsg, errCode = createOrUpdateEventFromDomainDataV3(projectID, userID, eventNameClickedAD.ID, domainData, U.LI_AD_CLICK_COUNT, domainData.Clicks, unixTimestamp+1, clicksEventsMapWithCampaign)
	if errMsg != "" {
		errMsg += " - click event"
		logCtx.Error(errMsg)
		return errMsg, errCode
	}
	existingImprCount := getExistingPropertyValueV3(domainData.OrgID, unixTimestamp, domainData.CampaignID, imprEventsMapWithCampaign)
	existingClickCount := getExistingPropertyValueV3(domainData.OrgID, unixTimestamp+1, domainData.CampaignID, clicksEventsMapWithCampaign)
	impr_diff := float64(domainData.Impressions) - existingImprCount
	clicks_diff := float64(domainData.Clicks) - existingClickCount

	userIDToUpdate := getUserIDFromEventsForUpdatingGroupUserV3(userID, domainData.OrgID, unixTimestamp, domainData.CampaignID, imprEventsMapWithCampaign)
	if userID != userIDToUpdate {
		log.WithFields(log.Fields{"projectID": projectID, "userID": userID, "userIDToUpdate": userIDToUpdate}).Info("Different user ID updated")
	}
	groupID := U.GetDomainGroupDomainName(projectID, domainData.Domain)
	if groupID == "" {
		return "Failed to get domain for raw domain", http.StatusNotImplemented
	}
	errMsg, errCode = updateAccountLevelPropertiesForGroupUser(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, groupID, userIDToUpdate, impr_diff, clicks_diff)
	if errMsg != "" {
		logCtx.Error(errMsg)
		return errMsg, errCode
	}

	err = store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
	if err != nil {
		logCtx.WithError(err).Error("Failed in updating user creation details after event and user update")
		return "Failed in updating user creation details after event and user update", http.StatusInternalServerError
	}
	return "", http.StatusOK
}

func createOrUpdateEventFromDomainDataV3(projectID int64, userID string, eventNameID string,
	domainData model.DomainDataResponse, propertyName string, propertyValue int64, timestamp int64,
	eventLookupMap map[int64]map[string]map[string]model.ValueForEventLookupMap) (string, int) {

	event := model.Event{
		EventNameId: eventNameID,
		Timestamp:   timestamp,
		ProjectId:   projectID,
		UserId:      userID,
	}
	eventPropertiesMap := U.PropertiesMap{
		propertyName:         propertyValue,
		U.EP_CAMPAIGN:        domainData.CampaignGroupName,
		U.EP_CAMPAIGN_ID:     domainData.CampaignGroupID,
		U.EP_ADGROUP_ID:      domainData.CampaignID,
		U.EP_ADGROUP:         domainData.CampaignName,
		U.EP_SKIP_SESSION:    U.PROPERTY_VALUE_TRUE,
		U.LI_ORGANIZATION_ID: domainData.ID,
		U.LI_RAW_URL:         domainData.RawDomain,
	}
	isEventReq, eventID, eventUserID := checkIfEventCreationReqV3(propertyValue, domainData.OrgID, timestamp, domainData.CampaignID, eventLookupMap)
	if isEventReq {
		eventPropertiesJsonB, err := U.EncodeStructTypeToPostgresJsonb(&eventPropertiesMap)
		if err != nil {
			return "Failed in encoding properties to JSONb", http.StatusInternalServerError
		}
		event.Properties = *eventPropertiesJsonB
		_, errCode := store.GetStore().CreateEvent(&event)
		if errCode != http.StatusCreated {
			return "Failed in creating event", errCode
		}
	} else if eventID != "" {
		errCode := store.GetStore().UpdateEventProperties(projectID, eventID, eventUserID, &eventPropertiesMap, timestamp, nil)
		if errCode != http.StatusAccepted {
			log.WithFields(log.Fields{"projectID": projectID, "eventID": eventID, "userID": eventUserID, "timestamp": timestamp, "props": eventPropertiesMap}).Error("Failed in updating event")
			return "Failed in updating event", errCode
		}
	}

	return "", http.StatusOK
}
