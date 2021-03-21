package tests

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	customerAccountId := U.RandomLowerAphaNumString(5)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId,
		IntAdwordsEnabledAgentUUID:  &agent.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	agentAdwordsToken := U.RandomLowerAphaNumString(10)
	errCode = store.GetStore().UpdateAgentIntAdwordsRefreshToken(agent.UUID, agentAdwordsToken)
	assert.Equal(t, http.StatusAccepted, errCode)

	adwordsLastSyncInfo, errCode := store.GetStore().GetAllAdwordsLastSyncInfoForAllProjects()
	assert.Equal(t, http.StatusOK, errCode)
	for _, alsi := range adwordsLastSyncInfo {
		if alsi.ProjectId == project.ID {
			assert.Equal(t, agentAdwordsToken, alsi.RefreshToken)
			assert.Equal(t, customerAccountId, alsi.CustomerAccountId)
		}
	}

	// Test project's corresponding access token map.
	project1, agent1, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	customerAccountId1 := U.RandomLowerAphaNumString(5)
	_, errCode = store.GetStore().UpdateProjectSettings(project1.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &customerAccountId1,
		IntAdwordsEnabledAgentUUID:  &agent1.UUID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)
	agentAdwordsToken1 := U.RandomLowerAphaNumString(10)
	errCode = store.GetStore().UpdateAgentIntAdwordsRefreshToken(agent1.UUID, agentAdwordsToken1)
	assert.Equal(t, http.StatusAccepted, errCode)

	adwordsLastSyncInfo1, errCode := store.GetStore().GetAllAdwordsLastSyncInfoForAllProjects()
	assert.Equal(t, http.StatusOK, errCode)
	for _, alsi := range adwordsLastSyncInfo1 {
		if alsi.ProjectId == project.ID {
			assert.Equal(t, agentAdwordsToken, alsi.RefreshToken)
			assert.Equal(t, customerAccountId, alsi.CustomerAccountId)
		}

		if alsi.ProjectId == project1.ID {
			assert.Equal(t, agentAdwordsToken1, alsi.RefreshToken)
			assert.Equal(t, customerAccountId1, alsi.CustomerAccountId)
		}
	}
}
