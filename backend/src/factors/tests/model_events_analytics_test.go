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
	"sort"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		result2, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query2, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result2)
		assert.Equal(t, "aggregate", result2.Headers[0])
		assert.Equal(t, float64(15), result2.Rows[0][0])

	})
}

func TestEventAnalyticsQueryWithFilterAndBreakdown(t *testing.T) {
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
		user1 -> event s0 with property1 -> s0 with property2 -> s1 with property2
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

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])

		//unique user count should return 2 for s0 to s1 with filter property2
		query.EventsWithProperties[0].Properties[0].Value = "B"
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//breakdown by user property should return property A with 1 count and property B with 2 count
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "$initial_source", result.Headers[0])
		assert.Equal(t, "aggregate", result.Headers[1])
		assert.Equal(t, "B", result.Rows[0][0])
		assert.Equal(t, float64(2), result.Rows[0][1])
		assert.Equal(t, "A", result.Rows[1][0])
		assert.Equal(t, float64(1), result.Rows[1][1])
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

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])

		query.EventsWithProperties[0].Properties[0].Value = "4321"
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, "$campaign_id", result.Headers[0])
		expectedKeys := []string{"1234", "4321"}
		actualKeys := []string{result.Rows[0][0].(string), result.Rows[1][0].(string)}
		sort.Strings(actualKeys)
		assert.Equal(t, expectedKeys, actualKeys)
		assert.Equal(t, float64(2), result.Rows[0][1])
		// Counting all occurrences instead of first. So for createdUserID1, both 4321 and 1234 will be counted.
		assert.Equal(t, float64(2), result.Rows[1][1])
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
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
		result, errCode, _ = store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "event_name", result.Headers[0])
		assert.Equal(t, "aggregate", result.Headers[1])
		assert.Equal(t, "s0", result.Rows[0][0])
		assert.Equal(t, float64(2), result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][0])
		assert.Equal(t, float64(3), result.Rows[1][1])
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
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
			icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
			payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
				`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":%d}}`,
				eventName1, icreatedUserID, startTimestamp+10, i, i)
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
		result, errCode, _ := store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		validateNumericalBucketRanges(t, result, numPropertyRangeStart, numPropertyRangeEnd, 0)

		/*
			New event with $page_load_time = 0
			total element 21

			User property numerical_property set as empty ($none).
			Will create 11 buckets. including 1 $none.
		*/
		icreatedUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, `+
			`"event_properties":{"$page_load_time":%d},"user_properties":{"numerical_property":""}}`,
			eventName1, icreatedUserID, startTimestamp+10, 0)
		w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
		assert.Equal(t, http.StatusOK, w.Code)

		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		validateNumericalBucketRanges(t, result, 0, numPropertyRangeEnd, 0)

		// Using group by numerical property.
		query.GroupByProperties[0].Entity = model.PropertyEntityUser
		query.GroupByProperties[0].Property = "numerical_property"
		result, errCode, _ = store.GetStore().Analyze(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
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
		assert.Equal(t, "aggregate", result1.Headers[0])
		assert.Equal(t, float64(5), result1.Rows[0][0])

		queryJson, err := json.Marshal(queryGroup)
		w = sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: "TestSave",
			Type:  model.QueryTypeSavedQuery,
			Query: &postgres.Jsonb{queryJson}})
		assert.Equal(t, http.StatusCreated, w.Code)

		var queries model.QueriesString
		decoder = json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&queries); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotEqual(t, "", queries.IdText)

		w = sendEventsQueryHandlerWithQueryId(r, project.ID, agent, queries.IdText)
		assert.Equal(t, http.StatusOK, w.Code)
		decoder = json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&resultGroup); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.NotNil(t, 1, len(resultGroup.Results))
		result1 = resultGroup.Results[0]
		assert.NotNil(t, result1)
		assert.Equal(t, "aggregate", result1.Headers[0])
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
		assert.Equal(t, "aggregate", result1.Headers[0])
		assert.Equal(t, float64(5), result1.Rows[0][0])

		result2 := resultGroup.Results[1]
		assert.NotNil(t, result2)
		assert.Equal(t, "aggregate", result2.Headers[0])
		assert.Equal(t, float64(15), result2.Rows[0][0])

	})
}

func sendEventsQueryHandler(r *gin.Engine, projectId int64, agent *model.Agent, queryGroup *model.QueryGroup) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/query", projectId)).
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

func sendEventsQueryHandlerWithQueryId(r *gin.Engine, projectId int64, agent *model.Agent, queryId string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/query?query_id=%v", projectId, queryId)).
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
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		if result.Rows[0][1] == "s1" {
			assert.Equal(t, float64(3), result.Rows[0][2])
		} else {
			assert.Equal(t, float64(2), result.Rows[0][2])
		}
		if result.Rows[1][1] == "s1" {
			assert.Equal(t, float64(3), result.Rows[1][2])
		} else {
			assert.Equal(t, float64(2), result.Rows[1][2])
		}
	})

	t.Run("AnalyticsEachEventsQueryEventOccurrenceBreakdownByDateEvents", func(t *testing.T) {
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

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "datetime", result.Headers[0])
		if result.Headers[1] == "s0" {
			assert.Equal(t, float64(2), result.Rows[0][1])
			assert.Equal(t, float64(3), result.Rows[0][2])
		} else {
			assert.Equal(t, float64(3), result.Rows[0][1])
			assert.Equal(t, float64(2), result.Rows[0][2])
		}
	})

	t.Run("TestingEventsWithZeroValuesAndEventsWithValues", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name:       "s0",
					Properties: []model.QueryProperty{},
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
			},
			Class: model.QueryClassEvents,

			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with fliter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "s0", result.Rows[0][1])
		assert.Equal(t, "s1", result.Rows[1][1])
		assert.Equal(t, "s3", result.Rows[2][1])
		assert.Equal(t, "s2", result.Rows[3][1])
	})

	t.Run("TestAnalyticsQueryWithContainshNone", func(t *testing.T) {
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
							Operator:  "contains",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "$none",
						},
					},
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}
		result, code, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, len(result.Rows), 0)
	})

	t.Run("TestAnalyticsQueryWithNotContainshNone", func(t *testing.T) {
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
							Operator:  "notContains",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "$none",
						},
					},
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}
		result, code, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, code)
		log.WithField("rows", result.Rows).Warn("kark2")
		assert.Equal(t, 1, len(result.Rows))
		assert.Equal(t, "s0", result.Rows[0][1])
		assert.Equal(t, float64(4), result.Rows[0][2])
	})

	t.Run("TestAnalyticsQueryWithNotContainsProperty", func(t *testing.T) {
		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 40,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityUser,
							Property:  "$intial_source",
							Operator:  "notEqual",
							Type:      "categorial",
							LogicalOp: "AND",
							Value:     "channel",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{Entity: "user", EventName: "$present", Property: "$hubspot_contact_hs_analytics_source_data_2", Type: "categorical"},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
		}
		result, code, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, code)
		log.WithField("rows", result.Rows).Warn("kark2")
		assert.Equal(t, 1, len(result.Rows))
		assert.Equal(t, "s0", result.Rows[0][1])
		assert.Equal(t, "$none", result.Rows[0][2])
		assert.Equal(t, float64(4), result.Rows[0][3])
	})
}

func TestCoalUniqueUsersEachEventQuery(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"
	customerIDUser3 := "customerIDUser3"

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	/*
		user1 -> event s0 with property1 -> s0 with property2 -> s1 with propterty2
		user2 -> event s0 with property1 -> s1 with property1
		user3 -> event s0 with property2 -> s1 with property2
	*/

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestCoalUniqueUsersEachEventQueryUser", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		if result.Rows[0][1] == "s1" {
			assert.Equal(t, float64(3), result.Rows[0][2])
		} else {
			assert.Equal(t, float64(3), result.Rows[0][2])
		}
		if result.Rows[1][1] == "s0" {
			assert.Equal(t, float64(3), result.Rows[1][2])
		} else {
			assert.Equal(t, float64(3), result.Rows[1][2])
		}
	})

	t.Run("TestCoalUniqueUsersEachEventQueryEvents", func(t *testing.T) {
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
			Class:            model.QueryClassEvents,
			GroupByTimestamp: "date",
			Type:             model.QueryTypeEventsOccurrence,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "datetime", result.Headers[0])
		if result.Headers[1] == "s0" {
			assert.Equal(t, float64(7), result.Rows[0][1])
			assert.Equal(t, float64(6), result.Rows[0][2])
		} else {
			assert.Equal(t, float64(6), result.Rows[0][1])
			assert.Equal(t, float64(7), result.Rows[0][2])
		}
	})
}

func TestCoalUniqueUsersEachEventQuerySingleUser(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestCoalUniqueUsersEachEventQuerySingleUserUser", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
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
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		if result.Rows[0][1] == "s1" {
			assert.Equal(t, float64(1), result.Rows[0][2])
		} else {
			assert.Equal(t, float64(1), result.Rows[0][2])
		}
		if result.Rows[1][1] == "s0" {
			assert.Equal(t, float64(1), result.Rows[1][2])
		} else {
			assert.Equal(t, float64(1), result.Rows[1][2])
		}
	})

	t.Run("TestCoalUniqueUsersEachEventQuerySingleUserEvents", func(t *testing.T) {
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
			Class:            model.QueryClassEvents,
			GroupByTimestamp: "date",
			Type:             model.QueryTypeEventsOccurrence,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "datetime", result.Headers[0])
		if result.Headers[1] == "s0" {
			assert.Equal(t, float64(7), result.Rows[0][1])
			assert.Equal(t, float64(6), result.Rows[0][2])
		} else {
			assert.Equal(t, float64(6), result.Rows[0][1])
			assert.Equal(t, float64(7), result.Rows[0][2])
		}
	})
}

func TestCoalUniqueUsersEachEventQuerySingleUserMultiEvents(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s5", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s6", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s7", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s8", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestCoalUniqueUsersEachEventQuerySingleUserMultiEventsUserEach", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
				model.QueryEventWithProperties{
					Name: "s5",
				},
				model.QueryEventWithProperties{
					Name: "s6",
				},
				model.QueryEventWithProperties{
					Name: "s7",
				},
				model.QueryEventWithProperties{
					Name: "s8",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			assert.Equal(t, float64(1), result.Rows[key][2])
		}
	})

	t.Run("TestCoalUniqueUsersEachEventQuerySingleUserMultiEventsUserAny", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
				model.QueryEventWithProperties{
					Name: "s5",
				},
				model.QueryEventWithProperties{
					Name: "s6",
				},
				model.QueryEventWithProperties{
					Name: "s7",
				},
				model.QueryEventWithProperties{
					Name: "s8",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(1), result.Rows[0][0])
	})

	t.Run("TestCoalUniqueUsersEachEventQuerySingleUserMultiEventsUserAll", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
				model.QueryEventWithProperties{
					Name: "s5",
				},
				model.QueryEventWithProperties{
					Name: "s6",
				},
				model.QueryEventWithProperties{
					Name: "s7",
				},
				model.QueryEventWithProperties{
					Name: "s8",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(1), result.Rows[0][0])
	})
}

func TestCoalUniqueUsersEachEventQueryMultiUserMultiEvents(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"
	customerIDUser3 := "customerIDUser3"
	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestCoalUniqueUsersEachEventQueryMultiUserMultiEventsUserEach", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			if result.Rows[key][1] == "s0" {
				assert.Equal(t, float64(1), result.Rows[key][2])
			} else if result.Rows[key][1] == "s1" {
				assert.Equal(t, float64(3), result.Rows[key][2])
			} else if result.Rows[key][1] == "s2" {
				assert.Equal(t, float64(2), result.Rows[key][2])
			} else if result.Rows[key][1] == "s3" {
				assert.Equal(t, float64(2), result.Rows[key][2])
			}
		}
	})
	t.Run("TestCoalUniqueUsersEachEventQueryMultiUserMultiEventsUserAny", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(3), result.Rows[0][0])
	})
	t.Run("TestCoalUniqueUsersEachEventQueryMultiUserMultiEventsUserAll", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(0), result.Rows[0][0])
	})
}

func TestCoalUniqueUsersEachEventQueryMultiUserMultiEventsWeekTimeGroup(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"
	customerIDUser3 := "customerIDUser3"

	startTimestamp := int64(1622636726)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestEventCountGroupByWeekTimeGroup", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + (30 * 24 * 60 * 60),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GroupByTimestamp: model.GroupByTimestampWeek,
			Class:            model.QueryClassEvents,
			Type:             model.QueryTypeEventsOccurrence,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "datetime", result.Headers[0])
		baseSunday := int64(1622332800)
		assert.Equal(t, baseSunday, result.Rows[0][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(7*24*60*60), result.Rows[1][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(14*24*60*60), result.Rows[2][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(21*24*60*60), result.Rows[3][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(28*24*60*60), result.Rows[4][0].(time.Time).Unix())

		// since meta metrics has the info on row column mapping for count, let's just check if first row is non-zero
		for i, _ := range result.Rows[0] {
			if i != 0 {
				assert.NotEqual(t, float64(0), result.Rows[0][i])
			}
		}
	})
	t.Run("TestCoalUniqueUsersCountGroupByWeekTimeGroup", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + (30 * 24 * 60 * 60),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GroupByTimestamp: model.GroupByTimestampWeek,
			Class:            model.QueryClassEvents,
			Type:             model.QueryTypeUniqueUsers,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "datetime", result.Headers[0])
		baseSunday := int64(1622332800)
		assert.Equal(t, baseSunday, result.Rows[0][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(7*24*60*60), result.Rows[1][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(14*24*60*60), result.Rows[2][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(21*24*60*60), result.Rows[3][0].(time.Time).Unix())
		assert.Equal(t, baseSunday+(28*24*60*60), result.Rows[4][0].(time.Time).Unix())

		// since meta metrics has the info on row column mapping for count, let's just check if first row is non-zero
		for i, _ := range result.Rows {
			if i != 0 {
				assert.NotEqual(t, float64(0), result.Rows[i][0])
			}
		}
	})
}

func TestCoalUniqueUsersEachEventQueryMultiUserMultiEventsQuarterTimeGroup(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"
	customerIDUser3 := "customerIDUser3"

	startTimestamp := int64(1622636726)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestEventCountGroupByQuarterTimeGroup", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + (200 * 24 * 60 * 60),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GroupByTimestamp: model.GroupByTimestampQuarter,
			Class:            model.QueryClassEvents,
			Type:             model.QueryTypeEventsOccurrence,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "datetime", result.Headers[0])
		april1 := int64(1617235200)
		july1 := int64(1625097600)
		oct1 := int64(1633046400)
		assert.Equal(t, april1, result.Rows[0][0].(time.Time).Unix())
		assert.Equal(t, july1, result.Rows[1][0].(time.Time).Unix())
		assert.Equal(t, oct1, result.Rows[2][0].(time.Time).Unix())

		// since meta metrics has the info on row column mapping for count, let's just check if first row is non-zero
		for i, _ := range result.Rows[0] {
			if i != 0 {
				assert.NotEqual(t, float64(0), result.Rows[0][i])
			}
		}
	})
	t.Run("TestCoalUniqueUsersCountGroupByQuarterTimeGroup", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + (200 * 24 * 60 * 60),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GroupByTimestamp: model.GroupByTimestampQuarter,
			Class:            model.QueryClassEvents,
			Type:             model.QueryTypeUniqueUsers,
			EventsCondition:  model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "datetime", result.Headers[0])
		april1 := int64(1617235200)
		july1 := int64(1625097600)
		oct1 := int64(1633046400)
		assert.Equal(t, april1, result.Rows[0][0].(time.Time).Unix())
		assert.Equal(t, july1, result.Rows[1][0].(time.Time).Unix())
		assert.Equal(t, oct1, result.Rows[2][0].(time.Time).Unix())

		// since meta metrics has the info on row column mapping for count, let's just check if first row is non-zero
		for i, _ := range result.Rows {
			if i != 0 {
				assert.NotEqual(t, float64(0), result.Rows[i][0])
			}
		}
	})
}

func TestGlobalFilterChanges(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"
	customerIDUser3 := "customerIDUser3"

	commonUserProperty := make(map[string]interface{})
	commonUserProperty[U.UP_BROWSER] = "Chrome"
	commonUserPropertyBytes, _ := json.Marshal(commonUserProperty)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
		CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
		CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestEachUniqueUsersGlobalFilter", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GlobalUserProperties: []model.QueryProperty{
				model.QueryProperty{
					Entity:    model.PropertyEntityUserGlobal,
					Property:  U.UP_BROWSER,
					Operator:  model.EqualsOpStr,
					Type:      U.PropertyTypeCategorical,
					LogicalOp: model.LOGICAL_OP_AND,
					Value:     "Chrome",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			if result.Rows[key][1] == "s0" {
				assert.Equal(t, float64(1), result.Rows[key][2])
			} else if result.Rows[key][1] == "s1" {
				assert.Equal(t, float64(2), result.Rows[key][2])
			} else if result.Rows[key][1] == "s2" {
				assert.Equal(t, float64(2), result.Rows[key][2])
			} else if result.Rows[key][1] == "s3" {
				assert.Equal(t, float64(1), result.Rows[key][2])
			} else if result.Rows[key][1] == "s4" {
				assert.Equal(t, float64(1), result.Rows[key][2])
			}
		}
	})
	t.Run("TestAnyUniqueUsersGlobalFilter", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GlobalUserProperties: []model.QueryProperty{
				model.QueryProperty{
					Entity:    model.PropertyEntityUserGlobal,
					Property:  U.UP_BROWSER,
					Operator:  model.EqualsOpStr,
					Type:      U.PropertyTypeCategorical,
					LogicalOp: model.LOGICAL_OP_AND,
					Value:     "Chrome",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAnyGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(2), result.Rows[0][0])
	})
	t.Run("TestCoalUniqueUsersEachEventQueryMultiUserMultiEventsUserAll", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			GlobalUserProperties: []model.QueryProperty{
				model.QueryProperty{
					Entity:    model.PropertyEntityUserGlobal,
					Property:  U.UP_BROWSER,
					Operator:  model.EqualsOpStr,
					Type:      U.PropertyTypeCategorical,
					LogicalOp: model.LOGICAL_OP_AND,
					Value:     "Chrome",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondAllGivenEvent,
		}
		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "aggregate", result.Headers[0])
		assert.Equal(t, float64(0), result.Rows[0][0])
	})
}

func TestNoneFilterGroup1(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"

	commonUserProperty := make(map[string]interface{})
	commonUserProperty[U.UP_BROWSER] = "Chrome"
	commonUserPropertyBytes, _ := json.Marshal(commonUserProperty)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$source":"%s"}}`,
		"s0", createdUserID1, stepTimestamp, "A", "google")
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$source":"%s"}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", "google")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$source":"%s"}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", "google")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestEachEventsNoneFilterCase1", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "$none",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{
					Property:       U.EP_SOURCE,
					Entity:         "event",
					Type:           "categorical",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},

			GlobalUserProperties: []model.QueryProperty{},
			Class:                model.QueryClassEvents,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsCondition:      model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			if result.Rows[key][1] == "s0" {
				assert.Equal(t, "google", result.Rows[key][2])
				assert.Equal(t, float64(1), result.Rows[key][3])
			}
		}
	})
	t.Run("TestEachEventsNoneFilterCase2", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "google",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{
					Property:       U.EP_SOURCE,
					Entity:         "event",
					Type:           "categorical",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},

			GlobalUserProperties: []model.QueryProperty{},
			Class:                model.QueryClassEvents,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsCondition:      model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		// no events where $source != "google"
		assert.Equal(t, 0, len(result.Rows))
	})
	t.Run("TestEachEventsNoneFilterCase3", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "$none",
						},
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "google",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{
					Property:       U.EP_SOURCE,
					Entity:         "event",
					Type:           "categorical",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},

			GlobalUserProperties: []model.QueryProperty{},
			Class:                model.QueryClassEvents,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsCondition:      model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		// no events where $source != "google"
		assert.Equal(t, 0, len(result.Rows))
	})
}

func TestNoneFilterGroup2(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"

	commonUserProperty := make(map[string]interface{})
	commonUserProperty[U.UP_BROWSER] = "Chrome"
	commonUserPropertyBytes, _ := json.Marshal(commonUserProperty)

	startTimestamp := U.UnixTimeBeforeDuration(time.Hour * 1)
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$source":"%s"}}`,
		"s0", createdUserID1, stepTimestamp, "A", "google")
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$source":"%s"}}`,
		"s0", createdUserID1, stepTimestamp+10, "A", "google")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$source":"%s"}}`,
		"s0", createdUserID1, stepTimestamp+20, "A", "facebook")
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestEachEventsNoneFilterCase1", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "$none",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{
					Property:       U.EP_SOURCE,
					Entity:         "event",
					Type:           "categorical",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},

			GlobalUserProperties: []model.QueryProperty{},
			Class:                model.QueryClassEvents,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsCondition:      model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			if result.Rows[key][1] == "s0" {
				if "google" == result.Rows[key][2] {
					assert.Equal(t, float64(2), result.Rows[key][3])
				}
				if "facebook" == result.Rows[key][2] {
					assert.Equal(t, float64(1), result.Rows[key][3])
				}
			}
		}
	})
	t.Run("TestEachEventsNoneFilterCase2", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "google",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{
					Property:       U.EP_SOURCE,
					Entity:         "event",
					Type:           "categorical",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},

			GlobalUserProperties: []model.QueryProperty{},
			Class:                model.QueryClassEvents,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsCondition:      model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			if result.Rows[key][1] == "s0" {
				assert.Equal(t, "facebook", result.Rows[key][2])
				assert.Equal(t, float64(1), result.Rows[key][3])
			}
		}
	})
	t.Run("TestEachEventsNoneFilterCase3", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + 200,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
					Properties: []model.QueryProperty{
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "$none",
						},
						model.QueryProperty{
							Entity:    model.PropertyEntityEvent,
							Property:  U.EP_SOURCE,
							Operator:  model.NotEqualOpStr,
							Type:      U.PropertyTypeCategorical,
							LogicalOp: model.LOGICAL_OP_AND,
							Value:     "google",
						},
					},
				},
			},
			GroupByProperties: []model.QueryGroupByProperty{
				{
					Property:       U.EP_SOURCE,
					Entity:         "event",
					Type:           "categorical",
					EventName:      "s0",
					EventNameIndex: 1,
				},
			},

			GlobalUserProperties: []model.QueryProperty{},
			Class:                model.QueryClassEvents,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsCondition:      model.EventCondEachGivenEvent,
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for key := range result.Rows {
			if result.Rows[key][1] == "s0" {
				assert.Equal(t, "facebook", result.Rows[key][2])
				assert.Equal(t, float64(1), result.Rows[key][3])
			}
		}
	})
}

func TestGroupByDateTimePropWeekTimeGroup(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	customerIDUser1 := "customerIDUser1"
	customerIDUser2 := "customerIDUser2"
	customerIDUser3 := "customerIDUser3"
	commonUserProperty := make(map[string]interface{})
	commonUserProperty[U.UP_BROWSER] = "Chrome"
	commonUserProperty["$custom_time"] = "1622636726" // Wednesday, 2 June 2021 17:55:26 GMT+05:30
	commonUserPropertyBytes, _ := json.Marshal(commonUserProperty)

	startTimestamp := int64(1622636726) // Wednesday, 2 June 2021 17:55:26 GMT+05:30
	stepTimestamp := startTimestamp

	// here createdUserID1, user4_1 have same customerID, same for 2, 5_2 and 3, 6_3
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID1)
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID2)
	createdUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID3)
	createdUserID4_1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser2, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID4_1)
	createdUserID5_2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser1, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID5_2)
	createdUserID6_3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: customerIDUser3, JoinTimestamp: startTimestamp - 10,
		Properties: postgres.Jsonb{RawMessage: commonUserPropertyBytes}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotEmpty(t, createdUserID6_3)

	payload := fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s0", createdUserID1, stepTimestamp, "A", 1234)
	w := ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID1, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID2, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s1", createdUserID3, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s2", createdUserID4_1, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s3", createdUserID4_1, stepTimestamp+10, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d,
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp+20, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID5_2, stepTimestamp, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp+10, "A", 1234)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	payload = fmt.Sprintf(`{"event_name": "%s", "user_id": "%s","timestamp": %d, 
	"user_properties": {"$initial_source" : "%s"}, "event_properties":{"$campaign_id":%d}}`,
		"s4", createdUserID6_3, stepTimestamp, "A", 4321)
	w = ServePostRequestWithHeaders(r, uri, []byte(payload), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	response = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, response["event_id"])

	t.Run("TestEventCountGroupByWeekTimeGroup", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + (30 * 24 * 60 * 60),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeEventsOccurrence,
			EventsCondition: model.EventCondEachGivenEvent,
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:      model.PropertyEntityUser,
					Property:    "$custom_time",
					Type:        U.PropertyTypeDateTime,
					EventName:   "present",
					Granularity: "week",
				},
			},
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		for i := 0; i < len(result.Headers); i++ {
			if "$custom_time" == result.Headers[i] {
				// Grouping starts from sunday
				assert.Equal(t, "2021-05-30T00:00:00+00:00", result.Rows[0][i].(string))
			}
		}
	})
	t.Run("TestCoalUniqueUsersCountGroupByWeekTimeGroup", func(t *testing.T) {

		query := model.Query{
			From: startTimestamp,
			To:   startTimestamp + (30 * 24 * 60 * 60),
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "s0",
				},
				model.QueryEventWithProperties{
					Name: "s1",
				},
				model.QueryEventWithProperties{
					Name: "s2",
				},
				model.QueryEventWithProperties{
					Name: "s3",
				},
				model.QueryEventWithProperties{
					Name: "s4",
				},
			},
			Class:           model.QueryClassEvents,
			Type:            model.QueryTypeUniqueUsers,
			EventsCondition: model.EventCondEachGivenEvent,
			GroupByProperties: []model.QueryGroupByProperty{
				model.QueryGroupByProperty{
					Entity:      model.PropertyEntityUser,
					Property:    "$custom_time",
					Type:        U.PropertyTypeDateTime,
					EventName:   "present",
					Granularity: "week",
				},
			},
		}

		//unique user count should return 2 for s0 to s1 with filter property1
		result, errCode, _ := store.GetStore().ExecuteEventsQuery(project.ID, query, C.EnableOptimisedFilterOnEventUserQuery())
		assert.Equal(t, http.StatusOK, errCode)
		assert.NotNil(t, result)
		assert.Equal(t, "event_index", result.Headers[0])
		assert.Equal(t, "$custom_time", result.Headers[2])
		// Grouping starts from sunday
		assert.Equal(t, "2021-05-30T00:00:00+00:00", result.Rows[0][2].(string))
	})
}
