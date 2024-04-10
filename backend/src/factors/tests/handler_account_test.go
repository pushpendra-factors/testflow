package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
)

func TestSDKTrackAccount(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAccountRoutes(r)
	uriCreate := "/sdk/account/create"
	uriUpdate := "/sdk/account/update"
	uriTrack := "/sdk/account/event/track"

	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	_, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)

	t.Run("TestForCreatingAccount", func(t *testing.T) {
		w := ServePostRequestWithHeaders(r, uriCreate,
			[]byte(fmt.Sprintf(`{"domain": "abc.com","properties": {"country": "United States",	"employee_count": "1M - 10M"}}`)), map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusCreated, w.Code)
		responseMap := DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)

		groupUser, status := store.GetStore().GetGroupUserByGroupID(project.ID, model.GROUP_NAME_DOMAINS, "abc.com")
		assert.Equal(t, http.StatusFound, status)
		assert.NotNil(t, groupUser)

		propertiesBytes, err := groupUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties map[string]interface{}
		json.Unmarshal(propertiesBytes.([]byte), &userProperties)

		assert.Equal(t, userProperties[U.UP_IS_OFFLINE], true)

		// create account failure

		w = ServePostRequestWithHeaders(r, uriCreate,
			[]byte(fmt.Sprintf(`{"domain": "abc.com","properties": {"country": "United States",	"employee_count": "1M - 10M"}}`)), map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusConflict, w.Code)

	})

	t.Run("TestForUpdatingAccount", func(t *testing.T) {

		// cheack failure while updating

		w := ServePostRequestWithHeaders(r, uriUpdate,
			[]byte(fmt.Sprintf(`{"domain": "abcd.com","properties": {"country": "United States",	"employee_count": "1M - 10M"}}`)), map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusNotFound, w.Code)

		groupUserID, status := store.GetStore().CreateOrGetDomainGroupUser(project.ID, model.GROUP_NAME_DOMAINS, "abcd.com", U.TimeNowUnix(), model.GetGroupUserSourceByGroupName(model.GROUP_NAME_DOMAINS))
		assert.Equal(t, http.StatusCreated, status)

		groupUser, status := store.GetStore().GetUser(project.ID, groupUserID)
		assert.Equal(t, http.StatusFound, status)

		propertiesBytes, err := groupUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties map[string]interface{}
		json.Unmarshal(propertiesBytes.([]byte), &userProperties)

		assert.NotEqual(t, userProperties[U.UP_IS_OFFLINE], true)

		// update account

		w = ServePostRequestWithHeaders(r, uriUpdate,
			[]byte(fmt.Sprintf(`{"domain": "abcd.com","properties": {"country": "United States",	"employee_count": "1M - 10M"}}`)), map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusOK, w.Code)

		groupUser, status = store.GetStore().GetUser(project.ID, groupUserID)
		assert.Equal(t, http.StatusFound, status)

		propertiesBytes, err = groupUser.Properties.Value()
		assert.Nil(t, err)
		json.Unmarshal(propertiesBytes.([]byte), &userProperties)

		assert.Equal(t, userProperties[U.UP_IS_OFFLINE], true)

	})

	t.Run("TestForTrackAccount", func(t *testing.T) {

		groupUserID, status := store.GetStore().CreateOrGetDomainGroupUser(project.ID, model.GROUP_NAME_DOMAINS, "abcde.com", U.TimeNowUnix(), model.GetGroupUserSourceByGroupName(model.GROUP_NAME_DOMAINS))
		assert.Equal(t, http.StatusCreated, status)

		groupUser, status := store.GetStore().GetUser(project.ID, groupUserID)
		assert.Equal(t, http.StatusFound, status)

		propertiesBytes, err := groupUser.Properties.Value()
		assert.Nil(t, err)
		var userProperties map[string]interface{}
		json.Unmarshal(propertiesBytes.([]byte), &userProperties)

		eventName1 := "ev1"
		_, status = store.GetStore().CreateOrGetEventName(&model.EventName{ProjectId: project.ID, Name: eventName1, Type: model.TYPE_USER_CREATED_EVENT_NAME})
		assert.Equal(t, http.StatusCreated, status)
		// track account

		w := ServePostRequestWithHeaders(r, uriTrack,
			[]byte(fmt.Sprintf(`{"domain": "abcde.com","event": { "name": "ev1" ,"timestamp" : %d,"properties": {"country": "United States" }}}`, U.TimeNowUnix())), map[string]string{"Authorization": project.PrivateToken})
		assert.Equal(t, http.StatusOK, w.Code)

		responseMap := DecodeJSONResponseToMap(w.Body)
		assert.NotEmpty(t, responseMap)

		trackedEvent, status := store.GetStore().GetEventById(project.ID, responseMap["eventID"].(string), "")
		assert.Equal(t, http.StatusFound, status)

		propertiesBytes, err = trackedEvent.Properties.Value()
		assert.Nil(t, err)
		var eventPropertiesBytes map[string]interface{}
		json.Unmarshal(propertiesBytes.([]byte), &eventPropertiesBytes)

		assert.Equal(t, eventPropertiesBytes["country"], "United States")

	})

}
