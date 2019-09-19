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

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	projectId := project.ID

	start := time.Now()

	// Test successful create user.
	user, errCode := M.CreateUser(&M.User{ProjectId: projectId})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.True(t, len(user.ID) > 30)
	assert.Equal(t, projectId, user.ProjectId)
	assert.True(t, user.JoinTimestamp >= start.Unix()-60)
	assert.InDelta(t, user.JoinTimestamp, start.Unix()-60, 3)
	assert.True(t, user.CreatedAt.After(start))
	assert.True(t, user.UpdatedAt.After(start))
	// Not more than 20ms difference.
	assert.InDelta(t, user.CreatedAt.UnixNano(), user.UpdatedAt.UnixNano(), 2.0e+7)
	// Test Get User on the created one.
	retUser, errCode := M.GetUser(projectId, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, user.JoinTimestamp >= start.Unix()-60)
	assert.InDelta(t, user.JoinTimestamp, start.Unix()-60, 3)
	assert.True(t, math.Abs(user.CreatedAt.Sub(retUser.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(user.UpdatedAt.Sub(retUser.UpdatedAt).Seconds()) < 0.1)
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	var userProperties, retUserProperties map[string]interface{}
	json.Unmarshal(user.Properties.RawMessage, &userProperties)
	json.Unmarshal(retUser.Properties.RawMessage, &retUserProperties)
	assert.Equal(t, 1, len(userProperties))
	assert.Equal(t, 1, len(retUserProperties))
	// nil gets changed to null.
	// A row in user_properties is created even when properties is nil.
	user.Properties = postgres.Jsonb{RawMessage: json.RawMessage([]byte(`null`))}
	retUser.CreatedAt = time.Time{}
	retUser.UpdatedAt = time.Time{}
	assert.Equal(t, user.ProjectId, retUser.ProjectId)
	// id of null user_properties row. updated as user_properties_id.
	assert.True(t, len(retUser.PropertiesId) > 0)
	// Test Get User with wrong project id.
	retUser, errCode = M.GetUser(projectId+1, user.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retUser)

	// Test successful create user with customer_user_id and properties.
	customerUserId := "customer_id"
	properties := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "india", "age": 30, "paid": true}`))}
	user, errCode = M.CreateUser(&M.User{ProjectId: projectId, CustomerUserId: customerUserId, Properties: properties})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, customerUserId, user.CustomerUserId)
	assert.True(t, len(user.ID) > 30)
	assert.Equal(t, projectId, user.ProjectId)
	assert.True(t, user.JoinTimestamp >= start.Unix()-60)
	assert.InDelta(t, user.JoinTimestamp, start.Unix()-60, 3)
	assert.True(t, user.CreatedAt.After(start))
	assert.True(t, user.UpdatedAt.After(start))
	// Not more than 20ms difference.
	assert.InDelta(t, user.CreatedAt.UnixNano(), user.UpdatedAt.UnixNano(), 5.0e+7)
	var retProperties map[string]interface{}
	err = json.Unmarshal(user.Properties.RawMessage, &retProperties)
	assert.Nil(t, err)
	assert.Contains(t, retProperties, "country")
	assert.Contains(t, retProperties, "age")
	assert.Contains(t, retProperties, "paid")
	assert.Contains(t, retProperties, U.UP_JOIN_TIME)

	// Creating again with the same customer_user_id with no properties.
	// Should respond with last user of customer_user instead of creating.
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	newUser, newUserErrorCode := M.CreateUser(&M.User{ProjectId: projectId, CustomerUserId: rCustomerUserId})
	assert.Equal(t, http.StatusCreated, newUserErrorCode)
	lastUser, lastUserErrorCode := M.GetUserLatestByCustomerUserId(projectId, rCustomerUserId)
	assert.Equal(t, http.StatusFound, lastUserErrorCode)
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
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

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
		assert.Equal(t, http.StatusCreated, errCode)
		assert.True(t, len(user.ID) > 30)
		users = append(users, *user)
	}

	retUsers, errCode = M.GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, limit, uint64(len(retUsers)))
	assertUsersWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	retUsers, errCode = M.GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, limit, uint64(len(retUsers)))
	assertUsersWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	retUsers, errCode = M.GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusFound, errCode)
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

		assert.Equal(t, expectedUser.ProjectId, actualUser.ProjectId)
		assert.Equal(t, expectedUser.ID, actualUser.ID)
		assert.Equal(t, expectedUser.CustomerUserId, actualUser.CustomerUserId)
		assert.Equal(t, expectedUser.PropertiesId, actualUser.PropertiesId)
		assert.Equal(t, expectedUser.SegmentAnonymousId, actualUser.SegmentAnonymousId)
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
	assert.Equal(t, http.StatusCreated, latestUserErrCode)
	getUser, getUserErrCode := M.GetUserLatestByCustomerUserId(project.ID, rCustomerUserId)
	assert.Equal(t, http.StatusFound, getUserErrCode)
	assert.Equal(t, latestUser.ID, getUser.ID)

	// Bad input. // Without project scope.
	_, errCode := M.GetUserLatestByCustomerUserId(0, rCustomerUserId)
	assert.NotEqual(t, http.StatusFound, errCode)

	// Bad input. // Unacceptable customer_user_id
	_, errCode = M.GetUserLatestByCustomerUserId(project.ID, " ")
	assert.NotEqual(t, http.StatusFound, errCode)

	sameCustomerId := "user_1"
	user1, errCode := M.CreateUser(&M.User{ProjectId: project.ID, CustomerUserId: sameCustomerId})
	assert.NotNil(t, user1)
	user2, errCode := M.CreateUser(&M.User{ProjectId: project.ID, CustomerUserId: sameCustomerId})
	assert.NotNil(t, user2)
	user3, errCode := M.CreateUser(&M.User{ProjectId: project.ID, CustomerUserId: sameCustomerId})
	assert.NotNil(t, user3)
	luser, errCode := M.GetUserLatestByCustomerUserId(project.ID, sameCustomerId)
	assert.Equal(t, user3.ID, luser.ID) // Should be the latest user with same customer_user_id.
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
	assert.Equal(t, http.StatusAccepted, errCode)
	assert.Equal(t, rCustomerUserId, cuUpdatedUser.CustomerUserId)
	// Using already tested GetUser method to validate update.
	gUser, gErrCode := M.GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, gErrCode)
	// Test CustomerUserId updated or not.
	assert.Equal(t, rCustomerUserId, gUser.CustomerUserId)
	// Update user should not create properties while updating
	// other fields (identify).
	assert.Equal(t, user.PropertiesId, gUser.PropertiesId)

	segAid := "seg_aid_1"
	_, errCode = M.UpdateUser(project.ID, user.ID, &M.User{SegmentAnonymousId: segAid})
	assert.Equal(t, http.StatusAccepted, errCode)
	gUser, gErrCode = M.GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, gErrCode)
	assert.Equal(t, segAid, gUser.SegmentAnonymousId)
	assert.Equal(t, rCustomerUserId, gUser.CustomerUserId) // Should not update cuid.

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
	assert.NotEqual(t, http.StatusAccepted, errCode)

	// Bad input. UserId.
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	_, errCode = M.UpdateUser(project.ID, "", &M.User{})
	assert.NotEqual(t, http.StatusAccepted, errCode)
}

func TestDBUpdateUserProperties(t *testing.T) {
	// Intialize.
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	assert.True(t, len(user.PropertiesId) > 0)

	// No change on empty json
	newProperties := &postgres.Jsonb{}
	var oldPropertiesId, newPropertiesId string
	newPropertiesId, status := M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusNotModified, status)

	oldPropertiesId = newPropertiesId
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "india", "age": 30.1, "paid": true}`))}
	newPropertiesId, status = M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldPropertiesId, newPropertiesId)

	oldPropertiesId = newPropertiesId
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "india", "age": 30.1, "paid": true}`))}
	newPropertiesId, status = M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusNotModified, status)
	assert.Equal(t, oldPropertiesId, newPropertiesId)

	oldPropertiesId = newPropertiesId
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"age": 30.1, "paid": true, "country": "usa"}`))}
	newPropertiesId, status = M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldPropertiesId, newPropertiesId)

	oldPropertiesId = newPropertiesId
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"device": "android"}`))}
	newPropertiesId, status = M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldPropertiesId, newPropertiesId)

	oldPropertiesId = newPropertiesId
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"age": 30.1, "country": "usa", "device": "android", "paid": true}`))}
	newPropertiesId, status = M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusNotModified, status)
	assert.Equal(t, oldPropertiesId, newPropertiesId)

	// New key should merge with existing keys.
	oldPropertiesId = newPropertiesId
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"prop1": "value1"}`))}
	newPropertiesId, status = M.UpdateUserProperties(project.ID, user.ID, newProperties)
	assert.Equal(t, http.StatusAccepted, status)
	properties, status := M.GetUserProperties(project.ID, user.ID, newPropertiesId)
	var propertiesMap map[string]interface{}
	err = json.Unmarshal((*properties).RawMessage, &propertiesMap)
	assert.Nil(t, err)
	assert.Len(t, propertiesMap, 6) // including joinTime.
	assert.Equal(t, "value1", propertiesMap["prop1"])
}

func TestDBFillUserDefaultProperties(t *testing.T) {
	propertiesMap := U.PropertiesMap{"prop_1": "value_1"}
	err := M.FillLocationUserProperties(&propertiesMap, "180.151.36.234") // Our gateway IP.
	assert.Nil(t, err)
	// IP is not stored in user properties.
	assert.Empty(t, propertiesMap[U.EP_INTERNAL_IP])
	assert.NotNil(t, propertiesMap[U.UP_COUNTRY])
	assert.Equal(t, "India", propertiesMap[U.UP_COUNTRY])
	assert.Equal(t, "Bengaluru", propertiesMap[U.UP_CITY])
	assert.NotNil(t, propertiesMap["prop_1"]) // Should append to existing values.

	propertiesMap = U.PropertiesMap{"prop_1": "value_1"}
	err = M.FillLocationUserProperties(&propertiesMap, "127.0.0.1")
	assert.Nil(t, err)
	assert.Empty(t, propertiesMap[U.EP_INTERNAL_IP])

	propertiesMap = U.PropertiesMap{"prop_1": "value_1"}
	err = M.FillLocationUserProperties(&propertiesMap, "::1")
	assert.Nil(t, err)
	assert.Empty(t, propertiesMap[U.EP_INTERNAL_IP])
}

func TestDBGetSegmentUser(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	user, errCode := M.GetSegmentUser(project.ID, "", "customer_1")
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, user)

	// no customer uid. create new user with seg_aid.
	segAid := U.RandomLowerAphaNumString(15)
	user1, errCode := M.GetSegmentUser(project.ID, segAid, "")
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, user1)
	assert.Equal(t, segAid, user1.SegmentAnonymousId)
	assert.Empty(t, user1.CustomerUserId)

	// exist return same user. using same segAid.
	user2, errCode := M.GetSegmentUser(project.ID, segAid, "")
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user2)
	assert.Equal(t, user1.ID, user2.ID)
	assert.Equal(t, user1.SegmentAnonymousId, user2.SegmentAnonymousId)
	assert.Empty(t, user2.CustomerUserId)

	// both provided. c_uid is empty. identify
	custId := U.RandomLowerAphaNumString(15)
	user3, errCode := M.GetSegmentUser(project.ID, segAid, custId)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user3)
	assert.Equal(t, user1.ID, user3.ID)
	assert.NotEmpty(t, user3.CustomerUserId)
	assert.Equal(t, custId, user3.CustomerUserId) // Update c_uid on existing user.

	// both seg_aid and c_uid matches.
	user4, errCode := M.GetSegmentUser(project.ID, segAid, user3.CustomerUserId)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user4)
	assert.Equal(t, user3.ID, user4.ID)

	// c_uid mismatch with existing seg_aid. should not update c_uid.
	custId1 := U.RandomLowerAphaNumString(15)
	user5, errCode := M.GetSegmentUser(project.ID, segAid, custId1)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user5)
	assert.Equal(t, user4.ID, user5.ID)                         // Should return existing user.
	assert.Equal(t, user4.CustomerUserId, user5.CustomerUserId) // Should not be updated.

	// user by seg_aid doesn't exist, but user exist with given c_uid.
	segAid1 := U.RandomLowerAphaNumString(15)
	user6, errCode := M.GetSegmentUser(project.ID, segAid1, user4.CustomerUserId) // new seg_aid.
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user6)
	assert.Equal(t, user4.ID, user6.ID) // Should not use existing user with same c_uid.

	// user by seg_aid and c_uid doesn't exist.
	custId2 := U.RandomLowerAphaNumString(15)
	segAid2 := U.RandomLowerAphaNumString(15)
	user7, errCode := M.GetSegmentUser(project.ID, segAid2, custId2)
	assert.Equal(t, http.StatusCreated, errCode)
	// new user with new seg_aid and c_uid.
	assert.Equal(t, segAid2, user7.SegmentAnonymousId)
	assert.Equal(t, custId2, user7.CustomerUserId)
}

func TestGetRecentUserPropertyKeys(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	props1 := json.RawMessage(`{"prop1": "value1", "prop2": 1}`)
	_, errCode1 := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{props1}})
	assert.Equal(t, http.StatusCreated, errCode1)
	props2 := json.RawMessage(`{"prop3": "value2", "prop4": 2}`)
	_, errCode2 := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{props2}})
	assert.Equal(t, http.StatusCreated, errCode2)

	// recent users limited to 1.
	props, errCode := M.GetRecentUserPropertyKeysWithLimits(project.ID, 1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Contains(t, props, U.PropertyTypeCategorical)
	assert.Contains(t, props, U.PropertyTypeNumerical)
	assert.Len(t, props[U.PropertyTypeCategorical], 1)
	assert.Len(t, props[U.PropertyTypeNumerical], 2) // including joinTime.
	// validates classification.
	assert.Contains(t, props[U.PropertyTypeCategorical], "prop3")
	assert.Contains(t, props[U.PropertyTypeNumerical], "prop4")
	// old user properties shoult not exist.
	assert.NotContains(t, props[U.PropertyTypeCategorical], "prop1")
	assert.NotContains(t, props[U.PropertyTypeNumerical], "prop2")
}

func TestGetRecentUserPropertyValues(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	props1 := json.RawMessage(`{"prop3": "value1", "prop4": "1"}`)
	_, errCode1 := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{props1}})
	assert.Equal(t, http.StatusCreated, errCode1)
	props2 := json.RawMessage(`{"prop3": "value2", "prop4": "2"}`)
	_, errCode2 := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{props2}})
	assert.Equal(t, http.StatusCreated, errCode2)
	// different user with same properties as previous and different values.
	props3 := json.RawMessage(`{"prop3": "value3", "prop4": "3"}`)
	_, errCode3 := M.CreateUser(&M.User{ProjectId: project.ID, Properties: postgres.Jsonb{props3}})
	assert.Equal(t, http.StatusCreated, errCode3)

	t.Run("RecentPropertyValuesLimitedByUsers", func(t *testing.T) {
		// recent users limited to 2.
		values, errCode := M.GetRecentUserPropertyValuesWithLimits(project.ID, "prop3", 2, 100)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, values, 2)
		assert.NotContains(t, values, "value1")
	})

	t.Run("RecentPropertyValuesLimitedByValues", func(t *testing.T) {
		// recent users limited to 3 but values limited to 2.
		values, errCode := M.GetRecentUserPropertyValuesWithLimits(project.ID, "prop4", 3, 2)
		assert.Equal(t, http.StatusFound, errCode)
		assert.Len(t, values, 2)
		assert.NotContains(t, values, "3")
	})
}
