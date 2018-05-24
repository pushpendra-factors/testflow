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

func TestDBCreateAndGetEventName(t *testing.T) {
	// Initialize a project for the event.
	randomProjectName := U.RandomLowerAphaNumString(15)
	project, errCode := M.CreateProject(&M.Project{Name: randomProjectName})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.NotNil(t, project)
	projectId := project.ID

	start := time.Now()

	// Test successful create eventName.
	eventName, errCode := M.CreateEventName(&M.EventName{Name: "test_event", ProjectId: projectId})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.Equal(t, projectId, eventName.ProjectId)
	assert.True(t, eventName.CreatedAt.After(start))
	// Test Get Project on the created one.
	retEventName, errCode := M.GetEventName(eventName.Name, projectId)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(eventName.CreatedAt.Sub(retEventName.CreatedAt).Seconds()) < 0.1)
	eventName.CreatedAt = time.Time{}
	retEventName.CreatedAt = time.Time{}
	assert.Equal(t, eventName, retEventName)

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
}
