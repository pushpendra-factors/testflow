package tests

import (
	M "factors/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBUpdateProjectSettings(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test UpdateProjectSetting.
	fieldsToBeUpdated := &M.ProjectSetting{AutoTrack: true}
	updatedPSettings, errCode := M.UpdateProjectSettings(project.ID, fieldsToBeUpdated)
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	assert.Equal(t, fieldsToBeUpdated.AutoTrack, updatedPSettings.AutoTrack)

	// Test UpdateProjectSetting with default value of a field. Covers a known bug on gorm with '.Updates'.
	fieldsToBeUpdated = &M.ProjectSetting{AutoTrack: false}
	updatedPSettings, errCode = M.UpdateProjectSettings(project.ID, fieldsToBeUpdated)
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.NotNil(t, updatedPSettings)
	projectSetting, errCode := M.GetProjectSetting(project.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotNil(t, projectSetting)
	assert.Equal(t, fieldsToBeUpdated.AutoTrack, projectSetting.AutoTrack)

	// Test UpdateProjectSetting without projectId.
	fieldsToBeUpdated = &M.ProjectSetting{AutoTrack: true}
	updatedPSettings, errCode = M.UpdateProjectSettings(0, fieldsToBeUpdated)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, updatedPSettings)

	// Todo(Dinesh): This would fail as update won't return error.
	// Not able to use RowsNotAffected which is always 0.

	// Test UpdateProjectSetting with invalid projectId.
	// fieldsToBeUpdated = &M.ProjectSetting{AutoTrack: true}
	// updatedPSettings, errCode = M.UpdateProjectSettings(999999999999, fieldsToBeUpdated)
	// assert.Equal(t, http.StatusNotFound, errCode)
	// assert.Nil(t, updatedPSettings)

}
