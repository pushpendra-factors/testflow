package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndUpdateCustomerUserId(t *testing.T) {
	// Create user with a customer_user_id, verify.

	project, agent, err := SetupProjectWithAgentDAO()
	assert.NotNil(t, agent)
	assert.Nil(t, err)

	cuid := "abc_customer_user_id"
	user1 := &model.User{
		ProjectId:      project.ID,
		CustomerUserId: cuid,
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
		Properties:     postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"city":"city1"}`))},
	}
	createdUserID1, status := store.GetStore().CreateUser(user1)
	assert.Equal(t, http.StatusCreated, status)
	getuser, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, cuid, getuser.CustomerUserId)
	assert.Equal(t, createdUserID1, getuser.ID)
	properties, status := store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, cuid, (*properties)[U.UP_USER_ID])

	// 	Update customer_user_id, verify.
	newCuid := "abc_customer_user_id_updated_1"
	segAid := "seg_aid_1"
	updateUser := &model.User{CustomerUserId: newCuid,
		SegmentAnonymousId: segAid,
		Properties:         postgres.Jsonb{json.RawMessage(`{"key": "value"}`)}}
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1,
		updateUser, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)

	getuser, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, newCuid, getuser.CustomerUserId)
	assert.Equal(t, segAid, getuser.SegmentAnonymousId)
	properties, status = store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, newCuid, (*properties)[U.UP_USER_ID])
	assert.Equal(t, "city1", (*properties)["city"])
	assert.Equal(t, "value", (*properties)["key"])

	// 	Update customer_user_id, verify.
	newCuid = "customer_user_id1"
	segAid = "new_seg_aid"
	updateUser = &model.User{CustomerUserId: newCuid,
		SegmentAnonymousId: segAid,
		Properties:         postgres.Jsonb{json.RawMessage(`{"key": "newvalue"}`)}}
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1,
		updateUser, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)

	getuser, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, newCuid, getuser.CustomerUserId)
	assert.Equal(t, segAid, getuser.SegmentAnonymousId)
	properties, status = store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, newCuid, (*properties)[U.UP_USER_ID])
	assert.Equal(t, "city1", (*properties)["city"])
	assert.Equal(t, "newvalue", (*properties)["key"])

	// Create User id2
	project1, agent, err := SetupProjectWithAgentDAO()
	assert.NotNil(t, agent)
	assert.Nil(t, err)

	user2 := &model.User{
		ProjectId:      project1.ID,
		CustomerUserId: newCuid,
		Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
		Properties:     postgres.Jsonb{RawMessage: json.RawMessage([]byte(`{"city":"city1"}`))},
	}
	createdUserID2, status := store.GetStore().CreateUser(user2)
	assert.Equal(t, http.StatusCreated, status)

	// SDK.Identify(id1, customer_user_id1)
	status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{UserId: createdUserID1, CustomerUserId: newCuid, RequestSource: model.UserSourceWeb}, false)
	assert.Equal(t, http.StatusOK, status)

	// SDK.Identify(id2, customer_user_id1)
	status, _ = SDK.Identify(project1.ID, &SDK.IdentifyPayload{UserId: createdUserID2, CustomerUserId: newCuid, RequestSource: model.UserSourceWeb}, false)
	assert.Equal(t, http.StatusOK, status)

	// Get user id1
	getuser, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, newCuid, getuser.CustomerUserId)
	assert.Equal(t, segAid, getuser.SegmentAnonymousId)
	properties, status = store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, newCuid, (*properties)[U.UP_USER_ID])
	assert.Equal(t, "city1", (*properties)["city"])
	assert.Equal(t, "newvalue", (*properties)["key"])

	// Test id1.customer_user_id == id1.$user_id
	assert.Equal(t, getuser.CustomerUserId, (*properties)[U.UP_USER_ID])

	// Get user id2
	getuser2, errCode := store.GetStore().GetUser(project1.ID, createdUserID2)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, newCuid, getuser2.CustomerUserId)
	properties, status = store.GetStore().GetLatestUserPropertiesOfUserAsMap(project1.ID, createdUserID2)
	assert.Equal(t, http.StatusFound, status)

	// Test id2.customer_user_id == id2.$user_id
	assert.Equal(t, getuser2.CustomerUserId, (*properties)[U.UP_USER_ID])

}

func TestDBCreateWithoutCustomerUserIdAndUpdateCustomerUserId(t *testing.T) {
	// Create user without customer_user_id, verify.
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	joinTimestamp := time.Now().AddDate(0, 0, -11)
	user1 := &model.User{
		ProjectId: project.ID,
		Source:    model.GetRequestSourcePointer(model.UserSourceWeb),
	}
	createdUserID1, status := store.GetStore().CreateUser(user1)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)

	getuser, errCode := store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)

	// 	Update customer_user_id, verify.
	cuid := "abc_customer_user_id_created_1"
	segAid := "seg_aid"
	updateUser := &model.User{CustomerUserId: cuid,
		SegmentAnonymousId: segAid,
		Properties:         postgres.Jsonb{json.RawMessage(`{"key": "value"}`)}}
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1, updateUser, joinTimestamp.Unix()+1)
	assert.Equal(t, http.StatusAccepted, errCode)

	getuser, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, cuid, getuser.CustomerUserId)
	assert.Equal(t, segAid, getuser.SegmentAnonymousId)

	properties, status := store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, cuid, (*properties)[U.UP_USER_ID])
	assert.Equal(t, "value", (*properties)["key"])

	// Update customer_user_id, verify.
	newCuid := "abc_customer_user_id_created_2"
	segAid = "new_seg_aid"
	updateUser = &model.User{CustomerUserId: newCuid,
		SegmentAnonymousId: segAid,
		Properties:         postgres.Jsonb{json.RawMessage(`{"key": "newvalue"}`)}}
	_, errCode = store.GetStore().UpdateUser(project.ID, createdUserID1,
		updateUser, time.Now().Unix())
	assert.Equal(t, http.StatusAccepted, errCode)

	getuser, errCode = store.GetStore().GetUser(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, newCuid, getuser.CustomerUserId)
	assert.Equal(t, segAid, getuser.SegmentAnonymousId)

	properties, status = store.GetStore().GetLatestUserPropertiesOfUserAsMap(project.ID, createdUserID1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, newCuid, (*properties)[U.UP_USER_ID])
	assert.Equal(t, "newvalue", (*properties)["key"])
}
