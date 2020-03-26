package sdk

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
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
}

type UpdateEventPropertiesResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

const RequestQueue = "sdk_request_queue"
const ProcessRequestTask = "process_sdk_request"

const (
	sdkRequestTypeEventTrack            = "sdk_event_track"
	sdkRequestTypeUserIdentify          = "sdk_user_identify"
	sdkRequestTypeUserAddProperties     = "sdk_user_add_properties"
	sdkRequestTypeEventUpdateProperties = "sdk_event_update_properties"
)

func ProcessQueueRequest(token, reqType, reqPayloadStr string) (float64, string, error) {
	// Todo(Dinesh): Retry on panic: Add payload back to queue as return
	// from defer is not possible and notify panic.

	// Todo(Dinesh): Add request_id for better tracing.

	logCtx := log.WithFields(log.Fields{"queue": RequestQueue, "token": token,
		"req_type": reqType, "req_payload": reqPayloadStr})

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

		status, response = TrackByToken(token, &reqPayload)

	case sdkRequestTypeUserIdentify:
		var reqPayload IdentifyPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}

		status, response = IdentifyByToken(token, &reqPayload)

	case sdkRequestTypeUserAddProperties:
		var reqPayload AddUserPropertiesPayload

		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}

		status, response = AddUserPropertiesByToken(token, &reqPayload)

	case sdkRequestTypeEventUpdateProperties:
		var reqPayload UpdateEventPropertiesPayload
		err := json.Unmarshal([]byte(reqPayloadStr), &reqPayload)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to unmarshal request payload on sdk process queue.")
			return http.StatusInternalServerError, "", nil
		}

		status, response = UpdateEventPropertiesByToken(token, &reqPayload)
	default:
		logCtx.Error("Invalid sdk request type on sdk process queue")
		return http.StatusInternalServerError, "", nil
	}

	responseBytes, _ := json.Marshal(response)
	logCtx = logCtx.WithField("status", status).WithField("response", string(responseBytes))

	// Log for analysing queue process status.
	logCtx.WithField("processed", "true").Info("Processed sdk request.")

	// Do not retry on below conditions.
	if status == http.StatusBadRequest || status == http.StatusNotAcceptable || status == http.StatusUnauthorized {
		logCtx.Info("Failed to process sdk request permanantly.")
		return float64(status), "", nil
	}

	// Return error only for retry. Retry after a period till it is successfull.
	// Retry dependencies not found and failures which can be successful on retries.
	if status == http.StatusNotFound || status == http.StatusInternalServerError {
		logCtx.WithField("retry", "true").Info("Failed to process sdk request on sdk process queue. Retry.")
		return http.StatusInternalServerError, "",
			tasks.NewErrRetryTaskLater("RETRY__REQUEST_PROCESSING_FAILURE", 5*time.Minute)
	}

	return http.StatusOK, string(responseBytes), nil
}

func enrichAfterTrack(projectId uint64, event *M.Event,
	userProperties *map[string]interface{}, reqTimestamp int64) int {

	if projectId == 0 || event == nil || userProperties == nil {
		return http.StatusBadRequest
	}

	if isAllPropertiesMissing := (*userProperties)[U.UP_HOUR_OF_FIRST_EVENT] == nil &&
		(*userProperties)[U.UP_DAY_OF_FIRST_EVENT] == nil; !isAllPropertiesMissing {
		return http.StatusOK
	}

	err := U.FillFirstEventUserProperties(userProperties, event.Timestamp)
	if err != nil {
		log.WithField("user_id", event.UserId).WithError(err).Error(
			"Failed to fill day of first event and hour of first event user properties on enrich after track.")
		return http.StatusInternalServerError
	}
	userPropsJSON, err := json.Marshal(userProperties)
	if err != nil {
		log.WithField("user_id", event.UserId).WithError(err).Error(
			"Failed to marshal existing user properties on enrich after track.")
		return http.StatusInternalServerError
	}

	_, errCode := M.UpdateUserProperties(projectId, event.UserId, &postgres.Jsonb{userPropsJSON}, reqTimestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		log.WithFields(log.Fields{"userProperties": userProperties,
			log.ErrorKey: errCode}).Error("Update user properties failed on enrich after track.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func Track(projectId uint64, request *TrackPayload,
	skipSession bool) (int, *TrackResponse) {
	logCtx := log.WithField("project_id", projectId)

	if projectId == 0 || request == nil {
		logCtx.WithField("request_payload", request).Error("Invalid track request.")
		return http.StatusBadRequest, &TrackResponse{}
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
		return http.StatusNotModified, &TrackResponse{}
	}

	var eventName *M.EventName
	var eventNameErrCode int

	// if auto_track enabled, auto_name = event_name else auto_name = "UCEN".
	// On 'auto' true event_name is the eventURL(e.g: factors.ai/u1/u2/u3) for JS_.
	if request.Auto {
		// Pass eventURL through filter and get corresponding event_name mapped by user.
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

	var userProperties *U.PropertiesMap

	response := &TrackResponse{}
	initialUserProperties := U.GetInitialUserProperties(eventProperties)
	isNewUser := request.IsNewUser

	// if create_user not true and user is not found,
	// allow to create_user.
	if !request.CreateUser && request.UserId != "" {
		_, errCode := M.GetUser(projectId, request.UserId)
		if errCode == http.StatusNotFound {
			request.CreateUser = true
		}
	}

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
		existingUserProperties, errCode := M.GetUserPropertiesAsMap(projectId, request.UserId)
		if errCode != http.StatusFound {
			logCtx.WithField("user_id", errCode).WithField("err_code",
				errCode).Error("Tracking failed. Get user properties as map failed.")
			return errCode, &TrackResponse{Error: "Tracking failed while getting user."}
		}

		// Is any initial user properties exists already.
		initialUserPropertyExists := false
		for k := range *initialUserProperties {
			if _, exists := (*existingUserProperties)[k]; exists {
				initialUserPropertyExists = true
				break
			}
		}

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
	U.FillUserAgentUserProperties(userProperties, request.UserAgent)
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

	userPropsJSON, err := json.Marshal(userProperties)
	if err != nil {
		log.WithFields(log.Fields{"userProperties": userProperties,
			log.ErrorKey: err}).Error("Update user properites on track failed. JSON marshal failed.")
		response.Error = "Failed updating user properties."
	}

	userPropertiesId, errCode := M.UpdateUserProperties(projectId, request.UserId,
		&postgres.Jsonb{userPropsJSON}, request.Timestamp)
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		log.WithFields(log.Fields{"userProperties": userProperties,
			log.ErrorKey: errCode}).Error("Update user properties on track failed. DB update failed.")
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

	if !skipSession {
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

		(*eventProperties)[U.EP_SESSION] = session.Count
		event.SessionId = &session.ID
	}
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

	existingUserProperties, errCode := M.GetUserPropertiesAsMap(projectId, event.UserId)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error(
			"Failed to get user properties for adding first event properties on track.")
	}

	// Todo: Try to use latest user properties, if available already.
	errCode = enrichAfterTrack(projectId, createdEvent, existingUserProperties, request.Timestamp)
	if errCode != http.StatusOK {
		// Logged and skipping failure response on after track enrichement failure.
		log.WithField("err_code", errCode).Error("Failed to enrich after track.")
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

	errCode = M.UpdateUserJoinTimePropertyForCustomerUser(projectId, request.CustomerUserId)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode, &IdentifyResponse{Error: "Identification failed."}
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

func TrackByToken(token string, reqPayload *TrackPayload) (int, *TrackResponse) {
	project, errCode := M.GetProjectByToken(token)
	if errCode == http.StatusFound {
		return Track(project.ID, reqPayload, false)
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

	return errCode, &UpdateEventPropertiesResponse{Error: "Update event properties failed."}
}

func UpdateEventPropertiesWithQueue(token string, reqPayload *UpdateEventPropertiesPayload,
	queueAllowedTokens []string) (int, *UpdateEventPropertiesResponse) {

	if U.UseQueue(token, queueAllowedTokens) {
		// add queued timestamp, if timestmap is not given.
		if reqPayload.Timestamp == 0 {
			reqPayload.Timestamp = time.Now().Unix()
		}

		err := enqueueRequest(token, sdkRequestTypeEventUpdateProperties, reqPayload)
		if err != nil {
			log.WithError(err).Error("Failed to queue updated event properties request.")
			return http.StatusInternalServerError,
				&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
		}

		return http.StatusOK,
			&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
	}

	return UpdateEventPropertiesByToken(token, reqPayload)
}

func UpdateEventProperties(projectId uint64,
	request *UpdateEventPropertiesPayload) (int, *UpdateEventPropertiesResponse) {

	logCtx := log.WithField("project_id", projectId)

	// add received timestamp before processing, if not given.
	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	updateAllowedProperties := U.GetUpdateAllowedEventProperties(&request.Properties)
	validatedProperties := U.GetValidatedEventProperties(updateAllowedProperties)
	if len(*validatedProperties) == 0 {
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "No valid properties given to update."}
	}

	errCode := M.UpdateEventPropertiesByTimestamp(projectId, request.EventId,
		validatedProperties, request.Timestamp)
	if errCode != http.StatusAccepted {
		return errCode,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	updatedEvent, errCode := M.GetEventById(projectId, request.EventId)
	if errCode == http.StatusNotFound && request.Timestamp > U.UnixTimeBeforeDuration(time.Hour*5) {
		logCtx.WithField("event_id", request.EventId).WithField("timestamp", request.Timestamp).Error(
			"Failed old update event properties request with unavailable event_id permanently.")
		return http.StatusBadRequest, &UpdateEventPropertiesResponse{
			Error: "Update event properties failed permanantly."}
	}
	if errCode != http.StatusFound {
		return errCode,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	logCtx = logCtx.WithField("event_id", request.EventId)

	if updatedEvent.SessionId == nil || *updatedEvent.SessionId == "" {
		logCtx.Error("Session id does not exist to update session properties on update event properties.")
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed. No session associated."}
	}

	newSessionProperties := U.GetSessionProperties(false, validatedProperties, &U.PropertiesMap{})
	sessionEvent, errCode := M.GetEventById(projectId, *updatedEvent.SessionId)
	if errCode != http.StatusFound {
		return errCode,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	updatedEventPropertiesMap, err := U.DecodePostgresJsonb(&updatedEvent.Properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal updated event properties on update event properties.")
		return http.StatusInternalServerError,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	sessionPropertiesMap, err := U.DecodePostgresJsonb(&sessionEvent.Properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal existing session properties on update event properties.")
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	eventPageURL, eventPageURLExists := (*updatedEventPropertiesMap)[U.EP_PAGE_RAW_URL]

	sessionInitialPageURL, sessionInitialPageURLExists := (*sessionPropertiesMap)[U.UP_INITIAL_PAGE_RAW_URL]
	if !eventPageURLExists || !sessionInitialPageURLExists {
		logCtx.Error("No page URL property to compare for session properties update.")
		return http.StatusBadRequest,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	// session properties updated only if page raw url and initial
	// page raw url matches.
	if eventPageURL != sessionInitialPageURL {
		return http.StatusAccepted,
			&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
	}

	isSessionPropertiesUpdateRequired := false
	for property, value := range *newSessionProperties {
		if _, exists := (*sessionPropertiesMap)[property]; !exists {
			(*sessionPropertiesMap)[property] = value
			isSessionPropertiesUpdateRequired = true
		}
	}

	// updates only when new properties added.
	if isSessionPropertiesUpdateRequired {
		updateSesssionProperties := U.PropertiesMap(*sessionPropertiesMap)
		errCode := M.UpdateEventPropertiesByTimestamp(projectId, sessionEvent.ID,
			&updateSesssionProperties, request.Timestamp)
		if errCode != http.StatusAccepted {
			return errCode,
				&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
		}
	}

	newUserProperties := U.GetInitialUserProperties(validatedProperties)
	user, errCode := M.GetUser(projectId, updatedEvent.UserId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get user properties of user on update event properties.")
		return errCode,
			&UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	userPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal existing user properties on update event properties.")
		return errCode, &UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	userInitialPageURL, userInitialPageURLExists := (*userPropertiesMap)[U.UP_INITIAL_PAGE_RAW_URL]
	if !userInitialPageURLExists {
		// skip error for old users.
		initialPageRawUrlAvailabilityTimestamp := 1569369600
		if user.JoinTimestamp < int64(initialPageRawUrlAvailabilityTimestamp) {
			return http.StatusAccepted,
				&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
		}

		logCtx.Error("No initial page raw url on user properties to compare on update event properties.")
		return errCode, &UpdateEventPropertiesResponse{Error: "Update event properties failed."}
	}

	// user properties updated only if initial_raw_page url of user_properties
	// and raw_page_url of event properties matches.
	if userInitialPageURL != eventPageURL {
		return http.StatusAccepted,
			&UpdateEventPropertiesResponse{Message: "Updated event properties successfully."}
	}

	isUserPropertiesUpdateRequired := false
	for property, value := range *newUserProperties {
		if _, exists := (*userPropertiesMap)[property]; !exists {
			(*userPropertiesMap)[property] = value
			isUserPropertiesUpdateRequired = true
		}
	}

	if isUserPropertiesUpdateRequired {
		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(userPropertiesMap)
		if err != nil {
			logCtx.Error("Failed to marshal user properties with initial user properties on updated event properties.")
			return errCode, &UpdateEventPropertiesResponse{Error: "Update event properties failed."}
		}

		_, errCode := M.UpdateUserProperties(projectId, updatedEvent.UserId,
			userPropertiesJsonb, request.Timestamp)
		if errCode != http.StatusAccepted {
			return errCode, &UpdateEventPropertiesResponse{Error: "Update event properties failed."}
		}
	}

	return http.StatusAccepted,
		&UpdateEventPropertiesResponse{Error: "Updated event properties successfully."}
}
