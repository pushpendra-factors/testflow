package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	U "factors/util"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"

	T "factors/task"
)

func TestFormFillHandlerSupport(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createFormFillSupport(t, project)
}

func createFormFillSupport(t *testing.T, project *model.Project) {
	// Empty/invalid projectID
	formFill := &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId1",
	}
	status, err := store.GetStore().CreateFormFillEventById(0, formFill)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	// Empty/invalid formId
	formFill = &model.SDKFormFillPayload{
		FormId:           "",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId1",
	}
	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	// Empty/invalid fieldId
	formFill = &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "",
	}

	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	// Fetching initial enabled projectIDs
	projectIDs, err := store.GetStore().GetFormFillEnabledProjectIDs()
	assert.Nil(t, err)

	// Fetching initial enabled projectIDs updated before 10 minutes
	initialRows, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.Nil(t, err)

	// Check if form fill created
	formFill = &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId1",
	}

	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Enabling AutoCaptureFormFills for the created record
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})

	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIDs, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotEmpty(t, projectIDs)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(1), len(rows))
	assert.Nil(t, err)

	startTime := U.UnixTimeBeforeDuration(15 * time.Minute)
	// Check if form fill created with same fields but different value
	formFill = &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value_2",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(15 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(15 * time.Minute),
		FieldId:          "fieldId1",
	}

	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusConflict, status)
	assert.Nil(t, err)

	// Enabling AutoCaptureFormFills for the created record
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})

	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIDs, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotEmpty(t, projectIDs)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(1), len(rows))
	assert.Nil(t, err)

	// Check if form fill created with different fields
	project, err = SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, errCode)

	formFill = &model.SDKFormFillPayload{
		FormId:           "formId2",
		Value:            "value2",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId2",
	}

	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})

	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIDs, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotEmpty(t, projectIDs)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(2), len(rows))
	assert.Nil(t, err)

	// Check for processing of FormFillProcessing
	errCode = T.FormFillProcessing()
	assert.Equal(t, errCode, http.StatusOK)

	// Check if records were deleted after processing
	projectIDs, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIDs)
	assert.Nil(t, err)
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.Greater(t, len(initialRows)+int(2), len(rows))
	assert.Nil(t, err)

	// Check if record created in events table
	eventName, errCode := store.GetStore().GetEventName("$form_fill", project.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, eventName)
	recordsCreated, errCode := store.GetStore().GetEventsByEventNameId(project.ID, eventName.ID, startTime, formFill.LastUpdatedTime)
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, int(1), len(recordsCreated))
}
