package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExplainV2(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	t.Run("CreateExplainV2Entity:valid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query := model.ExplainV2Query{rName1, "start_event_1", "end_event_1", []string{"ic_eve_1", "ic_eve_2"}, 1670371748, 1670371750}
		entity, errCode, errMsg := store.GetStore().CreateExplainV2Entity(agent.UUID, project.ID, &query)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)
	})

	t.Run("GetExplainV2Entity:valid", func(t *testing.T) {
		entity, errCode := store.GetStore().GetAllExplainV2EntityByProject(project.ID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.NotNil(t, entity)
	})

	t.Run("CreateExplainV2Entity:Title already present:Invalid", func(t *testing.T) {
		rName1 := U.RandomString(5)
		query1 := model.ExplainV2Query{rName1, "start_event_1", "end_event_1", []string{"ic_eve_1", "ic_eve_2"}, 1670371748, 1670371750}
		query2 := model.ExplainV2Query{rName1, "start_event_1", "end_event_1", []string{"ic_eve_1", "ic_eve_2"}, 1670371748, 1670371750}

		entity, errCode, errMsg := store.GetStore().CreateExplainV2Entity(agent.UUID, project.ID, &query1)
		assert.Equal(t, http.StatusCreated, errCode)
		assert.NotNil(t, entity)
		assert.Empty(t, errMsg)

		entity, errCode, _ = store.GetStore().CreateExplainV2Entity(agent.UUID, project.ID, &query2)
		assert.Equal(t, http.StatusConflict, errCode)
		assert.Nil(t, entity)
	})

}

func TestDeleteExplainV2(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	rName1 := U.RandomString(5)
	query1 := model.ExplainV2Query{rName1, "start_event_1", "end_event_1", []string{"ic_eve_1", "ic_eve_2"}, 1670371748, 1670371750}

	entity, errCode, errMsg := store.GetStore().CreateExplainV2Entity(agent.UUID, project.ID, &query1)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Empty(t, errMsg)
	assert.NotNil(t, entity)

	t.Run("DeleteExplainV2Entity:valid", func(t *testing.T) {
		errCode, errMsg = store.GetStore().DeleteExplainV2Entity(project.ID, entity.ID)
		assert.Equal(t, http.StatusAccepted, errCode)
		assert.Empty(t, errMsg)
	})
}
