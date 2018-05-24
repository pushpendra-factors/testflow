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
	random_project_name := U.RandomLowerAphaNumString(15)
	project, err_code := M.CreateProject(&M.Project{Name: random_project_name})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.NotNil(t, project)
	project_id := project.ID

	start := time.Now()

	// Test successful create event_name.
	event_name, err_code := M.CreateEventName(&M.EventName{Name: "test_event", ProjectId: project_id})
	assert.Equal(t, M.DB_SUCCESS, err_code)
	assert.Equal(t, project_id, event_name.ProjectId)
	assert.True(t, event_name.CreatedAt.After(start))
	// Test Get Project on the created one.
	get_event_name, err_code := M.GetEventName(event_name.Name, project_id)
	assert.Equal(t, M.DB_SUCCESS, err_code)
	// time.Time is not exactly same. Checking within an error threshold.
	assert.True(t, math.Abs(event_name.CreatedAt.Sub(get_event_name.CreatedAt).Seconds()) < 0.1)
	event_name.CreatedAt = time.Time{}
	get_event_name.CreatedAt = time.Time{}
	assert.Equal(t, event_name, get_event_name)

	// Test Get Event on non existent name.
	get_event_name, err_code = M.GetEventName("non_existent_event", project_id)
	assert.Equal(t, http.StatusNotFound, err_code)
	assert.Nil(t, get_event_name)

	// Test Get Event with only name.
	get_event_name, err_code = M.GetEventName(event_name.Name, 0)
	assert.Equal(t, http.StatusBadRequest, err_code)
	assert.Nil(t, get_event_name)

	// Test Get Event with only project_id.
	get_event_name, err_code = M.GetEventName("", project_id)
	assert.Equal(t, http.StatusBadRequest, err_code)
	assert.Nil(t, get_event_name)
}
