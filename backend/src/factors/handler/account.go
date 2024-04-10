package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
)

// CreateAccountHandler godoc
// @Summary Create a Groupuser
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body SDK.AccountPayload
// @Success 200
// @Router /sdk/account/create [post]
// @Security ApiKeyAuth
func CreateAccountHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Error": "Invalid token on sdk payload.",
		})
		return
	}

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid request. Request body unavailable.",
		})
		return
	}

	var accountPayload SDK.AccountPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&accountPayload); U.IsJsonError(err) {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Error": "Tracking failed. Json Decoding failed.",
		})
		return
	}

	if accountPayload.Domain == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid request. Domain Name unavailable",
		})
		return
	}

	_, status := store.GetStore().CreateOrGetDomainsGroup(projectID)
	if status != http.StatusCreated && status != http.StatusFound {
		logCtx.Error("Failed to CreateOrGetDomainsGroup ")
		c.JSON(status, gin.H{
			"Error": "Failed to CreateOrGetDomainsGroup",
		})
		return
	}

	domainName := U.GetDomainGroupDomainName(projectID, accountPayload.Domain)
	groupUserID, status := store.GetStore().CreateOrGetDomainGroupUser(projectID, model.GROUP_NAME_DOMAINS, domainName, U.TimeNowUnix(), model.GetGroupUserSourceByGroupName(model.GROUP_NAME_DOMAINS))

	if status != http.StatusCreated && status != http.StatusFound {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for  group user by group id.")
		c.JSON(status, gin.H{
			"Error": "Failed to check for  group user by group id.",
		})
		return
	} else if status == http.StatusFound {

		c.JSON(http.StatusConflict, gin.H{
			"message": "Account exists already.",
		})
		return
	}

	if accountPayload.Properties != nil {

		accountPayload.Properties[U.UP_IS_OFFLINE] = true

		accountPayloadJsonb, err := U.EncodeStructTypeToPostgresJsonb(accountPayload.Properties)
		if err != nil {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for  group user by group id.")
			c.JSON(status, gin.H{
				"Error": "Failed to check for  group user by group id.",
			})
			return
		}

		source := model.GetGroupUserSourceNameByGroupName(U.GROUP_NAME_DOMAINS)

		accountPayloadMaps, _ := U.DecodePostgresJsonb(accountPayloadJsonb)

		_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, U.GROUP_NAME_DOMAINS, domainName, groupUserID, accountPayloadMaps, U.TimeNowUnix(), U.TimeNowUnix(), source)
		if err != nil {
			logCtx.WithField("err", err).
				Error("Update user properties on track failed. DB update failed.")
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully.",
	})

}

// UpdateAccountHandler godoc
// @Summary update a Groupuser
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body SDK.AccountPayload
// @Success 200
// @Router /sdk/account/update [post]
// @Security ApiKeyAuth
func UpdateAccountHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Error": "Invalid token on sdk payload.",
		})
		return
	}

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid request. Request body unavailable.",
		})
		return
	}

	var accountPayload SDK.AccountPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&accountPayload); U.IsJsonError(err) {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Error": "Tracking failed. Json Decoding failed.",
		})
		return
	}

	if accountPayload.Domain == "" || accountPayload.Properties == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid request. Domain Name unavailable",
		})
		return
	}

	domainName := U.GetDomainGroupDomainName(projectID, accountPayload.Domain)
	groupUser, status := store.GetStore().GetGroupUserByGroupID(projectID, model.GROUP_NAME_DOMAINS, domainName)
	if status != http.StatusFound {
		logCtx.Error("Failed to Get DomainsGroupUser ")
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Account not found.",
		})
	}

	if accountPayload.Properties != nil {

		accountPayload.Properties[U.UP_IS_OFFLINE] = true

		accountPayloadJsonb, err := U.EncodeStructTypeToPostgresJsonb(accountPayload.Properties)
		if err != nil {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for  group user by group id.")

			c.AbortWithError(http.StatusBadRequest, errors.New("Failed to check for  group user by group id."))

			return
		}

		source := model.GetGroupUserSourceNameByGroupName(U.GROUP_NAME_DOMAINS)

		accountPayloadMaps, _ := U.DecodePostgresJsonb(accountPayloadJsonb)

		_, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, U.GROUP_NAME_DOMAINS, domainName, groupUser.ID, accountPayloadMaps, U.TimeNowUnix(), U.TimeNowUnix(), source)
		if err != nil {
			logCtx.WithField("err", err).
				Error("Update user properties on track failed. DB update failed.")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated account successfully.",
	})

}

// TrackAccountEventHandler godoc
// @Summary track event for account
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body SDK.AccountPayload
// @Success 200
// @Router /sdk/account/event/track [post]
// @Security ApiKeyAuth
func TrackAccountEventHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Error": "Invalid token on sdk payload.",
		})
		return
	}

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")

		c.AbortWithError(http.StatusBadRequest, errors.New("Invalid request. Request body unavailable."))

		return
	}

	var accountTrackPayload SDK.AccountTrackPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&accountTrackPayload); U.IsJsonError(err) {
		logCtx.WithError(err).Error("Tracking failed. Json Decoding failed.")

		c.AbortWithError(http.StatusBadRequest, errors.New("Tracking failed. Json Decoding failed."))
		return
	}

	if accountTrackPayload.Domain == "" || accountTrackPayload.Event == nil {

		c.AbortWithError(http.StatusBadRequest, errors.New("Invalid request. Domain Name unavailable"))
		return
	}

	domainName := U.GetDomainGroupDomainName(projectID, accountTrackPayload.Domain)
	groupUser, status := store.GetStore().GetGroupUserByGroupID(projectID, model.GROUP_NAME_DOMAINS, domainName)
	if status != http.StatusFound {
		logCtx.Error("Failed to Get DomainsGroupUser ")
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Account not found.",
		})
	}

	timestamp, _ := U.GetPropertyValueAsInt64(accountTrackPayload.Event["timestamp"])

	accountPayloadJsonb, err := U.EncodeStructTypeToPostgresJsonb(accountTrackPayload.Event["properties"])
	if err != nil {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to check for  group user by group id.")
		c.AbortWithError(status, errors.New("Failed to check for  group user by group id."))
		return
	}

	eventDetails, _ := store.GetStore().GetEventNameIDFromEventName(U.GetPropertyValueAsString(accountTrackPayload.Event["name"]), projectID)

	createdEvent, errCode := store.GetStore().CreateEvent(&model.Event{
		ProjectId:   projectID,
		EventNameId: eventDetails.ID,
		UserId:      groupUser.ID,
		Properties:  *accountPayloadJsonb,
		Timestamp:   timestamp,
	})

	if errCode == http.StatusNotAcceptable {
		c.AbortWithError(errCode, errors.New("Tracking failed. Event creation failed. Invalid payload."))
		return
	} else if errCode != http.StatusCreated {
		c.AbortWithError(errCode, errors.New("Tracking failed. Event creation failed."))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event tracked successfully.",
		"eventID": createdEvent.ID,
	})

}
