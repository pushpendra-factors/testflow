package task

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

func CreateGroupUserAndEvents(linkedinProjectSetting model.LinkedinProjectSettings) (string, int) {
	projectID, err := strconv.ParseInt(linkedinProjectSetting.ProjectId, 10, 64)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}
	log.Info("Starting processing for project ", projectID)

	eventNameViewedAD, eventNameClickedAD, errMsg, errCode := createDependentEventNames(projectID)
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

	distinctTimestamps, errCode := store.GetStore().GetDistinctTimestampsForEventCreation(linkedinProjectSetting.ProjectId)
	if errCode != http.StatusOK {
		return "Failed to get distinct timestamps for event creation from linkedin", errCode
	}
	for _, timestamp := range distinctTimestamps {
		domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
		if errCode != http.StatusOK {
			return "Failed to get domain data from linkedin", errCode
		}
		log.WithFields(log.Fields{"count": len(domainDataSet), "timestamp": timestamp}).Info("DebugMetric log 1")
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
			domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
			if errCode != http.StatusOK {
				return "Failed to get domain data from linkedin", errCode
			}
		}
	}
	log.Info("Ended processing for project ", projectID)
	return "", http.StatusOK
}

func createDependentEventNames(projectID int64) (*model.EventName, *model.EventName, string, int) {
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
				err = store.GetStore().UpdateLinkedinGroupUserCreationDetails(domainData)
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

		err := store.GetStore().UpdateLinkedinGroupUserCreationDetails(domainData)
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

	eventNameViewedAD, eventNameClickedAD, errMsg, errCode := createDependentEventNames(projectID)
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

	distinctTimestamps, errCode := store.GetStore().GetDistinctTimestampsForEventCreation(linkedinProjectSetting.ProjectId)
	if errCode != http.StatusOK {
		return "Failed to get distinct timestamps for event creation from linkedin", errCode
	}
	for _, timestamp := range distinctTimestamps {
		domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
		if errCode != http.StatusOK {
			return "Failed to get domain data from linkedin", errCode
		}
		log.WithFields(log.Fields{"count": len(domainDataSet), "timestamp": timestamp}).Info("DebugMetric log 1")
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
			domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinForTimestamp(linkedinProjectSetting.ProjectId, timestamp)
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
		err := store.GetStore().UpdateLinkedinGroupUserCreationDetails(domainData)
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
		err = store.GetStore().UpdateLinkedinGroupUserCreationDetails(domainData)
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
	err = store.GetStore().UpdateLinkedinGroupUserCreationDetails(domainData)
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
		log.WithFields(log.Fields{"projectID": projectID, "eventID": eventID, "timestamp": timestamp, "props": eventPropertiesMap}).Error("Failed in updating event")
		if errCode != http.StatusAccepted {
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
