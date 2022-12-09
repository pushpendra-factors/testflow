package tests

import (
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendSegmentPostReq(r *gin.Engine, request model.SegmentPayload, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/segments", projectId)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create segment req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendAllSegmentGetReq(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/segments", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating get all segments req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendSegmentGetByIdReq(r *gin.Engine, projectId int64, id string, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/segments/%s", projectId, id)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating get segment by id req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendUpdateSegmentPutReq(r *gin.Engine, request model.SegmentPayload, projectId int64, id string, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/segments/%s", projectId, id)).
		WithPostParams(request).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating put segment req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPostAPISegmentHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// Create new segment.
	events := make([]model.EventWithProperty, 1)
	properties := make([]model.QueryProperty, 1)
	prop := model.QueryProperty{
		Value:     "1",
		Property:  "prop1",
		Operator:  "op1",
		LogicalOp: "logicalop1",
		Type:      "type1",
	}
	properties[0] = prop
	event := model.EventWithProperty{
		Names:           "eventName1",
		Properties:      properties,
		LogicalOperator: "logicalOp1",
	}
	events[0] = event
	querySegment := model.SegmentQuery{
		EventsWithProperties: events,
		GlobalProperties:     properties,
	}

	segment := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w := sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// To check whether segemnent created
	getSegement, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, segment.Name, getSegement["event"][0].Name)
	assert.Equal(t, segment.Description, getSegement["event"][0].Description)
	assert.Equal(t, segment.Type, getSegement["event"][0].Type)
	querySegmentCheck, err := U.EncodeStructTypeToPostgresJsonb(getSegement["event"][0].Query)
	assert.Nil(t, err)
	getQueryMap, err := U.DecodePostgresJsonb(querySegmentCheck)
	assert.Nil(t, err)
	queryMap, _ := U.EncodeStructTypeToMap(querySegment)
	assert.Equal(t, &queryMap, getQueryMap)

	// Creating record with same name
	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Creating record without field query
	segment = &model.SegmentPayload{
		Name:        "Name2",
		Description: "dummy info",
		Type:        "event",
		Query:       model.SegmentQuery{},
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Creating record without field type
	segment = &model.SegmentPayload{
		Name:        "Name2",
		Description: "dummy info",
		Query:       querySegment,
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Creating record without field description
	segment = &model.SegmentPayload{
		Name:  "Name2",
		Query: querySegment,
		Type:  "event",
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Creating record without field name
	segment = &model.SegmentPayload{
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// creating record with only EventsWithProperties in Query
	querySegment = model.SegmentQuery{
		EventsWithProperties: events,
	}
	segment = &model.SegmentPayload{
		Name:        "Name3",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// creating record with only GlobalProperties in Query
	querySegment = model.SegmentQuery{
		GlobalProperties: properties,
	}
	segment = &model.SegmentPayload{
		Name:        "Name4",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

}

func TestGetAPIAllSegmentsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// Create new segment.
	events := make([]model.EventWithProperty, 1)
	properties := make([]model.QueryProperty, 1)
	prop := model.QueryProperty{
		Value:     "1",
		Property:  "prop1",
		Operator:  "op1",
		LogicalOp: "logicalop1",
		Type:      "type1",
	}
	properties[0] = prop
	event := model.EventWithProperty{
		Names:           "eventName1",
		Properties:      properties,
		LogicalOperator: "logicalOp1",
	}
	events[0] = event
	querySegment := model.SegmentQuery{
		EventsWithProperties: events,
		GlobalProperties:     properties,
	}

	segment := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w := sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	// creating one more record
	prop = model.QueryProperty{
		Value:     "1",
		Property:  "prop1",
		Operator:  "op1",
		LogicalOp: "logicalop1",
		Type:      "type1",
	}
	properties[0] = prop
	event = model.EventWithProperty{
		Names:           "eventName1",
		Properties:      properties,
		LogicalOperator: "logicalOp1",
	}
	events[0] = event
	querySegment = model.SegmentQuery{
		EventsWithProperties: events,
		GlobalProperties:     properties,
	}

	segment = &model.SegmentPayload{
		Name:        "Name2",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event1",
	}

	w = sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)
	// Get all segments
	w = sendAllSegmentGetReq(r, project.ID, agent)
	assert.Equal(t, http.StatusFound, w.Code)

	getSegement, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, int(2), len(getSegement))
}

func TestGetAPISegmentByIdHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// Create new segment.
	events := make([]model.EventWithProperty, 1)
	properties := make([]model.QueryProperty, 1)
	prop := model.QueryProperty{
		Value:     "1",
		Property:  "prop1",
		Operator:  "op1",
		LogicalOp: "logicalop1",
		Type:      "type1",
	}
	properties[0] = prop
	event := model.EventWithProperty{
		Names:           "eventName1",
		Properties:      properties,
		LogicalOperator: "logicalOp1",
	}
	events[0] = event
	querySegment := model.SegmentQuery{
		EventsWithProperties: events,
		GlobalProperties:     properties,
	}

	segment := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w := sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	getSegement, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	// Get segment by id
	w = sendSegmentGetByIdReq(r, project.ID, getSegement["event"][0].Id, agent)
	assert.Equal(t, http.StatusFound, w.Code)

	// To check the segemnent recieved from api
	getSegement1, status := store.GetStore().GetSegmentById(project.ID, getSegement["event"][0].Id)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, segment.Name, getSegement1.Name)
	assert.Equal(t, segment.Description, getSegement1.Description)
	assert.Equal(t, segment.Type, getSegement1.Type)
	querySegmentCheck, err := U.EncodeStructTypeToPostgresJsonb(getSegement1.Query)
	assert.Nil(t, err)
	getQueryMap, err := U.DecodePostgresJsonb(querySegmentCheck)
	assert.Nil(t, err)
	queryMap, _ := U.EncodeStructTypeToMap(querySegment)
	assert.Equal(t, &queryMap, getQueryMap)
}

func TestPutAPISegmentHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	// Create new segment.
	events := make([]model.EventWithProperty, 1)
	properties := make([]model.QueryProperty, 1)
	prop := model.QueryProperty{
		Value:     "1",
		Property:  "prop1",
		Operator:  "op1",
		LogicalOp: "logicalop1",
		Type:      "type1",
	}
	properties[0] = prop
	event := model.EventWithProperty{
		Names:           "eventName1",
		Properties:      properties,
		LogicalOperator: "logicalOp1",
	}
	events[0] = event
	querySegment := model.SegmentQuery{
		EventsWithProperties: events,
		GlobalProperties:     properties,
	}

	segment := &model.SegmentPayload{
		Name:        "Name1",
		Description: "dummy info",
		Query:       querySegment,
		Type:        "event",
	}

	w := sendSegmentPostReq(r, *segment, project.ID, agent)
	assert.Equal(t, http.StatusCreated, w.Code)

	getSegement, status := store.GetStore().GetAllSegments(project.ID)
	assert.Equal(t, http.StatusFound, status)

	// Update segment by id
	segment = &model.SegmentPayload{
		Description: "dummy dummy info",
		Type:        "new event",
	}

	// Updating with only type test
	w = sendUpdateSegmentPutReq(r, *segment, project.ID, getSegement["event"][0].Id, agent)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// updating with both type and query
	prop = model.QueryProperty{
		Value:     "1111111",
		Property:  "new prop",
		Operator:  "new op1",
		LogicalOp: "new logicalop1",
		Type:      "new type1",
	}
	properties[0] = prop
	querySegment = model.SegmentQuery{
		GlobalProperties:     properties,
		EventsWithProperties: events,
	}

	newSegment := &model.SegmentPayload{
		Name:        "Name2",
		Description: "Dummy description ",
		Type:        "event2",
		Query:       querySegment,
	}

	w = sendUpdateSegmentPutReq(r, *newSegment, project.ID, getSegement["event"][0].Id, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	// To check whether segemnent created
	getSegement1, status := store.GetStore().GetSegmentById(project.ID, getSegement["event"][0].Id)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, newSegment.Name, getSegement1.Name)
	assert.Equal(t, newSegment.Description, getSegement1.Description)
	assert.Equal(t, newSegment.Type, getSegement1.Type)
	querySegmentCheck, err := U.EncodeStructTypeToPostgresJsonb(getSegement1.Query)
	assert.Nil(t, err)
	getQueryMap, err := U.DecodePostgresJsonb(querySegmentCheck)
	assert.Nil(t, err)
	queryMap, _ := U.EncodeStructTypeToMap(querySegment)
	assert.Equal(t, &queryMap, getQueryMap)
}
