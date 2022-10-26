package tests

import (
	"encoding/json"
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

func TestSDKFormFillEventHandlerSupport(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/form_fill"

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	// Test without project_id scope and with non-existing project.
	w := ServePostRequest(r, uri, []byte("{}"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap := DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test with invalid token.
	w = ServePostRequestWithHeaders(r, uri, []byte(`{}`), map[string]string{"Authorization": "INVALID_TOKEN"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Test with empty/invalid payload.
	w = ServePostRequestWithHeaders(r, uri, []byte(``), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["error"])

	// Fetching of initial enabled projectIDs
	projectIds, err := store.GetStore().GetFormFillEnabledProjectIDs()
	assert.Nil(t, err)

	// Fetching initial rows with enabled projectIDs updated before 10 minutes
	initialRows, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.Nil(t, err)

	// Test form fill event - create new form fill event.
	startTime := U.UnixTimeBeforeDuration(15 * time.Minute)
	payload := &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(15 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(15 * time.Minute),
		FieldId:          "fieldId1",
	}
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(t, err)
	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status := store.GetStore().GetFormFillEventById(project.ID, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	assert.Equal(t, project.ID, event.ProjectID)
	assert.NotEqual(t, "", event.Id)
	assert.Equal(t, payload.FormId, event.FormId)
	assert.Equal(t, payload.Value, event.Value)
	assert.Equal(t, payload.FieldId, event.FieldId)
	assert.Equal(t, uint64(100), event.TimeSpentOnField)
	assert.Equal(t, payload.FirstUpdatedTime, event.FirstUpdatedTime)
	assert.Equal(t, payload.LastUpdatedTime, event.LastUpdatedTime)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())

	// Check fetching of enabled projectIDs
	enable := true
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotEmpty(t, projectIds)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err := store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.NotNil(t, rows)
	assert.Equal(t, len(initialRows)+int(1), len(rows))
	assert.Nil(t, err)

	// Create new form fill event with same fields but different value
	payload = &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value_2",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId1",
	}
	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusConflict, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status = store.GetStore().GetFormFillEventById(project.ID, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	// Check fetching of enabled projectIDs
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIds)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(1), len(rows))
	assert.Nil(t, err)

	// Creating records for same formId but different fieldId
	payload = &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId2",
	}

	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status = store.GetStore().GetFormFillEventById(project.ID, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	// Check fetching of enabled projectIDs
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIds)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(2), len(rows))
	assert.Nil(t, err)

	// Checking case for same fieldId but different formId
	payload = &model.SDKFormFillPayload{
		FormId:           "formId2",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId1",
	}

	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status = store.GetStore().GetFormFillEventById(project.ID, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	// Check fetching of enabled projectIDs
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIds)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(3), len(rows))
	assert.Nil(t, err)

	// Checking case for same fieldId and formId but different projectId
	project, err = SetupProjectReturnDAO()
	assert.Nil(t, err)

	payload = &model.SDKFormFillPayload{
		FormId:           "formId1",
		Value:            "value1",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId1",
	}

	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status = store.GetStore().GetFormFillEventById(project.ID, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	// Check fetching of enabled projectIDs
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIds)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(4), len(rows))
	assert.Nil(t, err)

	// Create new form fill event with different fields
	project, err = SetupProjectReturnDAO()
	assert.Nil(t, err)

	payload = &model.SDKFormFillPayload{
		FormId:           "formId2",
		Value:            "value2",
		TimeSpent:        0,
		FirstUpdatedTime: U.UnixTimeBeforeDuration(10 * time.Minute),
		LastUpdatedTime:  U.UnixTimeBeforeDuration(10 * time.Minute),
		FieldId:          "fieldId2",
	}
	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status = store.GetStore().GetFormFillEventById(project.ID, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	// Check fetching of enabled projectIDs
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	enable = true
	_, errCode = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		AutoCaptureFormFills: &enable,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	// Check fetching of enabled projectIDs
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIds)
	assert.Nil(t, err)

	// Check fetching of rows with enabled projectIDs updated before 10 minutes
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.NotEmpty(t, rows)
	assert.Equal(t, len(initialRows)+int(5), len(rows))
	assert.Nil(t, err)

	// Check for processing of FormFillProcessing
	errCode = T.FormFillProcessing()
	assert.Equal(t, errCode, http.StatusOK)

	// Check if records were deleted after processing
	projectIds, err = store.GetStore().GetFormFillEnabledProjectIDs()
	assert.NotNil(t, projectIds)
	assert.Nil(t, err)
	rows, err = store.GetStore().GetFormFillEventsUpdatedBeforeTenMinutes(projectIds)
	assert.Greater(t, len(initialRows)+int(5), len(rows))
	assert.Nil(t, err)

	// Check if record created in events table
	eventName, errCode := store.GetStore().GetEventName("$form_fill", project.ID)
	assert.Equal(t, errCode, http.StatusFound)
	assert.NotEmpty(t, eventName)
	recordsCreated, errCode := store.GetStore().GetEventsByEventNameId(project.ID, eventName.ID, startTime, payload.LastUpdatedTime)
	assert.Equal(t, errCode, http.StatusFound)
	assert.Equal(t, int(1), len(recordsCreated))
}
