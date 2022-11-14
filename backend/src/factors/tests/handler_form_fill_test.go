package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
)

func TestSDKFormFillEventHandler(t *testing.T) {
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

	// Test form fill event - create new form fill event.
	payload := &model.SDKFormFillPayload{
		FormId:  "formId1",
		Value:   "value1",
		FieldId: "fieldId1",
		UserId:  "1",
	}
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if form fill created.
	event, status := store.GetStore().GetFormFillEventById(project.ID, payload.UserId, payload.FormId, payload.FieldId)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	assert.Equal(t, project.ID, event.ProjectID)
	assert.NotEqual(t, "", event.Id)
	assert.Equal(t, payload.FormId, event.FormId)
	assert.Equal(t, payload.Value, event.Value)
	assert.Equal(t, payload.FieldId, event.FieldId)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())
	assert.True(t, event.UpdatedAt.Equal(event.CreatedAt))

	// No duplicates. The new row should be allowed.
	payload = &model.SDKFormFillPayload{
		FormId:  "formId1",
		Value:   "value1",
		FieldId: "fieldId1",
		UserId:  "1",
	}
	status, err = store.GetStore().CreateFormFillEventById(project.ID, payload)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)
}
