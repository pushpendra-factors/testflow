package tests

import (
	"encoding/json"
	"factors/handler"
	H "factors/handler"
	M "factors/model"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSDKTrackHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKRoutes(r)
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
	assert.Equal(t, http.StatusFound, w.Code)

	// Test auto tracked event.
	rEventName := U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true"}, "user_properties": {"$os": "Mac OS"}}`, user.ID, rEventName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	rEvent, errCode := M.GetEvent(project.ID, user.ID, responseMap["event_id"].(string))
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rEvent)
	eventPropertiesBytes, err := rEvent.Properties.Value()
	assert.Nil(t, err)
	var eventProperties map[string]interface{}
	json.Unmarshal(eventPropertiesBytes.([]byte), &eventProperties)
	assert.Nil(t, eventProperties["$dollar_property"])
	assert.NotNil(t, eventProperties[fmt.Sprintf("%s$dollar_property", U.NAME_PREFIX_ESCAPE_CHAR)]) // escaped property should exist.
	assert.NotNil(t, eventProperties["$qp_search"])                                                 // $qp should exist.
	assert.Nil(t, eventProperties[fmt.Sprintf("%s$qp_search", U.NAME_PREFIX_ESCAPE_CHAR)])          // $qp should not be escaped.
	assert.NotNil(t, eventProperties["mobile"])                                                     // no dollar properties should exist.
	assert.True(t, len(rEvent.UserPropertiesId) > 0)
	rUser, errCode := M.GetUser(rEvent.ProjectId, rEvent.UserId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, rUser)
	userPropertiesBytes, err := rUser.Properties.Value()
	assert.Nil(t, err)
	var userProperties map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userProperties)
	assert.NotNil(t, userProperties["$os"])

	// Should not allow $ prefixes apart from default properties.
	rEventName = U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "event_properties": {"mobile": "true", "$referrer": "http://google.com", "$pageRawURL": "https://factors.ai/login", "$pageTitle": "Login"}, "user_properties": {"$dollar_key": "unknow_value", "$os": "mac osx", "$osVersion": "1_2_3", "$screenWidth": 10, "$screenHeight": 11, "$browser": "mozilla", "$platform": "web", "$browserVersion": "10_2_3"}}`,
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
	assert.NotNil(t, eventProperties1["$referrer"])
	assert.NotNil(t, eventProperties1["$pageRawURL"])
	assert.NotNil(t, eventProperties1["$pageTitle"])
	assert.NotNil(t, eventProperties1["mobile"])
	assert.Nil(t, eventProperties1["_$referrer"])
	assert.Nil(t, eventProperties1["_$pageRawURL"])
	assert.Nil(t, eventProperties1["_$pageTitle"])
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
	assert.NotNil(t, userPropertiesMap3["$os"])
	assert.NotNil(t, userPropertiesMap3["$osVersion"])
	assert.NotNil(t, userPropertiesMap3["$browser"])
	assert.NotNil(t, userPropertiesMap3["$platform"])
	assert.NotNil(t, userPropertiesMap3["$screenWidth"])
	assert.NotNil(t, userPropertiesMap3["$screenHeight"])
	// should not be escaped.
	assert.Nil(t, userPropertiesMap3["_$os"])
	assert.Nil(t, userPropertiesMap3["_$osVersion"])
	assert.Nil(t, userPropertiesMap3["_$browser"])
	assert.Nil(t, userPropertiesMap3["_$platform"])
	assert.Nil(t, userPropertiesMap3["_$screenWidth"])
	assert.Nil(t, userPropertiesMap3["_$screenHeight"])

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
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$osVersion": "1_2_3"}}`,
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
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$osVersion": "1_2_3"}}`,
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
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$osVersion": "1_2_3"}}`,
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
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "auto": true, "event_properties": {"mobile": "true"}, "user_properties": {"$os": "mac osx", "$osVersion": "1_2_3"}}`,
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

}

func TestSDKIdentifyHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKRoutes(r)
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

	rCustomerUserId := U.RandomLowerAphaNumString(15)

	// Test without user_id in the payload.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"c_uid": "%s"}`, rCustomerUserId)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.NotNil(t, responseMap["user_id"])
	assert.NotEmpty(t, responseMap["user_id"].(string))

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
	H.InitSDKRoutes(r)
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
	H.InitSDKRoutes(r)
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

	// Test bad input with non exiting user id.
	uniqueName = U.RandomLowerAphaNumString(16)
	fakeUserId := U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s" , "properties": {"name": "%s"}}`, fakeUserId, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)

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
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"$dollar_key": "unknow_value", "$os": "mac osx", "$osVersion": "1_2_3", "$platform": "web", "$screenWidth": 10, "$screenHeight": 11, "$browser": "mozilla", "$browserVersion": "10_2_3"}}`,
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
	assert.NotNil(t, userPropertiesMap3["$os"])
	assert.NotNil(t, userPropertiesMap3["$osVersion"])
	assert.NotNil(t, userPropertiesMap3["$browser"])
	assert.NotNil(t, userPropertiesMap3["$platform"])
	assert.NotNil(t, userPropertiesMap3["$screenWidth"])
	assert.NotNil(t, userPropertiesMap3["$screenHeight"])
	// should not be escaped.
	assert.Nil(t, userPropertiesMap3["_$os"])
	assert.Nil(t, userPropertiesMap3["_$osVersion"])
	assert.Nil(t, userPropertiesMap3["_$browser"])
	assert.Nil(t, userPropertiesMap3["_$platform"])
	assert.Nil(t, userPropertiesMap3["_$screenWidth"])
	assert.Nil(t, userPropertiesMap3["_$screenHeight"])
}

func TestSDKGetProjectSettingsHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKRoutes(r)
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

	// Test Get project settings with invalid token.
	randomToken := U.RandomLowerAphaNumString(32)
	w = ServeGetRequestWithHeaders(r, uri, map[string]string{"Authorization": randomToken})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotNil(t, jsonResponseMap["error"])
}

func TestSDKBulk(t *testing.T) {
	r := gin.Default()
	H.InitSDKRoutes(r)

	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	uri := "/sdk/event/track/bulk"

	t.Run("Success", func(t *testing.T) {
		payload := fmt.Sprintf("[%s,%s]", `{"event_name": "signup", "event_properties": {"mobile" : "true"}}`, `{"event_name":"test", "event_properties": {"mobile" : "true"}}`)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		resp := make([]handler.SDKTrackResponse, 0, 0)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &resp)
		assert.Equal(t, 2, len(resp))
	})

	t.Run("DuplicateCustomerEventId", func(t *testing.T) {
		payload := fmt.Sprintf("[%s,%s,%s]", `{"event_name": "signup", "event_properties": {"mobile" : "true"}}`, `{"event_name":"test","c_event_id":"1", "event_properties": {"mobile" : "true"}}`, `{"event_name":"test2","c_event_id":"1", "event_properties": {"mobile" : "true"}}`)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		resp := make([]handler.SDKTrackResponse, 0, 0)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(jsonResponse, &resp)

		assert.Equal(t, 3, len(resp))

		assert.NotEmpty(t, resp[1].UserId)

		assert.Equal(t, "1", *resp[2].CustomerEventId)
		assert.Equal(t, "Tracking failed. Event creation failed. Duplicate CustomerEventID", resp[2].Error)
	})

}
