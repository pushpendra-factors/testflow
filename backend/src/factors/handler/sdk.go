package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	C "factors/config"
	mid "factors/middleware"
	M "factors/model"
	SDK "factors/sdk"
	U "factors/util"
)

func SDKStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "I'm ok."})
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
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.TrackResponse{Error: "Tracking failed. Missing request body."})
		return
	}

	var request SDK.TrackPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.TrackResponse{Error: "Tracking failed. Invalid payload."})
		return
	}

	if request.EventId != "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.TrackResponse{Error: "Tracking failed. Invalid payload."})
		return
	}

	// Add client_ip and user_agent from context
	// to track request.
	request.ClientIP = c.ClientIP()
	request.UserAgent = c.Request.UserAgent()

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	c.JSON(SDK.TrackWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens()))
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
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.TrackResponse{Error: "Tracking failed. Missing request body."})
		return
	}

	var sdkTrackPayloads []SDK.TrackPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sdkTrackPayloads); err != nil {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.TrackResponse{Error: "Tracking failed. Invalid payload."})
		return
	}

	if len(sdkTrackPayloads) > 50000 {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge,
			&SDK.TrackResponse{Error: "Tracking failed. Invalid payload. Request Exceeds more than 1000 events."})
		return
	}

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	response := make([]*SDK.TrackResponse, len(sdkTrackPayloads), len(sdkTrackPayloads))
	hasError := false

	for i, sdkTrackPayload := range sdkTrackPayloads {
		sdkTrackPayload.ClientIP = clientIP
		sdkTrackPayload.UserAgent = userAgent

		errCode, resp := SDK.TrackWithQueue(projectToken, &sdkTrackPayload, C.GetSDKRequestQueueAllowedTokens())
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
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.IdentifyResponse{Error: "Identification failed. Missing request body."})
		return
	}

	var request SDK.IdentifyPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Identification failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.IdentifyResponse{Error: "Identification failed. Invalid payload."})
		return
	}

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	c.JSON(SDK.IdentifyWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens()))
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
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.AddUserPropertiesResponse{Error: "Adding user properities failed. Missing request body."})
		return
	}

	var request SDK.AddUserPropertiesPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Add user properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.AddUserPropertiesResponse{Error: "Add user properties failed. Invalid payload."})
		return
	}

	request.ClientIP = c.ClientIP()

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	c.JSON(SDK.AddUserPropertiesWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens()))
}

type sdkSettingsResponse struct {
	AutoTrack       *bool `json:"auto_track"`
	AutoFormCapture *bool `json:"auto_form_capture"`
	ExcludeBot      *bool `json:"exclude_bot"`
}

// Test command.
// curl -i -H "Content-Type: application/json" -H "Authorization: YOUR_TOKEN" -X GET http://localhost:8080/sdk/project/get_settings
func SDKGetProjectSettingsHandler(c *gin.Context) {
	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	projectSetting, errCode := M.GetProjectSettingByTokenWithCacheAndDefault(projectToken)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Get project settings failed."})
		return
	}

	response := sdkSettingsResponse{
		AutoTrack:       projectSetting.AutoTrack,
		AutoFormCapture: projectSetting.AutoFormCapture,
		ExcludeBot:      projectSetting.ExcludeBot,
	}

	c.JSON(http.StatusOK, response)
}

func SDKUpdateEventPropertiesHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &SDK.UpdateEventPropertiesResponse{
			Error: "Updating event properities failed. Missing request body."})
		return
	}

	var request SDK.UpdateEventPropertiesPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Update event properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &SDK.UpdateEventPropertiesResponse{
			Error: "Update event properties failed. Invalid payload."})
		return
	}

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	c.JSON(SDK.UpdateEventPropertiesWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens()))
}
