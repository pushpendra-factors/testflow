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
	propsMap := []map[string]interface{}{
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 100, "$session_spent_time": 2500},
		{"$browser": "Chrome", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 105, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "London", "$country": "UK", "$device_type": "desktop", "$page_count": 120, "$session_spent_time": 2000},
		{"$browser": "Brave", "$city": "Paris", "$country": "France", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 3000},
		{"$browser": "Edge", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Firefox", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 100, "$session_spent_time": 3000},
		{"$browser": "Firefox", "$city": "Dubai", "$country": "UAE", "$device_type": "desktop", "$page_count": 150, "$session_spent_time": 2100},
		{"$browser": "Chrome", "$city": "Delhi", "$country": "India", "$device_type": "iPad", "$page_count": 150, "$session_spent_time": 2100},
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
		payload.Source = source
		w := sendGetProfileUserRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), count)
		assert.Condition(t, func() bool {
			for i, user := range resp {
				if source == "All" {
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
					assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
				}
			}
			return true
		})
	}

	// 2. UserSourceWeb (1 filter, no segment applied)
	payload = model.TimelinePayload{
		Source: "web",
		Filters: []model.QueryProperty{{
			Entity:    "user_g",
			Type:      "categorical",
			Property:  "$country",
			Operator:  "equals",
			Value:     "India",
			LogicalOp: "AND",
		}},
		SegmentId: "",
	}
	w := sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	resp := make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	assert.Condition(t, func() bool {
		for i, user := range resp {
			for _, prop := range timelinesConfig.UserConfig.TableProps {
				assert.NotNil(t, user.TableProps[prop])
			}
			assert.NotNil(t, user.LastActivity)
			if i > 0 {
				assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
			}
		}
		return true
	})

	// 2. UserSourceWeb (1 search filter applied)
	payload = model.TimelinePayload{
		Source: "web",
		SearchFilter: []model.QueryProperty{{
			Entity:    "user_g",
			Type:      "categorical",
			Property:  "$user_id",
			Operator:  "contains",
			Value:     "user2",
			LogicalOp: "AND",
		}},
		SegmentId: "",
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 1)
	assert.Equal(t, resp[0].Identity, "user2@example.com")
	assert.Condition(t, func() bool {
		for _, prop := range timelinesConfig.UserConfig.TableProps {
			assert.NotNil(t, resp[0].TableProps[prop])
		}
		assert.NotNil(t, resp[0].LastActivity)
		return true
	})

	// 3. UserSourceWeb (Segment Applied, no filters)
	// creating a segment
	segmentPayload := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$browser",
					Operator:  "equals",
					Value:     "Chrome",
					LogicalOp: "AND",
				},
			},
			TableProps: []string{"$country", "$page_count"},
		},
		Type: "web",
	}
	status, err := store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "web",
		SegmentId: segments["web"][0].Id,
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 4)
	assert.Condition(t, func() bool {
		for i, user := range resp {
			if i == 3 {
				assert.Equal(t, user.IsAnonymous, true)
			} else {
				assert.Equal(t, user.IsAnonymous, false)
			}
			for _, prop := range timelinesConfig.UserConfig.TableProps {
				assert.NotNil(t, user.TableProps[prop])
			}
			assert.NotNil(t, user.LastActivity)
			if i > 0 {
				assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
			}
		}
		return true
	})

	// 4. UserSourceWeb (Segment with multiple filters applied, no filters)
	segmentPayload = &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query: model.Query{
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
			TableProps: []string{"$country", "$page_count"},
		},
		Type: "web",
	}

	err, status = store.GetStore().UpdateSegmentById(project.ID, segments["web"][0].Id, *segmentPayload)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, err)

	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	payload = model.TimelinePayload{
		Source:    "web",
		SegmentId: segments["web"][0].Id,
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 4)
	assert.Condition(t, func() bool {
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
				assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
			}
		}
		return true
	})

	// 5. UserSourceWeb (Segment with multiple filters applied, 1 filter)
	payload = model.TimelinePayload{
		Source: "web",
		Filters: []model.QueryProperty{
			{
				Entity:    "user_g",
				Type:      "categorical",
				Property:  "$session_spent_time",
				Operator:  "greaterThanOrEqual",
				Value:     "2500",
				LogicalOp: "AND",
			},
		},
		SegmentId: segments["web"][0].Id,
	}
	w = sendGetProfileUserRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	assert.Condition(t, func() bool {
		for i, user := range resp {
			if i == 0 {
				assert.Equal(t, user.IsAnonymous, false)
			} else {
				assert.Equal(t, user.IsAnonymous, true)
			}
			for _, prop := range timelinesConfig.UserConfig.TableProps {
				assert.NotNil(t, user.TableProps[prop])
			}
			assert.NotNil(t, user.LastActivity)
			if i > 0 {
				assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
			}
		}
		return true
	})
}

func sendGetProfileUserRequest(r *gin.Engine, projectId int64, agent *model.Agent, payload model.TimelinePayload) *httptest.ResponseRecorder {

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

	var timelinesConfig model.TimelinesConfig

	timelinesConfig.UserConfig.LeftpaneProps = []string{"$email", "$page_count", "$user_id", "$name", "$session_spent_time"}
	timelinesConfig.UserConfig.Milestones = []string{"$milesone_1", "$milesone_2", "$milesone_3"}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

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
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &isGroupUser,
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
		assert.Condition(t, func() bool {
			for i, property := range resp.LeftPaneProps {
				assert.Equal(t, (*userPropsDecoded)[i], property)
			}
			for i, property := range resp.Milestones {
				assert.Equal(t, (*userPropsDecoded)[i], property)
			}
			return true
		})
		assert.NotNil(t, resp.GroupInfos)
		assert.Equal(t, resp.GroupInfos[0], model.GroupsInfo{GroupName: U.STANDARD_GROUP_DISPLAY_NAMES[U.GROUP_NAME_HUBSPOT_COMPANY], AssociatedGroup: "Freshworks"})
		assert.Equal(t, resp.GroupInfos[1], model.GroupsInfo{GroupName: U.STANDARD_GROUP_DISPLAY_NAMES[U.GROUP_NAME_SALESFORCE_ACCOUNT], AssociatedGroup: ""})
		assert.NotNil(t, resp.UserActivity)
		assert.Condition(t, func() bool {
			for i, activity := range resp.UserActivity {
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
				assert.Condition(t, func() bool { return activity.Timestamp <= uint64(time.Now().UTC().Unix()) })
				if i > 1 {
					if resp.UserActivity[i].Timestamp > resp.UserActivity[i-1].Timestamp {
						return false
					}
				}
				assert.Condition(t, func() bool {
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
							assert.Condition(t, func() bool { return i < len(lookInProps) })
						}
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
		{"$salesforce_account_name": "Pepper Content", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "peppercontent.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target"},
		{"$salesforce_account_name": "o9 Solutions", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "o9solutions.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "GoLinks Reporting", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "golinks.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "Cin7", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "cin7.com", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor"},
		{"$salesforce_account_name": "Repair Desk", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "repairdesk.co", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer"},
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet"},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development"},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services"},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development"},
		{U.SIX_SIGNAL_NAME: "Heyflow", U.SIX_SIGNAL_COUNTRY: "Germany", U.SIX_SIGNAL_DOMAIN: "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development"},
		{U.SIX_SIGNAL_NAME: "Clientjoy Ads", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services"},
		{U.SIX_SIGNAL_NAME: "Adapt.IO", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services"},
	}

	// Create 5 Salesforce Accounts
	accounts := make([]model.User, 0)
	numUsers := 5
	groupUser := true
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
			Group2ID:       "2",
			CustomerUserId: fmt.Sprintf("sfuser%d@%s", i+1, propertiesMap[i]["$salesforce_account_website"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
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
			Group1ID:       "1",
			CustomerUserId: fmt.Sprintf("hsuser%d@%s", i+1, propertiesMap[i+5]["$hubspot_company_domain"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
	}
	assert.Equal(t, len(accounts), 10)

	// Create 5 Six Signal Domains
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(propertiesMap[i+5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSixSignal)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group3ID:       "3",
			CustomerUserId: fmt.Sprintf("6siguser%d@%s", i+1, propertiesMap[i+10][U.SIX_SIGNAL_DOMAIN]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
	}
	assert.Equal(t, len(accounts), 15)

	var payload model.TimelinePayload

	// Test Cases :-

	//1. Source: All, 1 group exists
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)

	payload.Source = "All"
	w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	resp := make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 5)
	assert.Condition(t, func() bool {
		for i, user := range resp {
			assert.Equal(t, user.Name, propertiesMap[9-i][U.GP_HUBSPOT_COMPANY_NAME])
			assert.Equal(t, user.HostName, propertiesMap[9-i][U.GP_HUBSPOT_COMPANY_DOMAIN])
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

	//2 more groups
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.NotNil(t, group3)
	assert.Equal(t, http.StatusCreated, status)

	// 2. Accounts from Different Sources (No filter, no segment applied)
	sourceToUserCountMap := map[string]int{"All": 15, U.GROUP_NAME_HUBSPOT_COMPANY: 5, U.GROUP_NAME_SALESFORCE_ACCOUNT: 5, U.GROUP_NAME_SIX_SIGNAL: 5}
	for source, count := range sourceToUserCountMap {
		payload.Source = source
		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err = json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), count)
		assert.Condition(t, func() bool {
			for i, user := range resp {
				if source == "All" {
					if i < 5 {
						assert.Equal(t, user.Name, propertiesMap[14-i][U.SIX_SIGNAL_NAME])
						assert.Equal(t, user.HostName, propertiesMap[14-i][U.SIX_SIGNAL_DOMAIN])
					} else if i >= 5 && i < 10 {
						assert.Equal(t, user.Name, propertiesMap[14-i]["$hubspot_company_name"])
						assert.Equal(t, user.HostName, propertiesMap[14-i]["$hubspot_company_domain"])
					} else {
						assert.Equal(t, user.Name, propertiesMap[14-i]["$salesforce_account_name"])
						assert.Equal(t, user.HostName, propertiesMap[14-i]["$salesforce_account_website"])
					}
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
				if i > 0 {
					assert.Condition(t, func() bool { return resp[i].LastActivity.Unix() <= resp[i-1].LastActivity.Unix() })
				}
				for _, prop := range timelinesConfig.UserConfig.TableProps {
					assert.NotNil(t, user.TableProps[prop])
				}
			}
			return true
		})
	}

	// 3. Segment with multiple $hubspot_company filters
	segmentPayload := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
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
		},
		Type: "$hubspot_company",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "$hubspot_company",
		SegmentId: segments["$hubspot_company"][0].Id,
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 3)
	assert.Condition(t, func() bool {
		filteredCompaniesNameHostNameMap := map[string]string{"Adapt.IO": "adapt.io", "Clientjoy Ads": "clientjoy.io", "AdPushup": "adpushup.com"}
		for i, user := range resp {
			assert.Condition(t, func() bool {
				for name, hostName := range filteredCompaniesNameHostNameMap {
					if name == user.Name && hostName == user.HostName {
						return true
					}
				}
				return false
			})
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

	// Search Filter:-
	payload = model.TimelinePayload{
		Source: "$hubspot_company",
		SearchFilter: []model.QueryProperty{
			{
				Entity:    "user_g",
				Type:      "categorical",
				Property:  "$hubspot_company_name",
				Operator:  "equals",
				Value:     "AdPushup",
				LogicalOp: "AND",
			},
			{
				Entity:    "user_g",
				Type:      "categorical",
				Property:  "$hubspot_company_name",
				Operator:  "equals",
				Value:     "Heyflow",
				LogicalOp: "OR",
			},
		},
	}

	w = sendGetProfileAccountRequest(r, project.ID, agent, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	resp = make([]model.Profile, 0)
	err = json.Unmarshal(jsonResponse, &resp)
	assert.Nil(t, err)
	assert.Equal(t, len(resp), 2)
	assert.Condition(t, func() bool {
		filteredCompaniesNameHostNameMap := map[string]string{"AdPushup": "adpushup.com", "Heyflow": "heyflow.app"}
		for i, user := range resp {
			assert.Condition(t, func() bool {
				for name, hostName := range filteredCompaniesNameHostNameMap {
					if name == user.Name && hostName == user.HostName {
						return true
					}
				}
				return false
			})
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

}

func sendGetProfileAccountRequest(r *gin.Engine, projectId int64, agent *model.Agent, payload model.TimelinePayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/profiles/accounts", projectId)).
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

	createdUserID, _ := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
		Group1ID:       "1",
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &isGroupUser,
	})
	projectID := project.ID
	accountID := createdUserID
	user, errCode := store.GetStore().GetUser(projectID, accountID)
	assert.Equal(t, user.ID, accountID)
	assert.Equal(t, http.StatusFound, errCode)
	group1, status := store.GetStore().CreateGroup(projectID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)

	// 10  Associated Users
	m := map[string]string{"$name": "Some Name"}
	userProps, err := json.Marshal(m)
	if err != nil {
		log.WithError(err).Fatal("Marshal error.")
	}
	properties = postgres.Jsonb{RawMessage: userProps}
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
	numUsers := 10
	for i := 1; i <= numUsers; i++ {

		jobTitle := "Boss"
		if i > 1 {
			jobTitle = "Employee"
		}
		userProps := map[string]interface{}{
			"$hubspot_contact_jobtitle": jobTitle,
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

		associatedUserId, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      projectID,
			Properties:     userPropsEncoded,
			IsGroupUser:    &isGroupUser,
			Group1ID:       "1",
			Group1UserID:   accountID,
			CustomerUserId: customerUserID,
			Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
		})

		user, errCode := store.GetStore().GetUser(project.ID, associatedUserId)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)

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
	assert.Equal(t, len(users), numUsers)

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileAccountDetailsRequest(r, projectID, agent, accountID)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := &model.AccountDetails{}
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Contains(t, resp.Name, "Freshworks")
		assert.Equal(t, resp.HostName, "google.com")
		assert.Equal(t, len(resp.AccountTimeline), 9)
		assert.NotNil(t, resp.LeftPaneProps)
		assert.Condition(t, func() bool {
			for i, property := range resp.LeftPaneProps {
				assert.Equal(t, props[i], property)
			}
			for i, property := range resp.Milestones {
				assert.Equal(t, props[i], property)
			}
			return true
		})
		assert.Condition(t, func() bool {
			assert.Condition(t, func() bool { return len(resp.AccountTimeline) > 0 })
			for _, userTimeline := range resp.AccountTimeline {
				assert.Condition(t, func() bool {
					assert.NotNil(t, userTimeline.UserId)
					assert.NotNil(t, userTimeline.UserName)
					for i, activity := range userTimeline.UserActivities {
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
						}
						assert.NotNil(t, activity.Timestamp)
						assert.Condition(t, func() bool { return activity.Timestamp <= uint64(time.Now().UTC().Unix()) })
						if i > 1 {
							if userTimeline.UserActivities[i].Timestamp > userTimeline.UserActivities[i-1].Timestamp {
								return false
							}
						}
						assert.Condition(t, func() bool {
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
									assert.Condition(t, func() bool { return i < len(lookInProps) })
								}
							}
							return true
						})
					}
					return true
				})
			}
			return true
		})

	})
}

func sendGetProfileAccountDetailsRequest(r *gin.Engine, projectId int64, agent *model.Agent, id string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/profiles/accounts/%s", projectId, id)).
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

	// creating a segment with only global properties

	startTimestamp := U.UnixTimeBeforeDuration(24 * 28 * time.Hour)
	segmentPayload := &model.SegmentPayload{
		Name:        "Name0",
		Description: "dummy info",
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
			Class:           model.QueryClassProfiles,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "user_profiles",
			Source:          "salesforce",
			TableProps:      []string{"$country", "$page_count"},
		},
		Type: "salesforce",
	}
	status, err := store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "salesforce",
		SegmentId: segments["salesforce"][0].Id,
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

	// creating a segment

	segmentPayload = &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[2],
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "user_profiles",
			Source:          "web",
			TableProps:      []string{"$country", "$page_count"},
		},
		Type: "web",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "web",
		SegmentId: segments["web"][0].Id,
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

	// creating a segment with only ewp
	segmentPayload = &model.SegmentPayload{
		Name:        "Name2",
		Description: "dummy info",
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "user_profiles",
			Source:          "web",
			TableProps:      []string{"$country", "$page_count"},
		},
		Type: "web",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	var id string
	for _, segment := range segments["web"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}

	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "web",
		SegmentId: id,
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

	// creating a segment

	segmentPayload = &model.SegmentPayload{
		Name:        "Name3",
		Description: "dummy info",
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "user_profiles",
			Source:          "salesforce",
			TableProps:      []string{"$country", "$page_count"},
		},
		Type: "web",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["web"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}
	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "salesforce",
		SegmentId: id,
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

	// creating a segment

	segmentPayload = &model.SegmentPayload{
		Name:        "Name4",
		Description: "dummy info",
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "user_profiles",
			Source:          "salesforce",
			TableProps:      []string{"$country", "$page_count"},
		},
		Type: "salesforce",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["salesforce"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}
	// add segmentId to timeline payload
	payload = model.TimelinePayload{
		Source:    "salesforce",
		SegmentId: id,
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
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	propertiesMap := []map[string]interface{}{
		{"$salesforce_account_name": "Pepper Content", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "peppercontent.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target"},
		{"$salesforce_account_name": "o9 Solutions", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "o9solutions.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "GoLinks Reporting", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "golinks.io", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "Cin7", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "cin7.com", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor"},
		{"$salesforce_account_name": "Repair Desk", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "repairdesk.co", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer"},
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$country": "US"},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$country": "US"},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$country": "US"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$country": "US"},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services", "$country": "US"},
	}

	// Create 5 Salesforce Accounts
	accounts := make([]model.User, 0)
	numUsers = 5
	groupUser := true
	timestamp := U.UnixTimeBeforeDuration(1 * time.Hour)
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
			Group2ID:       "2",
			CustomerUserId: fmt.Sprintf("sfuser%d@%s", i+1, propertiesMap[i]["$salesforce_account_website"]),
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
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
			Group1ID:       "1",
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
		val := i + (i % 2)
		trackPayload := SDK.TrackPayload{
			UserId:          createdUserID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            randomURLs[val],
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
	assert.Equal(t, len(accounts), 10)

	// Test Cases :-
	//1. gpb and ewp
	segmentPayload = &model.SegmentPayload{
		Name:        "Name5",
		Description: "dummy info",
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
					Name: randomURLs[0],
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
					Name: randomURLs[2],
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "account_profiles",
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
		Type: "$hubspot_company",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["$hubspot_company"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}

	payload = model.TimelinePayload{
		Source:    "$hubspot_company",
		SegmentId: id,
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

	segmentPayload = &model.SegmentPayload{
		Name:        "Name6",
		Description: "dummy info",
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[2],
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
					Name: randomURLs[0],
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "account_profiles",
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
		Type: "$hubspot_company",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["$hubspot_company"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}

	payload = model.TimelinePayload{
		Source:    "$hubspot_company",
		SegmentId: id,
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

	segmentPayload = &model.SegmentPayload{
		Name:        "Name7",
		Description: "dummy info",
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
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[4],
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "account_profiles",
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
		Type: "$hubspot_company",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["$hubspot_company"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}
	payload = model.TimelinePayload{
		Source:    "$hubspot_company",
		SegmentId: id,
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
	}

	segmentPayload = &model.SegmentPayload{
		Name:        "Name8",
		Description: "dummy info",
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[2],
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
					Name: randomURLs[2],
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "account_profiles",
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
		Type: "$hubspot_company",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["$hubspot_company"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}
	payload = model.TimelinePayload{
		Source:    "$hubspot_company",
		SegmentId: id,
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
	}

	segmentPayload = &model.SegmentPayload{
		Name:        "Name9",
		Description: "dummy info",
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name: randomURLs[1],
				},
				{
					Name: randomURLs[2],
				},
			},

			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
			Timezone:        "America/Vancouver",
			From:            startTimestamp,
			To:              time.Now().UTC().Unix(),
			Caller:          "account_profiles",
			Source:          "$hubspot_company",
			TableProps:      []string{"$country", "$hubspot_company_num_associated_contacts", "$hour_of_first_event"},
		},
		Type: "$hubspot_company",
	}
	status, err = store.GetStore().CreateSegment(project.ID, segmentPayload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
	// fetch segments
	segments, status = store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	for _, segment := range segments["$hubspot_company"] {
		if segment.Name == segmentPayload.Name {
			id = segment.Id
		}
	}

	payload = model.TimelinePayload{
		Source:    "$hubspot_company",
		SegmentId: id,
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
