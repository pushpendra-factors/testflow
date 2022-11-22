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
	"factors/metrics"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	"factors/sdk"
	SDK "factors/sdk"
	"factors/util"
	U "factors/util"
)

// SDKStatusHandler godoc
// @Summary To check the status and availability of the sdk service.
// @Tags SDK
// @Accept  json
// @Produce json
// @Success 200 {string} json "{"message": "I'm ok."}"
// @Router /sdk/service/status [get]
func SDKStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "I'm ok."})
}

// Test command.
// curl -i -H "Content-UnitType: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/event/track -d '{"user_id": "YOUR_USER_ID", "event_name": "login", "auto": false, "event_properties": {"ip": "10.0.0.1", "mobile": true}, "user_properties": {"$os": "Mac OS"}}'
// SDKTrackHandler godoc
// @Summary Create a new track request.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body sdk.TrackPayload true "Track payload"
// @Success 200 {object} sdk.TrackResponse
// @Router /sdk/event/track [post]
// @Security ApiKeyAuth
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

	request.RequestSource = model.UserSourceWeb
	// Add client_ip and user_agent from context
	// to track request.
	request.ClientIP = c.ClientIP()
	request.UserAgent = c.Request.UserAgent()

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	status, response := SDK.TrackWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens())
	if status == http.StatusOK {
		metrics.Increment(metrics.IncrSDKRequestOverallCount)
		metrics.Increment(metrics.IncrSDKRequestTypeTrack)
	}
	c.JSON(status, response)
}

// Test command.
// curl -i -H "Content-UnitType: application/json" -H "Authorization: PROJECT_TOKEN" -X POST http://localhost:8080/sdk/event/bulk -d '[{"user_id": "YOUR_USER_ID", "event_name": "login", "auto": false, "event_properties": {"ip": "10.0.0.1", "mobile": true}, "user_properties": {"$os": "Mac OS"}}]'
// SDKBulkEventHandler godoc
// @Summary Create a new bulk events track request.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body []sdk.TrackPayload true "Array of Track payload"
// @Success 200 {array} sdk.TrackResponse
// @Router /sdk/event/track/bulk [post]
// @Security ApiKeyAuth
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
		sdkTrackPayload.RequestSource = model.UserSourceWeb

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
// curl -i -H "Content-UnitType: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/user/identify -d '{"user_id":"USER_ID", "c_uid": "CUSTOMER_USER_ID"}'
// SDKIdentifyHandler godoc
// @Summary To identify a factors user id with customer user id.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body sdk.IdentifyPayload true "Identify payload"
// @Success 200 {object} sdk.IdentifyResponse
// @Router /sdk/user/identify [post]
// @Security ApiKeyAuth
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
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Identification failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.IdentifyResponse{Error: "Identification failed. Invalid payload."})
		return
	}

	request.RequestSource = model.UserSourceWeb
	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	status, response := SDK.IdentifyWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens())
	if status == http.StatusOK {
		metrics.Increment(metrics.IncrSDKRequestOverallCount)
		metrics.Increment(metrics.IncrSDKRequestTypeIdentifyUser)
	}
	c.JSON(status, response)
}

// Test command.
// curl -i -H "Content-UnitType: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/user/add_properties -d '{"id": "USER_ID", "properties": {"name": "USER_NAME"}}'
// SDKAddUserPropertiesHandler godoc
// @Summary To update properties of a factors user.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body sdk.AddUserPropertiesPayload true "Add user properties payload"
// @Success 200 {object} sdk.AddUserPropertiesResponse
// @Router /sdk/user/add_properties [post]
// @Security ApiKeyAuth
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
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Add user properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.AddUserPropertiesResponse{Error: "Add user properties failed. Invalid payload."})
		return
	}

	request.ClientIP = c.ClientIP()
	request.RequestSource = model.UserSourceWeb

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	status, response := SDK.AddUserPropertiesWithQueue(projectToken, &request, C.GetSDKRequestQueueAllowedTokens())
	if status == http.StatusOK {
		metrics.Increment(metrics.IncrSDKRequestOverallCount)
		metrics.Increment(metrics.IncrSDKRequestTypeAddUserProperties)
	}
	c.JSON(status, response)
}

type sdkGetInfoPayload struct {
	UserID string `json:"user_id"`
}

type sdkGetInfoResponse struct {
	AutoTrack              *bool  `json:"auto_track"`
	AutoTrackSPAPageView   *bool  `json:"auto_track_spa_page_view"`
	AutoFormCapture        *bool  `json:"auto_form_capture"`
	AutoClickCapture       *bool  `json:"auto_click_capture"`
	ExcludeBot             *bool  `json:"exclude_bot"`
	AutoFormFillCapture    *bool  `json:"auto_capture_form_fills"`
	IntDrift               *bool  `json:"int_drift"`
	IntClearBit            *bool  `json:"int_clear_bit"`
	IntClientSixSignalKey  *bool  `json:"int_client_six_signal_key"`
	IntFactorsSixSignalKey *bool  `json:"int_factors_six_signal_key"`
	UserID                 string `json:"user_id,omitempty"`
}

// Test command.
// curl -i -H "Content-UnitType: application/json" -H "Authorization: YOUR_TOKEN" -X POST http://localhost:8080/sdk/get_info -d '{"user_id": "YOUR_USER_ID"}'
// SDKGetSettingsAndUserIDInfoHandler godoc
// @Summary To get project settings and user_id (if not available). Used only by JS SDK.
// @Tags SDK
// @Accept  json
// @Produce json
// @Success 200 {object} handler.sdkGetInfoResponse
// @Router /sdk/project/get_info [get]
// @Security ApiKeyAuth
func SDKGetInfoHandler(c *gin.Context) {
	r := c.Request
	logCtx := log.WithField("reqId", U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID))

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Get info request failed. Missing request body."})
		return
	}

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	logCtx = logCtx.WithField("token", projectToken)

	var request sdkGetInfoPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Failed decoding get info payload.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Failed decoding get info payload. Invalid payload."})
		return
	}

	projectSetting, errCode := store.GetStore().GetProjectSettingByTokenWithCacheAndDefault(projectToken)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, &SDK.Response{Error: "Get info failed."})
		return
	}

	response := sdkGetInfoResponse{
		AutoTrack:            projectSetting.AutoTrack,
		AutoTrackSPAPageView: projectSetting.AutoTrackSPAPageView,
		AutoFormCapture:      projectSetting.AutoFormCapture,
		AutoClickCapture:     projectSetting.AutoClickCapture,
		AutoFormFillCapture:  projectSetting.AutoCaptureFormFills,
		ExcludeBot:           projectSetting.ExcludeBot,
		IntDrift:             projectSetting.IntDrift,
		IntClearBit:          projectSetting.IntClearBit,
	}

	// Adds new user_id to response, if not available on the request.
	// The responded user_id is expected to be set on all JS_SDK requests.
	if request.UserID == "" {
		response.UserID = U.GetUUID()
	}

	c.JSON(http.StatusOK, response)
}

// DEPRECATED: Current JS_SDK is using /get_info instead of /get_settings.
// For backward compatibility, old NPM installations might be using /get_settings.
func SDKGetProjectSettingsHandler(c *gin.Context) {
	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)

	projectSetting, errCode := store.GetStore().GetProjectSettingByTokenWithCacheAndDefault(projectToken)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, &SDK.Response{Error: "Get project settings failed."})
		return
	}

	response := sdkGetInfoResponse{
		AutoTrack:            projectSetting.AutoTrack,
		AutoTrackSPAPageView: projectSetting.AutoTrackSPAPageView,
		AutoFormCapture:      projectSetting.AutoFormCapture,
		AutoClickCapture:     projectSetting.AutoClickCapture,
		AutoFormFillCapture:  projectSetting.AutoCaptureFormFills,
		ExcludeBot:           projectSetting.ExcludeBot,
		IntDrift:             projectSetting.IntDrift,
		IntClearBit:          projectSetting.IntClearBit,
	}

	c.JSON(http.StatusOK, response)
}

// SDKUpdateEventPropertiesHandler godoc
// @Summary To update event properties for an existing event.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body sdk.UpdateEventPropertiesPayload true "Update properties payload"
// @Success 202 {object} sdk.UpdateEventPropertiesResponse
// @Router /sdk/event/update_properties [post]
// @Security ApiKeyAuth
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
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Update event properties failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &SDK.UpdateEventPropertiesResponse{
			Error: "Update event properties failed. Invalid payload."})
		return
	}

	request.UserAgent = c.Request.UserAgent()
	request.RequestSource = model.UserSourceWeb

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	status, response := SDK.UpdateEventPropertiesWithQueue(projectToken, &request,
		C.GetSDKRequestQueueAllowedTokens())
	if status == http.StatusOK {
		metrics.Increment(metrics.IncrSDKRequestOverallCount)
		metrics.Increment(metrics.IncrSDKRequestTypeUpdateEventProperties)
	}
	c.JSON(status, response)
}

/*
AMPSDKTrackHandler - Tracks event from AMP Pages with query params

Sample Track URL
https://app.factors.ai/sdk/amp/event/track?token=${token}&title=${title}&referrer=${documentReferrer}
&screen_height=${screenHeight}&screen_width=${screenWidth}&page_load_time_in_ms=${pageLoadTime}
&client_id=${clientId(_factorsai_amp_id)}&source_url=${sourceUrl}
*/
// SDKAMPTrackHandler godoc
// @Summary Create a new AMP track request.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param token query string true "SDK token"
// @Param client_id query string true "Client id"
// @Param source_url query string false "Source url"
// @Param referrer query string false "Referrer"
// @Param title query string false "Title"
// @Param page_load_time_in_ms query number false "Page load time in milliseconds"
// @Param screen_height query number false "Screen height"
// @Param screen_width query number false "Screen width"
// @Success 200 {object} sdk.Response
// @Router /sdk/amp/event/track [post]
func SDKAMPTrackHandler(c *gin.Context) {

	// List of known query parameters.
	TOKEN := "token"
	CLIENT_ID := "client_id"
	SOURCE_URL := "source_url"
	REFERRER := "referrer"
	TITLE := "title"
	PAGE_LOAD_TIME_IN_MS := "page_load_time_in_ms"
	SCREEN_HEIGHT := "screen_height"
	SCREEN_WIDTH := "screen_width"
	EVENT_NAME := "event_name"

	trackedProperties := make(map[string]bool)
	token := c.Query(TOKEN)
	token = strings.TrimSpace(token)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			&SDK.Response{Error: "Track failed. Invalid token"})
		return
	}

	trackedProperties[TOKEN] = true
	logCtx := log.WithField(TOKEN, token)

	settings, errCode := store.GetStore().GetProjectSettingByTokenWithCacheAndDefault(token)
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

	ampClientId := c.Query(CLIENT_ID)
	ampClientId = strings.TrimSpace(ampClientId)
	if ampClientId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Track failed. Invalid client_id."})
		return
	}

	trackedProperties[CLIENT_ID] = true
	logCtx = logCtx.WithField(CLIENT_ID, ampClientId)

	sourceURL := c.Query(SOURCE_URL)
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			&SDK.Response{Error: "Track failed. Invalid source_url."})
		return
	}
	trackedProperties[SOURCE_URL] = true

	paramReferrerURL := c.Query(REFERRER)
	paramReferrerURL = strings.TrimSpace(paramReferrerURL)
	trackedProperties[REFERRER] = true

	pageTitle := c.Query(TITLE)
	trackedProperties[TITLE] = true

	var pageLoadTimeInSecs float64
	paramPageLoadTime := c.Query(PAGE_LOAD_TIME_IN_MS)
	paramPageLoadTime = strings.TrimSpace(paramPageLoadTime)
	pageLoadTimeInMs, err := strconv.ParseFloat(paramPageLoadTime, 64)
	trackedProperties[PAGE_LOAD_TIME_IN_MS] = true
	if paramPageLoadTime != "" && err != nil {
		logCtx.WithError(err).WithField(PAGE_LOAD_TIME_IN_MS, paramPageLoadTime).Error(
			"Failed to convert page_load_time to number on amp sdk track")
	}
	if pageLoadTimeInMs > 0 {
		pageLoadTimeInSecs = pageLoadTimeInMs / 1000
	}

	paramScreenHeight := c.Query(SCREEN_HEIGHT)
	screenHeight, err := strconv.ParseFloat(paramScreenHeight, 64)
	trackedProperties[SCREEN_HEIGHT] = true
	if paramScreenHeight != "" && err != nil {
		logCtx.WithError(err).WithField(SCREEN_HEIGHT, paramScreenHeight).Error(
			"Failed to convert screen_height to number on amp sdk track")
	}

	paramScreenWidth := c.Query(SCREEN_WIDTH)
	screenWidth, err := strconv.ParseFloat(paramScreenWidth, 64)
	trackedProperties[SCREEN_WIDTH] = true
	if paramScreenWidth != "" && err != nil {
		logCtx.WithError(err).WithField(SCREEN_WIDTH, paramScreenWidth).Error(
			"Failed to convert screen_width to number on amp sdk track")
	}

	queryParams := c.Request.URL.Query()
	payload := &SDK.AMPTrackPayload{
		ClientID:           ampClientId,
		SourceURL:          sourceURL,
		Title:              pageTitle,
		Referrer:           paramReferrerURL,
		ScreenHeight:       screenHeight,
		ScreenWidth:        screenWidth,
		PageLoadTimeInSecs: pageLoadTimeInSecs,

		Timestamp:     util.TimeNowUnix(), // request timestamp.
		UserAgent:     c.Request.UserAgent(),
		ClientIP:      c.ClientIP(),
		RequestSource: model.UserSourceWeb,
	}
	customProperties := make(map[string]interface{})

	for k, v := range queryParams {
		if k == EVENT_NAME {
			payload.EventName = v[0]
		} else {
			if !trackedProperties[k] {
				customProperties[k] = v[0]
			}
		}
	}
	payload.CustomProperties = customProperties
	status, response := SDK.AMPTrackWithQueue(token, payload, C.GetSDKRequestQueueAllowedTokens())
	if status == http.StatusOK {
		metrics.Increment(metrics.IncrSDKRequestOverallCount)
		metrics.Increment(metrics.IncrSDKRequestTypeAMPTrack)
	}
	c.JSON(status, response)
}

// SDKAMPUpdateEventPropertiesHandler godoc
// @Summary To update AMP event properties for an existing event.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param token query string true "SDK token"
// @Param client_id query string true "Client id"
// @Param source_url query string false "Source url"
// @Param page_spent_time query number false "Page spent time"
// @Param page_scroll_percent query number false "Page scroll percent"
// @Success 202 {object} sdk.Response
// @Router /sdk/amp/event/update_properties [post]
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

		Timestamp:     time.Now().Unix(), // request timestamp.
		UserAgent:     c.Request.UserAgent(),
		RequestSource: model.UserSourceWeb,
	}

	status, response := SDK.AMPUpdateEventPropertiesWithQueue(token, payload, C.GetSDKRequestQueueAllowedTokens())
	if status == http.StatusOK {
		metrics.Increment(metrics.IncrSDKRequestOverallCount)
		metrics.Increment(metrics.IncrSDKRequestTypeAMPUpdateEventProperties)
	}
	c.JSON(status, response)
}

// SDKAMPIdentifyHandler Test command.
// curl -i GET 'http://localhost:8085/sdk/amp/user/identify?token=<token>&client_id=<amp_id>&customer_user_id=<customer_user_id>
// SDKAMPIdentifyHandler godoc
// @Summary To identify factors user with customer user id from AMP pages.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param token query string true "SDK token"
// @Param client_id query string true "Client id"
// @Param customer_user_id query string false "Customer user id"
// @Success 200 {object} sdk.IdentifyResponse
// @Router /sdk/amp/user/identify [post]
func SDKAMPIdentifyHandler(c *gin.Context) {
	token := c.Query("token")
	customerUserID := c.Query("customer_user_id")
	clientID := c.Query("client_id")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, &SDK.Response{Error: "Identificational failed. Missing token"})
		return
	}

	if customerUserID == "" || clientID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, &SDK.Response{Error: "Identificational failed. Missing required params"})
		return
	}

	var payload SDK.AMPIdentifyPayload
	payload.CustomerUserID = customerUserID
	payload.ClientID = clientID
	payload.Timestamp = util.TimeNowUnix()
	payload.RequestSource = model.UserSourceWeb

	c.JSON(SDK.AMPIdentifyWithQueue(token, &payload, C.GetSDKRequestQueueAllowedTokens()))
}

type SDKError struct {
	UserId       string `json:"user_id"`
	URL          string `json:"url"`
	AutoTrackURL string `json:"auto_track_url"`
	Domain       string `json:"domain"`
	Error        string `json:"error"`
}

// SDKErrorHandler godoc
// @Summary To report errors on SDK requests.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param error body handler.SDKError true "Error payload"
// @Success 200
// @Router /sdk/service/error [post]
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

// SDKCaptureClickAsEvent godoc
// @Summary To capture clicks as events.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body CaptureClickPayload true "Capture click payload"
// @Success 200 {object} sdk.CaptureClickResponse
// @Router /sdk/capture_click [post]
func SDKCaptureClickHandler(c *gin.Context) {
	/* Behaviour: Clicks recorded on clickable_elements. Queued for tracking events once enabled.
	* Clicks discovery (collecting clickable_elements) does not use queue. Non-critical data. Queue
	will be bloated, if there is spurious click event, which is more likely.
	* If a clickable_element is enabled, then a custom track call will be queued with existing
	queue logic.
	* If database is down: the enabled/disabled data can be fetched from redis and custom track
	queue will keep working. This requires caching and refreshing cache on change of
	clickable_elements state. */

	r := c.Request
	logCtx := log.WithField("reqId", U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID))

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.CaptureClickResponse{
			Error: "Updating click failed. Missing request body."})
		return
	}

	var request model.CaptureClickPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Updating click failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.CaptureClickResponse{
			Error: "Updating click failed. Invalid payload."})
		return
	}

	// TODO: Move clickable_elements to use project_token and implement
	// cache based config to make the flow working when the db is down.
	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	projectID, errCode := store.GetStore().GetProjectIDByToken(projectToken)
	if errCode == http.StatusNotFound {
		if SDK.IsValidTokenString(projectToken) {
			logCtx.Error("Failed to get project from sdk project token.")
		} else {
			logCtx.WithField("token", projectToken).Warn("Invalid token on sdk payload.")
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized,
			&model.CaptureClickResponse{Error: "Update click failed. Invalid Token."})
		return
	}

	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, &model.CaptureClickResponse{Error: "Update click failed."})
		return
	}

	var response *model.CaptureClickResponse
	isEnabled, status, err := store.GetStore().UpsertCountAndCheckEnabledClickableElement(projectID, &request)
	if err != nil {
		logCtx.WithError(err).Error("Failed to capture click.")
		response = &model.CaptureClickResponse{Error: "Failed to capture click."}
		c.AbortWithStatusJSON(status, response)
		return
	}

	if !isEnabled {
		c.JSON(status, &model.CaptureClickResponse{Message: "Captured click successfully."})
		return
	}

	model.AddAllowedElementAttributes(projectID,
		request.ElementAttributes, &request.EventProperties)

	payload := sdk.TrackPayload{
		UserId:          request.UserID,
		Name:            strings.TrimSpace(request.DisplayName),
		EventProperties: request.EventProperties,
		UserProperties:  request.UserProperties,
		Auto:            false,
		RequestSource:   model.UserSourceWeb,
		Timestamp:       util.TimeNowUnix(),
	}

	trackStatus, trackResponse := sdk.TrackWithQueue(projectToken, &payload,
		C.GetSDKRequestQueueAllowedTokens())
	trackResponse.Message = "Tracked click as event."
	c.JSON(trackStatus, trackResponse)
}

func SDKFormFillHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.CaptureFormFillResponse{
			Error: "Form fill failed. Missing request body."})
		return
	}

	var request model.SDKFormFillPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Form fill event failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.CaptureFormFillResponse{
			Error: "Form fill event failed. Invalid payload."})
		return
	}

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	if !SDK.IsValidTokenString(projectToken) {
		logCtx.Error("Form fill event failed. Token invalid")
		c.AbortWithStatusJSON(http.StatusUnauthorized, &model.CaptureFormFillResponse{Error: "Form fill event failed, Unauthorized."})
		return

	}
	projectID, errCode := store.GetStore().GetProjectIDByToken(projectToken)
	if errCode == http.StatusNotFound {
		logCtx.WithField("token", projectToken).Warn("Invalid token on sdk payload.")
		c.AbortWithStatusJSON(http.StatusUnauthorized, &model.CaptureFormFillResponse{Error: "Form fill event failed. Invalid Token."})
		return
	}
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, &model.CaptureFormFillResponse{Error: "Form fill event failed."})
		return
	}

	var response *model.CaptureFormFillResponse

	status, err := store.GetStore().CreateFormFillEventById(projectID, &request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.CaptureFormFillResponse{Error: "Creation of form fill event by projectId failed."})
		return
	} else {
		response = &model.CaptureFormFillResponse{Message: "Form fill event successful."}
	}

	c.JSON(status, response)
}
