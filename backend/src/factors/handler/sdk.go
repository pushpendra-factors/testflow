package handler

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type IdentifiedUser struct {
	UserId         string `json:"user_id"`
	CustomerUserId string `json:"CustomerUserId"`
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

	// Get ProjecId scope for the request.
	scopeProjectIdIntf := U.GetScopeByKey(c, "projectId")
	if scopeProjectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tracking failed. Invalid project."})
		return
	}
	scopeProjectId := scopeProjectIdIntf.(uint64)

	var response gin.H
	// Create new user if user_id not specified on payload.
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

	event.ProjectId = scopeProjectId
	_, errCode = M.CreateEvent(&event)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatus(errCode)
	} else {
		if response == nil {
			response = gin.H{"event_id": event.ID}
		} else {
			response["event_id"] = event.ID
		}
		response["message"] = "User event tracked successfully."
		c.JSON(http.StatusOK, response)
	}
}

//Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/sdk/user/identify -d '{"user_id":"user_id", "c_uid": "customer_user_id"}'
func SDKIdentifyHandler(c *gin.Context) {
	r := c.Request

	identifiedUser := IdentifiedUser{}
	err := json.NewDecoder(r.Body).Decode(&identifiedUser)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Identification failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Invalid payload."})
		return
	}

	if identifiedUser.UserId == "" || identifiedUser.CustomerUserId == "" {
		log.Error("Identification failed. Missing user_id or c_uid.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Identification failed. Missing mandatory keys user_id or c_uid."})
		return
	}

	// Get ProjecId scope for the request.
	scopeProjectIdIntf := U.GetScopeByKey(c, "projectId")
	if scopeProjectIdIntf == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tracking failed. Invalid project."})
		return
	}
	scopeProjectId := scopeProjectIdIntf.(uint64)

	customerUserId, errCode := M.GetCustomerUserIdById(scopeProjectId, identifiedUser.UserId)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Finding user failed."})
		return
	}

	// Handled if user already identified as customer.
	if customerUserId != "" {
		newUser := M.User{ProjectId: scopeProjectId, CustomerUserId: customerUserId}
		_, errCode := M.CreateUser(&newUser)
		if errCode != M.DB_SUCCESS {
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. User creation failed."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": newUser.ID})
		return
	}

	// Updates userId by customerId
	errCode = M.UpdateCustomerUserIdById(scopeProjectId, identifiedUser.CustomerUserId, identifiedUser.UserId)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Identification failed. Failed mapping customer_user to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been identified successfully"})
}
