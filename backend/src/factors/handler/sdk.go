package handler

import (
	"encoding/json"
	C "factors/config"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type sdkTrackPayload struct {
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

type sdkIdentifyPayload struct {
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

func sdkTrack(projectId uint64, request *sdkTrackPayload, clientIP, userAgent string) (int, *SDKTrackResponse) {
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
		if eventNameErrCode != http.StatusFound {
			// create a auto tracked event name if no filter_expr match.
			eventName, eventNameErrCode = M.CreateOrGetAutoTrackedEventName(
				&M.EventName{Name: request.Name, ProjectId: projectId})
		}

		err := M.FillEventPropertiesByFilterExpr(&request.EventProperties, eventName.FilterExpr, request.Name)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "filter_expr": eventName.FilterExpr,
				"event_url": request.Name, log.ErrorKey: err}).Error(
				"Failed to fill event url properties for auto tracked event.")
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
	definedEventProperties := U.MapEventPropertiesToDefinedProperties(&request.EventProperties)
	eventProperties := U.GetValidatedEventProperties(definedEventProperties)
	if ip, ok := (*eventProperties)[U.EP_INTERNAL_IP]; ok && ip != "" {
		clientIP = ip.(string)
	}
	// Added IP to event proerties for internal usage.
	(*eventProperties)[U.EP_INTERNAL_IP] = clientIP
	eventPropsJSON, err := json.Marshal(eventProperties)
	if err != nil {
		return http.StatusBadRequest, &SDKTrackResponse{Error: "Tracking failed. Invalid properties."}
	}

	response := &SDKTrackResponse{}

	// Precondition: if user_id not given, create new user and respond.
	isUserFirstSession := request.IsNewUser
	if request.UserId == "" {
		// initial user properties defined from event properties on user create.
		initialUserProperties := U.GetInitialUserProperties(eventProperties)
		initialUserPropsJSON, err := json.Marshal(initialUserProperties)
		if err != nil {
			log.WithFields(log.Fields{"initialUserProperties": initialUserProperties,
				log.ErrorKey: err}).Error("Add initial user properties failed. JSON marshal failed.")
			response.Error = "Failed adding initial user properties."
		}

		newUser := M.User{ProjectId: projectId, Properties: postgres.Jsonb{initialUserPropsJSON}}
		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, &SDKTrackResponse{Error: "Tracking failed. User creation failed."}
		}

		request.UserId = newUser.ID
		response.UserId = newUser.ID
		isUserFirstSession = true
	}

	userProperties := U.GetValidatedUserProperties(&request.UserProperties)
	_ = M.FillLocationUserProperties(userProperties, clientIP)
	U.FillUserAgentUserProperties(userProperties, userAgent)
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

	session, errCode := M.CreateOrGetSessionEvent(projectId, request.UserId, isUserFirstSession, request.Timestamp,
		eventProperties, userProperties, userPropertiesId)
	if errCode != http.StatusCreated && errCode != http.StatusFound {
		response.Error = "Failed to associate with a session."
	}

	if session == nil || errCode == http.StatusNotFound || errCode == http.StatusInternalServerError {
		log.Error("Session is nil even after CreateOrGetSessionEvent.")
		return errCode, &SDKTrackResponse{Error: "Tracking failed. Unable to associate with a session."}
	}
	currentSessionId := session.ID

	createdEvent, errCode := M.CreateEvent(&M.Event{
		EventNameId:      eventName.ID,
		CustomerEventId:  request.CustomerEventId,
		Timestamp:        request.Timestamp,
		Properties:       postgres.Jsonb{eventPropsJSON},
		ProjectId:        projectId,
		UserId:           request.UserId,
		UserPropertiesId: userPropertiesId,
		SessionId:        &currentSessionId,
	})

	if errCode == http.StatusFound {
		return errCode, &SDKTrackResponse{Error: "Tracking failed. Event creation failed. Duplicate CustomerEventID",
			CustomerEventId: request.CustomerEventId}
	} else if errCode != http.StatusCreated {
		return errCode, &SDKTrackResponse{Error: "Tracking failed. Event creation failed."}
	}

	// Success response.
	response.EventId = createdEvent.ID
	response.Message = "User event tracked successfully."
	response.CustomerEventId = request.CustomerEventId
	return http.StatusOK, response
}

func sdkIdentify(projectId uint64, request *sdkIdentifyPayload) (int, gin.H) {
	// Todo(Dinesh): Add a mandatory field validator and move this.
	// Precondition: Fails to identify if customer_user_id not present.
	if request.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		return http.StatusBadRequest, gin.H{"error": "Identification failed. Missing mandatory keys c_uid."}
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
		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, gin.H{"error": "Identification failed. User creation failed."}
		}

		return http.StatusOK, gin.H{"user_id": newUser.ID, "message": "User has been identified successfully"}
	}

	// Happy path. Maps customer_user to an user.
	_, errCode = M.UpdateUser(projectId, request.UserId, &M.User{CustomerUserId: request.CustomerUserId,
		JoinTimestamp: request.JoinTimestamp})
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

	var trackPayload sdkTrackPayload

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

	c.JSON(sdkTrack(projectId, &trackPayload, c.ClientIP(), c.Request.UserAgent()))
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

	var sdkTrackPayloads []sdkTrackPayload
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
		errCode, resp := sdkTrack(projectId, &sdkTrackPayload, clientIP, userAgent)
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

	var request sdkIdentifyPayload

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

	c.JSON(sdkIdentify(projectId, &request))
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

	c.JSON(http.StatusOK, projectSetting)
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

	c.JSON(http.StatusAccepted, gin.H{"message": "Updated event properties successfully."})
}
