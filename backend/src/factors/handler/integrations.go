package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	C "factors/config"
	IntSegment "factors/integration/segment"
	IntShopify "factors/integration/shopify"
	mid "factors/middleware"
	M "factors/model"
	SDK "factors/sdk"
	U "factors/util"
)

func IntSegmentHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	var event IntSegment.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		logCtx.WithError(err).Error("Segment JSON decode failed")
	}
	logCtx.WithFields(log.Fields{"event": event}).Debug("Segment webhook request")

	token := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_PRIVATE_TOKEN)

	status, response := IntSegment.ReceiveEventWithQueue(token, &event,
		C.GetSegmentRequestQueueAllowedTokens())

	// send error on StatusInternalServerError (db unavailability and redis error),
	// for segment to retry.
	if status == http.StatusInternalServerError {
		c.AbortWithStatus(status)
	}

	// Always send StatusOK for failure on direct processing.
	c.JSON(http.StatusOK, response)
}

// Verifies agent access to projects using middlewares.
func IsAgentAuthorizedToAccessProject(projectId uint64, c *gin.Context) bool {
	agentAuthorizedProjectIds := U.GetScopeByKey(c, mid.SCOPE_AUTHORIZED_PROJECTS)
	for _, authorizedProjectId := range agentAuthorizedProjectIds.([]uint64) {
		if projectId == authorizedProjectId {
			return true
		}
	}

	return false
}

type AdwordsAddRefreshTokenPayload struct {
	// project_id conv from string to uint64 explicitly.
	ProjectId    string `json:"project_id"`
	RefreshToken string `json:"refresh_token"`
}

type AdwordsRequestPayload struct {
	ProjectId string `json:"project_id"`
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

	if requestPayload.ProjectId == "" || requestPayload.RefreshToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. empty mandatory fields."})
		return
	}

	projectId, err := strconv.ParseUint(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid json payload. empty mandatory fields."})
		return
	}

	if !IsAgentAuthorizedToAccessProject(projectId, c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authorized to access project"})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	errCode := M.UpdateAgentIntAdwordsRefreshToken(currentAgentUUID, requestPayload.RefreshToken)
	if errCode != http.StatusAccepted {
		log.WithField("agent_uuid", currentAgentUUID).Error("Failed to update adwords refresh token for agent.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed updating adwords refresh token for agent"})
		return
	}

	_, errCode = M.UpdateProjectSettings(projectId, &M.ProjectSetting{IntAdwordsEnabledAgentUUID: &currentAgentUUID})
	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectId).Error("Failed to update project settings adwords enable agent uuid.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed updating adwords enabled agent uuid project settings"})
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

	projectId, err := strconv.ParseUint(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	refreshToken, errCode := M.GetIntAdwordsRefreshTokenForProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get adwords refresh token for project."})
		return
	}

	c.JSON(http.StatusFound, gin.H{"refresh_token": refreshToken})
}

// IntEnableAdwordsHandler - Checks for refresh_token for the
// agent if exists: then add the agent_uuid as adwords_enabled_agent_uuid
// on project settings. if not exists: return 304.
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

	projectId, err := strconv.ParseUint(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}

	currentAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	agent, errCode := M.GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	if agent.IntAdwordsRefreshToken == "" {
		c.JSON(http.StatusNotModified, gin.H{})
		return
	}

	addEnableAgentUUIDSetting := M.ProjectSetting{IntAdwordsEnabledAgentUUID: &currentAgentUUID}
	_, errCode = M.UpdateProjectSettings(projectId, &addEnableAgentUUIDSetting)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to enable adwords"})
		return
	}

	c.JSON(http.StatusOK, addEnableAgentUUIDSetting)
}

func IntShopifyHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "Shopify webhook failed. Invalid project."})
		return
	}

	if !M.IsPSettingsIntShopifyEnabled(projectId) {
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
		}
		status, response := SDK.Track(projectId, request, false)
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
		}
		status, response := SDK.Track(projectId, request, false)
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
		}
		status, response := SDK.Track(projectId, request, false)
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
		}
		status, response := SDK.Track(projectId, request, false)
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
		}
		status, response := SDK.Track(projectId, request, false)
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
		}
		status, response := SDK.Track(projectId, request, false)
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
		}
		status, response := SDK.Track(projectId, request, false)
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
	AdAccount   string `json:"int_facebook_ad_account"`
}

type FacebookLongLivedTokenResponse struct {
	AccessToken string `json:"access_token"`
}

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
	_, errCode := M.GetAgentByUUID(currentAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid agent."})
		return
	}

	resp, err := http.Get("https://graph.facebook.com/v6.0/oauth/access_token?" +
		"grant_type=fb_exchange_token&client_id=" + C.GetFacebookAppId() + "&client_secret=" + C.GetFacebookAppSecret() +
		"&fb_exchange_token=" + requestPayload.AccessToken)
	if err != nil {
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

	projectId, err := strconv.ParseUint(requestPayload.ProjectId, 10, 64)
	if err != nil {
		log.WithError(err).Error("Failed to convert project_id as uint64.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project."})
		return
	}
	_, errCode = M.UpdateProjectSettings(projectId, &M.ProjectSetting{
		IntFacebookEmail: requestPayload.Email, IntFacebookAccessToken: newBody.AccessToken,
		IntFacebookAgentUUID: &currentAgentUUID, IntFacebookUserID: requestPayload.UserID,
		IntFacebookAdAccount: requestPayload.AdAccount})
	if errCode != http.StatusAccepted {
		log.WithField("project_id", projectId).Error("Failed to update project settings with facebook email and access token")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed updating facebook email and access token in project settings"})
		return
	}

	c.JSON(errCode, gin.H{})
}

func IntShopifySDKHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "Shopify webhook failed. Invalid project."})
		return
	}

	if !M.IsPSettingsIntShopifyEnabled(projectId) {
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
	errCode := M.SetCacheShopifyCartTokenToUserId(
		projectId, cartTokenPayload.CartToken, cartTokenPayload.UserId)
	if errCode != http.StatusOK && errCode != http.StatusCreated {
		c.AbortWithStatus(errCode)
		return
	}
	c.JSON(http.StatusOK, nil)
}
