package sdk

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"

	"factors/vendor_custom/machinery/v1/tasks"

	C "factors/config"
	M "factors/model"
	U "factors/util"
)

type TrackPayload struct {
	EventId string `json:"event_id"`
	UserId  string `json:"user_id"`
	// internal: create user with the user_id given, if true.
	CreateUser bool `json:"create_user"`
	// internal: indicates new user creation by external methods.
	IsNewUser       bool            `json:"-"`
	Name            string          `json:"event_name"`
	CustomerEventId *string         `json:"c_event_id"`
	EventProperties U.PropertiesMap `json:"event_properties"`
	UserProperties  U.PropertiesMap `json:"user_properties"`
	Timestamp       int64           `json:"timestamp`
	ProjectId       uint64          `json:"project_id"`
	Auto            bool            `json:"auto"`
	ClientIP        string          `json:"client_ip"`
	UserAgent       string          `json:"user_agent"`
}

type TrackResponse struct {
	EventId         string  `json:"event_id,omitempty"`
	UserId          string  `json:"user_id,omitempty"`
	Type            string  `json:"type,omitempty"`
	CustomerEventId *string `json:"c_event_id,omitempty"`
	Message         string  `json:"message,omitempty"`
	Error           string  `json:"error,omitempty"`
}

type IdentifyPayload struct {
	UserId string `json:"user_id"`
	// if create_user is true, create user with given id.
	CreateUser     bool   `json:"create_user"`
	CustomerUserId string `json:"c_uid"`
	JoinTimestamp  int64  `json:"join_timestamp"`
	Timestamp      int64  `json:"timestamp"`
}

type IdentifyResponse struct {
	UserId  string `json:"user_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AddUserPropertiesPayload struct {
	UserId string `json:"user_id"`
	// if create_user is true, create user with given id.
	CreateUser bool            `json:"create_user"`
	Timestamp  int64           `json:"timestamp"`
	Properties U.PropertiesMap `json:"properties"`
	ClientIP   string          `json:"client_ip"`
}

type AddUserPropertiesResponse struct {
	UserId  string `json:"user_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UpdateEventPropertiesPayload struct {
	EventId    string          `json:"event_id"`
	Properties U.PropertiesMap `json:"properties"`
	Timestamp  int64           `json:"timestamp"`
	UserAgent  string          `json:"user_agent"`
}

type UpdateEventPropertiesResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Response struct {
	EventId string `json:"event_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

const (
	SourceJSSDK  = "js_sdk"
	SourceAMPSDK = "amp_sdk"

	SourceSegment    = "segment"
	SourceShopify    = "shopify"
	SourceHubspot    = "hubspot"
	SourceSalesforce = "salesforce"
)

const RequestQueue = "sdk_request_queue"
const ProcessRequestTask = "process_sdk_request"

const (
	sdkRequestTypeEventTrack                = "sdk_event_track"
	sdkRequestTypeUserIdentify              = "sdk_user_identify"
	sdkRequestTypeUserAddProperties         = "sdk_user_add_properties"
	sdkRequestTypeEventUpdateProperties     = "sdk_event_update_properties"
	sdkRequestTypeAMPEventTrack             = "sdk_amp_event_track"
	sdkRequestTypedAMPEventUpdateProperties = "sdk_amp_event_update_properties"
)

func ProcessQueueRequest(token, reqType, reqPayloadStr string) (float64, string, error) {
	// Todo(Dinesh): Retry on panic: Add payload back to queue as return
	// from defer is not possible and notify panic.

	// Todo(Dinesh): Add request_id for better tracing.

	logCtx := log.WithFields(log.Fields{"queue": RequestQueue, "token": token,
		"req_type": reqType, "req_payload_str": reqPayloadStr})

	var response interface{}
	var status int

	switch reqType {
	case sdkRequestTypeEventTrack:
		var reqPayload TrackPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = TrackByToken(token, &reqPayload)

	case sdkRequestTypeUserIdentify:
		var reqPayload IdentifyPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = IdentifyByToken(token, &reqPayload)

	case sdkRequestTypeUserAddProperties:
		var reqPayload AddUserPropertiesPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = AddUserPropertiesByToken(token, &reqPayload)

	case sdkRequestTypeEventUpdateProperties:
		var reqPayload UpdateEventPropertiesPayload
		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = UpdateEventPropertiesByToken(token, &reqPayload)

	case sdkRequestTypeAMPEventTrack:
		var reqPayload AMPTrackPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = AMPTrackByToken(token, &reqPayload)

	case sdkRequestTypedAMPEventUpdateProperties:
		var reqPayload AMPUpdateEventPropertiesPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = AMPUpdateEventPropertiesByToken(token, &reqPayload)

	default:
		logCtx.Error("Invalid sdk request type on sdk process queue")
		return http.StatusInternalServerError, "", nil
	}

	responseBytes, _ := json.Marshal(response)
	logCtx = logCtx.WithField("status", status).WithField("response", string(responseBytes))

	// Do not retry on below conditions.
	if status == http.StatusBadRequest || status == http.StatusNotAcceptable || status == http.StatusUnauthorized {
		logCtx.WithField("processed", "true").Info("Failed to process sdk request permanantly.")
		return float64(status), "", nil
	}

	// Return error only for retry. Retry after a period till it is successfull.
	// Retry dependencies not found and failures which can be successful on retries.
	if status == http.StatusNotFound || status == http.StatusInternalServerError {
		logCtx.WithField("retry", "true").Info("Failed to process sdk request on sdk process queue. Retry.")
		return http.StatusInternalServerError, "",
			tasks.NewErrRetryTaskExp("EXP_RETRY__REQUEST_PROCESSING_FAILURE")
	}

	// Log for analysing queue process status.
	logCtx.WithField("processed", "true").Info("Processed sdk request.")

	return http.StatusOK, string(responseBytes), nil
}

func isRealtimeSessionRequired(skipSession bool, projectId uint64, skipProjectIds []uint64) bool {
	if skipSession {
		return false
	}

	for _, skipProjectId := range skipProjectIds {
		if skipProjectId == projectId {
			return false
		}
	}

	return true
}

func BackFillEventDataInCacheFromDb(project_id uint64, currentTime time.Time, no_of_days int, eventsLimit, propertyLimit, valuesLimit int, rowsLimit int, perQueryPullRange int, skipExpiryForCache bool) {

	// Preload EventNames-count-lastseen
	// TODO: Janani Make this 30 configurable, limit in cache, limit in ui
	logCtx := log.WithField("project_id", project_id)
	logCtx.Info("Refresh Event Properties Cache started")
	expiry := float64(U.EVENT_USER_CACHE_EXPIRY_SECS)
	if skipExpiryForCache {
		logCtx.Info("Setting Cache keys of this run to no-expiry")
		expiry = 0
	}
	allevents := make(map[string]bool)
	for i := 1; i <= no_of_days; i++ {
		var eventNames M.CacheEventNamesWithTimestamp
		eventNames.EventNames = make(map[string]U.CountTimestampTuple)

		dateFormat := currentTime.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		eventNamesKey, err := M.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(project_id, dateFormat)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key")
			return
		}

		logCtx.WithField("dateFormat", dateFormat).Info("Begin: Event names - DB query by occurence")
		begin := U.TimeNow()
		events, err := M.GetOrderedEventNamesFromDb(
			project_id,
			currentTime.AddDate(0, 0, -(i+perQueryPullRange)).Unix(),
			currentTime.AddDate(0, 0, -(i-1)).Unix(),
			eventsLimit)
		end := U.TimeNow()
		logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Event names - DB query by occurence")
		if err != nil {
			logCtx.WithError(err).Error("Failed to get values from DB - All event names")
			return
		}

		for item, element := range events {
			eventNames.EventNames[element.Name] = U.CountTimestampTuple{int64(element.LastSeen), element.Count}
			if item == eventsLimit {
				break
			}
		}
		enEventCache, err := json.Marshal(eventNames)
		if err != nil {
			logCtx.WithError(err).Error("Failed to marshall event names")
			return
		}
		logCtx.Info("Begin:EN:SB")
		begin = U.TimeNow()
		err = cacheRedis.SetPersistent(eventNamesKey, string(enEventCache), expiry)
		end = U.TimeNow()
		logCtx.WithFields(log.Fields{"timeTaken": end.Sub(begin).Milliseconds()}).Info("End:EN:SB")
		if err != nil {
			logCtx.WithError(err).Error("Failed to set events in cache")
			return
		}
		for event, _ := range eventNames.EventNames {
			allevents[event] = true
		}
	}

	for event := range allevents {
		for i := 1; i <= no_of_days; i++ {

			eventPropertyValuesInCache := make(map[*cacheRedis.Key]string)
			var eventProperties U.CachePropertyWithTimestamp
			eventProperties.Property = make(map[string]U.PropertyWithTimestamp)
			dateFormat := currentTime.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)

			logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event}).Info("Begin: Get event Properties DB call")
			begin := U.TimeNow()
			properties, err := M.GetRecentEventPropertyKeysWithLimits(
				project_id, event,
				currentTime.AddDate(0, 0, -(i+perQueryPullRange)).Unix(),
				currentTime.AddDate(0, 0, -(i-1)).Unix(),
				propertyLimit)
			end := U.TimeNow()
			logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Get event Properties DB call")
			if err != nil {
				logCtx.WithError(err).Error("Failed to fetch values from DB - user properties")
				return
			}

			if len(properties) > 0 {
				eventPropertiesKey, err := M.GetPropertiesByEventCategoryRollUpCacheKey(project_id, event, dateFormat)
				if err != nil {
					logCtx.WithError(err).Error("Failed to get cache key - properties")
					return
				}

				for _, property := range properties {

					logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event, "property": property.Key}).Info("Begin: Get event Property values DB call")
					begin := U.TimeNow()
					values, category, err := M.GetRecentEventPropertyValuesWithLimits(project_id, event, property.Key, valuesLimit, rowsLimit,
						currentTime.AddDate(0, 0, -(i+perQueryPullRange)).Unix(),
						currentTime.AddDate(0, 0, -(i-1)).Unix())
					end := U.TimeNow()
					logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event, "property": property.Key, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Get event Property values DB call")
					if err != nil {
						logCtx.WithError(err).Error("Failed to get values from db - property values")
						return
					}

					categoryMap := make(map[string]int64)
					categoryMap[category] = property.Count
					eventProperties.Property[property.Key] = U.PropertyWithTimestamp{
						category,
						categoryMap, // Setting precomputed ones to empty
						U.CountTimestampTuple{
							int64(property.LastSeen),
							property.Count}}

					var eventPropertyValues U.CachePropertyValueWithTimestamp
					eventPropertyValues.PropertyValue = make(map[string]U.CountTimestampTuple)

					if category == U.PropertyTypeCategorical {
						eventPropertyValuesKey, _ := M.GetValuesByEventPropertyRollUpCacheKey(project_id, event, property.Key, dateFormat)
						for _, value := range values {
							if value.Value != "" {
								eventPropertyValues.PropertyValue[value.Value] = U.CountTimestampTuple{
									int64(value.LastSeen),
									value.Count}
							}
						}
						enEventPropertyValuesCache, err := json.Marshal(eventPropertyValues)
						if err != nil {
							logCtx.WithError(err).Error("Failed to marshall - property values")
							return
						}
						eventPropertyValuesInCache[eventPropertyValuesKey] = string(enEventPropertyValuesCache)
					}
				}
				enEventPropertiesCache, err := json.Marshal(eventProperties)
				if err != nil {
					logCtx.WithError(err).Error("Failed to marshall - event properties")
					return
				}
				eventPropertyValuesInCache[eventPropertiesKey] = string(enEventPropertiesCache)
				logCtx.Info("Begin:EPV:SB")
				begin = U.TimeNow()
				err = cacheRedis.SetPersistentBatch(eventPropertyValuesInCache, expiry)
				end = U.TimeNow()
				logCtx.WithFields(log.Fields{"timeTaken": end.Sub(begin).Milliseconds()}).Info("End:EN:SB")
				if err != nil {
					logCtx.WithError(err).Error("Failed to set property values in cache")
					return
				}
			}
		}
	}
	logCtx.Info("Refresh Event Properties Cache Done!!!")
}

func setDefaultValuesToEventPropertiesBySource(eventProperties *U.PropertiesMap,
	source string, isAutoTracked bool) {

	if isAutoTracked && (source == SourceJSSDK || source == SourceAMPSDK) {
		U.SetDefaultValuesToEventProperties(eventProperties)
	}
}

func Track(projectId uint64, request *TrackPayload,
	skipSession bool, source string) (int, *TrackResponse) {
	logCtx := log.WithField("project_id", projectId)

	if projectId == 0 || request == nil {
		logCtx.WithField("request_payload", request).
			Error("Invalid track request.")
		return http.StatusBadRequest, &TrackResponse{}
	}

	// Add event_id if not available.
	// For queue, event_id is added before queueing.
	if request.EventId == "" {
		request.EventId = U.GetUUID()
	}

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	// Skipping track for configured projects.
	for _, skipProjectId := range C.GetSkipTrackProjectIds() {
		if skipProjectId == projectId {
			// Todo: Change status to StatusBadRequest, using StatusOk to avoid retries.
			return http.StatusOK, &TrackResponse{Error: "Tracking skipped."}
		}
	}

	// Precondition: Fails if event_name not provided.
	request.Name = strings.TrimSpace(request.Name) // Discourage whitespace on the end.
	if request.Name == "" {
		return http.StatusBadRequest,
			&TrackResponse{Error: "Tracking failed. Event name cannot be omitted or left empty."}
	}

	projectSettings, errCode := M.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError, &TrackResponse{Error: "Tracking failed. Invalid project."}
	}

	// Terminate track calls from bot user_agent.
	if *projectSettings.ExcludeBot && U.IsBotUserAgent(request.UserAgent) {
		return http.StatusNotModified, &TrackResponse{Message: "Tracking skipped. Bot request."}
	}

	var eventName *M.EventName
	var eventNameErrCode int

	// if auto_track enabled, auto_name = event_name else auto_name = "UCEN".
	// On 'auto' true event_name is the eventURL(e.g: factors.ai/u1/u2/u3) for JS_.
	if request.Auto {
		// Pass eventURL through filter and get corresponding event_name mapped by user.
		request.Name = strings.TrimSuffix(request.Name, "/")
		eventName, eventNameErrCode = M.FilterEventNameByEventURL(projectId, request.Name)
		if eventName != nil && eventNameErrCode == http.StatusFound {
			err := M.FillEventPropertiesByFilterExpr(&request.EventProperties, eventName.FilterExpr, request.Name)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectId, "filter_expr": eventName.FilterExpr,
					"event_url": request.Name, log.ErrorKey: err}).Error(
					"Failed to fill event url properties for auto tracked event.")
			}
		} else {
			// create a auto tracked event name, if no filter_expr match.
			eventName, eventNameErrCode = M.CreateOrGetAutoTrackedEventName(
				&M.EventName{Name: request.Name, ProjectId: projectId})
		}
	} else {
		eventName, eventNameErrCode = M.CreateOrGetUserCreatedEventName(
			&M.EventName{Name: request.Name, ProjectId: projectId})
	}

	if eventNameErrCode != http.StatusCreated && eventNameErrCode != http.StatusConflict &&
		eventNameErrCode != http.StatusFound {
		return eventNameErrCode, &TrackResponse{Error: "Tracking failed. Creating event_name failed."}
	}

	// Event Properties
	clientIP := request.ClientIP
	U.UnEscapeQueryParamProperties(&request.EventProperties)
	definedEventProperties, hasDefinedMarketingProperty := U.MapEventPropertiesToDefinedProperties(
		&request.EventProperties)
	eventProperties := U.GetValidatedEventProperties(definedEventProperties)
	if ip, ok := (*eventProperties)[U.EP_INTERNAL_IP]; ok && ip != "" {
		clientIP = ip.(string)
	}
	// Added IP to event properties for internal usage.
	(*eventProperties)[U.EP_INTERNAL_IP] = clientIP

	U.SanitizeProperties(eventProperties)

	var userProperties *U.PropertiesMap
	if request.UserProperties == nil {
		request.UserProperties = U.PropertiesMap{}
	}
	FillUserAgentUserProperties(&request.UserProperties, request.UserAgent)

	response := &TrackResponse{}
	initialUserProperties := U.GetInitialUserProperties(request.EventId, eventProperties)
	isNewUser := request.IsNewUser

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		_, errCode := M.GetUser(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

	var existingUserProperties *map[string]interface{}
	if request.CreateUser || request.UserId == "" {
		newUser := &M.User{ProjectId: projectId}

		// create user with given id.
		if request.CreateUser {
			if request.UserId == "" {
				logCtx.Error("Track request create user is true but user_id is not given.")
				return http.StatusInternalServerError,
					&TrackResponse{Error: "Tracking failed. User creation failed."}
			}

			newUser.ID = request.UserId
			// use event occurrence timestamp as
			// user create timestamp.
			newUser.JoinTimestamp = request.Timestamp
		}

		// Precondition: create new user, if user_id not given or create_user is true.
		createdUser, errCode := M.CreateUser(newUser)
		if errCode != http.StatusCreated {
			return errCode, &TrackResponse{Error: "Tracking failed. User creation failed."}
		}

		request.UserId = createdUser.ID
		response.UserId = createdUser.ID
		isNewUser = true

		// Initialize with initial user properties.
		userProperties = initialUserProperties
	} else {
		// Adding initial user properties if user_id exists,
		// but initial properties are not. i.e user created on identify.
		existingUserProperties, errCode = M.GetLatestUserPropertiesOfUserAsMap(projectId, request.UserId)
		if errCode != http.StatusFound {
			logCtx.WithField("user_id", errCode).WithField("err_code",
				errCode).Error("Tracking failed. Get user properties as map failed.")
			return errCode, &TrackResponse{Error: "Tracking failed while getting user."}
		}

		// Checking any initial user properties exists already.
		// Initial user properties should not be overwritten,
		// even if one exists already.
		initialUserPropertyExists := false
		for k := range *initialUserProperties {
			if _, exists := (*existingUserProperties)[k]; exists {
				initialUserPropertyExists = true
				break
			}
		}

		// UpdateUserProperties takes care merging new properites,
		// with existing user properites. So setting only the
		// intialProperites.
		if !initialUserPropertyExists {
			userProperties = initialUserProperties
		}
	}

	requestUserProperties := U.GetValidatedUserProperties(&request.UserProperties)
	if userProperties != nil {
		for k, v := range *requestUserProperties {
			if _, exists := (*userProperties)[k]; !exists {
				(*userProperties)[k] = v
			}
		}
	} else {
		userProperties = requestUserProperties
	}

	_ = M.FillLocationUserProperties(userProperties, clientIP)
	// Add latest touch user properties.
	if hasDefinedMarketingProperty {
		U.FillLatestTouchUserProperties(userProperties, eventProperties)
	}
	// Add user properties from form submit event properties.
	if request.Name == U.EVENT_NAME_FORM_SUBMITTED {
		customerUserId, errCode := M.FillUserPropertiesAndGetCustomerUserIdFromFormSubmit(
			projectId, request.UserId, userProperties, eventProperties)
		if errCode == http.StatusInternalServerError {
			log.WithFields(log.Fields{"userProperties": userProperties,
				"eventProperties": eventProperties}).Error(
				"Failed adding user properties from form submitted event.")
			response.Error = "Failed adding user properties."
		}

		if customerUserId != "" {
			errCode, _ := Identify(projectId, &IdentifyPayload{
				UserId: request.UserId, CustomerUserId: customerUserId})
			if errCode != http.StatusOK {
				log.WithFields(log.Fields{"projectId": projectId, "userId": request.UserId,
					"customerUserId": customerUserId}).Error("Failed to identify user on form submit event.")
			}
		}
	}

	if existingUserProperties == nil {
		existingUserProperties, errCode = M.GetLatestUserPropertiesOfUserAsMap(projectId, request.UserId)
		if errCode == http.StatusInternalServerError {
			logCtx.WithField("user_id", errCode).WithField("err_code",
				errCode).Error("Tracking failed. Get user properties as map failed.")
		}
	}

	err := U.FillFirstEventUserPropertiesIfNotExist(existingUserProperties, userProperties, request.Timestamp)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to fill day of first event user_properties on track.")
	}

	logCtx = logCtx.WithField("user_properties", userProperties)
	userPropsJSON, err := json.Marshal(userProperties)
	if err != nil {
		logCtx.WithError(err).Error("Update user properites on track failed. JSON marshal failed.")
		response.Error = "Failed updating user properties."
	}

	userPropertiesId, errCode := M.UpdateUserProperties(projectId, request.UserId,
		&postgres.Jsonb{userPropsJSON}, request.Timestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		logCtx.WithField("err_code", errCode).
			Error("Update user properties on track failed. DB update failed.")
		response.Error = "Failed updating user properties."
	}

	event := &M.Event{
		ID:               request.EventId,
		EventNameId:      eventName.ID,
		CustomerEventId:  request.CustomerEventId,
		Timestamp:        request.Timestamp,
		ProjectId:        projectId,
		UserId:           request.UserId,
		UserPropertiesId: userPropertiesId,
	}

	// Property used as flag for skipping session on offline session worker.
	if skipSession {
		(*eventProperties)[U.EP_SKIP_SESSION] = U.PROPERTY_VALUE_TRUE
	}

	skipSessionForAllProjects, skipSessionProjectIds := C.GetSkipSessionProjects()
	skipSessionRealtime := skipSession || skipSessionForAllProjects

	if isRealtimeSessionRequired(skipSessionRealtime, projectId, skipSessionProjectIds) {
		session, errCode := M.CreateOrGetSessionEvent(projectId, request.UserId,
			isNewUser, hasDefinedMarketingProperty, request.Timestamp,
			eventProperties, userProperties, userPropertiesId)
		if errCode != http.StatusCreated && errCode != http.StatusFound {
			response.Error = "Failed to associate with a session."
		}

		if session == nil || errCode == http.StatusNotFound || errCode == http.StatusInternalServerError {
			log.Error("Session is nil even after CreateOrGetSessionEvent.")
			return errCode, &TrackResponse{Error: "Tracking failed. Unable to associate with a session."}
		}

		(*eventProperties)[U.EP_SESSION_COUNT] = session.Count
		event.SessionId = &session.ID
	}

	setDefaultValuesToEventPropertiesBySource(eventProperties, source, request.Auto)
	eventPropsJSON, err := json.Marshal(eventProperties)
	if err != nil {
		return http.StatusBadRequest, &TrackResponse{Error: "Tracking failed. Invalid properties."}
	}
	event.Properties = postgres.Jsonb{eventPropsJSON}

	createdEvent, errCode := M.CreateEvent(event)
	if errCode == http.StatusNotAcceptable {
		return errCode, &TrackResponse{Error: "Tracking failed. Event creation failed. Invalid payload.",
			CustomerEventId: request.CustomerEventId}
	} else if errCode != http.StatusCreated {
		return errCode, &TrackResponse{Error: "Tracking failed. Event creation failed."}
	}

	// Success response.
	response.EventId = createdEvent.ID
	response.Message = "User event tracked successfully."
	response.CustomerEventId = request.CustomerEventId
	return http.StatusOK, response
}

func getIdentifiedUserPropertiesAsJsonb(customerUserId string) (*postgres.Jsonb, error) {
	if customerUserId == "" {
		return nil, errors.New("invalid customer user id")
	}

	properties := map[string]interface{}{
		U.UP_USER_ID: customerUserId,
	}

	if U.IsEmail(customerUserId) {
		properties[U.UP_EMAIL] = customerUserId
	}

	return U.EncodeToPostgresJsonb(&properties)
}

func Identify(projectId uint64, request *IdentifyPayload) (int, *IdentifyResponse) {
	// Precondition: Fails to identify if customer_user_id not present.
	if request.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		return http.StatusBadRequest, &IdentifyResponse{
			Error: "Identification failed. Missing mandatory keys c_uid."}
	}

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"user_id": request.UserId, "customer_user_id": request.CustomerUserId})

	userProperties, err := getIdentifiedUserPropertiesAsJsonb(request.CustomerUserId)
	if err != nil || userProperties == nil {
		logCtx.WithError(err).Error("Failed to get and add identified user properties on identify.")
	}

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		_, errCode := M.GetUser(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

	// Precondition: customer_user_id present, user_id not.
	// if customer_user has user already : respond with same user.
	// else : creating a new_user with the given customer_user_id and respond with new_user_id.
	if request.CreateUser || request.UserId == "" {
		response := &IdentifyResponse{}

		userLatest, errCode := M.GetUserLatestByCustomerUserId(projectId, request.CustomerUserId)
		switch errCode {
		case http.StatusInternalServerError:
			return errCode, &IdentifyResponse{
				Error: "Identification failed. Processing without user_id failed."}

		case http.StatusNotFound:
			newUser := M.User{
				ProjectId:      projectId,
				CustomerUserId: request.CustomerUserId,
				JoinTimestamp:  request.JoinTimestamp,
			}

			// create user with given user id.
			if request.CreateUser {
				if request.UserId == "" {
					logCtx.Error("Identify request payload with create_user true without user_id.")
					return http.StatusInternalServerError,
						&IdentifyResponse{Error: "Identification failed. User creation failed."}
				}
				newUser.ID = request.UserId
			}

			if userProperties != nil {
				newUser.Properties = *userProperties
			}

			_, errCode := M.CreateUser(&newUser)
			if errCode != http.StatusCreated {
				return errCode, &IdentifyResponse{
					Error: "Identification failed. User creation failed."}
			}
			response.UserId = newUser.ID

		case http.StatusFound:
			response.UserId = userLatest.ID
		}

		response.Message = "User has been identified successfully."
		return http.StatusOK, response
	}

	scopeUser, errCode := M.GetUser(projectId, request.UserId)
	if errCode != http.StatusFound {
		return errCode, &IdentifyResponse{Error: "Identification failed. Invalid user_id."}
	}

	// Precondition: Given user already identified as given customer_user.
	if scopeUser.CustomerUserId == request.CustomerUserId {
		return http.StatusOK, &IdentifyResponse{Message: "Identified already."}
	}

	// Precondition: user is already identified with different customer_user.
	// Creating a new user with the given customer_user_id and respond with new_user_id.
	if scopeUser.CustomerUserId != "" {
		newUser := M.User{
			ProjectId:      projectId,
			CustomerUserId: scopeUser.CustomerUserId,
			JoinTimestamp:  request.JoinTimestamp,
		}

		// create user with given user id.
		if request.CreateUser {
			if request.UserId == "" {
				logCtx.Error("Identify request payload with create_user true without user_id.")
				return http.StatusInternalServerError,
					&IdentifyResponse{Error: "Identification failed. User creation failed."}
			}
			newUser.ID = request.UserId
		}

		if userProperties != nil {
			newUser.Properties = *userProperties
		}

		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, &IdentifyResponse{Error: "Identification failed. User creation failed."}
		}

		return http.StatusOK, &IdentifyResponse{UserId: newUser.ID,
			Message: "User has been identified successfully"}
	}

	// Happy path. Maps customer_user to an user.
	updateUser := &M.User{CustomerUserId: request.CustomerUserId,
		JoinTimestamp: request.JoinTimestamp}
	if userProperties != nil {
		updateUser.Properties = *userProperties
	}

	_, errCode = M.UpdateUser(projectId, request.UserId, updateUser, request.Timestamp)
	if errCode != http.StatusAccepted {
		return errCode, &IdentifyResponse{Error: "Identification failed. Failed mapping customer_user to user"}
	}

	return http.StatusOK, &IdentifyResponse{Message: "User has been identified successfully."}
}

func AddUserProperties(projectId uint64,
	request *AddUserPropertiesPayload) (int, *AddUserPropertiesResponse) {

	logCtx := log.WithField("project_id", projectId)

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	// Validate properties.
	validProperties := U.GetValidatedUserProperties(&request.Properties)
	_ = M.FillLocationUserProperties(validProperties, request.ClientIP)
	propertiesJSON, err := json.Marshal(validProperties)
	if err != nil {
		return http.StatusBadRequest,
			&AddUserPropertiesResponse{Error: "Add user properties failed. Invalid properties."}
	}

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		_, errCode := M.GetUser(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

	// Precondition: user_id not given.
	if request.CreateUser || request.UserId == "" {
		newUser := &M.User{
			ProjectId:  projectId,
			Properties: postgres.Jsonb{propertiesJSON},
		}

		// create user with given user id.
		if request.CreateUser {
			if request.UserId == "" {
				logCtx.Error("Add user properties request is with create_user true and without user_id.")
				return http.StatusInternalServerError,
					&AddUserPropertiesResponse{Error: "Add user properties failed. User create failed"}
			}
			newUser.ID = request.UserId
			newUser.JoinTimestamp = request.Timestamp
		}

		// Create user with properties and respond user_id. Only properties allowed on create.
		newUser, errCode := M.CreateUser(newUser)
		if errCode != http.StatusCreated {
			return errCode, &AddUserPropertiesResponse{Error: "Add user properties failed. User create failed"}
		}
		return http.StatusOK, &AddUserPropertiesResponse{UserId: newUser.ID,
			Message: "Added user properties successfully."}
	}

	user, errCode := M.GetUser(projectId, request.UserId)
	if errCode == http.StatusNotFound {
		return http.StatusBadRequest,
			&AddUserPropertiesResponse{Error: "Add user properties failed. Invalid user_id."}
	} else if errCode == http.StatusInternalServerError {
		return errCode,
			&AddUserPropertiesResponse{Error: "Add user properties failed"}
	}

	_, errCode = M.UpdateUserPropertiesByCurrentProperties(projectId, user.ID,
		user.PropertiesId, &postgres.Jsonb{propertiesJSON}, request.Timestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return errCode,
			&AddUserPropertiesResponse{Error: "Add user properties failed."}
	}

	return http.StatusOK,
		&AddUserPropertiesResponse{Message: "Added user properties successfully."}
}

func enqueueRequest(token, reqType, reqPayload interface{}) error {
	reqPayloadJson, err := json.Marshal(reqPayload)
	if err != nil {
		log.WithError(err).WithField("token", token).Error(
			"Failed to marshal sdk request queue payload")
		return err
	}

	queueClient := C.GetServices().QueueClient
	_, err = queueClient.SendTask(&tasks.Signature{
		Name:                 ProcessRequestTask,
		RoutingKey:           RequestQueue, // queue to send.
		RetryLaterOnPriority: true,         // allow delayed tasks to run on priority.
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: token,
			},
			{
				Type:  "string",
				Value: reqType,
			},
			{
				Type:  "string",
				Value: string(reqPayloadJson),
			},
		},
	})

	return err
}

func excludeBotRequestBySetting(token, userAgent string) bool {
	settings, errCode := M.GetProjectSettingByTokenWithCacheAndDefault(token)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).
			Error("Failed to get project settings on excludeBotRequestBeforeQueue.")
		return false
	}

	return settings != nil && *settings.ExcludeBot && U.IsBotUserAgent(userAgent)
}

func TrackByToken(token string, reqPayload *TrackPayload) (int, *TrackResponse) {
	project, errCode := M.GetProjectByToken(token)
	if errCode == http.StatusFound {
		return Track(project.ID, reqPayload, false, SourceJSSDK)
	}

	if errCode == http.StatusNotFound {
		log.WithField("token", token).Error(
			"Failed to get project from sdk project token.")
		return http.StatusUnauthorized,
			&TrackResponse{Error: "Tracking failed. Invalid token."}
	}

	return errCode, &TrackResponse{Error: "Tracking failed."}
}

func TrackWithQueue(token string, reqPayload *TrackPayload,
	queueAllowedTokens []string) (int, *TrackResponse) {

	if excludeBotRequestBySetting(token, reqPayload.UserAgent) {
		return http.StatusNotModified,
			&TrackResponse{Message: "Tracking skipped. Bot request."}
	}

	if U.UseQueue(token, queueAllowedTokens) {
		reqPayload.EventId = U.GetUUID()

		response := &TrackResponse{EventId: reqPayload.EventId}

		// create user with given id,
		// if user_id not given on request.
		if reqPayload.UserId == "" {
			reqPayload.CreateUser = true
			reqPayload.UserId = U.GetUUID()
			// add user_id to response.
			response.UserId = reqPayload.UserId
		}

		// Add request received timestamp as
		// event timestamp, if not given.
		if reqPayload.Timestamp == 0 {
			reqPayload.Timestamp = time.Now().Unix()
		}

		err := enqueueRequest(token, sdkRequestTypeEventTrack, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue track request.")
			return http.StatusInternalServerError, &TrackResponse{Message: "Tracking failed."}
		}

		response.Message = "User event tracked successfully."

		return http.StatusOK, response
	}

	return TrackByToken(token, reqPayload)
}

func IdentifyByToken(token string, reqPayload *IdentifyPayload) (int, *IdentifyResponse) {
	project, errCode := M.GetProjectByToken(token)
	if errCode == http.StatusFound {
		return Identify(project.ID, reqPayload)
	}

	if errCode == http.StatusNotFound {
		log.WithField("token", token).Error(
			"Failed to get project from sdk project token.")
		return http.StatusUnauthorized,
			&IdentifyResponse{Error: "Identify failed. Invalid token."}
	}

	return errCode, &IdentifyResponse{Error: "Identify failed."}
}

func IdentifyWithQueue(token string, reqPayload *IdentifyPayload,
	queueAllowedTokens []string) (int, *IdentifyResponse) {

	if U.UseQueue(token, queueAllowedTokens) {
		response := &IdentifyResponse{}

		if reqPayload.UserId == "" {
			reqPayload.CreateUser = true
			reqPayload.UserId = U.GetUUID()
			// add user_id to response.
			response.UserId = reqPayload.UserId
		}

		if reqPayload.Timestamp == 0 {
			reqPayload.Timestamp = time.Now().Unix()
		}

		if reqPayload.JoinTimestamp == 0 {
			reqPayload.JoinTimestamp = reqPayload.Timestamp
		}

		err := enqueueRequest(token, sdkRequestTypeUserIdentify, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue identify request.")
			return http.StatusInternalServerError,
				&IdentifyResponse{Error: "Identify failed."}
		}

		response.Message = "User has been identified successfully"

		return http.StatusOK, response
	}

	return IdentifyByToken(token, reqPayload)
}

func AddUserPropertiesByToken(token string,
	reqPayload *AddUserPropertiesPayload) (int, *AddUserPropertiesResponse) {

	project, errCode := M.GetProjectByToken(token)
	if errCode == http.StatusFound {
		return AddUserProperties(project.ID, reqPayload)
	}

	if errCode == http.StatusNotFound {
		log.WithField("token", token).Error(
			"Failed to get project from sdk project token.")
		return http.StatusUnauthorized, &AddUserPropertiesResponse{
			Error: "Add user properties failed. Invalid token."}
	}

	return errCode, &AddUserPropertiesResponse{Error: "Add user properties failed."}
}

func AddUserPropertiesWithQueue(token string, reqPayload *AddUserPropertiesPayload,
	queueAllowedTokens []string) (int, *AddUserPropertiesResponse) {

	if U.UseQueue(token, queueAllowedTokens) {
		response := &AddUserPropertiesResponse{}

		if reqPayload.UserId == "" {
			reqPayload.CreateUser = true
			reqPayload.UserId = U.GetUUID()
			// add user_id to response.
			response.UserId = reqPayload.UserId
		}

		if reqPayload.Timestamp == 0 {
			reqPayload.Timestamp = time.Now().Unix()
		}

		err := enqueueRequest(token, sdkRequestTypeUserAddProperties, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue add user properties request.")
			return http.StatusInternalServerError,
				&AddUserPropertiesResponse{Error: "Add user properties failed."}
		}

		response.Message = "Added user properties successfully."

		return http.StatusOK, response
	}

	return AddUserPropertiesByToken(token, reqPayload)
}

func UpdateEventPropertiesByToken(token string,
	reqPayload *UpdateEventPropertiesPayload) (int, *UpdateEventPropertiesResponse) {

	project, errCode := M.GetProjectByToken(token)
	if errCode == http.StatusFound {
		return UpdateEventProperties(project.ID, reqPayload)
	}

	if errCode == http.StatusNotFound {
		log.WithField("token", token).Error(
			"Failed to get project from sdk project token.")
		return http.StatusUnauthorized, &UpdateEventPropertiesResponse{
			Error: "Update event properties failed. Invalid token."}
	}

	return errCode, &UpdateEventPropertiesResponse{Error: "Failed to update event properties using token."}
}

func UpdateEventPropertiesWithQueue(token string, reqPayload *UpdateEventPropertiesPayload,
	queueAllowedTokens []string) (int, *UpdateEventPropertiesResponse) {

	if excludeBotRequestBySetting(token, reqPayload.UserAgent) {
		return http.StatusNotModified, &UpdateEventPropertiesResponse{
			Message: "Update event properties skipped. Bot request."}
	}

	if U.UseQueue(token, queueAllowedTokens) {
		// add queued timestamp, if timestmap is not given.
		if reqPayload.Timestamp == 0 {
			reqPayload.Timestamp = time.Now().Unix()
		}

		err := enqueueRequest(token, sdkRequestTypeEventUpdateProperties, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue updated event properties request.")
			return http.StatusInternalServerError,
				&UpdateEventPropertiesResponse{
					Error: "Update event properties failed. Request reception failure."}
		}

		return http.StatusOK,
			&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
	}

	return UpdateEventPropertiesByToken(token, reqPayload)
}

func updateInitialUserPropertiesFromUpdateEventProperties(projectID uint64,
	eventID, userID, userPropertiesID string, newInitialUserProperties *U.PropertiesMap) int {

	logCtx := log.WithField("project_id", projectID).WithField("event_id", eventID)

	userPropertiesJsonb, errCode := M.GetUserProperties(projectID, userID, userPropertiesID)
	if errCode != http.StatusFound {
		return errCode
	}

	userProperties, err := U.DecodePostgresJsonb(userPropertiesJsonb)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode user_properties.")
		return http.StatusBadRequest
	}

	initialPageEventID, initialPageEventIDExists := (*userProperties)[U.UP_INITIAL_PAGE_EVENT_ID]

	// Skip the update, if initial properties exist and initial_page_event_id
	// doesn't exists, for backward compatibility.
	_, initialPageRawURLExists := (*userProperties)[U.UP_INITIAL_PAGE_RAW_URL]
	if !initialPageEventIDExists && initialPageRawURLExists {
		return http.StatusAccepted
	}

	// Do not update if the initial_page_event_id on user_properites is
	// not the current event id.
	if initialPageEventIDExists && initialPageEventID != eventID {
		return http.StatusAccepted
	}

	isUpdateRequired := false
	for key, value := range *newInitialUserProperties {
		if value == (*userProperties)[key] {
			continue
		}

		(*userProperties)[key] = value
		isUpdateRequired = true
	}

	if !isUpdateRequired {
		return http.StatusAccepted
	}

	updateUserPropertiesJson, err := U.EncodeToPostgresJsonb(userProperties)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to marshal user_properties after adding initial user_properties.")
		return http.StatusBadRequest
	}

	errCode = M.OverwriteUserProperties(projectID, userID,
		userPropertiesID, updateUserPropertiesJson)
	if errCode != http.StatusAccepted {
		logCtx.WithField("err_code", errCode).
			Error("Failed to overwrite user_properties after adding initial user_properties.")
		return errCode
	}

	return http.StatusAccepted
}

func UpdateEventProperties(projectId uint64,
	request *UpdateEventPropertiesPayload) (int, *UpdateEventPropertiesResponse) {

	// add received timestamp before processing, if not given.
	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	updateAllowedProperties := U.GetUpdateAllowedEventProperties(&request.Properties)
	properitesToBeUpdated := U.GetValidatedEventProperties(updateAllowedProperties)
	if len(*properitesToBeUpdated) == 0 {
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "No valid properties given to update."}
	}

	event, errCode := M.GetEventById(projectId, request.EventId)
	if errCode == http.StatusNotFound && request.Timestamp > U.UnixTimeBeforeDuration(time.Hour*5) {
		log.WithField("event_id", request.EventId).
			WithField("timestamp", request.Timestamp).
			Error("Failed old update event properties request with unavailable event_id permanently.")
		return http.StatusBadRequest, &UpdateEventPropertiesResponse{
			Error: "Update event properties failed permanantly."}
	}
	if errCode != http.StatusFound {
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed. Invalid event."}
	}

	user, errCode := M.GetUser(projectId, event.UserId)
	if errCode != http.StatusFound {
		return errCode, &UpdateEventPropertiesResponse{
			Error: "Update event properties failed. User not found."}
	}

	errCode = M.UpdateEventProperties(projectId, request.EventId,
		properitesToBeUpdated, request.Timestamp)
	if errCode != http.StatusAccepted {
		return errCode,
			&UpdateEventPropertiesResponse{
				Error: "Update event properties failed. Failed to update given properties."}
	}

	newInitialUserProperties := U.GetInitialUserProperties(event.ID, properitesToBeUpdated)

	// Update user_properties state associate to event.
	errCode = updateInitialUserPropertiesFromUpdateEventProperties(projectId, event.ID,
		event.UserId, event.UserPropertiesId, newInitialUserProperties)
	if errCode != http.StatusAccepted {
		return errCode,
			&UpdateEventPropertiesResponse{
				Error: "Update event properties failed. Failed to update event user properties."}
	}

	if event.UserPropertiesId == user.PropertiesId {
		return http.StatusAccepted,
			&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
	}

	// Update lastest user properties state of user.
	errCode = updateInitialUserPropertiesFromUpdateEventProperties(projectId, event.ID,
		event.UserId, user.PropertiesId, newInitialUserProperties)
	if errCode != http.StatusAccepted {
		return errCode,
			&UpdateEventPropertiesResponse{
				Error: "Update event properties failed. Failed to update latest user properties."}
	}

	return http.StatusAccepted,
		&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
}

type AMPTrackPayload struct {
	ClientID           string  `json:"client_id"` // amp user_id
	SourceURL          string  `json:"source_url"`
	Title              string  `json:"title"`
	Referrer           string  `json:"referrer"`
	ScreenHeight       float64 `json:"screen_height"`
	ScreenWidth        float64 `json:"screen_width"`
	PageLoadTimeInSecs float64 `json:"page_load_time_in_secs"`

	// internal
	Timestamp int64  `json:"timestamp"`
	UserAgent string `json:"user_agent"`
	ClientIP  string `json:"client_ip"`
}
type AMPUpdateEventPropertiesPayload struct {
	ClientID          string  `json:"client_id"` // amp user_id
	SourceURL         string  `json:"source_url"`
	PageScrollPercent float64 `json:"page_scroll_percent"`
	PageSpentTime     float64 `json:"page_spent_time"`

	// internal
	Timestamp int64  `json:"timestamp"`
	UserAgent string `json:"user_agent"`
}
type AMPTrackResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func AMPUpdateEventPropertiesByToken(token string,
	reqPayload *AMPUpdateEventPropertiesPayload) (int, *Response) {

	project, errCode := M.GetProjectByToken(token)
	if errCode != http.StatusFound {
		return http.StatusUnauthorized, &Response{Error: "Invalid token"}
	}

	logCtx := log.WithField("project_id", project.ID)

	parsedSourceURL, err := U.ParseURLStable(reqPayload.SourceURL)

	if err != nil {
		logCtx.WithField("canonical_url", reqPayload.SourceURL).WithError(err).Error(
			"Failed to parsing page url from canonical_url query param on amp sdk update event properties")
		return http.StatusBadRequest, &Response{Error: "Invalid page url"}
	}

	pageURL := U.CleanURI(parsedSourceURL.Host + parsedSourceURL.Path)

	user, errCode := M.CreateOrGetAMPUser(project.ID, reqPayload.ClientID, reqPayload.Timestamp)
	if errCode != http.StatusFound {
		return errCode, &Response{Error: "Invalid amp user."}
	}

	logCtx = logCtx.WithField("user_id", user.ID).WithField("page_url", pageURL)

	eventID, errCode := GetCacheAMPSDKEventIDByPageURL(project.ID, user.ID, pageURL)
	if errCode != http.StatusFound {
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to get eventId by page_url from cache.")
		}

		// Do not retry on failure or cache miss.
		return http.StatusNotModified,
			&Response{Error: "Failed to update event properties. Invalid request."}
	}

	updateEventProperties := U.PropertiesMap{}

	if reqPayload.PageSpentTime != 0 {
		updateEventProperties[U.EP_PAGE_SPENT_TIME] = reqPayload.PageSpentTime
	}

	if reqPayload.PageScrollPercent != 0 {
		updateEventProperties[U.EP_PAGE_SCROLL_PERCENT] = reqPayload.PageScrollPercent
	}

	errCode = M.UpdateEventProperties(project.ID, eventID, &updateEventProperties, time.Now().Unix())

	if errCode != http.StatusAccepted {
		logCtx.WithFields(log.Fields{"project_id": project.ID, "event_id": eventID}).
			Error("Failed to update event properties")
		return errCode, &Response{Error: "Failed to update event properties."}
	}

	return http.StatusAccepted, &Response{Message: "Updated event properties successfully."}
}

func AMPTrackByToken(token string, reqPayload *AMPTrackPayload) (int, *Response) {
	project, errCode := M.GetProjectByToken(token)
	if errCode != http.StatusFound {
		return http.StatusUnauthorized, &Response{Error: "Invalid token"}
	}

	logCtx := log.WithField("project_id", project.ID)

	var isNewUser bool
	user, errCode := M.CreateOrGetAMPUser(project.ID, reqPayload.ClientID, reqPayload.Timestamp)
	if errCode != http.StatusFound && errCode != http.StatusCreated {
		return errCode, &Response{Error: "Invalid user"}
	}

	if errCode == http.StatusCreated {
		isNewUser = true
	}

	parsedSourceURL, err := U.ParseURLStable(reqPayload.SourceURL)
	if err != nil {
		logCtx.WithField("canonical_url", reqPayload.SourceURL).WithError(err).Error(
			"Failed to parsing page url from canonical_url query param on amp sdk track")
		return http.StatusBadRequest, &Response{Error: "Invalid page url"}
	}

	pageURL := U.CleanURI(parsedSourceURL.Host + parsedSourceURL.Path)

	var referrerRawURL, referrerURL, referrerDomain string
	if reqPayload.Referrer != "" {
		parsedParamReferrerURL, err := U.ParseURLStable(reqPayload.Referrer)
		if err == nil {
			referrerRawURL = reqPayload.Referrer
			referrerURL = parsedParamReferrerURL.Host + parsedParamReferrerURL.Path
			referrerDomain = parsedParamReferrerURL.Host
		} else {
			logCtx.WithError(err).Error(
				"Failed parsing referrer_url query param on amp sdk track")
		}
	}

	eventProperties := U.PropertiesMap{}
	eventProperties[U.EP_PAGE_RAW_URL] = reqPayload.SourceURL
	eventProperties[U.EP_PAGE_URL] = pageURL
	eventProperties[U.EP_PAGE_DOMAIN] = parsedSourceURL.Host
	eventProperties[U.EP_PAGE_TITLE] = reqPayload.Title
	eventProperties[U.EP_REFERRER] = referrerRawURL
	eventProperties[U.EP_REFERRER_URL] = referrerURL
	eventProperties[U.EP_REFERRER_DOMAIN] = referrerDomain
	U.FillPropertiesFromURL(&eventProperties, parsedSourceURL)

	if reqPayload.PageLoadTimeInSecs > 0 {
		eventProperties[U.EP_PAGE_LOAD_TIME] = reqPayload.PageLoadTimeInSecs
	}
	userProperties := U.PropertiesMap{}
	if reqPayload.ScreenHeight > 0 {
		userProperties[U.UP_SCREEN_HEIGHT] = reqPayload.ScreenHeight
	}
	if reqPayload.ScreenWidth > 0 {
		userProperties[U.UP_SCREEN_WIDTH] = reqPayload.ScreenWidth
	}

	err = FillUserAgentUserProperties(&userProperties, reqPayload.UserAgent)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fill user agent user properties on amp track.")
	}

	trackPayload := TrackPayload{
		Auto:            true,
		UserId:          user.ID,
		IsNewUser:       isNewUser,
		Name:            pageURL,
		EventProperties: eventProperties,
		UserProperties:  userProperties,
		ClientIP:        reqPayload.ClientIP,
		UserAgent:       reqPayload.UserAgent,
		Timestamp:       reqPayload.Timestamp,
	}

	errCode, trackResponse := Track(project.ID, &trackPayload, false, SourceAMPSDK)
	if trackResponse.EventId != "" {
		cacheErrCode := SetCacheAMPSDKEventIDByPageURL(project.ID, user.ID,
			trackResponse.EventId, pageURL)
		if cacheErrCode != http.StatusAccepted {
			logCtx.Error("Failed to set cache event_id by page_url on AMP.")
		}
	} else {
		logCtx.WithFields(log.Fields{"user_id": user.ID, "event_id": trackResponse.EventId}).
			Error("Missing event_id from response of track on AMP track.")
	}

	return errCode, &Response{EventId: trackResponse.EventId,
		Message: trackResponse.Message, Error: trackResponse.Error}
}

func getAMPSDKByEventIDCacheKey(projectId uint64, userId string, pageURL string) (*cacheRedis.Key, error) {
	prefix := "amp_sdk_user_event"
	suffix := "uid:" + userId + ":url:" + pageURL
	return cacheRedis.NewKey(projectId, prefix, suffix)
}

func SetCacheAMPSDKEventIDByPageURL(projectId uint64, userId string, eventId string, pageURL string) int {
	logctx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	if projectId == 0 || userId == "" || eventId == "" || pageURL == "" {
		logctx.Error("Invalid scope ids.")
		return http.StatusBadRequest
	}

	resultCacheKey, err := getAMPSDKByEventIDCacheKey(projectId, userId, U.CleanURI(pageURL))
	if err != nil {
		logctx.WithError(err).Error("Failed to getAMPSDKByEventIdCacheKey.")
		return http.StatusNotFound
	}

	err = cacheRedis.Set(resultCacheKey, string(eventId), 60*60) // 60 mins
	if err != nil {
		logctx.WithError(err).Error("Failed to set cache amp sdk event_id by page_url.")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func GetCacheAMPSDKEventIDByPageURL(projectId uint64, userId string, pageURL string) (string, int) {
	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId).
		WithField("page_url", pageURL)

	var cacheResult string
	if projectId == 0 || userId == "" || pageURL == "" {
		return cacheResult, http.StatusBadRequest
	}

	resultCacheKey, err := getAMPSDKByEventIDCacheKey(projectId, userId, pageURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key on GetCacheAMPSDKEventIDByPageURL.")
		return cacheResult, http.StatusBadRequest
	}

	cacheResult, err = cacheRedis.Get(resultCacheKey)
	if err != nil {
		if err == redis.ErrNil {
			return cacheResult, http.StatusNotFound
		}
		return cacheResult, http.StatusInternalServerError
	}

	if cacheResult == "" {
		return cacheResult, http.StatusNotFound
	}

	return cacheResult, http.StatusFound
}

func AMPTrackWithQueue(token string, reqPayload *AMPTrackPayload,
	queueAllowedTokens []string) (int, *Response) {

	if excludeBotRequestBySetting(token, reqPayload.UserAgent) {
		return http.StatusNotModified,
			&Response{Message: "Track skipped. Bot request."}
	}

	if U.UseQueue(token, queueAllowedTokens) {
		err := enqueueRequest(token, sdkRequestTypeAMPEventTrack, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue amp sdk track event request.")
			return http.StatusInternalServerError, &Response{Error: "Track event failed"}
		}

		return http.StatusOK, &Response{Message: "Tracked event successfully"}
	}

	return AMPTrackByToken(token, reqPayload)
}

func AMPUpdateEventPropertiesWithQueue(token string, reqPayload *AMPUpdateEventPropertiesPayload,
	queueAllowedTokens []string) (int, *Response) {

	if excludeBotRequestBySetting(token, reqPayload.UserAgent) {
		return http.StatusNotModified,
			&Response{Message: "Update event properties skipped. Bot request."}
	}

	if U.UseQueue(token, queueAllowedTokens) {
		err := enqueueRequest(token, sdkRequestTypedAMPEventUpdateProperties, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue amp sdk update event request.")
			return http.StatusInternalServerError, &Response{Error: "Update event properties failed"}
		}

		return http.StatusOK, &Response{Message: "Updated event successfully"}
	}

	return AMPUpdateEventPropertiesByToken(token, reqPayload)
}

func FillUserAgentUserProperties(userProperties *U.PropertiesMap, userAgentStr string) error {
	if userAgentStr == "" {
		return errors.New("invalid user agent")
	}

	(*userProperties)[U.UP_USER_AGENT] = userAgentStr

	userAgent := user_agent.New(userAgentStr)
	(*userProperties)[U.UP_OS] = userAgent.OSInfo().Name
	(*userProperties)[U.UP_OS_VERSION] = userAgent.OSInfo().Version
	(*userProperties)[U.UP_OS_WITH_VERSION] = fmt.Sprintf("%s-%s",
		(*userProperties)[U.UP_OS], (*userProperties)[U.UP_OS_VERSION])

	if U.IsBotUserAgent(userAgentStr) {
		(*userProperties)[U.UP_BROWSER] = "Bot"
		return nil
	}

	browserName, browserVersion := userAgent.Browser()
	(*userProperties)[U.UP_BROWSER] = browserName
	(*userProperties)[U.UP_BROWSER_VERSION] = browserVersion
	(*userProperties)[U.UP_BROWSER_WITH_VERSION] = fmt.Sprintf("%s-%s",
		(*userProperties)[U.UP_BROWSER], (*userProperties)[U.UP_BROWSER_VERSION])

	dd := C.GetServices().DeviceDetector
	info := dd.Parse(userAgentStr)
	(*userProperties)[U.UP_DEVICE_BRAND] = info.Brand
	(*userProperties)[U.UP_DEVICE_TYPE] = info.Type
	(*userProperties)[U.UP_DEVICE_MODEL] = info.Model

	return nil
}
