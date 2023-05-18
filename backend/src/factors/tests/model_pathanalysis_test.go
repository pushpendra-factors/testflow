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

func TestCreatePathAnalysis(t *testing.T) {

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	t.Run("CreatePathAnalysisEntity:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, "")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)
	})

	t.Run("GetPathAnalysisEntity:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, "")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		entityList, errCode := store.GetStore().GetAllPathAnalysisEntityByProject(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, entityList)
	})

	t.Run("CreatePathAnalysisEntity:Title already present:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, "")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		entity, errCode, _ = store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve1", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, "")
		assert.Equal(t, http.StatusConflict, errCode)
		assert.Nil(t, entity)
	})

	t.Run("CreatePathAnalysisEntity:PathAnalysis entity already present in DB:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, "")
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		rName2 := U.RandomString(5)
		entity, errCode, _ = store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName2, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}}, "")
		assert.Equal(t, http.StatusConflict, errCode)
		assert.Nil(t, entity)
	})
}

func TestDeletePathAnalysis(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	rName1 := U.RandomString(5)
	entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
		Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"}, model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
			{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
		}}, "")
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, entity)

	t.Run("DeletePathAnalysisEntity:valid", func(t *testing.T) {
		errCode, errMsg = store.GetStore().DeletePathAnalysisEntity(project.ID, entity.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Empty(t, errMsg)
	})
}

func TestCreatePathAnalysisHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	t.Run("CreatePathAnalysis:WithValidQuery", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query := &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1},
			NumberOfSteps: 4,
			ExcludeEvents: []model.PathAnalysisEvent{{Label: "e1"}, {Label: "E2"}},
			Filter:        []model.QueryProperty{{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"}}}

		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendCreatePathAnalysisReques(r, project.ID, agent, &postgres.Jsonb{RawMessage: queryJson})
		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("CreatePathAnalysisEntity:Include & Exclude both Events Provided: Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query := &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4,
			IncludeEvents: []model.PathAnalysisEvent{{Label: "e1"}, {Label: "E2"}},
			ExcludeEvents: []model.PathAnalysisEvent{{Label: "e1"}, {Label: "E2"}},
			Filter:        []model.QueryProperty{{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"}}}

		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendCreatePathAnalysisReques(r, project.ID, agent, &postgres.Jsonb{RawMessage: queryJson})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func TestPathAnalysisLimitHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)
	t.Run("CreatePathAnalysisEntity:PathAnalysis limit reached:Invalid", func(t *testing.T) {

		limit := model.BuildLimit
		for i := 0; i < limit; i++ {
			rName1 := U.RandomString(5)
			query := &model.PathAnalysisQuery{
				Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4,
				IncludeEvents: []model.PathAnalysisEvent{{Label: "e1"}, {Label: "E2"}},
				Filter: []model.QueryProperty{
					{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
				},
			}

			queryJson, err := json.Marshal(query)
			assert.Nil(t, err)

			w := sendCreatePathAnalysisReques(r, project.ID, agent, &postgres.Jsonb{RawMessage: queryJson})
			assert.Equal(t, http.StatusCreated, w.Code)
		}

		rName2 := U.RandomString(5)
		query := &model.PathAnalysisQuery{
			Title: rName2, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName2}, NumberOfSteps: 4,
			IncludeEvents: []model.PathAnalysisEvent{{Label: "e1"}, {Label: "E2"}},
			Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			},
		}
		queryJson, err := json.Marshal(query)
		assert.Nil(t, err)

		w := sendCreatePathAnalysisReques(r, project.ID, agent, &postgres.Jsonb{RawMessage: queryJson})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
func sendCreatePathAnalysisReques(r *gin.Engine, projectId int64, agent *model.Agent, query *postgres.Jsonb) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, fmt.Sprintf("/projects/%d/v1/pathanalysis", projectId)).
		WithPostParams(query).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error creating create pathanalysis request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
