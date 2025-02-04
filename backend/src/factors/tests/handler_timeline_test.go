package tests

import (
	"encoding/base64"
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"io"
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

func TestTimelineGetProfileUserHandler(t *testing.T) {
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
	lastEventTime := time.Now().Add(time.Duration(-6) * time.Hour)
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propsMap[9-i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:   project.ID,
			Source:      model.GetRequestSourcePointer(model.UserSourceWeb),
			Properties:  properties,
			LastEventAt: &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 5)

	// Create 5 Identified Users from UserSourceWeb
	numUsers = 5
	lastEventTime = time.Now().Add(time.Duration(-5) * time.Hour)
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
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 10)

	// Create 2 Identified Users from UserSourceSalesforce
	numUsers = 2
	lastEventTime = time.Now().Add(time.Duration(-4) * time.Hour)
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
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 12)

	// Create 3 Identified Users from UserSourceHubspot
	numUsers = 3
	lastEventTime = time.Now().Add(time.Duration(-3) * time.Hour)
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
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(users), 15)

	var payload model.TimelinePayload
	payload.SegmentId = ""

	// Test Cases :-
	// 1. Users from Different Sources (No filter, no segment applied)
	sourceToUserCountMap := map[string]int{"All": 15, model.UserSourceSalesforceString: 2, model.UserSourceHubspotString: 3}
	for source, count := range sourceToUserCountMap {
		payload.Query.Source = source
		w := sendGetProfileUserRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := io.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), count)
		for i, user := range resp {
			if model.IsSourceAllUsers(source) {
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
		SegmentId: "",
	}
	w := sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := io.ReadAll(w.Body)
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
		}, SearchFilter: []string{"user2"},
		SegmentId: "",
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
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
		SegmentId: "",
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
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
		SegmentId: "",
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
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
		SegmentId: "",
	}

	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
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
		SegmentId: "",
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
	jsonResponse, _ = io.ReadAll(w.Body)
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

func TestTimelineGetProfileUserDetailsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.NotNil(t, agent)
	assert.Nil(t, err)

	timelinesConfig, err := store.GetStore().GetTimelinesConfig(project.ID)
	assert.Nil(t, err)

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
	lastEventTime := time.Now()

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
		LastEventAt: &lastEventTime,
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
		LastEventAt:    &lastEventTime,
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
		jsonResponse, _ := io.ReadAll(w.Body)
		resp := &model.ContactDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.UserId, userId)
		assert.Contains(t, resp.Name, "Cameron")
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
			eventFromMap, eventExistsInMap := model.TIMELINE_EVENT_PROPERTIES_CONFIG[activity.EventName]
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
					lookInProps = model.TIMELINE_EVENT_PROPERTIES_CONFIG["PageView"]
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

func sendGetProfileAccountRequest(r *gin.Engine, projectId int64, agent *model.Agent, payload model.TimelinePayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/profiles/accounts?download=true&user_marker=true", projectId)).
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

func TestTimelineGetProfileAccountDetailsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.NotNil(t, agent)
	assert.Nil(t, err)

	timelinesConfig, err := store.GetStore().GetTimelinesConfig(project.ID)
	assert.Nil(t, err)

	timelinesConfig.AccountConfig.TableProps = []string{U.GP_HUBSPOT_COMPANY_INDUSTRY, U.GP_HUBSPOT_COMPANY_COUNTRY,
		U.DP_ENGAGEMENT_LEVEL, U.DP_ENGAGEMENT_SCORE, U.DP_TOTAL_ENGAGEMENT_SCORE, U.DP_DOMAIN_NAME}
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

	domProps := []map[string]interface{}{
		{"$domain_name": "example1.com", "$engagement_level": "Hot", "$engagement_score": 125.300000,
			"$joinTime": 1681211371, "$total_enagagement_score": 196.000000},
		{"$domain_name": "example2.com", "$engagement_level": "Cold", "$engagement_score": 50.300000,
			"$total_enagagement_score": 96.000000},
	}
	propertiesJSON, err := json.Marshal(props)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}

	isGroupUser := true
	customerEmail := "abc@example1.com"

	// Create domain group
	group1, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	// create a domain

	dom1PropertiesJSON, err := json.Marshal(domProps[0])
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	domainID1, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Source:      model.GetRequestSourcePointer(model.UserSourceDomains),
		Group1ID:    "example1.com",
		IsGroupUser: &isGroupUser,
		Properties:  postgres.Jsonb{RawMessage: dom1PropertiesJSON},
	})
	domainUser1, errCode := store.GetStore().GetUser(project.ID, domainID1)
	assert.Equal(t, domainID1, domainUser1.ID)
	assert.Equal(t, http.StatusFound, errCode)

	// create hubspot group
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group2)

	// hubspot account associated to domain
	hubspotAccID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:    project.ID,
		Source:       model.GetRequestSourcePointer(model.UserSourceHubspot),
		Group1UserID: domainUser1.ID,
		Properties:   properties,
		IsGroupUser:  &isGroupUser,
	})
	user, errCode := store.GetStore().GetUser(project.ID, hubspotAccID)
	assert.Equal(t, user.ID, hubspotAccID)
	assert.Equal(t, http.StatusFound, errCode)

	// create another domain account
	dom2PropertiesJSON, err := json.Marshal(domProps[1])
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	domainID2, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:   project.ID,
		Source:      model.GetRequestSourcePointer(model.UserSourceDomains),
		Group1ID:    "example2.com",
		IsGroupUser: &isGroupUser,
		Properties:  postgres.Jsonb{RawMessage: dom2PropertiesJSON},
	})
	domainUser2, errCode := store.GetStore().GetUser(project.ID, domainID2)
	assert.Equal(t, domainID2, domainUser2.ID)
	assert.Equal(t, http.StatusFound, errCode)

	// Hubspot Group Events
	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
	trackPayload := SDK.TrackPayload{
		UserId:        hubspotAccID,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceHubspot,
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	trackPayload = SDK.TrackPayload{
		UserId:        hubspotAccID,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceHubspot,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	// account associated to domain
	salesforceAccID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
		Group1UserID:   domainUser2.ID,
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &isGroupUser,
	})
	user2, errCode := store.GetStore().GetUser(project.ID, salesforceAccID)
	assert.Equal(t, user2.ID, salesforceAccID)
	assert.Equal(t, http.StatusFound, errCode)
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group3)

	// Salesforce Group Events
	trackPayload = SDK.TrackPayload{
		UserId:        salesforceAccID,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	trackPayload = SDK.TrackPayload{
		UserId:        salesforceAccID,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED,
		Timestamp:     timestamp,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceSalesforce,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

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

	// Create Associated Users With Events
	randomURL := RandomURL()
	customerEmail = "@example1.com"
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
			customerEmail = "@example2.com"
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
				ProjectId:      project.ID,
				Properties:     userPropsEncoded,
				IsGroupUser:    &isGroupUser,
				Group2ID:       "2",
				Group2UserID:   hubspotAccID,
				CustomerUserId: customerUserID,
				Group1UserID:   domainUser1.ID,
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

	}
	assert.Equal(t, len(users), 10)

	t.Run("DomainUser1Details", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, project.ID, agent, domainID1, U.GROUP_NAME_DOMAINS)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := io.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.DomainName, "example1.com")
		assert.Equal(t, len(resp.AccountTimeline) > 0, true)
		assert.Equal(t, len(resp.AccountTimeline), 10)
		assert.NotNil(t, resp.LeftPaneProps)
		leftPaneProps := map[string]interface{}{U.GP_HUBSPOT_COMPANY_COUNTRY: "India",
			U.GP_HUBSPOT_COMPANY_INDUSTRY: "Freshworks-HS",
		}
		for key, value := range resp.LeftPaneProps {
			if expectedValue, ok := domProps[0][key]; ok {
				assert.Equal(t, expectedValue, value)
			}
			if expectedValue, ok := leftPaneProps[key]; ok {
				assert.Equal(t, expectedValue, value)
			}
		}
		for key, value := range resp.Milestones {
			if expectedValue, ok := props[key]; ok {
				assert.Equal(t, expectedValue, value)
			}
		}
		for _, userTimeline := range resp.AccountTimeline {
			if userTimeline.UserName == model.GROUP_ACTIVITY_USERNAME {
				assert.Equal(t, userTimeline.ExtraProp, "All")
				assert.Equal(t, userTimeline.IsAnonymous, false)
				assert.Equal(t, 1, len(userTimeline.UserActivities))
			}
		}
	})

	t.Run("domainID2Details", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, project.ID, agent, domainID2, U.GROUP_NAME_DOMAINS)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := io.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.DomainName, "example2.com")
		assert.Equal(t, len(resp.AccountTimeline) > 0, true)
		assert.Equal(t, len(resp.AccountTimeline), 4)
		assert.Equal(t, len(domProps[1]), len(resp.LeftPaneProps))
		for _, userTimeline := range resp.AccountTimeline {
			if userTimeline.UserName != model.GROUP_ACTIVITY_USERNAME {
				assert.Equal(t, userTimeline.IsAnonymous, false)
				assert.Equal(t, len(userTimeline.UserActivities), 13)
			}
		}

		for key, value := range resp.LeftPaneProps {
			if expectedValue, ok := domProps[1][key]; ok {
				assert.Equal(t, expectedValue, value)
			}
		}
		for key, value := range resp.Milestones {
			if expectedValue, ok := props[key]; ok {
				assert.Equal(t, expectedValue, value)
			}
		}
	})

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, project.ID, agent, domainID1, model.GROUP_NAME_DOMAINS)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := io.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, resp.DomainName, "example1.com")
		assert.Equal(t, len(resp.AccountTimeline), 10)
		assert.NotNil(t, resp.LeftPaneProps)
		for key, value := range resp.LeftPaneProps {
			if expectedValue, ok := domProps[0][key]; ok {
				assert.Equal(t, expectedValue, value)
			}
		}
		for key, value := range resp.Milestones {
			if expectedValue, ok := props[key]; ok {
				assert.Equal(t, expectedValue, value)
			}
		}

		assert.True(t, len(resp.AccountTimeline) > 0)

		// Loop through the AccountTimeline and perform assertions on each User Timeline
		for index, userTimeline := range resp.AccountTimeline {
			assert.NotNil(t, userTimeline.UserId)
			assert.NotNil(t, userTimeline.UserName)

			// Separate check the 10th element (Intent Activity)
			if index == 9 {
				assert.Equal(t, userTimeline.UserName, model.GROUP_ACTIVITY_USERNAME)
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

				eventFromMap, eventExistsInMap := model.TIMELINE_EVENT_PROPERTIES_CONFIG[activity.EventName]
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
						lookInProps = model.TIMELINE_EVENT_PROPERTIES_CONFIG["PageView"]
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
func sendGetProfileAccountOverviewRequest(r *gin.Engine, projectId int64, agent *model.Agent, id string, group string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/accounts/overview/%s/%s", projectId, group, id)).
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

func TestTimelineSegmentSupportEventAnalyticsQuery(t *testing.T) {
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

	// creating domain group
	group, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)

	// 5 domainn groups
	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}
	source := model.GetRequestSourcePointer(model.UserSourceDomains)
	groupUser := true
	domainAccounts := make([]string, 0)

	companyNames := []string{"FreshWorks", "CleverTap", "Adsup", "ChargeBee", "Heyflow"}
	for i := 0; i < 5; i++ {
		domID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:   project.ID,
			Source:      source,
			Group3ID:    fmt.Sprintf("%s@domainid.com", companyNames[i]),
			Properties:  domProperties,
			IsGroupUser: &groupUser,
		})
		_, errCode := store.GetStore().GetUser(project.ID, domID)
		assert.Equal(t, http.StatusFound, errCode)
		domainAccounts = append(domainAccounts, domID)
	}

	// 2 domain performed events
	for i := 0; i < 2; i++ {
		timestamp := U.UnixTimeBeforeDuration(time.Duration(1+i) * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:        domainAccounts[i],
			CreateUser:    false,
			IsNewUser:     false,
			Name:          "Third Party Intent Visit",
			Timestamp:     timestamp,
			ProjectId:     project.ID,
			Auto:          false,
			RequestSource: model.UserSourceDomains,
		}
		status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)
	}

	// 5 hubspot accounts
	var accounts []string
	cities := []string{"London", "London", "DC", "Delhi", "Paris"}
	lastEventTime := time.Now()
	for i := 0; i < 5; i++ {
		props := map[string]interface{}{
			U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED: companyNames[i],
			"$city":                         cities[i],
			"$hubspot_company_hs_object_id": 124,
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
			ProjectId:    project.ID,
			Source:       source,
			Group1ID:     fmt.Sprintf("%d", group1.ID),
			Properties:   accProps,
			IsGroupUser:  &groupUser,
			Group3UserID: domainAccounts[i],
			Group3ID:     fmt.Sprintf("%s@domainid.com", companyNames[i]),
			LastEventAt:  &lastEventTime,
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
			UserProperties:  props,
			EventProperties: eventProperties,
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
			ProjectId:    project.ID,
			Source:       source,
			Group2ID:     fmt.Sprintf("%d", group2.ID),
			Properties:   accProps,
			IsGroupUser:  &groupUser,
			Group3UserID: domainAccounts[i],
			LastEventAt:  &lastEventTime,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, account.ID)

		timestamp := U.UnixTimeBeforeDuration(time.Duration(1+i) * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:          account.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
			Timestamp:       timestamp,
			ProjectId:       project.ID,
			Auto:            false,
			RequestSource:   model.UserSourceSalesforce,
			EventProperties: props,
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
			if i > 6 {
				src = model.GetRequestSourcePointer(model.UserSourceWeb)
			}
			createdUserID, _ = store.GetStore().CreateUser(&model.User{
				ProjectId:      project.ID,
				Source:         src,
				Properties:     properties,
				Group1UserID:   accounts[i%5],
				Group1ID:       "1",
				CustomerUserId: fmt.Sprintf("hubspot@%daccount", (i%10)+1),
				Group3UserID:   domainAccounts[i%5],
				LastEventAt:    &lastEventTime,
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
				Group3UserID:   domainAccounts[i%5],
				LastEventAt:    &lastEventTime,
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
	jsonResponse, _ := io.ReadAll(w.Body)
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
					Range: int64(5),
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
	jsonResponse, _ = io.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 4)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps[U.UP_COUNTRY])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
	}

	// All accounts

	// 1. group performed event
	payload = model.TimelinePayload{
		Query: model.Query{
			GroupAnalysis:   model.GROUP_NAME_DOMAINS,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: "all_given_event",
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis: "Hubspot Companies",
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
							Operator:  "equals",
							Value:     "Heyflow",
							LogicalOp: "AND",
							Entity:    "user",
							GroupName: "$hubspot_company",
						},
					},
					FrequencyOperator: model.GreaterThanOrEqualOpStr,
					Frequency:         "1",
					IsEventPerformed:  true,
				},
			},
			Caller: "account_profiles",
			Source: "$domains",
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	var result model.AccountsProfileQueryResponsePayload
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	assert.Equal(t, result.IsPreview, true)
	resp = result.Profiles
	assert.Equal(t, len(resp), 1)
	for _, profile := range resp {
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
		assert.Equal(t, "Heyflow@domainid.com", profile.DomainName)
	}

	// 2. user performed event

	payload = model.TimelinePayload{
		Query: model.Query{
			GroupAnalysis:   model.GROUP_NAME_DOMAINS,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: "any_given_event",
			Class:           "events",
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:             U.EVENT_NAME_SESSION,
					GroupAnalysis:    "Others",
					IsEventPerformed: true,
				},
				{
					Name:             U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis:    "Hubspot Company",
					IsEventPerformed: false,
				},
			},
			Caller: "account_profiles",
			Source: "$domains",
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$in_hubspot",
					Operator:  "equals",
					Value:     "true",
					LogicalOp: "AND",
					GroupName: "$domains",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$visited_website",
					Operator:  "equals",
					Value:     "true",
					LogicalOp: "AND",
					GroupName: "$domains",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_created",
					Operator:  "equals",
					Value:     "ChargeBee",
					LogicalOp: "AND",
					GroupName: "$salesforce_account",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_created",
					Operator:  "equals",
					Value:     "Heyflow",
					LogicalOp: "OR",
					GroupName: "$salesforce_account",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_created",
					Operator:  "equals",
					Value:     "Adsup",
					LogicalOp: "OR",
					GroupName: "$salesforce_account",
				},
			},
			Timezone:   "America/Chicago",
			TableProps: []string{"$hubspot_company_created", "$hubspot_company_hs_object_id"},
		},
	}

	hostNames := []string{"ChargeBee@domainid.com", "Adsup@domainid.com", "Heyflow@domainid.com"}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	assert.Equal(t, result.IsPreview, true)
	resp = result.Profiles
	assert.Equal(t, len(resp), 3)
	for _, profile := range resp {
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
		assert.Contains(t, hostNames, profile.DomainName)
	}

	// 3. group and user events with props

	payload = model.TimelinePayload{
		Query: model.Query{
			GroupAnalysis:   model.GROUP_NAME_DOMAINS,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: "all_given_event",
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.EVENT_NAME_SESSION,
					GroupAnalysis: "Page views",
					Range:         int64(5),
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.EP_HUBSPOT_ENGAGEMENT_FROM,
							Operator:  "equals",
							Value:     "Somewhere",
							LogicalOp: "AND",
							Entity:    "event",
							GroupName: "event",
						},
					},
					IsEventPerformed: true,
				},
				{
					Name:             U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis:    "Hubspot Companies",
					Range:            int64(5),
					IsEventPerformed: true,
					Properties: []model.QueryProperty{
						{
							Type:      "categorical",
							Property:  U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
							Operator:  "equals",
							Value:     "Heyflow",
							LogicalOp: "AND",
							Entity:    "user",
							GroupName: "$hubspot_company",
						},
						{
							Type:      "categorical",
							Property:  U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
							Operator:  "equals",
							Value:     "ChargeBee",
							LogicalOp: "OR",
							Entity:    "user",
							GroupName: "$hubspot_company",
						},
					},
				},
			},
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$in_hubspot",
					Operator:  "equals",
					Value:     "true",
					LogicalOp: "AND",
					GroupName: "$hubspot_company",
				},
			},
			Caller:     "account_profiles",
			Source:     "$domains",
			TableProps: []string{"$country", "$hubspot_company_created", "$hour_of_first_event"},
		},
		SearchFilter: []string{"adsup", "charge", "hey"},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	assert.Equal(t, result.IsPreview, true)
	resp = result.Profiles

	hostNames = []string{"ChargeBee@domainid.com", "Heyflow@domainid.com"}
	assert.Equal(t, len(resp), 2)
	for _, profile := range resp {
		assert.NotNil(t, profile.TableProps["$hubspot_company_created"])
		assert.NotNil(t, profile.Identity)
		assert.NotNil(t, profile.LastActivity)
		assert.Contains(t, hostNames, profile.DomainName)
	}

	// 4. domain events

	payload = model.TimelinePayload{
		Query: model.Query{
			GroupAnalysis:   model.GROUP_NAME_DOMAINS,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: "all_given_event",
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:             "Third Party Intent Visit",
					GroupAnalysis:    model.GROUP_NAME_DOMAINS,
					IsEventPerformed: true,
				},
			},
			Caller:     "account_profiles",
			Source:     "$domains",
			TableProps: []string{"$country", "$hubspot_company_created", "$hour_of_first_event"},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	assert.Equal(t, result.IsPreview, true)
	resp = result.Profiles
	assert.Equal(t, len(resp), 2)

}

func TestTimelineAllAccountDefaultGroupProperties(t *testing.T) {
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
		{"$salesforce_account_name": "Cin7", "$page_count": 4, "$salesforce_account_id": "123", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "cin7.com", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor", "$browser": "Chrome", "$device_type": "PC", "$salesforce_account_city": "New Delhi"},
		{"$salesforce_account_name": "Repair Desk", "$page_count": 5, "$salesforce_account_id": "123", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "repairdesk.co", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "AdPushup", "$page_count": 5, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Mad Street Den", "$page_count": 5, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Heyflow", "$page_count": 4, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC", "$hubspot_company_is_public": "true"},
		{"$hubspot_company_name": "Adapt.IO", "$page_count": 4, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Clientjoy Ads", "$page_count": 4, "$hubspot_company_hs_object_id": 123, "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", "$salesforce_account_id": "123", U.SIX_SIGNAL_DOMAIN: "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", "$page_count": 4, "$salesforce_account_id": "123", U.SIX_SIGNAL_DOMAIN: "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Heyflow", U.SIX_SIGNAL_COUNTRY: "Germany", "$page_count": 4, "$hubspot_company_hs_object_id": 123, U.SIX_SIGNAL_DOMAIN: "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Clientjoy Ads", U.SIX_SIGNAL_COUNTRY: "India", "$page_count": 3, "$hubspot_company_hs_object_id": 123, U.SIX_SIGNAL_DOMAIN: "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "Adapt.IO", U.SIX_SIGNAL_COUNTRY: "India", "$page_count": 3, U.SIX_SIGNAL_DOMAIN: "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
	}

	// Creating domain Account and Group
	domProperties := postgres.Jsonb{RawMessage: json.RawMessage(`{}`)}

	accounts := make([]model.User, 0)

	var payload model.TimelinePayload

	group, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)

	numUsers := 5

	groupUser := true
	customerUserId := U.RandomLowerAphaNumString(5)
	commonDomId, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceDomains),
		Group1ID:       "1",
		CustomerUserId: customerUserId,
		Properties:     domProperties,
		IsGroupUser:    &groupUser,
	})

	_, errCode = store.GetStore().GetUser(project.ID, commonDomId)
	assert.Equal(t, http.StatusFound, errCode)
	HsDomIds := make([]string, 0)
	SfDomIds := make([]string, 0)
	VisitedDomIds := make([]string, 0)

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

		HsDomIds = append(HsDomIds, U.IfThenElse(i > 3, commonDomId, domId).(string))
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
			Group1UserID:   U.IfThenElse(i > 3, commonDomId, domId).(string),
			CustomerUserId: fmt.Sprintf("hsuser%d@%s", i+1, propertiesMap[i+5]["$hubspot_company_domain"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
	}
	// Create 5 Salesforce Accounts
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

		SfDomIds = append(SfDomIds, U.IfThenElse(i > 3, commonDomId, domId).(string))
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
			Group1UserID:   U.IfThenElse(i > 3, commonDomId, domId).(string),
			CustomerUserId: fmt.Sprintf("sfuser%d@%s", i+1, propertiesMap[i]["$salesforce_account_website"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
	}
	// Create 5 Six Signal Domains
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

		VisitedDomIds = append(VisitedDomIds, U.IfThenElse(i > 3, commonDomId, domId).(string))

		groupUser = false

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
			Group1ID:       "1",
			Group1UserID:   U.IfThenElse(i > 3, commonDomId, domId).(string),
			CustomerUserId: fmt.Sprintf("6siguser%d@%s", i+1, propertiesMap[i+10][U.SIX_SIGNAL_DOMAIN]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
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
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "Equal",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := io.ReadAll(w.Body)
		var result model.AccountsProfileQueryResponsePayload
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		resp := result.Profiles
		assert.Equal(t, result.IsPreview, true)
		assert.Equal(t, 5, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(HsDomIds, r.Identity))
		}

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "equals",
						Value:     "false",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.False(t, U.ContainsStringInArray(HsDomIds, r.Identity))
		}

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "equals",
						Value:     "$none",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.False(t, U.ContainsStringInArray(HsDomIds, r.Identity))
		}

		// in hubspot for notEqual
		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "notEqual",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.False(t, U.ContainsStringInArray(HsDomIds, r.Identity))
		}

		// in hubspot for notEqual
		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_hubspot",
						Operator:  "notEqual",
						Value:     "true",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		resp = make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.False(t, U.ContainsStringInArray(HsDomIds, r.Identity))
		}

	})

	// test in salesforce properties with single filter
	t.Run("TestForInSalesforceProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, len(resp), 5)
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(SfDomIds, r.Identity))
		}

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "false",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, len(resp), 4)
		for _, r := range resp {
			assert.False(t, U.ContainsStringInArray(SfDomIds, r.Identity))
		}

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$in_salesforce",
						Operator:  "equals",
						Value:     "$none",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, len(resp), 4)
		for _, r := range resp {
			assert.False(t, U.ContainsStringInArray(SfDomIds, r.Identity))
		}
	})

	// test in Visited website properties with single filter
	t.Run("TestForInVisitedProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(VisitedDomIds, r.Identity))
		}

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$visited_website",
						Operator:  "equals",
						Value:     "false",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, 1, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(VisitedDomIds, r.Identity))
		}

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_g",
						Type:      "categorical",
						Property:  "$visited_website",
						Operator:  "equals",
						Value:     "$none",
						LogicalOp: "AND",
					},
				},
			},
		}

		w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ = io.ReadAll(w.Body)
		result = model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp = result.Profiles
		assert.Equal(t, 1, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(VisitedDomIds, r.Identity))
		}
	})

	// test in salesforce and in hubspot properties with multiple filter
	t.Run("TestInPropertiesWithValueMultipleFilters", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 1, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(SfDomIds, r.Identity))
			assert.True(t, U.ContainsStringInArray(HsDomIds, r.Identity))
		}

	})

	// test in salesforce and in hubspot properties with multiple filter with false
	t.Run("TestInPropertiesWithValueMultipleFiltersWithFalse", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
						Value:     "false",
						LogicalOp: "AND",
					},
				},
			},
		}

		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(SfDomIds, r.Identity))
		}

	})

	// test in salesforce and user properties properties with multiple filter
	t.Run("TestInPropertiesWithValueMultipleFiltersWithUserProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
						Entity:    "user_group",
						Type:      "numerical",
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 1, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(SfDomIds, r.Identity))
		}
	})

	// test in salesforce and visited website properties with multiple filter
	t.Run("TestInVisitedWebsitePropertiesWithValueMultipleFilters", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 1, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(SfDomIds, r.Identity))
		}
	})

	// test in column properties with single filter

	t.Run("TestInPropertiesWithColumn", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_group",
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 5, len(resp))
	})
	// test in column properties with multiple filter
	t.Run("TestInPropertiesMultipleFilterWithColumn", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
				GlobalUserProperties: []model.QueryProperty{
					{
						Entity:    "user_group",
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 1, len(resp))
	})

	// test user properties and visited webite properties with multiple filter

	t.Run("TestInVisitedWebsitePropertiesWithValueMultipleFiltersWithUserProperties", func(t *testing.T) {

		payload = model.TimelinePayload{
			Query: model.Query{
				Source: U.GROUP_NAME_DOMAINS,
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
		jsonResponse, _ := io.ReadAll(w.Body)
		result := model.AccountsProfileQueryResponsePayload{}
		err = json.Unmarshal(jsonResponse, &result)
		assert.Nil(t, err)
		assert.Equal(t, result.IsPreview, true)
		resp := result.Profiles
		assert.Equal(t, 4, len(resp))
		for _, r := range resp {
			assert.True(t, U.ContainsStringInArray(VisitedDomIds, r.Identity))

		}
	})
}

// Move TestAPIGetProfileAccountHandler Test Cases here.
func TestTimelineAllAccounts(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// Create Domain Group
	domaindGroup, status := store.GetStore().CreateOrGetDomainsGroup(project.ID)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, domaindGroup)

	var payload model.TimelinePayload

	// Test :- CRM not enabled
	payload.Query.Source = U.GROUP_NAME_HUBSPOT_COMPANY
	w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// Create 5 Domains
	numDomains := 5
	domains := []string{"adpushup.com", "madstreetden.com", "heyflow.app", "clientjoy.io", "adapt.io"}
	domainUsers := make([]model.User, 0)
	for i := 0; i < numDomains; i++ {
		var domProperties postgres.Jsonb
		if i > 1 {
			domProperties = postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"$domain_name":"%s",
		"$engagement_level":"Hot","$engagement_score":125.300000,"$joinTime":1681211371,
		"$total_enagagement_score":196.000000}`, domains[i]))}
		} else {
			domProperties = postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"$domain_name":"%s",
		"$engagement_level":"Cold","$engagement_score":5.300000,"$joinTime":1681211371,
		"$total_enagagement_score":120.000000}`, domains[i]))}
		}
		source := model.GetRequestSourcePointer(model.UserSourceDomains)
		groupUser := true
		domID, _ := store.GetStore().CreateUser(&model.User{
			ID:          fmt.Sprintf("dom-%s", base64.StdEncoding.EncodeToString([]byte(domains[i]))),
			ProjectId:   project.ID,
			Source:      source,
			Group1ID:    domains[i],
			Properties:  domProperties,
			IsGroupUser: &groupUser,
		})
		domainUser, errCode := store.GetStore().GetUser(project.ID, domID)
		assert.Equal(t, http.StatusFound, errCode)
		domainUsers = append(domainUsers, *domainUser)
	}

	// Create CRM Groups
	hubspotGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, hubspotGroup)
	salesforceGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, salesforceGroup)
	assert.Equal(t, http.StatusCreated, status)
	sixSignalGroup, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.NotNil(t, sixSignalGroup)
	assert.Equal(t, http.StatusCreated, status)

	// Properties Map
	dummyPropsMap := []map[string]interface{}{
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome", "$device_type": "PC", "$hubspot_company_is_public": "true"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC"},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome", "$device_type": "PC", "$hubspot_company_notes_last_updated": 1710848309},
		{"$salesforce_account_name": "AdPushup", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "adpushup.com", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target", "$browser": "Chrome", "$device_type": "PC", "$salesforce_account_target_account__c": true},
		{"$salesforce_account_name": "Mad Street Den", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "madstreetden.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown", "$browser": "Chrome", "$device_type": "PC", "$salesforce_account_target_account__c": true},
		{"$salesforce_account_name": "Heyflow", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "heyflow.app", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown", "$browser": "Chrome", "$device_type": "PC"},
		{"$salesforce_account_name": "Clientjoy Ads", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "clientjoy.io", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor", "$browser": "Chrome", "$device_type": "PC", "$salesforce_account_city": "New Delhi"},
		{"$salesforce_account_name": "Adapt.IO", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "adapt.io", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer", "$browser": "Chrome", "$device_type": "PC"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Heyflow", U.SIX_SIGNAL_COUNTRY: "Germany", U.SIX_SIGNAL_DOMAIN: "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Clientjoy Ads", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$browser": "Chrome"},
		{U.SIX_SIGNAL_NAME: "Adapt.IO", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$browser": "Chrome"},
	}
	userPropsMap := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000, U.UP_COMPANY: "XYZ Company"},
		{"$browser": "Chrome", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 105, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 120, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 110, "$session_spent_time": 2500},
	}

	// Create Associated Accounts
	groupUsers := make([]model.User, 0)
	users := make([]model.User, 0)
	numUsers := 15
	// Create 5 Hubspot Companies
	for i := 0; i < numUsers; i++ {
		isGroupUser := true
		propertiesJSON, err := json.Marshal(dummyPropsMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		var source *int
		var isHubspot, isSalesforce, isSixSignal string
		var hsUserID, sfUserID, sixsUserID string
		var customerUserID string
		if i < 5 {
			source = model.GetRequestSourcePointer(model.UserSourceHubspot)
			isHubspot = "2"
			customerUserID = fmt.Sprintf("hsuser%d@%s", i+1, dummyPropsMap[i]["$hubspot_company_domain"])
		} else if i < 10 {
			source = model.GetRequestSourcePointer(model.UserSourceSalesforce)
			isSalesforce = "3"
			customerUserID = fmt.Sprintf("sfuser%d@%s", i+1, dummyPropsMap[i]["$salesforce_account_website"])
		} else if i < 15 {
			source = model.GetRequestSourcePointer(model.UserSourceSixSignal)
			isSixSignal = "4"
			customerUserID = fmt.Sprintf("6suser%d@%s", i+1, dummyPropsMap[i][U.SIX_SIGNAL_DOMAIN])
		}

		lastEventTime := time.Now()
		createdGroupUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:    project.ID,
			Properties:   properties,
			IsGroupUser:  &isGroupUser,
			Group1ID:     domainUsers[i%5].Group1ID,
			Group1UserID: domainUsers[i%5].ID,
			Group2ID:     isHubspot,
			Group3ID:     isSalesforce,
			Group4ID:     isSixSignal,
			Source:       source,
			LastEventAt:  &lastEventTime,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdGroupUserID)
		assert.Equal(t, http.StatusFound, errCode)
		if i < 5 {
			hsUserID = createdGroupUserID
		} else if i < 10 {
			sfUserID = createdGroupUserID
		} else if i < 15 {
			sixsUserID = createdGroupUserID
		}

		groupUsers = append(groupUsers, *account)

		// user associated to the account
		isGroupUser = false
		propertiesJSON, err = json.Marshal(userPropsMap[i%5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		userProperties := postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Properties:     userProperties,
			IsGroupUser:    &isGroupUser,
			Group1UserID:   domainUsers[i%5].ID,
			Group2UserID:   hsUserID,
			Group3UserID:   sfUserID,
			Group4UserID:   sixsUserID,
			CustomerUserId: customerUserID,
			Source:         source,
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(groupUsers), 15)
	assert.Equal(t, len(users), 15)

	// check total users created

	userCount, status := store.GetStore().GetAccountAssociatedUserCountByProjectID(project.ID, 1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, int64(30), userCount)

	// Test Cases :-

	// Search a Hubspot Company
	payload = model.TimelinePayload{
		Query: model.Query{
			Source: "$domains",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name",
				"$domain_name", "$engagement_level", "$engagement_score", "$total_enagagement_score"},
		},
		SearchFilter: []string{"hey"},
	}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := io.ReadAll(w.Body)
	var result model.AccountsProfileQueryResponsePayload
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp := result.Profiles
	assert.Equal(t, result.IsPreview, false)
	assert.Equal(t, len(resp), 1)
	assert.Contains(t, resp[0].DomainName, "hey")
	engagementMap := map[string]interface{}{"$engagement_level": "Hot", "$engagement_score": 125.3}
	assert.Equal(t, engagementMap, resp[0].TableProps[U.DP_ENGAGEMENT_LEVEL])

	// Search a Domain
	payload = model.TimelinePayload{
		Query: model.Query{
			Source: "$domains",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name",
				"$domain_name", "$engagement_level", "$engagement_score", "$total_enagagement_score"},
		},
		SearchFilter: []string{"maruti", "hey", "adapt"},
	}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, false)
	assert.Equal(t, len(resp), 2)
	searchNames := []string{"heyflow.app", "adapt.io"}
	assert.Contains(t, searchNames, resp[0].DomainName)
	assert.Contains(t, searchNames, resp[1].DomainName)
	for i := range resp {
		assert.Equal(t, engagementMap, resp[i].TableProps["$engagement_level"])
		assert.Equal(t, 125.3, resp[i].TableProps["$engagement_score"])
		assert.Equal(t, float64(196), resp[i].TableProps["$total_enagagement_score"])
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			Source: "$domains",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name",
				"$domain_name", "$engagement_level", "$engagement_score", "$total_enagagement_score"},
		},
	}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, false)
	assert.Equal(t, len(resp), 5)
	assert.Equal(t, result.Count, int64(0))
	for i := range resp {
		assert.NotEmpty(t, resp[i].TableProps["$engagement_level"])
		assert.NotEmpty(t, resp[i].TableProps["$engagement_score"])
		assert.NotEmpty(t, resp[i].TableProps["$total_enagagement_score"])
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_domain",
					Operator:  "notEqual",
					Value:     "$none",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_HUBSPOT_COMPANY,
				}, {
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_HUBSPOT_COMPANY,
				},
			}, Source: "$domains",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name"},
		},
		SearchFilter: []string{"adapt", "hey"},
	}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, true)
	assert.Equal(t, len(resp), 1)
	assert.Contains(t, resp[0].DomainName, "adapt")
	assert.Greater(t, resp[0].LastActivity, U.TimeNowZ().AddDate(0, 0, -1))
	assert.NotEmpty(t, resp[0].TableProps[U.SIX_SIGNAL_NAME])
	assert.NotEmpty(t, resp[0].TableProps["$hubspot_company_name"])
	assert.NotEmpty(t, resp[0].TableProps["$salesforce_account_name"])

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_name",
					Operator:  "equals",
					Value:     "Adapt.IO",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_SALESFORCE_ACCOUNT,
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "notEqual",
					Value:     "India",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_HUBSPOT_COMPANY,
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_country",
					Operator:  "equals",
					Value:     "Pakistan",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_HUBSPOT_COMPANY,
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$li_country",
					Operator:  "equals",
					Value:     "Germany",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_LINKEDIN_COMPANY,
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$engagement_score",
					Operator:  "equals",
					Value:     "50",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_DOMAINS,
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$g2_entity",
					Operator:  "equals",
					Value:     "something",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_G2,
				},
			},
			Source: "$domains",
		},
		SearchFilter: []string{"adapt", "hey"},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, true)
	assert.Equal(t, len(resp), 0)

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$engagement_level",
					Operator:  "equals",
					Value:     "Cold",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_DOMAINS,
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_target_account__c",
					Operator:  "equals",
					Value:     "true",
					LogicalOp: "AND",
					GroupName: "$salesforce_account",
				},
			}, Source: "$domains",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name",
				"$domain_name", "$engagement_level", "$engagement_score", "$total_enagagement_score"}},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, true)
	assert.Equal(t, len(resp), 2)
	assert.Equal(t, result.Count, int64(0))
	engagementMap = map[string]interface{}{"$engagement_level": "Cold", "$engagement_score": 5.3}
	for i := range resp {
		assert.Greater(t, resp[i].LastActivity, U.TimeNowZ().AddDate(0, 0, -1))
		assert.NotEmpty(t, resp[i].TableProps[U.SIX_SIGNAL_NAME])
		assert.NotEmpty(t, resp[i].TableProps["$hubspot_company_name"])
		assert.NotEmpty(t, resp[i].TableProps["$salesforce_account_name"])
		assert.Equal(t, engagementMap, resp[i].TableProps["$engagement_level"])
		assert.Equal(t, 5.3, resp[i].TableProps["$engagement_score"])
		assert.Equal(t, float64(120), resp[i].TableProps["$total_enagagement_score"])
	}

	payload = model.TimelinePayload{
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "datetime",
					Property:  "$hubspot_company_notes_last_updated",
					Operator:  "notInLast",
					Value:     "{\"fr\":1712428200,\"to\":1713032999,\"ovp\":false,\"num\":1,\"gran\":\"week\"}",
					LogicalOp: "AND",
					GroupName: U.GROUP_NAME_HUBSPOT_COMPANY,
				},
			}, Source: "$domains",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name"},
		},
	}
	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	assert.Equal(t, result.IsPreview, true)
	assert.Equal(t, result.Count, int64(0))
	resp = result.Profiles
	assert.Equal(t, len(resp), 1)
	assert.Contains(t, resp[0].DomainName, "adapt")
	assert.Greater(t, resp[0].LastActivity, U.TimeNowZ().AddDate(0, 0, -1))
	assert.NotEmpty(t, resp[0].TableProps[U.SIX_SIGNAL_NAME])
	assert.NotEmpty(t, resp[0].TableProps["$hubspot_company_name"])
	assert.NotEmpty(t, resp[0].TableProps["$salesforce_account_name"])

	// Create Events
	trackPayload := SDK.TrackPayload{
		UserId:          groupUsers[0].ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
		EventProperties: map[string]interface{}{},
		UserProperties:  map[string]interface{}{},
		ProjectId:       project.ID,
		Auto:            false,
		RequestSource:   model.UserSourceWeb,
		Timestamp:       time.Now().Unix(),
	}
	status, response := SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)
	// Create Events
	trackPayload = SDK.TrackPayload{
		UserId:          users[0].ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SESSION,
		EventProperties: map[string]interface{}{},
		UserProperties:  map[string]interface{}{},
		ProjectId:       project.ID,
		Auto:            false,
		RequestSource:   model.UserSourceWeb,
		Timestamp:       time.Now().Unix(),
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	trackPayload = SDK.TrackPayload{
		UserId:        groupUsers[0].ID,
		CreateUser:    false,
		IsNewUser:     false,
		Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
		Timestamp:     time.Now().Unix() + 100,
		ProjectId:     project.ID,
		Auto:          false,
		RequestSource: model.UserSourceHubspot,
	}
	status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotEmpty(t, response)
	assert.Equal(t, http.StatusOK, status)

	w = sendGetTopEventsForADomainRequest(r, project.ID, agent, domainUsers[0].Group1ID)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	newResp := make([]model.TimelineEvent, 0)
	err = json.Unmarshal(jsonResponse, &newResp)
	assert.Nil(t, err)
	assert.Equal(t, len(newResp), 2)
	assert.Equal(t, newResp[0].Name, U.EVENT_NAME_SESSION)
	assert.Equal(t, newResp[1].Name, U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED)
	assert.Contains(t, newResp[0].UserID, domainUsers[0].Group1ID)
	assert.False(t, newResp[0].IsGroupUser)
	assert.True(t, newResp[1].IsGroupUser)
}

func sendGetProfileAccountRequestConsumingMarker(r *gin.Engine, projectId int64, agent *model.Agent, payload model.TimelinePayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/profiles/accounts?user_marker=true", projectId)).
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

func TestTimelineAccountsConsumingMarker(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	SegmentMarkerTest(t, project, agent, r)

	segments, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	var segmentID string
	for _, segment := range segments["$domains"] {
		if segment.Name == "User Group props" {
			segmentID = segment.Id
			break
		}
	}

	// global props type segment
	payload := model.TimelinePayload{
		SegmentId: segmentID,
		Query: model.Query{
			Source: "$domains",
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_group",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "contains",
					Value:     "India",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_group",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "US",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_group",
					Type:      "categorical",
					Property:  "$device_type",
					Operator:  "notEqual",
					Value:     "macBook",
					LogicalOp: "AND",
				},
			},
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name"},
		},
	}

	w := sendGetProfileAccountRequestConsumingMarker(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := io.ReadAll(w.Body)
	var result model.AccountsProfileQueryResponsePayload
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp := result.Profiles
	assert.Equal(t, result.IsPreview, false)
	assert.Equal(t, len(resp), 3)
	assert.Equal(t, result.Count, int64(3))
	domNames := []string{"domain0id.com", "domain1id.com", "domain2id.com"}

	for _, profile := range resp {
		assert.Contains(t, domNames, profile.DomainName)
		assert.NotEmpty(t, profile.Identity)
		assert.Greater(t, profile.LastActivity, U.TimeNowZ().AddDate(0, 0, -1))
		assert.NotEmpty(t, profile.TableProps[U.SIX_SIGNAL_NAME])
		assert.NotEmpty(t, profile.TableProps["$hubspot_company_name"])
		assert.NotEmpty(t, profile.TableProps["$salesforce_account_name"])
	}

	for _, segment := range segments["$domains"] {
		if segment.Name == "Hubspot Group Performed Event" {
			segmentID = segment.Id
			break
		}
	}

	// adding a search filter (performed event segment)
	payload = model.TimelinePayload{
		SegmentId: segmentID,
		Query: model.Query{
			Source: "$domains",
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:              U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis:     "Most Recent",
					IsEventPerformed:  true,
					Range:             int64(180),
					Frequency:         "0",
					FrequencyOperator: model.GreaterThanOpStr,
				},
				{
					Name:              U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
					GroupAnalysis:     "Most Recent",
					IsEventPerformed:  true,
					Range:             int64(180),
					Frequency:         "0",
					FrequencyOperator: model.GreaterThanOpStr,
				},
			},
			TableProps:      []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name"},
			EventsCondition: model.EventCondAllGivenEvent,
		},
		SearchFilter: []string{"domain0id.com", "domain1id.com"},
	}

	w = sendGetProfileAccountRequestConsumingMarker(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, false)
	assert.Equal(t, len(resp), 2)
	assert.Equal(t, result.Count, int64(2))

	for _, profile := range resp {
		assert.Contains(t, domNames, profile.DomainName)
		assert.NotEmpty(t, profile.Identity)
		assert.Greater(t, profile.LastActivity, U.TimeNowZ().AddDate(0, 0, -1))
		assert.NotEmpty(t, profile.TableProps[U.SIX_SIGNAL_NAME])
		assert.NotEmpty(t, profile.TableProps["$hubspot_company_name"])
		assert.NotEmpty(t, profile.TableProps["$salesforce_account_name"])
	}

	// adding a search filter (performed event segment) and additional filters
	today := time.Now().UTC()
	dayOfWeek := today.Weekday()
	payload = model.TimelinePayload{
		SegmentId: segmentID,
		Query: model.Query{
			Type:            "unique_users",
			EventsCondition: "all_given_event",
			Caller:          "account_profiles",
			Source:          "$domains",
			GroupAnalysis:   "$domains",

			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:              U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
					GroupAnalysis:     "Most Recent",
					IsEventPerformed:  true,
					Range:             int64(180),
					Frequency:         "0",
					FrequencyOperator: model.GreaterThanOpStr,
				},
				{
					Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis: "Hubspot Company Created",
					Properties: []model.QueryProperty{
						{
							Entity:    "event",
							GroupName: "event",
							Type:      "categorical",
							Property:  "$day_of_week",
							Operator:  "equals",
							Value:     dayOfWeek.String(),
							LogicalOp: "AND",
						},
					},
					IsEventPerformed:  true,
					Range:             int64(180),
					Frequency:         "0",
					FrequencyOperator: model.GreaterThanOpStr,
				},
			},
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name"},
		},
		SearchFilter: []string{"domain0id.com", "domain1id.com"},
	}

	w = sendGetProfileAccountRequestConsumingMarker(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	result = model.AccountsProfileQueryResponsePayload{}
	err = json.Unmarshal(jsonResponse, &result)
	assert.Nil(t, err)
	resp = result.Profiles
	assert.Equal(t, result.IsPreview, true)
	assert.Equal(t, len(resp), 2)

	// query where source is All
	payload = model.TimelinePayload{
		SegmentId: segmentID,
		Query: model.Query{
			Source:     "All",
			TableProps: []string{"$hubspot_company_name", U.SIX_SIGNAL_NAME, "$salesforce_account_name"},
		},
		SearchFilter: []string{"domain0id.com", "domain1id.com"},
	}

	w = sendGetProfileAccountRequestConsumingMarker(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusBadRequest, w.Code)

}

func sendUpdateEventConfigRequest(r *gin.Engine, projectId int64, agent *model.Agent, eventName string, payload []string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/v1/profiles/events_config/%s", projectId, eventName)).
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

func TestTimelineUpdateEventsConfig(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	newSessionEventProperties := []string{
		U.EP_CHANNEL,
		U.EP_PAGE_URL,
		U.EP_REFERRER_URL,
		U.EP_PAGE_COUNT,
		U.SP_SPENT_TIME,
		U.EP_SOURCE,
	}
	w := sendUpdateEventConfigRequest(r, project.ID, agent, U.EVENT_NAME_SESSION, newSessionEventProperties)
	assert.Equal(t, http.StatusOK, w.Code)
}

func sendUpdateTablePropertiesRequest(r *gin.Engine, projectId int64, agent *model.Agent, profileType string, payload []string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/v1/profiles/%s/table_properties", projectId, profileType)).
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

func TestTimelineUpdateTablePropertiesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	newTableProperties := []string{
		U.SIX_SIGNAL_NAME,
		U.SIX_SIGNAL_INDUSTRY,
		U.SIX_SIGNAL_EMPLOYEE_RANGE,
		U.SIX_SIGNAL_ANNUAL_REVENUE,
		U.SIX_SIGNAL_EMPLOYEE_COUNT,
	}

	w := sendUpdateTablePropertiesRequest(r, project.ID, agent, model.PROFILE_TYPE_ACCOUNT, []string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = sendUpdateTablePropertiesRequest(r, project.ID, agent, model.PROFILE_TYPE_ACCOUNT, newTableProperties)
	assert.Equal(t, http.StatusOK, w.Code)

	timelinesConfig, err := store.GetStore().GetTimelinesConfig(project.ID)
	assert.Nil(t, err)
	assert.Equal(t, newTableProperties, timelinesConfig.AccountConfig.TableProps)
}

func sendUpdateSegmentTablePropertiesRequest(r *gin.Engine, projectId int64, agent *model.Agent, segmentID string, payload []string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/v1/profiles/segments/%s/table_properties", projectId, segmentID)).
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

func TestTimelineUpdateSegmentTablePropertiesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	newTableProperties := []string{
		U.SIX_SIGNAL_NAME,
		U.SIX_SIGNAL_INDUSTRY,
		U.SIX_SIGNAL_EMPLOYEE_RANGE,
		U.SIX_SIGNAL_ANNUAL_REVENUE,
		U.SIX_SIGNAL_EMPLOYEE_COUNT,
	}

	segment, status := store.GetStore().GetSegmentByName(project.ID, U.ALL_ACCOUNT_DEFAULT_PROPERTIES_DISPLAY_NAMES[U.VISITED_WEBSITE])
	assert.Equal(t, http.StatusFound, status)

	w := sendUpdateSegmentTablePropertiesRequest(r, project.ID, agent, segment.Id, []string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = sendUpdateSegmentTablePropertiesRequest(r, project.ID, agent, segment.Id, newTableProperties)
	assert.Equal(t, http.StatusOK, w.Code)

	updatedSegment, status := store.GetStore().GetSegmentById(project.ID, segment.Id)
	assert.Equal(t, http.StatusFound, status)

	var segmentQuery model.Query
	err = U.DecodePostgresJsonbToStructType(updatedSegment.Query, &segmentQuery)
	assert.Nil(t, err)
	assert.Equal(t, newTableProperties, segmentQuery.TableProps)
}

func TestTimelineGetConfiguredUserPropertiesWithValuesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var timelinesConfig model.TimelinesConfig

	timelinesConfig.UserConfig.TableProps = []string{"$page_count", "$session_spent_time"}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	props := map[string]interface{}{
		U.UP_NAME:             "Cameron Williomson",
		U.UP_COMPANY:          "Freshworks",
		U.UP_COUNTRY:          "Australia",
		U.UP_SESSION_COUNT:    int(8),
		U.UP_TOTAL_SPENT_TIME: int(500),
		U.UP_PAGE_COUNT:       int(10),
	}
	propertiesJSON, err := json.Marshal(props)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties := postgres.Jsonb{RawMessage: propertiesJSON}
	customerEmail := "abc@example.com"
	lastEventTime := time.Now().Add(time.Duration(-6) * time.Hour)

	createdUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
		CustomerUserId: customerEmail,
		Properties:     properties,
		LastEventAt:    &lastEventTime,
	})
	user, errCode := store.GetStore().GetUser(project.ID, createdUserID)
	assert.Equal(t, user.ID, createdUserID)
	assert.Equal(t, http.StatusFound, errCode)

	w := sendGetProfileUserPropertiesRequest(r, project.ID, agent, "randomuser", "")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = sendGetProfileUserPropertiesRequest(r, project.ID, agent, "randomuser", "false")
	assert.Equal(t, http.StatusNotFound, w.Code)
	jsonResponse, _ := io.ReadAll(w.Body)
	resp := map[string]interface{}{}
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)

	w = sendGetProfileUserPropertiesRequest(r, project.ID, agent, customerEmail, "false")
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = io.ReadAll(w.Body)
	resp = map[string]interface{}{}
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	for _, prop := range timelinesConfig.UserConfig.TableProps {
		_, exists := resp[prop]
		assert.Equal(t, true, exists)
	}
}

func sendGetProfileUserPropertiesRequest(r *gin.Engine, projectId int64, agent *model.Agent, userId string, isAnonymous string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/user_properties/%s?is_anonymous=%s", projectId, userId, isAnonymous)).
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

func sendGetTopEventsForADomainRequest(r *gin.Engine, projectID int64, agent *model.Agent, domainName string) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/accounts/top_events/%s", projectID, base64.StdEncoding.EncodeToString([]byte(domainName)))).
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
