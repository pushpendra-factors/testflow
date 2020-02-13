package handler

import (
	"encoding/json"
	"errors"
	C "factors/config"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type SDKTrackPayload struct {
	Name            string          `json:"event_name"`
	CustomerEventId *string         `json:"c_event_id"`
	EventProperties U.PropertiesMap `json:"event_properties"`
	UserProperties  U.PropertiesMap `json:"user_properties"`
	ProjectId       uint64          `json:"project_id"`
	UserId          string          `json:"user_id"`
	IsNewUser       bool            `json:"-"` // Not part of request json payload.
	Auto            bool            `json:"auto"`
	Timestamp       int64           `json:"timestamp`
}

type SDKTrackResponse struct {
	EventId         string  `json:"event_id,omitempty"`
	Type            string  `json:"type,omitempty"`
	CustomerEventId *string `json:"c_event_id,omitempty"`
	UserId          string  `json:"user_id,omitempty"`
	Message         string  `json:"message,omitempty"`
	Error           string  `json:"error,omitempty"`
}

type SDKIdentifyPayload struct {
	UserId         string `json:"user_id"`
	CustomerUserId string `json:"c_uid"`
	JoinTimestamp  int64  `json:"join_timestamp"`
}

type sdkAddUserPropertiesPayload struct {
	UserId     string          `json:"user_id"`
	Properties U.PropertiesMap `json:"properties"`
}

type sdkUpdateEventPropertiesPayload struct {
	EventId    string          `json:"event_id"`
	Properties U.PropertiesMap `json:"properties"`
}

func enrichAfterTrack(projectId uint64, event *M.Event, userProperties *map[string]interface{}) int {
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

	existingUserPropsJSON, err := json.Marshal(userProperties)
	if err != nil {
		log.WithField("user_id", event.UserId).Error(
			"Failed to marshal existing user properties on enrich after track.")
		return http.StatusInternalServerError
	}

	_, errCode := M.UpdateUserProperties(projectId, event.UserId, &postgres.Jsonb{existingUserPropsJSON})
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		log.WithFields(log.Fields{"userProperties": userProperties,
			log.ErrorKey: errCode}).Error("Update user properties failed on enrich after track.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func SDKTrack(projectId uint64, request *SDKTrackPayload, clientIP,
	userAgent string, skipSession bool) (int, *SDKTrackResponse) {

	if projectId == 0 || request == nil {
		log.WithFields(log.Fields{"project_id": projectId,
			"request_payload": request}).Error("Invalid track request.")
		return http.StatusBadRequest, &SDKTrackResponse{}
	}

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	// Skipping track for configured projects.
	for _, skipProjectId := range C.GetSkipTrackProjectIds() {
		if skipProjectId == projectId {
			// Todo: Change status to StatusBadRequest, using StatusOk to avoid retries.
			return http.StatusOK, &SDKTrackResponse{Error: "Tracking skipped."}
		}
	}

	// Precondition: Fails if event_name not provided.
	request.Name = strings.TrimSpace(request.Name) // Discourage whitespace on the end.
	if request.Name == "" {
		return http.StatusBadRequest, &SDKTrackResponse{Error: "Tracking failed. Event name cannot be omitted or left empty."}
	}

	projectSettings, errCode := M.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError, &SDKTrackResponse{Error: "Tracking failed. Invalid project."}
	}

	// Terminate track calls from bot user_agent.
	if *projectSettings.ExcludeBot && U.IsBotUserAgent(userAgent) {
		return http.StatusNotModified, &SDKTrackResponse{}
	}

	var eventName *M.EventName
	var eventNameErrCode int

	// if auto_track enabled, auto_name = event_name else auto_name = "UCEN".
	// On 'auto' true event_name is the eventURL(e.g: factors.ai/u1/u2/u3) for JS_SDK.
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
		return eventNameErrCode, &SDKTrackResponse{Error: "Tracking failed. Creating event_name failed."}
	}

	// Event Properties
	U.UnEscapeQueryParamProperties(&request.EventProperties)
	definedEventProperties, hasDefinedMarketingProperty := U.MapEventPropertiesToDefinedProperties(&request.EventProperties)
	eventProperties := U.GetValidatedEventProperties(definedEventProperties)
	if ip, ok := (*eventProperties)[U.EP_INTERNAL_IP]; ok && ip != "" {
		clientIP = ip.(string)
	}
	// Added IP to event properties for internal usage.
	(*eventProperties)[U.EP_INTERNAL_IP] = clientIP
	eventPropsJSON, err := json.Marshal(eventProperties)
	if err != nil {
		return http.StatusBadRequest, &SDKTrackResponse{Error: "Tracking failed. Invalid properties."}
	}

	var userProperties *U.PropertiesMap

	response := &SDKTrackResponse{}
	initialUserProperties := U.GetInitialUserProperties(eventProperties)
	isNewUser := request.IsNewUser

	if request.UserId == "" {
		// Precondition: if user_id not given, create new user and respond.

		newUser, errCode := M.CreateUser(&M.User{ProjectId: projectId})
		if errCode != http.StatusCreated {
			return errCode, &SDKTrackResponse{Error: "Tracking failed. User creation failed."}
		}

		request.UserId = newUser.ID
		response.UserId = newUser.ID
		isNewUser = true

		// Initialize with initial user properties.
		userProperties = initialUserProperties
	} else {
		// Adding initial user properties if user_id exists,
		// but initial properties are not. i.e user created on identify.
		existingUserProperties, errCode := M.GetUserPropertiesAsMap(projectId, request.UserId)
		if errCode != http.StatusFound {
			return errCode, &SDKTrackResponse{Error: "Tracking failed while getting user."}
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
	U.FillUserAgentUserProperties(userProperties, userAgent)
	// Add user properties from form submit event properties.
	if request.Name == U.EVENT_NAME_FORM_SUBMITTED {
		customerUserId, errCode := M.FillUserPropertiesAndGetCustomerUserIdFromFormSubmit(
			projectId, request.UserId, userProperties, eventProperties)
		if errCode == http.StatusInternalServerError {
			log.WithFields(log.Fields{"userProperties": userProperties,
				"eventProperties": eventProperties}).WithError(err).Error(
				"Failed adding user properties from form submitted event.")
			response.Error = "Failed adding user properties."
		}

		if customerUserId != "" {
			errCode, _ := SDKIdentify(projectId, &SDKIdentifyPayload{
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

	userPropertiesId, errCode := M.UpdateUserProperties(projectId, request.UserId, &postgres.Jsonb{userPropsJSON})
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		log.WithFields(log.Fields{"userProperties": userProperties,
			log.ErrorKey: errCode}).Error("Update user properties on track failed. DB update failed.")
		response.Error = "Failed updating user properties."
	}

	event := &M.Event{
		EventNameId:      eventName.ID,
		CustomerEventId:  request.CustomerEventId,
		Timestamp:        request.Timestamp,
		Properties:       postgres.Jsonb{eventPropsJSON},
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
			return errCode, &SDKTrackResponse{Error: "Tracking failed. Unable to associate with a session."}
		}

		event.SessionId = &session.ID
	}

	createdEvent, errCode := M.CreateEvent(event)
	if errCode == http.StatusFound {
		return errCode, &SDKTrackResponse{Error: "Tracking failed. Event creation failed. Duplicate CustomerEventID",
			CustomerEventId: request.CustomerEventId}
	} else if errCode != http.StatusCreated {
		return errCode, &SDKTrackResponse{Error: "Tracking failed. Event creation failed."}
	}

	// Todo: Try to use latest user properties, if available already.
	existingUserProperties, errCode := M.GetUserPropertiesAsMap(projectId, event.UserId)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error(
			"Failed to get user properties for adding first event properties on track.")
	}

	errCode = enrichAfterTrack(projectId, createdEvent, existingUserProperties)
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

func SDKIdentify(projectId uint64, request *SDKIdentifyPayload) (int, gin.H) {
	// Todo(Dinesh): Add a mandatory field validator and move this.
	// Precondition: Fails to identify if customer_user_id not present.
	if request.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		return http.StatusBadRequest, gin.H{"error": "Identification failed. Missing mandatory keys c_uid."}
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"user_id": request.UserId, "customer_user_id": request.CustomerUserId})

	userProperties, err := getIdentifiedUserPropertiesAsJsonb(request.CustomerUserId)
	if err != nil || userProperties == nil {
		logCtx.WithError(err).Error("Failed to get and add identified user properties on identify.")
	}

	// Precondition: customer_user_id present, user_id not.
	// if customer_user has user already : respond with same user.
	// else : creating a new_user with the given customer_user_id and respond with new_user_id.
	if request.UserId == "" {
		response := gin.H{}

		userLatest, errCode := M.GetUserLatestByCustomerUserId(projectId, request.CustomerUserId)
		switch errCode {
		case http.StatusInternalServerError:
			return errCode, gin.H{"error": "Identification failed. Processing without user_id failed."}

		case http.StatusNotFound:
			newUser := M.User{ProjectId: projectId,
				CustomerUserId: request.CustomerUserId,
				JoinTimestamp:  request.JoinTimestamp,
			}
			if userProperties != nil {
				newUser.Properties = *userProperties
			}

			_, errCode := M.CreateUser(&newUser)
			if errCode != http.StatusCreated {
				return errCode, gin.H{"error": "Identification failed. User creation failed."}
			}
			response["user_id"] = newUser.ID

		case http.StatusFound:
			response["user_id"] = userLatest.ID
		}

		response["message"] = "User has been identified successfully."
		return http.StatusOK, response
	}

	scopeUser, errCode := M.GetUser(projectId, request.UserId)
	if errCode != http.StatusFound {
		return errCode, gin.H{"error": "Identification failed. Invalid user_id."}
	}

	// Precondition: Given user already identified as given customer_user.
	if scopeUser.CustomerUserId == request.CustomerUserId {
		return http.StatusOK, gin.H{"message": "Identified already."}
	}

	// Precondition: user is already identified with different customer_user.
	// Creating a new user with the given customer_user_id and respond with new_user_id.
	if scopeUser.CustomerUserId != "" {
		newUser := M.User{
			ProjectId:      projectId,
			CustomerUserId: scopeUser.CustomerUserId,
			JoinTimestamp:  request.JoinTimestamp,
		}
		if userProperties != nil {
			newUser.Properties = *userProperties
		}

		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, gin.H{"error": "Identification failed. User creation failed."}
		}

		return http.StatusOK, gin.H{"user_id": newUser.ID, "message": "User has been identified successfully"}
	}

	// Happy path. Maps customer_user to an user.
	updateUser := &M.User{CustomerUserId: request.CustomerUserId,
		JoinTimestamp: request.JoinTimestamp}
	if userProperties != nil {
		updateUser.Properties = *userProperties
	}

	_, errCode = M.UpdateUser(projectId, request.UserId, updateUser)
	if errCode != http.StatusAccepted {
		return errCode, gin.H{"error": "Identification failed. Failed mapping customer_user to user"}
	}

	errCode = M.UpdateUserJoinTimePropertyForCustomerUser(projectId, request.CustomerUserId)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode, gin.H{"error": "Identification failed."}
	}

	return http.StatusOK, gin.H{"message": "User has been identified successfully."}
}

func sdkAddUserProperties(projectId uint64, request *sdkAddUserPropertiesPayload, clientIP string) (int, gin.H) {
	// Validate properties.
	validProperties := U.GetValidatedUserProperties(&request.Properties)
	_ = M.FillLocationUserProperties(validProperties, clientIP)
	propertiesJSON, err := json.Marshal(validProperties)
	if err != nil {
		return http.StatusBadRequest, gin.H{"error": "Add user properties failed. Invalid properties."}
	}

	// Precondition: user_id not given.
	if request.UserId == "" {
		// Create user with properties and respond user_id. Only properties allowed on create.
		newUser, errCode := M.CreateUser(&M.User{ProjectId: projectId,
			Properties: postgres.Jsonb{propertiesJSON}})
		if errCode != http.StatusCreated {
			return errCode, gin.H{"error": "Add user properties failed. User create failed"}
		}
		return http.StatusOK, gin.H{"user_id": newUser.ID, "message": "User properties added successfully."}
	}

	user, errCode := M.GetUser(projectId, request.UserId)
	if errCode == http.StatusNotFound {
		return http.StatusBadRequest, gin.H{"error": "Add user properties failed. Invalid user_id."}
	} else if errCode == http.StatusInternalServerError {
		return errCode, gin.H{"error": "Add user properties failed"}
	}

	_, errCode = M.UpdateUserPropertiesByCurrentProperties(projectId, user.ID,
		user.PropertiesId, &postgres.Jsonb{propertiesJSON})
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		return errCode, gin.H{"error": "Add user properties failed."}
	}

	return http.StatusOK, gin.H{"message": "Added user properties successfully."}
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/event/track -d '{"user_id": "YOUR_USER_ID", "event_name": "login", "auto": false, "event_properties": {"ip": "10.0.0.1", "mobile": true}, "user_properties": {"$os": "Mac OS"}}'
func SDKTrackHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Missing request body."})
		return
	}

	var trackPayload SDKTrackPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&trackPayload); err != nil {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid payload."})
		return
	}

	// Get ProjecId scope for the request.
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tracking failed. Invalid project."})
		return
	}

	c.JSON(SDKTrack(projectId, &trackPayload, c.ClientIP(), c.Request.UserAgent(), false))
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: PROJECT_TOKEN" -X POST http://localhost:8080/sdk/event/bulk -d '[{"user_id": "YOUR_USER_ID", "event_name": "login", "auto": false, "event_properties": {"ip": "10.0.0.1", "mobile": true}, "user_properties": {"$os": "Mac OS"}}]'
func SDKBulkEventHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Missing request body."})
		return
	}

	var sdkTrackPayloads []SDKTrackPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sdkTrackPayloads); err != nil {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid payload."})
		return
	}

	if len(sdkTrackPayloads) > 50000 {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Tracking failed. Invalid payload. Request Exceeds more than 1000 events."})
		return
	}

	// Get ProjecId scope for the request.
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tracking failed. Invalid project."})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	response := make([]*SDKTrackResponse, len(sdkTrackPayloads), len(sdkTrackPayloads))
	hasError := false

	for i, sdkTrackPayload := range sdkTrackPayloads {
		errCode, resp := SDKTrack(projectId, &sdkTrackPayload, clientIP, userAgent, false)
		if errCode != http.StatusOK {
			hasError = true
		}
		response[i] = resp
	}

	respCode := http.StatusOK
	if hasError {
		respCode = http.StatusInternalServerError
	}

	c.JSON(respCode, response)
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/user/identify -d '{"user_id":"USER_ID", "c_uid": "CUSTOMER_USER_ID"}'
func SDKIdentifyHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Missing request body."})
		return
	}

	var request SDKIdentifyPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Identification failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Invalid payload."})
		return
	}

	// Get ProjecId scope for the request.
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Identification failed. Invalid project."})
		return
	}

	c.JSON(SDKIdentify(projectId, &request))
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/user/add_properties -d '{"id": "USER_ID", "properties": {"name": "USER_NAME"}}'
func SDKAddUserPropertiesHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Adding user properities failed. Missing request body."})
		return
	}

	var request sdkAddUserPropertiesPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Add user properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Add user properties failed. Invalid payload."})
		return
	}

	// Get ProjecId scope for the request.
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Add user properties failed. Invalid project."})
		return
	}

	c.JSON(sdkAddUserProperties(projectId, &request, c.ClientIP()))
}

type sdkSettingsResponse struct {
	AutoTrack       *bool `json:"auto_track"`
	AutoFormCapture *bool `json:"auto_form_capture"`
	IntSegment      *bool `json:"int_segment"`
	ExcludeBot      *bool `json:"exclude_bot"`
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X GET http://localhost:8080/sdk/project/get_settings
func SDKGetProjectSettingsHandler(c *gin.Context) {
	// Get ProjecId scope for the request.
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get project settings failed. Invalid project."})
		return
	}

	projectSetting, errCode := M.GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get project settings failed."})
		return
	}

	response := sdkSettingsResponse{
		AutoTrack:       projectSetting.AutoTrack,
		AutoFormCapture: projectSetting.AutoFormCapture,
		IntSegment:      projectSetting.IntSegment,
		ExcludeBot:      projectSetting.ExcludeBot,
	}

	c.JSON(http.StatusOK, response)
}

func SDKUpdateEventProperties(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Updating event properities failed. Missing request body."})
		return
	}

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Update event properties failed. Invalid project."})
		return
	}

	logCtx = logCtx.WithField("project_id", projectId)

	var request sdkUpdateEventPropertiesPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Update event properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Update event properties failed. Invalid payload."})
		return
	}

	updateAllowedProperties := U.GetUpdateAllowedEventProperties(&request.Properties)
	validatedProperties := U.GetValidatedEventProperties(updateAllowedProperties)
	if len(*validatedProperties) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No valid properties given to update."})
	}

	errCode := M.UpdateEventProperties(projectId, request.EventId, validatedProperties)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
		return
	}

	updatedEvent, errCode := M.GetEventById(projectId, request.EventId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
		return
	}

	logCtx = logCtx.WithField("event_id", request.EventId)

	if updatedEvent.SessionId == nil || *updatedEvent.SessionId == "" {
		logCtx.Error("Session id does not exist to update session properties on update event properties.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Update event properties failed. No session associated."})
		return
	}

	newSessionProperties := U.GetSessionProperties(false, validatedProperties, &U.PropertiesMap{})
	sessionEvent, errCode := M.GetEventById(projectId, *updatedEvent.SessionId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
		return
	}

	updatedEventPropertiesMap, err := U.DecodePostgresJsonb(&updatedEvent.Properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal updated event properties on update event properties.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Update event properties failed."})
		return
	}

	sessionPropertiesMap, err := U.DecodePostgresJsonb(&sessionEvent.Properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal existing session properties on update event properties.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Update event properties failed."})
		return
	}

	eventPageURL, eventPageURLExists := (*updatedEventPropertiesMap)[U.EP_PAGE_RAW_URL]

	sessionInitialPageURL, sessionInitialPageURLExists := (*sessionPropertiesMap)[U.UP_INITIAL_PAGE_RAW_URL]
	if !eventPageURLExists || !sessionInitialPageURLExists {
		logCtx.Error("No page URL property to compare for session properties update.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Update event properties failed."})
		return
	}

	// session properties updated only if page raw url and initial
	// page raw url matches.
	if eventPageURL != sessionInitialPageURL {
		c.JSON(http.StatusAccepted, gin.H{"message": "Updated event properties successfully."})
		return
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
		errCode := M.UpdateEventProperties(projectId, sessionEvent.ID, &updateSesssionProperties)
		if errCode != http.StatusAccepted {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
			return
		}
	}

	newUserProperties := U.GetInitialUserProperties(validatedProperties)
	user, errCode := M.GetUser(projectId, updatedEvent.UserId)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get user properties of user on update event properties.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
		return
	}

	userPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	if err != nil {
		logCtx.Error("Failed to unmarshal existing user properties on update event properties.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
		return
	}

	userInitialPageURL, userInitialPageURLExists := (*userPropertiesMap)[U.UP_INITIAL_PAGE_RAW_URL]
	if !userInitialPageURLExists {
		// skip error for old users.
		initialPageRawUrlAvailabilityTimestamp := 1569369600
		if user.JoinTimestamp < int64(initialPageRawUrlAvailabilityTimestamp) {
			c.JSON(http.StatusAccepted, gin.H{"message": "Updated event properties successfully."})
			return
		}

		logCtx.Error("No initial page raw url on user properties to compare on update event properties.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
		return
	}

	// user properties updated only if initial_raw_page url of user_properties
	// and raw_page_url of event properties matches.
	if userInitialPageURL != eventPageURL {
		c.JSON(http.StatusAccepted, gin.H{"message": "Updated event properties successfully."})
		return
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
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
			return
		}

		_, errCode := M.UpdateUserProperties(projectId, updatedEvent.UserId, userPropertiesJsonb)
		if errCode != http.StatusAccepted {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Update event properties failed."})
			return
		}
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Updated event properties successfully."})
}
