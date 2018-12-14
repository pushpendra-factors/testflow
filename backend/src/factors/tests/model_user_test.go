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
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
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
	// Not more than 10ms difference.
	assert.InDelta(t, user.CreatedAt.UnixNano(), user.UpdatedAt.UnixNano(), 1.0e+7)
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
	// Not more than 10ms difference.
	assert.InDelta(t, user.CreatedAt.UnixNano(), user.UpdatedAt.UnixNano(), 1.0e+7)
	assert.Equal(t, properties, user.Properties)

	// Creating again with the same customer_user_id with no properties.
	// Should respond with last user of customer_user instead of creating.
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	newUser, newUserErrorCode := M.CreateUser(&M.User{ProjectId: projectId, CustomerUserId: rCustomerUserId})
	assert.Equal(t, M.DB_SUCCESS, newUserErrorCode)
	lastUser, lastUserErrorCode := M.GetUserLatestByCustomerUserId(projectId, rCustomerUserId)
	assert.Equal(t, M.DB_SUCCESS, lastUserErrorCode)
	assert.Equal(t, newUser.ID, lastUser.ID)

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
	project, errCode := M.CreateProjectWithDependencies(&M.Project{Name: randomProjectName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	var offset uint64 = 0
	var limit uint64 = 10
	// no users have been created
	retUsers, errCode := M.GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Create 100 users.
	var users []M.User
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		user, errCode := M.CreateUser(&M.User{ProjectId: projectId})
		assert.Equal(t, M.DB_SUCCESS, errCode)
		assert.True(t, len(user.ID) > 30)
		users = append(users, *user)
	}

	retUsers, errCode = M.GetUsers(projectId, offset, limit)
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

func TestDBGetUserLatestByCustomerUserId(t *testing.T) {
	// Intialize.
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test latest user return for the customer_user.
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	latestUser, latestUserErrCode := M.CreateUser(&M.User{ProjectId: project.ID, CustomerUserId: rCustomerUserId})
	assert.Equal(t, M.DB_SUCCESS, latestUserErrCode)
	getUser, getUserErrCode := M.GetUserLatestByCustomerUserId(project.ID, rCustomerUserId)
	assert.Equal(t, M.DB_SUCCESS, getUserErrCode)
	assert.Equal(t, latestUser.ID, getUser.ID)

	// Bad input. // Without project scope.
	_, errCode := M.GetUserLatestByCustomerUserId(0, rCustomerUserId)
	assert.NotEqual(t, M.DB_SUCCESS, errCode)

	// Bad input. // Unacceptable customer_user_id
	_, errCode = M.GetUserLatestByCustomerUserId(project.ID, " ")
	assert.NotEqual(t, M.DB_SUCCESS, errCode)
}

func TestDBUpdateUserById(t *testing.T) {
	// Intialize.
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)

	// Test updating a field.
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	updateUser := &M.User{CustomerUserId: rCustomerUserId}
	cuUpdatedUser, errCode := M.UpdateUser(project.ID, user.ID, updateUser)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, rCustomerUserId, cuUpdatedUser.CustomerUserId)
	// Using already tested GetUser method to validate update.
	gUser, gErrCode := M.GetUser(project.ID, user.ID)
	assert.Equal(t, M.DB_SUCCESS, gErrCode)
	// Test CustomerUserId updated or not.
	assert.Equal(t, rCustomerUserId, gUser.CustomerUserId)

	// Test updating ProjectId with other fields
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	uProject, uErr := SetupProjectReturnDAO()
	assert.Nil(t, uErr)
	assert.NotNil(t, uProject)
	updateUser = &M.User{ProjectId: uProject.ID, CustomerUserId: rCustomerUserId}
	_, errCode = M.UpdateUser(project.ID, user.ID, updateUser)
	assert.Equal(t, http.StatusBadRequest, errCode)

	// Bad input. ProjectId.
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	_, errCode = M.UpdateUser(0, user.ID, &M.User{})
	assert.NotEqual(t, M.DB_SUCCESS, errCode)

	// Bad input. UserId.
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	_, errCode = M.UpdateUser(project.ID, "", &M.User{})
	assert.NotEqual(t, M.DB_SUCCESS, errCode)
}

func TestAddUserDefaultProperties(t *testing.T) {
	propertiesMap := U.PropertiesMap{"prop_1": "value_1"}
	err := M.FillUserDefaultProperties(&propertiesMap, "180.151.36.234") // Our gateway IP.
	assert.Nil(t, err)
	assert.NotNil(t, propertiesMap[U.UP_INTERNAL_IP])
	assert.NotNil(t, propertiesMap[U.UP_COUNTRY])
	assert.Equal(t, "IN", propertiesMap[U.UP_COUNTRY])
	assert.NotNil(t, propertiesMap["prop_1"])

	propertiesMap = U.PropertiesMap{"prop_1": "value_1"}
	err = M.FillUserDefaultProperties(&propertiesMap, "127.0.0.1")
	assert.Nil(t, err)
	assert.NotEmpty(t, propertiesMap[U.UP_INTERNAL_IP])
	assert.Empty(t, propertiesMap[U.UP_COUNTRY])

	propertiesMap = U.PropertiesMap{"prop_1": "value_1"}
	err = M.FillUserDefaultProperties(&propertiesMap, "::1")
	assert.Nil(t, err)
	assert.NotEmpty(t, propertiesMap[U.UP_INTERNAL_IP])
	assert.Empty(t, propertiesMap[U.UP_COUNTRY])
}
