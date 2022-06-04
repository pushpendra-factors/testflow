package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAPIGetProfileUserHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	m := map[string]string{"$name": "Batman!"}
	propertiesJSON, err := json.Marshal(m)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}

	boolTrue := true
	customerEmail := "@example.com"

	// Create 5 Users.
	users := make([]model.User, 0)
	numUsers := 5
	for i := 0; i < numUsers; i++ {
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			Group1ID:       "1",
			Group2ID:       "2",
			CustomerUserId: strconv.Itoa(i) + customerEmail,
			Properties:     properties,
			IsGroupUser:    &boolTrue,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileUserRequest(r, project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Contact, 0)
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(users), 5)
		assert.Contains(t, resp[0].Identity, customerEmail)
	})
}

func sendGetProfileUserRequest(r *gin.Engine, projectId uint64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/users", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIGetProfileUserDetailsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.NotNil(t, agent)
	assert.Nil(t, err)

	props := map[string]string{
		"$name":          "Cameron Williomson",
		"$role":          "Head of Marketing",
		"$company":       "Freshworks",
		"$country":       "Australia",
		"$session_count": "8",
	}
	propertiesJSON, err := json.Marshal(props)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}

	boolTrue := true
	customerEmail := "abc@example.com"

	createdUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
		Group1ID:       "1",
		Group2ID:       "2",
		Group3ID:       "3",
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &boolTrue,
	})
	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	assert.Equal(t, user.ID, createdUserID)
	assert.Equal(t, http.StatusFound, errCode)
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, model.AllowedGroupNames)
	assert.NotNil(t, group3)
	assert.Equal(t, http.StatusCreated, status)

	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	randomeEventName := RandomURL()
	trackPayload := SDK.TrackPayload{
		EventId:         "",
		UserId:          createdUserID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            randomeEventName,
		CustomerEventId: new(string),
		EventProperties: U.PropertiesMap{"$qp_utm_campaign": "campaign1"},
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       project.ID,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	timestamp = timestamp + 10000
	trackPayload = SDK.TrackPayload{
		Name:          U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:     timestamp,
		UserId:        user.ID,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	timestamp = timestamp + 10000
	trackPayload = SDK.TrackPayload{
		Name:      U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp: timestamp,
		UserId:    user.ID,
		EventProperties: U.PropertiesMap{
			"$qp_utm_campaign": "campaign2",
		},
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	timestamp = timestamp + 10000
	trackPayload = SDK.TrackPayload{
		Name:          U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:     timestamp,
		UserId:        user.ID,
		RequestSource: model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	isAnonymous := "true"
	if len(user.CustomerUserId) != 0 {
		isAnonymous = "false"
	}

	userId := user.CustomerUserId
	if isAnonymous == "true" {
		userId = user.ID
	}

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileUserDetailsRequest(r, project.ID, agent, userId, isAnonymous)
		// log.Fatal("Output::", w)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.ContactDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.UserId, userId)
		assert.Contains(t, resp.Name, "Cameron")
		assert.Equal(t, resp.Company, "Freshworks")
		assert.Contains(t, resp.Role, "Head")
		assert.Equal(t, resp.Email, customerEmail)
		assert.Equal(t, resp.Country, "Australia")
		assert.Equal(t, resp.WebSessionsCount, uint64(8))
		assert.NotNil(t, resp.GroupInfos)
		assert.Condition(t, func() bool { return len(resp.GroupInfos) <= 4 })
		assert.Equal(t, resp.GroupInfos[0].GroupName, model.GROUP_NAME_HUBSPOT_COMPANY)
		assert.Equal(t, resp.GroupInfos[len(resp.GroupInfos)-1].GroupName, model.GROUP_NAME_SALESFORCE_OPPORTUNITY)
		assert.NotNil(t, resp.UserActivity)
		assert.Condition(t, func() bool {
			if resp.UserActivity == nil {
				return false
			}
			n := len(resp.UserActivity)
			for i, activity := range resp.UserActivity {
				assert.NotNil(t, activity.EventName)
				if i < n-1 {
					assert.Equal(t, activity.EventName, U.EVENT_NAME_FORM_SUBMITTED)
				}
				assert.NotNil(t, activity.Timestamp)
				if i > 1 {
					if resp.UserActivity[i].Timestamp > resp.UserActivity[i-1].Timestamp {
						return false
					}
				}

			}
			assert.Equal(t, resp.UserActivity[n-1].EventName, randomeEventName)
			return true
		})
	})
}

func sendGetProfileUserDetailsRequest(r *gin.Engine, projectId uint64, agent *model.Agent, userId string, isAnonymous string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/users/%s?is_anonymous=%s", projectId, userId, isAnonymous)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}
