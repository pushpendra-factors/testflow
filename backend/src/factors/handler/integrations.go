package handler

import (
	"encoding/json"
	I "factors/integration"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func IntSegmentHandler(c *gin.Context) {
	r := c.Request

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	if !M.IsPSettingsIntSegmentEnabled(projectId) {
		log.WithField("project_id", projectId).Error("Segment webhook failure. Integration not enabled.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	var event I.SegmentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "error": err}).Error("Segment JSON decode failed")
	}
	log.WithFields(log.Fields{"event": event}).Debug("Segment webhook request")

	logCtx := log.WithFields(log.Fields{
		"project_id":   projectId,
		"type":         event.Type,
		"anonymous_id": event.AnonymousID,
		"user_id":      event.UserId,
	})

	user, errCode := M.GetSegmentUser(projectId, event.AnonymousID, event.UserId)
	if errCode != http.StatusOK {
		c.AbortWithStatus(errCode)
		return
	}

	var unixTimestamp int64
	if parsedTimestamp, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		logCtx.WithField("timestamp", event.Timestamp).Error("Failed parsing segment event timestamp.")
	} else {
		unixTimestamp = parsedTimestamp.Unix()
	}

	var response gin.H
	var status int

	switch event.Type {
	case "track":
		userProperties := U.PropertiesMap{}
		I.FillSegmentGenericUserProperties(&userProperties, &event)
		I.FillSegmentWebUserProperties(&userProperties, &event)
		I.FillSegmentMobileUserProperties(&userProperties, &event)

		var eventProperties U.PropertiesMap
		if event.Properties != nil {
			// Initialized with already existing event props.
			eventProperties = event.Properties
		} else {
			eventProperties = make(U.PropertiesMap, 0)
		}
		I.FillSegmentGenericEventProperties(&eventProperties, &event)
		I.FillSegmentWebEventProperties(&eventProperties, &event)
		I.FillSegmentMobileEventProperties(&eventProperties, &event)

		request := &sdkTrackPayload{
			Name:            event.TrackName,
			UserId:          user.ID,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		if status, response = sdkTrack(projectId, request, event.Context.IP); status != http.StatusOK {
			logCtx.WithFields(log.Fields{"track_payload": request, "error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "page":
		userProperties := U.PropertiesMap{}
		I.FillSegmentGenericUserProperties(&userProperties, &event)
		I.FillSegmentWebUserProperties(&userProperties, &event)

		eventProperties := U.PropertiesMap{}
		I.FillSegmentGenericEventProperties(&eventProperties, &event)
		I.FillSegmentWebEventProperties(&eventProperties, &event)

		request := &sdkTrackPayload{
			Name:            event.Context.Page.RawURL,
			UserId:          user.ID,
			Auto:            true,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		if status, response = sdkTrack(projectId, request, event.Context.IP); status != http.StatusOK {
			logCtx.WithFields(log.Fields{"track_payload": request, "error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "screen":
		userProperties := U.PropertiesMap{}
		I.FillSegmentGenericUserProperties(&userProperties, &event)
		I.FillSegmentMobileUserProperties(&userProperties, &event)

		// Initialized with already existing event props.
		eventProperties := event.Properties
		I.FillSegmentGenericEventProperties(&eventProperties, &event)
		I.FillSegmentMobileEventProperties(&eventProperties, &event)

		request := &sdkTrackPayload{
			Name:            event.ScreenName,
			UserId:          user.ID,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		if status, response = sdkTrack(projectId, request, event.Context.IP); status != http.StatusOK {
			logCtx.WithFields(log.Fields{"track_payload": request, "error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "identify":
		// Identification happens on every call before type switch.
		// Updates the user properties with the traits, here.
		response = gin.H{"user_id": user.ID}

		_, status := M.UpdateUserProperties(projectId, user.ID, &event.Traits)
		if status != http.StatusAccepted && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"user_properties": event.Traits, "error_code": status}).Error("Segment event failure. Updating user_properties failed.")
			response["error"] = "Segment identification failed."
		}

	default:
		response = gin.H{"error": "Segment event failure. Unknown event type.", "type": event.Type}
		logCtx.Error("Unknown segment event type.")
	}

	// Always return HTTP STATUS_OK with original response.
	c.JSON(http.StatusOK, response)
}
