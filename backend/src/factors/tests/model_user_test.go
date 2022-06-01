package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/integration/clear_bit"
	"factors/model/model"
	"factors/model/store"
	"factors/sdk"
	SDK "factors/sdk"
	"factors/util"
	U "factors/util"
	"fmt"
	"github.com/clearbit/clearbit-go/clearbit"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
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
	createUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	user, errCode := store.GetStore().GetUser(projectId, createUserID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.True(t, len(user.ID) > 30)
	assert.Equal(t, projectId, user.ProjectId)
	assert.True(t, user.JoinTimestamp >= start.Unix()-60)
	assert.InDelta(t, user.JoinTimestamp, start.Unix()-60, 3)
	assert.True(t, user.CreatedAt.After(start))
	assert.True(t, user.UpdatedAt.After(start))
	// Not more than 20ms difference.
	assert.InDelta(t, user.CreatedAt.UnixNano(), user.UpdatedAt.UnixNano(), 2.0e+7)
	// Test Get User on the created one.
	retUser, errCode := store.GetStore().GetUser(projectId, user.ID)
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
	assert.NotEmpty(t, retUser.Properties)
	// Test Get User with wrong project id.
	retUser, errCode = store.GetStore().GetUser(projectId+1, user.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retUser)

	// Test successful create user with customer_user_id and properties.
	customerUserId := "customer_id"
	properties := postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"country": "india", "age": 30, "paid": true}`))}
	createUserID, errCode = store.GetStore().CreateUser(&model.User{ProjectId: projectId, CustomerUserId: customerUserId, Properties: properties, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	user, errCode = store.GetStore().GetUser(projectId, createUserID)
	assert.Equal(t, http.StatusFound, errCode)
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
	createUserID, newUserErrorCode := store.GetStore().CreateUser(&model.User{ProjectId: projectId, CustomerUserId: rCustomerUserId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, newUserErrorCode)
	lastUser, lastUserErrorCode := store.GetStore().GetUserLatestByCustomerUserId(projectId, rCustomerUserId, model.UserSourceWeb)
	assert.Equal(t, http.StatusFound, lastUserErrorCode)
	assert.Equal(t, createUserID, lastUser.ID)

	// Test Get User on random id.
	randomId := U.RandomLowerAphaNumString(15)
	retUser, errCode = store.GetStore().GetUser(projectId, randomId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retUser)

	// Test external UUID as id.
	uuid := U.GetUUID()
	createUserID, errCode = store.GetStore().CreateUser(&model.User{ID: uuid, ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	// User should be create with given id.
	assert.Equal(t, uuid, createUserID)

	// Use an existing user_id to create. Should get and return the user.
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ID: uuid, ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, createUserID, createdUserID2)
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
	retUsers, errCode := store.GetStore().GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusNotFound, errCode)

	// Create 100 users.
	var users []model.User
	numUsers := 100
	for i := 0; i < numUsers; i++ {
		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		user, errCode := store.GetStore().GetUser(projectId, createdUserID)
		assert.Equal(t, http.StatusFound, errCode)
		assert.True(t, len(user.ID) > 30)
		users = append(users, *user)
	}

	retUsers, errCode = store.GetStore().GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, limit, uint64(len(retUsers)))
	assertUsersWithOffset(t, users[offset:offset+limit], retUsers)

	offset = 25
	limit = 20
	retUsers, errCode = store.GetStore().GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, limit, uint64(len(retUsers)))
	assertUsersWithOffset(t, users[offset:offset+limit], retUsers)

	// Overflow
	offset = 95
	limit = 10
	retUsers, errCode = store.GetStore().GetUsers(projectId, offset, limit)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, numUsers-95, len(retUsers))
	assertUsersWithOffset(t, users[offset:numUsers], retUsers)
}

func assertUsersWithOffset(t *testing.T, expectedUsers []model.User, actualUsers []model.User) {
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
	createUserID1, latestUserErrCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: rCustomerUserId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, latestUserErrCode)
	getUser, getUserErrCode := store.GetStore().GetUserLatestByCustomerUserId(project.ID, rCustomerUserId, model.UserSourceWeb)
	assert.Equal(t, http.StatusFound, getUserErrCode)
	assert.Equal(t, createUserID1, getUser.ID)

	// Bad input. // Without project scope.
	_, errCode := store.GetStore().GetUserLatestByCustomerUserId(0, rCustomerUserId, model.UserSourceWeb)
	assert.NotEqual(t, http.StatusFound, errCode)

	// Bad input. // Unacceptable customer_user_id
	_, errCode = store.GetStore().GetUserLatestByCustomerUserId(project.ID, " ", model.UserSourceWeb)
	assert.NotEqual(t, http.StatusFound, errCode)

	sameCustomerId := "user_1"
	createUserID1, errCode = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: sameCustomerId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createUserID1)
	createUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: sameCustomerId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createUserID2)
	createUserID3, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: sameCustomerId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.NotEmpty(t, createUserID3)
	luser, errCode := store.GetStore().GetUserLatestByCustomerUserId(project.ID, sameCustomerId, model.UserSourceWeb)
	assert.Equal(t, createUserID3, luser.ID) // Should be the latest user with same customer_user_id.
}

func TestDBUpdateUserById(t *testing.T) {
	// Intialize.
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)

	// Test updating a field.
	rCustomerUserId := U.RandomLowerAphaNumString(15)
	updateUser := &model.User{CustomerUserId: rCustomerUserId}
	_, errCode := store.GetStore().UpdateUser(project.ID, user.ID,
		updateUser, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)
	// Using already tested GetUser method to validate update.
	gUser, gErrCode := store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, gErrCode)
	// Test CustomerUserId updated or not.
	assert.Equal(t, rCustomerUserId, gUser.CustomerUserId)
	// Update user should not create properties while updating
	// other fields (identify).
	assert.Equal(t, DecodePostgresJsonbWithoutError(&gUser.Properties),
		DecodePostgresJsonbWithoutError(&user.Properties))

	segAid := "seg_aid_1"
	_, errCode = store.GetStore().UpdateUser(project.ID, user.ID, &model.User{SegmentAnonymousId: segAid,
		Properties: postgres.Jsonb{json.RawMessage(`{"key": "value"}`)}}, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)
	gUser, gErrCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, gErrCode)
	assert.Equal(t, segAid, gUser.SegmentAnonymousId)
	assert.Equal(t, rCustomerUserId, gUser.CustomerUserId) // Should not update cuid.

	// Test overwriting of user's properties with empty when not given.
	segAid = "seg_aid_2"
	_, errCode = store.GetStore().UpdateUser(project.ID, user.ID, &model.User{SegmentAnonymousId: segAid}, time.Now().Unix()+1)
	assert.Equal(t, http.StatusAccepted, errCode)
	gUser, gErrCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.NotEmpty(t, gUser.Properties)
	propertiesMap, err := U.DecodePostgresJsonb(&gUser.Properties)
	assert.Nil(t, err)
	assert.Equal(t, (*propertiesMap)["key"], "value")

	// Test updating ProjectId with other fields
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	uProject, uErr := SetupProjectReturnDAO()
	assert.Nil(t, uErr)
	assert.NotNil(t, uProject)
	updateUser = &model.User{ProjectId: uProject.ID, CustomerUserId: rCustomerUserId}
	_, errCode = store.GetStore().UpdateUser(project.ID, user.ID, updateUser, time.Now().Unix())
	assert.Equal(t, http.StatusBadRequest, errCode)

	// Bad input. ProjectId.
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	_, errCode = store.GetStore().UpdateUser(0, user.ID, &model.User{}, time.Now().Unix())
	assert.NotEqual(t, http.StatusAccepted, errCode)

	// Bad input. UserId.
	rCustomerUserId = U.RandomLowerAphaNumString(15)
	_, errCode = store.GetStore().UpdateUser(project.ID, "", &model.User{}, time.Now().Unix())
	assert.NotEqual(t, http.StatusAccepted, errCode)
}

func TestDBUpdateUserProperties(t *testing.T) {
	// Intialize.
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Properties)

	// No change on empty json
	newProperties := &postgres.Jsonb{}
	oldUpdatedProperties, status := store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, http.StatusNotModified, status)

	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "india", "age": 30.1, "paid": true, "$hubspot_contact_lead_guid": "lead-guid1"}`))}
	newUpdatedProperties, status := store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldUpdatedProperties, newUpdatedProperties)
	newUpdatedPropertiesMap, err := U.DecodePostgresJsonb(newUpdatedProperties)
	assert.Nil(t, err)
	assert.Equal(t, "india", (*newUpdatedPropertiesMap)["country"])
	assert.Equal(t, "lead-guid1", (*newUpdatedPropertiesMap)[model.UserPropertyHubspotContactLeadGUID])

	oldUpdatedProperties = newUpdatedProperties
	// do not allow overwrite existing user properties from past timestamp.
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "US", "age": 30.1, "paid": true}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix()-60)
	assert.Equal(t, status, http.StatusAccepted)
	assert.Equal(t, oldUpdatedProperties, newUpdatedProperties)

	oldUpdatedProperties = newUpdatedProperties
	// allow adding new keys from past timestamp.
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "US", "age": 30.1, "paid": true, "past": true}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix()-60)
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldUpdatedProperties, newUpdatedProperties)

	oldUpdatedProperties = newUpdatedProperties
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "india", "age": 30.1, "paid": true}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, status, http.StatusAccepted)
	assert.Equal(t, oldUpdatedProperties, newUpdatedProperties)

	oldUpdatedProperties = newUpdatedProperties
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"age": 30.1, "paid": true, "country": "usa"}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldUpdatedProperties, newUpdatedProperties)

	oldUpdatedProperties = newUpdatedProperties
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"device": "android", "$hubspot_contact_lead_guid": "lead-guid2"}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, status)
	assert.NotEqual(t, oldUpdatedProperties, newUpdatedProperties)
	newUpdatedPropertiesMap, err = U.DecodePostgresJsonb(newUpdatedProperties)
	assert.Nil(t, err)
	assert.Equal(t, "usa", (*newUpdatedPropertiesMap)["country"])
	assert.Equal(t, "android", (*newUpdatedPropertiesMap)["device"])
	// Property should be skipped on merge. Should be same as earlier.
	assert.Equal(t, "lead-guid2", (*newUpdatedPropertiesMap)[model.UserPropertyHubspotContactLeadGUID])

	oldUpdatedProperties = newUpdatedProperties
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"age": 30.1, "country": "usa", "device": "android", "paid": true}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, status, http.StatusAccepted)
	assert.Equal(t, oldUpdatedProperties, newUpdatedProperties)

	// New key should merge with existing keys.
	oldUpdatedProperties = newUpdatedProperties
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"prop1": "value1"}`))}
	newUpdatedProperties, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, status)
	user, status = store.GetStore().GetUser(project.ID, user.ID)
	var propertiesMap map[string]interface{}
	err = json.Unmarshal((user.Properties).RawMessage, &propertiesMap)
	assert.Nil(t, err)
	assert.Len(t, propertiesMap, 8) // including joinTime.
	assert.Equal(t, "value1", propertiesMap["prop1"])
}

func TestPropertiesUpdatedTimestamp(t *testing.T) {
	// Intialize the project and the user. Also capture old timestamp in old_time.
	oldTimestamp := time.Now().Unix() - 1000
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Properties)
	storedUser, errCode := store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	storedTimestamp := storedUser.PropertiesUpdatedTimestamp

	// Update user properties using the older timestamp. The PropertiesUpdatedTimestamp
	// should not get updated.
	newProperties := &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"country": "india", "age": 30.1, "paid": true, "$hubspot_contact_lead_guid": "lead-guid1"}`))}
	_, status := store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, oldTimestamp)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.NotEqual(t, oldTimestamp, storedUser.PropertiesUpdatedTimestamp)
	assert.Equal(t, storedTimestamp, storedUser.PropertiesUpdatedTimestamp)

	// Update user properties using the current timestamp. The PropertiesUpdatedTimestamp
	// should get updated with the current timestamp.
	current_time := time.Now().Unix()
	newProperties = &postgres.Jsonb{RawMessage: json.RawMessage([]byte(
		`{"device": "android"}`))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user.ID,
		newProperties, current_time)
	assert.Equal(t, http.StatusAccepted, status)
	storedUser, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, current_time, storedUser.PropertiesUpdatedTimestamp)
}

func TestDBFillUserDefaultProperties(t *testing.T) {
	propertiesMap := U.PropertiesMap{"prop_1": "value_1"}
	err := model.FillLocationUserProperties(&propertiesMap, "180.151.36.234") // Our gateway IP.
	assert.Nil(t, err)
	// IP is not stored in user properties.
	assert.Empty(t, propertiesMap[U.EP_INTERNAL_IP])
	assert.NotNil(t, propertiesMap[U.UP_COUNTRY])
	assert.Equal(t, "India", propertiesMap[U.UP_COUNTRY])
	assert.Equal(t, "Bengaluru", propertiesMap[U.UP_CITY])
	assert.NotNil(t, propertiesMap[U.UP_CONTINENT])
	assert.Equal(t, "Asia", propertiesMap[U.UP_CONTINENT])
	assert.NotNil(t, propertiesMap[U.UP_POSTAL_CODE])
	assert.Equal(t, "560076", propertiesMap[U.UP_POSTAL_CODE])
	assert.NotNil(t, propertiesMap["prop_1"]) // Should append to existing values.

	propertiesMap = U.PropertiesMap{"prop_1": "value_1"}
	err = model.FillLocationUserProperties(&propertiesMap, "127.0.0.1")
	assert.Nil(t, err)
	assert.Empty(t, propertiesMap[U.EP_INTERNAL_IP])

	propertiesMap = U.PropertiesMap{"prop_1": "value_1"}
	err = model.FillLocationUserProperties(&propertiesMap, "::1")
	assert.Nil(t, err)
	assert.Empty(t, propertiesMap[U.EP_INTERNAL_IP])
}
func clearbitAnalysisTestDBClearbit(t *testing.T) {

	// Ip= "89.76.236.199"
	projectId := uint64(5)
	clientIP := "89.76.236.199"
	propertiesMap1 := U.PropertiesMap{"prop_1": "value_1"}
	executeClearBitStatusChannel := make(chan int)

	clearbitKey, errCode := store.GetStore().GetClearbitKeyFromProjectSetting(projectId)
	if errCode != http.StatusFound {
		log.Info("Get clear_bit key failed.")
	}
	go clear_bit.ExecuteClearBitEnrich(clearbitKey, &propertiesMap1, clientIP, executeClearBitStatusChannel) // Our gateway IP.
	select {
	case response, ok := <-executeClearBitStatusChannel:
		if ok {
			fmt.Println(response)
		}
	case <-time.After(300 * time.Second):
		fmt.Println("time Out :(")

	}
	// IP is not stored in user properties.
	assert.NotNil(t, propertiesMap1[U.CLR_COMPANY_GEO_COUNTRY])
	assert.Equal(t, "Poland", propertiesMap1[U.CLR_COMPANY_GEO_COUNTRY])
	assert.Equal(t, "Wrocław", propertiesMap1[U.CLR_COMPANY_GEO_CITY])
	assert.Equal(t, "50-203", propertiesMap1[U.CLR_COMPANY_GEO_POSTALCODE])
	assert.Equal(t, "89.76.236.199", propertiesMap1[U.CLR_IP])
	assert.Equal(t, "divante.pl", propertiesMap1[U.CLR_COMPANY_PARENT_DOMAIN])

	assert.NotNil(t, propertiesMap1["prop_1"]) // Should append to existing values.
}
func TestDBCreateOrGetSegmentUserWithSDKIdentify(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// No seg_aid but c_uid provided. should create new user without c_uid.
	// Later user will be identified with SDK.Identify.
	customerUserID := U.RandomLowerAphaNumString(15) + "@example.com"
	user, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, "", customerUserID, time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.CustomerUserId)
	status, _ := SDK.Identify(project.ID, &SDK.IdentifyPayload{UserId: user.ID, CustomerUserId: customerUserID, RequestSource: model.UserSourceWeb}, false)
	assert.Equal(t, http.StatusOK, status)
	user, errCode = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, customerUserID, user.CustomerUserId)
	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, user.CustomerUserId, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, user.CustomerUserId, (*userProperties)[U.UP_EMAIL])

	// no customer uid. create new user with seg_aid.
	segAid := U.RandomLowerAphaNumString(15)
	user1, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, "", time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusCreated, errCode)
	assert.NotNil(t, user1)
	assert.Equal(t, segAid, user1.SegmentAnonymousId)
	assert.Empty(t, user1.CustomerUserId)

	// exist return same user. using same segAid.
	user2, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, "", time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user2)
	assert.Equal(t, user1.ID, user2.ID)
	assert.Equal(t, user1.SegmentAnonymousId, user2.SegmentAnonymousId)
	assert.Empty(t, user2.CustomerUserId)

	// both provided. c_uid is empty. identify
	custId := U.RandomLowerAphaNumString(15) + "@example.com"
	user3, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, custId, time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user3)
	assert.Equal(t, user1.ID, user3.ID)
	status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{UserId: user3.ID, CustomerUserId: custId, RequestSource: model.UserSourceWeb}, false)
	assert.Equal(t, http.StatusOK, status)
	user3, errCode = store.GetStore().GetUser(project.ID, user3.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, custId, user3.CustomerUserId)
	userProperties, err = U.DecodePostgresJsonb(&user3.Properties)
	assert.Nil(t, err)
	assert.Equal(t, user3.CustomerUserId, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, user3.CustomerUserId, (*userProperties)[U.UP_EMAIL])

	// both seg_aid and c_uid matches.
	user4, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, user3.CustomerUserId, time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user4)
	assert.Equal(t, user3.ID, user4.ID)

	// c_uid mismatch with existing seg_aid. should not update c_uid.
	custId1 := U.RandomLowerAphaNumString(15)
	user5, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, custId1, time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user5)
	assert.Equal(t, user4.ID, user5.ID)                         // Should return existing user.
	assert.Equal(t, user4.CustomerUserId, user5.CustomerUserId) // Should not be updated.

	// user by seg_aid doesn't exist, but user exist with given c_uid.
	segAid1 := U.RandomLowerAphaNumString(15)
	user6, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid1, user4.CustomerUserId, time.Now().Unix(), model.UserSourceWeb) // new seg_aid.
	assert.Equal(t, http.StatusOK, errCode)
	assert.NotNil(t, user6)
	assert.Equal(t, user4.ID, user6.ID) // Should not use existing user with same c_uid.

	// user by seg_aid and c_uid doesn't exist.
	custId2 := U.RandomLowerAphaNumString(15) + "@example.com"
	segAid2 := U.RandomLowerAphaNumString(15)
	user7, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid2, custId2, time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusCreated, errCode)
	// new user with new seg_aid and identified with cuid.
	assert.Equal(t, segAid2, user7.SegmentAnonymousId)
	status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{UserId: user7.ID, CustomerUserId: custId2, RequestSource: model.UserSourceWeb}, false)
	assert.Equal(t, http.StatusOK, status)
	user7, errCode = store.GetStore().GetUser(project.ID, user7.ID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, custId2, user7.CustomerUserId)
	userProperties, err = U.DecodePostgresJsonb(&user7.Properties)
	assert.Nil(t, err)
	assert.Equal(t, user7.CustomerUserId, (*userProperties)[U.UP_USER_ID])
	assert.Equal(t, user7.CustomerUserId, (*userProperties)[U.UP_EMAIL])
	assert.Equal(t, custId2, user7.CustomerUserId)
}

func TestGetRecentUserPropertyKeys(t *testing.T) {
	r := gin.Default()
	project, err := SetupProjectReturnDAO()
	H.InitSDKServiceRoutes(r)
	assert.Nil(t, err)

	// Test successful CreateEvent.
	props1 := json.RawMessage(`{"prop1": "value1", "prop2": 1}`)
	props2 := json.RawMessage(`{"prop3": "value2", "prop4": 2}`)
	createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{props1}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	createdUserID2, errCode2 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{props2}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode1)
	assert.Equal(t, http.StatusCreated, errCode2)

	uri := "/sdk/event/track"

	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"prop1": "value1", "prop2": 1}}`, createdUserID1, "e1")),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"prop3": "value2", "prop4": 2}}`, createdUserID2, "e2")),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	props, err := store.GetStore().GetRecentUserPropertyKeysWithLimits(project.ID, 10, 100, time.Now().UTC())
	assert.Equal(t, nil, err)
	propertyMap := make(map[string]bool)
	for _, property := range props {
		propertyMap[property.Key] = true
	}
	assert.Equal(t, propertyMap["prop1"], true)
	assert.Equal(t, propertyMap["prop2"], true)
	assert.Equal(t, propertyMap["prop3"], true)
	assert.Equal(t, propertyMap["prop4"], true)

	// recent users limited to 1.
	props, err = store.GetStore().GetRecentUserPropertyKeysWithLimits(project.ID, 1, 100, time.Now().UTC())
	assert.Equal(t, nil, err)
	propertyMap = make(map[string]bool)
	for _, property := range props {
		propertyMap[property.Key] = true
	}
	user1Prop := propertyMap["prop1"] == true && propertyMap["prop2"] == true
	user2Prop := propertyMap["prop3"] == true && propertyMap["prop4"] == true
	assert.Equal(t, user1Prop || user2Prop, true)
	assert.Equal(t, user1Prop && user2Prop, false)

}

func TestGetRecentUserPropertyValues(t *testing.T) {
	r := gin.Default()
	project, err := SetupProjectReturnDAO()
	H.InitSDKServiceRoutes(r)
	assert.Nil(t, err)

	// Test successful CreateEvent.
	props1 := json.RawMessage(`{"prop1": "value1", "prop2": 1}`)
	props2 := json.RawMessage(`{"prop3": "value2", "prop4": 2}`)
	createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{props1}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	createdUserID2, errCode2 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Properties: postgres.Jsonb{props2}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode1)
	assert.Equal(t, http.StatusCreated, errCode2)

	uri := "/sdk/event/track"

	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"prop1": "value1", "prop2": 1}}`, createdUserID1, "e1")),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	_ = ServePostRequestWithHeaders(r, uri,
		[]byte(fmt.Sprintf(`{"user_id": "%s",  "event_name": "%s", "auto": true, "event_properties": {"$dollar_property": "dollarValue", "$qp_search": "mobile", "mobile": "true", "$qp_encoded": "google%%20search", "$qp_utm_keyword": "google%%20search"}, "user_properties": {"prop3": "value2", "prop4": 2}}`, createdUserID2, "e2")),
		map[string]string{
			"Authorization": project.Token,
			"User-Agent":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
		})

	t.Run("RecentPropertyValuesLimitedByUsers", func(t *testing.T) {
		values, category, err := store.GetStore().GetRecentUserPropertyValuesWithLimits(project.ID, "prop4", 10, 100, time.Now().UTC())
		assert.Equal(t, nil, err)
		assert.Len(t, values, 1)
		valuesMap := make(map[string]bool)
		for _, value := range values {
			valuesMap[value.Value] = true
		}
		assert.Equal(t, valuesMap["2"], true)
		assert.Equal(t, category, U.PropertyTypeNumerical)
	})
}

func TestFillFormSubmitEventUserProperties(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	t.Run("UserWithoutProperties", func(t *testing.T) {
		createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode1)
		formSubmitProperties := U.PropertiesMap{
			U.UP_EMAIL: "xxx@example.com",
			U.UP_PHONE: "99999999999",
		}
		customerUserId, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(project.ID,
			createdUserID1, &formSubmitProperties)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "xxx@example.com", customerUserId)
		assert.Equal(t, "xxx@example.com", (*formSubmitUserProperties)[U.UP_EMAIL])
		assert.Equal(t, "99999999999", (*formSubmitUserProperties)[U.UP_PHONE])
	})

	t.Run("FormSubmitWithEmailAndUserPropertiesWithSameEmail", func(t *testing.T) {
		createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Properties: postgres.Jsonb{json.RawMessage(`{"$email": "xxx@example.com"}`)}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode1)
		formSubmitProperties := U.PropertiesMap{
			U.UP_EMAIL:   "xxx@example.com",
			U.UP_PHONE:   "99999999999",
			U.UP_COMPANY: "Example Inc",
		}
		customerUserId, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(project.ID,
			createdUserID1, &formSubmitProperties)
		assert.Equal(t, http.StatusOK, errCode)
		// Should add phone and other properties from
		// form submit as email matches.
		assert.Equal(t, "xxx@example.com", customerUserId)
		assert.Equal(t, "xxx@example.com", (*formSubmitUserProperties)[U.UP_EMAIL])
		assert.Equal(t, "99999999999", (*formSubmitUserProperties)[U.UP_PHONE])
		assert.Equal(t, "Example Inc", (*formSubmitUserProperties)[U.UP_COMPANY])
	})

	t.Run("FormSubmitWithEmailAndUserPropertiesWithDifferentEmail", func(t *testing.T) {
		createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Properties: postgres.Jsonb{json.RawMessage(`{"$email": "yyy@example.com"}`)}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode1)
		formSubmitProperties := U.PropertiesMap{
			U.UP_EMAIL:   "xxx@example.com",
			U.UP_PHONE:   "99999999999",
			U.UP_COMPANY: "Example Inc",
		}
		customerUserId, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(project.ID,
			createdUserID1, &formSubmitProperties)
		// free email overwrite will be avoided
		assert.Equal(t, http.StatusBadRequest, errCode)
		assert.Equal(t, "", customerUserId)
		assert.Nil(t, formSubmitUserProperties)
	})

	t.Run("FormSubmitWithPhoneAndUserPropertiesWithSamePhone", func(t *testing.T) {
		createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Properties: postgres.Jsonb{json.RawMessage(`{"$phone": "99999999999"}`)}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode1)
		formSubmitProperties := U.PropertiesMap{
			U.UP_EMAIL:   "xxx@example.com",
			U.UP_PHONE:   "99999999999",
			U.UP_COMPANY: "Example Inc",
		}
		customerUserId, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(project.ID,
			createdUserID1, &formSubmitProperties)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "xxx@example.com", customerUserId)
		// Should add all other properties from form submit as phone matches.
		assert.Equal(t, "xxx@example.com", (*formSubmitUserProperties)[U.UP_EMAIL])
		assert.Equal(t, "99999999999", (*formSubmitUserProperties)[U.UP_PHONE])
		assert.Equal(t, "Example Inc", (*formSubmitUserProperties)[U.UP_COMPANY])
	})

	t.Run("FormSubmitWithPhoneAndUserPropertiesWithDifferentPhone", func(t *testing.T) {
		createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID,
			Properties: postgres.Jsonb{json.RawMessage(`{"$phone": "99999999999"}`)}, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode1)
		formSubmitProperties := U.PropertiesMap{
			U.UP_EMAIL:   "xxx@example.com",
			U.UP_PHONE:   "8888888888",
			U.UP_COMPANY: "Example Inc",
		}
		customerUserId, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(project.ID,
			createdUserID1, &formSubmitProperties)
		assert.Equal(t, http.StatusOK, errCode)
		assert.Equal(t, "xxx@example.com", customerUserId)
		// Should add all other properties from form submit as phone matches.
		assert.NotNil(t, formSubmitUserProperties)
	})

	t.Run("FormSubmitWithCaseSensitiveEmailAndLeadingZeroPhoneNo", func(t *testing.T) {
		createdUserID1, errCode1 := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode1)
		formSubmitProperties := U.PropertiesMap{
			U.UP_EMAIL:   "Xyz@Example.com",
			U.UP_PHONE:   "0123456789",
			U.UP_COMPANY: "Example Inc",
		}

		customerUserId, formSubmitUserProperties, errCode := store.GetStore().GetCustomerUserIDAndUserPropertiesFromFormSubmit(project.ID,
			createdUserID1, &formSubmitProperties)
		assert.Equal(t, http.StatusOK, errCode)

		// email translated to lower case
		assert.Equal(t, "xyz@example.com", customerUserId)
		// Should add all other properties from form submit as phone matches.
		assert.Equal(t, "xyz@example.com", (*formSubmitUserProperties)[U.UP_EMAIL])
		// sanatized phone number
		assert.Equal(t, "123456789", (*formSubmitUserProperties)[U.UP_PHONE])
		assert.Equal(t, "Example Inc", (*formSubmitUserProperties)[U.UP_COMPANY])
	})

}

func TestGetUserPropertiesAsMap(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.Properties)

	userProperties, errCode := store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, user.ID)
	assert.Equal(t, errCode, http.StatusFound)
	decodedUserProperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, userProperties, decodedUserProperties)
}

func TestUserIdentityPropertiesOnCreateUser(t *testing.T) {
	project, _, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	cuid := "abcd@xyz.com"
	createdUserID1, status := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		CustomerUserId: cuid,
		Properties:     postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"city":"city1"}`))},
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
	})
	assert.Equal(t, http.StatusCreated, status)
	properties, status := store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, cuid, (*properties)[U.UP_EMAIL])
	assert.Equal(t, cuid, (*properties)[U.UP_USER_ID])
	assert.Equal(t, "city1", (*properties)["city"])

}

func TestIdentificationOrderPrecedence(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)
	timestamp := time.Now().Unix()

	phone := "1234567890"
	email1 := "ma@mail.com"
	// identification by phone
	properties := U.PropertiesMap{"$phone": phone}
	trackPayload := &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timestamp,
		EventProperties: properties,
		RequestSource:   model.UserSourceWeb,
	}

	status, response := SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.NotEmpty(t, response.UserId)
	assert.Equal(t, http.StatusOK, status)
	userId := response.UserId
	user, status = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, phone, user.CustomerUserId)
	propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, phone, (*propertiesMap)[U.UP_USER_ID])
	assert.Equal(t, phone, (*propertiesMap)[U.UP_PHONE])
	assert.Nil(t, (*propertiesMap)[U.UP_EMAIL])

	// adding email property
	timestamp = timestamp + 1
	properties = U.PropertiesMap{"$phone": phone, "$email": email1}
	trackPayload = &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timestamp,
		EventProperties: properties,
		UserId:          userId,
		RequestSource:   model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Equal(t, http.StatusOK, status)

	// email should be new customer_user_id
	user, status = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, email1, user.CustomerUserId)
	propertiesMap, err = U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, email1, (*propertiesMap)[U.UP_USER_ID])
	assert.Equal(t, phone, (*propertiesMap)[U.UP_PHONE])
	assert.Equal(t, email1, (*propertiesMap)[U.UP_EMAIL])

	// new email should overwrite
	email2 := "ma1@mail.com"
	timestamp = timestamp + 10
	properties = U.PropertiesMap{"$phone": phone, "$email": email2}
	trackPayload = &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timestamp,
		EventProperties: properties,
		UserId:          userId,
		RequestSource:   model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Equal(t, http.StatusOK, status)
	user, status = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, email2, user.CustomerUserId)
	propertiesMap, err = U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, email2, (*propertiesMap)[U.UP_USER_ID])
	assert.Equal(t, phone, (*propertiesMap)[U.UP_PHONE])
	assert.Equal(t, email2, (*propertiesMap)[U.UP_EMAIL])

	// phone number change shouldn't affect customer_user_id
	phone2 := "1234567899"
	timestamp = timestamp + 10
	properties = U.PropertiesMap{"$phone": phone2, "$email": email2}
	trackPayload = &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timestamp,
		EventProperties: properties,
		UserId:          userId,
		RequestSource:   model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Equal(t, http.StatusOK, status)
	user, status = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, email2, user.CustomerUserId)
	propertiesMap, err = U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, email2, (*propertiesMap)[U.UP_USER_ID])
	assert.Equal(t, phone2, (*propertiesMap)[U.UP_PHONE])
	assert.Equal(t, email2, (*propertiesMap)[U.UP_EMAIL])

	/*
		New user with email initially
	*/

	timestamp = timestamp + 10
	email1 = "ma2@mail.com"
	properties = U.PropertiesMap{U.UP_EMAIL: email1}
	trackPayload = &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timestamp,
		EventProperties: properties,
		RequestSource:   model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, response.EventId)
	assert.NotEmpty(t, response.UserId)
	userId = response.UserId
	user, status = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, email1, user.CustomerUserId)
	propertiesMap, err = U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, email1, (*propertiesMap)[U.UP_USER_ID])
	assert.Nil(t, (*propertiesMap)[U.UP_PHONE])
	assert.Equal(t, email1, (*propertiesMap)[U.UP_EMAIL])

	// Only phone property in form submit, should not change customer_user_id. Phone property
	// added.

	timestamp = timestamp + 10
	phone = "1234567899"
	properties = U.PropertiesMap{U.UP_PHONE: phone}
	trackPayload = &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timestamp,
		EventProperties: properties,
		UserId:          userId,
		RequestSource:   model.UserSourceWeb,
	}

	status, response = SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Equal(t, http.StatusOK, status)
	user, status = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, email1, user.CustomerUserId)
	propertiesMap, err = U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, email1, (*propertiesMap)[U.UP_USER_ID])
	// phone property added
	assert.Equal(t, phone, (*propertiesMap)[U.UP_PHONE])
	assert.Equal(t, email1, (*propertiesMap)[U.UP_EMAIL])
}

func TestGetUserByPropertyKey(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, user)

	errCode := store.GetStore().OverwriteUserPropertiesByID(project.ID, user.ID,
		&postgres.Jsonb{RawMessage: json.RawMessage([]byte(
			`{"$hubspot_contact_lead_guid": "xxx"}`))}, false, 0, "")
	assert.Equal(t, http.StatusAccepted, errCode)

	leadUser, errCode := store.GetStore().GetUserByPropertyKey(project.ID,
		model.UserPropertyHubspotContactLeadGUID, "xxx")
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, user.ID, leadUser.ID)
}

func TestUsersUniquenessConstraints(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	segAid := "seg_anon_id_1"
	createdUser1, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, "", time.Now().Unix()-2, model.UserSourceWeb)
	assert.Equal(t, http.StatusCreated, errCode)

	// Should not create new user. Should return same user_id.
	createdUser2, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAid, "", time.Now().Unix()-2, model.UserSourceWeb)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, createdUser1.ID, createdUser2.ID)

	ampUserID := "amp_user_id_1"
	createdUserID11, errCode := store.GetStore().CreateOrGetAMPUser(project.ID, ampUserID, time.Now().Unix()-2, model.UserSourceWeb)
	assert.Equal(t, http.StatusCreated, errCode)

	// Should not create new user. Should return same user_id.
	createdUserID12, errCode := store.GetStore().CreateOrGetAMPUser(project.ID, ampUserID, time.Now().Unix()-2, model.UserSourceWeb)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, createdUserID11, createdUserID12)

	userID := U.GetUUID()
	createdUserID1, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)

	// Should not create new user. Should return same user_id.
	createdUserID2, errCode := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	assert.Equal(t, createdUserID1, createdUserID2)

	_, errCode = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID, SegmentAnonymousId: segAid, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusBadRequest, errCode)

	_, errCode = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID, AMPUserId: ampUserID, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusBadRequest, errCode)
}

func TestUserPropertySkipOnMerge(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	leadGUID1 := "123-45"
	leadGUID2 := "12-345"
	cUID1 := getRandomEmail()
	joinTimestamp := time.Now().AddDate(0, 0, -11)

	// Test user-1 lead guid1 user-2 no lead guid
	user1, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: cUID1, JoinTimestamp: joinTimestamp.Unix(), Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, status)

	properties := &postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"%s":"%s","%s":"%s"}`, model.UserPropertyHubspotContactLeadGUID, leadGUID1, "$hubspot_contact_id", "1"))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user1, properties, joinTimestamp.Unix()+1)

	user2, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: cUID1, JoinTimestamp: joinTimestamp.Unix() + 2, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, status)

	user, status := store.GetStore().GetUser(project.ID, user2)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Nil(t, (*userProperties)[model.UserPropertyHubspotContactLeadGUID])
	assert.Equal(t, "1", (*userProperties)["$hubspot_contact_id"])

	user, status = store.GetStore().GetUser(project.ID, user1)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, leadGUID1, (*userProperties)[model.UserPropertyHubspotContactLeadGUID])

	// Test user-3 lead guid2, same customer user id
	user3, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, CustomerUserId: cUID1, JoinTimestamp: joinTimestamp.Unix(), Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, status)

	properties = &postgres.Jsonb{RawMessage: []byte(fmt.Sprintf(`{"%s":"%s"}`, model.UserPropertyHubspotContactLeadGUID, leadGUID2))}
	_, status = store.GetStore().UpdateUserProperties(project.ID, user3, properties, joinTimestamp.Unix()+3)
	user, status = store.GetStore().GetUser(project.ID, user3)
	assert.Equal(t, http.StatusFound, status)
	userProperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, leadGUID2, (*userProperties)[model.UserPropertyHubspotContactLeadGUID])
}

func TestIdentifiersSkipOnMerge(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	cuid := getRandomEmail()
	userID1, status := store.GetStore().CreateUser(&model.User{
		ProjectId: project.ID,
		Source:    model.GetRequestSourcePointer(model.UserSourceWeb),
	})
	assert.Equal(t, http.StatusCreated, status)

	status, _ = sdk.Identify(project.ID, &SDK.IdentifyPayload{
		UserId: userID1, CustomerUserId: cuid, Source: sdk.SourceJSSDK, RequestSource: model.UserSourceWeb,
	}, true)
	assert.Equal(t, http.StatusOK, status)

	cuid2 := getRandomEmail()
	userID2, status := store.GetStore().CreateUser(&model.User{
		ProjectId: project.ID,
		Source:    model.GetRequestSourcePointer(model.UserSourceWeb),
	})
	assert.Equal(t, http.StatusCreated, status)

	status, _ = sdk.Identify(project.ID, &SDK.IdentifyPayload{
		UserId: userID2, CustomerUserId: cuid2, Source: sdk.SourceJSSDK, RequestSource: model.UserSourceWeb,
	}, true)
	assert.Equal(t, http.StatusOK, status)

	status, _ = sdk.Identify(project.ID, &SDK.IdentifyPayload{
		UserId: userID2, CustomerUserId: cuid, Source: sdk.SourceJSSDK, RequestSource: model.UserSourceWeb,
	}, true)
	assert.Equal(t, http.StatusOK, status)
	user1, status := store.GetStore().GetUser(project.ID, userID1)
	assert.Equal(t, http.StatusFound, status)
	user2, status := store.GetStore().GetUser(project.ID, userID2)
	assert.Equal(t, http.StatusFound, status)
	user1PropertiesMap, err := U.DecodePostgresJsonb(&user1.Properties)
	assert.Nil(t, err)
	user2PropertiesMap, err := U.DecodePostgresJsonb(&user2.Properties)
	assert.Nil(t, err)
	user1MetaObject, err := model.GetDecodedUserPropertiesIdentifierMetaObject(user1PropertiesMap)
	assert.Nil(t, err)
	user2MetaObject, err := model.GetDecodedUserPropertiesIdentifierMetaObject(user2PropertiesMap)
	assert.Nil(t, err)
	assert.Contains(t, *user1MetaObject, cuid)
	assert.NotContains(t, *user1MetaObject, cuid2)

	assert.Contains(t, *user2MetaObject, cuid)
	assert.Contains(t, *user2MetaObject, cuid2)

}

func TestGetSelectedUsersByCustomerUserID(t *testing.T) {
	// Initialize a project for the user.
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	assert.NotNil(t, project)
	projectId := project.ID

	// Create 10 users
	// Set the limit to fetch top and bottom users
	var limit uint64 = 10
	var numUsers uint64 = 4

	var users []model.User
	customer_id := "Taashish"
	for i := 0; i < int(limit); i++ {
		createdUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: projectId, CustomerUserId: customer_id, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
		assert.Equal(t, http.StatusCreated, errCode)
		lastUser, lastUserErrorCode := store.GetStore().GetUserLatestByCustomerUserId(projectId, customer_id, model.UserSourceWeb)
		assert.Equal(t, http.StatusFound, lastUserErrorCode)
		assert.Equal(t, createdUserID, lastUser.ID)
		users = append(users, *lastUser)
	}

	retUsers, errCode := store.GetStore().GetSelectedUsersByCustomerUserID(projectId, customer_id, limit, numUsers)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, numUsers, uint64(len(retUsers)))

	for i := 0; i < int(numUsers/2); i++ {
		assert.Equal(t, users[i].ID, retUsers[i].ID)
		assert.Equal(t, users[int(limit)-i-1].ID, retUsers[int(numUsers)-i-1].ID)
	}

}

func TestUserIntialPropertiesOnOldTimestamp(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	currentTime := time.Now()
	u1JointTimestamp := currentTime.AddDate(0, 0, -3)
	u2JointTimestamp := currentTime.AddDate(0, 0, -2)
	cuid := getRandomEmail()
	user1, status := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		CustomerUserId: cuid,
		JoinTimestamp:  u1JointTimestamp.Unix(),
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
	})
	assert.Equal(t, http.StatusCreated, status)

	_, status = store.GetStore().UpdateUserProperties(project.ID, user1, &postgres.Jsonb{[]byte(`{"city":"A"}`)}, currentTime.Unix())
	assert.Equal(t, http.StatusAccepted, status)

	user2, status := store.GetStore().CreateUser(&model.User{
		ProjectId:      project.ID,
		CustomerUserId: cuid,
		JoinTimestamp:  u2JointTimestamp.Unix(),
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
	})
	assert.Equal(t, http.StatusCreated, status)

	_, status = store.GetStore().UpdateUserProperties(project.ID, user1, &postgres.Jsonb{[]byte(`{"city":"B"}`)}, currentTime.Unix())
	assert.Equal(t, http.StatusAccepted, status)

	user, status := store.GetStore().GetUser(project.ID, user1)
	assert.Equal(t, http.StatusFound, status)
	userproperties, err := U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)

	assert.Equal(t, float64(u1JointTimestamp.Unix()), (*userproperties)["$joinTime"])
	assert.Equal(t, "B", (*userproperties)["city"])

	user, status = store.GetStore().GetUser(project.ID, user2)
	assert.Equal(t, http.StatusFound, status)
	userproperties, err = U.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)

	assert.Equal(t, float64(u1JointTimestamp.Unix()), (*userproperties)["$joinTime"])
	assert.Equal(t, "B", (*userproperties)["city"])
}

func TestUserPropertiesUpdateByGroupColumnName(t *testing.T) {
	project, user, err := SetupProjectUserReturnDAO()
	assert.Nil(t, err)

	userProperties, err := json.Marshal(`{"a":"b"}`)
	assert.Nil(t, err)
	timestamp := time.Now().Unix()
	segmentAnonymousId := getRandomEmail()
	ampUserID := getRandomEmail()
	testUser := model.User{
		ProjectId:                  project.ID,
		ID:                         user.ID,
		CustomerUserId:             getRandomEmail(),
		Properties:                 postgres.Jsonb{userProperties},
		PropertiesUpdatedTimestamp: timestamp,
		SegmentAnonymousId:         segmentAnonymousId,
		AMPUserId:                  ampUserID,
		JoinTimestamp:              timestamp,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	/*
		Test update by group column name should not affect other fileds
	*/
	testUserCopy := testUser
	assert.Equal(t, "", testUserCopy.Group1ID)
	processed, updated, err := model.SetUserGroupFieldByColumnName(&testUserCopy, "group_1_id", "g1")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_1_user_id", "g1user")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)
	// assert values are assigned
	assert.Equal(t, "g1", testUserCopy.Group1ID)
	assert.Equal(t, "g1user", testUserCopy.Group1UserID)

	assert.NotEqual(t, testUser, testUserCopy)
	// remove field property for equality check
	testUserCopy.Group1ID = ""
	testUserCopy.Group1UserID = ""
	assert.Equal(t, testUser, testUserCopy)

	/*
	 Test Multipele group column updates
	*/
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_1_id", "g1")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_2_id", "g2")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)

	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_1_user_id", "g1user")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_2_user_id", "g2user")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)

	assert.Equal(t, "g1", testUserCopy.Group1ID)
	assert.Equal(t, "g1user", testUserCopy.Group1UserID)
	assert.Equal(t, "g2", testUserCopy.Group2ID)
	assert.Equal(t, "g2user", testUserCopy.Group2UserID)

	assert.NotEqual(t, testUser, testUserCopy)
	testUserCopy.Group1ID = ""
	testUserCopy.Group1UserID = ""
	testUserCopy.Group2ID = ""
	testUserCopy.Group2UserID = ""
	assert.Equal(t, testUser, testUserCopy)

	/*
	 Test update not allowed for already set value
	*/
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_1_user_id", "g1user")
	assert.Nil(t, err)
	assert.Equal(t, true, processed)
	assert.Equal(t, true, updated)
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_1_user_id", "g1user2")
	assert.Nil(t, err)
	// got processed but didn't allow update
	assert.Equal(t, true, processed)
	assert.Equal(t, false, updated)
	assert.Equal(t, "g1user", testUserCopy.Group1UserID)
	testUserCopy.Group1UserID = ""
	assert.Equal(t, testUser, testUserCopy)

	/*
		Test invalid column name
	*/
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "group_5_user_id", "g1user")
	assert.NotNil(t, err)
	assert.Equal(t, false, processed) // didn't find field
	assert.Equal(t, false, updated)
	assert.Equal(t, testUser, testUserCopy)

	/*
		Test update non group column
	*/
	processed, updated, err = model.SetUserGroupFieldByColumnName(&testUserCopy, "amp_user_id", "g1user")
	assert.NotNil(t, err)
	assert.Equal(t, false, processed)
	assert.Equal(t, false, updated)
	assert.Equal(t, testUser, testUserCopy)
}

func TestUserGroupsPropertiesUpdate(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	groupName := "g1"
	groupID := "g1ID"
	allowedGroupsMap := map[string]bool{groupName: true}
	group, status := store.GetStore().CreateGroup(project.ID, groupName, allowedGroupsMap)
	assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName))
	assert.Equal(t, 1, group.ID)
	timestamp := time.Now().AddDate(0, 0, -1)

	groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
		ProjectId: project.ID, JoinTimestamp: timestamp.Unix() - 10, Source: model.GetRequestSourcePointer(model.UserSourceHubspot),
	}, groupName, groupID)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().UpdateUserGroupProperties(project.ID, groupUserID, &postgres.Jsonb{json.RawMessage([]byte(`{"hour":1,"count":2,"city":"Bengalore"}`))}, timestamp.Unix())
	assert.Equal(t, http.StatusAccepted, status)
	user, status := store.GetStore().GetUser(project.ID, groupUserID)
	assert.Equal(t, http.StatusFound, status)
	userPropertiesMap, err := util.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.NotNil(t, user.IsGroupUser)
	assert.Equal(t, true, *user.IsGroupUser)
	assert.Equal(t, float64(1), (*userPropertiesMap)["hour"])
	assert.Equal(t, float64(2), (*userPropertiesMap)["count"])
	assert.Equal(t, "Bengalore", (*userPropertiesMap)["city"])

	_, status = store.GetStore().UpdateUserGroupProperties(project.ID, groupUserID, &postgres.Jsonb{json.RawMessage([]byte(`{"city":"Delhi"}`))}, timestamp.Unix()-10)
	assert.Equal(t, http.StatusAccepted, status)
	user, status = store.GetStore().GetUser(project.ID, groupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, user.IsGroupUser)
	assert.Equal(t, true, *user.IsGroupUser)
	assert.Equal(t, timestamp.Unix(), user.PropertiesUpdatedTimestamp)
	userPropertiesMap, err = util.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*userPropertiesMap)["hour"])
	assert.Equal(t, float64(2), (*userPropertiesMap)["count"])
	assert.Equal(t, "Bengalore", (*userPropertiesMap)["city"])

	_, status = store.GetStore().UpdateUserGroupProperties(project.ID, groupUserID, &postgres.Jsonb{json.RawMessage([]byte(`{"city":"Delhi"}`))}, timestamp.Unix()+10)
	assert.Equal(t, http.StatusAccepted, status)
	user, status = store.GetStore().GetUser(project.ID, groupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, user.IsGroupUser)
	assert.Equal(t, true, *user.IsGroupUser)
	assert.Equal(t, timestamp.Unix()+10, user.PropertiesUpdatedTimestamp)
	userPropertiesMap, err = util.DecodePostgresJsonb(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, float64(1), (*userPropertiesMap)["hour"])
	assert.Equal(t, float64(2), (*userPropertiesMap)["count"])
	assert.Equal(t, "Delhi", (*userPropertiesMap)["city"])

	// test isGroupUser property value
	docID := "1"
	userID, status := store.GetStore().CreateUser(&model.User{
		ProjectId: project.ID,
		Source:    model.GetRequestSourcePointer(model.UserSourceWeb),
	})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().UpdateUserGroup(project.ID, userID, groupName, docID, groupUserID)
	assert.Equal(t, http.StatusAccepted, status)
	user, status = store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, user.IsGroupUser)
	assert.Equal(t, false, *user.IsGroupUser)
}

func TestUserSourceColumn(t *testing.T) {
	// Initialize a project for the user.
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	projectId := project.ID

	// Test successful create user, with source value getting successfully stored
	createUserID, errCode := store.GetStore().CreateUser(&model.User{ProjectId: projectId, Source: model.GetRequestSourcePointer(model.UserSourceWeb)})
	assert.Equal(t, http.StatusCreated, errCode)
	user, errCode := store.GetStore().GetUser(projectId, createUserID)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, model.UserSourceWeb, *user.Source)

	// Test un-successful create user when source is not passed
	_, errCode = store.GetStore().CreateUser(&model.User{ProjectId: projectId})
	assert.Equal(t, http.StatusInternalServerError, errCode)

	// Test for successfull create group user
	groupName := "g1"
	groupID := "g1ID"
	allowedGroupsMap := map[string]bool{groupName: true}
	group, status := store.GetStore().CreateGroup(project.ID, groupName, allowedGroupsMap)
	assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName))
	assert.Equal(t, 1, group.ID)
	timestamp := time.Now().AddDate(0, 0, -1)
	groupUserID, status := store.GetStore().CreateGroupUser(&model.User{
		ProjectId: project.ID, JoinTimestamp: timestamp.Unix() - 10, Source: model.GetRequestSourcePointer(model.UserSourceWeb),
	}, groupName, groupID)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().UpdateUserGroupProperties(project.ID, groupUserID, &postgres.Jsonb{json.RawMessage([]byte(`{"hour":1,"count":2,"city":"Bengalore"}`))}, timestamp.Unix())
	assert.Equal(t, http.StatusAccepted, status)
	user, status = store.GetStore().GetUser(project.ID, groupUserID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, model.UserSourceWeb, *user.Source)

	// Test for successfull CreateOrGetAMPUser
	userAgentStr := "Mozilla/5.0 (Linux; Android 8.0.0; SM-G960F Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.84 Mobile Safari/537.36"
	ampClientId := "amp-1xxAGEAL-irIHu4qMW8j3A"
	payload := &SDK.AMPTrackPayload{
		ClientID:      ampClientId,
		SourceURL:     "abcd.com/",
		Title:         "Test",
		Timestamp:     time.Now().Unix(),
		UserAgent:     userAgentStr,
		ClientIP:      "10.10.0.1",
		RequestSource: model.UserSourceWeb,
	}
	errCode, _ = SDK.AMPTrackByToken(project.Token, payload)
	assert.Equal(t, errCode, http.StatusOK)
	userID, errCode := store.GetStore().CreateOrGetAMPUser(project.ID, ampClientId, payload.Timestamp, model.UserSourceWeb)
	assert.Equal(t, errCode, http.StatusFound)
	user, status = store.GetStore().GetUser(project.ID, userID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, model.UserSourceWeb, *user.Source)

	// Test for successfull CreateOrGetSegmentUser
	customerUserID := U.RandomLowerAphaNumString(15) + "@example.com"
	user, errCode = store.GetStore().CreateOrGetSegmentUser(project.ID, "", customerUserID, time.Now().Unix(), model.UserSourceWeb)
	assert.Equal(t, http.StatusCreated, errCode)
	user, status = store.GetStore().GetUser(project.ID, user.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, model.UserSourceWeb, *user.Source)
}

func TestUserDuplicates(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		ampUserID := getRandomEmail()

		timeStamp := time.Now().Unix()
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				userID, errCode := store.GetStore().CreateOrGetAMPUser(project.ID, ampUserID, timeStamp, model.UserSourceWeb)
				assert.Contains(t, []int{http.StatusCreated, http.StatusFound}, errCode)
				assert.NotEqual(t, "", userID)
			}()

		}
	}
	wg.Wait()

	users, _ := store.GetStore().GetUsers(project.ID, 0, 50)
	assert.Len(t, users, 5)

	for i := range users {
		assert.NotEqual(t, "", users[i].AMPUserId)
	}

	// test segment_user
	project, err = SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	for i := 0; i < 5; i++ {
		segAnonID := getRandomEmail()

		timeStamp := time.Now().Unix()
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				user, errCode := store.GetStore().CreateOrGetSegmentUser(project.ID, segAnonID, "", timeStamp, model.UserSourceWeb)
				assert.Contains(t, []int{http.StatusCreated, http.StatusFound}, errCode)
				assert.NotNil(t, user)
			}()

		}
	}
	wg.Wait()

	users, _ = store.GetStore().GetUsers(project.ID, 0, 50)
	assert.Len(t, users, 5)

	for i := range users {
		assert.NotEqual(t, "", users[i].SegmentAnonymousId)
	}
}

func clearbit1(ip string, c chan int) {
	projectId := uint64(4)
	clearbitKey, errCode := store.GetStore().GetClearbitKeyFromProjectSetting(projectId)

	if errCode != http.StatusFound {
		log.Error("Get clear_bit key failed. Invalid project.")
	}
	client := clearbit.NewClient(clearbit.WithAPIKey(clearbitKey))
	_, _, _ = client.Reveal.Find(clearbit.RevealFindParams{
		IP: ip,
	})
	res := 5
	c <- res
}
func clearbitAnalysisTestDBClearBit(t *testing.T) {
	ip := []string{
		"89.76.236.199",
		"137.59.242.126",
		"66.249.81.247",
		"66.102.8.22",
		"59.88.57.114",
		"66.249.81.245",
		"185.229.59.141",
		"49.32.238.36",
		"84.203.150.136",
		"162.253.209.154",
		"68.74.220.125",
		"185.53.227.164",
		"35.88.129.120",
		"90.63.220.183",
		"185.229.59.141",
		"24.162.23.57",
		"223.190.80.87",
		"66.249.88.30",
		"185.229.59.141",
		"66.249.93.1",
		"66.102.8.20",
		"188.146.239.41",
		"66.249.93.5",
		"31.13.127.8",
		"42.111.163.64",
		"157.51.142.107",
		"98.242.148.59",
		"49.37.219.158",
		"213.32.243.78",
		"193.239.59.85",
		"66.249.83.60",
		"94.190.204.222",
		"34.221.93.80",
		"66.249.93.3",
		"117.216.103.89",
		"174.92.160.98",
		"74.6.40.75",
		"152.57.192.207",
		"5.57.52.47",
		"180.129.121.178",
		"66.102.6.180",
		"66.102.9.20",
		"79.21.248.188",
		"66.249.81.243",
		"66.249.93.1",
		"92.77.126.233",
		"81.19.223.22",
		"66.102.9.20",
		"66.249.93.3",
		"182.72.76.34",
		"124.253.113.83",
		"66.249.83.56",
		"66.249.93.3",
		"34.105.204.94",
		"49.207.202.238",
		"66.102.6.184",
		"12.167.12.226",
		"66.102.8.22",
		"66.249.88.30",
		"168.149.146.88",
		"81.128.133.67",
		"2.84.215.126",
		"66.249.83.56",
		"77.58.53.241",
		"213.123.54.101",
		"68.82.33.200",
		"106.51.1.148",
		"80.59.251.252",
		"109.245.127.240",
		"31.10.147.39",
		"176.61.95.239",
		"52.143.249.184",
		"31.13.127.31",
		"49.207.212.78",
		"49.37.210.172",
		"94.44.225.229",
		"87.196.73.101",
		"106.198.83.3",
		"103.211.22.54",
		"35.86.97.98",
		"66.249.93.5",
		"142.129.183.129",
		"66.249.93.3",
		"44.193.77.14",
		"76.112.99.78",
		"59.103.76.123",
		"66.249.81.247",
		"66.249.83.56",
		"185.204.179.225",
		"5.91.254.62",
		"66.249.93.3",
		"85.246.71.150",
		"62.85.92.110",
		"45.28.131.51",
		"183.83.208.186",
		"168.149.165.60",
		"103.238.109.174",
		"173.238.75.179",
		"218.185.235.242",
		"66.102.8.24",
	}

	for i := 0; i < 1; i++ {
		c := make(chan int)
		go clearbit1(ip[i], c)
		select {
		case response := <-c:
			fmt.Println(response)
			//fmt.Println(response.length)
		case <-time.After(1 / 10 * time.Second):
			fmt.Println("time Out :(")
		}
		//fmt.Println("Main function exited!")
	}

}
