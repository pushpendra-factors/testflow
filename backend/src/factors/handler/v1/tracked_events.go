package v1

import (
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetAllFactorsTrackedEventsHandler - Get All tracked events handler
// GetAllFactorsTrackedEventsHandler godoc
// @Summary Get all tracked events
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.FactorsTrackedEvent
// @Router /{project_id}/v1/factors/tracked_event [GET]
func GetAllFactorsTrackedEventsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	trackedEvents, errCode := store.GetStore().GetAllFactorsTrackedEventsByProject(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, trackedEvents)
}

type CreateFactorsTrackedEventParams struct {
	EventName string `json:"event_name" binding:"required"`
}

func getcreateFactorsTrackedEventParams(c *gin.Context) (*CreateFactorsTrackedEventParams, error) {
	params := CreateFactorsTrackedEventParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// CreateFactorsTrackedEventsHandler - Handler for creating tracked event
// CreateFactorsTrackedEventsHandler godoc
// @Summary Create a tracked event
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param create body v1.CreateFactorsTrackedEventParams true "create"
// @Success 201 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/tracked_event [POST]
func CreateFactorsTrackedEventsHandler(c *gin.Context) {
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	params, err := getcreateFactorsTrackedEventParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode := store.GetStore().CreateFactorsTrackedEvent(projectID, params.EventName, loggedInAgentUUID)
	if !(errCode == http.StatusCreated || errCode == http.StatusOK) {
		logCtx.Errorln("Tracked event creation failed")
		if errCode == http.StatusConflict {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tracked Event already exist"})
			return
		}
		if errCode == http.StatusNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Event Not found"})
			return
		}
		if errCode == http.StatusBadRequest {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tracked Events Count Exceeded"})
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

type RemoveFactorsTrackedEventParams struct {
	ID int64 `json:"id" binding:"required"`
}

func getRemoveFactorsTrackedEventParams(c *gin.Context) (*RemoveFactorsTrackedEventParams, error) {
	params := RemoveFactorsTrackedEventParams{}
	err := c.BindJSON(&params)
	if err != nil {
		return nil, err
	}
	return &params, nil
}

// RemoveFactorsTrackedEventsHandler - remove a tracked event handler
// RemoveFactorsTrackedEventsHandler godoc
// @Summary Remove a tracked event
// @Tags V1FactorsApi
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param remove body v1.RemoveFactorsTrackedEventParams true "remove"
// @Success 200 {string} json "{"id": uint64, "status": string}"
// @Router /{project_id}/v1/factors/tracked_event/remove [DELETE]
func RemoveFactorsTrackedEventsHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	logCtx := log.WithFields(log.Fields{
		"projectId": projectID,
	})
	params, err := getRemoveFactorsTrackedEventParams(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id, errCode := store.GetStore().DeactivateFactorsTrackedEvent(params.ID, projectID)
	if errCode != http.StatusOK {
		logCtx.Errorln("Removing Tracked event failed")
		if errCode == http.StatusConflict {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tracked Event already deleted"})
			return
		}
		if errCode == http.StatusNotFound {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Event Not found"})
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
