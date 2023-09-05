package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
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
		{"$salesforce_account_name": "AdPushup", "$salesforce_account_billingcountry": "India", "$salesforce_account_website": "adpushup.com", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Target"},
		{"$salesforce_account_name": "Mad Street Den", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "madstreetden.com", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "Heyflow", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "heyflow.app", "$salesforce_account_sales_play": "Penetrate", "$salesforce_account_status": "Unknown"},
		{"$salesforce_account_name": "Clientjoy Ads", "$salesforce_account_billingcountry": "New Zealand", "$salesforce_account_website": "clientjoy.io", "$salesforce_account_sales_play": "Win", "$salesforce_account_status": "Vendor", "$salesforce_city": "New Delhi"},
		{"$salesforce_account_name": "Adapt.IO", "$salesforce_account_billingcountry": "US", "$salesforce_account_website": "adapt.io", "$salesforce_account_sales_play": "Shape", "$salesforce_account_status": "Customer"},
		{"$hubspot_company_name": "AdPushup", "$hubspot_company_country": "US", "$hubspot_company_domain": "adpushup.com", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "Technology, Information and Internet"},
		{"$hubspot_company_name": "Mad Street Den", "$hubspot_company_country": "US", "$hubspot_company_domain": "madstreetden.com", "$hubspot_company_num_associated_contacts": 100, "$hubspot_company_industry": "Software Development"},
		{"$hubspot_company_name": "Heyflow", "$hubspot_company_country": "Germany", "$hubspot_company_domain": "heyflow.app", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "Software Development"},
		{"$hubspot_company_name": "Clientjoy Ads", "$hubspot_company_country": "India", "$hubspot_company_domain": "clientjoy.io", "$hubspot_company_num_associated_contacts": 20, "$hubspot_company_industry": "IT Services"},
		{"$hubspot_company_name": "Adapt.IO", "$hubspot_company_country": "India", "$hubspot_company_domain": "adapt.io", "$hubspot_company_num_associated_contacts": 50, "$hubspot_company_industry": "IT Services"},
		{U.SIX_SIGNAL_NAME: "AdPushup", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "adpushup.com"},
		{U.SIX_SIGNAL_NAME: "Mad Street Den", U.SIX_SIGNAL_COUNTRY: "US", U.SIX_SIGNAL_DOMAIN: "madstreetden.com"},
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
		domID, _ := store.GetStore().CreateUser(&model.User{
			ProjectId:   project.ID,
			Source:      source,
			Group4ID:    fmt.Sprintf("domain%did.com", i),
			Properties:  domProperties,
			IsGroupUser: &groupUser,
		})
		_, errCode := store.GetStore().GetUser(project.ID, domID)
		assert.Equal(t, http.StatusFound, errCode)
		domainAccounts = append(domainAccounts, domID)
	}

	// 5 salesforce accounts
	numUsers := 5
	accounts := make([]model.User, 0)
	users := make([]model.User, 0)
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
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
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
		})
		user, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
		assert.Equal(t, http.StatusFound, errCode)
		users = append(users, *user)
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
			GroupAnalysis:   "All",
			Source:          "All",
			Type:            "unique_users",
			From:            time.Now().AddDate(0, 0, -28).Unix(),
			To:              time.Now().Unix(),
			Timezone:        "Asia/Kolkata",
			GlobalUserProperties: []model.QueryProperty{
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
		Type: "All",
	}

	w := createSegmentPostReq(r, *segment1, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound := false

	for _, segment := range getSegement["All"] {
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
			Source: "All",
		},
		Type: "All",
	}

	w = createSegmentPostReq(r, *segments3, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement2, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound1 := false

	for _, segment := range getSegement2["All"] {
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
			Source: "All",
		},
		Type: "All",
	}

	w = createSegmentPostReq(r, *segments5, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegementFinal, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	nameFound2 := false

	for _, segment := range getSegementFinal["All"] {
		if segments5.Name == segment.Name {
			nameFound2 = true
			break
		}
	}
	assert.True(t, nameFound2)

	// Process all user
	errCode := T.SegmentMarker(project.ID)
	assert.Equal(t, errCode, http.StatusOK)

	// to verify updated users
	// segment1 -> domain1id.com, domain2id.com
	// segment2 -> hbgroupuser3@heyflow.app, hbgroupuser4@clientjoy.io, hbgroupuser5@adapt.io
	// segment3 -> domain0id.com, domain1id.com, domain2id.com
	// segment4 -> hubspot@1user (customer_user_id)
	// segment5 -> domain1id.com, domain0id.com

	updatedUsers, status := store.GetStore().GetUsers(project.ID, 0, 35)
	assert.Equal(t, http.StatusFound, status)

	resultUsers := make(map[string][]model.User)

	for _, user := range updatedUsers {
		if user.Group4ID != "" && (user.Group4ID == "domain0id.com" || user.Group4ID == "domain1id.com" || user.Group4ID == "domain2id.com") {
			resultUsers["All"] = append(resultUsers["All"], user)
		} else if user.Group2ID != "" && (user.Group2ID == "hbgroupuser3@heyflow.app" || user.Group2ID == "hbgroupuser4@clientjoy.io" || user.Group2ID == "hbgroupuser5@adapt.io") {
			resultUsers["$hubspot_company"] = append(resultUsers["$hubspot_company"], user)
		} else if user.CustomerUserId != "" && (user.CustomerUserId == "hubspot@1user") {
			resultUsers["hubspot"] = append(resultUsers["hubspot"], user)
		}
	}

	assert.Equal(t, 3, len(resultUsers))

	allAccountsSegmentNameIDs := make(map[string]string)
	for _, segment := range getSegementFinal["All"] {
		allAccountsSegmentNameIDs[segment.Name] = segment.Id
	}

	for groupName, checkUser := range resultUsers {
		for _, user := range checkUser {
			associatedUsers, err := U.DecodePostgresJsonb(&user.AssociatedSegments)
			assert.Nil(t, err)
			if groupName == "All" {
				if user.Group4ID == "domain0id.com" {
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["User Group props"])
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["Group AND User Group props"])
				} else if user.Group4ID == "domain1id.com" {
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["User Group props"])
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["Group AND User Group props"])
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["All accounts segment"])
				} else if user.Group4ID == "domain2id.com" {
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["User Group props"])
					assert.Contains(t, *associatedUsers, allAccountsSegmentNameIDs["All accounts segment"])
				}
			} else if groupName == "$hubspot_company" {
				assert.Contains(t, *associatedUsers, getSegementFinal["$hubspot_company"][0].Id)
			} else if groupName == "hubspot" {
				assert.Contains(t, *associatedUsers, getSegementFinal["hubspot"][0].Id)
			}
		}
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
