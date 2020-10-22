package tests

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"

	C "factors/config"
	H "factors/handler"
	M "factors/model"
	SDK "factors/sdk"
	U "factors/util"
)

func TestSDKTrackHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// Test without project_id scope and with non-existing project.
	w := ServePostRequest(r, uri, []byte("{}"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with invalid token.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`),
		map[string]string{"Authorization": "INVALID_TOKEN"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test without user_id in the payload.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{"event_name": "signup", "event_properties": {"mobile" : "true"}}`),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["user_id"])

	// Test without event_name in the payload.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{"event_properties": {"mobile" : "true"}}`),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test track bot with exclude_bot on.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_name": "%s", "event_properties": {"mobile" : "true"}}`, U.RandomString(8))),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		})
	assert.Equal(t, http.StatusNotModified, w.Code)

	// Test track bot with exclude_bot off.
	botState := false
	M.UpdateProjectSettings(project.ID, &M.ProjectSetting{ExcludeBot: &botState})
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_name": "%s", "event_properties": {"mobile" : "true"}}`, U.RandomString(8))),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	// Test without properties and with empty properites in the payload.
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_1", "event_properties": {}, "user_properties": {"$os": "Mac OS"}}`, user.ID)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_2"}`, user.ID)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with empty user properties in the payload.
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_1", "event_properties": {"mobile" : "true"}, "user_properties": {}}`, user.ID)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_2"}`, user.ID)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Create with customer_event_id
	CustEventId := U.RandomString(8)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_2", "c_event_id":"%s"}`, user.ID, CustEventId)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	// Duplicate customer_event_id
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_2", "c_event_id":"%s"}`, user.ID, CustEventId)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusNotAcceptable, w.Code)

	// Test auto tracked event.
	rEventName := "example.com/"
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$page_load_time": 0, "$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, user.ID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	rEvent, errCode := M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent)
	retEventName, err := M.GetEventNameFromEventNameId(rEvent.EventNameId, project.ID)
	assert.Nil(t, err)
	assert.Equal(t, retEventName.Name, "example.com")
	eventPropertiesBytes, err := rEvent.Properties.Value()
	assert.Nil(t, err)
	var eventProperties map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventProperties)
	assert.Nil(t, eventProperties["$dollar_property"])
	assert.NotNil(t, eventProperties[fmt.Sprintf("%s$dollar_property", U.NAME_PREFIX_ESCAPE_CHAR)]) // escaped property should exist.
	assert.NotNil(t, eventProperties["$qp_search"])                                                 // $qp should exist.
	assert.Nil(t, eventProperties[fmt.Sprintf("%s$qp_search", U.NAME_PREFIX_ESCAPE_CHAR)])          // $qp should not be escaped.
	assert.NotNil(t, eventProperties["mobile"])                                                     // no dollar properties should exist.
	assert.NotNil(t, eventProperties["$qp_encoded"])                                                // URL encoded property should exist.
	assert.Equal(t, "google search", eventProperties["$qp_encoded"])                                // decoded property value should have been stored.
	assert.Nil(t, eventProperties["$qp_utm_keyword"])                                               // $qp_utm_keyword mapped to $keyword should also be decoded.
	assert.Equal(t, "google search", eventProperties[U.EP_KEYWORD])
	assert.Equal(t, float64(1), eventProperties[U.EP_PAGE_SPENT_TIME])     // Should be default value.
	assert.Equal(t, float64(1), eventProperties[U.EP_PAGE_LOAD_TIME])      // Should be default value.
	assert.Equal(t, float64(0), eventProperties[U.EP_PAGE_SCROLL_PERCENT]) // Should be default value.
	assert.True(t, eventProperties[U.EP_IS_PAGE_VIEW].(bool))
	assert.True(t, len(rEvent.UserPropertiesId) > 0)
	rUser, errCode := M.GetUser(rEvent.ProjectId, rEvent.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rUser)
	userPropertiesBytes, err := rUser.Properties.Value()
	assert.Nil(t, err)
	var userProperties map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userProperties)
	assert.NotNil(t, userProperties["name"])
	// OS and Browser Properties should be filled on backend using user agent.
	assert.Equal(t, "Mac OS X", userProperties[U.UP_OS])
	assert.Equal(t, "10.13.6", userProperties[U.UP_OS_VERSION])
	assert.Equal(t, "Mac OS X-10.13.6", userProperties[U.UP_OS_WITH_VERSION])
	assert.Equal(t, "Chrome", userProperties[U.UP_BROWSER])
	assert.Equal(t, "79.0.3945.130", userProperties[U.UP_BROWSER_VERSION])
	assert.Equal(t, "Chrome-79.0.3945.130", userProperties[U.UP_BROWSER_WITH_VERSION])

	// Should not allow $ prefixes apart from default properties.
	rEventName = U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "event_properties": {"mobile": "true", "$referrer": "http://google.com", "$page_raw_url": "https://factors.ai/login/", "$page_title": "Login", "$page_load_time": 10}, "user_properties": {"$dollar_key": "unknow_value", "$os": "mac osx", "$os_version": "1_2_3", "$screen_width": 10, "$screen_height": 11, "$browser": "mozilla", "$platform": "web", "$browser_version": "10_2_3"}}`,
		user.ID, rEventName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	propsResponseMap1 := DecodeJSONResponseToMap(w.Body)
	assert.Nil(t, propsResponseMap1["user_id"])
	assert.NotNil(t, propsResponseMap1["event_id"])
	retEvent, errCode := M.GetEvent(project.ID, user.ID, propsResponseMap1["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	// check event default properties and tracked properties.
	eventPropertiesBytes, err = retEvent.Properties.Value()
	assert.Nil(t, err)
	var eventProperties1 map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventProperties1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, eventProperties1["mobile"])
	assert.NotNil(t, eventProperties1[U.EP_REFERRER])
	assert.NotNil(t, eventProperties1[U.EP_PAGE_RAW_URL])
	assert.NotNil(t, eventProperties1[U.EP_PAGE_TITLE])
	assert.Nil(t, eventProperties1[U.NAME_PREFIX_ESCAPE_CHAR+U.EP_REFERRER])
	assert.Nil(t, eventProperties1[U.NAME_PREFIX_ESCAPE_CHAR+U.EP_PAGE_RAW_URL])
	assert.Nil(t, eventProperties1[U.NAME_PREFIX_ESCAPE_CHAR+U.EP_PAGE_TITLE])
	// Should not overwrite non-zero value of page_load_time to default.
	assert.NotEqual(t, float64(10), eventProperties[U.EP_PAGE_LOAD_TIME])
	// Should assign page_load_time to page_spent_time when page_spent_time is not available.
	assert.NotEqual(t, float64(10), eventProperties[U.EP_PAGE_SPENT_TIME])
	// check user default properties.
	retUser, errCode := M.GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesBytes, err = retUser.Properties.Value()
	assert.Nil(t, err)
	var userPropertiesMap3 map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap3)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, userPropertiesMap3["$dollar_key"])                                              // dollar prefix not allowed.
	assert.NotNil(t, userPropertiesMap3[fmt.Sprintf("%s$dollar_key", U.NAME_PREFIX_ESCAPE_CHAR)]) // Escaped key should exist.
	assert.NotNil(t, userPropertiesMap3[U.UP_OS])
	assert.NotNil(t, userPropertiesMap3[U.UP_OS_VERSION])
	assert.NotNil(t, userPropertiesMap3[U.UP_BROWSER])
	assert.NotNil(t, userPropertiesMap3[U.UP_PLATFORM])
	assert.NotNil(t, userPropertiesMap3[U.UP_SCREEN_WIDTH])
	assert.NotNil(t, userPropertiesMap3[U.UP_SCREEN_HEIGHT])
	assert.NotNil(t, userPropertiesMap3[U.UP_DAY_OF_FIRST_EVENT])
	assert.Equal(t, time.Unix(rEvent.Timestamp, 0).Weekday().String(), userPropertiesMap3[U.UP_DAY_OF_FIRST_EVENT])
	retUserFirstVisitHour, _, _ := time.Unix(rEvent.Timestamp, 0).Clock()
	assert.NotNil(t, userPropertiesMap3[U.UP_HOUR_OF_FIRST_EVENT])
	assert.Equal(t, float64(retUserFirstVisitHour), userPropertiesMap3[U.UP_HOUR_OF_FIRST_EVENT])

	// should not be escaped.
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_OS])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_OS_VERSION])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_BROWSER])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_PLATFORM])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_SCREEN_WIDTH])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_SCREEN_HEIGHT])

	// Test event is using existing filter or not.
	// Created filter.
	expr := "a.com/u1/u2/:prop1"
	name := "login"
	filterEventName, errCode := M.CreateOrGetFilterEventName(&M.EventName{
		ProjectId:  project.ID,
		FilterExpr: expr,
		Name:       name,
	})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, filterEventName)
	assert.NotZero(t, filterEventName.ID)
	assert.Equal(t, name, filterEventName.Name)
	assert.Equal(t, expr, filterEventName.FilterExpr)
	assert.Equal(t, M.TYPE_FILTER_EVENT_NAME, filterEventName.Type)

	// Test filter_event_name hit with exact match.
	rEventName = "a.com/u1/u2/i1"
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
		user.ID, rEventName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	rEvent, errCode = M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent)
	assert.Equal(t, filterEventName.ID, rEvent.EventNameId)
	var rEventProperties map[string]interface{}
	json.Unmarshal(rEvent.Properties.RawMessage, &rEventProperties)
	assert.NotNil(t, rEventProperties["prop1"])
	assert.Equal(t, "i1", rEventProperties["prop1"]) // Event property filled using expression.

	// Test filter_event_name hit with raw event_url.
	rEventName = "https://a.com/u1/u2/i2/u4/u5?q=search_string"
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
		user.ID, rEventName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	rEvent, errCode = M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent)
	assert.Equal(t, filterEventName.ID, rEvent.EventNameId)
	var rEventProperties2 map[string]interface{}
	json.Unmarshal(rEvent.Properties.RawMessage, &rEventProperties2)
	assert.NotNil(t, rEventProperties2["prop1"])
	assert.Equal(t, "i2", rEventProperties2["prop1"])

	// Test filter_event_name miss created auto_tracked event_name.
	rEventName = fmt.Sprintf("%s/%s", "a.com/u1/u5/u7", U.RandomLowerAphaNumString(5))
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
		user.ID, rEventName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	rEvent, errCode = M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent)
	eventName, errCode := M.GetEventName(rEventName, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, eventName)
	assert.Equal(t, M.TYPE_AUTO_TRACKED_EVENT_NAME, eventName.Type)

	// Test filter_event_name miss after filter deleted by user.
	errCode = M.DeleteFilterEventName(project.ID, filterEventName.ID)
	assert.Equal(t, http.StatusAccepted, errCode)
	rEventName = "a.com/u1/u2/i1"
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
		user.ID, rEventName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	rEvent, errCode = M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent)
	eventName, errCode = M.GetEventName(rEventName, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, eventName)
	assert.NotEqual(t, filterEventName.ID, eventName.ID)            // should not use deleted filter.
	assert.Equal(t, M.TYPE_AUTO_TRACKED_EVENT_NAME, eventName.Type) // should create auto created event.

	t.Run("FilterExpressionWithHash", func(t *testing.T) {
		expr := "factors-dev.com/#/reports/:report_id"
		name := "seen_reports"
		filterEventName, errCode := M.CreateOrGetFilterEventName(&M.EventName{
			ProjectId:  project.ID,
			FilterExpr: expr,
			Name:       name,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, filterEventName)
		assert.NotZero(t, filterEventName.ID)
		assert.Equal(t, name, filterEventName.Name)
		assert.Equal(t, expr, filterEventName.FilterExpr)
		assert.Equal(t, M.TYPE_FILTER_EVENT_NAME, filterEventName.Type)

		rEventName = "factors-dev.com/#/reports/1234"
		w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(
			`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
			user.ID, rEventName)), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.Nil(t, responseMap["user_id"])
		rEvent, errCode = M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rEvent)
		assert.Equal(t, filterEventName.ID, rEvent.EventNameId)
		var rEventProperties map[string]interface{}
		json.Unmarshal(rEvent.Properties.RawMessage, &rEventProperties)
		assert.NotNil(t, rEventProperties["report_id"])
		assert.Equal(t, "1234", rEventProperties["report_id"])
	})

	t.Run("MapEventPropertiesToDefaultProperties", func(t *testing.T) {
		rEventName := "https://example.com/" + U.RandomLowerAphaNumString(10)
		w = ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "event_properties": {"mobile": "true", "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_source": "google","$qp_utm_term":"analytics", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroup_id": "xyz123", "$qp_utm_creativeid": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$os": "Mac OS"}}`, user.ID, rEventName)),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		rEvent, errCode := M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rEvent)
		eventPropertiesBytes, err := rEvent.Properties.Value()
		assert.Nil(t, err)
		var eventProperties map[string]interface{}
		json.Unmarshal(eventPropertiesBytes.([]byte), &eventProperties)
		// other properties should be present.
		assert.NotNil(t, eventProperties["mobile"])

		// property name should be replaced with corresponding
		// default property.
		assert.Nil(t, eventProperties["$qp_utm_campaign"])
		assert.NotNil(t, eventProperties[U.EP_CAMPAIGN])
		assert.Nil(t, eventProperties["$qp_utm_campaignid"])
		assert.NotNil(t, eventProperties[U.EP_CAMPAIGN_ID])
		assert.Nil(t, eventProperties["$qp_utm_source"])
		assert.NotNil(t, eventProperties[U.EP_SOURCE])
		assert.Nil(t, eventProperties["$qp_utm_medium"])
		assert.NotNil(t, eventProperties[U.EP_MEDIUM])
		assert.Nil(t, eventProperties["$qp_utm_keyword"])
		assert.NotNil(t, eventProperties[U.EP_KEYWORD])
		assert.Nil(t, eventProperties["$qp_utm_term"])
		assert.NotNil(t, eventProperties[U.EP_TERM])
		assert.Nil(t, eventProperties["$qp_utm_matchtype"])
		assert.NotNil(t, eventProperties[U.EP_KEYWORD_MATCH_TYPE])
		assert.Nil(t, eventProperties["$qp_utm_content"])
		assert.NotNil(t, eventProperties[U.EP_CONTENT])
		assert.Nil(t, eventProperties["$qp_utm_adgroup"])
		assert.NotNil(t, eventProperties[U.EP_ADGROUP])
		assert.Nil(t, eventProperties["$qp_gclid"])
		assert.NotNil(t, eventProperties[U.EP_GCLID])
		assert.Nil(t, eventProperties["$qp_fbclid"])
		assert.NotNil(t, eventProperties[U.EP_FBCLIID])
		// test map from second option.
		assert.Nil(t, eventProperties["$qp_utm_adgroup_id"])
		assert.NotNil(t, eventProperties[U.EP_ADGROUP_ID])
		assert.Nil(t, eventProperties["$qp_utm_creativeid"])
		assert.NotNil(t, eventProperties[U.EP_CREATIVE])
	})

	t.Run("AddInitialUserPropertiesFromEventProperties", func(t *testing.T) {
		rEventName := "https://example.com/" + U.RandomLowerAphaNumString(10)
		w := ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"event_name": "%s", "event_properties": {"mobile": "true", "$page_url": "https://example.com/xyz/", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$referrer_domain": "gartner.com", "$referrer_url": "https://gartner.com/product_of_the_month/", "$referrer": "https://gartner.com/product_of_the_month/", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$os": "Mac OS"}}`, rEventName)),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.NotNil(t, responseMap["user_id"])
		eventUserId := responseMap["user_id"].(string)
		rUser, errCode := M.GetUser(project.ID, eventUserId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err := rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties)
		// other user properties should exist after adding initial.
		assert.NotNil(t, userProperties["$os"])
		assert.NotNil(t, userProperties[U.UP_JOIN_TIME])
		// initial user properties.
		assert.Equal(t, responseMap["event_id"].(string), userProperties[U.UP_INITIAL_PAGE_EVENT_ID])
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_URL])
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_RAW_URL])
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_DOMAIN])
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_LOAD_TIME])
		assert.Equal(t, float64(100), userProperties[U.UP_INITIAL_PAGE_LOAD_TIME])
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_SPENT_TIME])
		assert.NotNil(t, userProperties[U.UP_INITIAL_CAMPAIGN])
		assert.NotNil(t, userProperties[U.UP_INITIAL_CAMPAIGN_ID])
		assert.NotNil(t, userProperties[U.UP_INITIAL_SOURCE])
		assert.NotNil(t, userProperties[U.UP_INITIAL_MEDIUM])
		assert.NotNil(t, userProperties[U.UP_INITIAL_KEYWORD])
		assert.NotNil(t, userProperties[U.UP_INITIAL_KEYWORD_MATCH_TYPE])
		assert.NotNil(t, userProperties[U.UP_INITIAL_CONTENT])
		assert.NotNil(t, userProperties[U.UP_INITIAL_ADGROUP])
		assert.NotNil(t, userProperties[U.UP_INITIAL_ADGROUP_ID])
		assert.NotNil(t, userProperties[U.UP_INITIAL_CREATIVE])
		assert.NotNil(t, userProperties[U.UP_INITIAL_GCLID])
		assert.NotNil(t, userProperties[U.UP_INITIAL_FBCLID])
		assert.NotNil(t, userProperties[U.UP_INITIAL_REFERRER])
		assert.Equal(t, "https://gartner.com/product_of_the_month", userProperties[U.UP_INITIAL_REFERRER])
		assert.NotNil(t, userProperties[U.UP_INITIAL_REFERRER_URL])
		assert.Equal(t, "https://gartner.com/product_of_the_month", userProperties[U.UP_INITIAL_REFERRER_URL])
		assert.NotNil(t, userProperties[U.UP_INITIAL_REFERRER_DOMAIN])
		assert.Equal(t, "gartner.com", userProperties[U.UP_INITIAL_REFERRER_DOMAIN])
		assert.Nil(t, userProperties[U.UP_INITIAL_COST])
		assert.Nil(t, userProperties[U.UP_INITIAL_REVENUE])
		assert.NotNil(t, userProperties[U.UP_DAY_OF_FIRST_EVENT])
		assert.Equal(t, time.Unix(rEvent.Timestamp, 0).Weekday().String(), userProperties[U.UP_DAY_OF_FIRST_EVENT])
		retUserFirstVisitHour, _, _ := time.Unix(rEvent.Timestamp, 0).Clock()
		assert.NotNil(t, userProperties[U.UP_HOUR_OF_FIRST_EVENT])
		assert.Equal(t, float64(retUserFirstVisitHour), userProperties[U.UP_HOUR_OF_FIRST_EVENT])

		// Initial user properties should not get updated on existing user's track call.
		rEventName = "https://example.com/" + U.RandomLowerAphaNumString(10)
		w = ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {"$qp_utm_campaign": "producthunt", "$qp_utm_campaignid": "78910"}, "user_properties": {"$os": "Mac OS"}}`,
				rEventName, eventUserId)), map[string]string{"Authorization": project.Token}) // user from prev track used.
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.Nil(t, responseMap["user_id"]) // no new user.
		rUser, errCode = M.GetUser(project.ID, eventUserId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err = rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties2 map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties2)
		assert.NotNil(t, userProperties2["$os"])
		assert.NotNil(t, userProperties[U.UP_JOIN_TIME])
		// initial user properties.
		assert.NotNil(t, userProperties2[U.UP_INITIAL_CAMPAIGN])
		assert.Equal(t, "google", userProperties2[U.UP_INITIAL_CAMPAIGN])
		assert.NotEqual(t, "producthunt", userProperties2[U.UP_INITIAL_CAMPAIGN])
		assert.NotNil(t, userProperties2[U.UP_INITIAL_CAMPAIGN_ID])
		assert.Equal(t, "12345", userProperties2[U.UP_INITIAL_CAMPAIGN_ID])
		assert.NotEqual(t, "78910", userProperties2[U.UP_INITIAL_CAMPAIGN_ID])

		// Should set default values for properties.
		rEventName = "example.com/" + U.RandomLowerAphaNumString(10)
		w = ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"event_name": "%s", "auto": true, "event_properties": {"$page_raw_url": "%s", "$qp_utm_campaign": "producthunt", "$qp_utm_campaignid": "78910"}, "user_properties": {"$os": "Mac OS"}}`,
				rEventName, rEventName)),
			map[string]string{"Authorization": project.Token}) // user from prev track used.
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.NotNil(t, responseMap["user_id"]) // no new user.
		eventUserId = responseMap["user_id"].(string)
		rUser, errCode = M.GetUser(project.ID, eventUserId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err = rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties3 map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties3)
		assert.NotNil(t, userProperties3["$os"])
		assert.NotNil(t, userProperties3[U.UP_JOIN_TIME])
		// initial user properties.
		assert.Equal(t, rEventName, userProperties3[U.UP_INITIAL_PAGE_RAW_URL])
		assert.Equal(t, float64(1), userProperties3[U.UP_INITIAL_PAGE_SPENT_TIME])
		assert.Equal(t, float64(1), userProperties3[U.UP_INITIAL_PAGE_LOAD_TIME])
		assert.Equal(t, float64(0), userProperties3[U.UP_INITIAL_PAGE_SCROLL_PERCENT])
	})

	t.Run("AddLatestTouchUserPropertiesFromEventPropertiesIfHasMarketingProperties", func(t *testing.T) {
		// New user.
		rEventName := "https://example.com/" + U.RandomLowerAphaNumString(10)
		w := ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"event_name": "%s", "event_properties": {"mobile": "true", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$referrer_domain": "gartner.com", "$referrer_url": "https://gartner.com/product_of_the_month?ref=google", "$referrer": "https://gartner.com/product_of_the_month", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$os": "Mac OS"}}`, rEventName)),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.NotNil(t, responseMap["user_id"])
		eventUserId := responseMap["user_id"].(string)
		rUser, errCode := M.GetUser(project.ID, eventUserId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err = rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties1 map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties1)
		assert.NotNil(t, userProperties1[U.UP_LATEST_CAMPAIGN])
		assert.Equal(t, "google", userProperties1[U.UP_LATEST_CAMPAIGN])
		assert.NotEqual(t, "producthunt", userProperties1[U.UP_LATEST_CAMPAIGN])
		assert.NotNil(t, userProperties1[U.UP_LATEST_CAMPAIGN_ID])
		assert.Equal(t, "12345", userProperties1[U.UP_LATEST_CAMPAIGN_ID])

		// Existing user.
		w = ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "event_properties": {"mobile": "true", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=facebook", "$page_domain": "example.com", "$referrer_domain": "gartner.com", "$referrer_url": "https://gartner.com/product_of_the_month?ref=google", "$referrer": "https://gartner.com/product_of_the_month", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "facebook", "$qp_utm_campaignid": "7890", "$qp_utm_source": "facebook", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$os": "Mac OS"}}`,
				eventUserId, rEventName)), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.Nil(t, responseMap["user_id"])
		// latest user properties should have the new campaign values.
		rUser, errCode = M.GetUser(project.ID, eventUserId)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err = rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties2 map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties2)
		assert.NotNil(t, userProperties2[U.UP_LATEST_CAMPAIGN])
		assert.Equal(t, "facebook", userProperties2[U.UP_LATEST_CAMPAIGN])
		assert.NotNil(t, userProperties2[U.UP_LATEST_CAMPAIGN_ID])
		assert.Equal(t, "7890", userProperties2[U.UP_LATEST_CAMPAIGN_ID])
	})

	t.Run("IgnoreFilterPropertyAtTheEndOnmatch", func(t *testing.T) {
		expr := "example.com/profile/id"
		name := "seen_reports"
		filterEventName, errCode := M.CreateOrGetFilterEventName(&M.EventName{
			ProjectId:  project.ID,
			FilterExpr: expr,
			Name:       name,
		})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, filterEventName)
		assert.NotZero(t, filterEventName.ID)
		assert.Equal(t, name, filterEventName.Name)
		assert.Equal(t, expr, filterEventName.FilterExpr)
		assert.Equal(t, M.TYPE_FILTER_EVENT_NAME, filterEventName.Type)

		rEventName = "example.com/profile/id/1"
		w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(
			`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$os_version": "1_2_3"}}`,
			user.ID, rEventName)), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.Nil(t, responseMap["user_id"])
		rEvent, errCode = M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rEvent)
		assert.Equal(t, filterEventName.ID, rEvent.EventNameId)
	})

	t.Run("InitialUserPropertiesAfterUserCreation", func(t *testing.T) {
		project, user, err := SetupProjectUserReturnDAO()
		assert.Nil(t, err)

		rEventName := "https://example.com/xyz"
		w := ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {"mobile": "true", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$os": "Mac OS"}}`,
				rEventName, user.ID)), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.Nil(t, responseMap["user_id"])
		rUser, errCode := M.GetUser(project.ID, user.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err := rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties)
		// Other user properties should also be there.
		assert.NotNil(t, userProperties[U.UP_OS])
		// Initial user properties should be there.
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_URL])
		assert.NotNil(t, userProperties[U.UP_INITIAL_PAGE_RAW_URL])
		assert.Equal(t, "https://example.com/xyz", userProperties[U.UP_INITIAL_PAGE_URL])
		assert.Equal(t, "https://example.com/xyz?utm_campaign=google", userProperties[U.UP_INITIAL_PAGE_RAW_URL])
		assert.Equal(t, responseMap["event_id"].(string), userProperties[U.UP_INITIAL_PAGE_EVENT_ID])
		initialEventID := responseMap["event_id"].(string)

		// initial properties should not be overwritten
		// on consecutive track calls.
		rEventName = "https://domain.com/xyz"
		w = ServePostRequestWithHeaders(r, uri,
			[]byte(fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {"$page_url": "https://domain.com/xyz", "$page_raw_url": "https://domain.com/xyz?utm_campaign=pd", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$os": "Mac OS"}}`,
				rEventName, user.ID)), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		responseMap = DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)
		assert.NotNil(t, responseMap["event_id"])
		assert.Nil(t, responseMap["user_id"])
		rUser, errCode = M.GetUser(project.ID, user.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, rUser)
		userPropertiesBytes, err = rUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties2 map[string]interface{}
		json.Unmarshal(userPropertiesBytes.([]byte), &userProperties2)
		// Other user properties should also be there.
		assert.NotNil(t, userProperties2[U.UP_OS])
		// Initial user properties should be there.
		assert.NotNil(t, userProperties2[U.UP_INITIAL_PAGE_URL])
		assert.NotNil(t, userProperties2[U.UP_INITIAL_PAGE_RAW_URL])
		// values should not be updated with current event properties.
		assert.Equal(t, "https://example.com/xyz", userProperties2[U.UP_INITIAL_PAGE_URL])
		assert.Equal(t, "https://example.com/xyz?utm_campaign=google", userProperties2[U.UP_INITIAL_PAGE_RAW_URL])
		assert.Equal(t, initialEventID, userProperties2[U.UP_INITIAL_PAGE_EVENT_ID])
	})
}

func TestSDKTrackWithExternalEventIdUserIdAndTimestamp(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	t.Run("WithUserIdAndCreateUserAsTrue", func(t *testing.T) {
		eventId := U.GetUUID()
		userId := U.GetUUID()
		timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
		randomeEventName := U.RandomLowerAphaNumString(10)
		trackPayload := SDK.TrackPayload{
			EventId:    eventId,
			UserId:     userId,
			CreateUser: true,
			Name:       randomeEventName,
			Timestamp:  timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		// Event should be created with the given event_id.
		assert.Equal(t, eventId, response.EventId)
		// User should be created with the given user id, as create_user is set.
		assert.Equal(t, userId, response.UserId)
		event, _ := M.GetEventById(project.ID, response.EventId)
		assert.NotNil(t, event)
		// Event timestamp should be externaly given timestamp.
		assert.Equal(t, timestamp, event.Timestamp)
		user, _ := M.GetUser(project.ID, response.UserId)
		assert.NotNil(t, user)
		// Event should be associated with created user.
		assert.Equal(t, user.ID, event.UserId)
		// User join timestamp should be event timestamp, as create_user is set.
		assert.Equal(t, timestamp, user.JoinTimestamp)
	})

	t.Run("WithUserIdAndCreateUserAsFalse", func(t *testing.T) {
		eventId := U.GetUUID()
		timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
		randomeEventName := U.RandomLowerAphaNumString(10)
		trackPayload := SDK.TrackPayload{
			EventId:    eventId,
			UserId:     user.ID,
			CreateUser: false,
			Name:       randomeEventName,
			Timestamp:  timestamp,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
		assert.Equal(t, http.StatusOK, status)
		// Event should be created with the given event_id.
		assert.Equal(t, eventId, response.EventId)
		// User should not be created with the given user id, as create_user is false.
		assert.Empty(t, response.UserId)
		event, _ := M.GetEventById(project.ID, response.EventId)
		assert.NotNil(t, event)
		// Event should be associated with the given existing user.
		assert.Equal(t, user.ID, event.UserId)
	})

}

func TestSDKWithQueue(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	t.Run("TrackWithoutUserId", func(t *testing.T) {
		randomeEventName := U.RandomLowerAphaNumString(10)
		payload := SDK.TrackPayload{
			Name: randomeEventName,
		}
		status, response := SDK.TrackWithQueue(project.Token,
			&payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
		// Should respond event id.
		assert.NotEmpty(t, response.EventId)
		// Should respond user id as user id is not given on request.
		assert.NotEmpty(t, response.UserId)
	})

	t.Run("TrackWithUserId", func(t *testing.T) {
		randomeEventName := U.RandomLowerAphaNumString(10)
		payload := SDK.TrackPayload{
			Name:   randomeEventName,
			UserId: user.ID,
		}
		status, response := SDK.TrackWithQueue(project.Token,
			&payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
		// Should respond event id.
		assert.NotEmpty(t, response.EventId)
		// Should respond user_id as it is given.
		assert.Empty(t, response.UserId)
	})

	t.Run("IdentifyWithoutUserId", func(t *testing.T) {
		randomeUserId := U.RandomLowerAphaNumString(10)
		payload := SDK.IdentifyPayload{
			CustomerUserId: randomeUserId,
		}
		status, response := SDK.IdentifyWithQueue(project.Token,
			&payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
		// Should respond user id as user id is not given on request.
		assert.NotEmpty(t, response.UserId)
	})

	t.Run("IdentifyWithUserId", func(t *testing.T) {
		randomeUserId := U.RandomLowerAphaNumString(10)
		payload := SDK.IdentifyPayload{
			UserId:         U.GetUUID(),
			CustomerUserId: randomeUserId,
		}
		status, response := SDK.IdentifyWithQueue(project.Token,
			&payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
		// Should not respond user id as user id given on request.
		assert.Empty(t, response.UserId)
	})

	t.Run("AddUserPropertiesWithoutUserId", func(t *testing.T) {
		payload := SDK.AddUserPropertiesPayload{
			Properties: U.PropertiesMap{},
		}
		status, response := SDK.AddUserPropertiesWithQueue(project.Token,
			&payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
		// Should respond user id as user id is not given on request.
		assert.NotEmpty(t, response.UserId)
	})

	t.Run("AddUserPropertiesWithUserId", func(t *testing.T) {
		payload := SDK.AddUserPropertiesPayload{
			UserId:     U.GetUUID(),
			Properties: U.PropertiesMap{},
		}
		status, response := SDK.AddUserPropertiesWithQueue(project.Token,
			&payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
		// Should not respond user id as user id given on request.
		assert.Empty(t, response.UserId)
	})

	// Update event
	t.Run("UpdateEventProperties", func(t *testing.T) {
		payload := SDK.UpdateEventPropertiesPayload{
			Properties: U.PropertiesMap{},
		}
		status, _ := SDK.UpdateEventPropertiesWithQueue(
			project.Token, &payload, []string{project.Token})
		assert.Equal(t, http.StatusOK, status)
	})
}

func TestSDKIdentifyWithExternalUserAndTimestamp(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	userID := U.GetUUID()
	customerUserID := U.RandomLowerAphaNumString(10)
	t.Run("WithUserIdAndCreateUserAsTrue", func(t *testing.T) {
		timestamp := U.UnixTimeBeforeDuration(2 * time.Hour)
		payload := &SDK.IdentifyPayload{
			UserId:         userID,
			CreateUser:     true,
			CustomerUserId: customerUserID,
			JoinTimestamp:  timestamp,
		}
		status, response := SDK.Identify(project.ID, payload)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, userID, response.UserId)
		user, _ := M.GetUser(project.ID, response.UserId)
		assert.NotNil(t, user)
		assert.Equal(t, customerUserID, user.CustomerUserId)
		assert.Equal(t, timestamp, user.JoinTimestamp)
	})

	// Should always create a new user even when customer_user_id
	// already exists but create_user is set to true.
	userID2 := U.GetUUID()
	t.Run("WithUserIDCreateUserAsTrueAndExistingCustomerUserID", func(t *testing.T) {
		timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
		payload := &SDK.IdentifyPayload{
			UserId:         userID2,
			CreateUser:     true,
			CustomerUserId: customerUserID,
			JoinTimestamp:  timestamp,
		}
		status, response := SDK.Identify(project.ID, payload)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, userID2, response.UserId)
		user, _ := M.GetUser(project.ID, response.UserId)
		assert.NotNil(t, user)
		assert.Equal(t, customerUserID, user.CustomerUserId)
		// Should be equal to request Join timestamp.
		assert.Equal(t, timestamp, user.JoinTimestamp)
	})

	t.Run("WithUserIdAndCreateUserAsFalse", func(t *testing.T) {
		customerUserId := U.RandomLowerAphaNumString(10)
		payload := &SDK.IdentifyPayload{
			UserId:         user.ID,
			CreateUser:     false,
			CustomerUserId: customerUserId,
		}
		status, response := SDK.Identify(project.ID, payload)
		assert.Equal(t, http.StatusOK, status)
		// Should use the existing user.
		assert.Empty(t, response.UserId)
		user, _ := M.GetUser(project.ID, user.ID)
		assert.NotNil(t, user)
		assert.Equal(t, customerUserId, user.CustomerUserId)
		assert.NotEmpty(t, user.JoinTimestamp)
	})
}

func TestSDKAddUserPropertiesWithExternalUserIdAndTimestamp(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	t.Run("WithUserIdAndCreateUserAsTrue", func(t *testing.T) {
		userId := U.GetUUID()
		timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
		payload := &SDK.AddUserPropertiesPayload{
			UserId:     userId,
			Timestamp:  timestamp,
			CreateUser: true,
			Properties: U.PropertiesMap{
				"key": "value1",
			},
		}
		status, response := SDK.AddUserProperties(project.ID, payload)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, userId, response.UserId)
		user, _ := M.GetUser(project.ID, response.UserId)
		assert.NotNil(t, user)
		assert.Equal(t, timestamp, user.JoinTimestamp)
		properties, err := U.DecodePostgresJsonb(&user.Properties)
		assert.NotNil(t, properties)
		assert.Nil(t, err)
		assert.Equal(t, "value1", (*properties)["key"])
	})

	t.Run("WithUserIdAndCreateUserAsFalse", func(t *testing.T) {
		payload := &SDK.AddUserPropertiesPayload{
			UserId:     user.ID,
			CreateUser: false,
			Properties: U.PropertiesMap{
				"key": "value1",
			},
			Timestamp: time.Now().Unix(),
		}
		status, response := SDK.AddUserProperties(project.ID, payload)
		assert.Equal(t, http.StatusOK, status)
		// Should use the existing user given.
		assert.Empty(t, response.UserId)
		user, _ := M.GetUser(project.ID, user.ID)
		assert.NotNil(t, user)
		assert.NotEmpty(t, user.JoinTimestamp)
		properties, err := U.DecodePostgresJsonb(&user.Properties)
		assert.NotNil(t, properties)
		assert.Nil(t, err)
		assert.Equal(t, "value1", (*properties)["key"])
	})
}

func TestTrackHandlerWithUserSession(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	eventName := U.RandomLowerAphaNumString(10)
	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "event_properties": {"$referrer": "https://example.com/abc?ref=1", "$referrer_url": "https://example.com/abc", "$referrer_domain": "example.com", "$page_url": "https://example.com/xyz", "$page_raw_url": "https://example.com/xyz?utm_campaign=google", "$page_domain": "example.com", "$page_load_time": 100, "$page_spent_time": 120, "$qp_utm_campaign": "google", "$qp_utm_campaignid": "12345", "$qp_utm_source": "google", "$qp_utm_medium": "email", "$qp_utm_keyword": "analytics", "$qp_utm_matchtype": "exact", "$qp_utm_content": "analytics", "$qp_utm_adgroup": "ad-xxx", "$qp_utm_adgroupid": "xyz123", "$qp_utm_creative": "creative-xxx", "$qp_gclid": "xxx123", "$qp_fbclid": "zzz123"}, "user_properties": {"$platform": "web", "$browser": "Mozilla", "$browser_version": "v0.1", "$browser_with_version": "Mozilla_v0.1", "$user_agent": "browser", "$os": "Linux", "$os_version": "v0.1", "$os_with_version": "Linux_v0.1", "$country": "india", "$region": "karnataka", "$city": "bengaluru", "$timezone": "Asia/Calcutta"}}`,
			eventName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	assert.NotNil(t, responseMap["user_id"])
	responseEventId := responseMap["event_id"].(string)
	responseUserId := responseMap["user_id"].(string)
	sessionEventName, errCode := M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, sessionEventName)
	userSessionEvents, errCode := M.GetUserEventsByEventNameId(project.ID,
		responseMap["user_id"].(string), sessionEventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.True(t, len(userSessionEvents) == 1)
	sessionPropertiesBytes, err := userSessionEvents[0].Properties.Value()
	assert.Nil(t, err)
	var sessionProperties map[string]interface{}
	json.Unmarshal(sessionPropertiesBytes.([]byte), &sessionProperties)
	assert.NotEmpty(t, sessionProperties[U.SP_IS_FIRST_SESSION])
	assert.True(t, sessionProperties[U.SP_IS_FIRST_SESSION].(bool))
	// session properties from user properties.
	assert.NotEmpty(t, sessionProperties[U.UP_BROWSER])
	assert.NotEmpty(t, sessionProperties[U.UP_BROWSER_VERSION])
	assert.NotEmpty(t, sessionProperties[U.UP_BROWSER_WITH_VERSION])
	assert.NotEmpty(t, sessionProperties[U.UP_USER_AGENT])
	assert.NotEmpty(t, sessionProperties[U.UP_OS])
	assert.NotEmpty(t, sessionProperties[U.UP_OS_VERSION])
	assert.NotEmpty(t, sessionProperties[U.UP_OS_WITH_VERSION])
	assert.NotEmpty(t, sessionProperties[U.UP_COUNTRY])
	assert.NotEmpty(t, sessionProperties[U.UP_CITY])
	assert.NotEmpty(t, sessionProperties[U.UP_REGION])
	assert.NotEmpty(t, sessionProperties[U.UP_TIMEZONE])
	// session properties from event properties.
	assert.NotEmpty(t, sessionProperties[U.UP_INITIAL_PAGE_URL])
	assert.NotEmpty(t, sessionProperties[U.UP_INITIAL_PAGE_RAW_URL])
	assert.NotEmpty(t, sessionProperties[U.UP_INITIAL_PAGE_DOMAIN])
	assert.NotEmpty(t, sessionProperties[U.SP_INITIAL_REFERRER])
	assert.NotEmpty(t, sessionProperties[U.SP_INITIAL_REFERRER_URL])
	assert.NotEmpty(t, sessionProperties[U.SP_INITIAL_REFERRER_DOMAIN])
	assert.NotEmpty(t, sessionProperties[U.UP_INITIAL_PAGE_LOAD_TIME])
	assert.Equal(t, float64(100), sessionProperties[U.UP_INITIAL_PAGE_LOAD_TIME])
	assert.NotEmpty(t, sessionProperties[U.UP_INITIAL_PAGE_SPENT_TIME])
	assert.NotEmpty(t, sessionProperties[U.EP_CAMPAIGN])
	assert.NotEmpty(t, sessionProperties[U.EP_CAMPAIGN_ID])
	assert.NotEmpty(t, sessionProperties[U.EP_SOURCE])
	assert.NotEmpty(t, sessionProperties[U.EP_MEDIUM])
	assert.NotEmpty(t, sessionProperties[U.EP_KEYWORD])
	assert.NotEmpty(t, sessionProperties[U.EP_KEYWORD_MATCH_TYPE])
	assert.NotEmpty(t, sessionProperties[U.EP_CONTENT])
	assert.NotEmpty(t, sessionProperties[U.EP_ADGROUP])
	assert.NotEmpty(t, sessionProperties[U.EP_ADGROUP_ID])
	assert.NotEmpty(t, sessionProperties[U.EP_CREATIVE])
	assert.NotEmpty(t, sessionProperties[U.EP_GCLID])
	assert.NotEmpty(t, sessionProperties[U.EP_FBCLIID])
	// Tracked event should have latest session of user associated with it.
	rEvent, errCode := M.GetEvent(project.ID, responseUserId, responseEventId)
	assert.Equal(t, http.StatusFound, errCode)
	latestSessionEvent, errCode := M.GetLatestEventOfUserByEventNameId(rEvent.ProjectId,
		rEvent.UserId, sessionEventName.ID, rEvent.Timestamp-86400, rEvent.Timestamp)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent.SessionId)
	assert.NotEmpty(t, *rEvent.SessionId)
	assert.Equal(t, latestSessionEvent.ID, *rEvent.SessionId)

	eventPropertiesMap, _ := U.DecodePostgresJsonb(&rEvent.Properties)
	assert.NotNil(t, (*eventPropertiesMap)[U.EP_SESSION_COUNT])
	assert.Equal(t, (*eventPropertiesMap)[U.EP_SESSION_COUNT], float64(latestSessionEvent.Count))

	// session with existing user and active.
	eventName = U.RandomLowerAphaNumString(10)
	// using user created on prev request.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {}}`, eventName, responseUserId)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	assert.Nil(t, responseMap["user_id"])
	responseEventId2 := responseMap["event_id"].(string)
	sessionEventName, errCode = M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, sessionEventName)
	userSessionEvents2, errCode := M.GetUserEventsByEventNameId(project.ID, responseUserId,
		sessionEventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	// should not create new session.
	assert.True(t, len(userSessionEvents2) == 1)
	assert.Equal(t, userSessionEvents[0].ID, userSessionEvents2[0].ID)
	// Tracked event should have latest session of active user associated with it.
	rEvent2, errCode := M.GetEvent(project.ID, responseUserId, responseEventId2)
	eventPropertiesMap, _ = U.DecodePostgresJsonb(&rEvent2.Properties)
	assert.NotNil(t, (*eventPropertiesMap)[U.EP_SESSION_COUNT])
	assert.Equal(t, (*eventPropertiesMap)[U.EP_SESSION_COUNT], float64(latestSessionEvent.Count))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent2.SessionId)
	assert.NotEmpty(t, *rEvent2.SessionId)
	assert.Equal(t, latestSessionEvent.ID, *rEvent2.SessionId)
}

func TestTrackHandlerUserSessionWithTimestamp(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	timestampBeforeOneDay := U.UnixTimeBeforeDuration(time.Hour * 24)
	user, errCode := M.CreateUser(&M.User{ProjectId: project.ID,
		JoinTimestamp: timestampBeforeOneDay})
	assert.Equal(t, http.StatusCreated, errCode)

	// New session has to created.
	payload := fmt.Sprintf(`{"user_id": "%s", "timestamp": %d, "event_name": "event_1", "event_properties": {}, "user_properties": {"$os": "Mac OS"}}`,
		user.ID, timestampBeforeOneDay)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	event1, errCode := M.GetEventById(project.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, event1.SessionId)

	// Existing session has to be used.
	lastEventTimestamp := timestampBeforeOneDay + 10
	payload = fmt.Sprintf(`{"user_id": "%s", "timestamp": %d, "event_name": "event_1", "event_properties": {}, "user_properties": {"$os": "Mac OS"}}`,
		user.ID, lastEventTimestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	event2, errCode := M.GetEventById(project.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, event2.SessionId)
	assert.Equal(t, event1.SessionId, event2.SessionId)
	// No of sessions should be 1.
	sessionEventName, errCode := M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, sessionEventName)
	sessionEvents, errCode := M.GetUserEventsByEventNameId(project.ID, user.ID,
		sessionEventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Len(t, sessionEvents, 1)

	// New session has to be created by even timestamp
	// as user was inactive.
	lastEventTimestamp = lastEventTimestamp + 1800
	payload = fmt.Sprintf(`{"user_id": "%s", "timestamp": %d, "event_name": "event_1", "event_properties": {}, "user_properties": {"$os": "Mac OS"}}`,
		user.ID, lastEventTimestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	event3, errCode := M.GetEventById(project.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, event3.SessionId)
	assert.NotEqual(t, event2.SessionId, event3.SessionId) // new session.
	// No of sessions should be 2.
	sessionEventName, errCode = M.GetEventName(U.EVENT_NAME_SESSION, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, sessionEventName)
	sessionEvents, errCode = M.GetUserEventsByEventNameId(project.ID, user.ID,
		sessionEventName.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Len(t, sessionEvents, 2)

}

func TestBlockSDKRequestByToken(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Adds project token to blocked list.
	C.GetConfig().BlockedSDKRequestProjectTokens = []string{project.Token}

	// Should block sdk track request.
	w := ServePostRequestWithHeaders(r, uri, []byte(`{"event_name": "signup", "event_properties": {"mobile" : "true"}}`),
		map[string]string{"Authorization": project.Token})
	// StatusOK intentional to avoid changing customer app behaviour.
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["event_id"])
	assert.Equal(t, "Request failed. Blocked.", responseMap["error"])
}

func TestOldUserSessionProperties(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	timestampBeforeOneDay := U.UnixTimeBeforeDuration(time.Hour * 24)
	user, errCode := M.CreateUser(&M.User{ProjectId: project.ID,
		JoinTimestamp: timestampBeforeOneDay})
	assert.Equal(t, http.StatusCreated, errCode)

	// Checking for old users
	payload := fmt.Sprintf(`{"user_id": "%s", "timestamp": %d, "event_name": "event_1", "event_properties": {}, "user_properties": {"$os": "Mac OS"}}`,
		user.ID, timestampBeforeOneDay)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	eventOldUser, errCode := M.GetEventById(project.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, eventOldUser.SessionId)
	oldUserPropertiesMap, errCode := M.GetLatestUserPropertiesOfUserAsMap(project.ID, user.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.Nil(t, (*oldUserPropertiesMap)[U.UP_PAGE_COUNT])
	assert.Nil(t, (*oldUserPropertiesMap)[U.UP_TOTAL_SPENT_TIME])
	assert.Equal(t, float64(1), (*oldUserPropertiesMap)[U.UP_SESSION_COUNT])
}

func TestTrackHandlerWithFormSubmit(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// track form submit event.
	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "event_properties": {"$email": "xxx@example.com", "$company": "Example Inc"}}`,
			U.EVENT_NAME_FORM_SUBMITTED)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	assert.NotNil(t, responseMap["user_id"])
	userId := responseMap["user_id"].(string)
	userProperties := postgres.Jsonb{json.RawMessage(`{"plan": "enterprise"}`)}
	_, errCode := M.UpdateUserProperties(project.ID, userId, &userProperties, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)
	// form submit event name created.
	formSubmitEventName, errCode := M.GetEventName(U.EVENT_NAME_FORM_SUBMITTED, project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, formSubmitEventName)
	// form submit event properties added as user properties.
	user, errCode := M.GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, user)
	userPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, "xxx@example.com", (*userPropertiesMap)[U.UP_EMAIL])
	assert.Equal(t, "Example Inc", (*userPropertiesMap)[U.UP_COMPANY])
	assert.Equal(t, "enterprise", (*userPropertiesMap)["plan"]) // other properties should not be affected.
	// identify with form submitted email.
	assert.Equal(t, "xxx@example.com", user.CustomerUserId)
}

func TestSDKIdentifyHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/user/identify"

	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// Note: Should tolerate anything other than improper c_uid.

	// Test without project_id scope and with non-existing project.
	w := ServePostRequest(r, uri, []byte("{}"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test with invalid token.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`),
		map[string]string{"Authorization": "INVALID_TOKEN"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Email as customer_user_id.
	rCustomerUserId := fmt.Sprintf("%s@example.com", U.RandomLowerAphaNumString(5))

	// Test without user_id in the payload.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s"}`, rCustomerUserId)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["user_id"])
	assert.NotEmpty(t, responseMap["user_id"].(string))
	retUser, errCode := M.GetUser(project.ID, responseMap["user_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, retUser)
	userProperties, err := U.DecodePostgresJsonb(&retUser.Properties)
	assert.Nil(t, err)
	assert.NotNil(t, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, rCustomerUserId, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, rCustomerUserId, (*userProperties)[U.UP_EMAIL])

	rUserId := U.RandomLowerAphaNumString(15)
	// Test without c_uid in the payload and with non-existing c_uid.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s"}`, rUserId)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test c_uid and user_id not present.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test identify with all conditions satisfied.
	r1CustomerUserId := U.RandomLowerAphaNumString(15)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		r1CustomerUserId, user.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Test re-identify an identified user with same customer_user_id and user_id
	// responds saying identified already.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		r1CustomerUserId, user.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Test re-identify an identified user only with same customer_user_id
	// should respond with latest user_id for the customer_user. reusing.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s"}`,
		r1CustomerUserId)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["user_id"])
	assert.Equal(t, user.ID, responseMap["user_id"])
	retUser, errCode = M.GetUser(project.ID, responseMap["user_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, retUser)
	userProperties, err = U.DecodePostgresJsonb(&retUser.Properties)
	assert.Nil(t, err)
	assert.NotNil(t, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, r1CustomerUserId, (*userProperties)[U.UP_USER_ID])

	// Test re-identify an identified user with different customer_user
	// should respond with new user_id mapped to customer_user_id
	r2CustomerUserId := U.RandomLowerAphaNumString(15)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		r2CustomerUserId, user.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["user_id"])
	assert.NotEmpty(t, responseMap["user_id"].(string))
	assert.NotEqual(t, responseMap["user_id"], user.ID)
	retUser, errCode = M.GetUser(project.ID, responseMap["user_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, retUser)
	userProperties, err = U.DecodePostgresJsonb(&retUser.Properties)
	assert.Nil(t, err)
	assert.NotNil(t, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, r2CustomerUserId, (*userProperties)[U.UP_USER_ID])
}

func assertEqualJoinTimePropertyOnAllRecords(t *testing.T, records []M.UserProperties, expectedJoinTime int64) {
	for _, userProperties := range records {
		var propertiesMap map[string]interface{}
		err := json.Unmarshal(userProperties.Properties.RawMessage, &propertiesMap)
		assert.Nil(t, err)

		assert.Contains(t, propertiesMap, U.UP_JOIN_TIME)
		assert.Equal(t, float64(expectedJoinTime), propertiesMap[U.UP_JOIN_TIME])
	}
}

func TestUpdateJoinTimeOnSDKIdentify(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/user/identify"

	project, user1, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)

	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)

	// identify all users with same c_uid.
	customerUserId := U.RandomLowerAphaNumString(15)
	w := ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user1.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user2.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// all user properties of all users with same c_uid should have joinTime as min of joinTime
	// among users.
	userPropertiesRecords, errCode := M.GetUserPropertyRecordsByUserId(project.ID, user1.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assertEqualJoinTimePropertyOnAllRecords(t, userPropertiesRecords, user1.JoinTimestamp)
	userPropertiesRecords, errCode = M.GetUserPropertyRecordsByUserId(project.ID, user2.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assertEqualJoinTimePropertyOnAllRecords(t, userPropertiesRecords, user1.JoinTimestamp)

	// identify with same customer user id after new user properties addition,
	// should update join time on new user_properties record also.
	addPropertiesURI := "/sdk/user/add_properties"
	uniqueName := U.RandomLowerAphaNumString(16)
	uniqueEmail := fmt.Sprintf(`%s@example.com`, U.RandomLowerAphaNumString(10))
	w = ServePostRequestWithHeaders(r, addPropertiesURI, []byte(fmt.Sprintf(
		`{"user_id": "%s", "properties": {"name": "%s", "email": "%s"}}`, user3.ID, uniqueName, uniqueEmail)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user3.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	userPropertiesRecords, errCode = M.GetUserPropertyRecordsByUserId(project.ID, user3.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assertEqualJoinTimePropertyOnAllRecords(t, userPropertiesRecords, user1.JoinTimestamp)
}

func TestSDKAddUserPropertiesHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/user/add_properties"

	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	// Test without project_id scope and with non-existing project.
	w := ServePostRequest(r, uri, []byte("{}"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test with invalid token.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`), map[string]string{"Authorization": "INVALID_TOKEN"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test with user_id in the payload.
	uniqueName := U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"name": "%s"}}`,
		user.ID, uniqueName)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	// user_id should not exist on response.
	assert.Nil(t, responseMap["user_id"])

	// Test with new property. email.
	uniqueEmail := fmt.Sprintf(`%s@example.com`, U.RandomLowerAphaNumString(10))
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"email": "%s"}}`,
		user.ID, uniqueEmail)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	// user_id should not exist on response.
	assert.Nil(t, responseMap["user_id"])

	// Test with multiple properties. name and email.
	uniqueName = U.RandomLowerAphaNumString(16)
	uniqueEmail = fmt.Sprintf(`%s@example.com`, U.RandomLowerAphaNumString(10))
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(
		`{"user_id": "%s", "properties": {"name": "%s", "email": "%s"}}`, user.ID, uniqueName, uniqueEmail)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	// user_id should not exist on response.
	assert.Nil(t, responseMap["user_id"])

	// Test without user_id in the payload.
	uniqueName = U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"properties": {"name": "%s"}}`, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["user_id"])
	assert.NotEmpty(t, responseMap["user_id"].(string))

	// Test bad payload - updating project_id as existing user.
	uniqueName = U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "project_id": "99999999", "properties": {"name": "%s"}}`, user.ID, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test bad payload - updating project_id as new user.
	uniqueName = U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"project_id": "99999999", "properties": {"name": "%s"}}`, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Non exiting user id.
	uniqueName = U.RandomLowerAphaNumString(16)
	fakeUserId := U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s" , "properties": {"name": "%s"}}`, fakeUserId, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code) // Should create if not exist, to support queue.

	// Test default user properties.
	uniqueName = U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"prop_1": "%s"}}`, user.ID, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	retUser, errCode := M.GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesBytes, err := retUser.Properties.Value()
	assert.Nil(t, err)
	var userPropertiesMap map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
	// Expected to test this. ClientIP nil on tests. HttpRequest RemoteAddr assignment is not working.
	// assert.NotEmpty(t, userPropertiesMap[U.UP_INTERNAL_IP])
	// assert.NotNil(t, userPropertiesMap[U.UP_COUNTRY])

	// Test properties type.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"int_prop": 100, "long_prop": 10000000000, "float_prop": 10.23, "string_prop": "string_value", "boolean_prop": false, "map_prop": {"key": "value"}}}`,
		user.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	propsResponseMap := DecodeJSONResponseToMap(w.Body)
	assert.Nil(t, propsResponseMap["user_id"])
	retUser, errCode = M.GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesBytes, err = retUser.Properties.Value()
	assert.Nil(t, err)
	var userPropertiesMap2 map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap2)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, userPropertiesMap2["int_prop"])
	assert.NotNil(t, userPropertiesMap2["long_prop"])
	assert.NotNil(t, userPropertiesMap2["string_prop"])
	assert.NotNil(t, userPropertiesMap2["float_prop"])
	assert.NotNil(t, userPropertiesMap2["boolean_prop"])
	// Types not allowed.
	assert.Nil(t, userPropertiesMap2["map_prop"])

	// Should not allow $ prefixes apart from default properties.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"$dollar_key": "unknow_value", "$os": "mac osx", "$os_version": "1_2_3", "$platform": "web", "$screen_width": 10, "$screen_height": 11, "$browser": "mozilla", "$browser_version": "10_2_3"}}`,
		user.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	propsResponseMap1 := DecodeJSONResponseToMap(w.Body)
	assert.Nil(t, propsResponseMap1["user_id"])
	retUser, errCode = M.GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	userPropertiesBytes, err = retUser.Properties.Value()
	assert.Nil(t, err)
	var userPropertiesMap3 map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap3)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Nil(t, userPropertiesMap3["$dollar_key"])                                              // dollar prefix not allowed.
	assert.NotNil(t, userPropertiesMap3[fmt.Sprintf("%s$dollar_key", U.NAME_PREFIX_ESCAPE_CHAR)]) // Escaped key should exist.
	// check for default props. Hardcoded property name as request payload.
	assert.NotNil(t, userPropertiesMap3[U.UP_OS])
	assert.NotNil(t, userPropertiesMap3[U.UP_OS_VERSION])
	assert.NotNil(t, userPropertiesMap3[U.UP_BROWSER])
	assert.NotNil(t, userPropertiesMap3[U.UP_PLATFORM])
	assert.NotNil(t, userPropertiesMap3[U.UP_SCREEN_WIDTH])
	assert.NotNil(t, userPropertiesMap3[U.UP_SCREEN_HEIGHT])
	// should not be escaped.
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_OS])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_OS_VERSION])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_BROWSER])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_PLATFORM])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_SCREEN_WIDTH])
	assert.Nil(t, userPropertiesMap3[U.NAME_PREFIX_ESCAPE_CHAR+U.UP_SCREEN_HEIGHT])
}

func TestSDKGetProjectSettingsHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/project/get_settings"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test Get project settings.
	w := ServeGetRequestWithHeaders(r, uri, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotEqual(t, 0, jsonResponseMap["id"])
	assert.NotNil(t, jsonResponseMap["auto_track"])

	// Test Get project settings with random token.
	// Returns default settings.
	randomToken := U.RandomLowerAphaNumString(32)
	w = ServeGetRequestWithHeaders(r, uri, map[string]string{"Authorization": randomToken})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSDKBulk(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	uri := "/sdk/event/track/bulk"

	t.Run("Success", func(t *testing.T) {
		payload := fmt.Sprintf("[%s,%s]", `{"event_name": "signup", "event_properties": {"mobile" : "true"}}`, `{"event_name":"test", "event_properties": {"mobile" : "true"}}`)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		resp := make([]SDK.TrackResponse, 0, 0)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &resp)
		assert.Equal(t, 2, len(resp))
	})

	t.Run("DuplicateCustomerEventId", func(t *testing.T) {
		payload := fmt.Sprintf("[%s,%s,%s]",
			`{"event_name": "signup", "event_properties": {"mobile" : "true"}}`,
			`{"event_name":"test","c_event_id":"1", "event_properties": {"mobile" : "true"}}`,
			`{"event_name":"test2","c_event_id":"1", "event_properties": {"mobile" : "true"}}`)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		resp := make([]SDK.TrackResponse, 0, 0)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &resp)

		assert.Equal(t, 3, len(resp))

		assert.NotEmpty(t, resp[1].UserId)

		assert.Equal(t, "1", *resp[2].CustomerEventId)
		assert.Equal(t, "Tracking failed. Event creation failed. Invalid payload.", resp[2].Error)
	})

}

func getAutoTrackedEventIdWithPageRawURL(t *testing.T, projectAuthToken, pageRawURL string) (string, string) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "auto": true, "event_properties": {"mobile" : "true", "$page_raw_url": "%s"}}`,
			"https://example.com/", pageRawURL)), map[string]string{"Authorization": projectAuthToken})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])
	assert.NotNil(t, responseMap["user_id"])
	return responseMap["event_id"].(string), responseMap["user_id"].(string)
}

func getAutoTrackedEventIdWithUserIdAndPageRawURL(t *testing.T, projectAuthToken, userId, pageRawURL string) string {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	w := ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "auto": true, "event_properties": {"mobile" : "true", "$page_raw_url": "%s"}}`,
			"https://example.com/", userId, pageRawURL)), map[string]string{"Authorization": projectAuthToken})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["event_id"])

	return responseMap["event_id"].(string)
}

func TestSDKUpdateEventPropertiesHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/update_properties"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Test without project_id scope and with non-existing project.
	w := ServePostRequest(r, uri, []byte("{}"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test with invalid token.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`), map[string]string{"Authorization": "INVALID_TOKEN"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	rawPageUrl := "https://example.com/url"

	// Test with invalid event_id in the payload.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$page_spent_time": "%d"}}`,
		"", 1)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test with disallowed property in the payload.
	eventId, _ := getAutoTrackedEventIdWithPageRawURL(t, project.Token, rawPageUrl)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$not_allowed": "%d"}}`,
		eventId, 1)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotNil(t, responseMap["error"])

	// Test update event properties and initial user properites added
	// from update event properties.
	eventId, _ = getAutoTrackedEventIdWithPageRawURL(t, project.Token, rawPageUrl)
	event, errCode := M.GetEventById(project.ID, eventId)
	assert.NotNil(t, event)
	user, errCode := M.GetUser(project.ID, event.UserId)
	assert.NotNil(t, user)
	latestUserProperitesIDBeforeUpdateEvent := user.PropertiesId
	// Trigger update event properties
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$page_spent_time": %d}}`,
		eventId, 100)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusAccepted, w.Code)
	event, _ = M.GetEventById(project.ID, eventId)
	assert.NotNil(t, event)
	user, _ = M.GetUser(project.ID, event.UserId)
	assert.NotNil(t, user)
	latestUserProperitesIDAfterUpdateEvent := user.PropertiesId
	// Latest user_properties_id of user, before and after
	// update_event_properties should be same.
	assert.Equal(t, latestUserProperitesIDBeforeUpdateEvent,
		latestUserProperitesIDAfterUpdateEvent)
	// Event user_properties_id should be the latest
	// user_properties_id of user
	assert.Equal(t, latestUserProperitesIDAfterUpdateEvent,
		event.UserPropertiesId)
	// initial_user_properites should be added.
	userPropertiesRecord, _ := M.GetUserPropertiesRecord(project.ID,
		event.UserId, event.UserPropertiesId)
	assert.NotNil(t, userPropertiesRecord)
	userProperties, _ := U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.NotNil(t, userProperties)
	assert.Equal(t, float64(100), (*userProperties)[U.UP_INITIAL_PAGE_SPENT_TIME])
	assert.Equal(t, event.ID, (*userProperties)[U.UP_INITIAL_PAGE_EVENT_ID])
	// Creating new user_properties state for the event user.
	newUserPropertiesJson := postgres.Jsonb{json.RawMessage(`{"plan": "enterprise"}`)}
	newUserPropertiesID, _ := M.UpdateUserProperties(project.ID, event.UserId, &newUserPropertiesJson, U.TimeNowUnix())
	// Trigger update event properties again after user properties update.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$page_spent_time": %d}}`,
		eventId, 200)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusAccepted, w.Code)
	user, _ = M.GetUser(project.ID, event.UserId)
	assert.NotNil(t, user)
	latestUserProperitesIDAfterUpdateEvent2 := user.PropertiesId
	// Latest user_properties of user should be new.
	assert.Equal(t, newUserPropertiesID, latestUserProperitesIDAfterUpdateEvent2)
	// Event user_properties should be updated.
	userPropertiesRecord, _ = M.GetUserPropertiesRecord(project.ID,
		event.UserId, event.UserPropertiesId)
	assert.NotNil(t, userPropertiesRecord)
	userProperties, _ = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.NotNil(t, userProperties)
	assert.Equal(t, float64(200), (*userProperties)[U.UP_INITIAL_PAGE_SPENT_TIME])
	// Latest user_properites should also be updated.
	userPropertiesRecord, _ = M.GetUserPropertiesRecord(project.ID,
		event.UserId, latestUserProperitesIDAfterUpdateEvent2)
	assert.NotNil(t, userPropertiesRecord)
	userProperties, _ = U.DecodePostgresJsonb(&userPropertiesRecord.Properties)
	assert.NotNil(t, userProperties)
	assert.Equal(t, float64(200), (*userProperties)[U.UP_INITIAL_PAGE_SPENT_TIME])

	eventId, _ = getAutoTrackedEventIdWithPageRawURL(t, project.Token, rawPageUrl)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$page_spent_time": %d, "$page_scroll_percent": %d}}`,
		eventId, 1, 10)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusAccepted, w.Code)
	updatedEvent, errCode := M.GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, updatedEvent)
	propertiesMap, err := U.DecodePostgresJsonb(&updatedEvent.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*propertiesMap)[U.EP_PAGE_SPENT_TIME])
	assert.Equal(t, float64(10), (*propertiesMap)[U.EP_PAGE_SCROLL_PERCENT])
}

func TestSessionAndUserInitialPropertiesUpdateOnSDKUpdateEventPropertiesHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/update_properties"

	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	pageRawURL := "https://page.url.com/1"
	eventId, userId := getAutoTrackedEventIdWithPageRawURL(t, project.Token, pageRawURL)
	w := ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$page_spent_time": %d, "$page_scroll_percent": %d}}`,
		eventId, 100, 10)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusAccepted, w.Code)
	updatedEvent, errCode := M.GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, *updatedEvent.SessionId)
	// Should update initial user properties on initial call.
	user, errCode := M.GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, errCode)
	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, pageRawURL, (*userProperties)[U.UP_INITIAL_PAGE_RAW_URL])
	assert.Equal(t, float64(100), (*userProperties)[U.UP_INITIAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(10), (*userProperties)[U.UP_INITIAL_PAGE_SCROLL_PERCENT])

	// same page_raw_url with same user and session should not
	// update $initial_page_spent_time again.
	eventId = getAutoTrackedEventIdWithUserIdAndPageRawURL(t, project.Token, userId, pageRawURL)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"event_id": "%s", "properties": {"$page_spent_time": %d}}`,
		eventId, 200)), map[string]string{"Authorization": project.Token})
	updatedEvent2, errCode := M.GetEventById(project.ID, eventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEmpty(t, *updatedEvent.SessionId)
	// Should use the same session.
	assert.Equal(t, *updatedEvent.SessionId, *updatedEvent2.SessionId)
	// Should not update user properties on consequtive calls.
	user, errCode = M.GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, errCode)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, pageRawURL, (*userProperties)[U.UP_INITIAL_PAGE_RAW_URL])
	assert.NotEqual(t, float64(200), (*userProperties)[U.UP_INITIAL_PAGE_SPENT_TIME])
	assert.Equal(t, float64(100), (*userProperties)[U.UP_INITIAL_PAGE_SPENT_TIME])
}

func TestAMPTrackByTokenHandler(t *testing.T) {
	userAgentStr := "Mozilla/5.0 (Linux; Android 8.0.0; SM-G960F Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.84 Mobile Safari/537.36"
	ampClientId := "amp-1xxAGEAL-irIHu4qMW8j3A"
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	payload := &SDK.AMPTrackPayload{
		ClientID:  ampClientId,
		SourceURL: "abcd.com/",
		Title:     "Test",

		Timestamp: time.Now().Unix(), // request timestamp.
		UserAgent: userAgentStr,
		ClientIP:  "10.10.0.1",
	}
	errCode, _ := SDK.AMPTrackByToken(project.Token, payload)
	assert.Equal(t, errCode, http.StatusOK)
	user, errCode := M.CreateOrGetAMPUser(project.ID, ampClientId, payload.Timestamp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEqual(t, user.ID, "")

	payload1 := &SDK.AMPTrackPayload{
		ClientID:  ampClientId,
		SourceURL: "abcd.com/1/",
		Title:     "Test1",

		Timestamp: time.Now().Unix(), // request timestamp.
		UserAgent: userAgentStr,
		ClientIP:  "10.10.0.1",
	}
	errCode, _ = SDK.AMPTrackByToken(project.Token, payload1)
	assert.Equal(t, errCode, http.StatusOK)
	user1, errCode := M.CreateOrGetAMPUser(project.ID, ampClientId, payload1.Timestamp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEqual(t, user1.ID, "")

	payload2 := &SDK.AMPTrackPayload{
		ClientID:  ampClientId,
		SourceURL: "abcd.com/xy_z",
		Title:     "Test2",

		Timestamp: time.Now().Unix(), // request timestamp.
		UserAgent: userAgentStr,
		ClientIP:  "10.10.0.1",
	}
	errCode, _ = SDK.AMPTrackByToken(project.Token, payload2)
	assert.Equal(t, errCode, http.StatusOK)
	user2, errCode := M.CreateOrGetAMPUser(project.ID, ampClientId, payload2.Timestamp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEqual(t, user2.ID, "")

	// with query param.
	url3 := fmt.Sprintf("abcd.com/%s", U.RandomLowerAphaNumString(5))
	payload3 := &SDK.AMPTrackPayload{
		ClientID:  ampClientId,
		SourceURL: url3 + "/?a=3", // with query param.
		Title:     "Test2",

		Timestamp: time.Now().Unix(), // request timestamp.
		UserAgent: userAgentStr,
		ClientIP:  "10.10.0.1",
	}
	errCode, _ = SDK.AMPTrackByToken(project.Token, payload3)
	assert.Equal(t, errCode, http.StatusOK)
	user3, errCode := M.CreateOrGetAMPUser(project.ID, ampClientId, payload3.Timestamp)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEqual(t, user3.ID, "")
}

func TestSDKAMPTrackByToken(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := U.UnixTimeBeforeAWeek()
	clientId := U.RandomLowerAphaNumString(5)
	request := &SDK.AMPTrackPayload{
		ClientID:  clientId,
		SourceURL: "https://example.com/a/b",
		Timestamp: timestamp,
	}
	errCode, response := SDK.AMPTrackByToken(project.Token, request)
	assert.Equal(t, http.StatusOK, errCode)
	event, errCode := M.GetEventById(project.ID, response.EventId)
	assert.Equal(t, http.StatusFound, errCode)
	// AMP Tracked event should use the given timestamp.
	assert.Equal(t, timestamp, event.Timestamp)
}

func TestSDKAMPIdentifyHandler(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/amp/user/identify"
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := U.UnixTimeBeforeAWeek()
	clientID := U.RandomLowerAphaNumString(5)
	request := &SDK.AMPTrackPayload{
		ClientID:  clientID,
		SourceURL: "https://example.com/a/b",
		Timestamp: timestamp,
	}
	errCode, _ := SDK.AMPTrackByToken(project.Token, request)
	assert.Equal(t, http.StatusOK, errCode)

	cUID := "1234"
	params := fmt.Sprintf("token=%s&client_id=%s&customer_user_id=%s", project.Token, clientID, cUID)
	response := ServePostRequest(r, uri+"?"+params, []byte{})
	assert.Equal(t, http.StatusOK, response.Code)
	jsonResponse, _ := ioutil.ReadAll(response.Body)
	var jsonResponseMap map[string]interface{}
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, "User has been identified successfully.", jsonResponseMap["message"])
	user, errCode := M.CreateOrGetAMPUser(project.ID, clientID, timestamp)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, cUID, user.CustomerUserId)

	// test old timestamp for user creation
	cUID = "12345"
	clientID = U.RandomLowerAphaNumString(5)
	oldTimestamp := time.Now().AddDate(0, 0, -10).Unix()
	payload := SDK.AMPIdentifyPayload{
		CustomerUserID: cUID,
		ClientID:       clientID,
		Timestamp:      oldTimestamp,
	}
	status, message := SDK.AMPIdentifyByToken(project.Token, &payload)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "User has been identified successfully.", message.Message)
	user, errCode = M.CreateOrGetAMPUser(project.ID, clientID, timestamp)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, oldTimestamp, user.JoinTimestamp)
}

func TestAddUserPropertiesMerge(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerUserID := getRandomEmail()
	user1, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "india",
			"age": 30,
			"paid": true,
			"gender": "m",
			"$initial_campaign": "campaign1",
			"$page_count": 10,
			"$session_spent_time": 2.2}`,
		))},
	})

	user2, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "canada",
			"age": 30,
			"paid": false,
			"$initial_campaign": "campaign2",
			"$page_count": 15,
			"$session_spent_time": 4.4}`,
		))},
	})

	// Test AddUserProperties handler call.
	errCode, _ := SDK.AddUserPropertiesByToken(
		project.Token,
		&SDK.AddUserPropertiesPayload{
			UserId: user1.ID,
			Properties: U.PropertiesMap{
				"revenue": 42,
			},
		},
	)
	assert.Equal(t, http.StatusOK, errCode)
	user1DBAfterAdd, _ := M.GetUser(project.ID, user1.ID)
	user2DBAfterAdd, _ := M.GetUser(project.ID, user2.ID)
	user1DBAfterAddProperties, _ := U.DecodePostgresJsonb(&user1DBAfterAdd.Properties)
	user2DBAfterAddProperties, _ := U.DecodePostgresJsonb(&user2DBAfterAdd.Properties)
	// Merge must have got called and updated user2 as well.
	assert.Equal(t, user1DBAfterAddProperties, user2DBAfterAddProperties)
	// PropertyId must have got updated after new property add.
	assert.NotEqual(t, user1.PropertiesId, user1DBAfterAdd.PropertiesId)
	assert.NotEqual(t, user2.PropertiesId, user2DBAfterAdd.PropertiesId)
	// New property revenue must be present in user properties.
	assert.Equal(t, float64(42), (*user1DBAfterAddProperties)["revenue"])
	assert.Equal(t, float64(42), (*user2DBAfterAddProperties)["revenue"])
}

func TestIdentifyUserPropertiesMerge(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerUserID := getRandomEmail()
	user1, _ := M.CreateUser(&M.User{
		ID:             U.GetUUID(),
		ProjectId:      project.ID,
		CustomerUserId: customerUserID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "india",
			"age": 30,
			"paid": true,
			"gender": "m",
			"$initial_campaign": "campaign1",
			"$page_count": 10,
			"$session_spent_time": 2.2}`,
		))},
	})

	// Without CustomerUserID
	user2, _ := M.CreateUser(&M.User{
		ID:        U.GetUUID(),
		ProjectId: project.ID,
		Properties: postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{
			"country": "canada",
			"age": 30,
			"paid": false,
			"$initial_campaign": "campaign2",
			"$page_count": 15,
			"$session_spent_time": 4.4}`,
		))},
	})
	// Before identify, properties are different for the users.
	user1DB, _ := M.GetUser(project.ID, user1.ID)
	user2DB, _ := M.GetUser(project.ID, user2.ID)
	user1DBProperties, _ := U.DecodePostgresJsonb(&user1DB.Properties)
	user2DBProperties, _ := U.DecodePostgresJsonb(&user2DB.Properties)
	assert.NotEqual(t, user1DBProperties, user2DBProperties)

	identifyPayload := &SDK.IdentifyPayload{
		UserId:         user2.ID,
		CustomerUserId: customerUserID,
	}

	errCode, _ := SDK.IdentifyByToken(project.Token, identifyPayload)
	assert.Equal(t, http.StatusOK, errCode)
	user1DB, _ = M.GetUser(project.ID, user1.ID)
	user2DB, _ = M.GetUser(project.ID, user2.ID)
	user1DBProperties, _ = U.DecodePostgresJsonb(&user1DB.Properties)
	user2DBProperties, _ = U.DecodePostgresJsonb(&user2DB.Properties)
	// Merge must have got called and updated user2 as well.
	assert.Equal(t, user1DBProperties, user2DBProperties)
	// PropertyId must have got updated after new property add.
	assert.NotEqual(t, user1.PropertiesId, user1DB.PropertiesId)
	assert.NotEqual(t, user2.PropertiesId, user2DB.PropertiesId)

	// Should not change on retry.
	errCode, _ = SDK.IdentifyByToken(project.Token, identifyPayload)
	assert.Equal(t, http.StatusOK, errCode)
	user1DBRetry, _ := M.GetUser(project.ID, user1.ID)
	user2DBRetry, _ := M.GetUser(project.ID, user2.ID)
	user1DBRetryProperties, _ := U.DecodePostgresJsonb(&user1DBRetry.Properties)
	user2DBRetryProperties, _ := U.DecodePostgresJsonb(&user2DBRetry.Properties)
	// Merge must have got called and updated user2 as well.
	assert.Equal(t, user1DBRetryProperties, user2DBRetryProperties)
	// PropertyId must have got updated after new property add.
	assert.Equal(t, user1DB.PropertiesId, user1DBRetry.PropertiesId)
	assert.Equal(t, user2DB.PropertiesId, user2DBRetry.PropertiesId)
}

func TestSDKTrackFirstEventUserProperties(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	randomEventURL := "https://example.com/" + U.RandomLowerAphaNumString(5)
	trackPayload := SDK.TrackPayload{
		Name:      randomEventURL,
		Timestamp: timestamp,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK)
	assert.Equal(t, http.StatusOK, status)

	event, errCode := M.GetEventById(project.ID, response.EventId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.UserId)
	assert.NotEmpty(t, event.UserPropertiesId)

	user, errCode := M.GetUser(project.ID, event.UserId)
	assert.Equal(t, http.StatusFound, errCode)

	// After first event tracking, the event associated user_properites_id
	// should be same as lastest user_properties_id of the user.
	assert.Equal(t, user.PropertiesId, event.UserPropertiesId)

	// Should contain first event properties.
	userPropertiesMap, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotEmpty(t, (*userPropertiesMap)[U.UP_DAY_OF_FIRST_EVENT])
	assert.NotEmpty(t, (*userPropertiesMap)[U.UP_HOUR_OF_FIRST_EVENT])
}
