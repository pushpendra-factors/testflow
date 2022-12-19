package segment

import (
	"encoding/json"
	"factors/model/model"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	Int "factors/integration"
	"factors/metrics"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"factors/vendor_custom/machinery/v1/tasks"
)

// Note:
// (userId) = Factors(customerUserId), (AnonymousId) = Factors(userId).
// Property mappings are defined on corresponding fill*Properities method.

type Device struct {
	ID                string `json:"id"`
	Manufacturer      string `json:"manufacturer"`
	Model             string `json:"model"`
	Type              string `json:"type"`
	Name              string `json:"name"`
	AdvertisingID     string `json:"advertisingId"`
	AdTrackingEnabled bool   `json:"adTrackingEnabled"`
	Token             string `json:"token"`
}

type Page struct {
	Referrer string `json:"referrer"`
	RawURL   string `json:"url"`
	Title    string `json:"title"`
	// Path     string `json:"path"`   // redundant. part of rawURL already.
	// Search   string `json:"search"` // redundant. part of rawURL already.
}

type App struct {
	Name      string       `json:"name"`
	Version   *interface{} `json:"version"`
	Build     *interface{} `json:"build"`
	Namespace string       `json:"namespace"`
}

type Location struct {
	City    string      `json:"city"`
	Country string      `json:"country"`
	Region  string      `json:"region"`
	Lat     interface{} `json:"latitude"`
	Long    interface{} `json:"longitude"`
}

type OS struct {
	Name    string       `json:"name"`
	Version *interface{} `json:"version"`
}

type Screen struct {
	// Changed height to interface{} as part of hot-fix, as
	// it is sent as string sometimes.
	// https://github.com/Slashbit-Technologies/factors/issues/1600
	Width   interface{} `json:"width"`
	Height  interface{} `json:"height"`
	Density interface{} `json:"density"`
}

type Network struct {
	Bluetooth bool   `json:"bluetooth"`
	Carrier   string `json:"carrier"`
	Cellular  bool   `json:"cellular"`
	Wifi      bool   `json:"wifi"`
}

type Library struct {
	Name string `json:"name"`
}

type Campaign struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Medium  string `json:"medium"`
	Term    string `json:"term"`
	Content string `json:"content"`
}

type Context struct {
	Campaign  interface{} `json:"campaign"`
	IP        string      `json:"ip"`
	Location  Location    `json:"location"`
	Page      Page        `json:"page"`
	UserAgent string      `json:"userAgent"`
	OS        OS          `json:"os"`
	Screen    Screen      `json:"screen"`
	Locale    string      `json:"locale"`
	Device    Device      `json:"device"`
	Network   Network     `json:"network"`
	App       App         `json:"app"`
	Library   Library     `json:"library"`
	Timezone  string      `json:"timezone"`
}

type Event struct {
	TrackName   string          `json:"event"`
	ScreenName  string          `json:"name"`
	UserId      string          `json:"userId"`
	AnonymousID string          `json:"anonymousId"`
	MessageID   *string         `json:"messageId"`
	GroupID     *string         `json:"groupId"`
	Channel     string          `json:"channel"`
	Context     Context         `json:"context"`
	Timestamp   string          `json:"timestamp"`
	Type        string          `json:"type"`
	Properties  U.PropertiesMap `json:"properties"`
	Traits      postgres.Jsonb  `json:"traits"`
	Version     *interface{}    `json:"version"`
}

type EventResponse struct {
	EventId string `json:"event_id,omitempty"`
	UserId  string `json:"user_id,omitempty"`
	Type    string `json:"type,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

const PagePropertyURL = "url"

func GetURLFromPageEvent(event *Event) string {
	if event.Context.Page.RawURL != "" {
		return event.Context.Page.RawURL
	}

	url, exists := event.Properties[PagePropertyURL]
	if exists && url != nil {
		return url.(string)
	}

	return ""
}

func getPropertyValuesAsFloat64WithDefault(key string, value interface{}) float64 {
	// TODO(Dinesh): Add check for empty interface{} and try not to convert if it is empty.
	valueFloat64, err := U.GetPropertyValueAsFloat64(value)
	if err != nil {
		log.WithField("key", key).WithError(err).Error("Failed to property value to float.")
		return 0
	}

	return valueFloat64
}

func fillGenericEventProperties(properties *U.PropertiesMap, event *Event) {
	if event.Version != nil {
		(*properties)[U.EP_SEGMENT_EVENT_VERSION] = *event.Version
	}
	if event.Context.Library.Name != "" {
		(*properties)[U.EP_SEGMENT_SOURCE_LIBRARY] = event.Context.Library.Name
	}
	if event.Channel != "" {
		(*properties)[U.EP_SEGMENT_SOURCE_CHANNEL] = event.Channel
	}

	latitude := getPropertyValuesAsFloat64WithDefault("latitude", event.Context.Location.Lat)
	if latitude != 0 {
		(*properties)[U.EP_LOCATION_LATITUDE] = latitude
	}

	longitude := getPropertyValuesAsFloat64WithDefault("longitude", event.Context.Location.Long)
	if longitude != 0 {
		(*properties)[U.EP_LOCATION_LONGITUDE] = longitude
	}
}

func fillGenericUserProperties(properties *U.PropertiesMap, event *Event) {
	(*properties)[U.UP_PLATFORM] = U.PLATFORM_WEB
	if event.Context.UserAgent != "" {
		(*properties)[U.UP_USER_AGENT] = event.Context.UserAgent
	}
	if event.Context.Location.Country != "" {
		(*properties)[U.UP_COUNTRY] = event.Context.Location.Country
	}
	if event.Context.Location.City != "" {
		(*properties)[U.UP_CITY] = event.Context.Location.City
	}
	if event.Context.Location.Region != "" {
		(*properties)[U.UP_REGION] = event.Context.Location.Region
	}

	// Added to generic event though it is mobile specific on segment.
	if event.Context.OS.Name != "" {
		(*properties)[U.UP_OS] = event.Context.OS.Name
	}
	if event.Context.OS.Version != nil {
		(*properties)[U.UP_OS_VERSION] = *event.Context.OS.Version
	}

	screenWidth := getPropertyValuesAsFloat64WithDefault("screen_width", event.Context.Screen.Width)
	if screenWidth != 0 {
		(*properties)[U.UP_SCREEN_WIDTH] = screenWidth
	}

	screenHeight := getPropertyValuesAsFloat64WithDefault("screen_height", event.Context.Screen.Height)
	if screenHeight != 0 {
		(*properties)[U.UP_SCREEN_HEIGHT] = screenHeight
	}

}

func fillMobileUserProperties(properties *U.PropertiesMap, event *Event) {
	if event.Context.App.Name != "" {
		(*properties)[U.UP_APP_NAME] = event.Context.App.Name
	}
	if event.Context.App.Namespace != "" {
		(*properties)[U.UP_APP_NAMESPACE] = event.Context.App.Namespace
	}
	if event.Context.App.Build != nil {
		(*properties)[U.UP_APP_BUILD] = *event.Context.App.Build
	}
	if event.Context.App.Version != nil {
		(*properties)[U.UP_APP_VERSION] = *event.Context.App.Version
	}
	if event.Context.Device.ID != "" {
		(*properties)[U.UP_DEVICE_ID] = event.Context.Device.ID
	}
	if event.Context.Device.Name != "" {
		(*properties)[U.UP_DEVICE_NAME] = event.Context.Device.Name
	}
	if event.Context.Device.AdvertisingID != "" {
		(*properties)[U.UP_DEVICE_ADVERTISING_ID] = event.Context.Device.AdvertisingID
	}
	if event.Context.Device.Model != "" {
		(*properties)[U.UP_DEVICE_MODEL] = event.Context.Device.Model
	}
	if event.Context.Device.Type != "" {
		(*properties)[U.UP_DEVICE_TYPE] = event.Context.Device.Type
	}
	if event.Context.Device.Manufacturer != "" {
		(*properties)[U.UP_DEVICE_MANUFACTURER] = event.Context.Device.Manufacturer
	}
	if event.Context.Network.Carrier != "" {
		(*properties)[U.UP_NETWORK_CARRIER] = event.Context.Network.Carrier
	}

	density := getPropertyValuesAsFloat64WithDefault("screen_density", event.Context.Screen.Density)
	if density != 0 {
		(*properties)[U.UP_SCREEN_DENSITY] = density
	}
	if event.Context.Timezone != "" {
		(*properties)[U.UP_TIMEZONE] = event.Context.Timezone
	}
	if event.Context.Locale != "" {
		(*properties)[U.UP_LOCALE] = event.Context.Locale
	}

	// Boolean values added without check.
	(*properties)[U.UP_DEVICE_ADTRACKING_ENABLED] = event.Context.Device.AdTrackingEnabled
	(*properties)[U.UP_NETWORK_BLUETOOTH] = event.Context.Network.Bluetooth
	(*properties)[U.UP_NETWORK_CELLULAR] = event.Context.Network.Cellular
	(*properties)[U.UP_NETWORK_WIFI] = event.Context.Network.Wifi
}

func fillWebEventProperties(properties *U.PropertiesMap, event *Event) {
	if url := GetURLFromPageEvent(event); url != "" {
		(*properties)[U.EP_PAGE_RAW_URL] = url
		pageURL, _ := U.ParseURLStable(url)
		(*properties)[U.EP_PAGE_DOMAIN] = pageURL.Host
		(*properties)[U.EP_PAGE_URL] = pageURL.Host + pageURL.Path + U.GetPathAppendableURLHash(pageURL.Fragment)
	}

	if event.Context.Page.Title != "" {
		(*properties)[U.EP_PAGE_TITLE] = event.Context.Page.Title
	}

	if event.Context.Page.Referrer != "" {
		(*properties)[U.EP_REFERRER] = event.Context.Page.Referrer
		referrerURL, _ := U.ParseURLStable(event.Context.Page.Referrer)
		(*properties)[U.EP_REFERRER_DOMAIN] = referrerURL.Host
		(*properties)[U.EP_REFERRER_URL] = referrerURL.Host + referrerURL.Path + U.GetPathAppendableURLHash(referrerURL.Fragment)
	}

	campaign := make(map[string]interface{})
	if event.Context.Campaign != nil {
		campaignType := fmt.Sprintf("%T", event.Context.Campaign)
		if campaignType == "map[string]interface {}" {
			campaign = event.Context.Campaign.(map[string]interface{})
			value, ok := campaign["name"]
			valueType := fmt.Sprintf("%T", value)
			if ok && valueType == "string" && value.(string) != "" {
				(*properties)[U.EP_CAMPAIGN] = value.(string)
			}
			value, ok = campaign["medium"]
			valueType = fmt.Sprintf("%T", value)
			if ok && valueType == "string" && value.(string) != "" {
				(*properties)[U.EP_MEDIUM] = value.(string)
			}
			value, ok = campaign["source"]
			valueType = fmt.Sprintf("%T", value)
			if ok && valueType == "string" && value.(string) != "" {
				(*properties)[U.EP_SOURCE] = value.(string)
			}
			value, ok = campaign["term"]
			valueType = fmt.Sprintf("%T", value)
			if ok && valueType == "string" && value.(string) != "" {
				(*properties)[U.EP_KEYWORD] = value.(string)
			}
			value, ok = campaign["content"]
			valueType = fmt.Sprintf("%T", value)
			if ok && valueType == "string" && value.(string) != "" {
				(*properties)[U.EP_CONTENT] = value.(string)
			}
		}
	}
}

func fillWebUserProperties(properties *U.PropertiesMap, event *Event) {}

func ReceiveEventWithQueue(token string, event *Event,
	queueAllowedTokens []string) (int, *EventResponse) {

	if token == "" || event == nil {
		return http.StatusBadRequest, &EventResponse{Error: "Invalid payload"}
	}

	logCtx := log.WithField("token", token)

	projectSetting, errCode := store.GetStore().GetProjectSettingByPrivateTokenWithCacheAndDefault(token)
	if errCode != http.StatusFound || projectSetting == nil {
		logCtx.Error("Failed to get project settings on segment/rudderstack ReceiveEventWithQueue.")
		return http.StatusBadRequest, &EventResponse{}
	}

	isSegmentEnabled := false
	if projectSetting.IntSegment != nil {
		isSegmentEnabled = *projectSetting.IntSegment
	}

	isRudderstackEnabled := false
	if projectSetting.IntRudderstack != nil {
		isRudderstackEnabled = *projectSetting.IntRudderstack
	}

	isEnabled := isSegmentEnabled || isRudderstackEnabled
	if !isEnabled {
		return http.StatusBadRequest, &EventResponse{Error: "Integration not enabled."}
	}

	bothEnabled := isSegmentEnabled && isRudderstackEnabled
	if bothEnabled {
		return http.StatusBadRequest, &EventResponse{Error: "Both integrations enabled."}
	}

	intType := ""
	if isSegmentEnabled {
		intType = Int.TypeSegment
	} else if isRudderstackEnabled {
		intType = Int.TypeRudderstack
	} else {
		return http.StatusBadRequest, &EventResponse{Error: "Invalid integration."}
	}

	if U.UseQueue(token, queueAllowedTokens) {
		err := Int.EnqueueRequest(token, intType, event)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to queue %s event request.", intType))

			// StatusInternalServerError will be forwarded to segment, to retry.
			return http.StatusInternalServerError,
				&EventResponse{Message: fmt.Sprintf("Receive %s event failed.", intType)}
		}

		return http.StatusOK, &EventResponse{
			Message: fmt.Sprintf("Successfully fully received %s event.", intType)}
	}

	return ReceiveEvent(token, event)
}

func shouldSkipSession(payload *SDK.TrackPayload) bool {
	if payload.EventProperties == nil {
		return true
	}

	return payload.EventProperties[U.EP_PAGE_RAW_URL] == nil ||
		payload.EventProperties[U.EP_PAGE_RAW_URL] == ""
}

func ReceiveEvent(token string, event *Event) (int, *EventResponse) {
	if token == "" || event == nil {
		return http.StatusBadRequest, &EventResponse{Error: "Invalid payload"}
	}

	project, errCode := store.GetStore().GetProjectByPrivateToken(token)
	if errCode == http.StatusNotFound {
		return http.StatusUnauthorized, &EventResponse{Error: "Invalid token."}
	}

	if errCode != http.StatusFound {
		return errCode, &EventResponse{Error: "Failed to get project by token"}
	}

	logCtx := log.WithFields(log.Fields{
		"project_id":   project.ID,
		"type":         event.Type,
		"anonymous_id": event.AnonymousID,
	})

	response := &EventResponse{Type: event.Type}

	parsedTimestamp, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		logCtx.WithFields(log.Fields{"timestamp": event.Timestamp,
			log.ErrorKey: err}).Error("Failed parsing segment event timestamp.")

		response.Error = "Invalid event timestamp"
		return http.StatusBadRequest, response
	}
	requestTimestamp := parsedTimestamp.Unix()

	// Reset the userId to avoid using same customer_user_id.
	if event.UserId != "" && C.IsSegmentExcludedCustomerUserID(project.ID, event.UserId) {
		event.UserId = ""
	}

	user, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, event.AnonymousID,
		event.UserId, requestTimestamp, model.UserSourceWeb)
	if errCode != http.StatusOK && errCode != http.StatusCreated {
		response.Error = "Invalid user"
		return errCode, response
	}
	isNewUser := errCode == http.StatusCreated
	userID := user.ID

	// Always try to identify when the event.UserId is available.
	if user.CustomerUserId == "" && event.UserId != "" {
		status, identifyResponse := SDK.Identify(project.ID,
			&SDK.IdentifyPayload{
				UserId:         user.ID,
				CustomerUserId: event.UserId,
				Timestamp:      requestTimestamp,
				RequestSource:  model.UserSourceWeb}, false)
		// Log and continue to track, if identification fails.
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", event.UserId).
				Error("Failed to identify segment user.")
		}

		// overwrite user_id, if new_user returned on identify.
		if identifyResponse.UserId != "" && identifyResponse.UserId != user.ID {
			userID = identifyResponse.UserId
		}
	}
	response.UserId = userID

	switch event.Type {
	case "track":

		userProperties := U.PropertiesMap{}
		fillGenericUserProperties(&userProperties, event)
		fillWebUserProperties(&userProperties, event)
		fillMobileUserProperties(&userProperties, event)

		var eventProperties U.PropertiesMap
		if event.Properties != nil {
			// Initialized with already existing event props.
			eventProperties = event.Properties
		} else {
			eventProperties = make(U.PropertiesMap, 0)
		}
		fillGenericEventProperties(&eventProperties, event)
		fillWebEventProperties(&eventProperties, event)

		pageURL := GetURLFromPageEvent(event)
		parsedPageURL, err := U.ParseURLStable(pageURL)
		if err != nil {
			// logging URL parse error but continuing for track requests
			logCtx.WithFields(log.Fields{log.ErrorKey: err, "page_url": pageURL}).Error(
				"Failed parsing URL from segment.")
		} else {
			enrichEventPropertyUsingURL(parsedPageURL, &eventProperties)
		}

		request := &SDK.TrackPayload{
			Name:            event.TrackName,
			CustomerEventId: event.MessageID,
			IsNewUser:       isNewUser,
			UserId:          userID,
			Auto:            false,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       requestTimestamp,
			ClientIP:        event.Context.IP,
			UserAgent:       event.Context.UserAgent,
			RequestSource:   model.UserSourceWeb,
		}

		status, trackResponse := SDK.Track(project.ID, request, shouldSkipSession(request), SDK.SourceSegment, "")
		if status != http.StatusOK &&
			status != http.StatusFound &&
			status != http.StatusNotModified &&
			status != http.StatusNotAcceptable {
			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status}).Error("Segment event failure. sdk_track call failed.")

			response.Error = "Reception of track event failed"
			return status, response
		}
		response.EventId = trackResponse.EventId

	case "page":

		userProperties := U.PropertiesMap{}
		fillGenericUserProperties(&userProperties, event)
		fillWebUserProperties(&userProperties, event)

		eventProperties := U.PropertiesMap{}
		fillGenericEventProperties(&eventProperties, event)
		fillWebEventProperties(&eventProperties, event)

		pageURL := GetURLFromPageEvent(event)
		parsedPageURL, err := U.ParseURLStable(pageURL)
		if err != nil {
			// logging URL parse error but continuing for page requests
			logCtx.WithFields(log.Fields{log.ErrorKey: err, "page_url": pageURL}).Error(
				"Failed parsing URL from segment.")
		} else {
			enrichEventPropertyUsingURL(parsedPageURL, &eventProperties)
		}

		name := U.GetURLHostAndPath(parsedPageURL)
		request := &SDK.TrackPayload{
			Name:            name,
			UserId:          userID,
			IsNewUser:       isNewUser,
			Auto:            true,
			CustomerEventId: event.MessageID,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       requestTimestamp,
			ClientIP:        event.Context.IP,
			UserAgent:       event.Context.UserAgent,
			RequestSource:   model.UserSourceWeb,
		}

		status, trackResponse := SDK.Track(project.ID, request, shouldSkipSession(request), SDK.SourceSegment, "")
		if status != http.StatusOK &&
			status != http.StatusFound &&
			status != http.StatusNotModified &&
			status != http.StatusNotAcceptable {

			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status}).Error("Segment event failure. sdk_track call failed.")

			response.Error = "Reception of page event failed"
			return status, response
		}
		response.EventId = trackResponse.EventId

	case "screen":
		userProperties := U.PropertiesMap{}
		fillGenericUserProperties(&userProperties, event)
		fillMobileUserProperties(&userProperties, event)

		var eventProperties U.PropertiesMap
		if event.Properties != nil {
			// Initialized with already existing event props.
			eventProperties = event.Properties
		} else {
			eventProperties = make(U.PropertiesMap, 0)
		}
		fillGenericEventProperties(&eventProperties, event)

		request := &SDK.TrackPayload{
			Name:            event.ScreenName,
			UserId:          userID,
			IsNewUser:       isNewUser,
			Auto:            false,
			CustomerEventId: event.MessageID,
			EventProperties: eventProperties,
			UserProperties:  userProperties,
			Timestamp:       requestTimestamp,
			ClientIP:        event.Context.IP,
			UserAgent:       event.Context.UserAgent,
			RequestSource:   model.UserSourceWeb,
		}

		status, trackResponse := SDK.Track(project.ID, request, shouldSkipSession(request), SDK.SourceSegment, "")
		if status != http.StatusOK &&
			status != http.StatusFound &&
			status != http.StatusNotModified &&
			status != http.StatusNotAcceptable {

			logCtx.WithFields(log.Fields{"track_payload": request,
				"error_code": status}).Error("Segment event failure. sdk_track call failed.")

			response.Error = "Reception of screen event failed"
			return status, response
		}
		response.EventId = trackResponse.EventId

	case "identify":
		// Identification happens on every call before type switch.
		// Updates the user properties with the traits, here.
		_, status := store.GetStore().UpdateUserProperties(project.ID, userID, &event.Traits, requestTimestamp)
		if status != http.StatusAccepted && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"user_properties": event.Traits,
				"error_code": status}).Error("Segment event failure. Updating user_properties failed.")

			response.Error = "Reception of identify event failed."
			return status, response
		}

	case "group":
		logCtx.WithField("tag", "unsupported").Info("Segment group event.")
		return http.StatusBadRequest, nil

	case "alias":
		logCtx.WithField("tag", "unsupported").Info("Segment alias event.")
		return http.StatusBadRequest, nil
	default:
		response.Error = fmt.Sprintf("Unknown event type %s", event.Type)
		logCtx.Error("Unknown segment event type.")
		return http.StatusBadRequest, response
	}

	response.Message = "Successfully received event"
	return http.StatusOK, response
}

func enrichEventPropertyUsingURL(parsedPageURL *url.URL, eventProperties *U.PropertiesMap) {

	var eventPropertiesURL U.PropertiesMap
	eventPropertiesURL = make(U.PropertiesMap, 0)
	U.FillPropertiesFromURL(&eventPropertiesURL, parsedPageURL)
	if eventPropertiesURL != nil {
		urlToEventPropMap := model.DefaultURLPropertiesToMarketingPropertiesMap()
		// Initialized with already existing event props.
		for urlProp, value := range eventPropertiesURL {
			if eventProp, exists := urlToEventPropMap[urlProp]; exists {
				if _, ok := (*eventProperties)[eventProp]; ok {
					// property seen already
					continue
				} else {
					(*eventProperties)[urlProp] = value
				}
			} else {
				(*eventProperties)[urlProp] = value
			}
		}
	}
}

func ProcessQueueEvent(token, eventJson string) (float64, string, error) {
	logCtx := log.WithField("token", token).WithField("event", eventJson)

	if token == "" || eventJson == "" {
		logCtx.Error("Invalid queue args on segment event queue.")
		return http.StatusInternalServerError, "", nil
	}

	var event Event
	err := json.Unmarshal([]byte(eventJson), &event)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal segment event from queue.")
		// Do not return error to avoid retry.
		return http.StatusInternalServerError, "", nil
	}

	status, response := ReceiveEvent(token, &event)
	responseJsonBytes, _ := json.Marshal(response)
	logCtx = logCtx.WithField("status", status).WithField("response", string(responseJsonBytes))

	metrics.Increment(C.GetSDKAndIntegrationMetricNameByConfig(metrics.IncrIntegrationRequestQueueProcessed))

	// Do not retry on below conditions.
	if status == http.StatusBadRequest ||
		status == http.StatusNotAcceptable ||
		status == http.StatusUnauthorized {

		logCtx.Info("Failed to process segment event permanantly.")
		return float64(status), "", nil
	}

	// Return error only for retry. Retry after a period till it is successfull.
	if status == http.StatusInternalServerError {
		metrics.Increment(C.GetSDKAndIntegrationMetricNameByConfig(metrics.IncrIntegrationRequestQueueRetry))
		return http.StatusInternalServerError, "",
			tasks.NewErrRetryTaskExp("EXP_RETRY_SEGMENT_EVENT_PROCESSING_FAILURE")
	}

	return http.StatusOK, string(responseJsonBytes), nil
}
