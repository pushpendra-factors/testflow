package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	T "factors/task"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"

	U "factors/util"
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
		FormId:  "formId1",
		Value:   "value1",
		FieldId: "fieldId1",
	}
	status, err := store.GetStore().CreateFormFillEventById(0, formFill)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	// Empty/invalid formId
	formFill = &model.SDKFormFillPayload{
		FormId:  "",
		Value:   "value1",
		FieldId: "fieldId1",
	}
	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	// Empty/invalid fieldId
	formFill = &model.SDKFormFillPayload{
		FormId:  "formId1",
		Value:   "value1",
		FieldId: "",
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

	timeBeforeOneDay := time.Now().AddDate(0, 0, -1)
	// Check if form fill created
	formFill = &model.SDKFormFillPayload{
		FormId:    "formId1",
		Value:     "abc@gmail.com",
		FieldId:   "fieldId1",
		UserId:    "userId1",
		UpdatedAt: &(timeBeforeOneDay),
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

	startTime := U.UnixTimeBeforeAWeek()

	// Check if form fill created with different fields
	project1, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, errCode)

	timeNow := time.Now().UTC()
	// Creating new record
	formFill = &model.SDKFormFillPayload{
		FormId:    "formId2",
		Value:     "xyz@gmail.com",
		FieldId:   "fieldId2",
		UserId:    "userId2",
		UpdatedAt: &(timeNow),
	}

	status, err = store.GetStore().CreateFormFillEventById(project1.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Enabling AutoCaptureFormFills for the created record
	_, errCode = store.GetStore().UpdateProjectSettings(project1.ID, &model.ProjectSetting{
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
	project2, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check if form fill created with different updated_at
	timeBeforeTwoDay := time.Now().AddDate(0, 0, -2)

	formFill = &model.SDKFormFillPayload{
		FormId:    "formId3",
		Value:     "mno@gmail.com",
		FieldId:   "fieldId3",
		UserId:    "userId3",
		UpdatedAt: &(timeBeforeTwoDay),
	}

	status, err = store.GetStore().CreateFormFillEventById(project2.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	_, errCode = store.GetStore().UpdateProjectSettings(project2.ID, &model.ProjectSetting{
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
	eventName, errCode := store.GetStore().GetEventName(U.EVENT_NAME_FORM_FILL, project.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, eventName)
	recordsCreated, errCode := store.GetStore().GetEventsByEventNameId(project.ID, eventName.ID, startTime, U.TimeNowUnix())
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, int(1), len(recordsCreated))
}
