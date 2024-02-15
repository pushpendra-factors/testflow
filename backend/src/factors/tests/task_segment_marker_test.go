package tests

import (
	"encoding/json"
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
	"testing"
	"time"

	T "factors/task"

	C "factors/config"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestTaskSegmentMarker(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	SegmentMarkerTest(t, project, agent, r)
}

func SegmentMarkerTest(t *testing.T, project *model.Project, agent *model.Agent, r *gin.Engine) {
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{SegmentMarkerLastRun: U.TimeNowZ().Add(time.Duration(-3) * time.Hour)})
	assert.Equal(t, errCode, http.StatusAccepted)

	// user property map
	userPropsMap := []map[string]interface{}{
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
		{"$browser": "Safari", "$city": "Paris", "$country": "France", "$device_type": "iPad", "$page_count": 120, "$session_spent_time": 3000},
		{"$browser": "Chromium", "$city": "New York", "$country": "US", "$device_type": "desktop", "$page_count": 110, "$session_spent_time": 2500},
		{"$browser": "Brave", "$city": "London", "$country": "UK", "$device_type": "iPad", "$page_count": 100, "$session_spent_time": 3000},
		{"$browser": "Firefox", "$city": "Dubai", "$country": "UAE", "$device_type": "desktop", "$page_count": 150, "$session_spent_time": 2100},
		{"$browser": "Chromium", "$city": "Delhi", "$country": "India", "$device_type": "macBook", "$page_count": 150, "$session_spent_time": 2100},
	}

	// Properties Map
	accountPropertiesMap := []map[string]interface{}{
		{"$salesforce_account_name": "AdPushup", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "adpushup.com", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target", "$csv_properties": "A"},
		{"$salesforce_account_name": "Mad Street Den", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "madstreetden.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown", "$csv_properties": "X"},
		{"$salesforce_account_name": "Heyflow", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "heyflow.app", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "Clientjoy Ads", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "clientjoy.io", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor", "$salesforce_city": "New Delhi"},
		{"$salesforce_account_name": "Adapt.IO", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "adapt.io", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer"},
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet", "$csv_properties": "A"},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development", "$csv_properties": "X"},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development", "$csv_properties": "A"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services", "$csv_properties": "X"},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "adpushup.com", "$csv_properties": "A"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "madstreetden.com", "$csv_properties": "X"},
		{U.SIX_SIGNAL_NAME: "Heyflow", U.SIX_SIGNAL_COUNTRY: "Germany", U.SIX_SIGNAL_DOMAIN: "heyflow.app"},
		{U.SIX_SIGNAL_NAME: "Clientjoy Ads", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "clientjoy.io"},
		{U.SIX_SIGNAL_NAME: "Adapt.IO", U.SIX_SIGNAL_COUNTRY: "India", U.SIX_SIGNAL_DOMAIN: "adapt.io"},
	}

	// groups creation
	group1, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group1)
	group2, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.NotNil(t, group2)
	assert.Equal(t, http.StatusCreated, status)
	group3, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.NotNil(t, group3)
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

	for i := 0; i < 5; i++ {
		var props postgres.Jsonb
		if i == 1 || i == 2 {
			props = postgres.Jsonb{RawMessage: json.RawMessage(fmt.Sprintf(`{"$domain_name":"%s",
			"$engagement_level":"Hot","$engagement_score":125.300000,"$joinTime":1681211371,
			"$total_enagagement_score":196.000000}`, accountPropertiesMap[i]["$salesforce_account_website"]))}
		} else {
			props = domProperties
		}
		domID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:   project.ID,
			Source:      source,
			Group4ID:    fmt.Sprintf("domain%did.com", i),
			Properties:  props,
			IsGroupUser: &groupUser,
		})
		_, errCode := store.GetStore().GetUser(project.ID, domID)
		assert.Equal(t, http.StatusFound, errCode)
		domainAccounts = append(domainAccounts, domID)
	}

	eventProperties := map[string]interface{}{
		U.EP_SALESFORCE_CAMPAIGNMEMBER_STATUS:  "CurrentStatus",
		U.EP_HUBSPOT_ENGAGEMENT_FROM:           "Somewhere",
		U.EP_HUBSPOT_ENGAGEMENT_TYPE:           "Some Engagement Type",
		U.EP_HUBSPOT_ENGAGEMENT_MEETINGOUTCOME: "Some Outcome",
		U.EP_SALESFORCE_CAMPAIGN_NAME:          "Some Salesforce Campaign Name",
	}

	// 5 salesforce accounts
	numUsers := 5
	accounts := make([]model.User, 0)
	users := make([]model.User, 0)
	lastEventTime := time.Now().Add(time.Duration(-15) * time.Minute)
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(accountPropertiesMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSalesforce)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:    project.ID,
			Source:       source,
			Group1ID:     fmt.Sprintf("sfgroupuser%d@%s", i+1, accountPropertiesMap[i]["$salesforce_account_website"]),
			Group4UserID: domainAccounts[i],
			Properties:   properties,
			IsGroupUser:  &groupUser,
			LastEventAt:  &lastEventTime,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// 5 users associated to the account
		propertiesJSON, err = json.Marshal(userPropsMap[i])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
			Properties:     properties,
			Group1ID:       "1",
			Group1UserID:   account.ID,
			Group4UserID:   domainAccounts[i],
			CustomerUserId: fmt.Sprintf("salesforce@%duser", (i%5)+1),
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)

		if i > 1 {
			continue
		}

		// SalesForce Group Events
		timestamp := U.UnixTimeBeforeDuration(time.Duration(1+i) * time.Hour)
		trackPayload := SDK.TrackPayload{
			UserId:        createdUserID,
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

	// Create 5 Hubspot Companies
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(accountPropertiesMap[i+5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceHubspot)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:    project.ID,
			Source:       source,
			Group2ID:     fmt.Sprintf("hbgroupuser%d@%s", i+1, accountPropertiesMap[i+5]["$hubspot_company_domain"]),
			Group4UserID: domainAccounts[i],
			Properties:   properties,
			IsGroupUser:  &groupUser,
			LastEventAt:  &lastEventTime,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// 5 users associated to the account
		propertiesJSON, err = json.Marshal(userPropsMap[i+5])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceHubspot),
			Properties:     properties,
			Group2UserID:   account.ID,
			Group4UserID:   domainAccounts[i],
			CustomerUserId: fmt.Sprintf("hubspot@%duser", (i%5)+1),
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
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

		if i > 1 {
			continue
		}

		// Hubspot Group Events
		trackPayload = SDK.TrackPayload{
			UserId:        createdUserID,
			CreateUser:    false,
			IsNewUser:     false,
			Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
			Timestamp:     timestamp,
			ProjectId:     project.ID,
			Auto:          false,
			RequestSource: model.UserSourceHubspot,
		}
		status, response = SDK.Track(project.ID, &trackPayload, false, SDK.SourceJSSDK, "")
		assert.NotEmpty(t, response)
		assert.Equal(t, http.StatusOK, status)

		trackPayload = SDK.TrackPayload{
			UserId:        createdUserID,
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

	}

	// Create 5 Six Signal Domains
	for i := 0; i < numUsers; i++ {
		propertiesJSON, err := json.Marshal(accountPropertiesMap[i+10])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties := postgres.Jsonb{RawMessage: propertiesJSON}
		source := model.GetRequestSourcePointer(model.UserSourceSixSignal)

		createdUserID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:    project.ID,
			Source:       source,
			Group3ID:     fmt.Sprintf("6siguser%d@%s", i+1, accountPropertiesMap[i+10][U.SIX_SIGNAL_DOMAIN]),
			Group4UserID: domainAccounts[i],
			Properties:   properties,
			IsGroupUser:  &groupUser,
			LastEventAt:  &lastEventTime,
		})
		account, errCode := store.GetStore().GetUser(project.ID, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		accounts = append(accounts, *account)

		// 5 users associated to the account
		propertiesJSON, err = json.Marshal(userPropsMap[i+10])
		if err != nil {
			log.WithError(err).Fatal("Marshal error.")
		}
		properties = postgres.Jsonb{RawMessage: propertiesJSON}
		createdUserID1, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			Source:         model.GetRequestSourcePointer(model.UserSourceSixSignal),
			Properties:     properties,
			Group3UserID:   account.ID,
			Group4UserID:   domainAccounts[i],
			CustomerUserId: fmt.Sprintf("sixsignal@%duser", (i%5)+1),
			LastEventAt:    &lastEventTime,
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
	}
	assert.Equal(t, len(accounts), 15)

	// Segment creation

	// 1. All Accounts segment (different sources, gup)
	segment1 := &model.SegmentPayload{
		Name: "All accounts segment",
		Query: model.Query{
			Caller:          "account_profiles",
			Class:           "events",
			EventsCondition: "any_given_event",
			GroupAnalysis:   "$domains",
			Source:          "$domains",
			Type:            "unique_users",
			From:            time.Now().AddDate(0, 0, -28).Unix(),
			To:              time.Now().Unix(),
			Timezone:        "Asia/Kolkata",
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$domain_name",
					Operator:  "equals",
					Value:     "madstreetden.com",
					LogicalOp: "AND",
					GroupName: "$domains",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$domain_name",
					Operator:  "equals",
					Value:     "heyflow.app",
					LogicalOp: "OR",
					GroupName: "$domains",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_billingcountry",
					Operator:  "equals",
					Value:     "US",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_billingcountry",
					Operator:  "equals",
					Value:     "New Zealand",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$hubspot_company_industry",
					Operator:  "equals",
					Value:     "Software Development",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  U.SIX_SIGNAL_COUNTRY,
					Operator:  "equals",
					Value:     "US",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  U.SIX_SIGNAL_COUNTRY,
					Operator:  "equals",
					Value:     "Germany",
					LogicalOp: "OR",
				},
			},
		},
		Type: "$domains",
	}

	w := createSegmentPostReq(r, *segment1, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound := false

	for _, segment := range getSegement["$domains"] {
		if segment1.Name == segment.Name {
			nameFound = true
			break
		}
	}
	assert.True(t, nameFound)

	// 2. Hubspot company segment
	segment2 := &model.SegmentPayload{
		Name: "Hubspot Company segment",
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
					Value:     "Germany",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$hubspot_company_num_associated_contacts",
					Operator:  model.GreaterThanOrEqualOpStr,
					Value:     "20",
					LogicalOp: "AND",
				},
			},
			Caller: "account_profiles",
		},
		Type: "$hubspot_company",
	}

	w = createSegmentPostReq(r, *segment2, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement1, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, segment2.Name, getSegement1["$hubspot_company"][0].Name)

	// 3. all accounts user props

	segments3 := &model.SegmentPayload{
		Name: "User Group props",
		Query: model.Query{
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
			Source: "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segments3, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement2, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound1 := false

	for _, segment := range getSegement2["$domains"] {
		if segments3.Name == segment.Name {
			nameFound1 = true
			break
		}
	}
	assert.True(t, nameFound1)

	// 4. User Segment

	segments4 := &model.SegmentPayload{
		Name: "User Segment",
		Query: model.Query{
			Source: "hubspot",
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "contains",
					Value:     "India",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "France",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_g",
					Type:      "numerical",
					Property:  "$session_spent_time",
					Operator:  "equals",
					Value:     "3000",
					LogicalOp: "AND",
				},
			},
			GroupAnalysis: "users",
		},
		Type: "hubspot",
	}

	w = createSegmentPostReq(r, *segments4, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement3, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, segments4.Name, getSegement3["hubspot"][0].Name)

	// 5. all accounts user props

	segments5 := &model.SegmentPayload{
		Name: "Group AND User Group props",
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_name",
					Operator:  "equals",
					Value:     "AdPushup",
					LogicalOp: "AND",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$salesforce_account_name",
					Operator:  "equals",
					Value:     "Mad Street Den",
					LogicalOp: "OR",
				},
				{
					Entity:    "user_group",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "equals",
					Value:     "India",
					LogicalOp: "AND",
				},
			},
			Source: "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segments5, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement4, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound2 := false

	for _, segment := range getSegement4["$domains"] {
		if segments5.Name == segment.Name {
			nameFound2 = true
			break
		}
	}
	assert.True(t, nameFound2)

	// 6. Hubspot user event segment
	segment6 := &model.SegmentPayload{
		Name: "Hubspot User Performed Event",
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$country",
					Operator:  "contains",
					Value:     "India",
					LogicalOp: "AND",
				},
			},
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.EVENT_NAME_SESSION,
					GroupAnalysis: "Most Recent",
					Properties: []model.QueryProperty{
						{
							Entity:    "user",
							Type:      "categorical",
							Property:  "$country",
							Operator:  "contains",
							Value:     "India",
							LogicalOp: "AND",
						},
					},
				},
				{
					Name:          U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL,
					GroupAnalysis: "Most Recent",
				},
				{
					Name:          U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_CREATED,
					GroupAnalysis: "Most Recent",
				},
			},
			Caller:          model.USER_PROFILES,
			EventsCondition: model.EventCondAllGivenEvent,
			GroupAnalysis:   "users",
		},
		Type: "hubspot",
	}

	w = createSegmentPostReq(r, *segment6, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement5, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound3 := false

	for _, segment := range getSegement5["hubspot"] {
		if segment6.Name == segment.Name {
			nameFound3 = true
			break
		}
	}
	assert.True(t, nameFound3)

	// 7. All event segment
	segment7 := &model.SegmentPayload{
		Name: "Hubspot Group Performed Event",
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis: "Most Recent",
				},
				{
					Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED,
					GroupAnalysis: "Most Recent",
				},
			},
			Caller:          model.ACCOUNT_PROFILES,
			EventsCondition: model.EventCondAllGivenEvent,
			GroupAnalysis:   "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segment7, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement6, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound4 := false

	for _, segment := range getSegement6["$domains"] {
		if segment7.Name == segment.Name {
			nameFound4 = true
			break
		}
	}
	assert.True(t, nameFound4)

	// 8. All event different segment
	segment8 := &model.SegmentPayload{
		Name: "All Group Performed Event",
		Query: model.Query{
			EventsWithProperties: []model.QueryEventWithProperties{
				{
					Name:          U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
					GroupAnalysis: "Most Recent",
				},
				{
					Name:          U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
					GroupAnalysis: "Most Recent",
				},
				{
					Name:          U.EVENT_NAME_SESSION,
					GroupAnalysis: "Most Recent",
					Properties: []model.QueryProperty{
						{
							Entity:    "user",
							Type:      "categorical",
							Property:  "$browser",
							Operator:  "contains",
							Value:     "Edge",
							LogicalOp: "AND",
						},
					},
				},
			},
			Caller:          model.ACCOUNT_PROFILES,
			EventsCondition: model.EventCondAllGivenEvent,
			GroupAnalysis:   "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segment8, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement7, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound5 := false

	for _, segment := range getSegement7["$domains"] {
		if segment8.Name == segment.Name {
			nameFound5 = true
			break
		}
	}
	assert.True(t, nameFound5)

	//uploading file
	csvFilePath := "/Users/apple/repos/factors/backend/src/factors/tests/data"
	csvFilename := "test_inlist.csv"

	csvFile, err := ioutil.ReadFile(csvFilePath + "/" + csvFilename)
	if err != nil {
		fmt.Println(err)
	}
	w = sendUploadListForFilters(r, project.ID, agent, csvFile, csvFilename)
	assert.Equal(t, http.StatusOK, w.Code)
	var res map[string]string
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&res); err != nil {
		assert.NotNil(t, nil, err)
	}

	// 9. All accounts segment With inList Support
	segment9 := &model.SegmentPayload{
		Name: "All accounts segment With inList Support",
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$csv_properties",
					Operator:  model.InList,
					Value:     res["file_reference"],
					LogicalOp: "AND",
				},
			},
			Source: "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segment9, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement8, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound6 := false

	for _, segment := range getSegement8["$domains"] {
		if segment9.Name == segment.Name {
			nameFound6 = true
			break
		}
	}
	assert.True(t, nameFound6)

	//uploading file
	csvFilename = "test_notinlist.csv"

	csvFile, err = ioutil.ReadFile(csvFilePath + "/" + csvFilename)
	if err != nil {
		fmt.Println(err)
	}
	w = sendUploadListForFilters(r, project.ID, agent, csvFile, csvFilename)
	assert.Equal(t, http.StatusOK, w.Code)
	decoder = json.NewDecoder(w.Body)
	if err := decoder.Decode(&res); err != nil {
		assert.NotNil(t, nil, err)
	}

	// 10. All accounts segment With NotinList Support
	segment10 := &model.SegmentPayload{
		Name: "All accounts segment With notInList Support",
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$csv_properties",
					Operator:  model.NotInList,
					Value:     res["file_reference"],
					LogicalOp: "AND",
				},
			},
			Source: "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segment10, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 11. Domain level filters
	segment11 := &model.SegmentPayload{
		Name: "Domain Level Support",
		Query: model.Query{
			GlobalUserProperties: []model.QueryProperty{
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$domain_name",
					Operator:  "equals",
					Value:     "madstreetden.com",
					LogicalOp: "AND",
					GroupName: "$domains",
				},
				{
					Entity:    "user_g",
					Type:      "categorical",
					Property:  "$domain_name",
					Operator:  "equals",
					Value:     "heyflow.app",
					LogicalOp: "OR",
					GroupName: "$domains",
				},
			},
			Source: "$domains",
		},
		Type: "$domains",
	}

	w = createSegmentPostReq(r, *segment11, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegementFinal, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound7 := false

	for _, segment := range getSegementFinal["$domains"] {
		if segment10.Name == segment.Name {
			nameFound7 = true
			break
		}
	}
	assert.True(t, nameFound7)

	// Process all user
	errCode = T.SegmentMarker(project.ID)
	assert.Equal(t, errCode, http.StatusOK)

	status, updatedUsers, associatedSegmentsList := store.GetStore().FetchAssociatedSegmentsFromUsers(project.ID)
	assert.Equal(t, http.StatusFound, status)

	allAccountsSegmentNameIDs := make(map[string]string)
	for _, segmentList := range getSegementFinal {
		for _, segment := range segmentList {
			allAccountsSegmentNameIDs[segment.Name] = segment.Id
		}
	}

	// to verify updated users
	// segment1 -> domain1id.com, domain2id.com
	// segment2 -> hbgroupuser3@heyflow.app, hbgroupuser4@clientjoy.io, hbgroupuser5@adapt.io
	// segment3 -> domain0id.com, domain1id.com, domain2id.com
	// segment4 -> hubspot@1user (customer_user_id)
	// segment5 -> domain1id.com, domain0id.com
	// segment7 -> domain1id.com, domain0id.com
	// segment6 -> domain1id.com
	// segment8 -> hubspot@5user
	// segment9 -> domain0id.com, domain2id.com
	// segment10 -> domain0id.com, domain2id.com, domain4id.com
	// segment11 -> domain1id.com, domain2id.com

	for index, checkUser := range updatedUsers {
		if checkUser.Group4ID == "domain0id.com" {
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["User Group props"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Group AND User Group props"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Hubspot Group Performed Event"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment With inList Support"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment With notInList Support"])
		} else if checkUser.Group4ID == "domain1id.com" {
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["User Group props"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Group AND User Group props"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All Group Performed Event"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Hubspot Group Performed Event"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Domain Level Support"])
		} else if checkUser.Group4ID == "domain2id.com" {
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["User Group props"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment With inList Support"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment With notInList Support"])
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Domain Level Support"])
		} else if checkUser.Group4ID == "domain4id.com" {
			assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["All accounts segment With notInList Support"])
		} else if checkUser.Group2ID == "hbgroupuser3@heyflow.app" || checkUser.Group2ID == "hbgroupuser4@clientjoy.io" || checkUser.Group2ID == "hbgroupuser5@adapt.io" {
			assert.Contains(t, associatedSegmentsList[index], getSegementFinal["$hubspot_company"][0].Id)
		}
		//  else if checkUser.CustomerUserId == "hubspot@5user" {
		// 	assert.Contains(t, associatedSegmentsList[index], allAccountsSegmentNameIDs["Hubspot User Performed Event"])
		// }
	}
}

func createSegmentPostReq(r *gin.Engine, request model.SegmentPayload, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/segments", projectId)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create segment req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
