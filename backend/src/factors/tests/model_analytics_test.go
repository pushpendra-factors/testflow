package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	H "factors/handler"
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

		user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, user.ID)

		occurrenceByIndex := []int{0, 1, 2}
		for index, eventIndex := range occurrenceByIndex {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], user.ID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := M.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []M.QueryProperty{},
				},
			},
			Class: M.QueryClassFunnel,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		// steps headers avalilable.
		assert.Equal(t, M.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, M.StepPrefix+"1", result.Headers[1])
		// no.of users should be 1.
		assert.Equal(t, int64(1), result.Rows[0][0].(int64))
		assert.Equal(t, int64(1), result.Rows[0][1].(int64))
	})

	t.Run("NoOfUsersDidNotCompleteFunnelOnFirstTimeOfStart:1", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 3; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, user.ID)

		// user did only 0 first few times, did only 1 few times then 2.
		occurrenceByIndexUser1 := []int{0, 0, 0, 1, 1, 2}

		for index, eventIndex := range occurrenceByIndexUser1 {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], user.ID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := M.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[2],
					Properties: []M.QueryProperty{},
				},
			},
			Class: M.QueryClassFunnel,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		assert.Equal(t, M.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, M.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result.Headers[2])
		assert.Equal(t, M.StepPrefix+"2", result.Headers[3])

		assert.Equal(t, int64(1), result.Rows[0][0], "step0")
		assert.Equal(t, int64(1), result.Rows[0][1], "step1")
		assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")
		assert.Equal(t, int64(1), result.Rows[0][3], "step3")
	})

	t.Run("NoOfUsersDidNotCompleteFunnelOnFirstTimeOfStart:2", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 4; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, user.ID)

		occurrenceByIndexUser1 := []int{0, 0, 0, 1, 1, 0, 2}
		for index, eventIndex := range occurrenceByIndexUser1 {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], user.ID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := M.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[2],
					Properties: []M.QueryProperty{},
				},
			},
			Class: M.QueryClassFunnel,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		assert.Equal(t, M.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, M.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result.Headers[2])
		assert.Equal(t, M.StepPrefix+"2", result.Headers[3])

		assert.Equal(t, int64(1), result.Rows[0][0], "step0")
		assert.Equal(t, int64(1), result.Rows[0][1], "step1")
		assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")
		assert.Equal(t, int64(1), result.Rows[0][3], "step2")
	})

	t.Run("NoOfUsersDidNotCompleteFunnelOnFirstTimeOfStart:3", func(t *testing.T) {
		project, err := SetupProjectReturnDAO()
		assert.Nil(t, err)

		eventNames := make([]string, 0, 0)
		for i := 0; i < 4; i++ {
			eventNames = append(eventNames, U.RandomLowerAphaNumString(8))
		}
		eventTimestamp := U.UnixTimeBeforeDuration(24 * 10 * time.Hour) // 10 days before.

		user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotEmpty(t, user.ID)

		occurrenceByIndexUser1 := []int{0, 0, 0, 1, 1, 0, 2, 1}
		for index, eventIndex := range occurrenceByIndexUser1 {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
				eventNames[eventIndex], user.ID, eventTimestamp+int64(index))
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
		}

		query := M.Query{
			From: eventTimestamp,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name:       eventNames[0],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[1],
					Properties: []M.QueryProperty{},
				},
				M.QueryEventWithProperties{
					Name:       eventNames[2],
					Properties: []M.QueryProperty{},
				},
			},
			Class: M.QueryClassFunnel,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)

		assert.Equal(t, M.StepPrefix+"0", result.Headers[0])
		assert.Equal(t, M.StepPrefix+"1", result.Headers[1])
		assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result.Headers[2])
		assert.Equal(t, M.StepPrefix+"2", result.Headers[3])

		assert.Equal(t, int64(1), result.Rows[0][0], "step0")
		assert.Equal(t, int64(1), result.Rows[0][1], "step1")
		assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")
		assert.Equal(t, int64(1), result.Rows[0][3], "step2")
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

	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user3.ID)

	user4, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user4.ID)

	payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		eventNames[2], user3.ID, eventTimestamp+100)
	w1 := ServePostRequestWithHeaders(r, trackURI, []byte(payload1), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w1.Code)
	response1 := DecodeJSONResponseToMap(w1.Body)
	assert.NotNil(t, response1["event_id"])

	payload2 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		eventNames[3], user4.ID, eventTimestamp+200)
	w2 := ServePostRequestWithHeaders(r, trackURI, []byte(payload2), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w2.Code)
	response2 := DecodeJSONResponseToMap(w2.Body)
	assert.NotNil(t, response2["event_id"])

	// identify users with same customer_user_id.
	identifyURI := "/sdk/user/identify"
	customerUserId := U.RandomLowerAphaNumString(15)
	w := ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user3.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user4.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	query := M.Query{
		From: eventTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       eventNames[2],
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       eventNames[3],
				Properties: []M.QueryProperty{},
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result1, errCode, _ := M.Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, result1)

	// steps headers avalilable.
	assert.Equal(t, M.StepPrefix+"0", result1.Headers[0])
	assert.Equal(t, M.StepPrefix+"1", result1.Headers[1])
	// no.of users should be 1 after identification.
	assert.Equal(t, int64(1), result1.Rows[0][0].(int64))
	assert.Equal(t, int64(1), result1.Rows[0][1].(int64))
}

func TestAnalyticsFunnelQueryWithFilterCondition(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user.ID)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// s0 event property value with 5.
	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 5}}`,
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
		payload1 := fmt.Sprintf(`{"event_name": "%s", "timestamp": %d, "event_properties": {"value": 10}}`,
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
		payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
			"s1", userIds[i], stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		stepTimestamp = stepTimestamp + 10
	}

	query := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:    M.PropertyEntityEvent,
						Property:  "value",
						Operator:  "greaterThan",
						Value:     "5",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result, errCode, _ := M.Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, M.StepPrefix+"0", result.Headers[0])
	assert.Equal(t, M.StepPrefix+"1", result.Headers[1])
	// all 5 users who performed s0 with value greater
	// 5 has performed s1.
	assert.Equal(t, int64(5), result.Rows[0][0], "step0")
	assert.Equal(t, int64(5), result.Rows[0][1], "step1")

	query1 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:    M.PropertyEntityEvent,
						Property:  "value",
						Operator:  "lesserThan",
						Value:     "11",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result1, errCode, _ := M.Analyze(project.ID, query1)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, M.StepPrefix+"0", result1.Headers[0])
	assert.Equal(t, M.StepPrefix+"1", result1.Headers[1])
	// among 10 users who performed s0 with value lesser
	// than 11, 5 users has performed s1.
	assert.Equal(t, int64(10), result1.Rows[0][0], "step0")
	assert.Equal(t, int64(5), result1.Rows[0][1], "step1")
	assert.Equal(t, "50.0", result1.Rows[0][2], "conversion_step_0_step_1")

	query2 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:    M.PropertyEntityEvent,
						Property:  "value",
						Operator:  "equals",
						Value:     "10",
						Type:      U.PropertyTypeNumerical,
						LogicalOp: "AND",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result2, errCode, _ := M.Analyze(project.ID, query2)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, M.StepPrefix+"0", result2.Headers[0])
	assert.Equal(t, M.StepPrefix+"1", result2.Headers[1])
	// all users performed s0 with value=10 has performed s1.
	assert.Equal(t, int64(5), result2.Rows[0][0], "step0")
	assert.Equal(t, int64(5), result2.Rows[0][1], "step1")
}

func TestAnalyticsInsightsQuery(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	t.Run("OperatorsWithNumericalPropertiesOnWhere", func(t *testing.T) {
		project, user, eventName, err := SetupProjectUserEventNameReturnDAO()
		assert.Nil(t, err)

		var firstEvent *M.Event

		// 10 times: page_spent_time as 5
		for i := 0; i < 10; i++ {
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "event_properties": {"$page_spent_time" : %d}}`, eventName.Name, user.ID, 5)
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
			response := DecodeJSONResponseToMap(w.Body)
			assert.NotNil(t, response["event_id"])
			if i == 0 {
				event, errCode := M.GetEventById(project.ID, response["event_id"].(string))
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
		query := M.Query{
			From: firstEvent.Timestamp - 10,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: eventName.Name,
					Properties: []M.QueryProperty{
						M.QueryProperty{
							Entity:    M.PropertyEntityEvent,
							Property:  "$page_spent_time",
							Operator:  "greaterThan",
							Type:      "numerical",
							LogicalOp: "AND",
							Value:     "11",
						},
					},
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeEventsOccurrence,
			EventsCondition: M.EventCondAnyGivenEvent,
		}

		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(5), result.Rows[0][0])

		// Query count of events: page_spent_time > 11
		query2 := M.Query{
			From: firstEvent.Timestamp - 10,
			To:   time.Now().UTC().Unix(),
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: eventName.Name,
					Properties: []M.QueryProperty{
						M.QueryProperty{
							Entity:    M.PropertyEntityEvent,
							Property:  "$page_spent_time",
							Operator:  "greaterThan",
							Type:      "numerical",
							LogicalOp: "AND",
							Value:     "4",
						},
					},
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeEventsOccurrence,
			EventsCondition: M.EventCondAnyGivenEvent,
		}

		result2, errCode, _ := M.Analyze(project.ID, query2)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, "count", result2.Headers[0])
		assert.Equal(t, int64(15), result2.Rows[0][0])
	})
}

func TestAnalyticsInsightsQueryWithFilterAndBreakdown(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)
	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)
	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user3.ID)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	/*
		user1 -> event s0 with property1 -> s0 with property2 -> s1 with propterty2
		user2 -> event s0 with property1 -> s1 with property1
		user3 -> event s0 with property2 -> s1 with property2
	*/

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", user1.ID, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", user1.ID, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s1", user1.ID, stepTimestamp+20, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", user2.ID, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s1", user2.ID, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s0", user3.ID, stepTimestamp, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, "user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`, "s1", user3.ID, stepTimestamp+10, "B", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("AnalyticsInsightsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

		query := M.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "s0",
					Properties: []M.QueryProperty{
						M.QueryProperty{
							Entity:    M.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "A",
						},
					},
				},
				M.QueryEventWithProperties{
					Name: "s1",
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

		//unique user count should return 2 for s0 to s1 with fliter property2
		query.EventsWithProperties[0].Properties[0].Value = "B"
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

		query = M.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "s0",
				},
				M.QueryEventWithProperties{
					Name: "s1",
				},
			},
			GroupByProperties: []M.QueryGroupByProperty{
				M.QueryGroupByProperty{
					Entity:   M.PropertyEntityUser,
					Property: "$initial_source",
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		//breakdown by user property should return property A with 1 count and property B with 2 count
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "$initial_source", result.Headers[0])
		assert.Equal(t, "count", result.Headers[1])
		assert.Equal(t, "B", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "A", result.Rows[1][0])
		assert.Equal(t, int64(1), result.Rows[1][1])
	})
	t.Run("AnalyticsInsightsQueryUniqueUserWithEventPropertyFilterAndBreakdown", func(t *testing.T) {
		query := M.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "s0",
					Properties: []M.QueryProperty{
						M.QueryProperty{
							Entity:    M.PropertyEntityEvent,
							Property:  "$campaign_id",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "1234",
						},
					},
				},
				M.QueryEventWithProperties{
					Name: "s1",
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

		query.EventsWithProperties[0].Properties[0].Value = "4321"
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

		query = M.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "s0",
				},
				M.QueryEventWithProperties{
					Name: "s1",
				},
			},
			GroupByProperties: []M.QueryGroupByProperty{
				M.QueryGroupByProperty{
					Entity:   M.PropertyEntityEvent,
					Property: "$campaign_id",
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeUniqueUsers,
			EventsCondition: M.EventCondAllGivenEvent,
		}
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, "$campaign_id", result.Headers[0])
		assert.Equal(t, "1234", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "4321", result.Rows[1][0])
		assert.Equal(t, int64(1), result.Rows[1][1])
	})

	t.Run("AnalyticsInsightsQueryEventOccurrenceWithCountEventOccurrences", func(t *testing.T) {
		query := M.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []M.QueryEventWithProperties{
				M.QueryEventWithProperties{
					Name: "s0",
					Properties: []M.QueryProperty{
						M.QueryProperty{
							Entity:    M.PropertyEntityUser,
							Property:  "$initial_source",
							Operator:  "equals",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "B",
						},
					},
				},
				M.QueryEventWithProperties{
					Name: "s1",
				},
			},
			Class: M.QueryClassInsights,

			Type:            M.QueryTypeEventsOccurrence,
			EventsCondition: M.EventCondAllGivenEvent,
		}

		/*
			Event occurrence with user property should give 5
			user1 -> 		 -> s0 with property2 -> s1 with propterty2
			user2 -> 		 -> s1 with property1
			user3 -> event s0 with property2 -> s1 with property2
		*/
		result, errCode, _ := M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(5), result.Rows[0][0])

		query.GroupByProperties = []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:   M.PropertyEntityUser,
				Property: "$initial_source",
			},
		}
		// property2 -> 4, property1 ->1
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "$initial_source", result.Headers[0])
		assert.Equal(t, "B", result.Rows[0][0])
		assert.Equal(t, int64(4), result.Rows[0][1])
		assert.Equal(t, "A", result.Rows[1][0])
		assert.Equal(t, int64(1), result.Rows[1][1])

		//Count should be same as when done with user property = 5
		query.EventsWithProperties[0].Properties[0].Entity = M.PropertyEntityEvent
		query.EventsWithProperties[0].Properties[0].Property = "$campaign_id"
		query.EventsWithProperties[0].Properties[0].Value = "1234"
		query.GroupByProperties = []M.QueryGroupByProperty{}
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(5), result.Rows[0][0])
	})
}
