package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
)

// SDKCaptureButtonClickAsEvent godoc
// @Summary To capture button clicks as events.
// @Tags SDK
// @Accept  json
// @Produce json
// @Param request body SDKCaptureButtonClickAsEventPayload true "Capture button click payload"
// @Success 200 {object} sdk.SDKButtonElementAttributesResponse
// @Router /sdk/capture_click [post]
func SDKCaptureButtonClickAsEventHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	if r.Body == nil {
		logCtx.Error("Invalid request. Request body unavailable.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SDKButtonElementAttributesResponse{
			Error: "Updating Button click failed. Missing request body."})
		return
	}

	var request model.SDKButtonElementAttributesPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		logCtx.WithError(err).Error("Updating Button click failed. JSON Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, &model.SDKButtonElementAttributesResponse{
			Error: "Updating Button click failed. Invalid payload."})
		return
	}

	projectToken := U.GetScopeByKeyAsString(c, mid.SCOPE_PROJECT_TOKEN)
	projectID, errCode := store.GetStore().GetProjectIDByToken(projectToken)
	if errCode == http.StatusNotFound {
		if SDK.IsValidTokenString(projectToken) {
			logCtx.Error("Failed to get project from sdk project token.")
		} else {
			logCtx.WithField("token", projectToken).Warn("Invalid token on sdk payload.")
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, &model.SDKButtonElementAttributesResponse{Error: "Update button click failed. Invalid Token."})
		return
	}

	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, &model.SDKButtonElementAttributesResponse{Error: "Update button click failed."})
		return
	}

	var response *model.SDKButtonElementAttributesResponse

	status, err := store.GetStore().UpdateButtonClickEventById(projectID, &request)
	if err != nil {
		response = &model.SDKButtonElementAttributesResponse{Error: err.Error()}
	} else {
		response = &model.SDKButtonElementAttributesResponse{Message: "Updated button click successfully."}
	}

	c.JSON(status, response)
}
