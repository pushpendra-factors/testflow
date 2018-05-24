package tests

import (
	"encoding/json"
	"math"
	M "model"
	"net/http"
	"testing"
	"time"
	U "util"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetUser(t *testing.T) {
	// Initialize a project for the user.
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.NotNil(t, project)
	project_id := project.ID

	start := time.Now()

	// Test successful create user.
	user, err_code := M.CreateUser(&M.User{ProjectId: project_id})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.True(t, len(user.ID) > 30)
	assert.Equal(t, project_id, user.ProjectId)
	assert.True(t, user.CreatedAt.After(start))
	assert.True(t, user.UpdatedAt.After(start))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
	assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, user.Properties)
	// Test Get Project on the created one.
	get_user, err_code := M.GetUser(user.ID)
	assert.Equal(t, M.DB_SUCCESS, err_code)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(user.CreatedAt.Sub(get_user.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(user.UpdatedAt.Sub(get_user.UpdatedAt).Seconds()) < 0.1)
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	get_user.CreatedAt = time.Time{}
	get_user.UpdatedAt = time.Time{}
	assert.Equal(t, user, get_user)

	// Test Get User on random id.
	random_id := U.RandomLowerAphaNumString(15)
	get_user, err_code = M.GetUser(random_id)
	assert.Equal(t, http.StatusNotFound, err_code)
	assert.Nil(t, get_user)

	// Test Bad input by providing id.
	user, err_code = M.CreateUser(&M.User{ID: random_id, ProjectId: project_id})
	assert.Equal(t, http.StatusBadRequest, err_code)
	assert.Nil(t, user)
}
