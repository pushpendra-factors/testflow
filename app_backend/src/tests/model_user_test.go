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
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProject(&M.Project{Name: randomProjectName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	start := time.Now()

	// Test successful create user.
	user, errCode := M.CreateUser(&M.User{ProjectId: projectId})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.True(t, len(user.ID) > 30)
	assert.Equal(t, projectId, user.ProjectId)
	assert.True(t, user.CreatedAt.After(start))
	assert.True(t, user.UpdatedAt.After(start))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
	assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, user.Properties)
	// Test Get User on the created one.
	retUser, errCode := M.GetUser(projectId, user.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(user.CreatedAt.Sub(retUser.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(user.UpdatedAt.Sub(retUser.UpdatedAt).Seconds()) < 0.1)
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	retUser.CreatedAt = time.Time{}
	retUser.UpdatedAt = time.Time{}
	assert.Equal(t, user, retUser)
	// Test Get User with wrong project id.
	retUser, errCode = M.GetUser(projectId+1, user.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retUser)

	// Test successful create user with customer_user_id and properties.
	customerUserId := "customer_id"
	properties := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "india", "age": 30, "paid": true}`))}
	user, errCode = M.CreateUser(&M.User{ProjectId: projectId, CustomerUserId: customerUserId, Properties: properties})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, customerUserId, user.CustomerUserId)
	assert.True(t, len(user.ID) > 30)
	assert.Equal(t, projectId, user.ProjectId)
	assert.True(t, user.CreatedAt.After(start))
	assert.True(t, user.UpdatedAt.After(start))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
	assert.Equal(t, properties, user.Properties)
	// Creating again with the same customer_user_id with no properties.
	user, errCode = M.CreateUser(&M.User{ProjectId: projectId, CustomerUserId: customerUserId})
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, user)

	// Test Get User on random id.
	randomId := U.RandomLowerAphaNumString(15)
	retUser, errCode = M.GetUser(projectId, randomId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retUser)

	// Test Bad input by providing id.
	user, errCode = M.CreateUser(&M.User{ID: randomId, ProjectId: projectId})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, user)
}
