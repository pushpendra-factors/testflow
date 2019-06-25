package handler

import (
	"encoding/json"
	I "factors/integration"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
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

	if !M.IsPSettingsIntSegmentEnabled(projectId) {
		logCtx.WithField("project_id", projectId).Error("Segment webhook failure. Integration not enabled.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Creating event_name failed. Invalid project."})
		return
	}

	var event I.SegmentEvent
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
	if errCode != http.StatusOK {
		c.AbortWithStatus(errCode)
		return
	}

	var unixTimestamp int64
	if parsedTimestamp, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		logCtx.WithFields(log.Fields{
			"timestamp":  event.Timestamp,
			log.ErrorKey: err,
		}).Error("Failed parsing segment event timestamp.")
	} else {
		unixTimestamp = parsedTimestamp.Unix()
	}

	response := &SDKTrackResponse{}
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

		request := &sdkTrackPayload{
			Name:            event.TrackName,
			CustomerEventId: event.MessageID,
			UserId:          user.ID,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		if status, response = sdkTrack(projectId, request, event.Context.IP); status != http.StatusOK {
			//logCtx.WithFields(log.Fields{"track_payload": request, "error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "page":
		userProperties := U.PropertiesMap{}
		I.FillSegmentGenericUserProperties(&userProperties, &event)
		I.FillSegmentWebUserProperties(&userProperties, &event)

		eventProperties := U.PropertiesMap{}
		I.FillSegmentGenericEventProperties(&eventProperties, &event)
		I.FillSegmentWebEventProperties(&eventProperties, &event)

		url := I.GetURLFromPageEvent(&event)
		name, err := U.GetURLHostAndPath(url)
		if err != nil {
			logCtx.WithFields(log.Fields{log.ErrorKey: err, "name": name}).Error("Falied parsing URL from segment.")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		request := &sdkTrackPayload{
			Name:            name,
			UserId:          user.ID,
			Auto:            true,
			CustomerEventId: event.MessageID,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		if status, response = sdkTrack(projectId, request, event.Context.IP); status != http.StatusOK {
			//logCtx.WithFields(log.Fields{"track_payload": request, "error_code": status}).Error("Segment event failure. sdk_track call failed.")
		}

	case "screen":
		userProperties := U.PropertiesMap{}
		I.FillSegmentGenericUserProperties(&userProperties, &event)
		I.FillSegmentMobileUserProperties(&userProperties, &event)

		// Initialized with already existing event props.
		eventProperties := event.Properties
		I.FillSegmentGenericEventProperties(&eventProperties, &event)

		request := &sdkTrackPayload{
			Name:            event.ScreenName,
			UserId:          user.ID,
			Auto:            false,
			CustomerEventId: event.MessageID,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       unixTimestamp,
		}
		if status, response = sdkTrack(projectId, request, event.Context.IP); status != http.StatusOK {
			//logCtx.WithFields(log.Fields{"track_payload": request, "error_code": status}).Error("Segment event failure. sdk_track call failed.")
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
