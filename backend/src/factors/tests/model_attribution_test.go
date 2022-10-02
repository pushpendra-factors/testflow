package tests

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	C "factors/config"
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
func createSession(projectId int64, userId string, timestamp int64, campaignName string, adgroupName string,
	keyword string, gclid string, source string, landingPage string) (*model.Event, int) {
	randomEventName := RandomURL()

	properties := U.PropertiesMap{}

	if campaignName != "" {
		properties["$qp_utm_campaign"] = campaignName
	}
	if adgroupName != "" {
		properties["$qp_utm_adgroup"] = adgroupName
	}
	if keyword != "" {
		properties["$qp_utm_keyword"] = keyword
	}
	if gclid != "" {
		properties["$qp_gclid"] = gclid
	}
	if source != "" {
		properties["$qp_utm_source"] = source
	}
	if landingPage != "" {
		properties["$page_url"] = landingPage
	}

	trackPayload := SDK.TrackPayload{
		Auto:            true,
		Name:            randomEventName,
		Timestamp:       timestamp - 1, // session is created OneSec before the event.
		UserId:          userId,
		EventProperties: properties,
		RequestSource:   model.UserSourceWeb,
	}
	_, response := SDK.Track(projectId, &trackPayload, false, SDK.SourceJSSDK, "")
	TaskSession.AddSession([]int64{projectId}, timestamp-60, 0, 0, 0, 0, 1)

	event, errCode := store.GetStore().GetEventById(projectId, response.EventId, "")
	if errCode != http.StatusFound {
		return nil, errCode
	}

	if event.SessionId == nil || *event.SessionId == "" {
		return nil, http.StatusInternalServerError
	}

	sessionEvent, errCode := store.GetStore().GetEventById(projectId, *event.SessionId, event.UserId)
	if errCode == http.StatusFound {
		return sessionEvent, http.StatusCreated
	}

	return sessionEvent, errCode
}

func createEventWithSession(projectId int64, conversionEventName string, userId string, timestamp int64,
	sessionCampaignName string, adgroupName string, keyword string, gclid string, source string, landingPage string) int {

	userSession, errCode := createSession(projectId, userId, timestamp, sessionCampaignName, adgroupName, keyword, gclid, source, landingPage)
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
	assert.NotNil(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "111111", "", "", "", "", "")
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
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
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

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
	})

	// Events with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+5*U.SECONDS_IN_A_DAY, "222222", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+5*U.SECONDS_IN_A_DAY, "333333", "", "", "", "", "")
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

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "333333"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", createdUserID1,
		timestamp+6*U.SECONDS_IN_A_DAY, "1234567", "", "", "", "", "")
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
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "333333"))
		// no hit for campaigns 1234567 or none
		// Assert below won't work with ExecuteAttributionQueryV1, because for camp '123456' event is 'event2'
		// While attributing, we pull users for 'event1' and not by default all sessions. Hence no longer valid.
		// assert.Equal(t, float64(0), getConversionUserCount(query.AttributionKey, result, "1234567"))
	})
}

func TestAttributionLandingPage(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)

	//Adding content groups in project

	request := model.ContentGroup{}
	request.ContentGroupName = fmt.Sprintf("%v", "cg_123")
	request.ContentGroupDescription = "description"
	value := model.ContentGroupValue{}
	value.Operator = "startsWith"
	value.LogicalOp = "OR"
	value.Value = "123"
	filters := make([]model.ContentGroupValue, 0)
	filters = append(filters, value)
	contentGroupValueArray := make([]model.ContentGroupRule, 0)
	contentGroupValue := model.ContentGroupRule{
		ContentGroupValue: "value_123",
		Rule:              filters,
	}
	contentGroupValueArray = append(contentGroupValueArray, contentGroupValue)
	contentGroupValueJson, err := json.Marshal(contentGroupValueArray)
	request.Rule = &postgres.Jsonb{contentGroupValueJson}
	w := sendCreateContentGroupRequest(r, request, agent, project.ID)
	assert.Equal(t, http.StatusCreated, w.Code)

	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})

	assert.Equal(t, http.StatusAccepted, errCode)
	valueAdwords := []byte(`{"cost": "0","clicks": "0","campaign_id":"123456","impressions": "0", "campaign_name": "test"}`)
	document := &model.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: customerAccountId,
		TypeAlias:         "campaign_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: valueAdwords},
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
	assert.NotNil(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "111111", "", "", "", "", "lp_111111")
	assert.Equal(t, http.StatusCreated, errCode)

	//Update user1 and user2 properties with the latest campaign
	t.Run("AttributionQueryFirstTouchWithinTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyLandingPage,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
			TacticOfferType:        model.MarketingEventTypeOffer,
		}

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_111111"))
	})

	t.Run("AttributionQueryFirstTouchOutOfTimestampRangeNoLookBack", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp + 3*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyLandingPage,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
			TacticOfferType:        model.MarketingEventTypeOffer,
		}

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
	})

	// Event with +3 Days
	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+3*U.SECONDS_IN_A_DAY, "222222", "", "", "", "", "lp_222222")
	assert.Equal(t, http.StatusCreated, errCode)

	// Event with +5 Days
	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+5*U.SECONDS_IN_A_DAY, "333333", "", "", "", "", "lp_333333")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryLastTouchLandingPageNoLookbackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyLandingPage,
			AttributionMethodology: model.AttributionMethodLastTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           10,
			TacticOfferType:        model.MarketingEventTypeOffer,
		}

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_111111"))
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_222222"))
		assert.Equal(t, int64(-1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_333333"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", createdUserID1,
		timestamp+6*U.SECONDS_IN_A_DAY, "1234567", "", "", "", "", "lp_1234567")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("TestFirstTouchLandingPageWithLookbackDays", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 10*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyLandingPage,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           20,
			TacticOfferType:        model.MarketingEventTypeOffer,
		}

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_111111"))
		assert.Equal(t, int64(-1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_222222"))
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_333333"))
		// no hit for landing page  lp_1234567 or none
		assert.Equal(t, float64(0), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_1234567"))
	})

	createdUserID4, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +12 Days
	errCode = createEventWithSession(project.ID, "event2", createdUserID4,
		timestamp+12*U.SECONDS_IN_A_DAY, "c111111", "", "", "", "", "123_lp_111111")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryLandingPageContentGroup", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                     timestamp + 15,
			To:                       timestamp + 20*U.SECONDS_IN_A_DAY,
			AttributionKey:           model.AttributionKeyLandingPage,
			AttributionMethodology:   model.AttributionMethodFirstTouch,
			ConversionEvent:          model.QueryEventWithProperties{Name: "event2"},
			LookbackDays:             10,
			AttributionContentGroups: []string{"cg_123"},
			TacticOfferType:          model.MarketingEventTypeOffer,
		}

		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		//testing content group level report. Content group "cg_123" has value "value_123" if landing page url starts with "123"(As per rule defined for cg_123)
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "value_123"))
	})

	t.Run("AttributionQueryInfluenceLandingMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyLandingPage,
			AttributionMethodology: model.AttributionMethodInfluence,

			ConversionEvent: model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:    10,
			TacticOfferType: model.MarketingEventTypeOffer,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_111111"))
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_222222"))
		assert.Equal(t, int64(-1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_333333"))
	})

	t.Run("AttributionQueryWShapedLandingMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                   timestamp,
			To:                     timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyLandingPage,
			AttributionMethodology: model.AttributionMethodWShaped,

			ConversionEvent: model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:    10,
			TacticOfferType: model.MarketingEventTypeOffer,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_111111"))
		assert.Equal(t, float64(1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_222222"))
		assert.Equal(t, int64(-1), getConversionUserCountLandingPage(query.AttributionKey, result, "lp_333333"))
	})

}

func TestAttributionEngagementModel(t *testing.T) {
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
	assert.NotNil(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "111111", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

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
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
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
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
	})

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+5*U.SECONDS_IN_A_DAY, "222222", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+5*U.SECONDS_IN_A_DAY, "333333", "", "", "", "", "")
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
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "333333"))
	})

	// linked event for user1
	errCode = createEventWithSession(project.ID, "event2", createdUserID1,
		timestamp+6*U.SECONDS_IN_A_DAY, "1234567", "", "", "", "", "")
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
		var debugQueryKey string
		//Should only have user2 with no 0 linked event count
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "111111"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "222222"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "333333"))
		// no hit for campaigns 1234567 or none
		// Assert below won't work with ExecuteAttributionQueryV1, because for camp '123456' event is 'event2'
		// While attributing, we pull users for 'event1' and not by default all sessions. Hence no longer valid.
		// assert.Equal(t, float64(0), getConversionUserCount(query.AttributionKey, result, "1234567"))
	})
}
func TestWShapedAttributionModel(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	addMarketingData(t, project)

	timestamp := int64(1589068800)

	//Creating 3 users
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryWShapedCampaignMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodWShaped,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(0.000001), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))

	})

	t.Run("AttributionQueryWShapedAdgroupMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodWShaped,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionQueryWShapedKeywordMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodWShaped,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// with "$none" match type for 2 users
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"$none"+model.KeyDelimiter+"Keyword_Adwords_300"))
		// with 'Board' match type for 1 user1
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})

}

func TestInfluenceAttributionModel(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	addMarketingData(t, project)

	timestamp := int64(1589068800)

	//Creating 3 users
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryInfluenceCampaignMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodInfluence,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(0.000001), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))

	})

	t.Run("AttributionQueryInfluenceAdgroupMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodInfluence,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionQueryInfluenceKeywordMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodInfluence,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// with "$none" match type for 2 users
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"$none"+model.KeyDelimiter+"Keyword_Adwords_300"))
		// with 'Board' match type for 1 user1
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})

}

func TestAttributionModelEndToEndWithEnrichment(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	addMarketingData(t, project)

	timestamp := int64(1589068800)

	// Creating 3 users
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	/*t.Run("AttributionWithMarketingPropertyCampaign", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodFirstTouch,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey)
		assert.Nil(t, err)
		assert.Equal(t, "adwords", result.Rows[1][0])
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(0.000001), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
	})

	t.Run("AttributionWithMarketingPropertySource", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeySource,
			AttributionMethodology:  model.AttributionMethodFirstTouch,
			AttributionKeyDimension: []string{model.FieldSource},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey)
		assert.Nil(t, err)
		assert.Equal(t, "google", result.Rows[1][0])

	})

	t.Run("AttributionWithMarketingPropertyChannel", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyChannel,
			AttributionMethodology:  model.AttributionMethodFirstTouch,
			AttributionKeyDimension: []string{model.FieldChannelGroup},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey)
		assert.Nil(t, err)
		assert.Equal(t, "Paid Search", result.Rows[1][0])
	})

	t.Run("AttributionWithMarketingPropertyAdgroup", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodFirstTouch,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey)
		assert.Nil(t, err)
		// Conversion.
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionWithMarketingPropertyKeyword", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 3*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodFirstTouch,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey)
		assert.Nil(t, err)
		// Added keys.
		assert.Equal(t, "adwords", result.Rows[1][0])
		assert.Equal(t, "Campaign_Adwords_100", result.Rows[1][1])
		assert.Equal(t, "Adgroup_Adwords_200", result.Rows[1][2])
		assert.Equal(t, "Broad", result.Rows[1][3])
		// Conversion.
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})*/

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_100",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryLinearCampaignMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodLinear,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, int64(1), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
		assert.Equal(t, float64(0.000001), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"))
	})

	t.Run("AttributionQueryLinearAdgroupMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodLinear,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionQueryLinearKeywordMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodLinear,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// with "$none" match type for 2 users
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"$none"+model.KeyDelimiter+"Keyword_Adwords_300"))
		// with 'Board' match type for 1 user1
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_100"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})
}

func TestAttributionModelWithDifferentCampaignNames(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitAppRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	addMarketingDataWithDifferentCampaignNames(t, project)

	timestamp := int64(1589068800)

	// Creating 3 users
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createdUserID1)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID2)
	assert.Equal(t, http.StatusCreated, errCode)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotNil(t, createdUserID3)
	assert.Equal(t, http.StatusCreated, errCode)

	// Events with +1 Days
	errCode = createEventWithSession(project.ID, "event1", createdUserID1,
		timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_Old",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID2, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_Old",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1",
		createdUserID3, timestamp+1*U.SECONDS_IN_A_DAY, "Campaign_Adwords_Old",
		"Adgroup_Adwords_200", "Keyword_Adwords_300", "", "google", "")
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("AttributionQueryWShapedCampaignMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodWShaped,
			AttributionKeyDimension: []string{model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "Campaign_Adwords_Old"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "Campaign_Adwords_New"))

	})
	t.Run("AttributionQueryWShapedAdgroupMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodWShaped,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_Old"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionQueryWShapedKeywordMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodWShaped,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// with "$none" match type for 2 users
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_Old"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"$none"+model.KeyDelimiter+"Keyword_Adwords_300"))
		// with 'Board' match type for 1 user1
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})

	t.Run("AttributionQueryInfluenceCampaignMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodInfluence,
			AttributionKeyDimension: []string{model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "Campaign_Adwords_Old"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "Campaign_Adwords_New"))

	})

	t.Run("AttributionQueryInfluenceAdgroupMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodInfluence,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_Old"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionQueryInfluenceKeywordMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodInfluence,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// with "$none" match type for 2 users
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_Old"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"$none"+model.KeyDelimiter+"Keyword_Adwords_300"))
		// with 'Board' match type for 1 user1
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})

	t.Run("AttributionQueryLinearCampaignMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyCampaign,
			AttributionMethodology:  model.AttributionMethodLinear,
			AttributionKeyDimension: []string{model.FieldCampaignName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "Campaign_Adwords_Old"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "Campaign_Adwords_New"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "Campaign_Adwords_New"))
	})

	t.Run("AttributionQueryLinearAdgroupMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyAdgroup,
			AttributionMethodology:  model.AttributionMethodLinear,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_Old"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, int64(2), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
		assert.Equal(t, float64(0.000002), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"))
	})

	t.Run("AttributionQueryLinearKeywordMultiUserSession", func(t *testing.T) {
		query := &model.AttributionQuery{
			From:                    timestamp,
			To:                      timestamp + 4*U.SECONDS_IN_A_DAY,
			AttributionKey:          model.AttributionKeyKeyword,
			AttributionMethodology:  model.AttributionMethodLinear,
			AttributionKeyDimension: []string{model.FieldChannelName, model.FieldCampaignName, model.FieldAdgroupName, model.FieldKeywordMatchType, model.FieldKeyword},
			ConversionEvent:         model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:            10,
		}
		var debugQueryKey string
		result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// with "$none" match type for 2 users
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"+model.KeyDelimiter+"Campaign_Adwords_Old"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"$none"+model.KeyDelimiter+"Keyword_Adwords_300"))
		// with 'Board' match type for 1 user1
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getImpressions(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, int64(3), getClicks(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
		assert.Equal(t, float64(0.000003), getSpend(query.AttributionKey, result, "adwords"+model.KeyDelimiter+"Campaign_Adwords_New"+model.KeyDelimiter+"Adgroup_Adwords_200"+model.KeyDelimiter+"Broad"+model.KeyDelimiter+"Keyword_Adwords_300"))
	})
}

func addMarketingData(t *testing.T, project *model.Project) {

	adwordsCustomerAccountId := U.RandomLowerAphaNumString(5)
	fbCustomerAccountId := U.RandomLowerAphaNumString(5)
	linkedinCustomerAccountId := U.RandomLowerAphaNumString(5)

	// Setting up Adwords, FB and Linkedin account
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &adwordsCustomerAccountId,
		IntFacebookAdAccount:        fbCustomerAccountId,
		IntLinkedinAdAccount:        linkedinCustomerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	/*
		Adwords:
		CampaignID: 100, CampaignName: "Campaign_Adwords_100"
		AdgroupID: 200, AdgroupName: "Adgroup_Adwords_200"
		KeywordID: 300, KeywordName: "Keyword_Adwords_300"
		AdID: 400, AdName: "AdName_Adwords_400"
		GCLID: Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs
		FB:
		CampaignID: 1000, CampaignName: "Campaign_Facebook_1000"
		AdgroupID: 2000, AdgroupName: "Adgroup_Facebook_2000"

		Linkedin:
		CampaignID: 10000, CampaignName: "Campaign_Linkedin_10000"
		AdgroupID: 20000, AdgroupName: "Adgroup_Linkedin_20000"
	*/

	// Adwords Campaign performance report
	value := []byte(`{"cost": "1","clicks": "1","campaign_id":"100","impressions": "1", "campaign_name": "Campaign_Adwords_100"}`)
	document := &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "100",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.CampaignPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        100,
	}
	status := store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// Adwords Adgroup performance report
	value = []byte(`{"cost": "2","clicks": "2","ad_group_id":"200","impressions": "2", "ad_group_name": "Adgroup_Adwords_200", "campaign_id":"100"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "200",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.AdGroupPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        100,
		AdGroupID:         200,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// Adwords Keyword performance report
	value = []byte(`{"cost": "3","clicks": "3","id":"300","impressions": "3", "criteria": "Keyword_Adwords_300", "ad_group_id":"200", "campaign_id":"100", "keyword_match_type":"Broad"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "300",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.KeywordPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        100,
		AdGroupID:         200,
		KeywordID:         300,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// GCLID performance report
	value = []byte(`{"gcl_id": "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs","ad_group_id": "200","campaign_id":"100","creative_id": "400", "criteria_id": "300", "slot": "Google search: Top"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         "click_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        100,
		AdGroupID:         200,
		KeywordID:         300,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// Facebook Campaign performance report
	value = []byte(`{"spend": "1","clicks": "1","campaign_id":"1000","impressions": "1", "campaign_name": "Campaign_Facebook_1000", "platform":"facebook"}`)
	documentFB := &model.FacebookDocument{
		ProjectID:           project.ID,
		ID:                  "1000",
		CustomerAdAccountID: fbCustomerAccountId,
		TypeAlias:           "campaign_insights",
		Timestamp:           20200510,
		Value:               &postgres.Jsonb{RawMessage: value},
		CampaignID:          "1000",
		Platform:            "facebook",
	}
	status = store.GetStore().CreateFacebookDocument(project.ID, documentFB)
	assert.Equal(t, http.StatusCreated, status)

	// Facebook Adgroup performance report
	value = []byte(`{"spend": "2","clicks": "2","adset_id":"2000","impressions": "2", "adset_name": "Adgroup_Facebook_2000", "platform":"facebook", "campaign_id":"1000"}`)
	documentFB = &model.FacebookDocument{
		ProjectID:           project.ID,
		ID:                  "2000",
		CustomerAdAccountID: fbCustomerAccountId,
		TypeAlias:           "ad_set_insights",
		Timestamp:           20200510,
		Value:               &postgres.Jsonb{RawMessage: value},
		CampaignID:          "1000",
		AdSetID:             "2000",
		Platform:            "facebook",
	}
	status = store.GetStore().CreateFacebookDocument(project.ID, documentFB)
	assert.Equal(t, http.StatusCreated, status)

	// Linkedin Campaign performance report
	value = []byte(`{"costInLocalCurrency": "1","clicks": "1","campaign_group_id":"10000","impressions": "1", "campaign_group_name": "Campaign_Linkedin_10000"}`)
	documentLinkedin := &model.LinkedinDocument{
		ProjectID:           project.ID,
		ID:                  "10000",
		CustomerAdAccountID: linkedinCustomerAccountId,
		TypeAlias:           "campaign_group_insights",
		Timestamp:           20200510,
		Value:               &postgres.Jsonb{RawMessage: value},
		CampaignGroupID:     "10000",
	}
	status = store.GetStore().CreateLinkedinDocument(project.ID, documentLinkedin)
	assert.Equal(t, http.StatusCreated, status)

	// Linkedin Adgroup performance report
	value = []byte(`{"costInLocalCurrency": "2","clicks": "2","campaign_id":"20000","impressions": "2", "campaign_name": "Adgroup_Linkedin_20000", "campaign_group_id":"10000"}`)
	documentLinkedin = &model.LinkedinDocument{
		ProjectID:           project.ID,
		ID:                  "20000",
		CustomerAdAccountID: linkedinCustomerAccountId,
		TypeAlias:           "campaign_insights",
		Timestamp:           20200510,
		Value:               &postgres.Jsonb{RawMessage: value},
		CampaignGroupID:     "10000",
		CampaignID:          "20000",
	}
	status = store.GetStore().CreateLinkedinDocument(project.ID, documentLinkedin)
	assert.Equal(t, http.StatusCreated, status)
}

func addMarketingDataWithDifferentCampaignNames(t *testing.T, project *model.Project) {
	adwordsCustomerAccountId := U.RandomLowerAphaNumString(5)

	// Setting up Adwords account
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &adwordsCustomerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	/*
			Adwords:
			CampaignID: 100, CampaignName: "Campaign_Adwords_Old"
		    CampaignID: 100, CampaignName: "Campaign_Adwords_New"
			AdgroupID: 200, AdgroupName: "Adgroup_Adwords_200"
			KeywordID: 300, KeywordName: "Keyword_Adwords_300"
			AdID: 400, AdName: "AdName_Adwords_400"
			GCLID: Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs

	*/

	// Adwords Campaign performance report for old campaign name
	value := []byte(`{"cost": "1","clicks": "1","campaign_id":"101","impressions": "1", "campaign_name": "Campaign_Adwords_Old"}`)
	document := &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "101",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.CampaignPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        101,
	}
	status := store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// Adwords Campaign performance report for new campaign name
	value = []byte(`{"cost": "1","clicks": "1","campaign_id":"101","impressions": "1", "campaign_name": "Campaign_Adwords_New"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "101",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.CampaignPerformanceReport,
		Timestamp:         20200510 + 1,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        101,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// Adwords Adgroup performance report
	value = []byte(`{"cost": "2","clicks": "2","ad_group_id":"200","impressions": "2", "ad_group_name": "Adgroup_Adwords_200", "campaign_id":"101"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "200",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.AdGroupPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        101,
		AdGroupID:         200,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// Adwords Keyword performance report
	value = []byte(`{"cost": "3","clicks": "3","id":"300","impressions": "3", "criteria": "Keyword_Adwords_300", "ad_group_id":"200", "campaign_id":"101", "keyword_match_type":"Broad"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "300",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.KeywordPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        101,
		AdGroupID:         200,
		KeywordID:         300,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	// GCLID performance report
	value = []byte(`{"gcl_id": "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs","ad_group_id": "200","campaign_id":"101","creative_id": "400", "criteria_id": "300", "slot": "Google search: Top"}`)
	document = &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "Cj0KCQjwmpb0BRCBARIsAG7y4zbZArcUWztiqP5bs",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         "click_performance_report",
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        101,
		AdGroupID:         200,
		KeywordID:         300,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

}

func getConversionUserCount(attributionKey string, result *model.QueryResult, key interface{}) interface{} {

	addedKeysSize := model.GetLastKeyValueIndex(result.Headers)
	conversionIndex := model.GetConversionIndex(result.Headers)

	for _, row := range result.Rows {
		rowKey := getRowKey(addedKeysSize, row)
		if rowKey == key {
			return row[conversionIndex]
		}
	}
	return int64(-1)
}

func getConversionUserCountLandingPage(attributionKey string, result *model.QueryResult, key interface{}) interface{} {

	addedKeysSize := model.GetLastKeyValueIndexLandingPage(result.Headers)
	conversionIndex := model.GetConversionIndex(result.Headers)

	for _, row := range result.Rows {
		rowKey := getRowKey(addedKeysSize, row)
		if rowKey == key {
			return row[conversionIndex]
		}
	}
	return int64(-1)
}
func getRowKey(addedKeysSize int, row []interface{}) string {
	rowKey := ""
	for i := 0; i <= addedKeysSize; i++ {
		rowKey = rowKey + row[i].(string)
		if i < addedKeysSize {
			rowKey = rowKey + model.KeyDelimiter
		}
	}
	return rowKey
}

func getClicks(attributionKey string, result *model.QueryResult, key interface{}) interface{} {

	addedKeysSize := model.GetLastKeyValueIndex(result.Headers)
	for _, row := range result.Rows {
		rowKey := getRowKey(addedKeysSize, row)
		if rowKey == key {
			return row[addedKeysSize+2]
		}
	}
	return int64(-1)
}

func getSpend(attributionKey string, result *model.QueryResult, key interface{}) interface{} {

	addedKeysSize := model.GetLastKeyValueIndex(result.Headers)
	for _, row := range result.Rows {
		rowKey := getRowKey(addedKeysSize, row)
		if rowKey == key {
			return row[addedKeysSize+3]
		}
	}
	return int64(-1)
}

func getImpressions(attributionKey string, result *model.QueryResult, key interface{}) interface{} {

	addedKeysSize := model.GetLastKeyValueIndex(result.Headers)
	impressionIndex := model.GetImpressionsIndex(result.Headers)
	for _, row := range result.Rows {
		rowKey := getRowKey(addedKeysSize, row)
		if rowKey == key {
			return row[impressionIndex]
		}
	}
	return int64(-1)
}

func getCompareConversionUserCount(attributionKey string, result *model.QueryResult, key interface{}) interface{} {

	addedKeysSize := model.GetLastKeyValueIndex(result.Headers)
	ccuIndex := model.GetCompareConversionUserCountIndex(result.Headers)
	for _, row := range result.Rows {
		rowKey := getRowKey(addedKeysSize, row)
		if rowKey == key {
			return row[ccuIndex]
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
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{RawMessage: user1PropertiesBytes},
		JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	_, errCode = createSession(project.ID, createdUserID1, timestamp+4*U.SECONDS_IN_A_DAY, "", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)
	userEventName, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "event1"})
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "event2"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: userEventName.ID, UserId: createdUserID1,
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
	var debugQueryKey string
	// Should find within look back window
	_, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)

	_, errCode = store.GetStore().CreateEvent(&model.Event{ProjectId: project.ID, EventNameId: userEventName.ID,
		UserId: createdUserID1, Timestamp: timestamp + 5*U.SECONDS_IN_A_DAY})
	assert.Equal(t, http.StatusCreated, errCode)

	_, errCode = createSession(project.ID, createdUserID1, timestamp+8*U.SECONDS_IN_A_DAY, "", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)
	query.From = timestamp + 5*U.SECONDS_IN_A_DAY
	// Event beyond lookback window.
	_, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
}

func TestAttributionTacticOffer(t *testing.T) {

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)

	timestamp := int64(1589068800)

	errCode = createEventWithSession(project.ID, "event1", createdUserID1, timestamp, "", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", createdUserID2, timestamp, "", "", "", "", "", "")
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
	var debugQueryKey string
	//both user should be treated different
	result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
	// Lookback is 0. There should be no attribution.
	// Attribution Time: 1589068798, Conversion Time: 1589068800, diff = 2 secs

	customerUserId := U.RandomLowerAphaNumString(15)
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1, &model.User{CustomerUserId: customerUserId}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID2, &model.User{CustomerUserId: customerUserId}, timestamp+1*U.SECONDS_IN_A_DAY)
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
	result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
	// both user should be treated same
	assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "$none"))
	// continuation to previous users
	_, status := store.GetStore().UpdateUserProperties(project.ID, createdUserID1, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
	assert.Equal(t, http.StatusAccepted, status)
	_, status = store.GetStore().UpdateUserProperties(project.ID, createdUserID2, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
	assert.True(t, status == http.StatusAccepted)
	/*
		t+3day -> first time $initial_campaign set with event for user1 and user2
		t+6day -> session event for user1 and user2
	*/
	status = createEventWithSession(project.ID, "event1", createdUserID1, timestamp+3*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, status)
	status = createEventWithSession(project.ID, "event1", createdUserID2, timestamp+3*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, status)
	status = createEventWithSession(project.ID, "event1", createdUserID1, timestamp+6*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, status)
	status = createEventWithSession(project.ID, "event1", createdUserID2, timestamp+6*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, status)

	t.Run("TestAttributionTacticOffer", func(t *testing.T) {

		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 7*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           4,
			QueryType:              model.AttributionQueryTypeEngagementBased,
			TacticOfferType:        model.MarketingEventTypeTacticOffer,
		}
		var debugQueryKey string
		result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// The attribution didn't happen in the query period. First touch is on 3rd day and which
		// is not between 4th to 7th (query period). Hence count is 0.
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "12345"))
	})
	t.Run("TestAttributionTactic", func(t *testing.T) {

		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 7*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           4,
			QueryType:              model.AttributionQueryTypeEngagementBased,
			TacticOfferType:        model.MarketingEventTypeTactic,
		}
		var debugQueryKey string
		result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// The attribution didn't happen in the query period. First touch is on 3rd day and which
		// is not between 4th to 7th (query period). Hence count is 0.
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "12345"))
	})
	t.Run("TestAttributionOffer", func(t *testing.T) {

		query := &model.AttributionQuery{
			From:                   timestamp + 4*U.SECONDS_IN_A_DAY,
			To:                     timestamp + 7*U.SECONDS_IN_A_DAY,
			AttributionKey:         model.AttributionKeyCampaign,
			AttributionMethodology: model.AttributionMethodFirstTouch,
			ConversionEvent:        model.QueryEventWithProperties{Name: "event1"},
			LookbackDays:           4,
			QueryType:              model.AttributionQueryTypeEngagementBased,
			TacticOfferType:        model.MarketingEventTypeOffer,
		}
		var debugQueryKey string
		result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		// The query contains sessions only, hence no attribution for OTP i.e. for "Offer"
		assert.Equal(t, int64(-1), getConversionUserCount(query.AttributionKey, result, "12345"))
	})

}

func TestAttributionWithUserIdentification(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)

	timestamp := int64(1589068800)

	errCode = createEventWithSession(project.ID, "event1", createdUserID1, timestamp, "", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", createdUserID2, timestamp, "", "", "", "", "", "")
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
	var debugQueryKey string
	//both user should be treated different
	result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
	// Lookback is 0. There should be no attribution.
	// Attribution Time: 1589068798, Conversion Time: 1589068800, diff = 2 secs

	customerUserId := U.RandomLowerAphaNumString(15)
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1, &model.User{CustomerUserId: customerUserId}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID2, &model.User{CustomerUserId: customerUserId}, timestamp+1*U.SECONDS_IN_A_DAY)
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
	result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
	// both user should be treated same
	assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "$none"))

	t.Run("TestAttributionUserIdentificationWithLookbackDays0", func(t *testing.T) {
		// continuation to previous users
		_, status := store.GetStore().UpdateUserProperties(project.ID, createdUserID1, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
		assert.Equal(t, http.StatusAccepted, status)
		_, status = store.GetStore().UpdateUserProperties(project.ID, createdUserID2, &postgres.Jsonb{RawMessage: json.RawMessage(`{"$initial_campaign":12345}`)}, timestamp+3*86400)
		assert.True(t, status == http.StatusAccepted)
		/*
			t+3day -> first time $initial_campaign set with event for user1 and user2
			t+6day -> session event for user1 and user2
		*/
		status = createEventWithSession(project.ID, "event1", createdUserID1, timestamp+3*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", createdUserID2, timestamp+3*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", createdUserID1, timestamp+6*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "event1", createdUserID2, timestamp+6*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
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
		var debugQueryKey string
		result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())

		assert.Nil(t, err)
		// The attribution didn't happen in the query period. First touch is on 3rd day and which
		// is not between 4th to 7th (query period). Hence count is 0.
		assert.Equal(t, float64(1), getConversionUserCount(query.AttributionKey, result, "12345"))
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

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)

	timestamp := int64(1589068800)

	errCode = createEventWithSession(project.ID, "event1", createdUserID1, timestamp, "", "", "", "", "", "")
	assert.Equal(t, http.StatusCreated, errCode)

	errCode = createEventWithSession(project.ID, "event1", createdUserID2, timestamp, "", "", "", "", "", "")
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
	var debugQueryKey string
	// Both user should be treated different
	result, err := store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)

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
	result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
	// Lookback days = 2
	assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"))

	customerUserId1 := U.RandomLowerAphaNumString(15) + "_one"
	customerUserId2 := U.RandomLowerAphaNumString(15) + "_two"
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1, &model.User{CustomerUserId: customerUserId1}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID2, &model.User{CustomerUserId: customerUserId2}, timestamp+1*U.SECONDS_IN_A_DAY)
	assert.Equal(t, http.StatusAccepted, errCode)
	result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
	assert.Nil(t, err)
	assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "$none"))

	t.Run("TestAttributionUserIdentificationWithLookbackDays", func(t *testing.T) {
		// 3 days is out of query window, but should be considered as it falls under Engagement window
		status := createEventWithSession(project.ID, "eventNewX", createdUserID1, timestamp+5*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
		assert.Equal(t, http.StatusCreated, status)
		status = createEventWithSession(project.ID, "eventNewX", createdUserID2, timestamp+6*U.SECONDS_IN_A_DAY, "12345", "", "", "", "", "")
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
		var debugQueryKey string
		result, err = store.GetStore().ExecuteAttributionQueryV1(project.ID, query, debugQueryKey, C.EnableOptimisedFilterOnProfileQuery(), C.EnableOptimisedFilterOnEventUserQuery())
		assert.Nil(t, err)
		assert.Equal(t, float64(2), getConversionUserCount(query.AttributionKey, result, "12345"))
	})
}

func isEqualGotAndWantAttribution(gotAttribution []model.AttributionKeyWeight, wantAttribution []model.AttributionKeyWeight) bool {
	wantCopy := make([]model.AttributionKeyWeight, len(wantAttribution))
	copy(wantCopy, wantAttribution)
	if len(gotAttribution) != len(wantCopy) {
		return false
	}
	for _, item1 := range gotAttribution {
		foundSameItem := false
		for i := 0; i < len(wantCopy); i++ {
			if item1 == wantCopy[i] {
				wantCopy = append(wantCopy[:i], wantCopy[i+1:]...)
				foundSameItem = true
				break
			}
		}
		if !foundSameItem {
			return false
		}
	}
	return true
}

func TestAttributionMethodologies(t *testing.T) {

	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp1 := "adwords:-:campaign1"
	camp2 := "adwords:-:campaign2"
	camp3 := "adwords:-:campaign3"

	queryFrom := 0
	queryTo := 10000000
	userSession := make(map[string]map[string]model.UserSessionData)
	userSession[user1] = make(map[string]model.UserSessionData)
	userSession[user1][camp1] = model.UserSessionData{MinTimestamp: 105, MaxTimestamp: 200, TimeStamps: []int64{100, 200}}
	userSession[user1][camp2] = model.UserSessionData{MinTimestamp: 150, MaxTimestamp: 300, TimeStamps: []int64{150, 300}}
	userSession[user1][camp3] = model.UserSessionData{MinTimestamp: 50, MaxTimestamp: 604950, TimeStamps: []int64{50, 604950}}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 604900
	//seconds in 7 days=604800
	lookbackDays := 20

	userSession2 := make(map[string]map[string]model.UserSessionData)
	userSession2[user1] = make(map[string]model.UserSessionData)
	userSession2[user1][camp1] = model.UserSessionData{MinTimestamp: 110, MaxTimestamp: 200, TimeStamps: []int64{100, 200}}
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionData
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
		queryType                     string
		attributionKey                string
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]model.AttributionKeyWeight
		wantLinkedEventUserCampaign map[string]map[string][]model.AttributionKeyWeight
		wantErr                     bool
	}{

		// Test for LINEAR_TOUCH
		{"linear_touch",
			args{model.AttributionMethodLinear,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp1, Weight: 0.2}, {Key: camp1, Weight: 0.2}, {Key: camp2, Weight: 0.2}, {Key: camp2, Weight: 0.2}, {Key: camp3, Weight: 0.2}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{model.AttributionMethodFirstTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp3, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{model.AttributionMethodLastTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test1 for U_SHAPED
		{"u_shaped",
			args{model.AttributionMethodUShaped,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp3, Weight: 0.5}, {Key: camp2, Weight: 0.5}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test2 for U_SHAPED
		{"u_shaped",
			args{model.AttributionMethodUShaped,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession2,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp1, Weight: 0.5}, {Key: camp1, Weight: 0.5}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test1 for Time_Decay
		{"time_decay",
			args{model.AttributionMethodTimeDecay,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp1, Weight: 0.20783766956583852}, {Key: camp2, Weight: 0.20783766956583852}, {Key: camp1, Weight: 0.18824349565124227}, {Key: camp2, Weight: 0.20783766956583852}, {Key: camp3, Weight: 0.18824349565124227}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test2 for Time_Decay
		{"time_decay",
			args{model.AttributionMethodTimeDecay,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession2,
				coalUserIdConversionTimestamp,
				lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp1, Weight: 0.4752649511825976}, {Key: camp1, Weight: 0.5247350488174024}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := model.ApplyAttribution(tt.args.queryType, tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession, tt.args.coalUserIdConversionTimestamp,
				tt.args.lookbackDays, int64(queryFrom), int64(queryTo), tt.args.attributionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for key, _ := range tt.wantUsersAttribution {
				if isEqualGotAndWantAttribution(got[key], tt.wantUsersAttribution[key]) == false {
					t.Errorf("applyAttribution() got = %v, want %v", got, tt.wantUsersAttribution)
				}
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
	camp0 := "$none:-:$none"
	camp1 := "adwords:-:campaign1"
	camp2 := "adwords:-:campaign2"
	camp3 := "adwords:-:campaign3"
	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionData)
	userSession[user1] = make(map[string]model.UserSessionData)
	userSession[user1][camp0] = model.UserSessionData{MinTimestamp: 10, MaxTimestamp: 40, TimeStamps: []int64{10, 40}}
	userSession[user1][camp1] = model.UserSessionData{MinTimestamp: 100, MaxTimestamp: 200, TimeStamps: []int64{100, 200}}
	userSession[user1][camp2] = model.UserSessionData{MinTimestamp: 150, MaxTimestamp: 300, TimeStamps: []int64{150, 300}}
	userSession[user1][camp3] = model.UserSessionData{MinTimestamp: 50, MaxTimestamp: 100, TimeStamps: []int64{50, 100}}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionData
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
		queryType                     string
		attributionKey                string
	}
	const weightOneBySix = float64(1.0 / 6.0)
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]model.AttributionKeyWeight
		wantLinkedEventUserCampaign map[string]map[string][]model.AttributionKeyWeight
		wantErr                     bool
	}{
		// Test for LINEAR_TOUCH
		{"linear_touch",
			args{model.AttributionMethodLinear,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp0, Weight: weightOneBySix}, {Key: camp0, Weight: weightOneBySix}, {Key: camp1, Weight: weightOneBySix}, {Key: camp2, Weight: weightOneBySix}, {Key: camp3, Weight: weightOneBySix}, {Key: camp3, Weight: weightOneBySix}}},

			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{model.AttributionMethodFirstTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp0, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{model.AttributionMethodLastTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{model.AttributionMethodFirstTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp3, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{model.AttributionMethodLastTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := model.ApplyAttribution(tt.args.queryType, tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession,
				tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays,
				int64(queryFrom), int64(queryTo), tt.args.attributionKey)
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
	camp1 := "adwords:-:campaign1"
	camp2 := "adwords:-:campaign2"
	camp3 := "adwords:-:campaign3"
	camp4 := "$none:-:$none"
	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionData)
	userSession[user1] = make(map[string]model.UserSessionData)
	userSession[user1][camp1] = model.UserSessionData{MinTimestamp: 100, MaxTimestamp: 200, TimeStamps: []int64{100, 200}}
	userSession[user1][camp2] = model.UserSessionData{MinTimestamp: 150, MaxTimestamp: 300, TimeStamps: []int64{150, 300}}
	userSession[user1][camp3] = model.UserSessionData{MinTimestamp: 50, MaxTimestamp: 100, TimeStamps: []int64{50, 100}}
	userSession[user1][camp4] = model.UserSessionData{MinTimestamp: 10, MaxTimestamp: 400, TimeStamps: []int64{10, 400}}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionData
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
		queryType                     string
		attributionKey                string
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]model.AttributionKeyWeight
		wantLinkedEventUserCampaign map[string]map[string][]model.AttributionKeyWeight
		wantErr                     bool
	}{
		// Test for LINEAR_TOUCH
		{"linear_touch",
			args{model.AttributionMethodLinear,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp1, Weight: 0.2}, {Key: camp2, Weight: 0.2}, {Key: camp3, Weight: 0.2}, {Key: camp3, Weight: 0.2}, {Key: camp4, Weight: 0.2}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for FIRST_TOUCH
		{"first_touch",
			args{model.AttributionMethodFirstTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp4, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH
		{"last_touch",
			args{model.AttributionMethodLastTouch,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{model.AttributionMethodFirstTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp3, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{model.AttributionMethodLastTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyCampaign,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := model.ApplyAttribution(tt.args.queryType, tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession,
				tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays, int64(queryFrom), int64(queryTo), tt.args.attributionKey)
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

func TestAttributionMethodologiesNonDirectAdgroup(t *testing.T) {
	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp0 := "$none:-:$none:-:$none"
	camp1 := "adwords:-:campaign1:-:adgroup1"
	camp2 := "adwords:-:campaign2:-:adgroup2"
	camp3 := "adwords:-:campaign3:-:adgroup3"
	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionData)
	userSession[user1] = make(map[string]model.UserSessionData)
	userSession[user1][camp0] = model.UserSessionData{MinTimestamp: 10, MaxTimestamp: 40, TimeStamps: []int64{10, 40}}
	userSession[user1][camp1] = model.UserSessionData{MinTimestamp: 100, MaxTimestamp: 200, TimeStamps: []int64{100, 200}}
	userSession[user1][camp2] = model.UserSessionData{MinTimestamp: 150, MaxTimestamp: 300, TimeStamps: []int64{150, 300}}
	userSession[user1][camp3] = model.UserSessionData{MinTimestamp: 50, MaxTimestamp: 100, TimeStamps: []int64{50, 100}}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionData
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
		queryType                     string
		attributionKey                string
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]model.AttributionKeyWeight
		wantLinkedEventUserCampaign map[string]map[string][]model.AttributionKeyWeight
		wantErr                     bool
	}{
		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{model.AttributionMethodFirstTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyAdgroup,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp3, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{model.AttributionMethodLastTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyAdgroup,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := model.ApplyAttribution(tt.args.queryType, tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession, tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays,
				int64(queryFrom), int64(queryTo), tt.args.attributionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantUsersAttribution) {
				t.Errorf("applyAttribution() got = %v, want %v", got, tt.wantUsersAttribution)
			}
			if !reflect.DeepEqual(got1, tt.wantLinkedEventUserCampaign) {
				t.Errorf("applyAttribution() got1 = %v, want %v", got1, tt.wantLinkedEventUserCampaign)
			}
		})
	}
}

func TestAttributionMethodologiesNonDirectKeyword(t *testing.T) {
	conversionEvent := "$Form_Submitted"
	user1 := "user1"
	camp0 := "$none:-:$none:-:$none:-:$none:-:$none"
	camp1 := "adwords:-:campaign1:-:adgroup1:-:match_type1:-:key_word_name1"
	camp2 := "adwords:-:campaign2:-:adgroup2:-:match_type2:-:key_word_name2"
	camp3 := "adwords:-:campaign3:-:adgroup3:-:match_type3:-:key_word_name3"
	queryFrom := 0
	queryTo := 1000
	userSession := make(map[string]map[string]model.UserSessionData)
	userSession[user1] = make(map[string]model.UserSessionData)
	userSession[user1][camp0] = model.UserSessionData{MinTimestamp: 10, MaxTimestamp: 40, TimeStamps: []int64{10, 40}}
	userSession[user1][camp1] = model.UserSessionData{MinTimestamp: 100, MaxTimestamp: 200, TimeStamps: []int64{100, 200}}
	userSession[user1][camp2] = model.UserSessionData{MinTimestamp: 150, MaxTimestamp: 300, TimeStamps: []int64{150, 300}}
	userSession[user1][camp3] = model.UserSessionData{MinTimestamp: 50, MaxTimestamp: 100, TimeStamps: []int64{50, 100}}
	coalUserIdConversionTimestamp := make(map[string]int64)
	coalUserIdConversionTimestamp[user1] = 150
	lookbackDays := 1
	type args struct {
		method                        string
		conversionEvent               string
		usersToBeAttributed           []model.UserEventInfo
		userInitialSession            map[string]map[string]model.UserSessionData
		coalUserIdConversionTimestamp map[string]int64
		lookbackDays                  int
		queryType                     string
		attributionKey                string
	}
	tests := []struct {
		name                        string
		args                        args
		wantUsersAttribution        map[string][]model.AttributionKeyWeight
		wantLinkedEventUserCampaign map[string]map[string][]model.AttributionKeyWeight
		wantErr                     bool
	}{
		// Test for FIRST_TOUCH_ND
		{"first_touch_nd",
			args{model.AttributionMethodFirstTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyKeyword,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp3, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},

		// Test for LAST_TOUCH_ND
		{"last_touch_nd",
			args{model.AttributionMethodLastTouchNonDirect,
				conversionEvent,
				[]model.UserEventInfo{{user1, conversionEvent, coalUserIdConversionTimestamp[user1], model.EventTypeGoalEvent}},
				userSession,
				coalUserIdConversionTimestamp, lookbackDays,
				model.AttributionQueryTypeConversionBased,
				model.AttributionKeyKeyword,
			},
			map[string][]model.AttributionKeyWeight{user1: {{Key: camp2, Weight: 1}}},
			map[string]map[string][]model.AttributionKeyWeight{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := model.ApplyAttribution(tt.args.queryType, tt.args.method, tt.args.conversionEvent,
				tt.args.usersToBeAttributed, tt.args.userInitialSession, tt.args.coalUserIdConversionTimestamp, tt.args.lookbackDays,
				int64(queryFrom), int64(queryTo), tt.args.attributionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAttribution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantUsersAttribution) {
				t.Errorf("applyAttribution() got = %v, want %v", got, tt.wantUsersAttribution)
			}
			if !reflect.DeepEqual(got1, tt.wantLinkedEventUserCampaign) {
				t.Errorf("applyAttribution() got1 = %v, want %v", got1, tt.wantLinkedEventUserCampaign)
			}
		})
	}
}

func TestMergeDataRowsHavingSameKey(t *testing.T) {

	rows := make([][]interface{}, 0)
	row1 := []interface{}{"Campaign1", int64(2), int64(2), float64(2),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(2), float64(2), float64(2), float64(2),
		// ConversionEventCount, CostPerConversion, ConversionEventCompareCount, CostPerConversionCompareCount
		float64(2), float64(2), float64(2), float64(2), float64(2), float64(2)}
	row2 := []interface{}{"Campaign1", int64(3), int64(3), float64(3),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(3), float64(3), float64(3), float64(3),
		// ConversionEventCount, CostPerConversion, ConversionEventCompareCount, CostPerConversionCompareCount
		float64(3), float64(3), float64(3), float64(3), float64(3), float64(3)}
	rows = append(rows, row1, row2)

	mergedRows := make([][]interface{}, 0)
	row3 := []interface{}{"Campaign1", int64(5), int64(5), float64(5),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(100), float64(1), float64(1000), float64(100),
		// ConversionEventCount, CostPerConversion, ConversionEventCompareCount, CostPerConversionCompareCount
		float64(5), float64(5), float64(1), float64(5), float64(5), float64(1)}

	logCtx := log.Entry{}
	mergedRows = append(mergedRows, row3)
	type args struct {
		rows [][]interface{}
	}
	tests := []struct {
		name string
		args args
		want [][]interface{}
	}{
		{"SimpleX", args{rows}, mergedRows},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.MergeDataRowsHavingSameKey(tt.args.rows, 0, model.AttributionKeyCampaign, model.AnalyzeTypeUsers, nil, logCtx)
			for rowNo, _ := range got {
				for colNo, _ := range got[rowNo] {
					if got[rowNo][colNo] != tt.want[rowNo][colNo] {
						t.Errorf("MergeDataRowsHavingSameKey()col No: %v=,  = %v, want %v", colNo, got[rowNo][colNo], tt.want[rowNo][colNo])
					}
				}
			}
		})
	}
}

func TestAddGrandTotalRow(t *testing.T) {
	headers := []string{"Campaign", "Impressions", "Clicks", "Spend",
		"CTR(%)", "Average CPC", "CPM", "ClickConversionRate(%)",
		"$session - Users", "$session - Users (Influence)", "Cost Per Conversion", "Compare - Users", "Compare - Users (Influence)", "Compare Cost Per Conversion", "Key"}
	rows := make([][]interface{}, 0)
	// Name, impression, clicks, spend
	row1 := []interface{}{"Campaign1", int64(2), int64(2), float64(2),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(2), float64(2), float64(2), float64(2),
		// ConversionEventCount,ConversionEventCount(influence), CostPerConversion, ConversionEventCompareCount, ConversionEventCompareCountInfluence,CostPerConversionCompareCount
		float64(2), float64(2), float64(2), float64(2), float64(2), float64(2), "key1"}
	row2 := []interface{}{"Campaign2", int64(3), int64(3), float64(3),
		// (4_CTR, 5_AvgCPC, 6_CPM, 7_ClickConversionRate)
		float64(3), float64(3), float64(3), float64(3),
		// 12_ConversionEventCount, 13_ConversionEventCount, 14_CostPerConversion, 15_ConversionEventCompareCount, 16_CostPerConversionCompareCount
		float64(3), float64(3), float64(3), float64(3), float64(3), float64(3), "key2"}
	rows = append(rows, row1, row2)

	row3 := []interface{}{"Grand Total", int64(5), int64(5), float64(5),
		// (CTR, AvgCPC, CPM, ClickConversionRate)
		float64(100), float64(1), float64(1000), float64(100),
		// ConversionEventCount, CostPerConversion, ConversionEventCompareCount, CostPerConversionCompareCount
		float64(5), float64(5), float64(1), float64(5), float64(5), float64(1), "Grand Total"}

	resultWant := append([][]interface{}{row3}, rows...)
	type args struct {
		headers []string
		rows    [][]interface{}
		method  string
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{"SimpleX", args{headers, rows, ""}, resultWant[0]},
		//for first touch
		{"First_Touch", args{headers, rows, model.AttributionMethodFirstTouch}, resultWant[0]},
		//for last touch
		{"Last_Touch", args{headers, rows, model.AttributionMethodLastTouch}, resultWant[0]},
		//for first touch non-direct
		{"First_Touch_ND", args{headers, rows, model.AttributionMethodFirstTouchNonDirect}, resultWant[0]},
		//for last touch non-direct
		{"Last_Touch_ND", args{headers, rows, model.AttributionMethodLastTouchNonDirect}, resultWant[0]},
		//for linear touch
		{"Linear", args{headers, rows, model.AttributionMethodLinear}, resultWant[0]},
		//for U Shape
		{"U_Shaped", args{headers, rows, model.AttributionMethodUShaped}, resultWant[0]},
		//for time decay
		{"Time_Decay", args{headers, rows, model.AttributionMethodTimeDecay}, resultWant[0]},
		//for Influence
		{"Influence", args{headers, rows, model.AttributionMethodInfluence}, resultWant[0]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resultGot := model.AddGrandTotalRow(tt.args.headers, tt.args.rows, 0, model.AnalyzeTypeUsers, nil, tt.args.method, "")
			got := resultGot[0]
			for colNo, _ := range got {
				if got[colNo] != tt.want[colNo] {
					t.Errorf("TestAddGrandTotalRow()col No: %v=,  = %v, want %v", colNo, got[colNo], tt.want[colNo])
				}
			}
		})
	}
}

func TestSortInteractionTime(t *testing.T) {
	type args struct {
		interactions []model.Interaction
		sortingType  string
	}
	sortedAsc1 := []model.Interaction{
		{"key1", int64(111)},
		{"key2", int64(112)},
		{"key3", int64(113)},
		{"key4", int64(114)},
	}
	sortedDesc1 := []model.Interaction{
		{"key4", int64(114)},
		{"key3", int64(113)},
		{"key2", int64(112)},
		{"key1", int64(111)},
	}

	sortedEqual1 := []model.Interaction{
		{"key1", int64(112)},
		{"key2", int64(112)},
		{"key3", int64(112)},
		{"key4", int64(112)},
	}

	sortedSomeEqual1 := []model.Interaction{
		{"key1", int64(111)},
		{"key2", int64(112)},
		{"key3", int64(112)},
		{"key4", int64(114)},
	}
	sortedDescSomeEqual1 := []model.Interaction{
		{"key4", int64(114)},
		{"key2", int64(112)},
		{"key3", int64(112)},
		{"key1", int64(111)},
	}
	/*sortedEqual1 := []model.Interaction{
		{"key1", int64(112)},
		{"key2", int64(112)},
		{"key3", int64(112)},
		{"key4", int64(112)},
	}*/

	tests := []struct {
		name string
		args args
		want []model.Interaction
	}{
		{"Test1", args{interactions: sortedAsc1, sortingType: model.SortASC}, sortedAsc1},
		{"Test2", args{interactions: sortedAsc1, sortingType: model.SortDESC}, sortedDesc1},
		{"Test2", args{interactions: sortedEqual1, sortingType: model.SortASC}, sortedEqual1},
		{"Test2", args{interactions: sortedEqual1, sortingType: model.SortDESC}, sortedEqual1},
		{"Test2", args{interactions: sortedSomeEqual1, sortingType: model.SortASC}, sortedSomeEqual1},
		{"Test2", args{interactions: sortedSomeEqual1, sortingType: model.SortDESC}, sortedDescSomeEqual1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := model.SortInteractionTime(tt.args.interactions, tt.args.sortingType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortInteractionTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterRows(t *testing.T) {
	type args struct {
		rows           [][]interface{}
		attributionKey string
		keyIndex       int
	}
	rowIn1 := [][]interface{}{[]interface{}{"adwords", "camp1", "adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}, []interface{}{"adwords", "camp1", "adgroup1", "matchType1", "keyword2", 1, 2, 3, 4, 5, 65}}
	rowOut1 := [][]interface{}{[]interface{}{"adwords", "camp1", "adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}, []interface{}{"adwords", "camp1", "adgroup1", "matchType1", "keyword2", 1, 2, 3, 4, 5, 65}}

	rowIn2 := [][]interface{}{[]interface{}{"adwords", "camp1", "adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}, []interface{}{"adwords", "camp1", "adgroup1", "matchType1", "$none", 1, 2, 3, 4, 5, 65}}
	rowOut2 := [][]interface{}{[]interface{}{"adwords", "camp1", "adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}}

	rowIn3 := [][]interface{}{[]interface{}{"adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}, []interface{}{"adgroup1", "matchType1", "keyword2", 1, 2, 3, 4, 5, 65}}
	rowOut3 := [][]interface{}{[]interface{}{"adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}, []interface{}{"adgroup1", "matchType1", "keyword2", 1, 2, 3, 4, 5, 65}}
	rowIn4 := [][]interface{}{[]interface{}{"adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}, []interface{}{"adgroup1", "matchType1", "$none", 1, 2, 3, 4, 5, 65}}
	rowOut4 := [][]interface{}{[]interface{}{"adgroup1", "matchType1", "keyword1", 1, 2, 3, 4, 5, 65}}

	rowIn5 := [][]interface{}{[]interface{}{"adwords", "camp1", "adgroup1", 1, 2, 3, 4, 5, 65}, []interface{}{"adwords", "camp1", "adgroup1", 1, 2, 3, 4, 5, 65}}
	rowOut5 := [][]interface{}{[]interface{}{"adwords", "camp1", "adgroup1", 1, 2, 3, 4, 5, 65}, []interface{}{"adwords", "camp1", "adgroup1", 1, 2, 3, 4, 5, 65}}

	tests := []struct {
		name string
		args args
		want [][]interface{}
	}{
		{"Test1", args{rows: rowIn1, attributionKey: model.AttributionKeyKeyword, keyIndex: 4}, rowOut1},
		{"Test2", args{rows: rowIn2, attributionKey: model.AttributionKeyKeyword, keyIndex: 4}, rowOut2},
		{"Test3", args{rows: rowIn3, attributionKey: model.AttributionKeyKeyword, keyIndex: 2}, rowOut3},
		{"Test4", args{rows: rowIn4, attributionKey: model.AttributionKeyKeyword, keyIndex: 2}, rowOut4},
		{"Test5", args{rows: rowIn5, attributionKey: model.AttributionKeyAdgroup, keyIndex: 2}, rowOut5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := model.FilterRows(tt.args.rows, tt.args.attributionKey, tt.args.keyIndex); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterRows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddKPIKeyDataInMap(t *testing.T) {
	type args struct {
		kpiQueryResult     model.QueryResult
		logCtx             log.Entry
		keyIdx             int
		datetimeIdx        int
		from               int64
		to                 int64
		valIdx             int
		kpiValueHeaders    []string
		kpiAggFunctionType []string
		kpiData            *map[string]model.KPIInfo
	}

	kpiData := make(map[string]model.KPIInfo)

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"test_kpi_opp",
			args{
				model.QueryResult{
					Headers: []string{"datetime",
						"$salesforce_opportunity_id",
						"Total Pipeline",
						"Salesforce Opportunities",
						"Stage 4 Opps (Stage 4 Date)"},
					Rows: [][]interface{}{{
						"2022-01-14T10:00:00-08:00",
						"0063m00000opRhqAAE",
						9000,
						1,
						0,
					},
						{
							"2022-01-16T10:00:00-08:00",
							"0063m00000opRhqAAE",
							0,
							0,
							1,
						},
						{
							"2022-01-18T09:00:00-08:00",
							"0063m00000opWnnAAE",
							69999,
							1,
							0,
						}},
				},
				*log.WithFields(log.Fields{"KPIAttribution": "Debug"}),
				1,
				0,
				1641024000,
				1643702399,
				2,
				[]string{},
				[]string{},
				&kpiData,
			},
			[]string{"0063m00000opRhqAAE", "0063m00000opWnnAAE"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := model.AddKPIKeyDataInMap(tt.args.kpiQueryResult, tt.args.logCtx, tt.args.keyIdx, tt.args.datetimeIdx, tt.args.from, tt.args.to, tt.args.valIdx, tt.args.kpiValueHeaders, tt.args.kpiAggFunctionType, tt.args.kpiData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddKPIKeyDataInMap() = %v, want %v", got, tt.want)
			} else {
				val := (*tt.args.kpiData)["0063m00000opRhqAAE"]
				assert.Equal(t, 9000.0, val.KpiValues[0])
				assert.Equal(t, 1.0, val.KpiValues[1])
				assert.Equal(t, 1.0, val.KpiValues[2])

				val2 := (*tt.args.kpiData)["0063m00000opWnnAAE"]
				assert.Equal(t, 69999.0, val2.KpiValues[0])
				assert.Equal(t, 1.0, val2.KpiValues[1])
				assert.Equal(t, 0.0, val2.KpiValues[2])
				log.Info("found valid keys")
			}
		})
	}
}
