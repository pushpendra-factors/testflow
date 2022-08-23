package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
)

func TestSDKCaptureButtonClickAsEventHandler(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	uri := "/sdk/capture_click"

	project, user, err := SetupProjectUserReturnDAO()
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
	payload := &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "BUTTON",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
			"element_type": "BUTTON",
			"class":        "style-1",
			"id":           "id1",
			"rel":          "rel1",
			"role":         "role1",
			"target":       "target1",
			"href":         "http://href1.com",
			"media":        "media1",
			"type":         "type1",
			"name":         "name1",

			"not_allowed": "not_allowed1",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if button click created.
	event, status := store.GetStore().GetClickableElement(project.ID, payload.DisplayName, payload.ElementType)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	elementAttributes, err := U.DecodePostgresJsonbAsPropertiesMap(event.ElementAttributes)
	assert.Nil(t, err)

	assert.Equal(t, project.ID, event.ProjectID)
	assert.NotEqual(t, "", event.Id)
	assert.Equal(t, payload.DisplayName, event.DisplayName)
	assert.Equal(t, payload.ElementType, event.ElementType)

	assert.Equal(t, payload.ElementAttributes["display_text"], (*elementAttributes)["display_text"])
	assert.Equal(t, payload.ElementAttributes["class"], (*elementAttributes)["class"])
	assert.Equal(t, payload.ElementAttributes["id"], (*elementAttributes)["id"])
	assert.Equal(t, payload.ElementAttributes["rel"], (*elementAttributes)["rel"])
	assert.Equal(t, payload.ElementAttributes["role"], (*elementAttributes)["role"])
	assert.Equal(t, payload.ElementAttributes["target"], (*elementAttributes)["target"])
	assert.Equal(t, payload.ElementAttributes["href"], (*elementAttributes)["href"])
	assert.Equal(t, payload.ElementAttributes["media"], (*elementAttributes)["media"])
	assert.Equal(t, payload.ElementAttributes["type"], (*elementAttributes)["type"])
	assert.Equal(t, payload.ElementAttributes["name"], (*elementAttributes)["name"])
	assert.Nil(t, (*elementAttributes)["not_allowed"])

	assert.NotZero(t, event.ClickCount)
	clickCount := event.ClickCount
	assert.False(t, event.Enabled)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())
	assert.True(t, event.UpdatedAt.Equal(event.CreatedAt))

	// Test button click - update button click event
	payload = &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "BUTTON",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
			"element_type": "BUTTON",
			"class":        "style-1",
			"id":           "id1",
			"rel":          "rel1",
			"role":         "role1",
			"target":       "target1",
			"href":         "http://href1.com",
			"media":        "media1",
			"type":         "type1",
			"name":         "name1",
		},
	}
	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, []byte(payloadBytes), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusAccepted, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])

	// Check if button click updated.
	event, status = store.GetStore().GetClickableElement(project.ID, payload.DisplayName, payload.ElementType)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, event)

	elementAttributes, err = U.DecodePostgresJsonbAsPropertiesMap(event.ElementAttributes)
	assert.Nil(t, err)

	assert.Equal(t, project.ID, event.ProjectID)
	assert.NotEqual(t, "", event.Id)
	assert.Equal(t, payload.DisplayName, event.DisplayName)
	assert.Equal(t, payload.ElementType, event.ElementType)

	assert.Equal(t, payload.ElementAttributes["display_text"], (*elementAttributes)["display_text"])
	assert.Equal(t, payload.ElementAttributes["class"], (*elementAttributes)["class"])
	assert.Equal(t, payload.ElementAttributes["id"], (*elementAttributes)["id"])
	assert.Equal(t, payload.ElementAttributes["rel"], (*elementAttributes)["rel"])
	assert.Equal(t, payload.ElementAttributes["role"], (*elementAttributes)["role"])
	assert.Equal(t, payload.ElementAttributes["target"], (*elementAttributes)["target"])
	assert.Equal(t, payload.ElementAttributes["href"], (*elementAttributes)["href"])
	assert.Equal(t, payload.ElementAttributes["media"], (*elementAttributes)["media"])
	assert.Equal(t, payload.ElementAttributes["type"], (*elementAttributes)["type"])
	assert.Equal(t, payload.ElementAttributes["name"], (*elementAttributes)["name"])

	// Button with same name and type should increment count by 1.
	assert.True(t, event.ClickCount > clickCount)
	assert.False(t, event.Enabled)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())

	// Enable element for capturing.
	status = store.GetStore().ToggleEnabledClickableElement(project.ID, event.Id)
	assert.Equal(t, http.StatusAccepted, status)

	// Test button click_capture after enabling. Track.
	payload = &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "BUTTON",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
			"element_type": "BUTTON",
			"class":        "style2",
			"id":           "id2",
			"rel":          "rel2",
			"role":         "role2",
			"target":       "target2",
			"href":         "http://href2.com",
			"media":        "media2",
			"type":         "type2",
			"name":         "name2",

			"not_allowed": "not_allowed1",
		},

		UserID: user.ID,
		EventProperties: U.PropertiesMap{
			U.EP_PAGE_URL: "http://example.com",
			U.EP_REFERRER: "https://google.com",
		},
		UserProperties: U.PropertiesMap{
			U.UP_SCREEN_WIDTH:  80,
			U.UP_SCREEN_HEIGHT: 100,
		},
	}
	payloadBytes, err = json.Marshal(payload)
	assert.Nil(t, err)

	w = ServePostRequestWithHeaders(r, uri, []byte(payloadBytes), map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusOK, w.Code)
	responseMap = DecodeJSONResponseToMap(w.Body)
	assert.NotNil(t, responseMap["message"])
	assert.NotNil(t, responseMap["event_id"])

	clickEvent, status := store.GetStore().GetEventById(project.ID, responseMap["event_id"].(string), user.ID)
	assert.Equal(t, http.StatusFound, status)
	assert.NotNil(t, clickEvent)

	clickEventProperties, err := U.DecodePostgresJsonbAsPropertiesMap(&clickEvent.Properties)
	assert.Nil(t, err)

	assert.Equal(t, payload.ElementAttributes["display_text"], (*clickEventProperties)["display_text"])
	assert.Equal(t, payload.ElementAttributes["element_type"], (*clickEventProperties)["element_type"])
	assert.Equal(t, payload.ElementAttributes["class"], (*clickEventProperties)["class"])
	assert.Equal(t, payload.ElementAttributes["id"], (*clickEventProperties)["id"])
	assert.Equal(t, payload.ElementAttributes["rel"], (*clickEventProperties)["rel"])
	assert.Equal(t, payload.ElementAttributes["role"], (*clickEventProperties)["role"])
	assert.Equal(t, payload.ElementAttributes["target"], (*clickEventProperties)["target"])
	assert.Equal(t, payload.ElementAttributes["href"], (*clickEventProperties)["href"])
	assert.Equal(t, payload.ElementAttributes["media"], (*clickEventProperties)["media"])
	assert.Equal(t, payload.ElementAttributes["type"], (*clickEventProperties)["type"])
	assert.Equal(t, payload.ElementAttributes["name"], (*clickEventProperties)["name"])
	assert.Nil(t, (*elementAttributes)["not_allowed"])

	eventName, err := store.GetStore().GetEventNameFromEventNameId(clickEvent.EventNameId, project.ID)
	assert.Nil(t, err)
	assert.Equal(t, payload.DisplayName, eventName.Name)
}

func sendGetClickableElementsRequest(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/clickable_elements", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error getting clickable elements.")
		return nil
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendToggleClickableElementRequest(r *gin.Engine, projectId int64, id string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error creating cookie data.")
	}

	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/clickable_elements/%s/toggle", projectId, id)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Failed to toggle the clickable element.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestClickableElementsSettingsUIHandlers(t *testing.T) {
	// Initialize routes and dependent data.
	r := gin.Default()
	H.InitSDKServiceRoutes(r)
	H.InitAppRoutes(r)
	uri := "/sdk/capture_click"

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.NotNil(t, agent)

	// Element 1
	payload := &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "BUTTON",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
			"element_type": "BUTTON",
			"class":        "style-1",
			"id":           "id1",
			"rel":          "rel1",
			"role":         "role1",
			"target":       "target1",
			"href":         "http://href1.com",
			"media":        "media1",
			"type":         "type1",
			"name":         "name1",

			"not_allowed": "not_allowed1",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(t, err)
	w := ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)

	// Element 2
	payload1 := &model.CaptureClickPayload{
		DisplayName: "Submit-2",
		ElementType: "BUTTON",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-2",
			"element_type": "BUTTON",
			"class":        "style-1",
			"id":           "id1",
			"rel":          "rel1",
			"role":         "role1",
			"target":       "target1",
			"href":         "http://href1.com",
			"media":        "media1",
			"type":         "type1",
			"name":         "name1",

			"not_allowed": "not_allowed1",
		},
	}
	payloadBytes, err = json.Marshal(payload1)
	assert.Nil(t, err)
	w = ServePostRequestWithHeaders(r, uri, payloadBytes, map[string]string{"Authorization": project.Token})
	assert.Equal(t, http.StatusCreated, w.Code)

	// Send get clickable elements API request.
	w = sendGetClickableElementsRequest(r, project.ID, agent)
	assert.Equal(t, http.StatusFound, w.Code)
	var elements []model.ClickableElements
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse, &elements)
	assert.Nil(t, err)
	assert.Len(t, elements, 2)

	// Send toggle clickable elements API request for enabling.
	assert.False(t, elements[0].Enabled)
	assert.False(t, elements[1].Enabled)

	w = sendToggleClickableElementRequest(r, project.ID, elements[0].Id, agent)
	assert.Equal(t, http.StatusAccepted, w.Code)
	w = sendToggleClickableElementRequest(r, project.ID, elements[1].Id, agent)
	assert.Equal(t, http.StatusAccepted, w.Code)

	w = sendGetClickableElementsRequest(r, project.ID, agent)
	assert.Equal(t, http.StatusFound, w.Code)
	var elements1 []model.ClickableElements
	jsonResponse1, _ := ioutil.ReadAll(w.Body)
	err = json.Unmarshal(jsonResponse1, &elements1)
	assert.Nil(t, err)
	assert.Len(t, elements1, 2)

	assert.True(t, elements1[0].Enabled)
	assert.True(t, elements1[1].Enabled)
}
