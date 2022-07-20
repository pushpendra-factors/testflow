package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
)

func TestCreateButtonClickHandler(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createButtonClick(t, project)
}

func createButtonClick(t *testing.T, project *model.Project) {
	buttonClick := &model.SDKButtonElementAttributesPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: model.SDKButtonElementAttributes{
			DisplayText: "Submit-1",
			Timestamp:   time.Now().Unix(), // request timestamp.
		},
	}
	status, err := store.GetStore().CreateButtonClickEventById(0, buttonClick)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	buttonClick = &model.SDKButtonElementAttributesPayload{
		ElementAttributes: model.SDKButtonElementAttributes{
			DisplayText: "Submit-1",
			Timestamp:   time.Now().Unix(), // request timestamp.
		},
	}
	status, err = store.GetStore().CreateButtonClickEventById(project.ID, buttonClick)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	buttonClick = &model.SDKButtonElementAttributesPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: model.SDKButtonElementAttributes{
			DisplayText: "Submit-1",
			Timestamp:   time.Now().Unix(), // request timestamp.
		},
	}
	status, err = store.GetStore().CreateButtonClickEventById(project.ID, buttonClick)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	status, err = store.GetStore().CreateButtonClickEventById(project.ID, buttonClick)
	assert.Equal(t, http.StatusConflict, status)
	assert.NotNil(t, err)
}
