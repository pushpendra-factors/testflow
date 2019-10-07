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
	H.InitSDKRoutes(r)
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
		assert.Equal(t, M.StepPrefix+"2", result.Headers[2])

		assert.Equal(t, int64(1), result.Rows[0][0], "step0")
		assert.Equal(t, int64(1), result.Rows[0][1], "step1")
		assert.Equal(t, int64(1), result.Rows[0][2], "step2")
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
		assert.Equal(t, M.StepPrefix+"2", result.Headers[2])

		assert.Equal(t, int64(1), result.Rows[0][0], "step0")
		assert.Equal(t, int64(1), result.Rows[0][1], "step1")
		assert.Equal(t, int64(1), result.Rows[0][2], "step2")
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
		assert.Equal(t, M.StepPrefix+"2", result.Headers[2])

		assert.Equal(t, int64(1), result.Rows[0][0], "step0")
		assert.Equal(t, int64(1), result.Rows[0][1], "step1")
		assert.Equal(t, int64(1), result.Rows[0][2], "step2")
	})
}
