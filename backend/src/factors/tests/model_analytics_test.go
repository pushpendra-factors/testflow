package tests

import (
	"encoding/json"
	C "factors/config"
	TaskSession "factors/task/session"
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
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID, groupName, groupID, groupUserID)
		assert.Equal(t, http.StatusAccepted, status)
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID1, groupName, groupID, groupUserID)
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
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID, groupName, groupID, groupUserID)
		assert.Equal(t, http.StatusAccepted, status)
		_, status = store.GetStore().UpdateUserGroup(project.ID, createdUserID1, groupName, groupID, groupUserID)
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

	// s0 event property value with 5.
	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 5,"id": 1}, "user_properties": {"gender": "M", "age": 18}}`,
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
			icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
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
		model.QueryClassInsights:  `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:    `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassChannel:   `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
		model.QueryClassEvents:    `{"query_group":[{"cl":"events","ty":"unique_users","ec":"each_given_event","fr":1583001000,"to":1585679399,"ewp":[{"na":"$session","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]},{"na":"MagazineViews","pr":[{"en":"event","pr":"$source","op":"equals","va":"google","ty":"categorical","lop":"AND"},{"en":"user","pr":"$country","op":"equals","va":"India","ty":"categorical","lop":"AND"}]}],"gbp":[{"pr":"$browser","en":"event","pty":"categorical","ena":"$session","eni":1},{"pr":"$campaign","en":"event","pty":"categorical","ena":"MagazineViews","eni":2}],"gbt":"","tz":"Asia/Calcutta"}]}`,
		model.QueryClassChannelV1: `{ "query_group":[{ "channel": "google_ads", "select_metrics": ["impressions"], "filters": [], "group_by": [], "gbt": "hour", "fr": 1585679400, "to": 1585765800 }], "cl": "channel_v1" }`,
		model.QueryClassWeb:       `{"units":[{"unit_id":194,"query_name":"bounce_rate"},{"unit_id":195,"query_name":"unique_users"},{"unit_id":196,"query_name":"avg_session_duration"},{"unit_id":197,"query_name":"avg_pages_per_session"},{"unit_id":200,"query_name":"sessions"},{"unit_id":201,"query_name":"total_page_view"},{"unit_id":199,"query_name":"traffic_channel_report"},{"unit_id":198,"query_name":"top_pages_report"}],"custom_group_units":[],"from":1609612200,"to":1610044199}`,
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

func TestQueryCaching(t *testing.T) {
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
		model.QueryClassInsights:    `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrence", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:      `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_users", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
		model.QueryClassChannel:     `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
		model.QueryClassEvents:      `{"query_group":[{"cl":"events","ty":"events_occurrence","fr":1612031400,"to":1612376999,"ewp":[{"na":"$hubspot_contact_created","pr":[]}],"gbt":"date","gbp":[{"pr":"$hubspot_contact_revenue_segment_fs_","en":"event","pty":"categorical","ena":"$hubspot_contact_created","eni":1},{"pr":"$hubspot_contact_revenue_segment_fs_","en":"user","pty":"categorical","ena":"$present"}],"ec":"each_given_event","tz":"Asia/Kolkata"},{"cl":"events","ty":"events_occurrence","fr":1612031400,"to":1612376999,"ewp":[{"na":"$hubspot_contact_created","pr":[]}],"gbt":"","gbp":[{"pr":"$hubspot_contact_revenue_segment_fs_","en":"event","pty":"categorical","ena":"$hubspot_contact_created","eni":1},{"pr":"$hubspot_contact_revenue_segment_fs_","en":"user","pty":"categorical","ena":"$present"}],"ec":"each_given_event","tz":"Asia/Kolkata"}]}`,
		model.QueryClassChannelV1:   `{"query_group":[{"channel":"facebook_ads","select_metrics":["clicks"],"group_by":[{"name":"ad_group","property":"name"}],"filters":[],"gbt":"date","fr":1611426600,"to":1612031399},{"channel":"facebook_ads","select_metrics":["clicks"],"group_by":[{"name":"ad_group","property":"name"}],"filters":[],"gbt":"","fr":1611426600,"to":1612031399}],"cl":"channel_v1"}`,
		model.QueryClassWeb:         `{"units":[{"unit_id":194,"query_name":"bounce_rate"},{"unit_id":195,"query_name":"unique_users"},{"unit_id":196,"query_name":"avg_session_duration"},{"unit_id":197,"query_name":"avg_pages_per_session"},{"unit_id":200,"query_name":"sessions"},{"unit_id":201,"query_name":"total_page_view"},{"unit_id":199,"query_name":"traffic_channel_report"},{"unit_id":198,"query_name":"top_pages_report"}],"custom_group_units":[],"from":1609612200,"to":1610044199}`,
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

func TestQueryCachingFailedCondition(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	var waitGroup sync.WaitGroup
	var badQueriesStr = map[string]string{
		// Bad query type for insights and funnel query.
		model.QueryClassInsights: `{"cl": "insights", "ec": "any_given_event", "fr": 1393612200, "to": 1396290599, "ty": "events_occurrences", "tz": "", "ewp": [{"na": "$session", "pr": []}], "gbp": [], "gbt": ""}`,
		model.QueryClassFunnel:   `{"cl": "funnel", "ec": "any_given_event", "fr": 1594492200, "to": 1594578599, "ty": "unique_userss", "tz": "Asia/Calcutta", "ewp": [{"na": "$session", "pr": []}, {"na": "www.chargebee.com/schedule-a-demo", "pr": []}], "gbp": [], "gbt": ""}`,

		// Attribution and channel query should fail as no customer account id is created for project in test.
		model.QueryClassAttribution: `{"cl": "attribution", "meta": {"metrics_breakdown": true}, "query": {"ce": {"na": "$session", "pr": []}, "cm": ["Impressions", "Clicks", "Spend"], "to": 1596479399, "lbw": 1, "lfe": [], "from": 1595874600, "attribution_key": "Campaign", "attribution_methodology": "Last_Touch"}}`,
		model.QueryClassChannel:     `{"cl": "channel", "meta": {"metric": "total_cost"}, "query": {"to": 1576060774, "from": 1573468774, "channel": "google_ads", "filter_key": "campaign", "filter_value": "all"}}`,
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
		_, status = store.GetStore().UpdateUserGroup(project.ID, userID, model.GROUP_NAME_HUBSPOT_COMPANY, group1ID, groupUserIDMap[group1ID])
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
