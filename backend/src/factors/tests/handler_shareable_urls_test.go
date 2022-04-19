package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
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

func executeSharedQueryReq(r *gin.Engine, projectId uint64, agent *model.Agent, shareString string) *httptest.ResponseRecorder {

	var rb *U.RequestBuilder
	if agent != nil {
		cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
		if err != nil {
			log.WithError(err).Error("Error creating cookie data.")
		}
		rb = C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/query?query_id=%s", projectId, shareString)).
			WithCookie(&http.Cookie{
				Name:   C.GetFactorsCookieName(),
				Value:  cookieData,
				MaxAge: 1000,
			})
	} else {
		rb = C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/query?query_id=%s", projectId, shareString))
	}

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating get shared queries req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendGetShareableUrlReq(r *gin.Engine, projectID uint64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/shareable_url", projectID)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating get shareable url req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendCreateShareableUrlReq(r *gin.Engine, projectID uint64, agent *model.Agent, share *H.ShareableURLParams) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/shareable_url", projectID)).
		WithPostParams(share).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create shareable url req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendDeleteShareableUrlReq(r *gin.Engine, projectID uint64, agent *model.Agent, shareId string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/shareable_url/%s", projectID, shareId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating delete shareable url req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendRevokeShareableUrlReq(r *gin.Engine, projectID uint64, agent *model.Agent, shareId string) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, fmt.Sprintf("/projects/%d/shareable_url/revoke/%s", projectID, shareId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating revoke shareable url req.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPIGetShareableURLsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent1, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent1)

	agent2, err := SetupAgentWithProject(project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, agent2)

	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent1, &H.SavedQueryRequestPayload{
		Title: U.RandomString(5),
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)
	queryId := uint64(responseMap["id"].(float64))
	
	w = sendCreateShareableUrlReq(r, project.ID, agent1, &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)
	
	w = sendGetShareableUrlReq(r, project.ID, agent1)
	assert.Equal(t, http.StatusOK, w.Code)

	var shares []model.ShareableURL
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, agent2)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendCreateShareableUrlReq(r, project.ID, agent2, &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent2)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendDeleteSavedQueryReq(r, project.ID, queryId, agent1)
	assert.Equal(t, http.StatusAccepted, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent1)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, agent2)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))
}

func TestAPICreateShareableURLHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent1, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent1)

	agent2, err := SetupAgentWithProject(project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, agent2)

	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent1, &H.SavedQueryRequestPayload{
		Title: U.RandomString(5),
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)
	queryId := uint64(responseMap["id"].(float64))

	shareableUrl := &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	}

	w = sendCreateShareableUrlReq(r, project.ID, agent1, shareableUrl)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent1)
	assert.Equal(t, http.StatusOK, w.Code)

	var shares []model.ShareableURL
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, agent2)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendCreateShareableUrlReq(r, project.ID, agent1, shareableUrl)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = sendCreateShareableUrlReq(r, project.ID, agent2, shareableUrl)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent1)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, agent2)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendCreateShareableUrlReq(r, project.ID, agent2, shareableUrl)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = sendDeleteSavedQueryReq(r, project.ID, queryId, agent1)
	assert.Equal(t, http.StatusAccepted, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent1)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, agent2)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendCreateShareableUrlReq(r, project.ID, agent1, shareableUrl)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPIDeleteShareableURLHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{
		Title: U.RandomString(5),
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)
	queryId := uint64(responseMap["id"].(float64))
	shareString := responseMap["id_text"].(string)

	shareableUrl := &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	}

	w = sendCreateShareableUrlReq(r, project.ID, agent, shareableUrl)
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap = DecodeJSONResponseToMap(w.Body)
	shareableUrlId := responseMap["id"].(string)

	w = sendGetShareableUrlReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	var shares []model.ShareableURL
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = executeSharedQueryReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendDeleteShareableUrlReq(r, project.ID, agent, shareableUrlId)
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = executeSharedQueryReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIRevokeAllShareableURLHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, admin, err := SetupProjectWithAdminAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, admin)

	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, admin, &H.SavedQueryRequestPayload{
		Title: U.RandomString(5),
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)
	queryId := uint64(responseMap["id"].(float64))
	shareString1 := responseMap["id_text"].(string)
	
	w = sendCreateShareableUrlReq(r, project.ID, admin, &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)
	
	w = sendCreateQueryReq(r, project.ID, admin, &H.SavedQueryRequestPayload{
		Title: U.RandomString(5),
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap = DecodeJSONResponseToMap(w.Body)
	queryId = uint64(responseMap["id"].(float64))
	shareString2 := responseMap["id_text"].(string)
	
	w = sendCreateShareableUrlReq(r, project.ID, admin, &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, admin)
	assert.Equal(t, http.StatusOK, w.Code)

	var shares []model.ShareableURL
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 2, len(shares))

	agent, err := SetupAgentWithProject(project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, agent)

	w = sendGetShareableUrlReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendRevokeShareableUrlReq(r, project.ID, agent, "all")
	assert.Equal(t, http.StatusForbidden, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, admin)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 2, len(shares))

	w = sendRevokeShareableUrlReq(r, project.ID, admin, "all")
	assert.Equal(t, http.StatusOK, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, admin)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	// Revoke all will reset all query id_texts
	w = executeSharedQueryReq(r, project.ID, agent, shareString1)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = executeSharedQueryReq(r, project.ID, agent, shareString2)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = executeSharedQueryReq(r, project.ID, nil, shareString1)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = executeSharedQueryReq(r, project.ID, nil, shareString2)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIRevokeShareableURLHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, admin, err := SetupProjectWithAdminAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, admin)

	agent, err := SetupAgentWithProject(project.ID)
	assert.Nil(t, err)
	assert.NotNil(t, agent)

	query := model.Query{
		EventsCondition: model.EventCondAnyGivenEvent,
		From:            1556602834,
		To:              1557207634,
		Type:            model.QueryTypeEventsOccurrence,
		EventsWithProperties: []model.QueryEventWithProperties{
			{
				Name: "event1",
			},
		},
		OverridePeriod: true,
	}
	queryJson, err := json.Marshal(query)
	assert.Nil(t, err)

	w := sendCreateQueryReq(r, project.ID, agent, &H.SavedQueryRequestPayload{
		Title: U.RandomString(5),
		Type:  model.QueryTypeSavedQuery,
		Query: &postgres.Jsonb{RawMessage: json.RawMessage(queryJson)},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	responseMap := DecodeJSONResponseToMap(w.Body)
	queryId := uint64(responseMap["id"].(float64))
	shareString := responseMap["id_text"].(string)
	
	w = sendCreateShareableUrlReq(r, project.ID, agent, &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)
	
	var shares []model.ShareableURL
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendCreateShareableUrlReq(r, project.ID, admin, &H.ShareableURLParams{
		EntityID: queryId,
		EntityType: model.ShareableURLEntityTypeQuery,
		ShareType: model.ShareableURLShareTypePublic,
		IsExpirationSet: false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, admin)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendRevokeShareableUrlReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusForbidden, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)
	
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, admin)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 1, len(shares))

	w = sendRevokeShareableUrlReq(r, project.ID, admin, shareString)
	assert.Equal(t, http.StatusOK, w.Code)

	// Query id text is reset after a revoke
	w = executeSharedQueryReq(r, project.ID, agent, shareString)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = executeSharedQueryReq(r, project.ID, nil, shareString)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = sendGetShareableUrlReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)
	
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))

	w = sendGetShareableUrlReq(r, project.ID, admin)
	assert.Equal(t, http.StatusOK, w.Code)

	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&shares); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, 0, len(shares))
}