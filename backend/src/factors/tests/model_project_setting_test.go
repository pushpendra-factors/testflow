package tests

import (
	M "factors/model"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBUpdateProjectSettings(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test UpdateProjectSetting.
	autoTrack := true
	updatedPSettings, errCode := M.UpdateProjectSettings(project.ID,
		&M.ProjectSetting{AutoTrack: &autoTrack})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	assert.Equal(t, autoTrack, *updatedPSettings.AutoTrack)

	// Test updating one column and another column should not be
	// updated with default value.
	intSegment := true
	updatedPSettings, errCode = M.UpdateProjectSettings(project.ID,
		&M.ProjectSetting{IntSegment: &intSegment})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	projectSetting, errCode := M.GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	// auto_track should stay false.
	assert.Equal(t, autoTrack, *projectSetting.AutoTrack)
	assert.Equal(t, intSegment, *projectSetting.IntSegment)
	assert.Equal(t, true, *projectSetting.ExcludeBot) // default state

	agentUUID := agent.UUID
	accountId := U.RandomLowerAphaNumString(6)
	updatedPSettings, errCode = M.UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntAdwordsCustomerAccountId: &accountId, IntAdwordsEnabledAgentUUID: &agentUUID})
	assert.Equal(t, errCode, http.StatusAccepted)
	projectSetting, errCode = M.GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	assert.Equal(t, true, *projectSetting.ExcludeBot)
	assert.Equal(t, accountId, *projectSetting.IntAdwordsCustomerAccountId)
	assert.Equal(t, agentUUID, *projectSetting.IntAdwordsEnabledAgentUUID)

	// Test UpdateProjectSetting without projectId.
	autoTrack = true
	updatedPSettings, errCode = M.UpdateProjectSettings(0,
		&M.ProjectSetting{AutoTrack: &autoTrack})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, updatedPSettings)
}
