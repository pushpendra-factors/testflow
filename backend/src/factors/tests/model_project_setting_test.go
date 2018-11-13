package tests

import (
	M "factors/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetProjectSetting(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test CreateProjectSetting with project id.
	projectSetting, errCode := M.CreateProjectSetting(&M.ProjectSetting{ProjectId: project.ID})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, projectSetting)
	// Check auto track default as disabled.
	assert.EqualValues(t, M.AUTO_TRACK_DISABLED, projectSetting.AutoTrack)

	project1, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test CreateProjectSetting with project id.
	projectSetting, errCode = M.CreateProjectSetting(&M.ProjectSetting{ProjectId: project1.ID, AutoTrack: M.AUTO_TRACK_ENABLED})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, projectSetting)
	// Check auto track default as disabled.
	assert.EqualValues(t, M.AUTO_TRACK_ENABLED, projectSetting.AutoTrack)

	// Test CreateProjectSetting without project id.
	projectSetting, errCode = M.CreateProjectSetting(&M.ProjectSetting{ProjectId: 0})
	assert.Equal(t, http.StatusBadRequest, errCode)

	// Test GetProjectSetting with invalid project id.
	projectSetting, errCode = M.GetProjectSetting(project.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, projectSetting)

	// Test GetProjectSetting with invalid project id.
	projectSetting, errCode = M.GetProjectSetting(0)
	assert.Equal(t, http.StatusBadRequest, errCode)
}
