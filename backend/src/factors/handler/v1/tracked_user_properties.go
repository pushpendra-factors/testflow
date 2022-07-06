package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetAllFactorsTrackedUserPropertiesHandler - Get All tracked user properties handler
// GetAllTrackedEventsHandler godoc
// @Summary Get all tracked user properties
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.FactorsTrackedUserProperty
// @Router /{project_id}/v1/factors/tracked_user_property [GET]
func GetAllFactorsTrackedUserPropertiesHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	trackedUserProperties, errCode := store.GetStore().GetAllFactorsTrackedUserPropertiesByProject(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, trackedUserProperties)
}

type CreateTrackeduserPropertyParams struct {
	UserPropertyName string `json:"user_property_name" binding:"required"`
}

func getcreateFactorsTrackedUserPropertyParams(c *gin.Context) (*CreateTrackeduserPropertyParams, error) {
	params := CreateTrackeduserPropertyParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

//CreateFactorsTrackedUserPropertyHandler - Handler for creating tracked user property
// CreateTrackedEventsHandler godoc
// @Summary Create a tracked user property
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param create body v1.CreateTrackeduserPropertyParams true "create"
// @Success 201 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/tracked_user_property [POST]
func CreateFactorsTrackedUserPropertyHandler(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	params, err := getcreateFactorsTrackedUserPropertyParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode := store.GetStore().CreateFactorsTrackedUserProperty(projectID, params.UserPropertyName, loggedInAgentUUID)
	if !(errCode == http.StatusCreated || errCode == http.StatusOK) {
		logCtx.Errorln("Tracked user property creation failed")
		if errCode == http.StatusConflict {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tracked user property already exist"})
			return
		}
		if errCode == http.StatusNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user property Not found"})
			return
		}
		if errCode == http.StatusBadRequest {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tracked User Properties Count Exceeded"})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["id"] = id
	c.JSON(http.StatusCreated, response)
}

type RemoveFactorsTrackedUserPropertyParams struct {
	ID int64 `json:"id" binding:"required"`
}

func getRemoveFactorsTrackedUserPropertyParams(c *gin.Context) (*RemoveFactorsTrackedUserPropertyParams, error) {
	params := RemoveFactorsTrackedUserPropertyParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// RemoveFactorsTrackedUserPropertyHandler - remove a tracked user properrty handler
// RemoveFactorsTrackedUserPropertyHandler godoc
// @Summary Remove a tracked user property
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param remove body v1.RemoveFactorsTrackedUserPropertyParams true "remove"
// @Success 200 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/tracked_user_property/remove [DELETE]
func RemoveFactorsTrackedUserPropertyHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	params, err := getRemoveFactorsTrackedUserPropertyParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode := store.GetStore().RemoveFactorsTrackedUserProperty(params.ID, projectID)
	if errCode != http.StatusOK {
		logCtx.Errorln("Removing Tracked event failed")
		if errCode == http.StatusConflict {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tracked user property already deleted"})
			return
		}
		if errCode == http.StatusNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user property Not found"})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := make(map[string]interface{})
	response["status"] = "success"
	response["id"] = id
	c.JSON(http.StatusOK, response)
}
