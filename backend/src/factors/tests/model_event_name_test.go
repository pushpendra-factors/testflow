package tests

import (
	M "factors/model"
	U "factors/util"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetEventName(t *testing.T) {
	// Initialize a project for the event.
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProject(&M.Project{Name: randomProjectName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	start := time.Now()

	// Test successful create eventName.
	eventName, errCode := M.CreateOrGetEventName(&M.EventName{Name: "test_event", ProjectId: projectId})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, projectId, eventName.ProjectId)
	assert.True(t, eventName.CreatedAt.After(start))
	// Trying to create again should return the old one.
	expectedEventName := &M.EventName{}
	copier.Copy(expectedEventName, eventName)
	retryEventName, errCode := M.CreateOrGetEventName(&M.EventName{Name: "test_event", ProjectId: projectId})
	assert.Equal(t, http.StatusConflict, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(expectedEventName.CreatedAt.Sub(retryEventName.CreatedAt).Seconds()) < 0.1)
	expectedEventName.CreatedAt = time.Time{}
	retryEventName.CreatedAt = time.Time{}
	assert.Equal(t, expectedEventName, retryEventName)
	// Test Get EventName on the created one.
	expectedEventName = &M.EventName{}
	copier.Copy(expectedEventName, eventName)
	retEventName, errCode := M.GetEventName(expectedEventName.Name, projectId)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(expectedEventName.CreatedAt.Sub(retEventName.CreatedAt).Seconds()) < 0.1)
	expectedEventName.CreatedAt = time.Time{}
	retEventName.CreatedAt = time.Time{}
	assert.Equal(t, expectedEventName, retEventName)

	// Test Get Event on non existent name.
	retEventName, errCode = M.GetEventName("non_existent_event", projectId)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEventName)

	// Test Get Event with only name.
	retEventName, errCode = M.GetEventName(eventName.Name, 0)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test Get Event with only projectId.
	retEventName, errCode = M.GetEventName("", projectId)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test Validate auto_name on CreateOrGetUserCreatedEventName.
	randomName := U.RandomLowerAphaNumString(10)
	ucEventName := &M.EventName{Name: randomName, ProjectId: project.ID}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, M.USER_CREATED_EVENT_NAME, retEventName.AutoName)

	// Test Duplicate creation of user created event name. Should be unique by project.
	duplicateEventName, errCode := M.CreateOrGetUserCreatedEventName(&M.EventName{Name: randomName, ProjectId: project.ID})
	assert.Equal(t, http.StatusConflict, errCode) // Should return conflict with the conflicted object.
	assert.Equal(t, M.USER_CREATED_EVENT_NAME, retEventName.AutoName)
	assert.Equal(t, retEventName.ID, duplicateEventName.ID)

	// Test CreateOrGetUserCreatedEventName without ProjectId.
	ucEventName = &M.EventName{Name: U.RandomLowerAphaNumString(10)}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)

	// Test CreateOrGetUserCreatedEventName without name.
	ucEventName = &M.EventName{Name: "", ProjectId: project.ID}
	retEventName, errCode = M.CreateOrGetUserCreatedEventName(ucEventName)
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, retEventName)
}
