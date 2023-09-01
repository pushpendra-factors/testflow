package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/memsql"
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

	timelinesConfig := &model.TimelinesConfig{
		UserConfig: model.UserConfig{
			TableProps: []string{"$country", "$page_count"},
		},
	}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	// Properties Map
	hbMeetingTime := time.Now().AddDate(0, 0, -10).Unix()
	hbMeetingTimeNow := time.Now().Unix()
	propsMap := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 105, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 120, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Paris", "$country": "France", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 3000, "$hubspot_contact_rh_meeting_time": hbMeetingTimeNow},
		{"$browser": "Edge", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 110, "$session_spent_time": 2500, "$hubspot_contact_rh_meeting_time": hbMeetingTimeNow},
		{"$browser": "Firefox", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 100, "$session_spent_time": 3000, "$hubspot_contact_rh_meeting_time": hbMeetingTime},
		{"$browser": "Firefox", "$city": "Dubai", "$country": "UAE", "$device_type": "desktop", "$page_count": 150, "$session_spent_time": 2100, "$hubspot_contact_rh_meeting_time": hbMeetingTime},
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 150, "$session_spent_time": 2100, "$hubspot_contact_rh_meeting_time": hbMeetingTime},
	}

	// Create 5 Unidentified Users
	users := make([]model.User, 0)
	numUsers := 5
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[9-i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:  project.ID,
			Source:     model.GetRequestSourcePointer(model.UserSourceWeb),
			Properties: properties,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 5)

	// Create 5 Identified Users from UserSourceWeb
	numUsers = 5
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			CustomerUserId: "user" + strconv.Itoa(i+1) + "@example.com",
			Properties:     properties,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 10)

	// Create 2 Identified Users from UserSourceSalesforce
	numUsers = 2
	for i := 5; i < 5+numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
			CustomerUserId: "user" + strconv.Itoa(i+1) + "@example.com",
			Properties:     properties,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 12)

	// Create 3 Identified Users from UserSourceHubspot
	numUsers = 3
	for i := 7; i < 7+numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
			CustomerUserId: "user" + strconv.Itoa(i+1) + "@example.com",
			Properties:     properties,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 15)

	var payload model.TimelinePayload

	// Test Cases :-
	// 1. Users from Different Sources (No filter, no segment applied)
	sourceToUserCountMap := map[string]int{"All": 15, model.UserSourceSalesforceString: 2, model.UserSourceHubspotString: 3}
	for source, count := range sourceToUserCountMap {
		payload.Query.Source = source
		w := sendGetProfileUserRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), count)
		for i, user := range resp {
			if model.IsDomainGroup(source) {
				if i < 10 {
					assert.Equal(t, user.IsAnonymous, false)
				} else {
					assert.Equal(t, user.IsAnonymous, true)
				}
			} else {
				assert.Equal(t, user.IsAnonymous, false)
			}
			for _, prop := range timelinesConfig.UserConfig.TableProps {
				assert.NotNil(t, user.TableProps[prop])
			}
			assert.NotNil(t, user.LastActivity)
			if i > 0 {
				assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
			}
		}
	}

	// 2. UserSourceWeb (1 filter, no segment applied)
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
				},
			},
			Source: "web",
		},
	}
	w := sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	resp := make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	for i, user := range resp {
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, user.TableProps[prop])
		}
		assert.NotNil(t, user.LastActivity)
		if i > 0 {
			assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
		}
	}

	// 2. UserSourceWeb (1 search filter applied)
	payload = model.TimelinePayload{
		Query: model.Query{
			Source: "web",
		}, SearchFilter: []model.QueryProperty{
			{
				Entity:    "user_g",
				Type:      "categorical",
				Property:  "$user_id",
				Operator:  "contains",
				Value:     "user2",
				LogicalOp: "AND",
			},
		},
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	assert.Equal(t, resp[0].Identity, "user2@example.com")
	for _, prop := range timelinesConfig.UserConfig.TableProps {
		assert.NotNil(t, resp[0].TableProps[prop])
	}
	assert.NotNil(t, resp[0].LastActivity)

	// // 3. UserSourceWeb (Query from Segment, no filters)
	payload = model.TimelinePayload{
		Query: model.Query{
			Type:                 "unique_users",
			EventsCondition:      "any_given_event",
			EventsWithProperties: []model.QueryEventWithProperties{},
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$browser",
					Operator:  "equals",
					Value:     "Chrome",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$browser",
					Operator:  "equals",
					Value:     "Firefox",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$city",
					Operator:  "equals",
					Value:     "Delhi",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$device_type",
					Operator:  "equals",
					Value:     "iPad",
					LogicalOp: "OR",
				},
			},
			GroupAnalysis: "users",
			Source:        "web",
			TableProps:    []string{"$country", "$page_count"},
		},
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 4)
	for i, user := range resp {
		if i < 2 {
			assert.Equal(t, user.IsAnonymous, false)
		} else {
			assert.Equal(t, user.IsAnonymous, true)
		}
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, user.TableProps[prop])
		}
		assert.NotNil(t, user.LastActivity)
		if i > 0 {
			assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
		}
	}

	year, month, day := time.Now().AddDate(0, 0, -1).Date()
	jointest := time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location()).UnixMilli()
	// 6. (a) Test for dateTime type filters (since)
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "datetime",
					Property:  "$hubspot_contact_rh_meeting_time",
					Operator:  "since",
					Value:     fmt.Sprintf("{\"fr\":%d}", jointest),
					LogicalOp: "AND",
				},
			},
			Source: "web",
		},
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))
	for i, user := range resp {
		assert.Equal(t, user.IsAnonymous, true)
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, user.TableProps[prop])
		}
		assert.NotNil(t, user.LastActivity)
		if i > 0 {
			assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
		}
	}

	// 6. (b) Test for dateTime type filters (before)
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "datetime",
					Property:  "$hubspot_contact_rh_meeting_time",
					Operator:  "before",
					Value:     fmt.Sprintf("{\"to\":%d}", jointest),
					LogicalOp: "AND",
				},
			},
			Source: "web",
		},
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(resp))
	for i, user := range resp {
		assert.Equal(t, user.IsAnonymous, true)
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, user.TableProps[prop])
		}
		assert.NotNil(t, user.LastActivity)
		if i > 0 {
			assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
		}
	}

	// 6. (c) Test for dateTime type filters (inCurrent)
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "datetime",
					Property:  "$hubspot_contact_rh_meeting_time",
					Operator:  "inCurrent",
					Value:     "{\"gran\":\"week\"}",
					LogicalOp: "AND",
				},
			},
			Source: "web",
		},
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp))
	for i, user := range resp {
		assert.Equal(t, user.IsAnonymous, true)
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, user.TableProps[prop])
		}
		assert.NotNil(t, user.LastActivity)
		if i > 0 {
			assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
		}
	}
}

func sendGetProfileUserRequest(r *gin.Engine, projectId int64, agent *model.Agent, payload model.TimelinePayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/profiles/users?score=true&debug=true", projectId)).
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

	var timelinesConfig model.TimelinesConfig

	timelinesConfig.UserConfig.LeftpaneProps = []string{"$email", "$page_count", "$user_id", "$name", "$session_spent_time"}
	timelinesConfig.UserConfig.Milestones = []string{"$milesone_1", "$milesone_2", "$milesone_3"}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	// Create Groups
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
	group5, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group5)

	// Create Domain Group
	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
	source := model.GetRequestSourcePointer(model.UserSourceDomains)
	groupUser := true
	domID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         source,
		Group5ID:       "abc.xyz",
		CustomerUserId: "domainuser",
		Properties:     domProperties,
		IsGroupUser:    &groupUser,
	})
	_, errCode = store.GetStore().GetUser(project.ID, domID)
	assert.Equal(t, http.StatusFound, errCode)

	// Create Associated Account
	props := map[string]interface{}{
		"$hubspot_company_name": "Freshworks",
		"$country":              "Australia",
	}
	propertiesJSON, err := json.Marshal(props)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	accProps := postgres.Jsonb{RawMessage: propertiesJSON}
	isGroupUser := true

	accountID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Properties:  accProps,
		IsGroupUser: &isGroupUser,
		Group1ID:    "1",
		Group2ID:    "2",
		Source:      model.GetRequestSourcePointer(model.UserSourceHubspot),
	})

	props = map[string]interface{}{
		"$name":               "Cameron Williomson",
		"$company":            "Freshworks",
		"$country":            "Australia",
		"$session_count":      8,
		"$session_spent_time": 500,
		"$page_count":         10,
		"$milesone_1":         U.UnixTimeBeforeDuration(1 * time.Hour),
		"$milesone_2":         U.UnixTimeBeforeDuration(2 * time.Hour),
		"$milesone_3":         U.UnixTimeBeforeDuration(3 * time.Hour),
		"$milesone_4":         U.UnixTimeBeforeDuration(4 * time.Hour),
		"$milesone_5":         U.UnixTimeBeforeDuration(5 * time.Hour),
	}
	propertiesJSON, err = json.Marshal(props)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}
	isGroupUser = false
	customerEmail := "abc@example.com"

	createdUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
		Group1ID:       "1",
		Group2ID:       "2",
		Group1UserID:   accountID,
		Group5UserID:   domID,
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &isGroupUser,
	})
	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	assert.Equal(t, user.ID, createdUserID)
	assert.Equal(t, http.StatusFound, errCode)

	// event properties map
	eventProperties := map[string]interface{}{
		U.EP_PAGE_COUNT:                              5,
		U.EP_CHANNEL:                                 "ChannelName",
		U.EP_CAMPAIGN:                                "CampaignName",
		U.SP_SPENT_TIME:                              120,
		U.EP_REFERRER_URL:                            RandomURL(),
		U.EP_FORM_NAME:                               "Form Name",
		U.EP_PAGE_URL:                                RandomURL(),
		U.EP_SALESFORCE_CAMPAIGN_TYPE:                "Some Type",
		U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:        "CurrentStatus",
		U.EP_HUBSPOT_ENGAGEMENT_SOURCE:               "Some Engagement Source",
		U.EP_HUBSPOT_ENGAGEMENT_FROM:                 "Somewhere",
		U.EP_HUBSPOT_ENGAGEMENT_TYPE:                 "Some Engagement Type",
		U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME:       "Some Outcome",
		U.EP_HUBSPOT_ENGAGEMENT_STARTTIME:            "Start time",
		U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS: 10000000000,
		U.EP_HUBSPOT_ENGAGEMENT_STATUS:               "Testing",
		U.EP_HUBSPOT_FORM_SUBMISSION_FORMTYPE:        "Some HS Form Submission Type",
		U.EP_HUBSPOT_FORM_SUBMISSION_PAGEURL:         RandomURL(),
		U.EP_HUBSPOT_ENGAGEMENT_ENDTIME:              "End Time",
		U.EP_SALESFORCE_CAMPAIGN_NAME:                "Some Salesforce Campaign Name",
		U.EP_HUBSPOT_FORM_SUBMISSION_TITLE:           "Some form submission title",
		U.EP_HUBSPOT_ENGAGEMENT_SUBJECT:              "Some Engagement Subject",
		U.EP_HUBSPOT_ENGAGEMENT_TITLE:                "Some Engagement Title",
		U.EP_SF_TASK_TYPE:                            "Some Task Type",
		U.EP_SF_TASK_SUBTYPE:                         "Some Task SubType",
		U.EP_SF_TASK_COMPLETED_DATETIME:              1660875887,
		U.EP_SF_EVENT_TYPE:                           "Some Event Type",
		U.EP_SF_EVENT_SUBTYPE:                        "Some Event Subtype",
		U.EP_SF_EVENT_COMPLETED_DATETIME:             1660875887,
		"$curr_prop_value":                           "Current Custom Value",
		"$prev_prop_value":                           "Previous Custom Value",
	}
	// Event 1 : Page View
	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	randomURL := RandomURL()
	trackPayload := SDK.TrackPayload{
		UserId:          createdUserID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            randomURL,
		EventProperties: map[string]interface{}{},
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		ProjectId:       project.ID,
		Auto:            true,
		RequestSource:   model.UserSourceWeb,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	// Event 2 : Web Session
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SESSION,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 3 : Form Submit
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 4 : Offline Touchpoint
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceWeb,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 5 : Campaign Member Created
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 6 : Campaign Member Responded to Campaign
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 7 : Hubspot Form Submission
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 8 : Engagement Email
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 9 : Engagement Meeting Created
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 10 : Engagement Call Created
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceHubspot,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 11 : Salesforce Task Created
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_TASK_CREATED,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 11 : Salesforce Task Created
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_EVENT_CREATED,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Event 13 : Random Event
	timestamp = timestamp - 10000
	randomProperties := map[string]interface{}{}
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
		EventProperties: randomProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		RequestSource:   model.UserSourceHubspot,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Empty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)

	// Create Smart Events
	filter := &model.SmartCRMEventFilter{
		Source:               model.SmartCRMEventSourceSalesforce,
		ObjectType:           "contact",
		Description:          "salesforce contact",
		FilterEvaluationType: model.FilterEvaluationTypeAny,
		Filters: []model.PropertyFilter{
			{
				Name:  "page_spent_time",
				Rules: []model.CRMFilterRule{},
			},
		},
		LogicalOp:               model.LOGICAL_OP_AND,
		TimestampReferenceField: "time",
	}
	filter.Filters[0].Name = "Property 0"
	eventName, status := store.GetStore().CreateOrGetCRMSmartEventFilterEventName(project.ID,
		&model.EventName{ProjectId: project.ID, Name: fmt.Sprintf("Smart Event Name %d", 0)}, filter)
	assert.Equal(t, http.StatusCreated, status)
	// Event 14: Custom CRM Event
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            eventName.Name,
		EventProperties: eventProperties,
		UserProperties:  map[string]interface{}{},
		Timestamp:       timestamp,
		Auto:            false,
		SmartEventType:  eventName.Type,
		RequestSource:   model.UserSourceSalesforce,
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
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.ContactDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.UserId, userId)
		assert.Contains(t, resp.Name, "Cameron")
		assert.Equal(t, resp.Company, "Freshworks")
		assert.NotNil(t, resp.LeftPaneProps)
		userProps := user.Properties
		userPropsDecoded, err := U.DecodePostgresJsonb(&userProps)
		assert.Nil(t, err)
		for i, property := range resp.LeftPaneProps {
			assert.Equal(t, (*userPropsDecoded)[i], property)
		}
		for i, property := range resp.Milestones {
			assert.Equal(t, (*userPropsDecoded)[i], property)
		}
		assert.Equal(t, resp.Account, "abc.xyz")
		assert.NotNil(t, resp.UserActivity)
		for _, activity := range resp.UserActivity {
			assert.NotNil(t, activity.EventName)
			assert.NotNil(t, activity.DisplayName)
			eventFromMap, eventExistsInMap := model.HOVER_EVENTS_NAME_PROPERTY_MAP[activity.EventName]
			if activity.EventName == randomURL {
				assert.Equal(t, activity.DisplayName, "Page View")
				assert.Equal(t, activity.AliasName, "")
			} else if eventExistsInMap {
				assert.Equal(t, activity.DisplayName, U.STANDARD_EVENTS_DISPLAY_NAMES[activity.EventName])
				if activity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("Added to %s", eventProperties[U.EP_SALESFORCE_CAMPAIGN_NAME]))
				} else if activity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("Responded to %s", eventProperties[U.EP_SALESFORCE_CAMPAIGN_NAME]))
				} else if activity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_FORM_SUBMISSION_TITLE]))
				} else if activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("%s: %s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TYPE], eventProperties[U.EP_HUBSPOT_ENGAGEMENT_SUBJECT]))
				} else if activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED ||
					activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TITLE]))
				} else if activity.EventName == U.EVENT_NAME_SALESFORCE_TASK_CREATED {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("Created Task - %s", eventProperties[U.EP_SF_TASK_SUBJECT]))
				} else if activity.EventName == U.EVENT_NAME_SALESFORCE_EVENT_CREATED {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("Created Event - %s", eventProperties[U.EP_SF_EVENT_SUBJECT]))
				} else if activity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_LIST {
					assert.Equal(t, activity.AliasName, fmt.Sprintf("Added to Hubspot List - %s", eventProperties[U.EP_HUBSPOT_CONTACT_LIST_LIST_NAME]))
				}
			} else if model.IsEventNameTypeSmartEvent(activity.EventType) {
				assert.Equal(t, activity.EventName, "Smart Event Name 0")
				assert.Equal(t, activity.DisplayName, "Smart Event Name 0")
				assert.NotNil(t, activity.Properties)
			}
			assert.NotNil(t, activity.Timestamp)
			assert.True(t, activity.Timestamp <= uint64(time.Now().UTC().Unix()))
			if activity.DisplayName == "Page View" || eventExistsInMap {
				var lookInProps []string
				if activity.DisplayName == "Page View" {
					lookInProps = model.PAGE_VIEW_HOVERPROPS_LIST
				} else if eventExistsInMap {
					lookInProps = eventFromMap
				}
				assert.NotNil(t, activity.Properties)
				properties, err := U.DecodePostgresJsonb(activity.Properties)
				assert.Nil(t, err)
				for key := range *properties {
					sort.Strings(lookInProps)
					i := sort.SearchStrings(lookInProps, key)
					assert.True(t, i < len(lookInProps))
				}
			}
		}
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

func TestAPIGetProfileAccountHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	timelinesConfig := &model.TimelinesConfig{
		AccountConfig: model.AccountConfig{
			TableProps: []string{"$salesforce_account_billingcountry", "$hubspot_company_country", U.SIX_SIGNAL_COUNTRY},
		},
	}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	// Properties Map
	propertiesMap := []map[string]interface{}{
		{"$salesforce_account_name": "AdPushup", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "adpushup.com", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "Mad Street Den", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "madstreetden.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "Heyflow", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "heyflow.app", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "Clientjoy Ads", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "clientjoy.io", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor", "$browser": "Chrome", "$device_type": "PC", "$salesforce_city": "New Delhi"},
		{"$salesforce_account_name": "Adapt.IO", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "adapt.io", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC", "$hubspot_company_is_public": "true"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Heyflow", U.SIX_SIGNAL_COUNTRY: "Germany", U.SIX_SIGNAL_DOMAIN: "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Clientjoy Ads", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Adapt.IO", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome"},
	}

	userProps := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000, U.UP_COMPANY: "XYZ Company"},
		{"$browser": "Chrome", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 105, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 120, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 110, "$session_spent_time": 2500},
	}
	// Creating domain Account and Group
	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
	source := model.GetRequestSourcePointer(model.UserSourceDomains)
	groupUser := true
	accounts := make([]model.User, 0)
	domID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Source:      source,
		Group1ID:    "clientjoy.io",
		Properties:  domProperties,
		IsGroupUser: &groupUser,
	})
	_, errCode = store.GetStore().GetUser(project.ID, domID)
	assert.Equal(t, http.StatusFound, errCode)

	var payload model.TimelinePayload

	// Test :- No CRMs enabled
	payload.Query.Source = "$hubspot_company"
	w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	group, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)

	// Create 5 Salesforce Accounts
	numUsers := 5
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propertiesMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSalesforce)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group3ID:       "3",
			Group1ID:       "1",
			Group1UserID:   domID,
			CustomerUserId: fmt.Sprintf("sfuser%d@%s", i+1, propertiesMap[i]["$salesforce_account_website"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// 5 users associated to the account
		propertiesJSON, err = json.Marshal(userProps[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
			Properties:     properties,
			Group3ID:       "3",
			Group3UserID:   account.ID,
			CustomerUserId: fmt.Sprintf("salesforce@%duser", (i%5)+1),
		})
		_, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
	}

	// Create 5 Hubspot Companies
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propertiesMap[i+5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceHubspot)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group2ID:       "2",
			Group1ID:       "1",
			Group1UserID:   domID,
			CustomerUserId: fmt.Sprintf("hsuser%d@%s", i+1, propertiesMap[i+5]["$hubspot_company_domain"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// 5 users associated to the account
		propertiesJSON, err = json.Marshal(userProps[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
			Properties:     properties,
			Group2ID:       "2",
			Group2UserID:   account.ID,
			Group1UserID:   domID,
			CustomerUserId: fmt.Sprintf("hubspot@%duser", (i%5)+1),
		})
		_, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
	}

	// creating another domain account
	domID2, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Source:      source,
		Group1ID:    "1",
		IsGroupUser: &groupUser,
	})
	_, errCode = store.GetStore().GetUser(project.ID, domID2)
	assert.Equal(t, http.StatusFound, errCode)

	// creating a web user associated to domain
	propertiesJSON, err := json.Marshal(userProps[0])
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}
	createdWebUser, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
		Properties:     properties,
		Group1UserID:   domID2,
		CustomerUserId: "webuser@ymail.com",
	})
	_, errCode = store.GetStore().GetUser(project.ID, createdWebUser)
	assert.Equal(t, http.StatusFound, errCode)

	// Create 5 Six Signal Domains
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propertiesMap[i+10])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSixSignal)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group4ID:       "4",
			CustomerUserId: fmt.Sprintf("6siguser%d@%s", i+1, propertiesMap[i+10][U.SIX_SIGNAL_DOMAIN]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// 5 users associated to the account
		propertiesJSON, err = json.Marshal(userProps[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceSixSignal),
			Properties:     properties,
			Group4ID:       "4",
			Group4UserID:   account.ID,
			CustomerUserId: fmt.Sprintf("sixsignal@%duser", (i%5)+1),
		})
		_, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
	}
	assert.Equal(t, len(accounts), 15)

	//Source: $hubspot_company, 2 group exists
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)

	// 2 more groups
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.NotNil(t, group3)
	assert.Equal(t, http.StatusCreated, status)

	// Test Cases :-
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{{
				Entity:    "user_group",
				Type:      "categorical",
				Property:  "$browser",
				Operator:  "equals",
				Value:     "Chrome",
				LogicalOp: "AND",
			}},
			GroupAnalysis: "$hubspot_company",
			Source:        "$hubspot_company",
			TableProps:    []string{},
		},
		SearchFilter: []model.QueryProperty{},
	}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	resp := make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	assert.Condition(t, func() bool {
		for i, user := range resp {
			assert.Equal(t, user.Name, propertiesMap[7-i][U.GP_HUBSPOT_COMPANY_NAME])
			assert.Equal(t, user.HostName, propertiesMap[7-i][U.GP_HUBSPOT_COMPANY_DOMAIN])
			assert.NotNil(t, user.LastActivity)
			if i > 0 {
				assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
			}
			for _, prop := range timelinesConfig.UserConfig.TableProps {
				assert.NotNil(t, user.TableProps[prop])
			}
		}
		return true
	})

	// 1. Accounts from Different Sources (1 user filter, no segment applied)
	sourceToUserCountMap := map[string]int{"All": 2, U.GROUP_NAME_HUBSPOT_COMPANY: 3, U.GROUP_NAME_SALESFORCE_ACCOUNT: 3, U.GROUP_NAME_SIX_SIGNAL: 3}

	for source, count := range sourceToUserCountMap {
		payload = model.TimelinePayload{
			Query: model.Query{
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_group",
						Type:      "categorical",
						Property:  "$browser",
						Operator:  "equals",
						Value:     "Chrome",
						LogicalOp: "AND",
					},
				},
				GroupAnalysis: source,
				Source:        source,
				TableProps:    []string{},
			},
			SearchFilter: []model.QueryProperty{},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), count)
		for i, user := range resp {
			if model.IsDomainGroup(source) {
				assert.NotEmpty(t, user.Name)
				assert.NotEmpty(t, user.HostName)
			}
			if source == U.GROUP_NAME_HUBSPOT_COMPANY {
				assert.Equal(t, user.Name, propertiesMap[count+4-i]["$hubspot_company_name"])
				assert.Equal(t, user.HostName, propertiesMap[count+4-i]["$hubspot_company_domain"])
			}
			if source == U.GROUP_NAME_SALESFORCE_ACCOUNT {
				assert.Equal(t, user.Name, propertiesMap[count-i-1]["$salesforce_account_name"])
				assert.Equal(t, user.HostName, propertiesMap[count-i-1]["$salesforce_account_website"])
			}
			if source == U.GROUP_NAME_SIX_SIGNAL {
				assert.Equal(t, user.Name, propertiesMap[count+9-i][U.SIX_SIGNAL_NAME])
				assert.Equal(t, user.HostName, propertiesMap[count+9-i][U.SIX_SIGNAL_DOMAIN])
			}
			assert.NotNil(t, user.LastActivity)
			for _, prop := range timelinesConfig.UserConfig.TableProps {
				assert.NotNil(t, user.TableProps[prop])
			}
		}
	}

	// 3. Segment with multiple $hubspot_company filters
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_group",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "UK",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_group",
					Type:      "categorical",
					Property:  "$device_type",
					Operator:  "equals",
					Value:     "desktop",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "US",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  "equals",
					Value:     "50",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  "equals",
					Value:     "20",
					LogicalOp: "OR",
				},
			},
			Source: "$hubspot_company",
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	filteredCompaniesNameHostNameMap := map[string]string{"Adapt.IO": "adapt.io", "Clientjoy Ads": "clientjoy.io", "AdPushup": "adpushup.com"}
	for i, user := range resp {
		assert.Contains(t, filteredCompaniesNameHostNameMap, user.Name, user.HostName)
		assert.NotNil(t, user.LastActivity)
		if i > 0 {
			assert.True(t, resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix())
		}
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, user.TableProps[prop])
		}
	}

	// 5. Accounts from All Sources (filters applied)

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_name",
					Operator:  "equals",
					Value:     "Adshup",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_name",
					Operator:  "equals",
					Value:     "Adapt.IO",
					LogicalOp: "OR",
				},
			},
			Source: "All",
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)

	timelinesConfig = &model.TimelinesConfig{
		AccountConfig: model.AccountConfig{
			TableProps: []string{"$salesforce_account_billingcountry", "$hubspot_company_country", U.SIX_SIGNAL_COUNTRY, "$salesforce_city", "$hubspot_company_is_public"},
		},
	}

	tlConfigEncoded, err = U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	// 6. Accounts from All Sources (filters applied)

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "notEqual",
					Value:     "Pakistan",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_name",
					Operator:  "equals",
					Value:     "Adapt.IO",
					LogicalOp: "AND",
				},
			},
			Source: "All",
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)

	// 7. Accounts from All Sources (filters applied)

	payload = model.TimelinePayload{
		Query: model.Query{
			Source: "All",
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_name",
					Operator:  "equals",
					Value:     "Adapt.IO",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "Pakistan",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "Germany",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  "equals",
					Value:     "50",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  "equals",
					Value:     "150",
					LogicalOp: "OR",
				},
			},
		},
	}

	filteredCompaniesNameHostNameMap = map[string]string{"Adapt.IO": "adapt.io", "o9 Solutions": "o9solutions.com", "GoLinks Reporting": "golinks.io", "Clientjoy Ads": "clientjoy.io", "Cin7": "cin7.com", "Repair Desk": "repairdesk.co", "AdPushup": "adpushup.com", "Mad Street Den": "madstreetden.com", "Heyflow": "heyflow.app"}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	assert.Equal(t, resp[0].Identity, domID)
	assert.NotNil(t, resp[0].LastActivity)
	assert.Contains(t, filteredCompaniesNameHostNameMap, resp[0].Name)
	assert.NotNil(t, resp[0].HostName)
	assert.Equal(t, resp[0].TableProps["$salesforce_city"], "New Delhi")
	assert.Equal(t, resp[0].TableProps["$hubspot_company_is_public"], "true")

	// Testing base64 conversion
	hostString, err := memsql.ConvertDomainIdToHostName("dom-Ni1wcm8tY2FwaXRhLmNvbQ==")
	assert.Nil(t, err)
	assert.Equal(t, hostString, "pro-capita.com")
}

func sendGetProfileAccountRequest(r *gin.Engine, projectId int64, agent *model.Agent, payload model.TimelinePayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/profiles/accounts?score=true&debug=true", projectId)).
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

func TestAPIGetProfileAccountDetailsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.NotNil(t, agent)
	assert.Nil(t, err)

	var timelinesConfig model.TimelinesConfig

	timelinesConfig.AccountConfig.LeftpaneProps = []string{"$hubspot_company_industry", "$hubspot_company_country"}
	timelinesConfig.AccountConfig.UserProp = "$hubspot_contact_jobtitle"
	timelinesConfig.UserConfig.Milestones = []string{"$milesone_1", "$milesone_2", "$milesone_3"}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	props := map[string]interface{}{
		"$company":                                "Freshworks",
		U.GP_HUBSPOT_COMPANY_NAME:                 "Freshworks-HS",
		U.GP_SALESFORCE_ACCOUNT_NAME:              "Freshworks-SF",
		U.GP_SALESFORCE_ACCOUNT_WEBSITE:           "freshworks.com",
		U.GP_HUBSPOT_COMPANY_DOMAIN:               "google.com",
		U.GP_HUBSPOT_COMPANY_COUNTRY:              "India",
		U.GP_SALESFORCE_ACCOUNT_BILLINGCOUNTRY:    "India",
		U.GP_HUBSPOT_COMPANY_INDUSTRY:             "Freshworks-HS",
		U.GP_SALESFORCE_ACCOUNT_INDUSTRY:          "Freshworks-SF",
		U.GP_HUBSPOT_COMPANY_NUMBEROFEMPLOYEES:    "",
		U.GP_SALESFORCE_ACCOUNT_NUMBEROFEMPLOYEES: "",
		"$milesone_1":                             U.UnixTimeBeforeDuration(1 * time.Hour),
		"$milesone_2":                             U.UnixTimeBeforeDuration(2 * time.Hour),
		"$milesone_3":                             U.UnixTimeBeforeDuration(3 * time.Hour),
		"$milesone_4":                             U.UnixTimeBeforeDuration(4 * time.Hour),
		"$milesone_5":                             U.UnixTimeBeforeDuration(5 * time.Hour),
	}
	propertiesJSON, err := json.Marshal(props)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}

	isGroupUser := true
	customerEmail := "abc@example.com"

	// create a domain
	createdDomainUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Source:      model.GetRequestSourcePointer(model.UserSourceDomains),
		Group1ID:    "1",
		IsGroupUser: &isGroupUser,
		Properties: postgres.Jsonb{
			RawMessage: json.RawMessage(`{}`)},
	})
	domainUser, errCode := store.GetStore().GetUser(project.ID, createdDomainUserID)
	assert.Equal(t, createdDomainUserID, domainUser.ID)
	assert.Equal(t, http.StatusFound, errCode)
	group1, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)

	// account associated to domain
	createdUserID1, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
		Group1ID:       "1",
		Group2ID:       "2",
		Group1UserID:   domainUser.ID,
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &isGroupUser,
	})
	projectID := project.ID
	user, errCode := store.GetStore().GetUser(projectID, createdUserID1)
	assert.Equal(t, user.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	group2, status := store.GetStore().CreateGroup(projectID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group2)

	// create another domain account
	createdDomainUserID2, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Source:      model.GetRequestSourcePointer(model.UserSourceDomains),
		Group1ID:    "chargebee.com",
		IsGroupUser: &isGroupUser,
		Properties: postgres.Jsonb{
			RawMessage: json.RawMessage(`{}`)},
	})
	domainUser2, errCode := store.GetStore().GetUser(project.ID, createdDomainUserID2)
	assert.Equal(t, createdDomainUserID2, domainUser2.ID)
	assert.Equal(t, http.StatusFound, errCode)

	// Hubspot Group Events
	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	trackPayload := SDK.TrackPayload{
		UserId:        createdUserID1,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceHubspot,
	}
	status, response := SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	trackPayload = SDK.TrackPayload{
		UserId:        createdUserID1,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceHubspot,
	}
	status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	// account associated to domain
	createdUserID2, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
		Group1ID:       "1",
		Group3ID:       "3",
		Group1UserID:   domainUser.ID,
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &isGroupUser,
	})
	user2, errCode := store.GetStore().GetUser(projectID, createdUserID2)
	assert.Equal(t, user2.ID, createdUserID2)
	assert.Equal(t, http.StatusFound, errCode)
	group3, status := store.GetStore().CreateGroup(projectID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group3)

	// Salesforce Group Events
	trackPayload = SDK.TrackPayload{
		UserId:        createdUserID2,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceSalesforce,
	}
	status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	trackPayload = SDK.TrackPayload{
		UserId:        createdUserID2,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceSalesforce,
	}
	status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	// 10  Associated Users
	// m := map[string]interface{}{U.UP_NAME: "Some Name"}
	// userProps, err := json.Marshal(m)
	// if err != nil {
	// 	log.WithError(err).Fatal("Marshal error.")
	// }
	// properties = postgres.Jsonb{RawMessage: userProps}
	// event properties map
	eventProperties := map[string]interface{}{
		U.EP_PAGE_COUNT:                              5,
		U.EP_CHANNEL:                                 "ChannelName",
		U.EP_CAMPAIGN:                                "CampaignName",
		U.SP_SPENT_TIME:                              120,
		U.EP_REFERRER_URL:                            RandomURL(),
		U.EP_FORM_NAME:                               "Form Name",
		U.EP_PAGE_URL:                                RandomURL(),
		U.EP_SALESFORCE_CAMPAIGN_TYPE:                "Some Type",
		U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:        "CurrentStatus",
		U.EP_HUBSPOT_ENGAGEMENT_SOURCE:               "Some Engagement Source",
		U.EP_HUBSPOT_ENGAGEMENT_FROM:                 "Somewhere",
		U.EP_HUBSPOT_ENGAGEMENT_TYPE:                 "Some Engagement Type",
		U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME:       "Some Outcome",
		U.EP_HUBSPOT_ENGAGEMENT_STARTTIME:            "Start time",
		U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS: 10000000000,
		U.EP_HUBSPOT_ENGAGEMENT_STATUS:               "Testing",
		U.EP_HUBSPOT_FORM_SUBMISSION_FORMTYPE:        "Some HS Form Submission Type",
		U.EP_HUBSPOT_FORM_SUBMISSION_PAGEURL:         RandomURL(),
		U.EP_HUBSPOT_ENGAGEMENT_ENDTIME:              "End Time",
		U.EP_SALESFORCE_CAMPAIGN_NAME:                "Some Salesforce Campaign Name",
		U.EP_HUBSPOT_FORM_SUBMISSION_TITLE:           "Some form submission title",
		U.EP_HUBSPOT_ENGAGEMENT_SUBJECT:              "Some Engagement Subject",
		U.EP_HUBSPOT_ENGAGEMENT_TITLE:                "Some Engagement Title",
		U.EP_SF_TASK_TYPE:                            "Some Task Type",
		U.EP_SF_TASK_SUBTYPE:                         "Some Task SubType",
		U.EP_SF_TASK_COMPLETED_DATETIME:              1660875887,
		U.EP_SF_EVENT_TYPE:                           "Some Event Type",
		U.EP_SF_EVENT_SUBTYPE:                        "Some Event Subtype",
		U.EP_SF_EVENT_COMPLETED_DATETIME:             1660875887,
	}
	randomURL := RandomURL()
	customerEmail = "@example.com"
	isGroupUser = false
	users := make([]model.User, 0)
	numUsers := 13
	for i := 1; i <= numUsers; i++ {

		jobTitle := "Boss"
		if i > 1 {
			jobTitle = "Employee"
		}
		userProps := map[string]interface{}{
			"$hubspot_contact_jobtitle": jobTitle,
			U.UP_TOTAL_SPENT_TIME:       100,
		}
		userPropsJSON, err := json.Marshal(userProps)
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		userPropsEncoded := postgres.Jsonb{RawMessage: userPropsJSON}
		var customerUserID string
		if i < 6 || i > 10 {
			customerUserID = "user" + strconv.Itoa(i) + customerEmail
		}
		if i == 6 {
			customerUserID = "user5" + customerEmail
		}

		var associatedUserId string

		if i > 10 {
			// users associated to domain2
			customerEmail = "@domain2.com"
			notGroupUser := false
			userProps := map[string]interface{}{
				"$page_count": i * 10, "$company": "ChargeBee", U.UP_TOTAL_SPENT_TIME: 100,
			}
			userPropsJSON, err := json.Marshal(userProps)
			if err != nil {
				log.WithError(err).Fatal("Marshal error.")
			}
			userPropsEncoded := postgres.Jsonb{RawMessage: userPropsJSON}
			associatedUserId, _ = store.GetStore().CreateUser(&model.User{
				ProjectId:      project.ID,
				Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
				Group1UserID:   domainUser2.ID,
				IsGroupUser:    &notGroupUser,
				Properties:     userPropsEncoded,
				CustomerUserId: fmt.Sprintf("user%d", i) + customerEmail,
			})
			user, errCode = store.GetStore().GetUser(project.ID, associatedUserId)
			assert.Equal(t, associatedUserId, user.ID)
			assert.Equal(t, http.StatusFound, errCode)
		} else {
			associatedUserId, _ = store.GetStore().CreateUser(&model.User{
				ProjectId:      projectID,
				Properties:     userPropsEncoded,
				IsGroupUser:    &isGroupUser,
				Group2ID:       "2",
				Group2UserID:   createdUserID1,
				CustomerUserId: customerUserID,
				Group1UserID:   domainUser.ID,
				Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
			})

			user, errCode = store.GetStore().GetUser(project.ID, associatedUserId)
			assert.Equal(t, http.StatusFound, errCode)
			users = append(users, *user)
		}

		// Event 1 : Page View
		timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:          associatedUserId,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            randomURL,
			EventProperties: map[string]interface{}{},
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			ProjectId:       project.ID,
			Auto:            true,
			RequestSource:   model.UserSourceWeb,
		}
		status, response := SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)

		// Event 2 : Web Session
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SESSION,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceWeb,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 3 : Form Submit
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_FORM_SUBMITTED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceWeb,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 4 : Offline Touchpoint
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceWeb,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 5 : Campaign Member Created
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 6 : Campaign Member Responded to Campaign
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 7 : Hubspot Form Submission
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 8 : Engagement Email
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 9 : Engagement Meeting Created
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 10 : Engagement Call Created
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceHubspot,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 11 : Salesforce Task Created
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SALESFORCE_TASK_CREATED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 11 : Salesforce Task Created
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SALESFORCE_EVENT_CREATED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(projectID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 13 : Random Event
		timestamp = timestamp - 10000
		randomProperties := map[string]interface{}{}
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
			EventProperties: randomProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceHubspot,
		}
		status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

	}
	assert.Equal(t, len(users), 10)

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, projectID, agent, domainUser.ID, "All")
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Contains(t, resp.Name, "Freshworks")
		assert.Equal(t, resp.HostName, "google.com")
		assert.Equal(t, len(resp.AccountTimeline) > 0, true)
		assert.Equal(t, len(resp.AccountTimeline), 10)
		assert.NotNil(t, resp.LeftPaneProps)
		for i, property := range resp.LeftPaneProps {
			assert.Equal(t, props[i], property)
		}
		for i, property := range resp.Milestones {
			assert.Equal(t, props[i], property)
		}
		for _, userTimeline := range resp.AccountTimeline {
			if userTimeline.UserName == model.GROUP_ACTIVITY_USERNAME {
				assert.Equal(t, userTimeline.AdditionalProp, "All")
				assert.Equal(t, userTimeline.IsAnonymous, false)
				assert.Equal(t, len(userTimeline.UserActivities), 2)
			}
		}

		//Top Users
		assert.Len(t, resp.Overview.TopUsers, 6)
		expectedNames := []string{"user5@example.com", "user4@example.com", "user2@example.com", "user3@example.com", "user1@example.com"}
		expectedPageViews := []int{2, 1, 1, 1, 1}
		expectedActiveTime := []int{400, 100, 100, 100, 100}
		expectedNumOfPages := []int{1, 1, 1, 1, 1}

		for i, expectedName := range expectedNames {
			assert.Equal(t, expectedName, resp.Overview.TopUsers[i].Name)
			assert.Equal(t, int64(expectedPageViews[i]), resp.Overview.TopUsers[i].NumPageViews)
			assert.Equal(t, float64(expectedActiveTime[i]), resp.Overview.TopUsers[i].ActiveTime)
			assert.Equal(t, int64(expectedNumOfPages[i]), resp.Overview.TopUsers[i].NumOfPages)
		}
		//Anonymous User
		assert.Equal(t, "4 Anonymous Users", resp.Overview.TopUsers[5].Name)
		assert.Equal(t, int64(4), resp.Overview.TopUsers[5].NumPageViews)
		assert.Equal(t, float64(400), resp.Overview.TopUsers[5].ActiveTime)
		assert.Equal(t, int64(1), resp.Overview.TopUsers[5].NumOfPages)

		//Top Pages
		assert.Len(t, resp.Overview.TopPages, 1)
		assert.Equal(t, "", resp.Overview.TopPages[0].PageUrl)
		assert.Equal(t, int64(10), resp.Overview.TopPages[0].Views)
		assert.Equal(t, int64(9), resp.Overview.TopPages[0].UsersCount)
		assert.Equal(t, float64(10), resp.Overview.TopPages[0].TotalTime)
		assert.Equal(t, float64(0), resp.Overview.TopPages[0].AvgScrollPercent)
	})

	t.Run("Success2", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, projectID, agent, domainUser2.ID, "All")
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Contains(t, resp.Name, "chargebee")
		assert.Equal(t, resp.HostName, "chargebee.com")
		assert.Equal(t, len(resp.AccountTimeline) > 0, true)
		assert.Equal(t, len(resp.AccountTimeline), 4)
		assert.Equal(t, resp.Overview.UsersCount, int64(len(resp.AccountTimeline)-1))
		assert.Equal(t, resp.Overview.TimeActive, float64((len(resp.AccountTimeline)-1)*100))
		for _, userTimeline := range resp.AccountTimeline {
			if userTimeline.UserName != model.GROUP_ACTIVITY_USERNAME {
				assert.Equal(t, userTimeline.IsAnonymous, false)
				assert.Equal(t, len(userTimeline.UserActivities), 13)
			}
		}
	})

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, projectID, agent, createdUserID1, model.GROUP_NAME_HUBSPOT_COMPANY)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Contains(t, resp.Name, "Freshworks")
		assert.Equal(t, resp.HostName, "google.com")
		assert.Equal(t, len(resp.AccountTimeline), 10)
		assert.Equal(t, resp.Overview.UsersCount, int64(len(resp.AccountTimeline)-1))
		assert.Equal(t, resp.Overview.TimeActive, float64((len(resp.AccountTimeline))*100))
		assert.NotNil(t, resp.LeftPaneProps)
		for i, property := range resp.LeftPaneProps {
			assert.Equal(t, props[i], property)
		}
		for i, property := range resp.Milestones {
			assert.Equal(t, props[i], property)
		}

		assert.True(t, len(resp.AccountTimeline) > 0)

		// Loop through the AccountTimeline and perform assertions on each User Timeline
		for index, userTimeline := range resp.AccountTimeline {
			assert.NotNil(t, userTimeline.UserId)
			assert.NotNil(t, userTimeline.UserName)

			// Separate check the 10th element (Intent Activity)
			if index == 9 {
				assert.Equal(t, userTimeline.UserName, model.GROUP_ACTIVITY_USERNAME)
				assert.Equal(t, userTimeline.AdditionalProp, U.STANDARD_GROUP_DISPLAY_NAMES[model.GROUP_NAME_HUBSPOT_COMPANY])
				assert.Equal(t, userTimeline.IsAnonymous, false)
				assert.Equal(t, len(userTimeline.UserActivities), 1)
				continue
			}

			// Loop through UserActivities and perform assertions
			for _, activity := range userTimeline.UserActivities {
				assert.NotNil(t, activity.EventName)
				assert.NotNil(t, activity.DisplayName)
				assert.NotNil(t, activity.Timestamp)
				assert.True(t, activity.Timestamp <= uint64(time.Now().UTC().Unix()))

				eventFromMap, eventExistsInMap := model.HOVER_EVENTS_NAME_PROPERTY_MAP[activity.EventName]
				if activity.EventName == randomURL {
					assert.Equal(t, activity.DisplayName, "Page View")
					assert.Equal(t, activity.AliasName, "")
				} else if eventExistsInMap {
					assert.Equal(t, activity.DisplayName, U.STANDARD_EVENTS_DISPLAY_NAMES[activity.EventName])
					// Check alias name based on event name
					switch activity.EventName {
					case U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("Added to %s", eventProperties[U.EP_SALESFORCE_CAMPAIGN_NAME]))
					case U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("Responded to %s", eventProperties[U.EP_SALESFORCE_CAMPAIGN_NAME]))
					case U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_FORM_SUBMISSION_TITLE]))
					case U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("%s: %s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TYPE], eventProperties[U.EP_HUBSPOT_ENGAGEMENT_SUBJECT]))
					case U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED, U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TITLE]))
					case U.EVENT_NAME_SALESFORCE_TASK_CREATED:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("Created Task - %s", eventProperties[U.EP_SF_TASK_SUBJECT]))
					case U.EVENT_NAME_SALESFORCE_EVENT_CREATED:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("Created Event - %s", eventProperties[U.EP_SF_EVENT_SUBJECT]))
					case U.EVENT_NAME_HUBSPOT_CONTACT_LIST:
						assert.Equal(t, activity.AliasName, fmt.Sprintf("Added to Hubspot List - %s", eventProperties[U.EP_HUBSPOT_CONTACT_LIST_LIST_NAME]))
					}
				}

				if activity.DisplayName == "Page View" || eventExistsInMap {
					var lookInProps []string
					if activity.DisplayName == "Page View" {
						lookInProps = model.PAGE_VIEW_HOVERPROPS_LIST
					} else if eventExistsInMap {
						lookInProps = eventFromMap
					}
					assert.NotNil(t, activity.Properties)
					properties, err := U.DecodePostgresJsonb(activity.Properties)
					assert.Nil(t, err)
					for key := range *properties {
						sort.Strings(lookInProps)
						i := sort.SearchStrings(lookInProps, key)
						assert.True(t, i < len(lookInProps))
					}
				}
			}
		}
	})
}

func sendGetProfileAccountDetailsRequest(r *gin.Engine, projectId int64, agent *model.Agent, id string, group string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/accounts/%s/%s", projectId, group, id)).
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

func TestSegmentEventAnalyticsQuery(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var createdUserID string
	// Properties Map
	propsMap := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 105, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 120, "$session_spent_time": 2000},
		{"$browser": "Firefox", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 100, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2500},
		{"$browser": "Edge", "$city": "DC", "$country": "US", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 150, "$session_spent_time": 2100},
		{"$browser": "Chrome", "$city": "UP", "$country": "India", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Brave", "$city": "Paris", "$country": "France", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 3000},
		{"$browser": "Chrome", "$city": "Paris", "$country": "France", "$device_type": "desktop", "$page_count": 110, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Cannes", "$country": "France", "$device_type": "iPad", "$page_count": 150, "$session_spent_time": 2500},
		{"$browser": "Firefox", "$city": "Dubai", "$country": "UAE", "$device_type": "desktop", "$page_count": 150, "$session_spent_time": 2100},
		{"$browser": "Chrome", "$city": "Abu Dhabi", "$country": "UAE", "$device_type": "tablet", "$page_count": 110, "$session_spent_time": 2200},
		{"$browser": "Firefox", "$city": "Dubai", "$country": "UAE", "$device_type": "tablet", "$page_count": 120, "$session_spent_time": 2800},
	}
	// Create 15 Users
	users := make([]model.User, 0)
	numUsers := 15
	var randomURLs []string
	for i := 0; i < numUsers; i++ {
		randomURLs = append(randomURLs, RandomURL())
	}
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		var src *int
		if i%2 == 0 {
			src = model.GetRequestSourcePointer(model.UserSourceSalesforce)
		} else {
			src = model.GetRequestSourcePointer(model.UserSourceWeb)
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ = store.GetStore().CreateUser(&model.User{
			ProjectId:  project.ID,
			Source:     src,
			Properties: properties,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)

		timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
		//randomURL := RandomURL()
		eventProperties := []map[string]interface{}{
			{
				U.EP_PAGE_COUNT:                              5,
				U.EP_CHANNEL:                                 "ChannelName",
				U.EP_CAMPAIGN:                                "CampaignName",
				U.SP_SPENT_TIME:                              120,
				U.EP_REFERRER_URL:                            RandomURL(),
				U.EP_FORM_NAME:                               "Form Name",
				U.EP_PAGE_URL:                                RandomURL(),
				U.EP_SALESFORCE_CAMPAIGN_TYPE:                "Some Type",
				U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:        "CurrentStatus",
				U.EP_HUBSPOT_ENGAGEMENT_SOURCE:               "Some Engagement Source",
				U.EP_HUBSPOT_ENGAGEMENT_TYPE:                 "Some Engagement Type",
				U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME:       "Some Outcome",
				U.EP_HUBSPOT_ENGAGEMENT_STARTTIME:            "Start time",
				U.EP_HUBSPOT_ENGAGEMENT_DURATIONMILLISECONDS: 10000000000,
				U.EP_HUBSPOT_FORM_SUBMISSION_FORMTYPE:        "Some HS Form Submission Type",
				U.EP_HUBSPOT_FORM_SUBMISSION_PAGEURL:         RandomURL(),
				U.EP_HUBSPOT_ENGAGEMENT_ENDTIME:              "End Time",
				U.EP_SALESFORCE_CAMPAIGN_NAME:                "Some Salesforce Campaign Name",
				U.EP_HUBSPOT_FORM_SUBMISSION_TITLE:           "Some form submission title",
				U.EP_HUBSPOT_ENGAGEMENT_SUBJECT:              "Some Engagement Subject",
				U.EP_HUBSPOT_ENGAGEMENT_TITLE:                "Some Engagement Title",
			},
		}
		val := i + (i % 2)
		trackPayload := SDK.TrackPayload{
			UserId:          createdUserID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            randomURLs[val],
			EventProperties: eventProperties[0],
			UserProperties:  propsMap[i],
			Timestamp:       timestamp,
			ProjectId:       project.ID,
			Auto:            false,
			RequestSource:   *src,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)
	}
	assert.Equal(t, len(users), 15)

	var payload model.TimelinePayload

	// query with only global properties
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "UK",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
				{
					Type:      "categorical",
					Property:  "$device_type",
					Operator:  "equals",
					Value:     "iPad",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
			},
			Source:     "salesforce",
			TableProps: []string{"$country", "$page_count"},
		},
	}

	w := sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	resp := make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	for _, profile := range resp {
		assert.Equal(t, "UK", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[2],
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Source:          "web",
			TableProps:      []string{"$country", "$page_count"},
		},
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	for _, profile := range resp {
		assert.Equal(t, "UK", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// with EWP
	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[1],
				},
				{
					Name: randomURLs[4],
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGN_TYPE,
							Operator:  "equals",
							Value:     "Some Type",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Source:          "web",
			TableProps:      []string{"$country", "$page_count"},
		},
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// EWP with GUP
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[6],
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGN_TYPE,
							Operator:  "equals",
							Value:     "Some Type",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: randomURLs[8],
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_FORM_NAME,
							Operator:  "equals",
							Value:     "Form Name",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: randomURLs[7],
				},
				{
					Name: randomURLs[6],
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,

			Source:     "salesforce",
			TableProps: []string{"$country", "$page_count"},
		},
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	for _, profile := range resp {
		assert.Equal(t, "India", profile.TableProps["$country"])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "France",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
				{
					Type:      "categorical",
					Property:  "$device_type",
					Operator:  "equals",
					Value:     "desktop",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[10],
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGN_TYPE,
							Operator:  "equals",
							Value:     "Some Type",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: randomURLs[10],
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_FORM_NAME,
							Operator:  "equals",
							Value:     "Form Name",
							LogicalOp: "OR",
							Entity:    "event",
						},
					},
				},
				{
					Name: randomURLs[10],
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_CHANNEL,
							Operator:  "equals",
							Value:     "ChannelName",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: randomURLs[10],
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
			Source:          "salesforce",
			TableProps:      []string{"$country", "$page_count"},
		},
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse1, _ := ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse1, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	for _, profile := range resp {
		assert.Equal(t, "France", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// ACCOUNT FILTERS

	// creating domain group
	group, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)

	// creating 2 more groups
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	hbMeetingTime := time.Now().AddDate(0, 0, -5).Unix()
	hbMeetingTimeNow := time.Now().Unix()
	propertiesMap := []map[string]interface{}{
		{"$salesforce_account_name": "Pepper Content", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "peppercontent.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target"},
		{"$salesforce_account_name": "o9 Solutions", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "o9solutions.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "GoLinks Reporting", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "golinks.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "Cin7", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "cin7.com", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor"},
		{"$salesforce_account_name": "Repair Desk", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "repairdesk.co", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer"},
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$country": "US", "$hubspot_contact_rh_meeting_time": hbMeetingTimeNow},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$country": "US", "$hubspot_contact_rh_meeting_time": hbMeetingTime},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$country": "US"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$country": "US", "$hubspot_contact_rh_meeting_time": hbMeetingTimeNow},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$country": "US", "$hubspot_contact_rh_meeting_time": hbMeetingTime},
	}

	// creating domain account

	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
	domSource := model.GetRequestSourcePointer(model.UserSourceDomains)
	groupUser := true
	accounts := make([]model.User, 0)

	domID1, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         domSource,
		Group1ID:       "1",
		CustomerUserId: "domainuser",
		Properties:     domProperties,
		IsGroupUser:    &groupUser,
	})
	_, errCode := store.GetStore().GetUser(project.ID, domID1)
	assert.Equal(t, http.StatusFound, errCode)

	domID2, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         domSource,
		Group1ID:       "1",
		CustomerUserId: "domainuser2",
		Properties:     domProperties,
		IsGroupUser:    &groupUser,
	})
	_, errCode = store.GetStore().GetUser(project.ID, domID2)
	assert.Equal(t, http.StatusFound, errCode)

	domID3, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         domSource,
		Group1ID:       "1",
		CustomerUserId: "domainuser3",
		Properties:     domProperties,
		IsGroupUser:    &groupUser,
	})
	_, errCode = store.GetStore().GetUser(project.ID, domID3)
	assert.Equal(t, http.StatusFound, errCode)

	domArray := []string{domID1, domID2, domID3}
	// Create 5 Salesforce Accounts
	numUsers = 5
	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propertiesMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSalesforce)

		domID := domArray[(i % 3)]
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group3ID:       "3",
			Group1ID:       "1",
			Group1UserID:   domID,
			CustomerUserId: fmt.Sprintf("sfuser%d@%s", i+1, propertiesMap[i]["$salesforce_account_website"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		eventProperties := map[string]interface{}{
			U.EP_CHANNEL:                           "ChannelName1",
			U.EP_CAMPAIGN:                          "CampaignName1",
			U.EP_FORM_NAME:                         "Form Name For Accountts",
			U.EP_SALESFORCE_CAMPAIGN_TYPE:          "Some Salesforce Type",
			U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:  "CurrentStatusNow",
			U.EP_HUBSPOT_ENGAGEMENT_SOURCE:         "Some Engagement Source For Accounts",
			U.EP_HUBSPOT_ENGAGEMENT_TYPE:           "Some Engagement Type",
			U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME: "Some Outcome",
			U.EP_HUBSPOT_ENGAGEMENT_STARTTIME:      "Start time",
		}
		trackPayload := SDK.TrackPayload{
			UserId:          createdUserID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
			EventProperties: eventProperties,
			UserProperties:  propertiesMap[i+5],
			Timestamp:       timestamp,
			ProjectId:       project.ID,
			Auto:            false,
			RequestSource:   *source,
		}

		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)
	}

	// Create 5 Hubspot Companies
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propertiesMap[i+5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceHubspot)

		domID := domArray[(i % 3)]
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group2ID:       "2",
			Group1ID:       "1",
			Group1UserID:   domID,
			CustomerUserId: fmt.Sprintf("hsuser%d@%s", i+1, propertiesMap[i+5]["$hubspot_company_domain"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		eventProperties := map[string]interface{}{
			U.EP_CHANNEL:                           "ChannelName1",
			U.EP_CAMPAIGN:                          "CampaignName1",
			U.EP_FORM_NAME:                         "Form Name For Accountts",
			U.EP_SALESFORCE_CAMPAIGN_TYPE:          "Some Salesforce Type",
			U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:  "CurrentStatusNow",
			U.EP_HUBSPOT_ENGAGEMENT_SOURCE:         "Some Engagement Source For Accounts",
			U.EP_HUBSPOT_ENGAGEMENT_TYPE:           "Some Engagement Type",
			U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME: "Some Outcome",
			U.EP_HUBSPOT_ENGAGEMENT_STARTTIME:      "Start time",
		}
		trackPayload := SDK.TrackPayload{
			UserId:          createdUserID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
			EventProperties: eventProperties,
			UserProperties:  propertiesMap[i+5],
			Timestamp:       timestamp,
			ProjectId:       project.ID,
			Auto:            false,
			RequestSource:   *source,
		}

		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)

		if i >= 2 {
			continue
		}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Properties:     properties,
			Group2UserID:   createdUserID,
			CustomerUserId: fmt.Sprintf("hubspot@%duser", (i%10)+1),
		})

		user1, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Equal(t, createdUserID1, user1.ID)

		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user1.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)
	}
	assert.Equal(t, len(accounts), 10)

	// Test Cases :-
	//1. gpb and ewp
	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "US",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  "equals",
					Value:     "50",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  "equals",
					Value:     "20",
					LogicalOp: "OR",
				},
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGN_TYPE,
							Operator:  "equals",
							Value:     "Some Salesforce Type",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_CHANNEL,
							Operator:  "equals",
							Value:     "ChannelName1",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	for _, profile := range resp {
		assert.Equal(t, "US", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGN_TYPE,
							Operator:  "equals",
							Value:     "Some Salesforce Type",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_CHANNEL,
							Operator:  "equals",
							Value:     "ChannelName1",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 0)
	for _, profile := range resp {
		assert.Equal(t, "US", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
				{
					Entity:    "user_g",
					Type:      "datetime",
					Property:  "$hubspot_contact_rh_meeting_time",
					Operator:  "inLast",
					Value:     "{\"num\":7,\"gran\":\"days\"}",
					LogicalOp: "AND",
				},
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_FORM_NAME,
							Operator:  "equals",
							Value:     "Form Name For Accountts",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	for _, profile := range resp {
		assert.Equal(t, "US", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGN_TYPE,
							Operator:  "equals",
							Value:     "Some Salesforce Type",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_CHANNEL,
							Operator:  "equals",
							Value:     "ChannelName1",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 5)
	for _, profile := range resp {
		assert.Equal(t, "US", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
				},
			},

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
			GroupAnalysis:   "$hubspot_company",
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	for _, profile := range resp {
		assert.Equal(t, "US", profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
		assert.NotNil(t, profile.Name)
		assert.NotNil(t, profile.HostName)
	}
}

func TestSegmentSupportEventAnalyticsQuery(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var createdUserID string
	// Properties Map
	propsMap := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 105, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 120, "$session_spent_time": 2000},
		{"$browser": "Firefox", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 100, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2500},
		{"$browser": "Edge", "$city": "DC", "$country": "US", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 150, "$session_spent_time": 2100},
		{"$browser": "Chrome", "$city": "UP", "$country": "India", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Brave", "$city": "Paris", "$country": "France", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 3000},
	}
	eventProperties := map[string]interface{}{
		U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:  "CurrentStatus",
		U.EP_HUBSPOT_ENGAGEMENT_FROM:           "Somewhere",
		U.EP_HUBSPOT_ENGAGEMENT_TYPE:           "Some Engagement Type",
		U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME: "Some Outcome",
		U.EP_SALESFORCE_CAMPAIGN_NAME:          "Some Salesforce Campaign Name",
	}

	// groups creation
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)

	// 5 hubspot accounts
	var accounts []string
	groupUser := true
	companyNames := []string{"FreshWorks", "CleverTap", "Adsup", "ChargeBee", "Heyflow"}
	cities := []string{"London", "London", "DC", "Delhi", "Paris"}
	for i := 0; i < 5; i++ {
		props := map[string]interface{}{
			U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED: companyNames[i],
			"$city": cities[i],
		}
		propertiesJSON, err := json.Marshal(props)
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		accProps := postgres.Jsonb{RawMessage: propertiesJSON}
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		source := model.GetRequestSourcePointer(model.UserSourceHubspot)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:   project.ID,
			Source:      source,
			Group1ID:    fmt.Sprintf("%d", group1.ID),
			Properties:  accProps,
			IsGroupUser: &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, account.ID)

		timestamp := U.UnixTimeBeforeDuration(time.Duration(1+i) * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:          account.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
			Timestamp:       timestamp,
			ProjectId:       project.ID,
			Auto:            false,
			RequestSource:   model.UserSourceHubspot,
			EventProperties: props,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)
	}

	// 5 salesforce accounts
	for i := 0; i < 5; i++ {
		props := map[string]interface{}{
			U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED: companyNames[i],
			"$city": cities[i],
		}
		propertiesJSON, err := json.Marshal(props)
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		accProps := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSalesforce)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:   project.ID,
			Source:      source,
			Group2ID:    fmt.Sprintf("%d", group2.ID),
			Properties:  accProps,
			IsGroupUser: &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, account.ID)

		timestamp := U.UnixTimeBeforeDuration(time.Duration(1+i) * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:        account.ID,
			CreateUser:    false,
			IsNewUser:     false,
			Name:          U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
			Timestamp:     timestamp,
			ProjectId:     project.ID,
			Auto:          false,
			RequestSource: model.UserSourceSalesforce,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)
	}

	// Create 20 Users
	users := make([]model.User, 0)
	numUsers := 20
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[i%10])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		var src *int
		if i > 10 {
			src = model.GetRequestSourcePointer(model.UserSourceHubspot)
			createdUserID, _ = store.GetStore().CreateUser(&model.User{
				ProjectId:      project.ID,
				Source:         src,
				Properties:     properties,
				Group1UserID:   accounts[i%5],
				Group1ID:       "1",
				CustomerUserId: fmt.Sprintf("hubspot@%daccount", (i%10)+1),
			})
		} else {
			src = model.GetRequestSourcePointer(model.UserSourceSalesforce)
			createdUserID, _ = store.GetStore().CreateUser(&model.User{
				ProjectId:      project.ID,
				Source:         src,
				Properties:     properties,
				Group2ID:       "2",
				Group2UserID:   accounts[5+i%5],
				CustomerUserId: fmt.Sprintf("salesforce@%daccount", (i%10)+1),
			})
		}
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)

		// Event 1 : Campaign Member Created
		timestamp := U.UnixTimeBeforeDuration(time.Duration(1+i) * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Event 2 : Engagement Email
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
		}
		status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)

		// Website session
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SESSION,
			EventProperties: eventProperties,
			UserProperties:  map[string]interface{}{},
			Timestamp:       timestamp,
			Auto:            false,
			RequestSource:   model.UserSourceHubspot,
		}
		status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotNil(t, response.EventId)
		assert.Empty(t, response.UserId)
		assert.Equal(t, http.StatusOK, status)
	}
	assert.Equal(t, len(users), 20)

	// user segments
	payload := model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS,
							Operator:  "equals",
							Value:     "CurrentStatus",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			GroupAnalysis:   "users",
			TableProps:      []string{"$country", "$page_count"},
			Source:          "salesforce",
		},
	}

	w := sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	resp := make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 10)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// user segment with ewp and gup
	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS,
							Operator:  "equals",
							Value:     "CurrentStatus",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$page_count",
					Operator:  "equals",
					Value:     "105",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$page_count",
					Operator:  "equals",
					Value:     "100",
					LogicalOp: "OR",
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Source:          "salesforce",
			GroupAnalysis:   "users",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// account segment with ewp and gup
	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
							Operator:  "equals",
							Value:     "FreshWorks",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_HUBSPOT_ENGAGEMENT_FROM,
							Operator:  "equals",
							Value:     "Somewhere",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
			},
			GlobalUserProperties: []model.QueryProperty{
				{
					Type:      "categorical",
					Property:  "$city",
					Operator:  "equals",
					Value:     "London",
					LogicalOp: "AND",
					Entity:    "user_g",
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Source:          model.GROUP_NAME_HUBSPOT_COMPANY,
			TableProps:      []string{"$hubspot_company_created", "$hour_of_first_event"},
			GroupAnalysis:   model.GROUP_NAME_HUBSPOT_COMPANY,
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	for idx, profile := range resp {
		assert.Equal(t, companyNames[1-idx], profile.TableProps["$hubspot_company_created"])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
		assert.NotNil(t, profile.Name)
	}

	// account segment with ewp (user event)

	payload = model.TimelinePayload{
		Query: model.Query{
			GroupAnalysis:   model.GROUP_NAME_HUBSPOT_COMPANY,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: "any_given_event",
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.EVENT_NAME_SESSION,
					GroupAnalysis: "Most Recent",
				},
			},
			Source:     model.GROUP_NAME_HUBSPOT_COMPANY,
			TableProps: []string{"$country", "$hubspot_company_created", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 5)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps["$hubspot_company_created"])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// account segment with only ewp
	payload = model.TimelinePayload{
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
							Operator:  "equals",
							Value:     "FreshWorks",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name: U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_HUBSPOT_ENGAGEMENT_FROM,
							Operator:  "equals",
							Value:     "Somewhere",
							LogicalOp: "AND",
							Entity:    "event",
						},
					},
				},
				{
					Name:          U.EVENT_NAME_SESSION,
					GroupAnalysis: "Most Recent",
				},
			},
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
			Source:          model.GROUP_NAME_HUBSPOT_COMPANY,
			TableProps:      []string{"$country", "$hubspot_company_created"},
			GroupAnalysis:   model.GROUP_NAME_HUBSPOT_COMPANY,
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)

	for _, profile := range resp {
		assert.Equal(t, "FreshWorks", profile.TableProps["$hubspot_company_created"])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// account segment with ewp (user event) and gup

	payload = model.TimelinePayload{
		Query: model.Query{
			GroupAnalysis:   model.GROUP_NAME_HUBSPOT_COMPANY,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: "any_given_event",
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$city",
					Operator:  "equals",
					Value:     "London",
					LogicalOp: "AND",
				},
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.EVENT_NAME_SESSION,
					GroupAnalysis: "Most Recent",
				},
			},
			Source:     model.GROUP_NAME_HUBSPOT_COMPANY,
			TableProps: []string{"$country", "$hubspot_company_created", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps["$hubspot_company_created"])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

}

func TestAllAccountDefaultGroupProperties(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var errCode int
	// Properties Map
	propertiesMap := []map[string]interface{}{
		{"$salesforce_account_name": "Adapt.IO", "$page_count": 4, "$salesforce_account_id": "123", "$salesforce_account_website": "adapt.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "o9 Solutions", "$page_count": 4, "$salesforce_account_id": "123", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "o9solutions.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "GoLinks Reporting", "$page_count": 4, "$salesforce_account_id": "123", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "golinks.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "Cin7", "$page_count": 4, "$salesforce_account_id": "123", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "cin7.com", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor", "$browser": "Chrome", "$device_type": "PC", "$salesforce_city": "New Delhi"},
		{"$salesforce_account_name": "Repair Desk", "$page_count": 5, "$salesforce_account_id": "123", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "repairdesk.co", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "AdPushup", "$page_count": 5, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Mad Street Den", "$page_count": 5, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Heyflow", "$page_count": 4, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC", "$hubspot_company_is_public": "true"},
		{"$hubspot_company_name": "Adapt.IO", "$page_count": 4, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Clientjoy Ads", "$page_count": 4, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", "$page_count": 4, "$salesforce_account_id": "123", U.SIX_SIGNAL_DOMAIN: "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", "$page_count": 4, "$salesforce_account_id": "123", U.SIX_SIGNAL_DOMAIN: "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Heyflow", U.SIX_SIGNAL_COUNTRY: "Germany", "$page_count": 4, "$hubspot_company_hs_object_id": 123, U.SIX_SIGNAL_DOMAIN: "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Clientjoy Ads", U.SIX_SIGNAL_COUNTRY: "India", "$page_count": 4, "$hubspot_company_hs_object_id": 123, U.SIX_SIGNAL_DOMAIN: "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Adapt.IO", U.SIX_SIGNAL_COUNTRY: "India", "$page_count": 4, U.SIX_SIGNAL_DOMAIN: "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
	}

	// Creating domain Account and Group
	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}

	accounts := make([]model.User, 0)

	var payload model.TimelinePayload

	group, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)

	numUsers := 5

	// Create 5 Hubspot Companies
	for i := 0; i < numUsers; i++ {
		groupUser := true
		customerUserId := U.RandomLowerAphaNumString(5)
		domId, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceDomains),
			Group1ID:       "1",
			CustomerUserId: customerUserId,
			Properties:     domProperties,
			IsGroupUser:    &groupUser,
		})

		_, errCode = store.GetStore().GetUser(project.ID, domId)
		assert.Equal(t, http.StatusFound, errCode)

		propertiesJSON, err := json.Marshal(propertiesMap[i+5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceHubspot)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group2ID:       "2",
			Group1ID:       "1",
			Group1UserID:   domId,
			CustomerUserId: fmt.Sprintf("hsuser%d@%s", i+1, propertiesMap[i+5]["$hubspot_company_domain"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// Create 5 Salesforce Accounts

		propertiesJSON, err = json.Marshal(propertiesMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		source = model.GetRequestSourcePointer(model.UserSourceSalesforce)

		createdUserID, _ = store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group3ID:       "3",
			Group1ID:       "1",
			Group1UserID:   domId,
			CustomerUserId: fmt.Sprintf("sfuser%d@%s", i+1, propertiesMap[i]["$salesforce_account_website"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode = store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		groupUser = false
		// Create 5 Six Signal Domains

		propertiesJSON, err = json.Marshal(propertiesMap[i+10])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		source = model.GetRequestSourcePointer(model.UserSourceSixSignal)

		createdUserID, _ = store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group4ID:       "4",
			Group1ID:       "1",
			Group1UserID:   domId,
			CustomerUserId: fmt.Sprintf("6siguser%d@%s", i+1, propertiesMap[i+10][U.SIX_SIGNAL_DOMAIN]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode = store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

	}

	assert.Equal(t, len(accounts), 15)

	// 3 group exists
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.NotNil(t, group3)
	assert.Equal(t, http.StatusCreated, status)

	// test in hubspot properties with single filter
	t.Run("TestForInHubspotProperties", func(t *testing.T) {
		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))

	})
	// test in salesforce properties with single filter
	t.Run("TestForInSalesforceProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), 5)
	})
	// test in Visited website properties with single filter
	t.Run("TestForInVisitedProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$visited_website",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))
	})

	// test in salesforce and in hubspot properties with multiple filter
	t.Run("TestInPropertiesWithValueMultipleFilters", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))
	})
	// test in salesforce and user properties properties with multiple filter
	t.Run("TestInPropertiesWithValueMultipleFiltersWithUserProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "false",
						LogicalOp: "AND",
					},
					{
						Entity:    "user_group",
						Type:      "categorical",
						Property:  "$page_count",
						Operator:  model.GreaterThanOpStr,
						Value:     "0",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))
	})

	// test in salesforce and visited website properties with multiple filter
	t.Run("TestInVisitedWebsitePropertiesWithValueMultipleFilters", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$visited_website",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))
	})

	// test in column properties with single filter

	t.Run("TestInPropertiesWithColumn", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  U.IDENTIFIED_USER_ID,
						Operator:  "notEqual",
						Value:     "$none",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 15, len(resp))
	})
	// test in column properties with multiple filter
	t.Run("TestInPropertiesMultipleFilterWithColumn", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  U.IDENTIFIED_USER_ID,
						Operator:  "notEqual",
						Value:     "$none",
						LogicalOp: "AND",
					},
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))
	})

	// test user properties and visited webite properties with multiple filter

	t.Run("TestInVisitedWebsitePropertiesWithValueMultipleFiltersWithUserProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: "All",
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$device_type",
						Operator:  model.EqualsOp,
						Value:     "PC",
						LogicalOp: "AND",
					},
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$visited_website",
						Operator:  "equals",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(resp))
	})
}
