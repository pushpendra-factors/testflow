package salesforce

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var AllowedSfEventTypeForOTP = []string{
	U.EVENT_NAME_SALESFORCE_TASK_CREATED,
	U.EVENT_NAME_SALESFORCE_TASK_UPDATED,
	U.EVENT_NAME_SALESFORCE_EVENT_CREATED,
	U.EVENT_NAME_SALESFORCE_EVENT_UPDATED,
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
	U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
}

// WorkerForSfOtp sync salesforce Events to otp events
func WorkerForSfOtp(projectID, startTime, endTime int64, wg *sync.WaitGroup) {

	defer wg.Done()

	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	if projectID == 0 {
		log.Info("no project found")
		return
	}

	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get otp Rules for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get OTP rules"})
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

	_startTime, errCode := store.GetStore().GetLatestEventTimeStampByEventNameId(project.ID, OtpEventName.ID, startTime, endTime)

	if errCode == http.StatusFound {
		startTime = _startTime
	}

	//batch time range day-wise

	daysTimeRange, _ := U.GetAllDaysAsTimestamp(startTime, endTime, string(timezoneString))

	for _, timeRange := range daysTimeRange {

		uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
		if errCode != http.StatusFound && errCode != http.StatusNotFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
			statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
				Status: "Failed to get OTP Unique Keys"})
			return
		}

		for _, eventName := range AllowedSfEventTypeForOTP {

			logCtx.Info(fmt.Sprintf("event name  %s", eventName))

			eventDetails, _ := store.GetStore().GetEventNameIDFromEventName(eventName, project.ID)

			switch eventName {

			case U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED:
				RunSFOfflineTouchPointRuleForCampaignMember(project, &otpRules, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)

			case U.EVENT_NAME_SALESFORCE_TASK_UPDATED, U.EVENT_NAME_SALESFORCE_TASK_CREATED:
				RunSFOfflineTouchPointRuleForTasks(project, &otpRules, &uniqueOTPEventKeys, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)

			case U.EVENT_NAME_SALESFORCE_EVENT_CREATED, U.EVENT_NAME_SALESFORCE_EVENT_UPDATED:
				RunSFOfflineTouchPointRuleForEvents(project, &otpRules, &uniqueOTPEventKeys, timeRange.Unix(), timeRange.Unix()+model.SecsInADay-1, eventDetails.ID, logCtx)
			default:
				continue

			}
		}
	}

}

func RunSFOfflineTouchPointRuleForCampaignMember(project *model.Project, otpRules *[]model.OTPRule, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(project.ID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}
	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		err := ApplySFOfflineTouchPointRuleForCampaignMemberV1(project, otpRules, events[ei])
		if err != nil {
			logCtx.WithField("event", events[ei]).Info("Fail to apply OTP")
			return

		}
	}

}

func RunSFOfflineTouchPointRuleForTasks(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(project.ID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}
	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		err := ApplySFOfflineTouchPointRuleForTasksV1(project, otpRules, uniqueOTPEventKeys, events[ei])
		if err != nil {
			logCtx.WithField("event", events[ei]).Info("Fail to apply OTP")
			return
		}
	}

}

func RunSFOfflineTouchPointRuleForEvents(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, startTime, endTime int64, eventNameId string, logCtx *log.Entry) {

	eventsIds, events, err := store.GetStore().PullEventIdsWithEventNameId(project.ID, startTime, endTime, eventNameId)
	if err != nil {
		logCtx.Warn("Failed to get events")
		return
	}
	if len(eventsIds) == 0 {
		logCtx.Warn("no event found")
		return
	}

	for ei := range events {

		err := ApplySFOfflineTouchPointRuleForEventsV1(project, otpRules, uniqueOTPEventKeys, events[ei])
		if err != nil {
			logCtx.WithField("event", events[ei]).Info("Fail to apply OTP")
			return
		}
	}

}

// Creates a unique key using ruleID, userID and salesforce task activity ID  as keyID for Salesforce Tasks.
func createOTPUniqueKeyForTasksV1(rule model.OTPRule, sfEvent model.EventIdToProperties, logCtx *log.Entry) (string, int) {

	ruleID := rule.ID
	userID := sfEvent.UserId
	var keyID string
	var uniqueKey string

	if _, exists := sfEvent.EventProperties[U.EP_SF_TASK_ID]; exists {
		keyID = fmt.Sprintf("%v", sfEvent.EventProperties[U.EP_SF_TASK_ID])
	} else {
		logCtx.Error("Event Property $salesforce_task_id does not exist.")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated
}

func filterCheckGeneralV1(rule model.OTPRule, event model.EventIdToProperties, logCtx *log.Entry) bool {
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

// CreateTouchPointEventForTasksAndEventsV1 - Creates offline touch-point for SF update events with given rule for SF Tasks/Events
func CreateTouchPointEventForTasksAndEventsV1(project *model.Project, sfEvent model.EventIdToProperties,
	rule model.OTPRule, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent"})
	logCtx.WithField("response", sfEvent).Info("CreateTouchPointEventForTasksAndEvents: creating salesforce OFFLINE TOUCH POINT document")
	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          sfEvent.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   model.UserSourceSalesforce,
	}

	var timestamp int64
	timeValue, exists := (sfEvent.EventProperties)[rule.TouchPointTimeRef]
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
		timestamp = sfEvent.Timestamp
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
		logCtx.WithField("Document", sfEvent).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for salesforce document.")
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed in method CreateTouchPointEventForTasksAndEvents for message %s", trackResponse.Error))
	}

	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := sfEvent.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = sfEvent.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceSalesforce, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", sfEvent).WithError(err).Error(fmt.Errorf("create salesforce OTP event track failed for message %s", trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed in method CreateTouchPointEventForTasksAndEvents for message %s", trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created salesforce offline touch point")
	return trackResponse, nil

}

// ApplySFOfflineTouchPointRuleForTasksV1 Check if the condition are satisfied for creating OTP events for each rule for SF Tasks Updated.
func ApplySFOfflineTouchPointRuleForTasksV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, sfEvent model.EventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRuleForTasks"})

	if otpRules == nil || project == nil {
		return nil
	}

	otpEventName, err := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)
	if err != nil {
		return err
	}

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForTasksV1(rule, sfEvent, logCtx)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		//Check if rule type is sf_tasks
		if rule.RuleType != model.TouchPointRuleTypeTasks {
			logCtx.Info("Rule Type is failing the OTP event creation for SF Tasks.")
			continue
		}

		// check if rule is applicable w.r.t filters
		if !filterCheckGeneralV1(rule, sfEvent, logCtx) {
			logCtx.Error("Filter check is failing for offline touch point rule for SF Tasks")
			continue
		}

		if C.GetOtpKeyWithQueryCheckEnabled() {

			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(project.ID, sfEvent.UserId, otpEventName.ID, otpUniqueKey)
			if !isUnique {
				continue
			}

		} else {
			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			if !isSalesforceOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
				continue
			}

		}

		_, err1 := CreateTouchPointEventForTasksAndEventsV1(project, sfEvent, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for salesforce tasks.")
			continue
		}

		*uniqueOTPEventKeys = append(*uniqueOTPEventKeys, otpUniqueKey)

	}
	return nil
}

// Creates a unique key using ruleID, userID and salesforce Event activity ID  as keyID for Salesforce Tasks.
func createOTPUniqueKeyForEventsV1(rule model.OTPRule, sfEvent model.EventIdToProperties, logCtx *log.Entry) (string, int) {

	ruleID := rule.ID
	userID := sfEvent.UserId
	var keyID string
	var uniqueKey string

	if _, exists := sfEvent.EventProperties[U.EP_SF_EVENT_ID]; exists {
		keyID = fmt.Sprintf("%v", sfEvent.EventProperties[U.EP_SF_EVENT_ID])
	} else {
		logCtx.Error("Event Property $salesforce_event_id does not exist.")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated
}

// ApplySFOfflineTouchPointRuleForEventsV1 Check if the condition are satisfied for creating OTP events for each rule for SF Event Updated.
func ApplySFOfflineTouchPointRuleForEventsV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, sfEvent model.EventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRuleForEvents"})

	if otpRules == nil || project == nil {
		return nil
	}

	otpEventName, err := store.GetStore().GetEventNameIDFromEventName(U.EVENT_NAME_OFFLINE_TOUCH_POINT, project.ID)
	if err != nil {
		return err
	}

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForEventsV1(rule, sfEvent, logCtx)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		//Check if rule type is sf_events
		if rule.RuleType != model.TouchPointRuleTypeEvents {
			logCtx.Info("Rule Type is failing the OTP event creation for SF Events.")
			continue
		}

		// check if rule is applicable w.r.t filters
		if !filterCheckGeneralV1(rule, sfEvent, logCtx) {
			logCtx.Error("Filter check is failing for offline touch point rule for SF Events")
			continue
		}

		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if C.GetOtpKeyWithQueryCheckEnabled() {

			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			isUnique, _ := store.GetStore().IsOTPKeyUniqueWithQuery(project.ID, sfEvent.UserId, otpEventName.ID, otpUniqueKey)
			if !isUnique {
				continue
			}

		} else {
			//Checks if the otpUniqueKey is already present in other OTP Event Properties
			if !isSalesforceOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
				continue
			}

		}

		_, err1 := CreateTouchPointEventForTasksAndEventsV1(project, sfEvent, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for salesforce events.")
			continue
		}

	}
	return nil
}

// ApplySFOfflineTouchPointRuleForCampaignMemberV1 Check if the condition are satisfied for creating OTP events for each rule for SF Campaign.
func ApplySFOfflineTouchPointRuleForCampaignMemberV1(project *model.Project, otpRules *[]model.OTPRule, sfEvent model.EventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRule", "event": sfEvent})

	if otpRules == nil || project == nil {
		return nil
	}

	fistRespondedRuleApplicable := true
	// Checking if the EP_SFCampaignMemberResponded has already been set as true for same customer id
	if sfEvent.Name == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED {

		// ignore to create a new touch point if last updated doc has EP_SFCampaignMemberResponded=true
		if val, exists := sfEvent.EventProperties[model.EP_SFCampaignMemberResponded]; exists {
			if val != nil && val.(bool) == true {
				logCtx.Info("found EP_SFCampaignMemberResponded=true for the document, skipping creating OTP.")
				fistRespondedRuleApplicable = false
			}
		}
	}

	for _, rule := range *otpRules {

		// check if rule is applicable
		if !filterCheckGeneralV1(rule, sfEvent, logCtx) {
			continue
		}

		// Run for create document rule
		if rule.TouchPointTimeRef == model.SFCampaignMemberCreated && sfEvent.Name == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED {
			_, err := CreateTouchPointEventCampaignMemberV1(project, sfEvent, rule)
			if err != nil {
				logCtx.WithError(err).Error("failed to create touch point for salesforce campaign member document. trying for responded rule")
			}
		}

		// Run for only first responded rules & documents where first responded is not set.
		if rule.TouchPointTimeRef == model.SFCampaignMemberResponded && fistRespondedRuleApplicable {

			logCtx.Info("Found existing salesforce campaign member document")
			if val, exists := sfEvent.EventProperties[model.EP_SFCampaignMemberResponded]; exists {
				if val.(bool) == true {
					_, err := CreateTouchPointEventCampaignMemberV1(project, sfEvent, rule)
					if err != nil {
						logCtx.WithError(err).Error("failed to create touch point for salesforce campaign member document.")
					}
				}
			}
		}
	}
	return nil
}

// CreateTouchPointEventCampaignMemberV1 - Creates offline touch point event for SF Campaign
func CreateTouchPointEventCampaignMemberV1(project *model.Project, sfEvent model.EventIdToProperties, rule model.OTPRule) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent", "rule": rule, "sfEvent": sfEvent})
	logCtx.WithField("sfEvent", sfEvent).Info("CreateTouchPointEvent: creating salesforce document")
	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          sfEvent.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   model.UserSourceSalesforce,
	}

	var timestamp int64

	timestamp, err = getSalesforceDocumentTimestampByEventV1(sfEvent)
	if err != nil {
		logCtx.WithError(err).Error("failed to timestamp for SF for offline touch point.")
		return trackResponse, err
	}
	payload.Timestamp = timestamp

	if rule.TouchPointTimeRef == model.SFCampaignMemberResponded {
		if val, exists := sfEvent.EventProperties[model.EP_SFCampaignMemberFirstRespondedDate]; exists {
			if tt, err := U.GetPropertyValueAsInt64(val); err != nil {
				payload.Timestamp = tt
			} else {
				logCtx.WithError(err).Error("failed to set timestamp for SF for offline touch point - First responded time.")
			}
		}
	}

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("event", sfEvent).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for salesforce document.")
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed for doc type , message %s", trackResponse.Error))
	}

	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := sfEvent.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = sfEvent.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}
	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceSalesforce, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("event", sfEvent).WithError(err).Error(fmt.Errorf("create salesforce touchpoint event track failed for doc type , message %s", trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed for doc type , message %s", trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created salesforce offline touch point")
	return trackResponse, nil
}

// getSalesforceDocumentTimestampByEventV1 returns created or last modified timestamp by SalesforceAction
func getSalesforceDocumentTimestampByEventV1(event model.EventIdToProperties) (int64, error) {

	if event.Name == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED {

		date, exists := event.EventProperties[model.EP_SFCampaignMemberUpdated]
		if !exists || date == nil {
			return 0, errors.New("failed to get date")
		}
		timestamp, ok := date.(float64)
		if !ok || timestamp == 0 {
			return 0, errors.New("invalid timestamp")
		}

		return int64(timestamp), nil
	}

	date, exists := event.EventProperties[model.EP_SFCampaignMemberCreated]
	if !exists || date == nil {
		return 0, errors.New("failed to get date")
	}

	timestamp, ok := date.(float64)
	if !ok || timestamp == 0 {
		return 0, errors.New("invalid timestamp")
	}

	return int64(timestamp), nil

}
