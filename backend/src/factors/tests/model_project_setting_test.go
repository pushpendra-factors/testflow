package tests

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBUpdateProjectSettings(t *testing.T) {
	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test UpdateProjectSetting.
	autoTrack := true
	updatedPSettings, errCode := store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{AutoTrack: &autoTrack})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	assert.Equal(t, autoTrack, *updatedPSettings.AutoTrack)

	// Test updating one column and another column should not be
	// updated with default value.
	intSegment := true
	updatedPSettings, errCode = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntSegment: &intSegment})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	projectSetting, errCode := store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	// auto_track should stay false.
	assert.Equal(t, autoTrack, *projectSetting.AutoTrack)
	assert.Equal(t, intSegment, *projectSetting.IntSegment)
	assert.Equal(t, true, *projectSetting.ExcludeBot) // default state

	intRudderstack := true
	updatedPSettings, errCode = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntRudderstack: &intRudderstack})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	projectSetting, errCode = store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	// auto_track should stay false.
	assert.Equal(t, autoTrack, *projectSetting.AutoTrack)
	assert.Equal(t, intSegment, *projectSetting.IntSegment)
	assert.Equal(t, intRudderstack, *projectSetting.IntRudderstack)
	assert.Equal(t, true, *projectSetting.ExcludeBot) // default state

	intSegment = false
	updatedPSettings, errCode = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntSegment: &intSegment})
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	projectSetting, errCode = store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	// auto_track should stay false.
	assert.Equal(t, autoTrack, *projectSetting.AutoTrack)
	assert.Equal(t, intSegment, *projectSetting.IntSegment)
	assert.Equal(t, intRudderstack, *projectSetting.IntRudderstack)
	assert.Equal(t, true, *projectSetting.ExcludeBot) // default state

	agentUUID := agent.UUID
	accountId := U.RandomLowerAphaNumString(6)
	updatedPSettings, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &accountId, IntAdwordsEnabledAgentUUID: &agentUUID})
	assert.Equal(t, errCode, http.StatusAccepted)
	projectSetting, errCode = store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	assert.Equal(t, true, *projectSetting.ExcludeBot)
	assert.Equal(t, accountId, *projectSetting.IntAdwordsCustomerAccountId)
	assert.Equal(t, agentUUID, *projectSetting.IntAdwordsEnabledAgentUUID)

	// Test UpdateProjectSetting without projectId.
	autoTrack = true
	updatedPSettings, errCode = store.GetStore().UpdateProjectSettings(0,
		&model.ProjectSetting{AutoTrack: &autoTrack})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, updatedPSettings)

	// Test clean adwords customer account id on update.
	adwordsCustomerAccountId := "899-900-900"
	updatedPSettings, errCode = store.GetStore().UpdateProjectSettings(project.ID,
		&model.ProjectSetting{IntAdwordsCustomerAccountId: &adwordsCustomerAccountId})
	assert.Equal(t, errCode, http.StatusAccepted)
	projectSetting, errCode = store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	assert.Equal(t, "899900900", *projectSetting.IntAdwordsCustomerAccountId)
}

func TestGetProjectSettingByKeyWithTimeout(t *testing.T) {
	project, _ := SetupProjectReturnDAO()
	assert.NotNil(t, project)

	// Should not timeout.
	// If the db queries are slow in local.
	// This case will start failing. So setting a high timeout period 5 seconds.
	projectSetting, errCode := store.GetStore().GetProjectSettingByKeyWithTimeout("token",
		project.Token, time.Second*5)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotNil(t, projectSetting)

	// Should timeout.
	// Assuming that no database environment can execute
	// the query in less than 1 micro seconds.
	projectSetting, errCode = store.GetStore().GetProjectSettingByKeyWithTimeout("token",
		project.Token, time.Microsecond*1)
	assert.Equal(t, errCode, http.StatusInternalServerError)
	assert.Nil(t, projectSetting)

	// Should return from default, as flag is set.
	C.GetConfig().UseDefaultProjectSettingForSDK = true
	projectSetting, errCode = store.GetStore().GetProjectSettingByKeyWithTimeout("token",
		project.Token, time.Microsecond*1)
	assert.Equal(t, errCode, http.StatusNotModified)
	assert.NotNil(t, projectSetting)
}

func TestEnableBigqueryArchivalForProject(t *testing.T) {
	project, _ := SetupProjectReturnDAO()
	assert.NotNil(t, project)

	projectSetting, errCode := store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	assert.False(t, *projectSetting.BigqueryEnabled)
	assert.False(t, *projectSetting.ArchiveEnabled)

	errCode = store.GetStore().EnableBigqueryArchivalForProject(project.ID)
	assert.Equal(t, http.StatusAccepted, errCode)

	projectSetting, errCode = store.GetStore().GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	assert.True(t, *projectSetting.BigqueryEnabled)
	assert.True(t, *projectSetting.ArchiveEnabled)
}
