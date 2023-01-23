package tests

import (
	C "factors/config"
	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	"factors/task/event_user_cache"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"

	V1 "factors/handler/v1"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createFactorsRule(project_id int64, agent_UUID string) ([]model.FactorsGoalRule, error) {

	fr := make([]model.FactorsGoalRule, 0)
	events_name := []string{"start_1", "start_2", "start_3", "end_1", "end_2", "end_3"}
	for _, ename := range events_name {
		_, _ = store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{Name: ename, ProjectId: project_id})
		_, _ = store.GetStore().CreateFactorsTrackedEvent(project_id, ename, agent_UUID)

	}
	var fr1 model.FactorsGoalRule
	var fr2 model.FactorsGoalRule
	var fr3 model.FactorsGoalRule

	fr1.StartEvent = events_name[0]
	fr1.EndEvent = events_name[3]

	fr2.StartEvent = events_name[1]
	fr2.EndEvent = events_name[4]

	fr3.StartEvent = events_name[2]
	fr3.EndEvent = events_name[5]
	fr = append(fr, fr1)

	return fr, nil
}

func createQueryV1(project_id int64, agent_UUID string) ([]model.ExplainV2Query, error) {

	qrys := make([]model.ExplainV2Query, 0)
	rulesFilter, err := createFactorsRule(project_id, agent_UUID)
	if err != nil {
		return nil, err
	}
	for _, ru := range rulesFilter {
		var qr model.ExplainV2Query
		qr.Query = ru
		qr.Title = U.RandomString(5)
		qr.StartTimestamp = 1670371748
		qr.EndTimestamp = 1670371750
		qrys = append(qrys, qr)
	}
	return qrys, nil
}

func createProjectAgentEventsFactorsTrackedEventsForExplainV2(r *gin.Engine) (int64, *model.Agent) {

	C.GetConfig().LookbackWindowForEventUserCache = 1

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, agent, _ := SetupProjectWithAgentDAO()

	createdUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})

	rEventName := "event1"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event2"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up1": "uv1"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event3"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "end1"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "end2"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up1": "uv1"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "end3"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "hubspot_account_created"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	rEventName = "hubspot_opportunity_created"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	rEventName = "hubspot_contact_created"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up2": "uv2"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})
	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	return project.ID, agent
}

func TestExplainV2(t *testing.T) {

	type errorObject struct {
		Error string `json:"error"`
	}
	type successObject struct {
		Id     int64  `json:"id"`
		Status string `json:"status"`
	}
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, agent := createProjectAgentEventsFactorsTrackedEventsForExplainV2(r)
	C.GetConfig().ActiveFactorsTrackedEventsLimit = 50
	project_ID := projectId
	agent_UUID := agent.UUID

	events_name := []string{"start_1", "start_2", "start_3", "end_1", "end_2", "end_3"}

	for _, en := range events_name {
		request := V1.CreateFactorsTrackedEventParams{}
		request.EventName = en
		_ = sendCreateFactorsTrackedEvent(r, request, agent, projectId)
		// assert.Equal(t, http.StatusCreated, w.Code)
	}

	t.Run("CreateExplainV2Entity:valid", func(t *testing.T) {
		queries, err := createQueryV1(project_ID, agent_UUID)
		assert.Nil(t, err)
		query := queries[0]
		entity, errCode, errMsg := store.GetStore().CreateExplainV2Entity(agent_UUID, project_ID, &query)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)
	})

	t.Run("GetExplainV2Entity:valid", func(t *testing.T) {
		entity, errCode := store.GetStore().GetAllExplainV2EntityByProject(project_ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, entity)
	})

	t.Run("CreateExplainV2Entity:Title already present:Invalid", func(t *testing.T) {

		queries, err := createQueryV1(project_ID, agent_UUID)
		assert.Nil(t, err)
		query1 := queries[0]
		query2 := queries[0]
		entity, errCode, errMsg := store.GetStore().CreateExplainV2Entity(agent_UUID, project_ID, &query1)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		entity, errCode, _ = store.GetStore().CreateExplainV2Entity(agent_UUID, project_ID, &query2)
		assert.Equal(t, http.StatusConflict, errCode)
		assert.Nil(t, entity)
	})

	t.Run("DeleteExplainV2Entity:valid", func(t *testing.T) {

		queries, err := createQueryV1(project_ID, agent_UUID)
		assert.Nil(t, err)
		query1 := queries[0]
		entity, errCode, errMsg := store.GetStore().CreateExplainV2Entity(agent.UUID, project_ID, &query1)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.Empty(t, errMsg)
		assert.NotNil(t, entity)

		errCode, errMsg = store.GetStore().DeleteExplainV2Entity(project_ID, entity.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Empty(t, errMsg)
	})

}
