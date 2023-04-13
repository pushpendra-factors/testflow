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
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
	"time"
)

const EmptyJsonStr = "{}"

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

func filterCheckGeneralV1(rule model.OTPRule, event eventIdToProperties, logCtx *log.Entry) bool {
	var ruleFilters []model.TouchPointFilter
	err := U.DecodePostgresJsonbToStructType(&rule.Filters, &ruleFilters)
	if err != nil {
		logCtx.WithFields(log.Fields{"event": event, "rule": rule}).WithError(err).Error("Failed to decode/fetch offline touch point rule FILTERS for salesforce document.")
		return false
	}

	filtersPassed := 0
	for _, filter := range ruleFilters {
		switch filter.Operator {
		case model.EqualsOpStr:
			if _, exists := event.EventProperties[filter.Property]; exists {
				if filter.Value != "" && event.EventProperties[filter.Property] == filter.Value {
					filtersPassed++
				}
			}
		case model.NotEqualOpStr:
			if _, exists := event.EventProperties[filter.Property]; exists {
				if filter.Value != "" && event.EventProperties[filter.Property] != filter.Value {
					filtersPassed++
				}
			}
		case model.ContainsOpStr:
			if _, exists := event.EventProperties[filter.Property]; exists {
				if filter.Property != "" {
					val, ok := event.EventProperties[filter.Property].(string)
					if ok && strings.Contains(val, filter.Value) {
						filtersPassed++
					}
				}
			}
		default:
			logCtx.WithField("Rule", rule).WithField("event", event).
				Error("No matching operator found for offline touch point rules for hubspot engagement document.")
			continue
		}
	}

	// return true if all the filters passed
	if filtersPassed != 0 && filtersPassed == len(ruleFilters) {
		return true
	}

	// When neither filters matched nor (filters matched but values are same)
	logCtx.WithField("Rule", rule).WithField("event", event).Warn("Filter check general is failing for offline touch point rule")
	return false
}

func PullEventIdsWithEventName(projectId int64, startTimestamp int64, endTimestamp int64, eventName string) ([]string, map[string]eventIdToProperties, error) {
	db := C.GetServices().Db

	events := make(map[string]eventIdToProperties, 0)
	eventsIds := make([]string, 0)

	rows, _ := db.Raw("SELECT events.id, events.user_id , event_names.name, events.timestamp, events.properties FROM events"+
		" "+"LEFT JOIN event_names ON event_names.id = events.event_name_id"+
		" "+"WHERE events.project_id = ? AND event_names.name = ? AND events.timestamp >= ? AND events.timestamp <= ? ORDER BY events.timestamp, events.created_at ASC", projectId, eventName, startTimestamp, endTimestamp).Rows()

	rowNum := 0
	for rows.Next() {
		var id string
		var userId string
		var name string
		var timestamp int64
		var properties *postgres.Jsonb

		if err := rows.Scan(&id, &userId, &name, &timestamp, &properties); err != nil {
			log.WithError(err).Error("Failed to scan rows")
			return nil, nil, err
		}

		var eventPropertiesBytes interface{}
		var err error
		if properties != nil {
			eventPropertiesBytes, err = properties.Value()
			if err != nil {
				log.WithError(err).Error("Failed to read event properties")
				return nil, nil, err
			}
		} else {
			eventPropertiesBytes = []byte(EmptyJsonStr)
		}

		var eventPropertiesMap map[string]interface{}
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		if err != nil {
			log.WithError(err).Error("Failed to marshal event properties")
			return nil, nil, err
		}

		eventsIds = append(eventsIds, id)

		events[id] = eventIdToProperties{
			ID:              id,
			UserId:          userId,
			ProjectId:       projectId,
			Name:            name,
			EventProperties: eventPropertiesMap,
			Timestamp:       timestamp,
		}

		rowNum++
	}

	return eventsIds, events, nil
}

func RunOTPHubspotForProjects(configs map[string]interface{}) (map[string]interface{}, bool) {

	projectIDList := configs["project_ids"].(string)
	disabledProjectIDList := configs["disabled_project_ids"].(string)
	defaultHealthcheckPingID := configs["health_check_ping_id"].(string)
	overrideHealthcheckPingID := configs["override_healthcheck_ping_id"].(string)
	numProjectRoutines := configs["num_project_routines"].(int)

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

	// Runs enrichment for list of project_ids as batch using go routines.
	batches := U.GetInt64ListAsBatch(projectIDs, numProjectRoutines)
	log.WithFields(log.Fields{"project_batches": batches}).Info("Running for batches.")
	syncStatus := SyncStatus{}
	for bi := range batches {
		batch := batches[bi]

		var wg sync.WaitGroup
		for pi := range batch {
			wg.Add(1)

			go syncWorkerForOTP(batch[pi], &wg)
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

func syncWorkerForOTP(projectID int64, wg *sync.WaitGroup) {
	defer wg.Done()

	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	logCtx.Info("Running sync for project.")

	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get otp Rules for Project")
		return
	}

	uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
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

	startTime, endTime, _ := U.GetQueryRangePresetYesterdayIn(timezoneString)

	for _, eventName := range AllowedHsEventTypeForOTP {

		switch eventName {

		case U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED:

			RunHSOfflineTouchContact(project, otpRules, uniqueOTPEventKeys, projectID, startTime, endTime, logCtx)

		case U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:

			RunHSOfflineTouchForms(project, otpRules, uniqueOTPEventKeys, projectID, startTime, endTime, logCtx)

		case U.EVENT_NAME_HUBSPOT_CONTACT_LIST:

			RunHSOfflineTouchContactList(project, otpRules, uniqueOTPEventKeys, projectID, startTime, endTime, logCtx)

		case U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL, U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
			U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED, U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
			U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:

			RunHSOfflineTouchEngagement(project, otpRules, uniqueOTPEventKeys, projectID, startTime, endTime, eventName, logCtx)

		}

	}

}

func RunHSOfflineTouchContact(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(projectID, startTime, endTime, U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		eventName := events[ei].Name

		logCtx.Info(fmt.Sprintf("event name  %s", eventName))

		errCode := ApplyHSOfflineTouchPointRuleV1(project, &otpRules, &uniqueOTPEventKeys, events[ei], events[ei].Timestamp)
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

func RunHSOfflineTouchForms(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(projectID, startTime, endTime, U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		eventName := events[ei].Name

		logCtx.Info(fmt.Sprintf("event name  %s", eventName))

		errCode := ApplyHSOfflineTouchPointRuleForFormsV1(project, &otpRules, &uniqueOTPEventKeys, events[ei], events[ei].Timestamp)
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

func RunHSOfflineTouchContactList(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(projectID, startTime, endTime, U.EVENT_NAME_HUBSPOT_CONTACT_LIST)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		eventName := events[ei].Name

		logCtx.Info(fmt.Sprintf("event name  %s", eventName))

		errCode := ApplyHSOfflineTouchPointRuleForContactListV1(project, &otpRules, &uniqueOTPEventKeys, events[ei])
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

func RunHSOfflineTouchEngagement(project *model.Project, otpRules []model.OTPRule, uniqueOTPEventKeys []string, projectID, startTime, endTime int64, eventName string, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(projectID, startTime, endTime, eventName)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		logCtx.Info(fmt.Sprintf("event name  %s", eventName))
		engagementType := events[ei].EventProperties[U.EP_HUBSPOT_ENGAGEMENT_TYPE]
		errCode := ApplyHSOfflineTouchPointRuleForEngagementV1(project, &otpRules, &uniqueOTPEventKeys, events[ei], fmt.Sprint(engagementType))
		if errCode != nil {
			log.Info("Fail to apply OTP")
		}
	}

}

type eventIdToProperties struct {
	ID                         string                 `gorm:"primary_key:true;type:uuid" json:"id"`
	ProjectId                  int64                  `gorm:"primary_key:true;" json:"project_id"`
	UserId                     string                 `json:"user_id"`
	Name                       string                 `json:"name"`
	PropertiesUpdatedTimestamp int64                  `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
	EventProperties            map[string]interface{} `json:"event_properties"`
	UserProperties             map[string]interface{} `json:"user_properties"`
	Timestamp                  int64                  `json:"timestamp"`
}

//ApplyHSOfflineTouchPointRuleV1 - Check if the condition are satisfied for creating OTP events for each rule for HS Contact
func ApplyHSOfflineTouchPointRuleV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event eventIdToProperties, eventTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRule",
		"document_action": fmt.Sprint("Contact")})

	if otpRules == nil || project == nil || &event == nil {
		return nil
	}

	eventTimestamp = U.CheckAndGetStandardTimestamp(eventTimestamp)

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForFormsAndContactsV1(rule, event)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable
		if !filterCheckGeneralV1(rule, event, logCtx) {
			continue
		}

		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
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
func CreateTouchPointEventForFormsAndContactsV1(project *model.Project, event eventIdToProperties,
	rule model.OTPRule, eventTimestamp int64, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent"})
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

//Creates a unique key using ruleID, userID and eventID as keyID for Forms and contacts
func createOTPUniqueKeyForFormsAndContactsV1(rule model.OTPRule, event eventIdToProperties) (string, int) {

	ruleID := rule.ID
	userID := event.UserId
	keyID := event.ID

	uniqueKey := userID + ruleID + keyID

	return uniqueKey, http.StatusCreated

}

//ApplyHSOfflineTouchPointRuleForEngagementV1 -Check if the condition are satisfied for creating OTP events for each rule for HS Engagements - Meetings/Calls/Emails
func ApplyHSOfflineTouchPointRuleForEngagementV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event eventIdToProperties, engagementType string) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForEngagement", "event": event})

	if otpRules == nil || project == nil {
		logCtx.Error("something is empty")
		return nil
	}
	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForEngagementsV1(rule, event, engagementType, logCtx)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		if !filterCheckGeneralV1(rule, event, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}

		_, err1 := CreateTouchPointEventForEngagementV1(project, event, rule, engagementType, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue

		}
	}
	return nil
}

//Creates a unique key using ruleID, userID and engagementID as keyID for Engagements {Emails, Calls and Meetings}
func createOTPUniqueKeyForEngagementsV1(rule model.OTPRule, event eventIdToProperties, engagementType string, logCtx *log.Entry) (string, int) {

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
func CreateTouchPointEventForEngagementV1(project *model.Project, event eventIdToProperties,
	rule model.OTPRule, engagementType string, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent",
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

//isEmailEngagementAlreadyTrackedV1 - Checks if the Email (a type of Engagement) is already tracked for creating OTP event.
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

//ApplyHSOfflineTouchPointRuleForContactListV1 - Check if the condition are satisfied for creating OTP events for each rule for HS Contact list
func ApplyHSOfflineTouchPointRuleForContactListV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event eventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForContactList", "event": event})

	if otpRules == nil || project == nil {
		logCtx.Error("something is empty")
		return nil
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
		if !filterCheckGeneralV1(rule, event, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}
		_, err1 := CreateTouchPointEventForListsV1(project, event, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot lists.")
			continue
		}

	}
	return nil
}

func createOTPUniqueKeyForContactListV1(rule model.OTPRule, event eventIdToProperties, logCtx *log.Entry) (string, int) {
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
func CreateTouchPointEventForListsV1(project *model.Project, event eventIdToProperties,
	rule model.OTPRule, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEventForLists"})

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

//ApplyHSOfflineTouchPointRuleForFormsV1 - Check if the condition are satisfied for creating OTP events for each rule for HS Forms Submission
func ApplyHSOfflineTouchPointRuleForFormsV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event eventIdToProperties, formTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForForms", "event": event})

	if otpRules == nil || project == nil {
		logCtx.Error("something is empty")
		return nil
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
		if !filterCheckGeneralV1(rule, event, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
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
