package handler

import (
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ShareableURLParams struct {
	EntityID   int64 `json:"entity_id"`
	EntityType int   `json:"entity_type"`
	ShareType  int   `json:"share_type"`
	// AllowedUsers    string `json:"allowed_users"`
	IsExpirationSet bool  `json:"is_expiration_set"`
	ExpirationTime  int64 `json:"expiration_time"`
}

type UpdateShareableURLParams struct {
	ShareType int `json:"share_type"`
	// AllowedUsers string `json:"allowed_users"`
}

func validateCreateShareableURLRequest(params *model.ShareableURL, projectID uint64, agentUUID string) (bool, string) {

	if params.EntityID == 0 {
		return false, "Invalid entity id."
	}

	if !model.ValidShareEntityTypes[params.EntityType] {
		return false, "Invalid entity type."
	}

	if !model.ValidShareTypes[params.ShareType] {
		return false, "Invalid share type."
	}

	if params.EntityType == model.ShareableURLEntityTypeQuery {
		query, err := store.GetStore().GetQueryWithQueryId(projectID, params.EntityID)
		if err != http.StatusFound {
			return false, "Invalid query id."
		}

		_, err = store.GetStore().GetShareableURLWithShareStringAndAgentID(projectID, query.IdText, agentUUID)
		if err == http.StatusBadRequest || err == http.StatusInternalServerError {
			return false, "Shareable query creation failed. DB error."
		} else if err == http.StatusFound {
			return false, "Shareable url already exists."
		}

		params.QueryID = query.IdText

	} else if params.EntityType == model.ShareableURLEntityTypeDashboard {

	} else if params.EntityType == model.ShareableURLEntityTypeTemplate {

	}
	return true, ""
}

func GetShareableURLsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": agentUUID,
		"projectId":     projectID,
	})

	shareableURLs, errCode := store.GetStore().GetAllShareableURLsWithProjectIDAndAgentID(projectID, agentUUID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get shareable urls")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get shareable urls."})
		return
	}

	c.JSON(http.StatusOK, shareableURLs)
}

func CreateShareableURLHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Shareable query creation failed. Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": agentUUID,
		"projectId":     projectID,
	})

	params := ShareableURLParams{}
	err := c.BindJSON(&params)
	if err != nil {
		logCtx.WithError(err).Error("Failed to parse request body")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Shareable query creation failed. Invalid params."})
		return
	}

	shareableUrlRequest := &model.ShareableURL{
		EntityType: params.EntityType,
		ShareType:  params.ShareType,
		// AllowedUsers: params.AllowedUsers,
		EntityID:  params.EntityID,
		ProjectID: projectID,
		CreatedBy: agentUUID,
	}

	if params.IsExpirationSet && params.ExpirationTime > time.Now().Unix() {
		shareableUrlRequest.ExpiresAt = params.ExpirationTime
	} else {
		shareableUrlRequest.ExpiresAt = time.Now().AddDate(0, 1, 0).Unix()
	}

	valid, errMsg := validateCreateShareableURLRequest(shareableUrlRequest, projectID, agentUUID)
	if !valid {
		logCtx.Error(errMsg)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	share, errCode := store.GetStore().CreateShareableURL(shareableUrlRequest)
	if errCode != http.StatusCreated {
		logCtx.WithError(err).Error("Failed to create shareable query")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Shareable query creation failed."})
		return
	}

	c.JSON(http.StatusCreated, share)
}

// func UpdateShareableURLHandler(c *gin.Context) {

// 	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
// 	if projectID == 0 {
// 		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid project."})
// 		return
// 	}

// 	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

// 	logCtx := log.WithFields(log.Fields{
// 		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
// 		"loggedInAgent": agentUUID,
// 		"projectId":     projectID,
// 	})

// 	shareId := c.Param("share_id")
// 	if shareId == "" {
// 		logCtx.Error("Invalid share id")
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid share id."})
// 		return
// 	}

// 	params := UpdateShareableURLParams{}
// 	err := c.BindJSON(&params)
// 	if err != nil {
// 		logCtx.WithError(err).Error("Failed to parse request body")
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid params."})
// 		return
// 	}

// 	if !model.ValidShareTypes[params.shareType] {
// 		logCtx.Error("Invalid share type")
// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid share type."})
// 		return
// 	}

// 	errCode := store.GetStore().UpdateShareableURLShareTypeWithShareIDandCreatedBy(shareId, agentUUID, params.ShareType)
// 	if errCode != http.StatusAccepted {
// 		logCtx.Error("Failed to update shareable query")
// 		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Successfully updated."})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Shareable query updated successfully."})
// }

func DeleteShareableURLHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": agentUUID,
		"projectId":     projectID,
	})

	shareId := c.Param("share_id")
	if shareId == "" {
		logCtx.Error("Invalid share id")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid share id."})
		return
	}

	errCode := store.GetStore().DeleteShareableURLWithShareIDandAgentID(projectID, shareId, agentUUID)
	if errCode == http.StatusNotFound {
		logCtx.Error("Shareable query not found")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Shareable query not found."})
		return
	} else if errCode != http.StatusAccepted {
		logCtx.Error("Failed to delete shareable query")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Share able query delete failed."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted"})
}

func RevokeShareableURLHandler(c *gin.Context) {

	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid project."})
		return
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{
		"reqId":         U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"loggedInAgent": agentUUID,
		"projectId":     projectID,
	})

	mapping, err := store.GetStore().GetProjectAgentMapping(projectID, agentUUID)
	if err != http.StatusFound {
		logCtx.Error("Failed to get project agent mapping")
		c.AbortWithStatusJSON(err, gin.H{"error": "Failed to get project agent mapping."})
		return
	}
	if mapping.Role != model.ADMIN {
		logCtx.Error("Not authorized")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Not authorized."})
		return
	}

	shareString := c.Param("query_id")
	if shareString == "" {
		logCtx.Error("Invalid share id")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid share id."})
		return
	}

	var errMsg string
	if shareString == "all" {
		err, errMsg = store.GetStore().RevokeShareableURLsWithProjectID(projectID)
	} else {
		err, errMsg = store.GetStore().RevokeShareableURLsWithShareString(projectID, shareString)
	}
	if err != http.StatusAccepted {
		logCtx.Error(errMsg)
		c.AbortWithStatusJSON(err, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully revoked"})
}
