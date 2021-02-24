package tests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	C "factors/config"
	H "factors/handler"
	IntSegment "factors/integration/segment"
	"factors/model/model"
	"factors/model/store"

	TaskSession "factors/task/session"
	U "factors/util"
)

func assertKeysExistAndNotEmpty(t *testing.T, obj map[string]interface{}, keys []string) {
	for _, k := range keys {
		assert.NotNil(t, obj[k], fmt.Sprintf("Key %s doesn't exist on %+v", k, obj))
		switch valueType := obj[k].(type) {
		case bool: // Skips empty check for bool.
			log.WithFields(log.Fields{"key": k, "type": valueType}).Debug("Skipping empty check for bool.")
		default:
			assert.NotEmpty(t, obj[k], fmt.Sprintf("Key %s is empty on %+v", k, obj))
		}
	}
}

// expected event properties from segment.
var genericEventProps = []string{U.EP_LOCATION_LATITUDE, U.EP_LOCATION_LONGITUDE, U.EP_SEGMENT_EVENT_VERSION,
	U.EP_SEGMENT_SOURCE_LIBRARY, U.EP_SEGMENT_SOURCE_CHANNEL}
var webEventProps = []string{U.EP_PAGE_RAW_URL, U.EP_PAGE_DOMAIN, U.EP_PAGE_URL, U.EP_PAGE_TITLE,
	U.EP_REFERRER, U.EP_REFERRER_DOMAIN, U.EP_REFERRER_URL, U.EP_CAMPAIGN, U.EP_SOURCE,
	U.EP_MEDIUM, U.EP_KEYWORD, U.EP_CONTENT}

// expected user properties from segment.
var genericUserProps = []string{U.UP_PLATFORM, U.UP_USER_AGENT, U.UP_COUNTRY, U.UP_CITY, U.UP_OS, U.UP_OS_VERSION,
	U.UP_SCREEN_WIDTH, U.UP_SCREEN_HEIGHT}
var webUserProps = []string{}
var mobileUserProps = []string{U.UP_APP_NAME, U.UP_APP_BUILD, U.UP_APP_NAMESPACE, U.UP_APP_VERSION,
	U.UP_DEVICE_ID, U.UP_DEVICE_NAME, U.UP_DEVICE_ADVERTISING_ID, U.UP_DEVICE_MODEL, U.UP_DEVICE_TYPE,
	U.UP_DEVICE_MANUFACTURER, U.UP_DEVICE_ADTRACKING_ENABLED, U.UP_NETWORK_CARRIER, U.UP_NETWORK_BLUETOOTH,
	U.UP_NETWORK_CELLULAR, U.UP_NETWORK_WIFI, U.UP_SCREEN_DENSITY, U.UP_TIMEZONE, U.UP_LOCALE}

func TestIntSegmentHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Not enabled.
	w := ServePostRequestWithHeaders(r, uri, []byte(`{"anonymousId": "ranon_2", "type": "random_type"}`),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)

	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Empty body.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code) // status ok with error
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Nil(t, jsonResponseMap["event_id"])
	assert.Nil(t, jsonResponseMap["user_id"])
	assert.NotNil(t, jsonResponseMap["error"])

	// Invalid type.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{"anonymousId": "ranon_1", "type": "random_type", "timestamp": "2015-02-23T22:28:55.111Z"}`),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code) // status ok with error
	jsonResponse1, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap1 map[string]interface{}
	json.Unmarshal(jsonResponse1, &jsonResponseMap1)
	assert.NotNil(t, jsonResponseMap1["error"])
	assert.NotNil(t, jsonResponseMap1["type"])
	assert.Equal(t, "random_type", jsonResponseMap1["type"])
	assert.Nil(t, jsonResponseMap1["event_id"])
	assert.NotNil(t, jsonResponseMap1["user_id"]) // Always return user_id.

	// Without both anonymousId and userId
	w = ServePostRequestWithHeaders(r, uri, []byte(`{"type": "track"}`),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)

	// With only anonymousId
	identifyPayloadWithoutUserId := `
	{
		"anonymousId": "507f191e810c19729de860ea",
		"channel": "browser",
		"context": {
			"ip": "8.8.8.8",
			"userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.115 Safari/537.36"
		},
		"integrations": {
			"All": false,
			"Mixpanel": true,
			"Salesforce": true
		},
		"messageId": "022bb90c-bbac-11e4-8dfc-aa07a5b093db",
		"receivedAt": "2015-02-23T22:28:55.387Z",
		"sentAt": "2015-02-23T22:28:55.111Z",
		"timestamp": "2015-02-23T22:28:55.111Z",
		"traits": {
			"email": "peter@initech.com",
			"plan": "premium",
			"address": {
				"street": "6th St",
				"city": "San Francisco"
			}
		},
		"type": "identify",
		"userId": "",
		"version": "1.1"
	}
	`
	w = ServePostRequestWithHeaders(r, uri, []byte(identifyPayloadWithoutUserId),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.NotNil(t, jsonResponseMap2["user_id"])

	// Test invalid event timestamp
	samplePayloadWithInvalidTimestamp := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "INVALID_TIMESTAMP",
		"name": "screen_1",
		"type": "screen",
		"userId": "",
		"version": "1.1"
	  }
	`

	w = ServePostRequestWithHeaders(r, uri, []byte(samplePayloadWithInvalidTimestamp),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.NotNil(t, jsonResponseMap2["error"])
}

func TestIntSegmentHandlerWithPageEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// invalid private token.
	w := ServePostRequestWithHeaders(r, uri, []byte(`{}`),
		map[string]string{"Authorization": "invalid_token"})
	assert.Equal(t, http.StatusOK, w.Code) // status ok with error
	jsonResponse9, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap9 map[string]interface{}
	json.Unmarshal(jsonResponse9, &jsonResponseMap9)
	assert.NotNil(t, jsonResponseMap9["error"])

	// Page.
	samplePagePayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"type": "page",
		"userId": "",
		"version": "1.1"
	  }
	`
	w = ServePostRequestWithHeaders(r, uri, []byte(samplePagePayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.Nil(t, jsonResponseMap2["error"])
	assert.NotNil(t, jsonResponseMap2["event_id"])
	// Check event properties added.
	retEvent, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	eventPropertiesBytes, err := retEvent.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap, genericEventProps)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap, webEventProps)
	// Check event properties added.
	retUser, errCode := store.GetStore().GetUser(project.ID, retEvent.UserId)
	assert.NotNil(t, retUser)
	userPropertiesBytes, err := retUser.Properties.Value()
	var userPropertiesMap map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, genericUserProps)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, webUserProps)

	t.Run("PageEventWithQueryParamsEventProperties", func(t *testing.T) {
		samplePagePayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://www.example.com/blog?token1=yyy&token2=xxx"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37a",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10", 
		  "title": "Segment Test",
		  "url": "https://www.example.com/blog?token1=yyy&token2=xxx"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"type": "page",
		"userId": "",
		"version": "1.1"
	  }
	`
		w = ServePostRequestWithHeaders(r, uri, []byte(samplePagePayload),
			map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse2, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap2 map[string]interface{}
		json.Unmarshal(jsonResponse2, &jsonResponseMap2)
		assert.Nil(t, jsonResponseMap2["error"])
		assert.NotNil(t, jsonResponseMap2["event_id"])
		// Check event properties added.
		retEvent, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
		assert.Equal(t, http.StatusFound, errCode)
		eventPropertiesBytes, err := retEvent.Properties.Value()
		assert.Nil(t, err)
		var eventPropertiesMap map[string]interface{}
		json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		assert.NotNil(t, eventPropertiesMap[U.QUERY_PARAM_PROPERTY_PREFIX+"token1"])
		assert.Equal(t, "yyy", eventPropertiesMap[U.QUERY_PARAM_PROPERTY_PREFIX+"token1"])
		assert.NotNil(t, eventPropertiesMap[U.QUERY_PARAM_PROPERTY_PREFIX+"token2"])
		assert.Equal(t, "xxx", eventPropertiesMap[U.QUERY_PARAM_PROPERTY_PREFIX+"token2"])
	})
}

func TestIntSegmentHandlePageEventWithFilterExpression(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	// disable := false
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntSegment: &enable, AutoTrack: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Filter.
	filterEventName, errCode := store.GetStore().CreateOrGetFilterEventName(&model.EventName{ProjectId: project.ID,
		Name: "MyAccountDiscover", FilterExpr: "www.livspace.com/my-account/discover/:id"})
	assert.NotNil(t, filterEventName)
	assert.Equal(t, http.StatusCreated, errCode)

	// Page.
	samplePagePayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://www.livspace.com/my-account/discover/1"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/my-account/discover/1",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "https://www.livspace.com/my-account/discover/1"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"type": "page",
		"userId": "",
		"version": "1.1"
	  }
	`
	w := ServePostRequestWithHeaders(r, uri, []byte(samplePagePayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.Nil(t, jsonResponseMap2["error"])
	assert.NotNil(t, jsonResponseMap2["event_id"])
	event, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	// event should use filter expr event name.
	assert.Equal(t, filterEventName.ID, event.EventNameId)

	// Filter1.
	filterEventName1, errCode := store.GetStore().CreateOrGetFilterEventName(&model.EventName{ProjectId: project.ID,
		Name: "MyAccountDiscover", FilterExpr: "www.livspace.com/:loc_id/magazine/*"})
	assert.NotNil(t, filterEventName1)
	assert.Equal(t, http.StatusCreated, errCode)

	// Page.
	samplePagePayload1 := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://www.livspace.com/in/magazine/best-livspace-blog-posts-2017"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37el",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/in/magazine/best-livspace-blog-posts-2017",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "https://www.livspace.com/in/magazine/best-livspace-blog-posts-2017"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"type": "page",
		"userId": "",
		"version": "1.1"
	  }
	`
	w = ServePostRequestWithHeaders(r, uri, []byte(samplePagePayload1),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ = ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap)
	assert.Nil(t, jsonResponseMap["error"])
	assert.NotNil(t, jsonResponseMap["event_id"])
	event1, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	// event should use filter expr event name.
	assert.NotEqual(t, filterEventName1.ID, event1.EventNameId)
}

func TestIntSegmentHandlerWithSession(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	t.Run("CreateNewSesssionForNewUser", func(t *testing.T) {
		timestamp := U.UnixTimeBeforeDuration(30 * 24 * time.Hour)
		eventTimestamp := time.Unix(timestamp, 0).Format(time.RFC3339)
		// Page.
		samplePagePayload := fmt.Sprintf(`
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d990",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "%s",
		"type": "page",
		"userId": "xxx123",
		"version": "1.1"
	  }
	`, eventTimestamp)

		w := ServePostRequestWithHeaders(r, uri, []byte(samplePagePayload),
			map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse2, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap2 map[string]interface{}
		json.Unmarshal(jsonResponse2, &jsonResponseMap2)
		assert.Nil(t, jsonResponseMap2["error"])
		assert.NotEmpty(t, jsonResponseMap2["event_id"])
		assert.NotEmpty(t, jsonResponseMap2["user_id"])

		_, err := TaskSession.AddSession([]uint64{project.ID}, timestamp-60, 0, 0, 0, 1)
		assert.Nil(t, err)

		event, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, event.SessionId)

		sessionEvent, errCode := store.GetStore().GetEventById(project.ID, *event.SessionId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, sessionEvent)

		sessionPropertiesBytes, err := sessionEvent.Properties.Value()
		assert.Nil(t, err)
		var sessionProperties map[string]interface{}
		json.Unmarshal(sessionPropertiesBytes.([]byte), &sessionProperties)

		assert.NotEmpty(t, sessionProperties[U.SP_IS_FIRST_SESSION])
		assert.True(t, sessionProperties[U.SP_IS_FIRST_SESSION].(bool))
	})
}

func TestSegmentTrackEventForBlockedToken(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Block by token.
	C.GetConfig().BlockedSDKRequestProjectTokens = []string{project.PrivateToken}

	sampleTrackPayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": 5.6,
			  "build": 1.1,
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": 2.4
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"event": "click_1",
		"type": "track",
		"userId": "",
		"version": 3.1
	  }
	`

	w := ServePostRequestWithHeaders(r, uri, []byte(sampleTrackPayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	response, _ := ioutil.ReadAll(w.Body)
	var responseMap map[string]interface{}
	json.Unmarshal(response, &responseMap)
	assert.Equal(t, "Request failed. Blocked.", responseMap["error"])
}

func TestIntSegmentHandlerWithTrackEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Inconsistent datatype tested with App(build, version),
	// OS(version) and Event(version) as numbers.
	sampleTrackPayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": 5.6,
			  "build": 1.1,
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": 2.4
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"event": "click_1",
		"type": "track",
		"userId": "",
		"version": 3.1
	  }
	`

	w := ServePostRequestWithHeaders(r, uri, []byte(sampleTrackPayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.Nil(t, jsonResponseMap2["error"])
	assert.NotNil(t, jsonResponseMap2["event_id"])
	// Check event properties added.
	retEvent, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	eventPropertiesBytes, err := retEvent.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap, genericEventProps)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap, webEventProps)
	// Check event properties added.
	retUser, errCode := store.GetStore().GetUser(project.ID, retEvent.UserId)
	assert.NotNil(t, retUser)
	userPropertiesBytes, err := retUser.Properties.Value()
	var userPropertiesMap map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, genericUserProps)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, webUserProps)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, mobileUserProps)

	// create track event with same messageId
	w = ServePostRequestWithHeaders(r, uri, []byte(sampleTrackPayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse4, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap4 map[string]interface{}
	json.Unmarshal(jsonResponse4, &jsonResponseMap4)
	assert.Nil(t, jsonResponseMap4["error"])

	sampleTrackPayloadWithoutProperties := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [

		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37a",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"event": "click_1",
		"type": "track",
		"userId": "",
		"version": "2"
	  }
	`

	w = ServePostRequestWithHeaders(r, uri, []byte(sampleTrackPayloadWithoutProperties),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse3, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap3 map[string]interface{}
	json.Unmarshal(jsonResponse3, &jsonResponseMap3)
	assert.Nil(t, jsonResponseMap3["error"])
	assert.NotNil(t, jsonResponseMap3["event_id"])
	// Check event properties added.
	retEvent1, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap3["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	eventPropertiesBytes1, err := retEvent1.Properties.Value()
	var eventPropertiesMap1 map[string]interface{}
	json.Unmarshal(eventPropertiesBytes1.([]byte), &eventPropertiesMap1)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap1, genericEventProps)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap1, webEventProps)
	// Check event properties added.
	retUser, errCode = store.GetStore().GetUser(project.ID, retEvent1.UserId)
	assert.NotNil(t, retUser)
	userPropertiesBytes1, err := retUser.Properties.Value()
	var userPropertiesMap1 map[string]interface{}
	json.Unmarshal(userPropertiesBytes1.([]byte), &userPropertiesMap1)
	assertKeysExistAndNotEmpty(t, userPropertiesMap1, genericUserProps)
	assertKeysExistAndNotEmpty(t, userPropertiesMap1, webUserProps)
	assertKeysExistAndNotEmpty(t, userPropertiesMap1, mobileUserProps)
}

func TestIntSegmentHandlerWithScreenEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	sampleScreenPayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"name": "screen_1",
		"type": "screen",
		"userId": "",
		"version": "2"
	  }
	`

	w := ServePostRequestWithHeaders(r, uri, []byte(sampleScreenPayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.Nil(t, jsonResponseMap2["error"])
	assert.NotNil(t, jsonResponseMap2["event_id"])
	// Check event properties added.
	retEvent, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	eventPropertiesBytes, err := retEvent.Properties.Value()
	var eventPropertiesMap map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
	assertKeysExistAndNotEmpty(t, eventPropertiesMap, genericEventProps)
	// Check event properties added.
	retUser, errCode := store.GetStore().GetUser(project.ID, retEvent.UserId)
	assert.NotNil(t, retUser)
	userPropertiesBytes, err := retUser.Properties.Value()
	var userPropertiesMap map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, genericUserProps)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, mobileUserProps)
}

func TestIntSegmentHandlerWithIdentifyEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	sampleIdentifyPayload := `
	{
		"anonymousId": "anon_99",
		"channel": "browser",
		"context": {
			"ip": "8.8.8.8",
			"userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.115 Safari/537.36"
		},
		"integrations": {
			"All": false,
			"Mixpanel": true,
			"Salesforce": true
		},
		"messageId": "022bb90c-bbac-11e4-8dfc-aa07a5b093db",
		"receivedAt": "2015-02-23T22:28:55.387Z",
		"sentAt": "2015-02-23T22:28:55.111Z",
		"timestamp": "2015-02-23T22:28:55.111Z",
		"traits": {
			"email": "peter@initech.com",
			"plan": "premium",
			"address": {
				"street": "6th St",
				"city": "San Francisco",
				"state": "CA",
				"postalCode": "94103",
				"country": "USA"
			}
		},
		"type": "identify",
		"userId": "user_99",
		"version": "1.1"
	}
	`

	w := ServePostRequestWithHeaders(r, uri, []byte(sampleIdentifyPayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.Nil(t, jsonResponseMap2["error"])
	assert.NotNil(t, jsonResponseMap2["user_id"])
	// Check event properties added.
	retUser, _ := store.GetStore().GetUser(project.ID, jsonResponseMap2["user_id"].(string))
	assert.NotNil(t, retUser)
	userPropertiesBytes, err := retUser.Properties.Value()
	var userPropertiesMap map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
	assertKeysExistAndNotEmpty(t, userPropertiesMap, []string{"email", "address", "plan"})
	// validates nested properties.
	assertKeysExistAndNotEmpty(t, userPropertiesMap["address"].(map[string]interface{}), []string{"street", "city"})
}

func TestIntSegmentHandlerWithPayloadFromSegmentPlatform(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment_platform"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// create basic auth token.
	tokenWithColon := fmt.Sprintf("%s:", project.PrivateToken)
	base64TokenWithColon := base64.StdEncoding.EncodeToString([]byte(tokenWithColon))
	basicAuthToken := fmt.Sprintf("Basic %s", base64TokenWithColon)

	t.Run("PlatformTestIdentifyPayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"userId": "identified-1",
			"type": "identify",
			"timestamp": "2019-06-24T15:32:33Z",
			"traits": {
				"email": "calvinfo@segment.com",
				"first_name": "Calvin",
				"last_name": "French-Owen",
				"phone": "555-555-5555"
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PlatformTestIdentify:UserSignupPayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"userId": "identified-1",
			"type": "track",
			"timestamp": "2019-06-24T15:32:33Z",
			"event": "Signed up",
			"properties": {
				"referrer": "paid"
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PlatformTestGroup:UserIsGroupedIntoMyUsersPayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"userId": "identified-1",
			"type": "group",
			"groupId": "myUsers",
			"timestamp": "2019-06-24T15:32:33Z",
			"traits": {
				"name": "Initech",
				"industry": "Technology",
				"employees": 329,
				"plan": "enterprise",
				"total billed": 830
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PlatformTestTrack:UserEnablesIntegrationPayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"userId": "identified-1",
			"type": "track",
			"timestamp": "2019-06-24T15:32:33Z",
			"event": "Integration Enabled",
			"properties": {
				"name": "Google Analytics"
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PlatformTestScreen:UsersOpensMobileApplicationPayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"userId": "identified-1",
			"type": "track",
			"timestamp": "2019-06-24T15:32:33Z",
			"event": "Integration Enabled",
			"properties": {
				"name": "Google Analytics"
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PlatformTestPage:AnonymousUserNavigationToHomePagePayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"type": "page",
			"timestamp": "2019-06-24T15:32:33Z",
			"name": "Home",
			"properties": {
				"url": "https://segment.com"
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PlatformTestPage:AnonymousUserNavigationToSignupPagePayload", func(t *testing.T) {
		payload := `
		{
			"anonymousId": "anonymous-1",
			"type": "page",
			"timestamp": "2019-06-24T15:32:33Z",
			"name": "Signup",
			"properties": {
				"url": "https://segment.com/signup" 
			}
		}
		`

		w := ServePostRequestWithHeaders(r, uri, []byte(payload),
			map[string]string{"Authorization": basicAuthToken})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIntSegmentHandlerWithTimestamp(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/integrations/segment"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	eventTimestamp := time.Now().Add(time.Hour * -1)
	eventTimestampInRFC3369 := eventTimestamp.UTC().Format(time.RFC3339)
	eventTimestampInUnix := eventTimestamp.UTC().Unix()

	// Page.
	samplePagePayload := fmt.Sprintf(`
{
	"_metadata": {
		"bundled": [
		"Segment.io"
		],
		"unbundled": []
	},
	"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
	"channel": "client",
	"context": {
		"active": true,
		"app": {
			"name": "InitechGlobal",
			"version": "545",
			"build": "3.0.1.545",
			"namespace": "com.production.segment"
		},
		"campaign": {
			"name": "TPS Innovation Newsletter",
			"source": "Newsletter",
			"medium": "email",
			"term": "tps reports",
			"content": "image link"
		},
		"device": {
			"id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			"advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			"adTrackingEnabled": true,
			"manufacturer": "Apple",
			"model": "iPhone7,2",
			"name": "maguro",
			"type": "ios",
			"token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
		},
		"ip": "8.8.8.8",
		"library": {
			"name": "analytics.js",
			"version": "2.11.1"
		},
		"locale": "nl-NL",
		"location": {
			"city": "San Francisco",
			"country": "United States",
			"latitude": 40.2964197,
			"longitude": -76.9411617,
			"speed": 0
		},
		"network": {
			"bluetooth": false,
			"carrier": "T-Mobile NL",
			"cellular": true,
			"wifi": false
		},
		"os": {
			"name": "iPhone OS",
			"version": "8.1.3"
		},
		"page": {
			"path": "/academy/",
			"referrer": "https://google.com",
			"search": "",
			"title": "Analytics Academy",
			"url": "https://segment.com/academy/"
		},
		"referrer": {
			"id": "ABCD582CDEFFFF01919",
			"type": "dataxu"
		},
		"screen": {
			"width": 320,
			"height": 568,
			"density": 2
		},
		"groupId": "12345",
		"timezone": "Europe/Amsterdam",
		"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
	},
	"integrations": {},
	"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
	"originalTimestamp": "2019-01-08T16:22:06.053Z",
	"projectId": "Zzft38QJhB",
	"properties": {
		"path": "/segment.test.html",
		"referrer": "",
		"search": "?a=10",
		"title": "Segment Test",
		"url": "http://localhost:8090/segment.test.html?a=10"
	},
	"receivedAt": "2019-01-08T16:21:54.106Z",
	"sentAt": "2019-01-08T16:22:06.058Z",
	"timestamp": "%s",
	"type": "page",
	"userId": "",
	"version": "1.1"
	}
`, eventTimestampInRFC3369)

	w := ServePostRequestWithHeaders(r, uri, []byte(samplePagePayload),
		map[string]string{"Authorization": project.PrivateToken})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse2, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap2 map[string]interface{}
	json.Unmarshal(jsonResponse2, &jsonResponseMap2)
	assert.Nil(t, jsonResponseMap2["error"])
	assert.NotNil(t, jsonResponseMap2["event_id"])
	// Check event properties added.
	retEvent, errCode := store.GetStore().GetEventById(project.ID, jsonResponseMap2["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, retEvent)
	assert.Equal(t, eventTimestampInUnix, retEvent.Timestamp)
}

func TestSegmentEventWithQueue(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{IntSegment: &enable})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Page.
	samplePagePayload := `
	{
		"_metadata": {
		  "bundled": [
			"Segment.io"
		  ],
		  "unbundled": [
			
		  ]
		},
		"anonymousId": "80444c7e-1580-4d3c-a77a-2f3427ed7d97",
		"channel": "client",
		"context": {
			"active": true,
			"app": {
			  "name": "InitechGlobal",
			  "version": "545",
			  "build": "3.0.1.545",
			  "namespace": "com.production.segment"
			},
			"campaign": {
			  "name": "TPS Innovation Newsletter",
			  "source": "Newsletter",
			  "medium": "email",
			  "term": "tps reports",
			  "content": "image link"
			},
			"device": {
			  "id": "B5372DB0-C21E-11E4-8DFC-AA07A5B093DB",
			  "advertisingId": "7A3CBEA0-BDF5-11E4-8DFC-AA07A5B093DB",
			  "adTrackingEnabled": true,
			  "manufacturer": "Apple",
			  "model": "iPhone7,2",
			  "name": "maguro",
			  "type": "ios",
			  "token": "ff15bc0c20c4aa6cd50854ff165fd265c838e5405bfeb9571066395b8c9da449"
			},
			"ip": "8.8.8.8",
			"library": {
			  "name": "analytics.js",
			  "version": "2.11.1"
			},
			"locale": "nl-NL",
			"location": {
			  "city": "San Francisco",
			  "country": "United States",
			  "latitude": 40.2964197,
			  "longitude": -76.9411617,
			  "speed": 0
			},
			"network": {
			  "bluetooth": false,
			  "carrier": "T-Mobile NL",
			  "cellular": true,
			  "wifi": false
			},
			"os": {
			  "name": "iPhone OS",
			  "version": "8.1.3"
			},
			"page": {
			  "path": "/academy/",
			  "referrer": "https://google.com",
			  "search": "",
			  "title": "Analytics Academy",
			  "url": "https://segment.com/academy/"
			},
			"referrer": {
			  "id": "ABCD582CDEFFFF01919",
			  "type": "dataxu"
			},
			"screen": {
			  "width": 320,
			  "height": 568,
			  "density": 2
			},
			"groupId": "12345",
			"timezone": "Europe/Amsterdam",
			"userAgent": "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
		},
		"integrations": {},
		"messageId": "ajs-19c084e2f80e70cf62bb62509e79b37e",
		"originalTimestamp": "2019-01-08T16:22:06.053Z",
		"projectId": "Zzft38QJhB",
		"properties": {
		  "path": "/segment.test.html",
		  "referrer": "",
		  "search": "?a=10",
		  "title": "Segment Test",
		  "url": "http://localhost:8090/segment.test.html?a=10"
		},
		"receivedAt": "2019-01-08T16:21:54.106Z",
		"sentAt": "2019-01-08T16:22:06.058Z",
		"timestamp": "2019-01-08T16:21:54.101Z",
		"type": "page",
		"userId": "",
		"version": "1.1"
	  }
	`

	var event IntSegment.Event
	err = json.Unmarshal([]byte(samplePagePayload), &event)
	assert.Nil(t, err)
	status, response := IntSegment.ReceiveEventWithQueue(project.PrivateToken,
		&event, []string{project.PrivateToken})
	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, response.Error)
}
