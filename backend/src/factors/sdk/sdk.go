package sdk

import (
	"encoding/json"
	"errors"
	"factors/cache"
	pCache "factors/cache/persistent"
	cacheRedis "factors/cache/redis"
	"factors/company_enrichment/clearbit"
	"factors/company_enrichment/demandbase"
	"factors/company_enrichment/factors_deanon"
	"factors/company_enrichment/sixsignal"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"

	"factors/vendor_custom/machinery/v1/tasks"

	C "factors/config"
	"factors/metrics"
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
	Timestamp       int64           `json:"timestamp"`
	ProjectId       int64           `json:"project_id"`
	Auto            bool            `json:"auto"`
	ClientIP        string          `json:"client_ip"`
	UserAgent       string          `json:"user_agent"`
	SmartEventType  string          `json:"smart_event"`
	// source of the user record (1 = WEB, 2 = HUBSPOT, 3 = SALESFORCE)
	RequestSource int  `json:"request_source"`
	IsPast        bool `json:"is_past"`
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
	CustomerUserId string `json:"c_uid"`
	Timestamp      int64  `json:"timestamp"`

	UserProperties postgres.Jsonb `json:"user_properties"`

	CreateUser bool `json:"create_user"`
	// join_timestamp to use at the time of creating user,
	// if not provided, request timestamp will be used.
	JoinTimestamp int64 `json:"join_timestamp"`

	// identify overwrite info
	PageURL string `json:"page_url"`
	Source  string `json:"source"`
	// source of the user record (1 = WEB, 2 = HUBSPOT, 3 = SALESFORCE)
	RequestSource int `json:"request_source"`
}

// AMPIdentifyPayload holds required fields for AMP identification
type AMPIdentifyPayload struct {
	CustomerUserID string `json:"customer_user_id"`
	ClientID       string `json:"client_id"`
	Timestamp      int64  `json:"timestamp"`
	RequestSource  int    `json:"request_source"`
}

type IdentifyResponse struct {
	UserId  string `json:"user_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AddUserPropertiesPayload struct {
	UserId string `json:"user_id"`
	// if create_user is true, create user with given id.
	CreateUser    bool            `json:"create_user"`
	Timestamp     int64           `json:"timestamp"`
	Properties    U.PropertiesMap `json:"properties"`
	ClientIP      string          `json:"client_ip"`
	RequestSource int             `json:"request_source"`
}

type AddUserPropertiesResponse struct {
	UserId  string `json:"user_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UpdateEventPropertiesPayload struct {
	UserId        string          `json:"user_id"`
	EventId       string          `json:"event_id"`
	Properties    U.PropertiesMap `json:"properties"`
	Timestamp     int64           `json:"timestamp"`
	UserAgent     string          `json:"user_agent"`
	RequestSource int             `json:"request_source"`
	ClientIP      string          `json:"client_ip"`
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

type AccountPayload struct {
	Domain     string          `json:"domain"`
	Properties U.PropertiesMap `json:"properties,omitempty"`
}

type AccountTrackPayload struct {
	Domain string                   `json:"domain"`
	Event  AccountTrackPayloadEvent `json:"event"`
}

type AccountTrackPayloadEvent struct {
	Name       string          `json:"name"`
	Timestamp  int64           `json:"timestamp"`
	Properties U.PropertiesMap `json:"properties"`
}

const (
	SourceJSSDK  = "js_sdk"
	SourceAMPSDK = "amp_sdk"

	SourceSegment     = "segment"
	SourceShopify     = "shopify"
	SourceHubspot     = "hubspot"
	SourceSalesforce  = "salesforce"
	SourceRudderstack = "rudderstack"
)

// RequestQueue - Name of the primary queue which will
// be queued with sdk requests.
const RequestQueue = "sdk_request_queue_2"

// RequestQueueDuplicate - Name of the secondary Queue which
// will be queued with copy of tasks sent RequestQueue, if enabled.
const RequestQueueDuplicate = "sdk_request_queue"

// ProcessRequestTask - Name of the task which has been
// queued to request queues.
const ProcessRequestTask = "process_sdk_request"

const (
	sdkRequestTypeEventTrack               = "sdk_event_track"
	sdkRequestTypeUserIdentify             = "sdk_user_identify"
	sdkRequestTypeUserAddProperties        = "sdk_user_add_properties"
	sdkRequestTypeEventUpdateProperties    = "sdk_event_update_properties"
	sdkRequestTypeAMPEventTrack            = "sdk_amp_event_track"
	sdkRequestTypeAMPEventUpdateProperties = "sdk_amp_event_update_properties"
	sdkRequestTypeAMPIdentify              = "sdk_amp_identify"
)

func IsBot(ProjectId int64, userAgent, eventName string) bool {

	if !(U.IsBotUserAgent(userAgent) || U.IsBotEventByPrefix(eventName)) && userAgent != "" {
		if C.AllowDeviceServiceByProjectID(ProjectId) {
			deviceInfo, status, err := GetDeviceInfoFromDeviceService(userAgent)
			if err != nil || status != http.StatusOK {
				return false
			}
			return deviceInfo.IsBot
		}

	}
	return (U.IsBotUserAgent(userAgent) || U.IsBotEventByPrefix(eventName))

}

func ProcessQueueRequest(token, reqType, reqPayloadStr string) (float64, string, error) {
	// Todo(Dinesh): Retry on panic: Add payload back to queue as return
	// from defer is not possible and notify panic.

	// Todo(Dinesh): Add request_id for better tracing.

	execStartTime := time.Now()
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

	case sdkRequestTypeAMPEventUpdateProperties:
		var reqPayload AMPUpdateEventPropertiesPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = AMPUpdateEventPropertiesByToken(token, &reqPayload)

	case sdkRequestTypeAMPIdentify:
		var reqPayload AMPIdentifyPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}
		logCtx = logCtx.WithField("req_payload", reqPayload)

		status, response = AMPIdentifyByToken(token, &reqPayload)

	default:
		logCtx.Error("Invalid sdk request type on sdk process queue")
		return http.StatusInternalServerError, "", nil
	}

	responseBytes, _ := json.Marshal(response)
	logCtx = logCtx.WithField("status", status).WithField("response", string(responseBytes))

	// Do not retry on below conditions.
	if status == http.StatusBadRequest || status == http.StatusNotAcceptable || status == http.StatusUnauthorized {
		recordSDKRequestProcessedMetrics(reqType, execStartTime)
		return float64(status), "", nil
	}

	// Return error only for retry. Retry after a period till it is successfull.
	// Retry dependencies not found and failures which can be successful on retries.
	if status == http.StatusNotFound || status == http.StatusInternalServerError {
		metrics.Increment(C.GetSDKAndIntegrationMetricNameByConfig(metrics.IncrSDKRequestQueueRetry))
		return http.StatusInternalServerError, "",
			tasks.NewErrRetryTaskExp("EXP_RETRY__REQUEST_PROCESSING_FAILURE")
	}

	// Log for analysing queue process status.
	recordSDKRequestProcessedMetrics(reqType, execStartTime)

	return http.StatusOK, string(responseBytes), nil
}

func recordSDKRequestProcessedMetrics(requestType string, execStartTime time.Time) {
	metrics.Increment(C.GetSDKAndIntegrationMetricNameByConfig(metrics.IncrSDKRequestQueueProcessed))
	recordLatencyMetricByRequestType(requestType, execStartTime)
}

func IsValidTokenString(token string) bool {
	validString := token != "" && token != "undefined" && token != "null" && token != "Null"
	// Check esde
	validExpression := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(token)
	return validString && validExpression
}

func recordLatencyMetricByRequestType(requestType string, execStartTime time.Time) {
	var metricName string
	switch requestType {
	case sdkRequestTypeEventTrack:
		metricName = metrics.LatencySDKRequestTypeTrack
	case sdkRequestTypeAMPEventTrack:
		metricName = metrics.LatencySDKRequestTypeAMPTrack
	case sdkRequestTypeEventUpdateProperties:
		metricName = metrics.LatencySDKRequestTypeUpdateEventProperties
	case sdkRequestTypeAMPEventUpdateProperties:
		metricName = metrics.LatencySDKRequestTypeAMPUpdateEventProperties
	case sdkRequestTypeUserAddProperties:
		metricName = metrics.LatencySDKRequestTypeAddUserProperties
	case sdkRequestTypeUserIdentify:
		metricName = metrics.LatencySDKRequestTypeIdentifyUser
	case sdkRequestTypeAMPIdentify:
		metricName = metrics.LatencySDKRequestTypeAMPIdentifyUser
	default:
		log.WithField("type", requestType).
			Info("Invalid request type on record latency.")
		return
	}

	latencyInMs := time.Now().Sub(execStartTime).Milliseconds()
	metrics.RecordLatency(C.GetSDKAndIntegrationMetricNameByConfig(metricName), float64(latencyInMs))
}

func BackFillEventDataInCacheFromDb(project_id int64, currentTime time.Time, no_of_days int,
	eventsLimit, propertyLimit, valuesLimit int, rowsLimit int, perQueryPullRange int, skipExpiryForCache bool) {

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
		var eventNames model.CacheEventNamesWithTimestamp
		eventNames.EventNames = make(map[string]U.CountTimestampTuple)

		dateFormat := currentTime.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		eventNamesKey, err := model.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(project_id, dateFormat)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key")
			return
		}

		logCtx.WithField("dateFormat", dateFormat).Info("Begin: Event names - DB query by occurence")
		begin := U.TimeNowZ()
		events, err := store.GetStore().GetOrderedEventNamesFromDb(
			project_id,
			currentTime.AddDate(0, 0, -(i+perQueryPullRange)).Unix(),
			currentTime.AddDate(0, 0, -(i-1)).Unix(),
			eventsLimit)
		end := U.TimeNowZ()
		logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Event names - DB query by occurence")
		if err != nil {
			logCtx.WithError(err).Error("Failed to get values from DB - All event names")
			return
		}

		for item, element := range events {
			eventNames.EventNames[element.Name] = U.CountTimestampTuple{LastSeenTimestamp: int64(element.LastSeen), Count: element.Count}
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
		begin = U.TimeNowZ()
		err = pCache.Set(eventNamesKey, string(enEventCache), expiry, true)
		end = U.TimeNowZ()
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

			eventPropertyValuesInCache := make(map[*cache.Key]string)
			var eventProperties U.CachePropertyWithTimestamp
			eventProperties.Property = make(map[string]U.PropertyWithTimestamp)
			dateFormat := currentTime.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)

			logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event}).Info("Begin: Get event Properties DB call")
			begin := U.TimeNowZ()
			properties, err := store.GetStore().GetRecentEventPropertyKeysWithLimits(
				project_id, event,
				currentTime.AddDate(0, 0, -(i+perQueryPullRange)).Unix(),
				currentTime.AddDate(0, 0, -(i-1)).Unix(),
				propertyLimit)
			end := U.TimeNowZ()
			logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Get event Properties DB call")
			if err != nil {
				logCtx.WithError(err).Error("Failed to fetch values from DB - user properties")
				return
			}

			if len(properties) > 0 {
				eventPropertiesKey, err := model.GetPropertiesByEventCategoryRollUpCacheKey(project_id, event, dateFormat)
				if err != nil {
					logCtx.WithError(err).Error("Failed to get cache key - properties")
					return
				}

				for _, property := range properties {

					logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event, "property": property.Key}).Info("Begin: Get event Property values DB call")
					begin := U.TimeNowZ()
					values, category, err := store.GetStore().GetRecentEventPropertyValuesWithLimits(project_id, event, property.Key, valuesLimit, rowsLimit,
						currentTime.AddDate(0, 0, -(i+perQueryPullRange)).Unix(),
						currentTime.AddDate(0, 0, -(i-1)).Unix())
					end := U.TimeNowZ()
					logCtx.WithFields(log.Fields{"dateFormat": dateFormat, "event": event, "property": property.Key, "timeTaken": end.Sub(begin).Milliseconds()}).Info("End: Get event Property values DB call")
					if err != nil {
						logCtx.WithError(err).Error("Failed to get values from db - property values")
						return
					}

					categoryMap := make(map[string]int64)
					categoryMap[category] = property.Count
					eventProperties.Property[property.Key] = U.PropertyWithTimestamp{
						Category:          category,
						CategorywiseCount: categoryMap, // Setting precomputed ones to empty
						CountTime: U.CountTimestampTuple{
							LastSeenTimestamp: int64(property.LastSeen),
							Count:             property.Count}}

					var eventPropertyValues U.CachePropertyValueWithTimestamp
					eventPropertyValues.PropertyValue = make(map[string]U.CountTimestampTuple)

					if category == U.PropertyTypeCategorical {
						eventPropertyValuesKey, _ := model.GetValuesByEventPropertyRollUpCacheKey(project_id, event, property.Key, dateFormat)
						for _, value := range values {
							if value.Value != "" {
								eventPropertyValues.PropertyValue[value.Value] = U.CountTimestampTuple{
									LastSeenTimestamp: int64(value.LastSeen),
									Count:             value.Count}
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
				begin = U.TimeNowZ()
				err = pCache.SetBatch(eventPropertyValuesInCache, expiry, true)
				end = U.TimeNowZ()
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

func isPropertiesDefaultableTrackRequest(source string, isAutoTracked bool) bool {
	return isAutoTracked && (source == SourceJSSDK || source == SourceAMPSDK)
}

func Track(projectId int64, request *TrackPayload,
	skipSession bool, source string, objectType string) (int, *TrackResponse) {
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

	if request.EventProperties == nil {
		request.EventProperties = make(U.PropertiesMap, 0)
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

	projectSettings, errCode := store.GetStore().GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError, &TrackResponse{Error: "Tracking failed. Invalid project."}
	}

	// Terminate track calls from bot user_agent and event_name prefix.
	if *projectSettings.ExcludeBot && IsBot(projectId, request.UserAgent, request.Name) {
		return http.StatusNotModified, &TrackResponse{Message: "Tracking skipped. Bot request."}
	}

	var eventName *model.EventName
	var eventNameErrCode int

	// if auto_track enabled, auto_name = event_name else auto_name = "UCEN".
	// On 'auto' true event_name is the eventURL(e.g: factors.ai/u1/u2/u3) for JS_.
	if request.Auto {
		request.Name = strings.TrimSuffix(request.Name, "/")
		request.EventProperties[U.EP_IS_PAGE_VIEW] = true

		// Pass eventURL through filter and get corresponding event_name mapped by user.
		eventName, eventNameErrCode = store.GetStore().FilterEventNameByEventURL(projectId, request.Name)
		if eventName != nil && eventNameErrCode == http.StatusFound {
			err := model.FillEventPropertiesByFilterExpr(&request.EventProperties, eventName.FilterExpr, request.Name)
			if err != nil {
				log.WithFields(log.Fields{"project_id": projectId, "filter_expr": eventName.FilterExpr,
					"event_url": request.Name, log.ErrorKey: err}).Error(
					"Failed to fill event url properties for auto tracked event.")
			}
		} else {
			// create a auto tracked event name, if no filter_expr match.
			eventName, eventNameErrCode = store.GetStore().CreateOrGetAutoTrackedEventName(
				&model.EventName{Name: request.Name, ProjectId: projectId})
		}

	} else if request.SmartEventType != "" {
		request.Name = strings.TrimSuffix(request.Name, "/")
		eventName, eventNameErrCode = store.GetStore().GetSmartEventEventName(&model.EventName{Name: request.Name, ProjectId: projectId, Type: request.SmartEventType})
	} else if request.Name == U.EVENT_NAME_OFFLINE_TOUCH_POINT {
		eventName, eventNameErrCode = store.GetStore().CreateOrGetOfflineTouchPointEventName(projectId)
	} else {
		eventName, eventNameErrCode = store.GetStore().CreateOrGetUserCreatedEventName(
			&model.EventName{Name: request.Name, ProjectId: projectId})
	}

	if eventNameErrCode != http.StatusCreated && eventNameErrCode != http.StatusConflict &&
		eventNameErrCode != http.StatusFound {
		return eventNameErrCode, &TrackResponse{Error: "Tracking failed. Creating event_name failed."}
	}

	// Parsing URL params for all the event sources
	pageURL := getURLFromPageEvent(request.EventProperties)
	parsedPageURL, err := U.ParseURLStable(pageURL)
	if err == nil {
		_ = U.FillPropertiesFromURL(&request.EventProperties, parsedPageURL)

	}

	// Event Properties
	clientIP := request.ClientIP
	U.UnEscapeQueryParamProperties(&request.EventProperties)
	definedEventProperties, hasDefinedMarketingProperty := MapEventPropertiesToProjectDefinedProperties(projectId,
		logCtx, &request.EventProperties)
	eventProperties := U.GetValidatedEventProperties(definedEventProperties)
	if ip, ok := (request.EventProperties)[U.EP_INTERNAL_IP]; ok && ip != "" {
		clientIP = ip.(string)
	}

	var isClickedEmailEvent bool
	var clickedEmailEventName *model.EventName

	if C.IsEmailUTMParameterAllowed(projectId) {
		U.UnEscapeQueryParamProperties(&request.EventProperties)
		if U.GetPropertyValueAsString((*eventProperties)[U.EP_EMAIL]) != "" {
			isClickedEmailEvent = true

			clickedEmailEventName, errCode = store.GetStore().CreateOrGetUserCreatedEventName(
				&model.EventName{Name: U.EVENT_NAME_CLICKED_EMAIL, ProjectId: projectId})

			if errCode != http.StatusCreated && errCode != http.StatusConflict &&
				errCode != http.StatusFound {
				isClickedEmailEvent = false
				logCtx.WithField("event_name", U.EVENT_NAME_CLICKED_EMAIL).Warn("Tracking failed. Creating  ClickedEmail Name  failed.")
			}
		}
	}

	U.SanitizeProperties(eventProperties)

	response := &TrackResponse{}

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		errCode := store.GetStore().IsUserExistByID(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

	var existingUserProperties *map[string]interface{}
	if request.CreateUser || request.UserId == "" {
		newUser := &model.User{ProjectId: projectId, Source: &request.RequestSource}

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
		createdUserID, errCode := store.GetStore().CreateUser(newUser)
		if errCode != http.StatusCreated {
			return errCode, &TrackResponse{Error: "Tracking failed. User creation failed."}
		}

		request.UserId = createdUserID
		response.UserId = createdUserID
	} else {
		// Adding initial user properties if user_id exists,
		// but initial properties are not. i.e user created on identify.
		existingUserProperties, errCode = store.GetStore().GetLatestUserPropertiesOfUserAsMap(projectId, request.UserId)
		if errCode != http.StatusFound {
			logCtx.WithField("user_id", errCode).WithField("err_code",
				errCode).Error("Tracking failed. Get user properties as map failed.")
			return errCode, &TrackResponse{Error: "Tracking failed while getting user."}
		}
	}

	newUserPropertiesMap := make(U.PropertiesMap, 0)
	userProperties := &newUserPropertiesMap
	FillUserAgentUserProperties(projectId, userProperties, request.UserAgent)
	U.FillInitialUserProperties(userProperties, request.EventId, eventProperties,
		existingUserProperties, isPropertiesDefaultableTrackRequest(source, request.Auto))

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

	_, _, isoCode := model.FillLocationUserProperties(userProperties, clientIP)
	pageURLProp := U.GetPropertyValueAsString((*eventProperties)[U.EP_PAGE_URL])

	if clientIP != "" {
		FillCompanyIdentificationUserProperties(projectId, clientIP, projectSettings, userProperties, eventProperties, request.UserId, isoCode, pageURLProp)
	}

	if C.EnableSixSignalGroupByProjectID(projectId) {
		groupProperties := U.FilterPropertiesByKeysByPrefix(userProperties, U.SIX_SIGNAL_PROPERTIES_PREFIX)
		if groupProperties != nil && len(*groupProperties) > 0 {
			status := TrackUserAccountGroup(projectId, request.UserId, U.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
			if status != http.StatusOK && status != http.StatusNotModified && status != http.StatusNotImplemented {
				logCtx.Error("Failed to TrackUserGroup.")
			}
		}
	}

	if C.IsEnableDebuggingForIP() {
		log.WithFields(log.Fields{"userProperties": userProperties,
			"eventProperties": eventProperties, "ip": clientIP}).Info("Debugging ip enrichment.")
	}

	// Add latest user properties without UTM parameters.
	U.FillLatestPageUserProperties(userProperties, eventProperties)
	// Add latest touch user properties.
	if hasDefinedMarketingProperty {
		U.FillLatestTouchUserProperties(userProperties, eventProperties)
	}

	// Identify users with form events.
	isFormEvent := request.Name == U.EVENT_NAME_FORM_SUBMITTED
	if C.IsFormFillIdentificationAllowedForProject(projectId) {
		isFormEvent = request.Name == U.EVENT_NAME_FORM_SUBMITTED || request.Name == U.EVENT_NAME_FORM_FILL
	}

	if C.IsEmailUTMParameterAllowed(projectId) {
		isFormEvent = isFormEvent || isClickedEmailEvent
	}

	if isFormEvent {
		customerUserID, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(
			projectId, request.UserId, eventProperties)
		if errCode == http.StatusInternalServerError {
			log.WithFields(log.Fields{"userProperties": userProperties,
				"eventProperties": eventProperties}).Error(
				"Failed adding user properties from form submitted event.")
			response.Error = "Failed adding user properties."
		}

		if customerUserID != "" {

			pageURL := U.GetPropertyValueAsString((*eventProperties)[U.EP_PAGE_URL])

			errCode, _ := Identify(projectId, &IdentifyPayload{
				UserId:         request.UserId,
				CustomerUserId: customerUserID,
				Timestamp:      request.Timestamp,
				PageURL:        pageURL,
				Source:         sdkRequestTypeEventTrack,
				RequestSource:  request.RequestSource,
			}, request.Name == U.EVENT_NAME_FORM_SUBMITTED)
			if errCode != http.StatusOK {
				log.WithFields(log.Fields{"projectId": projectId, "userId": request.UserId,
					"customerUserId": customerUserID}).Error("Failed to identify user on form submit event.")
			}

			// Add user properties from form submit event properties.
			if errCode == http.StatusOK {
				for k, v := range *formSubmitUserProperties {
					(*userProperties)[k] = v
				}
			}
		}
	}

	if existingUserProperties == nil {
		existingUserProperties, errCode = store.GetStore().GetLatestUserPropertiesOfUserAsMap(projectId, request.UserId)
		if errCode == http.StatusInternalServerError {
			logCtx.WithField("user_id", errCode).WithField("err_code",
				errCode).Error("Tracking failed. Get user properties as map failed.")
		}
	}

	err = U.FillFirstEventUserPropertiesIfNotExist(existingUserProperties, userProperties, request.Timestamp)
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to fill day of first event user_properties on track.")
	}

	logCtx = logCtx.WithField("user_properties", userProperties)
	userPropsJSON, err := json.Marshal(userProperties)
	if err != nil {
		logCtx.WithError(err).Error("Update user proprieties on track failed. JSON marshal failed.")
		response.Error = "Failed updating user properties."
	}

	newUserPropertiesJSON := &postgres.Jsonb{RawMessage: userPropsJSON}
	userPropertiesV2, errCode := store.GetStore().UpdateUserPropertiesV2(
		projectId, request.UserId, newUserPropertiesJSON, request.Timestamp, source, objectType)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		logCtx.WithField("err_code", errCode).
			Error("Update user properties on track failed. DB update failed.")
		response.Error = "Failed updating user properties."
	}

	if eventName == nil {
		logCtx.WithFields(log.Fields{"request_payload": request}).Error("Empty event name found. Skipping processing.")
		response.Error = "Failed processing event name."
		return http.StatusInternalServerError, &TrackResponse{Error: "Tracking failed. Failed processing event name."}
	}

	event := &model.Event{
		ID:              request.EventId,
		EventNameId:     eventName.ID,
		CustomerEventId: request.CustomerEventId,
		Timestamp:       request.Timestamp,
		ProjectId:       projectId,
		UserId:          request.UserId,
		// UserProperties - Computed using properties on users table.
		UserProperties: userPropertiesV2,
	}

	// Property used as flag for skipping session on offline session worker.
	if skipSession {
		(*eventProperties)[U.EP_SKIP_SESSION] = U.PROPERTY_VALUE_TRUE
	}

	if isPropertiesDefaultableTrackRequest(source, request.Auto) {
		U.SetDefaultValuesToEventProperties(eventProperties)
	}

	eventPropsJSON, err := json.Marshal(eventProperties)
	if err != nil {
		return http.StatusBadRequest, &TrackResponse{Error: "Tracking failed. Invalid properties."}
	}
	event.Properties = postgres.Jsonb{RawMessage: eventPropsJSON}

	if C.PastEventEnrichmentEnabled(projectId) && U.IsCRM(source) && request.IsPast {
		event.IsFromPast = true

		userProperties, err := U.EncodeToPostgresJsonb((*map[string]interface{})(&request.UserProperties))
		if err != nil {
			userProperties = &postgres.Jsonb{}
		}
		event.UserProperties = userProperties
	}

	createdEvent, errCode := store.GetStore().CreateEvent(event)
	if errCode == http.StatusNotAcceptable {
		return errCode, &TrackResponse{Error: "Tracking failed. Event creation failed. Invalid payload.",
			CustomerEventId: request.CustomerEventId}
	} else if errCode != http.StatusCreated {
		return errCode, &TrackResponse{Error: "Tracking failed. Event creation failed."}
	}

	if isClickedEmailEvent {
		clickedEmailevent := &model.Event{
			EventNameId:    clickedEmailEventName.ID,
			Timestamp:      request.Timestamp,
			ProjectId:      projectId,
			UserId:         request.UserId,
			UserProperties: userPropertiesV2,
			Properties:     event.Properties,
		}

		_, errCode := store.GetStore().CreateEvent(clickedEmailevent)
		if errCode != http.StatusCreated {
			logCtx.Warn("Tracking failed. Clicked email Event creation failed.")
		}

	}

	// Success response.
	response.EventId = createdEvent.ID
	response.Message = "User event tracked successfully."
	response.CustomerEventId = request.CustomerEventId
	return http.StatusOK, response
}

func FillCompanyIdentificationUserProperties(projectId int64, clientIP string, projectSettings *model.ProjectSetting, userProperties *U.PropertiesMap, eventProperties *U.PropertiesMap, userId string, isoCode string, pageUrl string) {

	logCtx := log.WithFields(log.Fields{
		"project_id": projectId,
		"logId":      fmt.Sprintf("%v+%v", projectId, clientIP)})

	var customerDemandbase demandbase.CustomerDemandbase
	var customerClearbit clearbit.CustomerClearbit
	var customerSixSignal sixsignal.CustomerSixSignal
	var factorsDeanon factors_deanon.FactorsDeanon

	if enrichByCustomerDemandbase, _ := customerDemandbase.IsEligible(projectSettings, logCtx); enrichByCustomerDemandbase {
		if _, ok := customerDemandbase.Enrich(projectSettings, userProperties, userId, clientIP, logCtx); ok == 1 {
			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(projectId, model.FEATURE_DEMANDBASE, model.SUCCESS)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": projectId}).Warn("Failed to update integration status")

			}
		}
	} else if enrichByCustomerClearbit, _ := customerClearbit.IsEligible(projectSettings, logCtx); enrichByCustomerClearbit {
		if _, ok := customerClearbit.Enrich(projectSettings, userProperties, userId, clientIP, logCtx); ok == 1 {
			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(projectId, model.FEATURE_CLEARBIT, model.SUCCESS)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": projectId}).Warn("Failed to update integration status")

			}
		}
	} else if enrichByCustomerSixSignal, _ := customerSixSignal.IsEligible(projectSettings, logCtx); enrichByCustomerSixSignal {
		if _, ok := customerSixSignal.Enrich(projectSettings, userProperties, userId, clientIP, logCtx); ok == 1 {
			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(projectId, model.FEATURE_SIX_SIGNAL, model.SUCCESS)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": projectId}).Warn("Failed to update integration status")

			}
		}
	} else if enrichByFactorsDeanon, _ := factorsDeanon.IsEligible(projectSettings, isoCode, pageUrl, logCtx); enrichByFactorsDeanon {
		domain, status := factorsDeanon.Enrich(projectSettings, userProperties, eventProperties, userId, clientIP, logCtx)
		if status == 1 {
			factorsDeanon.Meter(projectId, domain, logCtx)

			status := store.GetStore().UpdateProjectSettingsIntegrationStatus(projectId, model.FEATURE_FACTORS_DEANONYMISATION, model.SUCCESS)
			if status != http.StatusAccepted {
				log.WithFields(log.Fields{"project_id": projectId}).Warn("Failed to update integration status")

			}
		}

		errCode, err := factorsDeanon.HandleAccountLimitAlert(projectId, &http.Client{}, logCtx)
		if errCode != http.StatusOK && errCode != http.StatusForbidden {
			logCtx.WithField("error", err).Error("Failed to send account limit alert.")
		}
	} else if isDeanonQuotaAvailable, err := factors_deanon.CheckingFactorsDeanonQuotaLimit(projectId); err == nil && !isDeanonQuotaAvailable {
		status := store.GetStore().UpdateProjectSettingsIntegrationStatus(projectId, model.FEATURE_FACTORS_DEANONYMISATION, model.LIMIT_EXCEED)
		if status != http.StatusAccepted {
			log.WithFields(log.Fields{"project_id": projectId}).Warn("Failed to update integration status")
		}
	}

}

func getURLFromPageEvent(properties U.PropertiesMap) string {

	url, exists := properties["url"]
	if exists && url != nil {
		return url.(string)
	}
	url, exists = properties["$page_raw_url"]
	if exists && url != nil {
		return url.(string)
	}
	url, exists = properties["URL"]
	if exists && url != nil {
		return url.(string)
	}
	url, exists = properties["page_url"]
	if exists && url != nil {
		return url.(string)
	}
	return ""
}

type Rank struct {
	Rank  int
	Value string
}

func MapEventPropertiesToProjectDefinedProperties(projectID int64, logCtx *log.Entry, properties *U.PropertiesMap) (*U.PropertiesMap, bool) {
	mappedProperties := make(U.PropertiesMap)
	interactionSettings := model.InteractionSettings{}

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode == http.StatusFound {
		err := U.DecodePostgresJsonbToStructType(&project.InteractionSettings, &interactionSettings)
		if err != nil && err.Error() != "Empty jsonb object" {
			logCtx.WithField("projectID", projectID).WithField("err", err).Error("failed to Decode Postgres Jsonb")
		}
	} else {
		logCtx.WithField("projectID", projectID).WithField("err_code", errCode).Error("failed to fetch project")
	}

	// use default settings in case not provided
	if interactionSettings.UTMMappings == nil {
		interactionSettings = model.DefaultMarketingPropertiesMap()
	}

	ApplyRanking(interactionSettings, properties, &mappedProperties)
	return &mappedProperties, U.HasDefinedMarketingProperty(&mappedProperties)
}

func ApplyRanking(interactionSettings model.InteractionSettings, properties *U.PropertiesMap, mappedProperties *U.PropertiesMap) {

	// build a reverse map. Value = Standard Property; Rank = Key's value
	reverseMarketingTouchPoints := make(map[string]Rank)
	for k, v := range interactionSettings.UTMMappings {
		for rank, userDefinedTouchPoint := range v {
			// lower the rank, higher the priority
			reverseMarketingTouchPoints[userDefinedTouchPoint] = Rank{rank, k}
		}
	}

	// the rank tracker
	rankTracker := make(map[string]int)
	for k, v := range *properties {
		var property string
		if _, stdKeyExists := reverseMarketingTouchPoints[k]; stdKeyExists {
			if _, rankExists := rankTracker[reverseMarketingTouchPoints[k].Value]; rankExists {
				newRank := reverseMarketingTouchPoints[k].Rank
				existingRank := rankTracker[reverseMarketingTouchPoints[k].Value]
				if newRank > existingRank {
					// found a lower ranked query param
					continue
				}
				property = reverseMarketingTouchPoints[k].Value
				rankTracker[reverseMarketingTouchPoints[k].Value] = newRank
			} else {
				property = reverseMarketingTouchPoints[k].Value
				rankTracker[reverseMarketingTouchPoints[k].Value] = reverseMarketingTouchPoints[k].Rank
			}
		} else {
			property = k

		}
		(*mappedProperties)[property] = v
	}
}

func isUserAlreadyIdentifiedBySDKRequest(projectID int64, userID string) bool {
	userProperties, status := store.GetStore().GetLatestUserPropertiesOfUserAsMap(projectID, userID)
	if status != http.StatusFound {
		return false
	}

	metaObj, err := model.GetDecodedUserPropertiesIdentifierMetaObject(userProperties)
	if err != nil {
		return false
	}

	for _, customerUserIDMeta := range *metaObj {
		if customerUserIDMeta.Source == sdkRequestTypeUserIdentify {
			return true
		}
	}

	return false
}

func allowedCustomerUserIDSourceIdentificationOverwrite(incomingCustomerUseridSource, existingCustomerUseridSource int) bool {
	if existingCustomerUseridSource == model.UserSourceWeb &&
		incomingCustomerUseridSource == model.UserSourceWeb {
		return true
	}

	if model.IsUserSourceCRM(existingCustomerUseridSource) && model.IsUserSourceCRM(incomingCustomerUseridSource) && incomingCustomerUseridSource == existingCustomerUseridSource {
		return true
	}

	if existingCustomerUseridSource == model.UserSourceWeb && model.IsUserSourceCRM(incomingCustomerUseridSource) {
		return true
	}

	return false
}

func ShouldAllowIdentificationOverwrite(projectID int64, userID string,
	incomingCustomerUserid string, incomingRequestSource int, incomingSource string) bool {

	// sdk indentify source always overwrite the customer_user_id
	if incomingSource == sdkRequestTypeUserIdentify {
		return true
	}

	if isUserAlreadyIdentifiedBySDKRequest(projectID, userID) {
		return false
	}

	user, status := store.GetStore().GetUserWithoutJSONColumns(projectID, userID)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "customer_user_id": incomingCustomerUserid}).
				Error("Failed to get user for identification overwrite decicision. Allowing identification overwrite.")
		}
		return true
	}

	if user.CustomerUserId == "" {
		return true
	}

	if incomingRequestSource == model.UserSourceWeb {
		// Same CustomerUserId existing on other source should block overwrite
		_, status = store.GetStore().GetExistingUserByCustomerUserID(projectID, []string{user.CustomerUserId}, model.GetAllCRMUserSource()...)
		if status == http.StatusFound {
			return false
		}

		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "customer_user_id": incomingCustomerUserid}).
				Error("Failed to get user by customer user id and source.")
			return false
		}

		if user.CustomerUserIdSource != nil &&
			allowedCustomerUserIDSourceIdentificationOverwrite(incomingRequestSource, *user.CustomerUserIdSource) {
			return true
		}

		// allow identification once for users who don't have customer_user_id_source. It will allow setting customer_user_id_source
		if user.CustomerUserIdSource == nil {
			return true
		}

		return false
	}

	/*
		CRM flow
	*/
	if user.CustomerUserIdSource != nil &&
		allowedCustomerUserIDSourceIdentificationOverwrite(incomingRequestSource, *user.CustomerUserIdSource) {
		return true
	}

	// Same CustomerUserId existing on crm source should block overwrite. If the existing customer_user_id exists in any other crm then it will be blocked
	// For updating another crm user or web user taken over by another crm, that crm should tell to overwrite.
	_, status = store.GetStore().GetExistingUserByCustomerUserID(projectID, []string{user.CustomerUserId}, model.GetAllCRMUserSource()...)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "customer_user_id": incomingCustomerUserid}).
				Error("Failed to get user by customer user id and source.")
			return false
		}
		return true
	}

	// allow identification once for users who don't have customer_user_id_source. It will allow setting customer_user_id_source and taking priority
	if user.CustomerUserIdSource == nil {
		return true
	}

	return false
}

/*
Identify :-
If overwrite is false

	user will be identified once and customer_user_id will be set as per source

If overwrite is true

	customer_user_id_source will be set/updated when user is re-identified
		identification from sdk identfy source will always overwrite
		if user only identified in web then it would continue to re-identify
		if user identified in crm, web identification will be blocked
		crm source will be allowed to overwrite in all cases
*/
func Identify(projectId int64, request *IdentifyPayload, overwrite bool) (int, *IdentifyResponse) {
	// Precondition: Fails to identify if customer_user_id not present.
	if request.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		return http.StatusBadRequest, &IdentifyResponse{
			Error: "Identification failed. Missing mandatory keys c_uid."}
	}

	if projectId == 588 && strings.HasSuffix(request.CustomerUserId, "@mailmodo.com") {
		log.Warn("Identification failed. Blocked identification.")
		return http.StatusBadRequest, &IdentifyResponse{
			Error: "Blocked identification."}
	}

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	// if join_timestamp is not provided, use request
	// timestamp as user's join_timestamp during creation.
	if request.JoinTimestamp == 0 {
		request.JoinTimestamp = request.Timestamp
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"user_id": request.UserId, "customer_user_id": request.CustomerUserId})

	userProperties, err := model.GetIdentifiedUserPropertiesAsJsonb(request.CustomerUserId)
	if err != nil || userProperties == nil {
		logCtx.WithError(err).Error("Failed to get and add identified user properties on identify.")
	}

	allowSupportForUserPropertiesInIdentityCall := C.AllowSupportForUserPropertiesInIdentifyCall(projectId)

	if allowSupportForUserPropertiesInIdentityCall {
		incomingProperties, err := U.ConvertPostgresJSONBToMap(request.UserProperties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to convert Postgres JSONB object to Map.")
		}

		userProperties, err = U.AddToPostgresJsonb(userProperties, incomingProperties, true)
		if err != nil {
			logCtx.WithError(err).Error("Failed to merge incoming user properties with existing user properties.")
		}
	}

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		errCode := store.GetStore().IsUserExistByID(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

	// Create new user with given user_id and customer_user_id,
	// if the create user_id is set to true.
	if request.CreateUser {
		if request.UserId == "" {
			logCtx.Error("Identify request payload with create_user true without user_id.")
			return http.StatusInternalServerError,
				&IdentifyResponse{Error: "Identification failed. User creation failed."}
		}

		response := &IdentifyResponse{}

		newUser := model.User{
			ID:             request.UserId,
			ProjectId:      projectId,
			CustomerUserId: request.CustomerUserId,
			JoinTimestamp:  request.JoinTimestamp,
			Source:         &request.RequestSource,
		}

		if C.AllowIdentificationOverwriteUsingSource(projectId) {
			newUser.CustomerUserIdSource = &request.RequestSource
		}

		if overwrite {
			if request.Source == "" {
				logCtx.WithFields(log.Fields{"userId": request.UserId, "customerUserId": request.CustomerUserId}).Error("Source missing in indentify overwrite.")
			}

			err := store.GetStore().UpdateIdentifyOverwriteUserPropertiesMeta(projectId, request.CustomerUserId, request.UserId, request.PageURL,
				request.Source, userProperties, request.Timestamp, request.CreateUser)
			if err != nil {
				logCtx.WithFields(log.Fields{"userId": request.UserId,
					"customerUserId": request.CustomerUserId}).WithError(err).Error("Failed to add identify overwrite meta")
			}
		}

		if userProperties != nil {
			newUser.Properties = *userProperties
		}

		_, errCode := store.GetStore().CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, &IdentifyResponse{
				Error: "Identification failed. User creation failed."}
		}

		response.UserId = request.UserId

		response.Message = "User has been identified successfully."
		return http.StatusOK, response
	}

	// If identified without userID, try to re-use existing user of
	// customer_user_id, else create a new user. This is possible only
	// on non-queue requests. For queue requests, either create_user is
	// set to true or the user_id will be present.
	if request.UserId == "" {
		response := &IdentifyResponse{}

		userLatest, errCode := store.GetStore().GetUserLatestByCustomerUserId(projectId, request.CustomerUserId, request.RequestSource)
		switch errCode {
		case http.StatusInternalServerError:
			return errCode, &IdentifyResponse{
				Error: "Identification failed. Processing without user_id failed."}

		case http.StatusNotFound:
			newUser := model.User{
				ProjectId:      projectId,
				CustomerUserId: request.CustomerUserId,
				JoinTimestamp:  request.JoinTimestamp,
				Source:         &request.RequestSource,
			}

			if userProperties != nil {
				newUser.Properties = *userProperties
			}

			_, errCode := store.GetStore().CreateUser(&newUser)
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

	existingCustomerUserId, errCode := store.GetStore().GetCustomerUserIdFromUserId(projectId, request.UserId)
	if errCode != http.StatusFound {
		return errCode, &IdentifyResponse{Error: "Identification failed. Invalid user_id."}
	}

	// Precondition: Given user already identified as given customer_user.
	if existingCustomerUserId == request.CustomerUserId {
		return http.StatusOK, &IdentifyResponse{Message: "Identified already."}
	}

	if overwrite {
		if !C.AllowIdentificationOverwriteUsingSource(projectId) {
			// avoid overwrite if user was identified by sdk request identify
			if isUserAlreadyIdentifiedBySDKRequest(projectId, request.UserId) {
				if request.Source != sdkRequestTypeUserIdentify {
					overwrite = false
				}
			}
		} else {
			// avoid overwrite if user was identified by sdk request identify or crm source
			if !ShouldAllowIdentificationOverwrite(projectId, request.UserId, request.CustomerUserId, request.RequestSource, request.Source) {
				logCtx.WithFields(log.Fields{"project_id": projectId, "user_id": request.UserId, "customer_user_id": request.CustomerUserId,
					"source": request.Source, "request_source": request.RequestSource}).Info("Overwriting identification blocked.")
				overwrite = false
			}
		}

	}

	// Precondition: user is already identified with different customer_user.
	// Creating a new user with the given customer_user_id and respond with new_user_id.
	if existingCustomerUserId != "" && !overwrite {
		newUser := model.User{
			ProjectId:      projectId,
			CustomerUserId: request.CustomerUserId,
			JoinTimestamp:  request.JoinTimestamp,
			Source:         &request.RequestSource,
		}

		if C.AllowIdentificationOverwriteUsingSource(projectId) {
			newUser.CustomerUserIdSource = &request.RequestSource
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

		createdUserID, errCode := store.GetStore().CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, &IdentifyResponse{Error: "Identification failed. User creation failed."}
		}

		if C.EnableTotalSessionPropertiesV2ByProjectID(projectId) {
			status := store.GetStore().UpdateSessionProperties(projectId, newUser.CustomerUserId, "")
			if status != http.StatusOK {
				logCtx.Error("Failed to update total session property on identiy create user.")
			}
		}

		if C.EnableUserDomainsGroupByProjectID(projectId) {
			status := store.GetStore().AssociateUserDomainsGroup(projectId, createdUserID, "", "")
			if status != http.StatusOK && status != http.StatusNotModified {
				log.WithFields(log.Fields{"project_id": projectId, "user_id": createdUserID, "err_code": status}).
					Error("Failed to AssociateUserDomainsGroup on Identify create user.")
			}
		}

		return http.StatusOK, &IdentifyResponse{UserId: createdUserID,
			Message: "User has been identified successfully"}

	}

	if overwrite {
		err := store.GetStore().UpdateIdentifyOverwriteUserPropertiesMeta(projectId, request.CustomerUserId, request.UserId, request.PageURL, request.Source, userProperties, request.Timestamp, request.CreateUser)
		if err != nil {
			logCtx.WithFields(log.Fields{"userId": request.UserId,
				"customerUserId": request.CustomerUserId}).WithError(err).Error("Failed to add identify overwrite meta")
		}
	}

	// Happy path. Maps customer_user to an user.
	updateUser := &model.User{CustomerUserId: request.CustomerUserId}
	if userProperties != nil {
		updateUser.Properties = *userProperties
	}

	if C.AllowIdentificationOverwriteUsingSource(projectId) {
		updateUser.CustomerUserIdSource = &request.RequestSource
	}

	_, errCode = store.GetStore().UpdateUser(projectId, request.UserId, updateUser, request.Timestamp)
	if errCode != http.StatusAccepted {
		return errCode, &IdentifyResponse{Error: "Identification failed. Failed mapping customer_user to user"}
	}

	if C.EnableTotalSessionPropertiesV2ByProjectID(projectId) {
		status := store.GetStore().UpdateSessionProperties(projectId, request.CustomerUserId, "")
		if status != http.StatusOK {
			logCtx.WithFields(log.Fields{"new_customer_user_id": request.CustomerUserId}).
				Error("Failed to update total session property on identiy new customer user id .")
		}

		if existingCustomerUserId != "" {
			status = store.GetStore().UpdateSessionProperties(projectId, existingCustomerUserId, "")
			if status != http.StatusOK {
				logCtx.WithFields(log.Fields{"existing_customer_user_id": existingCustomerUserId}).
					Error("Failed to update total session property on identiy existing customer user id.")
			}
		}
	}

	if C.EnableUserDomainsGroupByProjectID(projectId) {
		status := store.GetStore().AssociateUserDomainsGroup(projectId, request.UserId, "", "")
		if status != http.StatusOK && status != http.StatusNotModified {
			log.WithFields(log.Fields{"project_id": projectId, "user_id": request.UserId, "err_code": status}).
				Error("Failed to AssociateUserDomainsGroup on user Identify.")
		}
	}

	return http.StatusOK, &IdentifyResponse{Message: "User has been identified successfully."}
}

func AddUserProperties(projectId int64,
	request *AddUserPropertiesPayload) (int, *AddUserPropertiesResponse) {

	logCtx := log.WithField("project_id", projectId)

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	// Validate properties.
	validProperties := U.GetValidatedUserProperties(&request.Properties)

	_, _, _ = model.FillLocationUserProperties(validProperties, request.ClientIP)
	propertiesJSON, err := json.Marshal(validProperties)
	if err != nil {
		return http.StatusBadRequest,
			&AddUserPropertiesResponse{Error: "Add user properties failed. Invalid properties."}
	}

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		errCode := store.GetStore().IsUserExistByID(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

	// Precondition: user_id not given.
	if request.CreateUser || request.UserId == "" {
		newUser := &model.User{
			ProjectId:  projectId,
			Properties: postgres.Jsonb{propertiesJSON},
			Source:     &request.RequestSource,
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
		createdUserID, errCode := store.GetStore().CreateUser(newUser)
		if errCode != http.StatusCreated {
			return errCode, &AddUserPropertiesResponse{Error: "Add user properties failed. User create failed"}
		}

		return http.StatusOK, &AddUserPropertiesResponse{UserId: createdUserID,
			Message: "Added user properties successfully."}
	}

	errCode := store.GetStore().IsUserExistByID(projectId, request.UserId)
	if errCode == http.StatusNotFound {
		return http.StatusBadRequest,
			&AddUserPropertiesResponse{Error: "Add user properties failed. Invalid user_id."}
	} else if errCode == http.StatusInternalServerError {
		return errCode,
			&AddUserPropertiesResponse{Error: "Add user properties failed"}
	}

	_, errCode = store.GetStore().UpdateUserProperties(projectId, request.UserId,
		&postgres.Jsonb{propertiesJSON}, request.Timestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return errCode,
			&AddUserPropertiesResponse{Error: "Add user properties failed."}
	}

	return http.StatusOK,
		&AddUserPropertiesResponse{Message: "Added user properties successfully."}
}

func enqueueRequest(token, reqType string, reqPayload interface{}) error {
	logCtx := log.WithField("token", token).WithField("payload", reqPayload)

	taskSignature, err := util.CreateTaskSignatureForQueue(ProcessRequestTask,
		RequestQueue, token, reqType, reqPayload)
	if err != nil {
		return err
	}

	queueClient := C.GetServices().QueueClient
	_, err = queueClient.SendTask(taskSignature)
	if err != nil {
		return err
	}

	if !C.IsSDKAndIntegrationRequestQueueDuplicationEnabled() {
		return nil
	}

	duplicateTaskSignature, err := util.CreateTaskSignatureForQueue(ProcessRequestTask,
		RequestQueueDuplicate, token, reqType, reqPayload)
	if err != nil {
		return err
	}

	duplicateQueueClient := C.GetServices().DuplicateQueueClient
	_, err = duplicateQueueClient.SendTask(duplicateTaskSignature)
	if err != nil {
		// Log and return duplicate task queue failure.
		// To avoid track failure response to the clients.
		logCtx.WithError(err).Error("Failed to send task to the duplicate queue.")
		return nil
	}

	return nil
}

func excludeBotRequestBySetting(token, userAgent string, eventName string, clientIP string) (excluded bool) {
	defer func() {
		if excluded == true {
			metrics.Increment(metrics.IncrSDKRquestQueueExcludedBot)
		}
	}()

	if clientIP != "" && C.IsExcludeBotIPV4AddressByRange(clientIP) {
		return true
	}

	settings, errCode := store.GetStore().GetProjectSettingByTokenWithCacheAndDefault(token)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).
			Error("Failed to get project settings on excludeBotRequestBeforeQueue.")
		return false
	}

	return settings != nil && *settings.ExcludeBot && IsBot(settings.ProjectId, userAgent, eventName)
}

func TrackByToken(token string, reqPayload *TrackPayload) (int, *TrackResponse) {
	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode == http.StatusFound {
		return Track(projectID, reqPayload, false, SourceJSSDK, "")
	}

	if errCode == http.StatusNotFound {
		logCtx := log.WithField("token", token).WithField("request_payload", reqPayload)
		if IsValidTokenString(token) {
			// This is tracked with a metric on the dashboard as it could cause error spikes.
			logCtx.WithField("tag", "invalid_sdk_token").Info("Failed to get project from sdk project token.")
		} else {
			log.WithField("token", token).Warn("Invalid token on sdk payload.")
		}

		return http.StatusUnauthorized,
			&TrackResponse{Error: "Tracking failed. Invalid token."}
	}

	return errCode, &TrackResponse{Error: "Tracking failed."}
}

func TrackWithQueue(token string, reqPayload *TrackPayload,
	queueAllowedTokens []string) (int, *TrackResponse) {

	if excludeBotRequestBySetting(token, reqPayload.UserAgent, reqPayload.Name, reqPayload.ClientIP) {
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
	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode == http.StatusFound {
		return Identify(projectID, reqPayload, true)
	}

	if errCode == http.StatusNotFound {
		logCtx := log.WithField("token", token).WithField("request_payload", reqPayload)
		if IsValidTokenString(token) {
			// This is tracked with a metric on the dashboard as it could cause error spikes.
			logCtx.WithField("tag", "invalid_sdk_token").Info("Failed to get project from sdk project token.")
		} else {
			log.WithField("token", token).Warn("Invalid token on sdk payload.")
		}

		return http.StatusUnauthorized,
			&IdentifyResponse{Error: "Identify failed. Invalid token."}
	}

	return errCode, &IdentifyResponse{Error: "Identify failed."}
}

// AMPIdentifyByToken identifies AMP user by project token
func AMPIdentifyByToken(token string, reqPayload *AMPIdentifyPayload) (int, *IdentifyResponse) {
	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode != http.StatusFound {
		log.WithField("token", token).Error("Failed to get project from AMP sdk project token.")

		if errCode == http.StatusInternalServerError {
			return errCode, &IdentifyResponse{Error: "Identify failed. Failed to get AMP user."}
		}

		return http.StatusUnauthorized, &IdentifyResponse{Error: "Identify failed. Invalid project id."}
	}

	userID, errCode := store.GetStore().CreateOrGetAMPUser(projectID, reqPayload.ClientID, reqPayload.Timestamp, reqPayload.RequestSource)
	if errCode != http.StatusCreated && errCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("Identify failed. Failed to CreateOrGetAMPUser.")
		return errCode, &IdentifyResponse{Error: "Identify failed. Failed to get AMP user."}
	}

	identifyPayload := &IdentifyPayload{
		UserId:         userID,
		CustomerUserId: reqPayload.CustomerUserID,
		Timestamp:      reqPayload.Timestamp,
		RequestSource:  reqPayload.RequestSource,
	}

	return Identify(projectID, identifyPayload, false)
}

// AMPIdentifyWithQueue identifies AMP user by customer_user_id. Uses queue if alowed for the poject
func AMPIdentifyWithQueue(token string, reqPayload *AMPIdentifyPayload,
	queueAllowedTokens []string) (int, *IdentifyResponse) {
	if token == "" {
		return http.StatusBadRequest, &IdentifyResponse{Error: "Identify failed. Invalid token"}
	}

	if reqPayload.ClientID == "" || reqPayload.CustomerUserID == "" {
		return http.StatusBadRequest, &IdentifyResponse{Error: "Identify failed. Invalid params"}
	}

	if U.UseQueue(token, queueAllowedTokens) {

		err := enqueueRequest(token, sdkRequestTypeAMPIdentify, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue identify request.")
			return http.StatusInternalServerError,
				&IdentifyResponse{Error: "Identify failed."}
		}

		return http.StatusOK, &IdentifyResponse{Message: "User has been identified successfully"}
	}

	return AMPIdentifyByToken(token, reqPayload)
}

func IdentifyWithQueue(token string, reqPayload *IdentifyPayload,
	queueAllowedTokens []string) (int, *IdentifyResponse) {
	if reqPayload.Source == "" {
		reqPayload.Source = sdkRequestTypeUserIdentify
	}

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

	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode == http.StatusFound {
		return AddUserProperties(projectID, reqPayload)
	}

	if errCode == http.StatusNotFound {
		logCtx := log.WithField("token", token).WithField("request_payload", reqPayload)
		if IsValidTokenString(token) {
			// This is tracked with a metric on the dashboard as it could cause error spikes.
			logCtx.WithField("tag", "invalid_sdk_token").Info("Failed to get project from sdk project token.")
		} else {
			log.WithField("token", token).Warn("Invalid token on sdk payload.")
		}

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

	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode == http.StatusFound {
		return UpdateEventProperties(projectID, reqPayload)
	}

	if errCode == http.StatusNotFound {
		logCtx := log.WithField("token", token).WithField("request_payload", reqPayload)
		if IsValidTokenString(token) {
			// This is tracked with a metric on the dashboard as it could cause error spikes.
			logCtx.WithField("tag", "invalid_sdk_token").Info("Failed to get project from sdk project token.")
		} else {
			log.WithField("token", token).Warn("Invalid token on sdk payload.")
		}

		return http.StatusUnauthorized, &UpdateEventPropertiesResponse{
			Error: "Update event properties failed. Invalid token."}
	}

	return errCode, &UpdateEventPropertiesResponse{Error: "Failed to update event properties using token."}
}

func UpdateEventPropertiesWithQueue(token string, reqPayload *UpdateEventPropertiesPayload,
	queueAllowedTokens []string) (int, *UpdateEventPropertiesResponse) {

	if excludeBotRequestBySetting(token, reqPayload.UserAgent, "", reqPayload.ClientIP) {
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

func updateInitialUserPropertiesFromUpdateEventProperties(projectID int64,
	eventID, userID string, newInitialUserProperties *U.PropertiesMap) int {

	logCtx := log.WithField("project_id", projectID).WithField("event_id", eventID)

	existingUserProperties, errCode := store.GetStore().GetUserPropertiesByUserID(projectID, userID)
	if errCode != http.StatusFound {
		return errCode
	}

	userProperties, err := U.DecodePostgresJsonb(existingUserProperties)
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

	return overwriteUserPropertiesOnTable(projectID, userID, eventID, existingUserProperties, updateUserPropertiesJson)
}

func overwriteUserPropertiesOnTable(projectID int64, userID string, eventID string, existingUserProperties,
	updateUserPropertiesJson *postgres.Jsonb) int {

	logCtx := log.WithField("project_id", projectID).
		WithField("user_id", userID).WithField("eventID", eventID)

	errCode := store.GetStore().OverwriteUserPropertiesByID(projectID, userID,
		existingUserProperties, updateUserPropertiesJson, false, 0, "")
	if errCode != http.StatusAccepted {
		logCtx.WithField("err_code", errCode).
			Error("Failed to overwrite user's properties with initial page properties.")
		return errCode
	}

	errCode = store.GetStore().OverwriteEventUserPropertiesByID(
		projectID, userID, eventID, updateUserPropertiesJson)
	if errCode != http.StatusAccepted {
		logCtx.WithField("err_code", errCode).
			Error("Failed to overwrite event's user properties with initial page properties.")
		return errCode
	}

	return http.StatusAccepted
}

func UpdateEventProperties(projectId int64,
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

	logCtx := log.WithField("project_id", projectId).
		WithField("event_id", request.EventId).
		WithField("timestamp", request.Timestamp)

	// TODO: Add support for user_id on SDK and use user_id on GetEventById for routing to a shard.
	eventID, userID, errCode := store.GetStore().GetUserIdFromEventId(projectId, request.EventId, request.UserId)
	if errCode == http.StatusNotFound && request.Timestamp > U.UnixTimeBeforeDuration(time.Hour*5) {
		logCtx.Warn("Failed old update event properties request with unavailable event_id permanently.")
		return http.StatusBadRequest, &UpdateEventPropertiesResponse{
			Error: "Update event properties failed permanantly."}
	}
	if errCode != http.StatusFound {
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed. Invalid event."}
	}

	errCode = store.GetStore().UpdateEventProperties(projectId, eventID, userID, properitesToBeUpdated, request.Timestamp, nil)
	if errCode != http.StatusAccepted {
		return errCode, &UpdateEventPropertiesResponse{Error: "Update event properties failed. Failed to update given properties."}
	}

	newInitialUserProperties := U.GetUpdateAllowedInitialUserProperties(properitesToBeUpdated)

	// Update user_properties state associate to event and lastest user properties state of user.
	errCode = updateInitialUserPropertiesFromUpdateEventProperties(projectId, eventID, userID, newInitialUserProperties)
	if errCode != http.StatusAccepted {
		return errCode,
			&UpdateEventPropertiesResponse{
				Error: "Update event properties failed. Failed to update event user properties."}
	}

	return http.StatusAccepted,
		&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
}

type AMPTrackPayload struct {
	ClientID           string                 `json:"client_id"` // amp user_id
	SourceURL          string                 `json:"source_url"`
	Title              string                 `json:"title"`
	Referrer           string                 `json:"referrer"`
	ScreenHeight       float64                `json:"screen_height"`
	ScreenWidth        float64                `json:"screen_width"`
	PageLoadTimeInSecs float64                `json:"page_load_time_in_secs"`
	EventName          string                 `json:"event_name"`
	CustomProperties   map[string]interface{} `json:"custom_properties"`

	// internal
	Timestamp     int64  `json:"timestamp"`
	UserAgent     string `json:"user_agent"`
	ClientIP      string `json:"client_ip"`
	RequestSource int    `json:"request_source"`
}
type AMPUpdateEventPropertiesPayload struct {
	ClientID          string  `json:"client_id"` // amp user_id
	SourceURL         string  `json:"source_url"`
	PageScrollPercent float64 `json:"page_scroll_percent"`
	PageSpentTime     float64 `json:"page_spent_time"`

	// internal
	Timestamp     int64  `json:"timestamp"`
	UserAgent     string `json:"user_agent"`
	RequestSource int    `json:"request_source"`
	ClientIP      string `json:"client_ip"`
}
type AMPTrackResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func AMPUpdateEventPropertiesByToken(token string,
	reqPayload *AMPUpdateEventPropertiesPayload) (int, *Response) {

	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode != http.StatusFound {
		return http.StatusUnauthorized, &Response{Error: "Invalid token"}
	}

	logCtx := log.WithField("project_id", projectID)

	parsedSourceURL, err := U.ParseURLStable(reqPayload.SourceURL)
	if err != nil {
		logCtx.WithField("canonical_url", reqPayload.SourceURL).WithError(err).Error(
			"Failed to parsing page url from canonical_url query param on amp sdk update event properties")
		return http.StatusBadRequest, &Response{Error: "Invalid page url"}
	}

	pageURL := U.CleanURI(parsedSourceURL.Host + parsedSourceURL.Path)

	userID, errCode := store.GetStore().GetUserIDByAMPUserID(projectID, reqPayload.ClientID)
	if errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			logCtx.WithField("client_id", reqPayload.ClientID).
				Warn("User not found on amp update event_properties.")
			return http.StatusBadRequest, &Response{Error: "Invalid amp user."}
		}

		return http.StatusInternalServerError, &Response{Error: "Invalid amp user."}
	}

	logCtx = logCtx.WithField("user_id", userID).WithField("page_url", pageURL)

	eventID, errCode := GetCacheAMPSDKEventIDByPageURL(projectID, userID, pageURL)
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

	errCode = store.GetStore().UpdateEventProperties(projectID, eventID, userID,
		&updateEventProperties, time.Now().Unix(), nil)
	if errCode != http.StatusAccepted {
		logCtx.WithFields(log.Fields{"project_id": projectID, "event_id": eventID}).
			Error("Failed to update event properties")
		return errCode, &Response{Error: "Failed to update event properties."}
	}

	return http.StatusAccepted, &Response{Message: "Updated event properties successfully."}
}

func AMPTrackByToken(token string, reqPayload *AMPTrackPayload) (int, *Response) {
	projectID, errCode := store.GetStore().GetProjectIDByToken(token)
	if errCode != http.StatusFound {
		return http.StatusUnauthorized, &Response{Error: "Invalid token"}
	}
	logCtx := log.WithField("project_id", projectID).WithField("client_id", reqPayload.ClientID)

	var isNewUser bool
	userID, errCode := store.GetStore().CreateOrGetAMPUser(projectID, reqPayload.ClientID, reqPayload.Timestamp, reqPayload.RequestSource)
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
	if reqPayload.Title != "" {
		eventProperties[U.EP_PAGE_TITLE] = reqPayload.Title
	}

	if referrerRawURL != "" {
		eventProperties[U.EP_REFERRER] = referrerRawURL
		eventProperties[U.EP_REFERRER_URL] = referrerURL
		eventProperties[U.EP_REFERRER_DOMAIN] = referrerDomain
	}

	U.FillPropertiesFromURL(&eventProperties, parsedSourceURL)

	for k, v := range reqPayload.CustomProperties {
		eventProperties[k] = v
	}

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

	err = FillUserAgentUserProperties(projectID, &userProperties, reqPayload.UserAgent)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fill user agent user properties on amp track.")
	}

	trackPayload := TrackPayload{
		Auto:            true,
		UserId:          userID,
		IsNewUser:       isNewUser,
		Name:            pageURL,
		EventProperties: eventProperties,
		UserProperties:  userProperties,
		ClientIP:        reqPayload.ClientIP,
		UserAgent:       reqPayload.UserAgent,
		Timestamp:       reqPayload.Timestamp,
		RequestSource:   model.UserSourceWeb,
	}

	// Support for custom event_name.
	if reqPayload.EventName != "" {
		trackPayload.Name = reqPayload.EventName
	}

	errCode, trackResponse := Track(projectID, &trackPayload, false, SourceAMPSDK, "")
	if trackResponse.EventId != "" {
		cacheErrCode := SetCacheAMPSDKEventIDByPageURL(projectID, userID,
			trackResponse.EventId, pageURL)
		if cacheErrCode != http.StatusAccepted {
			logCtx.WithField("err_code", errCode).WithField("user_id", userID).
				Error("Failed to set cache event_id by page_url on AMP.")
		}
	} else {
		logCtx.WithFields(log.Fields{"user_id": userID, "event_id": trackResponse.EventId}).
			Error("Missing event_id from response of track on AMP track.")
	}

	return errCode, &Response{EventId: trackResponse.EventId,
		Message: trackResponse.Message, Error: trackResponse.Error}
}

func getAMPSDKByEventIDCacheKey(projectId int64, userId string, pageURL string) (*cache.Key, error) {
	prefix := "amp_sdk_user_event"
	suffix := "uid:" + userId + ":url:" + pageURL
	return cache.NewKey(projectId, prefix, suffix)
}

func SetCacheAMPSDKEventIDByPageURL(projectId int64, userId string, eventId string, pageURL string) int {
	logctx := log.WithFields(log.Fields{"project_id": projectId,
		"user_id": userId, "event_id": eventId, "page_url": pageURL})

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

func GetCacheAMPSDKEventIDByPageURL(projectId int64, userId string, pageURL string) (string, int) {
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

	if excludeBotRequestBySetting(token, reqPayload.UserAgent, "", reqPayload.ClientIP) {
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

	if excludeBotRequestBySetting(token, reqPayload.UserAgent, "", reqPayload.ClientIP) {
		return http.StatusNotModified,
			&Response{Message: "Update event properties skipped. Bot request."}
	}

	if U.UseQueue(token, queueAllowedTokens) {
		err := enqueueRequest(token, sdkRequestTypeAMPEventUpdateProperties, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue amp sdk update event request.")
			return http.StatusInternalServerError, &Response{Error: "Update event properties failed"}
		}

		return http.StatusOK, &Response{Message: "Updated event successfully"}
	}

	return AMPUpdateEventPropertiesByToken(token, reqPayload)
}

// PostDeviceServiceAPI - Make POST request to Device Server
func PostDeviceServiceAPI(apiUrl string, userAgent string) (model.DeviceInfo, int, error) {
	logCtx := log.WithFields(log.Fields{
		"userAgent": userAgent,
		"Method":    "PostDeviceServiceAPI",
	})

	start := time.Now()
	var res model.DeviceInfo
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		msg := "failed to build request"
		logCtx.WithError(err).Error(msg)
		return res, http.StatusInternalServerError, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)

	if err != nil {
		msg := "Request to Device detection service failed"
		logCtx.WithError(err).Error(msg)
		return res, resp.StatusCode, err
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		msg := "User Agent Not found"
		logCtx.WithField("status_code", resp.StatusCode).Warning(msg)
		return res, resp.StatusCode, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "Error reading the response body"
		logCtx.WithError(err).Error(msg)
		return res, http.StatusInternalServerError, err
	}

	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		msg := "Failed to unmarshall"
		logCtx.WithError(err).Error(msg)
		return res, http.StatusInternalServerError, err
	}

	// check for time elapsed in the post request to device service
	elapsed := time.Since(start).Milliseconds()
	if elapsed > 100 {
		log.WithFields(log.Fields{"total_time": elapsed, "user_agent": userAgent}).Warning("PostDeviceServiceAPI call took more than 100 ms")
	}

	return res, http.StatusOK, nil
}

func GetDeviceInfoFromDeviceService(userAgent string) (deviceInfo model.DeviceInfo, status int, err error) {

	resp, errCode, err := model.GetCacheResultByUserAgent(userAgent)

	// check if device info for user agent exists in cache
	if errCode == http.StatusFound && err == nil {
		deviceInfo = *resp
	} else if userAgent == "" {
		return model.DeviceInfo{}, http.StatusBadRequest, err
	} else {

		deviceInfo, status, err = PostDeviceServiceAPI(C.GetConfig().DeviceServiceURL, userAgent)

		model.SetCacheResultByUserAgent(userAgent, &deviceInfo)
	}

	return deviceInfo, status, err
}

// FillDeviceInfoFromDeviceService - For given user agent it gets device info from php device service by default
// and fallbacks to existing device data in case of failure
func FillDeviceInfoFromDeviceService(userProperties *U.PropertiesMap, userAgent string) error {
	var allDeviceInfo *model.DeviceInfo

	deviceInfo, status, err := GetDeviceInfoFromDeviceService(userAgent)
	allDeviceInfo = &deviceInfo
	// check for status of POST request
	if err != nil || status != http.StatusOK {
		return FillDeviceInfoFromFallback(userProperties, userAgent)
	}

	(*userProperties)[U.UP_USER_AGENT] = userAgent

	// check for bot
	if allDeviceInfo.IsBot {
		(*userProperties)[U.UP_BROWSER] = "Bot"
		return nil
	}

	osInfo := allDeviceInfo.OsInfo

	(*userProperties)[U.UP_OS] = osInfo.Name
	(*userProperties)[U.UP_OS_VERSION] = osInfo.Version
	(*userProperties)[U.UP_OS_WITH_VERSION] = fmt.Sprintf("%s-%s",
		(*userProperties)[U.UP_OS], (*userProperties)[U.UP_OS_VERSION])

	clientInfo := allDeviceInfo.ClientInfo

	(*userProperties)[U.UP_BROWSER] = clientInfo.Name
	(*userProperties)[U.UP_BROWSER_VERSION] = clientInfo.Version
	(*userProperties)[U.UP_BROWSER_WITH_VERSION] = fmt.Sprintf("%s-%s",
		(*userProperties)[U.UP_BROWSER], (*userProperties)[U.UP_BROWSER_VERSION])

	(*userProperties)[U.UP_DEVICE_BRAND] = allDeviceInfo.DeviceBrand
	(*userProperties)[U.UP_DEVICE_TYPE] = allDeviceInfo.DeviceType
	(*userProperties)[U.UP_DEVICE_MODEL] = allDeviceInfo.DeviceModel
	return nil

}

// FillUserAgentUserProperties - Adds user_properties derived from user_agent.
// Note: Defined here to avoid cyclic import of config package on util.
func FillUserAgentUserProperties(projectID int64, userProperties *U.PropertiesMap, userAgentStr string) error {

	if userAgentStr == "" || projectID == 0 {
		return errors.New("invalid parameters")
	}

	if !C.AllowDeviceServiceByProjectID(projectID) {
		return FillDeviceInfoFromFallback(userProperties, userAgentStr)
	}

	return FillDeviceInfoFromDeviceService(userProperties, userAgentStr)

}

func FillDeviceInfoFromFallback(userProperties *U.PropertiesMap, userAgentStr string) error {
	userAgent := user_agent.New(userAgentStr)
	(*userProperties)[U.UP_USER_AGENT] = userAgent
	(*userProperties)[U.UP_OS] = userAgent.OSInfo().Name
	(*userProperties)[U.UP_OS_VERSION] = userAgent.OSInfo().Version
	(*userProperties)[U.UP_OS_WITH_VERSION] = fmt.Sprintf("%s-%s",
		(*userProperties)[U.UP_OS], (*userProperties)[U.UP_OS_VERSION])

	if U.IsBotUserAgent(userAgentStr) {
		(*userProperties)[U.UP_BROWSER] = "Bot"
		return nil
	}

	browserName, browserVersion := U.GetBrowser(userAgent)

	(*userProperties)[U.UP_BROWSER] = browserName
	(*userProperties)[U.UP_BROWSER_VERSION] = browserVersion
	(*userProperties)[U.UP_BROWSER_WITH_VERSION] = fmt.Sprintf("%s-%s",
		(*userProperties)[U.UP_BROWSER], (*userProperties)[U.UP_BROWSER_VERSION])

	dd := C.GetServices().DeviceDetector
	if info := dd.Parse(userAgentStr); info != nil {
		(*userProperties)[U.UP_DEVICE_BRAND] = info.Brand
		(*userProperties)[U.UP_DEVICE_TYPE] = info.Type
		(*userProperties)[U.UP_DEVICE_MODEL] = info.Model
	}
	return nil
}

// TrackGroupWithDomain tracks groups which will also get associated with domains group
func TrackGroupWithDomain(projectID int64, groupName, domainName string, properties U.PropertiesMap, timestamp int64) (string, int) {
	logFields := log.Fields{"project_id": projectID, "domain_name": domainName, "group_name": groupName, "properties": properties}
	logCtx := log.WithFields(logFields)
	if projectID == 0 || groupName == "" || domainName == "" || timestamp == 0 {
		logCtx.Error("Invalid parameters.")
		return "", http.StatusBadRequest
	}

	if !model.IsAllowedAccountGroupNames(groupName) {
		logCtx.Error("Invalid account group name.")
		return "", http.StatusBadRequest
	}

	_, status := store.GetStore().CreateOrGetGroupByName(projectID, groupName, model.AllowedGroupNames)
	if status != http.StatusFound && status != http.StatusCreated {
		logCtx.Error("Failed to CreateOrGetGroupByName on TrackGroupWithDomain.")
		return "", http.StatusInternalServerError
	}

	groupID := U.GetDomainGroupDomainName(projectID, domainName)
	if groupID == "" {
		logCtx.Warn("No domain available")
		return "", http.StatusNotImplemented
	}
	groupUserID, status := store.GetStore().CreateOrGetDomainGroupUser(projectID, groupName, groupID, timestamp, model.GetGroupUserSourceByGroupName(groupName))
	if status != http.StatusCreated && status != http.StatusFound {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for  group user by group id.")
		return "", http.StatusInternalServerError
	}

	propertiesMap := map[string]interface{}(properties)
	source := model.GetGroupUserSourceNameByGroupName(groupName)

	_, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID, groupUserID, &propertiesMap, timestamp, timestamp, source)
	if err != nil {
		logCtx.WithError(err).Error("Failed to create or update group user on TrackUserGroup.")
		return "", http.StatusInternalServerError
	}

	if C.IsAllowedDomainsGroupByProjectID(projectID) {
		status = TrackDomainsGroup(projectID, groupUserID, groupName, domainName, timestamp)
		if status != http.StatusOK {
			logCtx.Error("Failed to TrackDomainsGroup.")
			return "", http.StatusInternalServerError
		}
	}

	return groupUserID, http.StatusOK
}

// TrackUserAccountGroup to used only within Track call
func TrackUserAccountGroup(projectID int64, userID string, groupName string, groupProperties *U.PropertiesMap, timestamp int64) (status int) {
	logFields := log.Fields{"project_id": projectID, "group_name": groupName, "user_id": userID, "properties": groupProperties}
	logCtx := log.WithFields(logFields)
	defer func(startTime time.Time) {
		if status == http.StatusOK {
			totalTime := time.Now().Sub(startTime).Milliseconds()
			logCtx.WithFields(log.Fields{"total_time": totalTime}).Warning("TrackUserGroup completed successfully.")
		}

	}(time.Now())

	if projectID == 0 || groupName == "" || groupProperties == nil || userID == "" {
		logCtx.Error("Invalid arguments")
		return http.StatusBadRequest
	}

	groupIDPropertyKeys := model.GetDomainNameSourcePropertyKey(groupName)
	groupID := ""
	for i := range groupIDPropertyKeys {
		if groupID = U.GetDomainGroupDomainName(projectID, U.GetPropertyValueAsString((*groupProperties)[groupIDPropertyKeys[i]])); groupID != "" {
			break
		}
	}

	if groupID == "" {
		logCtx.Warning("No group id. Skip processing user group.")
		return http.StatusNotImplemented
	}

	groupUserID, status := TrackGroupWithDomain(projectID, groupName, groupID, *groupProperties, timestamp)
	if status != http.StatusOK {
		logCtx.Error("Failed to track TrackUserGroup.")
		return http.StatusInternalServerError
	}

	_, status = store.GetStore().UpdateUserGroup(projectID, userID, groupName, groupID, groupUserID, true)
	if status != http.StatusAccepted && status != http.StatusNotModified {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to update user association with group user on TrackUserAccountGroup.")
		return http.StatusInternalServerError
	}

	if C.EnableUserDomainsGroupByProjectID(projectID) {
		status = store.GetStore().AssociateUserDomainsGroup(projectID, userID, groupName, groupUserID)
		if status != http.StatusOK && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to AssociateUserDomainsGroup on TrackUserAccountGroup.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

// TrackDomainsGroup track $domains/All accounts group. Either group user id or domain name is required
func TrackDomainsGroup(projectID int64, groupUserID string, groupName string, domainName string, timestamp int64) int {
	logFields := log.Fields{"project_id": projectID,
		"domain_name": domainName, "timestamp": timestamp}
	logCtx := log.WithFields(logFields)

	if projectID == 0 || groupName == "" || timestamp == 0 {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest
	}

	if !model.IsAllowedGroupForDomainsGroup(groupName) {
		logCtx.Error("Group not allowed for domains group.")
		return http.StatusBadRequest
	}

	if groupUserID == "" && domainName == "" {
		logCtx.Error("Group user id and domain name empty.")
		return http.StatusBadRequest
	}

	_, status := store.GetStore().CreateOrGetDomainsGroup(projectID)
	if status != http.StatusFound && status != http.StatusCreated {
		logCtx.Error("Failed to CreateOrGetDomainsGroup.")
		return http.StatusInternalServerError
	}

	groupID := U.GetDomainGroupDomainName(projectID, domainName)
	if groupID == "" {
		logCtx.Warning("No group id. Skip processing user group.")
		return http.StatusOK
	}

	if groupUserID == "" {
		groupUserID, status = store.GetStore().CreateOrGetDomainGroupUser(projectID, groupName, groupID, timestamp,
			model.GetGroupUserSourceByGroupName(groupName))
		if status != http.StatusFound && status != http.StatusCreated {
			logCtx.Warning("Failed to CreateOrGetGroupUserID on TrackDomainsGroup.")
			return http.StatusInternalServerError
		}
	}

	domainsGroupUserID, status := store.GetStore().CreateOrGetDomainGroupUser(projectID, model.GROUP_NAME_DOMAINS, groupID, timestamp,
		model.GetGroupUserSourceByGroupName(model.GROUP_NAME_DOMAINS))
	if status != http.StatusFound && status != http.StatusCreated {
		logCtx.Warning("Failed to CreateOrGetDomainGroupUser on TrackDomainsGroup.")
		return http.StatusInternalServerError
	}

	_, status = store.GetStore().UpdateGroupUserDomainsGroup(projectID, groupUserID, domainsGroupUserID, groupID, true)
	if status != http.StatusAccepted && status != http.StatusNotModified {
		logCtx.WithFields(log.Fields{"status": status}).Error("Failed to update group user association with domains group user.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}
