package tests

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"math"
	"net/http"
	"testing"
	"time"

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
	// Should respond with last user of customer_user instead of creating.
	newUser, newUserErrorCode := M.CreateUser(&M.User{ProjectId: projectId, CustomerUserId: customerUserId})
	assert.Equal(t, M.DB_SUCCESS, newUserErrorCode)
	lastUser, lastUserErrorCode := M.GetUserLatestByCustomerUserId(projectId, customerUserId)
	assert.Equal(t, M.DB_SUCCESS, lastUserErrorCode)
	assert.Equal(t, lastUser.ID, newUser.ID)

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

func TestDBGetUsers(t *testing.T) {
	// Initialize a project for the user.
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProject(&M.Project{Name: randomProjectName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	// Create 100 users.
	var users []M.User
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		user, errCode := M.CreateUser(&M.User{ProjectId: projectId})
		assert.Equal(t, M.DB_SUCCESS, errCode)
		assert.True(t, len(user.ID) > 30)
		users = append(users, *user)
	}

	var offset uint64 = 0
	var limit uint64 = 10
	retUsers, errCode := M.GetUsers(projectId, offset, limit)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, limit, uint64(len(retUsers)))
	assertUsersWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	retUsers, errCode = M.GetUsers(projectId, offset, limit)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, limit, uint64(len(retUsers)))
	assertUsersWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	retUsers, errCode = M.GetUsers(projectId, offset, limit)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, numUsers-95, len(retUsers))
	assertUsersWithOffset(t, users[offset:numUsers], retUsers)
}

func assertUsersWithOffset(t *testing.T, expectedUsers []M.User, actualUsers []M.User) {
	assert.Equal(t, len(expectedUsers), len(actualUsers))
	for i := 0; i < len(actualUsers); i++ {
		expectedUser := expectedUsers[i]
		actualUser := actualUsers[i]
		// time.Time is not exactly same. Checking within an error threshold.
		assert.True(t, math.Abs(expectedUser.CreatedAt.Sub(actualUser.CreatedAt).Seconds()) < 0.1)
		assert.True(t, math.Abs(expectedUser.UpdatedAt.Sub(actualUser.UpdatedAt).Seconds()) < 0.1)
		expectedUser.CreatedAt = time.Time{}
		expectedUser.UpdatedAt = time.Time{}
		actualUser.CreatedAt = time.Time{}
		actualUser.UpdatedAt = time.Time{}
		assert.Equal(t, expectedUser, actualUser)
	}
}
