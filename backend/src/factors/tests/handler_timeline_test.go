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
	sourceToUserCountMap := map[string]interface{}{"All": 15, model.UserSourceSalesforceString: 2, model.UserSourceHubspotString: 3}
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

	// 3. UserSourceWeb (Segment Applied, no filters)
	// creating a segment
	segmentPayload := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query: model.SegmentQuery{
			GlobalProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$browser",
					Operator:  "equals",
					Value:     "Chrome",
					LogicalOp: "AND",
				},
			},
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
		Query: model.SegmentQuery{
			GlobalProperties: []model.QueryProperty{
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

	propsToShow := []string{"$email", "$page_count", "$user_id", "$name", "$session_spent_time"}
	timelinesConfig.UserConfig.LeftpaneProps = propsToShow

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	props := map[string]interface{}{
		"$name":               "Cameron Williomson",
		"$company":            "Freshworks",
		"$country":            "Australia",
		"$session_count":      8,
		"$session_spent_time": 500,
		"$page_count":         10,
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

	// Event 6 : Campaign Member Updated
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
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
	eventProperties[U.EP_HUBSPOT_FORM_SUBMISSION_TIMESTAMP] = timestamp
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
	eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP] = timestamp
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

	// Event 10 : Engagement Meeting Updated
	timestamp = timestamp - 10000
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
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

	// Event 11 : Engagement Call Created
	timestamp = timestamp - 10000
	eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP] = timestamp
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
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

	// Event 12 : Engagement Call Updated
	timestamp = timestamp - 10000
	eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP] = timestamp
	trackPayload = SDK.TrackPayload{
		UserId:          user.ID,
		CreateUser:      false,
		IsNewUser:       false,
		Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
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
			return true
		})
		assert.NotNil(t, resp.GroupInfos)
		assert.Condition(t, func() bool { return len(resp.GroupInfos) <= 4 })
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
					} else if activity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED {
						assert.Equal(t, activity.AliasName, fmt.Sprintf("Interacted with %s", eventProperties[U.EP_SALESFORCE_CAMPAIGN_NAME]))
					} else if activity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION {
						assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_FORM_SUBMISSION_TITLE]))
					} else if activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL {
						assert.Equal(t, activity.AliasName, fmt.Sprintf("%s: %s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TYPE], eventProperties[U.EP_HUBSPOT_ENGAGEMENT_SUBJECT]))
					} else if activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED ||
						activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED ||
						activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED ||
						activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED {
						assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TITLE]))
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
			TableProps: []string{"$page_count", "$browser"},
		},
	}

	tlConfigEncoded, err := U.EncodeStructTypeToPostgresJsonb(timelinesConfig)
	assert.Nil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{TimelinesConfig: tlConfigEncoded})
	assert.Equal(t, errCode, http.StatusAccepted)

	customerEmail := "@example.com"

	// Create 5 Users with Properties.
	accounts := make([]model.User, 0)
	numUsers := 6
	groupUser := true

	companies := []string{"FactorsAI", "Accenture", "Talentica", "Honeywell", "Meesho", ""}
	websites := []string{"factors.ai", "accenture.com", "talentica.com", "honeywell.com", "meesho.com"}
	countries := []string{"India", "Ireland", "India", "US", "India", "US"}
	browsers := []string{"Chrome", "Brave", "Firefox", "Edge", "Safari", "Opera"}
	for i := 0; i < numUsers; i++ {
		var propertiesMap map[string]interface{}
		if i%2 == 0 {
			propertiesMap = map[string]interface{}{
				U.GP_SALESFORCE_ACCOUNT_NAME:           companies[i],
				U.GP_SALESFORCE_ACCOUNT_BILLINGCOUNTRY: countries[i],
				U.GP_SALESFORCE_ACCOUNT_WEBSITE:        websites[i],
			}
		} else {
			if i == 5 {
				propertiesMap = map[string]interface{}{
					U.GP_HUBSPOT_COMPANY_COUNTRY:                 countries[i],
					U.GP_HUBSPOT_COMPANY_NUM_ASSOCIATED_CONTACTS: i * 2,
				}
			} else {
				propertiesMap = map[string]interface{}{
					U.GP_HUBSPOT_COMPANY_NAME:                    companies[i],
					U.GP_HUBSPOT_COMPANY_COUNTRY:                 countries[i],
					U.GP_HUBSPOT_COMPANY_DOMAIN:                  websites[i],
					U.GP_HUBSPOT_COMPANY_NUM_ASSOCIATED_CONTACTS: i * 2,
				}
			}

		}
		propertiesMap["$page_count"] = 1000 + i
		propertiesMap["$browser"] = browsers[i]
		propertiesJSON, err := json.Marshal(propertiesMap)
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		var source *int
		var gr1 string
		var gr2 string
		if i%2 == 0 {
			source = model.GetRequestSourcePointer(model.UserSourceSalesforce)
			gr1 = ""
			gr2 = "2"
		} else {
			source = model.GetRequestSourcePointer(model.UserSourceHubspot)
			gr1 = "1"
			gr2 = ""
		}

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         source,
			Group1ID:       gr1,
			Group2ID:       gr2,
			CustomerUserId: "user" + strconv.Itoa(i) + customerEmail,
			Properties:     properties,
			IsGroupUser:    &groupUser,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)
	}
	assert.Equal(t, len(accounts), numUsers)

	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)

	var payload model.TimelinePayload
	payload.Source = "All"

	// Without filters
	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), 6)
		assert.Condition(t, func() bool {
			for index, user := range resp {
				sort.Strings(companies)
				i := sort.SearchStrings(companies, user.Name)
				assert.Condition(t, func() bool { return i < len(companies) })
				sort.Strings(websites)
				i = sort.SearchStrings(websites, user.HostName)
				assert.Condition(t, func() bool { return i < len(websites) })
				sort.Strings(countries)
				assert.Condition(t, func() bool { return i < len(countries) })
				assert.NotNil(t, user.LastActivity)
				for _, prop := range timelinesConfig.AccountConfig.TableProps {
					assert.NotNil(t, user.TableProps[prop])
				}
				if index > 0 {
					assert.Condition(t, func() bool { return resp[index].LastActivity.Unix() <= resp[index-1].LastActivity.Unix() })
				}
			}
			return true
		})
	})

	//With filters
	filters := model.QueryProperty{
		Entity:    "user_g",
		Type:      "categorical",
		Property:  U.GP_SALESFORCE_ACCOUNT_BILLINGCOUNTRY,
		Operator:  "equals",
		Value:     "India",
		LogicalOp: "AND",
	}
	payload.Filters = append(payload.Filters, filters)
	filters = model.QueryProperty{
		Entity:    "user_g",
		Type:      "categorical",
		Property:  U.GP_HUBSPOT_COMPANY_COUNTRY,
		Operator:  "equals",
		Value:     "US",
		LogicalOp: "OR",
	}
	payload.Filters = append(payload.Filters, filters)

	t.Run("Success", func(t *testing.T) {
		w := sendGetProfileAccountRequest(r, project.ID, agent, payload)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		resp := make([]model.Profile, 0)
		err := json.Unmarshal(jsonResponse, &resp)
		assert.Nil(t, err)
		assert.Equal(t, len(resp), 5)
		assert.Condition(t, func() bool {
			for index, user := range resp {
				sort.Strings(companies)
				i := sort.SearchStrings(companies, user.Name)
				assert.Condition(t, func() bool { return i < len(companies) })
				sort.Strings(websites)
				i = sort.SearchStrings(websites, user.HostName)
				assert.Condition(t, func() bool { return i < len(websites) })
				sort.Strings(countries)
				assert.Condition(t, func() bool { return i < len(countries) })
				assert.NotNil(t, user.LastActivity)
				for _, prop := range timelinesConfig.AccountConfig.TableProps {
					assert.NotNil(t, user.TableProps[prop])
				}
				if index > 0 {
					assert.Condition(t, func() bool { return resp[index].LastActivity.Unix() <= resp[index-1].LastActivity.Unix() })
				}
			}
			return true
		})
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

	propsToShow := []string{"$hubspot_company_industry", "$hubspot_company_country"}
	timelinesConfig.AccountConfig.LeftpaneProps = propsToShow
	timelinesConfig.AccountConfig.UserProp = "$hubspot_contact_jobtitle"

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
		Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
		Group1ID:       "1",
		CustomerUserId: customerEmail,
		Properties:     properties,
		IsGroupUser:    &boolTrue,
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
	}
	randomURL := RandomURL()
	customerEmail = "@example.com"
	boolTrue = false
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
			IsGroupUser:    &boolTrue,
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

		// Event 6 : Campaign Member Updated
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED,
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
		eventProperties[U.EP_HUBSPOT_FORM_SUBMISSION_TIMESTAMP] = timestamp
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
		eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP] = timestamp
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

		// Event 10 : Engagement Meeting Updated
		timestamp = timestamp - 10000
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED,
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

		// Event 11 : Engagement Call Created
		timestamp = timestamp - 10000
		eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP] = timestamp
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED,
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

		// Event 12 : Engagement Call Updated
		timestamp = timestamp - 10000
		eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TIMESTAMP] = timestamp
		trackPayload = SDK.TrackPayload{
			UserId:          user.ID,
			CreateUser:      false,
			IsNewUser:       false,
			Name:            U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED,
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
							} else if activity.EventName == U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_UPDATED {
								assert.Equal(t, activity.AliasName, fmt.Sprintf("Interacted with %s", eventProperties[U.EP_SALESFORCE_CAMPAIGN_NAME]))
							} else if activity.EventName == U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION {
								assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_FORM_SUBMISSION_TITLE]))
							} else if activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL {
								assert.Equal(t, activity.AliasName, fmt.Sprintf("%s: %s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TYPE], eventProperties[U.EP_HUBSPOT_ENGAGEMENT_SUBJECT]))
							} else if activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED ||
								activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED ||
								activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED ||
								activity.EventName == U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED {
								assert.Equal(t, activity.AliasName, fmt.Sprintf("%s", eventProperties[U.EP_HUBSPOT_ENGAGEMENT_TITLE]))
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
