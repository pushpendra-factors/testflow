package tests

import (
	M "factors/model"
	U "factors/util"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetProject(t *testing.T) {
	start := time.Now()

	// Test successful create project.
	projectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: projectName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.True(t, project.ID > 0)
	assert.Equal(t, projectName, project.Name)
	assert.Equal(t, 32, len(project.Token))
	assert.True(t, project.CreatedAt.After(start))
	assert.True(t, project.UpdatedAt.After(start))
	assert.Equal(t, project.CreatedAt, project.UpdatedAt)

	// Test token is overwritten and cannot be provided.
	previousProjectId := project.ID
	// Random Token.
	providedToken := U.RandomLowerAphaNumString(32)
	// Reusing the same name. Name is not meant to be unique.
	project, errCode = M.CreateProjectWithDependencies(&M.Project{Name: projectName, Token: providedToken})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.True(t, project.ID > previousProjectId)
	assert.Equal(t, projectName, project.Name)
	assert.Equal(t, 32, len(project.Token))
	assert.NotEqual(t, providedToken, project.Token)
	assert.True(t, project.CreatedAt.After(start))
	assert.True(t, project.UpdatedAt.After(start))
	assert.Equal(t, project.CreatedAt, project.UpdatedAt)
	// Test Get Project on the created one.
	getProject, errCode := M.GetProject(project.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(project.CreatedAt.Sub(getProject.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(project.UpdatedAt.Sub(getProject.UpdatedAt).Seconds()) < 0.1)
	project.CreatedAt = time.Time{}
	project.UpdatedAt = time.Time{}
	getProject.CreatedAt = time.Time{}
	getProject.UpdatedAt = time.Time{}
	assert.Equal(t, project, getProject)

	// Test Get Project on random id.
	var randomId uint64 = 12345 // Assuming this to be random. Don't be surprised if this test fails some day.
	getProject, errCode = M.GetProject(randomId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, getProject)

	// Test Bad input by providing id.
	// Reusing the same name. Name is not meant to be unique.
	project, errCode = M.CreateProjectWithDependencies(&M.Project{Name: projectName, ID: previousProjectId + 10})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, project)

	// Test Get Project by a token.
	// Bad input.
	project, errCode = M.GetProjectByToken("")
	assert.Equal(t, http.StatusBadRequest, errCode)

	// RandomInput
	project, errCode = M.GetProjectByToken(U.RandomLowerAphaNumString(32))
	assert.Equal(t, http.StatusNotFound, errCode)

	// Check corresponding project returned with token.
	project, errCode = M.CreateProjectWithDependencies(&M.Project{Name: projectName})
	rProject, rErrCode := M.GetProjectByToken(project.Token)
	assert.Equal(t, M.DB_SUCCESS, rErrCode)
	assert.Equal(t, project.ID, rProject.ID)

	// Test CreateProjectWithDependencies
	start = time.Now()
	projectWithDeps, errCode := M.CreateProjectWithDependencies(&M.Project{Name: U.RandomLowerAphaNumString(15)})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.True(t, projectWithDeps.ID > 0)
	assert.Equal(t, 32, len(projectWithDeps.Token))
	assert.True(t, projectWithDeps.CreatedAt.After(start))
	assert.True(t, projectWithDeps.UpdatedAt.After(start))
	assert.Equal(t, projectWithDeps.CreatedAt, projectWithDeps.UpdatedAt)

	// Test depedencies creation - ProjectSettings.
	ps, errCode := M.GetProjectSetting(projectWithDeps.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, ps)
}
