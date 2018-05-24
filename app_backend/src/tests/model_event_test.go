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
	project_id, user_id, event_name, err := SetupProjectUserEventName()
	assert.Nil(t, err)

	start := time.Now()

	// Test successful CreateEvent.
	event, err_code := M.CreateEvent(&M.Event{EventName: event_name, ProjectId: project_id, UserId: user_id})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.True(t, len(event.ID) > 30)
	assert.Equal(t, project_id, event.ProjectId)
	assert.Equal(t, event_name, event.EventName)
	assert.True(t, event.CreatedAt.After(start))
	assert.True(t, event.UpdatedAt.After(start))
	assert.Equal(t, event.CreatedAt, event.UpdatedAt)
	assert.Equal(t, postgres.Jsonb{RawMessage: json.RawMessage(nil)}, event.Properties)
	// Test Get Project on the created.
	get_event, err_code := M.GetEvent(event.ID)
	assert.Equal(t, M.DB_SUCCESS, err_code)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(event.CreatedAt.Sub(get_event.CreatedAt).Seconds()) < 0.1)
	assert.True(t, math.Abs(event.UpdatedAt.Sub(get_event.UpdatedAt).Seconds()) < 0.1)
	event.CreatedAt = time.Time{}
	event.UpdatedAt = time.Time{}
	get_event.CreatedAt = time.Time{}
	get_event.UpdatedAt = time.Time{}
	assert.Equal(t, event, get_event)

	// Test Get Event on non existent id.
	get_event, err_code = M.GetEvent("random_id")
	assert.Equal(t, http.StatusNotFound, err_code)
	assert.Nil(t, get_event)

	// Test Create Event with id.
	event, err_code = M.CreateEvent(&M.Event{ID: "random_id", EventName: event_name, ProjectId: project_id, UserId: user_id})
	assert.Equal(t, http.StatusBadRequest, err_code)
	assert.Nil(t, event)

	// Test Create Event without project_id.
	event, err_code = M.CreateEvent(&M.Event{EventName: event_name, UserId: user_id})
	assert.Equal(t, http.StatusInternalServerError, err_code)
	assert.Nil(t, event)

	// Test Create Event without user_id.
	event, err_code = M.CreateEvent(&M.Event{EventName: event_name, ProjectId: project_id})
	assert.Equal(t, http.StatusInternalServerError, err_code)
	assert.Nil(t, event)

	// Test Create Event without event_name.
	event, err_code = M.CreateEvent(&M.Event{EventName: "", ProjectId: project_id, UserId: user_id})
	assert.Equal(t, http.StatusInternalServerError, err_code)
	assert.Nil(t, event)
}
