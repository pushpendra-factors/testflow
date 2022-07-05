package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	V1 "factors/handler/v1"
	"factors/model/model"
	"factors/model/store"
	"factors/task/event_user_cache"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateFactorsGoalRequest(r *gin.Engine, request V1.CreateFactorsGoalParams, agent *model.Agent, projectID int64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/factors/goals", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating goal")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetAllFactorsGoalsRequest(r *gin.Engine, agent *model.Agent, projectID int64) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/factors/goals", projectID)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting goals data.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendSearchFactorsGoalsRequest(r *gin.Engine, agent *model.Agent, projectID int64, request V1.SearchFactorsGoalParams) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/v1/factors/goals/search", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error searching goals data.")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendRemoveFactorsGoalRequest(r *gin.Engine, agent *model.Agent, projectID int64, request V1.RemoveFactorsGoalParams) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/v1/factors/goals/remove", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error removing goal")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendUpdateFactorsGoalRequest(r *gin.Engine, agent *model.Agent, projectID int64, request V1.UpdateFactorsGoalParams) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/v1/factors/goals/update", projectID)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error updating goal")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var id1, id2, id3 int64

func createProjectAgentEvents(r *gin.Engine) (int64, *model.Agent) {

	C.GetConfig().LookbackWindowForEventUserCache = 1

	H.InitSDKServiceRoutes(r)
	uri := "/sdk/event/track"

	project, agent, _ := SetupProjectWithAgentDAO()

	createdUserID, _ := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})

	rEventName := "event1"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"ep1": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon"}}`, createdUserID, rEventName)),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	rEventName = "event2"
	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"ep2": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"name": "Jhon", "up1": "uv1"}}`, createdUserID, rEventName)),
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
	configs := make(map[string]interface{})
	configs["rollupLookback"] = 1
	event_user_cache.DoRollUpSortedSet(configs)
	C.GetConfig().ActiveFactorsGoalsLimit = 50
	C.GetConfig().ActiveFactorsTrackedUserPropertiesLimit = 50
	C.GetConfig().ActiveFactorsTrackedEventsLimit = 50
	id1, _ = store.GetStore().CreateFactorsTrackedEvent(project.ID, "event1", agent.UUID)
	id2, _ = store.GetStore().CreateFactorsTrackedEvent(project.ID, "event2", agent.UUID)
	id3, _ = store.GetStore().CreateFactorsTrackedUserProperty(project.ID, "up1", agent.UUID)
	return project.ID, agent
}

func TestCreateFactorsGoalHandler(t *testing.T) {
	successFactorsGoalIds := make([]int64, 0)
	type errorObject struct {
		Error string `json:"error"`
	}
	type successObject struct {
		Id     int64  `json:"id"`
		Status string `json:"status"`
	}
	r := gin.Default()
	H.InitAppRoutes(r)
	projectId, agent := createProjectAgentEvents(r)

	request := V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal1"
	rule := model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	var globalFilters []model.KeyValueTuple
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)

	// Happy path
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal1"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w := sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusCreated, w.Code)
	var obj successObject
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	successFactorsGoalIds = append(successFactorsGoalIds, obj.Id)

	// Duplicate rule
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal2"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var err errorObject
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "Rule already exist", err.Error)

	// Duplicate name
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal1"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv2"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	fmt.Println(w.Body)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "duplicate name", err.Error)

	// non existing end event
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal4"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event_2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv3"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "event doesnt exist", err.Error)

	// non existing start event
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal5"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event_1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "event doesnt exist", err.Error)

	// start event not tracked
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal6"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event3"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "event not tracked", err.Error)

	// end event not tracked

	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal7"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event3"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "event not tracked", err.Error)

	// user property not exist
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal8"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up3", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "user property not associated to this project", err.Error)

	// user property not tracked
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal9"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up2", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "user property not tracked", err.Error)

	// get all goals
	w = sendGetAllFactorsGoalsRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	goals := []model.FactorsGoal{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &goals)
	assert.Equal(t, successFactorsGoalIds[0], int64(goals[0].ID))
	assert.Equal(t, true, goals[0].IsActive)

	// remove goal
	removeRequest := V1.RemoveFactorsGoalParams{
		ID: successFactorsGoalIds[0],
	}
	w = sendRemoveFactorsGoalRequest(r, agent, projectId, removeRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	obj = successObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	assert.Equal(t, successFactorsGoalIds[0], obj.Id)

	// get all goals
	w = sendGetAllFactorsGoalsRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	goals = []model.FactorsGoal{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &goals)
	assert.Equal(t, successFactorsGoalIds[0], int64(goals[0].ID))
	assert.Equal(t, false, goals[0].IsActive)

	// Happy path
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal10"
	rule = model.FactorsGoalRule{}
	rule.EndEvent = "event1"
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusCreated, w.Code)
	obj = successObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	successFactorsGoalIds = append(successFactorsGoalIds, obj.Id)

	// update goals
	updateRequest := V1.UpdateFactorsGoalParams{
		ID:   successFactorsGoalIds[1],
		Name: "Updated FactorsGoal",
	}
	w = sendUpdateFactorsGoalRequest(r, agent, projectId, updateRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	obj = successObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	assert.Equal(t, successFactorsGoalIds[1], obj.Id)

	goalObj, _ := store.GetStore().GetFactorsGoalByID(successFactorsGoalIds[1], projectId)
	assert.Equal(t, goalObj.Name, "Updated FactorsGoal")

	// Happy path
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal11"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event1"
	rule.EndEvent = "event2"
	rule.Visited = true
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	var stuserpr []model.KeyValueTuple
	stuserpr = append(stuserpr, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	var enuserpr []model.KeyValueTuple
	enuserpr = append(enuserpr, model.KeyValueTuple{Key: "up1", Value: "uv1"})
	var steventpr []model.KeyValueTuple
	steventpr = append(steventpr, model.KeyValueTuple{Key: "ep1", Value: "uv1"})
	var eneventpr []model.KeyValueTuple
	eneventpr = append(eneventpr, model.KeyValueTuple{Key: "ep2", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	rule.Rule.StartEnUserFitler = stuserpr
	rule.Rule.EndEnUserFitler = enuserpr
	rule.Rule.StartEnEventFitler = steventpr
	rule.Rule.EndEnEventFitler = eneventpr
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	fmt.Println(w)
	assert.Equal(t, http.StatusCreated, w.Code)
	obj = successObject{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &obj)
	successFactorsGoalIds = append(successFactorsGoalIds, obj.Id)

	store.GetStore().DeactivateFactorsTrackedEvent(id2, projectId)
	store.GetStore().RemoveFactorsTrackedUserProperty(id3, projectId)

	// get all goals
	w = sendGetAllFactorsGoalsRequest(r, agent, projectId)
	assert.Equal(t, http.StatusOK, w.Code)
	goals = []model.FactorsGoal{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &goals)
	assert.Equal(t, successFactorsGoalIds[0], int64(goals[0].ID))
	assert.Equal(t, false, goals[0].IsActive)
	assert.Equal(t, false, goals[1].IsActive)
	assert.Equal(t, false, goals[2].IsActive)

	// search goals
	searchRequest := V1.SearchFactorsGoalParams{
		SearchText: "FactorsGoal1",
	}
	w = sendSearchFactorsGoalsRequest(r, agent, projectId, searchRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	goals = []model.FactorsGoal{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &goals)
	assert.Equal(t, len(goals), 2)

	// search goals
	searchRequest = V1.SearchFactorsGoalParams{
		SearchText: "FactorsGoal",
	}
	w = sendSearchFactorsGoalsRequest(r, agent, projectId, searchRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	goals = []model.FactorsGoal{}
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &goals)
	assert.Equal(t, len(goals), 3)

	// nul agentID
	id, errCode, _ := store.GetStore().CreateFactorsGoal(projectId, "WithoutAgentID", model.FactorsGoalRule{EndEvent: "event1"}, "")
	assert.NotEqual(t, 0, id)
	assert.Equal(t, 201, errCode)

	// Limit exceeded
	C.GetConfig().ActiveFactorsGoalsLimit = 0
	request = V1.CreateFactorsGoalParams{}
	request.Name = "FactorsGoal12"
	rule = model.FactorsGoalRule{}
	rule.StartEvent = "event2"
	rule.Rule = model.FactorsGoalFilter{}
	globalFilters = nil
	globalFilters = append(globalFilters, model.KeyValueTuple{Key: "up2", Value: "uv1"})
	rule.Rule.GlobalFilters = globalFilters
	request.Rule = V1.ReverseMapRule(rule)
	w = sendCreateFactorsGoalRequest(r, request, agent, projectId)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &err)
	assert.Equal(t, "FactorsGoals count exceeded", err.Error)
}
