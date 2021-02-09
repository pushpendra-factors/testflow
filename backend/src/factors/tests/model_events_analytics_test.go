package tests

import (
	"encoding/json"
	C "factors/config"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"fmt"

	"github.com/gin-gonic/gin"
)

func TestEventAnalyticsQuery(t *testing.T) {
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
				event, errCode := store.GetStore().GetEventById(project.ID, response["event_id"].(string))
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(5), result.Rows[0][0])

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		result2, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query2)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, "count", result2.Headers[0])
		assert.Equal(t, int64(15), result2.Rows[0][0])

	})
}

func TestEventAnalyticsQueryWithFilterAndBreakdown(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)
	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)
	user3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
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

	t.Run("AnalyticsEventsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

		//unique user count should return 2 for s0 to s1 with fliter property2
		query.EventsWithProperties[0].Properties[0].Value = "B"
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//breakdown by user property should return property A with 1 count and property B with 2 count
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "$initial_source", result.Headers[0])
		assert.Equal(t, "count", result.Headers[1])
		assert.Equal(t, "B", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "A", result.Rows[1][0])
		assert.Equal(t, int64(1), result.Rows[1][1])
	})
	t.Run("AnalyticsEventsQueryUniqueUserWithEventPropertyFilterAndBreakdown", func(t *testing.T) {
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

		query.EventsWithProperties[0].Properties[0].Value = "4321"
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "count", result.Headers[0])
		assert.Equal(t, int64(2), result.Rows[0][0])

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, "$campaign_id", result.Headers[0])
		assert.Equal(t, "1234", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "4321", result.Rows[1][0])
		// Counting all occurrences instead of first. So for user1, both 4321 and 1234 will be counted.
		assert.Equal(t, int64(2), result.Rows[1][1])
	})

	t.Run("AnalyticsEventsQueryEventOccurrenceWithCountEventOccurrences", func(t *testing.T) {
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		/*
			Event occurrence with user property should give 5
			user1 -> 		 -> s0 with property2 -> s1 with propterty2
			user2 -> 		 -> s1 with property1
			user3 -> event s0 with property2 -> s1 with property2
		*/
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "count", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, int64(3), result.Rows[1][1])

		query.GroupByProperties = []model.QueryGroupByProperty{
			model.QueryGroupByProperty{
				Entity:   model.PropertyEntityUser,
				Property: "$initial_source",
			},
		}
		// property2 -> 4, property1 ->1
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query)
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
		query.EventsWithProperties[0].Properties[0].Entity = model.PropertyEntityEvent
		query.EventsWithProperties[0].Properties[0].Property = "$campaign_id"
		query.EventsWithProperties[0].Properties[0].Value = "1234"
		query.GroupByProperties = []model.QueryGroupByProperty{}
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "count", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, int64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, int64(3), result.Rows[1][1])
	})

	// Test for event filter on user property and group by user property at the same event.
	t.Run("AnalyticsEventsQueryUniqueUserWithUserPropertyFilterAndUserBreakdown", func(t *testing.T) {
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
	})
}

func TestEventAnalyticsQueryWithNumericalBucketing(t *testing.T) {
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
			iUser, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
				`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":%d}}`,
				eventName1, iUser.ID, startTimestamp+10, i, i)
			w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
			assert.Equal(t, http.StatusOK, w.Code)
		}

		query := model.Query{
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
		result, errCode, _ := store.GetStore().Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		validateNumericalBucketRanges(t, result, numPropertyRangeStart, numPropertyRangeEnd, 0)

		/*
			New event with $page_load_time = 0
			total element 21

			User property numerical_property set as empty ($none).
			Will create 11 buckets. including 1 $none.
		*/
		iUser, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":""}}`,
			eventName1, iUser.ID, startTimestamp+10, 0)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		result, errCode, _ = store.GetStore().Analyze(project.ID, query)
		validateNumericalBucketRanges(t, result, 0, numPropertyRangeEnd, 0)

		// Using group by numerical property.
		query.GroupByProperties[0].Entity = model.PropertyEntityUser
		query.GroupByProperties[0].Property = "numerical_property"
		result, errCode, _ = store.GetStore().Analyze(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		validateNumericalBucketRanges(t, result, numPropertyRangeStart, numPropertyRangeEnd, 1)
	})
}

func TestEventAnalyticsQueryGroupSingleQueryHandler(t *testing.T) {

	t.Run("GroupQuerySingleQuery", func(t *testing.T) {

		// Initialize routes and dependent data.
		r := gin.Default()
		H.InitAppRoutes(r)
		H.InitSDKServiceRoutes(r)
		uri := "/sdk/event/track"
		project, user, eventName, agent, err := SetupProjectUserEventNameAgentReturnDAO()
		assert.Nil(t, err)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.NotNil(t, agent)
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
				event, errCode := store.GetStore().GetEventById(project.ID, response["event_id"].(string))
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
		query1 := model.Query{
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		queryGroup := model.QueryGroup{}
		queryGroup.Queries = make([]model.Query, 0)
		queryGroup.Queries = append(queryGroup.Queries, query1)

		w := sendEventsQueryHandler(r, project.ID, agent, &queryGroup)
		assert.Equal(t, http.StatusOK, w.Code)
		var resultGroup model.ResultGroup
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&resultGroup); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, 1, len(resultGroup.Results))
		result1 := resultGroup.Results[0]
		assert.NotNil(t, result1)
		assert.Equal(t, "count", result1.Headers[0])
		assert.Equal(t, float64(5), result1.Rows[0][0])

	})
}

func TestEventAnalyticsQueryGroupMultiQueryHandler(t *testing.T) {

	t.Run("GroupQueryMultiQuery", func(t *testing.T) {

		// Initialize routes and dependent data.
		r := gin.Default()
		H.InitAppRoutes(r)
		H.InitSDKServiceRoutes(r)
		uri := "/sdk/event/track"
		project, user, eventName, agent, err := SetupProjectUserEventNameAgentReturnDAO()
		assert.Nil(t, err)
		assert.NotNil(t, project)
		assert.NotNil(t, user)
		assert.NotNil(t, agent)
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
				event, errCode := store.GetStore().GetEventById(project.ID, response["event_id"].(string))
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
		query1 := model.Query{
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		queryGroup := model.QueryGroup{}
		queryGroup.Queries = make([]model.Query, 0)
		queryGroup.Queries = append(queryGroup.Queries, query1)
		queryGroup.Queries = append(queryGroup.Queries, query2)

		w := sendEventsQueryHandler(r, project.ID, agent, &queryGroup)
		assert.Equal(t, http.StatusOK, w.Code)
		var resultGroup model.ResultGroup
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&resultGroup); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, 2, len(resultGroup.Results))

		result1 := resultGroup.Results[0]
		assert.NotNil(t, result1)
		assert.Equal(t, "count", result1.Headers[0])
		assert.Equal(t, float64(5), result1.Rows[0][0])

		result2 := resultGroup.Results[1]
		assert.NotNil(t, result2)
		assert.Equal(t, "count", result2.Headers[0])
		assert.Equal(t, float64(15), result2.Rows[0][0])

	})
}

func sendEventsQueryHandler(r *gin.Engine, projectId uint64, agent *model.Agent, queryGroup *model.QueryGroup) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := U.NewRequestBuilder(http.MethodPost, fmt.Sprintf("/projects/%d/v1/query", projectId)).
		WithPostParams(queryGroup).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error with request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEventAnalyticsEachEventQueryWithFilterAndBreakdown(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	user1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user1.ID)
	user2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, user2.ID)
	user3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID})
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

	t.Run("AnalyticsEachEventsQueryUniqueUserWithUserPropertyFilterAndBreakdown", func(t *testing.T) {

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		if result.Rows[0][1] == "s1" {
			assert.Equal(t, int64(3), result.Rows[0][2])
		} else {
			assert.Equal(t, int64(2), result.Rows[0][2])
		}
		if result.Rows[1][1] == "s1" {
			assert.Equal(t, int64(3), result.Rows[1][2])
		} else {
			assert.Equal(t, int64(2), result.Rows[1][2])
		}
	})

	t.Run("AnalyticsEachEventsQueryEventOccurrenceBreakdownByDate", func(t *testing.T) {
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
			Class:            model.QueryClassEvents,
			GroupByTimestamp: "date",
			Type:             model.QueryTypeEventsOccurrence,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "datetime", result.Headers[0])
		if result.Headers[1] == "s0" {
			assert.Equal(t, int64(2), result.Rows[0][1])
			assert.Equal(t, int64(3), result.Rows[0][2])
		} else {
			assert.Equal(t, int64(3), result.Rows[0][1])
			assert.Equal(t, int64(2), result.Rows[0][2])
		}
	})
}
