package tests

import (
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupCreation(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	group, status := store.GetStore().GetGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Nil(t, group)
	group, status = store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusCreated, status)
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.ID)
	group, status = store.GetStore().GetGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.ID)

	group, status = store.GetStore().CreateGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
	assert.Equal(t, http.StatusConflict, status)
	assert.Nil(t, group)
	group, status = store.GetStore().GetGroup(project.ID, model.GROUP_NAME_HUBSPOT_COMPANY)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, group)
	assert.Equal(t, 1, group.ID)

	allowedGroups := []string{
		"g1",
		"g2",
		"g3",
		"g4",
	}

	allowedGroupsMap := map[string]bool{
		"g1": true,
		"g2": true,
		"g3": true,
		"g4": true,
	}

	index := 2
	for _, groupName := range allowedGroups {
		if groupName == "g4" {
			continue
		}

		group, status = store.GetStore().CreateGroup(project.ID, groupName, allowedGroupsMap)
		assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName))
		assert.NotNil(t, group)
		assert.Equal(t, index, group.ID)
		index++
	}

	index = 2
	for _, groupName := range allowedGroups {
		if groupName == "g4" {
			continue
		}
		group, status = store.GetStore().CreateGroup(project.ID, groupName, allowedGroupsMap)
		assert.Equal(t, http.StatusConflict, status)
		assert.Nil(t, group)
		group, status = store.GetStore().GetGroup(project.ID, groupName)
		assert.Equal(t, http.StatusFound, status)
		assert.NotNil(t, group)
		assert.Equal(t, index, group.ID)
		index++
	}

	// should fail, only 4 groups allowed
	group, status = store.GetStore().CreateGroup(project.ID, "g4", allowedGroupsMap)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, group)
}

func assertUserGroupValueByColumnName(user *model.User, columnName string, value string) bool {
	switch columnName {
	case "group_1_id":
		return user.Group1ID == value
	case "group_1_user_id":
		return user.Group1UserID == value
	case "group_2_id":
		return user.Group2ID == value
	case "group_2_user_id":
		return user.Group2UserID == value
	case "group_3_id":
		return user.Group3ID == value
	case "group_3_user_id":
		return user.Group3UserID == value
	case "group_4_id":
		return user.Group4ID == value
	case "group_4_user_id":
		return user.Group4UserID == value
	default:
		return false
	}
}

func TestUserGroups(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	allowedGroups := []string{
		"g1",
		"g2",
		"g3",
		"g4",
	}

	allowedGroupsMap := map[string]bool{
		"g1": true,
		"g2": true,
		"g3": true,
		"g4": true,
	}

	index := 1
	for _, groupName := range allowedGroups {
		group, status := store.GetStore().CreateGroup(project.ID, groupName, allowedGroupsMap)
		assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName))
		assert.NotNil(t, group)
		assert.Equal(t, index, group.ID)
		index++
	}

	userIDs := make([]string, len(allowedGroups))
	groupIDs := []string{"1", "2", "3", "4"}
	for i := range allowedGroups {
		userID, status := store.GetStore().CreateGroupUser(&model.User{
			ProjectId: project.ID,
			Source:    model.GetRequestSourcePointer(model.UserSourceHubspot),
		}, allowedGroups[i], groupIDs[i])
		assert.Equal(t, http.StatusCreated, status)
		assert.NotEqual(t, "", userID)
		userIDs[i] = userID
	}

	for i := range userIDs {
		user, status := store.GetStore().GetUser(project.ID, userIDs[i])
		assert.Equal(t, http.StatusFound, status)
		assert.Equal(t, true, assertUserGroupValueByColumnName(user, fmt.Sprintf("group_%d_id", i+1), groupIDs[i]))
		assert.Equal(t, true, assertUserGroupValueByColumnName(user, fmt.Sprintf("group_%d_user_id", i+1), ""))
	}

}

func TestGroupRelationship(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	groupName1 := "g1"
	groupName2 := "g2"
	user1 := getRandomName()
	user2 := getRandomName()

	// fail if group name does not exist
	groupRelationship, status := store.GetStore().CreateGroupRelationship(project.ID, groupName1, user1, groupName2, user2)
	assert.Equal(t, http.StatusBadRequest, status)
	groupRelationships, status := store.GetStore().GetGroupRelationshipByUserID(project.ID, user1)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Len(t, groupRelationships, 0)

	allowedGroup := map[string]bool{
		groupName1: true,
		groupName2: true,
	}

	// Create groups for creating group relationships
	_, status = store.GetStore().CreateGroup(project.ID, groupName1, allowedGroup)
	assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName1))
	_, status = store.GetStore().CreateGroup(project.ID, groupName2, allowedGroup)
	assert.Equal(t, http.StatusCreated, status, fmt.Sprintf("failed creating group %s", groupName2))

	groupRelationship, status = store.GetStore().CreateGroupRelationship(project.ID, groupName1, user1, groupName2, user2)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, 1, groupRelationship.LeftGroupNameID)
	assert.Equal(t, user1, groupRelationship.LeftGroupUserID)
	assert.Equal(t, 2, groupRelationship.RightGroupNameID)
	assert.Equal(t, user2, groupRelationship.RightGroupUserID)

	_, status = store.GetStore().CreateGroupRelationship(project.ID, groupName1, user1, groupName2, user2)
	assert.Equal(t, http.StatusConflict, status)
	groupRelationships, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, user1)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationships, 1)
	assert.Equal(t, 1, groupRelationships[0].LeftGroupNameID)
	assert.Equal(t, user1, groupRelationships[0].LeftGroupUserID)
	assert.Equal(t, 2, groupRelationships[0].RightGroupNameID)
	assert.Equal(t, user2, groupRelationships[0].RightGroupUserID)

	//create reverse relationship
	_, status = store.GetStore().CreateGroupRelationship(project.ID, groupName2, user2, groupName1, user1)
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateGroupRelationship(project.ID, groupName2, user2, groupName1, user1)
	assert.Equal(t, http.StatusConflict, status)

	// verify previous relation exists
	groupRelationships, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, user1)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, 1, groupRelationships[0].LeftGroupNameID)
	assert.Equal(t, user1, groupRelationships[0].LeftGroupUserID)
	assert.Equal(t, 2, groupRelationships[0].RightGroupNameID)
	assert.Equal(t, user2, groupRelationships[0].RightGroupUserID)

	// verify new relationship
	groupRelationships, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, user2)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationships, 1)
	assert.Equal(t, 2, groupRelationships[0].LeftGroupNameID)
	assert.Equal(t, user2, groupRelationships[0].LeftGroupUserID)
	assert.Equal(t, 1, groupRelationships[0].RightGroupNameID)
	assert.Equal(t, user1, groupRelationships[0].RightGroupUserID)

	// create multiple group relationship for user2
	users := map[string]bool{getRandomName(): true, getRandomName(): true, getRandomName(): true, getRandomName(): true}
	for userID := range users {
		_, status = store.GetStore().CreateGroupRelationship(project.ID, groupName2, user2, groupName1, userID)
		assert.Equal(t, http.StatusCreated, status)
	}

	groupRelationships, status = store.GetStore().GetGroupRelationshipByUserID(project.ID, user2)
	assert.Equal(t, http.StatusFound, status)
	assert.Len(t, groupRelationships, 5)
	groupRelationshipsMap := make(map[string]bool)
	for i := range groupRelationships {
		assert.Equal(t, 2, groupRelationships[i].LeftGroupNameID)
		assert.Equal(t, user2, groupRelationships[i].LeftGroupUserID)
		assert.Equal(t, 1, groupRelationships[i].RightGroupNameID)
		groupRelationshipsMap[groupRelationships[i].RightGroupUserID] = true
	}

	assert.True(t, groupRelationshipsMap[user1])
	for user := range users {
		assert.True(t, groupRelationshipsMap[user])
	}
}
