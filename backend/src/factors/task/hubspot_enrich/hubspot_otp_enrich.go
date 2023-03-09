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
	"sync"
)

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

	eventsIds, events, err := PullEventIdsWithEventName(projectID, startTime, endTime, U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED)
	if err != nil {
		logCtx.Error("Failed to get events")
		return
	}

	if len(eventsIds) == 0 {
		logCtx.Error("no event found")
		return
	}

	batches := U.GetStringListAsBatch(eventsIds, 16)
	logCtx.WithField("events", events).Info("pulled events")
	for bi := range batches {
		batch := batches[bi]

		for ei := range batch {

			eventName := events[batch[ei]].Name

			log.Info(fmt.Sprintf("event name  %s", eventName))

			switch eventName {

			case U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED, U.EVENT_NAME_HUBSPOT_CONTACT_CREATED:
				errCode := ApplyHSOfflineTouchPointRuleV1(project, &otpRules, &uniqueOTPEventKeys, events[batch[ei]], events[batch[ei]].Timestamp)
				if errCode != nil {
					log.Info("Fail to apply OTP")
				}

			}

		}

	}

}

func ApplyHSOfflineTouchPointRuleV1(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, event eventIdToProperties, eventTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRule",
		"document_action": fmt.Sprint("Contact")})

	if otpRules == nil || project == nil || &event == nil {
		return nil
	}

	logCtx.WithField("rules", otpRules).Info("applying otp rule")

	eventTimestamp = U.CheckAndGetStandardTimestamp(eventTimestamp)

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForFormsAndContactsV1(rule, event)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable & the record has changed property w.r.t filters

		//since otp runs only for event type "model.HubspotDocumentActionUpdated"

		if !canCreateHSTouchPoint(model.HubspotDocumentActionUpdated) {
			continue
		}

		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IntHubspot.IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}

		_, err1 := CreateTouchPointEventForFormsAndContactsV1(project, event, rule, eventTimestamp, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue
		}

		*uniqueOTPEventKeys = append(*uniqueOTPEventKeys, otpUniqueKey)

	}
	return nil
}

//Creates a unique key using ruleID, userID and eventID as keyID for Forms and contacts
func createOTPUniqueKeyForFormsAndContactsV1(rule model.OTPRule, event eventIdToProperties) (string, int) {

	ruleID := rule.ID
	userID := event.UserId
	keyID := event.ID

	uniqueKey := userID + ruleID + keyID

	return uniqueKey, http.StatusCreated

}

//canCreateHSTouchPoint- Returns true if the document action type is Updated for HS Contacts.
func canCreateHSTouchPoint(documentActionType int) bool {
	// Ignore doc types other than HubspotDocumentActionUpdated
	if documentActionType != model.HubspotDocumentActionUpdated {
		return false
	}
	return true
}

func PullEventIdsWithEventName(projectId int64, startTimestamp int64, endTimestamp int64, eventName string) ([]string, map[string]eventIdToProperties, error) {
	db := C.GetServices().Db

	events := make(map[string]eventIdToProperties, 0)
	eventsIds := make([]string, 0)

	rows, _ := db.Raw("SELECT events.id, events.user_id , event_names.name, events.timestamp, events.properties FROM events"+
		" "+"LEFT JOIN event_names ON event_names.id = events.event_name_id"+
		" "+"WHERE events.project_id = ? AND event_names.name = ? AND events.timestamp >= ? AND events.timestamp <= ?", projectId, eventName, startTimestamp, endTimestamp).Rows()

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
