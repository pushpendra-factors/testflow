package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	response := make([]*SDK.TrackResponse, len(sdkTrackPayloads), len(sdkTrackPayloads))
	hasError := false

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)

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
		c.AbortWithStatusJSON(errCode, &SDK.Response{Error: "Get project settings failed."})
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

	request.UserAgent = c.Request.UserAgent()

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	c.JSON(SDK.UpdateEventPropertiesWithQueue(projectToken, &request,
		C.GetSDKRequestQueueAllowedTokens()))
}

/*
AMPSDKTrackHandler - Tracks event from AMP Pages with query params

Sample Track URL
https://app.factors.ai/sdk/amp/event/track?token=${token}&title=${title}&referrer=${documentReferrer}
&screen_height=${screenHeight}&screen_width=${screenWidth}&page_load_time_in_ms=${pageLoadTime}
&client_id=${clientId(_factorsai_amp_id)}&source_url=${sourceUrl}
*/
func SDKAMPTrackHandler(c *gin.Context) {
	token := c.Query("token")
	token = strings.TrimSpace(token)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			&SDK.Response{Error: "Track failed. Invalid token"})
		return
	}

	logCtx := log.WithField("token", token)

	settings, errCode := M.GetProjectSettingByTokenWithCacheAndDefault(token)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Track failed. Invalid request."})
		return
	}
	if !*settings.AutoTrack {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Message: "Track failed. Not enabled."})
		return
	}

	ampClientId := c.Query("client_id")
	ampClientId = strings.TrimSpace(ampClientId)
	if ampClientId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Track failed. Invalid client_id."})
		return
	}

	logCtx = logCtx.WithField("client_id", ampClientId)

	sourceURL := c.Query("source_url")
	sourceURL = strings.TrimSpace(sourceURL)

	paramReferrerURL := c.Query("referrer")
	paramReferrerURL = strings.TrimSpace(paramReferrerURL)

	pageTitle := c.Query("title")

	var pageLoadTimeInSecs float64
	paramPageLoadTime := c.Query("page_load_time_in_ms")
	paramPageLoadTime = strings.TrimSpace(paramPageLoadTime)
	pageLoadTimeInMs, err := strconv.ParseFloat(paramPageLoadTime, 64)
	if paramPageLoadTime != "" && err != nil {
		logCtx.WithError(err).WithField("page_load_time_in_ms", paramPageLoadTime).Error(
			"Failed to convert page_load_time to number on amp sdk track")
	}
	if pageLoadTimeInMs > 0 {
		pageLoadTimeInSecs = pageLoadTimeInMs / 1000
	}

	paramScreenHeight := c.Query("screen_height")
	screenHeight, err := strconv.ParseFloat(paramScreenHeight, 64)
	if paramScreenHeight != "" && err != nil {
		logCtx.WithError(err).WithField("screen_height", paramScreenHeight).Error(
			"Failed to convert screen_height to number on amp sdk track")
	}

	paramScreenWidth := c.Query("screen_width")
	screenWidth, err := strconv.ParseFloat(paramScreenWidth, 64)
	if paramScreenWidth != "" && err != nil {
		logCtx.WithError(err).WithField("screen_width", paramScreenWidth).Error(
			"Failed to convert screen_width to number on amp sdk track")
	}

	payload := &SDK.AMPTrackPayload{
		ClientID:           ampClientId,
		SourceURL:          sourceURL,
		Title:              pageTitle,
		Referrer:           paramReferrerURL,
		ScreenHeight:       screenHeight,
		ScreenWidth:        screenWidth,
		PageLoadTimeInSecs: pageLoadTimeInSecs,

		Timestamp: time.Now().Unix(), // request timestamp.
		UserAgent: c.Request.UserAgent(),
		ClientIP:  c.ClientIP(),
	}

	c.JSON(SDK.AMPTrackWithQueue(token, payload, C.GetSDKRequestQueueAllowedTokens()))
}

func SDKAMPUpdateEventPropertiesHandler(c *gin.Context) {
	token := c.Query("token")
	token = strings.TrimSpace(token)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			&SDK.Response{Error: "Update event properties failed. Invalid token"})
		return
	}

	logCtx := log.WithField("token", token)

	ampClientId := c.Query("client_id")
	ampClientId = strings.TrimSpace(ampClientId)
	if ampClientId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Update event properties failed. Invalid client_id"})
		return
	}

	spentTime := c.Query("page_spent_time")
	spentTime = strings.TrimSpace(spentTime)
	pageSpentTime, err := strconv.ParseFloat(spentTime, 64)

	scrollPercent := c.Query("page_scroll_percent")
	scrollPercent = strings.TrimSpace(scrollPercent)
	pageScrollPercent, err := strconv.ParseFloat(scrollPercent, 64)
	if scrollPercent != "" && err != nil {
		logCtx.WithError(err).WithField("page_scroll_percent", pageScrollPercent).Error(
			"Failed to convert scroll percent to number on amp sdk track")
	}

	sourceURL := c.Query("source_url")
	sourceURL = strings.TrimSpace(sourceURL)

	payload := &SDK.AMPUpdateEventPropertiesPayload{
		ClientID:          ampClientId,
		SourceURL:         sourceURL,
		PageScrollPercent: pageScrollPercent,
		PageSpentTime:     pageSpentTime,

		Timestamp: time.Now().Unix(), // request timestamp.
		UserAgent: c.Request.UserAgent(),
	}

	c.JSON(SDK.AMPUpdateEventPropertiesWithQueue(token, payload, C.GetSDKRequestQueueAllowedTokens()))
}

type SDKError struct {
	UserId       string `json:"user_id"`
	URL          string `json:"url"`
	AutoTrackURL string `json:"auto_track_url"`
	Domain       string `json:"domain"`
	Error        string `json:"error"`
}

func SDKErrorHandler(c *gin.Context) {
	var request SDKError

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		log.WithError(err).Error("Failed to unmarshal SDK Error.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	properties := make(U.PropertiesMap, 0)
	SDK.FillUserAgentUserProperties(&properties, c.Request.UserAgent())

	// Error logged for adding it to error email.
	log.WithFields(log.Fields{
		"domain":         request.Domain,
		"error":          request.Error,
		"url":            request.URL,
		"auto_track_url": request.AutoTrackURL,
		"properties":     properties,
		"tag":            "sdk_error",
	}).Info("Got JS SDK Error.")

	c.AbortWithStatus(http.StatusOK)
	return
}
