package handler

import (
	"encoding/json"
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
	EventProperties U.PropertiesMap `json:"event_properties"`
	UserProperties  U.PropertiesMap `json:"user_properties"`
	ProjectId       uint64          `json:"project_id"`
	UserId          string          `json:"user_id"`
	Auto            bool            `json:"auto"`
	Timestamp       int64           `json:"timestamp`
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

func sdkTrack(projectId uint64, request *sdkTrackPayload, clientIP string) (int, gin.H) {
	// Precondition: Fails if event_name not provided.
	request.Name = strings.TrimSpace(request.Name) // Discourage whitespace on the end.
	if request.Name == "" {
		return http.StatusBadRequest, gin.H{"error": "Tracking failed. Event name cannot be omitted or left empty."}
	}

	response := gin.H{}

	// Precondition: if user_id not given, create new user and respond.
	if request.UserId == "" {
		newUser := M.User{ProjectId: projectId}
		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			return errCode, gin.H{"error": "Tracking failed. User creation failed."}
		}
		request.UserId = newUser.ID
		response = gin.H{"user_id": newUser.ID}
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
			eventName, eventNameErrCode = M.CreateOrGetAutoTrackedEventName(&M.EventName{Name: request.Name, ProjectId: projectId})
		}

		err := M.FillEventPropertiesByFilterExpr(&request.EventProperties, eventName.FilterExpr, request.Name)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "filter_expr": eventName.FilterExpr,
				"event_url": request.Name, "error": err}).Error("Failed to fill event url properties for auto tracked event.")
		}
	} else {
		eventName, eventNameErrCode = M.CreateOrGetUserCreatedEventName(&M.EventName{Name: request.Name, ProjectId: projectId})
	}

	if eventNameErrCode != http.StatusCreated && eventNameErrCode != http.StatusConflict &&
		eventNameErrCode != http.StatusFound {
		return eventNameErrCode, gin.H{"error": "Tracking failed. Creating event_name failed."}
	}

	// Validate properties.
	validEventProperties := U.GetValidatedEventProperties(&request.EventProperties)
	eventPropsJSON, err := json.Marshal(validEventProperties)
	if err != nil {
		return http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid properties."}
	}

	// Update user properties on track.
	if ip, ok := (*validEventProperties)[U.EP_INTERNAL_IP]; ok && ip != "" {
		clientIP = ip.(string)
	}

	// Added IP to event proerties for internal usage.
	(*validEventProperties)[U.EP_INTERNAL_IP] = clientIP

	validUserProperties := U.GetValidatedUserProperties(&request.UserProperties)
	_ = M.FillLocationUserProperties(validUserProperties, clientIP)

	userPropsJSON, err := json.Marshal(validUserProperties)
	if err != nil {
		log.WithFields(log.Fields{"userProperties": validUserProperties,
			"error": err}).Error("Update user properites on track failed. Unmarshal json failed.")
		response["error"] = "Failed updating user properties."
	}

	userPropertiesId, errCode := M.UpdateUserProperties(projectId, request.UserId, &postgres.Jsonb{userPropsJSON})
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		log.WithFields(log.Fields{"userProperties": validUserProperties,
			"error": errCode}).Error("Update user properties on track failed. DB update failed.")
		response["error"] = "Failed updating user properties."
	}

	// Create Event.
	createdEvent, errCode := M.CreateEvent(&M.Event{
		EventNameId: eventName.ID,
		Timestamp:   request.Timestamp,
		Properties:  postgres.Jsonb{eventPropsJSON},
		ProjectId:   projectId, UserId: request.UserId, UserPropertiesId: userPropertiesId})
	if errCode != http.StatusCreated {
		return errCode, gin.H{"error": "Tracking failed. Event creation failed."}
	}

	// Success response.
	response["event_id"] = createdEvent.ID
	response["message"] = "User event tracked successfully."
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

	// Todo(Dinesh): Make UpdateUser to return 404 on 0 rows affected and remove this.
	scopeUser, errCode := M.GetUser(projectId, request.UserId)
	if errCode != http.StatusFound {
		return errCode, gin.H{"error": "Add user properties failed. Invalid user_id."}

	}

	if _, errCode = M.UpdateUser(projectId, scopeUser.ID,
		&M.User{Properties: postgres.Jsonb{propertiesJSON}}); errCode != http.StatusAccepted {
		return errCode, gin.H{"error": "Add user properties failed."}
	}

	return http.StatusOK, gin.H{"message": "Added user properties successfully."}
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/event/track -d '{"user_id": "YOUR_USER_ID", "event_name": "login", "auto": false, "event_properties": {"ip": "10.0.0.1", "mobile": true}, "user_properties": {"$os": "Mac OS"}}'
func SDKTrackHandler(c *gin.Context) {
	r := c.Request

	if r.Body == nil {
		log.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Missing request body."})
		return
	}

	var trackPayload sdkTrackPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&trackPayload); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid payload."})
		return
	}

	// Get ProjecId scope for the request.
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tracking failed. Invalid project."})
		return
	}

	c.JSON(sdkTrack(projectId, &trackPayload, c.ClientIP()))
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/user/identify -d '{"user_id":"USER_ID", "c_uid": "CUSTOMER_USER_ID"}'
func SDKIdentifyHandler(c *gin.Context) {
	r := c.Request

	if r.Body == nil {
		log.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Missing request body."})
		return
	}

	var request sdkIdentifyPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Identification failed. JSON Decoding failed.")
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

	if r.Body == nil {
		log.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Adding user properities failed. Missing request body."})
		return
	}

	var request sdkAddUserPropertiesPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Add user properties failed. JSON Decoding failed.")
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
