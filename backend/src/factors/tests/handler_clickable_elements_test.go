package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
)

func TestSDKCaptureButtonClickAsEventHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/capture_click"

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

	// Test button click - create new button click event
	payload := &model.SDKButtonElementAttributesPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: model.SDKButtonElementAttributes{
			DisplayText: "Submit-1",
			Timestamp:   time.Now().Unix(),
		},
	}
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if button click created.
	event, status := store.GetStore().GetButtonClickEventById(project.ID, payload.DisplayName, payload.ElementType)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	var elementAttributes model.SDKButtonElementAttributes
	error := U.DecodePostgresJsonbToStructType(event.ElementAttributes, &elementAttributes)
	assert.Nil(t, error)

	assert.Equal(t, project.ID, event.ProjectID)
	assert.NotEqual(t, "", event.Id)
	assert.Equal(t, payload.DisplayName, event.DisplayName)
	assert.Equal(t, payload.ElementType, event.ElementType)
	assert.Equal(t, payload.ElementAttributes.DisplayText, elementAttributes.DisplayText)
	assert.Equal(t, payload.ElementAttributes.Class, elementAttributes.Class)
	assert.Equal(t, payload.ElementAttributes.Id, elementAttributes.Id)
	assert.Equal(t, payload.ElementAttributes.Rel, elementAttributes.Rel)
	assert.Equal(t, payload.ElementAttributes.Role, elementAttributes.Role)
	assert.Equal(t, payload.ElementAttributes.Target, elementAttributes.Target)
	assert.Equal(t, payload.ElementAttributes.Href, elementAttributes.Href)
	assert.Equal(t, payload.ElementAttributes.Media, elementAttributes.Media)
	assert.Equal(t, payload.ElementAttributes.Type, elementAttributes.Type)
	assert.Equal(t, payload.ElementAttributes.Name, elementAttributes.Name)
	assert.Equal(t, payload.ElementAttributes.Timestamp, elementAttributes.Timestamp)
	assert.NotZero(t, event.ClickCount)
	assert.False(t, event.Enabled)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())
	assert.True(t, event.UpdatedAt.Equal(event.CreatedAt))

	// Test button click - update button click event
	payload = &model.SDKButtonElementAttributesPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: model.SDKButtonElementAttributes{
			DisplayText: "Submit-1",
		},
	}
	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, []byte(payloadBytes), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusAccepted, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if button click updated.
	event, status = store.GetStore().GetButtonClickEventById(project.ID, payload.DisplayName, payload.ElementType)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	error = U.DecodePostgresJsonbToStructType(event.ElementAttributes, &elementAttributes)
	assert.Nil(t, error)

	assert.Equal(t, project.ID, event.ProjectID)
	assert.NotEqual(t, "", event.Id)
	assert.Equal(t, payload.DisplayName, event.DisplayName)
	assert.Equal(t, payload.ElementType, event.ElementType)
	assert.Equal(t, payload.ElementAttributes.DisplayText, elementAttributes.DisplayText)
	assert.Equal(t, payload.ElementAttributes.Class, elementAttributes.Class)
	assert.Equal(t, payload.ElementAttributes.Id, elementAttributes.Id)
	assert.Equal(t, payload.ElementAttributes.Rel, elementAttributes.Rel)
	assert.Equal(t, payload.ElementAttributes.Role, elementAttributes.Role)
	assert.Equal(t, payload.ElementAttributes.Target, elementAttributes.Target)
	assert.Equal(t, payload.ElementAttributes.Href, elementAttributes.Href)
	assert.Equal(t, payload.ElementAttributes.Media, elementAttributes.Media)
	assert.Equal(t, payload.ElementAttributes.Type, elementAttributes.Type)
	assert.Equal(t, payload.ElementAttributes.Name, elementAttributes.Name)
	assert.NotZero(t, event.ClickCount)
	assert.False(t, event.Enabled)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())
	assert.True(t, event.UpdatedAt.After(event.CreatedAt))
}
