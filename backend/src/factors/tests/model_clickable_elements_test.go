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
	status, err := store.GetStore().CreateClickableElement(0, buttonClick)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	buttonClick = &model.CaptureClickPayload{
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
		},
	}
	status, err = store.GetStore().CreateClickableElement(project.ID, buttonClick)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotNil(t, err)

	buttonClick = &model.CaptureClickPayload{
		DisplayName: "Submit-1",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-1",
		},
	}
	status, err = store.GetStore().CreateClickableElement(project.ID, buttonClick)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	status, err = store.GetStore().CreateClickableElement(project.ID, buttonClick)
	assert.Equal(t, http.StatusConflict, status)
	assert.NotNil(t, err)
}

func TestDeleteClickableElementsOlderThanGivenDays(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	deleteClickableElementsOlderThanGivenDays(t, project)
}

func deleteClickableElementsOlderThanGivenDays(t *testing.T, project *model.Project) {
	timeBeforeEightDays := time.Now().AddDate(0, 0, -8)

	// creating a record with eight days old updated_at
	buttonClick := &model.CaptureClickPayload{
		DisplayName: "Submit-TEST",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-TEST",
		},
		UpdatedAt: &(timeBeforeEightDays),
	}
	status, err := store.GetStore().CreateClickableElement(project.ID, buttonClick)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// setting enabled as true
	clickableElement, status := store.GetStore().GetClickableElement(project.ID, buttonClick.DisplayName, buttonClick.ElementType)
	assert.Equal(t, http.StatusFound, status)
	status = store.GetStore().ToggleEnabledClickableElement(project.ID, clickableElement.Id)
	assert.Equal(t, http.StatusAccepted, status)

	// creating another record with enabled false (eight days old updated_at)
	buttonClick1 := &model.CaptureClickPayload{
		DisplayName: "Submit-TEST1",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-TEST1",
		},
		UpdatedAt: &(timeBeforeEightDays),
	}
	status, err = store.GetStore().CreateClickableElement(project.ID, buttonClick1)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// creating a record with one day old updated_at
	timeBeforeOneDay := time.Now().AddDate(0, 0, -1)
	buttonClick2 := &model.CaptureClickPayload{
		DisplayName: "Submit-TEST2",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-TEST2",
		},
		UpdatedAt: &(timeBeforeOneDay),
	}
	status, err = store.GetStore().CreateClickableElement(project.ID, buttonClick2)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// setting enabled as true
	clickableElement, status = store.GetStore().GetClickableElement(project.ID, buttonClick2.DisplayName, buttonClick2.ElementType)
	assert.Equal(t, http.StatusFound, status)
	status = store.GetStore().ToggleEnabledClickableElement(project.ID, clickableElement.Id)
	assert.Equal(t, http.StatusAccepted, status)

	// creating another record with enabled false (one day old updated_at)
	buttonClick3 := &model.CaptureClickPayload{
		DisplayName: "Submit-TEST3",
		ElementType: "Button",
		ElementAttributes: U.PropertiesMap{
			"display_text": "Submit-TEST3",
		},
		UpdatedAt: &(timeBeforeOneDay),
	}
	status, err = store.GetStore().CreateClickableElement(project.ID, buttonClick3)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// deleting 7 days older records
	status, err = store.GetStore().DeleteClickableElementsOlderThanGivenDays(7, project.ID, false)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, err)

	// check whether record 0 is not deleted
	_, status = store.GetStore().GetClickableElement(project.ID, buttonClick.DisplayName, buttonClick.ElementType)
	assert.Equal(t, http.StatusFound, status)

	// check whether record 1 is deleted
	_, status = store.GetStore().GetClickableElement(project.ID, buttonClick1.DisplayName, buttonClick1.ElementType)
	assert.Equal(t, http.StatusNotFound, status)

	// check whether record 2 is not deleted
	_, status = store.GetStore().GetClickableElement(project.ID, buttonClick2.DisplayName, buttonClick2.ElementType)
	assert.Equal(t, http.StatusFound, status)

	// check whether record 3 is not deleted
	_, status = store.GetStore().GetClickableElement(project.ID, buttonClick3.DisplayName, buttonClick3.ElementType)
	assert.Equal(t, http.StatusFound, status)

}
