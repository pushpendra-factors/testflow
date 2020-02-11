package handler

import (
	"encoding/json"
	C "factors/config"
	IntSegment "factors/integration/segment"
	IntShopify "factors/integration/shopify"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func IntSegmentHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	// Skipping configured projects.
	for _, skipProjectId := range C.GetSkipTrackProjectIds() {
		if skipProjectId == projectId {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"error": "Track skipped."})
			return
		}
	}

	if !M.IsPSettingsIntSegmentEnabled(projectId) {
		logCtx.WithField("project_id", projectId).Error("Segment webhook failure. Integration not enabled.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	var event IntSegment.SegmentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error("Segment JSON decode failed")
	}
	logCtx.WithFields(log.Fields{"event": event}).Debug("Segment webhook request")

	logCtx = logCtx.WithFields(log.Fields{
		"project_id":   projectId,
		"type":         event.Type,
		"anonymous_id": event.AnonymousID,
		"user_id":      event.UserId,
	})

	user, errCode := M.GetSegmentUser(projectId, event.AnonymousID, event.UserId)
	if errCode != http.StatusOK && errCode != http.StatusCreated {
		c.AbortWithStatus(errCode)
		return
	}
	isNewUser := errCode == http.StatusCreated

	response := &SDKTrackResponse{}

	parsedTimestamp, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		logCtx.WithFields(log.Fields{"timestamp": event.Timestamp,
			log.ErrorKey: err}).Error("Failed parsing segment event timestamp.")
		response.Error = "invalid event timestamp"
		response.Type = event.Type
		c.AbortWithStatusJSON(http.StatusOK, response) // For avoiding, segment retries.
		return
	}
	unixTimestamp := parsedTimestamp.Unix()

	var status int
	switch event.Type {
	case "track":
		userProperties := U.PropertiesMap{}
		IntSegment.FillSegmentGenericUserProperties(&userProperties, &event)
		IntSegment.FillSegmentWebUserProperties(&userProperties, &event)
		IntSegment.FillSegmentMobileUserProperties(&userProperties, &event)

		var eventProperties U.PropertiesMap
		if event.Properties != nil {
			// Initialized with already existing event props.
			eventProperties = event.Properties
		} else {
			eventProperties = make(U.PropertiesMap, 0)
		}
		IntSegment.FillSegmentGenericEventProperties(&eventProperties, &event)
		IntSegment.FillSegmentWebEventProperties(&eventProperties, &event)

		request := &SDKTrackPayload{
			Name:            event.TrackName,
			CustomerEventId: event.MessageID,
			IsNewUser:       isNewUser,
			UserId:          user.ID,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		status, response = SDKTrack(projectId, request, event.Context.IP, event.Context.UserAgent, false)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}
	case "page":
		pageURL := IntSegment.GetURLFromPageEvent(&event)
		parsedPageURL, err := U.ParseURLStable(pageURL)
		if err != nil {
			logCtx.WithFields(log.Fields{log.ErrorKey: err, "page_url": pageURL}).Error(
				"Failed parsing URL from segment.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		userProperties := U.PropertiesMap{}
		IntSegment.FillSegmentGenericUserProperties(&userProperties, &event)
		IntSegment.FillSegmentWebUserProperties(&userProperties, &event)

		eventProperties := U.PropertiesMap{}
		U.FillPropertiesFromURL(&eventProperties, parsedPageURL)
		IntSegment.FillSegmentGenericEventProperties(&eventProperties, &event)
		IntSegment.FillSegmentWebEventProperties(&eventProperties, &event)

		name := U.GetURLHostAndPath(parsedPageURL)
		request := &SDKTrackPayload{
			Name:            name,
			UserId:          user.ID,
			IsNewUser:       isNewUser,
			Auto:            true,
			CustomerEventId: event.MessageID,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		status, response = SDKTrack(projectId, request, event.Context.IP, event.Context.UserAgent, false)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "screen":
		userProperties := U.PropertiesMap{}
		IntSegment.FillSegmentGenericUserProperties(&userProperties, &event)
		IntSegment.FillSegmentMobileUserProperties(&userProperties, &event)

		var eventProperties U.PropertiesMap
		if event.Properties != nil {
			// Initialized with already existing event props.
			eventProperties = event.Properties
		} else {
			eventProperties = make(U.PropertiesMap, 0)
		}
		IntSegment.FillSegmentGenericEventProperties(&eventProperties, &event)

		request := &SDKTrackPayload{
			Name:            event.ScreenName,
			UserId:          user.ID,
			IsNewUser:       isNewUser,
			Auto:            false,
			CustomerEventId: event.MessageID,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		status, response = SDKTrack(projectId, request, event.Context.IP, event.Context.UserAgent, false)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "identify":
		// Identification happens on every call before type switch.
		// Updates the user properties with the traits, here.
		response.UserId = user.ID

		_, status := M.UpdateUserProperties(projectId, user.ID, &event.Traits)
		if status != http.StatusAccepted && status != http.StatusNotModified {
			//logCtx.WithFields(log.Fields{"user_properties": event.Traits, "error_code": status}).Error("Segment event failure. Updating user_properties failed.")
			response.Error = "Segment identification failed."
		}

	default:
		response.Error = fmt.Sprintf("Segment event failure. Unknown event type: %s.", event.Type)
		response.Type = event.Type
		logCtx.Error("Unknown segment event type.")
	}

	// Always return HTTP STATUS_OK with original response.
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

	switch shopifyTopic {
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
			projectId, IntShopify.ACTION_SHOPIFY_CHECKOUT_CREATED, &checkoutCreated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Checkout Create JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "checkout create failed."})
			return
		}
		request := &SDKTrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
		}
		status, response := SDKTrack(projectId, request, "", "", false)
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
			projectId, IntShopify.ACTION_SHOPIFY_CHECKOUT_UPDATED, &checkoutUpdated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Checkout Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "checkout update failed."})
			return
		}
		request := &SDKTrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
		}
		status, response := SDKTrack(projectId, request, "", "", false)
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
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_CREATED, &orderCreated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Create JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order create failed."})
			return
		}
		request := &SDKTrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
		}
		status, response := SDKTrack(projectId, request, "", "", false)
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
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_UPDATED, &orderUpdated)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order update failed."})
			return
		}
		request := &SDKTrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
		}
		status, response := SDKTrack(projectId, request, "", "", false)
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
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_PAID, &orderPaid)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order update failed."})
			return
		}
		request := &SDKTrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
		}
		status, response := SDKTrack(projectId, request, "", "", false)
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
			projectId, IntShopify.ACTION_SHOPIFY_ORDER_CANCELLED, &orderCancelled)
		if err != nil {
			logCtx.WithFields(log.Fields{"project_id": projectId, log.ErrorKey: err}).Error(
				"Shopify Order Update JSON track details failed")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "order cancelled failed."})
			return
		}
		request := &SDKTrackPayload{
			Name:            eventName,
			IsNewUser:       isNewUser,
			UserId:          userId,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       timestamp,
		}
		status, response := SDKTrack(projectId, request, "", "", false)
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
