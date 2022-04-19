package tests

import (
	"encoding/json"
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
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateQueryReq(r *gin.Engine, projectId uint64, agent *model.Agent, query *H.SavedQueryRequestPayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/queries", projectId)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create query req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendGetSavedQueriesReq(r *gin.Engine, projectId uint64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/queries", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create query req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendUpdateSavedQueryReq(r *gin.Engine, projectId uint64, queryId uint64, agent *model.Agent, query *H.SavedQueryUpdatePayload) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/queries/%d", projectId, queryId)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating update query req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendDeleteSavedQueryReq(r *gin.Engine, projectId uint64, queryId uint64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/queries/%d", projectId, queryId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating delete query req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendSearchQueryReq(r *gin.Engine, projectId uint64, searchString string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/queries/search?query=%s", projectId, searchString)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating search query req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestAPISearchQueryHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rTitle := "xyz"
	query := model.Query{
		EventsCondition:      model.EventCondAnyGivenEvent,
		From:                 1556602834,
		To:                   1557207634,
		Type:                 model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{}, // invalid, no events.
		OverridePeriod:       true,
	}

	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle,
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{queryJson}})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendSearchQueryReq(r, project.ID, "x", agent)
	assert.Equal(t, http.StatusFound, w.Code)

	var queries []model.Queries
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&queries); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(queries))
	assert.NotEqual(t, "", queries[0].IdText)

	w = sendSearchQueryReq(r, project.ID, "a", agent)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPICreateQueryHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	t.Run("CreateQuery:WithValidQuery", func(t *testing.T) {
		rTitle := U.RandomString(5)
		query := model.Query{
			EventsCondition: model.EventCondAnyGivenEvent,
			From:            1556602834,
			To:              1557207634,
			Type:            model.QueryTypeEventsOccurrence,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "event1",
				},
			},
			OverridePeriod: true,
		}

		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle,
			Type:  model.QueryTypeSavedQuery,
			Query: &postgres.Jsonb{queryJson}})
		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("CreateQuery:WithNoEventsQuery", func(t *testing.T) {
		rTitle := U.RandomString(5)
		query := model.Query{
			EventsCondition:      model.EventCondAnyGivenEvent,
			From:                 1556602834,
			To:                   1557207634,
			Type:                 model.QueryTypeEventsOccurrence,
			EventsWithProperties: []model.QueryEventWithProperties{}, // invalid, no events.
			OverridePeriod:       true,
		}

		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle,
			Type:  model.QueryTypeSavedQuery,
			Query: &postgres.Jsonb{queryJson}})
		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("CreateQuery:WithNoTitle", func(t *testing.T) {
		query := model.Query{
			EventsCondition: model.EventCondAnyGivenEvent,
			From:            1556602834,
			To:              1557207634,
			Type:            model.QueryTypeEventsOccurrence,
			EventsWithProperties: []model.QueryEventWithProperties{
				model.QueryEventWithProperties{
					Name: "event1",
				},
			},
			OverridePeriod: true,
		}

		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: "",
			Type:  model.QueryTypeSavedQuery,
			Query: &postgres.Jsonb{queryJson}})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAPIUpdateSavedQueryHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rTitle := U.RandomString(5)
	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}

	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle,
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: queryJson}})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)

	queryId := uint64(responseMap["id"].(float64))
	rTitle1 := U.RandomString(5)

	w = sendUpdateSavedQueryReq(r, project.ID, queryId, agent, &H.SavedQueryUpdatePayload{Title: rTitle1})
	assert.Equal(t, http.StatusAccepted, w.Code)

	query1, errCode := store.GetStore().GetQueryWithQueryId(project.ID, queryId)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, query1.Title, rTitle1)
	assert.Equal(t, 2, query1.Type)
}

func TestAPIDeleteSavedQueryHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rTitle := U.RandomString(5)
	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}

	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle,
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{queryJson}})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)

	queryId := uint64(responseMap["id"].(float64))
	shareString := responseMap["id_text"].(string)

	// Create public shareable url
	w = sendCreateShareableUrlReq(r, project.ID, agent, H.ShareableURLParams{
		EntityID:        queryId,
		EntityType:      model.ShareableURLEntityTypeQuery,
		ShareType:       model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	// Check if query is shared publicly
	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendDeleteSavedQueryReq(r, project.ID, queryId, agent)
	assert.Equal(t, http.StatusAccepted, w.Code)

	// Try to get deleted query via shareable url
	w = executeSharedQueryReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	_, errCode := store.GetStore().GetQueryWithQueryId(project.ID, queryId)
	assert.Equal(t, http.StatusNotFound, errCode)
}

func TestAPIGetQueriesHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	rTitle := U.RandomString(5)
	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			model.QueryEventWithProperties{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}

	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle,
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{queryJson}})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)

	queryId := uint64(responseMap["id"].(float64))
	shareString := responseMap["id_text"].(string)

	// Get query without agent
	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Get query with agent
	w = executeSharedQueryReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	// Create public shareable url

	w = sendCreateShareableUrlReq(r, project.ID, agent, H.ShareableURLParams{
		EntityID:        queryId,
		EntityType:      model.ShareableURLEntityTypeQuery,
		ShareType:       model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap = DecodeJSONResponseToMap(w.Body)
	shareID := responseMap["id"].(string)

	// Get query with shareable url without agent
	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete shareable url
	w = sendDeleteShareableUrlReq(r, project.ID, agent, shareID)
	assert.Equal(t, http.StatusOK, w.Code)

	// Get query without agent
	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Get query with agent
	w = executeSharedQueryReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	rTitle1 := U.RandomString(5)
	w = sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{Title: rTitle1,
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{queryJson}})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetSavedQueriesReq(r, project.ID, agent)
	assert.Equal(t, http.StatusFound, w.Code)

	var queries []model.Queries
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&queries); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 2, len(queries))

	// Test access of the agent to Demo project queries
	C.GetConfig().DemoProjectIds = append(C.GetConfig().DemoProjectIds, project.ID)
	b := true
	C.GetConfig().EnableDemoReadAccess = &b

	agent2, errCode := SetupAgentReturnDAO(getRandomEmail(), "+1343545")
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, agent2)

	w = executeSharedQueryReq(r, project.ID, agent2, queries[0].IdText)
	assert.Equal(t, http.StatusOK, w.Code)
}
