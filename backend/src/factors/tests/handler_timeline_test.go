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
	"sort"
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

	m := map[string]string{"$country": "Ukraine"}
	propertiesJSON, err := json.Marshal(m)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}

	customerEmail := "@example.com"

	// Create 5 Users with Properties.
	users := make([]model.User, 0)
	numUsers := 5
	for i := 1; i <= numUsers; i++ {
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			Group1ID:       "1",
			Group2ID:       "2",
			CustomerUserId: "user" + strconv.Itoa(i) + customerEmail,
			Properties:     properties,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), numUsers)
	// Create 5 Users without Properties.
	users = make([]model.User, 0)
	numUsers = 5
	for i := 1; i <= numUsers; i++ {
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			Group1ID:       "1",
			Group2ID:       "2",
			CustomerUserId: "user" + strconv.Itoa(i+5) + customerEmail,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), numUsers)

	var payload model.UTListPayload
	payload.Source = "web"

	filters := model.QueryProperty{
		Entity:    "user_g",
		Type:      "categorical",
		Property:  "$country",
		Operator:  "equals",
		Value:     "Ukraine",
		LogicalOp: "AND",
	}
	payload.Filters = append(payload.Filters, filters)

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileUserRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Contact, 0)
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), 5)
		assert.Condition(t, func() bool { return len(resp) <= 1000 })
		assert.Condition(t, func() bool {
			for i, user := range resp {
				assert.Equal(t, user.IsAnonymous, false)
				if i < 5 {
					assert.Equal(t, user.Country, "Ukraine")
				} else {
					assert.Equal(t, user.Country, "")
				}
				assert.NotNil(t, user.LastActivity)
			}
			return true
		})
	})
}

func sendGetProfileUserRequest(r *gin.Engine, projectId int64, agent *model.Agent, payload model.UTListPayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/profiles/users", projectId)).
		WithPostParams(payload).
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
		"$name":               "Cameron Williomson",
		"$company":            "Freshworks",
		"$country":            "Australia",
		"$session_count":      "8",
		"$session_spent_time": "500",
		"$page_count":         "10",
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
	group4, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_DEAL, model.AllowedGroupNames)
	assert.NotNil(t, group4)
	assert.Equal(t, http.StatusCreated, status)

	// Event 1 : Page View
	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	randomURL := RandomURL()
	trackPayload := SDK.TrackPayload{
		EventId:         "",
		UserId:          createdUserID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            randomURL,
		CustomerEventId: new(string),
		EventProperties: map[string]interface{}{},
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       project.ID,
		Auto:            true,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	// Event 2 : Web Session
	timestamp = timestamp - 10000
	sessionProperties := map[string]interface{}{
		U.EP_PAGE_COUNT:   "5",
		U.EP_CHANNEL:      "ChannelName",
		U.EP_CAMPAIGN:     "CampaignName",
		U.SP_SESSION_TIME: "120",
		U.EP_REFERRER_URL: RandomURL(),
	}
	trackPayload = SDK.TrackPayload{
		EventId:         "",
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SESSION,
		CustomerEventId: new(string),
		EventProperties: sessionProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       0,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 3 : Form Submit
	timestamp = timestamp - 10000
	formSubmitProperties := map[string]interface{}{
		U.EP_FORM_NAME: "FormName",
		U.EP_PAGE_URL:  RandomURL(),
	}
	trackPayload = SDK.TrackPayload{
		EventId:         "",
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		CustomerEventId: new(string),
		EventProperties: formSubmitProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       0,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 4 : Offline Touchpoint
	timestamp = timestamp - 10000
	touchpointProperties := map[string]interface{}{
		U.EP_CHANNEL:  "ChannelName",
		U.EP_CAMPAIGN: "CampaignName",
	}
	trackPayload = SDK.TrackPayload{
		EventId:         "",
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		CustomerEventId: new(string),
		EventProperties: touchpointProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       0,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 5 : Campaign Member Created
	timestamp = timestamp - 10000
	campCreatedProperties := map[string]interface{}{
		"$salesforce_campaign_name":     "Campaign Name",
		model.EP_SFCampaignMemberStatus: "CurrentStatus",
	}
	trackPayload = SDK.TrackPayload{
		EventId:         "",
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
		CustomerEventId: new(string),
		EventProperties: campCreatedProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       0,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 6 : Campaign Member Updated
	timestamp = timestamp - 10000
	campUpdatedProperties := map[string]interface{}{
		"$salesforce_campaign_name":     "Campaign Name",
		model.EP_SFCampaignMemberStatus: "CurrentStatus",
	}
	trackPayload = SDK.TrackPayload{
		EventId:         "",
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
		CustomerEventId: new(string),
		EventProperties: campUpdatedProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       0,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 7 : Random Event
	timestamp = timestamp - 10000
	randomProperties := map[string]interface{}{}
	trackPayload = SDK.TrackPayload{
		EventId:         "",
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
		CustomerEventId: new(string),
		EventProperties: randomProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       0,
		Auto:            false,
		ClientIP:        "",
		UserAgent:       "",
		SmartEventType:  "",
		RequestSource:   model.UserSourceHubspot,
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
	eventNamePropertiesMap := map[string][]string{
		U.EVENT_NAME_SESSION:                           {U.EP_PAGE_COUNT, U.EP_CHANNEL, U.EP_CAMPAIGN, U.SP_SESSION_TIME, U.EP_TIMESTAMP, U.EP_REFERRER_URL},
		U.EVENT_NAME_FORM_SUBMITTED:                    {U.EP_FORM_NAME, U.EP_PAGE_URL, U.EP_TIMESTAMP},
		U.EVENT_NAME_OFFLINE_TOUCH_POINT:               {U.EP_CHANNEL, U.EP_CAMPAIGN, U.EP_TIMESTAMP},
		U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED: {"$salesforce_campaign_name", model.EP_SFCampaignMemberStatus, U.EP_TIMESTAMP},
		U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED: {"$salesforce_campaign_name", model.EP_SFCampaignMemberStatus, U.EP_TIMESTAMP},
	}
	pageViewPropsList := []string{U.EP_IS_PAGE_VIEW, U.EP_PAGE_SPENT_TIME, U.EP_PAGE_SCROLL_PERCENT, U.EP_PAGE_LOAD_TIME}

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileUserDetailsRequest(r, project.ID, agent, userId, isAnonymous)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.ContactDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.UserId, userId)
		assert.Contains(t, resp.Name, "Cameron")
		assert.Equal(t, resp.Company, "Freshworks")
		assert.Equal(t, resp.Email, customerEmail)
		assert.Equal(t, resp.Country, "Australia")
		assert.Equal(t, resp.WebSessionsCount, float64(1))
		assert.Equal(t, resp.NumberOfPageViews, float64(10))
		assert.Equal(t, resp.TimeSpentOnSite, float64(500))
		assert.NotNil(t, resp.GroupInfos)
		assert.Condition(t, func() bool { return len(resp.GroupInfos) <= 4 })
		assert.NotNil(t, resp.UserActivity)
		assert.Condition(t, func() bool {
			if resp.UserActivity == nil {
				return false
			}
			for i, activity := range resp.UserActivity {
				assert.NotNil(t, activity.EventName)
				assert.NotNil(t, activity.DisplayName)
				if activity.EventName == randomURL {
					assert.Equal(t, activity.DisplayName, "Page View")
				} else {
					assert.Equal(t, U.STANDARD_EVENTS_DISPLAY_NAMES[activity.EventName], activity.DisplayName)
				}
				assert.NotNil(t, activity.Timestamp)
				assert.Condition(t, func() bool { return activity.Timestamp <= uint64(time.Now().UTC().Unix()) })
				if i > 1 {
					if resp.UserActivity[i].Timestamp > resp.UserActivity[i-1].Timestamp {
						return false
					}
				}
				assert.NotNil(t, activity.Properties)
				assert.Condition(t, func() bool {
					properties, err := U.DecodePostgresJsonb(&activity.Properties)
					_, eventExistsInMap := eventNamePropertiesMap[activity.EventName]
					assert.Nil(t, err)
					if activity.DisplayName == "Page View" {
						for key := range *properties {
							sort.Strings(pageViewPropsList)
							i := sort.SearchStrings(pageViewPropsList, key)
							assert.Condition(t, func() bool { return i < len(pageViewPropsList) })
						}
					} else if eventExistsInMap {
						for key := range *properties {
							sort.Strings(eventNamePropertiesMap[activity.EventName])
							i := sort.SearchStrings(eventNamePropertiesMap[activity.EventName], key)
							assert.Condition(t, func() bool { return i < len(eventNamePropertiesMap[activity.EventName]) })
						}
					} else {
						assert.Equal(t, *properties, randomProperties)
					}

					return true
				})

			}
			return true
		})
	})
}

func sendGetProfileUserDetailsRequest(r *gin.Engine, projectId int64, agent *model.Agent, userId string, isAnonymous string) *httptest.ResponseRecorder {
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
