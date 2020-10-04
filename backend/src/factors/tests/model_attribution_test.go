package tests

import (
	"net/http"
	"reflect"
	"testing"

	"encoding/json"
	H "factors/handler"
	M "factors/model"
	U "factors/util"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func createEventWithSession(projectId uint64, conversionEventName string, userId string, timestamp int64, userPropertiesId string, sessionCampaignName string) int {

	userSession, errCode := M.CreateOrGetSessionEvent(projectId, userId, false, false, timestamp,
		&U.PropertiesMap{"$campaign": sessionCampaignName}, &U.PropertiesMap{}, userPropertiesId)

	if errCode != http.StatusCreated {
		return errCode
	}

	userEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: projectId, Name: conversionEventName})
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

	// Creating 3 users
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{}, JoinTimestamp: timestamp})
	assert.NotNil(t, user1)
	assert.Equal(t, http.StatusCreated, errCode)
	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{}, JoinTimestamp: timestamp})
	assert.NotNil(t, user2)
	assert.Equal(t, http.StatusCreated, errCode)
	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{}, JoinTimestamp: timestamp})
	assert.NotNil(t, user3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp+1*day, user1.PropertiesId, "111111")
	assert.Equal(t, http.StatusCreated, errCode)

	//Update user1 and user2 properties with latest campaign
	t.Run("AttributionQueryFirstTouchWithinTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 3*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        M.QueryEventWithProperties{"event1", nil},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionEventCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionEventCount(result, "none"))
	})

	t.Run("AttributionQueryFirstTouchOutOfTimestampRangeNoLookBack", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                   timestamp + 3*day,
			To:                     timestamp + 3*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        M.QueryEventWithProperties{"event1", nil},
		}

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionEventCount(result, "111111"))
		assert.Equal(t, float64(0), getConversionEventCount(result, "none"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp+5*day, user2.PropertiesId, "222222")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user3.ID, timestamp+5*day, user3.PropertiesId, "333333")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryLastTouchCampaignNoLookbackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 4*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
			ConversionEvent:        M.QueryEventWithProperties{Name: "event1"},
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)

		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionEventCount(result, "111111"))
		assert.Equal(t, int64(-1), getConversionEventCount(result, "222222"))
		assert.Equal(t, int64(-1), getConversionEventCount(result, "333333"))
		assert.Equal(t, float64(0), getConversionEventCount(result, "none"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", user1.ID, timestamp+6*day, user1.PropertiesId, "1234567")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestLastTouchLinkedEventNoLookbackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 10*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
			ConversionEvent:        M.QueryEventWithProperties{Name: "event1"},
			LinkedEvents:           []M.QueryEventWithProperties{{"event2", nil}},
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)

	})

	t.Run("TestFirstTouchCampaignWithLookbackDays", func(t *testing.T) {
		query := &M.AttributionQuery{
			From:                   timestamp + 4*day,
			To:                     timestamp + 10*day,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        M.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           2,
		}

		userProperty, err := M.GetQueryUserProperty(query)
		assert.Nil(t, err)
		assert.NotEqual(t, "", userProperty)

		//Should only have user2 with no 0 linked event count
		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionEventCount(result, "111111"))
		assert.Equal(t, float64(1), getConversionEventCount(result, "222222"))
		assert.Equal(t, float64(1), getConversionEventCount(result, "333333"))
		// no hit for campaigns 1234567 or none
		assert.Equal(t, float64(0), getConversionEventCount(result, "1234567"))
		assert.Equal(t, float64(0), getConversionEventCount(result, "none"))
	})
}

func getConversionEventCount(result *M.QueryResult, key interface{}) interface{} {
	for _, row := range result.Rows {
		if row[0] == key {
			return row[5]
		}
	}
	return int64(-1)
}

func getLinkedEventCount(result *M.QueryResult, key interface{}) interface{} {
	for _, row := range result.Rows {
		if row[0] == key {
			linkedData := row[6].([]interface{})
			return linkedData[0]
		}
	}
	return int64(-1)
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
	_, errCode = M.CreateOrGetUserCreatedEventName(&M.EventName{ProjectId: project.ID, Name: "event2"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: user1.ID, Timestamp: timestamp + 3*day})
	assert.Equal(t, http.StatusCreated, errCode)

	query := &M.AttributionQuery{
		From:                   timestamp + 3*day,
		To:                     timestamp + 10*day,
		AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
		AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
		ConversionEvent:        M.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           2,
	}

	userProperty, err := M.GetQueryUserProperty(query)
	assert.Nil(t, err)
	assert.NotEqual(t, "", userProperty)

	//Should find withing lookback window
	_, err = M.ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)

	_, errCode = M.CreateEvent(&M.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: user1.ID, Timestamp: timestamp + 5*day})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = M.CreateOrGetSessionEvent(project.ID, user1.ID, false, false, timestamp+8*day,
		&U.PropertiesMap{}, &U.PropertiesMap{}, user1.PropertiesId)
	assert.Equal(t, http.StatusCreated, errCode)
	query.From = timestamp + 5*day

	//event beyond lookback window
	_, err = M.ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
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
	errCode = createEventWithSession(project.ID, "event1", user1.ID, timestamp, user1.PropertiesId, "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", user2.ID, timestamp, user2.PropertiesId, "")
	assert.Equal(t, http.StatusCreated, errCode)

	query := &M.AttributionQuery{
		From:                   timestamp - 86400,
		To:                     timestamp + 2*86400,
		AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
		AttributionMethodology: M.ATTRIBUTION_METHOD_LAST_TOUCH,
		ConversionEvent:        M.QueryEventWithProperties{Name: "event1"},
		LookbackDays:           0,
	}

	userProperty, err := M.GetQueryUserProperty(query)
	assert.Nil(t, err)
	assert.NotEqual(t, "", userProperty)

	//both user should be treated different
	result, err := M.ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	assert.Equal(t, float64(2), getConversionEventCount(result, "$none"))

	customerUserId := U.RandomLowerAphaNumString(15)
	_, errCode = M.UpdateUser(project.ID, user1.ID, &M.User{CustomerUserId: customerUserId}, timestamp+86400)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = M.UpdateUser(project.ID, user2.ID, &M.User{CustomerUserId: customerUserId}, timestamp+86400)
	assert.Equal(t, http.StatusAccepted, errCode)

	//both user should be treated same
	result, err = M.ExecuteAttributionQuery(project.ID, query)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), getConversionEventCount(result, "$none"))

	t.Run("TestAttributionUserIdentificationWithLookbackDays", func(t *testing.T) {
		//continuation to previous users
		user1NewPropertiesId, status := M.UpdateUserProperties(project.ID, user1.ID, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
		assert.Equal(t, http.StatusAccepted, status)		
		user2NewPropertiesId, status := M.UpdateUserProperties(project.ID, user2.ID, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
		// Status should be not_modified as user1 and user2 belong to 
		// same customer_user and user_properites merged on first UpdateUserProperties.
		assert.Equal(t, http.StatusNotModified, status)
		/*
			t+3day -> first time $initial_campaign set with event for user1 and user2
			t+6day -> sessioned event for user1 and user2
		*/
		status = createEventWithSession(project.ID, "event1", user1.ID, timestamp+3*86400, user1NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user2.ID, timestamp+3*86400, user2NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user1.ID, timestamp+6*86400, user1NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", user2.ID, timestamp+6*86400, user2NewPropertiesId, "12345")
		assert.Equal(t, http.StatusCreated, status)

		//should return 0 attribution data in 1 lookbackdays
		query := &M.AttributionQuery{
			From:                   timestamp + 4*86400,
			To:                     timestamp + 7*86400,
			AttributionKey:         M.ATTRIBUTION_KEY_CAMPAIGN,
			AttributionMethodology: M.ATTRIBUTION_METHOD_FIRST_TOUCH,
			ConversionEvent:        M.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           1,
		}
		//should get 1 unique user on 3 lookbackdays
		query.LookbackDays = 3
		result, err = M.ExecuteAttributionQuery(project.ID, query)
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionEventCount(result, "12345"))
	})
}

func TestAttributionMethodologies(t *testing.T) {

	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp1 := "campaign1"
	camp2 := "campaign2"
	camp3 := "campaign3"

	userSession := make(map[string]map[string]M.RangeTimestamp)
	userSession[user1] = make(map[string]M.RangeTimestamp)
	userSession[user1][camp1] = M.RangeTimestamp{MinTimestamp: 100, MaxTimestamp: 200}
	userSession[user1][camp2] = M.RangeTimestamp{MinTimestamp: 150, MaxTimestamp: 300}
	userSession[user1][camp3] = M.RangeTimestamp{MinTimestamp: 50, MaxTimestamp: 100}

	type args struct {
		method              string
		conversionEvent     string
		usersToBeAttributed []M.UserEventInfo
		userInitialSession  map[string]map[string]M.RangeTimestamp
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
			args{M.ATTRIBUTION_METHOD_LINEAR,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp1, camp2, camp3}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{M.ATTRIBUTION_METHOD_FIRST_TOUCH,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{M.ATTRIBUTION_METHOD_LAST_TOUCH,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp2}},
			map[string]map[string][]string{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := M.ApplyAttribution(tt.args.method, tt.args.conversionEvent, tt.args.usersToBeAttributed, tt.args.userInitialSession)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.method == M.ATTRIBUTION_METHOD_LINEAR {
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

	userSession := make(map[string]map[string]M.RangeTimestamp)
	userSession[user1] = make(map[string]M.RangeTimestamp)
	userSession[user1][camp0] = M.RangeTimestamp{MinTimestamp: 10, MaxTimestamp: 40}
	userSession[user1][camp1] = M.RangeTimestamp{MinTimestamp: 100, MaxTimestamp: 200}
	userSession[user1][camp2] = M.RangeTimestamp{MinTimestamp: 150, MaxTimestamp: 300}
	userSession[user1][camp3] = M.RangeTimestamp{MinTimestamp: 50, MaxTimestamp: 100}

	type args struct {
		method              string
		conversionEvent     string
		usersToBeAttributed []M.UserEventInfo
		userInitialSession  map[string]map[string]M.RangeTimestamp
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
			args{M.ATTRIBUTION_METHOD_LINEAR,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp0, camp1, camp2, camp3}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{M.ATTRIBUTION_METHOD_FIRST_TOUCH,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp0}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{M.ATTRIBUTION_METHOD_LAST_TOUCH,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp2}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{M.ATTRIBUTION_METHOD_FIRST_TOUCH_NON_DIRECT,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{M.ATTRIBUTION_METHOD_LAST_TOUCH_NON_DIRECT,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp2}},
			map[string]map[string][]string{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := M.ApplyAttribution(tt.args.method, tt.args.conversionEvent, tt.args.usersToBeAttributed, tt.args.userInitialSession)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.method == M.ATTRIBUTION_METHOD_LINEAR {
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

	userSession := make(map[string]map[string]M.RangeTimestamp)
	userSession[user1] = make(map[string]M.RangeTimestamp)
	userSession[user1][camp1] = M.RangeTimestamp{MinTimestamp: 100, MaxTimestamp: 200}
	userSession[user1][camp2] = M.RangeTimestamp{MinTimestamp: 150, MaxTimestamp: 300}
	userSession[user1][camp3] = M.RangeTimestamp{MinTimestamp: 50, MaxTimestamp: 100}
	userSession[user1][camp4] = M.RangeTimestamp{MinTimestamp: 10, MaxTimestamp: 400}

	type args struct {
		method              string
		conversionEvent     string
		usersToBeAttributed []M.UserEventInfo
		userInitialSession  map[string]map[string]M.RangeTimestamp
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
			args{M.ATTRIBUTION_METHOD_LINEAR,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp1, camp2, camp3, camp4}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{M.ATTRIBUTION_METHOD_FIRST_TOUCH,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp4}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{M.ATTRIBUTION_METHOD_LAST_TOUCH,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp4}},
			map[string]map[string][]string{},
			false},

		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{M.ATTRIBUTION_METHOD_FIRST_TOUCH_NON_DIRECT,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp3}},
			map[string]map[string][]string{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{M.ATTRIBUTION_METHOD_LAST_TOUCH_NON_DIRECT,
				conversionEvent,
				[]M.UserEventInfo{{user1, conversionEvent}},
				userSession,
			},
			map[string][]string{user1: {camp2}},
			map[string]map[string][]string{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := M.ApplyAttribution(tt.args.method, tt.args.conversionEvent, tt.args.usersToBeAttributed, tt.args.userInitialSession)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.method == M.ATTRIBUTION_METHOD_LINEAR {
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
