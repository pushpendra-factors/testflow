package handler

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type IdentifiedUser struct {
	UserId         string `json:"user_id"`
	CustomerUserId string `json:"c_uid"`
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/event/track -d '{"event_name": "login", "properties": {"ip": "10.0.0.1", "mobile": true}}'
func SDKTrackHandler(c *gin.Context) {
	r := c.Request

	if r.Body == nil {
		log.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Missing request body."})
		return
	}

	var event M.Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tracking failed. Invalid payload."})
		return
	}

	// Precondition: Fails if event_name not provided.
	event.EventName = strings.TrimSpace(event.EventName) // Discourage whitespace on the end.
	if event.EventName == "" {
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
		if errCode != M.DB_SUCCESS {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Tracking failed. User creation failed."})
			return
		}
		event.UserId = newUser.ID
		response = gin.H{"user_id": newUser.ID}
	}

	_, errCode := M.CreateOrGetEventName(&M.EventName{Name: event.EventName, ProjectId: scopeProjectId})
	if errCode != http.StatusConflict && errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Tracking failed. EventName creation failed."})
		return
	}

	// Create Event.
	event.ProjectId = scopeProjectId
	_, errCode = M.CreateEvent(&event)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Tracking failed. Event creation failed."})
	} else {
		response["event_id"] = event.ID
		response["message"] = "User event tracked successfully."
		c.JSON(http.StatusOK, response)
	}
}

//Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: c6b4useqmywdrvo7m8kqtqn2htvglvgj" -X POST http://localhost:8080/sdk/user/identify -d '{"user_id":"5624438c-18c4-4969-961e-078b4e83e516", "c_uid": "cid-001"}'
func SDKIdentifyHandler(c *gin.Context) {
	r := c.Request

	if r.Body == nil {
		log.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Missing request body."})
		return
	}

	identifiedUser := IdentifiedUser{}
	err := json.NewDecoder(r.Body).Decode(&identifiedUser)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Identification failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Invalid payload."})
		return
	}

	// Precondition: Fails to identify if customer_user_id not present.
	if identifiedUser.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Missing mandatory keys user_id or c_uid."})
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
	// if customer_user has user already : using the same user.
	// else : creating a new_user with the given customer_user_id and respond with new_user_id.
	if identifiedUser.UserId == "" {
		response := gin.H{}

		userLatest, errCode := M.GetUserLatestByCustomerUserId(scopeProjectId, identifiedUser.CustomerUserId)
		switch errCode {
		case http.StatusInternalServerError:
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Processing without user_id failed."})
			return

		case http.StatusNotFound:
			newUser := M.User{ProjectId: scopeProjectId, CustomerUserId: identifiedUser.CustomerUserId}
			_, errCode := M.CreateUser(&newUser)
			if errCode != M.DB_SUCCESS {
				c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. User creation failed."})
				return
			}
			response["user_id"] = newUser.ID

		case M.DB_SUCCESS:
			response["user_id"] = userLatest.ID
		}

		response["message"] = "User has been identified successfully."
		c.JSON(errCode, response)
		return
	}

	scopeUser, errCode := M.GetUser(scopeProjectId, identifiedUser.UserId)
	if errCode != M.DB_SUCCESS {
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
		newUser := M.User{ProjectId: scopeProjectId, CustomerUserId: scopeUser.CustomerUserId}
		_, errCode := M.CreateUser(&newUser)
		if errCode != M.DB_SUCCESS {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. User creation failed."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user_id": newUser.ID, "message": "User has been identified successfully"})
		return
	}

	// Happy path. Maps customer_user to an user.
	errCode = M.UpdateCustomerUserIdById(scopeProjectId, identifiedUser.UserId, identifiedUser.CustomerUserId)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Failed mapping customer_user to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been identified successfully."})
}
