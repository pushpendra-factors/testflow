package tests

import (
	"net/http"
	"testing"

	"encoding/json"
	H "factors/handler"
	M "factors/model"
	U "factors/util"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func createEventWithSession(projectId uint64, eventName string, userId string, timestamp int64, userPropertiesId string) int {
	userSession, errCode := M.CreateOrGetSessionEvent(projectId, userId, false, false, timestamp,
		&U.PropertiesMap{}, &U.PropertiesMap{}, userPropertiesId)
	if errCode != http.StatusCreated {
		return errCode
	}

	userEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: projectId, Name: eventName})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return errCode
	}

	_, errCode = M.CreateEvent(&M.Event{ProjectId: projectId, EventNameId: userEventName.ID, UserId: userId, Timestamp: timestamp, SessionId: &userSession.ID})
	if errCode != http.StatusCreated {
		return errCode
	}

	return http.StatusCreated
}

func TestAttributionModel(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

	//Should return error for non adwords customer account id
	result, err := M.ExecuteAttributionQuery(project.ID, &M.AttributionQuery{})
	assert.Nil(t, result)
	assert.NotNil(t, err)

	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	assert.Equal(t, http.StatusAccepted, errCode)
	value := []byte(`{"cost": "0","clicks": "0","campaign_id":"123456","impressions": "0", "campaign_name": "test"}`)
	document := &M.AdwordsDocument{
		ProjectId:         project.ID,
		CustomerAccountId: customerAccountId,
		TypeAlias:         "campaign_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{value},
	}

	status := M.CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	user1Properties := make(map[string]interface{})
	user1Properties[U.UP_INITIAL_CAMPAIGN] = 123456
	user1PropertiesBytes, _ := json.Marshal(user1Properties)

	user2Properties := make(map[string]interface{})
	user2Properties[U.UP_INITIAL_CAMPAIGN] = 54321
	user2PropertiesBytes, _ := json.Marshal(user2Properties)

	user3Properties := make(map[string]interface{})
	user3Properties[U.UP_INITIAL_CAMPAIGN] = 54321
	user3PropertiesBytes, _ := json.Marshal(user3Properties)
	/*
		timestamp(t)
		t				user1 ->first session + event intial_campaign -> 123456
		t+3day			user2 ->first session + event intial_campaign -> 54321
		t+3day			user3 ->first session intial_campaign -> 54321
		t+5day			user1 ->session + event latest_campaign -> 1234567
		t+5day			user2 ->session + event latest_campaign -> 123456
	*/
	timestamp := int64(1589068800)
	day := int64(86400)
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{user1PropertiesBytes}, JoinTimestamp: timestamp})
	assert.NotNil(t, user1)
	assert.Equal(t, http.StatusCreated, errCode)
	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp, user1.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{user2PropertiesBytes}, JoinTimestamp: timestamp})
	assert.NotNil(t, user2)
	assert.Equal(t, http.StatusCreated, errCode)

	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{user3PropertiesBytes}, JoinTimestamp: timestamp})
	assert.NotNil(t, user3)
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp+3*day, user1.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp+3*day, user2.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateOrGetSessionEvent(project.ID, user3.ID, false, false, timestamp+3*day,
		&U.PropertiesMap{}, &U.PropertiesMap{}, user3.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	//Update user1 and user2 properties with latest campaign

	user1Properties[U.UP_LATEST_CAMPAIGN] = 1234567
	user1PropertiesBytes, _ = json.Marshal(user1Properties)

	user2Properties[U.UP_LATEST_CAMPAIGN] = 123456
	user2PropertiesBytes, _ = json.Marshal(user2Properties)

	user1NewPropertiesId, status := M.UpdateUserProperties(project.ID, user1.ID, &postgres.Jsonb{user1PropertiesBytes}, timestamp+5*day)
	assert.Equal(t, status, http.StatusAccepted)
	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp+5*day, user1NewPropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	user2NewPropertiesId, status := M.UpdateUserProperties(project.ID, user2.ID, &postgres.Jsonb{user2PropertiesBytes}, timestamp+5*day)
	assert.Equal(t, http.StatusAccepted, status)
	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp+5*day, user2NewPropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryFirstTouchCampaignNoLookbackDays", func(t *testing.T) {
		attributionData := make(map[string]*M.AttributionData)
		query := &M.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 10*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        "event1",
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)
		//Possibly 2 user (user1(123456), user2(54321)), user3 only session was created
		err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), attributionData["54321"].ConversionEventCount)
		assert.Equal(t, int64(1), attributionData["123456"].ConversionEventCount)

		//All users
		err = M.AddWebsiteVisitorsByEventName(project.ID, attributionData, query.From, query.To, userProperty)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), attributionData["54321"].WebsiteVisitors)
		assert.Equal(t, int64(1), attributionData["123456"].WebsiteVisitors)

		//Only campaign id 123456 has matching adword document
		_, err = M.AddPerformanceReportByCampaign(project.ID, attributionData, query.From, query.To, &customerAccountId)
		assert.Nil(t, err)
		assert.Equal(t, "test", attributionData["123456"].Name)
		assert.Equal(t, "", attributionData["54321"].Name)

	})

	t.Run("AttributionQueryLastTouchCampaignNoLookbackDays", func(t *testing.T) {
		attributionData := make(map[string]*M.AttributionData)
		query := &M.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 10*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
			ConversionEvent:        "event1",
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)

		//Only user1 and user2 latest campaign set
		err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), attributionData["123456"].ConversionEventCount)
		assert.Equal(t, 0, len(attributionData["123456"].LinkedEventsCount))
		assert.Equal(t, int64(1), attributionData["1234567"].ConversionEventCount)
		assert.Equal(t, 0, len(attributionData["1234567"].LinkedEventsCount))

		//Only user1 and user2 latest campaign set
		err = M.AddWebsiteVisitorsByEventName(project.ID, attributionData, query.From, query.To, userProperty)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), attributionData["123456"].WebsiteVisitors)
		assert.Equal(t, int64(1), attributionData["1234567"].WebsiteVisitors)

		//Only campaign id 123456 has matching adword document
		_, err = M.AddPerformanceReportByCampaign(project.ID, attributionData, query.From, query.To, &customerAccountId)
		assert.Nil(t, err)
		assert.Equal(t, "test", attributionData["123456"].Name)
		assert.Equal(t, "", attributionData["1234567"].Name)
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", user1.ID, timestamp+6*day, user1NewPropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestLastTouchLinkedEventNoLookbackDays", func(t *testing.T) {
		attributionData := make(map[string]*M.AttributionData)
		query := &M.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 10*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
			LinkedEvents:           []string{"event2"},
			ConversionEvent:        "event1",
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)

		// 1 linked event count by user1
		err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), attributionData["123456"].ConversionEventCount)
		assert.Equal(t, 1, len(attributionData["123456"].LinkedEventsCount))
		assert.Equal(t, int64(0), attributionData["123456"].LinkedEventsCount[0])
		assert.Equal(t, int64(1), attributionData["1234567"].ConversionEventCount)
		assert.Equal(t, 1, len(attributionData["1234567"].LinkedEventsCount))
		assert.Equal(t, int64(1), attributionData["1234567"].LinkedEventsCount[0])
	})

	t.Run("TestFirstTouchCampaignWithLookbackDays", func(t *testing.T) {
		attributionData := make(map[string]*M.AttributionData)
		query := &M.AttributionQuery{
			From:                   timestamp + 4*day,
			To:                     timestamp + 10*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        "event1",
			LinkedEvents:           []string{"event2"},
			LoopbackDays:           2,
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)

		//Should only have user2 with no 0 linked event count
		err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(attributionData))
		assert.Equal(t, int64(1), attributionData["54321"].ConversionEventCount)
		assert.Equal(t, 1, len(attributionData["54321"].LinkedEventsCount))
		assert.Equal(t, int64(0), attributionData["54321"].LinkedEventsCount[0])
	})
}

func TestAttributionLastTouchWithLookbackWindow(t *testing.T) {

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	/*
		t+3day -> user event
		t+4day -> session event
		t+5day -> user event
		t+8day -> session event
	*/
	timestamp := int64(1589068800)
	day := int64(86400)

	user1Properties := make(map[string]interface{})
	user1Properties[U.UP_LATEST_CAMPAIGN] = 123456
	user1PropertiesBytes, _ := json.Marshal(user1Properties)
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{user1PropertiesBytes}, JoinTimestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, user1)

	_, errCode = M.CreateOrGetSessionEvent(project.ID, user1.ID, false, false, timestamp+4*day,
		&U.PropertiesMap{}, &U.PropertiesMap{}, user1.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)
	userEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "event1"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: user1.ID, Timestamp: timestamp + 3*day})
	assert.Equal(t, http.StatusCreated, errCode)

	attributionData := make(map[string]*M.AttributionData)
	query := &M.AttributionQuery{
		From:                   timestamp + 3*day,
		To:                     timestamp + 10*day,
		AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
		AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
		ConversionEvent:        "event1",
		LoopbackDays:           2,
	}

	userProperty, err := M.GetQueryUserProperty(query)
	assert.Nil(t, err)
	assert.NotEqual(t, "", userProperty)
	//Should find withing lookback window
	err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(attributionData))

	_, errCode = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: user1.ID, Timestamp: timestamp + 5*day})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateOrGetSessionEvent(project.ID, user1.ID, false, false, timestamp+8*day,
		&U.PropertiesMap{}, &U.PropertiesMap{}, user1.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)
	query.From = timestamp + 5*day
	attributionData = make(map[string]*M.AttributionData)
	//event beyond lookback window
	err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(attributionData))
}

func TestAttributionWithUserIdentification(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)

	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)

	timestamp := int64(1589068800)
	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp, user1.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp, user2.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)

	attributionData := make(map[string]*M.AttributionData)
	query := &M.AttributionQuery{
		From:                   timestamp - 86400,
		To:                     timestamp + 2*86400,
		AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
		AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
		ConversionEvent:        "event1",
		LoopbackDays:           0,
	}

	userProperty, err := M.GetQueryUserProperty(query)
	assert.Nil(t, err)
	assert.NotEqual(t, "", userProperty)

	//both user should be treated different
	err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), attributionData["$none"].ConversionEventCount)

	customerUserId := U.RandomLowerAphaNumString(15)
	_, errCode = M.UpdateUser(project.ID, user1.ID, &M.User{CustomerUserId: customerUserId}, timestamp+86400)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = M.UpdateUser(project.ID, user2.ID, &M.User{CustomerUserId: customerUserId}, timestamp+86400)
	assert.Equal(t, http.StatusAccepted, errCode)

	//both user should be treated same
	attributionData = make(map[string]*M.AttributionData)
	err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, userProperty)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), attributionData["$none"].ConversionEventCount)

	t.Run("TestAttributionUserIdentificationWithLookbackDays", func(t *testing.T) {
		//continuation to previous users
		user1NewPropertiesId, status := M.UpdateUserProperties(project.ID, user1.ID, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
		assert.Equal(t, http.StatusAccepted, status)
		user2NewPropertiesId, status := M.UpdateUserProperties(project.ID, user2.ID, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
		assert.Equal(t, http.StatusAccepted, status)
		/*
			t+3day -> first time $initial_campaign set with event for user1 and user2
			t+6day -> sessioned event for user1 and user2
		*/
		status = createEventWithSession(project.ID, "event1", user1.ID, timestamp+3*86400, user1NewPropertiesId)
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user2.ID, timestamp+3*86400, user2NewPropertiesId)
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user1.ID, timestamp+6*86400, user1NewPropertiesId)
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user2.ID, timestamp+6*86400, user2NewPropertiesId)
		assert.Equal(t, http.StatusCreated, status)

		//should return 0 attribution data in 1 lookbackdays
		attributionData = make(map[string]*M.AttributionData)
		query := &M.AttributionQuery{
			From:                   timestamp + 4*86400,
			To:                     timestamp + 7*86400,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        "event1",
			LoopbackDays:           1,
		}
		err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, U.UP_INITIAL_CAMPAIGN)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(attributionData))

		//should get 1 unique user on 3 lookbackdays
		attributionData = make(map[string]*M.AttributionData)
		query.LoopbackDays = 3
		err = M.GetUniqueUserByAttributionKeyAndLookbackWindow(project.ID, attributionData, query, U.UP_INITIAL_CAMPAIGN)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), attributionData["12345"].ConversionEventCount)
	})
}
