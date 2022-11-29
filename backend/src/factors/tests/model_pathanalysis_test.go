package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePathAnalysis(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	t.Run("CreatePathAnalysisEntity:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)
	})

	t.Run("GetPathAnalysisEntity:valid", func(t *testing.T) {
		entity, errCode := store.GetStore().GetAllPathAnalysisEntityByProject(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, entity)
	})

	t.Run("CreatePathAnalysisEntity:Include & Exclude both Events Provided: Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, _ := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}},
			ExcludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}},
			Filter:        []model.QueryProperty{{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"}}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, entity)
	})

	t.Run("CreatePathAnalysisEntity:Title already present:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		entity, errCode, _ = store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve1", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, entity)
	})

	t.Run("CreatePathAnalysisEntity:PathAnalysis entity already present in DB:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents:[]model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		rName2 := U.RandomString(5)
		entity, errCode, _ = store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName2, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, entity)
	})

	t.Run("CreatePathAnalysisEntity:PathAnalysis limit reached:Invalid", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			rName1 := U.RandomString(5)
			entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
				Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
					{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
				}})
			assert.Equal(t, http.StatusCreated, errCode)
			assert.NotNil(t, entity)
			assert.Empty(t, errMsg)
		}

		rName2 := U.RandomString(5)
		entity, errCode, _ := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
			Title: rName2, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName2}, NumberOfSteps: 4, IncludeEvents: []model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
				{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
			}})
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Nil(t, entity)
	})

}

func TestDeletePathAnalysis(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	rName1 := U.RandomString(5)
	entity, errCode, errMsg := store.GetStore().CreatePathAnalysisEntity(agent.UUID, project.ID, &model.PathAnalysisQuery{
		Title: rName1, EventType: "eve", Event: model.PathAnalysisEvent{Label: rName1}, NumberOfSteps: 4, IncludeEvents:[]model.PathAnalysisEvent{model.PathAnalysisEvent{Label: "e1"},model.PathAnalysisEvent{Label: "E2"}}, Filter: []model.QueryProperty{
			{Entity: "", Type: "categorical", Property: "campaign", Operator: "equals", LogicalOp: "AND"},
		}})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, entity)

	t.Run("DeletePathAnalysisEntity:valid", func(t *testing.T) {
		errCode, errMsg = store.GetStore().DeletePathAnalysisEntity(project.ID, entity.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Empty(t, errMsg)
	})
}
