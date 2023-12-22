package tests

import (
	"encoding/json"
	C "factors/config"
	SDK "factors/sdk"
	TaskSession "factors/task/session"
	"factors/util"
	U "factors/util"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	"fmt"

	"github.com/gin-gonic/gin"
)

func TestAnalyticsFunnelQuery(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	t.Run("NoOfUsersCompletedTheFunnelFirstTimeOfStart", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 3; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID)

		occurrenceByIndex := []int{0, 1, 2}
		for index, eventIndex := range occurrenceByIndex {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], createdUserID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := model.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []model.QueryProperty{},
				},
			},
			Class: model.QueryClassFunnel,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		// steps headers avalilable.
		assert.Equal(t, model.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, model.StepPrefix+"1", result.Headers[1])
		// no.of users should be 1.
		assert.Equal(t, float64(1), result.Rows[0][0].(float64))
		assert.Equal(t, float64(1), result.Rows[0][1].(float64))
	})

	t.Run("NoOfUsersDidNotCompleteFunnelOnFirstTimeOfStart:1", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 3; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID)

		// user did only 0 first few times, did only 1 few times then 2.
		occurrenceByIndexUser1 := []int{0, 0, 0, 1, 1, 2}

		for index, eventIndex := range occurrenceByIndexUser1 {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], createdUserID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := model.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[2],
					Properties: []model.QueryProperty{},
				},
			},
			Class: model.QueryClassFunnel,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		assert.Equal(t, model.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, model.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result.Headers[2])
		assert.Equal(t, model.StepPrefix+"2", result.Headers[3])

		assert.Equal(t, float64(1), result.Rows[0][0], "step0")
		assert.Equal(t, float64(1), result.Rows[0][1], "step1")
		assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")
		assert.Equal(t, float64(1), result.Rows[0][3], "step3")
	})

	t.Run("NoOfUsersDidNotCompleteFunnelOnFirstTimeOfStart:2", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 4; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID)

		occurrenceByIndexUser1 := []int{0, 0, 0, 1, 1, 0, 2}
		for index, eventIndex := range occurrenceByIndexUser1 {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], createdUserID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := model.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[2],
					Properties: []model.QueryProperty{},
				},
			},
			Class: model.QueryClassFunnel,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		assert.Equal(t, model.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, model.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result.Headers[2])
		assert.Equal(t, model.StepPrefix+"2", result.Headers[3])

		assert.Equal(t, float64(1), result.Rows[0][0], "step0")
		assert.Equal(t, float64(1), result.Rows[0][1], "step1")
		assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")
		assert.Equal(t, float64(1), result.Rows[0][3], "step2")
	})

	t.Run("NoOfUsersDidNotCompleteFunnelOnFirstTimeOfStart:3", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 4; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID)

		occurrenceByIndexUser1 := []int{0, 0, 0, 1, 1, 0, 2, 1}
		for index, eventIndex := range occurrenceByIndexUser1 {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], createdUserID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := model.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[2],
					Properties: []model.QueryProperty{},
				},
			},
			Class: model.QueryClassFunnel,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		assert.Equal(t, model.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, model.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result.Headers[2])
		assert.Equal(t, model.StepPrefix+"2", result.Headers[3])

		assert.Equal(t, float64(1), result.Rows[0][0], "step0")
		assert.Equal(t, float64(1), result.Rows[0][1], "step1")
		assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")
		assert.Equal(t, float64(1), result.Rows[0][3], "step2")
	})
}

func TestAnalyticsFunnelGroupUserQuery(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	t.Run("NoOfUsersCompletedTheFunnelFirstTimeOfStart", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		// create random eventNames
		eventNames := make([]string, 0)
		for i := 0; i < 2; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		// create normal users
		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID)
		createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID1)

		// create group with groupName = "$hubspot_company"
		groupName := model.GROUP_NAME_HUBSPOT_COMPANY
		timestamp := time.Now().AddDate(0, 0, 0).Unix() * 1000
		_, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)

		// create group user with random groupID
		groupID := U.RandomLowerAphaNumString(5)
		groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
			ProjectId: project.ID, JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceHubspot),
		}, groupName, groupID)
		assert.Equal(t, http.StatusCreated, status)

		// register a group event using groupUserID
		groupEventName := U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
			groupEventName, groupUserID, eventTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])

		// register user event using normal createdUsers
		occurrenceByIndex := []int{0, 1}
		createdUsers := []string{createdUserID, createdUserID1}
		for index, eventIndex := range occurrenceByIndex {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], createdUsers[eventIndex], eventTimestamp+int64(index+1))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		// associate normal users to group
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID, groupName, groupID, groupUserID, false)
		assert.Equal(t, http.StatusAccepted, status)
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID1, groupName, groupID, groupUserID, false)
		assert.Equal(t, http.StatusAccepted, status)

		// fire funnel query
		query := model.Query{
			From: eventTimestamp,
			To:   eventTimestamp + int64(3),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:       groupEventName,
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []model.QueryProperty{},
				},
			},
			Class: model.QueryClassFunnel,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		// steps headers and rows avalilable.
		assert.Equal(t, model.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, model.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, model.StepPrefix+"2", result.Headers[3])
		assert.Equal(t, float64(2), result.Rows[0][0].(float64))
		assert.Equal(t, float64(1), result.Rows[0][1].(float64))
		assert.Equal(t, float64(0), result.Rows[0][3].(float64))
	})
}

func TestAnalyticsUniqueUsersQueryWithGroupEvent(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	t.Run("UniqueUsersQueryWithCRMGroupEventsAndSDKWebEvents", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		// Create random eventNames
		eventNames := make([]string, 0)
		for i := 0; i < 2; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		// Create normal users
		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID)
		createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "usersx@example.com"})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID1)

		// Create group with groupName = "$hubspot_company"
		groupName := model.GROUP_NAME_HUBSPOT_COMPANY
		timestamp := time.Now().AddDate(0, 0, 0).Unix() * 1000
		_, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
		assert.Equal(t, http.StatusCreated, status)

		// Create group user with random groupID
		groupID := U.RandomLowerAphaNumString(5)
		groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
			ProjectId: project.ID, JoinTimestamp: timestamp, Source: model.GetRequestSourcePointer(model.UserSourceHubspot),
		}, groupName, groupID)
		assert.Equal(t, http.StatusCreated, status)

		// Register a group event using groupUserID
		groupEventName := U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
			groupEventName, groupUserID, eventTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])

		// Register user event using normal createdUsers.
		occurrenceByIndex := []int{0, 1}
		createdUsers := []string{createdUserID, createdUserID1}
		for index, eventIndex := range occurrenceByIndex {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], createdUsers[eventIndex], eventTimestamp+int64(index+1))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		// Associate normal users to group
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID, groupName, groupID, groupUserID, false)
		assert.Equal(t, http.StatusAccepted, status)
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID1, groupName, groupID, groupUserID, false)
		assert.Equal(t, http.StatusAccepted, status)

		// Non-group user with same customer_user_id
		createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "usersx@example.com"})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, createdUserID2)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
			eventNames[1], createdUserID2, eventTimestamp+10)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response = DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])

		// Unique users who performed all given events.
		query := model.Query{
			From: eventTimestamp,
			To:   eventTimestamp + 100,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:       groupEventName,
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []model.QueryProperty{},
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, model.AliasAggr, result.Headers[0])
		// Result explanation: Out of 3 users.
		// 2 users were part of a group and the group has performed an event.
		// 1 non-group user has performed another event.
		// One of the group user and the non-group user have same customer_user_id and
		// hence qualify as 1 user performed all both the events.
		assert.Equal(t, float64(1), result.Rows[0][0].(float64))
	})
}

func TestAnalyticsFunnelWithUserIdentification(t *testing.T) {
	// Test Funnel of 2 events done by 2 different factors users,
	// who has done 1 event each, but identified as 1 customer user.

	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	eventNames := make([]string, 0, 0)
	for i := 0; i < 6; i++ {
		eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
	}
	eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.
	trackURI := "/sdk/event/track"

	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)

	createdUserID4, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4)

	payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		eventNames[2], createdUserID3, eventTimestamp+100)
	w1 := ServePostRequestWithHeaders(r, trackURI, []byte(payload1), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w1.Code)
	response1 := DecodeJSONResponseToMap(w1.Body)
	assert.NotNil(t, response1["event_id"])

	payload2 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		eventNames[3], createdUserID4, eventTimestamp+200)
	w2 := ServePostRequestWithHeaders(r, trackURI, []byte(payload2), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w2.Code)
	response2 := DecodeJSONResponseToMap(w2.Body)
	assert.NotNil(t, response2["event_id"])

	// identify users with same customer_user_id.
	identifyURI := "/sdk/user/identify"
	customerUserId := U.RandomLowerAphaNumString(15)
	w := ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, createdUserID3)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, createdUserID4)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	query := model.Query{
		From: eventTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       eventNames[2],
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       eventNames[3],
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result1, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, result1)

	// steps headers avalilable.
	assert.Equal(t, model.StepPrefix+"0", result1.Headers[0])
	assert.Equal(t, model.StepPrefix+"1", result1.Headers[1])
	// no.of users should be 1 after identification.
	assert.Equal(t, float64(1), result1.Rows[0][0].(float64))
	assert.Equal(t, float64(1), result1.Rows[0][1].(float64))
}

func TestAnalyticsFunnelQueryWithFilterConditionNumericalProperty(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	timestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	startTimestamp := timestamp

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID)

	payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 5}, "user_properties": {"value": 5}}`,
		"s0", timestamp)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	timestamp = timestamp + 10
	payload1 = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 100}, "user_properties": {"value": "string"}}`,
		"s0", timestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	timestamp = timestamp + 10
	payload1 = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": "string"}, "user_properties": {"value": 200}}`,
		"s0", timestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	timestamp = timestamp + 10
	payload1 = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 1000}, "user_properties": {"value": 2000}}`,
		"s0", timestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	timestamp = timestamp + 10
	payload1 = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 1}, "user_properties": {"value": 2000}}`,
		"s0", timestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	query := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityEvent,
						Property: "value",
						Operator: "greaterThan",
						Value:    "50",
						Type:     U.PropertyTypeNumerical,
					},
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "value",
						Operator:  "greaterThan",
						Value:     "50",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "OR",
					},
				},
			},
		},
		Class:           model.QueryClassInsights,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Equal(t, "aggregate", result.Headers[0])
	assert.Equal(t, float64(4), result.Rows[0][0])
}

func TestInsightsAnalyticsQueryGroupingMultipleFilters(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID)

	payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"day": "Monday"}, "user_properties": {"hour": 5}}`,
		"s0", startTimestamp)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload1 = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"day": "Tuesday"}, "user_properties": {"day": "Monday", "hour": 10}}`,
		"s0", startTimestamp+10)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload1 = fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"day": "Wednesday"}, "user_properties": {"day": "Tuesday", "hour": 12}}`,
		"s0", startTimestamp+10)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload1),
		map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	query := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityEvent,
						Property: "day",
						Operator: "equal",
						Value:    "Monday",
						Type:     U.PropertyTypeCategorical,
					},
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "day",
						Operator:  "equal",
						Value:     "Tuesday",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "day",
						Operator:  "equal",
						Value:     "Monday",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "AND",
					},
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "day",
						Operator:  "equal",
						Value:     "Tuesday",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "hour",
						Operator:  "greaterThanOrEqual",
						Value:     "10",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
		},
		Class:           model.QueryClassInsights,
		Type:            model.QueryTypeEventsOccurrence,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Equal(t, "aggregate", result.Headers[0])
	assert.Equal(t, float64(1), result.Rows[0][0])

}

func TestAnalyticsFunnelQueryWithFilterCondition(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// s0 event property value with 5.
	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 5}, "user_properties": {"gender": "M"}}`,
			"s0", stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		assert.NotNil(t, response["user_id"])
		stepTimestamp = stepTimestamp + 10
	}

	// s0 event property value greater than 5.
	userIds := make([]interface{}, 0, 0)
	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 10}, "user_properties": {"gender": "F"}}`,
			"s0", stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		assert.NotNil(t, response["user_id"])

		userIds = append(userIds, response["user_id"])
		stepTimestamp = stepTimestamp + 10
	}

	// users with value 10 perform s1.
	for i := range userIds {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties": {"value": 10}, "user_properties": {"gender": "F"}}`,
			"s1", userIds[i], stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		stepTimestamp = stepTimestamp + 10
	}

	query := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "value",
						Operator:  "greaterThan",
						Value:     "5",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
					model.QueryProperty{
						Entity:   model.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, model.StepPrefix+"0", result.Headers[0])
	assert.Equal(t, model.StepPrefix+"1", result.Headers[1])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result.Headers[2])
	// all 5 users who performed s0 with value greater
	// 5 has performed s1.
	assert.Equal(t, float64(5), result.Rows[0][0], "step0")
	assert.Equal(t, float64(5), result.Rows[0][1], "step1")
	assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")

	query1 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "value",
						Operator:  "lesserThan",
						Value:     "11",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result1, errCode, _ := store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, model.StepPrefix+"0", result1.Headers[0])
	assert.Equal(t, model.StepPrefix+"1", result1.Headers[1])
	// among 10 users who performed s0 with value lesser
	// than 11, 5 users has performed s1.
	assert.Equal(t, float64(10), result1.Rows[0][0], "step0")
	assert.Equal(t, float64(5), result1.Rows[0][1], "step1")
	assert.Equal(t, "50.0", result1.Rows[0][2], "conversion_step_0_step_1")

	query2 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "value",
						Operator:  "equals",
						Value:     "10",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result2, errCode, _ := store.GetStore().Analyze(project.ID, query2, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, model.StepPrefix+"0", result2.Headers[0])
	assert.Equal(t, model.StepPrefix+"1", result2.Headers[1])
	// all users performed s0 with value=10 has performed s1.
	assert.Equal(t, float64(5), result2.Rows[0][0], "step0")
	assert.Equal(t, float64(5), result2.Rows[0][1], "step1")

	query3 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "value",
						Operator:  "equals",
						Value:     "10",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
			model.QueryEventWithProperties{
				Name: "s1",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
					},
				},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result3, errCode, _ := store.GetStore().Analyze(project.ID, query3, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, model.StepPrefix+"0", result3.Headers[0])
	assert.Equal(t, model.StepPrefix+"1", result3.Headers[1])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result3.Headers[2])
	assert.Equal(t, float64(5), result3.Rows[0][0], "step0")
	assert.Equal(t, float64(5), result3.Rows[0][1], "step1")
	assert.Equal(t, "100.0", result3.Rows[0][2], "conversion_step_0_step_1")
}

func TestAnalyticsFunnelQueryRepeatedEvents(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
			"s1", createdUserID1, stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		stepTimestamp = stepTimestamp + 10
	}

	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		"s1", createdUserID2, startTimestamp)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	query := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
	assert.Equal(t, "50.0", result.Rows[0][2])
	assert.Equal(t, "50.0", result.Rows[0][3])

	identifyURI := "/sdk/user/identify"
	customerUserId := U.RandomLowerAphaNumString(15)
	w = ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, createdUserID1)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, createdUserID2)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	query1 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result1, errCode, _ := store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Equal(t, float64(1), result1.Rows[0][0])
	assert.Equal(t, float64(1), result1.Rows[0][1])
	assert.Equal(t, "100.0", result1.Rows[0][2])
	assert.Equal(t, float64(1), result1.Rows[0][3])
	assert.Equal(t, "100.0", result1.Rows[0][4])
	assert.Equal(t, "100.0", result1.Rows[0][5])
}

func TestAnalyticsFunnelQueryCRMEventsWithSameTimestamp(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)

	// create 3 events with 2 users for the same timestamp
	// user1 : s1, user2 : s1,s2
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		"s1", createdUserID1, startTimestamp)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	payload2 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		"s1", createdUserID2, startTimestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload2), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload3 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		"s2", createdUserID2, startTimestamp)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload3), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	query := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// should result in 0 conversions for the same timestamp and same event name
	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(0), result.Rows[0][1])
	assert.Equal(t, "0.0", result.Rows[0][2])
	assert.Equal(t, "0.0", result.Rows[0][3])

	query1 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	// should have 1 conversion as events are different but the timestamp is same
	result1, errCode, _ := store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result1.Rows[0][0])
	assert.Equal(t, float64(1), result1.Rows[0][1])
	assert.Equal(t, "50.0", result1.Rows[0][2])
	assert.Equal(t, "50.0", result1.Rows[0][3])
}

func TestAnalyticsFunnelQueryWithFilterAndBreakDown(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	CustomerUserIds := make([]string, 0, 0)

	// s0 event property value with 5.
	for i := 0; i < 5; i++ {

		user_id := U.RandomLowerAphaNumString(5)
		cu_id := U.RandomString(4)

		payload1 := fmt.Sprintf(`{"user_id":"%s","event_name": "%s", "timestamp": %d, "event_properties": {"value": 5,"id": 1}, "user_properties": {"gender": "M", "age": 18}}`,
			user_id, "s0", stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		assert.NotNil(t, response["user_id"])
		stepTimestamp = stepTimestamp + 10

		CustomerUserIds = append(CustomerUserIds, cu_id)
		status, _ := SDK.Identify(project.ID, &SDK.IdentifyPayload{UserId: user_id, CustomerUserId: cu_id, RequestSource: model.UserSourceWeb}, true)
		assert.Equal(t, http.StatusOK, status)
	}

	// s0 event property value greater than 5.
	userIds := make([]interface{}, 0, 0)
	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 10,"id": 2}, "user_properties": {"gender": "F", "age": 20}}`,
			"s0", stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1),
			map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		assert.NotNil(t, response["user_id"])

		userIds = append(userIds, response["user_id"])
		stepTimestamp = stepTimestamp + 10
	}

	// users with value 10 perform s1.
	for i := range userIds {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties": {"value": 10, "id": 3}, "user_properties": {"gender": "F", "age":21}}`,
			"s1", userIds[i], stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		stepTimestamp = stepTimestamp + 10
	}

	// add session to created events.
	_, err = TaskSession.AddSession([]int64{project.ID}, startTimestamp-(60*30), 0, 0, 0, 1, 1)
	assert.Nil(t, err)

	//x1 -> x2
	// (breakdown by user_property u1)
	query := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "gender",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result.Headers[0])
	assert.Equal(t, model.StepPrefix+"0", result.Headers[1])
	assert.Equal(t, model.StepPrefix+"1", result.Headers[2])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result.Headers[3])

	var noGroupIndex, mIndex, fIndex int
	for index := range result.Rows {
		if result.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result.Rows[index][0] == "M" {
			mIndex = index
		} else if result.Rows[index][0] == "F" {
			fIndex = index
		}
	}
	assert.Equal(t, "$no_group", result.Rows[noGroupIndex][0])
	assert.Equal(t, float64(10), result.Rows[noGroupIndex][1])
	assert.Equal(t, float64(5), result.Rows[noGroupIndex][2])
	assert.Equal(t, "50.0", result.Rows[noGroupIndex][3])

	assert.Equal(t, "M", result.Rows[mIndex][0])
	assert.Equal(t, float64(5), result.Rows[mIndex][1])
	assert.Equal(t, float64(0), result.Rows[mIndex][2])
	assert.Equal(t, "0.0", result.Rows[mIndex][3])

	assert.Equal(t, "F", result.Rows[fIndex][0])
	assert.Equal(t, float64(5), result.Rows[fIndex][1])
	assert.Equal(t, float64(5), result.Rows[fIndex][2])
	assert.Equal(t, "100.0", result.Rows[fIndex][3])

	// 	x1 -> x2
	// (breakdown by event x1 event_property ep1)
	query1 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result1, errCode, _ := store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value", result1.Headers[0])
	assert.Equal(t, model.StepPrefix+"0", result1.Headers[1])
	assert.Equal(t, model.StepPrefix+"1", result1.Headers[2])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result1.Headers[3])

	assert.Equal(t, "$no_group", result1.Rows[0][0])
	assert.Equal(t, float64(10), result1.Rows[0][1])
	assert.Equal(t, float64(5), result1.Rows[0][2])
	assert.Equal(t, "50.0", result1.Rows[0][3])

	assert.Equal(t, "5", result1.Rows[1][0])
	assert.Equal(t, float64(5), result1.Rows[1][1])
	assert.Equal(t, float64(0), result1.Rows[1][2])
	assert.Equal(t, "0.0", result1.Rows[1][3])

	assert.Equal(t, "10", result1.Rows[2][0])
	assert.Equal(t, float64(5), result1.Rows[2][1])
	assert.Equal(t, float64(5), result1.Rows[2][2])
	assert.Equal(t, "100.0", result1.Rows[2][3])

	// 	x1 -> x2
	// (breakdown by event x1 event_property ep1) and (breakdown by event x2 event_property ep2)
	query2 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "id",
				EventName:      "s1",
				EventNameIndex: 2,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result2, errCode, _ := store.GetStore().Analyze(project.ID, query2, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value", result2.Headers[0])
	assert.Equal(t, "id", result2.Headers[1])
	assert.Equal(t, model.StepPrefix+"0", result2.Headers[2])
	assert.Equal(t, model.StepPrefix+"1", result2.Headers[3])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result2.Headers[4])

	var fiveIndex, tenIndex int
	for index := range result2.Rows {
		if result2.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result2.Rows[index][0] == "5" {
			fiveIndex = index
		} else if result2.Rows[index][0] == "10" {
			tenIndex = index
		}
	}
	assert.Equal(t, "$no_group", result2.Rows[noGroupIndex][0])
	assert.Equal(t, "$no_group", result2.Rows[noGroupIndex][1])
	assert.Equal(t, float64(10), result2.Rows[noGroupIndex][2])
	assert.Equal(t, float64(5), result2.Rows[noGroupIndex][3])
	assert.Equal(t, "50.0", result2.Rows[noGroupIndex][4])
	assert.Equal(t, 3, len(result2.Rows))

	assert.Equal(t, "5", result2.Rows[fiveIndex][0])
	assert.Equal(t, "$none", result2.Rows[fiveIndex][1])
	assert.Equal(t, float64(5), result2.Rows[fiveIndex][2])
	assert.Equal(t, float64(0), result2.Rows[fiveIndex][3])
	assert.Equal(t, "0.0", result2.Rows[fiveIndex][4])

	assert.Equal(t, "10", result2.Rows[tenIndex][0])
	assert.Equal(t, "3", result2.Rows[tenIndex][1])
	assert.Equal(t, float64(5), result2.Rows[tenIndex][2])
	assert.Equal(t, float64(5), result2.Rows[tenIndex][3])
	assert.Equal(t, "100.0", result2.Rows[tenIndex][4])

	// x1 -> x2
	// (breakdown by user_property up1) and (breakdown by event x1 event_property ep1)
	query3 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "gender",
				EventName: model.UserPropertyGroupByPresent,
			},
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result3, errCode, _ := store.GetStore().Analyze(project.ID, query3, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result3.Headers[0])
	assert.Equal(t, "value", result3.Headers[1])
	assert.Equal(t, model.StepPrefix+"0", result3.Headers[2])
	assert.Equal(t, model.StepPrefix+"1", result3.Headers[3])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result3.Headers[4])

	for index := range result3.Rows {
		if result3.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result3.Rows[index][0] == "M" {
			mIndex = index
		} else if result3.Rows[index][0] == "F" {
			fIndex = index
		}
	}
	assert.Equal(t, 3, len(result3.Rows))
	assert.Equal(t, "$no_group", result3.Rows[noGroupIndex][0])
	assert.Equal(t, "$no_group", result3.Rows[noGroupIndex][1])
	assert.Equal(t, float64(10), result3.Rows[noGroupIndex][2])
	assert.Equal(t, float64(5), result3.Rows[noGroupIndex][3])
	assert.Equal(t, "50.0", result3.Rows[noGroupIndex][4])

	assert.Equal(t, "M", result3.Rows[mIndex][0])
	assert.Equal(t, "5", result3.Rows[mIndex][1])
	assert.Equal(t, float64(5), result3.Rows[mIndex][2])
	assert.Equal(t, float64(0), result3.Rows[mIndex][3])
	assert.Equal(t, "0.0", result3.Rows[mIndex][4])

	assert.Equal(t, "F", result3.Rows[fIndex][0])
	assert.Equal(t, "10", result3.Rows[fIndex][1])
	assert.Equal(t, float64(5), result3.Rows[fIndex][2])
	assert.Equal(t, float64(5), result3.Rows[fIndex][3])
	assert.Equal(t, "100.0", result3.Rows[fIndex][4])

	// 	x1 (with event_property ep1 = ev1) -> x2
	// (breakdown by event x1 event_property ep1) and (breakdown by event x2 event_property ep2)
	query4 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityEvent,
						Property: "value",
						Operator: "equals",
						Value:    "10",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "id",
				EventName:      "s1",
				EventNameIndex: 2,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result4, errCode, _ := store.GetStore().Analyze(project.ID, query4, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value", result4.Headers[0])
	assert.Equal(t, "id", result4.Headers[1])
	assert.Equal(t, model.StepPrefix+"0", result4.Headers[2])
	assert.Equal(t, model.StepPrefix+"1", result4.Headers[3])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result4.Headers[4])

	for index := range result4.Rows {
		if result4.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result4.Rows[index][0] == "10" {
			tenIndex = index
		}
	}
	assert.Equal(t, 2, len(result4.Rows))
	assert.Equal(t, "$no_group", result4.Rows[noGroupIndex][0])
	assert.Equal(t, "$no_group", result4.Rows[noGroupIndex][1])
	assert.Equal(t, float64(5), result4.Rows[noGroupIndex][2])
	assert.Equal(t, float64(5), result4.Rows[noGroupIndex][3])
	assert.Equal(t, "100.0", result4.Rows[noGroupIndex][4])

	assert.Equal(t, "10", result4.Rows[tenIndex][0])
	assert.Equal(t, "3", result4.Rows[tenIndex][1])
	assert.Equal(t, float64(5), result4.Rows[tenIndex][2])
	assert.Equal(t, float64(5), result4.Rows[tenIndex][3])
	assert.Equal(t, "100.0", result4.Rows[tenIndex][4])

	// x1 (with event_property ep1 = ev1) -> x2
	// (breakdown by user_property up1) and (breakdown by user_property up2)
	query5 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityEvent,
						Property: "value",
						Operator: "equals",
						Value:    "10",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "gender",
				EventName: model.UserPropertyGroupByPresent,
			},
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "age",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result5, errCode, _ := store.GetStore().Analyze(project.ID, query5, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result5.Headers[0])
	assert.Equal(t, "age", result5.Headers[1])
	assert.Equal(t, model.StepPrefix+"0", result5.Headers[2])
	assert.Equal(t, model.StepPrefix+"1", result5.Headers[3])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result5.Headers[4])

	for index := range result5.Rows {
		if result5.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result5.Rows[index][0] == "F" {
			fIndex = index
		}
	}
	assert.Equal(t, 2, len(result5.Rows))
	assert.Equal(t, "$no_group", result5.Rows[noGroupIndex][0])
	assert.Equal(t, "$no_group", result5.Rows[noGroupIndex][1])
	assert.Equal(t, float64(5), result5.Rows[noGroupIndex][2])
	assert.Equal(t, float64(5), result5.Rows[noGroupIndex][3])
	assert.Equal(t, "100.0", result5.Rows[noGroupIndex][4])

	assert.Equal(t, "F", result5.Rows[fIndex][0])
	assert.Equal(t, "20", result5.Rows[fIndex][1])
	assert.Equal(t, float64(5), result5.Rows[fIndex][2])
	assert.Equal(t, float64(5), result5.Rows[fIndex][3])
	assert.Equal(t, "100.0", result5.Rows[fIndex][4])

	// 	x1 (user_property up1 = uv1) -> x2
	// (breakdown by user_property up1) and (breakdown by user_property up2)
	query6 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "gender",
				EventName: model.UserPropertyGroupByPresent,
			},
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "age",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result6, errCode, _ := store.GetStore().Analyze(project.ID, query6, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result6.Headers[0])
	assert.Equal(t, "age", result6.Headers[1])
	assert.Equal(t, model.StepPrefix+"0", result6.Headers[2])
	assert.Equal(t, model.StepPrefix+"1", result6.Headers[3])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result6.Headers[4])

	for index := range result6.Rows {
		if result6.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result6.Rows[index][0] == "F" {
			fIndex = index
		}
	}
	assert.Equal(t, 2, len(result6.Rows))
	assert.Equal(t, "$no_group", result6.Rows[noGroupIndex][0])
	assert.Equal(t, "$no_group", result6.Rows[noGroupIndex][1])
	assert.Equal(t, float64(5), result6.Rows[noGroupIndex][2])
	assert.Equal(t, float64(5), result6.Rows[noGroupIndex][3])
	assert.Equal(t, "100.0", result6.Rows[noGroupIndex][4])

	assert.Equal(t, "F", result6.Rows[fIndex][0])
	assert.Equal(t, "20", result6.Rows[fIndex][1])
	assert.Equal(t, float64(5), result6.Rows[fIndex][2])
	assert.Equal(t, float64(5), result6.Rows[fIndex][3])
	assert.Equal(t, "100.0", result6.Rows[fIndex][4])

	// 	x1 (user_property up1 = uv1) -> x2
	// (breakdown by user_property up1) and (breakdown by event x1 event_property ep1)
	query7 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityEvent,
						Property: "value",
						Operator: "equals",
						Value:    "10",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "gender",
				EventName: model.UserPropertyGroupByPresent,
			},
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result7, errCode, _ := store.GetStore().Analyze(project.ID, query7, C.EnableOptimisedFilterOnEventUserQuery(), true)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result7.Headers[0])
	assert.Equal(t, "value", result7.Headers[1])
	assert.Equal(t, model.StepPrefix+"0", result7.Headers[2])
	assert.Equal(t, model.StepPrefix+"1", result7.Headers[3])
	assert.Equal(t, model.FunnelConversionPrefix+model.StepPrefix+"0"+"_"+model.StepPrefix+"1", result7.Headers[4])

	for index := range result7.Rows {
		if result7.Rows[index][0] == "$no_group" {
			noGroupIndex = index
		} else if result7.Rows[index][0] == "F" {
			fIndex = index
		}
	}
	assert.Equal(t, 2, len(result7.Rows))
	assert.Equal(t, "$no_group", result7.Rows[noGroupIndex][0])
	assert.Equal(t, "$no_group", result7.Rows[noGroupIndex][1])
	assert.Equal(t, float64(5), result7.Rows[noGroupIndex][2])
	assert.Equal(t, float64(5), result7.Rows[noGroupIndex][3])
	assert.Equal(t, "100.0", result7.Rows[noGroupIndex][4])

	assert.Equal(t, "F", result7.Rows[fIndex][0])
	assert.Equal(t, "10", result7.Rows[fIndex][1])
	assert.Equal(t, float64(5), result7.Rows[fIndex][2])
	assert.Equal(t, float64(5), result7.Rows[fIndex][3])
	assert.Equal(t, "100.0", result7.Rows[fIndex][4])

	query8 := model.Query{
		From: startTimestamp - 1, // session created before timestamp of first event.
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "$session",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:             model.QueryClassFunnel,
		Type:              model.QueryTypeUniqueUsers,
		EventsCondition:   model.EventCondAllGivenEvent,
		SessionStartEvent: 1,
		SessionEndEvent:   2,
	}

	result8, errCode, _ := store.GetStore().Analyze(project.ID, query8, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(10), result8.Rows[0][0])
	assert.Equal(t, float64(5), result8.Rows[0][1])

	// Test for event filter on user property and group by user property at the same event.
	query9 := model.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "s0",
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:   model.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityUser,
				Property:       "age",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result9, errCode, _ := store.GetStore().Analyze(project.ID, query9, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, result9)

	query10 := model.Query{
		From: startTimestamp - 1, // session created before timestamp of first event.
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "$session",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
		},
		Class:             model.QueryClassFunnel,
		Type:              model.QueryTypeUniqueUsers,
		EventsCondition:   model.EventCondAllGivenEvent,
		SessionStartEvent: 2,
		SessionEndEvent:   3,
	}

	result10, errCode, _ := store.GetStore().Analyze(project.ID, query10, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(10), result10.Rows[0][0])
	assert.Equal(t, float64(5), result10.Rows[0][1])

	//column properties with event level and global level breakdown
	//breakdown with global level property  and column property
	query11 := model.Query{
		From: startTimestamp - 1, // session created before timestamp of first event.
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "$session",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{

			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  "gender",
				EventName: model.UserPropertyGroupByPresent,
			},

			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  U.IDENTIFIED_USER_ID,
				EventName: model.UserPropertyGroupByPresent,
			},
		},

		Class:             model.QueryClassFunnel,
		Type:              model.QueryTypeUniqueUsers,
		EventsCondition:   model.EventCondAllGivenEvent,
		SessionStartEvent: 2,
		SessionEndEvent:   3,
	}

	result11, errCode, _ := store.GetStore().Analyze(project.ID, query11, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 7, len(result11.Rows))
	assert.Equal(t, "F", result11.Rows[1][0])
	for i := 2; i < 7; i++ {
		assert.Equal(t, true, U.ContainsStringInArray(CustomerUserIds, fmt.Sprintf("%v", result11.Rows[i][1])))
		assert.Equal(t, "M", result11.Rows[i][0])
	}

	//break down with only column property
	query12 := model.Query{
		From: startTimestamp - 1, // session created before timestamp of first event.
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "$session",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{

			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  U.IDENTIFIED_USER_ID,
				EventName: model.UserPropertyGroupByPresent,
			},
		},

		Class:             model.QueryClassFunnel,
		Type:              model.QueryTypeUniqueUsers,
		EventsCondition:   model.EventCondAllGivenEvent,
		SessionStartEvent: 2,
		SessionEndEvent:   3,
	}

	result12, errCode, _ := store.GetStore().Analyze(project.ID, query12, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 7, len(result12.Rows))
	assert.Equal(t, float64(5), result12.Rows[1][1])

	for i := 2; i < 7; i++ {
		assert.Equal(t, true, U.ContainsStringInArray(CustomerUserIds, fmt.Sprintf("%v", result12.Rows[i][0])))
		assert.Equal(t, float64(1), result12.Rows[i][1])
	}

	//breakdown with event level and column property
	query13 := model.Query{
		From: startTimestamp - 1, // session created before timestamp of first event.
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "$session",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{

			model.QueryGroupByProperty{
				Entity:         model.PropertyEntityUser,
				Property:       "age",
				EventName:      "s0",
				EventNameIndex: 2,
			},

			model.QueryGroupByProperty{
				Entity:    model.PropertyEntityUser,
				Property:  U.IDENTIFIED_USER_ID,
				EventName: model.UserPropertyGroupByPresent,
			},
		},

		Class:             model.QueryClassFunnel,
		Type:              model.QueryTypeUniqueUsers,
		EventsCondition:   model.EventCondAllGivenEvent,
		SessionStartEvent: 2,
		SessionEndEvent:   3,
	}

	result13, errCode, _ := store.GetStore().Analyze(project.ID, query13, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 7, len(result13.Rows))
	assert.Equal(t, "20", result13.Rows[1][0])

	for i := 2; i < 7; i++ {
		assert.Equal(t, true, U.ContainsStringInArray(CustomerUserIds, fmt.Sprintf("%v", result13.Rows[i][1])))
		assert.Equal(t, "18", result13.Rows[i][0])
	}

}

func TestAnalyticsInsightsQuery(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	t.Run("OperatorsWithNumericalPropertiesOnWhere", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.Nil(t, err)

		var firstEvent *model.Event

		// 10 times: page_spent_time as 5
		for i := 0; i < 10; i++ {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {"$page_spent_time" : %d}}`, eventName.Name, user.ID, 5)
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
			if i == 0 {
				event, errCode := store.GetStore().GetEventById(project.ID, response["event_id"].(string), "")
				assert.Equal(t, http.StatusFound, errCode)
				assert.NotNil(t, event)
				firstEvent = event
			}
		}

		// 5 times: page_spent_time as 12.
		for i := 0; i < 5; i++ {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {"$page_spent_time" : %d}}`, eventName.Name, user.ID, 12)
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		// Query count of events: page_spent_time > 11
		query := model.Query{
			From: firstEvent.Timestamp - 10,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName.Name,
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  "$page_spent_time",
							Operator:  "greaterThan",
							Type:      "numerical",
							LogicalOp: "AND",
							Value:     "11",
						},
					},
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(5), result.Rows[0][0])

		// Query count of events: page_spent_time > 11
		query2 := model.Query{
			From: firstEvent.Timestamp - 10,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName.Name,
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  "$page_spent_time",
							Operator:  "greaterThan",
							Type:      "numerical",
							LogicalOp: "AND",
							Value:     "4",
						},
					},
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		result2, errCode, _ := store.GetStore().Analyze(project.ID, query2, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, "aggregate", result2.Headers[0])
		assert.Equal(t, float64(15), result2.Rows[0][0])

	})
}

func TestAnalyticsInsightsQueryForAliasName(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)

	/*
		user1 -> event www.factors.ai with property1 -> www.factors.ai with property2 -> www.factors.ai/pricing with propterty2
		user2 -> event www.factors.ai with property1 -> www.factors.ai/pricing with property1
		user3 -> event www.factors.ai with property2 -> www.factors.ai/pricing with property2
	*/

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai", createdUserID1, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai/pricing", createdUserID1, stepTimestamp+20, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai/pricing", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai", createdUserID3, stepTimestamp, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai/pricing", createdUserID3, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:      "www.factors.ai/pricing",
					AliasName: "a0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				model.QueryEventWithProperties{
					Name:      "www.factors.ai",
					AliasName: "a1",
				},
			},
			Class:            model.QueryClassInsights,
			GroupByTimestamp: model.GroupByTimestampDate,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		a0_Index := 0
		a1_Index := 0
		for index := range result.Headers {
			if result.Headers[index] == "a0" {
				a0_Index = index
			} else if result.Headers[index] == "a1" {
				a1_Index = index
			}
		}
		assert.Equal(t, float64(1), result.Rows[0][a0_Index])
		assert.Equal(t, float64(4), result.Rows[0][a1_Index])
	})

	// Test for the scenario with no alias_name
	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "www.factors.ai/pricing",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				model.QueryEventWithProperties{
					Name: "www.factors.ai",
				},
			},
			Class:            model.QueryClassInsights,
			GroupByTimestamp: model.GroupByTimestampDate,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		pricing_Index := 0
		ai_Index := 0
		for index := range result.Headers {
			if result.Headers[index] == "www.factors.ai/pricing" {
				pricing_Index = index
			} else if result.Headers[index] == "www.factors.ai" {
				ai_Index = index
			}
		}
		assert.Equal(t, float64(1), result.Rows[0][pricing_Index])
		assert.Equal(t, float64(4), result.Rows[0][ai_Index])
	})

	// Test for the scenario with alias_name having spaces in it
	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:      "www.factors.ai/pricing",
					AliasName: "a 0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				model.QueryEventWithProperties{
					Name:      "www.factors.ai",
					AliasName: "a 1",
				},
			},
			Class:            model.QueryClassInsights,
			GroupByTimestamp: model.GroupByTimestampDate,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		a0_Index := 0
		a1_Index := 0
		for index := range result.Headers {
			if result.Headers[index] == "a 0" {
				a0_Index = index
			} else if result.Headers[index] == "a 1" {
				a1_Index = index
			}
		}
		assert.Equal(t, float64(1), result.Rows[0][a0_Index])
		assert.Equal(t, float64(4), result.Rows[0][a1_Index])
	})

	// Test for verifying the counts by alias_name
	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:      "www.factors.ai/pricing",
					AliasName: "a0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				model.QueryEventWithProperties{
					Name:      "www.factors.ai",
					AliasName: "a1",
				},
			},
			Class:            model.QueryClassInsights,
			GroupByTimestamp: model.GroupByTimestampDate,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		a0_Index := 0
		a1_Index := 0
		for index := range result.Headers {
			if result.Headers[index] == "a0" {
				a0_Index = index
			} else if result.Headers[index] == "a1" {
				a1_Index = index
			}
		}
		assert.Equal(t, float64(1), result.Rows[0][a0_Index])
		assert.Equal(t, float64(4), result.Rows[0][a1_Index])
	})

	t.Run("AliasWithOnEventAndResultHavingNullEventName", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 2*86400,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:      "www.factors.ai/pricing",
					AliasName: "a0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
			},
			Class:            model.QueryClassInsights,
			GroupByTimestamp: model.GroupByTimestampDate,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		log.WithField("result", result).WithField("errCode", errCode).Warn("kark1")
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		a0_Index := 0
		for index := range result.Headers {
			if result.Headers[index] == "a0" {
				a0_Index = index
			}
		}
		assert.Equal(t, 2, len(result.Headers))
		assert.Equal(t, float64(1), result.Rows[0][a0_Index])
	})
}

func TestAnalyticsQueryWithAliasNameWithSomeNullResponses(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai/pricing", createdUserID1, stepTimestamp+20, "B", 4321)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai/pricing", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "www.factors.ai/pricing", createdUserID3, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 10000,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name:      "www.factors.ai/pricing",
					AliasName: "a0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
			},
			Class:            model.QueryClassInsights,
			GroupByTimestamp: model.GroupByTimestampHour,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		a0_Index := 0
		for index := range result.Headers {
			if result.Headers[index] == "a0" {
				a0_Index = index
			}
		}
		assert.Equal(t, float64(1), result.Rows[0][a0_Index])
	})
}

func TestAnalyticsInsightsQueryWithFilterAndBreakdown(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)

	/*
		user1 -> event s0 with property1 -> s0 with property2 -> s1 with propterty2
		user2 -> event s0 with property1 -> s1 with property1
		user3 -> event s0 with property2 -> s1 with property2
	*/

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID1, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s1", createdUserID1, stepTimestamp+20, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s1", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", createdUserID3, stepTimestamp, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s1", createdUserID3, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])

		//unique user count should return 2 for s0 to s1 with fliter property2
		query.EventsWithProperties[0].Properties[0].Value = "B"
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])

		query = model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityUser,
					Property: "$initial_source",
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//breakdown by user property should return property A with 1 count and property B with 2 count
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "$initial_source", result.Headers[0])
		assert.Equal(t, "aggregate", result.Headers[1])
		assert.Equal(t, "B", result.Rows[0][0])
		assert.Equal(t, float64(2), result.Rows[0][1])
		assert.Equal(t, "A", result.Rows[1][0])
		assert.Equal(t, float64(1), result.Rows[1][1])
	})
	t.Run("AnalyticsInsightsQueryUniqueUserWithEventPropertyFilterAndBreakdown", func(t *testing.T) {
		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  "$campaign_id",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "1234",
						},
					},
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])

		query.EventsWithProperties[0].Properties[0].Value = "4321"
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])

		query = model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:         model.PropertyEntityEvent,
					Property:       "$campaign_id",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, "$campaign_id", result.Headers[0])
		expectedKeys := []string{"1234", "4321"}
		actualKeys := []string{result.Rows[0][0].(string), result.Rows[1][0].(string)}
		sort.Strings(actualKeys)
		assert.Equal(t, expectedKeys, actualKeys)
		assert.Equal(t, float64(2), result.Rows[0][1])
		// Counting all occurrences instead of first. So for user1, both 4321 and 1234 will be counted.
		assert.Equal(t, float64(2), result.Rows[1][1])
	})

	t.Run("AnalyticsInsightsQueryEventOccurrenceWithCountEventOccurrences", func(t *testing.T) {
		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "B",
						},
					},
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		/*
			Event occurrence with user property should give 5
			user1 -> 		 -> s0 with property2 -> s1 with propterty2
			user2 -> 		 -> s1 with property1
			user3 -> event s0 with property2 -> s1 with property2
		*/
		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "aggregate", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, float64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, float64(3), result.Rows[1][1])

		query.GroupByProperties = []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:   model.PropertyEntityUser,
				Property: "$initial_source",
			},
		}
		// property2 -> 4, property1 ->1
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "$initial_source", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, "B", result.Rows[0][1])
		assert.Equal(t, float64(2), result.Rows[0][2])

		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, "B", result.Rows[1][1])
		assert.Equal(t, float64(2), result.Rows[1][2])

		assert.Equal(t, "s1", result.Rows[2][0])
		assert.Equal(t, "A", result.Rows[2][1])
		assert.Equal(t, float64(1), result.Rows[2][2])

		//Count should be same as when done with user property = 5
		query.EventsWithProperties[0].Properties[0].Entity = model.PropertyEntityEvent
		query.EventsWithProperties[0].Properties[0].Property = "$campaign_id"
		query.EventsWithProperties[0].Properties[0].Value = "1234"
		query.GroupByProperties = []model.QueryGroupByProperty{}
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "aggregate", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, float64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, float64(3), result.Rows[1][1])
	})

	// Test for event filter on user property and group by user property at the same event.
	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndUserBreakdown", func(t *testing.T) {
		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:         model.PropertyEntityUser,
					Property:       "$initial_source",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
	})

	t.Run("AnalyticsInsightsQueryWithQuestionMark", func(t *testing.T) {
		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0?",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  "$campaign_id",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "1234",
						},
					},
				},
				model.QueryEventWithProperties{
					Name: "s1?",
				},
			},
			Class: model.QueryClassInsights,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}
		_, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), false)
		assert.Equal(t, http.StatusOK, errCode)

		var query2 model.Query
		U.DeepCopy(&query, &query2)
		query2.Type = model.QueryTypeEventsOccurrence
		_, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), false)
		assert.Equal(t, http.StatusOK, errCode)

		var query3 model.Query
		U.DeepCopy(&query, &query3)
		query3.GroupByProperties = []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:   model.PropertyEntityUser,
				Property: "$initial_source",
			},
		}
		_, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), false)
		assert.Equal(t, http.StatusOK, errCode)

		var query4 model.Query
		U.DeepCopy(&query2, &query4)
		query4.Type = model.QueryTypeEventsOccurrence
		_, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), false)
		assert.Equal(t, http.StatusOK, errCode)

	})

}

func TestAnalyticsInsightsQueryWithNumericalBucketing(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	t.Run("EventOccurrenceSingleBreakdown", func(t *testing.T) {
		// 100 events with single incremented value.
		eventName1 := "event1"
		numPropertyRangeStart := 1
		numPropertyRangeEnd := 100
		for i := numPropertyRangeStart; i <= numPropertyRangeEnd; i++ {
			icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
				`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":%d}}`,
				eventName1, icreatedUserID, startTimestamp+10, i, i)
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Add "bad_number" string for numerical page_load_time and numerical_property.
		// Should get filteted out and existing tests should pass as is.
		icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"$page_load_time":"%s"},"user_properties":{"numerical_property":"%s"}}`,
			eventName1, icreatedUserID, startTimestamp+10, "bad_number", "bad_number")
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		query1 := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName1,
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityEvent,
					Property: "$page_load_time",
					Type:     U.PropertyTypeNumerical,
				},
			},
			Class:           model.QueryClassInsights,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		// Query with key GroupByType as with_buckets.
		query2 := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName1,
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:      model.PropertyEntityEvent,
					Property:    "$page_load_time",
					Type:        U.PropertyTypeNumerical,
					GroupByType: model.GroupByTypeWithBuckets,
				},
			},
			Class:           model.QueryClassInsights,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		// Query with key GroupByType as raw_values.
		query3 := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName1,
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:      model.PropertyEntityEvent,
					Property:    "$page_load_time",
					Type:        U.PropertyTypeNumerical,
					GroupByType: model.GroupByTypeRawValues,
				},
			},
			Class:           model.QueryClassInsights,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}
		result, errCode, _ := store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		validateNumericalBucketRanges(t, result, numPropertyRangeStart, numPropertyRangeEnd, 0)

		result, errCode, _ = store.GetStore().Analyze(project.ID, query2, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		validateNumericalBucketRanges(t, result, numPropertyRangeStart, numPropertyRangeEnd, 0)

		// Query 3 with raw values. Should have 100 rows for each $page_load_time value.
		result, errCode, _ = store.GetStore().Analyze(project.ID, query3, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, 101, len(result.Rows))

		/*
			New event with $page_load_time = 0
			total element 21

			User property numerical_property set as empty ($none).
			Will create 11 buckets. including 1 $none.
		*/
		icreatedUserID, _ = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":""}}`,
			eventName1, icreatedUserID, startTimestamp+10, 0)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		result, errCode, _ = store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)
		validateNumericalBucketRanges(t, result, 0, numPropertyRangeEnd, 0)

		result, errCode, _ = store.GetStore().Analyze(project.ID, query2, C.EnableOptimisedFilterOnEventUserQuery(), true)
		validateNumericalBucketRanges(t, result, 0, numPropertyRangeEnd, 0)

		// Using group by numerical property.
		query1.GroupByProperties[0].Entity = model.PropertyEntityUser
		query1.GroupByProperties[0].Property = "numerical_property"
		result, errCode, _ = store.GetStore().Analyze(project.ID, query1, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		validateNumericalBucketRanges(t, result, numPropertyRangeStart, numPropertyRangeEnd, 1)
	})
}

func TestAnalyticsFunnelQueryWithNumericalBucketing(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	t.Run("FunnelSingleBreakdown", func(t *testing.T) {
		// 20 events with single incremented value.
		eventName1 := "event1"
		eventName2 := "event2"
		numPropertyRangeStart := 1
		numPropertyRangeEnd := 100
		lowerPercentileValue := int(model.NumericalLowerBoundPercentile * float64(numPropertyRangeEnd))
		upperPercentileValue := int(model.NumericalUpperBoundPercentile * float64(numPropertyRangeEnd))
		// nonPercentileBucketRange := (upperPercentileValue - lowerPercentileValue) / (model.NumericalGroupByBuckets - 2)

		for i := numPropertyRangeStart; i <= numPropertyRangeEnd; i++ {
			cuid := RandomURL()
			icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
				`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":%d}}`,
				eventName1, icreatedUserID, startTimestamp+10, i, i)
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)

			// Event2 by 25 users with timestamp + 20 for funnel.
			if i%4 == 0 {
				payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d}`,
					eventName2, icreatedUserID, startTimestamp+20)
				w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
				assert.Equal(t, http.StatusOK, w.Code)
			}
		}

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName1,
				},
				model.QueryEventWithProperties{
					Name: eventName2,
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					EventName:      eventName1,
					EventNameIndex: 1,
					Entity:         model.PropertyEntityEvent,
					Property:       "$page_load_time",
					Type:           U.PropertyTypeNumerical,
				},
			},
			Class:           model.QueryClassFunnel,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		// Expected output should be 10 equal range buckets with 2 elements
		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		bucketStart := numPropertyRangeStart
		bucketEnd := lowerPercentileValue
		bucketRange := model.GetBucketRangeForStartAndEnd(bucketStart, bucketEnd)
		// First bucket range.
		assert.Equal(t, bucketRange, result.Rows[1][0])

		// Last bucket range.
		bucketStart = upperPercentileValue
		bucketEnd = numPropertyRangeEnd
		bucketRange = model.GetBucketRangeForStartAndEnd(bucketStart, bucketEnd)
		assert.Equal(t, bucketRange, result.Rows[10][0])

		// bucketStart = lowerPercentileValue + 1
		// for i := 2; i < 10; i++ {
		// 	if i == 9 {
		// 		bucketEnd = upperPercentileValue - 1
		// 	} else {
		// 		bucketEnd = int(bucketStart+nonPercentileBucketRange) - 1
		// 	}
		// 	bucketRange = model.GetBucketRangeForStartAndEnd(bucketStart, bucketEnd)
		// 	assert.Equal(t, bucketRange, result.Rows[i][0])

		// 	bucketStart = bucketEnd + 1
		// }
	})
}

func TestAnalyticsInsightsQueryWithDateTimeProperty(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	location, err := time.LoadLocation(string(U.TimeZoneStringIST))
	assert.Nil(t, err)
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	startTimestampString := time.Unix(U.GetBeginningOfDayTimestampIn(startTimestamp, U.TimeZoneStringIST), 0).
		In(location).Format(U.DATETIME_FORMAT_DB_WITH_TIMEZONE)
	startTimestampYesterday := U.UnixTimeBeforeDuration(time.Hour * 24)
	startTimestampStringYesterday := time.Unix(U.GetBeginningOfDayTimestampIn(startTimestampYesterday, U.TimeZoneStringIST), 0).
		In(location).Format(U.DATETIME_FORMAT_DB_WITH_TIMEZONE)
	t.Run("FunnelSingleBreakdown", func(t *testing.T) {
		// 20 events with single incremented value.
		eventName1 := "event1"
		icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"date_property":%d},"user_properties":{"date_property":%d}}`,
			eventName1, icreatedUserID, startTimestamp+10, startTimestamp, startTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"date_property":%d},"user_properties":{"date_property":%d}}`,
			eventName1, icreatedUserID, startTimestamp+10, startTimestampYesterday, startTimestampYesterday)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"date_property1":%d},"user_properties":{"date_property":%d}}`,
			eventName1, icreatedUserID, startTimestamp+10, startTimestampYesterday, startTimestampYesterday)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"date_property":%d},"user_properties":{"date_property":%d}}`,
			eventName1, icreatedUserID, startTimestamp+10, startTimestampYesterday, startTimestampYesterday)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"date_property1":%d},"user_properties":{"date_property":%d}}`,
			eventName1, icreatedUserID, startTimestamp+10, startTimestampYesterday, startTimestampYesterday)
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		query := model.Query{
			From:     startTimestamp - (24 * 60 * 60),
			To:       startTimestamp + 40,
			Timezone: string(U.TimeZoneStringIST),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName1,
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					EventName:      eventName1,
					EventNameIndex: 1,
					Entity:         model.PropertyEntityEvent,
					Property:       "date_property",
					Type:           U.PropertyTypeDateTime,
					Granularity:    U.DateTimeBreakdownDailyGranularity,
				},
			},
			Class:           model.QueryClassInsights,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		var noneIndex, timestmapIndex, timestmapYesterdayIndex int
		for index := range result.Rows {
			if result.Rows[index][0] == "$none" {
				noneIndex = index
			} else if result.Rows[index][0] == startTimestampString {
				timestmapIndex = index
			} else if result.Rows[index][0] == startTimestampStringYesterday {
				timestmapYesterdayIndex = index
			}
		}
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "$none", result.Rows[noneIndex][0])
		assert.Equal(t, startTimestampString, result.Rows[timestmapIndex][0])
		assert.Equal(t, startTimestampStringYesterday, result.Rows[timestmapYesterdayIndex][0])
		assert.Equal(t, float64(2), result.Rows[noneIndex][1])
		assert.Equal(t, float64(2), result.Rows[timestmapYesterdayIndex][1])
		assert.Equal(t, float64(1), result.Rows[timestmapIndex][1])
	})
}

func TestEventsFunnelChannelWebClassBaseQueryHashStringConsistency(t *testing.T) {
	var queriesStr = map[string]string{
		model.QueryClassInsights: `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:   `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		//model.QueryClassChannel:   `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
		model.QueryClassEvents: `{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2}],"gbt":"","tz":"Asia/Calcutta"}]}`,
		//model.QueryClassChannelV1: `{ "query_group":[{ "channel": "google_ads", "select_metrics": ["impressions"], "filters": [], "group_by": [], "gbt": "hour", "fr": 1585679400, "to": 1585765800 }], "cl": "channel_v1" }`,
		model.QueryClassWeb: `{"units":[{"unit_id":"194","query_name":"bounce_rate"},{"unit_id":"195","query_name":"unique_users"},{"unit_id":"196","query_name":"avg_session_duration"},{"unit_id":"197","query_name":"avg_pages_per_session"},{"unit_id":"200","query_name":"sessions"},{"unit_id":"201","query_name":"total_page_view"},{"unit_id":"199","query_name":"traffic_channel_report"},{"unit_id":"198","query_name":"top_pages_report"}],"custom_group_units":[],"from":1609612200,"to":1610044199}`,
	}
	for queryClass, queryString := range queriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		queryHashString, err := baseQuery.GetQueryCacheHashString()
		assert.Nil(t, err)
		for i := 0; i < 50; i++ {
			// Query hash should be consistent and same every time.
			tempQueryHashString, err := baseQuery.GetQueryCacheHashString()
			assert.Nil(t, err)
			assert.Equal(t, queryHashString, tempQueryHashString)
		}

		for rangeString, rangeFunction := range U.QueryDateRangePresets {
			from, to, errCode := rangeFunction(U.TimeZoneStringIST)
			assert.Nil(t, errCode)
			baseQuery.SetQueryDateRange(from, to)
			assertMsg := fmt.Sprintf("Failed for class:%s:range:%s", queryClass, rangeString)

			tempQueryHashString, err := baseQuery.GetQueryCacheHashString()
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, queryHashString, tempQueryHashString, assertMsg)
		}
	}
}

func TestAttributionClassBaseQueryHashStringConsistency(t *testing.T) {
	var queriesStr = map[string]string{
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
		model.QueryClassChannel:     `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
		model.QueryClassEvents:      `{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2}],"gbt":"","tz":"Asia/Calcutta"}]}`,
		model.QueryClassChannelV1:   `{ "query_group":[{ "channel": "google_ads", "select_metrics": ["impressions"], "filters": [], "group_by": [], "gbt": "hour", "fr": 1585679400, "to": 1585765800 }], "cl": "channel_v1" }`,
		model.QueryClassWeb:         `{"units":[{"unit_id":"194","query_name":"bounce_rate"},{"unit_id":"195","query_name":"unique_users"},{"unit_id":"196","query_name":"avg_session_duration"},{"unit_id":"197","query_name":"avg_pages_per_session"},{"unit_id":"200","query_name":"sessions"},{"unit_id":"201","query_name":"total_page_view"},{"unit_id":"199","query_name":"traffic_channel_report"},{"unit_id":"198","query_name":"top_pages_report"}],"custom_group_units":[],"from":1609612200,"to":1610044199}`,
	}
	for queryClass, queryString := range queriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		queryHashString, err := baseQuery.GetQueryCacheHashString()
		assert.Nil(t, err)
		for i := 0; i < 50; i++ {
			// Query hash should be consistent and same every time.
			tempQueryHashString, err := baseQuery.GetQueryCacheHashString()
			assert.Nil(t, err)
			assert.Equal(t, queryHashString, tempQueryHashString)
		}

		for rangeString, rangeFunction := range U.QueryDateRangePresets {
			from, to, errCode := rangeFunction(U.TimeZoneStringIST)
			assert.Nil(t, errCode)
			baseQuery.SetQueryDateRange(from, to)
			assertMsg := fmt.Sprintf("Failed for class:%s:range:%s", queryClass, rangeString)

			tempQueryHashString, err := baseQuery.GetQueryCacheHashString()
			assert.Nil(t, err, assertMsg)
			assert.Equal(t, queryHashString, tempQueryHashString, assertMsg)
		}
	}
}

func TestAnalyticsQueryCaching(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	_, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
		IntFacebookAdAccount:        customerAccountID,
	})

	errCode = store.GetStore().CreateWebAnalyticsDefaultDashboardWithUnits(project.ID, agent.UUID)
	assert.Equal(t, http.StatusCreated, errCode)

	var queriesStr = map[string]string{
		model.QueryClassInsights: `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:   `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		//model.QueryClassChannel:   `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
		model.QueryClassEvents: `{"query_group":[{"cl":"events","ty":"events_occurrence","fr":1612031400,"to":1612376999,"ewp":[{"na":"$hubspot_contact_created","pr":[]}],"gbt":"date","gbp":[{"pr":"$hubspot_contact_revenue_segment_fs_","en":"event","pty":"categorical","ena":"$hubspot_contact_created","eni":1},{"pr":"$hubspot_contact_revenue_segment_fs_","en":"user","pty":"categorical","ena":"$present"}],"ec":"each_given_event","tz":"Asia/Kolkata"},{"cl":"events","ty":"events_occurrence","fr":1612031400,"to":1612376999,"ewp":[{"na":"$hubspot_contact_created","pr":[]}],"gbt":"","gbp":[{"pr":"$hubspot_contact_revenue_segment_fs_","en":"event","pty":"categorical","ena":"$hubspot_contact_created","eni":1},{"pr":"$hubspot_contact_revenue_segment_fs_","en":"user","pty":"categorical","ena":"$present"}],"ec":"each_given_event","tz":"Asia/Kolkata"}]}`,
		//model.QueryClassChannelV1: `{"query_group":[{"channel":"facebook_ads","select_metrics":["clicks"],"group_by":[{"name":"ad_group","property":"name"}],"filters":[],"gbt":"date","fr":1611426600,"to":1612031399},{"channel":"facebook_ads","select_metrics":["clicks"],"group_by":[{"name":"ad_group","property":"name"}],"filters":[],"gbt":"","fr":1611426600,"to":1612031399}],"cl":"channel_v1"}`,
		model.QueryClassWeb: `{"units":[{"unit_id":"194","query_name":"bounce_rate"},{"unit_id":"195","query_name":"unique_users"},{"unit_id":"196","query_name":"avg_session_duration"},{"unit_id":"197","query_name":"avg_pages_per_session"},{"unit_id":"200","query_name":"sessions"},{"unit_id":"201","query_name":"total_page_view"},{"unit_id":"199","query_name":"traffic_channel_report"},{"unit_id":"198","query_name":"top_pages_report"}],"custom_group_units":[],"from":1609612200,"to":1610044199}`,
	}

	var waitGroup sync.WaitGroup
	for queryClass, queryString := range queriesStr {
		var dashboardID, unitID int64
		if queryClass == model.QueryClassWeb {
			dashboardID, _, _ = store.GetStore().GetWebAnalyticsQueriesFromDashboardUnits(project.ID)
		}
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		waitGroup.Add(1)
		go sendAnalyticsQueryFromRoutine(r, queryClass, project.ID, agent, dashboardID, unitID, baseQuery, false, false, 1, &waitGroup)

		// Another immediate query. Should return from cache after polling.
		time.Sleep(50 * time.Millisecond)
		w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboardID, unitID, "", baseQuery, false, false)
		assert.NotEmpty(t, w)
		assert.Equal(t, http.StatusOK, w.Code)
		if queryClass != model.QueryClassWeb {
			// For website analytics, it returns from Dashboard cache.
			assert.Equal(t, "true", w.HeaderMap.Get(model.QueryCacheResponseFromCacheHeader), queryClass+" "+w.Body.String())
		}
	}
}

func TestAttributionQueryCaching(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	_, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{ProjectId: project.ID, Name: "$session"})
	assert.Equal(t, http.StatusCreated, errCode)

	customerAccountID := U.RandomLowerAphaNumString(5)
	store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountID,
		IntFacebookAdAccount:        customerAccountID,
	})

	errCode = store.GetStore().CreateWebAnalyticsDefaultDashboardWithUnits(project.ID, agent.UUID)
	assert.Equal(t, http.StatusCreated, errCode)

	var queriesStr = map[string]string{
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
	}

	var waitGroup sync.WaitGroup
	for queryClass, queryString := range queriesStr {
		var dashboardID, unitID int64
		if queryClass == model.QueryClassWeb {
			dashboardID, _, _ = store.GetStore().GetWebAnalyticsQueriesFromDashboardUnits(project.ID)
		}
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		waitGroup.Add(1)
		go sendAnalyticsQueryFromRoutine(r, queryClass, project.ID, agent, dashboardID, unitID, baseQuery, false, false, 1, &waitGroup)

		// Another immediate query. Should return from cache after polling.
		time.Sleep(50 * time.Millisecond)
		w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, dashboardID, unitID, "", baseQuery, false, false)
		assert.NotEmpty(t, w)
		assert.Equal(t, http.StatusOK, w.Code)
		if queryClass != model.QueryClassWeb {
			// For website analytics, it returns from Dashboard cache.
			assert.Equal(t, "true", w.HeaderMap.Get(model.QueryCacheResponseFromCacheHeader), queryClass+" "+w.Body.String())
		}
	}
}

func TestAnalyticsQueryCachingFailedCondition(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var waitGroup sync.WaitGroup
	var badQueriesStr = map[string]string{
		// Bad query type for insights and funnel query.
		model.QueryClassInsights: `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrences", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:   `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_userss", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		//model.QueryClassChannel:  `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
	}

	for queryClass, queryString := range badQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		waitGroup.Add(1)
		go sendAnalyticsQueryFromRoutine(r, queryClass, project.ID, agent, 0, 0, baseQuery, false, false, 1, &waitGroup)

		// First query should will fail because of wrong query class. This query should return error after polling.
		time.Sleep(50 * time.Millisecond)
		w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, 0, 0, "", baseQuery, false, false)
		assert.NotEmpty(t, w)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		type errorObj struct {
			Err string `json:"error"`
		}
		var errData errorObj
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		decoder.Decode(&errData)

		assert.Equal(t, "Error Processing/Fetching data from Query cache", errData.Err)
	}
}

func TestAttributionQueryCachingFailedCondition(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var waitGroup sync.WaitGroup
	var badQueriesStr = map[string]string{
		// Attribution and channel query should fail as no customer account id is created for project in test.
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
	}

	for queryClass, queryString := range badQueriesStr {
		queryJSON := postgres.Jsonb{json.RawMessage(queryString)}
		baseQuery, err := model.DecodeQueryForClass(queryJSON, queryClass)
		assert.Nil(t, err)

		waitGroup.Add(1)
		go sendAnalyticsQueryFromRoutine(r, queryClass, project.ID, agent, 0, 0, baseQuery, false, false, 1, &waitGroup)

		// First query should will fail because of wrong query class. This query should return error after polling.
		time.Sleep(50 * time.Millisecond)
		w := sendAnalyticsQueryReq(r, queryClass, project.ID, agent, 0, 0, "", baseQuery, false, false)
		assert.NotEmpty(t, w)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		type errorObj struct {
			Err string `json:"error"`
		}
		var errData errorObj
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		decoder.Decode(&errData)

		assert.Equal(t, "Error Processing/Fetching data from Query cache", errData.Err)
	}
}

func TestNumericalBucketingRegex(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)

	for _, numericValue := range []float64{1, 1.2, 1.4678, -2, -2.86, 0} {
		eventName := U.RandomString(5)
		icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"numerical_property":%f}}`,
			eventName, icreatedUserID, startTimestamp+10, numericValue)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: eventName,
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:   model.PropertyEntityEvent,
					Property: "numerical_property",
					Type:     U.PropertyTypeNumerical,
				},
			},
			Class:           model.QueryClassInsights,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotEmpty(t, result)

		// Should have returned value with single row.
		expectedBucket, _ := U.FloatRoundOffWithPrecision(numericValue, 1)
		assert.Equal(t, fmt.Sprint(expectedBucket), result.Rows[0][1])
		assert.Equal(t, float64(1), result.Rows[0][2])
	}
}

func TestTransformQueryPlaceholdersForContext(t *testing.T) {
	sampleQueries := []string{
		"select * from users where id=?",
		"select * from users where id = ?",
		"SELECT COUNT(*) AS count FROM events  WHERE events.project_id=? AND timestamp\u003e=? AND timestamp\u003c=? AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id=? AND name=?) ORDER BY count DESC LIMIT 100000",
		"SELECT COUNT(*) AS count FROM events  WHERE events.project_id=? AND timestamp>=? AND timestamp<=? AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id=? AND name=?) ORDER BY count DESC LIMIT 100000",
	}
	expectedTransformedQueries := []string{
		"select * from users where id=$1",
		"select * from users where id = $1",
		"SELECT COUNT(*) AS count FROM events  WHERE events.project_id=$1 AND timestamp\u003e=$2 AND timestamp\u003c=$3 AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id=$4 AND name=$5) ORDER BY count DESC LIMIT 100000",
		"SELECT COUNT(*) AS count FROM events  WHERE events.project_id=$1 AND timestamp>=$2 AND timestamp<=$3 AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id=$4 AND name=$5) ORDER BY count DESC LIMIT 100000",
	}

	for i := range sampleQueries {
		transformedQuery := model.TransformQueryPlaceholdersForContext(sampleQueries[i])
		assert.Equal(t, expectedTransformedQueries[i], transformedQuery)
	}
}

func TestExpandArrayWithIndividualValues(t *testing.T) {
	query := "SELECT * FROM users WHERE id IN (?)"
	params := []interface{}{[]int{1, 2, 3, 4}}
	newQuery, newParams := model.ExpandArrayWithIndividualValues(query, params)
	assert.Equal(t, "SELECT * FROM users WHERE id IN (?, ?, ?, ?)", newQuery)
	assert.Equal(t, []interface{}{1, 2, 3, 4}, newParams)

	query = "SELECT * FROM users WHERE project_id = ? AND id IN (?)"
	params = []interface{}{10, []int{1, 2, 3, 4}}
	newQuery, newParams = model.ExpandArrayWithIndividualValues(query, params)
	assert.Equal(t, "SELECT * FROM users WHERE project_id = ? AND id IN (?, ?, ?, ?)", newQuery)
	assert.Equal(t, []interface{}{10, 1, 2, 3, 4}, newParams)

	query = "SELECT * FROM users WHERE project_id = ? AND id IN (?) AND properties_id IN (?)"
	params = []interface{}{10, []int{1, 2, 3, 4}, []string{"abc", "def"}}
	newQuery, newParams = model.ExpandArrayWithIndividualValues(query, params)
	assert.Equal(t, "SELECT * FROM users WHERE project_id = ? AND id IN (?, ?, ?, ?) AND properties_id IN (?, ?)", newQuery)
	assert.Equal(t, []interface{}{10, 1, 2, 3, 4, "abc", "def"}, newParams)
}

func sendAnalyticsQueryFromRoutine(r *gin.Engine, queryClass string, projectID int64, agent *model.Agent, dashboardID,
	unitID int64, baseQuery model.BaseQuery, refresh bool, withDashboardParams bool, queryWaitSeconds int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	sendAnalyticsQueryReqWithHeader(r, queryClass, projectID, agent, dashboardID, unitID, "",
		baseQuery, false, false, map[string]string{model.QueryCacheRequestSleepHeader: fmt.Sprint(queryWaitSeconds), model.QueryFunnelV2: "true"})
}

func validateNumericalBucketRanges(t *testing.T, result *model.QueryResult, numPropertyRangeStart,
	numPropertyRangeEnd, noneCount int) {
	lowerPercentileValue := int(model.NumericalLowerBoundPercentile * float64(numPropertyRangeEnd))
	upperPercentileValue := int(model.NumericalUpperBoundPercentile * float64(numPropertyRangeEnd))
	// nonPercentileBucketRange := (upperPercentileValue - lowerPercentileValue) / (model.NumericalGroupByBuckets - 2)

	bucketsIndexStart := 0
	bucketsIndexEnd := 9
	if noneCount > 0 {
		assert.Equal(t, model.PropertyValueNone, result.Rows[0][1]) // First bucket should be $none.
		assert.Equal(t, float64(noneCount), result.Rows[0][2])      // Count of $none should be 1.

		bucketsIndexStart = 1
		bucketsIndexEnd = 10
	}

	// Expected output should be:
	//     First bucket should be of the range start to lower bound percentile.
	//     Last bucket should be of the range upper bound percentile to range end.
	//     Others buckets to be divided in 8 equal ranges.
	bucketStart := numPropertyRangeStart
	bucketEnd := lowerPercentileValue
	bucketRange := model.GetBucketRangeForStartAndEnd(bucketStart, bucketEnd)
	countInBucket := bucketEnd - bucketStart + 1
	// First bucket range.
	assert.Equal(t, bucketRange, result.Rows[bucketsIndexStart][1])
	assert.Equal(t, float64(countInBucket), result.Rows[bucketsIndexStart][2])

	// bucketStart = lowerPercentileValue + 1
	// for i := bucketsIndexStart + 1; i < bucketsIndexEnd; i++ {
	// 	if i == bucketsIndexEnd-1 {
	// 		bucketEnd = upperPercentileValue - 1
	// 	} else {
	// 		bucketEnd = int(bucketStart+nonPercentileBucketRange) - 1
	// 	}
	// 	bucketRange = model.GetBucketRangeForStartAndEnd(bucketStart, bucketEnd)
	// 	countInBucket = bucketEnd - bucketStart + 1
	// 	assert.Equal(t, bucketRange, result.Rows[i][1])
	// 	assert.Equal(t, float64(countInBucket), result.Rows[i][2])

	// 	bucketStart = bucketEnd + 1
	// }
	// Last bucket range.
	bucketStart = upperPercentileValue
	bucketEnd = numPropertyRangeEnd
	bucketRange = model.GetBucketRangeForStartAndEnd(bucketStart, bucketEnd)
	countInBucket = bucketEnd - bucketStart + 1
	assert.Equal(t, bucketRange, result.Rows[bucketsIndexEnd][1])
	assert.Equal(t, float64(countInBucket), result.Rows[bucketsIndexEnd][2])
}

func TestAnalyticsFunnelAnyOrder(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	event1 := U.RandomString(5)
	event2 := U.RandomString(5)
	event3 := U.RandomString(5)
	event4 := U.RandomString(5)
	startTime := U.TimeNowZ().Add(-5 * time.Hour)
	user1OrderedEvents := map[string]int64{
		event1: startTime.Unix(),
		event2: startTime.Add(10 * time.Minute).Unix(),
		event3: startTime.Add(20 * time.Minute).Unix(),
		event4: startTime.Add(30 * time.Minute).Unix(),
	}

	user2OrderedEvents := map[string]int64{
		event2: startTime.Unix(),
		event4: startTime.Add(10 * time.Minute).Unix(),
		event3: startTime.Add(20 * time.Minute).Unix(),
		event1: startTime.Add(30 * time.Minute).Unix(),
	}

	user3OrderedEvents := map[string]int64{
		event1: startTime.Unix(),
		event2: startTime.Add(10 * time.Minute).Unix(),
	}

	user4OrderedEvents := map[string]int64{
		event3: startTime.Unix(),
		event4: startTime.Add(10 * time.Minute).Unix(),
	}

	user1ID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	num := 0
	for eventName, timestamp := range user1OrderedEvents {
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"numerical_property":%d}}`,
			eventName, user1ID, timestamp, num)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		num++
	}

	user2ID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	num = 0
	for eventName, timestamp := range user2OrderedEvents {
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"numerical_property":%d}}`,
			eventName, user2ID, timestamp, num)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		num++
	}

	user3ID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	num = 0
	for eventName, timestamp := range user3OrderedEvents {
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"numerical_property":%d}}`,
			eventName, user3ID, timestamp, num)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		num++
	}

	user4ID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	num = 0
	for eventName, timestamp := range user4OrderedEvents {
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"numerical_property":%d}}`,
			eventName, user4ID, timestamp, num)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		num++
	}

	// Funnel query 1 with order
	query := model.Query{
		From: startTime.Unix(),
		To:   startTime.Add(1 * time.Hour).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: event1,
			},
			{
				Name: event2,
			},
			{
				Name: event3,
			},
			{
				Name: event4,
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotEmpty(t, result)

	stepResult := make(map[string]interface{})
	for i := range result.Headers {
		header := U.GetPropertyValueAsString(result.Headers[i])
		if strings.HasPrefix(header, "step_") {
			stepResult[header] = result.Rows[0][i]
		}
	}
	assert.Equal(t, float64(3), stepResult["step_0"])
	assert.Equal(t, float64(2), stepResult["step_1"])
	assert.Equal(t, float64(1), stepResult["step_2"])
	assert.Equal(t, float64(1), stepResult["step_3"])

	// Funnel query 1 without order
	query.EventsCondition = model.EventCondFunnelAnyGivenEvent
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotEmpty(t, result)

	stepResult = make(map[string]interface{})
	for i := range result.Headers {
		header := U.GetPropertyValueAsString(result.Headers[i])
		if strings.HasPrefix(header, "step_") {
			stepResult[header] = result.Rows[0][i]
		}
	}
	assert.Equal(t, float64(3), stepResult["step_0"])
	assert.Equal(t, float64(3), stepResult["step_1"])
	assert.Equal(t, float64(2), stepResult["step_2"])
	assert.Equal(t, float64(2), stepResult["step_3"])

	//funnel query 2 with order
	query.EventsWithProperties = []model.QueryEventWithProperties{
		{
			Name: event4,
		},
		{
			Name: event3,
		},
		{
			Name: event2,
		},
		{
			Name: event1,
		},
	}
	query.EventsCondition = model.EventCondAnyGivenEvent

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotEmpty(t, result)

	stepResult = make(map[string]interface{})
	for i := range result.Headers {
		header := U.GetPropertyValueAsString(result.Headers[i])
		if strings.HasPrefix(header, "step_") {
			stepResult[header] = result.Rows[0][i]
		}
	}

	assert.Equal(t, float64(3), stepResult["step_0"])
	assert.Equal(t, float64(1), stepResult["step_1"])
	assert.Equal(t, float64(0), stepResult["step_2"])
	assert.Equal(t, float64(0), stepResult["step_3"])

	// Funnel query without order
	query.EventsCondition = model.EventCondFunnelAnyGivenEvent

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotEmpty(t, result)

	stepResult = make(map[string]interface{})
	for i := range result.Headers {
		header := U.GetPropertyValueAsString(result.Headers[i])
		if strings.HasPrefix(header, "step_") {
			stepResult[header] = result.Rows[0][i]
		}
	}
	assert.Equal(t, float64(3), stepResult["step_0"])
	assert.Equal(t, float64(3), stepResult["step_1"])
	assert.Equal(t, float64(2), stepResult["step_2"])
	assert.Equal(t, float64(2), stepResult["step_3"])
}

func TestAnalyticsFunnelGroupQuery(t *testing.T) {
	project, _, _, agent, err := SetupProjectUserEventNameAgentReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	_, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)

	/*
		group users
			group user 1 -> properties{"group1_id":group_1}
			group user 2 -> properties{"group1_id":group_2}
			group user 3 -> properties{"group1_id":group_3}
			group user 4 -> properties{"group1_id":group_4}
		hubspot contact users
			contact user1 -> customer_user_id("group_1"), group_1_user_id(group user 1)
			contact user2 -> customer_user_id("group_2"), group_1_user_id(group user 2)
			contact user3 -> customer_user_id("group_3"), group_1_user_id(group user 3)
			contact user4 -> customer_user_id(null), group_1_user_id(group user 4)

	*/
	group1ID1 := "group_1"
	group1ID2 := "group_2"
	group1ID3 := "group_3"
	group1ID4 := "group_4"

	nonIdentifiedContactUserByGroup := map[string]bool{
		group1ID4: true,
	}

	groupJoinTimestamp := time.Now().AddDate(0, 0, -1)
	groupUserIDMap := map[string]string{}
	hubspotContactUser := []string{}
	for _, group1ID := range []string{group1ID1, group1ID2, group1ID3, group1ID4} {
		groupProperties := &map[string]interface{}{
			"group1_id": group1ID,
		}
		properties, err := U.EncodeToPostgresJsonb(groupProperties)
		assert.Nil(t, err)

		// create group user
		groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
			ProjectId: project.ID, JoinTimestamp: groupJoinTimestamp.Unix(), Source: model.GetRequestSourcePointer(model.UserSourceHubspot), Properties: *properties,
		}, model.GROUP_NAME_HUBSPOT_COMPANY, group1ID)
		assert.Equal(t, http.StatusCreated, status)
		groupUserIDMap[group1ID] = groupUserID

		customerUserID := group1ID
		if nonIdentifiedContactUserByGroup[group1ID] {
			customerUserID = ""
		}
		// create hubspot contact user associated with group user
		userID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties, CustomerUserId: customerUserID,
			Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
		assert.Equal(t, http.StatusCreated, errCode)
		hubspotContactUser = append(hubspotContactUser, userID)
		_, status = store.GetStore().UpdateUserGroup(project.ID, userID, model.GROUP_NAME_HUBSPOT_COMPANY, group1ID, groupUserIDMap[group1ID], false)
		assert.Equal(t, http.StatusAccepted, status)
	}

	/*
		non group users
			user 1 -> properties{"global_user":1}, customer_user_id("group_1")
			user 2 -> properties{"global_user":2}, customer_user_id("group_2")
			user 3 -> properties{"global_user":3}, customer_user_id(null)
			user 4 -> properties{"global_user":4}, customer_user_id("group_3")
			user 5 -> properties{"global_user":1}, customer_user_id("group_1")
	*/

	propertiesMap := &map[string]interface{}{
		"global_user": "1",
	}
	properties, err := U.EncodeToPostgresJsonb(propertiesMap)
	assert.Nil(t, err)
	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties, CustomerUserId: group1ID1,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	propertiesMap = &map[string]interface{}{
		"global_user": "2",
	}
	properties, err = U.EncodeToPostgresJsonb(propertiesMap)
	assert.Nil(t, err)
	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties, CustomerUserId: group1ID2,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	propertiesMap = &map[string]interface{}{
		"global_user": "3",
	}
	properties, err = U.EncodeToPostgresJsonb(propertiesMap)
	assert.Nil(t, err)
	user3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	propertiesMap = &map[string]interface{}{
		"global_user": "4",
	}
	properties, err = U.EncodeToPostgresJsonb(propertiesMap)
	assert.Nil(t, err)
	user4, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties, CustomerUserId: group1ID3,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	propertiesMap = &map[string]interface{}{
		"global_user": "5",
	}
	properties, err = U.EncodeToPostgresJsonb(propertiesMap)
	assert.Nil(t, err)
	user5, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties, CustomerUserId: group1ID1,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	/*
		 Events
			group events
				hubspot company created -> group user 1, timesamp = 10x
				hubspot company created -> group user 2, timesamp = 10x
				hubspot company created -> group user 4, timesamp = 10x

			event1 -> non group user 1 -> event_properties{"$cost":1}, user_properies{"$city":"A"}, timestamp = x
			event1 -> non group user 2 -> event_properties{"$cost":3}, user_properies{"$city":"B"}, timestamp = 3x
			event1 -> non group user 3 -> event_properties{"$cost":6}, user_properies{"$city":"C"}, timestamp = 6x
			event1 -> non group user 4 -> event_properties{"$cost":1}, user_properies{"$city":"A"}, timestamp = 1x

			event2 -> non group user 1 -> event_properties{"$cost":2}, user_properies{"$city":"B"}, timestamp = 2x
			event2 -> non group user 2 -> event_properties{"$cost":7}, user_properies{"$city":"D"}, timestamp = 7x
			event2 -> non group user 3 -> event_properties{"$cost":7}, user_properies{"$city":"D"}, timestamp = 7x

			event3 -> non group user 3 -> event_properties{"$cost":2}, user_properies{"$city":"B"}, timstamp = 2x
			event3 -> non group user 5 -> event_properties{"$cost":4}, user_properies{"$city":"C"}, timstamp = 2x

			events by hubspot user
				hubspotContactEvent1 -> hubspot user 1 -> event_properties{"$hs_cost":1}, timestamp = x
				hubspotContactEvent1 -> hubspot user 2 -> event_properties{"$hs_cost":2}, timestamp = x
				hubspotContactEvent1 -> hubspot user 3 -> event_properties{"$hs_cost":3}, timestamp = x
				hubspotContactEvent1 -> hubspot user 4 -> event_properties{"$hs_cost":4}, timestamp = x

				hubspotContactEvent2 -> hubspot user 1 -> event_properties{"$hs_cost":5}, timestamp = 2x
				hubspotContactEvent2 -> hubspot user 2 -> event_properties{"$hs_cost":6}, timestamp = 2x
				hubspotContactEvent2 -> hubspot user 3 -> event_properties{"$hs_cost":7}, timestamp = 2x
				hubspotContactEvent2 -> hubspot user 4 -> event_properties{"$hs_cost":8}, timestamp = 2x

	*/

	eventTimestamp := groupJoinTimestamp
	// group events
	groupEventTimestamp := eventTimestamp.Add(10 * time.Hour).Unix()
	groupEventName := U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED
	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d}}`,
		groupEventName, groupUserIDMap[group1ID1], groupEventTimestamp, 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d}}`,
		groupEventName, groupUserIDMap[group1ID2], groupEventTimestamp, 1)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d}}`,
		groupEventName, groupUserIDMap[group1ID4], groupEventTimestamp, 1)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// user events
	// event1
	event1 := U.RandomString(4)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event1, user1, eventTimestamp.Unix(), 1, "A")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event1, user2, eventTimestamp.Add(3*time.Hour).Unix(), 3, "B")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event1, user3, eventTimestamp.Add(6*time.Hour).Unix(), 6, "C")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event1, user4, eventTimestamp.Add(1*time.Hour).Unix(), 1, "A")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// event2
	event2 := U.RandomString(4)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event2, user1, eventTimestamp.Add(2*time.Hour).Unix(), 2, "B")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event2, user2, eventTimestamp.Add(7*time.Hour).Unix(), 7, "D")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event2, user3, eventTimestamp.Add(7*time.Hour).Unix(), 7, "D")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// event3
	event3 := U.RandomString(4)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event3, user3, eventTimestamp.Add(2*time.Hour).Unix(), 2, "B")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$cost":%d},"user_properties":{"$city":"%s"}}`,
		event3, user5, eventTimestamp.Add(2*time.Hour).Unix(), 4, "C")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	/*
		hubspot contact events
	*/
	contactEvent1 := "contactEvent1"
	contactEvent2 := "contactEvent2"
	for id, userID := range hubspotContactUser {
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$hs_cost":%d},"user_properties":{"$city":"%s"}}`,
			contactEvent1, userID, eventTimestamp.Add(1*time.Hour).Unix(), id, "B")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$hs_cost":%d},"user_properties":{"$city":"%s"}}`,
			contactEvent2, userID, eventTimestamp.Add(2*time.Hour).Unix(), len(hubspotContactUser)+id, "C")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
	}

	/*
		event1 to group event
			user1(group_1_1), user2(group_2), user4(group_3) -> user1(group_1_1), user2(group_2)
	*/
	query := model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])

	/*
		event1 to group event where group events user properties "group1_id" equals "group_1"
			user1(group_1_1), user2(group_1_2), user4(group_1_3) -> user1(group_1_1)
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name: groupEventName,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "group1_id",
						Operator:  "equals",
						Value:     group1ID1,
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
				},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	/*
		negative check
		event1 to group event where group events user properties "group1_id" equals "10"
			user1(group_1_1), user2(group_1_2), user4(group_1_3) ->  none
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name: groupEventName,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "group1_id",
						Operator:  "equals",
						Value:     "group10", // no exist
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
				},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(0), result.Rows[0][1])

	/*
		event1 to group event where event1 "$cost" equals "1"
			user1(group_1_1), user4(group_1_3) -> user1(group_1_1)
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: event1,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "$cost",
						Operator:  "equals",
						Value:     "1",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	/*
		event1 to group event where event1 user properties "$city" equals "A"
			user1(group_1_1), user4(group_1_3) -> user1(group_1_1)
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: event1,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "$city",
						Operator:  "equals",
						Value:     "A",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	/*
		event1 to group event where event1 "$cost" equals "3"
			user2(group_1_2) -> user2(group_1_2)
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: event1,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "$cost",
						Operator:  "equals",
						Value:     "3",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	/*
		event1 to group event where event1 "$cost" equals "6"
			none -> none, user3 is not group user
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: event1,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityEvent,
						Property:  "$cost",
						Operator:  "equals",
						Value:     "6",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "OR",
					},
				},
			},
			model.QueryEventWithProperties{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(0), result.Rows[0][0])
	assert.Equal(t, float64(0), result.Rows[0][1])

	/*
		event1 to event2 where event1
			user1(group_1_1), user2(group_1_2), user4(group_1_3) -> user1(group_1_1), user2(group_1_2)
			user3 didn't belong to group
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       event2,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])

	// with headers for funnel group query
	w = sendAnalyticsQueryReqWithHeader(r, model.QueryClassFunnel, project.ID, agent, 0, 0, "", &query, false, false,
		map[string]string{model.QueryFunnelV2: "true"})
	assert.Equal(t, http.StatusOK, w.Code)
	result = DecodeJSONResponseToAnalyticsResult(w.Body)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])

	// without headers also group result
	w = sendAnalyticsQueryReqWithHeader(r, model.QueryClassFunnel, project.ID, agent, 0, 0, "", &query, true, false, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	result = DecodeJSONResponseToAnalyticsResult(w.Body)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])
	/*
		event1 to event2 where event1
			user1(group_1_1), user2(group_1_2), user3, user4(group_1_3) -> user5(group_1_1)
			user5 has same customer_user_id
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       event3,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	/*
		group event to event3
		group user 1, group user 2, group user 4 -> none
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       event3,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(0), result.Rows[0][1])

	/*
		event1 to event2
		user1(group_1_1), -> user1(group_1_1)
		global filter by
			group1_id = 'group_1'
		global group by 'group1_id' scope group 1
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       event2,
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{

			model.QueryProperty{
				Entity:    model.PropertyEntityUserGlobal,
				Property:  "group1_id",
				Operator:  "equals",
				Value:     "group_1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				EventName: model.UserPropertyGroupByPresent,
				Entity:    model.PropertyEntityUser,
				Property:  "group1_id",
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, "group_1", result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])

	/*
		event1 to event2
		user1(group_1_1), -> user1(group_1_1)
		global group by group1_id for group 1 scope

	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       event2,
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				EventName: model.UserPropertyGroupByPresent,
				Entity:    model.PropertyEntityUser,
				Property:  "group1_id",
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "group1_id", result.Headers[0])
	assert.Equal(t, "step_0", result.Headers[1])
	assert.Equal(t, "step_1", result.Headers[2])
	assert.Equal(t, "conversion_step_0_step_1", result.Headers[3])
	assert.Equal(t, "conversion_overall", result.Headers[4])
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, float64(3), result.Rows[0][1])
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, "group_1", result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, "group_2", result.Rows[2][0])
	assert.Equal(t, float64(1), result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
	assert.Equal(t, "group_3", result.Rows[3][0])
	assert.Equal(t, float64(1), result.Rows[3][1])
	assert.Equal(t, float64(0), result.Rows[3][2])

	/*
		event1 to event2
		user1(group_1_1), -> user1(group_1_1)
		global filter by
			group1_id = 'group_1'
		global group by
			'group1_id'
		event 1 group by
			'$cost'
		event 2 group by
			'$cost'
		event 1 group by
			'$cost1'
		event 1 group by
			'$city'
		global group by
			'$group_1_id' scope group 1
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       event1,
				Properties: []model.QueryProperty{},
			},

			{
				Name:       event2,
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				Property:  "group1_id",
				Operator:  "equals",
				Value:     "group_1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityEvent,
				EventName:      event1,
				EventNameIndex: 1,
				Property:       "$cost",
			},
			{
				Entity:         model.PropertyEntityEvent,
				EventName:      event2,
				EventNameIndex: 2,
				Property:       "$cost",
			},
			{
				Entity:         model.PropertyEntityEvent,
				EventName:      event1,
				EventNameIndex: 1,
				Property:       "$cost1",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventName:      event1,
				EventNameIndex: 1,
				Property:       "$city",
			},
			{
				EventName: model.UserPropertyGroupByPresent,
				Entity:    model.PropertyEntityUser,
				Property:  "group1_id",
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})

	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, "$no_group", result.Rows[0][1])
	assert.Equal(t, "$no_group", result.Rows[0][2])
	assert.Equal(t, "$no_group", result.Rows[0][3])
	assert.Equal(t, "$no_group", result.Rows[0][4])
	assert.Equal(t, float64(1), result.Rows[0][5])
	assert.Equal(t, float64(1), result.Rows[0][6])
	assert.Equal(t, "100.0", result.Rows[0][7])
	assert.Equal(t, "100.0", result.Rows[0][8])
	assert.Equal(t, "1", result.Rows[1][0])
	assert.Equal(t, "2", result.Rows[1][1])
	assert.Equal(t, "$none", result.Rows[1][2])
	assert.Equal(t, "A", result.Rows[1][3])
	assert.Equal(t, "group_1", result.Rows[1][4])
	assert.Equal(t, float64(1), result.Rows[1][5])
	assert.Equal(t, float64(1), result.Rows[1][6])
	assert.Equal(t, "100.0", result.Rows[1][7])
	assert.Equal(t, "100.0", result.Rows[1][8])

	/*
		hubspot contact event 1
		group user 1, group user 2, group user 4
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       contactEvent1,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(4), result.Rows[0][0])
	assert.Equal(t, "100.0", result.Rows[0][1])

	/*
		hubspot contact event 1 -> group event
		group user 1, group user 2, group user 4
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       contactEvent1,
				Properties: []model.QueryProperty{},
			},

			{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(4), result.Rows[0][0])
	assert.Equal(t, float64(3), result.Rows[0][1])

	/*
		group event -> hubspot contact event 1
		group user 1, group user 2, group user 4
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       groupEventName,
				Properties: []model.QueryProperty{},
			},

			{
				Name:       contactEvent1,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondFunnelAnyGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][0])
	assert.Equal(t, float64(3), result.Rows[0][1])

	/*
		hubspot contact event 1  -> hubspot contact event 2
		group user 1, group user 2, group user 4
	*/
	query = model.Query{
		From: eventTimestamp.Unix(),
		To:   eventTimestamp.AddDate(0, 0, 1).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       contactEvent1,
				Properties: []model.QueryProperty{},
			},

			{
				Name:       contactEvent2,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(4), result.Rows[0][0])
	assert.Equal(t, float64(4), result.Rows[0][1])
}

func TestAnalyticsFunnelBreakdownPropertyFirstOccurence(t *testing.T) {
	project, _, _, _, err := SetupProjectUserEventNameAgentReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID)
	startTime := time.Now().Add(-10 * time.Second).Unix()
	for i := 0; i < 5; i++ {
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":%d,"value2":%d},"user_properties":{"user_value1":%d,"user_value2":%d}}`,
			"s0", createdUserID, startTime+int64(2*i), i, i+1, i, i+2)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		assert.Nil(t, response["user_id"])
	}

	query := model.Query{
		From: startTime - 100,
		To:   startTime + 100,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityEvent,
				EventName:      "s0",
				EventNameIndex: 2,
				Property:       "value1",
			},
			{
				Entity:         model.PropertyEntityEvent,
				EventName:      "s0",
				EventNameIndex: 2,
				Property:       "value2",
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value1", result.Headers[0])
	assert.Equal(t, "value2", result.Headers[1])
	assert.Equal(t, "1", result.Rows[1][0])
	assert.Equal(t, "2", result.Rows[1][1])

	query = model.Query{
		From: startTime - 100,
		To:   startTime + 100,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 2,
				Property:       "user_value1",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 2,
				Property:       "user_value2",
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "user_value1", result.Headers[0])
	assert.Equal(t, "user_value2", result.Headers[1])
	assert.Equal(t, "1", result.Rows[1][0])
	assert.Equal(t, "3", result.Rows[1][1])

}

func TestAnalyticsFunnelConversionXTime(t *testing.T) {
	project, _, _, _, err := SetupProjectUserEventNameAgentReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	startTime := U.TimeNowIn(U.TimeZoneStringIST).AddDate(0, 0, -5)

	// user 1
	userID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, userID1)
	{
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s0", userID1, startTime.Unix(), "A", "B", "UA", "UB")
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.Nil(t, response["user_id"])

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s1", userID1, startTime.Add(15*time.Minute).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s2", userID1, startTime.Add(3*time.Hour).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s3", userID1, startTime.AddDate(0, 0, 1).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s4", userID1, startTime.AddDate(0, 0, 90).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s5", userID1, startTime.AddDate(0, 0, 91).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// user 2
	userID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, userID2)
	{
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s0", userID2, startTime.Add(5*time.Minute).Unix(), "A", "B", "UA", "UB")
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.Nil(t, response["user_id"])

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s1", userID2, startTime.Add(40*time.Minute).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s2", userID2, startTime.Add(5*time.Hour).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d,"event_properties":{"value1":"%s","value2":"%s"},"user_properties":{"user_value1":"%s","user_value2":"%s"}}`,
			"s3", userID2, startTime.AddDate(0, 0, 2).Unix(), "A", "B", "UA", "UB")
		w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 15 min conversion time
	query := model.Query{
		From: startTime.Unix(),
		To:   startTime.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s3",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Timezone:        string(U.TimeZoneStringIST),
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
		ConversionTime:  "15M",
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), false)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	query = model.Query{
		From: startTime.Unix(),
		To:   startTime.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s3",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 1,
				Property:       "user_value1",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 1,
				Property:       "user_value2",
			},
		},
		Class:           model.QueryClassFunnel,
		Timezone:        string(U.TimeZoneStringIST),
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
		ConversionTime:  "15M",
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, "$no_group", result.Rows[0][1])
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "UA", result.Rows[1][0])
	assert.Equal(t, "UB", result.Rows[1][1])
	assert.Equal(t, float64(2), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])

	// 3 hour conversion time
	query = model.Query{
		From: startTime.Unix(),
		To:   startTime.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s3",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 1,
				Property:       "user_value1",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 1,
				Property:       "user_value2",
			},
		},
		Class:           model.QueryClassFunnel,
		Timezone:        string(U.TimeZoneStringIST),
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
		ConversionTime:  "3H",
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "UA", result.Rows[1][0])
	assert.Equal(t, "UB", result.Rows[1][1])
	assert.Equal(t, float64(2), result.Rows[1][2])
	assert.Equal(t, float64(2), result.Rows[1][3])
	assert.Equal(t, float64(1), result.Rows[1][5])

	// 2 day conversion time
	query = model.Query{
		From: startTime.Unix(),
		To:   startTime.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s3",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 1,
				Property:       "user_value1",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s0",
				EventNameIndex: 1,
				Property:       "user_value2",
			},
		},
		Class:           model.QueryClassFunnel,
		Timezone:        string(U.TimeZoneStringIST),
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
		ConversionTime:  "2D",
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "UA", result.Rows[1][0])
	assert.Equal(t, "UB", result.Rows[1][1])
	assert.Equal(t, float64(2), result.Rows[1][2])
	assert.Equal(t, float64(2), result.Rows[1][3])
	assert.Equal(t, float64(2), result.Rows[1][5])
	assert.Equal(t, float64(2), result.Rows[1][7])

	// 1 day conversion time
	query.ConversionTime = "1D"
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "UA", result.Rows[1][0])
	assert.Equal(t, "UB", result.Rows[1][1])
	assert.Equal(t, float64(2), result.Rows[1][2])
	assert.Equal(t, float64(2), result.Rows[1][3])
	assert.Equal(t, float64(2), result.Rows[1][5])
	assert.Equal(t, float64(1), result.Rows[1][7])

	// default conversion time 90 days check
	// should fail, s5 outside 90 days
	query = model.Query{
		From: startTime.Unix(),
		To:   startTime.AddDate(0, 0, 91).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       "s5",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		Timezone:        string(U.TimeZoneStringIST),
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(0), result.Rows[0][1])

	// set 91 days, should passs 1 user
	query.ConversionTime = "91"
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	// any order 1 day any order converison time

	query = model.Query{
		From: startTime.Unix(),
		To:   startTime.AddDate(0, 0, 5).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "s2",
				Properties: []model.QueryProperty{},
			},

			model.QueryEventWithProperties{
				Name:       "s1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s0",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "s3",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s2",
				EventNameIndex: 1,
				Property:       "user_value1",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventName:      "s2",
				EventNameIndex: 1,
				Property:       "user_value2",
			},
		},
		Class:           model.QueryClassFunnel,
		Timezone:        string(U.TimeZoneStringIST),
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondFunnelAnyGivenEvent,
		ConversionTime:  "1D",
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "UA", result.Rows[1][0])
	assert.Equal(t, "UB", result.Rows[1][1])
	assert.Equal(t, float64(2), result.Rows[1][2])
	assert.Equal(t, float64(2), result.Rows[1][3])
	assert.Equal(t, float64(2), result.Rows[1][5])
	assert.Equal(t, float64(1), result.Rows[1][7])
}

func TestAnalyticsEventsGroupQuery(t *testing.T) {
	project, _, _, _, err := SetupProjectUserEventNameAgentReturnDAO()
	assert.Nil(t, err)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	groupProperties := map[string]map[string]interface{}{
		"group_1": {"group_id": "group_1"},
		"group_2": {"group_id": "group_2"},
		"group_3": {"group_id": "group_3"},
		"group_4": {"group_id": "group_4"},
	}

	groupGroupName := map[string]string{
		"group_1": model.GROUP_NAME_HUBSPOT_COMPANY,
		"group_2": model.GROUP_NAME_HUBSPOT_COMPANY,
		"group_3": model.GROUP_NAME_SALESFORCE_ACCOUNT,
		"group_4": model.GROUP_NAME_SALESFORCE_ACCOUNT,
	}

	for _, groupName := range groupGroupName {
		_, status := store.GetStore().CreateGroup(project.ID, groupName, model.AllowedGroupNames)
		assert.Contains(t, []int{http.StatusCreated, http.StatusConflict}, status)
	}

	groupUserIDMap := map[string]string{}

	for groupID, propertiesMap := range groupProperties {

		properties, err := U.EncodeToPostgresJsonb(&propertiesMap)
		assert.Nil(t, err)

		// create group user
		groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
			ProjectId: project.ID, JoinTimestamp: U.TimeNowUnix(),
			Source:     model.GetRequestSourcePointer(model.UserSourceHubspot),
			Properties: *properties,
		}, groupGroupName[groupID], groupID)
		assert.Equal(t, http.StatusCreated, status)
		groupUserIDMap[groupID] = groupUserID
	}

	userProperties := map[string]map[string]interface{}{
		"user_1": {"user": "user_1"},
		"user_2": {"user": "user_2"},
		"user_3": {"user": "user_3"},
		"user_4": {"user": "user_4"},
		"user_5": {"user": "user_5"},
		"user_6": {"user": "user_6"},
	}
	userCustomerUserID := map[string]string{
		"user_1": "cuid1",
		"user_2": "cuid2",
		"user_4": "cuid4",
	}

	userGroups := map[string]string{
		"user_1": "group_1",
		"user_2": "group_2",
		"user_3": "group_3",
		"user_4": "group_4",
		"user_6": "group_2",
	}

	userIDs := map[string]string{}
	for user, propertiesMap := range userProperties {
		properties, err := U.EncodeToPostgresJsonb(&propertiesMap)
		assert.Nil(t, err)

		userID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties,
			CustomerUserId: userCustomerUserID[user], Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		userIDs[user] = userID

		if userGroups[user] != "" {
			_, status := store.GetStore().UpdateUserGroup(project.ID, userID, groupGroupName[userGroups[user]], userGroups[user],
				groupUserIDMap[userGroups[user]], false)
			assert.Equal(t, http.StatusAccepted, status)
		}
	}

	groupEvents := map[string]string{
		"group_1": U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
		"group_2": U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
		"group_3": U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
		"group_4": U.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED,
	}

	for groupID, groupEventName := range groupEvents {
		// group events
		groupEventTimestamp := U.TimeNowZ().Unix()
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$company_cost":%d}}`,
			groupEventName, groupUserIDMap[groupID], groupEventTimestamp, 1)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
	}

	userEvents := map[string]string{
		"user_1": U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
		"user_2": "event_1",
		"user_5": "event_1",
		"user_3": "crm_event_1",
		"user_6": "crm_event_1",
	}
	for user, eventName := range userEvents {
		userEventTimestamp := U.TimeNowZ().Unix()
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d, "event_properties":{"$user_cost":%d}}`,
			eventName, userIDs[user], userEventTimestamp, 1)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
	}

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001677'
		                    AND timestamp >= '1684436272'
		                    AND timestamp <= '1684456272'
		                    AND (
		                        events.event_name_id = 'f83cb2f0-07ac-4b65-9806-e40a6361d8d6'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        GROUP BY
		            group_user_id
		    )
		    SELECT
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        step_0
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        100000
	*/
	// total groups with hubspot company created event
	query := model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])

	// invalid check- group analysis different than event group
	query.GroupAnalysis = model.GROUP_NAME_SALESFORCE_ACCOUNT
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusBadRequest, errCode)

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001679'
		                    AND timestamp >= '1684436839'
		                    AND timestamp <= '1684456839'
		                    AND (
		                        events.event_name_id = '44938ea7-2262-4143-89d7-fdc1e4ec0a84'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    step_1 AS (
		        SELECT
		            COALESCE(
		                step_1_event_users_view.group_user_id,
		                step_1_event_users_view.user_group_user_id
		            ) as group_user_id,
		            FIRST(
		                step_1_event_users_view.user_id,
		                FROM_UNIXTIME(step_1_event_users_view.timestamp)
		            ) as event_user_id
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties,
		                    user_groups.group_1_user_id as group_user_id,
		                    users.group_1_user_id as user_group_user_id
		                FROM
		                    events
		                    LEFT JOIN users ON events.user_id = users.id
		                    AND users.project_id = '16001679'
		                    LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                    AND user_groups.project_id = '16001679'
		                    AND user_groups.group_1_user_id IS NOT NULL
		                    AND user_groups.source = '2'
		                WHERE
		                    events.project_id = '16001679'
		                    AND timestamp >= '1684436839'
		                    AND timestamp <= '1684456839'
		                    AND (
		                        group_user_id IS NOT NULL
		                        OR user_group_user_id IS NOT NULL
		                    )
		                    AND (
		                        events.event_name_id = '1dd4d2fe-c5ed-419d-bb3b-b371795ef120'
		                    )
		                LIMIT
		                    10000000000
		            ) step_1_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    events_intersect AS (
		        SELECT
		            step_0.event_user_id as event_user_id,
		            step_0.group_user_id as group_user_id
		        FROM
		            step_0
		            JOIN step_1 ON step_1.group_user_id = step_0.group_user_id
		    )
		    SELECT
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        events_intersect
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        100000
	*/
	// total groups who performed all events, only group 2
	query = model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id,
		            '0_$hubspot_company_created' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001680'
		                    AND timestamp >= '1684437025'
		                    AND timestamp <= '1684457025'
		                    AND (
		                        events.event_name_id = '9a181ab3-378a-4731-b60f-b56284849ed0'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    step_1 AS (
		        SELECT
		            COALESCE(
		                step_1_event_users_view.group_user_id,
		                step_1_event_users_view.user_group_user_id
		            ) as group_user_id,
		            FIRST(
		                step_1_event_users_view.user_id,
		                FROM_UNIXTIME(step_1_event_users_view.timestamp)
		            ) as event_user_id,
		            '1_event_1' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties,
		                    user_groups.group_2_user_id as group_user_id,
		                    users.group_2_user_id as user_group_user_id
		                FROM
		                    events
		                    LEFT JOIN users ON events.user_id = users.id
		                    AND users.project_id = '16001680'
		                    LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                    AND user_groups.project_id = '16001680'
		                    AND user_groups.group_2_user_id IS NOT NULL
		                    AND user_groups.source = '2'
		                WHERE
		                    events.project_id = '16001680'
		                    AND timestamp >= '1684437025'
		                    AND timestamp <= '1684457025'
		                    AND (
		                        group_user_id IS NOT NULL
		                        OR user_group_user_id IS NOT NULL
		                    )
		                    AND (
		                        events.event_name_id = '50b4297e-942b-4832-9972-d5f7e3ad6611'
		                    )
		                LIMIT
		                    10000000000
		            ) step_1_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    each_events_union AS (
		        SELECT
		            step_0.event_name as event_name,
		            step_0.group_user_id as group_user_id,
		            step_0.event_user_id as event_user_id
		        FROM
		            step_0
		        UNION
		        ALL
		        SELECT
		            step_1.event_name as event_name,
		            step_1.group_user_id as group_user_id,
		            step_1.event_user_id as event_user_id
		        FROM
		            step_1
		    )
		    SELECT
		        event_name,
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        each_events_union
		    GROUP BY
		        event_name
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        100000
	*/
	// total groups who performed each events, hubspot company created - 2(group_1, group_2), event_1 - 1(group_2)
	query = model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[1][2])

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id,
		            '0_$hubspot_company_created' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001681'
		                    AND timestamp >= '1684437193'
		                    AND timestamp <= '1684457193'
		                    AND (
		                        events.event_name_id = 'c4ad6d57-cb30-4fce-bf25-d546ec9b47e7'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    step_1 AS (
		        SELECT
		            COALESCE(
		                step_1_event_users_view.group_user_id,
		                step_1_event_users_view.user_group_user_id
		            ) as group_user_id,
		            FIRST(
		                step_1_event_users_view.user_id,
		                FROM_UNIXTIME(step_1_event_users_view.timestamp)
		            ) as event_user_id,
		            '1_crm_event_1' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties,
		                    user_groups.group_2_user_id as group_user_id,
		                    users.group_2_user_id as user_group_user_id
		                FROM
		                    events
		                    LEFT JOIN users ON events.user_id = users.id
		                    AND users.project_id = '16001681'
		                    LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                    AND user_groups.project_id = '16001681'
		                    AND user_groups.group_2_user_id IS NOT NULL
		                    AND user_groups.source = '2'
		                WHERE
		                    events.project_id = '16001681'
		                    AND timestamp >= '1684437193'
		                    AND timestamp <= '1684457193'
		                    AND (
		                        group_user_id IS NOT NULL
		                        OR user_group_user_id IS NOT NULL
		                    )
		                    AND (
		                        events.event_name_id = 'a0ad2adc-a2d6-47f8-b104-3f849dd48020'
		                    )
		                LIMIT
		                    10000000000
		            ) step_1_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    each_events_union AS (
		        SELECT
		            step_0.event_name as event_name,
		            step_0.group_user_id as group_user_id,
		            step_0.event_user_id as event_user_id
		        FROM
		            step_0
		        UNION
		        ALL
		        SELECT
		            step_1.event_name as event_name,
		            step_1.group_user_id as group_user_id,
		            step_1.event_user_id as event_user_id
		        FROM
		            step_1
		    )
		    SELECT
		        event_name,
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        each_events_union
		    GROUP BY
		        event_name
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        100000
	*/
	// total groups who performed each events, hubspot company created - 2(group_1, group_2), crm_event_1 - 1(group_2)( user directly associated with group instead of using customer user id)
	query = model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "crm_event_1",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[1][2])

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001682'
		                    AND timestamp >= '1684437390'
		                    AND timestamp <= '1684457390'
		                    AND (
		                        events.event_name_id = '8d72806d-f60d-4663-bd47-58d4b860a9fb'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    step_1 AS (
		        SELECT
		            COALESCE(
		                step_1_event_users_view.group_user_id,
		                step_1_event_users_view.user_group_user_id
		            ) as group_user_id,
		            FIRST(
		                step_1_event_users_view.user_id,
		                FROM_UNIXTIME(step_1_event_users_view.timestamp)
		            ) as event_user_id
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties,
		                    user_groups.group_1_user_id as group_user_id,
		                    users.group_1_user_id as user_group_user_id
		                FROM
		                    events
		                    LEFT JOIN users ON events.user_id = users.id
		                    AND users.project_id = '16001682'
		                    LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                    AND user_groups.project_id = '16001682'
		                    AND user_groups.group_1_user_id IS NOT NULL
		                    AND user_groups.source = '2'
		                WHERE
		                    events.project_id = '16001682'
		                    AND timestamp >= '1684437390'
		                    AND timestamp <= '1684457390'
		                    AND (
		                        group_user_id IS NOT NULL
		                        OR user_group_user_id IS NOT NULL
		                    )
		                    AND (
		                        events.event_name_id = 'f23528ea-761c-4f19-8fb2-a4a75946c9c0'
		                    )
		                LIMIT
		                    10000000000
		            ) step_1_event_users_view
		        GROUP BY
		            group_user_id
		    ),
		    events_intersect AS (
		        SELECT
		            step_0.event_user_id as event_user_id,
		            step_0.group_user_id as group_user_id
		        FROM
		            step_0
		            JOIN step_1 ON step_1.group_user_id = step_0.group_user_id
		    )
		    SELECT
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        events_intersect
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        100000
	*/
	query.EventsCondition = model.EventCondAllGivenEvent
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id,
		            '0_$hubspot_company_created' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001683'
		                    AND timestamp >= '1684437595'
		                    AND timestamp <= '1684457595'
		                    AND (
		                        events.event_name_id = '67ca8555-ef97-46e2-9fb0-e594c10529ab'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        GROUP BY
		            group_user_id
		        ORDER BY
		            group_user_id,
		            step_0_event_users_view.timestamp ASC
		    ),
		    step_1 AS (
		        SELECT
		            COALESCE(
		                step_1_event_users_view.group_user_id,
		                step_1_event_users_view.user_group_user_id
		            ) as group_user_id,
		            FIRST(
		                step_1_event_users_view.user_id,
		                FROM_UNIXTIME(step_1_event_users_view.timestamp)
		            ) as event_user_id,
		            '1_event_1' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties,
		                    user_groups.group_1_user_id as group_user_id,
		                    users.group_1_user_id as user_group_user_id
		                FROM
		                    events
		                    LEFT JOIN users ON events.user_id = users.id
		                    AND users.project_id = '16001683'
		                    LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                    AND user_groups.project_id = '16001683'
		                    AND user_groups.group_1_user_id IS NOT NULL
		                    AND user_groups.source = '2'
		                WHERE
		                    events.project_id = '16001683'
		                    AND timestamp >= '1684437595'
		                    AND timestamp <= '1684457595'
		                    AND (
		                        group_user_id IS NOT NULL
		                        OR user_group_user_id IS NOT NULL
		                    )
		                    AND (
		                        events.event_name_id = '25b4fb1a-8797-4248-afc8-109c2adddafe'
		                    )
		                LIMIT
		                    10000000000
		            ) step_1_event_users_view
		        GROUP BY
		            group_user_id
		        ORDER BY
		            group_user_id,
		            step_1_event_users_view.timestamp ASC
		    ),
		    each_events_union AS (
		        SELECT
		            step_0.event_name as event_name,
		            step_0.group_user_id as group_user_id,
		            step_0.event_user_id as event_user_id
		        FROM
		            step_0
		        UNION
		        ALL
		        SELECT
		            step_1.event_name as event_name,
		            step_1.group_user_id as group_user_id,
		            step_1.event_user_id as event_user_id
		        FROM
		            step_1
		    ),
		    each_users_union AS (
		        SELECT
		            each_events_union.event_user_id,
		            each_events_union.group_user_id,
		            each_events_union.event_name,
		            CASE
		                WHEN JSON_EXTRACT_STRING(group_users.properties, 'group_id') IS NULL THEN '$none'
		                WHEN JSON_EXTRACT_STRING(group_users.properties, 'group_id') = '' THEN '$none'
		                ELSE JSON_EXTRACT_STRING(group_users.properties, 'group_id')
		            END AS _group_key_0
		        FROM
		            each_events_union
		            LEFT JOIN users AS group_users ON each_events_union.group_user_id = group_users.id
		            AND group_users.project_id = 16001683
		    )
		    SELECT
		        event_name,
		        _group_key_0,
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        each_users_union
		    GROUP BY
		        event_name,
		        _group_key_0
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        10000
	*/
	// global breakdown property filter test
	query = model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				Property:  "group_id",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		eventNameI := U.GetPropertyValueAsString(result.Rows[i][0])
		eventNameJ := U.GetPropertyValueAsString(result.Rows[j][0])
		eventNameI1, _ := U.GetPropertyValueAsFloat64(result.Rows[i][1])
		eventNameJ1, _ := U.GetPropertyValueAsFloat64(result.Rows[j][1])
		if eventNameI < eventNameJ {
			return true
		}

		if eventNameI1 > eventNameJ1 {
			return false
		}

		return eventNameI1 < eventNameJ1
	})

	assert.Equal(t, "$hubspot_company_created", result.Rows[0][1])
	assert.Equal(t, "group_1", result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "$hubspot_company_created", result.Rows[1][1])
	assert.Equal(t, "group_2", result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "event_1", result.Rows[2][1])
	assert.Equal(t, "group_2", result.Rows[2][2])
	assert.Equal(t, float64(1), result.Rows[2][3])

	/*
			        WITH step_0 AS (
		            SELECT
		                step_0_event_users_view.user_id as group_user_id,
		                FIRST(
		                    step_0_event_users_view.user_id,
		                    FROM_UNIXTIME(step_0_event_users_view.timestamp)
		                ) as event_user_id,
		                '0_$hubspot_company_created' AS event_name
		            FROM
		                (
		                    SELECT
		                        events.project_id,
		                        events.id,
		                        events.event_name_id,
		                        events.user_id,
		                        events.timestamp,
		                        events.properties as event_properties,
		                        events.user_properties as event_user_properties
		                    FROM
		                        events
		                    WHERE
		                        events.project_id = '16001683'
		                        AND timestamp >= '1684437595'
		                        AND timestamp <= '1684457595'
		                        AND (
		                            events.event_name_id = '67ca8555-ef97-46e2-9fb0-e594c10529ab'
		                        )
		                    LIMIT
		                        10000000000
		                ) step_0_event_users_view
		            GROUP BY
		                group_user_id
		            ORDER BY
		                group_user_id,
		                step_0_event_users_view.timestamp ASC
		        ),
		        step_1 AS (
		            SELECT
		                COALESCE(
		                    step_1_event_users_view.group_user_id,
		                    step_1_event_users_view.user_group_user_id
		                ) as group_user_id,
		                FIRST(
		                    step_1_event_users_view.user_id,
		                    FROM_UNIXTIME(step_1_event_users_view.timestamp)
		                ) as event_user_id,
		                '1_event_1' AS event_name
		            FROM
		                (
		                    SELECT
		                        events.project_id,
		                        events.id,
		                        events.event_name_id,
		                        events.user_id,
		                        events.timestamp,
		                        events.properties as event_properties,
		                        events.user_properties as event_user_properties,
		                        user_groups.group_1_user_id as group_user_id,
		                        users.group_1_user_id as user_group_user_id
		                    FROM
		                        events
		                        LEFT JOIN users ON events.user_id = users.id
		                        AND users.project_id = '16001683'
		                        LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                        AND user_groups.project_id = '16001683'
		                        AND user_groups.group_1_user_id IS NOT NULL
		                        AND user_groups.source = '2'
		                    WHERE
		                        events.project_id = '16001683'
		                        AND timestamp >= '1684437595'
		                        AND timestamp <= '1684457595'
		                        AND (
		                            group_user_id IS NOT NULL
		                            OR user_group_user_id IS NOT NULL
		                        )
		                        AND (
		                            events.event_name_id = '25b4fb1a-8797-4248-afc8-109c2adddafe'
		                        )
		                    LIMIT
		                        10000000000
		                ) step_1_event_users_view
		            GROUP BY
		                group_user_id
		            ORDER BY
		                group_user_id,
		                step_1_event_users_view.timestamp ASC
		        ),
		        each_events_union AS (
		            SELECT
		                step_0.event_name as event_name,
		                step_0.group_user_id as group_user_id,
		                step_0.event_user_id as event_user_id
		            FROM
		                step_0
		            UNION
		            ALL
		            SELECT
		                step_1.event_name as event_name,
		                step_1.group_user_id as group_user_id,
		                step_1.event_user_id as event_user_id
		            FROM
		                step_1
		        ),
		        each_users_union AS (
		            SELECT
		                each_events_union.event_user_id,
		                each_events_union.group_user_id,
		                each_events_union.event_name,
		                CASE
		                    WHEN JSON_EXTRACT_STRING(group_users.properties, 'group_id') IS NULL THEN '$none'
		                    WHEN JSON_EXTRACT_STRING(group_users.properties, 'group_id') = '' THEN '$none'
		                    ELSE JSON_EXTRACT_STRING(group_users.properties, 'group_id')
		                END AS _group_key_0
		            FROM
		                each_events_union
		                LEFT JOIN users AS group_users ON each_events_union.group_user_id = group_users.id
		                AND group_users.project_id = 16001683
		        )
		        SELECT
		            event_name,
		            _group_key_0,
		            COUNT(DISTINCT(group_user_id)) AS aggregate
		        FROM
		            each_users_union
		        GROUP BY
		            event_name,
		            _group_key_0
		        ORDER BY
		            aggregate DESC
		        LIMIT
		            10000
	*/
	// global property property filter test
	query = model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				Property:  "group_id",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GlobalUserProperties: []model.QueryProperty{

			model.QueryProperty{
				Entity:    model.PropertyEntityUserGlobal,
				Property:  "group_id",
				Operator:  "equals",
				Value:     "group_1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		eventNameI := U.GetPropertyValueAsString(result.Rows[i][0])
		eventNameJ := U.GetPropertyValueAsString(result.Rows[j][0])
		eventNameI1, _ := U.GetPropertyValueAsFloat64(result.Rows[i][1])
		eventNameJ1, _ := U.GetPropertyValueAsFloat64(result.Rows[j][1])
		if eventNameI < eventNameJ {
			return true
		}

		if eventNameI1 > eventNameJ1 {
			return false
		}

		return eventNameI1 < eventNameJ1
	})

	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "$hubspot_company_created", result.Rows[0][1])
	assert.Equal(t, "group_1", result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "event_1", result.Rows[1][1])
	assert.Equal(t, "group_1", result.Rows[1][2])
	assert.Equal(t, int(0), result.Rows[1][3])

	/*
			    WITH step_0 AS (
		        SELECT
		            step_0_event_users_view.user_id as group_user_id,
		            FIRST(
		                step_0_event_users_view.user_id,
		                FROM_UNIXTIME(step_0_event_users_view.timestamp)
		            ) as event_user_id,
		            '0_$hubspot_company_created' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties
		                FROM
		                    events
		                WHERE
		                    events.project_id = '16001685'
		                    AND timestamp >= '1684437958'
		                    AND timestamp <= '1684457958'
		                    AND (
		                        events.event_name_id = '768b0050-04cf-4ae0-af83-4418a46cbe88'
		                    )
		                LIMIT
		                    10000000000
		            ) step_0_event_users_view
		        WHERE
		            (
		                (
		                    JSON_EXTRACT_STRING(
		                        step_0_event_users_view.event_user_properties,
		                        'group_id'
		                    ) = 'group_1'
		                )
		            )
		        GROUP BY
		            group_user_id
		        ORDER BY
		            group_user_id,
		            step_0_event_users_view.timestamp ASC
		    ),
		    step_1 AS (
		        SELECT
		            COALESCE(
		                step_1_event_users_view.group_user_id,
		                step_1_event_users_view.user_group_user_id
		            ) as group_user_id,
		            FIRST(
		                step_1_event_users_view.user_id,
		                FROM_UNIXTIME(step_1_event_users_view.timestamp)
		            ) as event_user_id,
		            '1_event_1' AS event_name
		        FROM
		            (
		                SELECT
		                    events.project_id,
		                    events.id,
		                    events.event_name_id,
		                    events.user_id,
		                    events.timestamp,
		                    events.properties as event_properties,
		                    events.user_properties as event_user_properties,
		                    user_groups.group_2_user_id as group_user_id,
		                    users.group_2_user_id as user_group_user_id
		                FROM
		                    events
		                    LEFT JOIN users ON events.user_id = users.id
		                    AND users.project_id = '16001685'
		                    LEFT JOIN users AS user_groups ON users.customer_user_id = user_groups.customer_user_id
		                    AND user_groups.project_id = '16001685'
		                    AND user_groups.group_2_user_id IS NOT NULL
		                    AND user_groups.source = '2'
		                WHERE
		                    events.project_id = '16001685'
		                    AND timestamp >= '1684437958'
		                    AND timestamp <= '1684457958'
		                    AND (
		                        group_user_id IS NOT NULL
		                        OR user_group_user_id IS NOT NULL
		                    )
		                    AND (
		                        events.event_name_id = 'c46d6c76-2d6b-499c-aa68-eeefdbc697db'
		                    )
		                LIMIT
		                    10000000000
		            ) step_1_event_users_view
		        GROUP BY
		            group_user_id
		        ORDER BY
		            group_user_id,
		            step_1_event_users_view.timestamp ASC
		    ),
		    each_events_union AS (
		        SELECT
		            step_0.event_name as event_name,
		            step_0.group_user_id as group_user_id,
		            step_0.event_user_id as event_user_id
		        FROM
		            step_0
		        UNION
		        ALL
		        SELECT
		            step_1.event_name as event_name,
		            step_1.group_user_id as group_user_id,
		            step_1.event_user_id as event_user_id
		        FROM
		            step_1
		    ),
		    each_users_union AS (
		        SELECT
		            each_events_union.event_user_id,
		            each_events_union.group_user_id,
		            each_events_union.event_name,
		            CASE
		                WHEN JSON_EXTRACT_STRING(group_users.properties, 'group_id') IS NULL THEN '$none'
		                WHEN JSON_EXTRACT_STRING(group_users.properties, 'group_id') = '' THEN '$none'
		                ELSE JSON_EXTRACT_STRING(group_users.properties, 'group_id')
		            END AS _group_key_0
		        FROM
		            each_events_union
		            LEFT JOIN users AS group_users ON each_events_union.group_user_id = group_users.id
		            AND group_users.project_id = 16001685
		    )
		    SELECT
		        event_name,
		        _group_key_0,
		        COUNT(DISTINCT(group_user_id)) AS aggregate
		    FROM
		        each_users_union
		    GROUP BY
		        event_name,
		        _group_key_0
		    ORDER BY
		        aggregate DESC
		    LIMIT
		        10000
	*/
	// Event level filter test
	query = model.Query{
		From: U.TimeNowUnix() - 10000,
		To:   U.TimeNowUnix() + 10000,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED,
				Properties: []model.QueryProperty{
					model.QueryProperty{
						Entity:    model.PropertyEntityUser,
						Property:  "group_id",
						Operator:  "equals",
						Value:     "group_1",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "AND",
					},
				},
			},
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				Property:  "group_id",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_HUBSPOT_COMPANY,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		eventNameI := U.GetPropertyValueAsString(result.Rows[i][0])
		eventNameJ := U.GetPropertyValueAsString(result.Rows[j][0])
		eventNameI1, _ := U.GetPropertyValueAsFloat64(result.Rows[i][1])
		eventNameJ1, _ := U.GetPropertyValueAsFloat64(result.Rows[j][1])
		if eventNameI < eventNameJ {
			return true
		}

		if eventNameI1 > eventNameJ1 {
			return false
		}

		return eventNameI1 < eventNameJ1
	})

	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "$hubspot_company_created", result.Rows[0][1])
	assert.Equal(t, "group_1", result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "event_1", result.Rows[1][1])
	assert.Equal(t, "group_2", result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
}

func TestAnalyticsSixSignalGroupQuery(t *testing.T) {
	project, _, _, _, err := SetupProjectUserEventNameAgentReturnDAO()
	assert.Nil(t, err)

	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	_, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SIX_SIGNAL, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)

	// user 1 with 6signal group 1
	properties := map[string]interface{}{"group_user_no": 1}
	properties1JSONB, err := U.EncodeStructTypeToPostgresJsonb(properties)
	assert.Nil(t, err)
	userID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties1JSONB,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	groupProperties := &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc.com"}
	status = SDK.TrackUserAccountGroup(project.ID, userID, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	payload := fmt.Sprintf(`{"event_name": "event_1", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userID, U.TimeNowZ().Add(-10*time.Minute).Unix(), 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// user 2 with 6signal group 2
	properties = map[string]interface{}{"group_user_no": 2}
	properties1JSONB, err = U.EncodeStructTypeToPostgresJsonb(properties)
	assert.Nil(t, err)
	userID, errCode = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: *properties1JSONB,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc2.com"}
	status = SDK.TrackUserAccountGroup(project.ID, userID, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	payload = fmt.Sprintf(`{"event_name": "event_1", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userID, U.TimeNowZ().Add(-10*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "event_2", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userID, U.TimeNowZ().Add(-9*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Events query total groups who performed event_1
	query := model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "event_2",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_SIX_SIGNAL,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}
	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][0])

	// total groups who performed any given event
	query.EventsCondition = model.EventCondAnyGivenEvent
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])

	// total groups who performed each given event
	query.EventsCondition = model.EventCondEachGivenEvent
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[1][2])

	// Events query total groups who performed event_1, breakdown by
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "event_2",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:   model.PropertyEntityUser,
				Property: U.SIX_SIGNAL_DOMAIN,
			},
		},
		Class:           model.QueryClassEvents,
		GroupAnalysis:   U.GROUP_NAME_SIX_SIGNAL,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAnyGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, "abc.com", result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])
	assert.Equal(t, "abc2.com", result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])

	/*
		Funnel query event1 to event_2
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "event_2",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_SIX_SIGNAL,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name:       "event_1",
				Properties: []model.QueryProperty{},
			},
			model.QueryEventWithProperties{
				Name:       "event_2",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   U.GROUP_NAME_SIX_SIGNAL,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])
	assert.Equal(t, float64(1), result.Rows[0][2])

	assert.Equal(t, "abc.com", result.Rows[1][0])
	assert.Equal(t, float64(1), result.Rows[1][1])
	assert.Equal(t, float64(0), result.Rows[1][2])

	assert.Equal(t, "abc2.com", result.Rows[2][0])
	assert.Equal(t, float64(1), result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
}

func TestAnalyticsFunnelAllAccounts(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	/*
		userWeb1(domain: abc1.com) - event(xyz.com)
		userWeb2(domain: abc2.com) - event(xyz2.com)
		userWeb3(domain: abc3.com) - event(xyz.com)

		groupUserHubspot1(domain: abc1.com) - event(hubspot_contact_created, hubspot_contact_update)
		groupUserHubspot2(domain: abc2.com) - event(hubspot_contact_created, hubspot_contact_update)

		groupUserSalesforce1(domain: abc1.com) - event(salesforce_account_created, salesforce_account_updated)
		groupUserSalesforce2(domain: abc2.com) - event(salesforce_account_created, salesforce_account_updated)
	*/
	properties := postgres.Jsonb{[]byte(`{"user_no":"w1"}`)}
	userWeb1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w2"}`)}
	userWeb2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid2"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w3"}`)}
	userWeb3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid3"})
	assert.Equal(t, http.StatusCreated, errCode)

	dateTimeUTC := util.TimeNowZ()
	propertiesMap := U.PropertiesMap{"$hubspot_company_name": "abc1", "$hubspot_company_domain": "abc1.com", "$hubspot_company_region": "A", "hs_company_no": "h1", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	groupUserHubspot1, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc1.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	propertiesMap = U.PropertiesMap{"$hubspot_company_name": "abc2", "$hubspot_company_domain": "abc2.com", "$hubspot_company_region": "B", "hs_company_no": "h2", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	groupUserHubspot2, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc2.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	properties = postgres.Jsonb{[]byte(`{"user_no":"h1"}`)}
	userHubspot1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceHubspot), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().UpdateUserGroup(project.ID, userHubspot1, model.GROUP_NAME_HUBSPOT_COMPANY, "habc", groupUserHubspot1, false)
	assert.Equal(t, http.StatusAccepted, errCode)

	propertiesMap = U.PropertiesMap{"sf_account_no": "s123"}
	groupUserSalesforce1, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, "abc1.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	propertiesMap = U.PropertiesMap{"sf_account_no": "s234"}
	groupUserSalesforce2, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, "abc2.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties := &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc1.com", U.SIX_SIGNAL_REGION: "A"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb1, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc2.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb2, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc3.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb3, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	status = store.GetStore().AssociateUserDomainsGroup(project.ID, userWeb1, "", "")
	assert.Equal(t, http.StatusOK, status)
	status = store.GetStore().AssociateUserDomainsGroup(project.ID, userWeb2, "", "")
	assert.Equal(t, http.StatusOK, status)

	payload := fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb3, U.TimeNowZ().Add(-10*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 3)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 4)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_updated", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 5)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 6)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_updated", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 7)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$salesforce_account_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserSalesforce1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 8)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$salesforce_account_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserSalesforce2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 9)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	// Query: Total accounts performed www.xyz.com -> $hubspot_company_created
	// expected result: 			userweb1, userbweb3 -> userweb1
	/*
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 FROM  (SELECT events.project_id, events.id, events.event_name_id,
		events.user_id, events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
		user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000179'  WHERE
		events.project_id='38000179' AND timestamp>='1690359103' AND timestamp<='1690361503' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '6a70a933-3fad-411e-98c5-b84b987e970b'
		)  LIMIT 10000000000) step_0_event_users_view GROUP BY coal_group_user_id),  step_1 AS (SELECT
		step_1_event_users_view.group_user_id as coal_group_user_id, step_1_event_users_view.timestamp, 1 as
		step_1 FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
		events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
		user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000179'  WHERE
		events.project_id='38000179' AND timestamp>='1690359103' AND timestamp<='1690361503' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'cf055a4e-47ac-4cfb-a04d-c10f69650613'
		)  LIMIT 10000000000) step_1_event_users_view GROUP BY coal_group_user_id,timestamp) ,
		step_1_step_0_users AS (SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp,
		FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS step_0_timestamp ,
		FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN
		step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
		step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT step_0 , step_1 ,
		step_0_timestamp , step_1_timestamp FROM step_0   LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT
		SUM(step_0) AS step_0 , SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS
		step_0_1_time FROM funnel
	*/
	query := model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "$hubspot_company_created",
				Properties: []model.QueryProperty{},
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, float64(2), result.Rows[0][0])
	assert.Equal(t, float64(1), result.Rows[0][1])

	// Query: Total accounts performed www.xyz.com -> $hubspot_company_created, breakdown by six_signal_domain and hubspot_company_domain
	/*
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 FROM  (SELECT events.project_id, events.id, events.event_name_id,
		events.user_id, events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
		user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000180'  WHERE
		events.project_id='38000180' AND timestamp>='1690359555' AND timestamp<='1690361955' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '037aba1c-d3bc-43ed-80f0-1ecfdcb24f5b'
		)  LIMIT 10000000000) step_0_event_users_view GROUP BY coal_group_user_id),  step_1 AS (SELECT
		step_1_event_users_view.group_user_id as coal_group_user_id, step_1_event_users_view.timestamp, 1 as
		step_1 FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
		events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
		user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000180'  WHERE
		events.project_id='38000180' AND timestamp>='1690359555' AND timestamp<='1690361955' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '9e26945e-3c07-4501-a1e4-628fe2cd15bd'
		)  LIMIT 10000000000) step_1_event_users_view GROUP BY coal_group_user_id,timestamp) ,
		step_1_step_0_users AS (SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp,
		FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS step_0_timestamp ,
		FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN
		step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
		step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT
		step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE WHEN
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
		WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
		ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_0,
		CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
		(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
		WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
		ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
		BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
		group_users.join_timestamp END) = '' THEN '$none' ELSE
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
		BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
		group_users.join_timestamp END) END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on
		step_0.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id = '38000180' AND
		group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
		NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT *
		FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
		_group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS
		_group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "$hubspot_company_created",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, "$no_group", result.Rows[0][1])
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "50.0", result.Rows[0][4])
	assert.Equal(t, "50.0", result.Rows[0][5])
	assert.Equal(t, "abc1.com", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])
	assert.Equal(t, "100.0", result.Rows[1][5])
	assert.Equal(t, "abc3.com", result.Rows[2][0])
	assert.Equal(t, "$none", result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
	assert.Equal(t, float64(0), result.Rows[2][3])
	assert.Equal(t, "0.0", result.Rows[2][4])
	assert.Equal(t, "0.0", result.Rows[2][5])

	// Filter test, global group,six_signal_domain = "abc1.com"
	/*
		"WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 FROM  (SELECT events.project_id, events.id, events.event_name_id,
		events.user_id, events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_2_user_id as group_user_id , group_users.properties as
		group_properties FROM events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND
		user_groups.project_id = '38000181' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
		group_users.group_2_user_id AND group_users.project_id = '38000181' AND group_users.is_group_user =
		true AND group_users.source IN ( 8 ) AND ( group_users.group_4_id IS NOT NULL ) WHERE
		events.project_id='38000181' AND timestamp>='1690359791' AND timestamp<='1690362191' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'deb64e7e-6a98-4e01-a971-50abfe259b80'
		)  LIMIT 10000000000) step_0_event_users_view WHERE
		(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = 'abc1.com')
		GROUP BY coal_group_user_id),  step_1 AS (SELECT step_1_event_users_view.group_user_id as
		coal_group_user_id, step_1_event_users_view.timestamp, 1 as step_1 FROM  (SELECT events.project_id,
		events.id, events.event_name_id, events.user_id, events.timestamp , events.properties as
		event_properties, events.user_properties as event_user_properties , user_groups.group_2_user_id as
		group_user_id , group_users.properties as group_properties FROM events  LEFT JOIN users as
		user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000181' LEFT JOIN
		users as group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND
		group_users.project_id = '38000181' AND group_users.is_group_user = true AND group_users.source IN (
		8 ) AND ( group_users.group_4_id IS NOT NULL ) WHERE events.project_id='38000181' AND
		timestamp>='1690359791' AND timestamp<='1690362191' AND  ( group_user_id IS NOT NULL  )
		AND  ( events.event_name_id = '43f91fd4-80ad-467c-ac68-52024be0716f' )  LIMIT 10000000000)
		step_1_event_users_view WHERE (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') = 'abc1.com') GROUP BY coal_group_user_id,timestamp) , step_1_step_0_users AS
		(SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as
		timestamp, step_1 , step_0.timestamp AS step_0_timestamp , FIRST(step_1.timestamp,
		FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN step_1 ON
		step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
		step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT
		step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE WHEN
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
		WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
		ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_0,
		CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
		(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
		WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
		ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
		BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
		group_users.join_timestamp END) = '' THEN '$none' ELSE
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
		BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
		group_users.join_timestamp END) END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on
		step_0.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id = '38000181' AND
		group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
		NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT *
		FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
		_group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS
		_group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "$hubspot_company_created",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  "equals",
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Len(t, result.Rows, 2)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, "$no_group", result.Rows[0][1])
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "100.0", result.Rows[0][5])
	assert.Equal(t, "abc1.com", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])
	assert.Equal(t, "100.0", result.Rows[1][5])

	// filter test AND between two different groups
	/*
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 , MAX(group_1_id) as max_group_1_id , MAX(group_4_id) as max_group_4_id FROM
		(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
		events.properties as event_properties, events.user_properties as event_user_properties ,
		user_groups.group_2_user_id as group_user_id , group_users.properties as group_properties ,
		group_users.group_4_id as group_4_id , group_users.group_1_id as group_1_id FROM events  LEFT JOIN
		users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000182' LEFT
		JOIN users as group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND
		group_users.project_id = '38000182' AND group_users.is_group_user = true AND group_users.source IN (
		8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_1_id IS NOT NULL ) WHERE
		events.project_id='38000182' AND timestamp>='1690360129' AND timestamp<='1690362529' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'c7bd7705-ba63-4a14-aa6f-209c0d745088'
		)  LIMIT 10000000000) step_0_event_users_view WHERE (
		(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$hubspot_company_domain') =
		'abc1.com') ) OR ( (JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain')
		= 'abc1.com') ) GROUP BY coal_group_user_id HAVING max_group_1_id IS NOT NULL AND max_group_4_id IS
		NOT NULL),  step_1 AS (SELECT step_1_event_users_view.group_user_id as coal_group_user_id,
		step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_4_id) as max_group_4_id , MAX(group_1_id)
		as max_group_1_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
		events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_2_user_id as group_user_id , group_users.properties as
		group_properties , group_users.group_4_id as group_4_id , group_users.group_1_id as group_1_id FROM
		events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
		= '38000182' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
		group_users.group_2_user_id AND group_users.project_id = '38000182' AND group_users.is_group_user =
		true AND group_users.source IN ( 2,8 ) AND ( group_users.group_1_id IS NOT NULL OR
		group_users.group_4_id IS NOT NULL ) WHERE events.project_id='38000182' AND
		timestamp>='1690360129' AND timestamp<='1690362529' AND  ( group_user_id IS NOT NULL  )
		AND  ( events.event_name_id = 'af297be6-108b-4bf9-a5c4-cf430fca7dfd' )  LIMIT 10000000000)
		step_1_event_users_view WHERE ( (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') = 'abc1.com') ) OR (
		(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_domain') =
		'abc1.com') ) GROUP BY coal_group_user_id,timestamp HAVING max_group_4_id IS NOT NULL AND
		max_group_1_id IS NOT NULL) , step_1_step_0_users AS (SELECT step_1.coal_group_user_id,
		FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS
		step_0_timestamp , FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM
		step_0 LEFT JOIN step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE
		step_1.timestamp >= step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT
		DISTINCT step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE
		WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
		WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
		ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
		step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_0,
		CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
		(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
		WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
		ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
		BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
		group_users.join_timestamp END) = '' THEN '$none' ELSE
		FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
		BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
		group_users.join_timestamp END) END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on
		step_0.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id = '38000182' AND
		group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
		NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT *
		FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
		_group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS
		_group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "$hubspot_company_created",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "$hubspot_company_updated",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  "equals",
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				Operator:  "equals",
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		Class:           model.QueryClassFunnel,
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Len(t, result.Rows, 2)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][0])
		p2 := U.GetPropertyValueAsString(result.Rows[j][0])
		return p1 < p2
	})
	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, "$no_group", result.Rows[0][1])
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "100.0", result.Rows[0][5])
	assert.Equal(t, "abc1.com", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])
	assert.Equal(t, "100.0", result.Rows[1][5])

	// // filter test, OR between to different groups
	// /*
	// 	WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
	// 	FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	// 	timestamp, 1 as step_0 FROM  (SELECT events.project_id, events.id, events.event_name_id,
	// 	events.user_id, events.timestamp , events.properties as event_properties, events.user_properties as
	// 	event_user_properties , user_groups.group_2_user_id as group_user_id , group_users.properties as
	// 	group_properties FROM events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND
	// 	user_groups.project_id = '38000184' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
	// 	group_users.group_2_user_id AND group_users.project_id = '38000184' AND group_users.is_group_user =
	// 	true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS NOT NULL OR
	// 	group_users.group_1_id IS NOT NULL ) WHERE events.project_id='38000184' AND
	// 	timestamp>='1690360534' AND timestamp<='1690362934' AND  ( group_user_id IS NOT NULL  )
	// 	AND  ( events.event_name_id = '640e93cc-c066-4b1a-830c-ad954dff2f29' )  LIMIT 10000000000)
	// 	step_0_event_users_view WHERE (JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
	// 	'$6Signal_domain') = 'abc1.com' OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
	// 	'$hubspot_company_domain') = 'abc2.com') GROUP BY coal_group_user_id),  step_1 AS (SELECT
	// 	step_1_event_users_view.group_user_id as coal_group_user_id, step_1_event_users_view.timestamp, 1 as
	// 	step_1 FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	// 	events.timestamp , events.properties as event_properties, events.user_properties as
	// 	event_user_properties , user_groups.group_2_user_id as group_user_id , group_users.properties as
	// 	group_properties FROM events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND
	// 	user_groups.project_id = '38000184' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
	// 	group_users.group_2_user_id AND group_users.project_id = '38000184' AND group_users.is_group_user =
	// 	true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS NOT NULL OR
	// 	group_users.group_1_id IS NOT NULL ) WHERE events.project_id='38000184' AND
	// 	timestamp>='1690360534' AND timestamp<='1690362934' AND  ( group_user_id IS NOT NULL  )
	// 	AND  ( events.event_name_id = '5cf380b2-3f2a-4328-a3b3-66bdb097741f' )  LIMIT 10000000000)
	// 	step_1_event_users_view WHERE (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_domain') = 'abc1.com' OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$hubspot_company_domain') = 'abc2.com') GROUP BY coal_group_user_id,timestamp) ,
	// 	step_1_step_0_users AS (SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp,
	// 	FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS step_0_timestamp ,
	// 	FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN
	// 	step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
	// 	step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT
	// 	step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
	// 	WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
	// 	ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_0,
	// 	CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
	// 	(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
	// 	WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
	// 	ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) = '' THEN '$none' ELSE
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on
	// 	step_0.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id = '38000184' AND
	// 	group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
	// 	NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
	// 	step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT *
	// 	FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
	// 	AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
	// 	_group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS
	// 	_group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
	// 	AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	// */
	// query = model.Query{
	// 	From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
	// 	To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
	// 	EventsWithProperties: []model.QueryEventWithProperties{
	// 		{
	// 			Name:       "$hubspot_company_created",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 		{
	// 			Name:       "$hubspot_company_updated",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 	},
	// 	GlobalUserProperties: []model.QueryProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			Operator:  "equals",
	// 			Value:     "abc2.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 	},
	// 	GroupByProperties: []model.QueryGroupByProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 	},
	// 	Class:           model.QueryClassFunnel,
	// 	GroupAnalysis:   model.GROUP_NAME_DOMAINS,
	// 	Type:            model.QueryTypeUniqueUsers,
	// 	EventsCondition: model.EventCondAllGivenEvent,
	// }

	// result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	// assert.Equal(t, http.StatusOK, errCode)

	// assert.Len(t, result.Rows, 3)
	// sort.Slice(result.Rows, func(i, j int) bool {
	// 	p1 := U.GetPropertyValueAsString(result.Rows[i][0])
	// 	p2 := U.GetPropertyValueAsString(result.Rows[j][0])
	// 	return p1 < p2
	// })

	// assert.Equal(t, "$no_group", result.Rows[0][0])
	// assert.Equal(t, "$no_group", result.Rows[0][1])
	// assert.Equal(t, float64(2), result.Rows[0][2])
	// assert.Equal(t, float64(2), result.Rows[0][3])
	// assert.Equal(t, "100.0", result.Rows[0][4])
	// assert.Equal(t, "100.0", result.Rows[0][5])
	// assert.Equal(t, "abc1.com", result.Rows[1][0])
	// assert.Equal(t, "abc1.com", result.Rows[1][1])
	// assert.Equal(t, float64(1), result.Rows[1][2])
	// assert.Equal(t, float64(1), result.Rows[1][3])
	// assert.Equal(t, "100.0", result.Rows[1][4])
	// assert.Equal(t, "100.0", result.Rows[1][5])
	// assert.Equal(t, "abc2.com", result.Rows[2][0])
	// assert.Equal(t, "abc2.com", result.Rows[2][1])
	// assert.Equal(t, float64(1), result.Rows[2][2])
	// assert.Equal(t, float64(1), result.Rows[2][3])
	// assert.Equal(t, "100.0", result.Rows[2][4])
	// assert.Equal(t, "100.0", result.Rows[2][5])

	// // filter test when filter are already group
	// // (groupA.A or groupA.B) AND (groupB.A or groupB.B)
	// /*
	// 	WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
	// 	FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	// 	timestamp, 1 as step_0 , MAX(group_4_id) as max_group_4_id , MAX(group_1_id) as max_group_1_id FROM
	// 	(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
	// 	events.properties as event_properties, events.user_properties as event_user_properties ,
	// 	user_groups.group_2_user_id as group_user_id , group_users.properties as group_properties ,
	// 	group_users.group_4_id as group_4_id , group_users.group_1_id as group_1_id FROM events  LEFT JOIN
	// 	users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000185' LEFT
	// 	JOIN users as group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND
	// 	group_users.project_id = '38000185' AND group_users.is_group_user = true AND group_users.source IN (
	// 	8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_1_id IS NOT NULL ) WHERE
	// 	events.project_id='38000185' AND timestamp>='1690360835' AND timestamp<='1690363235' AND
	// 	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '4db2df87-1b4a-4144-9ad8-db0e08299a7e'
	// 	)  LIMIT 10000000000) step_0_event_users_view WHERE
	// 	(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$hubspot_company_domain') =
	// 	'abc2.com' OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
	// 	'$hubspot_company_region') = 'B') OR (JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
	// 	'$6Signal_domain') = 'abc1.com' OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
	// 	'$6Signal_region') = 'B') GROUP BY coal_group_user_id HAVING max_group_4_id IS NOT NULL AND
	// 	max_group_1_id IS NOT NULL),  step_1 AS (SELECT step_1_event_users_view.group_user_id as
	// 	coal_group_user_id, step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_4_id) as
	// 	max_group_4_id , MAX(group_1_id) as max_group_1_id FROM  (SELECT events.project_id, events.id,
	// 	events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
	// 	events.user_properties as event_user_properties , user_groups.group_2_user_id as group_user_id ,
	// 	group_users.properties as group_properties , group_users.group_4_id as group_4_id ,
	// 	group_users.group_1_id as group_1_id FROM events  LEFT JOIN users as user_groups on events.user_id =
	// 	user_groups.id AND user_groups.project_id = '38000185' LEFT JOIN users as group_users ON
	// 	user_groups.group_2_user_id = group_users.group_2_user_id AND group_users.project_id = '38000185'
	// 	AND group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id
	// 	IS NOT NULL OR group_users.group_1_id IS NOT NULL ) WHERE events.project_id='38000185' AND
	// 	timestamp>='1690360835' AND timestamp<='1690363235' AND  ( group_user_id IS NOT NULL  )
	// 	AND  ( events.event_name_id = '00545d8c-de69-40e3-826a-4f784ae62244' )  LIMIT 10000000000)
	// 	step_1_event_users_view WHERE (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$hubspot_company_domain') = 'abc2.com' OR
	// 	JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_region') = 'B') OR
	// 	(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_domain') = 'abc1.com' OR
	// 	JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_region') = 'B') GROUP BY
	// 	coal_group_user_id,timestamp HAVING max_group_4_id IS NOT NULL AND max_group_1_id IS NOT NULL) ,
	// 	step_1_step_0_users AS (SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp,
	// 	FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS step_0_timestamp ,
	// 	FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN
	// 	step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
	// 	step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT
	// 	step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
	// 	WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
	// 	ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_0,
	// 	CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
	// 	(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
	// 	WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
	// 	ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) = '' THEN '$none' ELSE
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on
	// 	step_0.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id = '38000185' AND
	// 	group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
	// 	NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
	// 	step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT *
	// 	FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
	// 	AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
	// 	_group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS
	// 	_group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
	// 	AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	// */
	// query = model.Query{
	// 	From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
	// 	To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
	// 	EventsWithProperties: []model.QueryEventWithProperties{
	// 		{
	// 			Name:       "$hubspot_company_created",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 		{
	// 			Name:       "$hubspot_company_updated",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 	},
	// 	GlobalUserProperties: []model.QueryProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  "$6Signal_region",
	// 			Operator:  "equals",
	// 			Value:     "B",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			Operator:  "equals",
	// 			Value:     "abc2.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_region",
	// 			Operator:  "equals",
	// 			Value:     "B",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 	},
	// 	GroupByProperties: []model.QueryGroupByProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 	},
	// 	Class:           model.QueryClassFunnel,
	// 	GroupAnalysis:   model.GROUP_NAME_DOMAINS,
	// 	Type:            model.QueryTypeUniqueUsers,
	// 	EventsCondition: model.EventCondAllGivenEvent,
	// }

	// result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	// assert.Equal(t, http.StatusOK, errCode)

	// assert.Len(t, result.Rows, 2)
	// sort.Slice(result.Rows, func(i, j int) bool {
	// 	p1 := U.GetPropertyValueAsString(result.Rows[i][0])
	// 	p2 := U.GetPropertyValueAsString(result.Rows[j][0])
	// 	return p1 < p2
	// })

	// // filter test, when group random order of groups
	// // filter will be re ordered in (groupA.A or groupA.B) AND (groupB.A or groupB.B)
	// /*
	// 	WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
	// 	FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	// 	timestamp, 1 as step_0 , MAX(group_4_id) as max_group_4_id , MAX(group_1_id) as max_group_1_id FROM
	// 	(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
	// 	events.properties as event_properties, events.user_properties as event_user_properties ,
	// 	user_groups.group_2_user_id as group_user_id , group_users.properties as group_properties ,
	// 	group_users.group_4_id as group_4_id , group_users.group_1_id as group_1_id FROM events  LEFT JOIN
	// 	users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000186' LEFT
	// 	JOIN users as group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND
	// 	group_users.project_id = '38000186' AND group_users.is_group_user = true AND group_users.source IN (
	// 	2,8 ) AND ( group_users.group_1_id IS NOT NULL OR group_users.group_4_id IS NOT NULL ) WHERE
	// 	events.project_id='38000186' AND timestamp\u003e='1690361078' AND timestamp\u003c='1690363478' AND
	// 	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'ffdb503c-9218-42e0-8323-f83648350be7'
	// 	)  LIMIT 10000000000) step_0_event_users_view WHERE
	// 	(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = 'abc1.com' OR
	// 	JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') = 'D' OR
	// 	JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_website_2') = 'abc2.com') OR
	// 	(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$hubspot_company_domain') =
	// 	'abc1.com') GROUP BY coal_group_user_id HAVING max_group_4_id IS NOT NULL AND max_group_1_id IS NOT
	// 	NULL),  step_1 AS (SELECT step_1_event_users_view.group_user_id as coal_group_user_id,
	// 	step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_1_id) as max_group_1_id , MAX(group_4_id)
	// 	as max_group_4_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	// 	events.timestamp , events.properties as event_properties, events.user_properties as
	// 	event_user_properties , user_groups.group_2_user_id as group_user_id , group_users.properties as
	// 	group_properties , group_users.group_1_id as group_1_id , group_users.group_4_id as group_4_id FROM
	// 	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	// 	= '38000186' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
	// 	group_users.group_2_user_id AND group_users.project_id = '38000186' AND group_users.is_group_user =
	// 	true AND group_users.source IN ( 2,8 ) AND ( group_users.group_1_id IS NOT NULL OR
	// 	group_users.group_4_id IS NOT NULL ) WHERE events.project_id='38000186' AND
	// 	timestamp\u003e='1690361078' AND timestamp\u003c='1690363478' AND  ( group_user_id IS NOT NULL  )
	// 	AND  ( events.event_name_id = '68d10781-5e9b-4590-b4b6-bb8ca9a4e915' )  LIMIT 10000000000)
	// 	step_1_event_users_view WHERE (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_domain') = 'abc1.com' OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_region') = 'D' OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_website_2') = 'abc2.com') OR
	// 	(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_domain') =
	// 	'abc1.com') GROUP BY coal_group_user_id,timestamp HAVING max_group_1_id IS NOT NULL AND
	// 	max_group_4_id IS NOT NULL) , step_1_step_0_users AS (SELECT step_1.coal_group_user_id,
	// 	FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS
	// 	step_0_timestamp , FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM
	// 	step_0 LEFT JOIN step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE
	// 	step_1.timestamp \u003e= step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT
	// 	DISTINCT step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE
	// 	WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
	// 	WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
	// 	ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_0,
	// 	CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
	// 	(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
	// 	WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
	// 	ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) = '' THEN '$none' ELSE
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on
	// 	step_0.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id = '38000186' AND
	// 	group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
	// 	NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
	// 	step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) \u003c= '90' ) SELECT *
	// 	FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
	// 	AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
	// 	_group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS
	// 	_group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
	// 	AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	// */
	// query = model.Query{
	// 	From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
	// 	To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
	// 	EventsWithProperties: []model.QueryEventWithProperties{
	// 		{
	// 			Name:       "$hubspot_company_created",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 		{
	// 			Name:       "$hubspot_company_updated",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 	},
	// 	GlobalUserProperties: []model.QueryProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  "$6Signal_region",
	// 			Operator:  "equals",
	// 			Value:     "D",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  "$6Signal_website_2",
	// 			Operator:  "equals",
	// 			Value:     "abc2.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 	},
	// 	GroupByProperties: []model.QueryGroupByProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 	},
	// 	Class:           model.QueryClassFunnel,
	// 	GroupAnalysis:   model.GROUP_NAME_DOMAINS,
	// 	Type:            model.QueryTypeUniqueUsers,
	// 	EventsCondition: model.EventCondAllGivenEvent,
	// }

	// result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	// assert.Equal(t, http.StatusOK, errCode)

	// assert.Len(t, result.Rows, 2)
	// sort.Slice(result.Rows, func(i, j int) bool {
	// 	p1 := U.GetPropertyValueAsString(result.Rows[i][0])
	// 	p2 := U.GetPropertyValueAsString(result.Rows[j][0])
	// 	return p1 < p2
	// })

	// assert.Equal(t, "$no_group", result.Rows[0][0])
	// assert.Equal(t, "$no_group", result.Rows[0][1])
	// assert.Equal(t, float64(1), result.Rows[0][2])
	// assert.Equal(t, float64(1), result.Rows[0][3])
	// assert.Equal(t, "100.0", result.Rows[0][4])
	// assert.Equal(t, "100.0", result.Rows[0][5])
	// assert.Equal(t, "abc1.com", result.Rows[1][0])
	// assert.Equal(t, "abc1.com", result.Rows[1][1])
	// assert.Equal(t, float64(1), result.Rows[1][2])
	// assert.Equal(t, float64(1), result.Rows[1][3])
	// assert.Equal(t, "100.0", result.Rows[1][4])
	// assert.Equal(t, "100.0", result.Rows[1][5])

	// // event level breakdown test
	// /*
	// 	WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
	// 	FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	// 	timestamp, 1 as step_0 , CASE WHEN
	// 	JSON_EXTRACT_STRING(FIRST(step_0_event_users_view.event_user_properties,
	// 	FROM_UNIXTIME(step_0_event_users_view.timestamp)), 'hs_company_no') IS NULL THEN '$none' WHEN
	// 	JSON_EXTRACT_STRING(FIRST(step_0_event_users_view.event_user_properties,
	// 	FROM_UNIXTIME(step_0_event_users_view.timestamp)), 'hs_company_no') = '' THEN '$none' ELSE
	// 	JSON_EXTRACT_STRING(FIRST(step_0_event_users_view.event_user_properties,
	// 	FROM_UNIXTIME(step_0_event_users_view.timestamp)), 'hs_company_no') END AS _group_key_0 ,
	// 	MAX(group_4_id) as max_group_4_id , MAX(group_1_id) as max_group_1_id FROM  (SELECT
	// 	events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
	// 	events.properties as event_properties, events.user_properties as event_user_properties ,
	// 	user_groups.group_2_user_id as group_user_id , group_users.properties as group_properties ,
	// 	group_users.group_4_id as group_4_id , group_users.group_1_id as group_1_id FROM events  LEFT JOIN
	// 	users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000187' LEFT
	// 	JOIN users as group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND
	// 	group_users.project_id = '38000187' AND group_users.is_group_user = true AND group_users.source IN (
	// 	8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_1_id IS NOT NULL ) WHERE
	// 	events.project_id='38000187' AND timestamp\u003e='1690361328' AND timestamp\u003c='1690363728' AND
	// 	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'd7375648-bc87-47b8-8827-2cdc1d812ba8'
	// 	)  LIMIT 10000000000) step_0_event_users_view WHERE
	// 	(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = 'abc1.com' OR
	// 	JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') = 'D' OR
	// 	JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_website_2') = 'abc2.com') OR
	// 	(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$hubspot_company_domain') =
	// 	'abc1.com') GROUP BY coal_group_user_id HAVING max_group_4_id IS NOT NULL AND max_group_1_id IS NOT
	// 	NULL),  step_1 AS (SELECT step_1_event_users_view.group_user_id as coal_group_user_id,
	// 	step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_4_id) as max_group_4_id , MAX(group_1_id)
	// 	as max_group_1_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	// 	events.timestamp , events.properties as event_properties, events.user_properties as
	// 	event_user_properties , user_groups.group_2_user_id as group_user_id , group_users.properties as
	// 	group_properties , group_users.group_4_id as group_4_id , group_users.group_1_id as group_1_id FROM
	// 	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	// 	= '38000187' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
	// 	group_users.group_2_user_id AND group_users.project_id = '38000187' AND group_users.is_group_user =
	// 	true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS NOT NULL OR
	// 	group_users.group_1_id IS NOT NULL ) WHERE events.project_id='38000187' AND
	// 	timestamp\u003e='1690361328' AND timestamp\u003c='1690363728' AND  ( group_user_id IS NOT NULL  )
	// 	AND  ( events.event_name_id = '049b3232-6fab-4b93-a246-0391212da67c' )  LIMIT 10000000000)
	// 	step_1_event_users_view WHERE (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_domain') = 'abc1.com' OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_region') = 'D' OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_website_2') = 'abc2.com') OR
	// 	(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_domain') =
	// 	'abc1.com') GROUP BY coal_group_user_id,timestamp HAVING max_group_4_id IS NOT NULL AND
	// 	max_group_1_id IS NOT NULL) , step_1_step_0_users AS (SELECT step_1.coal_group_user_id,
	// 	FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS
	// 	step_0_timestamp , FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM
	// 	step_0 LEFT JOIN step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE
	// 	step_1.timestamp \u003e= step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT
	// 	DISTINCT step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE
	// 	WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none'
	// 	WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none'
	// 	ELSE FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$6Signal_domain') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_1,
	// 	CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER
	// 	(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') IS NULL THEN 1000000000000
	// 	WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000
	// 	ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) = '' THEN '$none' ELSE
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain')) OVER (PARTITION
	// 	BY step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_domain') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_domain') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) END AS _group_key_2 ,  CASE WHEN _group_key_0 IS NULL THEN '$none'
	// 	WHEN _group_key_0 = '' THEN '$none' ELSE _group_key_0 END AS _group_key_0 FROM step_0  LEFT JOIN
	// 	users AS group_users on step_0.coal_group_user_id = group_users.group_2_user_id AND
	// 	group_users.project_id = '38000187' AND  group_users.is_group_user = true AND group_users.source IN
	// 	( 8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_1_id IS NOT NULL )  LEFT JOIN
	// 	step_1_step_0_users ON step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND
	// 	timestampdiff(DAY, DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) \u003c= '90' ) SELECT *
	// 	FROM ( SELECT _group_key_0, _group_key_1, _group_key_2, SUM(step_0) AS step_0 , SUM(step_1) AS
	// 	step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
	// 	_group_key_1, _group_key_2 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT
	// 	'$no_group' AS _group_key_0,'$no_group' AS _group_key_1,'$no_group' AS _group_key_2 , SUM(step_0) AS
	// 	step_0 , SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	// */
	// query = model.Query{
	// 	From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
	// 	To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
	// 	EventsWithProperties: []model.QueryEventWithProperties{
	// 		{
	// 			Name:       "$hubspot_company_created",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 		{
	// 			Name:       "$hubspot_company_updated",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 	},
	// 	GlobalUserProperties: []model.QueryProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  "$6Signal_region",
	// 			Operator:  "equals",
	// 			Value:     "D",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  "$6Signal_website_2",
	// 			Operator:  "equals",
	// 			Value:     "abc2.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "OR",
	// 		},
	// 	},
	// 	GroupByProperties: []model.QueryGroupByProperty{
	// 		{
	// 			Entity:         model.PropertyEntityUser,
	// 			Property:       "hs_company_no",
	// 			EventName:      "$hubspot_company_created",
	// 			EventNameIndex: 1,
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:  "$hubspot_company_domain",
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 	},
	// 	Class:           model.QueryClassFunnel,
	// 	GroupAnalysis:   model.GROUP_NAME_DOMAINS,
	// 	Type:            model.QueryTypeUniqueUsers,
	// 	EventsCondition: model.EventCondAllGivenEvent,
	// }

	// result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	// assert.Equal(t, http.StatusOK, errCode)
	// assert.Len(t, result.Rows, 2)

	// assert.Equal(t, "$no_group", result.Rows[0][0])
	// assert.Equal(t, "$no_group", result.Rows[0][1])
	// assert.Equal(t, "$no_group", result.Rows[0][2])
	// assert.Equal(t, float64(1), result.Rows[0][3])
	// assert.Equal(t, float64(1), result.Rows[0][4])
	// assert.Equal(t, "100.0", result.Rows[0][5])
	// assert.Equal(t, "100.0", result.Rows[0][6])
	// assert.Equal(t, "h1", result.Rows[1][0])
	// assert.Equal(t, "abc1.com", result.Rows[1][1])
	// assert.Equal(t, "abc1.com", result.Rows[1][2])
	// assert.Equal(t, float64(1), result.Rows[1][3])
	// assert.Equal(t, float64(1), result.Rows[1][4])
	// assert.Equal(t, "100.0", result.Rows[1][5])
	// assert.Equal(t, "100.0", result.Rows[1][6])

	// // event level breakdown and global datetime breakdown test
	// /*
	// 	WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
	// 	FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	// 	timestamp, 1 as step_0 , CASE WHEN
	// 	JSON_EXTRACT_STRING(FIRST(step_0_event_users_view.event_user_properties,
	// 	FROM_UNIXTIME(step_0_event_users_view.timestamp)), 'hs_company_no') IS NULL THEN '$none' WHEN
	// 	JSON_EXTRACT_STRING(FIRST(step_0_event_users_view.event_user_properties,
	// 	FROM_UNIXTIME(step_0_event_users_view.timestamp)), 'hs_company_no') = '' THEN '$none' ELSE
	// 	JSON_EXTRACT_STRING(FIRST(step_0_event_users_view.event_user_properties,
	// 	FROM_UNIXTIME(step_0_event_users_view.timestamp)), 'hs_company_no') END AS _group_key_1 FROM
	// 	(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
	// 	events.properties as event_properties, events.user_properties as event_user_properties ,
	// 	user_groups.group_2_user_id as group_user_id , group_users.properties as group_properties FROM
	// 	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	// 	= '38000178' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
	// 	group_users.group_2_user_id AND group_users.project_id = '38000178' AND group_users.is_group_user =
	// 	true AND group_users.source IN ( 8 ) AND ( group_users.group_4_id IS NOT NULL ) WHERE
	// 	events.project_id='38000178' AND timestamp>='1690358641' AND timestamp<='1690361041' AND
	// 	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'cfcda607-9704-4e70-9bc5-9a224dc89d3e'
	// 	)  LIMIT 10000000000) step_0_event_users_view WHERE
	// 	(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = 'abc1.com')
	// 	GROUP BY coal_group_user_id),  step_1 AS (SELECT step_1_event_users_view.group_user_id as
	// 	coal_group_user_id, step_1_event_users_view.timestamp, 1 as step_1 FROM  (SELECT events.project_id,
	// 	events.id, events.event_name_id, events.user_id, events.timestamp , events.properties as
	// 	event_properties, events.user_properties as event_user_properties , user_groups.group_2_user_id as
	// 	group_user_id , group_users.properties as group_properties FROM events  LEFT JOIN users as
	// 	user_groups on events.user_id = user_groups.id AND user_groups.project_id = '38000178' LEFT JOIN
	// 	users as group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND
	// 	group_users.project_id = '38000178' AND group_users.is_group_user = true AND group_users.source IN (
	// 	8 ) AND ( group_users.group_4_id IS NOT NULL ) WHERE events.project_id='38000178' AND
	// 	timestamp>='1690358641' AND timestamp<='1690361041' AND  ( group_user_id IS NOT NULL  )
	// 	AND  ( events.event_name_id = '5548335f-e5a8-4429-962b-28c4b97ca417' )  LIMIT 10000000000)
	// 	step_1_event_users_view WHERE (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
	// 	'$6Signal_domain') = 'abc1.com') GROUP BY coal_group_user_id,timestamp) , step_1_step_0_users AS
	// 	(SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as
	// 	timestamp, step_1 , step_0.timestamp AS step_0_timestamp , FIRST(step_1.timestamp,
	// 	FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN step_1 ON
	// 	step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
	// 	step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT
	// 	step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp , CASE WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate')) OVER
	// 	(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate') IS NULL THEN
	// 	1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate') = ''
	// 	THEN 1000000000000 ELSE group_users.join_timestamp END) IS NULL THEN '$none' WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate')) OVER
	// 	(PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate') IS NULL THEN
	// 	1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate') = ''
	// 	THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none' ELSE date_trunc('day',
	// 	CONVERT_TZ(FROM_UNIXTIME(CONVERT(SUBSTRING(FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties,
	// 	'$hubspot_company_createddate')) OVER (PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate') IS NULL THEN
	// 	1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_createddate') = ''
	// 	THEN 1000000000000 ELSE group_users.join_timestamp END),1,10), DECIMAL(10))), 'UTC', 'UTC')) END AS
	// 	_group_key_0, CASE WHEN FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, 'sf_account_no'))
	// 	OVER (PARTITION BY step_0.coal_group_user_id ORDER BY CASE WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, 'sf_account_no') IS NULL THEN 1000000000000 WHEN
	// 	JSON_EXTRACT_STRING(group_users.properties, 'sf_account_no') = '' THEN 1000000000000 ELSE
	// 	group_users.join_timestamp END) IS NULL THEN '$none' WHEN
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, 'sf_account_no')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'sf_account_no') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'sf_account_no') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) = '' THEN '$none' ELSE
	// 	FIRST_VALUE(JSON_EXTRACT_STRING(group_users.properties, 'sf_account_no')) OVER (PARTITION BY
	// 	step_0.coal_group_user_id ORDER BY CASE WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'sf_account_no') IS NULL THEN 1000000000000 WHEN JSON_EXTRACT_STRING(group_users.properties,
	// 	'sf_account_no') = '' THEN 1000000000000 ELSE group_users.join_timestamp END) END AS _group_key_2 ,
	// 	CASE WHEN _group_key_1 IS NULL THEN '$none' WHEN _group_key_1 = '' THEN '$none' ELSE _group_key_1
	// 	END AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on step_0.coal_group_user_id =
	// 	group_users.group_2_user_id AND group_users.project_id = '38000178' AND  group_users.is_group_user =
	// 	true AND group_users.source IN ( 2,3 ) AND ( group_users.group_1_id IS NOT NULL OR
	// 	group_users.group_3_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
	// 	step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
	// 	DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90' ) SELECT *
	// 	FROM ( SELECT _group_key_0, _group_key_1, _group_key_2, SUM(step_0) AS step_0 , SUM(step_1) AS
	// 	step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP BY _group_key_0,
	// 	_group_key_1, _group_key_2 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT
	// 	'$no_group' AS _group_key_0,'$no_group' AS _group_key_1,'$no_group' AS _group_key_2 , SUM(step_0) AS
	// 	step_0 , SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	// */
	// query = model.Query{
	// 	From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
	// 	To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
	// 	EventsWithProperties: []model.QueryEventWithProperties{
	// 		{
	// 			Name:       "$hubspot_company_created",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 		{
	// 			Name:       "$hubspot_company_updated",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 	},
	// 	GlobalUserProperties: []model.QueryProperty{
	// 		{
	// 			Entity:    model.PropertyEntityUserGlobal,
	// 			GroupName: model.GROUP_NAME_SIX_SIGNAL,
	// 			Property:  U.SIX_SIGNAL_DOMAIN,
	// 			Operator:  "equals",
	// 			Value:     "abc1.com",
	// 			Type:      U.PropertyTypeCategorical,
	// 			LogicalOp: "AND",
	// 		},
	// 	},
	// 	GroupByProperties: []model.QueryGroupByProperty{
	// 		{
	// 			Entity:      model.PropertyEntityUser,
	// 			GroupName:   model.GROUP_NAME_HUBSPOT_COMPANY,
	// 			Property:    "$hubspot_company_createddate",
	// 			EventName:   model.UserPropertyGroupByPresent,
	// 			Type:        U.PropertyTypeDateTime,
	// 			Granularity: U.DateTimeBreakdownDailyGranularity,
	// 		},
	// 		{
	// 			Entity:         model.PropertyEntityUser,
	// 			Property:       "hs_company_no",
	// 			EventName:      "$hubspot_company_created",
	// 			EventNameIndex: 1,
	// 		},
	// 		{
	// 			Entity:    model.PropertyEntityUser,
	// 			GroupName: model.GROUP_NAME_SALESFORCE_ACCOUNT,
	// 			Property:  "sf_account_no",
	// 			EventName: model.UserPropertyGroupByPresent,
	// 		},
	// 	},
	// 	Class:           model.QueryClassFunnel,
	// 	GroupAnalysis:   model.GROUP_NAME_DOMAINS,
	// 	Type:            model.QueryTypeUniqueUsers,
	// 	EventsCondition: model.EventCondAllGivenEvent,
	// }

	// result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	// assert.Equal(t, http.StatusOK, errCode)
	// assert.Len(t, result.Rows, 2)

	// assert.Equal(t, "$no_group", result.Rows[0][0])
	// assert.Equal(t, "$no_group", result.Rows[0][1])
	// assert.Equal(t, "$no_group", result.Rows[0][2])
	// assert.Equal(t, float64(1), result.Rows[0][3])
	// assert.Equal(t, float64(1), result.Rows[0][4])
	// assert.Equal(t, "100.0", result.Rows[0][5])
	// assert.Equal(t, "100.0", result.Rows[0][6])

	// assert.Equal(t, util.GetTimestampAsStrWithTimezone(now.New(dateTimeUTC).BeginningOfDay(), "UTC"), result.Rows[1][0])
	// assert.Equal(t, "h1", result.Rows[1][1])
	// assert.Equal(t, "s123", result.Rows[1][2])
	// assert.Equal(t, float64(1), result.Rows[1][3])
	// assert.Equal(t, float64(1), result.Rows[1][4])
	// assert.Equal(t, "100.0", result.Rows[1][5])
	// assert.Equal(t, "100.0", result.Rows[1][6])

	// // Test using API
	// enQuery, err := json.Marshal(query)
	// assert.Nil(t, err)
	// queryPJSON := postgres.Jsonb{json.RawMessage(enQuery)}
	// baseQuery, err := model.DecodeQueryForClass(queryPJSON, model.QueryClassFunnel)
	// assert.Nil(t, err)
	// w = sendAnalyticsQueryReq(r, model.QueryClassFunnel, project.ID, agent, 0, 0, "", baseQuery, false, false)
	// assert.NotEmpty(t, w)
	// assert.Equal(t, http.StatusOK, w.Code)

	// jsonResponse, err := ioutil.ReadAll(w.Body)
	// assert.Nil(t, err)
	// var querResult model.QueryResult
	// err = json.Unmarshal(jsonResponse, &querResult)
	// assert.Nil(t, err)
	// assert.Equal(t, http.StatusOK, w.Code)
	// assert.Len(t, querResult.Rows, 2)

	// assert.Equal(t, "$no_group", querResult.Rows[0][0])
	// assert.Equal(t, "$no_group", querResult.Rows[0][1])
	// assert.Equal(t, "$no_group", querResult.Rows[0][2])
	// assert.Equal(t, float64(1), querResult.Rows[0][3])
	// assert.Equal(t, float64(1), querResult.Rows[0][4])
	// assert.Equal(t, "100.0", querResult.Rows[0][5])
	// assert.Equal(t, "100.0", querResult.Rows[0][6])

	// assert.Equal(t, util.GetTimestampAsStrWithTimezone(now.New(dateTimeUTC).BeginningOfDay(), string(U.TimeZoneStringIST)), querResult.Rows[1][0])
	// assert.Equal(t, "h1", querResult.Rows[1][1])
	// assert.Equal(t, "s123", querResult.Rows[1][2])
	// assert.Equal(t, float64(1), querResult.Rows[1][3])
	// assert.Equal(t, float64(1), querResult.Rows[1][4])
	// assert.Equal(t, "100.0", querResult.Rows[1][5])
	// assert.Equal(t, "100.0", querResult.Rows[1][6])

	// // Queries without groupName should not affect for other scope groups
	// query = model.Query{
	// 	From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
	// 	To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
	// 	EventsWithProperties: []model.QueryEventWithProperties{
	// 		{
	// 			Name:       "www.xyz.com",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 		{
	// 			Name:       "$hubspot_company_created",
	// 			Properties: []model.QueryProperty{},
	// 		},
	// 	},
	// 	GroupByProperties: []model.QueryGroupByProperty{
	// 		{
	// 			Entity:      model.PropertyEntityUser,
	// 			Property:    "$hubspot_company_createddate",
	// 			EventName:   model.UserPropertyGroupByPresent,
	// 			Type:        U.PropertyTypeDateTime,
	// 			Granularity: U.DateTimeBreakdownDailyGranularity,
	// 		},
	// 	},
	// 	Class:           model.QueryClassFunnel,
	// 	GroupAnalysis:   model.GROUP_NAME_HUBSPOT_COMPANY,
	// 	Type:            model.QueryTypeUniqueUsers,
	// 	EventsCondition: model.EventCondAllGivenEvent,
	// }

	// enQuery, err = json.Marshal(query)
	// assert.Nil(t, err)
	// queryPJSON = postgres.Jsonb{json.RawMessage(enQuery)}
	// baseQuery, err = model.DecodeQueryForClass(queryPJSON, model.QueryClassFunnel)
	// assert.Nil(t, err)
	// w = sendAnalyticsQueryReq(r, model.QueryClassFunnel, project.ID, agent, 0, 0, "", baseQuery, false, false)
	// assert.NotEmpty(t, w)
	// assert.Equal(t, http.StatusOK, w.Code)

	// jsonResponse, err = ioutil.ReadAll(w.Body)
	// assert.Nil(t, err)
	// err = json.Unmarshal(jsonResponse, &querResult)
	// assert.Nil(t, err)
	// assert.Equal(t, http.StatusOK, w.Code)
	// assert.Len(t, querResult.Rows, 2)

	// assert.Equal(t, "$no_group", querResult.Rows[0][0])
	// assert.Equal(t, float64(1), querResult.Rows[0][1])
	// assert.Equal(t, float64(1), querResult.Rows[0][2])
	// assert.Equal(t, "100.0", querResult.Rows[0][3])
	// assert.Equal(t, "100.0", querResult.Rows[0][4])
	// assert.Equal(t, util.GetTimestampAsStrWithTimezone(now.New(dateTimeUTC).BeginningOfDay(), string(U.TimeZoneStringIST)), querResult.Rows[1][0])
	// assert.Equal(t, float64(1), querResult.Rows[1][1])
	// assert.Equal(t, float64(1), querResult.Rows[1][2])
	// assert.Equal(t, "100.0", querResult.Rows[1][3])
	// assert.Equal(t, "100.0", querResult.Rows[1][4])
}

func TestAnalyticsEventsAllAccounts(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"
	/*
		userWeb1(domain: abc1.com) - event(xyz.com)
		userWeb2(domain: abc2.com) - event(xyz2.com)
		userWeb3(domain: abc3.com) - event(xyz.com)

		groupUserHubspot1(domain: abc1.com) - event(hubspot_company_created, hubspot_company_update)
		groupUserHubspot2(domain: abc2.com) - event(hubspot_company_created, hubspot_company_update)

		groupUserSalesforce1(domain: abc1.com) - event(salesforce_account_created, salesforce_account_updated)
		groupUserSalesforce2(domain: abc2.com) - event(salesforce_account_created, salesforce_account_updated)
	*/
	properties := postgres.Jsonb{[]byte(`{"user_no":"w1"}`)}
	userWeb1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w2"}`)}
	userWeb2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid2"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w3"}`)}
	userWeb3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid3"})
	assert.Equal(t, http.StatusCreated, errCode)

	dateTimeUTC := util.TimeNowZ()
	propertiesMap := U.PropertiesMap{"$hubspot_company_name": "abc1", "$hubspot_company_domain": "abc1.com", "$hubspot_company_region": "A", "hs_company_no": "h1", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	groupUserHubspot1, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc1.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	propertiesMap = U.PropertiesMap{"$hubspot_company_name": "abc2", "$hubspot_company_domain": "abc2.com", "$hubspot_company_region": "B", "hs_company_no": "h2", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	groupUserHubspot2, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc2.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	properties = postgres.Jsonb{[]byte(`{"user_no":"h1"}`)}
	userHubspot1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceHubspot), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)
	_, errCode = store.GetStore().UpdateUserGroup(project.ID, userHubspot1, model.GROUP_NAME_HUBSPOT_COMPANY, "habc", groupUserHubspot1, false)
	assert.Equal(t, http.StatusAccepted, errCode)

	propertiesMap = U.PropertiesMap{"sf_account_no": "s123"}
	groupUserSalesforce1, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, "abc1.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	propertiesMap = U.PropertiesMap{"sf_account_no": "s234"}
	groupUserSalesforce2, status := SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, "abc2.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties := &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc1.com", U.SIX_SIGNAL_REGION: "A"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb1, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc2.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb2, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc3.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb3, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	status = store.GetStore().AssociateUserDomainsGroup(project.ID, userWeb1, "", "")
	assert.Equal(t, http.StatusOK, status)
	status = store.GetStore().AssociateUserDomainsGroup(project.ID, userWeb2, "", "")
	assert.Equal(t, http.StatusOK, status)

	payload := fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb3, U.TimeNowZ().Add(-10*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 3)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 4)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_updated", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 5)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 6)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$hubspot_company_updated", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserHubspot2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 7)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$salesforce_account_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserSalesforce1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 8)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "$salesforce_account_created", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		groupUserSalesforce2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 9)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	/* Unique accounts who performed all given events.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	events.timestamp , events.properties as event_properties, events.user_properties as
	event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
	user_groups on events.user_id = user_groups.id AND user_groups.project_id = '45000003'  WHERE
	events.project_id='45000003' AND timestamp\u003e='1692260832' AND timestamp\u003c='1692263232' AND
	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'bb0cba39-fec7-49dc-9ccb-309f29e3e9de'
	)  LIMIT 10000000000) step_0_event_users_view GROUP BY coal_group_user_id) SELECT
	COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM step_0 ORDER BY aggregate DESC LIMIT 100000
	*/
	query := model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, float64(2), result.Rows[0][0])

	/* Unique accounts who performed all events.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	events.timestamp , events.properties as event_properties, events.user_properties as
	event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
	user_groups on events.user_id = user_groups.id AND user_groups.project_id = '45000004'  WHERE
	events.project_id='45000004' AND timestamp\u003e='1692261119' AND timestamp\u003c='1692263519' AND
	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '360c6137-ae65-46db-a7c6-b66a4c72a456'
	)  LIMIT 10000000000) step_0_event_users_view GROUP BY coal_group_user_id), step_1 AS (SELECT
	step_1_event_users_view.group_user_id  as coal_group_user_id, FIRST(step_1_event_users_view.user_id,
	FROM_UNIXTIME(step_1_event_users_view.timestamp)) as event_user_id FROM  (SELECT events.project_id,
	events.id, events.event_name_id, events.user_id, events.timestamp , events.properties as
	event_properties, events.user_properties as event_user_properties , user_groups.group_2_user_id as
	group_user_id FROM events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND
	user_groups.project_id = '45000004'  WHERE events.project_id='45000004' AND
	timestamp\u003e='1692261119' AND timestamp\u003c='1692263519' AND  ( group_user_id IS NOT NULL  )
	AND  ( events.event_name_id = '8370eb09-c526-4e8a-8dc0-c97c09ba078d' )  LIMIT 10000000000)
	step_1_event_users_view GROUP BY coal_group_user_id) , events_intersect AS (SELECT
	step_0.event_user_id as event_user_id, step_0.coal_group_user_id as coal_group_user_id FROM step_0
	JOIN step_1 ON step_1.coal_group_user_id = step_0.coal_group_user_id) SELECT
	COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM events_intersect ORDER BY aggregate DESC LIMIT
	100000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, float64(0), result.Rows[0][0])

	/* Unique accounts who performed any events.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	events.timestamp , events.properties as event_properties, events.user_properties as
	event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
	user_groups on events.user_id = user_groups.id AND user_groups.project_id = '45000005'  WHERE
	events.project_id='45000005' AND timestamp\u003e='1692261327' AND timestamp\u003c='1692263727' AND
	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '03d0356c-5b24-4d75-8472-305d861ba438'
	)  LIMIT 10000000000) step_0_event_users_view GROUP BY coal_group_user_id), step_1 AS (SELECT
	step_1_event_users_view.group_user_id  as coal_group_user_id, FIRST(step_1_event_users_view.user_id,
	FROM_UNIXTIME(step_1_event_users_view.timestamp)) as event_user_id FROM  (SELECT events.project_id,
	events.id, events.event_name_id, events.user_id, events.timestamp , events.properties as
	event_properties, events.user_properties as event_user_properties , user_groups.group_2_user_id as
	group_user_id FROM events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND
	user_groups.project_id = '45000005'  WHERE events.project_id='45000005' AND
	timestamp\u003e='1692261327' AND timestamp\u003c='1692263727' AND  ( group_user_id IS NOT NULL  )
	AND  ( events.event_name_id = '84809b4f-dec8-4d5d-a641-ec86ba6793fc' )  LIMIT 10000000000)
	step_1_event_users_view GROUP BY coal_group_user_id) , events_union AS (SELECT step_0.event_user_id
	as event_user_id, step_0.coal_group_user_id as coal_group_user_id FROM step_0 UNION ALL SELECT
	step_1.event_user_id as event_user_id, step_1.coal_group_user_id as coal_group_user_id FROM step_1)
	SELECT COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM events_union ORDER BY aggregate DESC
	LIMIT 100000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAnyGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, float64(3), result.Rows[0][0])

	/* Unique accounts who performed all events.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
	events.timestamp , events.properties as event_properties, events.user_properties as
	event_user_properties , user_groups.group_2_user_id as group_user_id FROM events  LEFT JOIN users as
	user_groups on events.user_id = user_groups.id AND user_groups.project_id = '45000006'  WHERE
	events.project_id='45000006' AND timestamp\u003e='1692261480' AND timestamp\u003c='1692263880' AND
	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '2d820d12-7508-4539-a179-00bfeab009a8'
	)  LIMIT 10000000000) step_0_event_users_view GROUP BY coal_group_user_id), step_1 AS (SELECT
	step_1_event_users_view.group_user_id  as coal_group_user_id, FIRST(step_1_event_users_view.user_id,
	FROM_UNIXTIME(step_1_event_users_view.timestamp)) as event_user_id FROM  (SELECT events.project_id,
	events.id, events.event_name_id, events.user_id, events.timestamp , events.properties as
	event_properties, events.user_properties as event_user_properties , user_groups.group_2_user_id as
	group_user_id FROM events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND
	user_groups.project_id = '45000006'  WHERE events.project_id='45000006' AND
	timestamp\u003e='1692261480' AND timestamp\u003c='1692263880' AND  ( group_user_id IS NOT NULL  )
	AND  ( events.event_name_id = 'e7ce7df3-852f-4737-91d0-5efcdd63f7f9' )  LIMIT 10000000000)
	step_1_event_users_view GROUP BY coal_group_user_id) , events_intersect AS (SELECT
	step_0.event_user_id as event_user_id, step_0.coal_group_user_id as coal_group_user_id FROM step_0
	JOIN step_1 ON step_1.coal_group_user_id = step_0.coal_group_user_id) SELECT
	COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM events_intersect ORDER BY aggregate DESC LIMIT
	100000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "$hubspot_company_created",
				Properties: []model.QueryProperty{},
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, float64(1), result.Rows[0][0])

	/* Unique accounts who performed each events.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id , '0_www.xyz.com' AS event_name  FROM  (SELECT events.project_id, events.id,
	events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
	events.user_properties as event_user_properties , user_groups.group_2_user_id as group_user_id FROM
	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	= '45000007'  WHERE events.project_id='45000007' AND timestamp\u003e='1692261648' AND
	timestamp\u003c='1692264048' AND  ( group_user_id IS NOT NULL  ) AND  ( events.event_name_id =
	'4bb06821-42a4-48a4-86fc-bba7ce88ec57' )  LIMIT 10000000000) step_0_event_users_view GROUP BY
	coal_group_user_id), step_1 AS (SELECT step_1_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_1_event_users_view.user_id, FROM_UNIXTIME(step_1_event_users_view.timestamp)) as
	event_user_id , '1_www.xyz2.com' AS event_name  FROM  (SELECT events.project_id, events.id,
	events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
	events.user_properties as event_user_properties , user_groups.group_2_user_id as group_user_id FROM
	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	= '45000007'  WHERE events.project_id='45000007' AND timestamp\u003e='1692261648' AND
	timestamp\u003c='1692264048' AND  ( group_user_id IS NOT NULL  ) AND  ( events.event_name_id =
	'c4b40fb6-838a-4e70-b9af-d96863e2e34d' )  LIMIT 10000000000) step_1_event_users_view GROUP BY
	coal_group_user_id) , each_events_union AS (SELECT step_0.event_name as event_name,
	step_0.coal_group_user_id as coal_group_user_id, step_0.event_user_id as event_user_id FROM step_0
	UNION ALL SELECT step_1.event_name as event_name, step_1.coal_group_user_id as coal_group_user_id,
	step_1.event_user_id as event_user_id FROM step_1) SELECT event_name,
	COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM each_events_union GROUP BY event_name ORDER BY
	aggregate DESC LIMIT 100000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, float64(2), result.Rows[0][2])
	assert.Equal(t, "www.xyz2.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])

	/* Unique accounts who performed each events with breakdown.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id , '0_www.xyz.com' AS event_name  FROM  (SELECT events.project_id, events.id,
	events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
	events.user_properties as event_user_properties , user_groups.group_2_user_id as group_user_id FROM
	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	= '45000008'  WHERE events.project_id='45000008' AND timestamp\u003e='1692261853' AND
	timestamp\u003c='1692264253' AND  ( group_user_id IS NOT NULL  ) AND  ( events.event_name_id =
	'd1114298-de73-4b12-9694-6de5aa8f39ad' )  LIMIT 10000000000) step_0_event_users_view GROUP BY
	coal_group_user_id ORDER BY coal_group_user_id, step_0_event_users_view.timestamp ASC), step_1 AS
	(SELECT step_1_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_1_event_users_view.user_id, FROM_UNIXTIME(step_1_event_users_view.timestamp)) as
	event_user_id , '1_www.xyz2.com' AS event_name  FROM  (SELECT events.project_id, events.id,
	events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
	events.user_properties as event_user_properties , user_groups.group_2_user_id as group_user_id FROM
	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	= '45000008'  WHERE events.project_id='45000008' AND timestamp\u003e='1692261853' AND
	timestamp\u003c='1692264253' AND  ( group_user_id IS NOT NULL  ) AND  ( events.event_name_id =
	'c98aad90-ef41-4b34-b329-543bb809a6b0' )  LIMIT 10000000000) step_1_event_users_view GROUP BY
	coal_group_user_id ORDER BY coal_group_user_id, step_1_event_users_view.timestamp ASC) ,
	each_events_union AS (SELECT step_0.event_name as event_name, step_0.coal_group_user_id as
	coal_group_user_id, step_0.event_user_id as event_user_id FROM step_0 UNION ALL SELECT
	step_1.event_name as event_name, step_1.coal_group_user_id as coal_group_user_id,
	step_1.event_user_id as event_user_id FROM step_1) , each_users_union AS (SELECT
	each_events_union.event_user_id, each_events_union.coal_group_user_id,
	each_events_union.event_name, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
	'$hubspot_company_name') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
	'$hubspot_company_name') = '' then '$none' else CONCAT( join_timestamp, ':',
	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end), LOCATE(':', max( case
	when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null then '$none' when
	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then '$none' else CONCAT(
	join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end
	))+1) AS _group_key_0 FROM each_events_union  LEFT JOIN users AS group_users on
	each_events_union.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id =
	'45000008' AND  group_users.is_group_user = true AND group_users.source IN ( 2 ) AND (
	group_users.group_1_id IS NOT NULL ) GROUP BY each_events_union.coal_group_user_id) SELECT
	event_name, _group_key_0, COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM each_users_union
	GROUP BY event_name , _group_key_0 ORDER BY aggregate DESC LIMIT 10000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Len(t, result.Rows, 3)
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, "$none", result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "www.xyz.com", result.Rows[1][1])
	assert.Equal(t, "abc1", result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "www.xyz2.com", result.Rows[2][1])
	assert.Equal(t, "abc2", result.Rows[2][2])

	// Unique accounts who performed each event with multiple group breakdown.
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

	assert.Len(t, result.Rows, 3)
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, "$none", result.Rows[0][2])
	assert.Equal(t, "abc3.com", result.Rows[0][3])
	assert.Equal(t, float64(1), result.Rows[0][4])
	assert.Equal(t, "www.xyz.com", result.Rows[1][1])
	assert.Equal(t, "abc1", result.Rows[1][2])
	assert.Equal(t, "abc1.com", result.Rows[1][3])
	assert.Equal(t, float64(1), result.Rows[1][4])
	assert.Equal(t, "www.xyz2.com", result.Rows[2][1])
	assert.Equal(t, "abc2", result.Rows[2][2])
	assert.Equal(t, "abc2.com", result.Rows[2][3])
	assert.Equal(t, float64(1), result.Rows[2][4])

	/* Unique accounts who performed each event with breadkdown and filter.
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
	FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
	event_user_id , '0_www.xyz.com' AS event_name  FROM  (SELECT events.project_id, events.id,
	events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
	events.user_properties as event_user_properties , user_groups.group_2_user_id as group_user_id ,
	group_users.properties as group_properties FROM events  LEFT JOIN users as user_groups on
	events.user_id = user_groups.id AND user_groups.project_id = '45000009' LEFT JOIN users as
	group_users ON user_groups.group_2_user_id = group_users.group_2_user_id AND group_users.project_id
	= '45000009' AND group_users.is_group_user = true AND group_users.source IN ( 8 ) AND (
	group_users.group_4_id IS NOT NULL ) WHERE events.project_id='45000009' AND
	timestamp\u003e='1692262021' AND timestamp\u003c='1692264421' AND  ( group_user_id IS NOT NULL  )
	AND  ( events.event_name_id = 'cf37a547-fcd3-4fdf-b80f-ac46103a15d1' )  LIMIT 10000000000)
	step_0_event_users_view WHERE (JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
	'$6Signal_domain') = 'abc1.com') GROUP BY coal_group_user_id ORDER BY coal_group_user_id,
	step_0_event_users_view.timestamp ASC), step_1 AS (SELECT step_1_event_users_view.group_user_id  as
	coal_group_user_id, FIRST(step_1_event_users_view.user_id,
	FROM_UNIXTIME(step_1_event_users_view.timestamp)) as event_user_id , '1_www.xyz2.com' AS event_name
	FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
	events.properties as event_properties, events.user_properties as event_user_properties ,
	user_groups.group_2_user_id as group_user_id , group_users.properties as group_properties FROM
	events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
	= '45000009' LEFT JOIN users as group_users ON user_groups.group_2_user_id =
	group_users.group_2_user_id AND group_users.project_id = '45000009' AND group_users.is_group_user =
	true AND group_users.source IN ( 8 ) AND ( group_users.group_4_id IS NOT NULL ) WHERE
	events.project_id='45000009' AND timestamp\u003e='1692262021' AND timestamp\u003c='1692264421' AND
	( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '6c76d838-7bc6-409a-ac7b-d718a0d33d78'
	)  LIMIT 10000000000) step_1_event_users_view WHERE
	(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_domain') = 'abc1.com')
	GROUP BY coal_group_user_id ORDER BY coal_group_user_id, step_1_event_users_view.timestamp ASC) ,
	each_events_union AS (SELECT step_0.event_name as event_name, step_0.coal_group_user_id as
	coal_group_user_id, step_0.event_user_id as event_user_id FROM step_0 UNION ALL SELECT
	step_1.event_name as event_name, step_1.coal_group_user_id as coal_group_user_id,
	step_1.event_user_id as event_user_id FROM step_1) , each_users_union AS (SELECT
	each_events_union.event_user_id, each_events_union.coal_group_user_id,
	each_events_union.event_name, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
	'$hubspot_company_name') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
	'$hubspot_company_name') = '' then '$none' else CONCAT( join_timestamp, ':',
	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end), LOCATE(':', max( case
	when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null then '$none' when
	JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then '$none' else CONCAT(
	join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end
	))+1) AS _group_key_0, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
	'$6Signal_domain') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
	'$6Signal_domain') = '' then '$none' else CONCAT( join_timestamp, ':',
	JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end), LOCATE(':', max( case when
	JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null then '$none' when
	JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none' else CONCAT(
	join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end ))+1) AS
	_group_key_1 FROM each_events_union  LEFT JOIN users AS group_users on
	each_events_union.coal_group_user_id = group_users.group_2_user_id AND group_users.project_id =
	'45000009' AND  group_users.is_group_user = true AND group_users.source IN ( 2,8 ) AND (
	group_users.group_1_id IS NOT NULL OR group_users.group_4_id IS NOT NULL ) GROUP BY
	each_events_union.coal_group_user_id) SELECT event_name, _group_key_0, _group_key_1,
	COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM each_users_union GROUP BY event_name ,
	_group_key_0, _group_key_1 ORDER BY aggregate DESC LIMIT 100000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  "equals",
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, "abc1", result.Rows[0][2])
	assert.Equal(t, "abc1.com", result.Rows[0][3])
	assert.Equal(t, float64(1), result.Rows[0][4])
}

func TestAnalyticsAllAccountNegativeFilters(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	_, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)

	properties := postgres.Jsonb{[]byte(`{"user_no":"w1"}`)}
	userWeb1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w2"}`)}
	userWeb2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid2"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w3"}`)}
	userWeb3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid3"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w4"}`)}
	userWeb4, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid4"})
	assert.Equal(t, http.StatusCreated, errCode)

	dateTimeUTC := util.TimeNowZ()
	propertiesMap := U.PropertiesMap{"$hubspot_company_name": "abc1", "$hubspot_company_domain": "abc1.com", "$hubspot_company_region": "A", "hs_company_no": "h1", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	_, status = SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc1.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	propertiesMap = U.PropertiesMap{"$hubspot_company_name": "abc2", "$hubspot_company_domain": "abc2.com", "$hubspot_company_region": "B", "hs_company_no": "h2", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	_, status = SDK.TrackGroupWithDomain(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc2.com", propertiesMap, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties := &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc1.com", U.SIX_SIGNAL_REGION: "A"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb1, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc2.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb2, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc3.com", U.SIX_SIGNAL_REGION: "C"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb3, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{"$salesforce_account_website": "abc4.com", "$salesforce_account_no": "D"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb4, model.GROUP_NAME_SALESFORCE_ACCOUNT, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	status = store.GetStore().AssociateUserDomainsGroup(project.ID, userWeb1, "", "")
	assert.Equal(t, http.StatusOK, status)
	status = store.GetStore().AssociateUserDomainsGroup(project.ID, userWeb2, "", "")
	assert.Equal(t, http.StatusOK, status)

	payload := fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-9*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 3)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-9*time.Minute).Unix(), 4)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb3, U.TimeNowZ().Add(-10*time.Minute).Unix(), 5)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb4, U.TimeNowZ().Add(-10*time.Minute).Unix(), 5)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	query := model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}

	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(4), result.Rows[0][0])
	assert.Equal(t, float64(2), result.Rows[0][1])

	/*
		WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 , MAX(group_4_id) as max_group_4_id , MAX(group_2_id) as max_group_2_id FROM
		(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
		events.properties as event_properties, events.user_properties as event_user_properties ,
		user_groups.group_3_user_id as group_user_id , group_users.properties as group_properties ,
		group_users.group_4_id as group_4_id , group_users.group_2_id as group_2_id FROM events  LEFT JOIN
		users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '2000288' LEFT
		JOIN users as group_users ON user_groups.group_3_user_id = group_users.group_3_user_id AND
		group_users.project_id = '2000288' AND group_users.is_group_user = true AND group_users.source IN (
		8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_2_id IS NOT NULL ) WHERE
		events.project_id='2000288' AND timestamp>='1698817240' AND timestamp<='1698819640' AND  (
		group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '51ecfa82-30e1-4f51-8354-2f0c4842c771' )
		LIMIT 10000000000) step_0_event_users_view WHERE ( ( group_4_id is not null AND ( (
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') != 'abc1.com'  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = ''  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') IS NULL ) ) ) OR (
		group_2_id is not null AND ( ( JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') != 'abc1'  OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') = ''  OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') IS NULL ) ) ) OR ( group_4_id is NULL AND group_2_id is NULL ) ) GROUP BY
		coal_group_user_id),  step_1 AS (SELECT step_1_event_users_view.group_user_id as coal_group_user_id,
		step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_2_id) as max_group_2_id , MAX(group_4_id)
		as max_group_4_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
		events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_3_user_id as group_user_id , group_users.properties as
		group_properties , group_users.group_4_id as group_4_id , group_users.group_2_id as group_2_id FROM
		events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
		= '2000288' LEFT JOIN users as group_users ON user_groups.group_3_user_id =
		group_users.group_3_user_id AND group_users.project_id = '2000288' AND group_users.is_group_user =
		true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS NOT NULL OR
		group_users.group_2_id IS NOT NULL ) WHERE events.project_id='2000288' AND timestamp>='1698817240'
		AND timestamp<='1698819640' AND  ( group_user_id IS NOT NULL  ) AND  ( events.event_name_id =
		'0cb215c0-c149-44c1-a938-db877638e1a5' )  LIMIT 10000000000) step_1_event_users_view WHERE ( (
		group_4_id is not null AND ( ( JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') != 'abc1.com'  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') = ''  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') IS NULL ) ) ) OR ( group_2_id is not null AND ( (
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_name') != 'abc1'  OR
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_name') = ''  OR
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_name') IS NULL ) ) )
		OR ( group_4_id is NULL AND group_2_id is NULL ) ) GROUP BY coal_group_user_id,timestamp) ,
		step_1_step_0_users AS (SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp,
		FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS step_0_timestamp ,
		FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN
		step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >=
		step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT
		step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp ,
		SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null
		then '$none' when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then
		'$none' else CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_name') ) end), LOCATE(':', max( case when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then '$none' else CONCAT(
		join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end
		))+1) AS _group_key_0, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' then '$none' else CONCAT( join_timestamp, ':',
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end), LOCATE(':', max( case when
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none' else CONCAT(
		join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end ))+1) AS
		_group_key_1, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
		'$salesforce_account_website') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
		'$salesforce_account_website') = '' then '$none' else CONCAT( join_timestamp, ':',
		JSON_EXTRACT_STRING(group_users.properties, '$salesforce_account_website') ) end), LOCATE(':', max(
		case when JSON_EXTRACT_STRING(group_users.properties, '$salesforce_account_website') is null then
		'$none' when JSON_EXTRACT_STRING(group_users.properties, '$salesforce_account_website') = '' then
		'$none' else CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties,
		'$salesforce_account_website') ) end ))+1) AS _group_key_2 FROM step_0  LEFT JOIN users AS
		group_users on step_0.coal_group_user_id = group_users.group_3_user_id AND group_users.project_id =
		'2000288' AND  group_users.is_group_user = true AND group_users.source IN ( 2,8,3 ) AND (
		group_users.group_2_id IS NOT NULL OR group_users.group_4_id IS NOT NULL OR group_users.group_1_id
		IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90'  GROUP BY
		step_0.coal_group_user_id) SELECT * FROM ( SELECT _group_key_0, _group_key_1, _group_key_2,
		SUM(step_0) AS step_0 , SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS
		step_0_1_time FROM funnel GROUP BY _group_key_0, _group_key_1, _group_key_2 ORDER BY step_0 DESC
		LIMIT 10000 ) AS group_funnel UNION ALL SELECT '$no_group' AS _group_key_0,'$no_group' AS
		_group_key_1,'$no_group' AS _group_key_2 , SUM(step_0) AS step_0 , SUM(step_1) AS step_1 ,
		AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  model.NotEqualOpStr,
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.NotEqualOpStr,
				Value:     "abc1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SALESFORCE_ACCOUNT,
				Property:  "$salesforce_account_website",
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Equal(t, float64(3), result.Rows[0][3])
	assert.Equal(t, float64(1), result.Rows[0][4])
	assert.Equal(t, "33.3", result.Rows[0][5])
	assert.Equal(t, "$none", result.Rows[1][0])
	assert.Equal(t, "$none", result.Rows[1][1])
	assert.Equal(t, "abc4.com", result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, float64(0), result.Rows[1][4])
	assert.Equal(t, "0.0", result.Rows[1][5])
	assert.Equal(t, "abc2", result.Rows[2][0])
	assert.Equal(t, "abc2.com", result.Rows[2][1])
	assert.Equal(t, "$none", result.Rows[2][2])
	assert.Equal(t, float64(1), result.Rows[2][3])
	assert.Equal(t, float64(1), result.Rows[2][4])
	assert.Equal(t, "100.0", result.Rows[2][5])
	assert.Equal(t, "$none", result.Rows[3][0])
	assert.Equal(t, "abc3.com", result.Rows[3][1])
	assert.Equal(t, "$none", result.Rows[3][2])
	assert.Equal(t, float64(1), result.Rows[3][3])
	assert.Equal(t, float64(0), result.Rows[3][4])
	assert.Equal(t, "0.0", result.Rows[3][5])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				Operator:  model.NotEqualOpStr,
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Equal(t, float64(3), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "33.3", result.Rows[0][4])
	assert.Equal(t, "$none", result.Rows[1][0])
	assert.Equal(t, "$none", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(0), result.Rows[1][3])
	assert.Equal(t, "0.0", result.Rows[1][4])
	assert.Equal(t, "abc2", result.Rows[2][0])
	assert.Equal(t, "abc2.com", result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
	assert.Equal(t, float64(1), result.Rows[2][3])
	assert.Equal(t, "100.0", result.Rows[2][4])
	assert.Equal(t, "$none", result.Rows[3][0])
	assert.Equal(t, "abc3.com", result.Rows[3][1])
	assert.Equal(t, float64(1), result.Rows[3][2])
	assert.Equal(t, float64(0), result.Rows[3][3])
	assert.Equal(t, "0.0", result.Rows[3][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  model.NotEqualOpStr,
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.EqualsOpStr,
				Value:     "abc2",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "abc2", result.Rows[1][0])
	assert.Equal(t, "abc2.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])

	/*
			WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 , MAX(group_4_id) as max_group_4_id , MAX(group_2_id) as max_group_2_id FROM
		(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
		events.properties as event_properties, events.user_properties as event_user_properties ,
		user_groups.group_3_user_id as group_user_id , group_users.properties as group_properties ,
		group_users.group_4_id as group_4_id , group_users.group_2_id as group_2_id FROM events  LEFT JOIN
		users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '2000287' LEFT
		JOIN users as group_users ON user_groups.group_3_user_id = group_users.group_3_user_id AND
		group_users.project_id = '2000287' AND group_users.is_group_user = true AND group_users.source IN (
		8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_2_id IS NOT NULL ) WHERE
		events.project_id='2000287' AND timestamp>='1698817017' AND timestamp<='1698819417' AND  (
		group_user_id IS NOT NULL  ) AND  ( events.event_name_id = 'eabc2442-1183-4895-807b-16a5d977767c' )
		LIMIT 10000000000) step_0_event_users_view WHERE ( group_4_id is not null AND ( (
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') != 'abc1.com'  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = ''  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') IS NULL ) ) AND
		(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') = 'B') ) OR (
		group_2_id is not null AND (JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') = 'abc2') ) GROUP BY coal_group_user_id HAVING max_group_4_id IS NOT NULL
		AND max_group_2_id IS NOT NULL),  step_1 AS (SELECT step_1_event_users_view.group_user_id as
		coal_group_user_id, step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_4_id) as
		max_group_4_id , MAX(group_2_id) as max_group_2_id FROM  (SELECT events.project_id, events.id,
		events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
		events.user_properties as event_user_properties , user_groups.group_3_user_id as group_user_id ,
		group_users.properties as group_properties , group_users.group_4_id as group_4_id ,
		group_users.group_2_id as group_2_id FROM events  LEFT JOIN users as user_groups on events.user_id =
		user_groups.id AND user_groups.project_id = '2000287' LEFT JOIN users as group_users ON
		user_groups.group_3_user_id = group_users.group_3_user_id AND group_users.project_id = '2000287' AND
		group_users.is_group_user = true AND group_users.source IN ( 8,2 ) AND ( group_users.group_4_id IS
		NOT NULL OR group_users.group_2_id IS NOT NULL ) WHERE events.project_id='2000287' AND
		timestamp>='1698817017' AND timestamp<='1698819417' AND  ( group_user_id IS NOT NULL  ) AND  (
		events.event_name_id = '29803ea6-068a-4631-bd39-933a8fa889d2' )  LIMIT 10000000000)
		step_1_event_users_view WHERE ( group_4_id is not null AND ( (
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_domain') != 'abc1.com'  OR
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_domain') = ''  OR
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_domain') IS NULL ) ) AND
		(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$6Signal_region') = 'B') ) OR (
		group_2_id is not null AND (JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$hubspot_company_name') = 'abc2') ) GROUP BY coal_group_user_id,timestamp HAVING max_group_4_id IS
		NOT NULL AND max_group_2_id IS NOT NULL) , step_1_step_0_users AS (SELECT step_1.coal_group_user_id,
		FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS
		step_0_timestamp , FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM
		step_0 LEFT JOIN step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE
		step_1.timestamp >= step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT
		DISTINCT step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp ,
		SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null
		then '$none' when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then
		'$none' else CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_name') ) end), LOCATE(':', max( case when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then '$none' else CONCAT(
		join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end
		))+1) AS _group_key_0, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' then '$none' else CONCAT( join_timestamp, ':',
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end), LOCATE(':', max( case when
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none' else CONCAT(
		join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end ))+1) AS
		_group_key_1 FROM step_0  LEFT JOIN users AS group_users on step_0.coal_group_user_id =
		group_users.group_3_user_id AND group_users.project_id = '2000287' AND  group_users.is_group_user =
		true AND group_users.source IN ( 2,8 ) AND ( group_users.group_2_id IS NOT NULL OR
		group_users.group_4_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90'  GROUP BY
		step_0.coal_group_user_id) SELECT * FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0
		, SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP
		BY _group_key_0, _group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT
		'$no_group' AS _group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS
		step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  model.NotEqualOpStr,
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_REGION,
				Operator:  model.EqualsOpStr,
				Value:     "B",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.EqualsOpStr,
				Value:     "abc2",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "abc2", result.Rows[1][0])
	assert.Equal(t, "abc2.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  model.NotEqualOpStr,
				Value:     "abc2.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_REGION,
				Operator:  model.EqualsOpStr,
				Value:     "A",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.EqualsOpStr,
				Value:     "abc1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "abc1", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  model.NotEqualOpStr,
				Value:     "abc2.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.EqualsOpStr,
				Value:     "abc1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "abc1", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.NotEqualOpStr,
				Value:     "abc1",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				Operator:  model.NotEqualOpStr,
				Value:     "abc1.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(3), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "33.3", result.Rows[0][4])
	assert.Equal(t, "$none", result.Rows[1][0])
	assert.Equal(t, "$none", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(0), result.Rows[1][3])
	assert.Equal(t, "0.0", result.Rows[1][4])
	assert.Equal(t, "$none", result.Rows[2][0])
	assert.Equal(t, "abc3.com", result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
	assert.Equal(t, float64(0), result.Rows[2][3])
	assert.Equal(t, "0.0", result.Rows[2][4])
	assert.Equal(t, "abc2", result.Rows[3][0])
	assert.Equal(t, "abc2.com", result.Rows[3][1])
	assert.Equal(t, float64(1), result.Rows[3][2])
	assert.Equal(t, float64(1), result.Rows[3][3])

	/*
			WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0 , MAX(group_4_id) as max_group_4_id , MAX(group_2_id) as max_group_2_id FROM
		(SELECT events.project_id, events.id, events.event_name_id, events.user_id, events.timestamp ,
		events.properties as event_properties, events.user_properties as event_user_properties ,
		user_groups.group_3_user_id as group_user_id , group_users.properties as group_properties ,
		group_users.group_4_id as group_4_id , group_users.group_2_id as group_2_id FROM events  LEFT JOIN
		users as user_groups on events.user_id = user_groups.id AND user_groups.project_id = '2000286' LEFT
		JOIN users as group_users ON user_groups.group_3_user_id = group_users.group_3_user_id AND
		group_users.project_id = '2000286' AND group_users.is_group_user = true AND group_users.source IN (
		8,2 ) AND ( group_users.group_4_id IS NOT NULL OR group_users.group_2_id IS NOT NULL ) WHERE
		events.project_id='2000286' AND timestamp >= '1698816647' AND timestamp <='1698819047' AND  (
		group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '46fb743a-99ec-4e0e-9d71-40b50f3a36e1' )
		LIMIT 10000000000) step_0_event_users_view WHERE ( ( group_4_id is not null AND ( (
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') != 'abc2.com'  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') = ''  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_domain') IS NULL )  OR  (
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') != 'B'  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') = ''  OR
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') IS NULL ) ) ) OR (
		group_2_id is not null AND ( ( JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') != 'abc2'  OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') = ''  OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_name') IS NULL )  OR  (
		JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$hubspot_company_domain') !=
		'abc2.com'  OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_domain') = ''  OR JSON_EXTRACT_STRING(step_0_event_users_view.group_properties,
		'$hubspot_company_domain') IS NULL ) ) ) OR ( group_4_id is NULL AND group_2_id is NULL ) ) GROUP BY
		coal_group_user_id),  step_1 AS (SELECT step_1_event_users_view.group_user_id as coal_group_user_id,
		step_1_event_users_view.timestamp, 1 as step_1 , MAX(group_4_id) as max_group_4_id , MAX(group_2_id)
		as max_group_2_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
		events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_3_user_id as group_user_id , group_users.properties as
		group_properties , group_users.group_4_id as group_4_id , group_users.group_2_id as group_2_id FROM
		events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
		= '2000286' LEFT JOIN users as group_users ON user_groups.group_3_user_id =
		group_users.group_3_user_id AND group_users.project_id = '2000286' AND group_users.is_group_user =
		true AND group_users.source IN ( 2,8 ) AND ( group_users.group_2_id IS NOT NULL OR
		group_users.group_4_id IS NOT NULL ) WHERE events.project_id='2000286' AND timestamp >= '1698816647'
		AND timestamp <='1698819047' AND  ( group_user_id IS NOT NULL  ) AND  ( events.event_name_id =
		'429b3b0b-224b-4c4c-b533-844524fa7904' )  LIMIT 10000000000) step_1_event_users_view WHERE ( (
		group_4_id is not null AND ( ( JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') != 'abc2.com'  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') = ''  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_domain') IS NULL )  OR  ( JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_region') != 'B'  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_region') = ''  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$6Signal_region') IS NULL ) ) ) OR ( group_2_id is not null AND ( (
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_name') != 'abc2'  OR
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_name') = ''  OR
		JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_name') IS NULL )  OR
		( JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_domain') !=
		'abc2.com'  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$hubspot_company_domain') = ''  OR JSON_EXTRACT_STRING(step_1_event_users_view.group_properties,
		'$hubspot_company_domain') IS NULL ) ) ) OR ( group_4_id is NULL AND group_2_id is NULL ) ) GROUP BY
		coal_group_user_id,timestamp) , step_1_step_0_users AS (SELECT step_1.coal_group_user_id,
		FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as timestamp, step_1 , step_0.timestamp AS
		step_0_timestamp , FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM
		step_0 LEFT JOIN step_1 ON step_0.coal_group_user_id = step_1.coal_group_user_id WHERE
		step_1.timestamp  >=  step_0.timestamp GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT
		DISTINCT step_0.coal_group_user_id , step_0 , step_1 , step_0_timestamp , step_1_timestamp ,
		SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null
		then '$none' when JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then
		'$none' else CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_name') ) end), LOCATE(':', max( case when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') = '' then '$none' else CONCAT(
		join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_name') ) end
		))+1) AS _group_key_0, SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
		'$6Signal_domain') = '' then '$none' else CONCAT( join_timestamp, ':',
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end), LOCATE(':', max( case when
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none' else CONCAT(
		join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end ))+1) AS
		_group_key_1 FROM step_0  LEFT JOIN users AS group_users on step_0.coal_group_user_id =
		group_users.group_3_user_id AND group_users.project_id = '2000286' AND  group_users.is_group_user =
		true AND group_users.source IN ( 2,8 ) AND ( group_users.group_2_id IS NOT NULL OR
		group_users.group_4_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata')))  <= '90'  GROUP BY
		step_0.coal_group_user_id) SELECT * FROM ( SELECT _group_key_0, _group_key_1, SUM(step_0) AS step_0
		, SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel GROUP
		BY _group_key_0, _group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS group_funnel UNION ALL SELECT
		'$no_group' AS _group_key_0,'$no_group' AS _group_key_1 , SUM(step_0) AS step_0 , SUM(step_1) AS
		step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM funnel"
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				Operator:  model.NotEqualOpStr,
				Value:     "abc2.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_REGION,
				Operator:  model.NotEqualOpStr,
				Value:     "B",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				Operator:  model.NotEqualOpStr,
				Value:     "abc2",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_domain",
				Operator:  model.NotEqualOpStr,
				Value:     "abc2.com",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "OR",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_name",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Equal(t, float64(3), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "33.3", result.Rows[0][4])
	assert.Equal(t, "$none", result.Rows[1][0])
	assert.Equal(t, "$none", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(0), result.Rows[1][3])
	assert.Equal(t, "0.0", result.Rows[1][4])

	assert.Equal(t, "abc1", result.Rows[2][0])
	assert.Equal(t, "abc1.com", result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
	assert.Equal(t, float64(1), result.Rows[2][3])
	assert.Equal(t, "100.0", result.Rows[2][4])

	assert.Equal(t, "$none", result.Rows[3][0])
	assert.Equal(t, "abc3.com", result.Rows[3][1])
	assert.Equal(t, float64(1), result.Rows[3][2])
	assert.Equal(t, float64(0), result.Rows[3][3])
	assert.Equal(t, "0.0", result.Rows[3][4])

}

func TestAnalyticsAllAccountsFilterBreadkdown(t *testing.T) {
	project, _, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	_, status := store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_SALESFORCE_ACCOUNT, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)

	properties := postgres.Jsonb{[]byte(fmt.Sprintf(`{"user_no":"w1","%s":1}`, U.SP_PAGE_COUNT))}
	userWeb1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w2"}`)}
	userWeb2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid2"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w2"}`)}
	userWeb3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid3"})
	assert.Equal(t, http.StatusCreated, errCode)

	_, status = store.GetStore().CreateOrGetGroupByName(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)

	dateTimeUTC := util.TimeNowZ()
	propertiesMap := map[string]interface{}{"$hubspot_company_name": "abc1", "$hubspot_company_domain": "abc1.com", "$hubspot_company_region": "A",
		"hs_company_no": "h1", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	hsCompany1UserID, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "h1", "", &propertiesMap,
		dateTimeUTC.Unix(), dateTimeUTC.Unix(), model.UserSourceHubspotString)
	assert.Nil(t, err)
	status = SDK.TrackDomainsGroup(project.ID, hsCompany1UserID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc1.com", nil, dateTimeUTC.Unix())

	propertiesMap = map[string]interface{}{"$hubspot_company_name": "abc1", "$hubspot_company_domain": "abc1.com", "$hubspot_company_region": "B",
		"hs_company_no": "h2", "$hubspot_company_createddate": dateTimeUTC.Unix() + 10}
	hsCompany2UserID, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "h2", "", &propertiesMap,
		dateTimeUTC.Unix(), dateTimeUTC.Unix(), model.UserSourceHubspotString)
	assert.Nil(t, err)
	status = SDK.TrackDomainsGroup(project.ID, hsCompany2UserID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc1.com", nil, dateTimeUTC.Unix())

	propertiesMap = map[string]interface{}{"$hubspot_company_name": "abc2", "$hubspot_company_domain": "abc2.com", "$hubspot_company_region": "D",
		"hs_company_no": "h3", "$hubspot_company_createddate": dateTimeUTC.Unix()}
	hsCompany3UserID, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, "h1", "", &propertiesMap,
		dateTimeUTC.Unix(), dateTimeUTC.Unix(), model.UserSourceHubspotString)
	assert.Nil(t, err)
	status = SDK.TrackDomainsGroup(project.ID, hsCompany3UserID, model.GROUP_NAME_HUBSPOT_COMPANY, "abc2.com", nil, dateTimeUTC.Unix())

	groupProperties := &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc1.com", U.SIX_SIGNAL_REGION: "A"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb1, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	propertiesMap = map[string]interface{}{U.SIX_SIGNAL_DOMAIN: "abc1.com", U.SIX_SIGNAL_REGION: "AA"}
	sixSignalUser3, err := store.GetStore().CreateOrUpdateGroupPropertiesBySource(project.ID, model.GROUP_NAME_SIX_SIGNAL, "h1", "", &propertiesMap,
		dateTimeUTC.Unix(), dateTimeUTC.Unix()-10000, model.UserSourceSixSignalString)
	assert.Nil(t, err)
	status = SDK.TrackDomainsGroup(project.ID, sixSignalUser3, model.GROUP_NAME_SIX_SIGNAL, "abc1.com", nil, dateTimeUTC.Unix())

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc2.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb2, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	groupProperties = &U.PropertiesMap{U.SIX_SIGNAL_DOMAIN: "abc3.com", U.SIX_SIGNAL_REGION: "B"}
	status = SDK.TrackUserAccountGroup(project.ID, userWeb3, model.GROUP_NAME_SIX_SIGNAL, groupProperties, util.TimeNowUnix())
	assert.Equal(t, http.StatusOK, status)

	payload := fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-9*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 3)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-9*time.Minute).Unix(), 4)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz3.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb3, U.TimeNowZ().Add(-9*time.Minute).Unix(), 4)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	/*
			WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id as coal_group_user_id,
		FIRST(step_0_event_users_view.timestamp, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		timestamp, 1 as step_0, GROUP_CONCAT(step_0_event_users_view.group_users_user_id) as
		group_users_user_ids , MAX(group_2_id) as max_group_2_id FROM  (SELECT events.project_id, events.id,
		events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
		events.user_properties as event_user_properties , user_groups.group_3_user_id as group_user_id ,
		group_users.properties as group_properties , group_users.id as group_users_user_id ,
		group_users.group_2_id as group_2_id FROM events  LEFT JOIN users as user_groups on events.user_id =
		user_groups.id AND user_groups.project_id = '5000100' LEFT JOIN users as group_users ON
		user_groups.group_3_user_id = group_users.group_3_user_id AND group_users.project_id = '5000100' AND
		group_users.is_group_user = true AND group_users.source IN ( 2 ) AND ( group_users.group_2_id IS NOT
		NULL ) WHERE events.project_id='5000100' AND timestamp>='1699262536' AND timestamp<='1699264936' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '2d49b911-fc88-4762-b56f-eeffd130070e'
		)  LIMIT 10000000000) step_0_event_users_view WHERE ( group_2_id is not null AND
		(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$hubspot_company_region') = 'A') )
		GROUP BY coal_group_user_id HAVING max_group_2_id IS NOT NULL),  step_1 AS (SELECT
		step_1_event_users_view.group_user_id as coal_group_user_id, step_1_event_users_view.timestamp, 1 as
		step_1 , MAX(group_2_id) as max_group_2_id FROM  (SELECT events.project_id, events.id,
		events.event_name_id, events.user_id, events.timestamp , events.properties as event_properties,
		events.user_properties as event_user_properties , user_groups.group_3_user_id as group_user_id ,
		group_users.properties as group_properties , group_users.id as group_users_user_id ,
		group_users.group_2_id as group_2_id FROM events  LEFT JOIN users as user_groups on events.user_id =
		user_groups.id AND user_groups.project_id = '5000100' LEFT JOIN users as group_users ON
		user_groups.group_3_user_id = group_users.group_3_user_id AND group_users.project_id = '5000100' AND
		group_users.is_group_user = true AND group_users.source IN ( 2 ) AND ( group_users.group_2_id IS NOT
		NULL ) WHERE events.project_id='5000100' AND timestamp>='1699262536' AND timestamp<='1699264936' AND
		( group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '0246ad3e-4f0d-4095-9d00-bd3360567ba8'
		)  LIMIT 10000000000) step_1_event_users_view WHERE ( group_2_id is not null AND
		(JSON_EXTRACT_STRING(step_1_event_users_view.group_properties, '$hubspot_company_region') = 'A') )
		GROUP BY coal_group_user_id,timestamp HAVING max_group_2_id IS NOT NULL) , step_1_step_0_users AS
		(SELECT step_1.coal_group_user_id, FIRST(step_1.timestamp, FROM_UNIXTIME(step_1.timestamp)) as
		timestamp, step_1 , step_0.timestamp AS step_0_timestamp , FIRST(step_1.timestamp,
		FROM_UNIXTIME(step_1.timestamp)) AS step_1_timestamp FROM step_0 LEFT JOIN step_1 ON
		step_0.coal_group_user_id = step_1.coal_group_user_id WHERE step_1.timestamp >= step_0.timestamp
		GROUP BY step_1.coal_group_user_id) , funnel AS (SELECT DISTINCT step_0.coal_group_user_id , step_0
		, step_1 , step_0_timestamp , step_1_timestamp , SUBSTRING(max(case when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region') = '' then '$none' else
		CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region')
		) end), LOCATE(':', max( case when JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_region') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_region') = '' then '$none' else CONCAT( join_timestamp, ':',
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region') ) end ))+1) AS _group_key_0,
		SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null then
		'$none' when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none' else
		CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end),
		LOCATE(':', max( case when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null
		then '$none' when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none'
		else CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') )
		end ))+1) AS _group_key_1 FROM step_0  LEFT JOIN users AS group_users on step_0.coal_group_user_id =
		group_users.group_3_user_id AND group_users.project_id = '5000100' AND  group_users.is_group_user =
		true AND group_users.source IN ( 2,8 ) AND ( group_users.group_2_id IS NOT NULL OR
		group_users.group_4_id IS NOT NULL )  LEFT JOIN step_1_step_0_users ON
		step_0.coal_group_user_id=step_1_step_0_users.coal_group_user_id  AND timestampdiff(DAY,
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_0_timestamp), 'UTC', 'Asia/Kolkata')),
		DATE(CONVERT_TZ(FROM_UNIXTIME(step_1_timestamp), 'UTC', 'Asia/Kolkata'))) <= '90'  WHERE ( ((
		group_2_id is not null ) AND LOCATE( group_users.id,step_0.group_users_user_ids)>0 ) OR ( group_2_id
		is null )) GROUP BY step_0.coal_group_user_id) SELECT * FROM ( SELECT _group_key_0, _group_key_1,
		SUM(step_0) AS step_0 , SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS
		step_0_1_time FROM funnel GROUP BY _group_key_0, _group_key_1 ORDER BY step_0 DESC LIMIT 10000 ) AS
		group_funnel UNION ALL SELECT '$no_group' AS _group_key_0,'$no_group' AS _group_key_1 , SUM(step_0)
		AS step_0 , SUM(step_1) AS step_1 , AVG(step_1_timestamp-step_0_timestamp) AS step_0_1_time FROM
		funnel
	*/
	query := model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				Operator:  model.EqualsOpStr,
				Value:     "A",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "A", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])

	query.GlobalUserProperties = []model.QueryProperty{
		{
			Entity:    model.PropertyEntityUserGlobal,
			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
			Property:  "$hubspot_company_region",
			Operator:  model.EqualsOpStr,
			Value:     "B",
			Type:      U.PropertyTypeCategorical,
			LogicalOp: "AND",
		},
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, float64(1), result.Rows[0][2])
	assert.Equal(t, float64(1), result.Rows[0][3])
	assert.Equal(t, "100.0", result.Rows[0][4])
	assert.Equal(t, "B", result.Rows[1][0])
	assert.Equal(t, "abc1.com", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				Operator:  model.EqualsOpStr,
				Value:     "A",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, "A", result.Rows[0][2])
	assert.Equal(t, "abc1.com", result.Rows[0][3])
	assert.Equal(t, float64(1), result.Rows[0][4])
	assert.Equal(t, "www.xyz2.com", result.Rows[1][1])
	assert.Equal(t, "A", result.Rows[1][2])
	assert.Equal(t, "abc1.com", result.Rows[1][3])
	assert.Equal(t, float64(1), result.Rows[1][4])

	query.GlobalUserProperties = []model.QueryProperty{
		{
			Entity:    model.PropertyEntityUserGlobal,
			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
			Property:  "$hubspot_company_region",
			Operator:  model.EqualsOpStr,
			Value:     "B",
			Type:      U.PropertyTypeCategorical,
			LogicalOp: "AND",
		},
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, "B", result.Rows[0][2])
	assert.Equal(t, "abc1.com", result.Rows[0][3])
	assert.Equal(t, float64(1), result.Rows[0][4])
	assert.Equal(t, "www.xyz2.com", result.Rows[1][1])
	assert.Equal(t, "B", result.Rows[1][2])
	assert.Equal(t, "abc1.com", result.Rows[1][3])
	assert.Equal(t, float64(1), result.Rows[1][4])

	query.GlobalUserProperties = []model.QueryProperty{
		{
			Entity:    model.PropertyEntityUserGlobal,
			GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
			Property:  "$hubspot_company_region",
			Operator:  model.NotEqualOpStr,
			Value:     "B",
			Type:      U.PropertyTypeCategorical,
			LogicalOp: "AND",
		},
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 4)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		if p1 == p2 {
			return U.GetPropertyValueAsString(result.Rows[i][2]) < U.GetPropertyValueAsString(result.Rows[j][2])
		}
		return p1 < p2
	})
	assert.Equal(t, "www.xyz.com", result.Rows[0][1])
	assert.Equal(t, "A", result.Rows[0][2])
	assert.Equal(t, "abc1.com", result.Rows[0][3])
	assert.Equal(t, float64(1), result.Rows[0][4])
	assert.Equal(t, "www.xyz.com", result.Rows[1][1])
	assert.Equal(t, "D", result.Rows[1][2])
	assert.Equal(t, "abc2.com", result.Rows[1][3])
	assert.Equal(t, float64(1), result.Rows[1][4])
	assert.Equal(t, "www.xyz2.com", result.Rows[2][1])
	assert.Equal(t, "A", result.Rows[2][2])
	assert.Equal(t, "abc1.com", result.Rows[2][3])
	assert.Equal(t, float64(1), result.Rows[2][4])
	assert.Equal(t, "www.xyz2.com", result.Rows[3][1])
	assert.Equal(t, "D", result.Rows[3][2])
	assert.Equal(t, "abc2.com", result.Rows[3][3])
	assert.Equal(t, float64(1), result.Rows[3][4])

	query.EventsCondition = model.EventCondAllGivenEvent
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "A", result.Rows[0][0])
	assert.Equal(t, "abc1.com", result.Rows[0][1])
	assert.Equal(t, "D", result.Rows[1][0])
	assert.Equal(t, "abc2.com", result.Rows[1][1])

	query.EventsCondition = model.EventCondAnyGivenEvent
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "A", result.Rows[0][0])
	assert.Equal(t, "abc1.com", result.Rows[0][1])
	assert.Equal(t, "D", result.Rows[1][0])
	assert.Equal(t, "abc2.com", result.Rows[1][1])

	query.EventsWithProperties = append(query.EventsWithProperties, model.QueryEventWithProperties{
		Name: "www.xyz3.com", Properties: []model.QueryProperty{}})
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})
	assert.Equal(t, "A", result.Rows[0][0])
	assert.Equal(t, "abc1.com", result.Rows[0][1])
	assert.Equal(t, "D", result.Rows[1][0])
	assert.Equal(t, "abc2.com", result.Rows[1][1])
	assert.Equal(t, "$none", result.Rows[2][0])
	assert.Equal(t, "abc3.com", result.Rows[2][1])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				Operator:  model.EqualsOpStr,
				Value:     "A",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, "www.xyz2.com", result.Rows[0][1])
	assert.Equal(t, "A", result.Rows[0][2])
	assert.Equal(t, "abc1.com", result.Rows[0][3])

	/*
			WITH  step_0 AS (SELECT step_0_event_users_view.group_user_id  as coal_group_user_id,
		GROUP_CONCAT(step_0_event_users_view.group_users_user_id) as group_users_user_ids,
		FIRST(step_0_event_users_view.user_id, FROM_UNIXTIME(step_0_event_users_view.timestamp)) as
		event_user_id, CASE WHEN JSON_EXTRACT_STRING(step_0_event_users_view.event_user_properties,
		'$hubspot_company_region') IS NULL THEN '$none' WHEN
		JSON_EXTRACT_STRING(step_0_event_users_view.event_user_properties, '$hubspot_company_region') = ''
		THEN '$none' ELSE JSON_EXTRACT_STRING(step_0_event_users_view.event_user_properties,
		'$hubspot_company_region') END AS _group_key_0 , '0_www.xyz2.com' AS event_name  , MAX(group_4_id)
		as max_group_4_id FROM  (SELECT events.project_id, events.id, events.event_name_id, events.user_id,
		events.timestamp , events.properties as event_properties, events.user_properties as
		event_user_properties , user_groups.group_3_user_id as group_user_id , group_users.properties as
		group_properties , group_users.id as group_users_user_id , group_users.group_4_id as group_4_id FROM
		events  LEFT JOIN users as user_groups on events.user_id = user_groups.id AND user_groups.project_id
		= '5000005' LEFT JOIN users as group_users ON user_groups.group_3_user_id =
		group_users.group_3_user_id AND group_users.project_id = '5000005' AND group_users.is_group_user =
		true AND group_users.source IN ( 8 ) AND ( group_users.group_4_id IS NOT NULL ) WHERE
		events.project_id='5000005' AND timestamp>='1699258422' AND timestamp<='1699260822' AND  (
		group_user_id IS NOT NULL  ) AND  ( events.event_name_id = '07f3a671-040c-4d3e-8ec0-d571184fd9a4' )
		LIMIT 10000000000) step_0_event_users_view WHERE ( group_4_id is not null AND
		(JSON_EXTRACT_STRING(step_0_event_users_view.group_properties, '$6Signal_region') = 'AA') ) AND (
		(JSON_EXTRACT_STRING(step_0_event_users_view.event_user_properties, 'user_no') = 'w1') ) GROUP BY
		coal_group_user_id , _group_key_0 HAVING max_group_4_id IS NOT NULL ORDER BY coal_group_user_id,
		_group_key_0, step_0_event_users_view.timestamp ASC) , each_users_union AS (SELECT
		step_0.event_user_id, step_0.coal_group_user_id,  step_0.event_name, SUBSTRING(max(case when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region') is null then '$none' when
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region') = '' then '$none' else
		CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region')
		) end), LOCATE(':', max( case when JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_region') is null then '$none' when JSON_EXTRACT_STRING(group_users.properties,
		'$hubspot_company_region') = '' then '$none' else CONCAT( join_timestamp, ':',
		JSON_EXTRACT_STRING(group_users.properties, '$hubspot_company_region') ) end ))+1) AS _group_key_1,
		SUBSTRING(max(case when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null then
		'$none' when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none' else
		CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') ) end),
		LOCATE(':', max( case when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') is null
		then '$none' when JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') = '' then '$none'
		else CONCAT( join_timestamp, ':', JSON_EXTRACT_STRING(group_users.properties, '$6Signal_domain') )
		end ))+1) AS _group_key_2 , _group_key_0 FROM step_0  LEFT JOIN users AS group_users on
		step_0.coal_group_user_id = group_users.group_3_user_id AND group_users.project_id = '5000005' AND
		group_users.is_group_user = true AND group_users.source IN ( 2,8 ) AND ( group_users.group_2_id IS
		NOT NULL OR group_users.group_4_id IS NOT NULL ) WHERE ( (( group_4_id is not null ) AND LOCATE(
		group_users.id,step_0.group_users_user_ids)\u003e0 ) OR ( group_4_id is null )) GROUP BY
		step_0.event_name, step_0.coal_group_user_id) SELECT event_name, _group_key_0, _group_key_1,
		_group_key_2, COUNT(DISTINCT(coal_group_user_id)) AS aggregate FROM each_users_union GROUP BY
		event_name , _group_key_0, _group_key_1, _group_key_2 ORDER BY aggregate DESC LIMIT 100000
	*/
	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "www.xyz2.com",
				Properties: []model.QueryProperty{
					{
						Entity:    model.PropertyEntityUser,
						Property:  "user_no",
						Operator:  model.EqualsOpStr,
						Value:     "w1",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "AND",
					},
				},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_REGION,
				Operator:  model.EqualsOpStr,
				Value:     "AA",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventNameIndex: 1,
				Property:       "$hubspot_company_region",
				EventName:      "www.xyz2.com",
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, "www.xyz2.com", result.Rows[0][1])
	assert.Equal(t, "$none", result.Rows[0][2])
	assert.Equal(t, "B", result.Rows[0][3])
	assert.Equal(t, "abc1.com", result.Rows[0][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "www.xyz2.com",
				Properties: []model.QueryProperty{
					{
						Entity:    model.PropertyEntityUser,
						Property:  "user_no",
						Operator:  model.EqualsOpStr,
						Value:     "w1",
						Type:      U.PropertyTypeCategorical,
						LogicalOp: "AND",
					},
				},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_DOMAINS,
				Property:  U.VISITED_WEBSITE,
				Operator:  model.EqualsOpStr,
				Value:     "true",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				Operator:  model.EqualsOpStr,
				Value:     "A",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventNameIndex: 1,
				Property:       "$hubspot_company_region",
				EventName:      "www.xyz2.com",
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassEvents,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Nil(t, err)
	assert.Equal(t, "www.xyz2.com", result.Rows[0][1])
	assert.Equal(t, "$none", result.Rows[0][2])
	assert.Equal(t, "A", result.Rows[0][3])
	assert.Equal(t, "abc1.com", result.Rows[0][4])

	query.GlobalUserProperties = []model.QueryProperty{
		{
			Entity:    model.PropertyEntityUserGlobal,
			GroupName: model.GROUP_NAME_DOMAINS,
			Property:  U.VISITED_WEBSITE,
			Operator:  model.EqualsOpStr,
			Value:     "true",
			Type:      U.PropertyTypeCategorical,
			LogicalOp: "AND",
		},
	}

	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Len(t, result.Rows, 1)
	assert.Nil(t, err)
	assert.Equal(t, "www.xyz2.com", result.Rows[0][1])
	assert.Equal(t, "$none", result.Rows[0][2])
	assert.Equal(t, "B", result.Rows[0][3])
	assert.Equal(t, "abc1.com", result.Rows[0][4])

	query = model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},
		GlobalUserProperties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUserGlobal,
				GroupName: model.GROUP_NAME_DOMAINS,
				Property:  U.VISITED_WEBSITE,
				Operator:  model.EqualsOpStr,
				Value:     "true",
				Type:      U.PropertyTypeCategorical,
				LogicalOp: "AND",
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_HUBSPOT_COMPANY,
				Property:  "$hubspot_company_region",
				EventName: model.UserPropertyGroupByPresent,
			},
			{
				Entity:    model.PropertyEntityUser,
				GroupName: model.GROUP_NAME_SIX_SIGNAL,
				Property:  U.SIX_SIGNAL_DOMAIN,
				EventName: model.UserPropertyGroupByPresent,
			},
		},
		GroupAnalysis:   model.GROUP_NAME_DOMAINS,
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondEachGivenEvent,
	}
	result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery(), true)
	assert.Equal(t, http.StatusOK, errCode)

}

func TestAnalyticsFunnelValueLabel(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/event/track"

	status := store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_owner", "1", "o1")
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreateOrUpdateDisplayNameLabel(project.ID, "hubspot", "$hubspot_contact_owner", "2", "o2")
	assert.Equal(t, http.StatusCreated, status)

	properties := postgres.Jsonb{[]byte(fmt.Sprintf(`{"user_no":"w1","$hubspot_contact_owner":1}`))}
	userWeb1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid1"})
	assert.Equal(t, http.StatusCreated, errCode)

	properties = postgres.Jsonb{[]byte(`{"user_no":"w2","$hubspot_contact_owner":2}`)}
	userWeb2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: properties,
		Source: model.GetRequestSourcePointer(model.UserSourceWeb), CustomerUserId: "cuid2"})
	assert.Equal(t, http.StatusCreated, errCode)
	payload := fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-10*time.Minute).Unix(), 1)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb1, U.TimeNowZ().Add(-9*time.Minute).Unix(), 2)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-10*time.Minute).Unix(), 3)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	payload = fmt.Sprintf(`{"event_name": "www.xyz2.com", "user_id": "%s", "timestamp": %d,"event_properties":{"event_id":%d}}`,
		userWeb2, U.TimeNowZ().Add(-9*time.Minute).Unix(), 4)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	query := model.Query{
		From: U.TimeNowZ().Add(-20 * time.Minute).Unix(),
		To:   U.TimeNowZ().Add(20 * time.Minute).Unix(),
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name:       "www.xyz.com",
				Properties: []model.QueryProperty{},
			},
			{
				Name:       "www.xyz2.com",
				Properties: []model.QueryProperty{},
			},
		},

		GroupByProperties: []model.QueryGroupByProperty{
			{
				Entity:         model.PropertyEntityUser,
				EventNameIndex: 1,
				Property:       "$hubspot_contact_owner",
				EventName:      "www.xyz.com",
			},
			{
				Entity:         model.PropertyEntityUser,
				EventNameIndex: 1,
				Property:       "user_no",
				EventName:      "www.xyz.com",
			},
		},
		Class:           model.QueryClassFunnel,
		Type:            model.QueryTypeUniqueUsers,
		EventsCondition: model.EventCondAllGivenEvent,
	}

	w = sendAnalyticsQueryReq(r, model.QueryClassFunnel, project.ID, agent, 0, 0, "", &query, true, false)
	assert.NotEmpty(t, w)
	result := DecodeJSONResponseToAnalyticsResult(w.Body)
	sort.Slice(result.Rows, func(i, j int) bool {
		p1 := U.GetPropertyValueAsString(result.Rows[i][1])
		p2 := U.GetPropertyValueAsString(result.Rows[j][1])
		return p1 < p2
	})

	assert.Len(t, result.Rows, 3)
	assert.Nil(t, err)
	assert.Equal(t, "o1", result.Rows[1][0])
	assert.Equal(t, "w1", result.Rows[1][1])
	assert.Equal(t, float64(1), result.Rows[1][2])
	assert.Equal(t, float64(1), result.Rows[1][3])
	assert.Equal(t, "100.0", result.Rows[1][4])
	assert.Equal(t, "100.0", result.Rows[1][5])
	assert.Equal(t, "o2", result.Rows[2][0])
	assert.Equal(t, "w2", result.Rows[2][1])
	assert.Equal(t, float64(1), result.Rows[2][2])
	assert.Equal(t, float64(1), result.Rows[2][3])
	assert.Equal(t, "100.0", result.Rows[2][4])
	assert.Equal(t, "100.0", result.Rows[2][5])
}
