package tests

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	SDK "factors/sdk"
	T "factors/task"

	"factors/model/model"
	"factors/model/store"

	U "factors/util"
)

func TestTaskFormFillProcessing1(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

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

	// Form fill with current time.
	userId := U.GetUUID()
	timeNow := time.Now().UTC()
	formFill = &model.SDKFormFillPayload{
		FormId:    "formId2",
		FieldId:   "fieldId2",
		UserId:    userId,
		Value:     "email@gmail.com",
		UpdatedAt: &(timeNow),
	}
	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Check fetching of enabled projectIDs
	pwt, errCode := store.GetStore().GetFormFillEnabledProjectIDWithToken()
	projectIDs := U.GetKeysOfInt64StringMap(pwt)
	assert.Equal(t, errCode, http.StatusFound)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.Empty(t, rows)
	assert.Nil(t, err)

	project1, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Fetching initial enabled projectIDs
	pwt, errCode = store.GetStore().GetFormFillEnabledProjectIDWithToken()
	projectIDs = U.GetKeysOfInt64StringMap(pwt)
	assert.Equal(t, errCode, http.StatusFound)

	userId1 := U.GetUUID()
	userEmail1 := "email1@gmail.com"
	timeBeforeOneDay := time.Now().AddDate(0, 0, -1)
	// Check if form fill created
	formFill = &model.SDKFormFillPayload{
		FormId:    "formId1",
		FieldId:   "fieldId1",
		UserId:    userId1,
		Value:     userEmail1,
		UpdatedAt: &(timeBeforeOneDay),
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
	pwt, errCode = store.GetStore().GetFormFillEnabledProjectIDWithToken()
	projectIDs = U.GetKeysOfInt64StringMap(pwt)
	assert.Equal(t, errCode, http.StatusFound)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.NotEmpty(t, rows)
	assert.Equal(t, int(1), len(rows))
	assert.Nil(t, err)

	startTime := U.UnixTimeBeforeAWeek()

	// Check if form fill created with different fields
	project2, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Check if form fill created with different updated_at.
	userId2 := U.GetUUID()
	userEmail2 := "EMAIL2@gmail.com"
	timeBeforeTwoDay := time.Now().AddDate(0, 0, -2)
	formFill = &model.SDKFormFillPayload{
		FormId:    "formId3",
		FieldId:   "fieldId3",
		UserId:    userId2,
		Value:     userEmail2,
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
	pwt, errCode = store.GetStore().GetFormFillEnabledProjectIDWithToken()
	projectIDs = U.GetKeysOfInt64StringMap(pwt)
	assert.Equal(t, errCode, http.StatusFound)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes.
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.NotEmpty(t, rows)
	assert.Equal(t, int(2), len(rows))
	assert.Nil(t, err)

	// Process all form fills.
	errCode = T.FormFillProcessing()
	assert.Equal(t, errCode, http.StatusOK)

	// Check if form fills deleted after processing.
	pwt, errCode = store.GetStore().GetFormFillEnabledProjectIDWithToken()
	projectIDs = U.GetKeysOfInt64StringMap(pwt)
	assert.Equal(t, errCode, http.StatusFound)
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.Len(t, rows, 0)

	// Check $form_fill event for project1.
	eventName1, errCode := store.GetStore().GetEventName(U.EVENT_NAME_FORM_FILL, project1.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, eventName1)
	events1, errCode := store.GetStore().GetEventsByEventNameId(project1.ID, eventName1.ID, startTime, U.TimeNowUnix())
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, int(1), len(events1))
	// Check email identification for project1 userId1.
	user1, errCode := store.GetStore().GetUser(project1.ID, userId1)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, userEmail1, user1.CustomerUserId)

	// Check $form_fill event for project2.
	eventName2, errCode := store.GetStore().GetEventName(U.EVENT_NAME_FORM_FILL, project2.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, eventName2)
	events2, errCode := store.GetStore().GetEventsByEventNameId(project2.ID, eventName2.ID, startTime, U.TimeNowUnix())
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, int(1), len(events2))
	// Check email identification for project2 userId2.
	user2, errCode := store.GetStore().GetUser(project2.ID, userId2)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, strings.ToLower(userEmail2), user2.CustomerUserId)

	// Cleanup: Delete unprocessed form-fill for next run tests.
	status, err = store.GetStore().DeleteFormFillProcessedRecords(project.ID, userId, "formId2", "fieldId2")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, status)
}

func TestTaskFormFillProcessingAfterFormSubmit1(t *testing.T) {
	// Check if form fill created with different fields
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	userId := U.GetUUID()
	userEmail1 := "email1@gmail.com"

	timeBeforeThreeDays := time.Now().AddDate(0, 0, -3).Unix()
	properties := U.PropertiesMap{U.UP_EMAIL: userEmail1}
	trackPayload := &SDK.TrackPayload{
		Name:            U.EVENT_NAME_FORM_SUBMITTED,
		Timestamp:       timeBeforeThreeDays,
		EventProperties: properties,
		UserId:          userId,
		RequestSource:   model.UserSourceWeb,
	}

	status, response := SDK.Track(project.ID, trackPayload, false, SDK.SourceJSSDK, "")
	assert.NotNil(t, response.EventId)
	assert.Equal(t, http.StatusOK, status)

	// Verify identification from $form_submitted event.
	user, status := store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, userEmail1, user.CustomerUserId)
	propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&user.Properties)
	assert.Nil(t, err)
	assert.Equal(t, userEmail1, (*propertiesMap)[U.UP_USER_ID])
	assert.Equal(t, userEmail1, (*propertiesMap)[U.UP_EMAIL])

	// Check if form fill created with different updated_at.
	userEmail2 := "email2@gmail.com"
	timeBeforeTwoDay := time.Now().AddDate(0, 0, -2)
	formFill := &model.SDKFormFillPayload{
		FormId:    "formId3",
		FieldId:   "fieldId3",
		UserId:    userId,
		Value:     userEmail2,
		UpdatedAt: &(timeBeforeTwoDay),
	}

	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	pwt, errCode := store.GetStore().GetFormFillEnabledProjectIDWithToken()
	projectIDs := U.GetKeysOfInt64StringMap(pwt)
	assert.Equal(t, errCode, http.StatusFound)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes.
	rows, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIDs)
	assert.NotEmpty(t, rows)
	assert.Equal(t, int(1), len(rows))
	assert.Nil(t, err)

	// Process all form fills.
	errCode = T.FormFillProcessing()
	assert.Equal(t, errCode, http.StatusOK)

	// Check $form_fill event for project.
	eventName, errCode := store.GetStore().GetEventName(U.EVENT_NAME_FORM_FILL, project.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, eventName)
	events, errCode := store.GetStore().GetEventsByEventNameId(project.ID, eventName.ID, timeBeforeThreeDays, U.TimeNowUnix())
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, int(1), len(events))
	user, errCode = store.GetStore().GetUser(project.ID, userId)
	assert.Equal(t, http.StatusFound, errCode)
	// Customer User Id should be the one from form_submitted.
	assert.Equal(t, userEmail1, user.CustomerUserId)
}
