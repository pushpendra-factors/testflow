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
					M.QueryProperty{
						Entity:   M.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
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
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result.Headers[2])
	// all 5 users who performed s0 with value greater
	// 5 has performed s1.
	assert.Equal(t, int64(5), result.Rows[0][0], "step0")
	assert.Equal(t, int64(5), result.Rows[0][1], "step1")
	assert.Equal(t, "100.0", result.Rows[0][2], "conversion_step_0_step_1")

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

	query3 := M.Query{
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
				Name: "s1",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:   M.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
					},
				},
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result3, errCode, _ := M.Analyze(project.ID, query3)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, M.StepPrefix+"0", result3.Headers[0])
	assert.Equal(t, M.StepPrefix+"1", result3.Headers[1])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result3.Headers[2])
	assert.Equal(t, int64(5), result3.Rows[0][0], "step0")
	assert.Equal(t, int64(5), result3.Rows[0][1], "step1")
	assert.Equal(t, "100.0", result3.Rows[0][2], "conversion_step_0_step_1")
}

func TestAnalyticsFunnelQueryRepeatedEvents(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	for i := 0; i < 5; i++ {
		payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
			"s1", user1.ID, stepTimestamp)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)
		response := DecodeJSONResponseToMap(w.Body)
		assert.NotNil(t, response["event_id"])
		stepTimestamp = stepTimestamp + 10
	}

	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)
	payload1 := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s", "timestamp": %d}`,
		"s1", user2.ID, startTimestamp)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload1), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	query := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
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

	assert.Equal(t, int64(2), result.Rows[0][0])
	assert.Equal(t, int64(1), result.Rows[0][1])
	assert.Equal(t, "50.0", result.Rows[0][2])
	assert.Equal(t, "50.0", result.Rows[0][3])

	identifyURI := "/sdk/user/identify"
	customerUserId := U.RandomLowerAphaNumString(15)
	w = ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user1.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	w = ServePostRequestWithHeaders(r, identifyURI, []byte(fmt.Sprintf(`{"c_uid": "%s", "user_id": "%s"}`,
		customerUserId, user2.ID)), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)

	query1 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
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

	assert.Equal(t, int64(1), result1.Rows[0][0])
	assert.Equal(t, int64(1), result1.Rows[0][1])
	assert.Equal(t, "100.0", result1.Rows[0][2])
	assert.Equal(t, int64(1), result1.Rows[0][3])
	assert.Equal(t, "100.0", result1.Rows[0][4])
	assert.Equal(t, "100.0", result1.Rows[0][5])
}
func TestAnalyticsFunnelQueryWithFilterAndBreakDown(t *testing.T) {
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
	//x1 -> x2
	// (breakdown by user_property u1)
	query := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "s0",
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "gender",
				EventName: M.UserPropertyGroupByPresent,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result, errCode, _ := M.Analyze(project.ID, query)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result.Headers[0])
	assert.Equal(t, M.StepPrefix+"0", result.Headers[1])
	assert.Equal(t, M.StepPrefix+"1", result.Headers[2])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result.Headers[3])

	assert.Equal(t, "$no_group", result.Rows[0][0])
	assert.Equal(t, int64(10), result.Rows[0][1])
	assert.Equal(t, int64(5), result.Rows[0][2])
	assert.Equal(t, "50.0", result.Rows[0][3])

	assert.Equal(t, "M", result.Rows[1][0])
	assert.Equal(t, int64(5), result.Rows[1][1])
	assert.Equal(t, 0, result.Rows[1][2])
	assert.Equal(t, "0.0", result.Rows[1][3])

	assert.Equal(t, "F", result.Rows[2][0])
	assert.Equal(t, int64(5), result.Rows[2][1])
	assert.Equal(t, int64(5), result.Rows[2][2])
	assert.Equal(t, "100.0", result.Rows[2][3])

	// 	x1 -> x2
	// (breakdown by event x1 event_property ep1)
	query1 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "s0",
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}

	result1, errCode, _ := M.Analyze(project.ID, query1)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value", result1.Headers[0])
	assert.Equal(t, M.StepPrefix+"0", result1.Headers[1])
	assert.Equal(t, M.StepPrefix+"1", result1.Headers[2])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result1.Headers[3])

	assert.Equal(t, "$no_group", result1.Rows[0][0])
	assert.Equal(t, int64(10), result1.Rows[0][1])
	assert.Equal(t, int64(5), result1.Rows[0][2])
	assert.Equal(t, "50.0", result1.Rows[0][3])

	assert.Equal(t, "5", result1.Rows[1][0])
	assert.Equal(t, int64(5), result1.Rows[1][1])
	assert.Equal(t, 0, result1.Rows[1][2])
	assert.Equal(t, "0.0", result1.Rows[1][3])

	assert.Equal(t, "10", result1.Rows[2][0])
	assert.Equal(t, int64(5), result1.Rows[2][1])
	assert.Equal(t, int64(5), result1.Rows[2][2])
	assert.Equal(t, "100.0", result1.Rows[2][3])

	// 	x1 -> x2
	// (breakdown by event x1 event_property ep1) and (breakdown by event x2 event_property ep2)
	query2 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "s0",
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "id",
				EventName:      "s1",
				EventNameIndex: 2,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}
	result2, errCode, _ := M.Analyze(project.ID, query2)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value", result2.Headers[0])
	assert.Equal(t, "id", result2.Headers[1])
	assert.Equal(t, M.StepPrefix+"0", result2.Headers[2])
	assert.Equal(t, M.StepPrefix+"1", result2.Headers[3])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result2.Headers[4])

	assert.Equal(t, "$no_group", result2.Rows[0][0])
	assert.Equal(t, "$no_group", result2.Rows[0][1])
	assert.Equal(t, int64(10), result2.Rows[0][2])
	assert.Equal(t, int64(5), result2.Rows[0][3])
	assert.Equal(t, "50.0", result2.Rows[0][4])
	assert.Equal(t, 3, len(result2.Rows))

	assert.Equal(t, "5", result2.Rows[1][0])
	assert.Equal(t, "$none", result2.Rows[1][1])
	assert.Equal(t, int64(5), result2.Rows[1][2])
	assert.Equal(t, 0, result2.Rows[1][3])
	assert.Equal(t, "0.0", result2.Rows[1][4])

	assert.Equal(t, "10", result2.Rows[2][0])
	assert.Equal(t, "3", result2.Rows[2][1])
	assert.Equal(t, int64(5), result2.Rows[2][2])
	assert.Equal(t, int64(5), result2.Rows[2][3])
	assert.Equal(t, "100.0", result2.Rows[2][4])

	// x1 -> x2
	// (breakdown by user_property up1) and (breakdown by event x1 event_property ep1)
	query3 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "s0",
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "gender",
				EventName: M.UserPropertyGroupByPresent,
			},
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}
	result3, errCode, _ := M.Analyze(project.ID, query3)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result3.Headers[0])
	assert.Equal(t, "value", result3.Headers[1])
	assert.Equal(t, M.StepPrefix+"0", result3.Headers[2])
	assert.Equal(t, M.StepPrefix+"1", result3.Headers[3])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result3.Headers[4])

	assert.Equal(t, 3, len(result3.Rows))
	assert.Equal(t, "$no_group", result3.Rows[0][0])
	assert.Equal(t, "$no_group", result3.Rows[0][1])
	assert.Equal(t, int64(10), result3.Rows[0][2])
	assert.Equal(t, int64(5), result3.Rows[0][3])
	assert.Equal(t, "50.0", result3.Rows[0][4])

	assert.Equal(t, "M", result3.Rows[1][0])
	assert.Equal(t, "5", result3.Rows[1][1])
	assert.Equal(t, int64(5), result3.Rows[1][2])
	assert.Equal(t, 0, result3.Rows[1][3])
	assert.Equal(t, "0.0", result3.Rows[1][4])

	assert.Equal(t, "F", result3.Rows[2][0])
	assert.Equal(t, "10", result3.Rows[2][1])
	assert.Equal(t, int64(5), result3.Rows[2][2])
	assert.Equal(t, int64(5), result3.Rows[2][3])
	assert.Equal(t, "100.0", result3.Rows[2][4])

	// 	x1 (with event_property ep1 = ev1) -> x2
	// (breakdown by event x1 event_property ep1) and (breakdown by event x2 event_property ep2)
	query4 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:   M.PropertyEntityEvent,
						Property: "value",
						Operator: "equals",
						Value:    "10",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "id",
				EventName:      "s1",
				EventNameIndex: 2,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}
	result4, errCode, _ := M.Analyze(project.ID, query4)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "value", result4.Headers[0])
	assert.Equal(t, "id", result4.Headers[1])
	assert.Equal(t, M.StepPrefix+"0", result4.Headers[2])
	assert.Equal(t, M.StepPrefix+"1", result4.Headers[3])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result4.Headers[4])

	assert.Equal(t, 2, len(result4.Rows))
	assert.Equal(t, "$no_group", result4.Rows[0][0])
	assert.Equal(t, "$no_group", result4.Rows[0][1])
	assert.Equal(t, int64(5), result4.Rows[0][2])
	assert.Equal(t, int64(5), result4.Rows[0][3])
	assert.Equal(t, "100.0", result4.Rows[0][4])

	assert.Equal(t, "10", result4.Rows[1][0])
	assert.Equal(t, "3", result4.Rows[1][1])
	assert.Equal(t, int64(5), result4.Rows[1][2])
	assert.Equal(t, int64(5), result4.Rows[1][3])
	assert.Equal(t, "100.0", result4.Rows[1][4])

	// x1 (with event_property ep1 = ev1) -> x2
	// (breakdown by user_property up1) and (breakdown by user_property up2)
	query5 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:   M.PropertyEntityEvent,
						Property: "value",
						Operator: "equals",
						Value:    "10",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "gender",
				EventName: M.UserPropertyGroupByPresent,
			},
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "age",
				EventName: M.UserPropertyGroupByPresent,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}
	result5, errCode, _ := M.Analyze(project.ID, query5)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result5.Headers[0])
	assert.Equal(t, "age", result5.Headers[1])
	assert.Equal(t, M.StepPrefix+"0", result5.Headers[2])
	assert.Equal(t, M.StepPrefix+"1", result5.Headers[3])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result5.Headers[4])

	assert.Equal(t, 2, len(result5.Rows))
	assert.Equal(t, "$no_group", result5.Rows[0][0])
	assert.Equal(t, "$no_group", result5.Rows[0][1])
	assert.Equal(t, int64(5), result5.Rows[0][2])
	assert.Equal(t, int64(5), result5.Rows[0][3])
	assert.Equal(t, "100.0", result5.Rows[0][4])

	assert.Equal(t, "F", result5.Rows[1][0])
	assert.Equal(t, "21", result5.Rows[1][1])
	assert.Equal(t, int64(5), result5.Rows[1][2])
	assert.Equal(t, int64(5), result5.Rows[1][3])
	assert.Equal(t, "100.0", result5.Rows[1][4])

	// 	x1 (user_property up1 = uv1) -> x2
	// (breakdown by user_property up1) and (breakdown by user_property up2)
	query6 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:   M.PropertyEntityUser,
						Property: "gender",
						Operator: "equals",
						Value:    "F",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "gender",
				EventName: M.UserPropertyGroupByPresent,
			},
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "age",
				EventName: M.UserPropertyGroupByPresent,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}
	result6, errCode, _ := M.Analyze(project.ID, query6)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result6.Headers[0])
	assert.Equal(t, "age", result6.Headers[1])
	assert.Equal(t, M.StepPrefix+"0", result6.Headers[2])
	assert.Equal(t, M.StepPrefix+"1", result6.Headers[3])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result6.Headers[4])

	assert.Equal(t, 2, len(result6.Rows))
	assert.Equal(t, "$no_group", result6.Rows[0][0])
	assert.Equal(t, "$no_group", result6.Rows[0][1])
	assert.Equal(t, int64(5), result6.Rows[0][2])
	assert.Equal(t, int64(5), result6.Rows[0][3])
	assert.Equal(t, "100.0", result6.Rows[0][4])

	assert.Equal(t, "F", result6.Rows[1][0])
	assert.Equal(t, "21", result6.Rows[1][1])
	assert.Equal(t, int64(5), result6.Rows[1][2])
	assert.Equal(t, int64(5), result6.Rows[1][3])
	assert.Equal(t, "100.0", result6.Rows[1][4])

	// 	x1 (user_property up1 = uv1) -> x2
	// (breakdown by user_property up1) and (breakdown by event x1 event_property ep1)
	query7 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name: "s0",
				Properties: []M.QueryProperty{
					M.QueryProperty{
						Entity:   M.PropertyEntityEvent,
						Property: "value",
						Operator: "equals",
						Value:    "10",
					},
				},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		GroupByProperties: []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:    M.PropertyEntityUser,
				Property:  "gender",
				EventName: M.UserPropertyGroupByPresent,
			},
			M.QueryGroupByProperty{
				Entity:         M.PropertyEntityEvent,
				Property:       "value",
				EventName:      "s0",
				EventNameIndex: 1,
			},
		},
		Class:           M.QueryClassFunnel,
		Type:            M.QueryTypeUniqueUsers,
		EventsCondition: M.EventCondAllGivenEvent,
	}
	result7, errCode, _ := M.Analyze(project.ID, query7)

	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, "gender", result7.Headers[0])
	assert.Equal(t, "value", result7.Headers[1])
	assert.Equal(t, M.StepPrefix+"0", result7.Headers[2])
	assert.Equal(t, M.StepPrefix+"1", result7.Headers[3])
	assert.Equal(t, M.FunnelConversionPrefix+M.StepPrefix+"0"+"_"+M.StepPrefix+"1", result7.Headers[4])

	assert.Equal(t, 2, len(result7.Rows))
	assert.Equal(t, "$no_group", result7.Rows[0][0])
	assert.Equal(t, "$no_group", result7.Rows[0][1])
	assert.Equal(t, int64(5), result7.Rows[0][2])
	assert.Equal(t, int64(5), result7.Rows[0][3])
	assert.Equal(t, "100.0", result7.Rows[0][4])

	assert.Equal(t, "F", result7.Rows[1][0])
	assert.Equal(t, "10", result7.Rows[1][1])
	assert.Equal(t, int64(5), result7.Rows[1][2])
	assert.Equal(t, int64(5), result7.Rows[1][3])
	assert.Equal(t, "100.0", result7.Rows[1][4])

	query8 := M.Query{
		From: startTimestamp,
		To:   time.Now().UTC().Unix(),
		EventsWithProperties: []M.QueryEventWithProperties{
			M.QueryEventWithProperties{
				Name:       "$session",
				Properties: []M.QueryProperty{},
			},
			M.QueryEventWithProperties{
				Name:       "s1",
				Properties: []M.QueryProperty{},
			},
		},
		Class:             M.QueryClassFunnel,
		Type:              M.QueryTypeUniqueUsers,
		EventsCondition:   M.EventCondAllGivenEvent,
		SessionStartEvent: 1,
		SessionEndEvent:   2,
	}

	result8, errCode, _ := M.Analyze(project.ID, query8)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, int64(10), result8.Rows[0][0])
	assert.Equal(t, int64(5), result8.Rows[0][1])
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
					Entity:         M.PropertyEntityEvent,
					Property:       "$campaign_id",
					EventName:      "s0",
					EventNameIndex: 1,
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
		// Counting all occurrences instead of first. So for user1, both 4321 and 1234 will be counted.
		assert.Equal(t, int64(2), result.Rows[1][1])
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
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "count", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, int64(3), result.Rows[1][1])

		query.GroupByProperties = []M.QueryGroupByProperty{
			M.QueryGroupByProperty{
				Entity:   M.PropertyEntityUser,
				Property: "$initial_source",
			},
		}
		// property2 -> 4, property1 ->1
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "$initial_source", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, "B", result.Rows[0][1])
		assert.Equal(t, int64(2), result.Rows[0][2])

		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, "B", result.Rows[1][1])
		assert.Equal(t, int64(2), result.Rows[1][2])

		assert.Equal(t, "s1", result.Rows[2][0])
		assert.Equal(t, "A", result.Rows[2][1])
		assert.Equal(t, int64(1), result.Rows[2][2])

		//Count should be same as when done with user property = 5
		query.EventsWithProperties[0].Properties[0].Entity = M.PropertyEntityEvent
		query.EventsWithProperties[0].Properties[0].Property = "$campaign_id"
		query.EventsWithProperties[0].Properties[0].Value = "1234"
		query.GroupByProperties = []M.QueryGroupByProperty{}
		result, errCode, _ = M.Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "count", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, int64(3), result.Rows[1][1])
	})
}
