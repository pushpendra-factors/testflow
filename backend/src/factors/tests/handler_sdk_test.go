package tests

import (
	"encoding/json"
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

func TestSDKTrack(t *testing.T) {
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
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_1", "event_properties": {}}`, user.ID)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "event_2"}`, user.ID)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Test track user created event.
	rEventName := U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "event_name": "%s", "event_properties": {"mobile" : "true"}}`, user.ID, rEventName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	ren, errCode := M.GetEventNameByFilter(&M.EventName{ProjectId: project.ID, Name: rEventName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, M.USER_CREATED_EVENT_NAME, ren.AutoName)
	assert.NotEqual(t, ren.Name, ren.AutoName)

	// Test auto tracked event.
	rEventName = U.RandomLowerAphaNumString(10)
	w = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s", "auto": true, "event_name": "%s", "event_properties": {"mobile" : "true"}}`, user.ID, rEventName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	assert.Nil(t, responseMap["user_id"])
	ren, errCode = M.GetEventNameByFilter(&M.EventName{ProjectId: project.ID, Name: rEventName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotEqual(t, M.USER_CREATED_EVENT_NAME, ren.AutoName)
	assert.Equal(t, ren.Name, ren.AutoName)
}

func TestSDKIdentify(t *testing.T) {
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

func TestSDKAddUserProperties(t *testing.T) {
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
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test default user properties.
	uniqueName = U.RandomLowerAphaNumString(16)
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"prop_1": "%s"}}`, user.ID, uniqueName)),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotEmpty(t, responseMap)
	retUser, errCode := M.GetUser(project.ID, user.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
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
	assert.Equal(t, M.DB_SUCCESS, errCode)
	userPropertiesBytes, err = retUser.Properties.Value()
	assert.Nil(t, err)
	var userPropertiesMap2 map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap2)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, userPropertiesMap2["int_prop"])
	assert.NotNil(t, userPropertiesMap2["long_prop"])
	assert.NotNil(t, userPropertiesMap2["string_prop"])
	assert.NotNil(t, userPropertiesMap2["float_prop"])
	// Types not allowed.
	assert.Nil(t, userPropertiesMap2["boolean_prop"])
	assert.Nil(t, userPropertiesMap2["map_prop"])

	// Should not allow $ prefixes apart from default properties.
	w = ServePostRequestWithHeaders(r, uri, []byte(fmt.Sprintf(`{"user_id": "%s", "properties": {"$dollar_key": "unknow_value", "$os": "mac osx", "$osVersion": "1_2_3", "$referrer": "http://google.com", "$screenWidth": 10, "$screenHeight": 11, "$browser": "mozilla", "$browserVersion": "10_2_3"}}`,
		user.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	propsResponseMap1 := DecodeJSONResponseToMap(w.Body)
	assert.Nil(t, propsResponseMap1["user_id"])
	retUser, errCode = M.GetUser(project.ID, user.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	userPropertiesBytes, err = retUser.Properties.Value()
	assert.Nil(t, err)
	var userPropertiesMap3 map[string]interface{}
	json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap3)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Nil(t, userPropertiesMap3["$dollar_key"])                                              // dollar prefix not allowed.
	assert.NotNil(t, userPropertiesMap3[fmt.Sprintf("%s$dollar_key", U.NAME_PREFIX_ESCAPE_CHAR)]) // Escaped key should exist.
	// check for default props. Hardcoded property name as request payload.
	assert.NotNil(t, userPropertiesMap3["$os"])
	assert.NotNil(t, userPropertiesMap3["$osVersion"])
	assert.NotNil(t, userPropertiesMap3["$browser"])
	assert.NotNil(t, userPropertiesMap3["$referrer"])
	assert.NotNil(t, userPropertiesMap3["$screenWidth"])
	assert.NotNil(t, userPropertiesMap3["$screenHeight"])
}

func TestSDKGetProjectSettings(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKRoutes(r)
	uri := "/sdk/project/get_settings"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	_, err = SetupProjectDependenciesReturnDAO(project)
	assert.Nil(t, err)

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
