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
	"time"
)

// EnrichV1 sync salesforce documents to events
func EnrichV1(projectID int64) ([]Status, bool) {

	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	if projectID == 0 {
		return statusByProjectAndType, true
	}

	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get otp Rules for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get OTP rules"})
		return statusByProjectAndType, true
	}

	uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get OTP Unique Keys"})
		return statusByProjectAndType, true
	}

	//allowedDocTypes := model.GetSalesforceDocumentTypeAlias(projectID)

	//salesforceSmartEventNames := GetSalesforceSmartEventNames(projectID)

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project")
		return statusByProjectAndType, true
	}

	anyFailure := false
	overAllSyncStatus := make(map[string]bool)

	timeZoneStr, status := store.GetStore().GetTimezoneForProject(projectID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get timezone for project.")
		return statusByProjectAndType, true
	}

	//todo parthG get event

	_, endTime, _ := U.GetQueryRangePresetYesterdayIn(timeZoneStr)

	//const EVENT_NAME_SALESFORCE_TASK_CREATED = "$sf_task_created"
	//const EVENT_NAME_SALESFORCE_TASK_UPDATED = "$sf_task_updated"
	//const EVENT_NAME_SALESFORCE_EVENT_CREATED = "$sf_event_created"
	//const EVENT_NAME_SALESFORCE_EVENT_UPDATED = "$sf_event_updated"
	//const EVENT_NAME_SALESFORCE_CONTACT_CREATED = "$sf_contact_created"
	//const EVENT_NAME_SALESFORCE_CONTACT_UPDATED = "$sf_contact_updated"

	eventsIds, events, err := PullEventIdsWithEventName(projectID, 100, endTime, U.EVENT_NAME_SALESFORCE_TASK_UPDATED)
	if err != nil {
		logCtx.Error("Failed to get events")
		return statusByProjectAndType, true
	}

	if len(eventsIds) == 0 {
		logCtx.Error("no event found")
		return statusByProjectAndType, true
	}

	//batches := U.GetStringListAsBatch(eventsIds, 16)
	logCtx.Info("event_ids %v", eventsIds)

	var enrichStatus enrichWorkerStatus
	for _, eventId := range eventsIds {

		startTime := time.Now().Unix()
		switch model.SalesforceDocumentTypeTask { //todo parthG iterate over all task

		//case model.SalesforceDocumentTypeCampaign, model.SalesforceDocumentTypeCampaignMember:
		//	endTimestamp := timeRange[1]
		//	errCode = enrichCampaignV1(project, &otpRules, events[batch[ei]], endTimestamp)
		case model.SalesforceDocumentTypeTask:
			err = ApplySFOfflineTouchPointRuleForTasksV1(project, &otpRules, &uniqueOTPEventKeys, events[eventId])

		case model.SalesforceDocumentTypeEvent:
			err = ApplySFOfflineTouchPointRuleForEventsV1(project, &otpRules, &uniqueOTPEventKeys, events[eventId])
		default:
			continue
		}

		logCtx.WithField("time_taken_in_secs", time.Now().Unix()-startTime).Debugf(
			"Sync  add type of sf completed.")
		if err != nil {
			logCtx.WithField("time_taken_in_secs", time.Now().Unix()-startTime).Debugf(
				"OTP enrichment sucssesfull")
		}

		if err != nil {
			enrichStatus.HasFailure = true
		}
	}

	for docTypeAlias, failure := range overAllSyncStatus {
		status := Status{ProjectID: projectID,
			Type: docTypeAlias}
		if failure {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailure = true
		} else {
			status.Status = U.CRM_SYNC_STATUS_SUCCESS
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType, anyFailure
}

func enrichCampaignV1(project *model.Project, otpRules *[]model.OTPRule, document *model.SalesforceDocument, endTimestamp int64) int {
	if project.ID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type == model.SalesforceDocumentTypeCampaign {
		return enrichCampaignToAllCampaignMembers(project, otpRules, document, endTimestamp)
	}

	if document.Type == model.SalesforceDocumentTypeCampaignMember {
		return enrichCampaignMember(project, otpRules, document, endTimestamp)
	}

	return http.StatusBadRequest
}

func enrichTaskV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, sfEvent eventIdToProperties) int {
	logCtx := log.WithField("project_id", project.ID)

	if project.ID == 0 {
		logCtx.Error("Invalid parameters in enrich Task")
		return http.StatusBadRequest
	}

	return http.StatusOK
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

//filterCheck- Returns true if all the filters applied are passed.
func filterCheckV1(rule model.OTPRule, sfEvent eventIdToProperties, logCtx *log.Entry) bool {

	var ruleFilters []model.TouchPointFilter
	err := U.DecodePostgresJsonbToStructType(&rule.Filters, &ruleFilters)
	if err != nil {
		logCtx.WithField("Document", sfEvent).WithError(err).Error("Failed to decode/fetch offline touch point rule FILTERS for salesforce document.")
		return false
	}

	filtersPassed := 0
	for _, filter := range ruleFilters {
		switch filter.Operator {
		case model.EqualsOpStr:
			if _, exists := sfEvent.EventProperties[filter.Property]; exists {
				if filter.Value != "" && sfEvent.EventProperties[filter.Property] == filter.Value {
					filtersPassed++
				}
			}
		case model.NotEqualOpStr:
			if _, exists := sfEvent.EventProperties[filter.Property]; exists {
				if filter.Value != "" && sfEvent.EventProperties[filter.Property] != filter.Value {
					filtersPassed++
				}
			}
		case model.ContainsOpStr:
			if _, exists := sfEvent.EventProperties[filter.Property]; exists {
				if filter.Value != "" && strings.Contains(sfEvent.EventProperties[filter.Property].(string), filter.Value) {
					filtersPassed++
				}
			}
		default:
			logCtx.WithField("Rule", rule).WithField("response", sfEvent).Error("No matching operator found for offline touch point rules for salesforce document.")
			continue
		}
	}
	return filtersPassed != 0 && filtersPassed == len(ruleFilters)
}

// CreateTouchPointEventForTasksAndEventsV1 - Creates offline touchpoint for SF update events with given rule for SF Tasks/Events
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

//Check if the condition are satisfied for creating OTP events for each rule for SF Tasks Updated.
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
		if !filterCheckV1(rule, sfEvent, logCtx) {
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
	ID                         string          `gorm:"primary_key:true;type:uuid" json:"id"`
	ProjectId                  int64           `gorm:"primary_key:true;" json:"project_id"`
	UserId                     string          `json:"user_id"`
	Name                       string          `json:"name"`
	PropertiesUpdatedTimestamp int64           `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
	EventProperties            U.PropertiesMap `json:"event_properties"`
	UserProperties             U.PropertiesMap `json:"user_properties"`
	Timestamp                  int64           `json:"timestamp"`
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
		if !filterCheckV1(rule, sfEvent, logCtx) {
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

func enrichEventV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, sfEvent eventIdToProperties) int {
	logCtx := log.WithField("project_id", project.ID)

	if project.ID == 0 {
		logCtx.Error("Invalid parameters in enrich Event")
		return http.StatusBadRequest
	}

	err := ApplySFOfflineTouchPointRuleForEventsV1(project, otpRules, uniqueOTPEventKeys, sfEvent)
	if err != nil {
		// log and continue
		logCtx.WithField("EventID", sfEvent.ID).WithField("userID", sfEvent.UserId).WithField("error", err).Warn("Failed creating offline touch point event for SF Events")
	}

	return http.StatusOK
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
