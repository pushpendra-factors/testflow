package tests

import (
	"encoding/json"
	H "factors/handler"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"

	M "factors/model"
)

func TestAttributionModelCompare(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

	// Should return error for non adwords customer account id
	result, err := M.ExecuteAttributionQuery(project.ID, &M.AttributionQuery{})
	assert.Nil(t, result)
	assert.NotNil(t, err)

	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	assert.Equal(t, http.StatusAccepted, errCode)
	value := []byte(`{"cost": "0","clicks": "0","campaign_id":"123456","impressions": "0", "campaign_name": "test"}`)
	document := &M.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		TypeAlias:         "campaign_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
	}

	status := M.CreateAdwordsDocument(document)
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
	day := int64(86400)

	// Creating 3 users
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user1)
	assert.Equal(t, http.StatusCreated, errCode)
	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user2)
	assert.Equal(t, http.StatusCreated, errCode)
	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", user1.ID,
		timestamp+1*day, user1.PropertiesId, "111111")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                          timestamp - 2*day,
			To:                            timestamp + 3*day,
			LookbackDays:                  10,
			AttributionKey:                M.AttributionKeyCampaign,
			AttributionKeyFilter:          []M.AttributionKeyFilter{},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
			QueryType:                     M.AttributionQueryTypeEngagementBased,
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	t.Run("AttributionQueryCompareOutOfTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                          timestamp + 3*day,
			To:                            timestamp + 3*day,
			LookbackDays:                  10,
			AttributionKey:                M.AttributionKeyCampaign,
			AttributionKeyFilter:          []M.AttributionKeyFilter{},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		user2.ID, timestamp+5*day, user2.PropertiesId, "222222")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		user3.ID, timestamp+5*day, user3.PropertiesId, "333333")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareNoLookBackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                          timestamp,
			To:                            timestamp + 4*day,
			LookbackDays:                  10,
			AttributionKey:                M.AttributionKeyCampaign,
			AttributionKeyFilter:          []M.AttributionKeyFilter{},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
			QueryType:                     M.AttributionQueryTypeEngagementBased,
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "333333"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getCompareConversionUserCount(result, "222222"))
		assert.Equal(t, float64(0), getCompareConversionUserCount(result, "333333"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", user1.ID, timestamp+6*day, user1.PropertiesId, "1234567")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestFirstTouchCampaignCompareWithLookBackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                          timestamp + 4*day,
			To:                            timestamp + 10*day,
			LookbackDays:                  20,
			AttributionKey:                M.AttributionKeyCampaign,
			AttributionKeyFilter:          []M.AttributionKeyFilter{},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		//Should only have user2 with no 0 linked event count
		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "333333"))
		// compare conversion should be same as conversion count
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionUserCount(result, "1234567"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})
}

func TestAttributionCompareWithLookBackWindowX(t *testing.T) {

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
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{RawMessage: user1PropertiesBytes},
		JoinTimestamp: timestamp})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, user1)

	_, errCode = createSession(project.ID, user1.ID, timestamp+4*day, "")
	assert.Equal(t, http.StatusCreated, errCode)
	userEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "event1"})
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "event2"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: user1.ID,
		Timestamp: timestamp + 3*day})
	assert.Equal(t, http.StatusCreated, errCode)

	query := &M.AttributionQuery{
		From:                          timestamp + 3*day,
		To:                            timestamp + 10*day,
		LookbackDays:                  2,
		AttributionKey:                M.AttributionKeyCampaign,
		AttributionKeyFilter:          []M.AttributionKeyFilter{},
		LinkedEvents:                  []M.QueryEventWithProperties{},
		AttributionMethodology:        M.AttributionMethodFirstTouch,
		AttributionMethodologyCompare: M.AttributionMethodLastTouch,
		ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
		ConversionEventCompare:        M.QueryEventWithProperties{},
	}

	// Should find within lookBack window
	_, err = M.ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)

	_, errCode = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: userEventName.ID,
		UserId: user1.ID, Timestamp: timestamp + 5*day})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = createSession(project.ID, user1.ID, timestamp+8*day, "campaign1")
	assert.Equal(t, http.StatusCreated, errCode)
	query.From = timestamp + 5*day

	//event beyond look back window
	result, err := M.ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	assert.Equal(t, float64(0), getConversionUserCount(result, "campaign1"))
	assert.Equal(t, float64(0), getConversionUserCount(result, "none"))

}

func TestAttributionModelCompareFilter(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

	// Should return error for non adwords customer account id
	result, err := M.ExecuteAttributionQuery(project.ID, &M.AttributionQuery{})
	assert.Nil(t, result)
	assert.NotNil(t, err)

	_, errCode := M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	assert.Equal(t, http.StatusAccepted, errCode)
	value := []byte(`{"cost": "0","clicks": "0","campaign_id":"123456","impressions": "0", "campaign_name": "test"}`)
	document := &M.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		TypeAlias:         "campaign_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{value},
	}

	status := M.CreateAdwordsDocument(document)
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
	day := int64(86400)

	// Creating 3 users
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user1)
	assert.Equal(t, http.StatusCreated, errCode)
	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user2)
	assert.Equal(t, http.StatusCreated, errCode)
	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp})
	assert.NotNil(t, user3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", user1.ID,
		timestamp+1*day, user1.PropertiesId, "111111")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp - 2*day,
			To:             timestamp + 3*day,
			LookbackDays:   10,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Operator: M.EqualsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp - 2*day,
			To:             timestamp + 3*day,
			LookbackDays:   10,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Operator: M.NotEqualOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})
	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp - 2*day,
			To:             timestamp + 3*day,
			LookbackDays:   10,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Operator: M.ContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})
	t.Run("AttributionQueryCompareTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp - 2*day,
			To:             timestamp + 3*day,
			LookbackDays:   10,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Operator: M.NotContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		user2.ID, timestamp+5*day, user2.PropertiesId, "222222")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		user3.ID, timestamp+5*day, user3.PropertiesId, "333333")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryCompareNoLookBackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp,
			To:             timestamp + 4*day,
			LookbackDays:   10,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Operator: M.ContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "333333"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "333333"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	t.Run("AttributionQueryCompareNoLookBackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp,
			To:             timestamp + 4*day,
			LookbackDays:   10,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Operator: M.NotContainsOpStr,
					Value:    "111111"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "333333"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "333333"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", user1.ID, timestamp+6*day, user1.PropertiesId, "1234567")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestFirstTouchCampaignCompareWithLookBackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp + 4*day,
			To:             timestamp + 10*day,
			LookbackDays:   20,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.ContainsOpStr,
					Value:     "111111",
					LogicalOp: "OR"},
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.ContainsOpStr,
					Value:     "222222",
					LogicalOp: "OR"},
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.ContainsOpStr,
					Value:     "333333",
					LogicalOp: "OR"},
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.ContainsOpStr,
					Value:     "1234567",
					LogicalOp: "OR"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		//Should only have user2 with no 0 linked event count
		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(result, "333333"))
		// compare conversion should be same as conversion count
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "222222"))
		assert.Equal(t, float64(1), getCompareConversionUserCount(result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionUserCount(result, "1234567"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})

	t.Run("TestFirstTouchCampaignCompareWithLookBackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:           timestamp + 4*day,
			To:             timestamp + 10*day,
			LookbackDays:   20,
			AttributionKey: M.AttributionKeyCampaign,
			AttributionKeyFilter: []M.AttributionKeyFilter{
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.NotContainsOpStr,
					Value:     "111111",
					LogicalOp: "AND"},
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.NotContainsOpStr,
					Value:     "222222",
					LogicalOp: "AND"},
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.NotContainsOpStr,
					Value:     "333333",
					LogicalOp: "AND"},
				{AttributionKey: M.AttributionKeyCampaign,
					Property:  "name",
					Operator:  M.NotContainsOpStr,
					Value:     "1234567",
					LogicalOp: "AND"},
			},
			LinkedEvents:                  []M.QueryEventWithProperties{},
			AttributionMethodology:        M.AttributionMethodFirstTouch,
			AttributionMethodologyCompare: M.AttributionMethodLastTouch,
			ConversionEvent:               M.QueryEventWithProperties{Name: "event1"},
			ConversionEventCompare:        M.QueryEventWithProperties{},
		}

		//Should only have user2 with no 0 linked event count
		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(result, "333333"))
		// compare conversion should be same as conversion count
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "111111"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "222222"))
		assert.Equal(t, int64(-1), getCompareConversionUserCount(result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, int64(-1), getConversionUserCount(result, "1234567"))
		assert.Equal(t, float64(0), getConversionUserCount(result, "none"))
	})
}
