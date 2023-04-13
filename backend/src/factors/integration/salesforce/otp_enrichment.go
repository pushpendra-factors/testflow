package salesforce

import (
	"encoding/json"
	C "factors/config"
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
)

const EmptyJsonStr = "{}"

var AllowedSfEventTypeForOTP = []string{

	U.EVENT_NAME_SALESFORCE_TASK_CREATED,
	U.EVENT_NAME_SALESFORCE_TASK_UPDATED,
	U.EVENT_NAME_SALESFORCE_EVENT_CREATED,
	U.EVENT_NAME_SALESFORCE_EVENT_UPDATED,
	U.EVENT_NAME_SALESFORCE_CONTACT_CREATED,
	U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED,
}

// WorkerForSfOtp sync salesforce Events to otp events
func WorkerForSfOtp(projectID int64, wg *sync.WaitGroup) {

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

	uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get OTP Unique Keys"})
		return
	}

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project")
		return
	}

	timezoneString, status := store.GetStore().GetTimezoneForProject(projectID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get timezone for project.")
		return
	}

	startTime, endTime, _ := U.GetQueryRangePresetYesterdayIn(timezoneString)

	for _, eventName := range AllowedSfEventTypeForOTP {

		switch eventName {
		case U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED, U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED:

			RunSFOfflineTouchPointRuleForCampaignMember(project, &otpRules, startTime, endTime, eventName, logCtx)

		case U.EVENT_NAME_SALESFORCE_TASK_UPDATED, U.EVENT_NAME_SALESFORCE_TASK_CREATED:
			RunSFOfflineTouchPointRuleForTasks(project, &otpRules, &uniqueOTPEventKeys, startTime, endTime, eventName, logCtx)

		case U.EVENT_NAME_SALESFORCE_EVENT_CREATED, U.EVENT_NAME_SALESFORCE_EVENT_UPDATED:
			RunSFOfflineTouchPointRuleForEvents(project, &otpRules, &uniqueOTPEventKeys, startTime, endTime, eventName, logCtx)
		default:
			continue

		}
	}

}

func RunSFOfflineTouchPointRuleForCampaignMember(project *model.Project, otpRules *[]model.OTPRule, startTime, endTime int64, eventName string, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(project.ID, startTime, endTime, eventName)
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

		err := ApplySFOfflineTouchPointRuleForCampaignMemberV1(project, otpRules, events[ei])
		if err != nil {
			logCtx.WithField("event", events[ei]).Info("Fail to apply OTP")
			return

		}
	}

}

func RunSFOfflineTouchPointRuleForTasks(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, startTime, endTime int64, eventName string, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(project.ID, startTime, endTime, eventName)
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

		err := ApplySFOfflineTouchPointRuleForTasksV1(project, otpRules, uniqueOTPEventKeys, events[ei])
		if err != nil {
			logCtx.WithField("event", events[ei]).Info("Fail to apply OTP")
			return
		}
	}

}

func RunSFOfflineTouchPointRuleForEvents(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, startTime, endTime int64, eventName string, logCtx *log.Entry) {

	eventsIds, events, err := PullEventIdsWithEventName(project.ID, startTime, endTime, eventName)
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

		err := ApplySFOfflineTouchPointRuleForEventsV1(project, otpRules, uniqueOTPEventKeys, events[ei])
		if err != nil {
			logCtx.WithField("event", events[ei]).Info("Fail to apply OTP")
			return
		}
	}

}

//Creates a unique key using ruleID, userID and salesforce task activity ID  as keyID for Salesforce Tasks.
func createOTPUniqueKeyForTasksV1(rule model.OTPRule, sfEvent eventIdToProperties, logCtx *log.Entry) (string, int) {

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

// CreateTouchPointEventForTasksAndEventsV1 - Creates offline touch-point for SF update events with given rule for SF Tasks/Events
func CreateTouchPointEventForTasksAndEventsV1(project *model.Project, sfEvent eventIdToProperties,
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

//ApplySFOfflineTouchPointRuleForTasksV1 Check if the condition are satisfied for creating OTP events for each rule for SF Tasks Updated.
func ApplySFOfflineTouchPointRuleForTasksV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, sfEvent eventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRuleForTasks"})

	if otpRules == nil || project == nil {
		return nil
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

		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !isSalesforceOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
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

//Creates a unique key using ruleID, userID and salesforce Event activity ID  as keyID for Salesforce Tasks.
func createOTPUniqueKeyForEventsV1(rule model.OTPRule, sfEvent eventIdToProperties, logCtx *log.Entry) (string, int) {

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

//ApplySFOfflineTouchPointRuleForEventsV1 Check if the condition are satisfied for creating OTP events for each rule for SF Event Updated.
func ApplySFOfflineTouchPointRuleForEventsV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, sfEvent eventIdToProperties) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRuleForEvents"})

	if otpRules == nil || project == nil {
		return nil
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
		if !isSalesforceOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}

		_, err1 := CreateTouchPointEventForTasksAndEventsV1(project, sfEvent, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for salesforce events.")
			continue
		}

	}
	return nil
}

func PullEventIdsWithEventName(projectId int64, startTimestamp int64, endTimestamp int64, eventName string) ([]string, map[string]eventIdToProperties, error) {
	db := C.GetServices().Db

	events := make(map[string]eventIdToProperties, 0)
	eventsIds := make([]string, 0)

	rows, _ := db.Raw("SELECT events.id, events.user_id , event_names.name, events.timestamp, events.properties FROM events"+
		" "+"LEFT JOIN event_names ON event_names.id = events.event_name_id"+
		" "+"WHERE events.project_id = ? AND event_names.project_id = ? AND event_names.name = ? AND events.timestamp >= ? AND events.timestamp <= ?", projectId, projectId, eventName, startTimestamp, endTimestamp).Rows()

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
			eventPropertiesBytes = []byte("{}") //EmptyJsonStr
		}

		var eventProperties U.PropertiesMap
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventProperties)
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
			EventProperties: eventProperties,
			Timestamp:       timestamp,
		}

		rowNum++
	}

	return eventsIds, events, nil
}

//ApplySFOfflineTouchPointRuleForCampaignMemberV1 Check if the condition are satisfied for creating OTP events for each rule for SF Campaign.
func ApplySFOfflineTouchPointRuleForCampaignMemberV1(project *model.Project, otpRules *[]model.OTPRule, sfEvent eventIdToProperties) error {

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

//CreateTouchPointEventCampaignMemberV1 - Creates offline touch point event for SF Campaign
func CreateTouchPointEventCampaignMemberV1(project *model.Project, sfEvent eventIdToProperties, rule model.OTPRule) (*SDK.TrackResponse, error) {

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
			if tt, ok := val.(int64); ok {
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
func getSalesforceDocumentTimestampByEventV1(event eventIdToProperties) (int64, error) {

	if event.Name == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED {

		date, exists := event.EventProperties["salesforce_campaignmember_lastmodifieddate"]
		if !exists || date == nil {
			return 0, errors.New("failed to get date")
		}
		return model.GetSalesforceDocumentTimestamp(date)
	}

	date, exists := event.EventProperties["salesforce_campaignmember_createddate"]
	if !exists || date == nil {
		return 0, errors.New("failed to get date")
	}

	return model.GetSalesforceDocumentTimestamp(date)

}
