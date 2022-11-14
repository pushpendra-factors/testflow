package tests

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	H "factors/handler"
	"factors/model/model"
	"factors/model/store"
)

func TestCreateFormFillHandler(t *testing.T) {
	r := gin.Default()
	H.InitSDKServiceRoutes(r)

	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)

	createFormFill(t, project)
}

func createFormFill(t *testing.T, project *model.Project) {
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

	// Check if form fill created
	formFill = &model.SDKFormFillPayload{
		FormId:  "formId1",
		Value:   "value1",
		FieldId: "fieldId1",
	}
	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusCreated, status)
	assert.Nil(t, err)

	// Check for duplicate case.
	formFill = &model.SDKFormFillPayload{
		FormId:  "formId1",
		Value:   "value1",
		FieldId: "fieldId1",
	}
	status, err = store.GetStore().CreateFormFillEventById(project.ID, formFill)
	assert.Equal(t, http.StatusConflict, status)
	assert.Nil(t, err)
}
