package tests

import (
	"net/http"
	"reflect"
	"testing"

	"encoding/json"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	TaskSession "factors/task/session"
	U "factors/util"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

// Creating mock session with campaign.
func createSession(projectId uint64, userId string, timestamp int64, campaignName string) (*model.Event, int) {
	randomEventName := RandomURL()

	properties := U.PropertiesMap{}
	if campaignName != "" {
		properties["$qp_utm_campaign"] = campaignName
	}

	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp - 1, // session is created OneSec before the event.
		UserId:          userId,
		EventProperties: properties,
	}
	_, response := SDK.Track(projectId, &trackPayload, false, SDK.SourceJSSDK)
	TaskSession.AddSession([]uint64{projectId}, timestamp-60, 0, 0, 0, 0)

	event, errCode := store.GetStore().GetEventById(projectId, response.EventId)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	if event.SessionId == nil || *event.SessionId == "" {
		return nil, http.StatusInternalServerError
	}

	sessionEvent, errCode := store.GetStore().GetEventById(projectId, *event.SessionId)
	if errCode == http.StatusFound {
		return sessionEvent, http.StatusCreated
	}

	return sessionEvent, errCode
}

func createEventWithSession(projectId uint64, conversionEventName string, userId string, timestamp int64,
	userPropertiesId string, sessionCampaignName string) int {

	userSession, errCode := createSession(projectId, userId, timestamp, sessionCampaignName)
	if errCode != http.StatusCreated {
		return errCode
	}

	userEventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: projectId, Name: conversionEventName})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return errCode
	}

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: projectId, EventNameId: userEventName.ID,
		UserId: userId, Timestamp: timestamp, SessionId: &userSession.ID})
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

	// Should return error for non adwords customer account id
	result, err := store.GetStore().ExecuteAttributionQuery(project.ID, &model.AttributionQuery{})
	assert.Nil(t, result)
	assert.NotNil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	assert.Equal(t, http.StatusAccepted, errCode)
	value := []byte(`{"cost": "0","clicks": "0","campaign_id":"123456","impressions": "0", "campaign_name": "test"}`)
	document := &model.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		TypeAlias:         "campaign_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
	}

	status := store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	/*
		timestamp(t)
		t				user1 ->first session + event initial_campaign -> 123456
		t+3day			user2 ->first session + event initial_campaign -> 54321
		t+3day			user3 ->first session initial_campaign -> 54321
		t+5day			user1 ->session + event latest_campaign -> 1234567
		t+5day			user2 ->session + event latest_campaign -> 123456
	*/
	timestamp := int64(1589068800)

	// Creating 3 users
	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user1)
	assert.Equal(t, http.StatusCreated, errCode)
	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user2)
	assert.Equal(t, http.StatusCreated, errCode)
	user3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", user1.ID,
		timestamp+1*U.SECONDS_IN_A_DAY, user1.PropertiesId, "111111")
	assert.Equal(t, http.StatusCreated, errCode)

	//Update user1 and user2 properties with latest campaign
	t.Run("AttributionQueryFirstTouchWithinTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
		}

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	t.Run("AttributionQueryFirstTouchOutOfTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp + 3*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
		}

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		user2.ID, timestamp+5*U.SECONDS_IN_A_DAY, user2.PropertiesId, "222222")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		user3.ID, timestamp+5*U.SECONDS_IN_A_DAY, user3.PropertiesId, "333333")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryLastTouchCampaignNoLookbackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodLastTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
		}

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "333333"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", user1.ID, timestamp+6*U.SECONDS_IN_A_DAY, user1.PropertiesId, "1234567")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestFirstTouchCampaignWithLookbackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 10*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           20,
		}

		//Should only have user2 with no 0 linked event count
		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionUserCount(result, "1234567"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})
}

func TestAttributionEngagementModel(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

	// Should return error for non adwords customer account id
	result, err := store.GetStore().ExecuteAttributionQuery(project.ID, &model.AttributionQuery{})
	assert.Nil(t, result)
	assert.NotNil(t, err)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	assert.Equal(t, http.StatusAccepted, errCode)
	value := []byte(`{"cost": "0","clicks": "0","campaign_id":"123456","impressions": "0", "campaign_name": "test"}`)
	document := &model.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		TypeAlias:         "campaign_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
	}

	status := store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	/*
		timestamp(t)
		t				user1 ->first session + event initial_campaign -> 123456
		t+3day			user2 ->first session + event initial_campaign -> 54321
		t+3day			user3 ->first session initial_campaign -> 54321
		t+5day			user1 ->session + event latest_campaign -> 1234567
		t+5day			user2 ->session + event latest_campaign -> 123456
	*/
	timestamp := int64(1589068800)

	// Creating 3 users
	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user1)
	assert.Equal(t, http.StatusCreated, errCode)
	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user2)
	assert.Equal(t, http.StatusCreated, errCode)
	user3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", user1.ID,
		timestamp+1*U.SECONDS_IN_A_DAY, user1.PropertiesId, "111111")
	assert.Equal(t, http.StatusCreated, errCode)

	//Update user1 and user2 properties with latest campaign
	t.Run("TestAttributionEngagementQueryFirstTouchWithinTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
			QueryType:              model.AttributionQueryTypeEngagementBased,
		}

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	t.Run("TestAttributionEngagementQueryFirstTouchOutOfTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp + 3*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
			QueryType:              model.AttributionQueryTypeEngagementBased,
		}

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		user2.ID, timestamp+5*U.SECONDS_IN_A_DAY, user2.PropertiesId, "222222")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		user3.ID, timestamp+5*U.SECONDS_IN_A_DAY, user3.PropertiesId, "333333")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestAttributionEngagementQueryLastTouchCampaignNoLookbackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodLastTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
			QueryType:              model.AttributionQueryTypeEngagementBased,
		}

		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "333333"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", user1.ID, timestamp+6*U.SECONDS_IN_A_DAY, user1.PropertiesId, "1234567")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestEngagementFirstTouchCampaignWithLookbackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 10*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           20,
			QueryType:              model.AttributionQueryTypeEngagementBased,
		}

		//Should only have user2 with no 0 linked event count
		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionUserCount(result, "1234567"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})
}

func getConversionUserCount(result *model.QueryResult, key interface{}) interface{} {
	for _, row := range result.Rows {
		if row[0] == key {
			return row[5]
		}
	}
	return int64(-1)
}

func getCompareConversionUserCount(result *model.QueryResult, key interface{}) interface{} {
	for _, row := range result.Rows {
		if row[0] == key {
			return row[7]
		}
	}
	return int64(-1)
}

func TestAttributionLastTouchWithLookbackWindow(t *testing.T) {

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
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

	user1Properties := make(map[string]interface{})
	user1Properties[U.UP_LATEST_CAMPAIGN] = 123456
	user1PropertiesBytes, _ := json.Marshal(user1Properties)
	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{RawMessage: user1PropertiesBytes},
		JoinTimestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, user1)

	_, errCode = createSession(project.ID, user1.ID, timestamp+4*U.SECONDS_IN_A_DAY, "")
	assert.Equal(t, http.StatusCreated, errCode)
	userEventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "event1"})
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "event2"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: user1.ID,
		Timestamp: timestamp + 3*U.SECONDS_IN_A_DAY})
	assert.Equal(t, http.StatusCreated, errCode)

	query := &model.AttributionQuery{
		From:                   timestamp + 3*U.SECONDS_IN_A_DAY,
		To:                     timestamp + 10*U.SECONDS_IN_A_DAY,
		AttributionKey:         model.AttributionKeyCampaign,
		AttributionMethodology: model.AttributionMethodLastTouch,
		ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           2,
		QueryType:              model.AttributionQueryTypeConversionBased,
	}

	// Should find within look back window
	_, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: userEventName.ID,
		UserId: user1.ID, Timestamp: timestamp + 5*U.SECONDS_IN_A_DAY})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = createSession(project.ID, user1.ID, timestamp+8*U.SECONDS_IN_A_DAY, "")
	assert.Equal(t, http.StatusCreated, errCode)
	query.From = timestamp + 5*U.SECONDS_IN_A_DAY

	// Event beyond lookback window.
	_, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
}

func TestAttributionWithUserIdentification(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)

	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)

	timestamp := int64(1589068800)

	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp, user1.PropertiesId, "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp, user2.PropertiesId, "")
	assert.Equal(t, http.StatusCreated, errCode)

	query := &model.AttributionQuery{
		From:                   timestamp - 1*U.SECONDS_IN_A_DAY,
		To:                     timestamp + 2*U.SECONDS_IN_A_DAY,
		AttributionKey:         model.AttributionKeyCampaign,
		AttributionMethodology: model.AttributionMethodLastTouch,
		ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           0,
		QueryType:              model.AttributionQueryTypeConversionBased,
	}

	//both user should be treated different
	result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	// Lookback is 0. There should be no attribution.
	// Attribuiton Time: 1589068798, Conversion Time: 1589068800, diff = 2 secs
	assert.Equal(t, float64(0), getConversionUserCount(result, "$none"))

	customerUserId := U.RandomLowerAphaNumString(15)
	_, errCode = store.GetStore().UpdateUser(project.ID, user1.ID, &model.User{CustomerUserId: customerUserId}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = store.GetStore().UpdateUser(project.ID, user2.ID, &model.User{CustomerUserId: customerUserId}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)

	query = &model.AttributionQuery{
		From:                   timestamp - 1*U.SECONDS_IN_A_DAY,
		To:                     timestamp + 2*U.SECONDS_IN_A_DAY,
		AttributionKey:         model.AttributionKeyCampaign,
		AttributionMethodology: model.AttributionMethodLastTouch,
		ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           5,
		QueryType:              model.AttributionQueryTypeConversionBased,
	}

	result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	// both user should be treated same
	assert.Equal(t, float64(1), getConversionUserCount(result, "$none"))

	t.Run("TestAttributionUserIdentificationWithLookbackDays0", func(t *testing.T) {
		// continuation to previous users
		user1NewPropertiesId, status := store.GetStore().UpdateUserProperties(project.ID, user1.ID, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*U.SECONDS_IN_A_DAY)
		assert.Equal(t, http.StatusAccepted, status)
		user2NewPropertiesId, status := store.GetStore().UpdateUserProperties(project.ID, user2.ID, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*U.SECONDS_IN_A_DAY)
		// Status should be not_modified as user1 and user2 belong to
		// same customer_user and user_properties merged on first UpdateUserProperties.
		assert.Equal(t, http.StatusNotModified, status)
		/*
			t+3day -> first time $initial_campaign set with event for user1 and user2
			t+6day -> session event for user1 and user2
		*/
		status = createEventWithSession(project.ID, "event1", user1.ID, timestamp+3*U.SECONDS_IN_A_DAY, user1NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user2.ID, timestamp+3*U.SECONDS_IN_A_DAY, user2NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user1.ID, timestamp+6*U.SECONDS_IN_A_DAY, user1NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user2.ID, timestamp+6*U.SECONDS_IN_A_DAY, user2NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)

		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 7*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           4,
			QueryType:              model.AttributionQueryTypeEngagementBased,
		}
		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		// The attribution didn't happen in the query period. First touch is on 3rd day and which
		// is not between 4th to 7th (query period). Hence count is 0.
		assert.Equal(t, float64(0), getConversionUserCount(result, "12345"))
	})

}

func TestAttributionEngagementWithUserIdentification(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)

	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)

	timestamp := int64(1589068800)

	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp, user1.PropertiesId, "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp, user2.PropertiesId, "")
	assert.Equal(t, http.StatusCreated, errCode)

	query := &model.AttributionQuery{
		From:                   timestamp - 1*U.SECONDS_IN_A_DAY,
		To:                     timestamp + 2*U.SECONDS_IN_A_DAY,
		AttributionKey:         model.AttributionKeyCampaign,
		AttributionMethodology: model.AttributionMethodLastTouch,
		ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           0,
		QueryType:              model.AttributionQueryTypeEngagementBased,
	}

	// Both user should be treated different
	result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	// Lookback days = 0. Hence conversion couldn't happen withing lookback.
	assert.Equal(t, float64(0), getConversionUserCount(result, "$none"))

	query = &model.AttributionQuery{
		From:                   timestamp - 1*U.SECONDS_IN_A_DAY,
		To:                     timestamp + 2*U.SECONDS_IN_A_DAY,
		AttributionKey:         model.AttributionKeyCampaign,
		AttributionMethodology: model.AttributionMethodLastTouch,
		ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           2,
		QueryType:              model.AttributionQueryTypeEngagementBased,
	}

	// Both user should be treated different
	result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	// Lookback days = 2
	assert.Equal(t, float64(2), getConversionUserCount(result, "$none"))

	customerUserId1 := U.RandomLowerAphaNumString(15) + "_one"
	customerUserId2 := U.RandomLowerAphaNumString(15) + "_two"
	_, errCode = store.GetStore().UpdateUser(project.ID, user1.ID, &model.User{CustomerUserId: customerUserId1}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = store.GetStore().UpdateUser(project.ID, user2.ID, &model.User{CustomerUserId: customerUserId2}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)

	result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), getConversionUserCount(result, "$none"))

	t.Run("TestAttributionUserIdentificationWithLookbackDays", func(t *testing.T) {

		// 3 days is out of query window, but should be considered as it falls under Engagement window
		status := createEventWithSession(project.ID, "eventNewX", user1.ID, timestamp+5*U.SECONDS_IN_A_DAY, user1.PropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "eventNewX", user2.ID, timestamp+6*U.SECONDS_IN_A_DAY, user2.PropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)

		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 7*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodLastTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "eventNewX"},
			LookbackDays:           4,
			QueryType:              model.AttributionQueryTypeEngagementBased,
		}
		result, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(result, "12345"))
	})
}

func TestAttributionMethodologies(t *testing.T) {

	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp1 := "campaign1"
	camp2 := "campaign2"
	camp3 := "campaign3"

	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionTimestamp)
	userSession[user1] = make(map[string]model.UserSessionTimestamp)
	userSession[user1][camp1] = model.UserSessionTimestamp{MinTimestamp: 100, MaxTimestamp: 200}
	userSession[user1][camp2] = model.UserSessionTimestamp{MinTimestamp: 150, MaxTimestamp: 300}
	userSession[user1][camp3] = model.UserSessionTimestamp{MinTimestamp: 50, MaxTimestamp: 100}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionTimestamp
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]string
		wantLinkedEventUserCampaign map[string]map[string][]string
		wantErr                     bool
	}{

		// Test for LINEAR_TOUCH
		{"linear_touch",
			args{model.AttributionMethodLinear,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp,
				lookbackDays,
			},
			map[string][]string{user1: {camp1, camp2, camp3}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{model.AttributionMethodFirstTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{model.AttributionMethodLastTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := store.GetStore().ApplyAttribution(tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession,
				tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays, int64(queryFrom), int64(queryTo))
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.method == model.AttributionMethodLinear {
				for key, _ := range tt.wantUsersAttribution {
					if len(got[key]) != len(tt.wantUsersAttribution[key]) {
						t.Errorf("applyAttribution() Failed LINEAR TOUCH got = %v, want %v", len(got[key]), len(tt.wantUsersAttribution[key]))
					}
				}
			} else if !reflect.DeepEqual(got, tt.wantUsersAttribution) {
				t.Errorf("applyAttribution() got = %v, want %v", got, tt.wantUsersAttribution)
			}
			if !reflect.DeepEqual(got1, tt.wantLinkedEventUserCampaign) {
				t.Errorf("applyAttribution() got1 = %v, want %v", got1, tt.wantLinkedEventUserCampaign)
			}
		})
	}
}

func TestAttributionMethodologiesFirstTouchNonDirect(t *testing.T) {

	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp0 := "$none"
	camp1 := "campaign1"
	camp2 := "campaign2"
	camp3 := "campaign3"
	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionTimestamp)
	userSession[user1] = make(map[string]model.UserSessionTimestamp)
	userSession[user1][camp0] = model.UserSessionTimestamp{MinTimestamp: 10, MaxTimestamp: 40}
	userSession[user1][camp1] = model.UserSessionTimestamp{MinTimestamp: 100, MaxTimestamp: 200}
	userSession[user1][camp2] = model.UserSessionTimestamp{MinTimestamp: 150, MaxTimestamp: 300}
	userSession[user1][camp3] = model.UserSessionTimestamp{MinTimestamp: 50, MaxTimestamp: 100}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionTimestamp
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]string
		wantLinkedEventUserCampaign map[string]map[string][]string
		wantErr                     bool
	}{
		// Test for LINEAR_TOUCH
		{"linear_touch",
			args{model.AttributionMethodLinear,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp0, camp1, camp2, camp3}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{model.AttributionMethodFirstTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp0}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{model.AttributionMethodLastTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{model.AttributionMethodFirstTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{model.AttributionMethodLastTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := store.GetStore().ApplyAttribution(tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession,
				tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays,
				int64(queryFrom), int64(queryTo))
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.method == model.AttributionMethodLinear {
				for key, _ := range tt.wantUsersAttribution {
					if len(got[key]) != len(tt.wantUsersAttribution[key]) {
						t.Errorf("applyAttribution() Failed LINEAR TOUCH got = %v, want %v", len(got[key]), len(tt.wantUsersAttribution[key]))
					}
				}
			} else if !reflect.DeepEqual(got, tt.wantUsersAttribution) {
				t.Errorf("applyAttribution() got = %v, want %v", got, tt.wantUsersAttribution)
			}
			if !reflect.DeepEqual(got1, tt.wantLinkedEventUserCampaign) {
				t.Errorf("applyAttribution() got1 = %v, want %v", got1, tt.wantLinkedEventUserCampaign)
			}
		})
	}
}

func TestAttributionMethodologiesLastTouchNonDirect(t *testing.T) {

	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp1 := "campaign1"
	camp2 := "campaign2"
	camp3 := "campaign3"
	camp4 := "$none"
	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionTimestamp)
	userSession[user1] = make(map[string]model.UserSessionTimestamp)
	userSession[user1][camp1] = model.UserSessionTimestamp{MinTimestamp: 100, MaxTimestamp: 200}
	userSession[user1][camp2] = model.UserSessionTimestamp{MinTimestamp: 150, MaxTimestamp: 300}
	userSession[user1][camp3] = model.UserSessionTimestamp{MinTimestamp: 50, MaxTimestamp: 100}
	userSession[user1][camp4] = model.UserSessionTimestamp{MinTimestamp: 10, MaxTimestamp: 400}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionTimestamp
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]string
		wantLinkedEventUserCampaign map[string]map[string][]string
		wantErr                     bool
	}{
		// Test for LINEAR_TOUCH
		{"linear_touch",
			args{model.AttributionMethodLinear,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp1, camp2, camp3, camp4}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{model.AttributionMethodFirstTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp4}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{model.AttributionMethodLastTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{model.AttributionMethodFirstTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{model.AttributionMethodLastTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := store.GetStore().ApplyAttribution(tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession,
				tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays, int64(queryFrom), int64(queryTo))
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.method == model.AttributionMethodLinear {
				for key := range tt.wantUsersAttribution {
					if len(got[key]) != len(tt.wantUsersAttribution[key]) {
						t.Errorf("applyAttribution() Failed LINEAR TOUCH got = %v, want %v", len(got[key]), len(tt.wantUsersAttribution[key]))
					}
				}
			} else if !reflect.DeepEqual(got, tt.wantUsersAttribution) {
				t.Errorf("applyAttribution() got = %v, want %v", got, tt.wantUsersAttribution)
			}
			if !reflect.DeepEqual(got1, tt.wantLinkedEventUserCampaign) {
				t.Errorf("applyAttribution() got1 = %v, want %v", got1, tt.wantLinkedEventUserCampaign)
			}
		})
	}
}
