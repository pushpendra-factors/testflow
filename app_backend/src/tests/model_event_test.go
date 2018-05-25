package tests

import (
	"encoding/json"
	"math"
	M "model"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateAndGetEvent(t *testing.T) {
	// Initialize a project, user and  the event.
	projectId, userId, eventName, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	// Test successful CreateEvent.
	event, errCode := M.CreateEvent(&M.Event{EventName: eventName, ProjectId: projectId, UserId: userId})
	assert.Equal(t, M.DB_SUCCESS, errCode)
	assert.True(t, len(event.ID) > 30)
	assert.Equal(t, projectId, event.ProjectId)
	assert.Equal(t, eventName, event.EventName)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, event.UpdatedAt.After(start))
	assert.Equal(t, event.CreatedAt, event.UpdatedAt)
	assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, event.Properties)
	// Test Get Event on the created.
	retEvent, errCode := M.GetEvent(projectId, userId, event.ID)
	assert.Equal(t, M.DB_SUCCESS, errCode)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(event.CreatedAt.Sub(retEvent.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(event.UpdatedAt.Sub(retEvent.UpdatedAt).Seconds()) < 0.1)
	event.CreatedAt = time.Time{}
	event.UpdatedAt = time.Time{}
	retEvent.CreatedAt = time.Time{}
	retEvent.UpdatedAt = time.Time{}
	assert.Equal(t, event, retEvent)
	// Test Get Event with wrong project id.
	retEvent, errCode = M.GetEvent(projectId+1, userId, event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)
	// Test Get Event with wrong user id.
	retEvent, errCode = M.GetEvent(projectId, "randomId", event.ID)
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)

	// Test Get Event on non existent id.
	retEvent, errCode = M.GetEvent(projectId, userId, "random_id")
	assert.Equal(t, http.StatusNotFound, errCode)
	assert.Nil(t, retEvent)

	// Test Create Event with id.
	event, errCode = M.CreateEvent(&M.Event{ID: "random_id", EventName: eventName, ProjectId: projectId, UserId: userId})
	assert.Equal(t, http.StatusBadRequest, errCode)
	assert.Nil(t, event)

	// Test Create Event without projectId.
	event, errCode = M.CreateEvent(&M.Event{EventName: eventName, UserId: userId})
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, event)

	// Test Create Event without userId.
	event, errCode = M.CreateEvent(&M.Event{EventName: eventName, ProjectId: projectId})
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, event)

	// Test Create Event without eventName.
	event, errCode = M.CreateEvent(&M.Event{EventName: "", ProjectId: projectId, UserId: userId})
	assert.Equal(t, http.StatusInternalServerError, errCode)
	assert.Nil(t, event)
}
