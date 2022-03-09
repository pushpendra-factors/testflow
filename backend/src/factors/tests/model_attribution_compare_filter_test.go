package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestAttributionModelCompare(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

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
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "111111", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                          timestamp - 2*U.SECONDS_IN_A_DAY,
			To:                            timestamp + 3*U.SECONDS_IN_A_DAY,
			LookbackDays:                  10,
			AttributionKey:                model.AttributionKeyCampaign,
			AttributionKeyFilter:          []model.AttributionKeyFilter{},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
			QueryType:                     model.AttributionQueryTypeEngagementBased,
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
	})

	t.Run("AttributionQueryCompareOutOfTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                          timestamp + 3*U.SECONDS_IN_A_DAY,
			To:                            timestamp + 3*U.SECONDS_IN_A_DAY,
			LookbackDays:                  10,
			AttributionKey:                model.AttributionKeyCampaign,
			AttributionKeyFilter:          []model.AttributionKeyFilter{},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+5*U.SECONDS_IN_A_DAY, "222222", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+5*U.SECONDS_IN_A_DAY, "333333", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareNoLookBackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                          timestamp,
			To:                            timestamp + 4*U.SECONDS_IN_A_DAY,
			LookbackDays:                  10,
			AttributionKey:                model.AttributionKeyCampaign,
			AttributionKeyFilter:          []model.AttributionKeyFilter{},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
			QueryType:                     model.AttributionQueryTypeEngagementBased,
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "333333"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "333333"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", createdUserID1, timestamp+6*U.SECONDS_IN_A_DAY, "1234567", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestFirstTouchCampaignCompareWithLookBackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                          timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                            timestamp + 10*U.SECONDS_IN_A_DAY,
			LookbackDays:                  20,
			AttributionKey:                model.AttributionKeyCampaign,
			AttributionKeyFilter:          []model.AttributionKeyFilter{},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		//Should only have user2 with no 0 linked event count
		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "333333"))
		// compare conversion should be same as conversion count
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionUserCount(query.AttributionKey, result, "1234567"))
	})
}

func TestAttributionCompareWithLookBackWindowX(t *testing.T) {

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
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{RawMessage: user1PropertiesBytes},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	_, errCode = createSession(project.ID, createdUserID1, timestamp+4*U.SECONDS_IN_A_DAY, "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)
	userEventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "event1"})
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "event2"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: createdUserID1,
		Timestamp: timestamp + 3*U.SECONDS_IN_A_DAY})
	assert.Equal(t, http.StatusCreated, errCode)

	query := &model.AttributionQuery{
		From:                          timestamp + 3*U.SECONDS_IN_A_DAY,
		To:                            timestamp + 10*U.SECONDS_IN_A_DAY,
		LookbackDays:                  2,
		AttributionKey:                model.AttributionKeyCampaign,
		AttributionKeyFilter:          []model.AttributionKeyFilter{},
		LinkedEvents:                  []model.QueryEventWithProperties{},
		AttributionMethodology:        model.AttributionMethodFirstTouch,
		AttributionMethodologyCompare: model.AttributionMethodLastTouch,
		ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
		ConversionEventCompare:        model.QueryEventWithProperties{},
	}

	// Should find within lookBack window
	_, err = store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: userEventName.ID,
		UserId: createdUserID1, Timestamp: timestamp + 5*U.SECONDS_IN_A_DAY})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = createSession(project.ID, createdUserID1, timestamp+8*U.SECONDS_IN_A_DAY, "campaign1", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)
	query.From = timestamp + 5*U.SECONDS_IN_A_DAY

	//event beyond look back window
	result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	assert.Equal(t, float64(0), getConversionUserCount(query.AttributionKey, result, "campaign1"))
	assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "$none"))
}

func TestAttributionModelCompareFilter(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

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
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "111111", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp - 2*U.SECONDS_IN_A_DAY,
			To:             timestamp + 3*U.SECONDS_IN_A_DAY,
			LookbackDays:   10,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Operator: model.EqualsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
	})

	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp - 2*U.SECONDS_IN_A_DAY,
			To:             timestamp + 3*U.SECONDS_IN_A_DAY,
			LookbackDays:   10,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Operator: model.NotEqualOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
	})
	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp - 2*U.SECONDS_IN_A_DAY,
			To:             timestamp + 3*U.SECONDS_IN_A_DAY,
			LookbackDays:   10,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Operator: model.ContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
	})
	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp - 2*U.SECONDS_IN_A_DAY,
			To:             timestamp + 3*U.SECONDS_IN_A_DAY,
			LookbackDays:   10,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Operator: model.NotContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+5*U.SECONDS_IN_A_DAY, "222222", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+5*U.SECONDS_IN_A_DAY, "333333", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareNoLookBackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp,
			To:             timestamp + 4*U.SECONDS_IN_A_DAY,
			LookbackDays:   10,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Operator: model.ContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "333333"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "333333"))
	})

	t.Run("AttributionQueryCompareNoLookBackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp,
			To:             timestamp + 4*U.SECONDS_IN_A_DAY,
			LookbackDays:   10,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Operator: model.NotContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "333333"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "333333"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", createdUserID1, timestamp+6*U.SECONDS_IN_A_DAY, "1234567", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestFirstTouchCampaignCompareWithLookBackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp + 4*U.SECONDS_IN_A_DAY,
			To:             timestamp + 10*U.SECONDS_IN_A_DAY,
			LookbackDays:   20,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.ContainsOpStr,
					Value:     "111111",
					LogicalOp: "OR"},
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.ContainsOpStr,
					Value:     "222222",
					LogicalOp: "OR"},
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.ContainsOpStr,
					Value:     "333333",
					LogicalOp: "OR"},
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.ContainsOpStr,
					Value:     "1234567",
					LogicalOp: "OR"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		//Should only have user2 with no 0 linked event count
		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "333333"))
		// compare conversion should be same as conversion count
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(query.AttributionKey, result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionUserCount(query.AttributionKey, result, "1234567"))
	})

	t.Run("TestFirstTouchCampaignCompareWithLookBackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:           timestamp + 4*U.SECONDS_IN_A_DAY,
			To:             timestamp + 10*U.SECONDS_IN_A_DAY,
			LookbackDays:   20,
			AttributionKey: model.AttributionKeyCampaign,
			AttributionKeyFilter: []model.AttributionKeyFilter{
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.NotContainsOpStr,
					Value:     "111111",
					LogicalOp: "AND"},
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.NotContainsOpStr,
					Value:     "222222",
					LogicalOp: "AND"},
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.NotContainsOpStr,
					Value:     "333333",
					LogicalOp: "AND"},
				{AttributionKey: model.AttributionKeyCampaign,
					Property:  "name",
					Operator:  model.NotContainsOpStr,
					Value:     "1234567",
					LogicalOp: "AND"},
			},
			LinkedEvents:                  []model.QueryEventWithProperties{},
			AttributionMethodology:        model.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: model.AttributionMethodLastTouch,
			ConversionEvent:               model.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        model.QueryEventWithProperties{},
		}

		//Should only have user2 with no 0 linked event count
		result, err := store.GetStore().ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "333333"))
		// compare conversion should be same as conversion count
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(query.AttributionKey, result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "1234567"))
	})
}
