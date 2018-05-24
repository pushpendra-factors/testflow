package tests

import (
	"math"
	M "model"
	"net/http"
	"testing"
	"time"
	U "util"

	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetProject(t *testing.T) {
	start := time.Now()

	// Test successful create project.
	project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: project_name})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.True(t, project.ID > 0)
	assert.Equal(t, project_name, project.Name)
	assert.Equal(t, 32, len(project.APIKey))
	assert.True(t, project.CreatedAt.After(start))
	assert.True(t, project.UpdatedAt.After(start))
	assert.Equal(t, project.CreatedAt, project.UpdatedAt)

	// Test API key is overwritten and cannot be provided.
	previous_project_id := project.ID
	// Random API Key.
	provided_api_key := U.RandomLowerAphaNumString(32)
	// Reusing the same name. Name is not meant to be unique.
	project, err_code = M.CreateProject(&M.Project{Name: project_name, APIKey: provided_api_key})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.True(t, project.ID > previous_project_id)
	assert.Equal(t, project_name, project.Name)
	assert.Equal(t, 32, len(project.APIKey))
	assert.NotEqual(t, provided_api_key, project.APIKey)
	assert.True(t, project.CreatedAt.After(start))
	assert.True(t, project.UpdatedAt.After(start))
	assert.Equal(t, project.CreatedAt, project.UpdatedAt)
	// Test Get Project on the created one.
	get_project, err_code := M.GetProject(project.ID)
	assert.Equal(t, M.DB_SUCCESS, err_code)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(project.CreatedAt.Sub(get_project.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(project.UpdatedAt.Sub(get_project.UpdatedAt).Seconds()) < 0.1)
	project.CreatedAt = time.Time{}
	project.UpdatedAt = time.Time{}
	get_project.CreatedAt = time.Time{}
	get_project.UpdatedAt = time.Time{}
	assert.Equal(t, project, get_project)

	// Test Get Project on random id.
	var random_id uint64 = 12345 // Assuming this to be random. Don't be surprised if this test fails some day.
	get_project, err_code = M.GetProject(random_id)
	assert.Equal(t, http.StatusNotFound, err_code)
	assert.Nil(t, get_project)

	// Test Bad input by providing id.
	// Reusing the same name. Name is not meant to be unique.
	project, err_code = M.CreateProject(&M.Project{Name: project_name, ID: previous_project_id + 10})
	assert.Equal(t, http.StatusBadRequest, err_code)
	assert.Nil(t, project)
}
