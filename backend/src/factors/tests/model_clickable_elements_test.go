package tests

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
)

func TestCreateButtonClickHandler(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createButtonClick(t, project)
}

func createButtonClick(t *testing.T, project *model.Project) {
	buttonClick := &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
		},
	}
	status, err := store.GetStore().CreateClickableElementById(0, buttonClick)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	buttonClick = &model.CaptureClickPayload{
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
		},
	}
	status, err = store.GetStore().CreateClickableElementById(project.ID, buttonClick)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	buttonClick = &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
		},
	}
	status, err = store.GetStore().CreateClickableElementById(project.ID, buttonClick)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	status, err = store.GetStore().CreateClickableElementById(project.ID, buttonClick)
	assert.Equal(t, http.StatusConflict, status)
	assert.NotNil(t, err)
}
