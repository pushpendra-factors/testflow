package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func CreateContentGroupHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	r := c.Request

	var contentGroupPayload model.ContentGroup
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contentGroupPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on create content group handler.")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Failed to decode Json request on create content group handler.", true
	}

	contentGroupPayload.CreatedBy = U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	contentGroup, errCode, errMsg := store.GetStore().CreateContentGroup(projectId, contentGroupPayload)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": contentGroupPayload, "err-message": errMsg}).Error("Failed to insert content group on create create content group handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return contentGroup, http.StatusCreated, "", "", false
}

func UpdateContentGroupHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("UpdateContentGroup Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	contentGroupId := c.Params.ByName("id")
	if contentGroupId == "" {
		log.Error("UpdateContentGroup Failed. Id parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Id parse failed", true
	}

	r := c.Request
	var contentGroup model.ContentGroup
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contentGroup); err != nil {
		log.WithError(err).Error("Failed to decode Json request on update content handler.")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Failed to decode Json request on update content group handler.", true
	}

	_, errCode, errMsg := store.GetStore().UpdateContentGroup(contentGroupId, projectID, contentGroup)
	if errCode != http.StatusAccepted {
		log.WithFields(log.Fields{"document": contentGroup, "err-message": errMsg}).Error("Failed to insert content group on create create content group handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}
	contentGroup, errCode = store.GetStore().GetContentGroupById(contentGroupId, projectID)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get updated content group", true
	}
	return contentGroup, http.StatusOK, "", "", false
}

func GetContentGroupHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("Get ContentGroup Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	contentGroups, errCode := store.GetStore().GetAllContentGroups(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get content groups", true
	}
	return contentGroups, http.StatusOK, "", "", false
}

func GetContentGroupByIDHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetContentGroup Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	contentGroupId := c.Params.ByName("id")
	if contentGroupId == "" {
		log.Error("GetContentGroup Failed. Id parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Id parse failed", true
	}

	contentGroup, errCode := store.GetStore().GetContentGroupById(contentGroupId, projectID)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get content group", true
	}
	return contentGroup, http.StatusOK, "", "", false
}

func DeleteContentGroupHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("DeleteContentGroup Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	contentGroupId := c.Params.ByName("id")
	if contentGroupId == "" {
		log.Error("DeleteContentGroup Failed. Id parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "ID parse failed", true
	}

	errCode, errMsg := store.GetStore().DeleteContentGroup(contentGroupId, projectID)
	if errCode != http.StatusAccepted {
		log.Error("DeleteContentGroup Failed. Id parse failed." + errMsg)
		return nil, errCode, PROCESSING_FAILED, "Failed to delete content group", true
	}
	return nil, http.StatusOK, "", "", false
}
