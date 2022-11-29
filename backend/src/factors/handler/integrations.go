package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	C "factors/config"
	IntHubspot "factors/integration/hubspot"
	IntSalesforce "factors/integration/salesforce"
	IntSegment "factors/integration/segment"
	IntShopify "factors/integration/shopify"
	"factors/metrics"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
)

// IntSegmentHandler godoc
// @Summary To create event from segment/rudderstack.
// @Tags SDK,Integrations
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param request body segment.Event true "Event payload"
// @Success 200 {object} segment.EventResponse
// @Router /integrations/segment [post]
// @Security ApiKeyAuth
func IntSegmentHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	var bodyBuffer bytes.Buffer
	body := io.TeeReader(r.Body, &bodyBuffer)

	var event IntSegment.Event
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		logCtx.WithError(err).Error("Segment/Rudderstack JSON decode failed")
	}
	logCtx.WithFields(log.Fields{"event": event}).Debug("Segment/Rudderstack webhook request")

	token := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_PRIVATE_TOKEN)
	if C.IsBlockedSDKRequestProjectToken(token) {
		c.AbortWithStatusJSON(http.StatusOK,
			IntSegment.EventResponse{Error: "Request failed. Blocked."})
		return
	}

	// Debug raw request payload from segment.
	var rawRequestPayload map[string]interface{}
	logDebugCtx := logCtx.WithField("token", token).WithField("tag", "segment_request_payload")
	if err := json.Unmarshal(bodyBuffer.Bytes(), &rawRequestPayload); err != nil {
		logDebugCtx.WithError(err).Info("Failed to decode as raw request.")
	}

	status, response := IntSegment.ReceiveEventWithQueue(token, &event,
		C.GetSegmentRequestQueueAllowedTokens())

	// send error on StatusInternalServerError.
	// Possible when error from redis while queuing, if queue is enabled.
	// Possible on multiple conditions when processing directly.
	if status == http.StatusInternalServerError || status == http.StatusBadRequest {
		c.AbortWithStatusJSON(status, response)
		return
	}

	metrics.Increment(metrics.IncrIntegrationRequestOverallCount)
	// Always send StatusOK for failure on direct processing.
	c.JSON(http.StatusOK, response)
}

// IntSegmentPlatformHandler godoc
// @Summary To create event from segment.
// @Tags SDK,Integrations
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id path integer true "Dashboard ID"
// @Param request body segment.Event true "Event payload"
// @Success 200 {object} segment.EventResponse
// @Router /integrations/segment_platform [post]
// @Security ApiKeyAuth
// Wrapper to support documentation for segment_platform route.
func IntSegmentPlatformHandler(c *gin.Context) {
	IntSegmentHandler(c)
}

type AdwordsAddRefreshTokenPayload struct {
	// project_id conv from string to uint64 explicitly.
	ProjectId    string `json:"project_id"`
	AgentUUID    string `json:"agent_uuid"`
	RefreshToken string `json:"refresh_token"`
}

type GoogleOrganicAddRefreshTokenPayload struct {
	// project_id conv from string to uint64 explicitly.
	ProjectId    string `json:"project_id"`
	AgentUUID    string `json:"agent_uuid"`
	RefreshToken string `json:"refresh_token"`
}
type AdwordsRequestPayload struct {
	ProjectId string `json:"project_id"`
}

type GoogleOrganicRequestPayload struct {
	ProjectID string `json:"project_id"`
}

// Updates projects settings required for Adwords.
// Uses session cookie for auth on middleware.
func IntAdwordsAddRefreshTokenHandler(c *gin.Context) {
	r := c.Request

	var requestPayload AdwordsAddRefreshTokenPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Adwords update payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. update failed."})
		return
	}

	if requestPayload.ProjectId == "" ||
		requestPayload.AgentUUID == "" ||
		requestPayload.RefreshToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid json payload. empty mandatory fields."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid json payload. empty mandatory fields."})
		return
	}

	// Todo: Check agent has access to project or not, before adding refresh token?

	errCode := store.GetStore().UpdateAgentIntAdwordsRefreshToken(requestPayload.AgentUUID, requestPayload.RefreshToken)
	if errCode != http.StatusAccepted {
		log.WithField("agent_uuid", requestPayload.AgentUUID).
			Error("Failed to update adwords refresh token for agent.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating adwords refresh token for agent"})
		return
	}

	_, errCode = store.GetStore().UpdateProjectSettings(projectId,
		&model.ProjectSetting{IntAdwordsEnabledAgentUUID: &requestPayload.AgentUUID})
	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectId).
			Error("Failed to update project settings adwords enable agent uuid.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating adwords enabled agent uuid project settings"})
		return
	}

	c.JSON(errCode, gin.H{})
}

func IntAdwordsGetRefreshTokenHandler(c *gin.Context) {
	r := c.Request

	var requestPayload AdwordsRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Adwords get refresh token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. fetch failed."})
		return
	}

	if requestPayload.ProjectId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	refreshToken, errCode := store.GetStore().GetIntAdwordsRefreshTokenForProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get adwords refresh token for project."})
		return
	}

	c.JSON(http.StatusFound, gin.H{"refresh_token": refreshToken})
}

// Updates projects settings required for Google search console.
// Uses session cookie for auth on middleware.
func IntGoogleOrganicAddRefreshTokenHandler(c *gin.Context) {
	r := c.Request

	var requestPayload GoogleOrganicAddRefreshTokenPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("GoogleOrganic update payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. update failed."})
		return
	}

	if requestPayload.ProjectId == "" ||
		requestPayload.AgentUUID == "" ||
		requestPayload.RefreshToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid json payload. empty mandatory fields."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid json payload. empty mandatory fields."})
		return
	}

	// Todo: Check agent has access to project or not, before adding refresh token?

	errCode := store.GetStore().UpdateAgentIntGoogleOrganicRefreshToken(requestPayload.AgentUUID, requestPayload.RefreshToken)
	if errCode != http.StatusAccepted {
		log.WithField("agent_uuid", requestPayload.AgentUUID).
			Error("Failed to update gsc refresh token for agent.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating gsc refresh token for agent"})
		return
	}

	_, errCode = store.GetStore().UpdateProjectSettings(projectId,
		&model.ProjectSetting{IntGoogleOrganicEnabledAgentUUID: &requestPayload.AgentUUID})
	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectId).
			Error("Failed to update project settings GoogleOrganic enable agent uuid.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating GoogleOrganic enabled agent uuid project settings"})
		return
	}

	c.JSON(errCode, gin.H{})
}
func IntGoogleOrganicGetRefreshTokenHandler(c *gin.Context) {
	r := c.Request

	var requestPayload GoogleOrganicRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("GoogleOrganic get refresh token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. fetch failed."})
		return
	}

	if requestPayload.ProjectID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectID, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	refreshToken, errCode := store.GetStore().GetIntGoogleOrganicRefreshTokenForProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get gsc refresh token for project."})
		return
	}

	c.JSON(http.StatusFound, gin.H{"refresh_token": refreshToken})
}

// IntEnableAdwordsHandler - Checks for refresh_token for the
// agent if exists: then add the agent_uuid as adwords_enabled_agent_uuid
// on project settings. if not exists: return 304.
// IntEnableAdwordsHandler godoc
// @Summary To enable adwords for the project.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param request body AdwordsRequestPayload true "Request payload"
// @Success 200 {object} model.ProjectSetting
// @Router /integrations/adwords/enable [post]
func IntEnableAdwordsHandler(c *gin.Context) {
	r := c.Request

	var requestPayload AdwordsRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Adwords get refresh token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. enable failed."})
		return
	}

	if requestPayload.ProjectId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	agent, errCode := store.GetStore().GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	if agent.IntAdwordsRefreshToken == "" {
		c.JSON(http.StatusNotModified, gin.H{})
		return
	}

	addEnableAgentUUIDSetting := model.ProjectSetting{IntAdwordsEnabledAgentUUID: &currentAgentUUID}
	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &addEnableAgentUUIDSetting)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to enable adwords"})
		return
	}

	c.JSON(http.StatusOK, addEnableAgentUUIDSetting)
}
func IntEnableGoogleOrganicHandler(c *gin.Context) {
	r := c.Request
	var requestPayload AdwordsRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("GoogleOrganic get refresh token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. enable failed."})
		return
	}

	if requestPayload.ProjectId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	agent, errCode := store.GetStore().GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	if agent.IntGoogleOrganicRefreshToken == "" {
		c.JSON(http.StatusNotModified, gin.H{})
		return
	}

	addEnableAgentUUIDSetting := model.ProjectSetting{IntGoogleOrganicEnabledAgentUUID: &currentAgentUUID}
	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &addEnableAgentUUIDSetting)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to enable google search console"})
		return
	}

	c.JSON(http.StatusOK, addEnableAgentUUIDSetting)
}

type SalesforceEnableRequestPayload struct {
	ProjectId string `json:"project_id"`
}

// IntEnableSalesforceHandler - Checks for refresh_token for the
// agent if exists: then add the agent_uuid as int_salesforce_enabled_agent_uuid
// on project settings. if not exists: return 304.
// IntEnableSalesforceHandler godoc
// @Summary To enable Salesforce for a project.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param request body SalesforceEnableRequestPayload true "Request payload"
// @Success 200 {object} model.ProjectSetting
// @Router /integrations/salesforce/enable [post]
func IntEnableSalesforceHandler(c *gin.Context) {
	r := c.Request

	var requestPayload SalesforceEnableRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Salesforce get refresh token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. enable failed."})
		return
	}

	if requestPayload.ProjectId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	agent, errCode := store.GetStore().GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	if agent.IntSalesforceRefreshToken == "" && agent.IntSalesforceInstanceURL == "" {
		c.JSON(http.StatusNotModified, gin.H{"error": "agent not set for salesforce"})
		return
	}

	addEnableAgentUUIDSetting := model.ProjectSetting{IntSalesforceEnabledAgentUUID: &currentAgentUUID}
	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &addEnableAgentUUIDSetting)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to enable salesforce"})
		return
	}

	c.JSON(http.StatusOK, addEnableAgentUUIDSetting)
}

func IntShopifyHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "Shopify webhook failed. Invalid project."})
		return
	}

	if !store.GetStore().IsPSettingsIntShopifyEnabled(projectId) {
		logCtx.WithField("project_id", projectId).Error("Shopify webhook failure. Integration not enabled.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	shopifyTopic := c.Request.Header.Get("X-Shopify-Topic")
	shopifyTopic = strings.TrimSpace(shopifyTopic)

	decoder := json.NewDecoder(r.Body)
	shouldHashEmail := U.GetScopeByKeyAsBool(c, mid.SCOPE_SHOPIFY_HASH_EMAIL)

	switch shopifyTopic {
	case "carts/update":
		var cartUpdated IntShopify.CartObject
		if err := decoder.Decode(&cartUpdated); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Cart Update JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid json payload. cart update failed."})
			return
		}
		if len(cartUpdated.LineItems) == 0 {
			logCtx.WithFields(log.Fields{"project_id": projectId}).Info(
				"Ignoring Shopify Cart Update with zero items")
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromCartObject(
			projectId, IntShopify.ACTION_SHOPIFY_CART_UPDATED, &cartUpdated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Cart Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cart update failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify cart update failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cart update failed."})
		}
		logCtx.WithFields(log.Fields{"shopify cart updated": cartUpdated}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	case "checkouts/create":
		var checkoutCreated IntShopify.CheckoutObject
		if err := decoder.Decode(&checkoutCreated); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Checkout Create JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid json payload. checkout create failed."})
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromCheckoutObject(
			projectId, IntShopify.ACTION_SHOPIFY_CHECKOUT_CREATED, shouldHashEmail, &checkoutCreated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Checkout Create JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "checkout create failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify checkout create failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "checkout create failed."})
		}
		logCtx.WithFields(log.Fields{"shopify checkout created": checkoutCreated}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	case "checkouts/update":
		var checkoutUpdated IntShopify.CheckoutObject
		if err := decoder.Decode(&checkoutUpdated); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Checkout Update JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid json payload. checkout update failed."})
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromCheckoutObject(
			projectId, IntShopify.ACTION_SHOPIFY_CHECKOUT_UPDATED, shouldHashEmail, &checkoutUpdated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Checkout Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "checkout update failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify checkout update failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "checkout update failed."})
		}
		logCtx.WithFields(log.Fields{"shopify checkout updated": checkoutUpdated}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	case "orders/create":
		var orderCreated IntShopify.OrderObject
		if err := decoder.Decode(&orderCreated); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Create JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. order create failed."})
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromOrderObject(
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_CREATED, shouldHashEmail, &orderCreated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Create JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order create failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify order create failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "order create failed."})
		}
		logCtx.WithFields(log.Fields{"shopify order created": orderCreated}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	case "orders/updated":
		var orderUpdated IntShopify.OrderObject
		if err := decoder.Decode(&orderUpdated); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. order update failed."})
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromOrderObject(
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_UPDATED, shouldHashEmail, &orderUpdated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order update failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify order update failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "order update failed."})
		}
		logCtx.WithFields(log.Fields{"shopify order updated": orderUpdated}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	case "orders/paid":
		var orderPaid IntShopify.OrderObject
		if err := decoder.Decode(&orderPaid); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Paid JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid json payload. order paid failed."})
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromOrderObject(
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_PAID, shouldHashEmail, &orderPaid)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order update failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify order update failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "order paid failed."})
		}
		logCtx.WithFields(log.Fields{"shopify order paid": orderPaid}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	case "orders/cancelled":
		var orderCancelled IntShopify.OrderObject
		if err := decoder.Decode(&orderCancelled); err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify order cancelled JSON decode failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid json payload. order cancelled failed."})
			return
		}
		eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err := IntShopify.GetTrackDetailsFromOrderObject(
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_CANCELLED, shouldHashEmail, &orderCancelled)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order cancelled failed."})
			return
		}
		request := &SDK.TrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectId, request, false, SDK.SourceShopify, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status, "error": response.Error}).Error(
				"Shopify order update failure. sdk_track call failed.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "order cancelled failed."})
		}
		logCtx.WithFields(log.Fields{"shopify order cancelled": orderCancelled}).Debug("Shopify webhook request")
		c.JSON(http.StatusOK, response)
	}
}

type FacebookAddAccessTokenPayload struct {
	ProjectId   string `json:"project_id"`
	AccessToken string `json:"int_facebook_access_token"`
	Email       string `json:"int_facebook_email"`
	UserID      string `json:"int_facebook_user_id"`
	AdAccounts  string `json:"int_facebook_ad_account"`
}

type FacebookLongLivedTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// IntFacebookAddAccessTokenHandler godoc
// @Summary To add access token for Facebook.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param request body FacebookAddAccessTokenPayload true "Request payload"
// @Success 202 {string} json "{"error": "Error message"}"
// @Router /integrations/facebook/add_access_token [post]
func IntFacebookAddAccessTokenHandler(c *gin.Context) {
	r := c.Request

	var requestPayload FacebookAddAccessTokenPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Facebook get access token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. enable failed."})
		return
	}

	if requestPayload.ProjectId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	_, errCode := store.GetStore().GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	resp, err := http.Get("https://graph.facebook.com/v13.0/oauth/access_token?" +
		"grant_type=fb_exchange_token&client_id=" + C.GetFacebookAppId() + "&client_secret=" + C.GetFacebookAppSecret() +
		"&fb_exchange_token=" + requestPayload.AccessToken)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.WithError(err).Error("Failed to get long lived access token")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get long lived token"})
		return
	}
	defer resp.Body.Close()
	body := json.NewDecoder(resp.Body)

	var newBody FacebookLongLivedTokenResponse
	err = body.Decode(&newBody)
	if err != nil {
		log.WithError(err).Error("Facebook get long lived access token payload JSON decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. enable failed."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &model.ProjectSetting{
		IntFacebookEmail: requestPayload.Email, IntFacebookAccessToken: newBody.AccessToken,
		IntFacebookAgentUUID: &currentAgentUUID, IntFacebookUserID: requestPayload.UserID,
		IntFacebookAdAccount: requestPayload.AdAccounts, IntFacebookTokenExpiry: time.Now().Unix() + newBody.ExpiresIn})
	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectId).Error("Failed to update project settings with facebook email and access token")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed updating facebook email and access token in project settings"})
		return
	}

	c.JSON(errCode, gin.H{})
}

type LinkedinOauthCode struct {
	Code string `json:"code"`
}
type LinkedinOauthToken struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             uint64 `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn uint64 `json:"refresh_token_expires_in"`
}

/*
auth steps:
1. Get authorization code from client side
2. Use that code to get access token in the function below
*/
func IntLinkedinAuthHandler(c *gin.Context) {
	r := c.Request

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var linkedinCode LinkedinOauthCode
	if err := decoder.Decode(&linkedinCode); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload."})
		return
	}
	redirectURI := url.QueryEscape(C.GetProtocol() + C.GetAPPDomain())
	urloauth := "https://www.linkedin.com/oauth/v2/accessToken?grant_type=authorization_code&code=" + linkedinCode.Code + "&redirect_uri=" + redirectURI + "&client_id=" + C.GetLinkedinClientID() + "&client_secret=" + C.GetLinkedinClientSecret()
	resp, err := http.Get(urloauth)
	if err != nil {
		log.WithError(err).Error("Failed to get access token with golang error")
		c.AbortWithStatusJSON(resp.StatusCode, gin.H{"error": "failed to get access token"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(body)
		log.WithFields(log.Fields{"url": urloauth, "response_body": bodyString}).WithError(err).Error("Failed to get access token with response error")
		c.AbortWithStatusJSON(resp.StatusCode, gin.H{"error": "failed to get access token"})
		return
	}

	defer resp.Body.Close()
	decoder = json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	var responsePayload LinkedinOauthToken
	err = decoder.Decode(&responsePayload)
	if err != nil {
		log.WithError(err).Error("Linkedin access token payload JSON decode failure.", responsePayload)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. enable failed."})
		return
	}
	c.JSON(resp.StatusCode, responsePayload)
}

type LinkedinAccountPayload struct {
	AccessToken string `json:"access_token"`
}

func IntLinkedinAccountHandler(c *gin.Context) {
	r := c.Request

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var linkedinAccountPayload LinkedinAccountPayload
	if err := decoder.Decode(&linkedinAccountPayload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload."})
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.linkedin.com/v2/adAccountsV2?q=search", nil)
	req.Header.Set("Authorization", "Bearer "+linkedinAccountPayload.AccessToken)
	resp, err := client.Do(req)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read access token from response")
	}
	bodyString := string(bodyBytes)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.WithError(err).Error("Failed to get access token")
		c.AbortWithStatusJSON(resp.StatusCode, gin.H{"error": "failed to get access token"})
		return
	}
	c.JSON(resp.StatusCode, bodyString)
}

type LinkedinAccessTokenPayload struct {
	AccessToken        string `json:"int_linkedin_access_token"`
	RefreshToken       string `json:"int_linkedin_refresh_token"`
	RefreshTokenExpiry int64  `json:"int_linkedin_refresh_token_expiry"`
	AccessTokenExpiry  int64  `json:"int_linkedin_access_token_expiry"`
	ProjectID          string `json:"project_id"`
	AdAccount          string `json:"int_linkedin_ad_account"`
}

func IntLinkedinAddAccessTokenHandler(c *gin.Context) {
	r := c.Request

	var requestPayload LinkedinAccessTokenPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload."})
		return
	}

	projectId, err := strconv.ParseInt(requestPayload.ProjectID, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	_, errCode := store.GetStore().GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	_, errCode = store.GetStore().UpdateProjectSettings(projectId, &model.ProjectSetting{
		IntLinkedinAdAccount: requestPayload.AdAccount, IntLinkedinAccessToken: requestPayload.AccessToken,
		IntLinkedinAgentUUID: &currentAgentUUID, IntLinkedinAccessTokenExpiry: time.Now().Unix() + requestPayload.AccessTokenExpiry,
		IntLinkedinRefreshToken: requestPayload.RefreshToken, IntLinkedinRefreshTokenExpiry: time.Now().Unix() + requestPayload.RefreshTokenExpiry,
	})
	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectId).Error("Failed to update project settings with linkedin fields")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed updating linkedin fields in project settings"})
		return
	}

	c.JSON(errCode, gin.H{})
}
func IntShopifySDKHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "Shopify webhook failed. Invalid project."})
		return
	}

	if !store.GetStore().IsPSettingsIntShopifyEnabled(projectId) {
		logCtx.WithField("project_id", projectId).Error("Shopify sdk failure. Integration not enabled.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Request failed. Invalid token or project."})
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var cartTokenPayload IntShopify.CartTokenPayload
	if err := decoder.Decode(&cartTokenPayload); err != nil {
		logCtx.WithFields(log.Fields{log.ErrorKey: err}).Error(
			"Shopify Cart Token Payload decode failed")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload."})
		return
	}

	log.WithFields(log.Fields{"cart payload": cartTokenPayload}).Info("Received Cart Token")
	if cartTokenPayload.CartToken == "" || cartTokenPayload.UserId == "" {
		logCtx.WithFields(log.Fields{
			"cart token": cartTokenPayload.CartToken, "user id": cartTokenPayload.UserId}).Error(
			"Shopify Cart Token Payload decode failed")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "empty token or userId."})
	}
	errCode := model.SetCacheShopifyCartTokenToUserId(
		projectId, cartTokenPayload.CartToken, cartTokenPayload.UserId)
	if errCode != http.StatusOK && errCode != http.StatusCreated {
		c.AbortWithStatus(errCode)
		return
	}
	c.JSON(http.StatusOK, nil)
}

// SalesforceCallbackRoute holds oauth redirect route
const SalesforceCallbackRoute = "/salesforce/auth/callback"

type SalesforceRedirectRequestPayload struct {
	ProjectID string `json:"project_id"`
}

// GetSalesforceRedirectURL return the redirect URL based on environment
func GetSalesforceRedirectURL() string {
	return C.GetProtocol() + C.GetAPIDomain() + ROUTE_INTEGRATIONS_ROOT + SalesforceCallbackRoute
}

// SalesforceCallbackHandler handles the callback url from salesforce auth redirect url and requests access token
// SalesforceCallbackHandler godoc
// @Summary Handles the callback url from salesforce auth redirect url and requests access token.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param code query string true "Code"
// @Param state query string true "State"
// @Success 308 {string} json "redirectURL"
// @Router /integrations/salesforce/auth/callback [get]
func SalesforceCallbackHandler(c *gin.Context) {
	var oauthState IntSalesforce.OAuthState
	accessCode := c.Query("code")
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || *oauthState.AgentUUID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectID, "agent_uuid": oauthState.AgentUUID})
	salesforceTokenParams := IntSalesforce.AuthParams{
		GrantType:    "authorization_code",
		AccessCode:   accessCode,
		ClientID:     C.GetSalesforceAppId(),
		ClientSecret: C.GetSalesforceAppSecret(),
		RedirectURL:  GetSalesforceRedirectURL(),
	}

	userCredentials, err := IntSalesforce.GetSalesforceUserToken(&salesforceTokenParams)
	if err != nil {
		logCtx.WithError(err).Error("Failed to getSalesforceUserToken.")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	refreshToken, instancURL := getRequiredSalesforceCredentials(userCredentials)
	if refreshToken == "" || instancURL == "" {
		logCtx.Error("Failed to getRequiredSalesforceCredentials")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	errCode := store.GetStore().UpdateAgentIntSalesforce(*oauthState.AgentUUID,
		refreshToken,
		instancURL,
	)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce properties for agent.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating saleforce properties for agent"})
		return
	}

	_, errCode = store.GetStore().UpdateProjectSettings(oauthState.ProjectID,
		&model.ProjectSetting{IntSalesforceEnabledAgentUUID: oauthState.AgentUUID},
	)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update project settings salesforce enable agent uuid.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating salesforce enabled agent uuid project settings"})
		return
	}

	redirectURL := C.GetProtocol() + C.GetAPPDomain() + IntSalesforce.AppSettingsURL
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
}

func getRequiredSalesforceCredentials(credentials map[string]interface{}) (string, string) {
	if refreshToken, rValid := credentials[IntSalesforce.RefreshToken].(string); rValid { //could lead to error if refresh token not set on auth scope
		if instancURL, iValid := credentials[IntSalesforce.InstanceURL].(string); iValid {
			if refreshToken != "" && instancURL != "" {
				return refreshToken, instancURL
			}

		}
	}
	return "", ""
}

// SalesforceAuthRedirectHandler redirects to Salesforce oauth page
// SalesforceAuthRedirectHandler godoc
// @Summary For Salesforce authentication. Redirects to Salesforce oauth page.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param request body SalesforceRedirectRequestPayload true "Request payload"
// @Success 307 {string} json "{"redirectURL": redirectURL}"
// @Router /integrations/salesforce/auth [post]
func SalesforceAuthRedirectHandler(c *gin.Context) {
	r := c.Request

	var requestPayload SalesforceRedirectRequestPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Salesforce get redirect url payload decode failure.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload."})
		return
	}

	projectID, err := strconv.ParseInt(requestPayload.ProjectID, 10, 64)
	if err != nil || projectID == 0 {
		log.WithError(err).Error("Failed to get project_id on get SalesforceAuthRedirectHandler.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid project id."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if currentAgentUUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id."})
		return
	}

	oAuthState := &IntSalesforce.OAuthState{
		ProjectID: projectID,
		AgentUUID: &currentAgentUUID,
	}

	enOAuthState, err := json.Marshal(oAuthState)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	redirectURL := IntSalesforce.GetSalesforceAuthorizationURL(C.GetSalesforceAppId(), GetSalesforceRedirectURL(), "code", url.QueryEscape(string(enOAuthState)))
	c.JSON(http.StatusTemporaryRedirect, gin.H{"redirectURL": redirectURL})
}

func IntDeleteHandler(c *gin.Context) {

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID, err := strconv.ParseInt(c.Params.ByName("project_id"), 10, 64)
	if projectID == 0 || err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete Integration failed. Invalid project."})
		return
	}
	loggedInAgentPAM, errCode := store.GetStore().GetProjectAgentMapping(projectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch loggedInAgentPAM"})
		return
	}

	if loggedInAgentPAM.Role != model.ADMIN {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "operation denied for non-admins"})
		return
	}

	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}
	channelName := c.Params.ByName("channel_name")
	if channelName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid channel name."})
		return
	}

	errCode, err = store.GetStore().DeleteChannelIntegration(projectID, channelName)
	if err != nil || errCode != http.StatusOK {
		c.AbortWithStatusJSON(errCode, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully deleted the integration"})
}

type HubspotAuthRequestPayload struct {
	ProjectID string `json:"project_id"`
}

// HubspotOAuthState represent the state parameter for hubspot oAuth flow
type HubspotOAuthState struct {
	ProjectID int64  `json:"project_id"`
	AgentUUID string `json:"agent_uuid"`
}

// HubspotCallbackRoute holds oauth redirect route
const HubspotCallbackRoute = "/hubspot/auth/callback"

// GetHubspotRedirectURL return the redirect URL based on environment
func GetHubspotRedirectURL() string {
	return C.GetProtocol() + C.GetAPIDomain() + ROUTE_INTEGRATIONS_ROOT + HubspotCallbackRoute
}

// HubspotAuthRedirectHandler redirects to hubspot oauth page
// HubspotAuthRedirectHandler godoc
// @Summary For hubspot authentication. Redirects to hubspot oauth page.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param request body HubspotAuthRequestPayload true "Request payload"
// @Success 307 {string} json "{"redirect_url": redirectURL}"
// @Router /integrations/hubspot/auth [post]
func HubspotAuthRedirectHandler(c *gin.Context) {
	r := c.Request
	var requestPayload HubspotAuthRequestPayload

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		log.WithError(err).Error("Failed to decode request payload for hubspot auth.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request payload."})
		return
	}

	if requestPayload.ProjectID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	projectID, err := strconv.ParseInt(requestPayload.ProjectID, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as int64 for hubspot auth redirect url .")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "agent_uuid": currentAgentUUID})

	OAuthState := &HubspotOAuthState{
		ProjectID: projectID,
		AgentUUID: currentAgentUUID,
	}

	enOAuthState, err := json.Marshal(OAuthState)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal state for hubspot auth.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate auth state."})
		return
	}

	redirectURL := IntHubspot.GetHubspotAuthorizationURL(C.GetHubspotAppID(), GetHubspotRedirectURL(), string(enOAuthState))
	c.JSON(http.StatusTemporaryRedirect, gin.H{"redirect_url": redirectURL})
}

// HubspotCallbackHandler handles the callback url from hubspot auth redirect url and requests access token
// HubspotCallbackHandler godoc
// @Summary Handles the callback url from hubspot auth redirect url and requests access token.
// @Tags Integrations
// @Accept  json
// @Produce json
// @Param code query string true "Code"
// @Param state query string true "State"
// @Success 308 {string} json "redirectURL"
// @Router /integrations/hubspot/auth/callback [get]
func HubspotCallbackHandler(c *gin.Context) {
	var oauthState HubspotOAuthState
	accessCode := c.Query("code")
	state := c.Query("state")
	err := json.Unmarshal([]byte(state), &oauthState)
	if err != nil || oauthState.ProjectID == 0 || oauthState.AgentUUID == "" {
		log.WithFields(log.Fields{"state": state}).Error("Invalid auth state on hubspot auth callback handler.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid auth state."})
		return
	}

	logCtx := log.WithFields(log.Fields{"project_id": oauthState.ProjectID, "agent_uuid": oauthState.AgentUUID})

	_, errCode := store.GetStore().GetProjectAgentMapping(oauthState.ProjectID, oauthState.AgentUUID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get project agent mapping for requested hubspot auth user.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project agent."})
		return
	}

	credentials, err := IntHubspot.GetHubspotOAuthUserCredentials(C.GetHubspotAppID(), C.GetHubspotAppSecret(), GetHubspotRedirectURL(), accessCode)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get refresh token from hubspot auth token.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get hubspot user credentials."})
		return
	}

	if credentials.RefreshToken == "" {
		logCtx.WithError(err).Error("Empty refresh token on hubspot auth token response.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "invalid refresh token."})
		return
	}

	intHubspot := true
	_, errCode = store.GetStore().UpdateProjectSettings(oauthState.ProjectID,
		&model.ProjectSetting{IntHubspotRefreshToken: credentials.RefreshToken, IntHubspot: &intHubspot},
	)

	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update project settings for hubspot auth refresh token.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed updating hubspot auth project settings."})
		return
	}

	redirectURL := C.GetProtocol() + C.GetAPPDomain() + "/settings/integration"
	c.Redirect(http.StatusPermanentRedirect, redirectURL)
}
