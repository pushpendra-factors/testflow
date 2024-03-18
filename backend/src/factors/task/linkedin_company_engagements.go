package task

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func CreateGroupUserAndEvents(linkedinProjectSetting model.LinkedinProjectSettings) (string, int) {
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
		log.WithFields(log.Fields{"count": len(domainDataSet), "timestamp": timestamp}).Info("Number of documents to process for timestamp")
		timestampStr := strconv.FormatInt(timestamp, 10)
		timestampForEventLookup, err := time.ParseInLocation("20060102", timestampStr, location)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		unixTimestampForEventLookup := timestampForEventLookup.Unix()
		/*
			type LinkedinEventFields struct {
				Timestamp       int64  `json:"timestamp"`
				CampaignGroupID string `json:"campaign_group_id"`
				Domain          string `json:"domain"`
			}
		*/
		imprEventsMapWithCampaign, clicksEventsMapWithCampaign, err := store.GetStore().GetLinkedinEventFieldsBasedOnTimestamp(
			projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		for len(domainDataSet) > 0 {
			errMsg, errCode := createGroupUserAndEventsForGivenDomainData(projectID, eventNameViewedAD,
				eventNameClickedAD, location, domainDataSet, imprEventsMapWithCampaign, clicksEventsMapWithCampaign)
			if errCode != http.StatusOK {
				return errMsg, errCode
			}
			domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
			if errCode != http.StatusOK {
				return "Failed to get domain data from linkedin", errCode
			}
		}
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

func createGroupUserAndEventsForGivenDomainData(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, domainDataSet []model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]bool,
	clicksEventsMapWithCampaign map[int64]map[string]map[string]bool) (string, int) {

	for _, domainData := range domainDataSet {
		logFields := log.Fields{
			"project_id": projectID,
			"doument":    domainData,
		}
		defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
		logCtx := log.WithFields(logFields)

		if domainData.Domain != "" && domainData.Domain != "$none" {
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
				continue
			} else if errCode != http.StatusOK {
				logCtx.Error("Failed in TrackGroupWithDomain")
				return "Failed in TrackGroupWithDomain", errCode
			}

			isImprEventReq := checkIfEventCreationReq(domainData.Impressions, domainData.Domain, unixTimestamp, domainData.CampaignGroupID, imprEventsMapWithCampaign)
			if isImprEventReq {
				viewedADEvent := model.Event{
					EventNameId: eventNameViewedAD.ID,
					Timestamp:   unixTimestamp,
					ProjectId:   projectID,
					UserId:      userID,
				}
				viewedADEventPropertiesMap := map[string]interface{}{
					U.LI_AD_VIEW_COUNT: domainData.Impressions,
					U.EP_CAMPAIGN:      domainData.CampaignGroupName,
					U.EP_CAMPAIGN_ID:   domainData.CampaignGroupID,
					U.EP_SKIP_SESSION:  U.PROPERTY_VALUE_TRUE,
				}
				viewedADEventPropertiesJsonB, err := U.EncodeStructTypeToPostgresJsonb(&viewedADEventPropertiesMap)
				if err != nil {
					logCtx.WithError(err).Error("Failed in encoding viewed ad properties to JSONb")
					return "Failed in encoding viewed ad properties to JSONb", http.StatusInternalServerError
				}
				viewedADEvent.Properties = *viewedADEventPropertiesJsonB

				_, errCode = store.GetStore().CreateEvent(&viewedADEvent)
				if errCode != http.StatusCreated {
					logCtx.Error("Failed in creating viewed ad event")
					return "Failed in creating viewed ad event", errCode
				}
			}

			isCLickEventReq := checkIfEventCreationReq(domainData.Clicks, domainData.Domain, unixTimestamp+1, domainData.CampaignGroupID, clicksEventsMapWithCampaign)
			if isCLickEventReq {
				clickedADEvent := model.Event{
					EventNameId: eventNameClickedAD.ID,
					Timestamp:   unixTimestamp + 1,
					ProjectId:   projectID,
					UserId:      userID,
				}
				clickedADEventPropertiesMap := map[string]interface{}{
					U.LI_AD_CLICK_COUNT: domainData.Clicks,
					U.EP_CAMPAIGN:       domainData.CampaignGroupName,
					U.EP_CAMPAIGN_ID:    domainData.CampaignGroupID,
					U.EP_SKIP_SESSION:   U.PROPERTY_VALUE_TRUE,
				}
				clickedADEventPropertiesJsonB, err := U.EncodeStructTypeToPostgresJsonb(&clickedADEventPropertiesMap)
				if err != nil {
					logCtx.WithError(err).Error("Failed in encoding clicked ad properties to JSONb")
					return "Failed in encoding clicked ad properties to JSONb", http.StatusInternalServerError
				}
				clickedADEvent.Properties = *clickedADEventPropertiesJsonB

				_, errCode = store.GetStore().CreateEvent(&clickedADEvent)
				if errCode != http.StatusCreated {
					logCtx.Error("Failed in creating clicked ad event")
					return "Failed in creating clicked ad event", errCode
				}
			}
		}

		err := store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}
	}
	return "", http.StatusOK
}
func checkIfEventCreationReq(propertyValue int64, domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]bool) bool {
	if propertyValue <= 0 {
		return false
	}
	if _, exists := existingEventsWithCampaignData[timestamp][domain][campaignGroupID]; exists {
		return false
	}

	return true
}

type syncWorkerStatus struct {
	HasFailure bool
	ErrMsg     string
	StatusCode int
	Lock       sync.Mutex
}

func CreateGroupUserAndEventsV1(linkedinProjectSetting model.LinkedinProjectSettings, batchSize int) (string, int) {
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
		log.WithFields(log.Fields{"count": len(domainDataSet), "timestamp": timestamp}).Info("Number of documents to process for timestamp")
		timestampStr := strconv.FormatInt(timestamp, 10)
		timestampForEventLookup, err := time.ParseInLocation("20060102", timestampStr, location)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		unixTimestampForEventLookup := timestampForEventLookup.Unix()
		/*
			type LinkedinEventFieldsNew struct {
				Timestamp       int64  `json:"timestamp"`
				CampaignGroupID string `json:"campaign_group_id"`
				Domain          string `json:"domain"`
				ID 				string `json:"id"`
			}
		*/
		imprEventsMapWithCampaign, clicksEventsMapWithCampaign, err := store.GetStore().GetLinkedinEventFieldsBasedOnTimestampV1(
			projectID, unixTimestampForEventLookup, eventNameViewedAD.ID, eventNameClickedAD.ID)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		for len(domainDataSet) > 0 {
			var syncStatus syncWorkerStatus
			batchedDomainData := getBatchOfDomainDataV1(domainDataSet, batchSize)
			for _, domainDataBatch := range batchedDomainData {
				var wg sync.WaitGroup
				for _, domainData := range domainDataBatch {
					wg.Add(1)
					go createGroupUserAndEventsForGivenDomainDataBatchV1(projectID, eventNameViewedAD,
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
	log.Info("Ended processing for project ", projectID)
	return "", http.StatusOK
}

func getBatchOfDomainDataV1(domainDataSet []model.DomainDataResponse, batchSize int) [][]model.DomainDataResponse {
	batchedDomainData := make([][]model.DomainDataResponse, 0)
	for i := 0; i < len(domainDataSet); i += batchSize {
		end := i + batchSize

		if end > len(domainDataSet) {
			end = len(domainDataSet)
		}
		batchedDomainData = append(batchedDomainData, domainDataSet[i:end])
	}
	return batchedDomainData
}

func createGroupUserAndEventsForGivenDomainDataBatchV1(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, domainData model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]string,
	clicksEventsMapWithCampaign map[int64]map[string]map[string]string, wg *sync.WaitGroup, syncStatus *syncWorkerStatus) {
	defer wg.Done()
	errMsg, errCode := createGroupUserAndEventsForDomainDataV1(projectID, eventNameViewedAD,
		eventNameClickedAD, location, domainData, imprEventsMapWithCampaign, clicksEventsMapWithCampaign)

	syncStatus.Lock.Lock()
	defer syncStatus.Lock.Unlock()
	if errCode != http.StatusOK {
		syncStatus.HasFailure = true
		syncStatus.ErrMsg = errMsg
		syncStatus.StatusCode = errCode
	}
}
func createGroupUserAndEventsForDomainDataV1(projectID int64, eventNameViewedAD *model.EventName,
	eventNameClickedAD *model.EventName, location *time.Location, domainData model.DomainDataResponse,
	imprEventsMapWithCampaign map[int64]map[string]map[string]string,
	clicksEventsMapWithCampaign map[int64]map[string]map[string]string) (string, int) {

	logFields := log.Fields{
		"project_id": projectID,
		"doument":    domainData,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if domainData.Domain == "" || domainData.Domain == "$none" {
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
	errMsg, errCode := createOrUpdateEventFromDomainDataV1(projectID, userID, eventNameViewedAD.ID, domainData, U.LI_AD_VIEW_COUNT, domainData.Impressions, unixTimestamp, imprEventsMapWithCampaign)
	if errMsg != "" {
		logCtx.Error(errMsg)
		return errMsg, errCode
	}
	// creating/updating click event
	errMsg, errCode = createOrUpdateEventFromDomainDataV1(projectID, userID, eventNameClickedAD.ID, domainData, U.LI_AD_CLICK_COUNT, domainData.Clicks, unixTimestamp+1, clicksEventsMapWithCampaign)
	if errMsg != "" {
		logCtx.Error(errMsg)
		return errMsg, errCode
	}
	err = store.GetStore().UpdateSyncStatusLinkedinDocs(domainData)
	if err != nil {
		logCtx.WithError(err).Error("Failed in updating user creation details")
		return "Failed in updating user creation details", http.StatusInternalServerError
	}
	return "", http.StatusOK
}

func createOrUpdateEventFromDomainDataV1(projectID int64, userID string, eventNameID string,
	domainData model.DomainDataResponse, propertyName string, propertyValue int64, timestamp int64,
	eventLookupMap map[int64]map[string]map[string]string) (string, int) {
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
	isEventReq, eventID := checkIfEventCreationReqV1(propertyValue, domainData.Domain, timestamp, domainData.CampaignGroupID, eventLookupMap)
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
		errCode := store.GetStore().UpdateEventProperties(projectID, eventID, "", &eventPropertiesMap, timestamp, nil)
		if errCode != http.StatusAccepted {
			log.WithFields(log.Fields{"projectID": projectID, "eventID": eventID, "timestamp": timestamp, "props": eventPropertiesMap}).Error("Failed in updating event")
			return "Failed in updating event", errCode
		}
	}

	return "", http.StatusOK
}

func checkIfEventCreationReqV1(propertyValue int64, domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]string) (bool, string) {
	if propertyValue <= 0 {
		return false, ""
	}
	if id, exists := existingEventsWithCampaignData[timestamp][domain][campaignGroupID]; exists {
		return false, id
	}

	return true, ""
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

/*
linkedin websites, a.b.com, b.b.com, a.c.com, b.c.com, d.com
 1. getGroupingByDomainV2:
    [[a.b.com, b.b.com], [a.c.com, b.c.com], [d.com]]
 2. final return for batch length 2
    [[[a.b.com, b.b.com], [a.c.com, b.c.com]], [[d.com]]]
    `[[a.b.com, b.b.com], [a.c.com, b.c.com]]` is batch 1
 3. rows with same domain will run serially
*/
func getBatchOfDomainDataV2(projectID int64, domainDataSet []model.DomainDataResponse, batchSize int) [][][]model.DomainDataResponse {

	groupedDomainDataSet := getGroupingByDomainV2(projectID, domainDataSet)
	batchedDomainData := make([][][]model.DomainDataResponse, 0)
	for i := 0; i < len(groupedDomainDataSet); i += batchSize {
		end := i + batchSize

		if end > len(groupedDomainDataSet) {
			end = len(groupedDomainDataSet)
		}
		batchedDomainData = append(batchedDomainData, groupedDomainDataSet[i:end])
	}
	return batchedDomainData
}

func getGroupingByDomainV2(projectID int64, domainDataSet []model.DomainDataResponse) [][]model.DomainDataResponse {
	groupedDomainDataSet := make([][]model.DomainDataResponse, 0)
	mapOfDomainToDomainDataSet := make(map[string][]model.DomainDataResponse)
	for _, domainData := range domainDataSet {
		domain := U.GetDomainGroupDomainName(projectID, domainData.Domain)
		if _, exists := mapOfDomainToDomainDataSet[domain]; !exists {
			mapOfDomainToDomainDataSet[domain] = make([]model.DomainDataResponse, 0)
		}
		mapOfDomainToDomainDataSet[domain] = append(mapOfDomainToDomainDataSet[domain], domainData)
	}
	for key, groupedDomainData := range mapOfDomainToDomainDataSet {
		groupedDomainDataSet = append(groupedDomainDataSet, groupedDomainData)
		if len(groupedDomainData) > 10 {
			log.WithFields(log.Fields{"project_id": projectID, "domain": key, "length": len(groupedDomainData)}).Error("Number of rows per domain exceeding 10")
		}
	}
	return groupedDomainDataSet
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
	errMsg, errCode = updateAccountLevelPropertiesForGroupUser(projectID, U.GROUP_NAME_LINKEDIN_COMPANY, domainData.Domain, userIDToUpdate, impr_diff, clicks_diff, unixTimestamp+1)
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

func checkIfEventCreationReqV2(propertyValue int64, domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) (bool, string, string) {
	if propertyValue <= 0 {
		return false, "", ""
	}
	if value, exists := existingEventsWithCampaignData[timestamp][domain][campaignGroupID]; exists {
		return false, value["id"].(string), value["user_id"].(string)
	}

	return true, "", ""
}

func getExistingPropertyValue(domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) float64 {

	return U.SafeConvertToFloat64(existingEventsWithCampaignData[timestamp][domain][campaignGroupID]["p_value"])
}

func getUserIDFromEventsForUpdatingGroupUser(currUserID string, domain string, timestamp int64, campaignGroupID string, existingEventsWithCampaignData map[int64]map[string]map[string]map[string]interface{}) string {
	if _, exists := existingEventsWithCampaignData[timestamp][domain][campaignGroupID]; !exists {
		return currUserID
	}
	return existingEventsWithCampaignData[timestamp][domain][campaignGroupID]["user_id"].(string)
}

func updateAccountLevelPropertiesForGroupUser(projectID int64, groupName string, domainName string, groupUserID string, impr_diff float64, clicks_diff float64, timestamp int64) (string, int) {
	groupID := U.GetDomainGroupDomainName(projectID, domainName)
	if groupID == "" {
		return "", http.StatusNotImplemented
	}

	groupUser, errCode := store.GetStore().GetGroupUserByGroupID(projectID, groupName, groupID)
	if errCode != http.StatusFound {
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
	_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID, groupUserID, &propertiesMap, currTimestamp, currTimestamp, source)
	if err != nil {
		return "Failed to create or update group user on updateAccountLevelPropertiesForGroupUser.", http.StatusInternalServerError
	}
	return "", http.StatusOK
}

func getUserIDsWithDataMismatch(projectID int64) (string, string, int) {
	userIDs := ""
	db := C.GetServices().Db

	group, errCode := store.GetStore().GetGroup(projectID, model.GROUP_NAME_LINKEDIN_COMPANY)
	if errCode != http.StatusFound {
		return "", "Failed to get group.", http.StatusInternalServerError
	}
	source := model.GetGroupUserSourceByGroupName(U.GROUP_NAME_LINKEDIN_COMPANY)

	imprEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get impr event name")
		return "", "Failed to get impression eventname", http.StatusInternalServerError
	}
	clickEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD, projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get clicks event name")
		return "", "Failed to get click eventname", http.StatusInternalServerError
	}

	queryStr := "With users_0 as (SELECT id, group_%d_id, is_group_user, JSON_EXTRACT_STRING(properties, '$li_total_ad_view_count') as user_impressions, " +
		"JSON_EXTRACT_STRING(properties, '$li_total_ad_click_count') as user_clicks from users where project_id = ? and source = ?), " +
		"events_0 as (SELECT user_id, sum(JSON_EXTRACT_STRING(properties, '$li_ad_view_count')) as event_impressions, " +
		"Case when sum(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) is null then 0 else " +
		"sum(JSON_EXTRACT_STRING(properties, '$li_ad_click_count')) END as event_clicks from events where project_id = ? and event_name_id in (?,?)" +
		"group by user_id order by user_id) SELECT id, group_%d_id, is_group_user from users_0 join events_0 on id=user_id " +
		"where user_impressions != event_impressions or user_clicks != event_clicks"
	queryStr = fmt.Sprintf(queryStr, group.ID, group.ID)
	var users []model.User
	err = db.Raw(queryStr, projectID, source, projectID, imprEventName.ID, clickEventName.ID).Find(&users).Error
	if err != nil {
		log.WithError(err).Error("Failed to get group users")
		return "", "Failed to find group users", http.StatusInternalServerError
	}
	userIDsArr := make([]string, 0)
	for _, user := range users {
		userIDsArr = append(userIDsArr, user.ID)
	}
	userIDs = strings.Join(userIDsArr, ",")
	return userIDs, "", http.StatusOK
}
