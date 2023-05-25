package hubspot_enrich

import (
	"encoding/json"
	C "factors/config"
	IntHubspot "factors/integration/hubspot"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var AllowedHsEventTypeForOTP = []string{

	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION,
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
	U.EVENT_NAME_HUBSPOT_CONTACT_LIST,
}

func RunOTPHubspotForProjects(configs map[string]interface{}) (map[string]interface{}, bool) {

	projectIDList := configs["project_ids"].(string)
	disabledProjectIDList := configs["disabled_project_ids"].(string)
	defaultHealthcheckPingID := configs["health_check_ping_id"].(string)
	overrideHealthcheckPingID := configs["override_healthcheck_ping_id"].(string)
	numProjectRoutines := configs["num_project_routines"].(int)
	numDaysBackfill := configs["num_days_backfill"].(int)
	backfillStartTime := configs["backfill_start_timestamp"].(int)
	backfillEndTime := configs["backfill_end_timestamp"].(int)

	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, overrideHealthcheckPingID)

	hubspotEnabledProjectSettings, errCode := store.GetStore().GetAllHubspotProjectSettings()
	if errCode != http.StatusFound {
		log.Panic("No projects enabled hubspot integration.")
	}

	var propertyDetailSyncStatus []IntHubspot.Status
	anyFailure := false
	panicError := true
	jobStatus := make(map[string]interface{})
	defer func() {
		if panicError || anyFailure {
			C.PingHealthcheckForFailure(healthcheckPingID, jobStatus)
		} else {
			C.PingHealthcheckForSuccess(healthcheckPingID, jobStatus)
		}
	}()

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		projectIDList, disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	if len(disabledProjects) > 0 {
		log.WithField("excluded_projects", disabledProjectIDList).Info("Running with exclusion of projects.")
	}

	projectIDs := make([]int64, 0, 0)

	for _, settings := range hubspotEnabledProjectSettings {
		if exists := disabledProjects[settings.ProjectId]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[settings.ProjectId]; !exists {
				continue
			}
		}

		projectIDs = append(projectIDs, settings.ProjectId)

		log.WithFields(log.Fields{"projects": projectIDs}).Info("all project list")

	}

	endTime := U.TimeNowUnix()
	startTime := endTime - (int64(numDaysBackfill) * model.SecsInADay)

	//if backfill startTime and endTime is assigned

	if backfillStartTime <= backfillEndTime {

		startTime = int64(backfillStartTime)
		endTime = int64(backfillEndTime)
		if backfillEndTime == 0 {
			endTime = U.TimeNowUnix()
		}
	}

	backfillEnabled := backfillStartTime != 0

	// Runs enrichment for list of project_ids as batch using go routines.
	batches := U.GetInt64ListAsBatch(projectIDs, numProjectRoutines)
	log.WithFields(log.Fields{"project_batches": batches}).Info("Running for batches.")
	syncStatus := SyncStatus{}
	for bi := range batches {
		batch := batches[bi]

		var wg sync.WaitGroup
		for pi := range batch {
			wg.Add(1)

			go syncWorkerForOTP(batch[pi], startTime, endTime, backfillEnabled, &wg)
		}
		wg.Wait()
	}

	anyFailure = anyFailure || syncStatus.HasFailure

	jobStatus = map[string]interface{}{
		"document_sync":      syncStatus.Status,
		"property_type_sync": propertyDetailSyncStatus,
	}
	panicError = false

	return jobStatus, true

}

func syncWorkerForOTP(projectID int64, startTime, endTime int64, backfillEnabled bool, wg *sync.WaitGroup) {
	defer wg.Done()

	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	logCtx.Info("Running sync for project.")

	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get otp Rules for Project")
		return
	}

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project")
		return
	}

	timezoneString, statusCode := store.GetStore().GetTimezoneForProject(projectID)
	if statusCode != http.StatusFound {
		logCtx.Error("query failed. Failed to get Timezone from project")
		return
	}

	OtpEventName, _ := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)

	if !backfillEnabled {
		_startTime, errCode := store.GetStore().GetLatestEventTimeStampByEventNameId(project.ID, OtpEventName.ID, startTime, endTime)
		if errCode == http.StatusFound {
			startTime = _startTime
		}
	}

	logCtx.WithFields(log.Fields{"startTime": startTime, "endTime": endTime}).Info("starting otp creation job")
	//batch time range day-wise

	daysTimeRange, _ := U.GetAllDaysAsTimestamp(startTime, endTime, string(timezoneString))

	for _, timeRange := range daysTimeRange {

		uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
		if errCode != http.StatusFound && errCode != http.StatusNotFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
			return
		}

		for _, eventName := range AllowedHsEventTypeForOTP {

			logCtx.Info(fmt.Sprintf("event name  %s", eventName))

			eventDetails, _ := store.GetStore().GetEventNameIDFromEventName(eventName, project.ID)

			switch eventName {

			case U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED:

				RunHSOfflineTouchContact(project, otpRules, uniqueOTPEventKeys, projectID, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)

			case U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:

				RunHSOfflineTouchForms(project, otpRules, uniqueOTPEventKeys, projectID, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)

			case U.EVENT_NAME_HUBSPOT_CONTACT_LIST:

				RunHSOfflineTouchContactList(project, otpRules, uniqueOTPEventKeys, projectID, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)

			case U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL, U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
				U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED, U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
				U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:

				RunHSOfflineTouchEngagement(project, otpRules, uniqueOTPEventKeys, projectID, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)

			}

		}
	}
}

func RunHSOfflineTouchContact(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(projectID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		errCode := ApplyHSOfflineTouchPointRuleV1(project, &otpRules, &uniqueOTPEventKeys, events[ei], events[ei].Timestamp)
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

func RunHSOfflineTouchForms(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(projectID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		errCode := ApplyHSOfflineTouchPointRuleForFormsV1(project, &otpRules, &uniqueOTPEventKeys, events[ei], events[ei].Timestamp)
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

func RunHSOfflineTouchContactList(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(projectID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		errCode := ApplyHSOfflineTouchPointRuleForContactListV1(project, &otpRules, &uniqueOTPEventKeys, events[ei])
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

func RunHSOfflineTouchEngagement(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(projectID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		engagementType := events[ei].EventProperties[U.EP_HUBSPOT_ENGAGEMENT_TYPE]
		errCode := ApplyHSOfflineTouchPointRuleForEngagementV1(project, &otpRules, &uniqueOTPEventKeys, events[ei], fmt.Sprint(engagementType))
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

// ApplyHSOfflineTouchPointRuleV1 - Check if the condition are satisfied for creating OTP events for each rule for HS Contact
func ApplyHSOfflineTouchPointRuleV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event model.EventIdToProperties, eventTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRule",
		"document_action": fmt.Sprint("Contact")})

	if otpRules == nil || project == nil || &event == nil {
		return nil
	}

	eventTimestamp = U.CheckAndGetStandardTimestamp(eventTimestamp)

	otpEventName, err := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)
	if err != nil {
		return err
	}

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForFormsAndContactsV1(rule, event)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable
		if !model.EvaluateOTPFilterV1(rule, event, logCtx) {
			continue
		}

		if C.GetOtpKeyWithQueryCheckEnabled() {

			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(project.ID, event.UserId, otpEventName.ID, otpUniqueKey)
			if !isUnique {
				continue
			}

		} else {
			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
				continue
			}

		}

		_, err1 := CreateTouchPointEventForFormsAndContactsV1(project, event, rule, eventTimestamp, otpUniqueKey)
		if err1 != nil {
			logCtx.Warn("failed to create touch point for hubspot contact updated document.")
			continue
		}

		*uniqueOTPEventKeys = append(*uniqueOTPEventKeys, otpUniqueKey)

	}
	return nil
}

// CreateTouchPointEventForFormsAndContactsV1 - Creates offline touch-point for HS create/update events with given rule for HS Contacts and Forms
func CreateTouchPointEventForFormsAndContactsV1(project *model.Project, event model.EventIdToProperties,
	rule model.OTPRule, eventTimestamp int64, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEventForFormsAndContactsV1"})
	logCtx.WithField("event", event).Info("CreateTouchPointEvent: creating hubspot offline touch point document")

	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          event.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   model.UserSourceHubspot,
	}

	var timestamp int64
	if rule.TouchPointTimeRef == model.LastModifiedTimeRef {
		timestamp = eventTimestamp
	} else {
		timeValue, exists := (event.EventProperties)[rule.TouchPointTimeRef]
		if !exists {
			logCtx.Error("couldn't get the timestamp on hubspot contact properties using "+
				"given rule.TouchPointTimeRef-", rule.TouchPointTimeRef)
			return nil, errors.New(fmt.Sprintf("couldn't get the timestamp on hubspot "+
				"contact properties using given rule.TouchPointTimeRef - %s", rule.TouchPointTimeRef))
		}
		val, ok := timeValue.(int64)
		if !ok {
			logCtx.Error("couldn't convert the timestamp on hubspot contact properties. "+
				"using eventTimestamp instead, val", rule.TouchPointTimeRef, timeValue)
			timestamp = eventTimestamp
		} else {
			timestamp = val
		}
	}

	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID
	payload.EventProperties[U.EP_OTP_UNIQUE_KEY] = otpUniqueKey
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("Document", event).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for hubspot document.")
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event track failed for doc type, message %s", trackResponse.Error)) // add proper comment
	}

	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := event.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = event.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceHubspot, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", event).WithError(err).Error(fmt.Errorf("create "+
			"hubspot touchpoint event track failed for doc type message %s", trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event "+
			"track failed for doc type, message %s", trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created hubspot offline touch point")
	return trackResponse, nil
}

// Creates a unique key using ruleID, userID and eventID as keyID for Forms and contacts
func createOTPUniqueKeyForFormsAndContactsV1(rule model.OTPRule, event model.EventIdToProperties) (string, int) {

	ruleID := rule.ID
	userID := event.UserId
	keyID := event.ID

	uniqueKey := userID + ruleID + keyID

	return uniqueKey, http.StatusCreated

}

// ApplyHSOfflineTouchPointRuleForEngagementV1 -Check if the condition are satisfied for creating OTP events for each rule for HS Engagements - Meetings/Calls/Emails
func ApplyHSOfflineTouchPointRuleForEngagementV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event model.EventIdToProperties, engagementType string) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForEngagement", "event": event})

	if otpRules == nil || project == nil {
		logCtx.Error("something is empty")
		return nil
	}
	otpEventName, err := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)
	if err != nil {
		return err
	}

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForEngagementsV1(rule, event, engagementType, logCtx)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		if !model.EvaluateOTPFilterV1(rule, event, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if C.GetOtpKeyWithQueryCheckEnabled() {

			isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(project.ID, event.UserId, otpEventName.ID, otpUniqueKey)
			if !isUnique {
				continue
			}

		} else {

			if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
				continue
			}

		}

		_, err1 := CreateTouchPointEventForEngagementV1(project, event, rule, engagementType, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue

		}
	}
	return nil
}

// Creates a unique key using ruleID, userID and engagementID as keyID for Engagements {Emails, Calls and Meetings}
func createOTPUniqueKeyForEngagementsV1(rule model.OTPRule, event model.EventIdToProperties, engagementType string, logCtx *log.Entry) (string, int) {

	ruleID := rule.ID
	userID := event.UserId
	var keyID string
	var uniqueKey string

	switch engagementType {

	case IntHubspot.EngagementTypeEmail, IntHubspot.EngagementTypeIncomingEmail, IntHubspot.EngagementTypeMeeting, IntHubspot.EngagementTypeCall:
		if _, exists := event.EventProperties[U.EP_HUBSPOT_ENGAGEMENT_ID]; exists {
			keyID = fmt.Sprintf("%v", event.EventProperties[U.EP_HUBSPOT_ENGAGEMENT_ID])
		} else {
			logCtx.Error("Event Property $hubspot_engagement_id does not exist.")
			return uniqueKey, http.StatusNotFound
		}

	default:
		logCtx.Error("engagement type not supported yet for otp_unique_key creation")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated
}

// CreateTouchPointEventForEngagementV1 - Creates offline touch-point for HS engagements (calls, meetings, forms, emails) with give rule
func CreateTouchPointEventForEngagementV1(project *model.Project, event model.EventIdToProperties,
	rule model.OTPRule, engagementType string, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEventForEngagementV1",
		"event": event, "rule": rule})

	logCtx.WithField("rule", rule).WithField("event", event).
		Info("CreateTouchPointEvent: creating hubspot offline touch point document")
	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)

	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          event.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   model.UserSourceHubspot,
	}

	var timestamp int64

	switch engagementType {

	case IntHubspot.EngagementTypeEmail, IntHubspot.EngagementTypeIncomingEmail, IntHubspot.EngagementTypeMeeting, IntHubspot.EngagementTypeCall:
		{
			threadID, isPresent := event.EventProperties["threadId"]
			if !isPresent {
				logCtx.WithField("threadID", threadID).
					Error("couldn't get the threadID on hubspot email engagement, logging and continuing")
			} else {
				found, errT := isEmailEngagementAlreadyTrackedV1(project.ID, rule.ID, threadID.(string), logCtx)
				if found || errT != nil {
					return trackResponse, errT
				}
			}

			payload.EventProperties[U.EP_HUBSPOT_ENGAGEMENT_THREAD_ID] = threadID

			timeValue, exists := (event.EventProperties)[rule.TouchPointTimeRef]
			if !exists {
				logCtx.WithField("TouchPointTimeRef", rule.TouchPointTimeRef).
					Error("couldn't get the timestamp on hubspot contact properties using given rule.TouchPointTimeRef")
				return nil, errors.New(fmt.Sprintf("couldn't get the timestamp on hubspot contact properties "+
					"using given rule.TouchPointTimeRef - %s", rule.TouchPointTimeRef))
			}
			val, ok := timeValue.(int64)
			if !ok {
				logCtx.WithField("TouchPointTimeRef", rule.TouchPointTimeRef).WithField("TimeValue", timeValue).
					Error("couldn't convert the timestamp on hubspot contact properties. using trackPayload timestamp instead, val")
				timestamp = event.Timestamp
			} else {
				timestamp = val
			}
		}

	default:
		logCtx.Error("engagement type not supported yet for rule creation")

	}

	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID
	payload.EventProperties[U.EP_OTP_UNIQUE_KEY] = otpUniqueKey
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("event", event).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for hubspot document.")
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event track failed for doc type , message %s", trackResponse.Error))
	}
	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := event.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = event.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceHubspot, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("event", event).WithError(err).Error(fmt.Errorf("create "+
			"hubspot engagement touchpoint event track failed for doc type , message %s", trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create hubspot engagement touchpoint event "+
			"track failed for doc type , message %s", trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created hubspot engagement offline touch point")
	return trackResponse, nil
}

// isEmailEngagementAlreadyTrackedV1 - Checks if the Email (a type of Engagement) is already tracked for creating OTP event.
func isEmailEngagementAlreadyTrackedV1(projectID int64, ruleID string, threadID string, logCtx *log.Entry) (bool, error) {

	en, status := store.GetStore().CreateOrGetOfflineTouchPointEventName(projectID)
	if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
		logCtx.Error("failed to create event name on SF for offline touch point")
		return false, errors.New(fmt.Sprintf("failed to create event name on SF for offline touch point"))
	}

	last30DaysOTPEvents, err := store.GetStore().GetEventsByEventNameId(projectID, en.ID, time.Now().Unix()-U.MonthInSecs, time.Now().Unix())
	if err != http.StatusFound && err != http.StatusNotFound {
		logCtx.Info("no events found for engagement, continuing")
	} else {

		for _, event := range last30DaysOTPEvents {
			propertiesMap := make(map[string]interface{})
			err := json.Unmarshal(event.Properties.RawMessage, &propertiesMap)
			if err != nil {
				log.Error("Error occurred during unmarshal of otp event properties, continuing")
				continue
			}
			threadIDTracked, threadExists := propertiesMap[U.EP_HUBSPOT_ENGAGEMENT_THREAD_ID]
			if threadExists && threadID == threadIDTracked {
				ruleIDTracked, ruleExists := propertiesMap[U.EP_OTP_RULE_ID]
				if ruleExists && ruleIDTracked == ruleID {
					logCtx.Info("OTP already created for the rule and email engagement, skipping this thread")
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// ApplyHSOfflineTouchPointRuleForContactListV1 - Check if the condition are satisfied for creating OTP events for each rule for HS Contact list
func ApplyHSOfflineTouchPointRuleForContactListV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event model.EventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForContactList", "event": event})

	if otpRules == nil || project == nil {
		logCtx.Error("something is empty")
		return nil
	}

	otpEventName, err := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)
	if err != nil {
		return err
	}

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForContactListV1(rule, event, logCtx)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable & the record has changed property w.r.t filters
		if rule.RuleType != model.TouchPointRuleTypeContactList {
			logCtx.Info("Rule Type is failing the OTP event creation.")
			continue
		}
		if !model.EvaluateOTPFilterV1(rule, event, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if C.GetOtpKeyWithQueryCheckEnabled() {

			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(project.ID, event.UserId, otpEventName.ID, otpUniqueKey)
			if !isUnique {
				continue
			}

		} else {
			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
				continue
			}

		}

		_, err1 := CreateTouchPointEventForListsV1(project, event, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot lists.")
			continue
		}

	}
	return nil
}

func createOTPUniqueKeyForContactListV1(rule model.OTPRule, event model.EventIdToProperties, logCtx *log.Entry) (string, int) {
	ruleID := rule.ID
	userID := event.UserId
	var keyID string
	var uniqueKey string

	if _, exists := event.EventProperties[U.EP_HUBSPOT_CONTACT_LIST_LIST_ID]; exists {
		keyID = fmt.Sprintf("%v", event.EventProperties[U.EP_HUBSPOT_CONTACT_LIST_LIST_ID])
	} else {
		logCtx.Error("Event Property $hubspot_contact_list_list_id does not exist.")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated

}

// CreateTouchPointEventForListsV1 - Creates OTP for HS lists
func CreateTouchPointEventForListsV1(project *model.Project, event model.EventIdToProperties,
	rule model.OTPRule, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEventForListsV1"})
	logCtx.WithField("rule", rule).WithField("event", event).
		Info("CreateTouchPointEventForLists: creating hubspot offline touch point document")

	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          event.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   model.UserSourceHubspot,
	}

	var timestamp int64
	timeValue, exists := (event.EventProperties)[rule.TouchPointTimeRef]
	if !exists {
		logCtx.WithField("TouchPointTimeRef", rule.TouchPointTimeRef).
			Error("couldn't get the timestamp on hubspot contact properties using given rule.TouchPointTimeRef")
		return nil, errors.New(fmt.Sprintf("couldn't get the timestamp on hubspot contact properties "+
			"using given rule.TouchPointTimeRef - %s", rule.TouchPointTimeRef))
	}
	val, ok := timeValue.(int64)
	if !ok {
		logCtx.WithField("TouchPointTimeRef", rule.TouchPointTimeRef).WithField("TimeValue", timeValue).
			Error("couldn't convert the timestamp on hubspot contact properties. using trackPayload timestamp instead, val")
		timestamp = event.Timestamp
	} else {
		timestamp = val
	}

	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID
	payload.EventProperties[U.EP_OTP_UNIQUE_KEY] = otpUniqueKey
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("event", event).WithField("rule", rule).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for hubspot otp creation.")
		return trackResponse, errors.New(fmt.Sprintf("create hubspot list touchpoint event track failed for doc type , message %s", trackResponse.Error))
	}

	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := event.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = event.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceHubspot, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("event", event).WithError(err).Error(fmt.Errorf("create "+
			"hubspot list touchpoint event track failed for doc type , message %s", trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event "+
			"track failed for doc type , message %s", trackResponse.Error))
	}

	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created hubspot lists offline touch point")
	return trackResponse, nil
}

// ApplyHSOfflineTouchPointRuleForFormsV1 - Check if the condition are satisfied for creating OTP events for each rule for HS Forms Submission
func ApplyHSOfflineTouchPointRuleForFormsV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event model.EventIdToProperties, formTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForForms", "event": event})

	if otpRules == nil || project == nil {
		logCtx.Error("something is empty")
		return nil
	}

	otpEventName, err := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)
	if err != nil {
		return err
	}

	logCtx.WithFields(log.Fields{"ProjectID": project.ID}).Info("Inside method ApplyHSOfflineTouchPointRuleForForms")
	formTimestamp = U.CheckAndGetStandardTimestamp(formTimestamp)

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForFormsAndContactsV1(rule, event)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable & the record has changed property w.r.t filters
		if rule.RuleType != model.TouchPointRuleTypeForms {
			logCtx.Info("Rule Type is failing the OTP event creation.")
			continue
		}
		if !model.EvaluateOTPFilterV1(rule, event, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if C.GetOtpKeyWithQueryCheckEnabled() {

			isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(project.ID, event.UserId, otpEventName.ID, otpUniqueKey)
			if !isUnique {
				continue
			}

		} else {

			if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
				continue
			}

		}

		logCtx.WithFields(log.Fields{"ProjectID": project.ID, "rule": rule}).Info("Invoking method CreateTouchPointEvent")
		_, err1 := CreateTouchPointEventForFormsAndContactsV1(project, event, rule, formTimestamp, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue
		}

	}
	return nil
}
