package tests

import (
	"bufio"
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFilter() ([]model.FactorsGoalRule, error) {
	fr := make([]model.FactorsGoalRule, 0)
	file, err := os.Open("./data/rule.txt")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var fr1 model.FactorsGoalRule
		err := json.Unmarshal([]byte(line), &fr1)
		if err != nil {
			return nil, err
		}
		fr = append(fr, fr1)
	}

	return fr, nil
}

func createQueryV1() ([]model.ExplainV2Query, error) {
	qrys := make([]model.ExplainV2Query, 0)
	rulesFilter, err := readFilter()
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

func TestExplainV2(t *testing.T) {
	// project, agent, err := SetupProjectWithAgentDAO()
	// assert.Nil(t, err)
	project_ID := int64(1000001)
	agent_UUID := "059adfda-083d-4d90-9e34-5ef8ab8f8571"
	t.Run("CreateExplainV2Entity:valid", func(t *testing.T) {
		queries, err := createQueryV1()
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

		queries, err := createQueryV1()
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

}

func TestDeleteExplainV2(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)

	rName1 := U.RandomString(5)
	fr, err := readFilter()
	assert.Nil(t, err)
	query1 := model.ExplainV2Query{Title: rName1, Query: fr[0], StartTimestamp: 1670371748, EndTimestamp: 1670371750}

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
