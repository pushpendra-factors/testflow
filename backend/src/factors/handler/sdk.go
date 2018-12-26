package handler

import (
	"encoding/json"
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

type sdkAddPropertiesPayload struct {
	UserId     string          `json:"user_id"`
	Properties U.PropertiesMap `json:"properties"`
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/event/track -d '{"user_id": "YOUR_USER_ID", "event_name": "login", "event_properties": {"ip": "10.0.0.1", "mobile": true}, "user_properties": {"$os": "Mac OS"}}'
func SDKTrackHandler(c *gin.Context) {
	r := c.Request

	if r.Body == nil {
		log.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Missing request body."})
		return
	}

	var event sdkTrackPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid payload."})
		return
	}

	// Precondition: Fails if event_name not provided.
	event.Name = strings.TrimSpace(event.Name) // Discourage whitespace on the end.
	if event.Name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Tracking failed. Event name cannot be omitted or left empty.")
		return
	}

	// Get ProjecId scope for the request.
	scopeProjectIdIntf := U.GetScopeByKey(c, "projectId")
	if scopeProjectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tracking failed. Invalid project."})
		return
	}
	scopeProjectId := scopeProjectIdIntf.(uint64)

	response := gin.H{}

	// Precondition: if user_id not given, create new user and respond.
	if event.UserId == "" {
		newUser := M.User{ProjectId: scopeProjectId}
		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Tracking failed. User creation failed."})
			return
		}
		event.UserId = newUser.ID
		response = gin.H{"user_id": newUser.ID}
	}

	var eventName *M.EventName
	var errCode int
	// if auto_track enabled, auto_name = event_name else auto_name = "UCEN".
	if event.Auto {
		eventName, errCode = M.CreateOrGetEventName(&M.EventName{Name: event.Name,
			AutoName: event.Name, ProjectId: scopeProjectId})
	} else {
		eventName, errCode = M.CreateOrGetUserCreatedEventName(&M.EventName{Name: event.Name, ProjectId: scopeProjectId})
	}
	if errCode != http.StatusConflict && errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Tracking failed. EventName creation failed."})
		return
	}

	// Validate properties.
	validEventProperties := U.GetValidatedEventProperties(&event.EventProperties)
	eventPropsJSON, err := json.Marshal(validEventProperties)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid properties."})
		return
	}
	// Update user properties on track.
	var clientIP string
	if ip, ok := (*validEventProperties)[U.EP_INTERNAL_IP]; ok && ip != "" {
		clientIP = ip.(string)
	} else {
		clientIP = c.ClientIP()
	}
	// Added IP to event proerties for internal usage.
	(*validEventProperties)[U.EP_INTERNAL_IP] = clientIP

	validUserProperties := U.GetValidatedUserProperties(&event.UserProperties)
	_ = M.FillLocationUserProperties(validUserProperties, clientIP)

	userPropsJSON, err := json.Marshal(validUserProperties)
	if err != nil {
		log.WithFields(log.Fields{"userProperties": validUserProperties,
			"error": err}).Error("Update user properites on track failed. Unmarshal json failed.")
		response["error"] = "Failed updating user properties."
	}

	userPropertiesId, errCode := M.UpdateUserProperties(scopeProjectId, event.UserId, &postgres.Jsonb{userPropsJSON})
	if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
		log.WithFields(log.Fields{"userProperties": validUserProperties,
			"error": errCode}).Error("Update user properties on track failed. DB update failed.")
		response["error"] = "Failed updating user properties."
	}

	// Create Event.
	event.ProjectId = scopeProjectId
	createdEvent, errCode := M.CreateEvent(&M.Event{
		EventNameId: eventName.ID,
		Timestamp:   event.Timestamp,
		Properties:  postgres.Jsonb{eventPropsJSON},
		ProjectId:   scopeProjectId, UserId: event.UserId, UserPropertiesId: userPropertiesId})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Tracking failed. Event creation failed."})
		return
	}

	response["event_id"] = createdEvent.ID
	response["message"] = "User event tracked successfully."

	c.JSON(http.StatusOK, response)
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

	var identifiedUser sdkIdentifyPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&identifiedUser); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Identification failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Invalid payload."})
		return
	}

	// Todo(Dinesh): Add a mandatory field validator and move this.
	// Precondition: Fails to identify if customer_user_id not present.
	if identifiedUser.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Missing mandatory keys c_uid."})
		return
	}

	// Get ProjecId scope for the request.
	scopeProjectIdIntf := U.GetScopeByKey(c, "projectId")
	if scopeProjectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Identification failed. Invalid project."})
		return
	}
	scopeProjectId := scopeProjectIdIntf.(uint64)

	// Precondition: customer_user_id present, user_id not.
	// if customer_user has user already : respond with same user.
	// else : creating a new_user with the given customer_user_id and respond with new_user_id.
	if identifiedUser.UserId == "" {
		response := gin.H{}

		userLatest, errCode := M.GetUserLatestByCustomerUserId(scopeProjectId, identifiedUser.CustomerUserId)
		switch errCode {
		case http.StatusInternalServerError:
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Processing without user_id failed."})
			return

		case http.StatusNotFound:
			newUser := M.User{ProjectId: scopeProjectId,
				CustomerUserId: identifiedUser.CustomerUserId,
				JoinTimestamp:  identifiedUser.JoinTimestamp,
			}
			_, errCode := M.CreateUser(&newUser)
			if errCode != http.StatusCreated {
				c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. User creation failed."})
				return
			}
			response["user_id"] = newUser.ID

		case http.StatusFound:
			response["user_id"] = userLatest.ID
		}

		response["message"] = "User has been identified successfully."
		c.JSON(http.StatusOK, response)
		return
	}

	scopeUser, errCode := M.GetUser(scopeProjectId, identifiedUser.UserId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Invalid user_id."})
		return
	}

	// Precondition: Given user already identified as given customer_user.
	if scopeUser.CustomerUserId == identifiedUser.CustomerUserId {
		c.JSON(http.StatusOK, gin.H{"message": "Identified already."})
		return
	}

	// Precondition: user is already identified with different customer_user.
	// Creating a new user with the given customer_user_id and respond with new_user_id.
	if scopeUser.CustomerUserId != "" {
		newUser := M.User{
			ProjectId:      scopeProjectId,
			CustomerUserId: scopeUser.CustomerUserId,
			JoinTimestamp:  identifiedUser.JoinTimestamp,
		}
		_, errCode := M.CreateUser(&newUser)
		if errCode != http.StatusCreated {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. User creation failed."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user_id": newUser.ID, "message": "User has been identified successfully"})
		return
	}

	// Happy path. Maps customer_user to an user.
	_, errCode = M.UpdateUser(scopeProjectId, identifiedUser.UserId,
		&M.User{CustomerUserId: identifiedUser.CustomerUserId,
			JoinTimestamp: identifiedUser.JoinTimestamp})
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Failed mapping customer_user to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been identified successfully."})
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

	var addPropsUser sdkAddPropertiesPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&addPropsUser); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Add user properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Add user properties failed. Invalid payload."})
		return
	}

	// Get ProjecId scope for the request.
	scopeProjectIdIntf := U.GetScopeByKey(c, "projectId")
	if scopeProjectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Add user properties failed. Invalid project."})
		return
	}
	scopeProjectId := scopeProjectIdIntf.(uint64)

	// Validate properties.
	validProperties := U.GetValidatedUserProperties(&addPropsUser.Properties)
	_ = M.FillLocationUserProperties(validProperties, c.ClientIP())
	propertiesJSON, err := json.Marshal(validProperties)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Add user properties failed. Invalid properties."})
		return
	}

	// Precondition: user_id not given.
	if addPropsUser.UserId == "" {
		// Create user with properties and respond user_id. Only properties allowed on create.
		newUser, errCode := M.CreateUser(&M.User{ProjectId: scopeProjectId,
			Properties: postgres.Jsonb{propertiesJSON}})
		if errCode != http.StatusCreated {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Add user properties failed. User create failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": newUser.ID, "message": "User properties added successfully."})
		return
	}

	// Todo(Dinesh): Make UpdateUser to return 404 on 0 rows affected and remove this.
	scopeUser, errCode := M.GetUser(scopeProjectId, addPropsUser.UserId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Add user properties failed. Invalid user_id."})
		return
	}

	if _, errCode = M.UpdateUser(scopeProjectId, scopeUser.ID,
		&M.User{Properties: postgres.Jsonb{propertiesJSON}}); errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Add user properties failed."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added user properties successfully."})
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X GET http://localhost:8080/sdk/project/get_settings
func SDKGetProjectSettings(c *gin.Context) {
	// Get ProjecId scope for the request.
	scopeProjectIdIntf := U.GetScopeByKey(c, "projectId")
	if scopeProjectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Get project settings failed. Invalid project."})
		return
	}
	scopeProjectId := scopeProjectIdIntf.(uint64)

	projectSetting, errCode := M.GetProjectSetting(scopeProjectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get project settings failed."})
		return
	}

	c.JSON(http.StatusOK, projectSetting)
}
